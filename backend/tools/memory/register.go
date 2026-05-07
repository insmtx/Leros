package memory

import (
	"fmt"

	"github.com/insmtx/SingerOS/backend/tools"
)

// NewTools returns all built-in memory tools.
func NewTools() []tools.Tool {
	return []tools.Tool{
		NewTool(),
	}
}

// Register adds built-in memory tools to the runtime registry.
func Register(registry *tools.Registry) error {
	if registry == nil {
		return fmt.Errorf("tool registry is nil")
	}
	for _, tool := range NewTools() {
		if err := registry.Register(tool); err != nil {
			return err
		}
	}
	return nil
}
