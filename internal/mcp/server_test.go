package hopmcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/danmartuszewski/hop/internal/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gopkg.in/yaml.v3"
)

func writeConfig(t *testing.T, cfg *config.Config) string {
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

func TestNewHopServer_ReadOnly(t *testing.T) {
	cfg := &config.Config{Version: 1, Connections: []config.Connection{}, Groups: map[string][]string{}}
	path := writeConfig(t, cfg)

	server := NewHopServer("test", path, false)
	if server == nil {
		t.Fatal("expected non-nil server")
	}
}

func TestNewHopServer_WithExec(t *testing.T) {
	cfg := &config.Config{Version: 1, Connections: []config.Connection{}, Groups: map[string][]string{}}
	path := writeConfig(t, cfg)

	server := NewHopServer("test", path, true)
	if server == nil {
		t.Fatal("expected non-nil server")
	}
}

func TestIntegration_InitializeAndListTools(t *testing.T) {
	cfg := &config.Config{
		Version: 1,
		Connections: []config.Connection{
			{ID: "test-server", Host: "test.example.com", Port: 22, User: "testuser"},
		},
		Groups: map[string][]string{},
	}
	path := writeConfig(t, cfg)

	ctx := context.Background()

	// Test without exec
	t.Run("without exec", func(t *testing.T) {
		server := NewHopServer("test", path, false)
		clientTransport, serverTransport := mcp.NewInMemoryTransports()

		serverSession, err := server.Connect(ctx, serverTransport, nil)
		if err != nil {
			t.Fatalf("server connect: %v", err)
		}

		client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0"}, nil)
		clientSession, err := client.Connect(ctx, clientTransport, nil)
		if err != nil {
			t.Fatalf("client connect: %v", err)
		}

		// List tools
		toolsResult, err := clientSession.ListTools(ctx, nil)
		if err != nil {
			t.Fatalf("list tools: %v", err)
		}

		toolNames := map[string]bool{}
		for _, tool := range toolsResult.Tools {
			toolNames[tool.Name] = true
		}

		// Should have read-only tools
		expectedTools := []string{"list_connections", "search_connections", "get_connection", "resolve_target", "list_groups", "get_history", "build_ssh_command"}
		for _, name := range expectedTools {
			if !toolNames[name] {
				t.Errorf("missing tool: %s", name)
			}
		}

		// Should NOT have exec_command
		if toolNames["exec_command"] {
			t.Error("exec_command should not be available without --allow-exec")
		}

		clientSession.Close()
		serverSession.Wait()
	})

	// Test with exec
	t.Run("with exec", func(t *testing.T) {
		server := NewHopServer("test", path, true)
		clientTransport, serverTransport := mcp.NewInMemoryTransports()

		serverSession, err := server.Connect(ctx, serverTransport, nil)
		if err != nil {
			t.Fatalf("server connect: %v", err)
		}

		client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0"}, nil)
		clientSession, err := client.Connect(ctx, clientTransport, nil)
		if err != nil {
			t.Fatalf("client connect: %v", err)
		}

		// List tools
		toolsResult, err := clientSession.ListTools(ctx, nil)
		if err != nil {
			t.Fatalf("list tools: %v", err)
		}

		toolNames := map[string]bool{}
		for _, tool := range toolsResult.Tools {
			toolNames[tool.Name] = true
		}

		// Should have exec_command
		if !toolNames["exec_command"] {
			t.Error("exec_command should be available with --allow-exec")
		}

		clientSession.Close()
		serverSession.Wait()
	})
}

func TestIntegration_CallTool(t *testing.T) {
	cfg := &config.Config{
		Version: 1,
		Connections: []config.Connection{
			{ID: "test-server", Host: "test.example.com", Port: 22, User: "testuser", Project: "myproject", Env: "prod"},
		},
		Groups: map[string][]string{},
	}
	path := writeConfig(t, cfg)

	ctx := context.Background()
	server := NewHopServer("test", path, false)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer func() {
		clientSession.Close()
		serverSession.Wait()
	}()

	// Call list_connections
	result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name:      "list_connections",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("call tool: %v", err)
	}
	if result.IsError {
		t.Fatal("unexpected tool error")
	}

	text := result.Content[0].(*mcp.TextContent).Text
	var connections []ConnectionInfo
	if err := json.Unmarshal([]byte(text), &connections); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(connections) != 1 {
		t.Errorf("expected 1 connection, got %d", len(connections))
	}
	if connections[0].ID != "test-server" {
		t.Errorf("expected ID test-server, got %s", connections[0].ID)
	}

	// Call get_connection
	result, err = clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name:      "get_connection",
		Arguments: map[string]any{"id": "test-server"},
	})
	if err != nil {
		t.Fatalf("call tool: %v", err)
	}
	if result.IsError {
		t.Fatal("unexpected tool error")
	}

	// Call get_connection with nonexistent ID
	result, err = clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name:      "get_connection",
		Arguments: map[string]any{"id": "nonexistent"},
	})
	if err != nil {
		t.Fatalf("call tool: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError for nonexistent connection")
	}

	// Call build_ssh_command
	result, err = clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name:      "build_ssh_command",
		Arguments: map[string]any{"id": "test-server", "command": "uptime"},
	})
	if err != nil {
		t.Fatalf("call tool: %v", err)
	}
	if result.IsError {
		t.Fatal("unexpected tool error")
	}
	text = result.Content[0].(*mcp.TextContent).Text
	if text == "" {
		t.Error("expected non-empty SSH command")
	}
}

func TestIntegration_ReadResource(t *testing.T) {
	cfg := &config.Config{
		Version: 1,
		Connections: []config.Connection{
			{ID: "test-server", Host: "test.example.com", Port: 22, User: "testuser"},
		},
		Groups: map[string][]string{"all": {"test-server"}},
	}
	path := writeConfig(t, cfg)

	ctx := context.Background()
	server := NewHopServer("test", path, false)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer func() {
		clientSession.Close()
		serverSession.Wait()
	}()

	// Read config resource
	result, err := clientSession.ReadResource(ctx, &mcp.ReadResourceParams{URI: "hop://config"})
	if err != nil {
		t.Fatalf("read resource: %v", err)
	}
	if len(result.Contents) == 0 {
		t.Fatal("expected resource contents")
	}

	type configSummary struct {
		ConnectionCount int `json:"connection_count"`
		GroupCount       int `json:"group_count"`
	}
	var summary configSummary
	if err := json.Unmarshal([]byte(result.Contents[0].Text), &summary); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if summary.ConnectionCount != 1 {
		t.Errorf("expected 1 connection, got %d", summary.ConnectionCount)
	}
	if summary.GroupCount != 1 {
		t.Errorf("expected 1 group, got %d", summary.GroupCount)
	}
}

func TestIntegration_ContextCancellation(t *testing.T) {
	cfg := &config.Config{Version: 1, Connections: []config.Connection{}, Groups: map[string][]string{}}
	path := writeConfig(t, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	server := NewHopServer("test", path, false)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}

	// Cancel context
	cancel()
	clientSession.Close()
	serverSession.Wait()
}
