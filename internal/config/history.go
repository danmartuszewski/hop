package config

import (
	"os"
	"path/filepath"
	"sort"
	"time"

	"gopkg.in/yaml.v3"
)

// History tracks connection usage for recent connections feature
type History struct {
	Entries []HistoryEntry `yaml:"entries"`
}

// HistoryEntry represents a single connection usage record
type HistoryEntry struct {
	ID        string    `yaml:"id"`
	LastUsed  time.Time `yaml:"last_used"`
	UseCount  int       `yaml:"use_count"`
}

// DefaultHistoryPath returns the default path for the history file
func DefaultHistoryPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "hop", "history.yaml")
}

// LoadHistory loads the history from the default path
func LoadHistory() (*History, error) {
	return LoadHistoryFromPath(DefaultHistoryPath())
}

// LoadHistoryFromPath loads history from a specific path
func LoadHistoryFromPath(path string) (*History, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &History{Entries: []HistoryEntry{}}, nil
		}
		return nil, err
	}

	var history History
	if err := yaml.Unmarshal(data, &history); err != nil {
		return nil, err
	}

	return &history, nil
}

// Save saves the history to the default path
func (h *History) Save() error {
	return h.SaveToPath(DefaultHistoryPath())
}

// SaveToPath saves history to a specific path
func (h *History) SaveToPath(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(h)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// RecordUsage records a connection usage
func (h *History) RecordUsage(id string) {
	now := time.Now()

	for i := range h.Entries {
		if h.Entries[i].ID == id {
			h.Entries[i].LastUsed = now
			h.Entries[i].UseCount++
			return
		}
	}

	// New entry
	h.Entries = append(h.Entries, HistoryEntry{
		ID:        id,
		LastUsed:  now,
		UseCount:  1,
	})
}

// GetRecent returns the most recently used connection IDs
func (h *History) GetRecent(limit int) []string {
	// Sort by last used time (most recent first)
	sorted := make([]HistoryEntry, len(h.Entries))
	copy(sorted, h.Entries)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].LastUsed.After(sorted[j].LastUsed)
	})

	result := make([]string, 0, limit)
	for i := 0; i < len(sorted) && i < limit; i++ {
		result = append(result, sorted[i].ID)
	}
	return result
}

// GetMostUsed returns the most frequently used connection IDs
func (h *History) GetMostUsed(limit int) []string {
	sorted := make([]HistoryEntry, len(h.Entries))
	copy(sorted, h.Entries)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].UseCount > sorted[j].UseCount
	})

	result := make([]string, 0, limit)
	for i := 0; i < len(sorted) && i < limit; i++ {
		result = append(result, sorted[i].ID)
	}
	return result
}

// GetLastUsed returns the last used time for a connection
func (h *History) GetLastUsed(id string) (time.Time, bool) {
	for _, e := range h.Entries {
		if e.ID == id {
			return e.LastUsed, true
		}
	}
	return time.Time{}, false
}

// GetUseCount returns the use count for a connection
func (h *History) GetUseCount(id string) int {
	for _, e := range h.Entries {
		if e.ID == id {
			return e.UseCount
		}
	}
	return 0
}

// RemoveEntry removes an entry from history (for when connection is deleted)
func (h *History) RemoveEntry(id string) {
	for i, e := range h.Entries {
		if e.ID == id {
			h.Entries = append(h.Entries[:i], h.Entries[i+1:]...)
			return
		}
	}
}
