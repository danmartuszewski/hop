package ssh

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/hop-cli/hop/internal/config"
)

func TestPrefixWriter(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		input    string
		expected string
	}{
		{
			name:     "single line without newline",
			prefix:   "[host] ",
			input:    "hello world",
			expected: "[host] hello world",
		},
		{
			name:     "single line with newline",
			prefix:   "[host] ",
			input:    "hello world\n",
			expected: "[host] hello world\n",
		},
		{
			name:     "multiple lines",
			prefix:   "[host] ",
			input:    "line1\nline2\nline3\n",
			expected: "[host] line1\n[host] line2\n[host] line3\n",
		},
		{
			name:     "empty input",
			prefix:   "[host] ",
			input:    "",
			expected: "",
		},
		{
			name:     "only newlines",
			prefix:   "[host] ",
			input:    "\n\n",
			expected: "[host] \n[host] \n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			pw := &prefixWriter{
				prefix:      tt.prefix,
				writer:      &buf,
				atLineStart: true,
			}

			_, err := pw.Write([]byte(tt.input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if buf.String() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, buf.String())
			}
		})
	}
}

func TestPrefixWriterMultipleWrites(t *testing.T) {
	var buf bytes.Buffer
	pw := &prefixWriter{
		prefix:      "[host] ",
		writer:      &buf,
		atLineStart: true,
	}

	// Write partial line
	pw.Write([]byte("hello "))
	// Write rest of line with newline
	pw.Write([]byte("world\n"))
	// Write another line
	pw.Write([]byte("next line\n"))

	expected := "[host] hello world\n[host] next line\n"
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}
}

func TestFormatGroupedOutput(t *testing.T) {
	results := []ExecResult{
		{
			Connection: &config.Connection{ID: "server1"},
			Stdout:     "hello from server1\n",
			ExitCode:   0,
		},
		{
			Connection: &config.Connection{ID: "server2"},
			Stdout:     "hello from server2\n",
			ExitCode:   0,
		},
	}

	output := FormatGroupedOutput(results)

	if !strings.Contains(output, "═══ server1 ═══") {
		t.Error("expected server1 header in output")
	}
	if !strings.Contains(output, "═══ server2 ═══") {
		t.Error("expected server2 header in output")
	}
	if !strings.Contains(output, "hello from server1") {
		t.Error("expected server1 output")
	}
	if !strings.Contains(output, "hello from server2") {
		t.Error("expected server2 output")
	}
}

func TestFormatGroupedOutputWithErrors(t *testing.T) {
	results := []ExecResult{
		{
			Connection: &config.Connection{ID: "server1"},
			Stdout:     "partial output\n",
			Error:      &mockError{msg: "connection timeout"},
			ExitCode:   1,
		},
	}

	output := FormatGroupedOutput(results)

	if !strings.Contains(output, "═══ server1 ═══") {
		t.Error("expected server1 header in output")
	}
	if !strings.Contains(output, "partial output") {
		t.Error("expected partial output")
	}
	if !strings.Contains(output, "Error:") {
		t.Error("expected error message")
	}
}

func TestHasErrors(t *testing.T) {
	tests := []struct {
		name     string
		results  []ExecResult
		expected bool
	}{
		{
			name: "no errors",
			results: []ExecResult{
				{Connection: &config.Connection{ID: "s1"}, ExitCode: 0},
				{Connection: &config.Connection{ID: "s2"}, ExitCode: 0},
			},
			expected: false,
		},
		{
			name: "one error",
			results: []ExecResult{
				{Connection: &config.Connection{ID: "s1"}, ExitCode: 0},
				{Connection: &config.Connection{ID: "s2"}, Error: &mockError{msg: "failed"}, ExitCode: 1},
			},
			expected: true,
		},
		{
			name:     "empty results",
			results:  []ExecResult{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasErrors(tt.results); got != tt.expected {
				t.Errorf("HasErrors() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCountErrors(t *testing.T) {
	results := []ExecResult{
		{Connection: &config.Connection{ID: "s1"}, ExitCode: 0},
		{Connection: &config.Connection{ID: "s2"}, Error: &mockError{msg: "failed"}, ExitCode: 1},
		{Connection: &config.Connection{ID: "s3"}, Error: &mockError{msg: "timeout"}, ExitCode: -1},
		{Connection: &config.Connection{ID: "s4"}, ExitCode: 0},
	}

	if got := CountErrors(results); got != 2 {
		t.Errorf("CountErrors() = %d, want 2", got)
	}
}

func TestExecOptionsDefaults(t *testing.T) {
	// Test that nil options don't panic
	connections := []config.Connection{}
	results := Execute(connections, nil)
	if len(results) != 0 {
		t.Errorf("expected empty results for empty connections")
	}
}

func TestExecResultDuration(t *testing.T) {
	// Verify that duration is a reasonable type
	result := ExecResult{
		Duration: 100 * time.Millisecond,
	}
	if result.Duration != 100*time.Millisecond {
		t.Error("duration not set correctly")
	}
}

// mockError is a simple error implementation for testing
type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}
