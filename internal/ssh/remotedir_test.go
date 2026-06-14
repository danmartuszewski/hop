package ssh

import (
	"strings"
	"testing"

	"github.com/danmartuszewski/hop/internal/config"
)

func TestPosixQuote(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", "''"},
		{"plain path", "/var/www", "/var/www"},
		{"safe chars", "/opt/app-1.2_x:y", "/opt/app-1.2_x:y"},
		{"space", "/srv/my app", `'/srv/my app'`},
		{"single quote", "/opt/it's", `'/opt/it'\''s'`},
		{"semicolon", "/a;rm -rf", `'/a;rm -rf'`},
		{"dollar", "/home/$USER", `'/home/$USER'`},
		{"double quote", `/a"b`, `'/a"b'`},
		{"backtick", "/a`b", "'/a`b'"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := posixQuote(tt.in); got != tt.want {
				t.Errorf("posixQuote(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestRemoteDirCommand(t *testing.T) {
	tests := []struct {
		name string
		dir  string
		want string
	}{
		{
			"plain absolute",
			"/srv/app",
			`cd -- /srv/app; exec "${SHELL:-/bin/sh}" -l`,
		},
		{
			// Needs quoting; cd uses ";" (forgiving) not "&&".
			"with space",
			"/srv/my app",
			`cd -- '/srv/my app'; exec "${SHELL:-/bin/sh}" -l`,
		},
		{
			// Leading "-" must not be parsed as a cd flag: "cd --" guards it.
			"leading dash",
			"-p",
			`cd -- -p; exec "${SHELL:-/bin/sh}" -l`,
		},
		{
			// Tilde must stay unquoted so the remote shell expands it to $HOME.
			"tilde path",
			"~/projects",
			`cd -- ~/projects; exec "${SHELL:-/bin/sh}" -l`,
		},
		{
			// Tilde prefix unquoted, remainder quoted.
			"tilde path with space",
			"~/my projects",
			`cd -- ~/'my projects'; exec "${SHELL:-/bin/sh}" -l`,
		},
		{
			"tilde user",
			"~deploy/app",
			`cd -- ~deploy/app; exec "${SHELL:-/bin/sh}" -l`,
		},
		{
			"bare tilde",
			"~",
			`cd -- ~; exec "${SHELL:-/bin/sh}" -l`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RemoteDirCommand(tt.dir); got != tt.want {
				t.Errorf("RemoteDirCommand(%q) = %q, want %q", tt.dir, got, tt.want)
			}
		})
	}
}

func TestBuildCommand_RemoteDir(t *testing.T) {
	conn := &config.Connection{Host: "example.com", User: "admin", RemoteDir: "/srv/app"}
	args := BuildCommand(conn, nil)

	// -t must be forced so the interactive login shell gets a TTY.
	if args[0] != "-t" {
		t.Fatalf("expected -t to be forced, got args %v", args)
	}
	// The landing-dir command is the final argument and unmangled.
	last := args[len(args)-1]
	want := `cd -- /srv/app; exec "${SHELL:-/bin/sh}" -l`
	if last != want {
		t.Errorf("remote command = %q, want %q", last, want)
	}
	// Destination precedes the remote command.
	if args[len(args)-2] != "admin@example.com" {
		t.Errorf("destination misplaced in %v", args)
	}
}

func TestBuildCommand_ExplicitCommandWinsOverRemoteDir(t *testing.T) {
	conn := &config.Connection{Host: "example.com", RemoteDir: "/srv/app"}
	args := BuildCommand(conn, &ConnectOptions{Command: "uptime"})
	if last := args[len(args)-1]; last != "uptime" {
		t.Errorf("explicit command should win, got %q", last)
	}
}

func TestBuildCommand_NoRemoteDir_NoTTY(t *testing.T) {
	conn := &config.Connection{Host: "example.com"}
	args := BuildCommand(conn, nil)
	for _, a := range args {
		if a == "-t" {
			t.Fatalf("did not expect -t without RemoteDir/ForceTTY: %v", args)
		}
	}
}

// TestBuildCommandString_RemoteDir_TerminalSafe asserts the flattened command
// survives a single round-trip through "sh -c" — the embedding used by the
// majority of supported terminal launchers (Alacritty, Konsole, Kitty, GNOME,
// WezTerm, Ghostty-linux, Terminator, ...). The remote command must remain a
// single argument with its shell metacharacters intact.
func TestBuildCommandString_RemoteDir_TerminalSafe(t *testing.T) {
	conn := &config.Connection{Host: "example.com", User: "admin", RemoteDir: "/srv/my app"}
	cmdStr := BuildCommandString(conn, nil)

	// The whole landing-dir command is single-quoted as one unit so the local
	// shell does not re-interpret ";" or the inner quotes.
	wantFragment := `'cd -- '\''/srv/my app'\''; exec "${SHELL:-/bin/sh}" -l'`
	if !strings.Contains(cmdStr, wantFragment) {
		t.Errorf("BuildCommandString() = %q\n missing fragment %q", cmdStr, wantFragment)
	}
	if !strings.HasPrefix(cmdStr, "ssh -t ") {
		t.Errorf("expected forced TTY in %q", cmdStr)
	}
}

func TestBuildMoshCommand_RemoteDir_WrapsInShell(t *testing.T) {
	conn := &config.Connection{Host: "example.com", UseMosh: boolPtr(true), RemoteDir: "/srv/app"}
	_, args := BuildMoshCommand(conn, nil)

	// mosh execs the -- argv directly, so the shell-syntax landing command must
	// go through "sh -c".
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, `-- sh -c cd -- /srv/app; exec "${SHELL:-/bin/sh}" -l`) {
		t.Errorf("mosh args = %v, expected sh -c wrapping", args)
	}
}
