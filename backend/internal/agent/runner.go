package agent

import "context"

// Runner executes one normalized agent request.
type Runner interface {
	Run(ctx context.Context, req *RequestContext) (*RunResult, error)
}

// EventSink receives observable events emitted during a run.
type EventSink interface {
	Emit(ctx context.Context, event *Event) error
}
