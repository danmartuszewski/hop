package ssh

import (
	"fmt"
	"strings"

	"github.com/danmartuszewski/hop/internal/config"
)

// RemoteDirCommand returns a remote command that changes into dir and then
// replaces itself with an interactive login shell, so the SSH session lands in
// dir instead of $HOME.
//
// Design notes:
//   - cd failures are non-fatal (";" rather than "&&"): a stale or missing
//     directory prints an error but still drops the user into a shell in $HOME
//     rather than bouncing them off the host entirely.
//   - exec replaces the wrapper process so the session feels identical to a
//     normal login — no extra shell layer, signals and exit codes behave.
//   - "${SHELL:-/bin/sh}" honours the user's real login shell and falls back to
//     /bin/sh if SHELL is unset. This is POSIX parameter expansion, valid in
//     sh/bash/zsh/ksh/dash (the shells sshd runs the remote command under).
//   - dir is quoted via quoteCdTarget so paths with spaces or metacharacters
//     survive the remote shell. The whole command is quoted again by the local
//     terminal layer (see BuildCommandString), which is why the quoting must be
//     correct shell quoting, not Go's %q.
//   - "cd --" stops a directory whose name begins with "-" (e.g. "-p") from
//     being parsed as a cd option.
func RemoteDirCommand(dir string) string {
	return fmt.Sprintf("cd -- %s; exec \"${SHELL:-/bin/sh}\" -l", quoteCdTarget(dir))
}

// quoteCdTarget quotes dir for use as the operand of "cd --". A leading "~" or
// "~user" prefix is left unquoted so the remote shell still performs tilde
// expansion — users naturally type "~/path" in a "start in" field — while the
// remainder is POSIX-quoted so spaces and metacharacters are safe. Paths that
// do not start with "~" are quoted as a whole.
func quoteCdTarget(dir string) string {
	if strings.HasPrefix(dir, "~") {
		rest := dir[1:]
		if slash := strings.IndexByte(rest, '/'); slash == -1 {
			// Bare "~" or "~user" with no path component.
			if isShellSafe(rest) {
				return dir
			}
		} else if user := rest[:slash]; isShellSafe(user) {
			// "~" + optional user + "/" + quoted remainder.
			return "~" + user + "/" + posixQuote(rest[slash+1:])
		}
	}
	return posixQuote(dir)
}

// resolveRemoteCommand decides what command, if any, the connection should run
// on the remote host. An explicit opts.Command (e.g. "hop exec" or `connect --
// cmd`) always wins. Otherwise a configured RemoteDir produces a landing-dir
// command. The bool reports whether the command was auto-injected for RemoteDir,
// which the callers use to force a TTY (an interactive shell needs one) and, for
// mosh, to wrap it in `sh -c`.
func resolveRemoteCommand(conn *config.Connection, opts *ConnectOptions) (cmd string, autoDir bool) {
	if opts != nil && opts.Command != "" {
		return opts.Command, false
	}
	if conn.RemoteDir != "" {
		return RemoteDirCommand(conn.RemoteDir), true
	}
	return "", false
}

// posixQuote quotes s so it survives word-splitting and metacharacter
// interpretation by a POSIX shell. Strings made up solely of characters that are
// already safe are returned unchanged (keeps generated commands readable);
// everything else is wrapped in single quotes, with any embedded single quote
// closed-escaped-reopened (the standard close-quote, backslash-quote, reopen
// idiom).
func posixQuote(s string) string {
	if s == "" {
		return "''"
	}
	if isShellSafe(s) {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// isShellSafe reports whether s consists only of characters that need no
// quoting in any POSIX shell context.
func isShellSafe(s string) bool {
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		case strings.ContainsRune("_-./:@%+,=", r):
		default:
			return false
		}
	}
	return true
}
