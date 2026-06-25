package contract

import "context"

// ToggleSkillStatusRequest switches a skill between active/inactive.
type ToggleSkillStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// ToggleSkillStatusResponse returns the updated skill status.
type ToggleSkillStatusResponse struct {
	Code   string `json:"code"`
	Status string `json:"status"`
}

// SkillService defines the skill management contract.
type SkillService interface {
	ToggleSkillStatus(ctx context.Context, code string, req *ToggleSkillStatusRequest) (*ToggleSkillStatusResponse, error)
	ListRecentUsedSkills(ctx context.Context, orgID, uin uint, limit int) ([]SkillInstalledItem, error)
	GetSkillStatuses(ctx context.Context, codes []string) (map[string]string, error)
}
