package instinct_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/instinct"
)

func TestParseFrontmatter(t *testing.T) {
	refTime := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	refTimeStr := refTime.Format(time.RFC3339)

	tests := []struct {
		name    string
		input   string
		want    []instinct.Instinct
		wantErr bool
	}{
		{
			name: "single instinct with content",
			input: "---\n" +
				"id: prefer-early-return\n" +
				"trigger: nested conditionals detected\n" +
				"confidence: 0.7\n" +
				"domain: go\n" +
				"source: observation\n" +
				"source_repo: github.com/example/repo\n" +
				"created_at: " + refTimeStr + "\n" +
				"updated_at: " + refTimeStr + "\n" +
				"---\n" +
				"Use guard clauses to reduce nesting.\n",
			wantErr: false,
			want: []instinct.Instinct{
				{
					ID:         "prefer-early-return",
					Trigger:    "nested conditionals detected",
					Confidence: 0.7,
					Domain:     "go",
					Source:     "observation",
					SourceRepo: "github.com/example/repo",
					Content:    "Use guard clauses to reduce nesting.\n",
					CreatedAt:  refTime,
					UpdatedAt:  refTime,
				},
			},
		},
		{
			name: "instinct without content",
			input: "---\n" +
				"id: no-content\n" +
				"trigger: some trigger\n" +
				"confidence: 0.5\n" +
				"domain: general\n" +
				"source: manual\n" +
				"created_at: " + refTimeStr + "\n" +
				"updated_at: " + refTimeStr + "\n" +
				"---\n",
			wantErr: false,
			want: []instinct.Instinct{
				{
					ID:         "no-content",
					Trigger:    "some trigger",
					Confidence: 0.5,
					Domain:     "general",
					Source:     "manual",
					SourceRepo: "",
					Content:    "",
					CreatedAt:  refTime,
					UpdatedAt:  refTime,
				},
			},
		},
		{
			name:    "empty input",
			input:   "",
			want:    nil,
			wantErr: false,
		},
		{
			name:    "no frontmatter delimiters",
			input:   "just some text without delimiters",
			want:    nil,
			wantErr: false,
		},
		{
			name: "missing id is skipped",
			input: "---\n" +
				"trigger: something\n" +
				"confidence: 0.5\n" +
				"domain: go\n" +
				"source: observation\n" +
				"created_at: " + refTimeStr + "\n" +
				"updated_at: " + refTimeStr + "\n" +
				"---\n",
			want:    nil,
			wantErr: false,
		},
		{
			name: "quoted values",
			input: "---\n" +
				"id: \"quoted-id\"\n" +
				"trigger: \"trigger with: colon\"\n" +
				"confidence: 0.6\n" +
				"domain: \"go\"\n" +
				"source: \"manual\"\n" +
				"created_at: " + refTimeStr + "\n" +
				"updated_at: " + refTimeStr + "\n" +
				"---\n",
			wantErr: false,
			want: []instinct.Instinct{
				{
					ID:         "quoted-id",
					Trigger:    "trigger with: colon",
					Confidence: 0.6,
					Domain:     "go",
					Source:     "manual",
					SourceRepo: "",
					Content:    "",
					CreatedAt:  refTime,
					UpdatedAt:  refTime,
				},
			},
		},
		{
			name: "unquoted values with colons",
			input: "---\n" +
				"id: unquoted-id\n" +
				"trigger: trigger with: colon in value\n" +
				"confidence: 0.5\n" +
				"domain: go\n" +
				"source: observation\n" +
				"created_at: " + refTimeStr + "\n" +
				"updated_at: " + refTimeStr + "\n" +
				"---\n",
			wantErr: false,
			want: []instinct.Instinct{
				{
					ID:         "unquoted-id",
					Trigger:    "trigger with: colon in value",
					Confidence: 0.5,
					Domain:     "go",
					Source:     "observation",
					SourceRepo: "",
					Content:    "",
					CreatedAt:  refTime,
					UpdatedAt:  refTime,
				},
			},
		},
		{
			name: "timestamps parsed as time.Time",
			input: "---\n" +
				"id: time-test\n" +
				"trigger: test\n" +
				"confidence: 0.5\n" +
				"domain: go\n" +
				"source: manual\n" +
				"created_at: 2025-06-15T08:00:00Z\n" +
				"updated_at: 2025-06-15T09:30:00Z\n" +
				"---\n",
			wantErr: false,
			want: []instinct.Instinct{
				{
					ID:         "time-test",
					Trigger:    "test",
					Confidence: 0.5,
					Domain:     "go",
					Source:     "manual",
					SourceRepo: "",
					Content:    "",
					CreatedAt:  time.Date(2025, 6, 15, 8, 0, 0, 0, time.UTC),
					UpdatedAt:  time.Date(2025, 6, 15, 9, 30, 0, 0, time.UTC),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := instinct.ParseFrontmatter(tt.input)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.want == nil {
				assert.Nil(t, got)
				return
			}

			require.Len(t, got, len(tt.want))

			for i := range tt.want {
				assert.Equal(t, tt.want[i].ID, got[i].ID, "ID mismatch")
				assert.Equal(t, tt.want[i].Trigger, got[i].Trigger, "Trigger mismatch")
				assert.InDelta(t, tt.want[i].Confidence, got[i].Confidence, 0.001, "Confidence mismatch")
				assert.Equal(t, tt.want[i].Domain, got[i].Domain, "Domain mismatch")
				assert.Equal(t, tt.want[i].Source, got[i].Source, "Source mismatch")
				assert.Equal(t, tt.want[i].SourceRepo, got[i].SourceRepo, "SourceRepo mismatch")
				assert.Equal(t, tt.want[i].Content, got[i].Content, "Content mismatch")
				assert.True(t, tt.want[i].CreatedAt.Equal(got[i].CreatedAt), "CreatedAt mismatch")
				assert.True(t, tt.want[i].UpdatedAt.Equal(got[i].UpdatedAt), "UpdatedAt mismatch")
			}
		})
	}
}
