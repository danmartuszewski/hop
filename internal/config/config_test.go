package config

import (
	"os"
	"path/filepath"
	"testing"
)

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
