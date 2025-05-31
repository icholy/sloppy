package sloppy

import "context"

type AgentV2 interface {
	Run(ctx context.Context, input *RunInput) (*RunOutput, error)
}

type Driver struct {
}

func (d *Driver) Run(ctx context.Context, prompt string) error {
	return nil
}
