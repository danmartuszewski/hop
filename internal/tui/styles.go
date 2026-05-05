package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Header
	headerStyle lipgloss.Style

	titleStyle lipgloss.Style

	versionStyle lipgloss.Style

	// Filter
	filterPromptStyle lipgloss.Style

	filterInputStyle lipgloss.Style

	// List items
	itemStyle lipgloss.Style

	selectedItemStyle lipgloss.Style

	projectStyle lipgloss.Style

	envStyle lipgloss.Style

	hostStyle lipgloss.Style

	userStyle lipgloss.Style

	portStyle lipgloss.Style

	// Footer
	footerStyle lipgloss.Style

	helpKeyStyle lipgloss.Style

	helpDescStyle lipgloss.Style

	// Status
	statusStyle lipgloss.Style

	// Empty state
	emptyStyle lipgloss.Style

	// Tag style (used in filter bar and tag picker)
	panelTagStyle lipgloss.Style

	// Warning style (used for renamed entries in the import view)
	warningStyle lipgloss.Style

	// Health check indicators
	healthReachableStyle lipgloss.Style

	healthUnreachableStyle lipgloss.Style

	healthCheckingStyle lipgloss.Style
)

func init() {
	refreshStyles()
}

func refreshStyles() {
	headerStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(currentTheme.Primary).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(currentTheme.Secondary).
		Padding(0, 1)

	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(currentTheme.Primary)

	versionStyle = lipgloss.NewStyle().
		Foreground(currentTheme.Secondary)

	filterPromptStyle = lipgloss.NewStyle().
		Foreground(currentTheme.Primary).
		Bold(true)

	filterInputStyle = lipgloss.NewStyle().
		Foreground(currentTheme.Foreground)

	itemStyle = lipgloss.NewStyle().
		PaddingLeft(2)

	selectedItemStyle = lipgloss.NewStyle().
		PaddingLeft(1).
		Foreground(currentTheme.Primary).
		Bold(true)

	projectStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(currentTheme.Accent)

	envStyle = lipgloss.NewStyle().
		Foreground(currentTheme.Warning)

	hostStyle = lipgloss.NewStyle().
		Foreground(currentTheme.Secondary)

	userStyle = lipgloss.NewStyle().
		Foreground(currentTheme.Success)

	portStyle = lipgloss.NewStyle().
		Foreground(currentTheme.Muted)

	footerStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		BorderForeground(currentTheme.Secondary).
		Foreground(currentTheme.Secondary).
		Padding(0, 1)

	helpKeyStyle = lipgloss.NewStyle().
		Foreground(currentTheme.Primary).
		Bold(true)

	helpDescStyle = lipgloss.NewStyle().
		Foreground(currentTheme.Secondary)

	statusStyle = lipgloss.NewStyle().
		Foreground(currentTheme.Secondary).
		Italic(true)

	emptyStyle = lipgloss.NewStyle().
		Foreground(currentTheme.Secondary).
		Italic(true).
		Padding(2, 4)

	panelTagStyle = lipgloss.NewStyle().
		Foreground(currentTheme.Accent).
		Background(currentTheme.Selection).
		Padding(0, 1)

	warningStyle = lipgloss.NewStyle().
		Foreground(currentTheme.Warning)

	healthReachableStyle = lipgloss.NewStyle().
		Foreground(currentTheme.Success)

	healthUnreachableStyle = lipgloss.NewStyle().
		Foreground(currentTheme.Error)

	healthCheckingStyle = lipgloss.NewStyle().
		Foreground(currentTheme.Secondary)
}
