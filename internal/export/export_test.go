package export

import (
	"bytes"
	"strings"
	"testing"

	"github.com/danmartuszewski/hop/internal/config"
)

func TestBuildExportConfig(t *testing.T) {
	conns := []config.Connection{
		{ID: "web-1", Host: "10.0.0.1", User: "admin", Port: 22},
		{ID: "db-1", Host: "10.0.0.2", User: "root", Port: 5432},
	}

	cfg := BuildExportConfig(conns)

	if cfg.Version != 1 {
		t.Errorf("expected version 1, got %d", cfg.Version)
	}
	if len(cfg.Connections) != 2 {
		t.Errorf("expected 2 connections, got %d", len(cfg.Connections))
	}
	if cfg.Groups != nil {
		t.Error("expected nil groups")
	}
}

func TestBuildExportConfigEmpty(t *testing.T) {
	cfg := BuildExportConfig(nil)
	if cfg.Version != 1 {
		t.Errorf("expected version 1, got %d", cfg.Version)
	}
	if cfg.Connections != nil {
		t.Error("expected nil connections for nil input")
	}
}

func TestWriteYAML(t *testing.T) {
	conns := []config.Connection{
		{ID: "web-1", Host: "10.0.0.1", User: "admin"},
	}
	cfg := BuildExportConfig(conns)

	var buf bytes.Buffer
	if err := WriteYAML(&buf, cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "version: 1") {
		t.Error("expected output to contain 'version: 1'")
	}
	if !strings.Contains(out, "id: web-1") {
		t.Error("expected output to contain 'id: web-1'")
	}
	if !strings.Contains(out, "host: 10.0.0.1") {
		t.Error("expected output to contain 'host: 10.0.0.1'")
	}
}

func TestWriteYAMLRoundTrip(t *testing.T) {
	conns := []config.Connection{
		{
			ID:      "app-prod",
			Host:    "prod.example.com",
			User:    "deploy",
			Port:    2222,
			Project: "myapp",
			Env:     "production",
			Tags:    []string{"web", "critical"},
		},
	}
	cfg := BuildExportConfig(conns)

	var buf bytes.Buffer
	if err := WriteYAML(&buf, cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "project: myapp") {
		t.Error("expected output to contain project field")
	}
	if !strings.Contains(out, "env: production") {
		t.Error("expected output to contain env field")
	}
	if !strings.Contains(out, "- web") {
		t.Error("expected output to contain tags")
	}
}
