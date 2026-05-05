package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Preset is a named palette. Light and dark variants of the same theme are
// modelled as separate Presets (e.g. "everforest-dark" and "everforest-light")
// so users see and pick each one explicitly.
type Preset struct {
	Name        string
	Description string
	IsLight     bool
	Theme       Theme
}

// DefaultDarkPresetName / DefaultLightPresetName back the implicit default
// applied when theme_preset is not set; the choice between them is made by
// looking at the terminal background.
const (
	DefaultDarkPresetName  = "default-dark"
	DefaultLightPresetName = "default-light"
)

// presetList is the ordered registry of bundled presets. Keep dark/light
// pairs adjacent so the picker reads naturally.
var presetList = []Preset{
	{
		Name:        "default-dark",
		Description: "Built-in hop palette tuned for dark terminals",
		Theme:       defaultTheme,
	},
	{
		Name:        "default-light",
		Description: "Built-in hop palette tuned for light terminals",
		IsLight:     true,
		Theme: Theme{
			Primary:    lipgloss.Color("26"),  // DeepSkyBlue4
			Secondary:  lipgloss.Color("244"), // Grey50
			Accent:     lipgloss.Color("162"), // Magenta3
			Success:    lipgloss.Color("28"),  // Green4
			Warning:    lipgloss.Color("130"), // DarkOrange3
			Error:      lipgloss.Color("124"), // Red3
			Muted:      lipgloss.Color("244"),
			Selection:  lipgloss.Color("254"), // Grey89
			Foreground: lipgloss.Color("235"), // Grey15
		},
	},
	{
		Name:        "everforest-dark",
		Description: "Soft warm green-on-dark",
		Theme: Theme{
			Primary:    lipgloss.Color("#7FBBB3"),
			Secondary:  lipgloss.Color("#859289"),
			Accent:     lipgloss.Color("#D699B6"),
			Success:    lipgloss.Color("#A7C080"),
			Warning:    lipgloss.Color("#E69875"),
			Error:      lipgloss.Color("#E67E80"),
			Muted:      lipgloss.Color("#7A8478"),
			Selection:  lipgloss.Color("#3D484D"),
			Foreground: lipgloss.Color("#D3C6AA"),
		},
	},
	{
		Name:        "everforest-light",
		Description: "Soft warm cream-on-light",
		IsLight:     true,
		Theme: Theme{
			Primary:    lipgloss.Color("#3A94C5"),
			Secondary:  lipgloss.Color("#939F91"),
			Accent:     lipgloss.Color("#DF69BA"),
			Success:    lipgloss.Color("#8DA101"),
			Warning:    lipgloss.Color("#F57D26"),
			Error:      lipgloss.Color("#F85552"),
			Muted:      lipgloss.Color("#A6B0A0"),
			Selection:  lipgloss.Color("#EFEBD4"),
			Foreground: lipgloss.Color("#5C6A72"),
		},
	},
	{
		Name:        "gruvbox-dark",
		Description: "Retro warm contrast (dark)",
		Theme: Theme{
			Primary:    lipgloss.Color("#83a598"),
			Secondary:  lipgloss.Color("#928374"),
			Accent:     lipgloss.Color("#d3869b"),
			Success:    lipgloss.Color("#b8bb26"),
			Warning:    lipgloss.Color("#fe8019"),
			Error:      lipgloss.Color("#fb4934"),
			Muted:      lipgloss.Color("#7c6f64"),
			Selection:  lipgloss.Color("#3c3836"),
			Foreground: lipgloss.Color("#ebdbb2"),
		},
	},
	{
		Name:        "gruvbox-light",
		Description: "Retro warm contrast (light)",
		IsLight:     true,
		Theme: Theme{
			Primary:    lipgloss.Color("#076678"),
			Secondary:  lipgloss.Color("#7c6f64"),
			Accent:     lipgloss.Color("#8f3f71"),
			Success:    lipgloss.Color("#79740e"),
			Warning:    lipgloss.Color("#af3a03"),
			Error:      lipgloss.Color("#9d0006"),
			Muted:      lipgloss.Color("#928374"),
			Selection:  lipgloss.Color("#ebdbb2"),
			Foreground: lipgloss.Color("#3c3836"),
		},
	},
	{
		Name:        "catppuccin-mocha",
		Description: "Catppuccin Mocha (dark pastel)",
		Theme: Theme{
			Primary:    lipgloss.Color("#89b4fa"),
			Secondary:  lipgloss.Color("#6c7086"),
			Accent:     lipgloss.Color("#f5c2e7"),
			Success:    lipgloss.Color("#a6e3a1"),
			Warning:    lipgloss.Color("#fab387"),
			Error:      lipgloss.Color("#f38ba8"),
			Muted:      lipgloss.Color("#585b70"),
			Selection:  lipgloss.Color("#313244"),
			Foreground: lipgloss.Color("#cdd6f4"),
		},
	},
	{
		Name:        "catppuccin-latte",
		Description: "Catppuccin Latte (light pastel)",
		IsLight:     true,
		Theme: Theme{
			Primary:    lipgloss.Color("#1e66f5"),
			Secondary:  lipgloss.Color("#8c8fa1"),
			Accent:     lipgloss.Color("#ea76cb"),
			Success:    lipgloss.Color("#40a02b"),
			Warning:    lipgloss.Color("#fe640b"),
			Error:      lipgloss.Color("#d20f39"),
			Muted:      lipgloss.Color("#acb0be"),
			Selection:  lipgloss.Color("#ccd0da"),
			Foreground: lipgloss.Color("#4c4f69"),
		},
	},
	{
		Name:        "tokyo-night-storm",
		Description: "Tokyo Night Storm (dark)",
		Theme: Theme{
			Primary:    lipgloss.Color("#7aa2f7"),
			Secondary:  lipgloss.Color("#565f89"),
			Accent:     lipgloss.Color("#bb9af7"),
			Success:    lipgloss.Color("#9ece6a"),
			Warning:    lipgloss.Color("#ff9e64"),
			Error:      lipgloss.Color("#f7768e"),
			Muted:      lipgloss.Color("#565f89"),
			Selection:  lipgloss.Color("#2e3354"),
			Foreground: lipgloss.Color("#c0caf5"),
		},
	},
	{
		Name:        "tokyo-night-day",
		Description: "Tokyo Night Day (light)",
		IsLight:     true,
		Theme: Theme{
			Primary:    lipgloss.Color("#2e7de9"),
			Secondary:  lipgloss.Color("#6172b0"),
			Accent:     lipgloss.Color("#9854f1"),
			Success:    lipgloss.Color("#587539"),
			Warning:    lipgloss.Color("#b15c00"),
			Error:      lipgloss.Color("#f52a65"),
			Muted:      lipgloss.Color("#848cb5"),
			Selection:  lipgloss.Color("#c4c8da"),
			Foreground: lipgloss.Color("#3760bf"),
		},
	},
	{
		Name:        "solarized-dark",
		Description: "Ethan Schoonover's classic (dark)",
		Theme: Theme{
			Primary:    lipgloss.Color("#268bd2"),
			Secondary:  lipgloss.Color("#586e75"),
			Accent:     lipgloss.Color("#d33682"),
			Success:    lipgloss.Color("#859900"),
			Warning:    lipgloss.Color("#cb4b16"),
			Error:      lipgloss.Color("#dc322f"),
			Muted:      lipgloss.Color("#657b83"),
			Selection:  lipgloss.Color("#073642"),
			Foreground: lipgloss.Color("#93a1a1"),
		},
	},
	{
		Name:        "solarized-light",
		Description: "Ethan Schoonover's classic (light)",
		IsLight:     true,
		Theme: Theme{
			Primary:    lipgloss.Color("#268bd2"),
			Secondary:  lipgloss.Color("#93a1a1"),
			Accent:     lipgloss.Color("#d33682"),
			Success:    lipgloss.Color("#859900"),
			Warning:    lipgloss.Color("#cb4b16"),
			Error:      lipgloss.Color("#dc322f"),
			Muted:      lipgloss.Color("#839496"),
			Selection:  lipgloss.Color("#eee8d5"),
			Foreground: lipgloss.Color("#586e75"),
		},
	},
	{
		Name:        "nord",
		Description: "Arctic, north-bluish (dark)",
		Theme: Theme{
			Primary:    lipgloss.Color("#88c0d0"),
			Secondary:  lipgloss.Color("#4c566a"),
			Accent:     lipgloss.Color("#b48ead"),
			Success:    lipgloss.Color("#a3be8c"),
			Warning:    lipgloss.Color("#d08770"),
			Error:      lipgloss.Color("#bf616a"),
			Muted:      lipgloss.Color("#4c566a"),
			Selection:  lipgloss.Color("#3b4252"),
			Foreground: lipgloss.Color("#eceff4"),
		},
	},
	{
		Name:        "nord-light",
		Description: "Arctic snow-storm (light)",
		IsLight:     true,
		Theme: Theme{
			Primary:    lipgloss.Color("#5e81ac"),
			Secondary:  lipgloss.Color("#4c566a"),
			Accent:     lipgloss.Color("#b48ead"),
			Success:    lipgloss.Color("#a3be8c"),
			Warning:    lipgloss.Color("#d08770"),
			Error:      lipgloss.Color("#bf616a"),
			Muted:      lipgloss.Color("#4c566a"),
			Selection:  lipgloss.Color("#e5e9f0"),
			Foreground: lipgloss.Color("#2e3440"),
		},
	},
	{
		Name:        "dracula",
		Description: "High-contrast dark",
		Theme: Theme{
			Primary:    lipgloss.Color("#8be9fd"),
			Secondary:  lipgloss.Color("#6272a4"),
			Accent:     lipgloss.Color("#ff79c6"),
			Success:    lipgloss.Color("#50fa7b"),
			Warning:    lipgloss.Color("#ffb86c"),
			Error:      lipgloss.Color("#ff5555"),
			Muted:      lipgloss.Color("#6272a4"),
			Selection:  lipgloss.Color("#44475a"),
			Foreground: lipgloss.Color("#f8f8f2"),
		},
	},
	{
		Name:        "alucard",
		Description: "Dracula's light counterpart",
		IsLight:     true,
		Theme: Theme{
			Primary:    lipgloss.Color("#0184bc"),
			Secondary:  lipgloss.Color("#6272a4"),
			Accent:     lipgloss.Color("#a626a4"),
			Success:    lipgloss.Color("#50a14f"),
			Warning:    lipgloss.Color("#c18401"),
			Error:      lipgloss.Color("#e45649"),
			Muted:      lipgloss.Color("#a0a1a7"),
			Selection:  lipgloss.Color("#e5e5e6"),
			Foreground: lipgloss.Color("#383a42"),
		},
	},
}

// Presets returns the bundled preset list in display order. Callers must not
// mutate the returned slice.
func Presets() []Preset {
	return presetList
}

// FindPreset looks up a preset by case-insensitive name. Returns nil if no
// preset matches.
func FindPreset(name string) *Preset {
	for i := range presetList {
		if strings.EqualFold(presetList[i].Name, name) {
			return &presetList[i]
		}
	}
	return nil
}
