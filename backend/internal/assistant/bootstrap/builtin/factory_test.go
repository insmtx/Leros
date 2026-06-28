package builtin

import (
	"testing"

	engines "github.com/insmtx/Leros/backend/agent/runtime/provider"
	"github.com/insmtx/Leros/backend/config"
)

func TestNewRegistryFromConfigDetectsInstalledEngines(t *testing.T) {
	// Set workspace root to a temp dir so deps.New can create state dirs.
	tmpDir := t.TempDir()
	t.Setenv("LEROS_WORKSPACE_ROOT", tmpDir)

	registry, err := NewRegistryFromConfig(&config.CLIEnginesConfig{})
	if err != nil {
		t.Fatalf("build registry: %v", err)
	}
	if registry == nil {
		t.Fatal("expected registry")
	}
}

func TestNewEngineRejectsUnsupportedEngine(t *testing.T) {
	_, err := newEngine("unknown", "")
	if err == nil {
		t.Fatal("expected unsupported engine error")
	}
}

func TestNewEngineCreatesBuiltinEngines(t *testing.T) {
	for _, name := range []string{engines.EngineClaude, engines.EngineCodex} {
		engine, err := newEngine(name, name)
		if err != nil {
			t.Fatalf("build %s engine: %v", name, err)
		}
		if engine == nil {
			t.Fatalf("expected %s engine", name)
		}
	}
}
