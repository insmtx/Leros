package steps

import "context"

type Step interface {
	Name() string
	Run(ctx context.Context, state *State) error
}

type Pipeline []Step

func (p Pipeline) Run(ctx context.Context, state *State) error {
	for _, step := range p {
		if state != nil && state.Skipped {
			return nil
		}
		if err := step.Run(ctx, state); err != nil {
			return err
		}
	}
	return nil
}
