// Package dm 定义 SingerOS 服务之间共享的领域消息协议。
package dm

import "time"

// MessageType 表示领域消息的顶层类型。
type MessageType string

const (
	// MessageTypeWorkerTask 表示 Server 发送给 Worker 的任务消息。
	MessageTypeWorkerTask MessageType = "worker.task"
	// MessageTypeStream 表示 Worker 发送给 Server 并转发到 UI 的流式消息。
	MessageTypeStream MessageType = "message.stream"
)

// TraceContext 承载跨 UI、Server、Worker 和 Runtime 的链路追踪标识。
type TraceContext struct {
	TraceID   string `json:"trace_id"`
	RequestID string `json:"request_id,omitempty"`
	TaskID    string `json:"task_id,omitempty"`
	RunID     string `json:"run_id,omitempty"`
	ParentID  string `json:"parent_id,omitempty"`
}

// RouteContext 承载消息投递和租户隔离所需的路由信息。
type RouteContext struct {
	OrgID     string `json:"org_id"`
	SessionID string `json:"session_id,omitempty"`
	WorkerID  string `json:"worker_id,omitempty"`
}

// Envelope 是 MQ topic 上使用的通用领域消息信封。
type Envelope[T any] struct {
	ID        string      `json:"id"`
	Type      MessageType `json:"type"`
	CreatedAt time.Time   `json:"created_at"`

	Trace TraceContext `json:"trace"`
	Route RouteContext `json:"route"`

	Body     T              `json:"body"`
	Metadata map[string]any `json:"metadata,omitempty"`
}
