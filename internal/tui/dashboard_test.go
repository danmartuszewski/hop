package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/danmartuszewski/hop/internal/config"
)

func testConfig() *config.Config {
	return &config.Config{
		Version: 1,
		Defaults: config.Defaults{
			Port: 22,
		},
		Connections: []config.Connection{
			{ID: "dev-server", Host: "dev.example.com", User: "developer", Port: 22, Project: "myapp", Env: "dev"},
			{ID: "prod-server", Host: "prod.example.com", User: "admin", Port: 22, Project: "myapp", Env: "prod"},
			{ID: "staging", Host: "staging.example.com", User: "deploy", Port: 2222, Project: "myapp", Env: "staging"},
		},
		Groups: map[string][]string{},
	}
}

func emptyConfig() *config.Config {
	return &config.Config{
		Version:     1,
		Connections: []config.Connection{},
		Groups:      map[string][]string{},
	}
}

func TestNewModel(t *testing.T) {
	cfg := testConfig()
	m := NewModel(cfg, "1.0.0")

	if m.version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", m.version)
	}

	if len(m.filtered) != 3 {
		t.Errorf("expected 3 filtered items, got %d", len(m.filtered))
	}

	if m.view != viewList {
		t.Errorf("expected viewList, got %v", m.view)
	}
}

func TestNavigationKeys(t *testing.T) {
	cfg := testConfig()
	m := NewModel(cfg, "1.0.0")

	// Initial cursor should be 0
	if m.cursor != 0 {
		t.Errorf("expected cursor 0, got %d", m.cursor)
	}

	// Press down
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = newModel.(Model)
	if m.cursor != 1 {
		t.Errorf("expected cursor 1 after down, got %d", m.cursor)
	}

	// Press j (vim down)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = newModel.(Model)
	if m.cursor != 2 {
		t.Errorf("expected cursor 2 after j, got %d", m.cursor)
	}

	// Press up
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = newModel.(Model)
	if m.cursor != 1 {
		t.Errorf("expected cursor 1 after up, got %d", m.cursor)
	}

	// Press k (vim up)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = newModel.(Model)
	if m.cursor != 0 {
		t.Errorf("expected cursor 0 after k, got %d", m.cursor)
	}

	// Press G (go to end)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	m = newModel.(Model)
	if m.cursor != 2 {
		t.Errorf("expected cursor 2 after G, got %d", m.cursor)
	}

	// Press g (go to start)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	m = newModel.(Model)
	if m.cursor != 0 {
		t.Errorf("expected cursor 0 after g, got %d", m.cursor)
	}
}

func TestViewSwitching(t *testing.T) {
	cfg := testConfig()
	m := NewModel(cfg, "1.0.0")

	// Press ? to open help
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = newModel.(Model)
	if m.view != viewHelp {
		t.Errorf("expected viewHelp, got %v", m.view)
	}

	// Press any key to close help
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	m = newModel.(Model)
	if m.view != viewList {
		t.Errorf("expected viewList after closing help, got %v", m.view)
	}

	// Press a to open add form
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = newModel.(Model)
	if m.view != viewForm {
		t.Errorf("expected viewForm, got %v", m.view)
	}

	// Press esc to cancel form
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = newModel.(Model)
	if m.view != viewList {
		t.Errorf("expected viewList after esc, got %v", m.view)
	}

	// Press d to open delete confirm
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = newModel.(Model)
	if m.view != viewConfirmDelete {
		t.Errorf("expected viewConfirmDelete, got %v", m.view)
	}

	// Press n to cancel delete
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = newModel.(Model)
	if m.view != viewList {
		t.Errorf("expected viewList after cancel delete, got %v", m.view)
	}

	// Press p to open paste view
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	m = newModel.(Model)
	if m.view != viewPaste {
		t.Errorf("expected viewPaste, got %v", m.view)
	}
}

func TestFilterMode(t *testing.T) {
	cfg := testConfig()
	m := NewModel(cfg, "1.0.0")

	// Press / to enter filter mode
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = newModel.(Model)
	if !m.filtering {
		t.Error("expected filtering to be true")
	}

	// Type "prod"
	for _, r := range "prod" {
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = newModel.(Model)
	}

	// Should filter to 1 result
	if len(m.filtered) != 1 {
		t.Errorf("expected 1 filtered result, got %d", len(m.filtered))
	}

	// Press esc to clear filter
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = newModel.(Model)
	if m.filtering {
		t.Error("expected filtering to be false after esc")
	}
	if len(m.filtered) != 3 {
		t.Errorf("expected 3 filtered results after clear, got %d", len(m.filtered))
	}
}

func TestFuzzyFilterScoring(t *testing.T) {
	// Create config with connections that should have different fuzzy scores
	cfg := &config.Config{
		Version: 1,
		Connections: []config.Connection{
			{ID: "web-prod", Host: "web.prod.example.com", Project: "app", Env: "prod"},
			{ID: "prod-server", Host: "prod.example.com", Project: "app", Env: "prod"},
			{ID: "prod", Host: "prod-main.example.com", Project: "app", Env: "production"},
			{ID: "staging", Host: "staging.example.com", Project: "app", Env: "staging"},
		},
		Groups: map[string][]string{},
	}
	m := NewModel(cfg, "1.0.0")

	// Apply filter "prod"
	m.applyFilter("prod")

	// Should find 3 results (not staging)
	if len(m.filtered) != 3 {
		t.Errorf("expected 3 filtered results, got %d", len(m.filtered))
	}

	// First result should be "prod" (exact match, score 1000)
	if len(m.filtered) > 0 {
		idx := m.filtered[0]
		conn := m.items[idx].connection
		if conn.ID != "prod" {
			t.Errorf("expected first result to be 'prod' (exact match), got %q", conn.ID)
		}
	}

	// Second result should be "prod-server" (prefix match, higher score than substring)
	if len(m.filtered) > 1 {
		idx := m.filtered[1]
		conn := m.items[idx].connection
		if conn.ID != "prod-server" {
			t.Errorf("expected second result to be 'prod-server' (prefix match), got %q", conn.ID)
		}
	}
}

func TestFuzzyFilterSubsequence(t *testing.T) {
	// Test fuzzy subsequence matching
	cfg := &config.Config{
		Version: 1,
		Connections: []config.Connection{
			{ID: "myapp-prod-web", Host: "web.prod.example.com"},
			{ID: "other-server", Host: "other.example.com"},
		},
		Groups: map[string][]string{},
	}
	m := NewModel(cfg, "1.0.0")

	// Apply fuzzy filter "mpw" (m-yapp-p-rod-w-eb)
	m.applyFilter("mpw")

	// Should find 1 result via fuzzy subsequence matching
	if len(m.filtered) != 1 {
		t.Errorf("expected 1 filtered result from fuzzy match, got %d", len(m.filtered))
	}

	if len(m.filtered) > 0 {
		idx := m.filtered[0]
		conn := m.items[idx].connection
		if conn.ID != "myapp-prod-web" {
			t.Errorf("expected fuzzy match 'myapp-prod-web', got %q", conn.ID)
		}
	}
}

func TestTagFilterUI(t *testing.T) {
	// Create config with tagged connections
	cfg := &config.Config{
		Version: 1,
		Connections: []config.Connection{
			{ID: "web1", Host: "web1.example.com", Tags: []string{"web", "prod"}},
			{ID: "web2", Host: "web2.example.com", Tags: []string{"web", "staging"}},
			{ID: "db1", Host: "db1.example.com", Tags: []string{"database", "prod"}},
			{ID: "db2", Host: "db2.example.com", Tags: []string{"database", "staging"}},
		},
		Groups: map[string][]string{},
	}
	m := NewModel(cfg, "1.0.0")

	// Verify tags were collected
	if len(m.allTags) != 4 {
		t.Errorf("expected 4 unique tags, got %d", len(m.allTags))
	}

	// All connections should be visible initially
	if len(m.filtered) != 4 {
		t.Errorf("expected 4 filtered connections initially, got %d", len(m.filtered))
	}

	// Activate "web" tag
	m.activeTags["web"] = true
	m.applyFilter("")

	// Should show only 2 connections with "web" tag
	if len(m.filtered) != 2 {
		t.Errorf("expected 2 filtered connections with 'web' tag, got %d", len(m.filtered))
	}

	// Activate "prod" tag as well (AND logic)
	m.activeTags["prod"] = true
	m.applyFilter("")

	// Should show only 1 connection with both "web" AND "prod" tags
	if len(m.filtered) != 1 {
		t.Errorf("expected 1 filtered connection with 'web' AND 'prod' tags, got %d", len(m.filtered))
	}

	// Clear tags
	m.activeTags = make(map[string]bool)
	m.applyFilter("")

	// All connections visible again
	if len(m.filtered) != 4 {
		t.Errorf("expected 4 filtered connections after clearing tags, got %d", len(m.filtered))
	}
}

func TestTagFilterWithTextFilter(t *testing.T) {
	cfg := &config.Config{
		Version: 1,
		Connections: []config.Connection{
			{ID: "web-prod", Host: "web.prod.example.com", Tags: []string{"web"}},
			{ID: "web-staging", Host: "web.staging.example.com", Tags: []string{"web"}},
			{ID: "api-prod", Host: "api.prod.example.com", Tags: []string{"api"}},
		},
		Groups: map[string][]string{},
	}
	m := NewModel(cfg, "1.0.0")

	// Activate "web" tag and search for "prod"
	m.activeTags["web"] = true
	m.applyFilter("prod")

	// Should show only 1 connection matching both tag and search
	if len(m.filtered) != 1 {
		t.Errorf("expected 1 filtered connection, got %d", len(m.filtered))
	}

	if len(m.filtered) > 0 {
		idx := m.filtered[0]
		conn := m.items[idx].connection
		if conn.ID != "web-prod" {
			t.Errorf("expected 'web-prod', got %q", conn.ID)
		}
	}
}

func TestTagPickerView(t *testing.T) {
	cfg := &config.Config{
		Version: 1,
		Connections: []config.Connection{
			{ID: "server1", Host: "s1.example.com", Tags: []string{"web", "prod"}},
		},
		Groups: map[string][]string{},
	}
	m := NewModel(cfg, "1.0.0")

	// Press t to open tag picker
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	m = newModel.(Model)

	if m.view != viewTagPicker {
		t.Errorf("expected viewTagPicker, got %v", m.view)
	}

	// Render tag picker
	output := m.renderTagPicker()
	if !strings.Contains(output, "web") {
		t.Error("tag picker should show 'web' tag")
	}
	if !strings.Contains(output, "prod") {
		t.Error("tag picker should show 'prod' tag")
	}

	// Press esc to close
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = newModel.(Model)

	if m.view != viewList {
		t.Errorf("expected viewList after esc, got %v", m.view)
	}
}

func TestMouseScrollNavigation(t *testing.T) {
	cfg := testConfig()
	m := NewModel(cfg, "1.0.0")

	// Initial cursor at 0
	if m.cursor != 0 {
		t.Errorf("expected cursor 0, got %d", m.cursor)
	}

	// Scroll down with mouse wheel
	newModel, _ := m.handleMouse(tea.MouseMsg{
		Button: tea.MouseButtonWheelDown,
	})
	m = newModel.(Model)
	if m.cursor != 1 {
		t.Errorf("expected cursor 1 after wheel down, got %d", m.cursor)
	}

	// Scroll up with mouse wheel
	newModel, _ = m.handleMouse(tea.MouseMsg{
		Button: tea.MouseButtonWheelUp,
	})
	m = newModel.(Model)
	if m.cursor != 0 {
		t.Errorf("expected cursor 0 after wheel up, got %d", m.cursor)
	}

	// Scroll up at top should stay at 0
	newModel, _ = m.handleMouse(tea.MouseMsg{
		Button: tea.MouseButtonWheelUp,
	})
	m = newModel.(Model)
	if m.cursor != 0 {
		t.Errorf("expected cursor 0 at top, got %d", m.cursor)
	}
}

func TestMouseClickSelection(t *testing.T) {
	// Use simple config without project/env grouping
	cfg := &config.Config{
		Version: 1,
		Connections: []config.Connection{
			{ID: "server1", Host: "s1.example.com"},
			{ID: "server2", Host: "s2.example.com"},
			{ID: "server3", Host: "s3.example.com"},
		},
		Groups: map[string][]string{},
	}
	m := NewModel(cfg, "1.0.0")
	m.width = 100
	m.height = 30

	// Click on the rendered line containing server1.
	view := m.View()
	lines := strings.Split(view, "\n")
	server1Y := -1
	server2Y := -1
	for i, line := range lines {
		if strings.Contains(line, "server1") {
			server1Y = i
		}
		if strings.Contains(line, "server2") {
			server2Y = i
		}
	}
	if server1Y == -1 || server2Y == -1 {
		t.Fatalf("failed to find server lines in rendered view")
	}

	newModel, _ := m.handleMouse(tea.MouseMsg{
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionRelease,
		Y:      server1Y,
	})
	m = newModel.(Model)

	if m.cursor != 0 {
		t.Errorf("expected cursor 0 after clicking server1, got %d", m.cursor)
	}

	// Click on the rendered line containing server2.
	newModel, _ = m.handleMouse(tea.MouseMsg{
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionRelease,
		Y:      server2Y,
	})
	m = newModel.(Model)

	if m.cursor != 1 {
		t.Errorf("expected cursor 1 after click, got %d", m.cursor)
	}
}

func TestSelectConnection(t *testing.T) {
	cfg := testConfig()
	m := NewModel(cfg, "1.0.0")

	// Move to second item
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = newModel.(Model)

	// Press enter to select
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(Model)

	if m.selected == nil {
		t.Error("expected a connection to be selected")
	}

	// Should quit after selection
	if cmd == nil {
		t.Error("expected quit command")
	}
}

func TestEmptyConfig(t *testing.T) {
	cfg := emptyConfig()
	m := NewModel(cfg, "1.0.0")

	if len(m.filtered) != 0 {
		t.Errorf("expected 0 filtered items, got %d", len(m.filtered))
	}

	// View should still render without panic
	output := m.View()
	if output == "" {
		t.Error("expected non-empty view output")
	}
}

func TestEditConnection(t *testing.T) {
	cfg := testConfig()
	m := NewModel(cfg, "1.0.0")

	// Press e to edit
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m = newModel.(Model)

	if m.view != viewForm {
		t.Errorf("expected viewForm, got %v", m.view)
	}

	if !m.form.IsEditing() {
		t.Error("expected form to be in edit mode")
	}
}

func TestDuplicateConnection(t *testing.T) {
	cfg := testConfig()
	m := NewModel(cfg, "1.0.0")

	// Press c to duplicate
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	m = newModel.(Model)

	if m.view != viewForm {
		t.Errorf("expected viewForm, got %v", m.view)
	}

	if m.form.IsEditing() {
		t.Error("expected form NOT to be in edit mode for duplicate")
	}

	// Form should have host pre-filled from original but empty ID
	// The ID field should be empty (requiring user to fill it)
	if m.form.inputs[fieldID].Value() != "" {
		t.Errorf("expected empty ID for duplicate, got %s", m.form.inputs[fieldID].Value())
	}

	// Host should be copied from original
	if m.form.inputs[fieldHost].Value() == "" {
		t.Error("expected host to be copied from original")
	}
}

func TestQuit(t *testing.T) {
	cfg := testConfig()
	m := NewModel(cfg, "1.0.0")

	// Press q to quit
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = newModel.(Model)

	if !m.quitting {
		t.Error("expected quitting to be true")
	}

	if cmd == nil {
		t.Error("expected quit command")
	}
}

// Integration test using teatest
func TestDashboardIntegration(t *testing.T) {
	cfg := testConfig()
	m := NewModel(cfg, "1.0.0")

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))

	// Navigate down twice
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})

	// Open help
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})

	// Close help
	tm.Send(tea.KeyMsg{Type: tea.KeyEscape})

	// Quit
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	// Wait for program to finish
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))

	finalModel := tm.FinalModel(t).(Model)

	if finalModel.cursor != 2 {
		t.Errorf("expected cursor at 2, got %d", finalModel.cursor)
	}

	if !finalModel.quitting {
		t.Error("expected quitting to be true")
	}
}

func TestParseSSHString(t *testing.T) {
	tests := []struct {
		input    string
		wantHost string
		wantUser string
		wantPort int
	}{
		{"user@host.com", "host.com", "user", 22},
		{"admin@192.168.1.1", "192.168.1.1", "admin", 22},
		{"user@host.com:2222", "host.com", "user", 2222},
		{"host.com", "host.com", "", 22},
		{"host.com:2222", "host.com", "", 2222},
		{"ssh user@host.com", "host.com", "user", 22},
		{"ssh user@host.com -p 2222", "host.com", "user", 2222},
		{"ssh -p 2222 user@host.com", "host.com", "user", 2222},
		{"ssh://user@host.com:2222", "host.com", "user", 2222},
		{"", "", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			conn := ParseSSHString(tt.input)

			if tt.wantHost == "" {
				if conn != nil {
					t.Errorf("expected nil for input %q", tt.input)
				}
				return
			}

			if conn == nil {
				t.Fatalf("expected non-nil connection for input %q", tt.input)
			}

			if conn.Host != tt.wantHost {
				t.Errorf("host: got %q, want %q", conn.Host, tt.wantHost)
			}
			if conn.User != tt.wantUser {
				t.Errorf("user: got %q, want %q", conn.User, tt.wantUser)
			}
			if conn.Port != tt.wantPort {
				t.Errorf("port: got %d, want %d", conn.Port, tt.wantPort)
			}
		})
	}
}
