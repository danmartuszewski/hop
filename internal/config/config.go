package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Version     int                 `yaml:"version"`
	Defaults    Defaults            `yaml:"defaults,omitempty"`
	Connections []Connection        `yaml:"connections"`
	Groups      map[string][]string `yaml:"groups,omitempty"`
}

type Defaults struct {
	User string `yaml:"user,omitempty"`
	Port int    `yaml:"port,omitempty"`
}

type Connection struct {
	ID           string            `yaml:"id"`
	Host         string            `yaml:"host"`
	User         string            `yaml:"user,omitempty"`
	Port         int               `yaml:"port,omitempty"`
	Project      string            `yaml:"project,omitempty"`
	Env          string            `yaml:"env,omitempty"`
	IdentityFile string            `yaml:"identity_file,omitempty"`
	ProxyJump    string            `yaml:"proxy_jump,omitempty"`
	ForwardAgent bool              `yaml:"forward_agent,omitempty"`
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
	}
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
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func (c *Config) AddConnection(conn Connection) {
	c.Connections = append(c.Connections, conn)
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
