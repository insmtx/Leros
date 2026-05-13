package leros

import (
	"github.com/insmtx/Leros/backend/internal/agent"
	einoadapter "github.com/insmtx/Leros/backend/internal/agent/eino"
	"github.com/insmtx/Leros/backend/internal/agent/runtime/events"
)

type runState struct {
	req          *agent.RequestContext
	emitter      *events.Emitter
	userInput    string
	systemPrompt string
	toolBinding  einoadapter.ToolBinding
	maxStep      int
}
