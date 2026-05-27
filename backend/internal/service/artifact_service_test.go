package service

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/insmtx/Leros/backend/types"
)

func TestListTaskArtifactsDoesNotExposeDownloadURL(t *testing.T) {
	result := convertToContractArtifact(&types.Artifact{
		PublicID:     "art_result",
		Title:        "Result",
		Filename:     "result.md",
		ArtifactType: "file",
		MimeType:     "text/markdown",
		FileSize:     12,
		Sha256:       "abc123",
		StorageKey:   "projects/1/1/repo/result.md",
	})
	payload, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}
	if strings.Contains(string(payload), "download_url") {
		t.Fatalf("list response should not expose download_url: %s", payload)
	}
}
