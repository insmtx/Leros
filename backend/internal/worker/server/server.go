package server

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	infradb "github.com/insmtx/Leros/backend/internal/infra/db"
	"github.com/insmtx/Leros/backend/internal/worker"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WorkerManager struct {
	workers   map[string]*WorkerConnection
	mu        sync.RWMutex
	scheduler worker.WorkerScheduler
	db        *gorm.DB
}

func NewServer(scheduler worker.WorkerScheduler, db *gorm.DB) *WorkerManager {
	return &WorkerManager{
		workers:   make(map[string]*WorkerConnection),
		scheduler: scheduler,
		db:        db,
	}
}

func (s *WorkerManager) RegisterRoutes(r gin.IRouter) {
	r.GET("/ws/worker", s.handleWorkerWebSocket)
	r.POST("/ListWorkers", s.listWorkers)
	r.POST("/GetWorkerInfo", s.getWorkerInfo)
	r.POST("/ShutdownWorker", s.shutdownWorker)
	r.POST("/CreateWorker", s.createWorker)
}

func (s *WorkerManager) handleWorkerWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logs.Errorf("Failed to upgrade WebSocket: %v", err)
		return
	}
	defer conn.Close()

	ctx := c.Request.Context()

	var workerID string
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			logs.Errorf("Failed to read registration message: %v", err)
			return
		}

		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err != nil {
			logs.Errorf("Failed to parse message: %v", err)
			continue
		}

		if msgType, ok := msg["type"].(string); ok && msgType == "worker_register" {
			if payload, ok := msg["payload"].(map[string]interface{}); ok {
				if id, ok := payload["worker_id"].(string); ok {
					workerID = id
					break
				}
			}
		}
	}

	worker := &WorkerConnection{
		ID:         workerID,
		Conn:       conn,
		Send:       make(chan map[string]interface{}, 256),
		Status:     "active",
		Registered: time.Now(),
		LastSeen:   time.Now(),
	}

	s.mu.Lock()
	s.workers[workerID] = worker
	s.mu.Unlock()

	logs.Infof("Worker %s registered", workerID)

	welcomeMsg := map[string]interface{}{
		"type": "welcome",
		"payload": map[string]interface{}{
			"message":   "Connected to Leros worker server",
			"worker_id": workerID,
		},
	}
	if err := worker.SendJSON(welcomeMsg); err != nil {
		logs.Errorf("Failed to send welcome message: %v", err)
		return
	}

	go s.readPump(worker)
	go s.writePump(worker)
	go s.heartbeatChecker(worker)

	<-ctx.Done()
}

func (s *WorkerManager) readPump(worker *WorkerConnection) {
	defer func() {
		s.unregisterWorker(worker.ID)
		worker.Conn.Close()
	}()

	worker.Conn.SetReadLimit(512 * 1024)
	worker.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	worker.Conn.SetPongHandler(func(string) error {
		worker.mu.Lock()
		worker.LastSeen = time.Now()
		worker.mu.Unlock()
		worker.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := worker.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logs.Errorf("Worker %s WebSocket error: %v", worker.ID, err)
			}
			break
		}

		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err != nil {
			logs.Errorf("Failed to unmarshal message from worker %s: %v", worker.ID, err)
			continue
		}

		s.handleWorkerMessage(worker, msg)
	}
}

func (s *WorkerManager) writePump(worker *WorkerConnection) {
	ticker := time.NewTicker(54 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case message, ok := <-worker.Send:
			if !ok {
				worker.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			worker.mu.RLock()
			err := worker.Conn.WriteJSON(message)
			worker.mu.RUnlock()

			if err != nil {
				logs.Errorf("Failed to write to worker %s: %v", worker.ID, err)
				return
			}
		case <-ticker.C:
			worker.mu.RLock()
			worker.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			err := worker.Conn.WriteMessage(websocket.PingMessage, nil)
			worker.mu.RUnlock()

			if err != nil {
				return
			}
		}
	}
}

func (s *WorkerManager) handleWorkerMessage(worker *WorkerConnection, msg map[string]interface{}) {
	msgType, _ := msg["type"].(string)

	switch msgType {
	case "heartbeat":
		worker.mu.Lock()
		worker.LastSeen = time.Now()
		worker.Status = "active"
		worker.mu.Unlock()

		ack := map[string]interface{}{
			"type": "heartbeat_ack",
			"payload": map[string]interface{}{
				"timestamp": time.Now().Unix(),
			},
		}
		select {
		case worker.Send <- ack:
		default:
			logs.Warnf("Heartbeat ack dropped for worker %s", worker.ID)
		}

	case "worker_status":
		if payload, ok := msg["payload"].(map[string]interface{}); ok {
			if status, ok := payload["status"].(string); ok {
				worker.mu.Lock()
				worker.Status = status
				worker.mu.Unlock()
			}
		}

	case "getConfig":
		s.handleGetConfig(worker, msg)
	}
}

func (s *WorkerManager) handleGetConfig(worker *WorkerConnection, msg map[string]interface{}) {
	assistantCode := ""
	if payload, ok := msg["payload"].(map[string]interface{}); ok {
		if code, ok := payload["assistant_code"].(string); ok {
			assistantCode = code
		}
	}

	if assistantCode == "" {
		resp := map[string]interface{}{
			"type": "configResponse",
			"payload": map[string]interface{}{
				"config": nil,
				"error":  "assistant_code is required",
			},
		}
		select {
		case worker.Send <- resp:
		default:
			logs.Warnf("Config response dropped for worker %s", worker.ID)
		}
		return
	}

	if s.db == nil {
		resp := map[string]interface{}{
			"type": "configResponse",
			"payload": map[string]interface{}{
				"config": nil,
				"error":  "database not available",
			},
		}
		select {
		case worker.Send <- resp:
		default:
			logs.Warnf("Config response dropped for worker %s", worker.ID)
		}
		return
	}

	da, err := infradb.GetDigitalAssistantByCode(context.Background(), s.db, assistantCode)
	if err != nil {
		logs.Errorf("Failed to get digital assistant %s: %v", assistantCode, err)
		resp := map[string]interface{}{
			"type": "configResponse",
			"payload": map[string]interface{}{
				"config": nil,
				"error":  err.Error(),
			},
		}
		select {
		case worker.Send <- resp:
		default:
			logs.Warnf("Config response dropped for worker %s", worker.ID)
		}
		return
	}

	if da == nil {
		logs.Warnf("Digital assistant %s not found", assistantCode)
		resp := map[string]interface{}{
			"type": "configResponse",
			"payload": map[string]interface{}{
				"config": nil,
				"error":  "digital assistant not found",
			},
		}
		select {
		case worker.Send <- resp:
		default:
			logs.Warnf("Config response dropped for worker %s", worker.ID)
		}
		return
	}

	resp := map[string]interface{}{
		"type": "configResponse",
		"payload": map[string]interface{}{
			"config": da.Config,
			"error":  nil,
		},
	}
	select {
	case worker.Send <- resp:
		logs.Infof("Config sent to worker %s for assistant %s", worker.ID, assistantCode)
	default:
		logs.Warnf("Config response dropped for worker %s", worker.ID)
	}
}

func (s *WorkerManager) unregisterWorker(workerID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if worker, ok := s.workers[workerID]; ok {
		delete(s.workers, workerID)
		close(worker.Send)
		logs.Infof("Worker %s unregistered", workerID)
	}
}

func (s *WorkerManager) heartbeatChecker(worker *WorkerConnection) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		worker.mu.RLock()
		lastSeen := worker.LastSeen
		worker.mu.RUnlock()

		if time.Since(lastSeen) > 90*time.Second {
			logs.Warnf("Worker %s heartbeat timeout", worker.ID)
			s.unregisterWorker(worker.ID)
			worker.Conn.Close()
			return
		}
	}
}

type ListWorkersResponse struct {
	Workers []WorkerInfo `json:"workers"`
	Total   int          `json:"total"`
}

type WorkerInfo struct {
	ID         string    `json:"id"`
	Status     string    `json:"status"`
	Registered time.Time `json:"registered"`
	LastSeen   time.Time `json:"last_seen"`
}

func (s *WorkerManager) listWorkers(c *gin.Context) {
	var req struct{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s.mu.RLock()
	workers := make([]WorkerInfo, 0, len(s.workers))
	for _, worker := range s.workers {
		worker.mu.RLock()
		workers = append(workers, WorkerInfo{
			ID:         worker.ID,
			Status:     worker.Status,
			Registered: worker.Registered,
			LastSeen:   worker.LastSeen,
		})
		worker.mu.RUnlock()
	}
	s.mu.RUnlock()

	c.JSON(http.StatusOK, ListWorkersResponse{
		Workers: workers,
		Total:   len(workers),
	})
}

type GetWorkerInfoRequest struct {
	WorkerID string `json:"worker_id" binding:"required"`
}

func (s *WorkerManager) getWorkerInfo(c *gin.Context) {
	var req GetWorkerInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s.mu.RLock()
	worker, ok := s.workers[req.WorkerID]
	s.mu.RUnlock()

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "worker not found"})
		return
	}

	worker.mu.RLock()
	info := WorkerInfo{
		ID:         worker.ID,
		Status:     worker.Status,
		Registered: worker.Registered,
		LastSeen:   worker.LastSeen,
	}
	worker.mu.RUnlock()

	c.JSON(http.StatusOK, info)
}

type ShutdownWorkerRequest struct {
	WorkerID string `json:"worker_id" binding:"required"`
}

func (s *WorkerManager) shutdownWorker(c *gin.Context) {
	var req ShutdownWorkerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s.mu.RLock()
	worker, ok := s.workers[req.WorkerID]
	s.mu.RUnlock()

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "worker not found"})
		return
	}

	shutdownMsg := map[string]interface{}{
		"type": "shutdown",
		"payload": map[string]interface{}{
			"timestamp": time.Now().Unix(),
		},
	}

	select {
	case worker.Send <- shutdownMsg:
		c.JSON(http.StatusOK, gin.H{"message": "shutdown command sent"})
	default:
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "worker send buffer full"})
	}
}

type CreateWorkerRequest struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	EnvType     string            `json:"env_type"`
	Image       string            `json:"image"`
	Command     []string          `json:"command"`
	Args        []string          `json:"args"`
	Env         map[string]string `json:"env"`
	WorkingDir  string            `json:"working_dir"`
}

func (s *WorkerManager) createWorker(c *gin.Context) {
	var req CreateWorkerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if s.scheduler == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "worker scheduler not initialized"})
		return
	}

	spec := &worker.WorkerSpec{
		ID:          req.ID,
		Name:        req.Name,
		Labels:      req.Labels,
		Annotations: req.Annotations,
		EnvType:     worker.WorkerEnvType(req.EnvType),
		Image:       req.Image,
		Command:     req.Command,
		Args:        req.Args,
		Env:         req.Env,
		WorkingDir:  req.WorkingDir,
	}

	spec.EnvType = worker.WorkerEnvProcess

	instance, err := s.scheduler.Start(c.Request.Context(), spec)
	if err != nil {
		logs.Errorf("Failed to create worker: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, instance)
}
