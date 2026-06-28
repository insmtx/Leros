package provider

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/insmtx/Leros/backend/agent"
)

// SendEngineEvent sends an engine event to the channel, respecting context cancellation.
func SendEngineEvent(ctx context.Context, ch chan<- agent.Event, event agent.Event) bool {
	select {
	case ch <- event:
		return true
	case <-ctx.Done():
		return false
	}
}

// SendEngineEventTo sends a simple engine event with just type and content.
func SendEngineEventTo(ch chan<- agent.Event, eventType agent.EventType, content string) {
	select {
	case ch <- agent.Event{Type: eventType, Content: content}:
	default:
	}
}

// SendEngineEventPayloadTo sends an engine event with a typed payload.
func SendEngineEventPayloadTo(ch chan<- agent.Event, eventType agent.EventType, payload any) {
	raw, _ := json.Marshal(payload)
	select {
	case ch <- agent.Event{Type: eventType, Payload: raw}:
	default:
	}
}

// SendEngineEventDirect sends a pre-built engine event directly to the channel.
func SendEngineEventDirect(ch chan<- agent.Event, evt *agent.Event) {
	if evt == nil {
		return
	}
	select {
	case ch <- *evt:
	default:
	}
}

// ScanPlainEngineOutput reads plain text output and converts to engine events.
func ScanPlainEngineOutput(ctx context.Context, r interface{ Read([]byte) (int, error) }, evtChan chan<- agent.Event, eventType agent.EventType) string {
	var output strings.Builder
	// Use a simple message ID counter for plain output.
	var msgSeq int
	ScanJSONLines(r, func(line string) bool {
		line = strings.TrimSpace(line)
		if line == "" {
			return true
		}
		if output.Len() > 0 {
			output.WriteString("\n")
		}
		output.WriteString(line)
		if eventType == EngineEventMessageDelta {
			msgSeq++
			msgID := "msg_" + itoa(msgSeq)
			payload, _ := json.Marshal(engineDeltaPayload{
				MessageID: msgID,
				Role:      "assistant",
				Content:   line,
			})
			select {
			case evtChan <- agent.Event{Type: EngineEventMessageDelta, Payload: payload, Content: line}:
			case <-ctx.Done():
				return false
			}
		} else {
			select {
			case evtChan <- agent.Event{Type: eventType, Content: line}:
			case <-ctx.Done():
				return false
			}
		}
		return true
	})
	return output.String()
}

type engineDeltaPayload struct {
	MessageID string `json:"message_id,omitempty"`
	Role      string `json:"role,omitempty"`
	Content   string `json:"content"`
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := make([]byte, 0, 20)
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}
