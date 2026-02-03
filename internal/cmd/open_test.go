package cmd

import (
	"testing"

	"github.com/danmartuszewski/hop/internal/config"
)

func TestResolveConnections(t *testing.T) {
	cfg := &config.Config{
		Connections: []config.Connection{
			{ID: "myapp-prod-web1", Host: "web1.myapp.com", Project: "myapp", Env: "prod", Tags: []string{"web"}},
			{ID: "myapp-prod-web2", Host: "web2.myapp.com", Project: "myapp", Env: "prod", Tags: []string{"web"}},
			{ID: "myapp-prod-db", Host: "db.myapp.com", Project: "myapp", Env: "prod", Tags: []string{"database"}},
			{ID: "myapp-staging-web1", Host: "staging.myapp.com", Project: "myapp", Env: "staging", Tags: []string{"web"}},
			{ID: "client-prod-api", Host: "10.0.1.50", Project: "client", Env: "prod", Tags: []string{"api"}},
		},
		Groups: map[string][]string{
			"all-web": {"myapp-prod-web1", "myapp-prod-web2", "myapp-staging-web1"},
		},
	}

	tests := []struct {
		name      string
		queries   []string
		tag       string
		wantCount int
		wantIDs   []string
		wantErr   bool
	}{
		{
			name:      "named group",
			queries:   []string{"all-web"},
			tag:       "",
			wantCount: 3,
			wantIDs:   []string{"myapp-prod-web1", "myapp-prod-web2", "myapp-staging-web1"},
		},
		{
			name:      "project-env group",
			queries:   []string{"myapp-prod"},
			tag:       "",
			wantCount: 3,
			wantIDs:   []string{"myapp-prod-web1", "myapp-prod-web2", "myapp-prod-db"},
		},
		{
			name:      "project only",
			queries:   []string{"myapp"},
			tag:       "",
			wantCount: 4,
		},
		{
			name:      "env only",
			queries:   []string{"prod"},
			tag:       "",
			wantCount: 4,
		},
		{
			name:      "exact ID match",
			queries:   []string{"myapp-prod-web1"},
			tag:       "",
			wantCount: 1,
			wantIDs:   []string{"myapp-prod-web1"},
		},
		{
			name:      "multiple exact IDs",
			queries:   []string{"myapp-prod-web1", "myapp-prod-db"},
			tag:       "",
			wantCount: 2,
			wantIDs:   []string{"myapp-prod-web1", "myapp-prod-db"},
		},
		{
			name:      "tag only",
			queries:   []string{},
			tag:       "web",
			wantCount: 3,
		},
		{
			name:      "tag only database",
			queries:   []string{},
			tag:       "database",
			wantCount: 1,
			wantIDs:   []string{"myapp-prod-db"},
		},
		{
			name:      "group with tag filter",
			queries:   []string{"myapp-prod"},
			tag:       "web",
			wantCount: 2,
			wantIDs:   []string{"myapp-prod-web1", "myapp-prod-web2"},
		},
		{
			name:      "no duplicates",
			queries:   []string{"myapp-prod-web1", "myapp-prod-web1"},
			tag:       "",
			wantCount: 1,
			wantIDs:   []string{"myapp-prod-web1"},
		},
		{
			name:      "group and ID no duplicates",
			queries:   []string{"all-web", "myapp-prod-web1"},
			tag:       "",
			wantCount: 3,
		},
		{
			name:    "no matches",
			queries: []string{"nonexistent"},
			tag:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connections, err := resolveConnections(cfg, tt.queries, tt.tag)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(connections) != tt.wantCount {
				t.Errorf("got %d connections, want %d", len(connections), tt.wantCount)
				for _, c := range connections {
					t.Logf("  - %s", c.ID)
				}
			}

			if tt.wantIDs != nil {
				got := make(map[string]bool)
				for _, c := range connections {
					got[c.ID] = true
				}
				for _, wantID := range tt.wantIDs {
					if !got[wantID] {
						t.Errorf("expected connection %q not found", wantID)
					}
				}
			}
		})
	}
}

func TestFilterByTag(t *testing.T) {
	connections := []config.Connection{
		{ID: "web1", Tags: []string{"web", "frontend"}},
		{ID: "web2", Tags: []string{"web", "frontend"}},
		{ID: "db1", Tags: []string{"database", "postgres"}},
		{ID: "api1", Tags: []string{"api", "backend"}},
	}

	tests := []struct {
		name      string
		tag       string
		wantCount int
	}{
		{"web tag", "web", 2},
		{"database tag", "database", 1},
		{"case insensitive", "WEB", 2},
		{"no match", "nonexistent", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := filterByTag(connections, tt.tag)
			if len(filtered) != tt.wantCount {
				t.Errorf("filterByTag(%q) returned %d, want %d", tt.tag, len(filtered), tt.wantCount)
			}
		})
	}
}
