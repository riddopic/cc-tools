package superpowers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/superpowers"
)

// hookOutput mirrors the JSON structure emitted by Injector.Run.
type hookOutput struct {
	HookSpecificOutput hookSpecificOutput `json:"hookSpecificOutput"`
}

// hookSpecificOutput carries the event name and injected context.
type hookSpecificOutput struct {
	HookEventName     string `json:"hookEventName"`
	AdditionalContext string `json:"additionalContext"`
}

func TestInjectorRun(t *testing.T) {
	tests := []struct {
		name            string
		setupFunc       func(t *testing.T, projectDir string)
		wantOutput      bool
		wantEventName   string
		wantContentWrap bool
		wantErr         bool
	}{
		{
			name: "outputs correct hookSpecificOutput JSON with skill content",
			setupFunc: func(t *testing.T, projectDir string) {
				t.Helper()
				skillDir := filepath.Join(projectDir, ".claude", "skills", "using-superpowers")
				require.NoError(t, os.MkdirAll(skillDir, 0o755))
				require.NoError(t, os.WriteFile(
					filepath.Join(skillDir, "SKILL.md"),
					[]byte("Use /superpowers to discover skills."),
					0o600,
				))
			},
			wantOutput:      true,
			wantEventName:   "SessionStart",
			wantContentWrap: true,
			wantErr:         false,
		},
		{
			name:            "returns nil when skill file does not exist",
			setupFunc:       nil,
			wantOutput:      false,
			wantEventName:   "",
			wantContentWrap: false,
			wantErr:         false,
		},
		{
			name: "handles empty skill file",
			setupFunc: func(t *testing.T, projectDir string) {
				t.Helper()
				skillDir := filepath.Join(projectDir, ".claude", "skills", "using-superpowers")
				require.NoError(t, os.MkdirAll(skillDir, 0o755))
				require.NoError(t, os.WriteFile(
					filepath.Join(skillDir, "SKILL.md"),
					[]byte(""),
					0o600,
				))
			},
			wantOutput:      true,
			wantEventName:   "SessionStart",
			wantContentWrap: true,
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectDir := t.TempDir()

			if tt.setupFunc != nil {
				tt.setupFunc(t, projectDir)
			}

			inj := superpowers.NewInjector(projectDir)
			var buf bytes.Buffer
			ctx := context.Background()

			err := inj.Run(ctx, &buf)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			if !tt.wantOutput {
				assert.Empty(t, buf.String(), "expected no output when skill file is absent")
				return
			}

			assert.NotEmpty(t, buf.String(), "expected JSON output")

			var got hookOutput
			require.NoError(t, json.Unmarshal(buf.Bytes(), &got),
				"output should be valid JSON")

			assert.Equal(t, tt.wantEventName, got.HookSpecificOutput.HookEventName)

			if tt.wantContentWrap {
				assert.Contains(t, got.HookSpecificOutput.AdditionalContext,
					"<EXTREMELY_IMPORTANT>")
				assert.Contains(t, got.HookSpecificOutput.AdditionalContext,
					"</EXTREMELY_IMPORTANT>")
			}
		})
	}
}

func TestInjectorRunOutputIsValidJSON(t *testing.T) {
	projectDir := t.TempDir()
	skillDir := filepath.Join(projectDir, ".claude", "skills", "using-superpowers")
	require.NoError(t, os.MkdirAll(skillDir, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(skillDir, "SKILL.md"),
		[]byte("# Superpowers\nDiscover available skills."),
		0o600,
	))

	inj := superpowers.NewInjector(projectDir)
	var buf bytes.Buffer

	require.NoError(t, inj.Run(context.Background(), &buf))

	var parsed hookOutput
	require.NoError(t, json.Unmarshal(buf.Bytes(), &parsed),
		"Run output must be parseable JSON")

	assert.Equal(t, "SessionStart", parsed.HookSpecificOutput.HookEventName)
	assert.Contains(t, parsed.HookSpecificOutput.AdditionalContext,
		"Discover available skills.")
}

func TestInjectorRunWrapsContentInTags(t *testing.T) {
	projectDir := t.TempDir()
	skillDir := filepath.Join(projectDir, ".claude", "skills", "using-superpowers")
	require.NoError(t, os.MkdirAll(skillDir, 0o755))

	skillContent := "Look for .claude/skills/ directories."
	require.NoError(t, os.WriteFile(
		filepath.Join(skillDir, "SKILL.md"),
		[]byte(skillContent),
		0o600,
	))

	inj := superpowers.NewInjector(projectDir)
	var buf bytes.Buffer

	require.NoError(t, inj.Run(context.Background(), &buf))

	var got hookOutput
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))

	expected := "<EXTREMELY_IMPORTANT>\n" + skillContent + "\n</EXTREMELY_IMPORTANT>"
	assert.Equal(t, expected, got.HookSpecificOutput.AdditionalContext,
		"additionalContext must wrap content in EXTREMELY_IMPORTANT tags")
}
