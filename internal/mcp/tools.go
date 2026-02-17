package hopmcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/danmartuszewski/hop/internal/config"
	"github.com/danmartuszewski/hop/internal/fuzzy"
	"github.com/danmartuszewski/hop/internal/resolve"
	"github.com/danmartuszewski/hop/internal/ssh"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// configLoader provides config access to tool handlers.
type configLoader struct {
	cfgPath string
}

func (cl *configLoader) load() (*config.Config, error) {
	return config.Load(cl.cfgPath)
}

// toConnectionInfo converts a config.Connection to a ConnectionInfo, omitting IdentityFile.
func toConnectionInfo(c config.Connection) ConnectionInfo {
	return ConnectionInfo{
		ID:           c.ID,
		Host:         c.Host,
		User:         c.User,
		Port:         c.Port,
		Project:      c.Project,
		Env:          c.Env,
		ProxyJump:    c.ProxyJump,
		ForwardAgent: c.ForwardAgent,
		Tags:         c.Tags,
		Options:      c.Options,
	}
}

func textResult(text string) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}, nil, nil
}

func jsonTextResult(v any) (*mcp.CallToolResult, any, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("marshal result: %w", err)
	}
	return textResult(string(data))
}

func errorResult(msg string) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		IsError: true,
	}, nil, nil
}

// --- Tool handlers ---

func (cl *configLoader) handleListConnections(_ context.Context, _ *mcp.CallToolRequest, input ListConnectionsInput) (*mcp.CallToolResult, any, error) {
	cfg, err := cl.load()
	if err != nil {
		return errorResult(fmt.Sprintf("failed to load config: %v", err))
	}

	connections := cfg.Connections

	// Apply filters
	if input.Project != "" {
		var filtered []config.Connection
		for _, c := range connections {
			if strings.EqualFold(c.Project, input.Project) {
				filtered = append(filtered, c)
			}
		}
		connections = filtered
	}
	if input.Env != "" {
		var filtered []config.Connection
		for _, c := range connections {
			if strings.EqualFold(c.Env, input.Env) {
				filtered = append(filtered, c)
			}
		}
		connections = filtered
	}
	if input.Tag != "" {
		connections = fuzzy.MatchByTag(input.Tag, connections)
	}

	infos := make([]ConnectionInfo, len(connections))
	for i, c := range connections {
		infos[i] = toConnectionInfo(c)
	}
	return jsonTextResult(infos)
}

func (cl *configLoader) handleSearchConnections(_ context.Context, _ *mcp.CallToolRequest, input SearchConnectionsInput) (*mcp.CallToolResult, any, error) {
	cfg, err := cl.load()
	if err != nil {
		return errorResult(fmt.Sprintf("failed to load config: %v", err))
	}

	matches := fuzzy.FindMatches(input.Query, cfg.Connections)
	if len(matches) == 0 {
		return textResult(fmt.Sprintf("No connections matching '%s'.", input.Query))
	}

	type searchResult struct {
		ConnectionInfo
		Score int `json:"score"`
	}

	results := make([]searchResult, len(matches))
	for i, m := range matches {
		results[i] = searchResult{
			ConnectionInfo: toConnectionInfo(*m.Connection),
			Score:          m.Score,
		}
	}
	return jsonTextResult(results)
}

func (cl *configLoader) handleGetConnection(_ context.Context, _ *mcp.CallToolRequest, input GetConnectionInput) (*mcp.CallToolResult, any, error) {
	cfg, err := cl.load()
	if err != nil {
		return errorResult(fmt.Sprintf("failed to load config: %v", err))
	}

	conn := cfg.FindConnection(input.ID)
	if conn == nil {
		return errorResult(fmt.Sprintf("Connection '%s' not found.", input.ID))
	}

	return jsonTextResult(toConnectionInfo(*conn))
}

func (cl *configLoader) handleResolveTarget(_ context.Context, _ *mcp.CallToolRequest, input ResolveTargetInput) (*mcp.CallToolResult, any, error) {
	cfg, err := cl.load()
	if err != nil {
		return errorResult(fmt.Sprintf("failed to load config: %v", err))
	}

	result, err := resolve.ResolveTarget(input.Target, cfg)
	if err != nil {
		return errorResult(fmt.Sprintf("resolve error: %v", err))
	}

	connections := result.Connections
	if input.Tag != "" {
		connections = fuzzy.MatchByTag(input.Tag, connections)
	}

	type resolveOutput struct {
		Method      string           `json:"method"`
		Connections []ConnectionInfo `json:"connections"`
	}

	infos := make([]ConnectionInfo, len(connections))
	for i, c := range connections {
		infos[i] = toConnectionInfo(c)
	}

	return jsonTextResult(resolveOutput{
		Method:      result.Method.String(),
		Connections: infos,
	})
}

func (cl *configLoader) handleListGroups(_ context.Context, _ *mcp.CallToolRequest, _ ListGroupsInput) (*mcp.CallToolResult, any, error) {
	cfg, err := cl.load()
	if err != nil {
		return errorResult(fmt.Sprintf("failed to load config: %v", err))
	}

	type groupInfo struct {
		Name    string   `json:"name"`
		Members []string `json:"members"`
	}

	groups := make([]groupInfo, 0, len(cfg.Groups))
	for name, members := range cfg.Groups {
		groups = append(groups, groupInfo{Name: name, Members: members})
	}
	return jsonTextResult(groups)
}

func (cl *configLoader) handleGetHistory(_ context.Context, _ *mcp.CallToolRequest, input GetHistoryInput) (*mcp.CallToolResult, any, error) {
	history, err := config.LoadHistory()
	if err != nil {
		return errorResult(fmt.Sprintf("failed to load history: %v", err))
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 10
	}

	type historyItem struct {
		ID       string    `json:"id"`
		LastUsed time.Time `json:"last_used"`
		UseCount int       `json:"use_count"`
	}

	var ids []string
	switch input.SortBy {
	case "frequent":
		ids = history.GetMostUsed(limit)
	default:
		ids = history.GetRecent(limit)
	}

	items := make([]historyItem, 0, len(ids))
	for _, id := range ids {
		lastUsed, _ := history.GetLastUsed(id)
		items = append(items, historyItem{
			ID:       id,
			LastUsed: lastUsed,
			UseCount: history.GetUseCount(id),
		})
	}
	return jsonTextResult(items)
}

func (cl *configLoader) handleBuildSSHCommand(_ context.Context, _ *mcp.CallToolRequest, input BuildSSHCommandInput) (*mcp.CallToolResult, any, error) {
	cfg, err := cl.load()
	if err != nil {
		return errorResult(fmt.Sprintf("failed to load config: %v", err))
	}

	conn := cfg.FindConnection(input.ID)
	if conn == nil {
		return errorResult(fmt.Sprintf("Connection '%s' not found.", input.ID))
	}

	// Strip identity file to avoid leaking local SSH key paths
	sanitized := *conn
	sanitized.IdentityFile = ""

	opts := &ssh.ConnectOptions{
		Command:  input.Command,
		ForceTTY: input.ForceTTY,
	}

	cmdStr := ssh.BuildCommandString(&sanitized, opts)
	return textResult(cmdStr)
}

func (cl *configLoader) handleExecCommand(ctx context.Context, _ *mcp.CallToolRequest, input ExecCommandInput) (*mcp.CallToolResult, any, error) {
	if input.Command == "" {
		return errorResult("command is required")
	}
	if input.Target == "" {
		return errorResult("target is required")
	}

	cfg, err := cl.load()
	if err != nil {
		return errorResult(fmt.Sprintf("failed to load config: %v", err))
	}

	result, err := resolve.ResolveTarget(input.Target, cfg)
	if err != nil {
		return errorResult(fmt.Sprintf("resolve error: %v", err))
	}

	connections := result.Connections
	if input.Tag != "" {
		connections = fuzzy.MatchByTag(input.Tag, connections)
	}

	if len(connections) == 0 {
		return errorResult(fmt.Sprintf("No connections matching '%s'.", input.Target))
	}

	// Enforce host cap
	truncatedHosts := false
	if len(connections) > MaxHostsPerRequest {
		connections = connections[:MaxHostsPerRequest]
		truncatedHosts = true
	}

	// Parse timeout
	var timeout time.Duration
	if input.Timeout != "" {
		timeout, err = time.ParseDuration(input.Timeout)
		if err != nil {
			return errorResult(fmt.Sprintf("invalid timeout '%s': %v", input.Timeout, err))
		}
	}

	parallel := input.Parallel
	if parallel <= 0 {
		parallel = 10
	}

	log.Printf("[exec] target=%s command=%q hosts=%d", input.Target, input.Command, len(connections))

	execOpts := &ssh.ExecOptions{
		Command:  input.Command,
		Parallel: parallel,
		Timeout:  timeout,
		Stream:   false,
	}

	results := ssh.ExecuteContext(ctx, connections, execOpts)

	// Build structured response
	type hostResult struct {
		ID       string `json:"id"`
		Stdout   string `json:"stdout,omitempty"`
		Stderr   string `json:"stderr,omitempty"`
		ExitCode int    `json:"exit_code"`
		Duration string `json:"duration"`
		Error    string `json:"error,omitempty"`
	}

	hostResults := make([]hostResult, len(results))
	for i, r := range results {
		stdout := r.Stdout
		stderr := r.Stderr

		// Enforce output cap
		if len(stdout) > MaxBytesPerHost {
			stdout = stdout[:MaxBytesPerHost] + "\n[truncated]"
		}
		if len(stderr) > MaxBytesPerHost {
			stderr = stderr[:MaxBytesPerHost] + "\n[truncated]"
		}

		hr := hostResult{
			ID:       r.Connection.ID,
			Stdout:   stdout,
			Stderr:   stderr,
			ExitCode: r.ExitCode,
			Duration: r.Duration.Round(time.Millisecond).String(),
		}
		if r.Error != nil {
			hr.Error = r.Error.Error()
		}
		hostResults[i] = hr
	}

	type execResponse struct {
		Results        []hostResult `json:"results"`
		TotalHosts     int          `json:"total_hosts"`
		TruncatedHosts bool         `json:"truncated_hosts,omitempty"`
	}

	return jsonTextResult(execResponse{
		Results:        hostResults,
		TotalHosts:     len(hostResults),
		TruncatedHosts: truncatedHosts,
	})
}

// --- Resource handlers ---

func (cl *configLoader) handleConfigResource(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	cfg, err := cl.load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	projects := map[string]bool{}
	envs := map[string]bool{}
	for _, c := range cfg.Connections {
		if c.Project != "" {
			projects[c.Project] = true
		}
		if c.Env != "" {
			envs[c.Env] = true
		}
	}

	projectList := make([]string, 0, len(projects))
	for p := range projects {
		projectList = append(projectList, p)
	}
	envList := make([]string, 0, len(envs))
	for e := range envs {
		envList = append(envList, e)
	}

	type configSummary struct {
		Version         int      `json:"version"`
		ConnectionCount int      `json:"connection_count"`
		GroupCount       int      `json:"group_count"`
		Projects        []string `json:"projects"`
		Environments    []string `json:"environments"`
	}

	data, _ := json.MarshalIndent(configSummary{
		Version:         cfg.Version,
		ConnectionCount: len(cfg.Connections),
		GroupCount:       len(cfg.Groups),
		Projects:        projectList,
		Environments:    envList,
	}, "", "  ")

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      req.Params.URI,
			MIMEType: "application/json",
			Text:     string(data),
		}},
	}, nil
}

func (cl *configLoader) handleConnectionsResource(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	cfg, err := cl.load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	infos := make([]ConnectionInfo, len(cfg.Connections))
	for i, c := range cfg.Connections {
		infos[i] = toConnectionInfo(c)
	}

	data, _ := json.MarshalIndent(infos, "", "  ")

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      req.Params.URI,
			MIMEType: "application/json",
			Text:     string(data),
		}},
	}, nil
}

func (cl *configLoader) handleConnectionResource(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	cfg, err := cl.load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Extract ID from URI: hop://connections/{id}
	uri := req.Params.URI
	id := strings.TrimPrefix(uri, "hop://connections/")
	if id == "" || id == uri {
		return nil, fmt.Errorf("invalid connection URI: %s", uri)
	}

	conn := cfg.FindConnection(id)
	if conn == nil {
		return nil, fmt.Errorf("connection '%s' not found", id)
	}

	data, _ := json.MarshalIndent(toConnectionInfo(*conn), "", "  ")

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      req.Params.URI,
			MIMEType: "application/json",
			Text:     string(data),
		}},
	}, nil
}

func (cl *configLoader) handleGroupsResource(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	cfg, err := cl.load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	type groupInfo struct {
		Name    string   `json:"name"`
		Members []string `json:"members"`
	}

	groups := make([]groupInfo, 0, len(cfg.Groups))
	for name, members := range cfg.Groups {
		groups = append(groups, groupInfo{Name: name, Members: members})
	}

	data, _ := json.MarshalIndent(groups, "", "  ")

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      req.Params.URI,
			MIMEType: "application/json",
			Text:     string(data),
		}},
	}, nil
}
