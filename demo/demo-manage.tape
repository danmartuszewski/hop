# hop - Server Management Demo (Short)
# Run with: vhs demo-manage.tape
#
# No Docker required — TUI only, no SSH connections.

# --- Output ---
Output hop-manage.gif

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
Type `    tags: [web, prod]`
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
Type `    tags: [database, prod]`
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
# SCENE 1: Open dashboard
# ============================================================

Sleep 500ms

Type@80ms `hop`
Enter
Sleep 2s

# ============================================================
# SCENE 2: Add a new connection
# ============================================================

# Press 'a' to add
Type `a`
Sleep 1s

# Fill in the form
Type@60ms `prod-api-01`
Tab
Sleep 300ms
Type@60ms `api.us-east-1.example.com`
Tab
Sleep 300ms
Type@60ms `deploy`
Tab
Sleep 300ms
# Port — skip (default 22)
Tab
Sleep 300ms
# Project
Type@60ms `myapp`
Tab
Sleep 300ms
# Env
Type@60ms `production`
Tab
Sleep 300ms
# Tags
Type@60ms `api, prod`

Sleep 1s

# Submit the form
Enter
Sleep 1.5s

# ============================================================
# SCENE 3: Quick add with paste
# ============================================================

# Press 'p' to paste an SSH string
Type `p`
Sleep 1s

Type@60ms `ssh admin@monitoring.internal.example.com -p 2222`
Sleep 500ms
Enter
Sleep 1s

# Form pre-filled — just add project and tags
# NOTE: Tab count assumes form order: ID → Host → User → Port → Project → Env → Tags
# If form fields change, adjust the number of Tabs below.
Tab
Sleep 200ms
Tab
Sleep 200ms
Tab
Sleep 200ms
Tab
Sleep 300ms
Type@60ms `infra`
Tab
Sleep 300ms
Type@60ms `production`
Tab
Sleep 300ms
Type@60ms `monitoring`

Sleep 500ms
Enter
Sleep 1.5s

# ============================================================
# SCENE 4: Edit a connection
# ============================================================

# Navigate to a connection and edit
Down
Sleep 300ms
Down
Sleep 300ms

# Press 'e' to edit
Type `e`
Sleep 1s

# Navigate to tags field (6 tabs from ID)
# NOTE: Tab count assumes form order: ID → Host → User → Port → Project → Env → Tags
Tab
Sleep 200ms
Tab
Sleep 200ms
Tab
Sleep 200ms
Tab
Sleep 200ms
Tab
Sleep 200ms
Tab
Sleep 300ms

# We're on tags — append a tag
Type@60ms `, critical`

Sleep 500ms
Enter
Sleep 1.5s

# ============================================================
# SCENE 5: Delete a connection
# ============================================================

# Navigate down to a connection
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms

# Press 'd' to delete
Type `d`
Sleep 1s

# Confirm with 'y'
Type `y`
Sleep 3s
