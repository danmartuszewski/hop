package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/danmartuszewski/hop/internal/config"
	"github.com/danmartuszewski/hop/internal/fuzzy"
	"github.com/danmartuszewski/hop/internal/picker"
	"github.com/danmartuszewski/hop/internal/ssh"
	"github.com/spf13/cobra"
)

var (
	dryRun   bool
	forceTTY bool
)

var connectCmd = &cobra.Command{
	Use:   "connect <id>",
	Short: "Connect to a server by exact ID",
	Long:  "Connect to a server using its exact connection ID.",
	Args:  cobra.ExactArgs(1),
	RunE:  runConnect,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) >= 1 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return getConnectionCompletions(toComplete)
	},
}

func init() {
	rootCmd.AddCommand(connectCmd)

	// Add flags to both root command (for quick connect) and connect command
	for _, cmd := range []*cobra.Command{rootCmd, connectCmd} {
		cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print SSH command without executing")
		cmd.Flags().BoolVarP(&forceTTY, "t", "t", false, "force TTY allocation")
	}
}

func runConnect(cmd *cobra.Command, args []string) error {
	id := args[0]

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	conn := fuzzy.FindByID(id, cfg.Connections)
	if conn == nil {
		return fmt.Errorf("connection '%s' not found", id)
	}

	return connectToServer(conn, cmd)
}

func runQuickConnect(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}

	query := args[0]
	var remoteCmd string

	for i, arg := range args {
		if arg == "--" && i+1 < len(args) {
			remoteCmd = strings.Join(args[i+1:], " ")
			args = args[:i]
			break
		}
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	matches := fuzzy.FindMatches(query, cfg.Connections)
	if len(matches) == 0 {
		return fmt.Errorf("no connections matching '%s'", query)
	}

	var conn *config.Connection
	if len(matches) == 1 {
		conn = matches[0].Connection
	} else if matches[0].Score == 1000 {
		// Exact match takes priority
		conn = matches[0].Connection
	} else {
		// Multiple matches - show interactive picker
		if !quiet {
			fmt.Fprintf(os.Stderr, "Multiple matches for '%s':\n", query)
		}
		selected, err := picker.SelectConnection(matches)
		if err != nil {
			return err
		}
		conn = selected
	}

	opts := &ssh.ConnectOptions{
		DryRun:   dryRun,
		ForceTTY: forceTTY,
		Command:  remoteCmd,
	}

	if !quiet {
		fmt.Fprintf(os.Stderr, "Connecting to %s (%s)...\n", conn.ID, conn.Host)
	}

	return ssh.Connect(conn, opts)
}

func connectToServer(conn *config.Connection, cmd *cobra.Command) error {
	var remoteCmd string
	args := cmd.Flags().Args()
	for i, arg := range args {
		if arg == "--" && i+1 < len(args) {
			remoteCmd = strings.Join(args[i+1:], " ")
			break
		}
	}

	opts := &ssh.ConnectOptions{
		DryRun:   dryRun,
		ForceTTY: forceTTY,
		Command:  remoteCmd,
	}

	if !quiet {
		fmt.Fprintf(os.Stderr, "Connecting to %s (%s)...\n", conn.ID, conn.Host)
	}

	return ssh.Connect(conn, opts)
}

func loadConfig() (*config.Config, error) {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}
