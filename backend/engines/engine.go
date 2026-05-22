// Package engines defines the execution boundary for external agent CLI engines.
package engines

import (
	"context"
	"time"

	"github.com/insmtx/Leros/backend/internal/runtime/events"
)

const (
	// EngineClaude is the registry name for Claude Code.
	EngineClaude = "claude"
	// EngineCodex is the registry name for Codex CLI.
	EngineCodex = "codex"
)

const (
	// EventProviderSessionStarted indicates that the provider created or exposed a native session ID.
	EventProviderSessionStarted events.EventType = "provider_session.started"
)

// PrepareRequest contains engine-specific workspace preparation input.
type PrepareRequest struct {
	WorkDir string
}

// ModelConfig carries model settings injected into CLI processes.
type ModelConfig struct {
	Provider string
	Model    string
	APIKey   string
	BaseURL  string
}

// RunRequest contains all input needed to execute one external CLI run.
type RunRequest struct {
	ExecutionID  string
	SessionID    string
	Resume       bool
	WorkDir      string
	SystemPrompt string
	Prompt       string
	Model        ModelConfig
	ExtraEnv     []string
	Timeout      time.Duration
}

// Process is a running external CLI process handle.
type Process interface {
	PID() int
	Stop() error
}

// RunHandle is returned after an engine process starts.
type RunHandle struct {
	Process Process
	Events  <-chan events.Event
}

// Engine executes prompts through an external AI CLI.
type Engine interface {
	Prepare(ctx context.Context, req PrepareRequest) error
	RegisterMCP(ctx context.Context, cfg MCPServerConfig) error
	GetSkillDir() string
	Run(ctx context.Context, req RunRequest) (*RunHandle, error)
}
