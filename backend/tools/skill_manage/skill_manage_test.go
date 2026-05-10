package skillmanage

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	skillstore "github.com/insmtx/Leros/backend/internal/skill/store"
	"github.com/insmtx/Leros/backend/pkg/leros"
)

func TestToolExecuteCreate(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(leros.EnvHome, filepath.Join(home, ".leros"))

	store, err := skillstore.NewSkillStore(t.TempDir())
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	tool := NewToolWithStore(store)

	output, err := tool.Execute(context.Background(), map[string]interface{}{
		"action":  "create",
		"name":    "review-flow",
		"content": "---\nname: review-flow\ndescription: Review flow\n---\n# Review flow\n\n1. Inspect changes.\n",
	})
	if err != nil {
		t.Fatalf("execute create: %v", err)
	}

	var result skillstore.Result
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if !result.Success || result.Action != "create" || result.Name != "review-flow" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestToolValidateRequiresNewTextForPatch(t *testing.T) {
	tool := NewToolWithStore(nil)
	err := tool.Validate(map[string]interface{}{
		"action":   "patch",
		"name":     "review-flow",
		"old_text": "old",
	})
	if err == nil {
		t.Fatalf("expected validation error")
	}
}
