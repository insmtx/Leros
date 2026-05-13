package events

import "time"

// MessageType represents the top-level type of domain messages.
type MessageType string

const (
	// MessageTypeWorkerTask represents task messages from Server to Worker.
	MessageTypeWorkerTask MessageType = "worker.task"
	// MessageTypeStream represents stream messages from Worker to Server (forwarded to UI).
	MessageTypeStream MessageType = "message.stream"
)

// TraceContext carries distributed tracing identifiers across UI, Server, Worker, and Runtime.
type TraceContext struct {
	TraceID   string `json:"trace_id"`
	RequestID string `json:"request_id,omitempty"`
	TaskID    string `json:"task_id,omitempty"`
	RunID     string `json:"run_id,omitempty"`
	ParentID  string `json:"parent_id,omitempty"`
}

// RouteContext carries routing information for message delivery and tenant isolation.
type RouteContext struct {
	OrgID     uint   `json:"org_id"`
	SessionID string `json:"session_id,omitempty"`
	WorkerID  uint   `json:"worker_id,omitempty"`
}

// Envelope is the generic domain message envelope used on MQ topics.
type Envelope[T any] struct {
	ID        string      `json:"id"`
	Type      MessageType `json:"type"`
	CreatedAt time.Time   `json:"created_at"`

	Trace TraceContext `json:"trace"`
	Route RouteContext `json:"route"`

	Body     T              `json:"body"`
	Metadata map[string]any `json:"metadata,omitempty"`
}
