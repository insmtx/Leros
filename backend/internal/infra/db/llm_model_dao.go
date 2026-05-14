package db

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/insmtx/Leros/backend/types"
)

// CreateLLMModel 创建LLM模型配置
func CreateLLMModel(ctx context.Context, db *gorm.DB, model *types.LLMModel) error {
	return db.WithContext(ctx).Create(model).Error
}

// GetLLMModelByID 根据ID获取LLM模型配置
func GetLLMModelByID(ctx context.Context, db *gorm.DB, id uint) (*types.LLMModel, error) {
	var entity types.LLMModel
	err := db.WithContext(ctx).First(&entity, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

// GetLLMModelByCode 根据组织ID和编码获取LLM模型配置
func GetLLMModelByCode(ctx context.Context, db *gorm.DB, orgID uint, code string) (*types.LLMModel, error) {
	var entity types.LLMModel
	err := db.WithContext(ctx).Where("org_id = ? AND code = ?", orgID, code).First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

// GetDefaultLLMModel 获取组织默认LLM模型配置
func GetDefaultLLMModel(ctx context.Context, db *gorm.DB, orgID uint) (*types.LLMModel, error) {
	var entity types.LLMModel
	err := db.WithContext(ctx).
		Where("org_id = ? AND is_default = ? AND status = ?", orgID, true, string(types.LLMModelStatusActive)).
		Order("updated_at DESC").
		First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

// UpdateLLMModel 更新LLM模型配置
func UpdateLLMModel(ctx context.Context, db *gorm.DB, model *types.LLMModel) error {
	return db.WithContext(ctx).Save(model).Error
}

// DeleteLLMModel 删除LLM模型配置
func DeleteLLMModel(ctx context.Context, db *gorm.DB, id uint) error {
	return db.WithContext(ctx).Delete(&types.LLMModel{}, id).Error
}

// ListLLMModels 分页查询LLM模型配置列表
func ListLLMModels(ctx context.Context, db *gorm.DB, orgID *uint, provider *string, status *string, keyword *string, offset, limit int) ([]*types.LLMModel, int64, error) {
	var entities []*types.LLMModel
	var total int64

	query := db.WithContext(ctx).Model(&types.LLMModel{})

	if orgID != nil && *orgID > 0 {
		query = query.Where("org_id = ?", *orgID)
	}
	if provider != nil && *provider != "" {
		query = query.Where("provider = ?", *provider)
	}
	if status != nil && *status != "" {
		query = query.Where("status = ?", *status)
	}
	if keyword != nil && *keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ? OR model LIKE ? OR description LIKE ?",
			"%"+*keyword+"%", "%"+*keyword+"%", "%"+*keyword+"%", "%"+*keyword+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Offset(offset).Limit(limit).Order("is_default DESC, created_at DESC").Find(&entities).Error
	if err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}

// LLMModelCodeExists 检查组织内LLM模型编码是否存在（排除指定ID）
func LLMModelCodeExists(ctx context.Context, db *gorm.DB, orgID uint, code string, excludeID uint) (bool, error) {
	var count int64
	query := db.WithContext(ctx).Model(&types.LLMModel{}).Where("org_id = ? AND code = ?", orgID, code)
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	err := query.Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
