package notify_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/notify"
)

type mockPlayer struct {
	playFn func(filepath string) error
	called bool
	path   string
}

func (m *mockPlayer) Play(fp string) error {
	m.called = true
	m.path = fp
	return m.playFn(fp)
}

func TestAudioPlayRandom(t *testing.T) {
	noon := func() time.Time { return time.Date(2026, 1, 1, 12, 0, 0, 0, time.Local) }
	lateNight := func() time.Time { return time.Date(2026, 1, 1, 23, 0, 0, 0, time.Local) }

	tests := []struct {
		name       string
		setupDir   func(t *testing.T) string
		quietHours notify.QuietHours
		nowFunc    func() time.Time
		playerErr  error
		wantCalled bool
		wantErr    bool
	}{
		{
			name: "plays a file when MP3s exist",
			setupDir: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				require.NoError(t, os.WriteFile(filepath.Join(dir, "chime.mp3"), []byte("fake"), 0o600))
				return dir
			},
			quietHours: notify.QuietHours{Enabled: false, Start: "21:00", End: "07:30"},
			nowFunc:    noon,
			playerErr:  nil,
			wantCalled: true,
			wantErr:    false,
		},
		{
			name: "returns nil when no MP3s found",
			setupDir: func(t *testing.T) string {
				t.Helper()
				return t.TempDir()
			},
			quietHours: notify.QuietHours{Enabled: false, Start: "21:00", End: "07:30"},
			nowFunc:    noon,
			playerErr:  nil,
			wantCalled: false,
			wantErr:    false,
		},
		{
			name: "skips playback during quiet hours",
			setupDir: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				require.NoError(t, os.WriteFile(filepath.Join(dir, "alert.mp3"), []byte("fake"), 0o600))
				return dir
			},
			quietHours: notify.QuietHours{Enabled: true, Start: "21:00", End: "07:30"},
			nowFunc:    lateNight,
			playerErr:  nil,
			wantCalled: false,
			wantErr:    false,
		},
		{
			name: "returns player error",
			setupDir: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				require.NoError(t, os.WriteFile(filepath.Join(dir, "sound.mp3"), []byte("fake"), 0o600))
				return dir
			},
			quietHours: notify.QuietHours{Enabled: false, Start: "21:00", End: "07:30"},
			nowFunc:    noon,
			playerErr:  errors.New("speaker busy"),
			wantCalled: true,
			wantErr:    true,
		},
		{
			name: "returns error for nonexistent directory",
			setupDir: func(_ *testing.T) string {
				return "/tmp/nonexistent-audio-dir-xyz"
			},
			quietHours: notify.QuietHours{Enabled: false, Start: "21:00", End: "07:30"},
			nowFunc:    noon,
			playerErr:  nil,
			wantCalled: false,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setupDir(t)
			player := &mockPlayer{
				playFn: func(_ string) error { return tt.playerErr },
				called: false,
				path:   "",
			}

			a := notify.NewAudio(player, dir, tt.quietHours, tt.nowFunc)
			err := a.PlayRandom()

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.wantCalled, player.called)

			if tt.wantCalled && !tt.wantErr {
				assert.Contains(t, player.path, ".mp3")
			}
		})
	}
}
