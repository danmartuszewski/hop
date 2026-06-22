package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func boolPtr(v bool) *bool { return &v }

func TestLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	content := `version: 1

defaults:
  user: testuser
  port: 22

connections:
  - id: server1
    host: server1.example.com
    tags: [web]

  - id: server2
    host: server2.example.com
    user: admin
    port: 2222
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Version != 1 {
		t.Errorf("Version = %d, want 1", cfg.Version)
	}

	if cfg.Defaults.User != "testuser" {
		t.Errorf("Defaults.User = %s, want testuser", cfg.Defaults.User)
	}

	if len(cfg.Connections) != 2 {
		t.Errorf("len(Connections) = %d, want 2", len(cfg.Connections))
	}

	conn1 := cfg.Connections[0]
	if conn1.ID != "server1" {
		t.Errorf("Connections[0].ID = %s, want server1", conn1.ID)
	}
	if conn1.User != "testuser" {
		t.Errorf("Connections[0].User = %s, want testuser (from defaults)", conn1.User)
	}

	conn2 := cfg.Connections[1]
	if conn2.User != "admin" {
		t.Errorf("Connections[1].User = %s, want admin", conn2.User)
	}
	if conn2.Port != 2222 {
		t.Errorf("Connections[1].Port = %d, want 2222", conn2.Port)
	}
}

func TestSaveOmitsEmptyOptionalFields(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Config with minimal connection — only required fields + one optional
	content := `version: 1

defaults:
  user: deploy

connections:
  - id: homelab
    host: 192.168.1.100
    user: dan
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Save (round-trip)
	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	saved, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read saved config: %v", err)
	}
	output := string(saved)

	// Empty optional fields should NOT appear in saved config
	for _, field := range []string{"project:", "env:", "identity_file:", "tags:", "options:"} {
		if strings.Contains(output, field) {
			t.Errorf("saved config should not contain %q when field was not set, got:\n%s", field, output)
		}
	}
}

func TestLoadWithUseMosh(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	content := `version: 1
connections:
  - id: mosh-server
    host: example.com
    user: admin
    use_mosh: true

  - id: ssh-server
    host: other.example.com
    user: admin
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !cfg.Connections[0].Mosh() {
		t.Error("Connections[0].Mosh() = false, want true")
	}
	if cfg.Connections[1].Mosh() {
		t.Error("Connections[1].Mosh() = true, want false")
	}
}

func TestSaveOmitsUseMoshWhenFalse(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := &Config{
		Version: 1,
		Connections: []Connection{
			{ID: "server1", Host: "example.com", Port: 22},
		},
	}

	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	saved, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read saved config: %v", err)
	}

	if strings.Contains(string(saved), "use_mosh") {
		t.Errorf("saved config should not contain 'use_mosh' when false, got:\n%s", string(saved))
	}
}

func TestSaveIncludesUseMoshWhenTrue(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := &Config{
		Version: 1,
		Connections: []Connection{
			{ID: "server1", Host: "example.com", Port: 22, UseMosh: boolPtr(true)},
		},
	}

	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	saved, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read saved config: %v", err)
	}

	if !strings.Contains(string(saved), "use_mosh: true") {
		t.Errorf("saved config should contain 'use_mosh: true', got:\n%s", string(saved))
	}
}

func TestDefaultsUseMoshApplied(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	content := `version: 1
defaults:
  use_mosh: true

connections:
  - id: inherits-mosh
    host: example.com

  - id: overrides-mosh
    host: other.example.com
    use_mosh: false
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Connection without use_mosh should inherit the global default
	if !cfg.Connections[0].Mosh() {
		t.Error("Connections[0].Mosh() = false, want true (inherited from defaults)")
	}

	// Connection with explicit use_mosh: false should override the default
	if cfg.Connections[1].Mosh() {
		t.Error("Connections[1].Mosh() = true, want false (explicit override)")
	}
}

func TestDefaultsUseMoshNotAppliedWhenFalse(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	content := `version: 1
connections:
  - id: server1
    host: example.com
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Connections[0].Mosh() {
		t.Error("Connections[0].Mosh() = true, want false (no default set)")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Version: 1,
				Connections: []Connection{
					{ID: "server1", Host: "example.com"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid version",
			config: Config{
				Version: 2,
				Connections: []Connection{
					{ID: "server1", Host: "example.com"},
				},
			},
			wantErr: true,
		},
		{
			name: "missing id",
			config: Config{
				Version: 1,
				Connections: []Connection{
					{Host: "example.com"},
				},
			},
			wantErr: true,
		},
		{
			name: "missing host",
			config: Config{
				Version: 1,
				Connections: []Connection{
					{ID: "server1"},
				},
			},
			wantErr: true,
		},
		{
			name: "duplicate id",
			config: Config{
				Version: 1,
				Connections: []Connection{
					{ID: "server1", Host: "example1.com"},
					{ID: "server1", Host: "example2.com"},
				},
			},
			wantErr: true,
		},
		{
			name: "group references unknown connection",
			config: Config{
				Version: 1,
				Connections: []Connection{
					{ID: "server1", Host: "example.com"},
				},
				Groups: map[string][]string{
					"web": {"server1", "unknown"},
				},
			},
			wantErr: true,
		},
		{
			name: "host starting with dash is rejected",
			config: Config{
				Version: 1,
				Connections: []Connection{
					{ID: "evil", Host: "-oProxyCommand=touch /tmp/pwned"},
				},
			},
			wantErr: true,
		},
		{
			name: "user starting with dash is rejected",
			config: Config{
				Version: 1,
				Connections: []Connection{
					{ID: "evil", Host: "example.com", User: "-oProxyCommand=id"},
				},
			},
			wantErr: true,
		},
		{
			name: "proxy jump starting with dash is rejected",
			config: Config{
				Version: 1,
				Connections: []Connection{
					{ID: "evil", Host: "example.com", ProxyJump: "-oProxyCommand=id"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConnectionCheckSafety(t *testing.T) {
	tests := []struct {
		name    string
		conn    Connection
		wantErr bool
	}{
		{name: "plain host", conn: Connection{Host: "example.com"}, wantErr: false},
		{name: "user@host", conn: Connection{Host: "example.com", User: "admin"}, wantErr: false},
		{name: "proxy jump host", conn: Connection{Host: "example.com", ProxyJump: "bastion"}, wantErr: false},
		{name: "dash host", conn: Connection{Host: "-oProxyCommand=touch /tmp/pwned"}, wantErr: true},
		{name: "dash user", conn: Connection{Host: "example.com", User: "-x"}, wantErr: true},
		{name: "dash proxy jump", conn: Connection{Host: "example.com", ProxyJump: "-x"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.conn.CheckSafety()
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckSafety() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSuggestDuplicateID(t *testing.T) {
	tests := []struct {
		name     string
		existing []string
		sourceID string
		want     string
	}{
		{
			name:     "no collision appends -copy",
			existing: []string{"web-prod"},
			sourceID: "web-prod",
			want:     "web-prod-copy",
		},
		{
			name:     "copy taken falls back to -copy-2",
			existing: []string{"web-prod", "web-prod-copy"},
			sourceID: "web-prod",
			want:     "web-prod-copy-2",
		},
		{
			name:     "copy and copy-2 taken falls back to -copy-3",
			existing: []string{"web-prod", "web-prod-copy", "web-prod-copy-2"},
			sourceID: "web-prod",
			want:     "web-prod-copy-3",
		},
		{
			name:     "source already ends in -copy",
			existing: []string{"web-copy"},
			sourceID: "web-copy",
			want:     "web-copy-copy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Version: 1}
			for _, id := range tt.existing {
				cfg.Connections = append(cfg.Connections, Connection{ID: id, Host: "h"})
			}
			if got := cfg.SuggestDuplicateID(tt.sourceID); got != tt.want {
				t.Errorf("SuggestDuplicateID(%q) = %q, want %q", tt.sourceID, got, tt.want)
			}
		})
	}
}

func TestConnectionCloneIsDeepCopy(t *testing.T) {
	mosh := true
	src := Connection{
		ID:      "src",
		Host:    "src.example.com",
		Tags:    []string{"prod", "web"},
		Options: map[string]string{"ServerAliveInterval": "60"},
		UseMosh: &mosh,
	}

	clone := src.Clone()

	// Mutating the clone's reference fields must not affect the source.
	clone.Tags[0] = "mutated"
	clone.Options["ServerAliveInterval"] = "99"
	*clone.UseMosh = false

	if src.Tags[0] != "prod" {
		t.Errorf("expected source Tags unchanged, got %v", src.Tags)
	}
	if src.Options["ServerAliveInterval"] != "60" {
		t.Errorf("expected source Options unchanged, got %v", src.Options)
	}
	if src.UseMosh == nil || !*src.UseMosh {
		t.Error("expected source UseMosh unchanged")
	}
}
