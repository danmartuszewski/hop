#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "Stopping test server..."
docker compose -f "$SCRIPT_DIR/docker-compose.yaml" down

# Remove known_hosts entry for localhost:2222
ssh-keygen -R "[localhost]:2222" 2>/dev/null || true

echo "Done."
