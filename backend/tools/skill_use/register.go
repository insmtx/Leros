package skilluse

import (
	"fmt"

	skillcatalog "github.com/insmtx/SingerOS/backend/internal/skill/catalog"
	"github.com/insmtx/SingerOS/backend/tools"
)

// NewTools returns all skill catalog tools for registration.
func NewTools(catalog skillcatalog.SkillCatalog) []tools.Tool {
	return NewToolsWithProvider(skillcatalog.NewStaticCatalogProvider(catalog))
}

// NewToolsWithProvider returns all skill catalog tools using a reloadable provider.
func NewToolsWithProvider(provider skillcatalog.CatalogProvider) []tools.Tool {
	return []tools.Tool{
		NewSkillUseToolWithProvider(provider),
	}
}

// Register registers all skill catalog tools into the provided registry.
func Register(registry *tools.Registry, catalog skillcatalog.SkillCatalog) error {
	return RegisterWithProvider(registry, skillcatalog.NewStaticCatalogProvider(catalog))
}

// RegisterWithProvider registers all skill catalog tools into the provided registry.
func RegisterWithProvider(registry *tools.Registry, provider skillcatalog.CatalogProvider) error {
	if registry == nil {
		return fmt.Errorf("tool registry is required")
	}

	for _, tool := range NewToolsWithProvider(provider) {
		if err := registry.Register(tool); err != nil {
			return err
		}
	}

	return nil
}
