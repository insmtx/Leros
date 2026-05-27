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

// CreateArtifact persists one artifact record.
func CreateArtifact(ctx context.Context, db *gorm.DB, artifact *types.Artifact) error {
	if artifact == nil {
		return nil
	}
	return db.WithContext(ctx).Create(artifact).Error
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

// BindArtifactMessage attaches an existing artifact to a session message.
func BindArtifactMessage(ctx context.Context, db *gorm.DB, artifactID uint, sessionID uint, messageID uint) error {
	updates := &types.Artifact{
		SessionID: &sessionID,
		MessageID: &messageID,
	}
	return db.WithContext(ctx).
		Model(&types.Artifact{}).
		Where("id = ?", artifactID).
		Select("session_id", "message_id").
		Updates(updates).Error
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
