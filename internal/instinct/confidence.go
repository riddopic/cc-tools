package instinct

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
