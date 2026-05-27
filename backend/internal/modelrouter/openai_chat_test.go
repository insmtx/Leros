package modelrouter

import (
	"encoding/json"
	"testing"
)

func TestConvertAnthropicToolResultsToSeparateChatToolMessages(t *testing.T) {
	input := []byte(`{
		"model": "alias",
		"max_tokens": 1024,
		"messages": [
			{
				"role": "assistant",
				"content": [
					{"type": "tool_use", "id": "call_1", "name": "Write", "input": {"file_path": "a.md"}},
					{"type": "tool_use", "id": "call_2", "name": "Write", "input": {"file_path": "b.md"}}
				]
			},
			{
				"role": "user",
				"content": [
					{"type": "tool_result", "tool_use_id": "call_1", "content": "created a.md"},
					{"type": "tool_result", "tool_use_id": "call_2", "content": "created b.md"}
				]
			}
		]
	}`)

	converted, err := convertRequest(input, ProtocolAnthropicMessages, ProtocolOpenAIChat, "gpt-test")
	if err != nil {
		t.Fatalf("convertRequest() error = %v", err)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(converted, &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	messages, ok := body["messages"].([]interface{})
	if !ok {
		t.Fatalf("messages = %T, want []interface{}", body["messages"])
	}
	if len(messages) != 3 {
		t.Fatalf("len(messages) = %d, want assistant plus two tool messages: %#v", len(messages), messages)
	}

	assistant, ok := messages[0].(map[string]interface{})
	if !ok {
		t.Fatalf("messages[0] = %T, want object", messages[0])
	}
	if got := assistant["role"]; got != "assistant" {
		t.Fatalf("messages[0].role = %v, want assistant", got)
	}
	if toolCalls, ok := assistant["tool_calls"].([]interface{}); !ok || len(toolCalls) != 2 {
		t.Fatalf("messages[0].tool_calls = %#v, want two tool calls", assistant["tool_calls"])
	}

	for i, wantID := range []string{"call_1", "call_2"} {
		msg, ok := messages[i+1].(map[string]interface{})
		if !ok {
			t.Fatalf("messages[%d] = %T, want object", i+1, messages[i+1])
		}
		if got := msg["role"]; got != "tool" {
			t.Fatalf("messages[%d].role = %v, want tool", i+1, got)
		}
		if got := msg["tool_call_id"]; got != wantID {
			t.Fatalf("messages[%d].tool_call_id = %v, want %s", i+1, got, wantID)
		}
	}
}

func TestConvertChatStreamToAnthropicIgnoresDuplicateToolStart(t *testing.T) {
	state := newStreamConversionState()
	start := []byte(`{"id":"chatcmpl-1","object":"chat.completion.chunk","created":1,"model":"gpt-test","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"id":"call_1","type":"function","function":{"name":"Write","arguments":""}}]},"finish_reason":null}]}`)

	first, err := convertStreamEventWithState(start, ProtocolAnthropicMessages, ProtocolOpenAIChat, state)
	if err != nil {
		t.Fatalf("convert first start: %v", err)
	}
	if len(first) != 1 {
		t.Fatalf("len(first) = %d, want one content_block_start", len(first))
	}

	second, err := convertStreamEventWithState(start, ProtocolAnthropicMessages, ProtocolOpenAIChat, state)
	if err != nil {
		t.Fatalf("convert duplicate start: %v", err)
	}
	if len(second) != 0 {
		t.Fatalf("len(second) = %d, want duplicate start ignored: %s", len(second), string(second[0]))
	}
}
