#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

# Rebuild hop binary first
echo "==> Building hop..."
(cd .. && make build)

# Kill stale ttyd processes from previous runs
pkill -f ttyd 2>/dev/null || true
sleep 1

for f in *.tape; do
  echo "==> $f"
  vhs "$f"
  pkill -f ttyd 2>/dev/null || true
  sleep 1
done

# Update root hop.gif (README hero image)
cp hop-dashboard.gif ../hop.gif

echo "Done. Generated GIFs:"
ls -lh *.gif 2>/dev/null
