package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gorm.io/gorm"

	"github.com/insmtx/Leros/backend/internal/api/contract"
	"github.com/insmtx/Leros/backend/internal/infra/db"
	"github.com/insmtx/Leros/backend/pkg/leros"
	"github.com/insmtx/Leros/backend/types"
)

type artifactService struct {
	db *gorm.DB
}

// NewArtifactService creates a service for generated artifacts.
func NewArtifactService(db *gorm.DB) contract.ArtifactService {
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
	task, err := db.GetTaskByPublicID(ctx, s.db, caller.OrgID, taskPublicID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, errors.New("task not found")
	}
	artifacts, err := db.ListTaskArtifacts(ctx, s.db, caller.OrgID, task.ID)
	if err != nil {
		return nil, err
	}
	result := make([]contract.Artifact, 0, len(artifacts))
	for _, artifact := range artifacts {
		result = append(result, convertToContractArtifact(artifact))
	}
	return result, nil
}

func (s *artifactService) GetArtifactDownload(ctx context.Context, artifactPublicID string) (*contract.ArtifactDownload, error) {
	caller, err := requireCallerOrg(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(artifactPublicID) == "" {
		return nil, errors.New("artifact_id is required")
	}
	artifact, err := db.GetArtifactByPublicID(ctx, s.db, caller.OrgID, artifactPublicID)
	if err != nil {
		return nil, err
	}
	if artifact == nil {
		return nil, errors.New("artifact not found")
	}
	path, err := storagePath(artifact.StorageKey)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open artifact file: %w", err)
	}
	return &contract.ArtifactDownload{
		FileName: artifact.Title,
		MimeType: artifact.MimeType,
		Size:     artifact.FileSize,
		Reader:   file,
	}, nil
}

func convertToContractArtifact(artifact *types.Artifact) contract.Artifact {
	if artifact == nil {
		return contract.Artifact{}
	}
	return contract.Artifact{
		ArtifactID:   artifact.PublicID,
		Title:        artifact.Title,
		Description:  artifact.Description,
		ArtifactType: artifact.ArtifactType,
		MimeType:     artifact.MimeType,
		FileSize:     artifact.FileSize,
		Sha256:       artifact.Sha256,
		DownloadURL:  "/v1/artifacts/" + artifact.PublicID + "/download",
	}
}

func storagePath(storageKey string) (string, error) {
	key := strings.TrimSpace(storageKey)
	if key == "" || filepath.IsAbs(key) {
		return "", fmt.Errorf("invalid artifact storage key")
	}
	root, err := leros.WorkspaceRoot()
	if err != nil {
		return "", err
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	pathAbs, err := filepath.Abs(filepath.Join(rootAbs, filepath.FromSlash(key)))
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(rootAbs, pathAbs)
	if err != nil {
		return "", err
	}
	if rel == "." || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return "", fmt.Errorf("artifact storage key escapes workspace")
	}
	return pathAbs, nil
}

var _ contract.ArtifactService = (*artifactService)(nil)
