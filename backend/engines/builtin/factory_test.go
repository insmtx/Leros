package builtin

import (
	"testing"

	"github.com/insmtx/Leros/backend/config"
	"github.com/insmtx/Leros/backend/engines"
	"github.com/insmtx/Leros/backend/internal/runtime/deps"
)

func TestNewRegistryFromConfigDetectsInstalledEngines(t *testing.T) {
	// Set workspace root to a temp dir so deps.New can create state dirs.
	tmpDir := t.TempDir()
	t.Setenv("LEROS_WORKSPACE_ROOT", tmpDir)

	env, err := deps.New(t.Context(), deps.Options{})
	if err != nil {
		t.Fatalf("create deps container: %v", err)
	}
	registry, err := NewRegistryFromConfig(&config.CLIEnginesConfig{}, env)
	if err != nil {
		t.Fatalf("build registry: %v", err)
	}
	if registry == nil {
		t.Fatal("expected registry")
	}
	// Native engine should always be registered.
	if _, ok := registry.Get("leros"); !ok {
		t.Fatal("expected native engine (leros) in registry")
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
