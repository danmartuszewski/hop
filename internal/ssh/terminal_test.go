package ssh

import (
	"os"
	"runtime"
	"testing"
)

func TestParseTerminalType(t *testing.T) {
	tests := []struct {
		input    string
		expected TerminalType
	}{
		{"apple-terminal", TerminalAppleTerminal},
		{"terminal", TerminalAppleTerminal},
		{"Terminal.app", TerminalAppleTerminal},
		{"iterm", TerminalITerm2},
		{"iterm2", TerminalITerm2},
		{"iTerm2", TerminalITerm2},
		{"warp", TerminalWarp},
		{"Warp", TerminalWarp},
		{"alacritty", TerminalAlacritty},
		{"Alacritty", TerminalAlacritty},
		{"windows-terminal", TerminalWindowsTerminal},
		{"wt", TerminalWindowsTerminal},
		{"gnome-terminal", TerminalGNOMETerminal},
		{"gnome", TerminalGNOMETerminal},
		{"konsole", TerminalKonsole},
		{"kitty", TerminalKitty},
		{"unknown", TerminalUnknown},
		{"", TerminalUnknown},
		{"some-random-terminal", TerminalUnknown},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := parseTerminalType(tc.input)
			if result != tc.expected {
				t.Errorf("parseTerminalType(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestTerminalTypeString(t *testing.T) {
	tests := []struct {
		termType TerminalType
		expected string
	}{
		{TerminalAppleTerminal, "Apple Terminal"},
		{TerminalITerm2, "iTerm2"},
		{TerminalWarp, "Warp"},
		{TerminalAlacritty, "Alacritty"},
		{TerminalWindowsTerminal, "Windows Terminal"},
		{TerminalGNOMETerminal, "GNOME Terminal"},
		{TerminalKonsole, "Konsole"},
		{TerminalKitty, "Kitty"},
		{TerminalUnknown, "Unknown"},
	}

	for _, tc := range tests {
		t.Run(string(tc.termType), func(t *testing.T) {
			result := tc.termType.String()
			if result != tc.expected {
				t.Errorf("%q.String() = %q, want %q", tc.termType, result, tc.expected)
			}
		})
	}
}

func TestTerminalTypeSupportsNewTab(t *testing.T) {
	supportsTab := []TerminalType{
		TerminalAppleTerminal,
		TerminalITerm2,
		TerminalWarp,
		TerminalAlacritty,
		TerminalWindowsTerminal,
		TerminalGNOMETerminal,
		TerminalKonsole,
		TerminalKitty,
	}

	doesNotSupportTab := []TerminalType{
		TerminalUnknown,
	}

	for _, tt := range supportsTab {
		t.Run(string(tt)+"_supports", func(t *testing.T) {
			if !tt.SupportsNewTab() {
				t.Errorf("%q.SupportsNewTab() = false, want true", tt)
			}
		})
	}

	for _, tt := range doesNotSupportTab {
		t.Run(string(tt)+"_does_not_support", func(t *testing.T) {
			if tt.SupportsNewTab() {
				t.Errorf("%q.SupportsNewTab() = true, want false", tt)
			}
		})
	}
}

func TestDetectTerminalWithEnvOverride(t *testing.T) {
	// Save and restore original env
	origEnv := os.Getenv("HOP_TERMINAL")
	defer os.Setenv("HOP_TERMINAL", origEnv)

	tests := []struct {
		envValue string
		expected TerminalType
	}{
		{"iterm2", TerminalITerm2},
		{"warp", TerminalWarp},
		{"alacritty", TerminalAlacritty},
		{"gnome-terminal", TerminalGNOMETerminal},
	}

	for _, tc := range tests {
		t.Run(tc.envValue, func(t *testing.T) {
			os.Setenv("HOP_TERMINAL", tc.envValue)
			result := DetectTerminal()
			if result != tc.expected {
				t.Errorf("DetectTerminal() with HOP_TERMINAL=%q = %q, want %q",
					tc.envValue, result, tc.expected)
			}
		})
	}
}

func TestDetectMacOSTerminal(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test")
	}

	// Save original env vars
	origTermProgram := os.Getenv("TERM_PROGRAM")
	origLCTerminal := os.Getenv("LC_TERMINAL")
	origWarpLocal := os.Getenv("WARP_IS_LOCAL_SHELL_SESSION")
	origHopTerminal := os.Getenv("HOP_TERMINAL")
	defer func() {
		os.Setenv("TERM_PROGRAM", origTermProgram)
		os.Setenv("LC_TERMINAL", origLCTerminal)
		os.Setenv("WARP_IS_LOCAL_SHELL_SESSION", origWarpLocal)
		os.Setenv("HOP_TERMINAL", origHopTerminal)
	}()

	// Clear HOP_TERMINAL to test auto-detection
	os.Unsetenv("HOP_TERMINAL")

	tests := []struct {
		name        string
		termProgram string
		lcTerminal  string
		warpLocal   string
		expected    TerminalType
	}{
		{"iTerm2", "iTerm.app", "", "", TerminalITerm2},
		{"Apple Terminal", "Apple_Terminal", "", "", TerminalAppleTerminal},
		{"Warp via TERM_PROGRAM", "WarpTerminal", "", "", TerminalWarp},
		{"Alacritty", "Alacritty", "", "", TerminalAlacritty},
		{"Kitty", "kitty", "", "", TerminalKitty},
		{"Warp via LC_TERMINAL", "", "iTerm2", "true", TerminalWarp},
		{"iTerm2 via LC_TERMINAL", "", "iTerm2", "", TerminalITerm2},
		{"Default to Apple Terminal", "", "", "", TerminalAppleTerminal},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			os.Setenv("TERM_PROGRAM", tc.termProgram)
			os.Setenv("LC_TERMINAL", tc.lcTerminal)
			if tc.warpLocal != "" {
				os.Setenv("WARP_IS_LOCAL_SHELL_SESSION", tc.warpLocal)
			} else {
				os.Unsetenv("WARP_IS_LOCAL_SHELL_SESSION")
			}

			result := detectMacOSTerminal()
			if result != tc.expected {
				t.Errorf("detectMacOSTerminal() = %q, want %q", result, tc.expected)
			}
		})
	}
}

func TestDetectLinuxTerminal(t *testing.T) {
	// Save original env vars
	origTermProgram := os.Getenv("TERM_PROGRAM")
	origGnomeScreen := os.Getenv("GNOME_TERMINAL_SCREEN")
	origVTE := os.Getenv("VTE_VERSION")
	origKonsole := os.Getenv("KONSOLE_VERSION")
	origKitty := os.Getenv("KITTY_WINDOW_ID")
	origAlacritty := os.Getenv("ALACRITTY_SOCKET")
	defer func() {
		os.Setenv("TERM_PROGRAM", origTermProgram)
		os.Setenv("GNOME_TERMINAL_SCREEN", origGnomeScreen)
		os.Setenv("VTE_VERSION", origVTE)
		os.Setenv("KONSOLE_VERSION", origKonsole)
		os.Setenv("KITTY_WINDOW_ID", origKitty)
		os.Setenv("ALACRITTY_SOCKET", origAlacritty)
	}()

	// Clear all env vars first
	os.Unsetenv("TERM_PROGRAM")
	os.Unsetenv("GNOME_TERMINAL_SCREEN")
	os.Unsetenv("VTE_VERSION")
	os.Unsetenv("KONSOLE_VERSION")
	os.Unsetenv("KITTY_WINDOW_ID")
	os.Unsetenv("ALACRITTY_SOCKET")

	tests := []struct {
		name     string
		setup    func()
		expected TerminalType
	}{
		{
			name: "Alacritty via TERM_PROGRAM",
			setup: func() {
				os.Setenv("TERM_PROGRAM", "alacritty")
			},
			expected: TerminalAlacritty,
		},
		{
			name: "Kitty via TERM_PROGRAM",
			setup: func() {
				os.Setenv("TERM_PROGRAM", "kitty")
			},
			expected: TerminalKitty,
		},
		{
			name: "GNOME Terminal via screen",
			setup: func() {
				os.Setenv("GNOME_TERMINAL_SCREEN", "/org/gnome/Terminal/screen/0")
			},
			expected: TerminalGNOMETerminal,
		},
		{
			name: "GNOME Terminal via VTE",
			setup: func() {
				os.Setenv("VTE_VERSION", "6800")
			},
			expected: TerminalGNOMETerminal,
		},
		{
			name: "Konsole",
			setup: func() {
				os.Setenv("KONSOLE_VERSION", "220401")
			},
			expected: TerminalKonsole,
		},
		{
			name: "Kitty via KITTY_WINDOW_ID",
			setup: func() {
				os.Setenv("KITTY_WINDOW_ID", "1")
			},
			expected: TerminalKitty,
		},
		{
			name: "Alacritty via socket",
			setup: func() {
				os.Setenv("ALACRITTY_SOCKET", "/run/user/1000/Alacritty-:0-1234.sock")
			},
			expected: TerminalAlacritty,
		},
		{
			name:     "Unknown terminal",
			setup:    func() {},
			expected: TerminalUnknown,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Clear env vars for each test
			os.Unsetenv("TERM_PROGRAM")
			os.Unsetenv("GNOME_TERMINAL_SCREEN")
			os.Unsetenv("VTE_VERSION")
			os.Unsetenv("KONSOLE_VERSION")
			os.Unsetenv("KITTY_WINDOW_ID")
			os.Unsetenv("ALACRITTY_SOCKET")

			tc.setup()
			result := detectLinuxTerminal()
			if result != tc.expected {
				t.Errorf("detectLinuxTerminal() = %q, want %q", result, tc.expected)
			}
		})
	}
}

func TestDetectWindowsTerminal(t *testing.T) {
	// Save original env vars
	origWTSession := os.Getenv("WT_SESSION")
	origAlacritty := os.Getenv("ALACRITTY_SOCKET")
	defer func() {
		os.Setenv("WT_SESSION", origWTSession)
		os.Setenv("ALACRITTY_SOCKET", origAlacritty)
	}()

	// Clear all env vars first
	os.Unsetenv("WT_SESSION")
	os.Unsetenv("ALACRITTY_SOCKET")

	tests := []struct {
		name     string
		setup    func()
		expected TerminalType
	}{
		{
			name: "Windows Terminal",
			setup: func() {
				os.Setenv("WT_SESSION", "some-guid")
			},
			expected: TerminalWindowsTerminal,
		},
		{
			name: "Alacritty on Windows",
			setup: func() {
				os.Setenv("ALACRITTY_SOCKET", "some-socket")
			},
			expected: TerminalAlacritty,
		},
		{
			name:     "Unknown terminal",
			setup:    func() {},
			expected: TerminalUnknown,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			os.Unsetenv("WT_SESSION")
			os.Unsetenv("ALACRITTY_SOCKET")

			tc.setup()
			result := detectWindowsTerminal()
			if result != tc.expected {
				t.Errorf("detectWindowsTerminal() = %q, want %q", result, tc.expected)
			}
		})
	}
}

func TestEscapeAppleScript(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"with spaces", "with spaces"},
		{`with "quotes"`, `with \"quotes\"`},
		{`with \backslash`, `with \\backslash`},
		{`both "quotes" and \backslash`, `both \"quotes\" and \\backslash`},
		{"ssh user@host", "ssh user@host"},
		{`ssh -o "StrictHostKeyChecking=no" user@host`, `ssh -o \"StrictHostKeyChecking=no\" user@host`},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := escapeAppleScript(tc.input)
			if result != tc.expected {
				t.Errorf("escapeAppleScript(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}
