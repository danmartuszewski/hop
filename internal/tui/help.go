package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type HelpModel struct {
	width  int
	height int
}

func NewHelpModel() HelpModel {
	return HelpModel{}
}

func (m HelpModel) Init() tea.Cmd {
	return nil
}

func (m HelpModel) Update(msg tea.Msg) (HelpModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m HelpModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Keyboard Shortcuts"))
	b.WriteString("\n\n")

	sections := []struct {
		title string
		keys  [][]string
	}{
		{
			title: "Navigation",
			keys: [][]string{
				{"↑/k", "Move up"},
				{"↓/j", "Move down"},
				{"g", "Go to top"},
				{"G", "Go to bottom"},
				{"/", "Filter connections"},
				{"esc", "Clear filter"},
			},
		},
		{
			title: "Actions",
			keys: [][]string{
				{"enter", "Connect to selected"},
				{"a", "Add new connection"},
				{"p", "Paste SSH string (quick add)"},
				{"e", "Edit selected connection"},
				{"c", "Duplicate selected connection"},
				{"d", "Delete selected connection"},
				{"y", "Copy SSH command"},
			},
		},
		{
			title: "General",
			keys: [][]string{
				{"?", "Toggle this help"},
				{"q", "Quit"},
				{"ctrl+c", "Force quit"},
			},
		},
	}

	for _, section := range sections {
		b.WriteString(projectStyle.Render(section.title))
		b.WriteString("\n")
		for _, key := range section.keys {
			b.WriteString("  ")
			b.WriteString(helpKeyStyle.Render(padRight(key[0], 10)))
			b.WriteString(" ")
			b.WriteString(helpDescStyle.Render(key[1]))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	b.WriteString(helpDescStyle.Render("Press any key to close"))

	return b.String()
}

func padRight(s string, length int) string {
	if len(s) >= length {
		return s
	}
	return s + strings.Repeat(" ", length-len(s))
}
