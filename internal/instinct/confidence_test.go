package instinct_test

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/riddopic/cc-tools/internal/instinct"
)

func TestConfidenceFromObservations(t *testing.T) {
	tests := []struct {
		name  string
		count int
		want  float64
	}{
		{name: "1 observation", count: 1, want: 0.3},
		{name: "2 observations", count: 2, want: 0.3},
		{name: "3 observations", count: 3, want: 0.5},
		{name: "5 observations", count: 5, want: 0.5},
		{name: "6 observations", count: 6, want: 0.7},
		{name: "10 observations", count: 10, want: 0.7},
		{name: "11 observations", count: 11, want: 0.85},
		{name: "20 observations", count: 20, want: 0.85},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := instinct.ConfidenceFromObservations(tt.count)
			assert.InDelta(t, tt.want, got, 0.001)
		})
	}
}

func TestClampConfidence(t *testing.T) {
	tests := []struct {
		name  string
		input float64
		want  float64
	}{
		{name: "below min clamps to 0.3", input: 0.1, want: 0.3},
		{name: "above max clamps to 0.9", input: 0.95, want: 0.9},
		{name: "in range unchanged", input: 0.6, want: 0.6},
		{name: "at min boundary", input: 0.3, want: 0.3},
		{name: "at max boundary", input: 0.9, want: 0.9},
		{name: "NaN clamps to min", input: math.NaN(), want: 0.3},
		{name: "positive infinity clamps to max", input: math.Inf(1), want: 0.9},
		{name: "negative infinity clamps to min", input: math.Inf(-1), want: 0.3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := instinct.ClampConfidence(tt.input)
			assert.InDelta(t, tt.want, got, 0.001)
		})
	}
}

func TestDecayConfidence(t *testing.T) {
	tests := []struct {
		name       string
		confidence float64
		weeks      int
		rate       float64
		want       float64
	}{
		{
			name:       "basic decay",
			confidence: 0.7,
			weeks:      2,
			rate:       0.02,
			want:       0.66,
		},
		{
			name:       "decay below min clamps",
			confidence: 0.35,
			weeks:      10,
			rate:       0.05,
			want:       0.3,
		},
		{
			name:       "zero weeks no change",
			confidence: 0.7,
			weeks:      0,
			rate:       0.05,
			want:       0.7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := instinct.DecayConfidence(tt.confidence, tt.weeks, tt.rate)
			assert.InDelta(t, tt.want, got, 0.001)
		})
	}
}

func TestElapsedWeeks(t *testing.T) {
	now := time.Date(2026, 2, 24, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		updatedAt time.Time
		want      int
	}{
		{
			name:      "zero time returns 0",
			updatedAt: time.Time{},
			want:      0,
		},
		{
			name:      "future timestamp returns 0",
			updatedAt: now.Add(24 * time.Hour),
			want:      0,
		},
		{
			name:      "same day returns 0",
			updatedAt: now,
			want:      0,
		},
		{
			name:      "6 days ago returns 0",
			updatedAt: now.AddDate(0, 0, -6),
			want:      0,
		},
		{
			name:      "7 days ago returns 1",
			updatedAt: now.AddDate(0, 0, -7),
			want:      1,
		},
		{
			name:      "14 days ago returns 2",
			updatedAt: now.AddDate(0, 0, -14),
			want:      2,
		},
		{
			name:      "20 days ago returns 2",
			updatedAt: now.AddDate(0, 0, -20),
			want:      2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := instinct.ElapsedWeeks(tt.updatedAt, now)
			assert.Equal(t, tt.want, got)
		})
	}
}

// newDecayTestInstinct creates an Instinct with all required fields set for tests.
func newDecayTestInstinct(id string, confidence float64, updatedAt time.Time) instinct.Instinct {
	return instinct.Instinct{
		ID:         id,
		Trigger:    "test-trigger",
		Confidence: confidence,
		Domain:     "testing",
		Source:     "personal",
		SourceRepo: "",
		Content:    "",
		CreatedAt:  updatedAt,
		UpdatedAt:  updatedAt,
	}
}

func TestApplyDecay(t *testing.T) {
	now := time.Date(2026, 2, 24, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		inst instinct.Instinct
		rate float64
		want float64
	}{
		{
			name: "no decay when rate is zero",
			inst: newDecayTestInstinct("decay-zero-rate", 0.8, now.AddDate(0, 0, -28)),
			rate: 0,
			want: 0.8,
		},
		{
			name: "no decay when rate is negative",
			inst: newDecayTestInstinct("decay-neg-rate", 0.8, now.AddDate(0, 0, -28)),
			rate: -0.01,
			want: 0.8,
		},
		{
			name: "no decay when updatedAt is zero",
			inst: newDecayTestInstinct("decay-zero-time", 0.8, time.Time{}),
			rate: 0.02,
			want: 0.8,
		},
		{
			name: "no decay when updatedAt is in the future",
			inst: newDecayTestInstinct("decay-future", 0.8, now.Add(48*time.Hour)),
			rate: 0.02,
			want: 0.8,
		},
		{
			name: "2 weeks of decay at 0.02",
			inst: newDecayTestInstinct("decay-2w", 0.8, now.AddDate(0, 0, -14)),
			rate: 0.02,
			want: 0.76,
		},
		{
			name: "decay clamps to min",
			inst: newDecayTestInstinct("decay-clamp", 0.35, now.AddDate(0, 0, -70)),
			rate: 0.05,
			want: 0.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := instinct.ApplyDecay(tt.inst, now, tt.rate)
			assert.InDelta(t, tt.want, got, 0.001)
		})
	}
}

func TestApplyDecayToSlice(t *testing.T) {
	now := time.Date(2026, 2, 24, 12, 0, 0, 0, time.UTC)

	instincts := []instinct.Instinct{
		newDecayTestInstinct("test-1", 0.8, now.AddDate(0, 0, -14)),
		newDecayTestInstinct("test-2", 0.6, now.AddDate(0, 0, -7)),
	}

	result := instinct.ApplyDecayToSlice(instincts, now, 0.02)

	assert.Len(t, result, 2)
	assert.InDelta(t, 0.76, result[0].Confidence, 0.001)
	assert.InDelta(t, 0.58, result[1].Confidence, 0.001)

	// Originals should be unchanged
	assert.InDelta(t, 0.8, instincts[0].Confidence, 0.001)
	assert.InDelta(t, 0.6, instincts[1].Confidence, 0.001)
}
