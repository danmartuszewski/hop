package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/danmartuszewski/hop/internal/config"
)

func TestThemePicker_NavigatePreviewsLiveTheme(t *testing.T) {
	resetTheme(t)
	lipgloss.SetHasDarkBackground(true)
	cfg := &config.Config{}
	InitTheme(cfg)
	original := currentTheme

	m := NewThemePickerModel(cfg, 80, 24)
	if currentTheme != original {
		t.Fatalf("opening picker on the current preset must not change theme")
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if currentTheme == original {
		t.Errorf("navigating down should preview a different preset, but theme was unchanged")
	}
	if m.SelectedPreset() == DefaultDarkPresetName {
		t.Errorf("cursor should have advanced past default; got %q", m.SelectedPreset())
	}
}

func TestThemePicker_EscRestoresOriginalTheme(t *testing.T) {
	resetTheme(t)
	lipgloss.SetHasDarkBackground(true)
	cfg := &config.Config{}
	InitTheme(cfg)
	original := currentTheme

	m := NewThemePickerModel(cfg, 80, 24)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if currentTheme == original {
		t.Fatalf("preview should have changed theme before esc")
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if !m.Cancelled() {
		t.Errorf("esc should mark picker cancelled")
	}
	if currentTheme != original {
		t.Errorf("esc should restore original theme; got %+v want %+v", currentTheme, original)
	}
}

func TestThemePicker_EnterMarksConfirmedAtSelectedPreset(t *testing.T) {
	resetTheme(t)
	lipgloss.SetHasDarkBackground(true)
	cfg := &config.Config{}
	InitTheme(cfg)

	m := NewThemePickerModel(cfg, 80, 24)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	want := m.SelectedPreset()

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !m.Confirmed() {
		t.Errorf("enter should mark picker confirmed")
	}
	if got := m.SelectedPreset(); got != want {
		t.Errorf("SelectedPreset = %q, want %q", got, want)
	}
}

func TestThemePicker_StartsOnConfiguredPreset(t *testing.T) {
	resetTheme(t)
	lipgloss.SetHasDarkBackground(true)

	cfg := &config.Config{ThemePreset: "everforest-light"}
	InitTheme(cfg)
	m := NewThemePickerModel(cfg, 80, 24)

	if got := m.SelectedPreset(); got != "everforest-light" {
		t.Errorf("picker should open on configured preset; got %q", got)
	}
}

func TestThemePicker_BoundaryNavigation(t *testing.T) {
	resetTheme(t)
	lipgloss.SetHasDarkBackground(true)
	cfg := &config.Config{}
	InitTheme(cfg)

	m := NewThemePickerModel(cfg, 80, 24)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp}) // already at top
	if got := m.SelectedPreset(); got != DefaultDarkPresetName {
		t.Errorf("up at top should stay on default; got %q", got)
	}

	last := Presets()[len(Presets())-1].Name
	m, _ = m.Update(tea.KeyMsg{Runes: []rune{'G'}, Type: tea.KeyRunes})
	if got := m.SelectedPreset(); got != last {
		t.Errorf("G should jump to last preset; got %q want %q", got, last)
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown}) // already at bottom
	if got := m.SelectedPreset(); got != last {
		t.Errorf("down at bottom should stay on last; got %q", got)
	}
}
