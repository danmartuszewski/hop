package ssh

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/hop-cli/hop/internal/config"
)

type ConnectOptions struct {
	DryRun     bool
	ForceTTY   bool
	Command    string
	ExtraArgs  []string
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

	return cmd.Run()
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
