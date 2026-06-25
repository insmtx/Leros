package service

import (
	"testing"
	"time"

	"github.com/insmtx/Leros/backend/types"
)

func TestIsPathAllowed(t *testing.T) {
	tests := []struct {
		path    string
		allowed bool
	}{
		{"uploads/readme.md", true},
		{"uploads/sub/dir/file.txt", true},
		{"artifacts/report.pdf", true},
		{"artifacts/", true},
		{"src/main.go", false},
		{"", false},
		{"uploads", false},
		{"artifacts", false},
		{"config.yaml", false},
		{"uploads_evil/file.txt", false},
	}

	for _, tt := range tests {
		result := isPathAllowed(tt.path)
		if result != tt.allowed {
			t.Errorf("isPathAllowed(%q) = %v, want %v", tt.path, result, tt.allowed)
		}
	}
}

func TestBuildFileTreeFromProjectFiles(t *testing.T) {
	now := time.Now()
	files := []types.ProjectFile{
		{
			OriginalName: "readme.md",
			MimeType:     "text/markdown",
			FileSize:     100,
			Source:       "user_upload",
			PublicID:     "file_001",
		},
		{
			OriginalName: "logo.png",
			MimeType:     "image/png",
			FileSize:     2048,
			Source:       "user_upload",
			PublicID:     "file_002",
		},
		{
			OriginalName: "report.pdf",
			MimeType:     "application/pdf",
			FileSize:     4096,
			Source:       "worker_artifact",
			PublicID:     "file_003",
		},
	}
	for i := range files {
		files[i].CreatedAt = now
	}

	roots := buildFileTreeFromProjectFiles(files, "")

	if len(roots) != 3 {
		t.Fatalf("expected 3 root files, got %d", len(roots))
	}

	rd := roots[0]
	if rd.Name != "readme.md" || rd.Type != "file" || rd.Path != "uploads/readme.md" {
		t.Errorf("root[0] expected uploads/readme.md, got %+v", rd)
	}

	logo := roots[1]
	if logo.Name != "logo.png" || logo.Type != "file" || logo.Path != "uploads/logo.png" {
		t.Errorf("root[1] expected uploads/logo.png, got %+v", logo)
	}

	rpt := roots[2]
	if rpt.Name != "report.pdf" || rpt.Type != "file" || rpt.Path != "artifacts/report.pdf" {
		t.Errorf("root[2] expected artifacts/report.pdf, got %+v", rpt)
	}
}

func TestBuildFileTreeFromProjectFiles_Empty(t *testing.T) {
	roots := buildFileTreeFromProjectFiles(nil, "")
	if len(roots) != 0 {
		t.Errorf("expected empty roots, got %v", roots)
	}
}

func TestMimeTypeByExt(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"image.png", "image/png"},
		{"data.json", "application/json"},
	}

	for _, tt := range tests {
		got := mimeTypeByExt(tt.filename)
		if got != tt.want {
			t.Errorf("mimeTypeByExt(%q) = %q, want %q", tt.filename, got, tt.want)
		}
	}

	if got := mimeTypeByExt("script.js"); got == "" {
		t.Errorf("mimeTypeByExt(\"script.js\") should return non-empty mime type")
	}

	if got := mimeTypeByExt("noext"); got != "" {
		t.Errorf("mimeTypeByExt(\"noext\") = %q, want \"\"", got)
	}
}
