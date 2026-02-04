package sshconfig

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParse_BasicConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config")

	content := `
# Test SSH config
Host myserver
    HostName 192.168.1.100
    User admin
    Port 2222

Host webserver
    HostName web.example.com
    User deploy
    IdentityFile ~/.ssh/web_key
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	hosts, err := Parse(configPath)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(hosts) != 2 {
		t.Fatalf("expected 2 hosts, got %d", len(hosts))
	}

	// Check first host
	if hosts[0].Alias != "myserver" {
		t.Errorf("hosts[0].Alias = %q, want %q", hosts[0].Alias, "myserver")
	}
	if hosts[0].HostName != "192.168.1.100" {
		t.Errorf("hosts[0].HostName = %q, want %q", hosts[0].HostName, "192.168.1.100")
	}
	if hosts[0].User != "admin" {
		t.Errorf("hosts[0].User = %q, want %q", hosts[0].User, "admin")
	}
	if hosts[0].Port != 2222 {
		t.Errorf("hosts[0].Port = %d, want %d", hosts[0].Port, 2222)
	}

	// Check second host
	if hosts[1].Alias != "webserver" {
		t.Errorf("hosts[1].Alias = %q, want %q", hosts[1].Alias, "webserver")
	}
	if hosts[1].User != "deploy" {
		t.Errorf("hosts[1].User = %q, want %q", hosts[1].User, "deploy")
	}
}

func TestParse_WildcardFiltering(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config")

	content := `
Host *
    ServerAliveInterval 60

Host *.example.com
    User webadmin

Host myserver
    HostName 192.168.1.100
    User admin

Host dev?
    HostName dev.local
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	hosts, err := Parse(configPath)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Only myserver should be returned (wildcards filtered out)
	if len(hosts) != 1 {
		t.Fatalf("expected 1 host (wildcards filtered), got %d", len(hosts))
	}

	if hosts[0].Alias != "myserver" {
		t.Errorf("hosts[0].Alias = %q, want %q", hosts[0].Alias, "myserver")
	}
}

func TestParse_ProxyJumpAndForwardAgent(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config")

	content := `
Host bastion
    HostName bastion.example.com
    User jump

Host internal
    HostName 10.0.1.50
    User admin
    ProxyJump bastion
    ForwardAgent yes
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	hosts, err := Parse(configPath)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(hosts) != 2 {
		t.Fatalf("expected 2 hosts, got %d", len(hosts))
	}

	// Check internal host
	internal := hosts[1]
	if internal.ProxyJump != "bastion" {
		t.Errorf("ProxyJump = %q, want %q", internal.ProxyJump, "bastion")
	}
	if !internal.ForwardAgent {
		t.Error("ForwardAgent = false, want true")
	}
}

func TestParse_MissingHostName(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config")

	// When HostName is missing, the alias should be used as hostname
	content := `
Host myserver.example.com
    User admin
    Port 22
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	hosts, err := Parse(configPath)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(hosts) != 1 {
		t.Fatalf("expected 1 host, got %d", len(hosts))
	}

	conn := hosts[0].ToConnection()
	if conn.Host != "myserver.example.com" {
		t.Errorf("Host = %q, want %q", conn.Host, "myserver.example.com")
	}
}

func TestParse_Include(t *testing.T) {
	dir := t.TempDir()

	// Create included config
	includedConfig := `
Host included-server
    HostName included.example.com
    User included-user
`
	includedPath := filepath.Join(dir, "config.d", "included.conf")
	if err := os.MkdirAll(filepath.Dir(includedPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(includedPath, []byte(includedConfig), 0644); err != nil {
		t.Fatal(err)
	}

	// Create main config with Include
	mainConfig := `
Include config.d/*.conf

Host main-server
    HostName main.example.com
    User main-user
`
	mainPath := filepath.Join(dir, "config")
	if err := os.WriteFile(mainPath, []byte(mainConfig), 0644); err != nil {
		t.Fatal(err)
	}

	hosts, err := Parse(mainPath)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(hosts) != 2 {
		t.Fatalf("expected 2 hosts, got %d", len(hosts))
	}

	// Included hosts should come first
	if hosts[0].Alias != "included-server" {
		t.Errorf("hosts[0].Alias = %q, want %q", hosts[0].Alias, "included-server")
	}
	if hosts[1].Alias != "main-server" {
		t.Errorf("hosts[1].Alias = %q, want %q", hosts[1].Alias, "main-server")
	}
}

func TestParse_NonexistentFile(t *testing.T) {
	hosts, err := Parse("/nonexistent/path/config")
	if err != nil {
		t.Errorf("Parse() should not error on nonexistent file, got %v", err)
	}
	if hosts != nil && len(hosts) != 0 {
		t.Errorf("expected empty hosts for nonexistent file, got %d", len(hosts))
	}
}

func TestParse_EqualsSeparator(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config")

	// Some SSH configs use = as separator
	content := `
Host myserver
    HostName=192.168.1.100
    User=admin
    Port=2222
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	hosts, err := Parse(configPath)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(hosts) != 1 {
		t.Fatalf("expected 1 host, got %d", len(hosts))
	}

	if hosts[0].HostName != "192.168.1.100" {
		t.Errorf("HostName = %q, want %q", hosts[0].HostName, "192.168.1.100")
	}
	if hosts[0].User != "admin" {
		t.Errorf("User = %q, want %q", hosts[0].User, "admin")
	}
	if hosts[0].Port != 2222 {
		t.Errorf("Port = %d, want %d", hosts[0].Port, 2222)
	}
}

func TestParsedHost_ToConnection(t *testing.T) {
	host := ParsedHost{
		Alias:        "myserver",
		HostName:     "192.168.1.100",
		User:         "admin",
		Port:         2222,
		IdentityFile: "/home/user/.ssh/key",
		ProxyJump:    "bastion",
		ForwardAgent: true,
		Options: map[string]string{
			"serveraliveinterval": "60",
		},
	}

	conn := host.ToConnection()

	if conn.ID != "myserver" {
		t.Errorf("ID = %q, want %q", conn.ID, "myserver")
	}
	if conn.Host != "192.168.1.100" {
		t.Errorf("Host = %q, want %q", conn.Host, "192.168.1.100")
	}
	if conn.User != "admin" {
		t.Errorf("User = %q, want %q", conn.User, "admin")
	}
	if conn.Port != 2222 {
		t.Errorf("Port = %d, want %d", conn.Port, 2222)
	}
	if conn.ProxyJump != "bastion" {
		t.Errorf("ProxyJump = %q, want %q", conn.ProxyJump, "bastion")
	}
	if !conn.ForwardAgent {
		t.Error("ForwardAgent = false, want true")
	}
}

func TestToConnections(t *testing.T) {
	hosts := []ParsedHost{
		{Alias: "server1", HostName: "host1.example.com"},
		{Alias: "server2", HostName: "host2.example.com"},
	}

	connections := ToConnections(hosts)

	if len(connections) != 2 {
		t.Fatalf("expected 2 connections, got %d", len(connections))
	}

	if connections[0].ID != "server1" {
		t.Errorf("connections[0].ID = %q, want %q", connections[0].ID, "server1")
	}
	if connections[1].ID != "server2" {
		t.Errorf("connections[1].ID = %q, want %q", connections[1].ID, "server2")
	}
}

func TestIsWildcard(t *testing.T) {
	tests := []struct {
		pattern string
		want    bool
	}{
		{"*", true},
		{"*.example.com", true},
		{"server?", true},
		{"dev?-*", true},
		{"myserver", false},
		{"server.example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			got := isWildcard(tt.pattern)
			if got != tt.want {
				t.Errorf("isWildcard(%q) = %v, want %v", tt.pattern, got, tt.want)
			}
		})
	}
}

func TestResolveConflict(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		existingIDs map[string]bool
		want        string
	}{
		{
			name:        "no conflict with -imported suffix",
			id:          "myserver",
			existingIDs: map[string]bool{"myserver": true},
			want:        "myserver-imported",
		},
		{
			name:        "conflict with -imported, use -imported-2",
			id:          "myserver",
			existingIDs: map[string]bool{"myserver": true, "myserver-imported": true},
			want:        "myserver-imported-2",
		},
		{
			name:        "conflict up to -imported-3",
			id:          "myserver",
			existingIDs: map[string]bool{"myserver": true, "myserver-imported": true, "myserver-imported-2": true},
			want:        "myserver-imported-3",
		},
		{
			name:        "gap in numbering still finds next available",
			id:          "myserver",
			existingIDs: map[string]bool{"myserver": true, "myserver-imported": true, "myserver-imported-3": true},
			want:        "myserver-imported-2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveConflict(tt.id, tt.existingIDs)
			if got != tt.want {
				t.Errorf("ResolveConflict(%q) = %q, want %q", tt.id, got, tt.want)
			}
		})
	}
}
