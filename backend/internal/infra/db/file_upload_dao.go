package db

import (
	"context"

	"gorm.io/gorm"

	"github.com/insmtx/Leros/backend/types"
)

func CreateFileUpload(ctx context.Context, db *gorm.DB, file *types.FileUpload) error {
	return db.WithContext(ctx).Create(file).Error
}

func GetFileUploadByPublicID(ctx context.Context, db *gorm.DB, orgID uint, publicID string) (*types.FileUpload, error) {
	var file types.FileUpload
	err := db.WithContext(ctx).Where("public_id = ? AND org_id = ?", publicID, orgID).First(&file).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &file, nil
}

func ListFileUploads(ctx context.Context, db *gorm.DB, orgID uint, purpose string, offset, limit int) ([]types.FileUpload, int64, error) {
	var files []types.FileUpload
	query := db.WithContext(ctx).Model(&types.FileUpload{}).Where("org_id = ?", orgID)
	if purpose != "" {
		query = query.Where("purpose = ?", purpose)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&files).Error; err != nil {
		return nil, 0, err
	}
	return files, total, nil
}
