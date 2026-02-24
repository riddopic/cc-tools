package instinct_test

import (
	"math"
	"testing"

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
