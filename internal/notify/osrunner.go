package notify

import (
	"context"
	"os/exec"
)

// OSRunner executes commands using os/exec.
type OSRunner struct{}

// Run executes the named program with the given arguments.
func (r *OSRunner) Run(name string, args ...string) error {
	return exec.CommandContext(
		context.Background(), name, args...,
	).Run()
}
