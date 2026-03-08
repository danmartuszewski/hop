#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
KEY_FILE="$SCRIPT_DIR/id_test"
CONFIG_FILE="$SCRIPT_DIR/hop-test-config.yaml"

# Generate SSH key pair if not exists
if [ ! -f "$KEY_FILE" ]; then
    echo "Generating test SSH key pair..."
    ssh-keygen -t ed25519 -f "$KEY_FILE" -N "" -C "hop-test"
fi

# Build and start the container
echo "Starting test server..."
docker compose -f "$SCRIPT_DIR/docker-compose.yaml" up -d --build

# Wait for SSH to be ready
echo "Waiting for SSH to be ready..."
for i in $(seq 1 10); do
    if ssh -o StrictHostKeyChecking=no -o ConnectTimeout=2 -i "$KEY_FILE" -p 2222 testuser@localhost echo "ok" 2>/dev/null; then
        break
    fi
    if [ "$i" -eq 10 ]; then
        echo "ERROR: SSH server did not become ready in time"
        docker compose -f "$SCRIPT_DIR/docker-compose.yaml" logs
        exit 1
    fi
    sleep 1
done

echo ""
echo "=== Test server ready ==="
echo ""
echo "Test config: $CONFIG_FILE"
echo ""
echo "Usage (set HOP_CONFIG or use -c flag):"
echo "  export HOP_CONFIG=$CONFIG_FILE"
echo ""
echo "  # Test SSH connection"
echo "  hop test-ssh -- 'echo hello'"
echo ""
echo "  # Test mosh connection (requires mosh installed locally: brew install mosh)"
echo "  hop test-mosh"
echo ""
echo "  # Test --mosh flag override on SSH connection"
echo "  hop test-ssh --mosh"
echo ""
echo "  # Dry run to see generated commands"
echo "  hop connect test-ssh --dry-run"
echo "  hop connect test-mosh --dry-run"
echo "  hop connect test-ssh --mosh --dry-run"
echo ""
echo "Teardown:"
echo "  $SCRIPT_DIR/teardown.sh"
