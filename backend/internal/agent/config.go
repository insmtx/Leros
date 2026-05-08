package agent

import (
	skillcatalog "github.com/insmtx/SingerOS/backend/internal/skill/catalog"
	"github.com/insmtx/SingerOS/backend/tools"
)

// Config stores runtime dependencies that are orthogonal to the agent implementation.
type Config struct {
	SkillsCatalog         *skillcatalog.Catalog
	SkillsCatalogProvider skillcatalog.CatalogProvider
	ToolRegistry          *tools.Registry
}
