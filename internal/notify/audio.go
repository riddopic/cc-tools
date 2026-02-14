package notify

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// AudioPlayer abstracts audio file playback for testing.
type AudioPlayer interface {
	Play(filepath string) error
}

// Audio manages audio notification playback.
type Audio struct {
	player     AudioPlayer
	dir        string
	quietHours QuietHours
	nowFunc    func() time.Time
}

// NewAudio creates a new Audio notifier.
func NewAudio(player AudioPlayer, dir string, qh QuietHours, nowFunc func() time.Time) *Audio {
	if nowFunc == nil {
		nowFunc = time.Now
	}

	return &Audio{
		player:     player,
		dir:        dir,
		quietHours: qh,
		nowFunc:    nowFunc,
	}
}

// PlayRandom plays a random MP3 file from the audio directory.
// Returns nil if quiet hours are active or no MP3 files are found.
func (a *Audio) PlayRandom() error {
	if a.quietHours.IsActive(a.nowFunc()) {
		return nil
	}

	files, err := listMP3Files(a.dir)
	if err != nil {
		return fmt.Errorf("list mp3 files in %s: %w", a.dir, err)
	}

	if len(files) == 0 {
		return nil
	}

	idx, randErr := rand.Int(rand.Reader, big.NewInt(int64(len(files))))
	if randErr != nil {
		return fmt.Errorf("generate random index: %w", randErr)
	}

	chosen := files[idx.Int64()]

	playErr := a.player.Play(chosen)
	if playErr != nil {
		return fmt.Errorf("play audio %s: %w", chosen, playErr)
	}

	return nil
}

func listMP3Files(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read directory %s: %w", dir, err)
	}

	files := make([]string, 0, len(entries))

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if strings.EqualFold(filepath.Ext(entry.Name()), ".mp3") {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}

	return files, nil
}
