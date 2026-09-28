package hooks_test

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"

	"github.com/riddopic/cc-tools/internal/hooks"
)

func numberedLines(n int) string {
	lines := make([]string, n)
	for i := range lines {
		lines[i] = fmt.Sprintf("line %d", i+1)
	}
	return strings.Join(lines, "\n")
}

func TestSummarizeOutput(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty", in: "", want: ""},
		{name: "only whitespace", in: "\n\n  \n", want: ""},
		{name: "short output kept verbatim", in: "a.go:1: bad\n", want: "a.go:1: bad"},
		{name: "ANSI escapes stripped", in: "\x1b[0;31merror\x1b[0m: \x1b[1mbold\x1b[22m\n", want: "error: bold"},
		{name: "CRLF normalized", in: "one\r\ntwo\r\n\r\n", want: "one\ntwo"},
		{name: "exactly at line limit kept", in: numberedLines(60), want: numberedLines(60)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, hooks.SummarizeOutput(tt.in))
		})
	}
}

func TestSummarizeOutput_Truncation(t *testing.T) {
	tests := []struct {
		name         string
		in           string
		wantContains []string
		wantMissing  []string
		maxLen       int
	}{
		{
			name: "long output keeps head and tail",
			in:   numberedLines(200),
			wantContains: []string{
				"line 1\n", "line 20\n", "... 140 lines omitted ...", "line 161\n", "line 200",
			},
			wantMissing: []string{"line 21\n", "line 160\n"},
			maxLen:      3000,
		},
		{
			name:         "huge single line capped",
			in:           strings.Repeat("x", 50_000),
			wantContains: []string{"chars omitted"},
			wantMissing:  nil,
			maxLen:       3100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hooks.SummarizeOutput(tt.in)
			for _, s := range tt.wantContains {
				assert.Contains(t, got, s)
			}
			for _, s := range tt.wantMissing {
				assert.NotContains(t, got, s)
			}
			assert.LessOrEqual(t, len(got), tt.maxLen)
		})
	}
}

func TestSummarizeOutput_MultibyteCapStaysValidUTF8(t *testing.T) {
	got := hooks.SummarizeOutput(strings.Repeat("é", 10_000))
	assert.True(t, utf8.ValidString(got), "output must remain valid UTF-8")
	assert.Contains(t, got, "chars omitted")
}
