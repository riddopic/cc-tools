package instinct

import (
	"math"
	"time"
)

const (
	// MinConfidence is the lowest allowed confidence value.
	MinConfidence = 0.3
	// MaxConfidence is the highest allowed confidence value.
	MaxConfidence = 0.9
)

// Observation count thresholds for confidence scoring.
const (
	observationsHigh   = 11
	observationsMedium = 6
	observationsLow    = 3
)

// Confidence values for each observation tier.
const (
	confidenceHigh   = 0.85
	confidenceMedium = 0.7
	confidenceLow    = 0.5
)

// ConfidenceFromObservations returns base confidence for observation count.
func ConfidenceFromObservations(count int) float64 {
	switch {
	case count >= observationsHigh:
		return confidenceHigh
	case count >= observationsMedium:
		return confidenceMedium
	case count >= observationsLow:
		return confidenceLow
	default:
		return MinConfidence
	}
}

// ClampConfidence restricts a confidence value to [MinConfidence, MaxConfidence].
func ClampConfidence(c float64) float64 {
	if math.IsNaN(c) {
		return MinConfidence
	}

	if c < MinConfidence {
		return MinConfidence
	}

	if c > MaxConfidence {
		return MaxConfidence
	}

	return c
}

// DecayConfidence reduces confidence by rate per week for given weeks.
// The result is clamped to [MinConfidence, MaxConfidence].
func DecayConfidence(confidence float64, weeks int, rate float64) float64 {
	decayed := confidence - float64(weeks)*rate

	return ClampConfidence(decayed)
}

const daysPerWeek = 7

// ElapsedWeeks returns the number of complete weeks between updatedAt and now.
// Returns 0 if updatedAt is zero or in the future.
func ElapsedWeeks(updatedAt, now time.Time) int {
	if updatedAt.IsZero() || updatedAt.After(now) {
		return 0
	}

	days := int(now.Sub(updatedAt).Hours() / 24) //nolint:mnd // 24 hours per day
	return days / daysPerWeek
}

// ApplyDecay computes the decayed confidence for an instinct at the given time.
// Returns the original confidence when rate <= 0, updatedAt is zero, or
// updatedAt is in the future.
func ApplyDecay(inst Instinct, now time.Time, rate float64) float64 {
	if rate <= 0 {
		return inst.Confidence
	}

	weeks := ElapsedWeeks(inst.UpdatedAt, now)
	if weeks == 0 {
		return inst.Confidence
	}

	return DecayConfidence(inst.Confidence, weeks, rate)
}

// ApplyDecayToSlice returns a copy of the instinct slice with decayed
// confidence values. The original slice is not modified.
func ApplyDecayToSlice(instincts []Instinct, now time.Time, rate float64) []Instinct {
	result := make([]Instinct, len(instincts))
	copy(result, instincts)

	for i := range result {
		result[i].Confidence = ApplyDecay(result[i], now, rate)
	}

	return result
}
