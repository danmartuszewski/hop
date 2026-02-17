package resolve

import (
	"testing"

	"github.com/danmartuszewski/hop/internal/config"
)

func testConfig() *config.Config {
	cfg := &config.Config{
		Version: 1,
		Connections: []config.Connection{
			{ID: "web-prod-1", Host: "web1.prod.example.com", Port: 22, Project: "myapp", Env: "prod", Tags: []string{"web", "production"}},
			{ID: "web-prod-2", Host: "web2.prod.example.com", Port: 22, Project: "myapp", Env: "prod", Tags: []string{"web", "production"}},
			{ID: "db-prod", Host: "db.prod.example.com", Port: 22, Project: "myapp", Env: "prod", Tags: []string{"database", "production"}},
			{ID: "web-staging", Host: "web.staging.example.com", Port: 22, Project: "myapp", Env: "staging", Tags: []string{"web", "staging"}},
			{ID: "api-prod", Host: "api.prod.example.com", Port: 22, Project: "api", Env: "prod", Tags: []string{"api"}},
		},
		Groups: map[string][]string{
			"production": {"web-prod-1", "web-prod-2", "db-prod"},
			"web-tier":   {"web-prod-1", "web-prod-2", "web-staging"},
		},
	}
	return cfg
}

func TestResolveByNamedGroup(t *testing.T) {
	cfg := testConfig()
	result, err := ResolveTarget("production", cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Method != MatchNamedGroup {
		t.Errorf("expected MatchNamedGroup, got %v", result.Method)
	}
	if len(result.Connections) != 3 {
		t.Errorf("expected 3 connections, got %d", len(result.Connections))
	}
}

func TestResolveByProjectEnv(t *testing.T) {
	cfg := testConfig()
	result, err := ResolveTarget("myapp-prod", cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Method != MatchProjectEnv {
		t.Errorf("expected MatchProjectEnv, got %v", result.Method)
	}
	if len(result.Connections) == 0 {
		t.Error("expected at least one connection")
	}
}

func TestResolveByGlobPattern(t *testing.T) {
	cfg := testConfig()
	result, err := ResolveTarget("web-prod-*", cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Method != MatchGlob {
		t.Errorf("expected MatchGlob, got %v", result.Method)
	}
	if len(result.Connections) != 2 {
		t.Errorf("expected 2 connections, got %d", len(result.Connections))
	}
}

func TestResolveByFuzzyMatch(t *testing.T) {
	cfg := testConfig()
	result, err := ResolveTarget("api-prod", cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// api-prod matches project-env pattern for api project
	if len(result.Connections) == 0 {
		t.Error("expected at least one connection")
	}
}

func TestResolveNoMatches(t *testing.T) {
	cfg := testConfig()
	result, err := ResolveTarget("nonexistent-server-xyz", cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Method != MatchNone {
		t.Errorf("expected MatchNone, got %v", result.Method)
	}
	if len(result.Connections) != 0 {
		t.Errorf("expected 0 connections, got %d", len(result.Connections))
	}
}

func TestResolveGlobNoMatches(t *testing.T) {
	cfg := testConfig()
	result, err := ResolveTarget("zzz*", cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Method != MatchGlob {
		t.Errorf("expected MatchGlob, got %v", result.Method)
	}
	if len(result.Connections) != 0 {
		t.Errorf("expected 0 connections, got %d", len(result.Connections))
	}
}

func TestResolveGlobQuestionMark(t *testing.T) {
	cfg := testConfig()
	result, err := ResolveTarget("web-prod-?", cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Method != MatchGlob {
		t.Errorf("expected MatchGlob, got %v", result.Method)
	}
	if len(result.Connections) != 2 {
		t.Errorf("expected 2 connections, got %d", len(result.Connections))
	}
}

func TestGlobToRegex(t *testing.T) {
	tests := []struct {
		glob     string
		expected string
	}{
		{"web*", "^web.*$"},
		{"prod-?-db", "^prod-.-db$"},
		{"exact", "^exact$"},
		{"*", "^.*$"},
	}

	for _, tt := range tests {
		got := globToRegex(tt.glob)
		if got != tt.expected {
			t.Errorf("globToRegex(%q) = %q, want %q", tt.glob, got, tt.expected)
		}
	}
}

func TestMatchMethodString(t *testing.T) {
	tests := []struct {
		method   MatchMethod
		expected string
	}{
		{MatchNamedGroup, "named group"},
		{MatchProjectEnv, "project-env pattern"},
		{MatchGlob, "glob pattern"},
		{MatchFuzzy, "fuzzy match"},
		{MatchNone, "none"},
	}

	for _, tt := range tests {
		got := tt.method.String()
		if got != tt.expected {
			t.Errorf("MatchMethod(%d).String() = %q, want %q", tt.method, got, tt.expected)
		}
	}
}
