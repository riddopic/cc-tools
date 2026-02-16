package notify

import (
	"context"
	"os/exec"
)

// AFPlayer plays audio files using macOS afplay.
type AFPlayer struct{}

// Play plays the audio file at the given path using afplay.
func (p *AFPlayer) Play(filepath string) error {
	return exec.CommandContext(
		context.Background(), "afplay", filepath,
	).Run()
}
