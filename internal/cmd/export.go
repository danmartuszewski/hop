package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/danmartuszewski/hop/internal/config"
	"github.com/danmartuszewski/hop/internal/export"
	"github.com/spf13/cobra"
)

var (
	exportProject string
	exportEnv     string
	exportTag     string
	exportIDs     string
	exportOutput  string
	exportAll     bool
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export connections to YAML",
	Long: `Export connections to a YAML file that can be shared or re-imported.

At least one filter flag or --all is required to prevent accidental full dumps.
Filters combine with AND logic when multiple are specified.

Examples:
  hop export --all                          Export all connections to stdout
  hop export --all -o backup.yaml           Export all to a file
  hop export --project myapp -o myapp.yaml  Export by project
  hop export --tag database                 Export by tag
  hop export --env production               Export by environment
  hop export --id web-1,web-2               Export specific connections`,
	RunE: runExport,
}

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.Flags().StringVar(&exportProject, "project", "", "filter by project name")
	exportCmd.Flags().StringVar(&exportEnv, "env", "", "filter by environment")
	exportCmd.Flags().StringVar(&exportTag, "tag", "", "filter by tag")
	exportCmd.Flags().StringVar(&exportIDs, "id", "", "export specific connection IDs (comma-separated)")
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "output file path (default: stdout)")
	exportCmd.Flags().BoolVar(&exportAll, "all", false, "export all connections")
}

func runExport(cmd *cobra.Command, args []string) error {
	if !exportAll && exportProject == "" && exportEnv == "" && exportTag == "" && exportIDs == "" {
		return fmt.Errorf("at least one filter flag (--project, --env, --tag, --id) or --all is required")
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	var filtered []config.Connection

	if exportAll {
		filtered = cfg.Connections
	} else {
		// Build ID set if specified
		idSet := make(map[string]bool)
		if exportIDs != "" {
			for _, id := range strings.Split(exportIDs, ",") {
				id = strings.TrimSpace(id)
				if id != "" {
					idSet[id] = true
				}
			}
		}

		for _, conn := range cfg.Connections {
			if len(idSet) > 0 && !idSet[conn.ID] {
				continue
			}
			if exportProject != "" && !strings.EqualFold(conn.Project, exportProject) {
				continue
			}
			if exportEnv != "" && !strings.EqualFold(conn.Env, exportEnv) {
				continue
			}
			if exportTag != "" && !hasTag(conn, exportTag) {
				continue
			}
			filtered = append(filtered, conn)
		}
	}

	if len(filtered) == 0 {
		return fmt.Errorf("no connections match the given filters")
	}

	exportCfg := export.BuildExportConfig(filtered)

	if exportOutput == "" {
		return export.WriteYAML(os.Stdout, exportCfg)
	}

	f, err := os.Create(exportOutput)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	if err := export.WriteYAML(f, exportCfg); err != nil {
		return fmt.Errorf("failed to write export: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Exported %d connection(s) to %s\n", len(filtered), exportOutput)
	return nil
}

func hasTag(conn config.Connection, tag string) bool {
	for _, t := range conn.Tags {
		if strings.EqualFold(t, tag) {
			return true
		}
	}
	return false
}
