package cmd

import (
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
  hop open <group|ids...>     Open multiple terminal tabs
  hop exec <group> "<cmd>"    Execute on multiple servers
  hop list                    List all connections
  hop config                  Manage configuration`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			// TODO: Open TUI dashboard
			return cmd.Help()
		}
		return runQuickConnect(cmd, args)
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default: ~/.config/hop/config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suppress non-essential output")
}
