package seqtracker

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func newTestTracker(t *testing.T) SeqTracker {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test_tracker.db")
	tracker, err := NewSQLiteTracker(path)
	if err != nil {
		t.Fatalf("failed to create tracker: %v", err)
	}
	t.Cleanup(func() { tracker.Close() })
	return tracker
}

func TestTrackReceived(t *testing.T) {
	tracker := newTestTracker(t)
	ctx := context.Background()

	err := tracker.TrackReceived(ctx, "test.topic", 100, "session-1", "msg-1", "task-1", "run-1")
	if err != nil {
		t.Fatalf("TrackReceived failed: %v", err)
	}

	// Track same seq again should be idempotent.
	err = tracker.TrackReceived(ctx, "test.topic", 100, "session-2", "msg-2", "task-2", "run-2")
	if err != nil {
		t.Fatalf("second TrackReceived failed: %v", err)
	}
}

func TestMarkProcessingAndCompleted(t *testing.T) {
	tracker := newTestTracker(t)
	ctx := context.Background()

	tracker.TrackReceived(ctx, "test.topic", 200, "s1", "m1", "t1", "r1")

	err := tracker.MarkProcessing(ctx, "test.topic", 200)
	if err != nil {
		t.Fatalf("MarkProcessing failed: %v", err)
	}

	err = tracker.MarkCompleted(ctx, "test.topic", 200)
	if err != nil {
		t.Fatalf("MarkCompleted failed: %v", err)
	}
}

func TestMarkFailed(t *testing.T) {
	tracker := newTestTracker(t)
	ctx := context.Background()

	tracker.TrackReceived(ctx, "test.topic", 300, "s1", "m1", "t1", "r1")

	err := tracker.MarkFailed(ctx, "test.topic", 300, "something went wrong")
	if err != nil {
		t.Fatalf("MarkFailed failed: %v", err)
	}
}

func TestGetLastCompletedSeq(t *testing.T) {
	tracker := newTestTracker(t)
	ctx := context.Background()

	// No records yet.
	seq, err := tracker.GetLastCompletedSeq(ctx, "test.topic")
	if err != nil {
		t.Fatalf("GetLastCompletedSeq failed: %v", err)
	}
	if seq != 0 {
		t.Fatalf("expected 0, got %d", seq)
	}

	// Insert and complete seq 1 and 2.
	tracker.TrackReceived(ctx, "test.topic", 1, "s1", "m1", "t1", "r1")
	tracker.TrackReceived(ctx, "test.topic", 2, "s2", "m2", "t2", "r2")
	tracker.MarkCompleted(ctx, "test.topic", 1)
	tracker.MarkCompleted(ctx, "test.topic", 2)

	seq, err = tracker.GetLastCompletedSeq(ctx, "test.topic")
	if err != nil {
		t.Fatalf("GetLastCompletedSeq failed: %v", err)
	}
	if seq != 2 {
		t.Fatalf("expected 2, got %d", seq)
	}

	// Only completed count — seq 3 is still pending.
	tracker.TrackReceived(ctx, "test.topic", 3, "s3", "m3", "t3", "r3")
	seq, err = tracker.GetLastCompletedSeq(ctx, "test.topic")
	if err != nil {
		t.Fatalf("GetLastCompletedSeq failed: %v", err)
	}
	if seq != 2 {
		t.Fatalf("expected 2 (max completed), got %d", seq)
	}
}

func TestIsDuplicate(t *testing.T) {
	tracker := newTestTracker(t)
	ctx := context.Background()

	tracker.TrackReceived(ctx, "test.topic", 400, "s1", "m1", "t1", "r1")

	// Not completed yet — not a duplicate.
	isDup, err := tracker.IsDuplicate(ctx, "test.topic", 400)
	if err != nil {
		t.Fatalf("IsDuplicate failed: %v", err)
	}
	if isDup {
		t.Fatal("expected not duplicate for pending message")
	}

	// Mark completed — now it is a duplicate.
	tracker.MarkCompleted(ctx, "test.topic", 400)
	isDup, err = tracker.IsDuplicate(ctx, "test.topic", 400)
	if err != nil {
		t.Fatalf("IsDuplicate failed: %v", err)
	}
	if !isDup {
		t.Fatal("expected duplicate for completed message")
	}
}

func TestDifferentTopics(t *testing.T) {
	tracker := newTestTracker(t)
	ctx := context.Background()

	tracker.TrackReceived(ctx, "topic.a", 100, "s1", "m1", "t1", "r1")
	tracker.MarkCompleted(ctx, "topic.a", 100)

	// Same seq on different topic should NOT be a duplicate.
	isDup, err := tracker.IsDuplicate(ctx, "topic.b", 100)
	if err != nil {
		t.Fatalf("IsDuplicate failed: %v", err)
	}
	if isDup {
		t.Fatal("different topic should not be duplicate")
	}

	// getLastCompletedSeq should be per-topic.
	seq, _ := tracker.GetLastCompletedSeq(ctx, "topic.b")
	if seq != 0 {
		t.Fatalf("expected 0 for topic.b, got %d", seq)
	}
	seq, _ = tracker.GetLastCompletedSeq(ctx, "topic.a")
	if seq != 100 {
		t.Fatalf("expected 100 for topic.a, got %d", seq)
	}
}

func init() {
	// Suppress SQLite warnings about test db files.
	os.Setenv("CGO_ENABLED", "1")
}
