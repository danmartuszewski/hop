package ssh

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// TerminalType represents a detected terminal emulator
type TerminalType string

const (
	TerminalUnknown        TerminalType = "unknown"
	TerminalAppleTerminal  TerminalType = "apple-terminal"
	TerminalITerm2         TerminalType = "iterm2"
	TerminalWarp           TerminalType = "warp"
	TerminalAlacritty      TerminalType = "alacritty"
	TerminalWindowsTerminal TerminalType = "windows-terminal"
	TerminalGNOMETerminal  TerminalType = "gnome-terminal"
	TerminalKonsole        TerminalType = "konsole"
	TerminalKitty          TerminalType = "kitty"
)

// DetectTerminal detects the current terminal emulator.
// It first checks HOP_TERMINAL environment variable, then attempts auto-detection.
func DetectTerminal() TerminalType {
	// Check environment variable override first
	if envTerminal := os.Getenv("HOP_TERMINAL"); envTerminal != "" {
		return parseTerminalType(envTerminal)
	}

	switch runtime.GOOS {
	case "darwin":
		return detectMacOSTerminal()
	case "linux":
		return detectLinuxTerminal()
	case "windows":
		return detectWindowsTerminal()
	default:
		return TerminalUnknown
	}
}

func parseTerminalType(s string) TerminalType {
	switch strings.ToLower(s) {
	case "apple-terminal", "terminal", "terminal.app":
		return TerminalAppleTerminal
	case "iterm", "iterm2":
		return TerminalITerm2
	case "warp":
		return TerminalWarp
	case "alacritty":
		return TerminalAlacritty
	case "windows-terminal", "wt":
		return TerminalWindowsTerminal
	case "gnome-terminal", "gnome":
		return TerminalGNOMETerminal
	case "konsole":
		return TerminalKonsole
	case "kitty":
		return TerminalKitty
	default:
		return TerminalUnknown
	}
}

func detectMacOSTerminal() TerminalType {
	// Check TERM_PROGRAM environment variable (most reliable on macOS)
	termProgram := os.Getenv("TERM_PROGRAM")
	switch termProgram {
	case "iTerm.app":
		return TerminalITerm2
	case "Apple_Terminal":
		return TerminalAppleTerminal
	case "WarpTerminal":
		return TerminalWarp
	case "Alacritty":
		return TerminalAlacritty
	case "kitty":
		return TerminalKitty
	}

	// Check for Warp via LC_TERMINAL
	if os.Getenv("LC_TERMINAL") == "iTerm2" {
		// Warp sometimes sets this
		if os.Getenv("WARP_IS_LOCAL_SHELL_SESSION") != "" {
			return TerminalWarp
		}
		return TerminalITerm2
	}

	// Fallback to Apple Terminal on macOS
	return TerminalAppleTerminal
}

func detectLinuxTerminal() TerminalType {
	// Check common environment variables
	termProgram := os.Getenv("TERM_PROGRAM")
	if termProgram != "" {
		switch strings.ToLower(termProgram) {
		case "alacritty":
			return TerminalAlacritty
		case "warpTerminal":
			return TerminalWarp
		case "kitty":
			return TerminalKitty
		}
	}

	// Check GNOME Terminal
	if os.Getenv("GNOME_TERMINAL_SCREEN") != "" || os.Getenv("VTE_VERSION") != "" {
		return TerminalGNOMETerminal
	}

	// Check Konsole
	if os.Getenv("KONSOLE_VERSION") != "" {
		return TerminalKonsole
	}

	// Check Kitty
	if os.Getenv("KITTY_WINDOW_ID") != "" {
		return TerminalKitty
	}

	// Check Alacritty via ALACRITTY_SOCKET
	if os.Getenv("ALACRITTY_SOCKET") != "" {
		return TerminalAlacritty
	}

	return TerminalUnknown
}

func detectWindowsTerminal() TerminalType {
	// Check WT_SESSION environment variable (set by Windows Terminal)
	if os.Getenv("WT_SESSION") != "" {
		return TerminalWindowsTerminal
	}

	// Check for Alacritty
	if os.Getenv("ALACRITTY_SOCKET") != "" {
		return TerminalAlacritty
	}

	return TerminalUnknown
}

// SupportsNewTab returns true if the terminal supports opening new tabs
func (t TerminalType) SupportsNewTab() bool {
	switch t {
	case TerminalAppleTerminal, TerminalITerm2, TerminalWarp,
		TerminalWindowsTerminal, TerminalGNOMETerminal, TerminalKonsole, TerminalKitty:
		return true
	case TerminalAlacritty:
		// Alacritty doesn't support tabs, but can open new windows
		return true
	default:
		return false
	}
}

// String returns a human-readable name for the terminal
func (t TerminalType) String() string {
	switch t {
	case TerminalAppleTerminal:
		return "Apple Terminal"
	case TerminalITerm2:
		return "iTerm2"
	case TerminalWarp:
		return "Warp"
	case TerminalAlacritty:
		return "Alacritty"
	case TerminalWindowsTerminal:
		return "Windows Terminal"
	case TerminalGNOMETerminal:
		return "GNOME Terminal"
	case TerminalKonsole:
		return "Konsole"
	case TerminalKitty:
		return "Kitty"
	default:
		return "Unknown"
	}
}

// OpenNewTab opens a new terminal tab/window and runs the given command.
// The command should be the full SSH command string.
func (t TerminalType) OpenNewTab(command string) error {
	switch t {
	case TerminalAppleTerminal:
		return openAppleTerminalTab(command)
	case TerminalITerm2:
		return openITerm2Tab(command)
	case TerminalWarp:
		return openWarpTab(command)
	case TerminalAlacritty:
		return openAlacrittyWindow(command)
	case TerminalWindowsTerminal:
		return openWindowsTerminalTab(command)
	case TerminalGNOMETerminal:
		return openGNOMETerminalTab(command)
	case TerminalKonsole:
		return openKonsoleTab(command)
	case TerminalKitty:
		return openKittyTab(command)
	default:
		return fmt.Errorf("terminal %q does not support opening new tabs", t)
	}
}

func openAppleTerminalTab(command string) error {
	script := fmt.Sprintf(`
		tell application "Terminal"
			activate
			do script "%s"
		end tell
	`, escapeAppleScript(command))

	return exec.Command("osascript", "-e", script).Run()
}

func openITerm2Tab(command string) error {
	script := fmt.Sprintf(`
		tell application "iTerm"
			activate
			tell current window
				create tab with default profile
				tell current session
					write text "%s"
				end tell
			end tell
		end tell
	`, escapeAppleScript(command))

	return exec.Command("osascript", "-e", script).Run()
}

func openWarpTab(command string) error {
	// Warp uses a different approach - open new tab via AppleScript
	// then execute command
	script := fmt.Sprintf(`
		tell application "Warp"
			activate
			tell application "System Events"
				keystroke "t" using command down
				delay 0.3
				keystroke "%s"
				keystroke return
			end tell
		end tell
	`, escapeAppleScript(command))

	return exec.Command("osascript", "-e", script).Run()
}

func openAlacrittyWindow(command string) error {
	// Alacritty doesn't support tabs, opens a new window instead
	return exec.Command("alacritty", "-e", "sh", "-c", command).Start()
}

func openWindowsTerminalTab(command string) error {
	// Windows Terminal: wt -w 0 nt cmd /c "command"
	return exec.Command("wt", "-w", "0", "nt", "cmd", "/c", command).Start()
}

func openGNOMETerminalTab(command string) error {
	return exec.Command("gnome-terminal", "--tab", "--", "sh", "-c", command+"; exec bash").Start()
}

func openKonsoleTab(command string) error {
	return exec.Command("konsole", "--new-tab", "-e", "sh", "-c", command).Start()
}

func openKittyTab(command string) error {
	return exec.Command("kitty", "@", "new-window", "--", "sh", "-c", command).Start()
}

func escapeAppleScript(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}
