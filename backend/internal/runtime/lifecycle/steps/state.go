package steps

import (
	"time"

	"github.com/insmtx/Leros/backend/internal/agent"
	lifecyclejournal "github.com/insmtx/Leros/backend/internal/runtime/lifecycle/journal"
)

type State struct {
	OriginalRequest *agent.RequestContext
	Request         *agent.RequestContext
	Journal         *lifecyclejournal.RunJournal
	Result          *agent.RunResult
	Err             error
	StartedAt       time.Time
	Skipped         bool
}
