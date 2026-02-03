package tui

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/danmartuszewski/hop/internal/config"
)

// ParseSSHString parses various SSH connection string formats:
// - user@host
// - user@host:port
// - host
// - host:port
// - ssh user@host
// - ssh user@host -p port
// - ssh -p port user@host
// - ssh://user@host:port
func ParseSSHString(input string) *config.Connection {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}

	conn := &config.Connection{
		Port: 22,
	}

	// Remove ssh:// prefix
	input = strings.TrimPrefix(input, "ssh://")

	// Remove "ssh " prefix
	if strings.HasPrefix(input, "ssh ") {
		input = strings.TrimPrefix(input, "ssh ")
	}

	// Extract -p port flag
	portRegex := regexp.MustCompile(`-p\s*(\d+)`)
	if matches := portRegex.FindStringSubmatch(input); len(matches) > 1 {
		if p, err := strconv.Atoi(matches[1]); err == nil {
			conn.Port = p
		}
		input = portRegex.ReplaceAllString(input, "")
		input = strings.TrimSpace(input)
	}

	// Remove other common ssh flags we don't care about
	flagRegex := regexp.MustCompile(`-[A-Za-z]\s*\S*`)
	input = flagRegex.ReplaceAllString(input, "")
	input = strings.TrimSpace(input)

	// Now parse user@host:port or user@host or host:port or host
	if strings.Contains(input, "@") {
		parts := strings.SplitN(input, "@", 2)
		conn.User = strings.TrimSpace(parts[0])
		input = parts[1]
	}

	// Check for host:port
	if strings.Contains(input, ":") && !strings.Contains(input, "[") {
		// Not IPv6
		parts := strings.SplitN(input, ":", 2)
		conn.Host = strings.TrimSpace(parts[0])
		if p, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
			conn.Port = p
		}
	} else if strings.Contains(input, "]:") {
		// IPv6 with port: [::1]:22
		idx := strings.LastIndex(input, "]:")
		conn.Host = strings.TrimSpace(input[:idx+1])
		if p, err := strconv.Atoi(strings.TrimSpace(input[idx+2:])); err == nil {
			conn.Port = p
		}
	} else {
		conn.Host = strings.TrimSpace(input)
	}

	// Generate an ID from host
	if conn.Host != "" {
		conn.ID = generateID(conn.Host)
	}

	if conn.Host == "" {
		return nil
	}

	return conn
}

func generateID(host string) string {
	// Remove common suffixes
	id := host
	id = strings.TrimSuffix(id, ".local")
	id = strings.TrimSuffix(id, ".lan")

	// Replace dots with dashes
	id = strings.ReplaceAll(id, ".", "-")

	// Remove port-like suffixes
	id = regexp.MustCompile(`-\d+$`).ReplaceAllString(id, "")

	// Clean up
	id = strings.Trim(id, "-")

	return id
}
