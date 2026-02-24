package instinct_test

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/instinct"
)

func TestExportYAML(t *testing.T) {
	now := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
	instincts := []instinct.Instinct{
		{
			ID:         "test-first",
			Trigger:    "when adding features",
			Confidence: 0.8,
			Domain:     "testing",
			Source:     "observation",
			SourceRepo: "",
			Content:    "",
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}

	var buf bytes.Buffer
	require.NoError(t, instinct.ExportYAML(&buf, instincts))

	output := buf.String()
	assert.Contains(t, output, "id: test-first")
	assert.Contains(t, output, "trigger: when adding features")
	assert.Contains(t, output, "confidence: 0.8")
	assert.Contains(t, output, "domain: testing")
	assert.Contains(t, output, "source: observation")
	assert.Contains(t, output, "---")
}

func TestExportYAML_MultipleInstincts(t *testing.T) {
	now := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
	instincts := []instinct.Instinct{
		{
			ID:         "first",
			Trigger:    "trigger one",
			Confidence: 0.7,
			Domain:     "go",
			Source:     "observation",
			SourceRepo: "",
			Content:    "",
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:         "second",
			Trigger:    "trigger two",
			Confidence: 0.85,
			Domain:     "python",
			Source:     "inherited",
			SourceRepo: "github.com/example/repo",
			Content:    "Some content here",
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}

	var buf bytes.Buffer
	require.NoError(t, instinct.ExportYAML(&buf, instincts))

	output := buf.String()
	assert.Contains(t, output, "id: first")
	assert.Contains(t, output, "id: second")
	assert.Contains(t, output, "source_repo: github.com/example/repo")
	assert.Contains(t, output, "Some content here")
}

func TestExportYAML_EmptySlice(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, instinct.ExportYAML(&buf, nil))
	assert.Empty(t, buf.String())
}

func TestExportJSON(t *testing.T) {
	now := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
	instincts := []instinct.Instinct{
		{
			ID:         "test-first",
			Trigger:    "when adding features",
			Confidence: 0.8,
			Domain:     "testing",
			Source:     "observation",
			SourceRepo: "",
			Content:    "",
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}

	var buf bytes.Buffer
	require.NoError(t, instinct.ExportJSON(&buf, instincts))

	var parsed []instinct.Instinct
	require.NoError(t, json.Unmarshal(buf.Bytes(), &parsed))
	require.Len(t, parsed, 1)
	assert.Equal(t, "test-first", parsed[0].ID)
	assert.Equal(t, "when adding features", parsed[0].Trigger)
	assert.InDelta(t, 0.8, parsed[0].Confidence, 0.001)
	assert.Equal(t, "testing", parsed[0].Domain)
}

func TestExportJSON_MultipleInstincts(t *testing.T) {
	now := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
	instincts := []instinct.Instinct{
		{
			ID:         "alpha",
			Trigger:    "trigger a",
			Confidence: 0.6,
			Domain:     "go",
			Source:     "observation",
			SourceRepo: "",
			Content:    "",
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:         "beta",
			Trigger:    "trigger b",
			Confidence: 0.9,
			Domain:     "python",
			Source:     "inherited",
			SourceRepo: "github.com/example/repo",
			Content:    "content",
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}

	var buf bytes.Buffer
	require.NoError(t, instinct.ExportJSON(&buf, instincts))

	var parsed []instinct.Instinct
	require.NoError(t, json.Unmarshal(buf.Bytes(), &parsed))
	require.Len(t, parsed, 2)
	assert.Equal(t, "alpha", parsed[0].ID)
	assert.Equal(t, "beta", parsed[1].ID)
	assert.Equal(t, "github.com/example/repo", parsed[1].SourceRepo)
}

func TestExportJSON_EmptySlice(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, instinct.ExportJSON(&buf, nil))

	var parsed []instinct.Instinct
	require.NoError(t, json.Unmarshal(buf.Bytes(), &parsed))
	assert.Empty(t, parsed)
}

func TestExport(t *testing.T) {
	now := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
	input := []instinct.Instinct{
		{
			ID:         "exp-1",
			Trigger:    "test trigger",
			Confidence: 0.8,
			Domain:     "go",
			Source:     "observation",
			SourceRepo: "",
			Content:    "",
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}

	tests := []struct {
		name    string
		format  string
		wantErr string
	}{
		{name: "yaml format", format: "yaml", wantErr: ""},
		{name: "json format", format: "json", wantErr: ""},
		{name: "unsupported format", format: "xml", wantErr: "unsupported export format"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := instinct.Export(&buf, input, tt.format)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, buf.String())
		})
	}
}
