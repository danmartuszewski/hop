package hopmcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestExecCommandOutputLimit(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake SSH uses a POSIX shell")
	}
	for _, chunks := range []int{64, 65} {
		t.Run(fmt.Sprint(chunks), func(t *testing.T) {
			dir := t.TempDir()
			// A fake local SSH executable exercises capture and serialization without
			// opening a network connection. Write both streams beyond pipe capacity.
			script := fmt.Sprintf("#!/bin/sh\ni=0\nwhile [ \"$i\" -lt %d ]; do\nprintf '%%s' '%s'\nprintf '%%s' '%s' >&2\ni=$((i+1))\ndone\n", chunks, strings.Repeat("x", 1024), strings.Repeat("y", 1024))
			if err := os.WriteFile(filepath.Join(dir, "ssh"), []byte(script), 0700); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
			loader := &configLoader{cfgPath: writeTestConfig(t, fullTestConfig())}
			result, _, err := loader.handleExecCommand(context.Background(), nil, ExecCommandInput{Target: "db-prod", Command: "unused"})
			if err != nil {
				t.Fatal(err)
			}
			if result.IsError {
				t.Fatalf("tool failed: %+v", result)
			}
			var response struct {
				Results []struct {
					Stdout   string `json:"stdout"`
					Stderr   string `json:"stderr"`
					ExitCode int    `json:"exit_code"`
				}
			}
			if err := json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &response); err != nil {
				t.Fatal(err)
			}
			if len(response.Results) != 1 {
				t.Fatalf("expected one host, got %d", len(response.Results))
			}
			got := response.Results[0]
			marker := ""
			if chunks > 64 {
				marker = "\n[truncated]"
			}
			if got.Stdout != strings.Repeat("x", MaxBytesPerHost)+marker || got.Stderr != strings.Repeat("y", MaxBytesPerHost)+marker {
				t.Fatal("output cap or truncation marker incorrect")
			}
			if got.ExitCode != 0 {
				t.Fatalf("output was not drained: exit code %d", got.ExitCode)
			}
		})
	}
}
