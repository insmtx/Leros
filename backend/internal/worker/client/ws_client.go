package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ygpkg/yg-go/logs"

	"github.com/insmtx/SingerOS/backend/internal/worker/wsproto"
)

type WSClient struct {
	conn          *websocket.Conn
	workerID      string
	serverAddr    string
	send          chan *wsproto.WSMessage
	ctx           context.Context
	cancel        context.CancelFunc
	onConfigReady func(config map[string]interface{})
}

type WSClientOption func(*WSClient)

func WithWorkerID(workerID string) WSClientOption {
	return func(c *WSClient) {
		c.workerID = workerID
	}
}

func WithOnConfigReady(handler func(map[string]interface{})) WSClientOption {
	return func(c *WSClient) {
		c.onConfigReady = handler
	}
}

func NewWSClient(serverAddr, workerID string, opts ...WSClientOption) *WSClient {
	ctx, cancel := context.WithCancel(context.Background())
	c := &WSClient{
		workerID:   workerID,
		serverAddr: serverAddr,
		send:       make(chan *wsproto.WSMessage, 256),
		ctx:        ctx,
		cancel:     cancel,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *WSClient) Connect(ctx context.Context) error {
	wsURL := fmt.Sprintf("ws://%s/ws/worker", c.serverAddr)
	logs.Infof("Connecting to server WebSocket: %s", wsURL)

	dialer := websocket.DefaultDialer
	conn, _, err := dialer.DialContext(ctx, wsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to dial WebSocket: %w", err)
	}

	c.conn = conn
	logs.Infof("Connected to server successfully")

	registerMsg, err := wsproto.NewPayload(wsproto.MsgTypeWorkerRegister, wsproto.RegisterPayload{
		WorkerID:  c.workerID,
		Timestamp: time.Now().Unix(),
	})
	if err != nil {
		return fmt.Errorf("failed to create registration payload: %w", err)
	}

	if err := c.sendWSMessage(registerMsg); err != nil {
		return fmt.Errorf("failed to send registration: %w", err)
	}

	go c.readLoop(ctx)
	go c.writeLoop(ctx)

	return nil
}

func (c *WSClient) readLoop(ctx context.Context) {
	defer func() {
		logs.Info("WebSocket read loop exited")
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					logs.Info("WebSocket connection closed")
					return
				}
				logs.Errorf("WebSocket read error: %v", err)
				return
			}

			var msg wsproto.WSMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				logs.Errorf("Failed to unmarshal message: %v", err)
				continue
			}

			c.handleMessage(&msg)
		}
	}
}

func (c *WSClient) writeLoop(ctx context.Context) {
	defer func() {
		logs.Info("WebSocket write loop exited")
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-c.send:
			if !ok {
				return
			}
			if err := c.sendWSMessage(msg); err != nil {
				logs.Errorf("Failed to send message: %v", err)
				return
			}
		}
	}
}

func (c *WSClient) handleMessage(msg *wsproto.WSMessage) {
	switch msg.Type {
	case wsproto.MsgTypeWelcome:
		logs.Infof("Received welcome from server")
		if c.workerID != "" {
			c.requestConfig()
		}
	case wsproto.MsgTypeConfigResponse:
		c.handleConfigResponse(msg)
	case wsproto.MsgTypeConfigUpdate:
		logs.Infof("Received config update")
	default:
		logs.Debugf("Received message: %s", msg.Type)
	}
}

func (c *WSClient) requestConfig() {
	reqMsg, err := wsproto.NewPayload(wsproto.MsgTypeGetConfig, wsproto.GetConfigPayload{
		WorkerID: c.workerID,
	})
	if err != nil {
		logs.Errorf("Failed to create config request payload: %v", err)
		return
	}

	if err := c.sendWSMessage(reqMsg); err != nil {
		logs.Errorf("Failed to request config: %v", err)
	} else {
		logs.Infof("Requested config for worker %s", c.workerID)
	}
}

func (c *WSClient) handleConfigResponse(msg *wsproto.WSMessage) {
	var payload wsproto.ConfigResponsePayload
	if err := msg.GetPayload(&payload); err != nil {
		logs.Errorf("Failed to unmarshal config response payload: %v", err)
		return
	}

	if payload.Error != "" {
		logs.Errorf("Config response error: %s", payload.Error)
		return
	}

	if payload.Config == nil {
		logs.Errorf("Config not found in response")
		return
	}

	if c.onConfigReady != nil {
		configMap := make(map[string]interface{})
		bytes, _ := json.Marshal(payload.Config)
		json.Unmarshal(bytes, &configMap)
		c.onConfigReady(configMap)
		logs.Info("Config processed successfully")
	} else {
		logs.Warn("No config handler registered")
	}
}

func (c *WSClient) sendWSMessage(msg *wsproto.WSMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return c.conn.WriteMessage(websocket.TextMessage, data)
}

func (c *WSClient) Close() error {
	c.cancel()
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *WSClient) IsConnected() bool {
	return c.conn != nil
}
