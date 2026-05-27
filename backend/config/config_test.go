package config

import (
	"testing"

	"gopkg.in/yaml.v2"
)

func TestConfigParsesWorkspaceRoot(t *testing.T) {
	var cfg Config
	body := []byte("workspace_root: /tmp/leros\nserver:\n  port: \"8080\"\n")

	if err := yaml.Unmarshal(body, &cfg); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}

	if cfg.WorkspaceRoot != "/tmp/leros" {
		t.Fatalf("workspace root = %q", cfg.WorkspaceRoot)
	}
}
