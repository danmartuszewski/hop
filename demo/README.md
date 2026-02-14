# Demo Recordings

VHS tape files for generating hop demo GIFs. See [DEMOS.md](DEMOS.md) for the showcase.

## Prerequisites

- [VHS](https://github.com/charmbracelet/vhs): `brew install charmbracelet/tap/vhs`
- hop built locally: `make build`
- Docker (for demos with real SSH connections)

## Tapes

| Tape | Description | Needs Docker |
|------|-------------|:---:|
| `demo-short.tape` | Before/after with real SSH connection | yes |
| `demo-dashboard.tape` | TUI dashboard navigation + help | no |
| `demo-import.tape` | Import from ~/.ssh/config | no |
| `demo-exec.tape` | Multi-exec across servers | yes |
| `demo-manage.tape` | Add, paste, edit, delete connections | no |
| `demo-search.tape` | Text filtering, tag filtering, combined | no |

## Running

```bash
cd demo

# Generate a single demo
vhs demo-short.tape

# Generate all demos
./generate.sh
```

## Docker SSH Server Setup

Demos marked "yes" above require a local SSH server:

```bash
# Start the server
docker run -d --name hop-demo-ssh -p 2222:2222 \
  -e USER_NAME=deploy -e PASSWORD_ACCESS=true \
  -e USER_PASSWORD=demo \
  lscr.io/linuxserver/openssh-server:latest

# Set up key auth
mkdir -p /tmp/hop-demo
ssh-keygen -t ed25519 -f /tmp/hop-demo/demo_key -N ""
docker exec hop-demo-ssh mkdir -p /config/.ssh
docker cp /tmp/hop-demo/demo_key.pub hop-demo-ssh:/config/.ssh/authorized_keys
docker exec hop-demo-ssh chown -R 911:911 /config/.ssh
docker exec hop-demo-ssh chmod 700 /config/.ssh
docker exec hop-demo-ssh chmod 600 /config/.ssh/authorized_keys

# Verify
ssh -i /tmp/hop-demo/demo_key -o StrictHostKeyChecking=no -p 2222 deploy@127.0.0.1 hostname

# Cleanup when done
docker rm -f hop-demo-ssh
```
