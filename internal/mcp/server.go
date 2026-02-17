package hopmcp

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// NewHopServer creates and configures an MCP server with all hop tools and resources.
func NewHopServer(version, cfgPath string, allowExec bool) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "hop",
		Version: version,
	}, &mcp.ServerOptions{
		Instructions: "hop is an SSH connection manager. Use these tools to discover, search, and manage SSH connections.",
	})

	loader := &configLoader{cfgPath: cfgPath}

	// Read-only tools (always registered)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_connections",
		Description: "List SSH connections, optionally filtered by project, environment, or tag.",
	}, loader.handleListConnections)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_connections",
		Description: "Fuzzy search across all connections by ID, host, tags, project, or environment.",
	}, loader.handleSearchConnections)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_connection",
		Description: "Get detailed information about a specific connection by its ID.",
	}, loader.handleGetConnection)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "resolve_target",
		Description: "Resolve a target pattern to matching connections. Shows how a target would be resolved (named group, project-env, glob, or fuzzy match).",
	}, loader.handleResolveTarget)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_groups",
		Description: "List all named connection groups defined in the config.",
	}, loader.handleListGroups)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_history",
		Description: "Get connection usage history, sorted by recent use or frequency.",
	}, loader.handleGetHistory)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "build_ssh_command",
		Description: "Build the full SSH command string for a connection, useful for debugging or manual use.",
	}, loader.handleBuildSSHCommand)

	// Exec tool (gated by --allow-exec)
	if allowExec {
		mcp.AddTool(server, &mcp.Tool{
			Name:        "exec_command",
			Description: "Execute a shell command on one or more remote servers matched by a target pattern. Requires --allow-exec flag.",
		}, loader.handleExecCommand)
	}

	// Resources
	registerResources(server, loader)

	return server
}

func registerResources(server *mcp.Server, loader *configLoader) {
	server.AddResource(&mcp.Resource{
		Name:        "config",
		URI:         "hop://config",
		Description: "Hop configuration summary including connection count, group count, and available projects/environments.",
		MIMEType:    "application/json",
	}, loader.handleConfigResource)

	server.AddResource(&mcp.Resource{
		Name:        "connections",
		URI:         "hop://connections",
		Description: "List of all configured SSH connections.",
		MIMEType:    "application/json",
	}, loader.handleConnectionsResource)

	server.AddResourceTemplate(&mcp.ResourceTemplate{
		Name:        "connection",
		URITemplate: "hop://connections/{id}",
		Description: "Details of a specific SSH connection by ID.",
		MIMEType:    "application/json",
	}, loader.handleConnectionResource)

	server.AddResource(&mcp.Resource{
		Name:        "groups",
		URI:         "hop://groups",
		Description: "All named connection groups and their members.",
		MIMEType:    "application/json",
	}, loader.handleGroupsResource)
}
