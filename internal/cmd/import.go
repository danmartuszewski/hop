package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/danmartuszewski/hop/internal/config"
	"github.com/danmartuszewski/hop/internal/sshconfig"
	"github.com/spf13/cobra"
)

var (
	importFile   string
	importDryRun bool
	importYes    bool
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import connections from SSH config",
	Long: `Import SSH connections from ~/.ssh/config into hop.

By default, imports from ~/.ssh/config. Use --file to specify a different path.

Wildcard host patterns (*, ?) are automatically skipped.
Existing connections with the same ID are renamed with -imported suffix.`,
	RunE: runImport,
}

func init() {
	rootCmd.AddCommand(importCmd)

	importCmd.Flags().StringVarP(&importFile, "file", "f", "", "SSH config file path (default: ~/.ssh/config)")
	importCmd.Flags().BoolVar(&importDryRun, "dry-run", false, "Preview imports without saving")
	importCmd.Flags().BoolVarP(&importYes, "yes", "y", false, "Skip confirmation prompt")
}

func runImport(cmd *cobra.Command, args []string) error {
	// Parse SSH config
	hosts, err := sshconfig.Parse(importFile)
	if err != nil {
		return fmt.Errorf("failed to parse SSH config: %w", err)
	}

	if len(hosts) == 0 {
		fmt.Println("No importable connections found in SSH config.")
		fmt.Println("(Wildcard patterns like Host * are automatically skipped)")
		return nil
	}

	// Load current hop config
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	// Build a set of existing IDs
	existingIDs := make(map[string]bool)
	for _, conn := range cfg.Connections {
		existingIDs[conn.ID] = true
	}

	// Convert to connections and handle conflicts
	type importItem struct {
		original   string
		connection config.Connection
		renamed    bool
	}

	var imports []importItem
	for _, host := range hosts {
		conn := host.ToConnection()
		originalID := conn.ID

		// Handle ID conflicts
		if existingIDs[conn.ID] {
			conn.ID = sshconfig.ResolveConflict(conn.ID, existingIDs)
		}

		// Mark the new ID as taken
		existingIDs[conn.ID] = true

		imports = append(imports, importItem{
			original:   originalID,
			connection: conn,
			renamed:    originalID != conn.ID,
		})
	}

	// Display preview
	fmt.Printf("Found %d connection(s) to import:\n\n", len(imports))

	for _, item := range imports {
		conn := item.connection
		if item.renamed {
			fmt.Printf("  %s -> %s (renamed, original ID exists)\n", item.original, conn.ID)
		} else {
			fmt.Printf("  %s\n", conn.ID)
		}
		fmt.Printf("    Host: %s", conn.Host)
		if conn.User != "" {
			fmt.Printf(", User: %s", conn.User)
		}
		if conn.Port != 0 && conn.Port != 22 {
			fmt.Printf(", Port: %d", conn.Port)
		}
		if conn.ProxyJump != "" {
			fmt.Printf(", ProxyJump: %s", conn.ProxyJump)
		}
		if conn.ForwardAgent {
			fmt.Printf(", ForwardAgent: yes")
		}
		fmt.Println()
	}
	fmt.Println()

	if importDryRun {
		fmt.Println("(dry-run: no changes made)")
		return nil
	}

	// Confirm unless --yes flag is set
	if !importYes {
		fmt.Print("Import these connections? [y/N] ")
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("Import cancelled.")
			return nil
		}
	}

	// Add connections
	for _, item := range imports {
		cfg.AddConnection(item.connection)
	}

	// Save config
	if err := cfg.Save(cfgFile); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Successfully imported %d connection(s).\n", len(imports))
	return nil
}
