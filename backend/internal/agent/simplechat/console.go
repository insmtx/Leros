package simplechat

import (
	"context"
	"fmt"

	"github.com/insmtx/Leros/backend/internal/agent/runtime/events"
)

type ConsoleSink struct{}

func NewConsoleSink() *ConsoleSink {
	return &ConsoleSink{}
}

func (s *ConsoleSink) Emit(_ context.Context, event *events.Event) error {
	switch event.Type {
	case events.EventMessageDelta:
		fmt.Print(event.Content)
	case events.EventCompleted:
		fmt.Println()
	case events.EventToolCallStarted:
		fmt.Printf("\n[Tool Call Started] %s\n", event.Content)
	case events.EventToolCallOutput:
		fmt.Printf("\n[Tool Output] %s\n", event.Content)
	case events.EventFailed:
		fmt.Printf("\n[Error] %s\n", event.Content)
	}
	return nil
}
