package cmd

import (
	"fmt"
	"os"

	"github.com/danmartuszewski/hop/internal/ssh"
	"github.com/danmartuszewski/hop/internal/tui"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	verbose bool
	quiet   bool
)

var rootCmd = &cobra.Command{
	Use:   "hop [query]",
	Short: "SSH Connection Manager",
	Long: `hop - Quick, elegant SSH connection management for terminal enthusiasts

Usage:
  hop                         Open TUI dashboard
  hop <query>                 Quick connect (fuzzy match)
  hop connect <id>            Connect to exact ID
  hop open <target...>         Open multiple terminal tabs
  hop exec <target> "<cmd>"    Execute on multiple servers
  hop resolve <target>         Test which connections a target matches
  hop list                     List all connections`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return runDashboard()
		}
		return runQuickConnect(cmd, args)
	},
}

func Execute() error {
	// Execute the root command, but intercept "unknown command" errors
	// to treat them as quick-connect queries (e.g. "hop alex")
	rootCmd.InitDefaultHelpCmd()
	err := rootCmd.Execute()
	if err != nil {
		// Check if this was an unknown command error - if so, treat the
		// original os.Args as a quick-connect query
		args := os.Args[1:]
		if len(args) > 0 && !isKnownCommand(args[0]) {
			return runQuickConnect(rootCmd, args)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return nil
}

func isKnownCommand(name string) bool {
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == name || cmd.HasAlias(name) {
			return true
		}
	}
	return false
}

func runDashboard() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	selected, err := tui.Run(cfg, Version)
	if err != nil {
		return err
	}

	if selected == nil {
		return nil
	}

	if !quiet {
		fmt.Fprintf(os.Stderr, "Connecting to %s (%s)...\n", selected.ID, selected.Host)
	}

	return ssh.Connect(selected, &ssh.ConnectOptions{})
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default: ~/.config/hop/config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suppress non-essential output")
}
