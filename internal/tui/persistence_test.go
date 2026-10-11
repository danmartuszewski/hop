package tui

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/danmartuszewski/hop/internal/config"
)

func TestDashboardSavesSelectedConfig(t *testing.T) {
	for _, explicit := range []bool{false, true} {
		name := "environment"
		if explicit {
			name = "explicit"
		}
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			defaultPath := filepath.Join(dir, "default.yaml")
			t.Setenv("HOP_CONFIG", defaultPath)
			selectedPath := defaultPath
			constructorPath := ""
			if explicit {
				selectedPath = filepath.Join(dir, "selected.yaml")
				constructorPath = selectedPath
				if err := os.WriteFile(defaultPath, []byte("keep this untouched\n"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			cfg := testConfig()
			if err := cfg.Save(selectedPath); err != nil {
				t.Fatal(err)
			}
			m := NewModel(cfg, "test", constructorPath)
			m.view = viewConfirmDelete
			m.deleteTarget = cfg.FindConnection("dev-server")
			updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
			m = updated.(Model)
			if m.statusMsg != "Deleted: dev-server" {
				t.Fatalf("unexpected status: %s", m.statusMsg)
			}
			saved, err := config.Load(selectedPath)
			if err != nil {
				t.Fatal(err)
			}
			if saved.FindConnection("dev-server") != nil {
				t.Fatal("deletion was not saved to the selected config")
			}
			if explicit {
				data, err := os.ReadFile(defaultPath)
				if err != nil || string(data) != "keep this untouched\n" {
					t.Fatalf("default config was modified: %q, %v", data, err)
				}
			}
		})
	}
}

func TestEditRejectsDuplicateID(t *testing.T) {
	cfg := testConfig()
	m := NewModel(cfg, "test", filepath.Join(t.TempDir(), "config.yaml"))
	m.form = NewFormModel("Edit Connection", cfg.FindConnection("dev-server"))
	m.view = viewForm
	m.form.inputs[fieldID].SetValue("prod-server")
	m.form.inputs[fieldHost].SetValue("changed.example.com")
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	m = updated.(Model)
	if m.view != viewForm || m.form.errMsg == "" || m.form.Submitted() {
		t.Fatal("expected editable form with duplicate-ID error")
	}
	if m.form.inputs[fieldHost].Value() != "changed.example.com" {
		t.Fatal("form input was lost")
	}
	if cfg.FindConnection("dev-server") == nil || cfg.FindConnection("prod-server").Host != "prod.example.com" {
		t.Fatal("rejected edit modified config")
	}
}
