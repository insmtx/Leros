package utils

import (
	"context"
	"testing"
	"time"
)

func TestTrailingDebouncerRunsLatestAfterQuietWindow(t *testing.T) {
	ctx := context.Background()
	handled := make(chan int, 1)
	debouncer, err := NewTrailingDebouncer(30*time.Millisecond, func(ctx context.Context, value int) error {
		handled <- value
		return nil
	}, nil)
	if err != nil {
		t.Fatalf("NewTrailingDebouncer() error = %v", err)
	}

	debouncer.Call(ctx, "session-1", 1)
	time.Sleep(10 * time.Millisecond)
	debouncer.Call(ctx, "session-1", 2)
	time.Sleep(10 * time.Millisecond)
	debouncer.Call(ctx, "session-1", 3)

	select {
	case value := <-handled:
		t.Fatalf("handler ran before debounce window with value %d", value)
	case <-time.After(20 * time.Millisecond):
	}

	select {
	case value := <-handled:
		if value != 3 {
			t.Fatalf("handler value = %d, want 3", value)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for debounced handler")
	}

	select {
	case value := <-handled:
		t.Fatalf("handler ran more than once with value %d", value)
	case <-time.After(40 * time.Millisecond):
	}
}

func TestTrailingDebouncerKeepsKeysIndependent(t *testing.T) {
	type result struct {
		key   string
		value int
	}

	ctx := context.Background()
	handled := make(chan result, 2)
	debouncer, err := NewTrailingDebouncer(20*time.Millisecond, func(ctx context.Context, value result) error {
		handled <- value
		return nil
	}, nil)
	if err != nil {
		t.Fatalf("NewTrailingDebouncer() error = %v", err)
	}

	debouncer.Call(ctx, "session-1", result{key: "session-1", value: 1})
	debouncer.Call(ctx, "session-2", result{key: "session-2", value: 2})

	got := map[string]int{}
	for len(got) < 2 {
		select {
		case value := <-handled:
			got[value.key] = value.value
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("timed out waiting for handlers, got %v", got)
		}
	}
	if got["session-1"] != 1 || got["session-2"] != 2 {
		t.Fatalf("handled values = %v, want session-1=1 session-2=2", got)
	}
}
