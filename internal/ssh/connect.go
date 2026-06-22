package ssh

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/danmartuszewski/hop/internal/config"
)

type ConnectOptions struct {
	DryRun    bool
	ForceTTY  bool
	Command   string
	ExtraArgs []string
}

func BuildCommand(conn *config.Connection, opts *ConnectOptions) []string {
	args := []string{}

	remoteCmd, autoDir := resolveRemoteCommand(conn, opts)

	// A landing-dir command launches an interactive shell, which needs a TTY.
	if (opts != nil && opts.ForceTTY) || autoDir {
		args = append(args, "-t")
	}

	if conn.Port != 0 && conn.Port != 22 {
		args = append(args, "-p", fmt.Sprintf("%d", conn.Port))
	}

	if conn.IdentityFile != "" {
		identityFile := expandPath(conn.IdentityFile)
		args = append(args, "-i", identityFile)
	}

	if conn.ProxyJump != "" {
		args = append(args, "-J", conn.ProxyJump)
	}

	if conn.ForwardAgent {
		args = append(args, "-A")
	}

	for key, value := range conn.Options {
		args = append(args, "-o", fmt.Sprintf("%s=%s", key, value))
	}

	if opts != nil {
		args = append(args, opts.ExtraArgs...)
	}

	user := conn.EffectiveUser()
	destination := conn.Host
	if user != "" {
		destination = user + "@" + conn.Host
	}

	// "--" ends option parsing so a destination beginning with "-" can never be
	// interpreted as an ssh flag (e.g. "-oProxyCommand=..."), which would run an
	// arbitrary command on the local machine (CWE-88 argument injection).
	args = append(args, "--", destination)

	if remoteCmd != "" {
		args = append(args, remoteCmd)
	}

	return args
}

// BuildMoshCommand builds the mosh command arguments for a connection.
// It returns the binary name ("mosh") and the argument list.
func BuildMoshCommand(conn *config.Connection, opts *ConnectOptions) (string, []string) {
	var moshArgs []string

	remoteCmd, autoDir := resolveRemoteCommand(conn, opts)

	// Build inner SSH options for the --ssh flag
	var sshParts []string
	sshParts = append(sshParts, "ssh")

	if conn.Port != 0 && conn.Port != 22 {
		sshParts = append(sshParts, "-p", fmt.Sprintf("%d", conn.Port))
	}

	if conn.IdentityFile != "" {
		identityFile := expandPath(conn.IdentityFile)
		sshParts = append(sshParts, "-i", identityFile)
	}

	if conn.ProxyJump != "" {
		sshParts = append(sshParts, "-J", conn.ProxyJump)
	}

	if conn.ForwardAgent {
		sshParts = append(sshParts, "-A")
	}

	// Sort option keys for deterministic output
	if len(conn.Options) > 0 {
		keys := make([]string, 0, len(conn.Options))
		for k := range conn.Options {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			sshParts = append(sshParts, "-o", fmt.Sprintf("%s=%s", key, conn.Options[key]))
		}
	}

	// Only add --ssh if we have non-default SSH options
	if len(sshParts) > 1 {
		moshArgs = append(moshArgs, fmt.Sprintf("--ssh=%s", strings.Join(sshParts, " ")))
	}

	// Extra args go to mosh
	if opts != nil {
		moshArgs = append(moshArgs, opts.ExtraArgs...)
	}

	// Destination. mosh takes the destination as a positional argument and uses
	// "--" to separate its own options from the remote command, so — unlike ssh —
	// a leading "--" cannot guard the destination here. A destination beginning
	// with "-" is instead rejected up front by Connection.CheckSafety (enforced on
	// config load, at import time, and again in Connect before launch).
	user := conn.EffectiveUser()
	destination := conn.Host
	if user != "" {
		destination = user + "@" + conn.Host
	}
	moshArgs = append(moshArgs, destination)

	// Remote command (mosh uses -- to pass the server command). mosh execs this
	// argv directly rather than via a shell, so the landing-dir command — which
	// relies on shell syntax (";", "exec", parameter expansion) — must be run
	// through "sh -c". An explicit user command is passed as-is, matching prior
	// behaviour.
	if remoteCmd != "" {
		if autoDir {
			moshArgs = append(moshArgs, "--", "sh", "-c", remoteCmd)
		} else {
			moshArgs = append(moshArgs, "--", remoteCmd)
		}
	}

	return "mosh", moshArgs
}

func BuildCommandString(conn *config.Connection, opts *ConnectOptions) string {
	if conn.Mosh() {
		return buildMoshCommandString(conn, opts)
	}
	args := BuildCommand(conn, opts)
	quotedArgs := make([]string, len(args))
	for i, arg := range args {
		quotedArgs[i] = posixQuote(arg)
	}
	return "ssh " + strings.Join(quotedArgs, " ")
}

func buildMoshCommandString(conn *config.Connection, opts *ConnectOptions) string {
	binary, args := BuildMoshCommand(conn, opts)
	quotedArgs := make([]string, len(args))
	for i, arg := range args {
		quotedArgs[i] = posixQuote(arg)
	}
	return binary + " " + strings.Join(quotedArgs, " ")
}

func Connect(conn *config.Connection, opts *ConnectOptions) error {
	// Last line of defense against argument injection: refuse to launch a
	// connection whose host/user/proxy-jump could be parsed as an ssh/mosh option,
	// even if it somehow bypassed load- and import-time validation.
	if err := conn.CheckSafety(); err != nil {
		return err
	}

	if opts != nil && opts.DryRun {
		fmt.Println(BuildCommandString(conn, opts))
		return nil
	}

	var binary string
	var args []string
	if conn.Mosh() {
		binary, args = BuildMoshCommand(conn, opts)
	} else {
		binary = "ssh"
		args = BuildCommand(conn, opts)
	}

	cmd := exec.Command(binary, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return wrapSSHError(err, conn)
	}
	return nil
}

// SSHError wraps an SSH error with a helpful suggestion
type SSHError struct {
	Original   error
	Suggestion string
	Host       string
}

func (e *SSHError) Error() string {
	if e.Suggestion != "" {
		return fmt.Sprintf("%v\n\n  Suggestion: %s", e.Original, e.Suggestion)
	}
	return e.Original.Error()
}

func (e *SSHError) Unwrap() error {
	return e.Original
}

// wrapSSHError wraps common SSH errors with helpful suggestions
func wrapSSHError(err error, conn *config.Connection) error {
	errStr := err.Error()

	// Check for common SSH error patterns
	sshErr := &SSHError{Original: err, Host: conn.Host}

	switch {
	case strings.Contains(errStr, "Permission denied"):
		identityHint := "~/.ssh/id_rsa"
		if conn.IdentityFile != "" {
			identityHint = conn.IdentityFile
		}
		sshErr.Suggestion = fmt.Sprintf("Check your SSH key permissions.\n  Try: ssh-add %s\n  Or check that %s has the correct public key.", identityHint, conn.Host)
	case strings.Contains(errStr, "Connection refused"):
		port := conn.Port
		if port == 0 {
			port = 22
		}
		sshErr.Suggestion = fmt.Sprintf("SSH service may not be running on %s:%d.\n  Check if the SSH daemon is running and the port is correct.", conn.Host, port)
	case strings.Contains(errStr, "Host key verification failed"):
		sshErr.Suggestion = fmt.Sprintf("The host key has changed or is unknown.\n  If this is expected, run: ssh-keygen -R %s", conn.Host)
	case strings.Contains(errStr, "Connection timed out") || strings.Contains(errStr, "Operation timed out"):
		sshErr.Suggestion = fmt.Sprintf("Connection timed out reaching %s.\n  Check network connectivity and firewall rules.", conn.Host)
	case strings.Contains(errStr, "No route to host"):
		sshErr.Suggestion = fmt.Sprintf("Cannot reach %s.\n  Check network connectivity and that the hostname resolves correctly.", conn.Host)
	case strings.Contains(errStr, "Could not resolve hostname"):
		sshErr.Suggestion = fmt.Sprintf("Could not resolve hostname %s.\n  Check the hostname spelling or DNS configuration.", conn.Host)
	case strings.Contains(errStr, "Too many authentication failures"):
		sshErr.Suggestion = "Too many authentication failures.\n  Try: ssh-add -D && ssh-add ~/.ssh/your_key"
	default:
		// Return original error without modification for unknown errors
		return err
	}

	return sshErr
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return home + path[1:]
		}
	}
	return path
}
