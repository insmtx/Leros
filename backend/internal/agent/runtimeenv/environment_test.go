package runtimeenv

import (
	"context"
	"reflect"
	"testing"

	"github.com/insmtx/Leros/backend/tools"
)

type fakeTool struct {
	tools.BaseTool
}

func newFakeTool(name string) fakeTool {
	return fakeTool{
		BaseTool: tools.NewBaseTool(name, "fake tool", tools.Schema{Type: "object"}),
	}
}

func (t fakeTool) Execute(context.Context, map[string]interface{}) (string, error) {
	return "ok", nil
}

func TestAvailableToolNamesFiltersRegisteredTools(t *testing.T) {
	registry := tools.NewRegistry()
	if err := registry.Register(newFakeTool("memory")); err != nil {
		t.Fatalf("register memory: %v", err)
	}
	if err := registry.Register(newFakeTool("skill_manage")); err != nil {
		t.Fatalf("register skill_manage: %v", err)
	}

	env := &Environment{toolRegistry: registry}
	got := env.AvailableToolNames([]string{"", "memory", "missing", "memory", "skill_manage"})
	want := []string{"memory", "skill_manage"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("AvailableToolNames() = %v, want %v", got, want)
	}
}
