package notify_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/notify"
)

type mockRunner struct {
	runFn func(name string, args ...string) error
}

func (m *mockRunner) Run(name string, args ...string) error {
	return m.runFn(name, args...)
}

func TestDesktopSend(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		message     string
		runnerErr   error
		wantErr     bool
		wantContain []string
	}{
		{
			name:        "builds correct osascript command",
			title:       "Test Title",
			message:     "Test Message",
			runnerErr:   nil,
			wantErr:     false,
			wantContain: []string{"osascript", "Test Title", "Test Message"},
		},
		{
			name:        "escapes quotes in message",
			title:       `Say "hello"`,
			message:     `It's a "test"`,
			runnerErr:   nil,
			wantErr:     false,
			wantContain: []string{`Say \"hello\"`, `It's a \"test\"`},
		},
		{
			name:        "returns error when runner fails",
			title:       "Title",
			message:     "Msg",
			runnerErr:   errors.New("osascript not found"),
			wantErr:     true,
			wantContain: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var captured string
			runner := &mockRunner{runFn: func(name string, args ...string) error {
				captured = name + " " + strings.Join(args, " ")
				return tt.runnerErr
			}}

			d := notify.NewDesktop(runner)
			err := d.Send(tt.title, tt.message)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			for _, want := range tt.wantContain {
				assert.Contains(t, captured, want)
			}
		})
	}
}
