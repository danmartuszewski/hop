package ssh

import (
	"os"
	"reflect"
	"testing"

	"github.com/hop-cli/hop/internal/config"
)

func TestBuildCommand(t *testing.T) {
	tests := []struct {
		name string
		conn *config.Connection
		opts *ConnectOptions
		want []string
	}{
		{
			name: "basic connection",
			conn: &config.Connection{
				Host: "example.com",
				User: "admin",
			},
			opts: nil,
			want: []string{"admin@example.com"},
		},
		{
			name: "custom port",
			conn: &config.Connection{
				Host: "example.com",
				User: "admin",
				Port: 2222,
			},
			opts: nil,
			want: []string{"-p", "2222", "admin@example.com"},
		},
		{
			name: "with identity file",
			conn: &config.Connection{
				Host:         "example.com",
				User:         "deploy",
				IdentityFile: "/path/to/key.pem",
			},
			opts: nil,
			want: []string{"-i", "/path/to/key.pem", "deploy@example.com"},
		},
		{
			name: "with ssh options",
			conn: &config.Connection{
				Host: "example.com",
				User: "admin",
				Options: map[string]string{
					"StrictHostKeyChecking": "no",
				},
			},
			opts: nil,
			want: []string{"-o", "StrictHostKeyChecking=no", "admin@example.com"},
		},
		{
			name: "force tty",
			conn: &config.Connection{
				Host: "example.com",
				User: "admin",
			},
			opts: &ConnectOptions{ForceTTY: true},
			want: []string{"-t", "admin@example.com"},
		},
		{
			name: "with remote command",
			conn: &config.Connection{
				Host: "example.com",
				User: "admin",
			},
			opts: &ConnectOptions{Command: "uptime"},
			want: []string{"admin@example.com", "uptime"},
		},
		{
			name: "no user specified uses system user",
			conn: &config.Connection{
				Host: "example.com",
			},
			opts: nil,
			want: func() []string {
				if user := os.Getenv("USER"); user != "" {
					return []string{user + "@example.com"}
				}
				return []string{"example.com"}
			}(),
		},
		{
			name: "default port 22 not included",
			conn: &config.Connection{
				Host: "example.com",
				User: "admin",
				Port: 22,
			},
			opts: nil,
			want: []string{"admin@example.com"},
		},
		{
			name: "full complex connection",
			conn: &config.Connection{
				Host:         "10.0.1.50",
				User:         "deploy",
				Port:         2222,
				IdentityFile: "/keys/server.pem",
				Options: map[string]string{
					"ServerAliveInterval": "60",
				},
			},
			opts: &ConnectOptions{
				ForceTTY: true,
				Command:  "htop",
			},
			want: []string{"-t", "-p", "2222", "-i", "/keys/server.pem", "-o", "ServerAliveInterval=60", "deploy@10.0.1.50", "htop"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildCommand(tt.conn, tt.opts)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildCommandString(t *testing.T) {
	conn := &config.Connection{
		Host: "example.com",
		User: "admin",
		Port: 2222,
	}
	got := BuildCommandString(conn, nil)
	want := "ssh -p 2222 admin@example.com"
	if got != want {
		t.Errorf("BuildCommandString() = %q, want %q", got, want)
	}
}

func TestBuildCommandString_WithSpaces(t *testing.T) {
	conn := &config.Connection{
		Host: "example.com",
		User: "admin",
	}
	opts := &ConnectOptions{
		Command: "echo hello world",
	}
	got := BuildCommandString(conn, opts)
	want := `ssh admin@example.com "echo hello world"`
	if got != want {
		t.Errorf("BuildCommandString() = %q, want %q", got, want)
	}
}

func TestExpandPath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "absolute path unchanged",
			input: "/path/to/key.pem",
			want:  "/path/to/key.pem",
		},
		{
			name:  "relative path unchanged",
			input: "keys/server.pem",
			want:  "keys/server.pem",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandPath(tt.input)
			if got != tt.want {
				t.Errorf("expandPath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
