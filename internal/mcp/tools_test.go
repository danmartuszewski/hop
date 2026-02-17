package hopmcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gopkg.in/yaml.v3"

	"github.com/danmartuszewski/hop/internal/config"
)

// writeTestConfig creates a temporary config file and returns the path.
func writeTestConfig(t *testing.T, cfg *config.Config) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	data, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}

func fullTestConfig() *config.Config {
	return &config.Config{
		Version: 1,
		Defaults: config.Defaults{
			User: "deploy",
			Port: 22,
		},
		Connections: []config.Connection{
			{ID: "web-prod-1", Host: "web1.prod.example.com", Port: 22, User: "deploy", Project: "myapp", Env: "prod", Tags: []string{"web", "production"}, IdentityFile: "~/.ssh/id_rsa"},
			{ID: "web-prod-2", Host: "web2.prod.example.com", Port: 22, User: "deploy", Project: "myapp", Env: "prod", Tags: []string{"web", "production"}},
			{ID: "db-prod", Host: "db.prod.example.com", Port: 5432, User: "deploy", Project: "myapp", Env: "prod", Tags: []string{"database", "production"}},
			{ID: "web-staging", Host: "web.staging.example.com", Port: 22, User: "deploy", Project: "myapp", Env: "staging", Tags: []string{"web", "staging"}},
			{ID: "api-prod", Host: "api.prod.example.com", Port: 22, User: "deploy", Project: "api", Env: "prod", Tags: []string{"api"},
				ProxyJump: "bastion", ForwardAgent: true, Options: map[string]string{"StrictHostKeyChecking": "no"}},
		},
		Groups: map[string][]string{
			"production": {"web-prod-1", "web-prod-2", "db-prod"},
			"web-tier":   {"web-prod-1", "web-prod-2", "web-staging"},
		},
	}
}

func emptyTestConfig() *config.Config {
	return &config.Config{
		Version:     1,
		Connections: []config.Connection{},
		Groups:      map[string][]string{},
	}
}

// --- Tool handler tests ---

func TestListConnections_AllConnections(t *testing.T) {
	cfg := fullTestConfig()
	path := writeTestConfig(t, cfg)
	loader := &configLoader{cfgPath: path}

	result, _, err := loader.handleListConnections(context.Background(), nil, ListConnectionsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected tool error: %s", result.Content[0].(*mcp.TextContent).Text)
	}

	var connections []ConnectionInfo
	text := result.Content[0].(*mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &connections); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if len(connections) != 5 {
		t.Errorf("expected 5 connections, got %d", len(connections))
	}
}

func TestListConnections_EmptyConfig(t *testing.T) {
	cfg := emptyTestConfig()
	path := writeTestConfig(t, cfg)
	loader := &configLoader{cfgPath: path}

	result, _, err := loader.handleListConnections(context.Background(), nil, ListConnectionsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var connections []ConnectionInfo
	text := result.Content[0].(*mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &connections); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if len(connections) != 0 {
		t.Errorf("expected 0 connections, got %d", len(connections))
	}
}

func TestListConnections_FilterByProject(t *testing.T) {
	cfg := fullTestConfig()
	path := writeTestConfig(t, cfg)
	loader := &configLoader{cfgPath: path}

	result, _, err := loader.handleListConnections(context.Background(), nil, ListConnectionsInput{Project: "myapp"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var connections []ConnectionInfo
	text := result.Content[0].(*mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &connections); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if len(connections) != 4 {
		t.Errorf("expected 4 connections for project myapp, got %d", len(connections))
	}
}

func TestListConnections_FilterByEnv(t *testing.T) {
	cfg := fullTestConfig()
	path := writeTestConfig(t, cfg)
	loader := &configLoader{cfgPath: path}

	result, _, err := loader.handleListConnections(context.Background(), nil, ListConnectionsInput{Env: "staging"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var connections []ConnectionInfo
	text := result.Content[0].(*mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &connections); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if len(connections) != 1 {
		t.Errorf("expected 1 connection for env staging, got %d", len(connections))
	}
}

func TestListConnections_FilterByTag(t *testing.T) {
	cfg := fullTestConfig()
	path := writeTestConfig(t, cfg)
	loader := &configLoader{cfgPath: path}

	result, _, err := loader.handleListConnections(context.Background(), nil, ListConnectionsInput{Tag: "database"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var connections []ConnectionInfo
	text := result.Content[0].(*mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &connections); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if len(connections) != 1 {
		t.Errorf("expected 1 connection with tag database, got %d", len(connections))
	}
}

func TestListConnections_FilterCombined(t *testing.T) {
	cfg := fullTestConfig()
	path := writeTestConfig(t, cfg)
	loader := &configLoader{cfgPath: path}

	result, _, err := loader.handleListConnections(context.Background(), nil, ListConnectionsInput{
		Project: "myapp",
		Env:     "prod",
		Tag:     "web",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var connections []ConnectionInfo
	text := result.Content[0].(*mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &connections); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if len(connections) != 2 {
		t.Errorf("expected 2 connections (web+prod+myapp), got %d", len(connections))
	}
}

func TestSearchConnections_WithResults(t *testing.T) {
	cfg := fullTestConfig()
	path := writeTestConfig(t, cfg)
	loader := &configLoader{cfgPath: path}

	result, _, err := loader.handleSearchConnections(context.Background(), nil, SearchConnectionsInput{Query: "web"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected tool error")
	}

	text := result.Content[0].(*mcp.TextContent).Text
	var results []json.RawMessage
	if err := json.Unmarshal([]byte(text), &results); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected search results")
	}
}

func TestSearchConnections_NoResults(t *testing.T) {
	cfg := fullTestConfig()
	path := writeTestConfig(t, cfg)
	loader := &configLoader{cfgPath: path}

	result, _, err := loader.handleSearchConnections(context.Background(), nil, SearchConnectionsInput{Query: "zzzznonexistent"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := result.Content[0].(*mcp.TextContent).Text
	if text != "No connections matching 'zzzznonexistent'." {
		t.Errorf("unexpected message: %s", text)
	}
}

func TestGetConnection_Found(t *testing.T) {
	cfg := fullTestConfig()
	path := writeTestConfig(t, cfg)
	loader := &configLoader{cfgPath: path}

	result, _, err := loader.handleGetConnection(context.Background(), nil, GetConnectionInput{ID: "web-prod-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected tool error")
	}

	var conn ConnectionInfo
	text := result.Content[0].(*mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &conn); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if conn.ID != "web-prod-1" {
		t.Errorf("expected ID web-prod-1, got %s", conn.ID)
	}
}

func TestGetConnection_NotFound(t *testing.T) {
	cfg := fullTestConfig()
	path := writeTestConfig(t, cfg)
	loader := &configLoader{cfgPath: path}

	result, _, err := loader.handleGetConnection(context.Background(), nil, GetConnectionInput{ID: "nonexistent"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError for nonexistent connection")
	}
}

func TestListGroups(t *testing.T) {
	cfg := fullTestConfig()
	path := writeTestConfig(t, cfg)
	loader := &configLoader{cfgPath: path}

	result, _, err := loader.handleListGroups(context.Background(), nil, ListGroupsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	type groupInfo struct {
		Name    string   `json:"name"`
		Members []string `json:"members"`
	}
	var groups []groupInfo
	text := result.Content[0].(*mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &groups); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if len(groups) != 2 {
		t.Errorf("expected 2 groups, got %d", len(groups))
	}
}

func TestListGroups_Empty(t *testing.T) {
	cfg := emptyTestConfig()
	path := writeTestConfig(t, cfg)
	loader := &configLoader{cfgPath: path}

	result, _, err := loader.handleListGroups(context.Background(), nil, ListGroupsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	type groupInfo struct {
		Name    string   `json:"name"`
		Members []string `json:"members"`
	}
	var groups []groupInfo
	text := result.Content[0].(*mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &groups); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if len(groups) != 0 {
		t.Errorf("expected 0 groups, got %d", len(groups))
	}
}

func TestGetHistory_ReturnsValidJSON(t *testing.T) {
	result, _, err := (&configLoader{}).handleGetHistory(context.Background(), nil, GetHistoryInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("unexpected tool error")
	}

	var items []json.RawMessage
	text := result.Content[0].(*mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &items); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	// Just verify it returns valid JSON array (may or may not have entries depending on system state)
}

func TestBuildSSHCommand(t *testing.T) {
	cfg := fullTestConfig()
	path := writeTestConfig(t, cfg)
	loader := &configLoader{cfgPath: path}

	result, _, err := loader.handleBuildSSHCommand(context.Background(), nil, BuildSSHCommandInput{ID: "api-prod"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := result.Content[0].(*mcp.TextContent).Text
	if text == "" {
		t.Error("expected non-empty SSH command")
	}
	// Should contain proxy jump, forward agent, and options
	if !contains(text, "-J") || !contains(text, "-A") || !contains(text, "-o") {
		t.Errorf("expected SSH command with -J, -A, -o flags, got: %s", text)
	}
}

func TestBuildSSHCommand_NoIdentityFileLeak(t *testing.T) {
	cfg := fullTestConfig()
	path := writeTestConfig(t, cfg)
	loader := &configLoader{cfgPath: path}

	// web-prod-1 has IdentityFile set to ~/.ssh/id_rsa
	result, _, err := loader.handleBuildSSHCommand(context.Background(), nil, BuildSSHCommandInput{ID: "web-prod-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := result.Content[0].(*mcp.TextContent).Text
	if contains(text, "-i") || contains(text, "id_rsa") || contains(text, ".ssh") {
		t.Errorf("identity file leaked in build_ssh_command output: %s", text)
	}
}

func TestBuildSSHCommand_NotFound(t *testing.T) {
	cfg := fullTestConfig()
	path := writeTestConfig(t, cfg)
	loader := &configLoader{cfgPath: path}

	result, _, err := loader.handleBuildSSHCommand(context.Background(), nil, BuildSSHCommandInput{ID: "nonexistent"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError for nonexistent connection")
	}
}

func TestExecCommand_EmptyCommand(t *testing.T) {
	cfg := fullTestConfig()
	path := writeTestConfig(t, cfg)
	loader := &configLoader{cfgPath: path}

	result, _, err := loader.handleExecCommand(context.Background(), nil, ExecCommandInput{
		Target:  "production",
		Command: "",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError for empty command")
	}
}

func TestExecCommand_EmptyTarget(t *testing.T) {
	cfg := fullTestConfig()
	path := writeTestConfig(t, cfg)
	loader := &configLoader{cfgPath: path}

	result, _, err := loader.handleExecCommand(context.Background(), nil, ExecCommandInput{
		Target:  "",
		Command: "uptime",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError for empty target")
	}
}

func TestExecCommand_InvalidTimeout(t *testing.T) {
	cfg := fullTestConfig()
	path := writeTestConfig(t, cfg)
	loader := &configLoader{cfgPath: path}

	result, _, err := loader.handleExecCommand(context.Background(), nil, ExecCommandInput{
		Target:  "production",
		Command: "uptime",
		Timeout: "banana",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError for invalid timeout")
	}
}

func TestIdentityFileNeverExposed(t *testing.T) {
	cfg := fullTestConfig()
	path := writeTestConfig(t, cfg)
	loader := &configLoader{cfgPath: path}

	// Test list_connections
	result, _, _ := loader.handleListConnections(context.Background(), nil, ListConnectionsInput{})
	text := result.Content[0].(*mcp.TextContent).Text
	if contains(text, "identity_file") || contains(text, "id_rsa") {
		t.Error("identity file leaked in list_connections output")
	}

	// Test get_connection
	result, _, _ = loader.handleGetConnection(context.Background(), nil, GetConnectionInput{ID: "web-prod-1"})
	text = result.Content[0].(*mcp.TextContent).Text
	if contains(text, "identity_file") || contains(text, "id_rsa") {
		t.Error("identity file leaked in get_connection output")
	}
}

func TestResolveTarget(t *testing.T) {
	cfg := fullTestConfig()
	path := writeTestConfig(t, cfg)
	loader := &configLoader{cfgPath: path}

	result, _, err := loader.handleResolveTarget(context.Background(), nil, ResolveTargetInput{Target: "production"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected tool error")
	}

	type resolveOutput struct {
		Method      string           `json:"method"`
		Connections []ConnectionInfo `json:"connections"`
	}
	var output resolveOutput
	text := result.Content[0].(*mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &output); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if output.Method != "named group" {
		t.Errorf("expected 'named group', got %q", output.Method)
	}
	if len(output.Connections) != 3 {
		t.Errorf("expected 3 connections, got %d", len(output.Connections))
	}
}

// --- Resource handler tests ---

func TestConfigResource(t *testing.T) {
	cfg := fullTestConfig()
	path := writeTestConfig(t, cfg)
	loader := &configLoader{cfgPath: path}

	req := &mcp.ReadResourceRequest{Params: &mcp.ReadResourceParams{URI: "hop://config"}}

	result, err := loader.handleConfigResource(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	type configSummary struct {
		Version         int      `json:"version"`
		ConnectionCount int      `json:"connection_count"`
		GroupCount       int      `json:"group_count"`
		Projects        []string `json:"projects"`
		Environments    []string `json:"environments"`
	}
	var summary configSummary
	if err := json.Unmarshal([]byte(result.Contents[0].Text), &summary); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if summary.ConnectionCount != 5 {
		t.Errorf("expected 5 connections, got %d", summary.ConnectionCount)
	}
	if summary.GroupCount != 2 {
		t.Errorf("expected 2 groups, got %d", summary.GroupCount)
	}
}

func TestConfigResource_Empty(t *testing.T) {
	cfg := emptyTestConfig()
	path := writeTestConfig(t, cfg)
	loader := &configLoader{cfgPath: path}

	req := &mcp.ReadResourceRequest{Params: &mcp.ReadResourceParams{URI: "hop://config"}}

	result, err := loader.handleConfigResource(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	type configSummary struct {
		ConnectionCount int `json:"connection_count"`
		GroupCount       int `json:"group_count"`
	}
	var summary configSummary
	if err := json.Unmarshal([]byte(result.Contents[0].Text), &summary); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if summary.ConnectionCount != 0 {
		t.Errorf("expected 0 connections, got %d", summary.ConnectionCount)
	}
	if summary.GroupCount != 0 {
		t.Errorf("expected 0 groups, got %d", summary.GroupCount)
	}
}

func TestConnectionsResource(t *testing.T) {
	cfg := fullTestConfig()
	path := writeTestConfig(t, cfg)
	loader := &configLoader{cfgPath: path}

	req := &mcp.ReadResourceRequest{Params: &mcp.ReadResourceParams{URI: "hop://connections"}}

	result, err := loader.handleConnectionsResource(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var connections []ConnectionInfo
	if err := json.Unmarshal([]byte(result.Contents[0].Text), &connections); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(connections) != 5 {
		t.Errorf("expected 5 connections, got %d", len(connections))
	}
	// Verify no identity files
	if contains(result.Contents[0].Text, "identity_file") {
		t.Error("identity file leaked in connections resource")
	}
}

func TestConnectionResource_Found(t *testing.T) {
	cfg := fullTestConfig()
	path := writeTestConfig(t, cfg)
	loader := &configLoader{cfgPath: path}

	req := &mcp.ReadResourceRequest{Params: &mcp.ReadResourceParams{URI: "hop://connections/web-prod-1"}}

	result, err := loader.handleConnectionResource(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var conn ConnectionInfo
	if err := json.Unmarshal([]byte(result.Contents[0].Text), &conn); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if conn.ID != "web-prod-1" {
		t.Errorf("expected ID web-prod-1, got %s", conn.ID)
	}
}

func TestConnectionResource_NotFound(t *testing.T) {
	cfg := fullTestConfig()
	path := writeTestConfig(t, cfg)
	loader := &configLoader{cfgPath: path}

	req := &mcp.ReadResourceRequest{Params: &mcp.ReadResourceParams{URI: "hop://connections/nonexistent"}}

	_, err := loader.handleConnectionResource(context.Background(), req)
	if err == nil {
		t.Error("expected error for nonexistent connection")
	}
}

func TestGroupsResource(t *testing.T) {
	cfg := fullTestConfig()
	path := writeTestConfig(t, cfg)
	loader := &configLoader{cfgPath: path}

	req := &mcp.ReadResourceRequest{Params: &mcp.ReadResourceParams{URI: "hop://groups"}}

	result, err := loader.handleGroupsResource(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	type groupInfo struct {
		Name    string   `json:"name"`
		Members []string `json:"members"`
	}
	var groups []groupInfo
	if err := json.Unmarshal([]byte(result.Contents[0].Text), &groups); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(groups) != 2 {
		t.Errorf("expected 2 groups, got %d", len(groups))
	}
}

func TestGroupsResource_Empty(t *testing.T) {
	cfg := emptyTestConfig()
	path := writeTestConfig(t, cfg)
	loader := &configLoader{cfgPath: path}

	req := &mcp.ReadResourceRequest{Params: &mcp.ReadResourceParams{URI: "hop://groups"}}

	result, err := loader.handleGroupsResource(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	type groupInfo struct {
		Name    string   `json:"name"`
		Members []string `json:"members"`
	}
	var groups []groupInfo
	if err := json.Unmarshal([]byte(result.Contents[0].Text), &groups); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(groups) != 0 {
		t.Errorf("expected 0 groups, got %d", len(groups))
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
