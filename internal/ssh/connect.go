package ssh

import (
	"fmt"
	"os"
	"os/exec"
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

	if opts != nil && opts.ForceTTY {
		args = append(args, "-t")
	}

	if conn.Port != 0 && conn.Port != 22 {
		args = append(args, "-p", fmt.Sprintf("%d", conn.Port))
	}

	if conn.IdentityFile != "" {
		identityFile := expandPath(conn.IdentityFile)
		args = append(args, "-i", identityFile)
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
	args = append(args, destination)

	if opts != nil && opts.Command != "" {
		args = append(args, opts.Command)
	}

	return args
}

func BuildCommandString(conn *config.Connection, opts *ConnectOptions) string {
	args := BuildCommand(conn, opts)
	quotedArgs := make([]string, len(args))
	for i, arg := range args {
		if strings.ContainsAny(arg, " \t\n\"'") {
			quotedArgs[i] = fmt.Sprintf("%q", arg)
		} else {
			quotedArgs[i] = arg
		}
	}
	return "ssh " + strings.Join(quotedArgs, " ")
}

func Connect(conn *config.Connection, opts *ConnectOptions) error {
	args := BuildCommand(conn, opts)

	if opts != nil && opts.DryRun {
		fmt.Println(BuildCommandString(conn, opts))
		return nil
	}

	cmd := exec.Command("ssh", args...)
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
