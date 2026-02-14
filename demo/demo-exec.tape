# hop - Multi-exec Demo (Short)
# Run with: vhs demo-exec.tape
#
# Prerequisites: same as demo-short.tape (Docker SSH server + key auth)

# --- Output ---
Output hop-exec.gif

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

Type `mkdir -p /tmp/hop-demo/.config/hop`
Enter
Sleep 200ms

# Multiple connections pointing to the same Docker SSH server
# to simulate a real multi-server exec
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
Type `  - id: prod-web-02`
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
Type `  - id: prod-api-01`
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
Type `    tags: [api, prod]`
Enter
Type `    options:`
Enter
Type `      StrictHostKeyChecking: "no"`
Enter
Type `      UserKnownHostsFile: /dev/null`
Enter
Type `      LogLevel: ERROR`
Enter
Type `groups:`
Enter
Type `  production: [prod-web-01, prod-web-02, prod-api-01]`
Enter
Type `CONF`
Enter
Sleep 500ms

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
# SCENE 1: Execute on a named group
# ============================================================

Sleep 500ms

Type@80ms `hop exec production "uptime"`
Enter
Sleep 4s

# ============================================================
# SCENE 2: Stream mode — real-time output with host prefixes
# ============================================================

Type@80ms `hop exec production "df -h /" --stream`
Enter
Sleep 4s

# ============================================================
# SCENE 3: Glob pattern targeting
# ============================================================

Type@80ms `hop exec "prod-web*" "echo 'nginx ok'"`
Enter
Sleep 4s

# ============================================================
# SCENE 4: Resolve to preview targets
# ============================================================

Type@80ms `hop resolve production`
Enter
Sleep 3s
