# hop - Short Demo (YouTube Shorts / social media)
# Run with: vhs demo-short.tape
#
# Prerequisites:
#   1. Install VHS: brew install charmbracelet/tap/vhs
#   2. Build hop: make build
#   3. Start demo SSH server:
#      docker run -d --name hop-demo-ssh -p 2222:2222 \
#        -e USER_NAME=deploy -e PASSWORD_ACCESS=true \
#        -e USER_PASSWORD=demo \
#        lscr.io/linuxserver/openssh-server:latest
#   4. Set up key auth:
#      mkdir -p /tmp/hop-demo
#      ssh-keygen -t ed25519 -f /tmp/hop-demo/demo_key -N ""
#      docker exec hop-demo-ssh mkdir -p /config/.ssh
#      docker cp /tmp/hop-demo/demo_key.pub hop-demo-ssh:/config/.ssh/authorized_keys
#      docker exec hop-demo-ssh chown -R 911:911 /config/.ssh
#      docker exec hop-demo-ssh chmod 700 /config/.ssh
#      docker exec hop-demo-ssh chmod 600 /config/.ssh/authorized_keys
#   5. Cleanup after: docker rm -f hop-demo-ssh

# --- Output ---
Output hop-short.gif

# --- Terminal settings ---
Set Shell "bash"
Set FontSize 18
Set Width 1200
Set Height 700
Set Padding 20
Set FontFamily "JetBrains Mono"
Set Theme "coolnight"
Set TypingSpeed 40ms
Set CursorBlink false

# --- Hidden setup ---
Hide

# Create demo config in a fake HOME so hop picks it up automatically
# (the -c flag doesn't work with fuzzy connect, so we override HOME)
Type `mkdir -p /tmp/hop-demo/.config/hop`
Enter
Sleep 200ms

Type `cat > /tmp/hop-demo/.config/hop/config.yaml << 'CONF'`
Enter
Type `version: 1`
Enter
Type `connections:`
Enter
Type `  - id: prod-web-01`
Enter
Type `    host: 127.0.0.1`
Enter
Type `    user: deploy`
Enter
Type `    port: 2222`
Enter
Type `    identity_file: /tmp/hop-demo/demo_key`
Enter
Type `    project: myapp`
Enter
Type `    env: production`
Enter
Type `    tags: [web, prod]`
Enter
Type `    options:`
Enter
Type `      StrictHostKeyChecking: "no"`
Enter
Type `      UserKnownHostsFile: /dev/null`
Enter
Type `      LogLevel: ERROR`
Enter
Type `CONF`
Enter
Sleep 500ms

# Set HOME so hop uses our demo config, and alias hop to local build
Type `export HOME=/tmp/hop-demo`
Enter
Sleep 200ms
Type `export PATH=$PWD/../bin:$PATH`
Enter
Sleep 200ms
Type `clear`
Enter
Sleep 500ms

Show

# ============================================================
# SCENE 1: The Pain (typing long SSH command)
# ============================================================

Sleep 1s

Type@100ms `ssh -i ~/.ssh/work_key deploy@app-server-prod-03.us-east-1.example.com -p 2222`

Sleep 1.5s

# Backspace to simulate noticing a typo in the port
Backspace 4
Sleep 300ms
Type@150ms `22`

Sleep 500ms

# Give up
Ctrl+C
Sleep 500ms

Type `# there has to be a better way...`
Enter
Sleep 1.5s

# ============================================================
# SCENE 2: hop — real connection
# ============================================================

Type@80ms `hop prod`
Sleep 500ms
Enter

# Wait for SSH connection to establish and shell prompt to appear
Sleep 3s

# Run commands on the remote server to prove we're in
Type `hostname`
Enter
Sleep 1s

Type `uptime`
Enter
Sleep 1s

Type `whoami`
Enter
Sleep 1.5s

# Exit SSH
Type `exit`
Enter
Sleep 1s

# ============================================================
# SCENE 3: Tagline
# ============================================================

Type `# hop — just type hop and you're in.`
Enter
Sleep 500ms

Type `# github.com/danmartuszewski/hop`
Enter

Sleep 3s
