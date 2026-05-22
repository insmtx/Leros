package lifecycle

import (
	"context"

	"github.com/insmtx/Leros/backend/internal/runtime/lifecycle/steps"
)

type Step = steps.Step
type Pipeline = steps.Pipeline

func RunPipeline(ctx context.Context, pipeline Pipeline, state *RunState) error {
	return pipeline.Run(ctx, state)
}
