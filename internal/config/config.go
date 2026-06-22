package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Version     int                 `yaml:"version"`
	ThemePreset string              `yaml:"theme_preset,omitempty"`
	Theme       map[string]string   `yaml:"theme,omitempty"`
	ThemeDark   map[string]string   `yaml:"theme_dark,omitempty"`
	ThemeLight  map[string]string   `yaml:"theme_light,omitempty"`
	Defaults    Defaults            `yaml:"defaults,omitempty"`
	Connections []Connection        `yaml:"connections"`
	Groups      map[string][]string `yaml:"groups,omitempty"`
}

type Defaults struct {
	User        string `yaml:"user,omitempty"`
	Port        int    `yaml:"port,omitempty"`
	UseMosh     bool   `yaml:"use_mosh,omitempty"`
	HealthCheck *bool  `yaml:"health_check,omitempty"`
}

// HealthCheckEnabled returns whether health checks are enabled (default: true).
func (d *Defaults) HealthCheckEnabled() bool {
	if d.HealthCheck == nil {
		return true
	}
	return *d.HealthCheck
}

type Connection struct {
	ID           string            `yaml:"id"`
	Host         string            `yaml:"host"`
	User         string            `yaml:"user,omitempty"`
	Port         int               `yaml:"port,omitempty"`
	Project      string            `yaml:"project,omitempty"`
	Env          string            `yaml:"env,omitempty"`
	IdentityFile string            `yaml:"identity_file,omitempty"`
	RemoteDir    string            `yaml:"remote_dir,omitempty"`
	ProxyJump    string            `yaml:"proxy_jump,omitempty"`
	ForwardAgent bool              `yaml:"forward_agent,omitempty"`
	UseMosh      *bool             `yaml:"use_mosh,omitempty"`
	Tags         []string          `yaml:"tags,omitempty"`
	Options      map[string]string `yaml:"options,omitempty"`
}

func DefaultConfigPath() string {
	if path := os.Getenv("HOP_CONFIG"); path != "" {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "hop", "config.yaml")
}

func Load(path string) (*Config, error) {
	if path == "" {
		path = DefaultConfigPath()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty config if no config file exists
			cfg := &Config{
				Version:     1,
				Connections: []Connection{},
				Groups:      make(map[string][]string),
			}
			cfg.applyDefaults()
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	cfg.applyDefaults()
	return &cfg, nil
}

func (c *Config) applyDefaults() {
	if c.Defaults.Port == 0 {
		c.Defaults.Port = 22
	}

	for i := range c.Connections {
		conn := &c.Connections[i]
		if conn.User == "" {
			conn.User = c.Defaults.User
		}
		if conn.Port == 0 {
			conn.Port = c.Defaults.Port
		}
		if conn.UseMosh == nil && c.Defaults.UseMosh {
			v := true
			conn.UseMosh = &v
		}
	}
}

// Mosh returns whether mosh is enabled for this connection.
func (c *Connection) Mosh() bool {
	if c.UseMosh == nil {
		return false
	}
	return *c.UseMosh
}

// SetMosh sets the mosh flag on a connection.
func (c *Connection) SetMosh(v bool) {
	c.UseMosh = &v
}

func (c *Connection) EffectiveUser() string {
	if c.User != "" {
		return c.User
	}
	return os.Getenv("USER")
}

func (c *Config) Save(path string) error {
	if path == "" {
		path = DefaultConfigPath()
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	// 0700/0600: the config holds the full infrastructure inventory, so keep it
	// readable only by the owner, matching SSH's own conventions for ~/.ssh.
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Clone returns a deep copy of the connection. Reference-type fields (Tags,
// Options) and the UseMosh pointer are copied into fresh backing storage so the
// clone shares no mutable state with the source. This matters for duplication,
// where the source and the copy must be fully independent config entries.
func (c Connection) Clone() Connection {
	clone := c

	if c.Tags != nil {
		clone.Tags = append([]string(nil), c.Tags...)
	}
	if c.Options != nil {
		clone.Options = make(map[string]string, len(c.Options))
		for k, v := range c.Options {
			clone.Options[k] = v
		}
	}
	if c.UseMosh != nil {
		mosh := *c.UseMosh
		clone.UseMosh = &mosh
	}

	return clone
}

func (c *Config) AddConnection(conn Connection) {
	c.Connections = append(c.Connections, conn)
}

// SuggestDuplicateID returns a unique connection ID derived from sourceID for
// use when duplicating a connection. It tries "<sourceID>-copy" first, then
// "<sourceID>-copy-2", "<sourceID>-copy-3", ... until it finds an ID that no
// existing connection uses.
func (c *Config) SuggestDuplicateID(sourceID string) string {
	base := sourceID + "-copy"
	if c.FindConnection(base) == nil {
		return base
	}
	for n := 2; ; n++ {
		candidate := fmt.Sprintf("%s-%d", base, n)
		if c.FindConnection(candidate) == nil {
			return candidate
		}
	}
}

func (c *Config) UpdateConnection(id string, conn Connection) bool {
	for i, existing := range c.Connections {
		if existing.ID == id {
			c.Connections[i] = conn
			return true
		}
	}
	return false
}

func (c *Config) DeleteConnection(id string) bool {
	for i, conn := range c.Connections {
		if conn.ID == id {
			c.Connections = append(c.Connections[:i], c.Connections[i+1:]...)
			return true
		}
	}
	return false
}

func (c *Config) FindConnection(id string) *Connection {
	for i, conn := range c.Connections {
		if conn.ID == id {
			return &c.Connections[i]
		}
	}
	return nil
}
