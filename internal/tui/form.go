package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/danmartuszewski/hop/internal/config"
)

type formField int

const (
	fieldID formField = iota
	fieldHost
	fieldUser
	fieldPort
	fieldProject
	fieldEnv
	fieldTags
	fieldCount
)

type FormModel struct {
	inputs     []textinput.Model
	focused    formField
	title      string
	editing    bool
	originalID string
	width      int
	height     int
	cancelled  bool
	submitted  bool
}

func NewFormModel(title string, conn *config.Connection) FormModel {
	inputs := make([]textinput.Model, fieldCount)

	for i := range inputs {
		t := textinput.New()
		t.CharLimit = 100

		switch formField(i) {
		case fieldID:
			t.Placeholder = "_______________"
			t.Focus()
		case fieldHost:
			t.Placeholder = "_______________"
		case fieldUser:
			t.Placeholder = "(optional)"
		case fieldPort:
			// Port will be pre-filled with 22
		case fieldProject:
			t.Placeholder = "(optional)"
		case fieldEnv:
			t.Placeholder = "(optional)"
		case fieldTags:
			t.Placeholder = "(optional, comma-separated)"
		}

		inputs[i] = t
	}

	// Pre-fill port with default value
	inputs[fieldPort].SetValue("22")

	m := FormModel{
		inputs: inputs,
		title:  title,
	}

	if conn != nil {
		m.editing = true
		m.originalID = conn.ID
		m.inputs[fieldID].SetValue(conn.ID)
		m.inputs[fieldHost].SetValue(conn.Host)
		m.inputs[fieldUser].SetValue(conn.User)
		if conn.Port != 0 {
			m.inputs[fieldPort].SetValue(strconv.Itoa(conn.Port))
		}
		m.inputs[fieldProject].SetValue(conn.Project)
		m.inputs[fieldEnv].SetValue(conn.Env)
		if len(conn.Tags) > 0 {
			m.inputs[fieldTags].SetValue(strings.Join(conn.Tags, ", "))
		}
	}

	return m
}

func (m FormModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m FormModel) Update(msg tea.Msg) (FormModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			return m, nil
		case "tab", "down":
			m.focused = (m.focused + 1) % fieldCount
			return m, m.updateFocus()
		case "shift+tab", "up":
			m.focused = (m.focused - 1 + fieldCount) % fieldCount
			return m, m.updateFocus()
		case "enter":
			if m.focused == fieldCount-1 {
				m.submitted = true
				return m, nil
			}
			m.focused = (m.focused + 1) % fieldCount
			return m, m.updateFocus()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	cmd := m.updateInput(msg)
	return m, cmd
}

func (m *FormModel) updateFocus() tea.Cmd {
	for i := range m.inputs {
		if formField(i) == m.focused {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
	return textinput.Blink
}

func (m *FormModel) updateInput(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
	return cmd
}

func (m FormModel) View() string {
	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render(m.title))
	b.WriteString("\n\n")

	// Form fields
	labels := []string{"ID:", "Host:", "User:", "Port:", "Project:", "Env:", "Tags:"}

	for i, label := range labels {
		style := helpDescStyle
		if formField(i) == m.focused {
			style = primaryStyle
		}

		b.WriteString(style.Render(fmt.Sprintf("%-10s", label)))
		b.WriteString(m.inputs[i].View())
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Help
	help := helpKeyStyle.Render("tab") + " " + helpDescStyle.Render("next") + "  "
	help += helpKeyStyle.Render("enter") + " " + helpDescStyle.Render("save") + "  "
	help += helpKeyStyle.Render("esc") + " " + helpDescStyle.Render("cancel")
	b.WriteString(help)

	return b.String()
}

func (m FormModel) GetConnection() (*config.Connection, error) {
	id := strings.TrimSpace(m.inputs[fieldID].Value())
	host := strings.TrimSpace(m.inputs[fieldHost].Value())

	if id == "" {
		return nil, fmt.Errorf("ID is required")
	}
	if host == "" {
		return nil, fmt.Errorf("Host is required")
	}

	port := 22
	if portStr := strings.TrimSpace(m.inputs[fieldPort].Value()); portStr != "" {
		p, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid port: %s", portStr)
		}
		port = p
	}

	var tags []string
	if tagsStr := strings.TrimSpace(m.inputs[fieldTags].Value()); tagsStr != "" {
		for _, tag := range strings.Split(tagsStr, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tags = append(tags, tag)
			}
		}
	}

	return &config.Connection{
		ID:      id,
		Host:    host,
		User:    strings.TrimSpace(m.inputs[fieldUser].Value()),
		Port:    port,
		Project: strings.TrimSpace(m.inputs[fieldProject].Value()),
		Env:     strings.TrimSpace(m.inputs[fieldEnv].Value()),
		Tags:    tags,
	}, nil
}

func (m FormModel) Cancelled() bool {
	return m.cancelled
}

func (m FormModel) Submitted() bool {
	return m.submitted
}

func (m FormModel) OriginalID() string {
	return m.originalID
}

func (m FormModel) IsEditing() bool {
	return m.editing
}

var primaryStyle = titleStyle
