package db

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/insmtx/Leros/backend/types"
)

// CreateArtifacts persists artifact records in order.
func CreateArtifacts(ctx context.Context, db *gorm.DB, artifacts []*types.Artifact) error {
	if len(artifacts) == 0 {
		return nil
	}
	return db.WithContext(ctx).Create(&artifacts).Error
}

// GetArtifactByPublicID returns one artifact in an organization.
func GetArtifactByPublicID(ctx context.Context, db *gorm.DB, orgID uint, publicID string) (*types.Artifact, error) {
	var entity types.Artifact
	err := db.WithContext(ctx).
		Where("org_id = ? AND public_id = ?", orgID, publicID).
		First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

// ListTaskArtifacts returns completed artifacts for a task.
func ListTaskArtifacts(ctx context.Context, db *gorm.DB, orgID uint, taskID uint) ([]*types.Artifact, error) {
	var entities []*types.Artifact
	err := db.WithContext(ctx).
		Where("org_id = ? AND task_id = ? AND status = ?", orgID, taskID, string(types.ArtifactStatusCompleted)).
		Order("created_at ASC").
		Find(&entities).Error
	return entities, err
}
