package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestHistory_RecordUsage(t *testing.T) {
	h := &History{}

	// First usage
	h.RecordUsage("server1")
	if len(h.Entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(h.Entries))
	}
	if h.Entries[0].UseCount != 1 {
		t.Errorf("expected use count 1, got %d", h.Entries[0].UseCount)
	}

	// Second usage of same server
	h.RecordUsage("server1")
	if len(h.Entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(h.Entries))
	}
	if h.Entries[0].UseCount != 2 {
		t.Errorf("expected use count 2, got %d", h.Entries[0].UseCount)
	}

	// Usage of different server
	h.RecordUsage("server2")
	if len(h.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(h.Entries))
	}
}

func TestHistory_GetRecent(t *testing.T) {
	h := &History{}

	// Add entries with different timestamps
	h.Entries = []HistoryEntry{
		{ID: "oldest", LastUsed: time.Now().Add(-3 * time.Hour), UseCount: 1},
		{ID: "newest", LastUsed: time.Now(), UseCount: 1},
		{ID: "middle", LastUsed: time.Now().Add(-1 * time.Hour), UseCount: 1},
	}

	recent := h.GetRecent(2)
	if len(recent) != 2 {
		t.Errorf("expected 2 recent, got %d", len(recent))
	}
	if recent[0] != "newest" {
		t.Errorf("expected newest first, got %s", recent[0])
	}
	if recent[1] != "middle" {
		t.Errorf("expected middle second, got %s", recent[1])
	}
}

func TestHistory_GetMostUsed(t *testing.T) {
	h := &History{
		Entries: []HistoryEntry{
			{ID: "low", UseCount: 1},
			{ID: "high", UseCount: 10},
			{ID: "medium", UseCount: 5},
		},
	}

	mostUsed := h.GetMostUsed(2)
	if len(mostUsed) != 2 {
		t.Errorf("expected 2 most used, got %d", len(mostUsed))
	}
	if mostUsed[0] != "high" {
		t.Errorf("expected high first, got %s", mostUsed[0])
	}
	if mostUsed[1] != "medium" {
		t.Errorf("expected medium second, got %s", mostUsed[1])
	}
}

func TestHistory_GetLastUsed(t *testing.T) {
	now := time.Now()
	h := &History{
		Entries: []HistoryEntry{
			{ID: "server1", LastUsed: now, UseCount: 1},
		},
	}

	lastUsed, ok := h.GetLastUsed("server1")
	if !ok {
		t.Error("expected to find server1")
	}
	if !lastUsed.Equal(now) {
		t.Errorf("expected %v, got %v", now, lastUsed)
	}

	_, ok = h.GetLastUsed("nonexistent")
	if ok {
		t.Error("expected not to find nonexistent")
	}
}

func TestHistory_RemoveEntry(t *testing.T) {
	h := &History{
		Entries: []HistoryEntry{
			{ID: "keep", UseCount: 1},
			{ID: "remove", UseCount: 1},
		},
	}

	h.RemoveEntry("remove")
	if len(h.Entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(h.Entries))
	}
	if h.Entries[0].ID != "keep" {
		t.Errorf("expected 'keep', got %s", h.Entries[0].ID)
	}
}

func TestHistory_SaveAndLoad(t *testing.T) {
	tempDir := t.TempDir()
	historyPath := filepath.Join(tempDir, "history.yaml")

	// Create and save history
	h := &History{}
	h.RecordUsage("server1")
	h.RecordUsage("server2")

	if err := h.SaveToPath(historyPath); err != nil {
		t.Fatalf("failed to save history: %v", err)
	}

	// Load history
	loaded, err := LoadHistoryFromPath(historyPath)
	if err != nil {
		t.Fatalf("failed to load history: %v", err)
	}

	if len(loaded.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(loaded.Entries))
	}
}

func TestHistory_LoadNonexistent(t *testing.T) {
	tempDir := t.TempDir()
	historyPath := filepath.Join(tempDir, "nonexistent", "history.yaml")

	h, err := LoadHistoryFromPath(historyPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h == nil {
		t.Fatal("expected non-nil history")
	}
	if len(h.Entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(h.Entries))
	}
}

func TestHistory_SaveCreatesDirectory(t *testing.T) {
	tempDir := t.TempDir()
	historyPath := filepath.Join(tempDir, "subdir", "history.yaml")

	h := &History{}
	h.RecordUsage("server1")

	if err := h.SaveToPath(historyPath); err != nil {
		t.Fatalf("failed to save history: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(historyPath); os.IsNotExist(err) {
		t.Error("expected history file to exist")
	}
}
