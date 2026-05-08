package lifecycle

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"github.com/insmtx/SingerOS/backend/internal/agent"
	agentevents "github.com/insmtx/SingerOS/backend/internal/agent/events"
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
}

func (r *traceRecorder) observe(event *agentevents.RunEvent) {
	if r == nil || event == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	switch event.Type {
	case agentevents.RunEventToolCallStarted:
		r.toolCalls++
		if name := toolNameFromEventContent(event.Content); name != "" {
			r.toolNames = append(r.toolNames, name)
			// TODO 外部CLI 可能无法区分技能调用，后续优化
			if name == "skill_use" {
				r.usedSkillTool = true
			}
		}
	case agentevents.RunEventToolCallFailed:
		r.toolFailures++
	}
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
	next     agent.RunEventSink
	recorder *traceRecorder
}

func (s *traceSink) Emit(ctx context.Context, event *agentevents.RunEvent) error {
	if s != nil && s.recorder != nil {
		s.recorder.observe(event)
	}
	if s == nil || s.next == nil {
		return nil
	}
	return s.next.Emit(ctx, event)
}

func wrapSink(sink agent.RunEventSink, recorder *traceRecorder) agent.RunEventSink {
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
