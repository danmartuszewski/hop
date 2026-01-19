package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Version     int          `yaml:"version"`
	Defaults    Defaults     `yaml:"defaults"`
	Connections []Connection `yaml:"connections"`
	Groups      map[string][]string `yaml:"groups"`
}

type Defaults struct {
	User string `yaml:"user"`
	Port int    `yaml:"port"`
}

type Connection struct {
	ID           string            `yaml:"id"`
	Host         string            `yaml:"host"`
	User         string            `yaml:"user"`
	Port         int               `yaml:"port"`
	Project      string            `yaml:"project"`
	Env          string            `yaml:"env"`
	IdentityFile string            `yaml:"identity_file"`
	Tags         []string          `yaml:"tags"`
	Options      map[string]string `yaml:"options"`
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
