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
	TerminalGhostty        TerminalType = "ghostty"
	TerminalCmux           TerminalType = "cmux"
	TerminalWezTerm        TerminalType = "wezterm"
	TerminalTilix          TerminalType = "tilix"
	TerminalTerminator     TerminalType = "terminator"
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
	case "ghostty":
		return TerminalGhostty
	case "cmux":
		return TerminalCmux
	case "wezterm":
		return TerminalWezTerm
	case "tilix":
		return TerminalTilix
	case "terminator":
		return TerminalTerminator
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
	case "WezTerm":
		return TerminalWezTerm
	case "ghostty":
		// cmux embeds libghostty and sets TERM_PROGRAM=ghostty too; disambiguate
		// via CMUX_WORKSPACE_ID, which cmux guarantees in its spawned shells.
		if os.Getenv("CMUX_WORKSPACE_ID") != "" {
			return TerminalCmux
		}
		return TerminalGhostty
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
		case "ghostty":
			return TerminalGhostty
		case "wezterm":
			return TerminalWezTerm
		}
	}

	// WezTerm clears TERM_PROGRAM under sudo; WEZTERM_PANE is the more specific signal.
	if os.Getenv("WEZTERM_PANE") != "" {
		return TerminalWezTerm
	}

	// Ghostty sets GHOSTTY_RESOURCES_DIR in its shell integration
	if os.Getenv("GHOSTTY_RESOURCES_DIR") != "" {
		return TerminalGhostty
	}

	// Tilix and Terminator both set VTE_VERSION, so check them before the
	// GNOME Terminal branch below.
	if os.Getenv("TILIX_ID") != "" {
		return TerminalTilix
	}
	if os.Getenv("TERMINATOR_UUID") != "" {
		return TerminalTerminator
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

	// WezTerm sets TERM_PROGRAM=WezTerm (or WEZTERM_PANE) on Windows too.
	if strings.EqualFold(os.Getenv("TERM_PROGRAM"), "wezterm") || os.Getenv("WEZTERM_PANE") != "" {
		return TerminalWezTerm
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
		TerminalWindowsTerminal, TerminalGNOMETerminal, TerminalKonsole, TerminalKitty,
		TerminalGhostty, TerminalCmux, TerminalWezTerm, TerminalTilix, TerminalTerminator:
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
	case TerminalGhostty:
		return "Ghostty"
	case TerminalCmux:
		return "cmux"
	case TerminalWezTerm:
		return "WezTerm"
	case TerminalTilix:
		return "Tilix"
	case TerminalTerminator:
		return "Terminator"
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
	case TerminalGhostty:
		return openGhosttyTab(command)
	case TerminalCmux:
		return openCmuxTab(command)
	case TerminalWezTerm:
		return openWezTermTab(command)
	case TerminalTilix:
		return openTilixTab(command)
	case TerminalTerminator:
		return openTerminatorTab(command)
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

func openGhosttyTab(command string) error {
	// Ghostty has no AppleScript "do script" equivalent, so on macOS we
	// activate the app and drive a new tab via System Events: Cmd+T, type
	// the command, press Return. Requires Accessibility permission for
	// whichever process is running hop (same constraint as the Warp path).
	if runtime.GOOS == "darwin" {
		script := fmt.Sprintf(`
			tell application "Ghostty"
				activate
			end tell
			delay 0.2
			tell application "System Events"
				keystroke "t" using command down
				delay 0.3
				keystroke "%s"
				keystroke return
			end tell
		`, escapeAppleScript(command))

		return exec.Command("osascript", "-e", script).Run()
	}

	// Linux / other: spawn a new Ghostty window running the command.
	return exec.Command("ghostty", "-e", "sh", "-c", command).Start()
}

// cmuxBundledCLI is where the Homebrew cask install of cmux.app keeps its CLI
// when it isn't symlinked onto PATH.
const cmuxBundledCLI = "/Applications/cmux.app/Contents/Resources/bin/cmux"

func openCmuxTab(command string) error {
	// cmux is macOS-only (native Swift + AppKit app).
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("cmux is only supported on macOS")
	}

	bin, err := exec.LookPath("cmux")
	if err != nil {
		bin = cmuxBundledCLI
	}

	// `cmux new-workspace --command <cmd>` creates a new workspace tab and types
	// the command into its shell, atomic via cmux's socket API — no AppleScript
	// keystroke synthesis, no Accessibility permission required.
	return exec.Command(bin, "new-workspace", "--command", command).Start()
}

func openWezTermTab(command string) error {
	// `wezterm cli spawn` adds a tab to the running GUI via its mux socket. If
	// no GUI is running that fails, so fall back to `wezterm start` which
	// launches a fresh window.
	err := exec.Command("wezterm", "cli", "spawn", "--", "sh", "-c", command).Run()
	if err == nil {
		return nil
	}
	return exec.Command("wezterm", "start", "--", "sh", "-c", command).Start()
}

func openTilixTab(command string) error {
	// Tilix has no "new tab" action, only splits. session-add-right is the
	// closest "new pane alongside" primitive. `-e` takes a single string that
	// Tilix word-splits, so we build a shell invocation and escape single
	// quotes in the user command.
	escaped := strings.ReplaceAll(command, "'", `'\''`)
	arg := fmt.Sprintf("sh -c '%s; exec bash'", escaped)
	return exec.Command("tilix", "--action=session-add-right", "-e", arg).Start()
}

func openTerminatorTab(command string) error {
	// `--new-tab` routes via DBus to the first running Terminator instance,
	// or opens a new window with one tab if none is running. `-x` consumes
	// the rest of argv as the child command (no string-splitting surprises).
	return exec.Command("terminator", "--new-tab", "-x", "sh", "-c", command+"; exec bash").Start()
}

func escapeAppleScript(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}
