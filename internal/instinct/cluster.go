package instinct

import (
	"sort"
	"strings"
)

// Cluster groups instincts that share common trigger keywords.
type Cluster struct {
	Members       []Instinct
	Keywords      []string
	AvgConfidence float64
}

// isStopWord reports whether w is a common word that should be filtered
// during trigger normalization.
func isStopWord(w string) bool {
	switch w {
	case "when", "creating", "new", "a", "the", "in", "for", "of",
		"with", "to", "an", "is", "are", "on", "at", "by", "as", "it":
		return true
	default:
		return false
	}
}

// NormalizeTrigger extracts meaningful keywords from a trigger string.
// It lowercases the input, splits on whitespace, removes stop words,
// and returns the remaining words sorted alphabetically.
func NormalizeTrigger(s string) []string {
	if s == "" {
		return nil
	}

	words := strings.Fields(strings.ToLower(s))

	var keywords []string

	for _, w := range words {
		if !isStopWord(w) {
			keywords = append(keywords, w)
		}
	}

	if len(keywords) == 0 {
		return nil
	}

	sort.Strings(keywords)

	return keywords
}

// ClusterByTrigger groups instincts by shared normalized trigger keywords.
// Only clusters with at least minSize members are returned, sorted by
// member count descending.
func ClusterByTrigger(instincts []Instinct, minSize int) []Cluster {
	if len(instincts) == 0 {
		return nil
	}

	index := buildInvertedIndex(instincts)
	clusters := collectClusters(index, minSize)

	sort.Slice(clusters, func(i, j int) bool {
		return len(clusters[i].Members) > len(clusters[j].Members)
	})

	return clusters
}

// buildInvertedIndex maps each normalized keyword to the instincts
// containing that keyword.
func buildInvertedIndex(instincts []Instinct) map[string][]Instinct {
	index := make(map[string][]Instinct)

	for _, inst := range instincts {
		keywords := NormalizeTrigger(inst.Trigger)
		for _, kw := range keywords {
			index[kw] = append(index[kw], inst)
		}
	}

	return index
}

// collectClusters creates Cluster values for keywords with enough members.
func collectClusters(index map[string][]Instinct, minSize int) []Cluster {
	var clusters []Cluster

	for keyword, members := range index {
		if len(members) < minSize {
			continue
		}

		clusters = append(clusters, Cluster{
			Members:       members,
			Keywords:      []string{keyword},
			AvgConfidence: avgConfidence(members),
		})
	}

	return clusters
}

// avgConfidence computes the mean confidence across instincts.
func avgConfidence(instincts []Instinct) float64 {
	if len(instincts) == 0 {
		return 0
	}

	var sum float64

	for _, inst := range instincts {
		sum += inst.Confidence
	}

	return sum / float64(len(instincts))
}
