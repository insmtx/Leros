package native

import (
	"github.com/insmtx/Leros/backend/internal/agent"
	einoadapter "github.com/insmtx/Leros/backend/internal/runtime/eino"
	"github.com/insmtx/Leros/backend/internal/runtime/events"
)

type runState struct {
	req          *agent.RequestContext
	eventSink    events.Sink
	userInput    string
	systemPrompt string
	toolBinding  einoadapter.ToolBinding
	maxStep      int
}
