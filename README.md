# hop

Fast, elegant SSH connection manager with a TUI dashboard.

![hop TUI dashboard](hop.gif)

```
hop prod          # fuzzy match & connect
hop               # open TUI dashboard
hop list          # list all connections
```

## Features

- **Fuzzy matching** - Connect with minimal keystrokes
- **TUI dashboard** - Browse, add, edit, delete connections interactively
- **Quick paste** - Parse `user@host:port` strings instantly
- **Groups** - Organize connections by project and environment
- **Multi-exec** - Run commands across multiple servers
- **Mouse support** - Scroll and click in the dashboard
- **Zero dependencies** - Single binary, no runtime requirements

## Installation

### From source

```bash
go install github.com/danmartuszewski/hop/cmd/hop@latest
```

### Build locally

```bash
git clone https://github.com/danmartuszewski/hop.git
cd hop
make build
./bin/hop
```

## Quick Start

1. **Launch the dashboard:**
   ```bash
   hop
   ```

2. **Add a connection** - Press `a` and fill in the form, or press `p` to paste an SSH string like `user@host.com`

3. **Connect** - Select a connection and press `Enter`

## Configuration

Config file location: `~/.config/hop/config.yaml`

```yaml
version: 1

defaults:
  user: admin
  port: 22

connections:
  - id: prod-web
    host: web.example.com
    user: deploy
    project: myapp
    env: production
    tags: [web, prod]

  - id: prod-db
    host: db.example.com
    user: dbadmin
    port: 5432
    project: myapp
    env: production
    tags: [database, prod]

  - id: staging
    host: staging.example.com
    user: deploy
    project: myapp
    env: staging

groups:
  production: [prod-web, prod-db]
  web-servers: [prod-web, staging]
```

## TUI Dashboard

Launch with `hop` or `hop dashboard`.

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `↑/k` | Move up |
| `↓/j` | Move down |
| `g` | Go to top |
| `G` | Go to bottom |
| `/` | Filter connections |
| `t` | Filter by tags |
| `r` | Toggle sort by recent |
| `Enter` | Connect to selected |
| `a` | Add new connection |
| `p` | Paste SSH string (quick add) |
| `e` | Edit selected |
| `c` | Duplicate selected |
| `d` | Delete selected |
| `y` | Copy SSH command |
| `?` | Show help |
| `q` | Quit |

### Quick Add with Paste

Press `p` and paste any of these formats:

```
user@host.com
user@host.com:2222
ssh user@host.com -p 2222
ssh://user@host:port
```

The connection form opens with fields pre-filled.

## CLI Commands

```bash
hop                          # Open TUI dashboard
hop <query>                  # Fuzzy match and connect
hop connect <id>             # Connect by exact ID
hop list                     # List all connections
hop list --json              # List as JSON
hop list --flat              # Flat list without grouping
hop open <target...>         # Open multiple terminal tabs
hop exec <target> "cmd"      # Execute command on multiple servers
hop resolve <target>         # Test which connections a target matches
hop version                  # Show version
```

### Targeting

Commands like `exec` and `open` accept a **target** that resolves to one or more connections. The target is matched in this order:

1. **Named group** — an explicit list of connection IDs defined under `groups:` in config
2. **Project-env pattern** — matches connections by `project` and `env` fields (e.g. `myapp-prod` matches all connections with `project: myapp` and `env: prod`)
3. **Glob pattern** — wildcard matching on connection IDs (e.g. `web*`, `*-prod-*`)
4. **Fuzzy match** — falls back to fuzzy matching a single connection ID

You can also filter any target by tag with `--tag`.

Use `hop resolve` to preview which connections a target will match before running anything:

```bash
hop resolve production              # see what "production" resolves to
hop resolve "web*"                  # test a glob pattern
hop resolve myapp-prod --tag=web    # combine target + tag filter
```

### Examples

```bash
# Fuzzy connect
hop prod                # matches "prod-web", "prod-db", etc.
hop web                 # matches first *web* server

# Multi-exec with different target types
hop exec production "uptime"           # named group
hop exec myapp-prod "df -h"            # project-env pattern
hop exec "web*" "systemctl status"     # glob pattern
hop exec --tag=database "psql -c '\\l'" # tag filter

# Open multiple tabs
hop open production                    # named group
hop open web1 db1 api1                 # specific IDs
hop open myapp-prod -- "htop"          # with initial command

# List connections
hop list --flat
```

## Shell Completions

```bash
# Bash (Linux)
hop completion bash | sudo tee /etc/bash_completion.d/hop > /dev/null

# Bash (macOS with Homebrew)
hop completion bash > $(brew --prefix)/etc/bash_completion.d/hop

# Zsh (add to ~/.zshrc)
source <(hop completion zsh)

# Fish
hop completion fish > ~/.config/fish/completions/hop.fish
```

## Flags

```bash
-c, --config <path>    # Use custom config file
-v, --verbose          # Verbose output
-q, --quiet            # Suppress non-essential output
    --dry-run          # Print SSH command without executing
```

## Building

```bash
make build          # Build binary to ./bin/hop
make test           # Run tests
make test-docker    # Run tests in Docker (isolated)
make install        # Install to $GOPATH/bin
make docker         # Build Docker image
```

## Docker

```bash
# Build image
docker build -t hop .

# Run interactively
docker run -it --rm hop

# Run tests in container
docker build --target tester -t hop-test .
```

## Project Structure

```
hop/
├── cmd/hop/           # Main entry point
├── internal/
│   ├── cmd/           # CLI commands (cobra)
│   ├── config/        # Configuration loading/saving
│   ├── fuzzy/         # Fuzzy matching
│   ├── picker/        # Connection picker (promptui)
│   ├── ssh/           # SSH connection handling
│   └── tui/           # TUI dashboard (bubbletea)
├── Dockerfile
├── Makefile
└── README.md
```

## License

MIT License - see [LICENSE](LICENSE) for details.
