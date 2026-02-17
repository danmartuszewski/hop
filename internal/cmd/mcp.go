package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	hopmcp "github.com/danmartuszewski/hop/internal/mcp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

var (
	mcpAllowExec       bool
	mcpPrintClientConf bool
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP server for AI assistant integration",
	Long: `Start a Model Context Protocol (MCP) server over stdio.

This allows AI assistants like Claude Desktop and Claude Code to discover
and use hop's SSH connection management capabilities.

By default, only read-only tools are available. Use --allow-exec to enable
remote command execution.

Setup:
  claude mcp add hop -- hop mcp
  claude mcp add hop -- hop mcp --allow-exec`,
	RunE: runMCP,
}

func init() {
	rootCmd.AddCommand(mcpCmd)

	mcpCmd.Flags().BoolVar(&mcpAllowExec, "allow-exec", false, "enable the exec_command tool for remote command execution")
	mcpCmd.Flags().BoolVar(&mcpPrintClientConf, "print-client-config", false, "print client configuration JSON and exit")
}

func runMCP(cmd *cobra.Command, args []string) error {
	if mcpPrintClientConf {
		return printClientConfig()
	}

	// All logging goes to stderr to keep stdout clean for JSON-RPC
	log.SetOutput(os.Stderr)

	server := hopmcp.NewHopServer(Version, cfgFile, mcpAllowExec)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.Printf("hop MCP server starting (version=%s, allow-exec=%v)", Version, mcpAllowExec)

	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		return fmt.Errorf("MCP server error: %w", err)
	}

	return nil
}

func printClientConfig() error {
	execPath, err := os.Executable()
	if err != nil {
		execPath = "hop"
	}

	mcpArgs := []string{"mcp"}
	if mcpAllowExec {
		mcpArgs = append(mcpArgs, "--allow-exec")
	}

	type mcpServerConfig struct {
		Command string   `json:"command"`
		Args    []string `json:"args"`
	}

	type clientConfig struct {
		MCPServers map[string]mcpServerConfig `json:"mcpServers"`
	}

	configFile := "claude_desktop_config.json"
	configDir := ""
	switch runtime.GOOS {
	case "darwin":
		home, _ := os.UserHomeDir()
		configDir = home + "/Library/Application Support/Claude"
	case "windows":
		configDir = os.Getenv("APPDATA") + "\\Claude"
	default:
		home, _ := os.UserHomeDir()
		configDir = home + "/.config/claude"
	}

	cfg := clientConfig{
		MCPServers: map[string]mcpServerConfig{
			"hop": {
				Command: execPath,
				Args:    mcpArgs,
			},
		},
	}

	data, _ := json.MarshalIndent(cfg, "", "  ")

	fmt.Fprintf(os.Stderr, "Add to %s/%s:\n\n", configDir, configFile)
	fmt.Println(string(data))
	fmt.Fprintf(os.Stderr, "\nOr for Claude Code:\n  claude mcp add hop -- %s %s\n", execPath, joinArgs(mcpArgs))

	return nil
}

func joinArgs(args []string) string {
	result := ""
	for i, a := range args {
		if i > 0 {
			result += " "
		}
		result += a
	}
	return result
}
