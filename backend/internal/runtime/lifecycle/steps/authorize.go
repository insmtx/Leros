package steps

import "context"

type AuthorizeStep struct{}

func (AuthorizeStep) Name() string {
	return "authorize"
}

func (AuthorizeStep) Run(context.Context, *State) error {
	return nil
}
