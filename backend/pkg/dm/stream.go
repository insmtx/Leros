package dm

// StreamEventType 表示 Worker 执行流中的事件类型。
type StreamEventType string

const (
	// StreamEventRunStarted 表示一次运行已经开始。
	StreamEventRunStarted StreamEventType = "run.started"
	// StreamEventMessageDelta 表示助手输出的增量文本。
	StreamEventMessageDelta StreamEventType = "message.delta"
	// StreamEventToolCallStarted 表示工具调用已经开始。
	StreamEventToolCallStarted StreamEventType = "tool_call.started"
	// StreamEventToolCallDelta 表示流式工具调用参数片段。
	StreamEventToolCallDelta StreamEventType = "tool_call.delta"
	// StreamEventToolCallFinished 表示工具调用已经结束。
	StreamEventToolCallFinished StreamEventType = "tool_call.finished"
	// StreamEventMessageCompleted 表示最终助手消息已经生成。
	StreamEventMessageCompleted StreamEventType = "message.completed"
	// StreamEventRunCompleted 表示一次运行成功完成。
	StreamEventRunCompleted StreamEventType = "run.completed"
	// StreamEventRunFailed 表示一次运行失败。
	StreamEventRunFailed StreamEventType = "run.failed"
)

// MessageStreamMessage 是 Worker 发送给 Server 并转发到 UI 的流式消息协议。
type MessageStreamMessage = Envelope[StreamBody]

// StreamBody 是 Worker 到 Server 再到 UI 的单个流式事件载荷。
type StreamBody struct {
	Seq     int64           `json:"seq"`
	Event   StreamEventType `json:"event"`
	Payload StreamPayload   `json:"payload"`

	Usage *UsagePayload `json:"usage,omitempty"`
	Error *StreamError  `json:"error,omitempty"`
}

// StreamPayload 承载流式事件的具体内容。
type StreamPayload struct {
	Role       MessageRole      `json:"role,omitempty"`
	Content    string           `json:"content,omitempty"`
	ToolCall   *ToolCallEvent   `json:"tool_call,omitempty"`
	ToolResult *ToolResultEvent `json:"tool_result,omitempty"`
}

// ToolCallEvent 描述流式事件中的工具调用。
type ToolCallEvent struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

// ToolResultEvent 描述流式事件中的工具执行结果。
type ToolResultEvent struct {
	ToolCallID string         `json:"tool_call_id"`
	Name       string         `json:"name,omitempty"`
	Result     map[string]any `json:"result,omitempty"`
}

// UsagePayload 描述模型用量信息。
type UsagePayload struct {
	InputTokens  int `json:"input_tokens,omitempty"`
	OutputTokens int `json:"output_tokens,omitempty"`
	TotalTokens  int `json:"total_tokens,omitempty"`
}

// StreamError 描述流式执行中的终止错误或可恢复错误。
type StreamError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}
