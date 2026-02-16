//go:build testmode

package output_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/riddopic/cc-tools/internal/output"
	"github.com/riddopic/cc-tools/internal/shared"
)

func TestHookFormatter_NewHookFormatter(t *testing.T) {
	hf := output.NewHookFormatter()
	assert.NotNil(t, hf, "NewHookFormatter() should return a non-nil formatter")
}

func TestHookFormatter_FormatSuccess(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    string
	}{
		{
			name:    "non-empty message",
			message: "all checks passed",
			want:    shared.ANSIGreen + "all checks passed" + shared.ANSIReset,
		},
		{
			name:    "empty message",
			message: "",
			want:    shared.ANSIGreen + "" + shared.ANSIReset,
		},
	}

	hf := output.NewHookFormatter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hf.FormatSuccess(tt.message)
			assert.Equal(t, tt.want, got)
			assert.Contains(t, got, shared.ANSIGreen)
			assert.Contains(t, got, shared.ANSIReset)
			assert.Contains(t, got, tt.message)
		})
	}
}

func TestHookFormatter_FormatWarning(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    string
	}{
		{
			name:    "non-empty message",
			message: "something may be wrong",
			want:    shared.ANSIYellow + "something may be wrong" + shared.ANSIReset,
		},
		{
			name:    "empty message",
			message: "",
			want:    shared.ANSIYellow + "" + shared.ANSIReset,
		},
	}

	hf := output.NewHookFormatter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hf.FormatWarning(tt.message)
			assert.Equal(t, tt.want, got)
			assert.Contains(t, got, shared.ANSIYellow)
			assert.Contains(t, got, shared.ANSIReset)
			assert.Contains(t, got, tt.message)
		})
	}
}

func TestHookFormatter_FormatError(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    string
	}{
		{
			name:    "non-empty message",
			message: "validation failed",
			want:    shared.ANSIRed + "validation failed" + shared.ANSIReset,
		},
		{
			name:    "empty message",
			message: "",
			want:    shared.ANSIRed + "" + shared.ANSIReset,
		},
	}

	hf := output.NewHookFormatter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hf.FormatError(tt.message)
			assert.Equal(t, tt.want, got)
			assert.Contains(t, got, shared.ANSIRed)
			assert.Contains(t, got, shared.ANSIReset)
			assert.Contains(t, got, tt.message)
		})
	}
}

func TestHookFormatter_FormatBlockingError(t *testing.T) {
	tests := []struct {
		name   string
		format string
		args   []any
		want   string
	}{
		{
			name:   "format with string arg",
			format: "blocking: %s",
			args:   []any{"lint failed"},
			want:   shared.ANSIRed + "blocking: lint failed" + shared.ANSIReset,
		},
		{
			name:   "format with multiple args",
			format: "%d errors in %s",
			args:   []any{3, "main.go"},
			want:   shared.ANSIRed + "3 errors in main.go" + shared.ANSIReset,
		},
		{
			name:   "format with no args",
			format: "simple blocking error",
			args:   nil,
			want:   shared.ANSIRed + "simple blocking error" + shared.ANSIReset,
		},
	}

	hf := output.NewHookFormatter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hf.FormatBlockingError(tt.format, tt.args...)
			assert.Equal(t, tt.want, got)
			assert.Contains(t, got, shared.ANSIRed)
			assert.Contains(t, got, shared.ANSIReset)
		})
	}
}

func TestHookFormatter_FormatTestPass(t *testing.T) {
	hf := output.NewHookFormatter()
	got := hf.FormatTestPass()

	assert.Contains(t, got, "Tests pass")
	assert.Contains(t, got, shared.ANSIYellow)
	assert.Contains(t, got, shared.ANSIReset)

	expected := shared.ANSIYellow + "\U0001f449 Tests pass. Continue with your task." + shared.ANSIReset
	assert.Equal(t, expected, got)
}

func TestHookFormatter_FormatLintPass(t *testing.T) {
	hf := output.NewHookFormatter()
	got := hf.FormatLintPass()

	assert.Contains(t, got, "Lints pass")
	assert.Contains(t, got, shared.ANSIYellow)
	assert.Contains(t, got, shared.ANSIReset)

	expected := shared.ANSIYellow + "\U0001f449 Lints pass. Continue with your task." + shared.ANSIReset
	assert.Equal(t, expected, got)
}

func TestHookFormatter_FormatValidationPass(t *testing.T) {
	hf := output.NewHookFormatter()
	got := hf.FormatValidationPass()

	assert.Contains(t, got, "Validations pass")
	assert.Contains(t, got, shared.ANSIYellow)
	assert.Contains(t, got, shared.ANSIReset)

	expected := shared.ANSIYellow + "\U0001f449 Validations pass. Continue with your task." + shared.ANSIReset
	assert.Equal(t, expected, got)
}
