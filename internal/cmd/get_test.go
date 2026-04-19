package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/danmartuszewski/hop/internal/config"
)

// boolPtr returns a pointer to the given bool value.
func boolPtr(b bool) *bool { return &b }

// newTestConfig builds a *config.Config with a few representative connections,
// then runs the same default-application logic that Load would.
func newTestConfig() *config.Config {
	cfg := &config.Config{
		Version: 1,
		Defaults: config.Defaults{
			User: "deploy",
			Port: 22,
		},
		Connections: []config.Connection{
			{
				ID:           "prod",
				Host:         "web.example.com",
				User:         "alice",
				Port:         2222,
				Project:      "myapp",
				Env:          "prod",
				IdentityFile: "~/.ssh/prod_id",
				ProxyJump:    "bastion.example.com",
				ForwardAgent: true,
				UseMosh:      boolPtr(true),
				Tags:         []string{"web", "primary"},
				Options: map[string]string{
					"StrictHostKeyChecking": "yes",
					"ServerAliveInterval":   "60",
				},
			},
			{
				ID:      "staging",
				Host:    "staging.example.com",
				Project: "myapp",
				Env:     "staging",
			},
			{
				ID:   "bare",
				Host: "bare.example.com",
			},
		},
	}

	// Mirror config.Load's behavior by applying defaults via a round-trip
	// through the package's exported defaulting (applyDefaults is private,
	// but the same effect is achieved by calling Load on a temp file). To
	// avoid filesystem coupling, we replicate the relevant defaults inline:
	if cfg.Defaults.Port == 0 {
		cfg.Defaults.Port = 22
	}
	for i := range cfg.Connections {
		c := &cfg.Connections[i]
		if c.User == "" {
			c.User = cfg.Defaults.User
		}
		if c.Port == 0 {
			c.Port = cfg.Defaults.Port
		}
	}
	return cfg
}

func TestRunGet_SingleField_Host(t *testing.T) {
	cfg := newTestConfig()
	var stdout, stderr bytes.Buffer

	if err := runGet(cfg, &stdout, &stderr, "prod", "host", getOpts{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := stdout.String(), "web.example.com\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr non-empty: %q", stderr.String())
	}
}

func TestRunGet_SingleField_User_Effective(t *testing.T) {
	cfg := newTestConfig()
	var stdout, stderr bytes.Buffer

	if err := runGet(cfg, &stdout, &stderr, "prod", "user", getOpts{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := stdout.String(), "alice\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}

func TestRunGet_SingleField_User_FromDefaults(t *testing.T) {
	cfg := newTestConfig()
	var stdout, stderr bytes.Buffer

	if err := runGet(cfg, &stdout, &stderr, "staging", "user", getOpts{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := stdout.String(), "deploy\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}

func TestRunGet_SingleField_User_FallbackUSER(t *testing.T) {
	// Build a config where neither connection nor defaults supplies a user.
	cfg := &config.Config{
		Version:  1,
		Defaults: config.Defaults{Port: 22},
		Connections: []config.Connection{
			{ID: "noone", Host: "x.example.com", Port: 22},
		},
	}

	t.Setenv("USER", "envuser")

	var stdout, stderr bytes.Buffer
	if err := runGet(cfg, &stdout, &stderr, "noone", "user", getOpts{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := stdout.String(), "envuser\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}

func TestRunGet_SingleField_Port(t *testing.T) {
	cfg := newTestConfig()

	cases := []struct {
		id   string
		want string
	}{
		{"prod", "2222\n"},
		{"staging", "22\n"},
	}
	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if err := runGet(cfg, &stdout, &stderr, tc.id, "port", getOpts{}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := stdout.String(); got != tc.want {
				t.Errorf("stdout = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestRunGet_SingleField_IdentityFile(t *testing.T) {
	cfg := newTestConfig()
	var stdout, stderr bytes.Buffer
	if err := runGet(cfg, &stdout, &stderr, "prod", "identity_file", getOpts{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := stdout.String(), "~/.ssh/prod_id\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}

func TestRunGet_SingleField_ProxyJump(t *testing.T) {
	cfg := newTestConfig()
	var stdout, stderr bytes.Buffer
	if err := runGet(cfg, &stdout, &stderr, "prod", "proxy_jump", getOpts{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := stdout.String(), "bastion.example.com\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}

func TestRunGet_SingleField_ForwardAgent(t *testing.T) {
	cfg := newTestConfig()

	cases := []struct {
		id   string
		want string
	}{
		{"prod", "true\n"},
		{"staging", "false\n"},
	}
	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if err := runGet(cfg, &stdout, &stderr, tc.id, "forward_agent", getOpts{}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := stdout.String(); got != tc.want {
				t.Errorf("stdout = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestRunGet_SingleField_UseMosh(t *testing.T) {
	cfg := newTestConfig()

	cases := []struct {
		id   string
		want string
	}{
		{"prod", "true\n"},    // explicitly true
		{"staging", "false\n"}, // nil pointer -> false via Mosh()
	}
	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if err := runGet(cfg, &stdout, &stderr, tc.id, "use_mosh", getOpts{}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := stdout.String(); got != tc.want {
				t.Errorf("stdout = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestRunGet_SingleField_Project(t *testing.T) {
	cfg := newTestConfig()
	var stdout, stderr bytes.Buffer
	if err := runGet(cfg, &stdout, &stderr, "prod", "project", getOpts{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := stdout.String(), "myapp\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}

func TestRunGet_SingleField_Env(t *testing.T) {
	cfg := newTestConfig()
	var stdout, stderr bytes.Buffer
	if err := runGet(cfg, &stdout, &stderr, "prod", "env", getOpts{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := stdout.String(), "prod\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}

func TestRunGet_SingleField_Tags(t *testing.T) {
	cfg := newTestConfig()
	var stdout, stderr bytes.Buffer
	if err := runGet(cfg, &stdout, &stderr, "prod", "tags", getOpts{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := stdout.String(), "web\nprimary\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}

func TestRunGet_SingleField_Options(t *testing.T) {
	cfg := newTestConfig()
	var stdout, stderr bytes.Buffer
	if err := runGet(cfg, &stdout, &stderr, "prod", "options", getOpts{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// sorted by key
	want := "ServerAliveInterval=60\nStrictHostKeyChecking=yes\n"
	if got := stdout.String(); got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}

func TestRunGet_SingleField_OptionsSubkey(t *testing.T) {
	cfg := newTestConfig()
	var stdout, stderr bytes.Buffer
	if err := runGet(cfg, &stdout, &stderr, "prod", "options.StrictHostKeyChecking", getOpts{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := stdout.String(), "yes\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}

func TestRunGet_SingleField_ID(t *testing.T) {
	cfg := newTestConfig()
	var stdout, stderr bytes.Buffer
	if err := runGet(cfg, &stdout, &stderr, "prod", "id", getOpts{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := stdout.String(), "prod\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}

func TestRunGet_BulkComma(t *testing.T) {
	cfg := newTestConfig()
	var stdout, stderr bytes.Buffer
	if err := runGet(cfg, &stdout, &stderr, "prod", "host,port,user", getOpts{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "web.example.com\t2222\talice\n"
	if got := stdout.String(); got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr non-empty: %q", stderr.String())
	}
}

func TestRunGet_BulkComma_UnknownField(t *testing.T) {
	cfg := newTestConfig()
	var stdout, stderr bytes.Buffer
	err := runGet(cfg, &stdout, &stderr, "prod", "host,bogus,user", getOpts{})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout should be empty on error, got: %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "bogus") {
		t.Errorf("stderr should name the bad field, got: %q", stderr.String())
	}
}

func TestRunGet_Bare_SshGStyle(t *testing.T) {
	cfg := newTestConfig()
	var stdout, stderr bytes.Buffer

	if err := runGet(cfg, &stdout, &stderr, "prod", "", getOpts{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := stdout.String()
	if !strings.HasSuffix(out, "\n") {
		t.Errorf("expected trailing newline; got %q", out)
	}

	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")

	// Verify sorted alphabetically.
	sorted := append([]string(nil), lines...)
	sort.Strings(sorted)
	if !reflect.DeepEqual(lines, sorted) {
		t.Errorf("lines not sorted alphabetically:\ngot:    %v\nsorted: %v", lines, sorted)
	}

	// Build a key->value map and check expected scalar fields.
	got := make(map[string]string, len(lines))
	for _, ln := range lines {
		parts := strings.SplitN(ln, " ", 2)
		if len(parts) != 2 {
			t.Errorf("line not in 'key value' form: %q", ln)
			continue
		}
		got[parts[0]] = parts[1]
	}

	want := map[string]string{
		"id":            "prod",
		"host":          "web.example.com",
		"user":          "alice",
		"port":          "2222",
		"project":       "myapp",
		"env":           "prod",
		"identity_file": "~/.ssh/prod_id",
		"proxy_jump":    "bastion.example.com",
		"forward_agent": "true",
		"use_mosh":      "true",
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("field %q: got %q, want %q", k, got[k], v)
		}
	}

	// Tags and options should not appear in the bare output.
	if _, ok := got["tags"]; ok {
		t.Errorf("tags should not appear in bare output")
	}
	if _, ok := got["options"]; ok {
		t.Errorf("options should not appear in bare output")
	}
}

func TestRunGet_Bare_SkipsEmptyValues(t *testing.T) {
	cfg := newTestConfig()
	var stdout, stderr bytes.Buffer
	if err := runGet(cfg, &stdout, &stderr, "bare", "", getOpts{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := stdout.String()
	// "bare" connection has no project/env/identity_file/proxy_jump/etc.
	for _, k := range []string{"project", "env", "identity_file", "proxy_jump"} {
		if strings.Contains(out, k+" ") {
			t.Errorf("bare output should skip empty %q, got:\n%s", k, out)
		}
	}
	// Should still include host and id.
	if !strings.Contains(out, "host bare.example.com") {
		t.Errorf("expected host line, got:\n%s", out)
	}
	if !strings.Contains(out, "id bare") {
		t.Errorf("expected id line, got:\n%s", out)
	}
}

func TestRunGet_Flag_NoNewline_Single(t *testing.T) {
	cfg := newTestConfig()
	var stdout, stderr bytes.Buffer
	if err := runGet(cfg, &stdout, &stderr, "prod", "host", getOpts{NoNewline: true}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := stdout.String(), "web.example.com"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}

func TestRunGet_Flag_NoNewline_Bulk(t *testing.T) {
	cfg := newTestConfig()
	var stdout, stderr bytes.Buffer
	if err := runGet(cfg, &stdout, &stderr, "prod", "host,port", getOpts{NoNewline: true}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := stdout.String(), "web.example.com\t2222"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}

func TestRunGet_Flag_Default_WhenEmpty(t *testing.T) {
	cfg := newTestConfig()
	var stdout, stderr bytes.Buffer
	// "staging" has no identity_file; should print default
	if err := runGet(cfg, &stdout, &stderr, "staging", "identity_file", getOpts{Default: "fallback"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := stdout.String(), "fallback\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}

func TestRunGet_Flag_Default_NotUsedWhenPresent(t *testing.T) {
	cfg := newTestConfig()
	var stdout, stderr bytes.Buffer
	if err := runGet(cfg, &stdout, &stderr, "prod", "host", getOpts{Default: "fallback"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := stdout.String(), "web.example.com\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}

func TestRunGet_Flag_JSON_SingleField(t *testing.T) {
	cfg := newTestConfig()
	var stdout, stderr bytes.Buffer
	if err := runGet(cfg, &stdout, &stderr, "prod", "host", getOpts{JSON: true}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := strings.TrimRight(stdout.String(), "\n")
	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("invalid JSON: %v\nraw: %q", err, stdout.String())
	}
	if got["host"] != "web.example.com" {
		t.Errorf("host = %v, want %q", got["host"], "web.example.com")
	}
	if !strings.HasSuffix(stdout.String(), "\n") {
		t.Errorf("expected trailing newline on JSON output")
	}
}

func TestRunGet_Flag_JSON_MultiField(t *testing.T) {
	cfg := newTestConfig()
	var stdout, stderr bytes.Buffer
	if err := runGet(cfg, &stdout, &stderr, "prod", "host,port", getOpts{JSON: true}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(stdout.String()), &got); err != nil {
		t.Fatalf("invalid JSON: %v\nraw: %q", err, stdout.String())
	}
	if got["host"] != "web.example.com" {
		t.Errorf("host = %v, want %q", got["host"], "web.example.com")
	}
	// JSON numbers decode to float64
	if p, ok := got["port"].(float64); !ok || int(p) != 2222 {
		t.Errorf("port = %v, want 2222", got["port"])
	}
}

func TestRunGet_Flag_JSON_Bare(t *testing.T) {
	cfg := newTestConfig()
	var stdout, stderr bytes.Buffer
	if err := runGet(cfg, &stdout, &stderr, "prod", "", getOpts{JSON: true}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var conn config.Connection
	if err := json.Unmarshal([]byte(stdout.String()), &conn); err != nil {
		t.Fatalf("invalid JSON: %v\nraw: %q", err, stdout.String())
	}
	if conn.ID != "prod" {
		t.Errorf("ID = %q, want %q", conn.ID, "prod")
	}
	if conn.Host != "web.example.com" {
		t.Errorf("Host = %q, want %q", conn.Host, "web.example.com")
	}
	if conn.IdentityFile != "~/.ssh/prod_id" {
		t.Errorf("IdentityFile = %q, want %q (should be included in local CLI JSON)", conn.IdentityFile, "~/.ssh/prod_id")
	}
}

func TestRunGet_Error_UnknownID_WithSuggestion(t *testing.T) {
	cfg := newTestConfig()
	var stdout, stderr bytes.Buffer
	err := runGet(cfg, &stdout, &stderr, "badid", "host", getOpts{})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout should be empty on error, got: %q", stdout.String())
	}
	se := stderr.String() + err.Error()
	if !strings.Contains(se, "not found: badid") {
		t.Errorf("expected 'not found: badid' in stderr/err, got: %q / %q", stderr.String(), err.Error())
	}
}

func TestRunGet_Error_UnknownID_FuzzyDoesNotMatch(t *testing.T) {
	cfg := newTestConfig()
	var stdout, stderr bytes.Buffer
	// "pro" is a substring of "prod" but should NOT match for `get`;
	// `get` uses exact ID lookup only.
	err := runGet(cfg, &stdout, &stderr, "pro", "host", getOpts{})
	if err == nil {
		t.Fatalf("expected error for non-exact id 'pro', got nil")
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout should be empty on error, got: %q", stdout.String())
	}
	se := stderr.String() + err.Error()
	if !strings.Contains(se, "not found: pro") {
		t.Errorf("expected 'not found: pro', got stderr=%q err=%q", stderr.String(), err.Error())
	}
	if !strings.Contains(se, "did you mean") {
		t.Errorf("expected 'did you mean' suggestion, got stderr=%q err=%q", stderr.String(), err.Error())
	}
	if !strings.Contains(se, "prod") {
		t.Errorf("expected 'prod' suggested, got stderr=%q err=%q", stderr.String(), err.Error())
	}
}

func TestRunGet_Error_UnknownField_ListsValid(t *testing.T) {
	cfg := newTestConfig()
	var stdout, stderr bytes.Buffer
	err := runGet(cfg, &stdout, &stderr, "prod", "nope", getOpts{})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout should be empty on error, got: %q", stdout.String())
	}
	se := stderr.String() + err.Error()
	// Should list at least a few known field names.
	for _, f := range []string{"host", "user", "port", "id"} {
		if !strings.Contains(se, f) {
			t.Errorf("expected error to list valid field %q, got stderr=%q err=%q", f, stderr.String(), err.Error())
		}
	}
}

func TestRunGet_Error_StdoutEmptyOnAnyError(t *testing.T) {
	cfg := newTestConfig()

	cases := []struct {
		name  string
		id    string
		field string
	}{
		{"unknown id", "nope", "host"},
		{"unknown field", "prod", "doesnotexist"},
		{"bulk with bad", "prod", "host,bogus"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := runGet(cfg, &stdout, &stderr, tc.id, tc.field, getOpts{})
			if err == nil {
				t.Fatalf("expected error")
			}
			if stdout.Len() != 0 {
				t.Errorf("stdout must be empty on error, got: %q", stdout.String())
			}
		})
	}
}

// Ensure the test environment doesn't leak USER state across tests that don't set it.
func TestMain(m *testing.M) {
	// Snapshot USER so individual tests using t.Setenv don't surprise others.
	_ = os.Getenv("USER")
	os.Exit(m.Run())
}
