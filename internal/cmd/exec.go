package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/hop-cli/hop/internal/config"
	"github.com/hop-cli/hop/internal/fuzzy"
	"github.com/hop-cli/hop/internal/ssh"
	"github.com/spf13/cobra"
)

var (
	execParallel int
	execTimeout  string
	execFailFast bool
	execStream   bool
	execDryRun   bool
	execTag      string
)

var execCmd = &cobra.Command{
	Use:   "exec <group|pattern> <command>",
	Short: "Execute a command on multiple servers",
	Long: `Execute a command on multiple servers in parallel.

Examples:
  hop exec myapp-prod "uptime"              # Execute on all servers in group
  hop exec myapp-prod "df -h" --parallel=2  # Limit parallelism
  hop exec prod "hostname" --stream         # Stream output in real-time
  hop exec "web*" "systemctl restart nginx" # Pattern match
  hop exec --tag=database "psql -c 'SELECT 1'" # Filter by tag`,
	Args: cobra.MinimumNArgs(2),
	RunE: runExec,
}

func init() {
	rootCmd.AddCommand(execCmd)

	execCmd.Flags().IntVar(&execParallel, "parallel", 10, "maximum parallel connections")
	execCmd.Flags().StringVar(&execTimeout, "timeout", "", "command timeout (e.g., 30s, 5m)")
	execCmd.Flags().BoolVar(&execFailFast, "fail-fast", false, "stop on first error")
	execCmd.Flags().BoolVar(&execStream, "stream", false, "stream output in real-time with host prefixes")
	execCmd.Flags().BoolVar(&execDryRun, "dry-run", false, "print SSH commands without executing")
	execCmd.Flags().StringVar(&execTag, "tag", "", "filter connections by tag")
}

func runExec(cmd *cobra.Command, args []string) error {
	groupOrPattern := args[0]
	command := strings.Join(args[1:], " ")

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	// Resolve connections
	connections, err := resolveExecTargets(groupOrPattern, cfg)
	if err != nil {
		return err
	}

	// Filter by tag if specified
	if execTag != "" {
		connections = fuzzy.MatchByTag(execTag, connections)
		if len(connections) == 0 {
			return fmt.Errorf("no connections found with tag '%s'", execTag)
		}
	}

	if len(connections) == 0 {
		return fmt.Errorf("no connections matching '%s'", groupOrPattern)
	}

	// Parse timeout
	var timeout time.Duration
	if execTimeout != "" {
		timeout, err = time.ParseDuration(execTimeout)
		if err != nil {
			return fmt.Errorf("invalid timeout: %w", err)
		}
	}

	// Handle dry-run
	if execDryRun {
		fmt.Fprintf(os.Stderr, "Would execute on %d server(s):\n\n", len(connections))
		for _, conn := range connections {
			opts := &ssh.ConnectOptions{Command: command}
			fmt.Printf("  %s: %s\n", conn.ID, ssh.BuildCommandString(&conn, opts))
		}
		return nil
	}

	// Print header
	if !quiet {
		fmt.Fprintf(os.Stderr, "Executing on %d server(s)...\n", len(connections))
		if execStream {
			fmt.Fprintf(os.Stderr, "\n")
		}
	}

	// Execute
	opts := &ssh.ExecOptions{
		Command:  command,
		Parallel: execParallel,
		Timeout:  timeout,
		FailFast: execFailFast,
		Stream:   execStream,
	}

	results := ssh.Execute(connections, opts)

	// Output results
	if !execStream {
		fmt.Print(ssh.FormatGroupedOutput(results))
	}

	// Print summary
	if !quiet {
		errCount := ssh.CountErrors(results)
		if errCount > 0 {
			fmt.Fprintf(os.Stderr, "\n%d of %d server(s) failed\n", errCount, len(results))
		} else {
			fmt.Fprintf(os.Stderr, "\nCompleted on %d server(s)\n", len(results))
		}
	}

	// Return error if any failed
	if ssh.HasErrors(results) {
		return fmt.Errorf("command failed on %d server(s)", ssh.CountErrors(results))
	}

	return nil
}

// resolveExecTargets resolves a group name, pattern, or connection ID to a list of connections.
func resolveExecTargets(groupOrPattern string, cfg *config.Config) ([]config.Connection, error) {
	// First, try as a defined group
	if cfg.Groups != nil {
		if members, ok := cfg.Groups[groupOrPattern]; ok {
			var connections []config.Connection
			for _, id := range members {
				if conn := fuzzy.FindByID(id, cfg.Connections); conn != nil {
					connections = append(connections, *conn)
				}
			}
			return connections, nil
		}
	}

	// Try as project-env pattern (e.g., "myapp-prod")
	matches := fuzzy.MatchGroup(groupOrPattern, cfg)
	if len(matches) > 0 {
		return matches, nil
	}

	// Try as glob pattern (e.g., "web*")
	if strings.Contains(groupOrPattern, "*") {
		pattern := globToRegex(groupOrPattern)
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern '%s': %w", groupOrPattern, err)
		}

		var connections []config.Connection
		for _, conn := range cfg.Connections {
			if re.MatchString(conn.ID) {
				connections = append(connections, conn)
			}
		}
		return connections, nil
	}

	// Try as a single connection ID (fuzzy match)
	if conn := fuzzy.FindBestMatch(groupOrPattern, cfg.Connections); conn != nil {
		return []config.Connection{*conn}, nil
	}

	return nil, nil
}

// globToRegex converts a simple glob pattern to a regex pattern.
func globToRegex(glob string) string {
	pattern := regexp.QuoteMeta(glob)
	pattern = strings.ReplaceAll(pattern, "\\*", ".*")
	pattern = strings.ReplaceAll(pattern, "\\?", ".")
	return "^" + pattern + "$"
}
