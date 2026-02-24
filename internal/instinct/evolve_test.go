package instinct_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/instinct"
)

func makeEvolveInstinct(id, trigger, domain string, conf float64) instinct.Instinct {
	now := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)

	return instinct.Instinct{
		ID:         id,
		Trigger:    trigger,
		Confidence: conf,
		Domain:     domain,
		Source:     "observation",
		SourceRepo: "",
		Content:    "",
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func TestEvolve(t *testing.T) {
	instincts := []instinct.Instinct{
		makeEvolveInstinct("go-1", "error handling patterns", "go", 0.9),
		makeEvolveInstinct("go-2", "error wrapping patterns", "go", 0.85),
		makeEvolveInstinct("go-3", "error checking patterns", "go", 0.8),
		makeEvolveInstinct("test-1", "table driven tests", "testing", 0.75),
		makeEvolveInstinct("test-2", "table test patterns", "testing", 0.7),
		makeEvolveInstinct("test-3", "table assertions", "testing", 0.65),
		makeEvolveInstinct("wf-1", "commit workflow", "workflow", 0.8),
		makeEvolveInstinct("wf-2", "deploy workflow", "workflow", 0.75),
		makeEvolveInstinct("sec-1", "input validation", "security", 0.7),
		makeEvolveInstinct("sec-2", "path validation", "security", 0.65),
	}

	opts := instinct.EvolveOptions{
		ClusterThreshold:   2,
		CommandConfidence:  0.7,
		CommandDomain:      "workflow",
		AgentMinCluster:    3,
		AgentAvgConfidence: 0.75,
	}

	result := instinct.Evolve(instincts, opts)
	require.NotNil(t, result)

	assert.NotEmpty(t, result.Skills,
		"expected skill candidates from clustered instincts")

	assert.NotEmpty(t, result.Commands,
		"expected command candidates from workflow instincts")

	for _, cmd := range result.Commands {
		assert.GreaterOrEqual(t, cmd.Confidence, opts.CommandConfidence)
	}
}

func TestEvolve_EmptyInput(t *testing.T) {
	opts := instinct.EvolveOptions{
		ClusterThreshold:   2,
		CommandConfidence:  0,
		CommandDomain:      "",
		AgentMinCluster:    0,
		AgentAvgConfidence: 0,
	}
	result := instinct.Evolve(nil, opts)
	require.NotNil(t, result)
	assert.Empty(t, result.Skills)
	assert.Empty(t, result.Commands)
	assert.Empty(t, result.Agents)
}

func TestDominantDomain(t *testing.T) {
	tests := []struct {
		name    string
		members []instinct.Instinct
		want    string
	}{
		{
			name: "single domain",
			members: []instinct.Instinct{
				makeEvolveInstinct("a", "trigger", "go", 0.5),
				makeEvolveInstinct("b", "trigger", "go", 0.5),
			},
			want: "go",
		},
		{
			name:    "empty members",
			members: nil,
			want:    "",
		},
		{
			name: "empty domain defaults to general",
			members: []instinct.Instinct{
				makeEvolveInstinct("a", "trigger", "", 0.5),
				makeEvolveInstinct("b", "trigger", "", 0.5),
			},
			want: "general",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := instinct.DominantDomain(tt.members)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDominantDomain_DeterministicTiebreak(t *testing.T) {
	// Two domains with equal counts: "beta" and "alpha" each have 2 members.
	// The lexicographically first domain ("alpha") should always win.
	members := []instinct.Instinct{
		makeEvolveInstinct("a", "trigger", "beta", 0.5),
		makeEvolveInstinct("b", "trigger", "alpha", 0.5),
		makeEvolveInstinct("c", "trigger", "beta", 0.5),
		makeEvolveInstinct("d", "trigger", "alpha", 0.5),
	}

	// Run multiple times to verify determinism despite map iteration order.
	for range 100 {
		got := instinct.DominantDomain(members)
		assert.Equal(t, "alpha", got, "tie-breaking must pick lexicographically first domain")
	}
}

func TestSortedKeys(t *testing.T) {
	groups := map[string][]instinct.Instinct{
		"zebra":  {makeEvolveInstinct("z", "trigger", "zebra", 0.5)},
		"alpha":  {makeEvolveInstinct("a", "trigger", "alpha", 0.5)},
		"middle": {makeEvolveInstinct("m", "trigger", "middle", 0.5)},
		"beta":   {makeEvolveInstinct("b", "trigger", "beta", 0.5)},
	}

	keys := instinct.SortedKeys(groups)
	assert.Equal(t, []string{"alpha", "beta", "middle", "zebra"}, keys)
}

func TestGroupByDomain(t *testing.T) {
	instincts := []instinct.Instinct{
		makeEvolveInstinct("a", "trigger", "go", 0.5),
		makeEvolveInstinct("b", "trigger", "testing", 0.5),
		makeEvolveInstinct("c", "trigger", "go", 0.5),
		makeEvolveInstinct("d", "trigger", "", 0.5),
	}

	groups := instinct.GroupByDomain(instincts)
	assert.Len(t, groups["go"], 2)
	assert.Len(t, groups["testing"], 1)
	assert.Len(t, groups["general"], 1, "empty domain should map to general")
}
