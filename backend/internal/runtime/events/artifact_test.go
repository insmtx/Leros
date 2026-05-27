package events

import (
	"strings"
	"testing"
)

func TestNewArtifactDeclaredPayloadIsMinimal(t *testing.T) {
	event := NewArtifactDeclared(ArtifactPayload{
		ArtifactID:   "art_123",
		Title:        "Report",
		Filename:     "report.md",
		MimeType:     "text/markdown",
		ArtifactType: "file",
	})
	if event.Type != EventArtifactDeclared {
		t.Fatalf("event type = %q, want %q", event.Type, EventArtifactDeclared)
	}
	if string(event.Payload) == "" {
		t.Fatal("expected payload")
	}
	payload := string(event.Payload)
	for _, forbidden := range []string{"tool_call_id", "path", "description", "download_url", "status", "is_final"} {
		if strings.Contains(payload, forbidden) {
			t.Fatalf("payload should not contain %q: %s", forbidden, payload)
		}
	}
}
