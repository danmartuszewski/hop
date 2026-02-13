package cmd

import (
	"github.com/spf13/cobra"
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Open TUI dashboard",
	Long:  "Open the interactive TUI dashboard for browsing and connecting to servers. The dashboard persists after SSH sessions end, letting you hop between servers.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDashboard()
	},
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
}
