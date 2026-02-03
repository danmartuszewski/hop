package ssh

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/danmartuszewski/hop/internal/config"
)

// ExecOptions configures the behavior of parallel execution.
type ExecOptions struct {
	// Command is the command to execute on remote hosts.
	Command string
	// Parallel is the maximum number of concurrent connections (default: 10).
	Parallel int
	// Timeout is the maximum duration for each command (0 means no timeout).
	Timeout time.Duration
	// FailFast stops execution on the first error if true.
	FailFast bool
	// Stream enables real-time output streaming with host prefixes.
	Stream bool
	// DryRun prints commands without executing if true.
	DryRun bool
}

// ExecResult holds the result of executing a command on a single host.
type ExecResult struct {
	Connection *config.Connection
	Stdout     string
	Stderr     string
	ExitCode   int
	Error      error
	Duration   time.Duration
}

// Execute runs a command on multiple connections in parallel.
func Execute(connections []config.Connection, opts *ExecOptions) []ExecResult {
	if opts == nil {
		opts = &ExecOptions{}
	}

	// Set defaults
	parallel := opts.Parallel
	if parallel <= 0 {
		parallel = 10
	}

	results := make([]ExecResult, len(connections))
	resultChan := make(chan struct {
		index  int
		result ExecResult
	}, len(connections))

	// Create context for timeout and cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Semaphore for limiting parallelism
	sem := make(chan struct{}, parallel)

	// WaitGroup for tracking completion
	var wg sync.WaitGroup

	// Track if we should stop due to fail-fast
	var failFastTriggered bool
	var failFastMu sync.Mutex

	for i := range connections {
		wg.Add(1)
		go func(index int, conn config.Connection) {
			defer wg.Done()

			// Check fail-fast before acquiring semaphore
			if opts.FailFast {
				failFastMu.Lock()
				triggered := failFastTriggered
				failFastMu.Unlock()
				if triggered {
					resultChan <- struct {
						index  int
						result ExecResult
					}{
						index: index,
						result: ExecResult{
							Connection: &conn,
							Error:      fmt.Errorf("skipped due to fail-fast"),
							ExitCode:   -1,
						},
					}
					return
				}
			}

			// Acquire semaphore
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				resultChan <- struct {
					index  int
					result ExecResult
				}{
					index: index,
					result: ExecResult{
						Connection: &conn,
						Error:      ctx.Err(),
						ExitCode:   -1,
					},
				}
				return
			}

			result := executeOnHost(ctx, &conn, opts)

			// Check if we should trigger fail-fast
			if opts.FailFast && result.Error != nil {
				failFastMu.Lock()
				failFastTriggered = true
				failFastMu.Unlock()
				cancel()
			}

			resultChan <- struct {
				index  int
				result ExecResult
			}{
				index:  index,
				result: result,
			}
		}(i, connections[i])
	}

	// Close result channel when all goroutines are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	for r := range resultChan {
		results[r.index] = r.result
	}

	return results
}

// executeOnHost runs the command on a single host.
func executeOnHost(ctx context.Context, conn *config.Connection, opts *ExecOptions) ExecResult {
	start := time.Now()

	result := ExecResult{
		Connection: conn,
	}

	// Build SSH command args
	args := BuildCommand(conn, &ConnectOptions{
		Command: opts.Command,
	})

	// Create command with context
	var execCtx context.Context
	var execCancel context.CancelFunc
	if opts.Timeout > 0 {
		execCtx, execCancel = context.WithTimeout(ctx, opts.Timeout)
	} else {
		execCtx, execCancel = context.WithCancel(ctx)
	}
	defer execCancel()

	cmd := exec.CommandContext(execCtx, "ssh", args...)

	var stdout, stderr bytes.Buffer

	if opts.Stream {
		// In stream mode, we write to stdout/stderr with prefixes
		cmd.Stdout = &prefixWriter{prefix: fmt.Sprintf("[%s] ", conn.ID), writer: os.Stdout}
		cmd.Stderr = &prefixWriter{prefix: fmt.Sprintf("[%s] ", conn.ID), writer: os.Stderr}
	} else {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	}

	err := cmd.Run()
	result.Duration = time.Since(start)

	if !opts.Stream {
		result.Stdout = stdout.String()
		result.Stderr = stderr.String()
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = -1
		}
		result.Error = err
	}

	return result
}

// prefixWriter adds a prefix to each line written.
type prefixWriter struct {
	prefix      string
	writer      io.Writer
	mu          sync.Mutex
	atLineStart bool
}

func newPrefixWriter(prefix string, writer io.Writer) *prefixWriter {
	return &prefixWriter{
		prefix:      prefix,
		writer:      writer,
		atLineStart: true,
	}
}

func (pw *prefixWriter) Write(p []byte) (n int, err error) {
	pw.mu.Lock()
	defer pw.mu.Unlock()

	originalLen := len(p)

	for len(p) > 0 {
		if pw.atLineStart {
			_, err = pw.writer.Write([]byte(pw.prefix))
			if err != nil {
				return originalLen - len(p), err
			}
			pw.atLineStart = false
		}

		// Find the next newline
		idx := bytes.IndexByte(p, '\n')
		if idx == -1 {
			// No newline, write the rest
			_, err = pw.writer.Write(p)
			return originalLen, err
		}

		// Write up to and including the newline
		_, err = pw.writer.Write(p[:idx+1])
		if err != nil {
			return originalLen - len(p), err
		}
		p = p[idx+1:]
		pw.atLineStart = true
	}

	return originalLen, nil
}

// FormatGroupedOutput formats results in grouped mode.
func FormatGroupedOutput(results []ExecResult) string {
	var buf bytes.Buffer

	for i, result := range results {
		if i > 0 {
			buf.WriteString("\n")
		}
		fmt.Fprintf(&buf, "═══ %s ═══\n", result.Connection.ID)

		if result.Error != nil {
			if result.Stdout != "" {
				buf.WriteString(result.Stdout)
			}
			if result.Stderr != "" {
				buf.WriteString(result.Stderr)
			}
			fmt.Fprintf(&buf, "Error: %v (exit code: %d)\n", result.Error, result.ExitCode)
		} else {
			if result.Stdout != "" {
				buf.WriteString(result.Stdout)
			}
			if result.Stderr != "" {
				buf.WriteString(result.Stderr)
			}
		}
	}

	return buf.String()
}

// HasErrors returns true if any result has an error.
func HasErrors(results []ExecResult) bool {
	for _, r := range results {
		if r.Error != nil {
			return true
		}
	}
	return false
}

// CountErrors returns the number of failed executions.
func CountErrors(results []ExecResult) int {
	count := 0
	for _, r := range results {
		if r.Error != nil {
			count++
		}
	}
	return count
}
