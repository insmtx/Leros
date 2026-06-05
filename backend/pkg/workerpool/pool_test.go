package workerpool

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestPoolExecutesTasks(t *testing.T) {
	p := New(4)
	defer p.Close()

	var count atomic.Int32
	done := make(chan struct{}, 10)

	for i := 0; i < 10; i++ {
		p.Submit(func(ctx context.Context) error {
			count.Add(1)
			done <- struct{}{}
			return nil
		})
	}

	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("timeout waiting for tasks")
		}
	}

	if count.Load() != 10 {
		t.Fatalf("expected 10 tasks, got %d", count.Load())
	}
}

func TestPoolConcurrencyLimit(t *testing.T) {
	p := New(2)
	defer p.Close()

	var maxConcurrent atomic.Int32
	var current atomic.Int32
	started := make(chan struct{})
	block := make(chan struct{})

	// Submit 4 tasks — only 2 should run concurrently.
	for i := 0; i < 4; i++ {
		p.Submit(func(ctx context.Context) error {
			n := current.Add(1)
			for {
				old := maxConcurrent.Load()
				if n <= old || maxConcurrent.CompareAndSwap(old, n) {
					break
				}
			}
			started <- struct{}{}
			<-block // Block until released.
			current.Add(-1)
			return nil
		})
	}

	// Wait for 2 tasks to start.
	<-started
	<-started

	// Give a moment for any potential third task to start.
	time.Sleep(100 * time.Millisecond)

	if maxConcurrent.Load() != 2 {
		t.Fatalf("expected max 2 concurrent tasks, got %d", maxConcurrent.Load())
	}

	// Release workers.
	close(block)

	// Wait for remaining tasks.
	for i := 0; i < 2; i++ {
		select {
		case <-started:
		case <-time.After(3 * time.Second):
			t.Fatal("timeout waiting for remaining tasks")
		}
	}
}

func TestPoolCloseWaitsForInFlight(t *testing.T) {
	p := New(2)

	var completed atomic.Int32
	started := make(chan struct{})
	block := make(chan struct{})

	p.Submit(func(ctx context.Context) error {
		started <- struct{}{}
		<-block
		completed.Add(1)
		return nil
	})

	<-started

	closeDone := make(chan struct{})
	go func() {
		p.Close()
		close(closeDone)
	}()

	// Close should NOT complete while task is in-flight.
	select {
	case <-closeDone:
		t.Fatal("Close returned before task completed")
	case <-time.After(200 * time.Millisecond):
	}

	// Release task.
	close(block)

	// Close should now complete.
	select {
	case <-closeDone:
	case <-time.After(5 * time.Second):
		t.Fatal("Close did not return after task completed")
	}

	if completed.Load() != 1 {
		t.Fatalf("expected 1 completed task, got %d", completed.Load())
	}
}
