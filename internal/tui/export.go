package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/danmartuszewski/hop/internal/config"
)

// ExportItem represents a connection available for export
type ExportItem struct {
	Connection config.Connection
	Selected   bool
}

// ExportModel handles the export modal
type ExportModel struct {
	items     []ExportItem
	cursor    int
	width     int
	height    int
	cancelled bool
	confirmed bool
	scrollTop int
	// Path input
	pathInput  textinput.Model
	focusPath  bool
}

// NewExportModel creates a new export model pre-populated with connections.
// All items start selected.
func NewExportModel(connections []config.Connection, width, height int) ExportModel {
	pi := textinput.New()
	pi.Placeholder = "export.yaml"
	pi.CharLimit = 200
	pi.Width = 40
	pi.SetValue("export.yaml")

	items := make([]ExportItem, len(connections))
	for i, conn := range connections {
		items[i] = ExportItem{
			Connection: conn,
			Selected:   true,
		}
	}

	return ExportModel{
		items:     items,
		width:     width,
		height:    height,
		pathInput: pi,
	}
}

func (m ExportModel) Init() tea.Cmd {
	return nil
}

func (m ExportModel) Update(msg tea.Msg) (ExportModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		if m.focusPath {
			switch msg.String() {
			case "esc":
				m.cancelled = true
				return m, nil
			case "enter":
				m.confirmed = true
				return m, nil
			case "tab", "shift+tab":
				m.focusPath = false
				m.pathInput.Blur()
				return m, nil
			}
			var cmd tea.Cmd
			m.pathInput, cmd = m.pathInput.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "esc", "q":
			m.cancelled = true
			return m, nil

		case "enter":
			m.confirmed = true
			return m, nil

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.ensureVisible()
			}

		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
				m.ensureVisible()
			}

		case " ", "x":
			if m.cursor < len(m.items) {
				m.items[m.cursor].Selected = !m.items[m.cursor].Selected
			}

		case "a":
			for i := range m.items {
				m.items[i].Selected = true
			}

		case "n":
			for i := range m.items {
				m.items[i].Selected = false
			}

		case "tab":
			m.focusPath = true
			m.pathInput.Focus()
			return m, textinput.Blink
		}
	}

	return m, nil
}

func (m *ExportModel) ensureVisible() {
	visibleHeight := m.visibleHeight()
	if visibleHeight <= 0 {
		return
	}

	if m.cursor < m.scrollTop {
		m.scrollTop = m.cursor
	} else if m.cursor >= m.scrollTop+visibleHeight {
		m.scrollTop = m.cursor - visibleHeight + 1
	}
}

func (m ExportModel) visibleHeight() int {
	// Account for title, path input, help, and padding
	h := m.height - 12
	if h < 3 {
		h = 3
	}
	return h
}

func (m ExportModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Export Connections"))
	b.WriteString("\n\n")

	if len(m.items) == 0 {
		b.WriteString(emptyStyle.Render("No connections to export."))
		b.WriteString("\n\n")
		b.WriteString(helpKeyStyle.Render("esc") + " " + helpDescStyle.Render("close"))
		return b.String()
	}

	// Count selected
	selectedCount := 0
	for _, item := range m.items {
		if item.Selected {
			selectedCount++
		}
	}

	b.WriteString(helpDescStyle.Render(fmt.Sprintf("Select connections to export (%d/%d selected):", selectedCount, len(m.items))))
	b.WriteString("\n\n")

	// Display items with scrolling
	visibleHeight := m.visibleHeight()
	start := m.scrollTop
	end := start + visibleHeight
	if end > len(m.items) {
		end = len(m.items)
	}

	for i := start; i < end; i++ {
		item := m.items[i]
		isSelected := i == m.cursor

		checkbox := "[ ]"
		if item.Selected {
			checkbox = "[x]"
		}

		var line strings.Builder

		if isSelected && !m.focusPath {
			line.WriteString(selectedItemStyle.Render("> "))
		} else {
			line.WriteString("  ")
		}

		id := item.Connection.ID
		host := item.Connection.Host
		if item.Connection.User != "" {
			host = item.Connection.User + "@" + host
		}

		if isSelected && !m.focusPath {
			line.WriteString(selectedItemStyle.Render(checkbox + " " + id))
		} else {
			line.WriteString(checkbox + " " + itemStyle.Render(id))
		}

		line.WriteString("  ")
		line.WriteString(hostStyle.Render(host))

		var extras []string
		if item.Connection.Port != 0 && item.Connection.Port != 22 {
			extras = append(extras, fmt.Sprintf(":%d", item.Connection.Port))
		}
		if item.Connection.Project != "" {
			extras = append(extras, item.Connection.Project)
		}
		if len(item.Connection.Tags) > 0 {
			extras = append(extras, strings.Join(item.Connection.Tags, ","))
		}
		if len(extras) > 0 {
			line.WriteString(" ")
			line.WriteString(hostStyle.Render("(" + strings.Join(extras, ", ") + ")"))
		}

		b.WriteString(line.String())
		b.WriteString("\n")
	}

	// Scroll indicator
	if len(m.items) > visibleHeight {
		if m.scrollTop > 0 {
			b.WriteString(helpDescStyle.Render("  \u2191 more above"))
			b.WriteString("\n")
		}
		if end < len(m.items) {
			b.WriteString(helpDescStyle.Render(fmt.Sprintf("  \u2193 %d more below", len(m.items)-end)))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")

	// Path input
	pathLabel := "  Output file: "
	if m.focusPath {
		b.WriteString(selectedItemStyle.Render(pathLabel))
		b.WriteString(m.pathInput.View())
	} else {
		b.WriteString(helpDescStyle.Render(pathLabel))
		val := m.pathInput.Value()
		if val == "" {
			val = "export.yaml"
		}
		b.WriteString(itemStyle.Render(val))
	}
	b.WriteString("\n\n")

	// Help
	help := helpKeyStyle.Render("space") + " " + helpDescStyle.Render("toggle") + "  "
	help += helpKeyStyle.Render("a") + " " + helpDescStyle.Render("all") + "  "
	help += helpKeyStyle.Render("n") + " " + helpDescStyle.Render("none") + "  "
	help += helpKeyStyle.Render("tab") + " " + helpDescStyle.Render("edit path") + "  "
	help += helpKeyStyle.Render("enter") + " " + helpDescStyle.Render("export") + "  "
	help += helpKeyStyle.Render("esc") + " " + helpDescStyle.Render("cancel")
	b.WriteString(help)

	return b.String()
}

func (m ExportModel) Cancelled() bool {
	return m.cancelled
}

func (m ExportModel) Confirmed() bool {
	return m.confirmed
}

// SelectedConnections returns the connections that were selected for export
func (m ExportModel) SelectedConnections() []config.Connection {
	var selected []config.Connection
	for _, item := range m.items {
		if item.Selected {
			selected = append(selected, item.Connection)
		}
	}
	return selected
}

// OutputPath returns the user-specified output file path
func (m ExportModel) OutputPath() string {
	val := m.pathInput.Value()
	if val == "" {
		return "export.yaml"
	}
	return val
}

// HasItems returns true if there are items to export
func (m ExportModel) HasItems() bool {
	return len(m.items) > 0
}
