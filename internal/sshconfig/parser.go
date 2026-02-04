package sshconfig

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/danmartuszewski/hop/internal/config"
)

// ParsedHost represents a parsed SSH config host entry
type ParsedHost struct {
	Alias        string
	HostName     string
	User         string
	Port         int
	IdentityFile string
	ProxyJump    string
	ForwardAgent bool
	Options      map[string]string
}

// Parse reads an SSH config file and returns a list of parsed hosts
func Parse(path string) ([]ParsedHost, error) {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(home, ".ssh", "config")
	}

	return parseFile(path, make(map[string]bool))
}

// parseFile parses a single SSH config file, tracking visited files to avoid cycles
func parseFile(path string, visited map[string]bool) ([]ParsedHost, error) {
	// Resolve to absolute path for cycle detection
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	if visited[absPath] {
		return nil, nil // Already processed this file
	}
	visited[absPath] = true

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	var hosts []ParsedHost
	var current *ParsedHost
	baseDir := filepath.Dir(path)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split into keyword and argument
		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			// Try with = separator
			parts = strings.SplitN(line, "=", 2)
		}
		if len(parts) < 2 {
			continue
		}

		keyword := strings.ToLower(strings.TrimSpace(parts[0]))
		argument := strings.TrimSpace(parts[1])

		// Handle Include directive
		if keyword == "include" {
			includedHosts, err := handleInclude(argument, baseDir, visited)
			if err != nil {
				// Ignore include errors, continue parsing
				continue
			}
			hosts = append(hosts, includedHosts...)
			continue
		}

		// Handle Host directive
		if keyword == "host" {
			// Save previous host if exists
			if current != nil && !isWildcard(current.Alias) {
				hosts = append(hosts, *current)
			}

			current = &ParsedHost{
				Alias:   argument,
				Options: make(map[string]string),
			}
			continue
		}

		// Skip if we're not in a Host block
		if current == nil {
			continue
		}

		// Parse directives within Host block
		switch keyword {
		case "hostname":
			current.HostName = argument
		case "user":
			current.User = argument
		case "port":
			if port, err := strconv.Atoi(argument); err == nil {
				current.Port = port
			}
		case "identityfile":
			current.IdentityFile = expandPath(argument)
		case "proxyjump":
			current.ProxyJump = argument
		case "forwardagent":
			current.ForwardAgent = strings.ToLower(argument) == "yes"
		default:
			// Store other options
			current.Options[keyword] = argument
		}
	}

	// Don't forget the last host
	if current != nil && !isWildcard(current.Alias) {
		hosts = append(hosts, *current)
	}

	if err := scanner.Err(); err != nil {
		return hosts, err
	}

	return hosts, nil
}

// handleInclude processes Include directives
func handleInclude(pattern string, baseDir string, visited map[string]bool) ([]ParsedHost, error) {
	// Expand ~ in pattern
	pattern = expandPath(pattern)

	// If pattern is relative, make it relative to the config file's directory
	if !filepath.IsAbs(pattern) {
		pattern = filepath.Join(baseDir, pattern)
	}

	// Expand glob patterns
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	var allHosts []ParsedHost
	for _, match := range matches {
		hosts, err := parseFile(match, visited)
		if err != nil {
			continue // Ignore errors in included files
		}
		allHosts = append(allHosts, hosts...)
	}

	return allHosts, nil
}

// isWildcard returns true if the host pattern contains wildcards
func isWildcard(pattern string) bool {
	return strings.Contains(pattern, "*") || strings.Contains(pattern, "?")
}

// expandPath expands ~ to home directory
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

// ToConnection converts a ParsedHost to a hop Connection
func (h *ParsedHost) ToConnection() config.Connection {
	conn := config.Connection{
		ID:           h.Alias,
		Host:         h.HostName,
		User:         h.User,
		Port:         h.Port,
		IdentityFile: h.IdentityFile,
		ProxyJump:    h.ProxyJump,
		ForwardAgent: h.ForwardAgent,
	}

	// If HostName is not set, use the alias as hostname
	if conn.Host == "" {
		conn.Host = h.Alias
	}

	// Copy relevant options
	if len(h.Options) > 0 {
		conn.Options = make(map[string]string)
		for k, v := range h.Options {
			conn.Options[k] = v
		}
	}

	return conn
}

// ToConnections converts a slice of ParsedHosts to hop Connections
func ToConnections(hosts []ParsedHost) []config.Connection {
	connections := make([]config.Connection, 0, len(hosts))
	for _, host := range hosts {
		connections = append(connections, host.ToConnection())
	}
	return connections
}

// ResolveConflict generates a unique ID by appending -imported suffix
func ResolveConflict(id string, existingIDs map[string]bool) string {
	candidate := id + "-imported"
	if !existingIDs[candidate] {
		return candidate
	}

	// Try numbered suffixes
	for i := 2; ; i++ {
		candidate = fmt.Sprintf("%s-imported-%d", id, i)
		if !existingIDs[candidate] {
			return candidate
		}
	}
}
