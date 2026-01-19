# hop - Implementation Plan

> Track implementation progress by marking tasks as done: `[x]`

---

## Tech Stack

| Component | Technology | Rationale |
|-----------|------------|-----------|
| CLI + TUI | Go | Single binary, fast startup, excellent cross-platform |
| TUI Framework | [Bubble Tea](https://github.com/charmbracelet/bubbletea) | Best-in-class Go TUI library |
| Styling | [Lip Gloss](https://github.com/charmbracelet/lipgloss) | Companion styling library |
| Fuzzy matching | [go-fuzzyfinder](https://github.com/ktr0731/go-fuzzyfinder) or custom | Fast, accurate |
| Config parsing | [yaml.v3](https://gopkg.in/yaml.v3) | Standard YAML library |
| CLI framework | [Cobra](https://github.com/spf13/cobra) | Industry standard CLI |
| Raycast Extension | TypeScript | Required by Raycast |

---

## Project Structure

```
hop/
├── cmd/
│   └── hop/
│       └── main.go           # Entry point
├── internal/
│   ├── config/
│   │   ├── config.go         # Config parsing
│   │   └── validate.go       # Config validation
│   ├── ssh/
│   │   ├── connect.go        # SSH connection handling
│   │   ├── exec.go           # Multi-exec logic
│   │   └── terminal.go       # Terminal detection & tab opening
│   ├── tui/
│   │   ├── app.go            # Bubble Tea app
│   │   ├── list.go           # Connection list view
│   │   ├── filter.go         # Fuzzy filter component
│   │   ├── help.go           # Help overlay
│   │   └── styles.go         # Styling
│   └── fuzzy/
│       └── matcher.go        # Fuzzy matching logic
├── raycast/
│   └── hop-extension/        # Raycast extension (TypeScript)
│       ├── package.json
│       ├── src/
│       │   ├── connect.tsx   # Quick connect command
│       │   ├── open.tsx      # Open group command
│       │   ├── exec.tsx      # Exec command
│       │   └── dashboard.tsx # Dashboard command
│       └── assets/
├── Makefile
├── go.mod
├── go.sum
├── README.md
├── LICENSE
└── .goreleaser.yaml          # Release automation
```

---

## Phase 1: Project Setup & Core Infrastructure

### 1.1 Project Initialization
- [ ] Create Go module (`go mod init github.com/<username>/hop`)
- [ ] Set up project directory structure
- [ ] Create Makefile with build targets
- [ ] Add .gitignore
- [ ] Create LICENSE (MIT)
- [ ] Initialize git repository

### 1.2 Configuration System
- [ ] Define config structs (`Config`, `Connection`, `Defaults`, `Group`)
- [ ] Implement YAML config loading (`internal/config/config.go`)
- [ ] Implement config validation (`internal/config/validate.go`)
- [ ] Support `~/.config/hop/config.yaml` default path
- [ ] Support `HOP_CONFIG` environment variable override
- [ ] Support `--config` flag override
- [ ] Add config file creation on first run (with example)
- [ ] Write unit tests for config parsing

### 1.3 CLI Framework Setup
- [ ] Set up Cobra root command
- [ ] Implement `hop version` command
- [ ] Implement `hop help` command
- [ ] Implement global flags (`--config`, `--verbose`, `--quiet`)
- [ ] Set up command structure for subcommands

---

## Phase 2: Core SSH Functionality

### 2.1 Single Connection
- [ ] Implement SSH command builder (`internal/ssh/connect.go`)
- [ ] Support basic connection (host, user, port)
- [ ] Support identity file (`-i` flag)
- [ ] Support custom SSH options
- [ ] Implement `hop connect <id>` command (exact match)
- [ ] Implement `hop <query>` quick connect (fuzzy match)
- [ ] Add `--dry-run` flag (print SSH command)
- [ ] Add `-t` flag (force TTY)
- [ ] Add `-- <cmd>` support (run command on remote)
- [ ] Write unit tests for SSH command building

### 2.2 Fuzzy Matching
- [ ] Implement fuzzy matcher (`internal/fuzzy/matcher.go`)
- [ ] Match against connection ID
- [ ] Match against hostname
- [ ] Match against tags
- [ ] Implement scoring (shorter matches win)
- [ ] Implement interactive picker when multiple matches
- [ ] Write unit tests for fuzzy matching

### 2.3 Connection Listing
- [ ] Implement `hop list` command
- [ ] Display grouped by project → env
- [ ] Add `--json` flag for JSON output
- [ ] Add `--flat` flag for ungrouped list

---

## Phase 3: Multi-Open (Batch Connect)

### 3.1 Terminal Detection
- [ ] Implement terminal detection (`internal/ssh/terminal.go`)
- [ ] Detect macOS Terminal.app
- [ ] Detect iTerm2
- [ ] Detect Warp
- [ ] Detect Alacritty
- [ ] Detect Windows Terminal
- [ ] Detect GNOME Terminal
- [ ] Support `HOP_TERMINAL` override

### 3.2 Tab Opening
- [ ] Implement new tab opening for each terminal
- [ ] macOS: Terminal.app AppleScript
- [ ] macOS: iTerm2 AppleScript
- [ ] macOS: Warp (CLI integration)
- [ ] Linux: GNOME Terminal (`--tab`)
- [ ] Linux: Alacritty (new window fallback)
- [ ] Windows: Windows Terminal (`wt -w 0 nt`)
- [ ] Add configurable delay between tabs

### 3.3 Open Command
- [ ] Implement `hop open <group>` command
- [ ] Implement `hop open <id1> <id2> ...` (multiple IDs)
- [ ] Support fuzzy matching for groups and IDs
- [ ] Implement `-- "<cmd>"` for initial command
- [ ] Implement `--tag=<tag>` filter
- [ ] Implement `--dry-run` flag
- [ ] Write integration tests

---

## Phase 4: Multi-Exec (Non-Interactive)

### 4.1 Parallel Execution Engine
- [ ] Implement parallel SSH execution (`internal/ssh/exec.go`)
- [ ] Configurable parallelism (`--parallel=N`)
- [ ] Collect stdout/stderr per host
- [ ] Implement timeout handling (`--timeout`)
- [ ] Implement fail-fast mode (`--fail-fast`)

### 4.2 Output Formatting
- [ ] Implement grouped output mode (default)
- [ ] Implement stream mode (`--stream`)
- [ ] Add host prefixes in stream mode
- [ ] Colorize output per host
- [ ] Handle exit codes properly

### 4.3 Exec Command
- [ ] Implement `hop exec <group> "<command>"` command
- [ ] Support pattern matching for groups (`"web*"`)
- [ ] Support tag filtering
- [ ] Write integration tests

---

## Phase 5: TUI Dashboard

### 5.1 Basic TUI Structure
- [ ] Set up Bubble Tea app (`internal/tui/app.go`)
- [ ] Implement main model with states
- [ ] Set up Lip Gloss styles (`internal/tui/styles.go`)
- [ ] Implement header component
- [ ] Implement footer/status bar
- [ ] Implement quit handling (`q`, `Ctrl+C`)

### 5.2 Connection List View
- [ ] Implement tree view (`internal/tui/list.go`)
- [ ] Group connections by project → env
- [ ] Implement navigation (`↑/↓`, `j/k`)
- [ ] Implement group collapse/expand (`Tab`)
- [ ] Highlight selected item
- [ ] Display connection details (host, user)

### 5.3 Fuzzy Filter
- [ ] Implement filter input (`internal/tui/filter.go`)
- [ ] Activate filter with `/`
- [ ] Real-time filtering as you type
- [ ] Clear filter with `Esc`
- [ ] Highlight matching characters

### 5.4 Multi-Select Mode
- [ ] Implement selection state per connection
- [ ] Toggle selection with `Space`
- [ ] Visual indicator for selected items (`●`/`○`)
- [ ] Show selection count in header
- [ ] Select all visible (`a`)
- [ ] Deselect all (`A`)
- [ ] Clear selection with `Esc`

### 5.5 Actions
- [ ] Connect on `Enter` (single or multi)
- [ ] Execute command on `x` (prompt for command)
- [ ] Edit config on `e` (open in `$EDITOR`)
- [ ] Reload config on `r`
- [ ] Help overlay on `?`

### 5.6 Help Overlay
- [ ] Implement help view (`internal/tui/help.go`)
- [ ] Show all keybindings
- [ ] Toggle with `?`
- [ ] Dismiss with `Esc` or `?`

---

## Phase 6: Config Command

### 6.1 Config Subcommand
- [ ] Implement `hop config` (open in $EDITOR)
- [ ] Implement `hop config --validate`
- [ ] Implement `hop config --path` (print path)
- [ ] Implement `hop config --init` (create example config)
- [ ] Show validation errors with line numbers

---

## Phase 7: Raycast Extension

### 7.1 Extension Setup
- [ ] Initialize Raycast extension (`raycast/hop-extension/`)
- [ ] Set up TypeScript + React
- [ ] Configure package.json with commands
- [ ] Implement config file reading (shared with CLI)
- [ ] Add extension icon/assets

### 7.2 Connect Command
- [ ] Implement connection list view
- [ ] Group by project → env
- [ ] Fuzzy search
- [ ] `Enter` → open terminal tab
- [ ] `Cmd+Shift+C` → copy SSH command

### 7.3 Multi-Select
- [ ] Implement `Cmd+K` toggle multi-select
- [ ] Selection state management
- [ ] `Cmd+A` select all visible
- [ ] `Cmd+Enter` open all selected

### 7.4 Open with Command
- [ ] Implement `Cmd+Shift+O` action
- [ ] Prompt for initial command
- [ ] Open all with command running

### 7.5 Exec Command
- [ ] Implement `hop exec` Raycast command
- [ ] Server selection UI
- [ ] Command input
- [ ] Results display in Raycast
- [ ] `Cmd+C` copy results

### 7.6 Dashboard Command
- [ ] Implement `hop dashboard` command
- [ ] Launch TUI in terminal

---

## Phase 8: Build & Release

### 8.1 Build System
- [ ] Makefile: `make build`
- [ ] Makefile: `make test`
- [ ] Makefile: `make lint` (golangci-lint)
- [ ] Makefile: `make clean`
- [ ] Set up CI (GitHub Actions)
- [ ] Run tests on PR
- [ ] Run linter on PR

### 8.2 Release Automation
- [ ] Set up goreleaser (`.goreleaser.yaml`)
- [ ] Build for darwin/amd64
- [ ] Build for darwin/arm64
- [ ] Build for linux/amd64
- [ ] Build for linux/arm64
- [ ] Build for windows/amd64
- [ ] Generate checksums
- [ ] GitHub Release on tag

### 8.3 Distribution
- [ ] Create Homebrew tap repository
- [ ] Write Homebrew formula
- [ ] Test `brew install`
- [ ] Document binary download
- [ ] Document `go install`

### 8.4 Raycast Store
- [ ] Prepare extension for submission
- [ ] Add screenshots
- [ ] Write store description
- [ ] Submit to Raycast Store

---

## Phase 9: Documentation & Polish

### 9.1 README
- [ ] Installation instructions
- [ ] Quick start guide
- [ ] Feature overview with examples
- [ ] Configuration reference
- [ ] CLI reference
- [ ] Screenshots/GIFs of TUI

### 9.2 Polish
- [ ] Error messages (user-friendly)
- [ ] Edge case handling
- [ ] Performance optimization
- [ ] Memory usage check
- [ ] Cross-platform testing

### 9.3 Testing
- [ ] Unit test coverage > 70%
- [ ] Integration tests for CLI commands
- [ ] Manual testing on macOS
- [ ] Manual testing on Linux
- [ ] Manual testing on Windows

---

## Progress Summary

| Phase | Description | Status |
|-------|-------------|--------|
| 1 | Project Setup & Core Infrastructure | Not Started |
| 2 | Core SSH Functionality | Not Started |
| 3 | Multi-Open (Batch Connect) | Not Started |
| 4 | Multi-Exec (Non-Interactive) | Not Started |
| 5 | TUI Dashboard | Not Started |
| 6 | Config Command | Not Started |
| 7 | Raycast Extension | Not Started |
| 8 | Build & Release | Not Started |
| 9 | Documentation & Polish | Not Started |

---

## Notes

- Start with Phase 1-2 for MVP (quick connect works)
- Phase 3-4 adds power features
- Phase 5 adds visual dashboard
- Phase 7 (Raycast) can be developed in parallel after Phase 2
- Phase 8-9 are ongoing throughout development
