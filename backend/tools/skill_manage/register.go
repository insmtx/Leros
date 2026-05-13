package skillmanage

import (
	"fmt"

	skillmanageinternal "github.com/insmtx/Leros/backend/internal/skill/manage"
	"github.com/insmtx/Leros/backend/tools"
)

// NewTools returns all skill management tools.
func NewTools() []tools.Tool {
	return []tools.Tool{
		NewTool(),
	}
}

// NewToolsWithManager returns all skill management tools using an explicit manager.
func NewToolsWithManager(manager *skillmanageinternal.Manager) []tools.Tool {
	return []tools.Tool{
		NewToolWithManager(manager),
	}
}

// Register adds skill management tools to the runtime registry.
func Register(registry *tools.Registry) error {
	return RegisterWithManager(registry, nil)
}

// RegisterWithManager adds skill management tools to the runtime registry.
func RegisterWithManager(registry *tools.Registry, manager *skillmanageinternal.Manager) error {
	if registry == nil {
		return fmt.Errorf("tool registry is required")
	}
	var skillTools []tools.Tool
	if manager == nil {
		skillTools = NewTools()
	} else {
		skillTools = NewToolsWithManager(manager)
	}
	for _, tool := range skillTools {
		if err := registry.Register(tool); err != nil {
			return err
		}
	}
	return nil
}
