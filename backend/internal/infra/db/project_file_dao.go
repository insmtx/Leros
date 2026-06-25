package db

import (
	"context"

	"gorm.io/gorm"

	"github.com/insmtx/Leros/backend/types"
)

// CreateProjectFile 创建项目文件关联记录
func CreateProjectFile(ctx context.Context, db *gorm.DB, file *types.ProjectFile) error {
	return db.WithContext(ctx).Create(file).Error
}

// GetProjectFileByPublicID 通过 public_id 查询项目文件记录
func GetProjectFileByPublicID(ctx context.Context, db *gorm.DB, orgID uint, publicID string) (*types.ProjectFile, error) {
	var file types.ProjectFile
	err := db.WithContext(ctx).Where("public_id = ? AND org_id = ?", publicID, orgID).First(&file).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &file, nil
}

// GetProjectFileByArtifactID 通过 artifact_id 查询关联的项目文件记录
func GetProjectFileByArtifactID(ctx context.Context, db *gorm.DB, artifactID uint) (*types.ProjectFile, error) {
	var file types.ProjectFile
	err := db.WithContext(ctx).Where("artifact_id = ?", artifactID).First(&file).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &file, nil
}

// ListProjectFiles 查询项目下的文件列表，可按 source 过滤
func ListProjectFiles(ctx context.Context, db *gorm.DB, orgID uint, projectPublicID string, source string) ([]types.ProjectFile, error) {
	var files []types.ProjectFile
	query := db.WithContext(ctx).Model(&types.ProjectFile{}).
		Where("org_id = ? AND project_public_id = ?", orgID, projectPublicID)
	if source != "" {
		query = query.Where("source = ?", source)
	}
	if err := query.Order("created_at DESC").Find(&files).Error; err != nil {
		return nil, err
	}
	return files, nil
}

// DeleteProjectFile 软删除项目文件记录
func DeleteProjectFile(ctx context.Context, db *gorm.DB, publicID string) error {
	return db.WithContext(ctx).Where("public_id = ?", publicID).Delete(&types.ProjectFile{}).Error
}

// DeleteProjectFilesByArtifactID 按 artifact_id 删除关联的项目文件记录
func DeleteProjectFilesByArtifactID(ctx context.Context, db *gorm.DB, artifactID uint) error {
	return db.WithContext(ctx).Where("artifact_id = ?", artifactID).Delete(&types.ProjectFile{}).Error
}
