# hop - Search & Filtering Demo (Short)
# Run with: vhs demo-search.tape
#
# No Docker required — TUI only, no SSH connections.

# --- Output ---
Output hop-search.gif

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

# Create a config with many connections to showcase filtering
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
Type `  - id: prod-api-01`
Enter
Type `    host: api-01.us-east-1.example.com`
Enter
Type `    user: deploy`
Enter
Type `    project: myapp`
Enter
Type `    env: production`
Enter
Type `    tags: [api, prod]`
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
Type `  - id: staging-db`
Enter
Type `    host: db-staging.example.com`
Enter
Type `    user: dbadmin`
Enter
Type `    project: myapp`
Enter
Type `    env: staging`
Enter
Type `    tags: [database, staging]`
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
# SCENE 1: Open dashboard — full list
# ============================================================

Sleep 500ms

Type@80ms `hop`
Enter
Sleep 2s

# ============================================================
# SCENE 2: Text filter — single keyword
# ============================================================

# Press '/' to start filtering
Type `/`
Sleep 500ms

# Type "db" — should narrow to database connections
Type@80ms `db`
Sleep 2s

# Clear filter
Escape
Sleep 1s

# ============================================================
# SCENE 3: Text filter — multi-keyword AND search
# ============================================================

Type `/`
Sleep 500ms

# Type "prod web" — matches connections with BOTH keywords
Type@80ms `prod web`
Sleep 2s

# Clear filter
Escape
Sleep 1s

# ============================================================
# SCENE 4: Tag filtering
# ============================================================

# Press 't' to open tag picker
Type `t`
Sleep 1s

# NOTE: Tag picker shows tags in alphabetical order from the config.
# Down count targets a specific tag — adjust if config tags change.
Down
Sleep 200ms
Down
Sleep 200ms
Down
Sleep 200ms

# Select "database" tag with space
Space
Sleep 1.5s

# Close tag picker — list is now filtered
Escape
Sleep 2s

# Reopen tags to clear and pick another
Type `t`
Sleep 500ms

# Clear all tags
Type `c`
Sleep 500ms

# Pick "nginx" tag (index 10 in alphabetical tag list)
Down
Sleep 200ms
Down
Sleep 200ms
Down
Sleep 200ms
Down
Sleep 200ms
Down
Sleep 200ms
Down
Sleep 200ms
Down
Sleep 200ms
Down
Sleep 200ms
Down
Sleep 200ms
Down
Sleep 200ms

Space
Sleep 1s

Escape
Sleep 2s

# ============================================================
# SCENE 5: Combined text + tag filter
# ============================================================

# Tags are still active, now add a text filter on top
Type `/`
Sleep 500ms
Type@80ms `prod`
Sleep 2s

# Clear text filter — back to tag-filtered view
Escape
Sleep 3s
