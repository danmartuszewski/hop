---
name: publish-mcp
description: Publish updated MCP server metadata to the MCP Registry
disable-model-invocation: true
argument-hint: []
---

Publish the hop MCP server to the MCP Registry with the latest release version.

## Steps

1. Get the latest release version from GitHub:
   ```
   gh release view --repo danmartuszewski/hop --json tagName -q .tagName
   ```
   Strip the leading `v` prefix (e.g. `v1.2.0` → `1.2.0`).

2. Read `server.json` and check if the version already matches. If it does, inform the user and stop.

3. Update the `version` field in `server.json` to the latest release version.

4. Commit the change:
   ```
   git add server.json && git commit -m "chore: bump server.json version to <version>"
   ```

5. Push to remote:
   ```
   git push
   ```

6. Publish to the MCP Registry:
   ```
   mcp-publisher publish
   ```

7. Report the result to the user.
