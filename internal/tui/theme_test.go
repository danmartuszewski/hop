package tui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/danmartuszewski/hop/internal/config"
)

func resetTheme(t *testing.T) {
	t.Helper()
	prev := currentTheme
	t.Cleanup(func() {
		currentTheme = prev
		refreshStyles()
	})
}

func TestInitTheme_NoConfig_KeepsDefaults(t *testing.T) {
	resetTheme(t)
	lipgloss.SetHasDarkBackground(true)

	InitTheme(&config.Config{})

	if currentTheme != defaultTheme {
		t.Fatalf("expected defaults, got %+v", currentTheme)
	}
}

func TestInitTheme_SharedThemeAppliesToBothModes(t *testing.T) {
	resetTheme(t)
	cfg := &config.Config{
		Theme: map[string]string{"primary": "99"},
	}

	lipgloss.SetHasDarkBackground(true)
	InitTheme(cfg)
	if got := currentTheme.Primary; got != lipgloss.Color("99") {
		t.Errorf("dark: primary = %q, want %q", got, "99")
	}

	lipgloss.SetHasDarkBackground(false)
	InitTheme(cfg)
	if got := currentTheme.Primary; got != lipgloss.Color("99") {
		t.Errorf("light: primary = %q, want %q", got, "99")
	}
}

func TestInitTheme_DarkOverridesShared(t *testing.T) {
	resetTheme(t)
	lipgloss.SetHasDarkBackground(true)

	cfg := &config.Config{
		Theme:     map[string]string{"primary": "99", "error": "9"},
		ThemeDark: map[string]string{"primary": "200"},
	}
	InitTheme(cfg)

	if got := currentTheme.Primary; got != lipgloss.Color("200") {
		t.Errorf("primary = %q, want %q (theme_dark wins)", got, "200")
	}
	if got := currentTheme.Error; got != lipgloss.Color("9") {
		t.Errorf("error = %q, want %q (inherited from theme)", got, "9")
	}
	if got := currentTheme.Secondary; got != defaultTheme.Secondary {
		t.Errorf("secondary = %q, want default %q", got, defaultTheme.Secondary)
	}
}

func TestInitTheme_LightOverridesShared(t *testing.T) {
	resetTheme(t)
	lipgloss.SetHasDarkBackground(false)

	cfg := &config.Config{
		Theme:      map[string]string{"primary": "99", "error": "9"},
		ThemeLight: map[string]string{"primary": "21"},
	}
	InitTheme(cfg)

	if got := currentTheme.Primary; got != lipgloss.Color("21") {
		t.Errorf("primary = %q, want %q (theme_light wins)", got, "21")
	}
	if got := currentTheme.Error; got != lipgloss.Color("9") {
		t.Errorf("error = %q, want %q (inherited from theme)", got, "9")
	}
}

func TestInitTheme_EmptyMapDoesNotBreakFallback(t *testing.T) {
	resetTheme(t)
	lipgloss.SetHasDarkBackground(true)

	cfg := &config.Config{
		Theme:     map[string]string{"primary": "99"},
		ThemeDark: map[string]string{},
	}
	InitTheme(cfg)

	if got := currentTheme.Primary; got != lipgloss.Color("99") {
		t.Errorf("primary = %q, want %q (empty theme_dark must not block theme)", got, "99")
	}
}

func TestInitTheme_OppositeModeIsIgnored(t *testing.T) {
	resetTheme(t)
	lipgloss.SetHasDarkBackground(true)

	cfg := &config.Config{
		ThemeLight: map[string]string{"primary": "21"},
	}
	InitTheme(cfg)

	if got := currentTheme.Primary; got != defaultTheme.Primary {
		t.Errorf("primary = %q, want default %q (theme_light must not apply on dark)", got, defaultTheme.Primary)
	}
}

func TestApplyOverrides_AllKeys(t *testing.T) {
	m := map[string]string{
		"primary":    "1",
		"secondary":  "2",
		"accent":     "3",
		"success":    "4",
		"warning":    "5",
		"error":      "6",
		"muted":      "7",
		"selection":  "8",
		"foreground": "9",
	}
	got := applyOverrides(defaultTheme, m)
	want := Theme{
		Primary:    lipgloss.Color("1"),
		Secondary:  lipgloss.Color("2"),
		Accent:     lipgloss.Color("3"),
		Success:    lipgloss.Color("4"),
		Warning:    lipgloss.Color("5"),
		Error:      lipgloss.Color("6"),
		Muted:      lipgloss.Color("7"),
		Selection:  lipgloss.Color("8"),
		Foreground: lipgloss.Color("9"),
	}
	if got != want {
		t.Errorf("applyOverrides:\n got  %+v\n want %+v", got, want)
	}
}

func TestApplyOverrides_UnknownAndEmptyKeysIgnored(t *testing.T) {
	m := map[string]string{
		"primary": "",
		"unknown": "999",
	}
	got := applyOverrides(defaultTheme, m)
	if got != defaultTheme {
		t.Errorf("expected unchanged defaults, got %+v", got)
	}
}
