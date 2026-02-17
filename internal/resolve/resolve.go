package resolve

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/danmartuszewski/hop/internal/config"
	"github.com/danmartuszewski/hop/internal/fuzzy"
)

// MatchMethod describes how a target was resolved.
type MatchMethod int

const (
	MatchNamedGroup MatchMethod = iota
	MatchProjectEnv
	MatchGlob
	MatchFuzzy
	MatchNone
)

func (m MatchMethod) String() string {
	switch m {
	case MatchNamedGroup:
		return "named group"
	case MatchProjectEnv:
		return "project-env pattern"
	case MatchGlob:
		return "glob pattern"
	case MatchFuzzy:
		return "fuzzy match"
	default:
		return "none"
	}
}

// ResolveResult contains the resolved connections and how they were matched.
type ResolveResult struct {
	Connections []config.Connection
	Method      MatchMethod
}

// ResolveTarget resolves a target string to connections, returning
// both the matched connections and the method used.
func ResolveTarget(target string, cfg *config.Config) (*ResolveResult, error) {
	// 1. Named group
	if cfg.Groups != nil {
		if members, ok := cfg.Groups[target]; ok {
			var connections []config.Connection
			for _, id := range members {
				if conn := fuzzy.FindByID(id, cfg.Connections); conn != nil {
					connections = append(connections, *conn)
				}
			}
			return &ResolveResult{Connections: connections, Method: MatchNamedGroup}, nil
		}
	}

	// 2. Project-env pattern
	matches := fuzzy.MatchGroup(target, cfg)
	if len(matches) > 0 {
		return &ResolveResult{Connections: matches, Method: MatchProjectEnv}, nil
	}

	// 3. Glob pattern
	if strings.Contains(target, "*") || strings.Contains(target, "?") {
		pattern := globToRegex(target)
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern '%s': %w", target, err)
		}

		var connections []config.Connection
		for _, conn := range cfg.Connections {
			if re.MatchString(conn.ID) {
				connections = append(connections, conn)
			}
		}
		return &ResolveResult{Connections: connections, Method: MatchGlob}, nil
	}

	// 4. Fuzzy match
	if conn := fuzzy.FindBestMatch(target, cfg.Connections); conn != nil {
		return &ResolveResult{Connections: []config.Connection{*conn}, Method: MatchFuzzy}, nil
	}

	return &ResolveResult{Method: MatchNone}, nil
}

// globToRegex converts a simple glob pattern to a regex pattern.
func globToRegex(glob string) string {
	pattern := regexp.QuoteMeta(glob)
	pattern = strings.ReplaceAll(pattern, "\\*", ".*")
	pattern = strings.ReplaceAll(pattern, "\\?", ".")
	return "^" + pattern + "$"
}
