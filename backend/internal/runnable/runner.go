// Package runnable provides background task runners with recover and keep-alive.
package runnable

import (
	"context"
	"time"

	"github.com/ygpkg/yg-go/logs"
)

// Run starts a named background task with panic recovery and auto-restart on error.
func Run(ctx context.Context, name string, fn func(ctx context.Context)) {
	const restartDelay = 5 * time.Second
	for {
		select {
		case <-ctx.Done():
			logs.InfoContextf(ctx, "runnable %s stopped", name)
			return
		default:
		}

		finished := make(chan struct{}, 1)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					logs.ErrorContextf(ctx, "runnable %s panicked: %v", name, r)
				}
				finished <- struct{}{}
			}()
			fn(ctx)
		}()

		<-finished

		if ctx.Err() != nil {
			return
		}

		logs.WarnContextf(ctx, "runnable %s exited, restarting in %v", name, restartDelay)
		select {
		case <-ctx.Done():
			return
		case <-time.After(restartDelay):
		}
	}
}
