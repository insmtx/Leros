package server

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WorkerConnection struct {
	ID         string
	Conn       *websocket.Conn
	Send       chan map[string]interface{}
	Status     string
	Registered time.Time
	LastSeen   time.Time
	mu         sync.RWMutex
}

func (wc *WorkerConnection) SendJSON(msg map[string]interface{}) error {
	wc.mu.RLock()
	defer wc.mu.RUnlock()
	return wc.Conn.WriteJSON(msg)
}
