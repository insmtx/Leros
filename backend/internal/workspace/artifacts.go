package workspace

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/insmtx/Leros/backend/internal/runtime/events"
	"github.com/insmtx/Leros/backend/types"
)

// ManifestArtifact is one JSON Lines declaration written by an agent turn.
type ManifestArtifact struct {
	Path         string `json:"path"`
	Title        string `json:"title,omitempty"`
	Description  string `json:"description,omitempty"`
	MimeType     string `json:"mime_type,omitempty"`
	ArtifactType string `json:"artifact_type,omitempty"`
	IsFinal      bool   `json:"is_final,omitempty"`
}

// CollectFinalArtifacts reads and validates the final artifact manifest.
func CollectFinalArtifacts(ctx context.Context, plan *TaskWorkspace) ([]events.ArtifactPayload, error) {
	if plan == nil || strings.TrimSpace(plan.ArtifactManifestPath) == "" {
		return nil, nil
	}
	file, err := os.Open(plan.ArtifactManifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("open artifact manifest: %w", err)
	}
	defer file.Close()

	declared := make(map[string]ManifestArtifact)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var item ManifestArtifact
		if err := json.Unmarshal([]byte(line), &item); err != nil {
			return nil, fmt.Errorf("parse artifact manifest: %w", err)
		}
		if !item.IsFinal {
			continue
		}
		key := filepath.ToSlash(filepath.Clean(filepath.FromSlash(strings.TrimSpace(item.Path))))
		if key == "." || strings.HasPrefix(key, "../") || strings.HasPrefix(key, "/") {
			return nil, fmt.Errorf("invalid artifact path %q", item.Path)
		}
		item.Path = key
		declared[key] = item
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read artifact manifest: %w", err)
	}
	if len(declared) == 0 {
		return nil, nil
	}

	paths := make([]string, 0, len(declared))
	for path := range declared {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	artifacts := make([]events.ArtifactPayload, 0, len(declared))
	for _, path := range paths {
		item := declared[path]
		artifact, err := buildArtifactPayload(plan, item)
		if err != nil {
			return nil, err
		}
		artifacts = append(artifacts, artifact)
	}
	return artifacts, nil
}

func buildArtifactPayload(plan *TaskWorkspace, item ManifestArtifact) (events.ArtifactPayload, error) {
	absolute, err := SafeJoin(plan.RepoDir, item.Path)
	if err != nil {
		return events.ArtifactPayload{}, fmt.Errorf("validate artifact path %q: %w", item.Path, err)
	}
	info, err := os.Stat(absolute)
	if err != nil {
		return events.ArtifactPayload{}, fmt.Errorf("stat artifact %q: %w", item.Path, err)
	}
	if info.IsDir() {
		return events.ArtifactPayload{}, fmt.Errorf("artifact %q is a directory", item.Path)
	}
	sha, err := sha256File(absolute)
	if err != nil {
		return events.ArtifactPayload{}, err
	}
	storageKey, err := plan.StorageKey(item.Path)
	if err != nil {
		return events.ArtifactPayload{}, err
	}
	title := strings.TrimSpace(item.Title)
	if title == "" {
		title = filepath.Base(item.Path)
	}
	artifactType := strings.TrimSpace(item.ArtifactType)
	if artifactType == "" {
		artifactType = string(types.ArtifactTypeFile)
	}
	return events.ArtifactPayload{
		Title:        title,
		Description:  strings.TrimSpace(item.Description),
		ArtifactType: artifactType,
		RelativePath: item.Path,
		StorageKey:   storageKey,
		MimeType:     detectMimeType(absolute, item.MimeType),
		FileSize:     info.Size(),
		Sha256:       sha,
		Source:       string(types.ArtifactSourceAgentDeclared),
		Status:       string(types.ArtifactStatusCompleted),
	}, nil
}

func sha256File(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open artifact for hash: %w", err)
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("hash artifact: %w", err)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func detectMimeType(path string, declared string) string {
	if strings.TrimSpace(declared) != "" {
		return strings.TrimSpace(declared)
	}
	if ext := filepath.Ext(path); ext != "" {
		if value := mime.TypeByExtension(ext); value != "" {
			return value
		}
	}
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()
	var buf [512]byte
	n, err := file.Read(buf[:])
	if err != nil && err != io.EOF {
		return ""
	}
	return http.DetectContentType(buf[:n])
}
