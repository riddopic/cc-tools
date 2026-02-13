package output_test

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/riddopic/cc-tools/internal/output"
)

func TestNewTerminal(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	term := output.NewTerminal(stdout, stderr)

	if term == nil {
		t.Fatal("NewTerminal() returned nil")
	}

	// Verify stdout routing by writing a message.
	err := term.Write("hello")
	if err != nil {
		t.Fatalf("Write() returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "hello") {
		t.Error("expected Write output to appear in stdout buffer")
	}

	if stderr.Len() > 0 {
		t.Error("expected stderr to be empty after Write")
	}

	// Verify stderr routing by writing an error message.
	stdout.Reset()

	err = term.WriteError("oops")
	if err != nil {
		t.Fatalf("WriteError() returned error: %v", err)
	}

	if !strings.Contains(stderr.String(), "oops") {
		t.Error("expected WriteError output to appear in stderr buffer")
	}

	if stdout.Len() > 0 {
		t.Error("expected stdout to be empty after WriteError")
	}

	// Verify all levels have styles by rendering a styled string for each.
	for level := output.Info; level <= output.Debug; level++ {
		styled := term.Style(level, "check")
		if styled == "" {
			t.Errorf("Style() returned empty string for level %d", level)
		}
	}
}

func TestTerminalWrite(t *testing.T) {
	tests := []struct {
		name    string
		message string
		wantErr bool
	}{
		{
			name:    "writes simple message",
			message: "Hello, World!",
			wantErr: false,
		},
		{
			name:    "writes empty message",
			message: "",
			wantErr: false,
		},
		{
			name:    "writes message with newlines",
			message: "Line 1\nLine 2",
			wantErr: false,
		},
		{
			name:    "writes message with special characters",
			message: "Special: !@#$%^&*()",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			term := output.NewTerminal(stdout, stderr)

			err := term.Write(tt.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				// Check message was written to stdout with newline.
				expected := tt.message + "\n"
				if stdout.String() != expected {
					t.Errorf("stdout = %q, want %q", stdout.String(), expected)
				}

				// Check nothing was written to stderr.
				if stderr.Len() > 0 {
					t.Errorf("stderr should be empty, got %q", stderr.String())
				}
			}
		})
	}
}

func TestTerminalWriteError(t *testing.T) {
	tests := []struct {
		name    string
		message string
		wantErr bool
	}{
		{
			name:    "writes error message",
			message: "Error occurred",
			wantErr: false,
		},
		{
			name:    "writes empty error",
			message: "",
			wantErr: false,
		},
		{
			name:    "writes multiline error",
			message: "Error:\nDetails here",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			term := output.NewTerminal(stdout, stderr)

			err := term.WriteError(tt.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteError() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				// Check message was written to stderr with newline.
				expected := tt.message + "\n"
				if stderr.String() != expected {
					t.Errorf("stderr = %q, want %q", stderr.String(), expected)
				}

				// Check nothing was written to stdout.
				if stdout.Len() > 0 {
					t.Errorf("stdout should be empty, got %q", stdout.String())
				}
			}
		})
	}
}

func TestTerminalPrint(t *testing.T) {
	tests := []struct {
		name   string
		level  output.Level
		format string
		args   []any
	}{
		{
			name:   "prints info message",
			level:  output.Info,
			format: "Info: %s",
			args:   []any{"test"},
		},
		{
			name:   "prints success message",
			level:  output.Success,
			format: "Success: %d items",
			args:   []any{42},
		},
		{
			name:   "prints warning message",
			level:  output.Warning,
			format: "Warning: %v",
			args:   []any{true},
		},
		{
			name:   "prints with no args",
			level:  output.Info,
			format: "Simple message",
			args:   []any{},
		},
		{
			name:   "prints with multiple args",
			level:  output.Info,
			format: "%s: %d of %d",
			args:   []any{"Progress", 5, 10},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			term := output.NewTerminal(stdout, stderr)

			err := term.Print(tt.level, tt.format, tt.args...)
			if err != nil {
				t.Fatalf("Print() returned error: %v", err)
			}

			// Check something was written to stdout.
			if stdout.Len() == 0 {
				t.Error("Nothing written to stdout")
			}

			// Check the formatted message is in the output.
			expectedMsg := formatMessage(tt.format, tt.args...)
			if !strings.Contains(stdout.String(), expectedMsg) {
				t.Errorf("stdout should contain %q, got %q", expectedMsg, stdout.String())
			}

			// Check nothing was written to stderr.
			if stderr.Len() > 0 {
				t.Errorf("stderr should be empty, got %q", stderr.String())
			}
		})
	}
}

func TestTerminalPrintError(t *testing.T) {
	tests := []struct {
		name   string
		level  output.Level
		format string
		args   []any
	}{
		{
			name:   "prints error level",
			level:  output.Error,
			format: "Error: %s",
			args:   []any{"failed"},
		},
		{
			name:   "prints debug level",
			level:  output.Debug,
			format: "Debug: %v",
			args:   []any{map[string]int{"key": 1}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			term := output.NewTerminal(stdout, stderr)

			err := term.PrintError(tt.level, tt.format, tt.args...)
			if err != nil {
				t.Fatalf("PrintError() returned error: %v", err)
			}

			// Check something was written to stderr.
			if stderr.Len() == 0 {
				t.Error("Nothing written to stderr")
			}

			// Check the formatted message is in the output.
			expectedMsg := formatMessage(tt.format, tt.args...)
			if !strings.Contains(stderr.String(), expectedMsg) {
				t.Errorf("stderr should contain %q, got %q", expectedMsg, stderr.String())
			}

			// Check nothing was written to stdout.
			if stdout.Len() > 0 {
				t.Errorf("stdout should be empty, got %q", stdout.String())
			}
		})
	}
}

type convenienceTestCase struct {
	name         string
	method       func(*output.Terminal, string, ...any) error
	format       string
	args         []any
	expectStdout bool
	expectStderr bool
}

func convenienceTestCases() []convenienceTestCase {
	return []convenienceTestCase{
		{
			name:         "Info writes to stdout",
			method:       (*output.Terminal).Info,
			format:       "Info: %s",
			args:         []any{"message"},
			expectStdout: true,
			expectStderr: false,
		},
		{
			name:         "Success writes to stdout",
			method:       (*output.Terminal).Success,
			format:       "Success: %d",
			args:         []any{100},
			expectStdout: true,
			expectStderr: false,
		},
		{
			name:         "Warning writes to stdout",
			method:       (*output.Terminal).Warning,
			format:       "Warning: %v",
			args:         []any{true},
			expectStdout: true,
			expectStderr: false,
		},
		{
			name:         "Error writes to stderr",
			method:       (*output.Terminal).Error,
			format:       "Error: %s",
			args:         []any{"failed"},
			expectStdout: false,
			expectStderr: true,
		},
		{
			name:         "Debug writes to stderr",
			method:       (*output.Terminal).Debug,
			format:       "Debug: %d",
			args:         []any{42},
			expectStdout: false,
			expectStderr: true,
		},
	}
}

func assertConvenienceOutput(t *testing.T, tc convenienceTestCase) {
	t.Helper()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	term := output.NewTerminal(stdout, stderr)

	err := tc.method(term, tc.format, tc.args...)
	if err != nil {
		t.Fatalf("method returned error: %v", err)
	}

	assertOutputDestination(t, tc, stdout, stderr)

	expectedMsg := formatMessage(tc.format, tc.args...)
	combined := stdout.String() + stderr.String()

	if !strings.Contains(combined, expectedMsg) {
		t.Errorf("output should contain %q, got %q", expectedMsg, combined)
	}
}

func assertOutputDestination(t *testing.T, tc convenienceTestCase, stdout, stderr *bytes.Buffer) {
	t.Helper()

	if tc.expectStdout && stdout.Len() == 0 {
		t.Error("expected output to stdout but got none")
	}

	if !tc.expectStdout && stdout.Len() > 0 {
		t.Errorf("unexpected stdout output: %q", stdout.String())
	}

	if tc.expectStderr && stderr.Len() == 0 {
		t.Error("expected output to stderr but got none")
	}

	if !tc.expectStderr && stderr.Len() > 0 {
		t.Errorf("unexpected stderr output: %q", stderr.String())
	}
}

func TestTerminalConvenienceMethods(t *testing.T) {
	for _, tc := range convenienceTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			assertConvenienceOutput(t, tc)
		})
	}
}

func TestTerminalRaw(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{
			name:    "writes raw string",
			message: "Raw output",
		},
		{
			name:    "writes empty string",
			message: "",
		},
		{
			name:    "writes string without adding newline",
			message: "No newline",
		},
		{
			name:    "writes string with existing newline",
			message: "Has newline\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			term := output.NewTerminal(stdout, stderr)

			err := term.Raw(tt.message)
			if err != nil {
				t.Fatalf("Raw() returned error: %v", err)
			}

			// Check exact output (no newline added).
			if stdout.String() != tt.message {
				t.Errorf("stdout = %q, want %q", stdout.String(), tt.message)
			}

			// Check nothing written to stderr.
			if stderr.Len() > 0 {
				t.Errorf("stderr should be empty, got %q", stderr.String())
			}
		})
	}
}

func TestTerminalRawError(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{
			name:    "writes raw error string",
			message: "Raw error",
		},
		{
			name:    "writes empty error string",
			message: "",
		},
		{
			name:    "writes error without adding newline",
			message: "No newline error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			term := output.NewTerminal(stdout, stderr)

			err := term.RawError(tt.message)
			if err != nil {
				t.Fatalf("RawError() returned error: %v", err)
			}

			// Check exact output (no newline added).
			if stderr.String() != tt.message {
				t.Errorf("stderr = %q, want %q", stderr.String(), tt.message)
			}

			// Check nothing written to stdout.
			if stdout.Len() > 0 {
				t.Errorf("stdout should be empty, got %q", stdout.String())
			}
		})
	}
}

func TestTerminalStyle(t *testing.T) {
	tests := []struct {
		name   string
		level  output.Level
		format string
		args   []any
	}{
		{
			name:   "styles info message",
			level:  output.Info,
			format: "Info: %s",
			args:   []any{"test"},
		},
		{
			name:   "styles success message",
			level:  output.Success,
			format: "Success!",
			args:   []any{},
		},
		{
			name:   "styles warning message",
			level:  output.Warning,
			format: "Warning: %d%%",
			args:   []any{75},
		},
		{
			name:   "styles error message",
			level:  output.Error,
			format: "Error: %v",
			args:   []any{errors.New("test error")},
		},
		{
			name:   "styles debug message",
			level:  output.Debug,
			format: "Debug: %+v",
			args:   []any{struct{ Name string }{"test"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			term := output.NewTerminal(stdout, stderr)

			styled := term.Style(tt.level, tt.format, tt.args...)

			// Check that something was returned.
			if styled == "" {
				t.Error("Style() returned empty string")
			}

			// Check that the formatted message is in the styled output.
			expectedMsg := formatMessage(tt.format, tt.args...)
			if !strings.Contains(styled, expectedMsg) {
				t.Errorf("styled output should contain %q, got %q", expectedMsg, styled)
			}

			// Verify nothing was written to stdout or stderr.
			if stdout.Len() > 0 {
				t.Error("Style() should not write to stdout")
			}
			if stderr.Len() > 0 {
				t.Error("Style() should not write to stderr")
			}
		})
	}
}

func TestWriterInterface(t *testing.T) {
	// Verify Terminal implements Writer interface.
	var _ output.Writer = (*output.Terminal)(nil)

	// Test using the interface.
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	var w output.Writer = output.NewTerminal(stdout, stderr)

	// Test Write.
	err := w.Write("test message")
	if err != nil {
		t.Errorf("Write() error = %v", err)
	}

	if !strings.Contains(stdout.String(), "test message") {
		t.Error("Write() should write to stdout")
	}

	// Test WriteError.
	err = w.WriteError("error message")
	if err != nil {
		t.Errorf("WriteError() error = %v", err)
	}

	if !strings.Contains(stderr.String(), "error message") {
		t.Error("WriteError() should write to stderr")
	}
}

func TestWriteFailureHandling(t *testing.T) {
	failWriter := &failingWriter{shouldFail: true}

	term := output.NewTerminal(failWriter, failWriter)

	// Test Write error handling.
	err := term.Write("test")
	if err == nil {
		t.Error("Write() should return error when write fails")
	}

	if !strings.Contains(err.Error(), "write to stdout") {
		t.Errorf("error should mention stdout, got %v", err)
	}

	// Test WriteError error handling.
	err = term.WriteError("test")
	if err == nil {
		t.Error("WriteError() should return error when write fails")
	}

	if !strings.Contains(err.Error(), "write to stderr") {
		t.Errorf("error should mention stderr, got %v", err)
	}
}

func TestColorRendering(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	term := output.NewTerminal(stdout, stderr)

	message := "Test Message"

	// Verify each level produces output containing the message.
	for level := output.Info; level <= output.Debug; level++ {
		styled := term.Style(level, "%s", message)

		if !strings.Contains(styled, message) {
			t.Errorf("styled output for level %d doesn't contain message", level)
		}
	}

	// In environments with color support, each level should produce unique
	// output. In NO_COLOR/CI environments lipgloss strips ANSI codes, so all
	// levels render identically. Detect this and skip the uniqueness check.
	outputs := make(map[string]bool)
	allUnique := true

	for level := output.Info; level <= output.Debug; level++ {
		styled := term.Style(level, "%s", message)
		if outputs[styled] {
			allUnique = false

			break
		}

		outputs[styled] = true
	}

	if !allUnique {
		t.Skip("lipgloss not applying colors (NO_COLOR or CI environment)")
	}
}

func TestLipglossIntegration(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	term := output.NewTerminal(stdout, stderr)

	// Verify that Style returns output containing the original message.
	message := "Integration Test"
	styled := term.Style(output.Info, "%s", message)

	if !strings.Contains(styled, message) {
		t.Errorf("styled output should contain the message, got %q", styled)
	}

	// In color-capable environments, the styled output should differ from plain text.
	if styled == message {
		t.Skip("lipgloss styling not applied (possibly NO_COLOR mode)")
	}
}

func TestConcurrentWrites(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	term := output.NewTerminal(stdout, stderr)

	done := make(chan bool, 5)

	go func() {
		for i := range 10 {
			_ = term.Info("Info %d", i)
		}
		done <- true
	}()

	go func() {
		for i := range 10 {
			_ = term.Success("Success %d", i)
		}
		done <- true
	}()

	go func() {
		for i := range 10 {
			_ = term.Warning("Warning %d", i)
		}
		done <- true
	}()

	go func() {
		for i := range 10 {
			_ = term.Error("Error %d", i)
		}
		done <- true
	}()

	go func() {
		for i := range 10 {
			_ = term.Debug("Debug %d", i)
		}
		done <- true
	}()

	// Wait for all goroutines.
	for range 5 {
		<-done
	}

	// Check that we have output (exact content may be interleaved).
	if stdout.Len() == 0 {
		t.Error("should have stdout output from concurrent writes")
	}

	if stderr.Len() == 0 {
		t.Error("should have stderr output from concurrent writes")
	}
}

// Helper types and functions.

type failingWriter struct {
	shouldFail bool
}

func (f *failingWriter) Write(p []byte) (int, error) {
	if f.shouldFail {
		return 0, errors.New("write failed")
	}

	return len(p), nil
}

func formatMessage(format string, args ...any) string {
	if len(args) == 0 {
		return format
	}

	return fmt.Sprintf(format, args...)
}
