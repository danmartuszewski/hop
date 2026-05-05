package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/danmartuszewski/hop/internal/config"
)

// ThemePickerModel is the modal view for browsing and selecting bundled
// presets. Navigation triggers a live preview by mutating currentTheme; the
// snapshot taken at construction time is restored on cancel or after the
// caller has consumed the model.
type ThemePickerModel struct {
	cfg      *config.Config
	cursor   int
	width    int
	height   int
	original Theme

	// terminal state at end of life
	cancelled bool
	confirmed bool
}

// NewThemePickerModel snapshots the active theme and positions the cursor on
// the currently configured preset (or the auto-selected default if unset).
func NewThemePickerModel(cfg *config.Config, width, height int) ThemePickerModel {
	cursor := 0
	current := selectPreset(cfg.ThemePreset).Name
	for i, p := range Presets() {
		if strings.EqualFold(p.Name, current) {
			cursor = i
			break
		}
	}

	m := ThemePickerModel{
		cfg:      cfg,
		cursor:   cursor,
		width:    width,
		height:   height,
		original: currentTheme,
	}
	m.preview()
	return m
}

func (m ThemePickerModel) Init() tea.Cmd {
	return nil
}

func (m ThemePickerModel) Update(msg tea.Msg) (ThemePickerModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			restoreTheme(m.original)
			m.cancelled = true
			return m, nil
		case "enter":
			m.confirmed = true
			return m, nil
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.preview()
			}
		case "down", "j":
			if m.cursor < len(Presets())-1 {
				m.cursor++
				m.preview()
			}
		case "home", "g":
			m.cursor = 0
			m.preview()
		case "end", "G":
			m.cursor = len(Presets()) - 1
			m.preview()
		}
	}
	return m, nil
}

// preview applies the currently highlighted preset to the live theme so the
// preview pane below the list reflects the selection.
func (m ThemePickerModel) preview() {
	presets := Presets()
	if m.cursor < 0 || m.cursor >= len(presets) {
		return
	}
	previewPreset(m.cfg, presets[m.cursor].Name)
}

// SelectedPreset returns the name of the preset under the cursor.
func (m ThemePickerModel) SelectedPreset() string {
	presets := Presets()
	if m.cursor < 0 || m.cursor >= len(presets) {
		return DefaultDarkPresetName
	}
	return presets[m.cursor].Name
}

// Cancelled reports whether the user pressed esc.
func (m ThemePickerModel) Cancelled() bool { return m.cancelled }

// Confirmed reports whether the user pressed enter.
func (m ThemePickerModel) Confirmed() bool { return m.confirmed }

// Original returns the theme snapshot taken when the picker opened. Callers
// use it to restore the theme if persisting the choice fails.
func (m ThemePickerModel) Original() Theme { return m.original }

func (m ThemePickerModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Theme"))
	b.WriteString("\n")
	b.WriteString(helpDescStyle.Render("Pick a preset — colors update live as you navigate."))
	b.WriteString("\n\n")

	presets := Presets()
	currentName := currentPresetName(m.cfg)

	// Compute a scroll window so the cursor stays visible when the preset
	// list is taller than the available space. Reserve room for the title,
	// help line, divider, preview block, and footer.
	visibleRows := m.height - 14
	if visibleRows < 4 {
		visibleRows = len(presets)
	}
	if visibleRows > len(presets) {
		visibleRows = len(presets)
	}
	start := 0
	if m.cursor >= visibleRows {
		start = m.cursor - visibleRows + 1
	}
	if start+visibleRows > len(presets) {
		start = len(presets) - visibleRows
	}
	end := start + visibleRows

	for i := start; i < end; i++ {
		p := presets[i]
		marker := "  "
		nameStyle := itemStyle
		if i == m.cursor {
			marker = "> "
			nameStyle = selectedItemStyle
		}

		variant := envStyle.Render("dark")
		if p.IsLight {
			variant = warningStyle.Render("light")
		}

		name := p.Name
		if strings.EqualFold(name, currentName) {
			name += " (current)"
		}
		b.WriteString(marker)
		b.WriteString(nameStyle.Render(name))
		b.WriteString("  ")
		b.WriteString(variant)
		if p.Description != "" {
			b.WriteString("  ")
			b.WriteString(helpDescStyle.Render(p.Description))
		}
		b.WriteString("\n")
	}

	if start > 0 {
		b.WriteString(helpDescStyle.Render("  ↑ more above"))
		b.WriteString("\n")
	}
	if end < len(presets) {
		b.WriteString(helpDescStyle.Render(fmt.Sprintf("  ↓ %d more below", len(presets)-end)))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpDescStyle.Render(strings.Repeat("─", 40)))
	b.WriteString("\n")
	b.WriteString(renderThemePreview())
	b.WriteString("\n\n")

	help := helpKeyStyle.Render("↑/↓") + " " + helpDescStyle.Render("navigate") + "  "
	help += helpKeyStyle.Render("enter") + " " + helpDescStyle.Render("save") + "  "
	help += helpKeyStyle.Render("esc") + " " + helpDescStyle.Render("cancel")
	b.WriteString(help)

	return b.String()
}

// currentPresetName resolves what the picker should label as "(current)" —
// the preset persisted in cfg, or the auto-selected default when unset.
func currentPresetName(cfg *config.Config) string {
	if cfg == nil {
		return DefaultDarkPresetName
	}
	return selectPreset(cfg.ThemePreset).Name
}

// renderThemePreview shows a small mock-up of dashboard elements using the
// currently active styles, so the user can judge how the highlighted preset
// will look. Updates are implicit: refreshStyles() ran when the selection
// changed.
func renderThemePreview() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render("hop  preview"))
	b.WriteString("\n")
	b.WriteString(selectedItemStyle.Render("> "))
	b.WriteString(projectStyle.Render("project"))
	b.WriteString(" ")
	b.WriteString(envStyle.Render("prod"))
	b.WriteString("  ")
	b.WriteString(userStyle.Render("deploy"))
	b.WriteString(hostStyle.Render("@web1.example.com"))
	b.WriteString(portStyle.Render(":22"))
	b.WriteString("\n")

	b.WriteString("  ")
	b.WriteString(itemStyle.Render(""))
	b.WriteString(panelTagStyle.Render("nginx"))
	b.WriteString(" ")
	b.WriteString(panelTagStyle.Render("web"))
	b.WriteString("   ")
	b.WriteString(healthReachableStyle.Render("● healthy"))
	b.WriteString("   ")
	b.WriteString(healthUnreachableStyle.Render("● unreachable"))
	b.WriteString("\n")

	b.WriteString(footerStyle.Render(fmt.Sprintf("%s  %s",
		helpKeyStyle.Render("?"),
		helpDescStyle.Render("help"),
	)))
	return b.String()
}
