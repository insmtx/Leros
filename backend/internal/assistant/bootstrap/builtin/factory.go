// Package builtin 连接内置的外部 CLI 引擎适配器。
package builtin

import (
	"fmt"

	claudeprovider "github.com/insmtx/Leros/backend/agent/runtime/claude"
	codexprovider "github.com/insmtx/Leros/backend/agent/runtime/codex"
	opencodeprovider "github.com/insmtx/Leros/backend/agent/runtime/opencode"
	engines "github.com/insmtx/Leros/backend/agent/runtime/provider"
	"github.com/insmtx/Leros/backend/config"
)

// NewRegistryFromConfig creates a registry with every detected external CLI provider.
func NewRegistryFromConfig(cfg *config.CLIEnginesConfig) (*engines.Registry, error) {
	registry := engines.NewRegistry()

	// Discover and register external CLI engines.
	for _, status := range engines.DiscoverAvailableCLI() {
		if !status.Installed {
			continue
		}
		engine, err := newEngine(status.Name, status.Path)
		if err != nil {
			return nil, err
		}
		if err := registry.Register(status.Name, engine); err != nil {
			return nil, err
		}
	}
	return registry, nil
}

func newEngine(name string, path string) (engines.Engine, error) {
	switch name {
	case engines.EngineClaude:
		return claudeprovider.NewAdapter(path, nil), nil
	case engines.EngineCodex:
		return codexprovider.NewAdapter(path, nil), nil
	case engines.EngineOpenCode:
		return opencodeprovider.NewAdapter(path, nil), nil
	default:
		return nil, fmt.Errorf("unsupported CLI engine %q", name)
	}
}
