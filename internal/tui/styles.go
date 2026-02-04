package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	primaryColor   = lipgloss.Color("39")  // Cyan
	secondaryColor = lipgloss.Color("245") // Gray
	accentColor    = lipgloss.Color("212") // Pink
	successColor   = lipgloss.Color("82")  // Green
	warningColor   = lipgloss.Color("214") // Orange

	// Header
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(secondaryColor).
			Padding(0, 1)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor)

	versionStyle = lipgloss.NewStyle().
			Foreground(secondaryColor)

	// Filter
	filterPromptStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true)

	filterInputStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("255"))

	// List items
	itemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(1).
				Foreground(primaryColor).
				Bold(true)

	projectStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accentColor)

	envStyle = lipgloss.NewStyle().
			Foreground(warningColor)

	hostStyle = lipgloss.NewStyle().
			Foreground(secondaryColor)

	userStyle = lipgloss.NewStyle().
			Foreground(successColor)

	// Footer
	footerStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(secondaryColor).
			Foreground(secondaryColor).
			Padding(0, 1)

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(secondaryColor)

	// Status
	statusStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Italic(true)

	// Empty state
	emptyStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Italic(true).
			Padding(2, 4)

	// Tag style (used in filter bar and tag picker)
	panelTagStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Background(lipgloss.Color("236")).
			Padding(0, 1)
)
