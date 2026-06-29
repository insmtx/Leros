package service

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/insmtx/Leros/backend/internal/api/contract"
	infradb "github.com/insmtx/Leros/backend/internal/infra/db"
	"github.com/insmtx/Leros/backend/internal/infra/filestore"
	"github.com/insmtx/Leros/backend/types"
	"gorm.io/gorm"
)

type artifactService struct {
	db *gorm.DB
}

func NewArtifactService(db *gorm.DB, _ interface{}) contract.ArtifactService {
	return &artifactService{db: db}
}

func (s *artifactService) ListTaskArtifacts(ctx context.Context, taskPublicID string) ([]contract.Artifact, error) {
	caller, err := requireCallerOrg(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(taskPublicID) == "" {
		return nil, errors.New("task_id is required")
	}
	task, err := infradb.GetTaskByPublicID(ctx, s.db, caller.OrgID, taskPublicID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, errors.New("task not found")
	}
	if err := verifyUserPermission(task.OwnerID, caller.Uin); err != nil {
		return nil, err
	}
	artifacts, err := infradb.ListTaskArtifacts(ctx, s.db, caller.OrgID, task.ID)
	if err != nil {
		return nil, err
	}
	result := make([]contract.Artifact, 0, len(artifacts))
	for _, a := range artifacts {
		if converted := convertToContractArtifact(a); converted != nil {
			result = append(result, *converted)
		}
	}
	return result, nil
}

func (s *artifactService) GetArtifact(ctx context.Context, artifactPublicID string) (*contract.ArtifactDetail, error) {
	caller, err := requireCallerOrg(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(artifactPublicID) == "" {
		return nil, errors.New("artifact_id is required")
	}
	artifact, err := infradb.GetArtifactByPublicID(ctx, s.db, caller.OrgID, artifactPublicID)
	if err != nil {
		return nil, err
	}
	if artifact == nil {
		return nil, errors.New("artifact not found")
	}
	if err := verifyUserPermission(artifact.OwnerID, caller.Uin); err != nil {
		return nil, err
	}
	return convertToArtifactDetail(artifact), nil
}

func (s *artifactService) GetArtifactDownload(ctx context.Context, artifactPublicID string) (*contract.ArtifactDownload, error) {
	caller, err := requireCallerOrg(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(artifactPublicID) == "" {
		return nil, errors.New("artifact_id is required")
	}
	artifact, err := infradb.GetArtifactByPublicID(ctx, s.db, caller.OrgID, artifactPublicID)
	if err != nil {
		return nil, err
	}
	if artifact == nil {
		return nil, errors.New("artifact not found")
	}
	if err := verifyUserPermission(artifact.OwnerID, caller.Uin); err != nil {
		return nil, err
	}

	storageURI := strings.TrimSpace(artifact.FileURL)
	if storageURI == "" {
		return nil, errors.New("artifact has no file url")
	}

	bucket, key, err := filestore.ParseStorageURI(storageURI)
	if err != nil {
		return nil, fmt.Errorf("parse artifact storage uri: %w", err)
	}

	st := filestore.GetStorage()
	obj, err := st.GetObject(ctx, bucket, key)
	if err != nil {
		return nil, fmt.Errorf("read artifact from storage: %w", err)
	}

	return &contract.ArtifactDownload{
		FileName: artifactDownloadName(artifact),
		MimeType: artifact.MimeType,
		Size:     artifact.FileSize,
		Reader:   obj.Body,
	}, nil
}

func convertToContractArtifact(artifact *types.Artifact) *contract.Artifact {
	if artifact == nil {
		return nil
	}
	return &contract.Artifact{
		ArtifactID:   artifact.PublicID,
		Title:        artifact.Title,
		Filename:     artifact.Filename,
		Description:  artifact.Description,
		ArtifactType: artifact.ArtifactType,
		MimeType:     artifact.MimeType,
		FileSize:     artifact.FileSize,
		Sha256:       artifact.Sha256,
		CreatedAt:    artifact.CreatedAt,
	}
}

func convertToArtifactDetail(artifact *types.Artifact) *contract.ArtifactDetail {
	if artifact == nil {
		return nil
	}
	return &contract.ArtifactDetail{
		Artifact:     *convertToContractArtifact(artifact),
		RelativePath: artifact.RelativePath,
		FilePublicID: artifact.FilePublicID,
		Source:       artifact.Source,
		ExportFormat: artifact.ExportFormat,
		Version:      artifact.Version,
		Status:       artifact.Status,
	}
}

func artifactDownloadName(artifact *types.Artifact) string {
	if artifact == nil {
		return ""
	}
	if strings.TrimSpace(artifact.Filename) != "" {
		return strings.TrimSpace(artifact.Filename)
	}
	if strings.TrimSpace(artifact.Title) != "" {
		return strings.TrimSpace(artifact.Title)
	}
	return filepath.Base(strings.TrimSpace(artifact.RelativePath))
}

var _ contract.ArtifactService = (*artifactService)(nil)
