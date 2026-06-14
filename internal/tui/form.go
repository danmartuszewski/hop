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
	fieldIdentity
	fieldRemoteDir
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
	original   config.Connection
	width      int
	height     int
	cancelled  bool
	submitted  bool
	errMsg     string
}

func NewFormModel(title string, conn *config.Connection) FormModel {
	inputs := make([]textinput.Model, fieldCount)

	for i := range inputs {
		t := textinput.New()
		t.CharLimit = 100

		switch formField(i) {
		case fieldID:
			t.Focus()
		case fieldHost:
			// required, no placeholder
		case fieldUser:
			// optional
		case fieldPort:
			// Port will be pre-filled with 22
		case fieldIdentity:
			// optional - path to a private key
		case fieldRemoteDir:
			// optional - directory to land in after connecting
		case fieldProject:
			// optional
		case fieldEnv:
			// optional
		case fieldTags:
			// No placeholder - we'll show hint text instead
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
		m.original = conn.Clone()
		fillInputs(m.inputs, conn)
	}

	return m
}

// NewDuplicateFormModel builds a form pre-filled from src for creating a copy
// of an existing connection. Unlike the edit form, it is not in editing mode:
// the result is saved as a brand-new connection. The ID is pre-filled with a
// collision-free suggestion (e.g. "web-prod-copy") so the user can save
// immediately, and the original connection is deep-cloned so the copy shares no
// mutable state (Tags, Options, UseMosh) with the source.
func NewDuplicateFormModel(src *config.Connection, suggestedID string) FormModel {
	dup := src.Clone()
	dup.ID = suggestedID

	m := NewFormModel(fmt.Sprintf("Add Connection — copy of %q", src.ID), nil)
	m.original = dup
	fillInputs(m.inputs, &dup)

	return m
}

// fillInputs pre-fills the form's text inputs from a connection. The unexposed
// fields (proxy jump, agent forwarding, mosh, options) are not shown here; they
// are preserved separately via FormModel.original.
func fillInputs(inputs []textinput.Model, conn *config.Connection) {
	inputs[fieldID].SetValue(conn.ID)
	inputs[fieldHost].SetValue(conn.Host)
	inputs[fieldUser].SetValue(conn.User)
	if conn.Port != 0 {
		inputs[fieldPort].SetValue(strconv.Itoa(conn.Port))
	}
	inputs[fieldIdentity].SetValue(conn.IdentityFile)
	inputs[fieldRemoteDir].SetValue(conn.RemoteDir)
	inputs[fieldProject].SetValue(conn.Project)
	inputs[fieldEnv].SetValue(conn.Env)
	if len(conn.Tags) > 0 {
		inputs[fieldTags].SetValue(strings.Join(conn.Tags, ", "))
	}
}

func (m FormModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m FormModel) Update(msg tea.Msg) (FormModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Any keypress clears a stale validation error; a fresh one is set
		// again after this Update if the next save still fails.
		m.errMsg = ""
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			return m, nil
		case "ctrl+s":
			m.submitted = true
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
	labels := []string{"ID:", "Host:", "User:", "Port:", "Identity:", "Remote Dir:", "Project:", "Env:", "Tags:"}

	for i, label := range labels {
		style := helpDescStyle
		if formField(i) == m.focused {
			style = primaryStyle
		}

		b.WriteString(style.Render(fmt.Sprintf("%-10s", label)))
		b.WriteString(m.inputs[i].View())

		// Add hints for fields that benefit from clarification
		switch formField(i) {
		case fieldIdentity:
			b.WriteString(" ")
			b.WriteString(helpDescStyle.Render("(path to private key)"))
		case fieldRemoteDir:
			b.WriteString(" ")
			b.WriteString(helpDescStyle.Render("(directory to start in)"))
		case fieldTags:
			b.WriteString(" ")
			b.WriteString(helpDescStyle.Render("(comma-separated)"))
		}

		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Validation error (e.g. duplicate ID), if any
	if m.errMsg != "" {
		b.WriteString(warningStyle.Render(m.errMsg))
		b.WriteString("\n\n")
	}

	// Help
	help := helpKeyStyle.Render("tab") + " " + helpDescStyle.Render("next") + "  "
	help += helpKeyStyle.Render("enter") + " " + helpDescStyle.Render("next/save") + "  "
	help += helpKeyStyle.Render("ctrl+s") + " " + helpDescStyle.Render("save") + "  "
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

	// Start from the original connection so fields not exposed in the form
	// (proxy jump, agent forwarding, mosh, extra options) are preserved when
	// editing or duplicating. For a brand-new connection this is the zero value.
	// Clone so the returned connection shares no mutable state (Options,
	// UseMosh) with the source — important on the duplicate path.
	conn := m.original.Clone()
	conn.ID = id
	conn.Host = host
	conn.User = strings.TrimSpace(m.inputs[fieldUser].Value())
	conn.Port = port
	conn.IdentityFile = strings.TrimSpace(m.inputs[fieldIdentity].Value())
	conn.RemoteDir = strings.TrimSpace(m.inputs[fieldRemoteDir].Value())
	conn.Project = strings.TrimSpace(m.inputs[fieldProject].Value())
	conn.Env = strings.TrimSpace(m.inputs[fieldEnv].Value())
	conn.Tags = tags
	return &conn, nil
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
