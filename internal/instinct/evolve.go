package instinct

import "strings"

// EvolveOptions controls how instincts are analyzed for promotion candidates.
type EvolveOptions struct {
	ClusterThreshold   int
	CommandConfidence  float64
	CommandDomain      string
	AgentMinCluster    int
	AgentAvgConfidence float64
}

// SkillCandidate represents a cluster that could become a skill.
type SkillCandidate struct {
	Domain   string
	Count    int
	Keywords []string
}

// CommandCandidate represents an instinct suitable for promotion to a command.
type CommandCandidate struct {
	Trigger    string
	Confidence float64
}

// AgentCandidate represents a cluster suitable for promotion to an agent.
type AgentCandidate struct {
	Count         int
	AvgConfidence float64
	Keywords      []string
}

// EvolveResult holds all promotion candidates from an evolve analysis.
type EvolveResult struct {
	Skills   []SkillCandidate
	Commands []CommandCandidate
	Agents   []AgentCandidate
}

// Evolve analyzes instincts and returns skill, command, and agent candidates.
// It clusters instincts by trigger keywords and filters by the given options.
func Evolve(instincts []Instinct, opts EvolveOptions) *EvolveResult {
	result := &EvolveResult{
		Skills:   nil,
		Commands: nil,
		Agents:   nil,
	}

	if len(instincts) == 0 {
		return result
	}

	clusters := ClusterByTrigger(instincts, opts.ClusterThreshold)

	for _, c := range clusters {
		domain := DominantDomain(c.Members)
		if domain != "" {
			result.Skills = append(result.Skills, SkillCandidate{
				Domain:   domain,
				Count:    len(c.Members),
				Keywords: c.Keywords,
			})
		}

		if len(c.Members) >= opts.AgentMinCluster && c.AvgConfidence >= opts.AgentAvgConfidence {
			result.Agents = append(result.Agents, AgentCandidate{
				Count:         len(c.Members),
				AvgConfidence: c.AvgConfidence,
				Keywords:      c.Keywords,
			})
		}
	}

	for _, inst := range instincts {
		if inst.Confidence >= opts.CommandConfidence && inst.Domain == opts.CommandDomain {
			result.Commands = append(result.Commands, CommandCandidate{
				Trigger:    inst.Trigger,
				Confidence: inst.Confidence,
			})
		}
	}

	return result
}

// DominantDomain returns the most common domain among members.
// Empty domains are treated as "general". Returns "" for nil input.
func DominantDomain(members []Instinct) string {
	if len(members) == 0 {
		return ""
	}

	counts := make(map[string]int)

	for _, m := range members {
		d := m.Domain
		if d == "" {
			d = "general"
		}

		counts[d]++
	}

	var best string

	var bestCount int

	for d, count := range counts {
		if count > bestCount {
			best = d
			bestCount = count
		}
	}

	return best
}

// GroupByDomain groups instincts by their domain field.
// Instincts with an empty domain are placed in the "general" group.
func GroupByDomain(instincts []Instinct) map[string][]Instinct {
	groups := make(map[string][]Instinct)

	for _, inst := range instincts {
		d := inst.Domain
		if d == "" {
			d = "general"
		}

		groups[d] = append(groups[d], inst)
	}

	return groups
}

// SortedKeys returns map keys in sorted order.
func SortedKeys(m map[string][]Instinct) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sortStrings(keys)

	return keys
}

// sortStrings sorts a string slice in place.
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && strings.Compare(s[j-1], s[j]) > 0; j-- {
			s[j-1], s[j] = s[j], s[j-1]
		}
	}
}
