package agent

import (
	"encoding/json"
	"time"
)

// RawPayload stores structured event payload JSON while keeping Event easy to marshal.
type RawPayload json.RawMessage

// MarshalJSON serializes the raw payload bytes.
func (p RawPayload) MarshalJSON() ([]byte, error) {
	if len(p) == 0 {
		return []byte("null"), nil
	}
	return p, nil
}

// UnmarshalJSON stores the raw payload bytes.
func (p *RawPayload) UnmarshalJSON(data []byte) error {
	if p == nil {
		return nil
	}
	if string(data) == "null" {
		*p = nil
		return nil
	}
	*p = append((*p)[0:0], data...)
	return nil
}

// EventType identifies an observable runtime event emitted during execution.
type EventType string

// Event is the stable runtime event envelope emitted during execution.
type Event struct {
	ID        string     `json:"id,omitempty"`
	RunID     string     `json:"run_id,omitempty"`
	TraceID   string     `json:"trace_id,omitempty"`
	Seq       int64      `json:"seq,omitempty"`
	Type      EventType  `json:"type"`
	CreatedAt time.Time  `json:"created_at,omitempty"`
	Payload   RawPayload `json:"payload,omitempty"`
	Content   string     `json:"content,omitempty"`
}

// Usage describes model token usage when available.
type Usage struct {
	InputTokens  int `json:"input_tokens,omitempty"`
	OutputTokens int `json:"output_tokens,omitempty"`
	TotalTokens  int `json:"total_tokens,omitempty"`
}

// RunStatus is the final status returned from Run.
type RunStatus string

const (
	RunStatusCompleted RunStatus = "completed"
	RunStatusFailed    RunStatus = "failed"
	RunStatusCancelled RunStatus = "cancelled"
)

// RunResult is the final result of one agent run.
type RunResult struct {
	RunID       string           `json:"run_id"`
	TraceID     string           `json:"trace_id,omitempty"`
	Status      RunStatus        `json:"status"`
	Message     string           `json:"message,omitempty"`
	Error       string           `json:"error,omitempty"`
	Usage       *Usage           `json:"usage,omitempty"`
	ToolCalls   []ToolCallRecord `json:"tool_calls,omitempty"`
	Metadata    map[string]any   `json:"metadata,omitempty"`
	StartedAt   time.Time        `json:"started_at"`
	CompletedAt time.Time        `json:"completed_at,omitempty"`
}

// ToolCallRecord is a compact final tool call summary.
type ToolCallRecord struct {
	CallID string         `json:"call_id,omitempty"`
	Name   string         `json:"name,omitempty"`
	Result map[string]any `json:"result,omitempty"`
	Error  string         `json:"error,omitempty"`
}
