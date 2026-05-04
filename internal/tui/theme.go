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

// defaultTheme preserves the original palette from styles.go.
// It is used whenever no matching theme override is configured.
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

// InitTheme picks a palette based on the terminal background and the
// theme / theme_dark / theme_light maps in the config. theme is the
// shared base; theme_dark and theme_light layer on top of it for their
// respective backgrounds. Each layer only overrides the keys it
// defines; missing keys fall through to the layer below, ending at the
// built-in default palette.
func InitTheme(cfg *config.Config) {
	shared, dark, light := cfg.Theme, cfg.ThemeDark, cfg.ThemeLight

	t := defaultTheme
	if lipgloss.HasDarkBackground() {
		t = applyOverrides(t, shared)
		t = applyOverrides(t, dark)
	} else {
		t = applyOverrides(t, shared)
		t = applyOverrides(t, light)
	}

	currentTheme = t
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
