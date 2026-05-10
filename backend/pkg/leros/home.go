// Package leros provides shared Leros filesystem conventions.
package leros

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// EnvHome is the worker-local root used for Leros state.
	EnvHome = "LEROS_HOME"

	// DefaultHomeDirName is the fallback directory under the current user's home.
	DefaultHomeDirName = ".leros"
)

// HomeDir returns $LEROS_HOME, or ~/.leros when unset.
func HomeDir() (string, error) {
	home := strings.TrimSpace(os.Getenv(EnvHome))
	if home == "" {
		userHome, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve user home for %s: %w", EnvHome, err)
		}
		if strings.TrimSpace(userHome) == "" {
			return "", fmt.Errorf("user home is empty")
		}
		home = filepath.Join(userHome, DefaultHomeDirName)
	}

	absolute, err := filepath.Abs(home)
	if err != nil {
		return "", fmt.Errorf("resolve %s: %w", EnvHome, err)
	}
	return absolute, nil
}

// JoinHome joins path elements under the Leros home directory.
func JoinHome(elem ...string) (string, error) {
	home, err := HomeDir()
	if err != nil {
		return "", err
	}
	parts := append([]string{home}, elem...)
	return filepath.Join(parts...), nil
}

// SkillsDir returns the default Leros skills directory.
func SkillsDir() (string, error) {
	return JoinHome("skills")
}

// MemoryDir returns the default Leros memory directory.
func MemoryDir() (string, error) {
	return JoinHome("memory")
}
