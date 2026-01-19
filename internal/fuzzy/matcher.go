package fuzzy

import (
	"sort"
	"strings"

	"github.com/hop-cli/hop/internal/config"
)

type Match struct {
	Connection *config.Connection
	Score      int
}

func FindMatches(query string, connections []config.Connection) []Match {
	if query == "" {
		return nil
	}

	query = strings.ToLower(query)
	var matches []Match

	for i := range connections {
		conn := &connections[i]
		score := scoreConnection(query, conn)
		if score > 0 {
			matches = append(matches, Match{
				Connection: conn,
				Score:      score,
			})
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Score != matches[j].Score {
			return matches[i].Score > matches[j].Score
		}
		return len(matches[i].Connection.ID) < len(matches[j].Connection.ID)
	})

	return matches
}

func FindBestMatch(query string, connections []config.Connection) *config.Connection {
	matches := FindMatches(query, connections)
	if len(matches) == 0 {
		return nil
	}
	return matches[0].Connection
}

func FindByID(id string, connections []config.Connection) *config.Connection {
	for i := range connections {
		if connections[i].ID == id {
			return &connections[i]
		}
	}
	return nil
}

func scoreConnection(query string, conn *config.Connection) int {
	score := 0

	idLower := strings.ToLower(conn.ID)
	if idLower == query {
		return 1000
	}

	if strings.HasPrefix(idLower, query) {
		score = 100 + (100 - len(conn.ID))
	} else if strings.Contains(idLower, query) {
		score = 50 + (50 - len(conn.ID))
	} else if fuzzyMatch(query, idLower) {
		score = 25 + (25 - len(conn.ID))
	}

	hostLower := strings.ToLower(conn.Host)
	if strings.Contains(hostLower, query) {
		if score < 40 {
			score = 40
		}
	}

	for _, tag := range conn.Tags {
		tagLower := strings.ToLower(tag)
		if tagLower == query {
			if score < 80 {
				score = 80
			}
		} else if strings.Contains(tagLower, query) {
			if score < 30 {
				score = 30
			}
		}
	}

	if conn.Project != "" && strings.Contains(strings.ToLower(conn.Project), query) {
		if score < 35 {
			score = 35
		}
	}

	if conn.Env != "" && strings.Contains(strings.ToLower(conn.Env), query) {
		if score < 35 {
			score = 35
		}
	}

	return score
}

func fuzzyMatch(pattern, text string) bool {
	pIdx := 0
	for tIdx := 0; tIdx < len(text) && pIdx < len(pattern); tIdx++ {
		if text[tIdx] == pattern[pIdx] {
			pIdx++
		}
	}
	return pIdx == len(pattern)
}

func MatchGroup(query string, cfg *config.Config) []config.Connection {
	if cfg.Groups != nil {
		if members, ok := cfg.Groups[query]; ok {
			var connections []config.Connection
			for _, id := range members {
				if conn := FindByID(id, cfg.Connections); conn != nil {
					connections = append(connections, *conn)
				}
			}
			return connections
		}
	}

	queryLower := strings.ToLower(query)
	var matches []config.Connection

	for _, conn := range cfg.Connections {
		projectEnv := strings.ToLower(conn.Project + "-" + conn.Env)
		if projectEnv == queryLower || strings.HasPrefix(projectEnv, queryLower) {
			matches = append(matches, conn)
			continue
		}

		if conn.Project != "" && strings.ToLower(conn.Project) == queryLower {
			matches = append(matches, conn)
			continue
		}

		if conn.Env != "" && strings.ToLower(conn.Env) == queryLower {
			matches = append(matches, conn)
		}
	}

	return matches
}

func MatchByTag(tag string, connections []config.Connection) []config.Connection {
	tagLower := strings.ToLower(tag)
	var matches []config.Connection

	for _, conn := range connections {
		for _, t := range conn.Tags {
			if strings.ToLower(t) == tagLower {
				matches = append(matches, conn)
				break
			}
		}
	}

	return matches
}

// ContainsIgnoreCase returns true if text contains substr, case-insensitively.
func ContainsIgnoreCase(text, substr string) bool {
	return strings.Contains(strings.ToLower(text), strings.ToLower(substr))
}
