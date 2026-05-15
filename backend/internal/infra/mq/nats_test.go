package mq

import (
	"context"
	"testing"
	"time"
)

func TestContextWithDefaultDeadlineAddsDeadline(t *testing.T) {
	ctx, cancel := contextWithDefaultDeadline(context.Background(), time.Second)
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected context deadline")
	}
	if time.Until(deadline) <= 0 {
		t.Fatal("expected future context deadline")
	}
}

func TestContextWithDefaultDeadlinePreservesExistingDeadline(t *testing.T) {
	parent, parentCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer parentCancel()
	parentDeadline, _ := parent.Deadline()

	ctx, cancel := contextWithDefaultDeadline(parent, time.Second)
	defer cancel()
	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected context deadline")
	}
	if !deadline.Equal(parentDeadline) {
		t.Fatalf("deadline = %v, want %v", deadline, parentDeadline)
	}
}

func TestStreamConfigFromTopicSessionStreamUsesMaxAge(t *testing.T) {
	cfg := streamConfigFromTopic("test_stream", "org.1.session.sess_1.message.stream")
	if cfg.MaxAge != defaultSessionStreamMaxAge {
		t.Fatalf("max age = %s, want %s", cfg.MaxAge, defaultSessionStreamMaxAge)
	}
}

func TestStreamConfigFromTopicNonSessionStreamWithoutMaxAge(t *testing.T) {
	cfg := streamConfigFromTopic("test_stream", "org.1.worker.1.task")
	if cfg.MaxAge != 0 {
		t.Fatalf("max age = %s, want 0", cfg.MaxAge)
	}
}
