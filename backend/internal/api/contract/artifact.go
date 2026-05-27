package contract

import (
	"context"
	"io"
)

// ArtifactService defines task artifact query and download behavior.
type ArtifactService interface {
	ListTaskArtifacts(ctx context.Context, taskPublicID string) ([]Artifact, error)
	GetArtifactDownload(ctx context.Context, artifactPublicID string) (*ArtifactDownload, error)
}

// Artifact is the public response shape for a generated file.
type Artifact struct {
	ArtifactID   string `json:"artifact_id"`
	Title        string `json:"title"`
	Filename     string `json:"filename,omitempty"`
	Description  string `json:"description,omitempty"`
	ArtifactType string `json:"artifact_type"`
	MimeType     string `json:"mime_type,omitempty"`
	FileSize     int64  `json:"file_size,omitempty"`
	Sha256       string `json:"sha256,omitempty"`
}

// ArtifactDownload contains a file stream and HTTP response metadata.
type ArtifactDownload struct {
	FileName string
	MimeType string
	Size     int64
	Reader   io.ReadCloser
}
