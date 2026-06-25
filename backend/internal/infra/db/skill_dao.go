package db

import (
	"context"

	"gorm.io/gorm"

	"github.com/insmtx/Leros/backend/types"
)

// GetActiveSkillCodes returns a set of skill codes whose status is "active".
func GetActiveSkillCodes(ctx context.Context, db *gorm.DB) (map[string]bool, error) {
	var skills []types.Skill
	if err := db.WithContext(ctx).
		Select("code").
		Where("status = ?", types.SkillStatusActive).
		Find(&skills).Error; err != nil {
		return nil, err
	}

	active := make(map[string]bool, len(skills))
	for _, s := range skills {
		active[s.Code] = true
	}
	return active, nil
}

// GetSkillStatuses returns a map of skill code -> status for the given codes.
func GetSkillStatuses(ctx context.Context, db *gorm.DB, codes []string) (map[string]string, error) {
	if len(codes) == 0 {
		return map[string]string{}, nil
	}

	var skills []types.Skill
	if err := db.WithContext(ctx).
		Select("code", "status").
		Where("code IN ?", codes).
		Find(&skills).Error; err != nil {
		return nil, err
	}

	result := make(map[string]string, len(skills))
	for _, s := range skills {
		result[s.Code] = s.Status
	}
	return result, nil
}