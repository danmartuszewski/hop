package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/danmartuszewski/hop/internal/config"
	"github.com/danmartuszewski/hop/internal/fuzzy"
)

type viewState int

const (
	viewList viewState = iota
	viewHelp
	viewForm
	viewConfirmDelete
	viewPaste
	viewTagPicker
)

type listItem struct {
	connection *config.Connection
	isProject  bool
	isEnv      bool
	label      string
	indent     int
	expanded   bool
}

type Model struct {
	config       *config.Config
	configPath   string
	version      string
	items        []listItem
	filtered     []int
	cursor       int
	filter       textinput.Model
	filtering    bool
	width        int
	height       int
	selected     *config.Connection
	quitting     bool
	view         viewState
	form         FormModel
	help         HelpModel
	pasteInput   textinput.Model
	statusMsg    string
	deleteTarget *config.Connection
	// Tag filtering
	activeTags    map[string]bool
	allTags       []string
	tagCursor     int
	// Recent connections
	history       *config.History
	sortByRecent  bool
}

func NewModel(cfg *config.Config, version string) Model {
	ti := textinput.New()
	ti.Placeholder = "Type to filter..."
	ti.CharLimit = 50
	ti.Width = 30

	paste := textinput.New()
	paste.Placeholder = "user@host:port or ssh user@host -p 22"
	paste.CharLimit = 200
	paste.Width = 50

	// Load history (ignore errors - history is optional)
	history, _ := config.LoadHistory()
	if history == nil {
		history = &config.History{}
	}

	m := Model{
		config:     cfg,
		configPath: config.DefaultConfigPath(),
		version:    version,
		filter:     ti,
		pasteInput: paste,
		filtered:   []int{},
		view:       viewList,
		help:       NewHelpModel(),
		activeTags: make(map[string]bool),
		history:    history,
	}

	m.buildItems()
	m.collectTags()
	m.resetFilter()

	return m
}

func (m *Model) buildItems() {
	m.items = []listItem{}

	groups := m.groupConnections()
	for _, project := range sortedMapKeys(groups) {
		envs := groups[project]

		if project != "" {
			m.items = append(m.items, listItem{
				isProject: true,
				label:     project,
				indent:    0,
				expanded:  true,
			})
		}

		for _, env := range sortedMapKeys(envs) {
			conns := envs[env]

			if env != "" && project != "" {
				m.items = append(m.items, listItem{
					isEnv:    true,
					label:    env,
					indent:   1,
					expanded: true,
				})
			}

			indent := 0
			if project != "" {
				indent++
			}
			if env != "" {
				indent++
			}

			for _, conn := range conns {
				c := conn
				m.items = append(m.items, listItem{
					connection: &c,
					indent:     indent,
				})
			}
		}
	}
}

func (m *Model) groupConnections() map[string]map[string][]config.Connection {
	groups := make(map[string]map[string][]config.Connection)

	for _, conn := range m.config.Connections {
		project := conn.Project
		env := conn.Env

		if groups[project] == nil {
			groups[project] = make(map[string][]config.Connection)
		}
		groups[project][env] = append(groups[project][env], conn)
	}

	return groups
}

func (m *Model) collectTags() {
	tagSet := make(map[string]bool)
	for _, conn := range m.config.Connections {
		for _, tag := range conn.Tags {
			tagSet[tag] = true
		}
	}
	m.allTags = make([]string, 0, len(tagSet))
	for tag := range tagSet {
		m.allTags = append(m.allTags, tag)
	}
	// Sort tags alphabetically
	for i := 0; i < len(m.allTags)-1; i++ {
		for j := i + 1; j < len(m.allTags); j++ {
			if m.allTags[i] > m.allTags[j] {
				m.allTags[i], m.allTags[j] = m.allTags[j], m.allTags[i]
			}
		}
	}
}

func (m *Model) hasActiveTagFilter() bool {
	for _, active := range m.activeTags {
		if active {
			return true
		}
	}
	return false
}

func (m *Model) matchesTags(conn *config.Connection) bool {
	if !m.hasActiveTagFilter() {
		return true
	}
	for tag, active := range m.activeTags {
		if !active {
			continue
		}
		found := false
		for _, connTag := range conn.Tags {
			if strings.EqualFold(connTag, tag) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (m *Model) resetFilter() {
	m.filtered = []int{}
	for i, item := range m.items {
		if item.connection != nil && m.matchesTags(item.connection) {
			m.filtered = append(m.filtered, i)
		}
	}
	m.sortFiltered()
	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
}

func (m *Model) applyFilter(query string) {
	if query == "" && !m.hasActiveTagFilter() {
		m.resetFilter()
		return
	}

	// Build a map of connection ID to item index for efficient lookup
	idToIndex := make(map[string]int)
	for i, item := range m.items {
		if item.connection != nil {
			idToIndex[item.connection.ID] = i
		}
	}

	m.filtered = []int{}

	if query == "" {
		// Only tag filtering active
		for i, item := range m.items {
			if item.connection != nil && m.matchesTags(item.connection) {
				m.filtered = append(m.filtered, i)
			}
		}
	} else {
		// Use sophisticated fuzzy matching with scoring for consistency with CLI
		matches := fuzzy.FindMatches(query, m.config.Connections)

		// Convert matches to filtered indices, preserving score-based order
		// Also apply tag filter
		for _, match := range matches {
			if !m.matchesTags(match.Connection) {
				continue
			}
			if idx, ok := idToIndex[match.Connection.ID]; ok {
				m.filtered = append(m.filtered, idx)
			}
		}
	}
	m.sortFiltered()
	m.cursor = 0
}

// sortFiltered sorts the filtered list by recent usage if enabled
func (m *Model) sortFiltered() {
	if !m.sortByRecent || m.history == nil {
		return
	}

	sort.SliceStable(m.filtered, func(i, j int) bool {
		connI := m.items[m.filtered[i]].connection
		connJ := m.items[m.filtered[j]].connection
		if connI == nil || connJ == nil {
			return false
		}

		timeI, okI := m.history.GetLastUsed(connI.ID)
		timeJ, okJ := m.history.GetLastUsed(connJ.ID)

		// Connections with no history go to the bottom
		if !okI && !okJ {
			return false
		}
		if !okI {
			return false
		}
		if !okJ {
			return true
		}

		return timeI.After(timeJ)
	})
}

func (m *Model) selectedConnection() *config.Connection {
	if len(m.filtered) == 0 || m.cursor >= len(m.filtered) {
		return nil
	}
	idx := m.filtered[m.cursor]
	return m.items[idx].connection
}

func (m *Model) refresh() {
	m.buildItems()
	m.applyFilter(m.filter.Value())
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	switch m.view {
	case viewHelp:
		return m.updateHelp(msg)
	case viewForm:
		return m.updateForm(msg)
	case viewConfirmDelete:
		return m.updateConfirmDelete(msg)
	case viewPaste:
		return m.updatePaste(msg)
	case viewTagPicker:
		return m.updateTagPicker(msg)
	default:
		return m.updateList(msg)
	}
}

func (m Model) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Clear status message on any key
		m.statusMsg = ""

		if m.filtering {
			switch msg.String() {
			case "enter":
				m.filtering = false
				m.applyFilter(m.filter.Value())
				return m, nil
			case "esc":
				m.filtering = false
				m.filter.SetValue("")
				m.resetFilter()
				return m, nil
			}
			var cmd tea.Cmd
			m.filter, cmd = m.filter.Update(msg)
			m.applyFilter(m.filter.Value())
			return m, cmd
		}

		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "?":
			m.view = viewHelp
			return m, nil
		case "/":
			m.filtering = true
			m.filter.Focus()
			return m, textinput.Blink
		case "esc":
			if m.filter.Value() != "" {
				m.filter.SetValue("")
				m.resetFilter()
			}
			return m, nil
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
		case "home", "g":
			m.cursor = 0
		case "end", "G":
			if len(m.filtered) > 0 {
				m.cursor = len(m.filtered) - 1
			}
		case "enter":
			if conn := m.selectedConnection(); conn != nil {
				m.selected = conn
				return m, tea.Quit
			}
		case "a":
			m.form = NewFormModel("Add Connection", nil)
			m.view = viewForm
			return m, textinput.Blink
		case "e":
			if conn := m.selectedConnection(); conn != nil {
				m.form = NewFormModel("Edit Connection", conn)
				m.view = viewForm
				return m, textinput.Blink
			}
		case "d":
			if conn := m.selectedConnection(); conn != nil {
				m.deleteTarget = conn
				m.view = viewConfirmDelete
			}
			return m, nil
		case "c":
			if conn := m.selectedConnection(); conn != nil {
				// Duplicate with empty ID
				dup := *conn
				dup.ID = ""
				m.form = NewFormModel("Duplicate Connection", &dup)
				m.form.editing = false // Treat as new connection
				m.form.originalID = ""
				m.view = viewForm
				return m, textinput.Blink
			}
		case "p":
			m.pasteInput.SetValue("")
			m.pasteInput.Focus()
			m.view = viewPaste
			return m, textinput.Blink
		case "y":
			if conn := m.selectedConnection(); conn != nil {
				user := conn.EffectiveUser()
				var sshCmd string
				if user != "" {
					sshCmd = fmt.Sprintf("ssh %s@%s", user, conn.Host)
				} else {
					sshCmd = fmt.Sprintf("ssh %s", conn.Host)
				}
				if conn.Port != 0 && conn.Port != 22 {
					sshCmd += fmt.Sprintf(" -p %d", conn.Port)
				}
				m.statusMsg = fmt.Sprintf("Copied: %s", sshCmd)
			}
			return m, nil
		case "t":
			if len(m.allTags) > 0 {
				m.tagCursor = 0
				m.view = viewTagPicker
			} else {
				m.statusMsg = "No tags defined"
			}
			return m, nil
		case "r":
			m.sortByRecent = !m.sortByRecent
			if m.sortByRecent {
				m.statusMsg = "Sorted by recent"
			} else {
				m.statusMsg = "Sorted by name"
			}
			m.applyFilter(m.filter.Value())
			return m, nil
		}
	}

	return m, nil
}

func (m Model) updateHelp(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		m.view = viewList
		return m, nil
	}

	var cmd tea.Cmd
	m.help, cmd = m.help.Update(msg)
	return m, cmd
}

func (m Model) updateForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.form, cmd = m.form.Update(msg)

	if m.form.Cancelled() {
		m.view = viewList
		return m, nil
	}

	if m.form.Submitted() {
		conn, err := m.form.GetConnection()
		if err != nil {
			m.statusMsg = "Error: " + err.Error()
			m.view = viewList
			return m, nil
		}

		if m.form.IsEditing() {
			// Update existing
			m.config.UpdateConnection(m.form.OriginalID(), *conn)
			m.statusMsg = fmt.Sprintf("Updated: %s", conn.ID)
		} else {
			// Check for duplicate ID
			if m.config.FindConnection(conn.ID) != nil {
				m.statusMsg = fmt.Sprintf("Error: Connection '%s' already exists", conn.ID)
				m.view = viewList
				return m, nil
			}
			m.config.AddConnection(*conn)
			m.statusMsg = fmt.Sprintf("Added: %s", conn.ID)
		}

		// Save config
		if err := m.config.Save(m.configPath); err != nil {
			m.statusMsg = "Error saving: " + err.Error()
		}

		m.refresh()
		m.view = viewList
		return m, nil
	}

	return m, cmd
}

func (m Model) updateConfirmDelete(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			if m.deleteTarget != nil {
				id := m.deleteTarget.ID
				m.config.DeleteConnection(id)
				if err := m.config.Save(m.configPath); err != nil {
					m.statusMsg = "Error saving: " + err.Error()
				} else {
					m.statusMsg = fmt.Sprintf("Deleted: %s", id)
				}
				m.refresh()
			}
			m.deleteTarget = nil
			m.view = viewList
			return m, nil
		case "n", "N", "esc":
			m.deleteTarget = nil
			m.view = viewList
			return m, nil
		}
	}
	return m, nil
}

func (m Model) updatePaste(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.view = viewList
			return m, nil
		case "enter":
			input := m.pasteInput.Value()
			conn := ParseSSHString(input)
			if conn == nil {
				m.statusMsg = "Could not parse SSH string"
				m.view = viewList
				return m, nil
			}
			m.form = NewFormModel("Add Connection", conn)
			m.form.editing = false
			m.view = viewForm
			return m, textinput.Blink
		}
	}

	var cmd tea.Cmd
	m.pasteInput, cmd = m.pasteInput.Update(msg)
	return m, cmd
}

func (m Model) updateTagPicker(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "t":
			m.view = viewList
			return m, nil
		case "enter", " ":
			// Toggle selected tag
			if m.tagCursor < len(m.allTags) {
				tag := m.allTags[m.tagCursor]
				m.activeTags[tag] = !m.activeTags[tag]
				m.applyFilter(m.filter.Value())
			}
			return m, nil
		case "up", "k":
			if m.tagCursor > 0 {
				m.tagCursor--
			}
			return m, nil
		case "down", "j":
			if m.tagCursor < len(m.allTags)-1 {
				m.tagCursor++
			}
			return m, nil
		case "c":
			// Clear all tag filters
			m.activeTags = make(map[string]bool)
			m.applyFilter(m.filter.Value())
			return m, nil
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	switch m.view {
	case viewHelp:
		return m.renderHelp()
	case viewForm:
		return m.renderForm()
	case viewConfirmDelete:
		return m.renderConfirmDelete()
	case viewPaste:
		return m.renderPaste()
	case viewTagPicker:
		return m.renderTagPicker()
	default:
		return m.renderList()
	}
}

func (m Model) renderList() string {
	var b strings.Builder

	// Header
	header := m.renderHeader()
	b.WriteString(header)
	b.WriteString("\n")

	// Filter bar
	filterBar := m.renderFilterBar()
	b.WriteString(filterBar)
	b.WriteString("\n")

	// Calculate panel width (40% of terminal, min 30, max 50)
	panelWidth := m.width * 40 / 100
	if panelWidth < 30 {
		panelWidth = 30
	}
	if panelWidth > 50 {
		panelWidth = 50
	}

	// Only show panel if terminal is wide enough (80+ chars)
	showPanel := m.width >= 80 && m.selectedConnection() != nil

	if showPanel {
		// Calculate list width
		listWidth := m.width - panelWidth - 3 // 3 for border/padding

		// Render list and panel side by side
		listContent := m.renderListContentWithWidth(listWidth)
		panel := m.renderDetailsPanel(panelWidth)

		// Join side by side
		listLines := strings.Split(listContent, "\n")
		panelLines := strings.Split(panel, "\n")

		// Pad to same height
		maxLines := len(listLines)
		if len(panelLines) > maxLines {
			maxLines = len(panelLines)
		}
		for len(listLines) < maxLines {
			listLines = append(listLines, strings.Repeat(" ", listWidth))
		}
		for len(panelLines) < maxLines {
			panelLines = append(panelLines, "")
		}

		for i := 0; i < maxLines; i++ {
			// Pad list line to full width
			listLine := listLines[i]
			lineWidth := lipgloss.Width(listLine)
			if lineWidth < listWidth {
				listLine += strings.Repeat(" ", listWidth-lineWidth)
			}
			b.WriteString(listLine)
			b.WriteString(" ")
			b.WriteString(panelLines[i])
			b.WriteString("\n")
		}
	} else {
		// List only
		list := m.renderListContent()
		b.WriteString(list)
	}

	// Status message
	if m.statusMsg != "" {
		b.WriteString(statusStyle.Render(m.statusMsg))
		b.WriteString("\n")
	}

	// Footer
	footer := m.renderFooter()
	b.WriteString("\n")
	b.WriteString(footer)

	return b.String()
}

func (m Model) renderHeader() string {
	title := titleStyle.Render("hop")
	subtitle := " - SSH Connection Manager"
	version := versionStyle.Render("v" + m.version)

	left := title + subtitle
	right := version

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right) - 4
	if gap < 1 {
		gap = 1
	}

	return headerStyle.Width(m.width).Render(left + strings.Repeat(" ", gap) + right)
}

func (m Model) renderFilterBar() string {
	var filterView string
	if m.filtering {
		filterView = filterPromptStyle.Render(" / ") + m.filter.View()
	} else if m.filter.Value() != "" {
		filterView = filterPromptStyle.Render(" / ") + filterInputStyle.Render(m.filter.Value())
	} else {
		filterView = statusStyle.Render(fmt.Sprintf(" %d connections", len(m.filtered)))
	}

	// Append active tags
	var activeTags []string
	for _, tag := range m.allTags {
		if m.activeTags[tag] {
			activeTags = append(activeTags, tag)
		}
	}
	if len(activeTags) > 0 {
		filterView += " "
		for _, tag := range activeTags {
			filterView += " " + panelTagStyle.Render(tag)
		}
	}

	return filterView
}

func (m Model) renderListContent() string {
	if len(m.config.Connections) == 0 {
		return emptyStyle.Render("No connections configured.\n\nPress 'a' to add a connection, or '?' for help.")
	}

	if len(m.filtered) == 0 {
		return emptyStyle.Render("No connections match your filter.")
	}

	// Calculate max ID width for column alignment
	maxIDWidth := 0
	for _, i := range m.filtered {
		item := m.items[i]
		if item.connection != nil && len(item.connection.ID) > maxIDWidth {
			maxIDWidth = len(item.connection.ID)
		}
	}

	// Build visible lines including group headers
	type displayLine struct {
		text        string
		filterIndex int // -1 for group headers
	}
	var lines []displayLine

	lastProject := ""
	lastEnv := ""
	notFiltering := m.filter.Value() == ""

	for fi, idx := range m.filtered {
		item := m.items[idx]
		conn := item.connection
		if conn == nil {
			continue
		}

		// Insert project/env headers when grouping changes
		if notFiltering {
			if conn.Project != "" && conn.Project != lastProject {
				lines = append(lines, displayLine{
					text:        projectStyle.Render(conn.Project),
					filterIndex: -1,
				})
				lastEnv = "" // reset env on project change
			}
			if conn.Env != "" && (conn.Env != lastEnv || conn.Project != lastProject) {
				indent := ""
				if conn.Project != "" {
					indent = "  "
				}
				lines = append(lines, displayLine{
					text:        indent + envStyle.Render(conn.Env),
					filterIndex: -1,
				})
			}
			lastProject = conn.Project
			lastEnv = conn.Env
		}

		// Build connection line
		isSelected := fi == m.cursor
		id := conn.ID
		padded := id + strings.Repeat(" ", maxIDWidth-len(id))
		host := conn.Host
		user := conn.EffectiveUser()

		userHost := hostStyle.Render(host)
		if user != "" {
			userHost = userStyle.Render(user) + "@" + hostStyle.Render(host)
		}
		portStr := ""
		if conn.Port != 0 && conn.Port != 22 {
			portStr = "  " + hostStyle.Render(fmt.Sprintf(":%d", conn.Port))
		}

		// Calculate indent based on grouping
		indent := ""
		if conn.Project != "" {
			indent = "  " // Under project
			if conn.Env != "" {
				indent = "    " // Under project + env
			}
		}

		var line string
		if isSelected {
			line = indent + selectedItemStyle.Render(">") + " " + selectedItemStyle.Render(padded) + "  " + userHost + portStr
		} else {
			line = indent + "  " + itemStyle.Render(padded) + "  " + userHost + portStr
		}

		lines = append(lines, displayLine{text: line, filterIndex: fi})
	}

	// Calculate visible range based on cursor position in display lines
	listHeight := m.height - 7
	if listHeight < 5 {
		listHeight = 5
	}

	// Find the display line index that corresponds to the cursor
	cursorDisplayIdx := 0
	for i, dl := range lines {
		if dl.filterIndex == m.cursor {
			cursorDisplayIdx = i
			break
		}
	}

	start := 0
	if cursorDisplayIdx >= listHeight {
		start = cursorDisplayIdx - listHeight + 1
	}
	end := start + listHeight
	if end > len(lines) {
		end = len(lines)
	}

	var b strings.Builder
	for i := start; i < end; i++ {
		b.WriteString(lines[i].text)
		b.WriteString("\n")
	}

	return b.String()
}

func (m Model) renderListContentWithWidth(maxWidth int) string {
	if len(m.config.Connections) == 0 {
		return emptyStyle.Render("No connections configured.\n\nPress 'a' to add a connection, or '?' for help.")
	}

	if len(m.filtered) == 0 {
		return emptyStyle.Render("No connections match your filter.")
	}

	// Calculate max ID width for column alignment
	maxIDWidth := 0
	for _, i := range m.filtered {
		item := m.items[i]
		if item.connection != nil && len(item.connection.ID) > maxIDWidth {
			maxIDWidth = len(item.connection.ID)
		}
	}
	// Limit ID width to fit in available space
	if maxIDWidth > maxWidth/2 {
		maxIDWidth = maxWidth / 2
	}

	// Build visible lines including group headers
	type displayLine struct {
		text        string
		filterIndex int // -1 for group headers
	}
	var lines []displayLine

	lastProject := ""
	lastEnv := ""
	notFiltering := m.filter.Value() == ""

	for fi, idx := range m.filtered {
		item := m.items[idx]
		conn := item.connection
		if conn == nil {
			continue
		}

		// Insert project/env headers when grouping changes
		if notFiltering {
			if conn.Project != "" && conn.Project != lastProject {
				lines = append(lines, displayLine{
					text:        projectStyle.Render(conn.Project),
					filterIndex: -1,
				})
				lastEnv = "" // reset env on project change
			}
			if conn.Env != "" && (conn.Env != lastEnv || conn.Project != lastProject) {
				indent := ""
				if conn.Project != "" {
					indent = "  "
				}
				lines = append(lines, displayLine{
					text:        indent + envStyle.Render(conn.Env),
					filterIndex: -1,
				})
			}
			lastProject = conn.Project
			lastEnv = conn.Env
		}

		// Build connection line
		isSelected := fi == m.cursor
		id := conn.ID
		if len(id) > maxIDWidth {
			id = id[:maxIDWidth-1] + "…"
		}
		padded := id + strings.Repeat(" ", maxIDWidth-len(id))
		host := conn.Host
		user := conn.EffectiveUser()

		userHost := hostStyle.Render(host)
		if user != "" {
			userHost = userStyle.Render(user) + "@" + hostStyle.Render(host)
		}
		portStr := ""
		if conn.Port != 0 && conn.Port != 22 {
			portStr = "  " + hostStyle.Render(fmt.Sprintf(":%d", conn.Port))
		}

		// Calculate indent based on grouping
		indent := ""
		if conn.Project != "" {
			indent = "  " // Under project
			if conn.Env != "" {
				indent = "    " // Under project + env
			}
		}

		var line string
		if isSelected {
			line = indent + selectedItemStyle.Render(">") + " " + selectedItemStyle.Render(padded) + "  " + userHost + portStr
		} else {
			line = indent + "  " + itemStyle.Render(padded) + "  " + userHost + portStr
		}

		lines = append(lines, displayLine{text: line, filterIndex: fi})
	}

	// Calculate visible range based on cursor position in display lines
	listHeight := m.height - 7
	if listHeight < 5 {
		listHeight = 5
	}

	// Find the display line index that corresponds to the cursor
	cursorDisplayIdx := 0
	for i, dl := range lines {
		if dl.filterIndex == m.cursor {
			cursorDisplayIdx = i
			break
		}
	}

	start := 0
	if cursorDisplayIdx >= listHeight {
		start = cursorDisplayIdx - listHeight + 1
	}
	end := start + listHeight
	if end > len(lines) {
		end = len(lines)
	}

	var b strings.Builder
	for i := start; i < end; i++ {
		b.WriteString(lines[i].text)
		if i < end-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (m Model) renderDetailsPanel(width int) string {
	conn := m.selectedConnection()
	if conn == nil {
		return ""
	}

	var b strings.Builder

	// Title
	b.WriteString(panelTitleStyle.Render("Connection Details"))
	b.WriteString("\n\n")

	// Helper function for label: value pairs
	addField := func(label, value string) {
		if value == "" {
			return
		}
		b.WriteString(panelLabelStyle.Render(label + ": "))
		b.WriteString(panelValueStyle.Render(value))
		b.WriteString("\n")
	}

	// Basic info
	addField("ID", conn.ID)
	addField("Host", conn.Host)
	addField("User", conn.EffectiveUser())
	if conn.Port != 0 && conn.Port != 22 {
		addField("Port", fmt.Sprintf("%d", conn.Port))
	}

	// Project/Environment
	if conn.Project != "" || conn.Env != "" {
		b.WriteString("\n")
		addField("Project", conn.Project)
		addField("Environment", conn.Env)
	}

	// Tags
	if len(conn.Tags) > 0 {
		b.WriteString("\n")
		b.WriteString(panelLabelStyle.Render("Tags: "))
		for i, tag := range conn.Tags {
			if i > 0 {
				b.WriteString(" ")
			}
			b.WriteString(panelTagStyle.Render(tag))
		}
		b.WriteString("\n")
	}

	// Identity file
	if conn.IdentityFile != "" {
		b.WriteString("\n")
		addField("Identity", conn.IdentityFile)
	}

	// SSH options
	if len(conn.Options) > 0 {
		b.WriteString("\n")
		b.WriteString(panelLabelStyle.Render("SSH Options:"))
		b.WriteString("\n")
		for key, value := range conn.Options {
			b.WriteString("  ")
			b.WriteString(panelValueStyle.Render(key + "=" + value))
			b.WriteString("\n")
		}
	}

	// Apply panel style with border
	return panelStyle.Width(width - 4).Render(b.String())
}

func (m Model) renderFooter() string {
	var keys []string

	keys = append(keys, helpKeyStyle.Render("↑/↓")+" "+helpDescStyle.Render("nav"))
	keys = append(keys, helpKeyStyle.Render("/")+" "+helpDescStyle.Render("filter"))
	keys = append(keys, helpKeyStyle.Render("t")+" "+helpDescStyle.Render("tags"))
	recentLabel := "recent"
	if m.sortByRecent {
		recentLabel = "recent*"
	}
	keys = append(keys, helpKeyStyle.Render("r")+" "+helpDescStyle.Render(recentLabel))
	keys = append(keys, helpKeyStyle.Render("a")+" "+helpDescStyle.Render("add"))
	keys = append(keys, helpKeyStyle.Render("e")+" "+helpDescStyle.Render("edit"))
	keys = append(keys, helpKeyStyle.Render("d")+" "+helpDescStyle.Render("del"))
	keys = append(keys, helpKeyStyle.Render("enter")+" "+helpDescStyle.Render("connect"))
	keys = append(keys, helpKeyStyle.Render("?")+" "+helpDescStyle.Render("help"))

	return footerStyle.Width(m.width).Render(strings.Join(keys, "  "))
}

func (m Model) renderHelp() string {
	return m.help.View()
}

func (m Model) renderForm() string {
	return m.form.View()
}

func (m Model) renderConfirmDelete() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Delete Connection"))
	b.WriteString("\n\n")

	if m.deleteTarget != nil {
		b.WriteString(fmt.Sprintf("Are you sure you want to delete '%s'?\n", m.deleteTarget.ID))
		b.WriteString(helpDescStyle.Render(fmt.Sprintf("Host: %s", m.deleteTarget.Host)))
		b.WriteString("\n\n")
	}

	b.WriteString(helpKeyStyle.Render("y") + " " + helpDescStyle.Render("Yes, delete"))
	b.WriteString("  ")
	b.WriteString(helpKeyStyle.Render("n") + " " + helpDescStyle.Render("No, cancel"))

	return b.String()
}

func (m Model) renderPaste() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Quick Add - Paste SSH String"))
	b.WriteString("\n\n")

	b.WriteString(helpDescStyle.Render("Paste or type an SSH connection string:"))
	b.WriteString("\n\n")

	b.WriteString("  ")
	b.WriteString(m.pasteInput.View())
	b.WriteString("\n\n")

	b.WriteString(helpDescStyle.Render("Supported formats:"))
	b.WriteString("\n")
	b.WriteString(helpDescStyle.Render("  user@host"))
	b.WriteString("\n")
	b.WriteString(helpDescStyle.Render("  user@host:port"))
	b.WriteString("\n")
	b.WriteString(helpDescStyle.Render("  ssh user@host -p 22"))
	b.WriteString("\n\n")

	b.WriteString(helpKeyStyle.Render("enter") + " " + helpDescStyle.Render("parse & continue"))
	b.WriteString("  ")
	b.WriteString(helpKeyStyle.Render("esc") + " " + helpDescStyle.Render("cancel"))

	return b.String()
}

func (m Model) renderTagPicker() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Filter by Tags"))
	b.WriteString("\n\n")

	if len(m.allTags) == 0 {
		b.WriteString(emptyStyle.Render("No tags defined in your connections."))
		b.WriteString("\n\n")
	} else {
		b.WriteString(helpDescStyle.Render("Select tags to filter connections (AND logic):"))
		b.WriteString("\n\n")

		for i, tag := range m.allTags {
			isSelected := m.tagCursor == i
			isActive := m.activeTags[tag]

			// Checkbox
			checkbox := "[ ]"
			if isActive {
				checkbox = "[✓]"
			}

			line := "  " + checkbox + " " + tag

			if isSelected {
				b.WriteString(selectedItemStyle.Render("> " + checkbox + " " + tag))
			} else if isActive {
				b.WriteString("  " + panelTagStyle.Render(checkbox+" "+tag))
			} else {
				b.WriteString("  " + helpDescStyle.Render(checkbox+" "+tag))
			}
			b.WriteString("\n")
			_ = line // silence unused
		}
	}

	b.WriteString("\n")
	b.WriteString(helpKeyStyle.Render("space/enter") + " " + helpDescStyle.Render("toggle"))
	b.WriteString("  ")
	b.WriteString(helpKeyStyle.Render("c") + " " + helpDescStyle.Render("clear all"))
	b.WriteString("  ")
	b.WriteString(helpKeyStyle.Render("esc/t") + " " + helpDescStyle.Render("close"))

	return b.String()
}

func (m Model) Selected() *config.Connection {
	return m.selected
}

func sortedMapKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	// Sort with empty string first
	for i := 0; i < len(keys)-1; i++ {
		for j := i + 1; j < len(keys); j++ {
			swap := false
			if keys[i] == "" {
				swap = false
			} else if keys[j] == "" {
				swap = true
			} else if keys[i] > keys[j] {
				swap = true
			}
			if swap {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	return keys
}

func Run(cfg *config.Config, version string) (*config.Connection, error) {
	m := NewModel(cfg, version)
	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("TUI error: %w", err)
	}

	result := finalModel.(Model)
	selected := result.Selected()

	// Record usage in history
	if selected != nil && result.history != nil {
		result.history.RecordUsage(selected.ID)
		_ = result.history.Save() // Ignore save errors - history is optional
	}

	return selected, nil
}
