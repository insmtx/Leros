package taskconsumer

import (
	"context"

	"github.com/insmtx/Leros/backend/internal/runtime/events"
	agentworkspace "github.com/insmtx/Leros/backend/internal/workspace"
	"github.com/ygpkg/yg-go/logs"
)

type artifactStreamSink struct {
	next events.Sink
	plan *agentworkspace.TaskWorkspace
}

func newArtifactStreamSink(next events.Sink, plan *agentworkspace.TaskWorkspace) events.Sink {
	if plan == nil {
		return next
	}
	return &artifactStreamSink{next: next, plan: plan}
}

func (s *artifactStreamSink) Emit(ctx context.Context, event *events.Event) error {
	if s == nil || s.next == nil || event == nil {
		return nil
	}
	if event.Type != events.EventCompleted {
		return s.next.Emit(ctx, event)
	}
	artifacts, err := agentworkspace.CollectFinalArtifacts(ctx, s.plan)
	if err != nil {
		logs.WarnContextf(ctx, "collect final artifacts failed: %v", err)
		return s.next.Emit(ctx, event)
	}
	if len(artifacts) == 0 {
		return s.next.Emit(ctx, event)
	}
	payload, err := events.DecodePayload[events.RunCompletedPayload](event)
	if err != nil {
		logs.WarnContextf(ctx, "decode completed payload for artifacts failed: %v", err)
		return s.next.Emit(ctx, event)
	}
	payload.Artifacts = artifacts
	nextEvent := events.NewRunCompleted(payload, payload.Result.Message)
	nextEvent.ID = event.ID
	nextEvent.RunID = event.RunID
	nextEvent.TraceID = event.TraceID
	nextEvent.Seq = event.Seq
	nextEvent.CreatedAt = event.CreatedAt
	return s.next.Emit(ctx, nextEvent)
}

var _ events.Sink = (*artifactStreamSink)(nil)
