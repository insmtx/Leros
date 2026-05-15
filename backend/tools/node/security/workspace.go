package security

import (
	"fmt"
	"os"
	"path/filepath"
)

const defaultWorkingDir = "/workspace"

// WorkspaceRoot 获取工作区根目录，优先使用环境变量 LEROS_WORKSPACE_ROOT
func WorkspaceRoot() (string, error) {
	if root := os.Getenv("LEROS_WORKSPACE_ROOT"); root != "" {
		return filepath.Abs(root)
	}
	if info, err := os.Stat(defaultWorkingDir); err == nil && info.IsDir() {
		return defaultWorkingDir, nil
	}
	return os.Getwd()
}

// RealWorkspaceRoot 获取工作区根目录的真实路径（解析符号链接）
func RealWorkspaceRoot() (string, error) {
	root, err := WorkspaceRoot()
	if err != nil {
		return "", err
	}
	root, err = filepath.Abs(filepath.Clean(root))
	if err != nil {
		return "", fmt.Errorf("resolve workspace root: %w", err)
	}
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", fmt.Errorf("resolve workspace root symlinks: %w", err)
	}
	return filepath.Clean(realRoot), nil
}
