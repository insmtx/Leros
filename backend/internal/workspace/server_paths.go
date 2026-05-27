package workspace

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/insmtx/Leros/backend/pkg/leros"
)

// WorkerMountedWorkspacePath 返回 server 视角下某个 worker 的 workspace 挂载目录。
func WorkerMountedWorkspacePath(orgID uint, workerID uint) (string, error) {
	if orgID == 0 {
		return "", fmt.Errorf("org_id is required")
	}
	if workerID == 0 {
		return "", fmt.Errorf("worker_id is required")
	}
	root, err := leros.WorkspaceRoot()
	if err != nil {
		return "", err
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	return filepath.Join(rootAbs, fmt.Sprintf("%d", orgID), fmt.Sprintf("%d", workerID), "workspace"), nil
}

// ArtifactStoragePath 从 server 侧解析 worker workspace 相对的 storage key。
func ArtifactStoragePath(orgID uint, workerID uint, storageKey string) (string, error) {
	key := strings.TrimSpace(storageKey)
	if key == "" || filepath.IsAbs(key) {
		return "", fmt.Errorf("invalid artifact storage key")
	}
	workspacePath, err := WorkerMountedWorkspacePath(orgID, workerID)
	if err != nil {
		return "", err
	}
	workspaceAbs, err := filepath.Abs(workspacePath)
	if err != nil {
		return "", err
	}
	pathAbs, err := filepath.Abs(filepath.Join(workspaceAbs, filepath.FromSlash(key)))
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(workspaceAbs, pathAbs)
	if err != nil {
		return "", err
	}
	if rel == "." || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return "", fmt.Errorf("artifact storage key escapes workspace")
	}
	return pathAbs, nil
}
