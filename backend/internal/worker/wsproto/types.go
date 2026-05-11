package wsproto

import (
	"encoding/json"
	"fmt"

	"github.com/insmtx/Leros/backend/types"
)

// WebSocket Message Type Constants
const (
	// Client -> Server
	MsgTypeWorkerRegister = "worker_register"
	MsgTypeHeartbeat      = "heartbeat"
	MsgTypeGetConfig      = "getConfig"
	MsgTypeWorkerStatus   = "worker_status"

	// Server -> Client
	MsgTypeWelcome        = "welcome"
	MsgTypeHeartbeatAck   = "heartbeat_ack"
	MsgTypeConfigResponse = "configResponse"
	MsgTypeShutdown       = "shutdown"
	MsgTypeConfigUpdate   = "config_update"
)

// WSMessage defines the standard WebSocket message structure
// 所有 Server-Client 交互都使用此统一结构
type WSMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// GetPayload unmarshals the payload into the provided interface
func (m *WSMessage) GetPayload(v interface{}) error {
	if m.Payload == nil {
		return fmt.Errorf("payload is nil")
	}
	return json.Unmarshal(m.Payload, v)
}

// SetPayload marshals the provided interface into the payload
func (m *WSMessage) SetPayload(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	m.Payload = data
	return nil
}

// ===== Client -> Server Messages =====

// RegisterPayload defines the payload for worker registration (Client -> Server)
type RegisterPayload struct {
	WorkerID  string `json:"worker_id"`
	Timestamp int64  `json:"timestamp"`
}

// HeartbeatPayload defines the payload for heartbeat message (Client -> Server)
type HeartbeatPayload struct {
	Timestamp int64 `json:"timestamp"`
}

// GetConfigPayload defines the payload for config request (Client -> Server)
type GetConfigPayload struct {
	WorkerID uint `json:"worker_id"`
}

// WorkerStatusPayload defines the payload for worker status update (Client -> Server)
type WorkerStatusPayload struct {
	Status string `json:"status"`
}

// ===== Server -> Client Messages =====

// WelcomePayload defines the payload for welcome message (Server -> Client)
type WelcomePayload struct {
	Message  string `json:"message"`
	WorkerID string `json:"worker_id"`
}

// HeartbeatAckPayload defines the payload for heartbeat acknowledgment (Server -> Client)
type HeartbeatAckPayload struct {
	Timestamp int64 `json:"timestamp"`
}

// ConfigResponsePayload defines the payload for config response (Server -> Client)
type ConfigResponsePayload struct {
	Config *types.AssistantConfig `json:"config"`
	Error  string                 `json:"error,omitempty"`
}

// ShutdownPayload defines the payload for shutdown message (Server -> Client)
type ShutdownPayload struct {
	Timestamp int64 `json:"timestamp"`
}

// ConfigUpdatePayload defines the payload for config update notification (Server -> Client)
type ConfigUpdatePayload struct {
	WorkerID uint `json:"worker_id"`
	Version  int  `json:"version"`
}

// NewPayload creates a new WSMessage with the given type and payload
func NewPayload(msgType string, payload interface{}) (*WSMessage, error) {
	msg := &WSMessage{Type: msgType}
	if err := msg.SetPayload(payload); err != nil {
		return nil, err
	}
	return msg, nil
}
