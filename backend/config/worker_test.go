package config

import (
	"testing"

	"gopkg.in/yaml.v2"
)

func TestWorkerConfigParsesWorkspaceRoot(t *testing.T) {
	var cfg WorkerConfig
	body := []byte("org_id: 1\nworker_id: 2\nworkspace_root: /tmp/leros-workspace\n")

	if err := yaml.Unmarshal(body, &cfg); err != nil {
		t.Fatalf("unmarshal worker config: %v", err)
	}

	if cfg.WorkspaceRoot != "/tmp/leros-workspace" {
		t.Fatalf("workspace root = %q", cfg.WorkspaceRoot)
	}
}
