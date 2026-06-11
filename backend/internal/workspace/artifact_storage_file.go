package workspace

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"strings"

	appstorage "github.com/insmtx/Leros/backend/internal/infra/storage"
)

type ArtifactStorageFile struct {
	Path     string
	Filename string
	MimeType string
	FileSize int64
	Sha256   string
}

func ResolveArtifactStorageFile(ctx context.Context, orgID uint, workerID uint, storageKey string, declaredMimeType string) (*ArtifactStorageFile, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	st := appstorage.GetStorage()
	bucket := appstorage.DefaultBucket()

	result, err := st.GetObject(ctx, bucket, storageKey)
	if err != nil {
		return nil, fmt.Errorf("get artifact object: %w", err)
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("read artifact object: %w", err)
	}

	hash := sha256.Sum256(data)
	sha256Hex := hex.EncodeToString(hash[:])

	return &ArtifactStorageFile{
		Path:     result.Path.Path(),
		Filename: filepath.Base(result.Path.Key()),
		MimeType: detectMimeTypeFromKey(result.Path.Key(), declaredMimeType),
		FileSize: result.Size,
		Sha256:   sha256Hex,
	}, nil
}

func OpenArtifactStorageFile(ctx context.Context, orgID uint, workerID uint, storageKey string) (io.ReadCloser, error) {
	st := appstorage.GetStorage()
	bucket := appstorage.DefaultBucket()

	result, err := st.GetObject(ctx, bucket, storageKey)
	if err != nil {
		return nil, fmt.Errorf("get artifact object: %w", err)
	}
	return result.Body, nil
}

func RepoRelativePathFromStorageKey(storageKey string) string {
	key := filepath.ToSlash(strings.TrimSpace(storageKey))
	const marker = "/repo/"
	idx := strings.Index(key, marker)
	if idx < 0 {
		return ""
	}
	return strings.TrimPrefix(key[idx+len(marker):], "/")
}

func detectMimeTypeFromKey(key, declared string) string {
	if strings.TrimSpace(declared) != "" {
		return normalizeMimeType(declared)
	}
	if ext := filepath.Ext(key); ext != "" {
		if value := mime.TypeByExtension(ext); value != "" {
			return normalizeMimeType(value)
		}
	}
	return ""
}

func normalizeMimeType(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if mediaType, _, err := mime.ParseMediaType(value); err == nil {
		return mediaType
	}
	if index := strings.Index(value, ";"); index >= 0 {
		return strings.TrimSpace(value[:index])
	}
	return value
}
