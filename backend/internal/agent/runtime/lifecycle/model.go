package lifecycle

import (
	"context"
	"fmt"

	"github.com/insmtx/Leros/backend/internal/agent"
	infradb "github.com/insmtx/Leros/backend/internal/infra/db"
	"github.com/insmtx/Leros/backend/internal/worker/identity"
	"github.com/insmtx/Leros/backend/types"
)

// EnsureModelConfig fills req.Model from the persisted model table.
//
// If req.Model.ID is set, that model is used. Otherwise the current worker
// organization's default active model is used. A request that already carries a
// complete in-process model config is left unchanged.
func EnsureModelConfig(ctx context.Context, req *agent.RequestContext) error {
	if req == nil {
		return fmt.Errorf("request context is required")
	}
	if req.Model.Provider != "" && req.Model.Model != "" && req.Model.APIKey != "" {
		return nil
	}

	orgID := identity.OrgID()
	if orgID == 0 {
		return nil
	}

	database := infradb.GetDB()
	if database == nil {
		return fmt.Errorf("database is not initialized")
	}

	var (
		model *types.LLMModel
		err   error
	)
	if req.Model.ID > 0 {
		model, err = infradb.GetLLMModelByID(ctx, database, req.Model.ID)
	} else {
		model, err = infradb.GetDefaultLLMModel(ctx, database, orgID)
	}
	if err != nil {
		return err
	}
	if model == nil {
		return fmt.Errorf("llm model not found")
	}
	if model.OrgID != orgID {
		return fmt.Errorf("llm model does not belong to current org")
	}
	if model.Status != string(types.LLMModelStatusActive) {
		return fmt.Errorf("llm model is inactive")
	}

	req.Model.ID = model.ID
	req.Model.Provider = model.Provider
	req.Model.Model = model.ModelName
	req.Model.APIKey = model.APIKeyEncrypted
	req.Model.BaseURL = model.BaseURL
	if req.Model.Temperature == 0 {
		req.Model.Temperature = model.Temperature
	}
	return nil
}
