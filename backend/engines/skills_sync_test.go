package engines

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/insmtx/Leros/backend/pkg/leros"
)

func TestSyncBuiltinSkillsAlsoSyncsLerosUserSkills(t *testing.T) {
	builtinRoot := t.TempDir()
	targetRoot := t.TempDir()
	lerosOSHome := t.TempDir()
	t.Setenv(leros.EnvHome, lerosOSHome)

	writeSyncTestSkill(t, filepath.Join(builtinRoot, "review-flow"), "review-flow", "builtin body")
	writeSyncTestSkill(t, filepath.Join(lerosOSHome, "skills", "review-flow"), "review-flow", "user body")

	if err := SyncBuiltinSkills(builtinRoot, []string{targetRoot}); err != nil {
		t.Fatalf("sync skills: %v", err)
	}

	targetBody, err := os.ReadFile(filepath.Join(targetRoot, "review-flow", skillManifestFile))
	if err != nil {
		t.Fatalf("read synced skill: %v", err)
	}
	if !strings.Contains(string(targetBody), "user body") {
		t.Fatalf("expected user skill to overwrite builtin skill, got:\n%s", string(targetBody))
	}
}

func TestSyncLerosSkillsNoopsWhenNoSourcesExist(t *testing.T) {
	t.Setenv(leros.EnvHome, filepath.Join(t.TempDir(), "missing"))

	if err := SyncLerosSkills([]string{t.TempDir()}); err != nil {
		t.Fatalf("sync missing sources should no-op: %v", err)
	}
}

func writeSyncTestSkill(t *testing.T, dir string, name string, body string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir skill: %v", err)
	}
	content := "---\nname: " + name + "\ndescription: " + name + "\n---\n# " + name + "\n\n" + body + "\n"
	if err := os.WriteFile(filepath.Join(dir, skillManifestFile), []byte(content), 0o644); err != nil {
		t.Fatalf("write skill: %v", err)
	}
}
