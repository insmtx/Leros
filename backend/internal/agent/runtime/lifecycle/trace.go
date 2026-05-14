package lifecycle

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"github.com/insmtx/Leros/backend/internal/agent/runtime/events"
)

// RunTrace 记录一次运行中与自我学习判断相关的事实。
type RunTrace struct {
	ToolCalls     int
	ToolFailures  int
	ToolNames     []string
	UsedSkillTool bool
}

type traceRecorder struct {
	mu            sync.Mutex
	toolCalls     int
	toolFailures  int
	toolNames     []string
	usedSkillTool bool
	maxSeq        int64
}

func (r *traceRecorder) observe(event *events.Event) {
	if r == nil || event == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	if event.Seq > r.maxSeq {
		r.maxSeq = event.Seq
	}

	switch event.Type {
	case events.EventToolCallStarted:
		r.toolCalls++
		if name := toolNameFromEventContent(event.Content); name != "" {
			r.toolNames = append(r.toolNames, name)
			if name == "skill_use" {
				r.usedSkillTool = true
			}
		}
	case events.EventToolCallFailed:
		r.toolFailures++
	}
}

func (r *traceRecorder) nextSeq() int64 {
	if r == nil {
		return 1
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	r.maxSeq++
	return r.maxSeq
}

func (r *traceRecorder) trace() *RunTrace {
	if r == nil {
		return &RunTrace{}
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	return &RunTrace{
		ToolCalls:     r.toolCalls,
		ToolFailures:  r.toolFailures,
		ToolNames:     append([]string{}, r.toolNames...),
		UsedSkillTool: r.usedSkillTool,
	}
}

type traceSink struct {
	next     events.Sink
	recorder *traceRecorder
}

func (s *traceSink) Emit(ctx context.Context, event *events.Event) error {
	if s != nil && s.recorder != nil {
		s.recorder.observe(event)
	}
	if s == nil || s.next == nil {
		return nil
	}
	return s.next.Emit(ctx, event)
}

func wrapSink(sink events.Sink, recorder *traceRecorder) events.Sink {
	return &traceSink{
		next:     sink,
		recorder: recorder,
	}
}

func toolNameFromEventContent(content string) string {
	if strings.TrimSpace(content) == "" {
		return ""
	}
	var payload struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal([]byte(content), &payload); err != nil {
		return ""
	}
	return strings.TrimSpace(payload.Name)
}
