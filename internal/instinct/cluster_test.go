package instinct_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/riddopic/cc-tools/internal/instinct"
)

func newClusterInstinct(id, trigger, domain string, confidence float64) instinct.Instinct {
	now := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)

	return instinct.Instinct{
		ID:         id,
		Trigger:    trigger,
		Confidence: confidence,
		Domain:     domain,
		Source:     "observation",
		SourceRepo: "",
		Content:    "",
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func TestClusterInstincts(t *testing.T) {
	instincts := []instinct.Instinct{
		newClusterInstinct("a", "when writing functions", "code-style", 0.8),
		newClusterInstinct("b", "when writing tests", "testing", 0.7),
		newClusterInstinct("c", "when writing code", "code-style", 0.6),
		newClusterInstinct("d", "when writing modules", "code-style", 0.75),
		newClusterInstinct("e", "when debugging errors", "debugging", 0.5),
	}

	clusters := instinct.ClusterByTrigger(instincts, 3)
	assert.GreaterOrEqual(t, len(clusters), 1)

	// Verify cluster has 3+ members.
	for _, c := range clusters {
		assert.GreaterOrEqual(t, len(c.Members), 3)
	}
}

func TestClusterByTriggerSortedByMemberCount(t *testing.T) {
	instincts := []instinct.Instinct{
		newClusterInstinct("a", "when writing functions", "code-style", 0.8),
		newClusterInstinct("b", "when writing tests", "testing", 0.7),
		newClusterInstinct("c", "when writing code", "code-style", 0.6),
		newClusterInstinct("d", "debugging errors", "debugging", 0.5),
		newClusterInstinct("e", "debugging failures", "debugging", 0.4),
	}

	clusters := instinct.ClusterByTrigger(instincts, 2)
	assert.GreaterOrEqual(t, len(clusters), 2)

	// Clusters should be sorted by member count descending.
	for i := 1; i < len(clusters); i++ {
		assert.GreaterOrEqual(t, len(clusters[i-1].Members), len(clusters[i].Members))
	}
}

func TestClusterByTriggerAvgConfidence(t *testing.T) {
	instincts := []instinct.Instinct{
		newClusterInstinct("a", "writing functions", "", 0.8),
		newClusterInstinct("b", "writing tests", "", 0.6),
	}

	clusters := instinct.ClusterByTrigger(instincts, 2)
	assert.Len(t, clusters, 1)
	assert.InDelta(t, 0.7, clusters[0].AvgConfidence, 0.001)
}

func TestClusterByTriggerEmpty(t *testing.T) {
	clusters := instinct.ClusterByTrigger(nil, 3)
	assert.Empty(t, clusters)
}

func TestClusterByTriggerNoQualifying(t *testing.T) {
	instincts := []instinct.Instinct{
		newClusterInstinct("a", "when writing functions", "", 0.8),
		newClusterInstinct("b", "when debugging errors", "", 0.7),
	}

	// minSize 3 means no keyword can qualify (each appears once after stop word removal).
	clusters := instinct.ClusterByTrigger(instincts, 3)
	assert.Empty(t, clusters)
}

func TestNormalizeTrigger(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"when writing new functions", []string{"functions", "writing"}},
		{"when creating a module", []string{"module"}},
		{"when debugging errors in tests", []string{"debugging", "errors", "tests"}},
		{"", nil},
		{"when a the in", nil},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := instinct.NormalizeTrigger(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
