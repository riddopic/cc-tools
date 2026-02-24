package notify

import (
	"context"
	"os/exec"
	"time"
)

// afplayTimeout is the maximum time to wait for afplay to finish.
const afplayTimeout = 30 * time.Second

// AFPlayer plays audio files using macOS afplay.
type AFPlayer struct{}

// Play plays the audio file at the given path using afplay.
func (p *AFPlayer) Play(filepath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), afplayTimeout)
	defer cancel()

	return exec.CommandContext(ctx, "afplay", filepath).Run()
}
