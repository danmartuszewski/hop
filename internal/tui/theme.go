package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/danmartuszewski/hop/internal/config"
)

// Theme holds the colors used by the TUI styles. Field names map to the
// keys accepted under theme / theme_dark / theme_light in config.yaml.
type Theme struct {
	Primary    lipgloss.Color
	Secondary  lipgloss.Color
	Accent     lipgloss.Color
	Success    lipgloss.Color
	Warning    lipgloss.Color
	Error      lipgloss.Color
	Muted      lipgloss.Color
	Selection  lipgloss.Color
	Foreground lipgloss.Color
}

// defaultTheme preserves the original palette from styles.go and backs the
// "default-dark" preset. It is also the final fallback when both the
// configured preset and the auto-detected default are missing.
var defaultTheme = Theme{
	Primary:    lipgloss.Color("39"),  // Cyan
	Secondary:  lipgloss.Color("245"), // Gray
	Accent:     lipgloss.Color("212"), // Pink
	Success:    lipgloss.Color("82"),  // Green
	Warning:    lipgloss.Color("214"), // Orange
	Error:      lipgloss.Color("196"), // Red
	Muted:      lipgloss.Color("240"), // Dark gray
	Selection:  lipgloss.Color("236"), // Near-black gray
	Foreground: lipgloss.Color("255"), // Near-white
}

var currentTheme = defaultTheme

// InitTheme picks a palette based on the configured preset and any user
// overrides. Precedence (low to high):
//
//  1. Selected preset's palette
//  2. cfg.Theme (shared overrides)
//  3. cfg.ThemeDark or cfg.ThemeLight (only the one matching the preset's
//     light/dark classification)
//
// When theme_preset is unset, the default is auto-selected based on the
// terminal background (default-dark or default-light). Unknown preset names
// fall back to default-dark.
func InitTheme(cfg *config.Config) {
	currentTheme = resolveTheme(cfg)
	refreshStyles()
}

// resolveTheme builds the Theme that InitTheme would apply, without mutating
// global state. Exported for testing and the live-preview picker.
func resolveTheme(cfg *config.Config) Theme {
	preset := selectPreset(cfg.ThemePreset)
	t := preset.Theme
	t = applyOverrides(t, cfg.Theme)
	if preset.IsLight {
		t = applyOverrides(t, cfg.ThemeLight)
	} else {
		t = applyOverrides(t, cfg.ThemeDark)
	}
	return t
}

// selectPreset returns the preset matching name. If name is empty, the
// default preset for the current terminal background is returned. Unknown
// names also fall back to the default-dark preset.
func selectPreset(name string) *Preset {
	if name == "" {
		if lipgloss.HasDarkBackground() {
			if p := FindPreset(DefaultDarkPresetName); p != nil {
				return p
			}
		} else {
			if p := FindPreset(DefaultLightPresetName); p != nil {
				return p
			}
		}
	}
	if p := FindPreset(name); p != nil {
		return p
	}
	if p := FindPreset(DefaultDarkPresetName); p != nil {
		return p
	}
	// Should not happen — presetList always contains default-dark — but
	// keep the binary working if someone strips the registry.
	return &Preset{Name: DefaultDarkPresetName, Theme: defaultTheme}
}

// previewPreset temporarily swaps the active theme to show how a preset would
// render with the user's overrides on top. Used by the theme picker for live
// preview; revert by calling restoreTheme with the snapshot taken before the
// first preview call.
func previewPreset(cfg *config.Config, presetName string) {
	cfgCopy := *cfg
	cfgCopy.ThemePreset = presetName
	currentTheme = resolveTheme(&cfgCopy)
	refreshStyles()
}

// restoreTheme replaces the active theme with a snapshot and refreshes styles.
func restoreTheme(snapshot Theme) {
	currentTheme = snapshot
	refreshStyles()
}

// applyOverrides patches the given Theme with any non-empty values from
// the provided color map. Unknown keys and empty values are ignored.
func applyOverrides(t Theme, m map[string]string) Theme {
	for key, val := range m {
		if val == "" {
			continue
		}
		c := lipgloss.Color(val)
		switch key {
		case "primary":
			t.Primary = c
		case "secondary":
			t.Secondary = c
		case "accent":
			t.Accent = c
		case "success":
			t.Success = c
		case "warning":
			t.Warning = c
		case "error":
			t.Error = c
		case "muted":
			t.Muted = c
		case "selection":
			t.Selection = c
		case "foreground":
			t.Foreground = c
		}
	}
	return t
}
