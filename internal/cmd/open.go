package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/hop-cli/hop/internal/config"
	"github.com/hop-cli/hop/internal/fuzzy"
	"github.com/hop-cli/hop/internal/ssh"
	"github.com/spf13/cobra"
)

var (
	openDryRun   bool
	openTag      string
	openDelay    int
	openForceTTY bool
)

var openCmd = &cobra.Command{
	Use:   "open <group|id...> [-- \"<cmd>\"]",
	Short: "Open multiple terminal tabs",
	Long: `Open multiple terminal tabs, each connected to a different server.

Examples:
  hop open myapp-prod                    # Open all servers in group
  hop open web1 db1 api1                 # Open specific servers
  hop open myapp-prod -- "htop"          # Open with initial command
  hop open --tag=web                     # Open all servers tagged 'web'
  hop open prod --tag=database -- "psql" # Combine group and tag filter`,
	Args: func(cmd *cobra.Command, args []string) error {
		// Allow zero args if --tag is provided
		if openTag != "" {
			return nil
		}
		if len(args) == 0 {
			return fmt.Errorf("requires at least 1 argument (group or connection ID)")
		}
		return nil
	},
	RunE: runOpen,
}

func init() {
	rootCmd.AddCommand(openCmd)

	openCmd.Flags().BoolVar(&openDryRun, "dry-run", false, "show what would be opened without executing")
	openCmd.Flags().StringVar(&openTag, "tag", "", "filter connections by tag")
	openCmd.Flags().IntVar(&openDelay, "delay", 100, "delay in milliseconds between opening tabs")
	openCmd.Flags().BoolVarP(&openForceTTY, "t", "t", false, "force TTY allocation")
}

func runOpen(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	// Parse args to extract remote command (after --)
	var remoteCmd string
	var queries []string
	for i, arg := range args {
		if arg == "--" {
			if i+1 < len(args) {
				remoteCmd = strings.Join(args[i+1:], " ")
			}
			queries = args[:i]
			break
		}
	}
	if queries == nil {
		queries = args
	}

	// Collect connections to open
	connections, err := resolveConnections(cfg, queries, openTag)
	if err != nil {
		return err
	}

	if len(connections) == 0 {
		return fmt.Errorf("no connections found matching the criteria")
	}

	// Detect terminal
	terminal := ssh.DetectTerminal()
	if !terminal.SupportsNewTab() {
		return fmt.Errorf("terminal %q does not support opening new tabs", terminal)
	}

	if openDryRun {
		return printDryRun(connections, terminal, remoteCmd)
	}

	return openTabs(connections, terminal, remoteCmd)
}

// resolveConnections resolves queries and tag filters to a list of connections.
func resolveConnections(cfg *config.Config, queries []string, tag string) ([]config.Connection, error) {
	var connections []config.Connection
	seen := make(map[string]bool)

	// If tag is specified without queries, get all connections with that tag
	if len(queries) == 0 && tag != "" {
		connections = fuzzy.MatchByTag(tag, cfg.Connections)
		return connections, nil
	}

	// Process each query
	for _, query := range queries {
		// First, try to match as a group
		groupMatches := fuzzy.MatchGroup(query, cfg)
		if len(groupMatches) > 0 {
			for _, conn := range groupMatches {
				if !seen[conn.ID] {
					seen[conn.ID] = true
					connections = append(connections, conn)
				}
			}
			continue
		}

		// Then try fuzzy matching individual connections
		matches := fuzzy.FindMatches(query, cfg.Connections)
		if len(matches) == 0 {
			return nil, fmt.Errorf("no connections matching '%s'", query)
		}

		// If exact match (score 1000), use only that one
		if matches[0].Score == 1000 {
			conn := matches[0].Connection
			if !seen[conn.ID] {
				seen[conn.ID] = true
				connections = append(connections, *conn)
			}
		} else if len(matches) == 1 {
			// Single match
			conn := matches[0].Connection
			if !seen[conn.ID] {
				seen[conn.ID] = true
				connections = append(connections, *conn)
			}
		} else {
			// Multiple fuzzy matches - add all of them
			for _, m := range matches {
				if !seen[m.Connection.ID] {
					seen[m.Connection.ID] = true
					connections = append(connections, *m.Connection)
				}
			}
		}
	}

	// Apply tag filter if specified
	if tag != "" {
		connections = filterByTag(connections, tag)
	}

	return connections, nil
}

// filterByTag filters connections that have the specified tag.
func filterByTag(connections []config.Connection, tag string) []config.Connection {
	tagLower := strings.ToLower(tag)
	var filtered []config.Connection

	for _, conn := range connections {
		for _, t := range conn.Tags {
			if strings.ToLower(t) == tagLower {
				filtered = append(filtered, conn)
				break
			}
		}
	}

	return filtered
}

func printDryRun(connections []config.Connection, terminal ssh.TerminalType, remoteCmd string) error {
	fmt.Printf("Would open %d tab(s) in %s:\n\n", len(connections), terminal)

	for _, conn := range connections {
		opts := &ssh.ConnectOptions{
			ForceTTY: openForceTTY || remoteCmd != "",
			Command:  remoteCmd,
		}
		cmdStr := ssh.BuildCommandString(&conn, opts)
		fmt.Printf("  %s (%s)\n", conn.ID, conn.Host)
		fmt.Printf("    %s\n\n", cmdStr)
	}

	return nil
}

func openTabs(connections []config.Connection, terminal ssh.TerminalType, remoteCmd string) error {
	if !quiet {
		fmt.Fprintf(os.Stderr, "Opening %d tab(s) in %s...\n", len(connections), terminal)
	}

	for i, conn := range connections {
		opts := &ssh.ConnectOptions{
			ForceTTY: openForceTTY || remoteCmd != "",
			Command:  remoteCmd,
		}
		cmdStr := ssh.BuildCommandString(&conn, opts)

		if !quiet {
			fmt.Fprintf(os.Stderr, "  Opening %s (%s)...\n", conn.ID, conn.Host)
		}

		if err := terminal.OpenNewTab(cmdStr); err != nil {
			return fmt.Errorf("failed to open tab for %s: %w", conn.ID, err)
		}

		// Add delay between tabs (except for the last one)
		if i < len(connections)-1 && openDelay > 0 {
			time.Sleep(time.Duration(openDelay) * time.Millisecond)
		}
	}

	if !quiet {
		fmt.Fprintf(os.Stderr, "Done.\n")
	}

	return nil
}
