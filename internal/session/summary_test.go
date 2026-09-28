package session_test

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"

	"github.com/riddopic/cc-tools/internal/session"
)

func TestBuildSummary(t *testing.T) {
	tests := []struct {
		name  string
		ts    *session.TranscriptSummary
		cwd   string
		want  string
		check func(t *testing.T, got string)
	}{
		{
			name:  "nil transcript",
			ts:    nil,
			cwd:   "/repo",
			want:  "",
			check: nil,
		},
		{
			name: "no prompt means nothing worth recalling",
			ts: &session.TranscriptSummary{
				TotalMessages: 0,
				FirstPrompt:   "",
				ToolsUsed:     nil,
				FilesModified: []string{"/repo/a.go"},
			},
			cwd:   "/repo",
			want:  "",
			check: nil,
		},
		{
			name: "prompt only",
			ts: &session.TranscriptSummary{
				TotalMessages: 1,
				FirstPrompt:   "Fix the parser",
				ToolsUsed:     nil,
				FilesModified: nil,
			},
			cwd:   "/repo",
			want:  `Request: "Fix the parser".`,
			check: nil,
		},
		{
			name: "whitespace collapsed and files made project-relative",
			ts: &session.TranscriptSummary{
				TotalMessages: 1,
				FirstPrompt:   "Fix\n\n  the   parser",
				ToolsUsed:     nil,
				FilesModified: []string{"/repo/internal/a.go", "/elsewhere/b.go"},
			},
			cwd:   "/repo",
			want:  `Request: "Fix the parser". Files modified: internal/a.go, b.go.`,
			check: nil,
		},
		{
			name: "file list capped at five",
			ts: &session.TranscriptSummary{
				TotalMessages: 1,
				FirstPrompt:   "Refactor",
				ToolsUsed:     nil,
				FilesModified: []string{"/r/1", "/r/2", "/r/3", "/r/4", "/r/5", "/r/6", "/r/7"},
			},
			cwd:   "/r",
			want:  `Request: "Refactor". Files modified: 1, 2, 3, 4, 5, +2 more.`,
			check: nil,
		},
		{
			name: "long prompt truncated on a rune boundary",
			ts: &session.TranscriptSummary{
				TotalMessages: 1,
				FirstPrompt:   strings.Repeat("é", 400),
				ToolsUsed:     nil,
				FilesModified: nil,
			},
			cwd:  "/repo",
			want: "",
			check: func(t *testing.T, got string) {
				t.Helper()
				assert.True(t, utf8.ValidString(got))
				assert.Contains(t, got, "…")
				assert.LessOrEqual(t, utf8.RuneCountInString(got), 170)
			},
		},
		{
			name: "whole summary stays under the hard cap",
			ts: &session.TranscriptSummary{
				TotalMessages: 1,
				FirstPrompt:   strings.Repeat("word ", 100),
				ToolsUsed:     nil,
				FilesModified: []string{
					"/r/" + strings.Repeat("d/", 60) + "a.go",
					"/r/" + strings.Repeat("e/", 60) + "b.go",
				},
			},
			cwd:  "/r",
			want: "",
			check: func(t *testing.T, got string) {
				t.Helper()
				assert.LessOrEqual(t, utf8.RuneCountInString(got), 400)
				assert.True(t, strings.HasSuffix(got, "…"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := session.BuildSummary(tt.ts, tt.cwd)
			if tt.check != nil {
				tt.check(t, got)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
