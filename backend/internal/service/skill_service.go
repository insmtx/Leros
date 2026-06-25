package service

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/insmtx/Leros/backend/internal/api/contract"
	infradb "github.com/insmtx/Leros/backend/internal/infra/db"
	"github.com/insmtx/Leros/backend/types"
)

const defaultRecentSkillLimit = 10

type skillService struct {
	db *gorm.DB
}

// NewSkillService creates a new SkillService.
func NewSkillService(db *gorm.DB) contract.SkillService {
	return &skillService{db: db}
}

func (s *skillService) ToggleSkillStatus(ctx context.Context, code string, req *contract.ToggleSkillStatusRequest) (*contract.ToggleSkillStatusResponse, error) {
	if code == "" {
		return nil, fmt.Errorf("code is required")
	}
	if req.Status != string(types.SkillStatusActive) && req.Status != string(types.SkillStatusInactive) {
		return nil, fmt.Errorf("invalid status: %s (must be 'active' or 'inactive')", req.Status)
	}

	var skill types.Skill
	if err := s.db.WithContext(ctx).Where("code = ?", code).First(&skill).Error; err != nil {
		return nil, fmt.Errorf("skill not found: %s", code)
	}

	if skill.Status == req.Status {
		return &contract.ToggleSkillStatusResponse{Code: code, Status: req.Status}, nil
	}

	if err := s.db.WithContext(ctx).Model(&skill).Update("status", req.Status).Error; err != nil {
		return nil, fmt.Errorf("failed to update skill status: %w", err)
	}

	return &contract.ToggleSkillStatusResponse{Code: code, Status: req.Status}, nil
}

func (s *skillService) ListRecentUsedSkills(ctx context.Context, limit int) ([]types.Skill, error) {
	if limit <= 0 {
		limit = defaultRecentSkillLimit
	}

	keys, err := infradb.GetDistinctSkillCodes(ctx, s.db, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get distinct skill codes: %w", err)
	}
	if len(keys) == 0 {
		return nil, nil
	}

	codes := make([]string, 0, len(keys))
	for _, key := range keys {
		if idx := strings.Index(key, ":"); idx != -1 {
			codes = append(codes, key[idx+1:])
		} else {
			codes = append(codes, key)
		}
	}

	var skills []types.Skill
	if err := s.db.WithContext(ctx).Where("code IN ?", codes).Find(&skills).Error; err != nil {
		return nil, fmt.Errorf("failed to query skills: %w", err)
	}

	skillMap := make(map[string]types.Skill, len(skills))
	for _, s := range skills {
		skillMap[s.Code] = s
	}

	result := make([]types.Skill, 0, len(codes))
	for _, code := range codes {
		if s, ok := skillMap[code]; ok {
			result = append(result, s)
		}
	}

	return result, nil
}

func (s *skillService) GetSkillStatuses(ctx context.Context, codes []string) (map[string]string, error) {
	return infradb.GetSkillStatuses(ctx, s.db, codes)
}
