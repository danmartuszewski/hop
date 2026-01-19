# hop - SSH Connection Manager

> Quick, elegant SSH connection management for terminal enthusiasts

## Overview

**hop** is a cross-platform SSH connection manager that combines a beautiful TUI dashboard with lightning-fast CLI access. Manage your servers, connect in seconds, and execute commands across multiple hosts.

```
hop prod-web                # fuzzy match & connect
hop                         # open TUI dashboard
hop open prod               # open terminal tab for each server in group
hop open prod -- "htop"     # open all with initial command
hop exec prod "uptime"      # run command on all prod servers (non-interactive)
```

## Core Principles

1. **Speed first** - Connect with minimal keystrokes via fuzzy matching
2. **Visual when needed** - TUI dashboard for browsing and discovery
3. **Simple config** - Single YAML file, human-readable
4. **Cross-platform** - macOS, Linux, Windows
5. **No lock-in** - Uses standard SSH under the hood

---

## Features

### 1. Quick Connect (CLI)

```bash
# Fuzzy match connection ID
hop prod          # matches "myapp-prod-web1"
hop web1          # matches first *web1* server
hop myapp-prod    # exact or fuzzy match

# Direct connect by exact ID
hop myapp-prod-web1
```

**Fuzzy matching rules:**
- Matches against connection ID, host, and tags
- Shortest match wins on ambiguity
- Interactive picker shown if multiple matches

### 2. Multi-Open (Batch Connect)

Open multiple terminal tabs at once, each connected to a different server. Optionally run an initial command on each.

```bash
# Open all servers in a group (each in separate terminal tab)
hop open myapp-prod                    # Opens 3 tabs for 3 prod servers

# Open specific connections
hop open web1 db1 api1                 # Opens 3 specific servers

# Open with initial command running on each
hop open myapp-prod -- "htop"          # Opens all with htop running
hop open myapp-prod -- "tail -f /var/log/app.log"

# Fuzzy match group or multiple IDs
hop open prod                          # Matches myapp-prod group
hop open web1 web2                     # Fuzzy matches

# Combine with tags
hop open --tag=web                     # All servers tagged 'web'
hop open --tag=database -- "psql"      # All DB servers with psql
```

**Behavior:**
- Each connection opens in a new terminal tab (uses system terminal or `$HOP_TERMINAL`)
- Tabs open in quick succession (~100ms delay between each)
- With `-- "command"`: runs command after SSH connects (interactive session stays open)
- Without command: drops into normal shell

**Terminal support:**
| Terminal | macOS | Linux | Windows |
|----------|-------|-------|---------|
| Default system terminal | ✓ | ✓ | ✓ |
| iTerm2 | ✓ | - | - |
| Warp | ✓ | ✓ | - |
| Alacritty | ✓ | ✓ | ✓ |
| Windows Terminal | - | - | ✓ |
| GNOME Terminal | - | ✓ | - |

### 3. TUI Dashboard

```
┌─────────────────────────────────────────────────────────────────┐
│  hop - SSH Connection Manager                          v0.1.0  │
├─────────────────────────────────────────────────────────────────┤
│  🔍 Filter: _                                                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ▼ myapp (3)                                                    │
│    ├─ ▼ prod (2)                                                │
│    │    ├─ web1        web1.myapp.com          user@host       │
│    │    └─ db          db.myapp.com            admin@host      │
│    └─ ▼ staging (1)                                             │
│         └─ web1        staging.myapp.com       user@host       │
│                                                                 │
│  ▼ client-project (2)                                           │
│    └─ ▼ prod (2)                                                │
│         ├─ api         10.0.1.50               deploy@host     │
│         └─ worker      10.0.1.51               deploy@host     │
│                                                                 │
├─────────────────────────────────────────────────────────────────┤
│  ↑↓ navigate  │  enter connect  │  / filter  │  ? help  │  q quit│
└─────────────────────────────────────────────────────────────────┘
```

**TUI Features:**
- Grouped by project → environment
- Fuzzy filter (updates as you type)
- Collapsible groups
- Keyboard-driven navigation
- **Multi-select mode** for batch operations
- Connection details panel (optional)

**Keybindings:**
| Key | Action |
|-----|--------|
| `↑/↓` or `j/k` | Navigate |
| `Enter` | Connect to selected (or all if multi-selected) |
| `Space` | Toggle selection (multi-select mode) |
| `a` | Select all visible |
| `A` | Deselect all |
| `x` | Execute command on selected (prompts for command) |
| `/` | Focus filter input |
| `Esc` | Clear filter / clear selection / back |
| `Tab` | Toggle group collapse |
| `?` | Show help |
| `q` | Quit |
| `e` | Edit config |
| `r` | Reload config |

**Multi-select workflow:**
```
1. Navigate to first server, press Space (selected)
2. Navigate to second server, press Space (selected)
3. Press Enter → opens both in separate terminal tabs
   OR press x → prompts for command, opens both with that command running
```

**TUI with selection:**
```
┌─────────────────────────────────────────────────────────────────┐
│  hop - SSH Connection Manager                          v0.1.0  │
├─────────────────────────────────────────────────────────────────┤
│  🔍 Filter: _                                    [2 selected]   │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ▼ myapp (3)                                                    │
│    ├─ ▼ prod (2)                                                │
│    │    ├─ ● web1      web1.myapp.com          user@host       │
│    │    └─ ○ db        db.myapp.com            admin@host      │
│    └─ ▼ staging (1)                                             │
│         └─ ● web1      staging.myapp.com       user@host       │
│                                                                 │
├─────────────────────────────────────────────────────────────────┤
│  space select │ enter open all │ x exec │ a select all │ ? help │
└─────────────────────────────────────────────────────────────────┘
```
`●` = selected, `○` = not selected

### 4. Multi-Server Command Execution

```bash
# Execute on all servers in a group
hop exec myapp-prod "uptime"
hop exec myapp-prod "df -h"

# Execute on all servers matching a pattern
hop exec "web*" "sudo systemctl restart nginx"

# Execute with custom parallelism
hop exec myapp-prod "deploy.sh" --parallel=2
```

**Output modes:**
```bash
# Default: grouped output
hop exec prod "hostname"
# ═══ myapp-prod-web1 ═══
# web1.myapp.com
# ═══ myapp-prod-db ═══
# db.myapp.com

# Stream mode: live interleaved output with prefixes
hop exec prod "tail -f /var/log/app.log" --stream
# [web1] INFO: Request received
# [db] INFO: Query executed
# [web1] INFO: Response sent
```

### 5. Raycast Extension

Quick access to connections from Raycast launcher.

**Commands:**

| Command | Description |
|---------|-------------|
| `hop connect` | Browse and connect to servers |
| `hop open group` | Open all servers in a group |
| `hop exec` | Run command on multiple servers |
| `hop dashboard` | Open TUI dashboard |

**Connect View:**
```
┌─────────────────────────────────────────────────────────────────┐
│ 🔍 hop connect: prod                                            │
├─────────────────────────────────────────────────────────────────┤
│ ☐ 🖥  myapp-prod-web1                                           │
│      web1.myapp.com • user                                      │
│                                                                 │
│ ☐ 🖥  myapp-prod-db                                             │
│      db.myapp.com • admin                                       │
│                                                                 │
│ ☐ 🖥  client-prod-api                                           │
│      10.0.1.50 • deploy                                         │
├─────────────────────────────────────────────────────────────────┤
│ ↵ Connect  ⌘⇧C Copy  ⌘K Select Multiple  ⌘⇧O Open with Command │
└─────────────────────────────────────────────────────────────────┘
```

**Keyboard shortcuts:**

| Shortcut | Action |
|----------|--------|
| `Enter` | Connect to selected server (opens terminal tab) |
| `Cmd+K` | Toggle multi-select mode |
| `Cmd+A` | Select all visible (in multi-select mode) |
| `Cmd+Shift+C` | Copy SSH command to clipboard |
| `Cmd+Shift+O` | Open with command (prompts for initial command) |
| `Cmd+D` | Open TUI dashboard |
| `Cmd+E` | Quick exec (prompts for command, shows output in Raycast) |

**Multi-select flow:**
1. Press `Cmd+K` to enter multi-select mode
2. Navigate and press `Enter` to toggle selection on each server
3. Press `Cmd+Enter` to open all selected in separate terminal tabs
4. Or press `Cmd+Shift+O` to open all with an initial command

**Exec View (Cmd+E):**
```
┌─────────────────────────────────────────────────────────────────┐
│ 🔍 hop exec: uptime                                             │
├─────────────────────────────────────────────────────────────────┤
│ Select servers to run command on:                               │
│                                                                 │
│ ☑ myapp-prod-web1                                               │
│ ☑ myapp-prod-db                                                 │
│ ☐ client-prod-api                                               │
├─────────────────────────────────────────────────────────────────┤
│ ↵ Run on Selected                                               │
└─────────────────────────────────────────────────────────────────┘
```

After execution, results appear in Raycast:
```
┌─────────────────────────────────────────────────────────────────┐
│ hop exec: uptime                                    ✓ Complete  │
├─────────────────────────────────────────────────────────────────┤
│ ═══ myapp-prod-web1 ═══                                         │
│ 14:32:01 up 45 days, 3:21, 2 users, load average: 0.12         │
│                                                                 │
│ ═══ myapp-prod-db ═══                                           │
│ 14:32:01 up 45 days, 3:21, 1 user, load average: 0.08          │
├─────────────────────────────────────────────────────────────────┤
│ ⌘C Copy All  ↵ Done                                             │
└─────────────────────────────────────────────────────────────────┘
```

---

## Configuration

Config location: `~/.config/hop/config.yaml`

### Full Example

```yaml
# hop configuration
version: 1

# Default settings applied to all connections
defaults:
  user: dan
  port: 22

# Connection definitions
connections:
  # Simple connection
  - id: personal-vps
    host: vps.example.com

  # Grouped connections (project → environment)
  - id: myapp-prod-web1
    project: myapp
    env: prod
    host: web1.myapp.com
    user: deploy
    tags: [web, frontend]

  - id: myapp-prod-web2
    project: myapp
    env: prod
    host: web2.myapp.com
    user: deploy
    tags: [web, frontend]

  - id: myapp-prod-db
    project: myapp
    env: prod
    host: db.myapp.com
    user: admin
    port: 2222
    tags: [database, postgres]

  - id: myapp-staging-web1
    project: myapp
    env: staging
    host: staging.myapp.com
    user: deploy

  # Connection with custom key
  - id: client-prod-api
    project: client-project
    env: prod
    host: 10.0.1.50
    user: deploy
    identity_file: ~/.ssh/client-project.pem

  # Connection with extra SSH options
  - id: legacy-server
    host: old.example.com
    user: root
    options:
      StrictHostKeyChecking: "no"
      ServerAliveInterval: 60

# Groups for multi-exec (optional - auto-generated from project/env)
groups:
  all-web:
    - myapp-prod-web1
    - myapp-prod-web2
    - myapp-staging-web1
```

### Connection Fields

| Field | Required | Default | Description |
|-------|----------|---------|-------------|
| `id` | Yes | - | Unique identifier for the connection |
| `host` | Yes | - | Hostname or IP address |
| `user` | No | from defaults or current user | SSH username |
| `port` | No | 22 | SSH port |
| `project` | No | - | Project name for grouping |
| `env` | No | - | Environment name (prod, staging, dev) |
| `identity_file` | No | - | Path to SSH private key |
| `tags` | No | [] | Tags for filtering and searching |
| `options` | No | {} | Additional SSH options |

### Environment Variables

| Variable | Description |
|----------|-------------|
| `HOP_CONFIG` | Custom config file path |
| `HOP_DEFAULT_USER` | Override default SSH user |
| `HOP_TERMINAL` | Terminal to use (auto-detected) |

---

## CLI Reference

### Commands

```bash
hop                           # Open TUI dashboard
hop <query>                   # Quick connect (fuzzy match)
hop connect <id>              # Connect to exact ID
hop open <group|ids...>       # Open multiple terminal tabs (batch connect)
hop open <group> -- "<cmd>"   # Open all with initial command
hop list                      # List all connections
hop list --json               # List as JSON
hop exec <group> "<command>"  # Execute on multiple servers (non-interactive)
hop config                    # Open config in $EDITOR
hop config --validate         # Validate config file
hop config --path             # Print config path
hop version                   # Show version
hop help                      # Show help
```

### Flags

```bash
# Global flags
--config, -c <path>     # Use custom config file
--verbose, -v           # Verbose output
--quiet, -q             # Suppress non-essential output

# Connect flags
hop <query> --dry-run   # Show SSH command without executing
hop <query> -t          # Force TTY allocation
hop <query> -- <cmd>    # Run command on remote host

# Open flags (batch connect)
hop open <g> --tag=<tag>       # Filter by tag
hop open <g> -- "<cmd>"        # Run initial command on each
hop open <g> --dry-run         # Show what would be opened

# Exec flags (non-interactive)
hop exec <g> "<cmd>" --parallel=N    # Max parallel connections (default: 10)
hop exec <g> "<cmd>" --stream        # Stream output in real-time
hop exec <g> "<cmd>" --fail-fast     # Stop on first error
hop exec <g> "<cmd>" --timeout=30s   # Command timeout
```

---

## Installation

### Homebrew (macOS/Linux)

```bash
brew tap <username>/tap
brew install hop
```

### Binary Download

```bash
# macOS (Apple Silicon)
curl -L https://github.com/<username>/hop/releases/latest/download/hop_darwin_arm64 -o hop
chmod +x hop
sudo mv hop /usr/local/bin/

# Linux
curl -L https://github.com/<username>/hop/releases/latest/download/hop_linux_amd64 -o hop
chmod +x hop
sudo mv hop /usr/local/bin/
```

### From Source

```bash
go install github.com/<username>/hop/cmd/hop@latest
```

### Raycast Extension

1. Open Raycast
2. Search for "hop" in the Store
3. Install the extension
4. Extension reads from `~/.config/hop/config.yaml`

---

## Getting Started

### 1. Create config

```bash
mkdir -p ~/.config/hop
cat > ~/.config/hop/config.yaml << 'EOF'
version: 1

defaults:
  user: your-username

connections:
  - id: my-server
    host: server.example.com
EOF
```

### 2. Connect

```bash
hop my-server
# or just
hop my
```

### 3. Open dashboard

```bash
hop
```

---

## Future Ideas (v2+)

- SSH tunnel management (`hop tunnel myapp-prod-db -L 5432:localhost:5432`)
- Connection health checks / ping dashboard
- Import from `~/.ssh/config`
- SFTP quick commands (`hop cp local.txt myserver:/path/`)
- Session recording / audit log
- Team config sharing (git-based)
- tmux integration (auto-name sessions)
- Connection bookmarks / favorites
- Custom pre/post connect hooks

---

## License

MIT License - use it however you want.

---

## Contributing

PRs welcome! Please open an issue first for major changes.
