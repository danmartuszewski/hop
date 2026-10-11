package ssh

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/danmartuszewski/hop/internal/config"
)

func TestCaptureBuffer(t *testing.T) {
	for _, tt := range []struct {
		name      string
		limit     int
		chunks    []string
		want      string
		truncated bool
	}{
		{"empty", 4, nil, "", false},
		{"exact limit", 4, []string{"ab", "cd", ""}, "abcd", false},
		{"overflow", 4, []string{"ab", "cdef", "gh"}, "abcd", true},
		{"unlimited", 0, []string{"abcd", "efgh"}, "abcdefgh", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			b := captureBuffer{limit: tt.limit}
			for _, chunk := range tt.chunks {
				n, err := b.Write([]byte(chunk))
				if n != len(chunk) || err != nil {
					t.Fatalf("Write must drain all input: n=%d, err=%v", n, err)
				}
			}
			if b.String() != tt.want || b.truncated != tt.truncated {
				t.Fatalf("capture = %q, truncated=%v; want %q, %v", b.String(), b.truncated, tt.want, tt.truncated)
			}
		})
	}
}

func fakeSSH(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fake SSH uses a POSIX shell")
	}
	dir := t.TempDir()
	script := "#!/bin/sh\nprintf 'first\\nsecond\\n'\nprintf 'error\\nnext\\n' >&2\nexit 7\n"
	if err := os.WriteFile(filepath.Join(dir, "ssh"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestExecuteCapturesBoundedOutput(t *testing.T) {
	fakeSSH(t)
	conn := &config.Connection{ID: "test", Host: "unused.example.com"}
	result := executeOnHost(context.Background(), conn, &ExecOptions{Command: "unused", MaxOutputBytes: 4})
	if result.Stdout != "firs" || result.Stderr != "erro" || !result.StdoutTruncated || !result.StderrTruncated {
		t.Fatalf("unexpected capture: %+v", result)
	}
	if result.ExitCode != 7 || result.Error == nil {
		t.Fatalf("exit status lost after truncation: %+v", result)
	}
}

func TestExecuteStreamPrefixesFirstLine(t *testing.T) {
	fakeSSH(t)
	stdout, err := os.CreateTemp(t.TempDir(), "stdout")
	if err != nil {
		t.Fatal(err)
	}
	defer stdout.Close()
	stderr, err := os.CreateTemp(t.TempDir(), "stderr")
	if err != nil {
		t.Fatal(err)
	}
	defer stderr.Close()
	originalOut, originalErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = stdout, stderr
	t.Cleanup(func() { os.Stdout, os.Stderr = originalOut, originalErr })
	result := executeOnHost(context.Background(), &config.Connection{ID: "test", Host: "unused.example.com"}, &ExecOptions{Command: "unused", Stream: true})
	if result.ExitCode != 7 {
		t.Fatalf("unexpected result: %+v", result)
	}
	for _, file := range []*os.File{stdout, stderr} {
		data, err := os.ReadFile(file.Name())
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range strings.Split(strings.TrimSuffix(string(data), "\n"), "\n") {
			if !strings.HasPrefix(line, "[test] ") {
				t.Errorf("missing host prefix: %q", line)
			}
		}
	}
}
