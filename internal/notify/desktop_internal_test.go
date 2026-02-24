//go:build testmode

package notify

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEscapeAppleScript(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no special chars passes through",
			input: "Hello World",
			want:  "Hello World",
		},
		{
			name:  "escapes double quotes",
			input: `Say "hello"`,
			want:  `Say \"hello\"`,
		},
		{
			name:  "escapes backslash",
			input: `path\to\file`,
			want:  `path\\to\\file`,
		},
		{
			name:  "escapes newline",
			input: "line1\nline2",
			want:  `line1\nline2`,
		},
		{
			name:  "escapes carriage return",
			input: "line1\rline2",
			want:  `line1\rline2`,
		},
		{
			name:  "escapes tab",
			input: "col1\tcol2",
			want:  `col1\tcol2`,
		},
		{
			name:  "escapes all special chars together",
			input: "say \"hi\"\npath\\here\ttab\r",
			want:  `say \"hi\"\npath\\here\ttab\r`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := escapeAppleScript(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
