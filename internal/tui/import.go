package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/danmartuszewski/hop/internal/config"
	"github.com/danmartuszewski/hop/internal/sshconfig"
)

// ImportItem represents a connection to be imported
type ImportItem struct {
	Original   string
	Connection config.Connection
	Renamed    bool
	Selected   bool
}

// ImportModel handles the SSH config import modal
type ImportModel struct {
	items      []ImportItem
	cursor     int
	width      int
	height     int
	cancelled  bool
	confirmed  bool
	err        error
	scrollTop  int
	configPath string
}

// NewImportModel creates a new import model by parsing SSH config
func NewImportModel(existingIDs map[string]bool, sshConfigPath string, hopConfigPath string, width, height int) ImportModel {
	hosts, err := sshconfig.Parse(sshConfigPath)
	if err != nil {
		return ImportModel{err: err, configPath: hopConfigPath, width: width, height: height}
	}

	// Make a copy of existingIDs to track as we process
	usedIDs := make(map[string]bool)
	for k, v := range existingIDs {
		usedIDs[k] = v
	}

	var items []ImportItem
	for _, host := range hosts {
		conn := host.ToConnection()
		originalID := conn.ID
		renamed := false

		// Handle ID conflicts
		if usedIDs[conn.ID] {
			conn.ID = sshconfig.ResolveConflict(conn.ID, usedIDs)
			renamed = true
		}
		usedIDs[conn.ID] = true

		items = append(items, ImportItem{
			Original:   originalID,
			Connection: conn,
			Renamed:    renamed,
			Selected:   true, // All selected by default
		})
	}

	return ImportModel{
		items:      items,
		configPath: hopConfigPath,
		width:      width,
		height:     height,
	}
}

func (m ImportModel) Init() tea.Cmd {
	return nil
}

func (m ImportModel) Update(msg tea.Msg) (ImportModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
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
			// Toggle selection
			if m.cursor < len(m.items) {
				m.items[m.cursor].Selected = !m.items[m.cursor].Selected
			}

		case "a":
			// Select all
			for i := range m.items {
				m.items[i].Selected = true
			}

		case "n":
			// Deselect all
			for i := range m.items {
				m.items[i].Selected = false
			}
		}
	}

	return m, nil
}

func (m *ImportModel) ensureVisible() {
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

func (m ImportModel) visibleHeight() int {
	// Account for title, help, and padding
	h := m.height - 10
	if h < 3 {
		h = 3
	}
	return h
}

func (m ImportModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Import from SSH Config"))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(emptyStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		b.WriteString("\n\n")
		b.WriteString(helpKeyStyle.Render("esc") + " " + helpDescStyle.Render("close"))
		return b.String()
	}

	if len(m.items) == 0 {
		b.WriteString(emptyStyle.Render("No importable connections found in SSH config."))
		b.WriteString("\n")
		b.WriteString(helpDescStyle.Render("(Wildcard patterns like Host * are automatically skipped)"))
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

	b.WriteString(helpDescStyle.Render(fmt.Sprintf("Select connections to import (%d/%d selected):", selectedCount, len(m.items))))
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

		// Checkbox
		checkbox := "[ ]"
		if item.Selected {
			checkbox = "[x]"
		}

		// Build the line
		var line strings.Builder

		if isSelected {
			line.WriteString(selectedItemStyle.Render("> "))
		} else {
			line.WriteString("  ")
		}

		// Connection info
		id := item.Connection.ID
		if item.Renamed {
			id = fmt.Sprintf("%s (was: %s)", item.Connection.ID, item.Original)
		}

		host := item.Connection.Host
		if item.Connection.User != "" {
			host = item.Connection.User + "@" + host
		}

		if isSelected {
			line.WriteString(selectedItemStyle.Render(checkbox + " " + id))
		} else if item.Renamed {
			line.WriteString(warningStyle.Render(checkbox + " " + id))
		} else {
			line.WriteString(checkbox + " " + itemStyle.Render(id))
		}

		line.WriteString("  ")
		line.WriteString(hostStyle.Render(host))

		// Show extra info on same line
		var extras []string
		if item.Connection.Port != 0 && item.Connection.Port != 22 {
			extras = append(extras, fmt.Sprintf(":%d", item.Connection.Port))
		}
		if item.Connection.ProxyJump != "" {
			extras = append(extras, "via "+item.Connection.ProxyJump)
		}
		if item.Connection.ForwardAgent {
			extras = append(extras, "agent")
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
			b.WriteString(helpDescStyle.Render("  ↑ more above"))
			b.WriteString("\n")
		}
		if end < len(m.items) {
			b.WriteString(helpDescStyle.Render(fmt.Sprintf("  ↓ %d more below", len(m.items)-end)))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")

	// Help
	help := helpKeyStyle.Render("space") + " " + helpDescStyle.Render("toggle") + "  "
	help += helpKeyStyle.Render("a") + " " + helpDescStyle.Render("all") + "  "
	help += helpKeyStyle.Render("n") + " " + helpDescStyle.Render("none") + "  "
	help += helpKeyStyle.Render("enter") + " " + helpDescStyle.Render("import") + "  "
	help += helpKeyStyle.Render("esc") + " " + helpDescStyle.Render("cancel")
	b.WriteString(help)

	return b.String()
}

func (m ImportModel) Cancelled() bool {
	return m.cancelled
}

func (m ImportModel) Confirmed() bool {
	return m.confirmed
}

// SelectedConnections returns the connections that were selected for import
func (m ImportModel) SelectedConnections() []config.Connection {
	var selected []config.Connection
	for _, item := range m.items {
		if item.Selected {
			selected = append(selected, item.Connection)
		}
	}
	return selected
}

// HasItems returns true if there are items to import
func (m ImportModel) HasItems() bool {
	return len(m.items) > 0
}

// Error returns any error that occurred during parsing
func (m ImportModel) Error() error {
	return m.err
}

// Add warningStyle if not already defined
var warningStyle = envStyle // Reuse the orange env style for warnings
