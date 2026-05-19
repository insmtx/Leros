package utils

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// DebouncedHandler handles the final value that survives a debounce window.
type DebouncedHandler[T any] func(ctx context.Context, value T) error

// ErrorHandler observes handler errors without coupling callers to background execution.
type ErrorHandler func(ctx context.Context, err error)

// TrailingDebouncer keeps the latest call for each key and runs it after the window is quiet.
type TrailingDebouncer[T any] struct {
	window  time.Duration
	handler DebouncedHandler[T]
	onError ErrorHandler

	mu      sync.Mutex
	entries map[string]*debounceEntry[T]
}

type debounceEntry[T any] struct {
	ctx        context.Context
	pending    T
	hasPending bool
	timer      *time.Timer
	seq        uint64
	running    bool
}

// NewTrailingDebouncer creates a keyed trailing-edge debouncer.
func NewTrailingDebouncer[T any](window time.Duration, handler DebouncedHandler[T], onError ErrorHandler) (*TrailingDebouncer[T], error) {
	if window <= 0 {
		return nil, fmt.Errorf("debounce window must be positive")
	}
	if handler == nil {
		return nil, fmt.Errorf("debounced handler is required")
	}
	return &TrailingDebouncer[T]{
		window:  window,
		handler: handler,
		onError: onError,
		entries: map[string]*debounceEntry[T]{},
	}, nil
}

// Call resets the key's debounce window and keeps value as the latest pending work.
func (d *TrailingDebouncer[T]) Call(ctx context.Context, key string, value T) {
	d.mu.Lock()
	entry := d.entries[key]
	if entry == nil {
		entry = &debounceEntry[T]{}
		d.entries[key] = entry
	}
	entry.ctx = ctx
	entry.pending = value
	entry.hasPending = true
	entry.seq++
	seq := entry.seq
	if entry.timer != nil {
		entry.timer.Stop()
	}
	entry.timer = time.AfterFunc(d.window, func() {
		d.flush(key, seq)
	})
	d.mu.Unlock()
}

func (d *TrailingDebouncer[T]) flush(key string, seq uint64) {
	ctx, value, ok := d.takeReady(key, seq)
	if !ok {
		return
	}
	d.run(key, ctx, value)
}

func (d *TrailingDebouncer[T]) takeReady(key string, seq uint64) (context.Context, T, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	entry := d.entries[key]
	if entry == nil {
		var zero T
		return nil, zero, false
	}
	if entry.seq != seq {
		var zero T
		return nil, zero, false
	}
	entry.timer = nil
	if entry.running || !entry.hasPending {
		var zero T
		return nil, zero, false
	}
	return d.takePendingLocked(entry)
}

func (d *TrailingDebouncer[T]) run(key string, ctx context.Context, value T) {
	if err := d.handler(ctx, value); err != nil && d.onError != nil {
		d.onError(ctx, err)
	}

	for {
		nextCtx, nextValue, ok := d.finishRun(key)
		if !ok {
			return
		}
		if err := d.handler(nextCtx, nextValue); err != nil && d.onError != nil {
			d.onError(nextCtx, err)
		}
	}
}

func (d *TrailingDebouncer[T]) finishRun(key string) (context.Context, T, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	entry := d.entries[key]
	if entry == nil {
		var zero T
		return nil, zero, false
	}
	entry.running = false
	if entry.hasPending && entry.timer == nil {
		return d.takePendingLocked(entry)
	}
	if !entry.hasPending && entry.timer == nil {
		delete(d.entries, key)
	}
	var zero T
	return nil, zero, false
}

func (d *TrailingDebouncer[T]) takePendingLocked(entry *debounceEntry[T]) (context.Context, T, bool) {
	ctx := entry.ctx
	value := entry.pending
	var zero T
	entry.pending = zero
	entry.hasPending = false
	entry.running = true
	return ctx, value, true
}
