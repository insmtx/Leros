package dto

type SessionEventType string

const (
	SessionEventTypeMessageDelta    SessionEventType = "message.delta"
	SessionEventTypeMessageComplete SessionEventType = "message.complete"
	SessionEventTypeRunStarted      SessionEventType = "run.started"
	SessionEventTypeRunCompleted    SessionEventType = "run.completed"
	SessionEventTypeRunFailed       SessionEventType = "run.failed"
	SessionEventTypeToolCallStarted SessionEventType = "tool_call.started"
	SessionEventTypeToolCallDelta   SessionEventType = "tool_call.delta"
	SessionEventTypeToolCallResult  SessionEventType = "tool_call.result"
)

type SessionEvent struct {
	Type      SessionEventType `json:"type"`
	SessionID string           `json:"session_id"`
	Payload   interface{}      `json:"payload"`
	Sequence  int64            `json:"sequence"`
	Timestamp int64            `json:"timestamp"` // Unix timestamp in milliseconds
}

type MessageDeltaPayload struct {
	Role    string `json:"role"`
	Content string `json:"content"` // 增量文本
}

type MessageCompletePayload struct {
	ID        string                 `json:"id"`
	Role      string                 `json:"role"`
	Content   string                 `json:"content"`
	Status    string                 `json:"status"`
	ToolCalls []ToolCallResponse     `json:"tool_calls,omitempty"`
	Thinking  string                 `json:"thinking,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type RunStatusPayload struct {
	Status  string `json:"status"`
	RunID   string `json:"run_id,omitempty"`
	Message string `json:"message,omitempty"`
}

type ToolCallResponse struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
	Status    string                 `json:"status"`
	Result    interface{}            `json:"result,omitempty"`
	Duration  int                    `json:"duration,omitempty"`
}

type ToolCallDeltaPayload struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name,omitempty"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type ToolCallResultPayload struct {
	ID     string      `json:"id"`
	Name   string      `json:"name"`
	Result interface{} `json:"result"`
	Status string      `json:"status"` // success | error
}
