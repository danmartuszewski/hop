package tui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestWarningStyleHasWarningColor(t *testing.T) {
	resetTheme(t)
	currentTheme = defaultTheme
	refreshStyles()

	if got, want := warningStyle.GetForeground(), lipgloss.Color("214"); got != want {
		t.Errorf("warningStyle foreground = %q, want %q", got, want)
	}
}
