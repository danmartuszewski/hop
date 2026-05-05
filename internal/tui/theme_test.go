package tui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/danmartuszewski/hop/internal/config"
)

func resetTheme(t *testing.T) {
	t.Helper()
	prev := currentTheme
	prevDark := lipgloss.HasDarkBackground()
	t.Cleanup(func() {
		lipgloss.SetHasDarkBackground(prevDark)
		currentTheme = prev
		refreshStyles()
	})
}

func TestInitTheme_NoConfig_DarkTerminalUsesDefaultDark(t *testing.T) {
	resetTheme(t)
	lipgloss.SetHasDarkBackground(true)

	InitTheme(&config.Config{})

	if currentTheme != defaultTheme {
		t.Fatalf("expected default-dark palette, got %+v", currentTheme)
	}
}

func TestInitTheme_NoConfig_LightTerminalUsesDefaultLight(t *testing.T) {
	resetTheme(t)
	lipgloss.SetHasDarkBackground(false)

	InitTheme(&config.Config{})

	want := FindPreset(DefaultLightPresetName).Theme
	if currentTheme != want {
		t.Errorf("light terminal should auto-pick default-light; got %+v", currentTheme)
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

func TestInitTheme_LightPresetRespectsThemeLight(t *testing.T) {
	resetTheme(t)
	lipgloss.SetHasDarkBackground(true) // light preset overrides terminal detection

	cfg := &config.Config{
		ThemePreset: "everforest-light",
		Theme:       map[string]string{"primary": "99", "error": "9"},
		ThemeLight:  map[string]string{"primary": "21"},
	}
	InitTheme(cfg)

	if got := currentTheme.Primary; got != lipgloss.Color("21") {
		t.Errorf("primary = %q, want %q (theme_light applies to light preset)", got, "21")
	}
	if got := currentTheme.Error; got != lipgloss.Color("9") {
		t.Errorf("error = %q, want %q (inherited from shared theme)", got, "9")
	}
}

func TestInitTheme_DarkPresetIgnoresThemeLight(t *testing.T) {
	resetTheme(t)
	lipgloss.SetHasDarkBackground(false) // light terminal but dark preset selected

	cfg := &config.Config{
		ThemePreset: "everforest-dark",
		ThemeLight:  map[string]string{"primary": "21"},
	}
	InitTheme(cfg)

	want := FindPreset("everforest-dark").Theme.Primary
	if got := currentTheme.Primary; got != want {
		t.Errorf("primary = %q, want %q (theme_light must not apply to dark preset)", got, want)
	}
}

func TestInitTheme_PresetOverridesTerminalDetection(t *testing.T) {
	resetTheme(t)
	lipgloss.SetHasDarkBackground(true) // dark terminal but user picked light

	cfg := &config.Config{ThemePreset: "everforest-light"}
	InitTheme(cfg)

	want := FindPreset("everforest-light").Theme
	if currentTheme != want {
		t.Errorf("explicit light preset should apply on dark terminal; got %+v", currentTheme)
	}
}

func TestInitTheme_UnknownPresetFallsBackToDefaultDark(t *testing.T) {
	resetTheme(t)
	lipgloss.SetHasDarkBackground(true)

	cfg := &config.Config{ThemePreset: "no-such-preset"}
	InitTheme(cfg)

	if currentTheme != defaultTheme {
		t.Errorf("unknown preset should fall back to default-dark, got %+v", currentTheme)
	}
}

func TestInitTheme_UserOverridesLayerOnTopOfPreset(t *testing.T) {
	resetTheme(t)
	lipgloss.SetHasDarkBackground(true)

	cfg := &config.Config{
		ThemePreset: "everforest-dark",
		Theme:       map[string]string{"primary": "1"},
		ThemeDark:   map[string]string{"accent": "2"},
	}
	InitTheme(cfg)

	if got := currentTheme.Primary; got != lipgloss.Color("1") {
		t.Errorf("primary = %q, want shared override %q", got, "1")
	}
	if got := currentTheme.Accent; got != lipgloss.Color("2") {
		t.Errorf("accent = %q, want theme_dark override %q", got, "2")
	}
	want := FindPreset("everforest-dark").Theme.Success
	if got := currentTheme.Success; got != want {
		t.Errorf("success = %q, want everforest %q (no override)", got, want)
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

func TestDefaultDarkPresetMatchesHistoricalPalette(t *testing.T) {
	p := FindPreset(DefaultDarkPresetName)
	if p == nil {
		t.Fatal("default-dark preset must exist")
	}
	if p.Theme != defaultTheme {
		t.Error("default-dark preset must match defaultTheme byte-for-byte")
	}
	if p.IsLight {
		t.Error("default-dark preset must not be marked as light")
	}
}

func TestDefaultLightPresetIsMarkedLight(t *testing.T) {
	p := FindPreset(DefaultLightPresetName)
	if p == nil {
		t.Fatal("default-light preset must exist")
	}
	if !p.IsLight {
		t.Error("default-light preset must be marked as light")
	}
	if p.Theme == defaultTheme {
		t.Error("default-light should differ from default-dark palette")
	}
}

func TestPresets_AllNamesUnique(t *testing.T) {
	seen := make(map[string]bool)
	for _, p := range Presets() {
		if seen[p.Name] {
			t.Errorf("duplicate preset name: %s", p.Name)
		}
		seen[p.Name] = true
	}
}

func TestPresets_LightVariantsClassified(t *testing.T) {
	// Sanity: any preset with "-light" or "latte"/"day"/"alucard" suffix
	// should be marked IsLight; the rest should not.
	wantLight := map[string]bool{
		"default-light":     true,
		"everforest-light":  true,
		"gruvbox-light":     true,
		"catppuccin-latte":  true,
		"tokyo-night-day":   true,
		"solarized-light":   true,
		"nord-light":        true,
		"alucard":           true,
	}
	for _, p := range Presets() {
		if got, want := p.IsLight, wantLight[p.Name]; got != want {
			t.Errorf("preset %q: IsLight = %v, want %v", p.Name, got, want)
		}
	}
}

func TestFindPreset_CaseInsensitive(t *testing.T) {
	if FindPreset("EverForest-Dark") == nil {
		t.Error("FindPreset should be case-insensitive")
	}
	if FindPreset("nope") != nil {
		t.Error("FindPreset should return nil for unknown names")
	}
}
