# hop - SSH Config Import Demo (Short)
# Run with: vhs demo-import.tape

# --- Output ---
Output hop-import.gif

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

# Create empty hop config
Type `mkdir -p /tmp/hop-demo/.config/hop`
Enter
Sleep 200ms

Type `cat > /tmp/hop-demo/.config/hop/config.yaml << 'CONF'`
Enter
Type `version: 1`
Enter
Type `connections: []`
Enter
Type `CONF`
Enter
Sleep 300ms

# Create a fake ~/.ssh/config with realistic entries
Type `mkdir -p /tmp/hop-demo/.ssh`
Enter
Sleep 200ms

Type `cat > /tmp/hop-demo/.ssh/config << 'SSH'`
Enter
Type `Host prod-web-01`
Enter
Type `    HostName web-01.us-east-1.example.com`
Enter
Type `    User deploy`
Enter
Type `    IdentityFile ~/.ssh/prod_key`
Enter
Type ``
Enter
Type `Host prod-web-02`
Enter
Type `    HostName web-02.us-east-1.example.com`
Enter
Type `    User deploy`
Enter
Type `    IdentityFile ~/.ssh/prod_key`
Enter
Type ``
Enter
Type `Host prod-db`
Enter
Type `    HostName db.us-east-1.example.com`
Enter
Type `    User dbadmin`
Enter
Type `    Port 5432`
Enter
Type ``
Enter
Type `Host staging`
Enter
Type `    HostName staging.example.com`
Enter
Type `    User deploy`
Enter
Type ``
Enter
Type `Host bastion`
Enter
Type `    HostName bastion.example.com`
Enter
Type `    User admin`
Enter
Type `    IdentityFile ~/.ssh/bastion_key`
Enter
Type ``
Enter
Type `Host internal-app`
Enter
Type `    HostName 10.0.1.50`
Enter
Type `    User app`
Enter
Type `    ProxyJump bastion`
Enter
Type `    ForwardAgent yes`
Enter
Type ``
Enter
Type `Host ci-runner`
Enter
Type `    HostName ci.internal.example.com`
Enter
Type `    User runner`
Enter
Type `    ProxyJump bastion`
Enter
Type ``
Enter
Type `Host *`
Enter
Type `    ServerAliveInterval 60`
Enter
Type `SSH`
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
# SCENE 1: Show the SSH config we're importing from
# ============================================================

Sleep 500ms

Type@80ms `cat ~/.ssh/config`
Enter
Sleep 3s

# ============================================================
# SCENE 2: Open dashboard — empty
# ============================================================

Type@80ms `hop`
Enter
Sleep 2s

# ============================================================
# SCENE 3: Press 'i' to import
# ============================================================

Type `i`
Sleep 2s

# All connections are pre-selected, navigate to see them
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

# Deselect one (the wildcard should already be skipped)
Space
Sleep 500ms

# Go back up
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

# Import selected
Enter
Sleep 2s

# Dashboard now shows all imported connections
Sleep 3s
