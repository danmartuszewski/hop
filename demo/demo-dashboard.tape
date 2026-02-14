# hop - Dashboard Demo (Short)
# Run with: vhs demo-dashboard.tape
#
# No Docker required — this demo only shows the TUI, no SSH connections.

# --- Output ---
Output hop-dashboard.gif

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

Type `cat > /tmp/hop-demo/.config/hop/config.yaml << 'CONF'`
Enter
Type `version: 1`
Enter
Type `connections:`
Enter
Type `  - id: prod-web-01`
Enter
Type `    host: web-01.us-east-1.example.com`
Enter
Type `    user: deploy`
Enter
Type `    project: myapp`
Enter
Type `    env: production`
Enter
Type `    tags: [web, nginx, prod]`
Enter
Type `  - id: prod-web-02`
Enter
Type `    host: web-02.us-east-1.example.com`
Enter
Type `    user: deploy`
Enter
Type `    project: myapp`
Enter
Type `    env: production`
Enter
Type `    tags: [web, nginx, prod]`
Enter
Type `  - id: prod-db-master`
Enter
Type `    host: db-master.us-east-1.example.com`
Enter
Type `    user: dbadmin`
Enter
Type `    port: 5432`
Enter
Type `    project: myapp`
Enter
Type `    env: production`
Enter
Type `    tags: [database, postgres, prod]`
Enter
Type `  - id: prod-db-replica`
Enter
Type `    host: db-replica.us-east-1.example.com`
Enter
Type `    user: dbadmin`
Enter
Type `    port: 5432`
Enter
Type `    project: myapp`
Enter
Type `    env: production`
Enter
Type `    tags: [database, postgres, prod]`
Enter
Type `  - id: staging-web`
Enter
Type `    host: staging.example.com`
Enter
Type `    user: deploy`
Enter
Type `    project: myapp`
Enter
Type `    env: staging`
Enter
Type `    tags: [web, staging]`
Enter
Type `  - id: staging-api`
Enter
Type `    host: api-staging.example.com`
Enter
Type `    user: deploy`
Enter
Type `    project: myapp`
Enter
Type `    env: staging`
Enter
Type `    tags: [api, staging]`
Enter
Type `  - id: monitoring`
Enter
Type `    host: grafana.internal.example.com`
Enter
Type `    user: ops`
Enter
Type `    project: infra`
Enter
Type `    env: production`
Enter
Type `    tags: [monitoring, grafana]`
Enter
Type `  - id: bastion`
Enter
Type `    host: bastion.example.com`
Enter
Type `    user: admin`
Enter
Type `    project: infra`
Enter
Type `    env: production`
Enter
Type `    tags: [bastion, gateway]`
Enter
Type `  - id: ci-runner`
Enter
Type `    host: ci.internal.example.com`
Enter
Type `    user: runner`
Enter
Type `    project: infra`
Enter
Type `    env: production`
Enter
Type `    tags: [ci, jenkins]`
Enter
Type `  - id: dev-sandbox`
Enter
Type `    host: 192.168.1.100`
Enter
Type `    user: dan`
Enter
Type `    tags: [dev, local]`
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
# SCENE 1: Launch dashboard
# ============================================================

Sleep 500ms

Type@80ms `hop`
Enter
Sleep 2s

# Navigate down through connections
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms

Sleep 1s

# Navigate back up
Up
Sleep 200ms
Up
Sleep 200ms
Up
Sleep 200ms
Up
Sleep 200ms
Up
Sleep 200ms
Up
Sleep 200ms
Up
Sleep 200ms
Up
Sleep 200ms

Sleep 500ms

# ============================================================
# SCENE 2: Copy SSH command
# ============================================================

# Press 'y' to copy SSH command
Type `y`
Sleep 2s

# ============================================================
# SCENE 3: Show help
# ============================================================

Type `?`
Sleep 3s

# Close help — back to dashboard
Escape
Sleep 3s
