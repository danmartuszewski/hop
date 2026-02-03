package picker

import (
	"fmt"

	"github.com/danmartuszewski/hop/internal/config"
	"github.com/danmartuszewski/hop/internal/fuzzy"
	"github.com/manifoldco/promptui"
)

// SelectConnection presents an interactive picker for selecting a connection
// from multiple fuzzy matches. Returns the selected connection or an error
// if the user cancels or an error occurs.
func SelectConnection(matches []fuzzy.Match) (*config.Connection, error) {
	if len(matches) == 0 {
		return nil, fmt.Errorf("no matches to select from")
	}

	if len(matches) == 1 {
		return matches[0].Connection, nil
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "▸ {{ .Connection.ID | cyan }} ({{ .Connection.Host }})",
		Inactive: "  {{ .Connection.ID }} ({{ .Connection.Host }})",
		Selected: "✓ {{ .Connection.ID | green }}",
		Details: `
--------- Connection ----------
{{ "ID:" | faint }}	{{ .Connection.ID }}
{{ "Host:" | faint }}	{{ .Connection.Host }}
{{ "User:" | faint }}	{{ .Connection.User }}{{ if .Connection.Project }}
{{ "Project:" | faint }}	{{ .Connection.Project }}{{ end }}{{ if .Connection.Env }}
{{ "Env:" | faint }}	{{ .Connection.Env }}{{ end }}`,
	}

	searcher := func(input string, index int) bool {
		match := matches[index]
		return fuzzy.ContainsIgnoreCase(match.Connection.ID, input) ||
			fuzzy.ContainsIgnoreCase(match.Connection.Host, input)
	}

	prompt := promptui.Select{
		Label:     "Select connection",
		Items:     matches,
		Templates: templates,
		Size:      10,
		Searcher:  searcher,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		if err == promptui.ErrInterrupt {
			return nil, fmt.Errorf("selection cancelled")
		}
		return nil, fmt.Errorf("picker error: %w", err)
	}

	return matches[idx].Connection, nil
}
