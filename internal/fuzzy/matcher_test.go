package fuzzy

import (
	"testing"

	"github.com/hop-cli/hop/internal/config"
)

func TestFindMatches(t *testing.T) {
	connections := []config.Connection{
		{ID: "myapp-prod-web1", Host: "web1.myapp.com", Project: "myapp", Env: "prod", Tags: []string{"web"}},
		{ID: "myapp-prod-web2", Host: "web2.myapp.com", Project: "myapp", Env: "prod", Tags: []string{"web"}},
		{ID: "myapp-prod-db", Host: "db.myapp.com", Project: "myapp", Env: "prod", Tags: []string{"database"}},
		{ID: "myapp-staging-web1", Host: "staging.myapp.com", Project: "myapp", Env: "staging", Tags: []string{"web"}},
		{ID: "client-prod-api", Host: "10.0.1.50", Project: "client", Env: "prod", Tags: []string{"api"}},
	}

	tests := []struct {
		name      string
		query     string
		wantFirst string
		wantCount int
	}{
		{
			name:      "exact match",
			query:     "myapp-prod-web1",
			wantFirst: "myapp-prod-web1",
			wantCount: 1,
		},
		{
			name:      "prefix match",
			query:     "myapp-prod",
			wantFirst: "myapp-prod-db",
			wantCount: 3,
		},
		{
			name:      "partial match",
			query:     "web1",
			wantFirst: "myapp-prod-web1",
			wantCount: 2,
		},
		{
			name:      "tag match",
			query:     "web",
			wantFirst: "myapp-prod-web1",
			wantCount: 3, // web1, web2, staging-web1 (db has database tag, api has api tag)
		},
		{
			name:      "host match",
			query:     "staging",
			wantFirst: "myapp-staging-web1",
			wantCount: 1,
		},
		{
			name:      "no match",
			query:     "nonexistent",
			wantFirst: "",
			wantCount: 0,
		},
		{
			name:      "case insensitive",
			query:     "MYAPP",
			wantFirst: "myapp-prod-db",
			wantCount: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := FindMatches(tt.query, connections)
			if len(matches) != tt.wantCount {
				t.Errorf("FindMatches(%q) returned %d matches, want %d", tt.query, len(matches), tt.wantCount)
			}
			if tt.wantCount > 0 && matches[0].Connection.ID != tt.wantFirst {
				t.Errorf("FindMatches(%q) first match = %q, want %q", tt.query, matches[0].Connection.ID, tt.wantFirst)
			}
		})
	}
}

func TestFindBestMatch(t *testing.T) {
	connections := []config.Connection{
		{ID: "prod", Host: "prod.example.com"},
		{ID: "prod-web", Host: "web.prod.example.com"},
		{ID: "prod-web-1", Host: "web1.prod.example.com"},
	}

	match := FindBestMatch("prod", connections)
	if match == nil {
		t.Fatal("FindBestMatch returned nil")
	}
	if match.ID != "prod" {
		t.Errorf("FindBestMatch('prod') = %q, want 'prod' (exact match)", match.ID)
	}

	match = FindBestMatch("prod-web", connections)
	if match.ID != "prod-web" {
		t.Errorf("FindBestMatch('prod-web') = %q, want 'prod-web'", match.ID)
	}
}

func TestFindByID(t *testing.T) {
	connections := []config.Connection{
		{ID: "server1", Host: "server1.example.com"},
		{ID: "server2", Host: "server2.example.com"},
	}

	conn := FindByID("server1", connections)
	if conn == nil {
		t.Fatal("FindByID returned nil for existing connection")
	}
	if conn.ID != "server1" {
		t.Errorf("FindByID('server1').ID = %q, want 'server1'", conn.ID)
	}

	conn = FindByID("nonexistent", connections)
	if conn != nil {
		t.Error("FindByID should return nil for nonexistent ID")
	}
}

func TestFuzzyMatch(t *testing.T) {
	tests := []struct {
		pattern string
		text    string
		want    bool
	}{
		{"abc", "abc", true},
		{"abc", "aXbXc", true},
		{"abc", "abcd", true},
		{"abc", "ab", false},
		{"pw", "prod-web", true},
		{"mpw", "myapp-prod-web", true},
		{"xyz", "abc", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.text, func(t *testing.T) {
			got := fuzzyMatch(tt.pattern, tt.text)
			if got != tt.want {
				t.Errorf("fuzzyMatch(%q, %q) = %v, want %v", tt.pattern, tt.text, got, tt.want)
			}
		})
	}
}

func TestMatchGroup(t *testing.T) {
	cfg := &config.Config{
		Connections: []config.Connection{
			{ID: "myapp-prod-web1", Project: "myapp", Env: "prod"},
			{ID: "myapp-prod-web2", Project: "myapp", Env: "prod"},
			{ID: "myapp-staging-web1", Project: "myapp", Env: "staging"},
			{ID: "client-prod-api", Project: "client", Env: "prod"},
		},
		Groups: map[string][]string{
			"all-web": {"myapp-prod-web1", "myapp-prod-web2", "myapp-staging-web1"},
		},
	}

	matches := MatchGroup("all-web", cfg)
	if len(matches) != 3 {
		t.Errorf("MatchGroup('all-web') returned %d matches, want 3", len(matches))
	}

	matches = MatchGroup("myapp-prod", cfg)
	if len(matches) != 2 {
		t.Errorf("MatchGroup('myapp-prod') returned %d matches, want 2", len(matches))
	}

	matches = MatchGroup("myapp", cfg)
	if len(matches) != 3 {
		t.Errorf("MatchGroup('myapp') returned %d matches, want 3", len(matches))
	}

	matches = MatchGroup("prod", cfg)
	if len(matches) != 3 {
		t.Errorf("MatchGroup('prod') returned %d matches, want 3", len(matches))
	}
}

func TestMatchByTag(t *testing.T) {
	connections := []config.Connection{
		{ID: "web1", Tags: []string{"web", "frontend"}},
		{ID: "web2", Tags: []string{"web", "frontend"}},
		{ID: "db1", Tags: []string{"database", "postgres"}},
		{ID: "api1", Tags: []string{"api", "backend"}},
	}

	matches := MatchByTag("web", connections)
	if len(matches) != 2 {
		t.Errorf("MatchByTag('web') returned %d matches, want 2", len(matches))
	}

	matches = MatchByTag("database", connections)
	if len(matches) != 1 {
		t.Errorf("MatchByTag('database') returned %d matches, want 1", len(matches))
	}

	matches = MatchByTag("WEB", connections)
	if len(matches) != 2 {
		t.Errorf("MatchByTag('WEB') should be case insensitive, got %d matches", len(matches))
	}

	matches = MatchByTag("nonexistent", connections)
	if len(matches) != 0 {
		t.Errorf("MatchByTag('nonexistent') returned %d matches, want 0", len(matches))
	}
}

func TestContainsIgnoreCase(t *testing.T) {
	tests := []struct {
		text   string
		substr string
		want   bool
	}{
		{"hello world", "world", true},
		{"Hello World", "world", true},
		{"hello world", "WORLD", true},
		{"HELLO WORLD", "hello", true},
		{"myapp-prod-web1", "prod", true},
		{"myapp-prod-web1", "PROD", true},
		{"hello", "xyz", false},
		{"", "test", false},
		{"test", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.text+"_"+tt.substr, func(t *testing.T) {
			got := ContainsIgnoreCase(tt.text, tt.substr)
			if got != tt.want {
				t.Errorf("ContainsIgnoreCase(%q, %q) = %v, want %v", tt.text, tt.substr, got, tt.want)
			}
		})
	}
}
