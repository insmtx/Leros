package eventtypes

// StreamEventType represents event types in Worker execution streams.
type StreamEventType string

const (
	// StreamEventRunStarted indicates a run has started.
	StreamEventRunStarted StreamEventType = "run.started"
	// StreamEventMessageDelta indicates incremental text output from assistant.
	StreamEventMessageDelta StreamEventType = "message.delta"
	// StreamEventToolCallStarted indicates a tool call has started.
	StreamEventToolCallStarted StreamEventType = "tool_call.started"
	// StreamEventToolCallDelta represents streaming tool call argument fragments.
	StreamEventToolCallDelta StreamEventType = "tool_call.delta"
	// StreamEventToolCallFinished indicates a tool call has finished.
	StreamEventToolCallFinished StreamEventType = "tool_call.finished"
	// StreamEventMessageCompleted indicates the final assistant message is generated.
	StreamEventMessageCompleted StreamEventType = "message.completed"
	// StreamEventRunCompleted indicates a run completed successfully.
	StreamEventRunCompleted StreamEventType = "run.completed"
	// StreamEventRunFailed indicates a run failed.
	StreamEventRunFailed StreamEventType = "run.failed"
)

// MessageStreamMessage is the stream message protocol from Worker to Server (forwarded to UI).
type MessageStreamMessage = Envelope[StreamBody]

// StreamBody is a single streaming event payload from Worker to Server to UI.
type StreamBody struct {
	Seq     int64           `json:"seq"`
	Event   StreamEventType `json:"event"`
	Payload StreamPayload   `json:"payload"`

	Usage *UsagePayload `json:"usage,omitempty"`
	Error *StreamError  `json:"error,omitempty"`
}

// StreamPayload carries the specific content of streaming events.
type StreamPayload struct {
	Role       MessageRole      `json:"role,omitempty"`
	Content    string           `json:"content,omitempty"`
	ToolCall   *ToolCallEvent   `json:"tool_call,omitempty"`
	ToolResult *ToolResultEvent `json:"tool_result,omitempty"`
}

// ToolCallEvent describes a tool call in streaming events.
type ToolCallEvent struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

// ToolResultEvent describes a tool execution result in streaming events.
type ToolResultEvent struct {
	ToolCallID string         `json:"tool_call_id"`
	Name       string         `json:"name,omitempty"`
	Result     map[string]any `json:"result,omitempty"`
}

// UsagePayload describes model usage information.
type UsagePayload struct {
	InputTokens  int `json:"input_tokens,omitempty"`
	OutputTokens int `json:"output_tokens,omitempty"`
	TotalTokens  int `json:"total_tokens,omitempty"`
}

// StreamError describes terminal or recoverable errors in streaming execution.
type StreamError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}
