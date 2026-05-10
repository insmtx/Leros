package leros

import (
	"path/filepath"
	"testing"
)

func TestHomeDirUsesLerosHome(t *testing.T) {
	root := t.TempDir()
	t.Setenv(EnvHome, root)

	home, err := HomeDir()
	if err != nil {
		t.Fatalf("home dir: %v", err)
	}

	expected, err := filepath.Abs(root)
	if err != nil {
		t.Fatalf("abs root: %v", err)
	}
	if home != expected {
		t.Fatalf("expected %s, got %s", expected, home)
	}
}

func TestHomeDirDefaultsToUserLerosDir(t *testing.T) {
	userHome := t.TempDir()
	t.Setenv(EnvHome, "")
	t.Setenv("HOME", userHome)

	home, err := HomeDir()
	if err != nil {
		t.Fatalf("home dir: %v", err)
	}

	expected := filepath.Join(userHome, DefaultHomeDirName)
	if home != expected {
		t.Fatalf("expected %s, got %s", expected, home)
	}
}

func TestSkillsAndMemoryDirs(t *testing.T) {
	root := t.TempDir()
	t.Setenv(EnvHome, root)

	skillsDir, err := SkillsDir()
	if err != nil {
		t.Fatalf("skills dir: %v", err)
	}
	if skillsDir != filepath.Join(root, "skills") {
		t.Fatalf("unexpected skills dir: %s", skillsDir)
	}

	memoryDir, err := MemoryDir()
	if err != nil {
		t.Fatalf("memory dir: %v", err)
	}
	if memoryDir != filepath.Join(root, "memory") {
		t.Fatalf("unexpected memory dir: %s", memoryDir)
	}
}
