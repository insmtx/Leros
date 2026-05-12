package leros

import (
	"github.com/insmtx/Leros/backend/internal/agent"
	einoadapter "github.com/insmtx/Leros/backend/internal/agent/eino"
	agentevents "github.com/insmtx/Leros/backend/internal/agent/events"
)

type runState struct {
	req          *agent.RequestContext
	emitter      *agentevents.Emitter
	userInput    string
	systemPrompt string
	toolBinding  einoadapter.ToolBinding
	maxStep      int
}
