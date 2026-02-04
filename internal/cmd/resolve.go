package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/tabwriter"

	"github.com/danmartuszewski/hop/internal/config"
	"github.com/danmartuszewski/hop/internal/fuzzy"
	"github.com/spf13/cobra"
)

var resolveTag string

var resolveCmd = &cobra.Command{
	Use:   "resolve <target>",
	Short: "Show which connections a target matches",
	Long: `Test target resolution without connecting or executing anything.
Shows which connections would be matched and how the target was resolved.

Target is resolved in this order:
  1. Named group from config (e.g. "production")
  2. Project-env pattern (e.g. "myapp-prod")
  3. Glob pattern on connection IDs (e.g. "web*")
  4. Fuzzy match on connection IDs

Examples:
  hop resolve production
  hop resolve myapp-prod
  hop resolve "web*"
  hop resolve prod --tag=web`,
	Args: cobra.ExactArgs(1),
	RunE: runResolve,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) >= 1 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return getConnectionCompletions(toComplete)
	},
}

func init() {
	rootCmd.AddCommand(resolveCmd)

	resolveCmd.Flags().StringVar(&resolveTag, "tag", "", "filter connections by tag")
}

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

// resolveTarget resolves a target string to connections, returning
// both the matched connections and the method used.
func resolveTarget(target string, cfg *config.Config) (*ResolveResult, error) {
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

func runResolve(cmd *cobra.Command, args []string) error {
	target := args[0]

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	result, err := resolveTarget(target, cfg)
	if err != nil {
		return err
	}

	connections := result.Connections

	// Apply tag filter
	if resolveTag != "" {
		connections = fuzzy.MatchByTag(resolveTag, connections)
	}

	if len(connections) == 0 {
		if resolveTag != "" {
			return fmt.Errorf("no connections matching '%s' with tag '%s'", target, resolveTag)
		}
		return fmt.Errorf("no connections matching '%s'", target)
	}

	fmt.Fprintf(os.Stderr, "Target:   %s\n", target)
	fmt.Fprintf(os.Stderr, "Matched:  %s (%d connection(s))\n", result.Method, len(connections))
	if resolveTag != "" {
		fmt.Fprintf(os.Stderr, "Tag:      %s\n", resolveTag)
	}
	fmt.Fprintf(os.Stderr, "\n")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "ID\tHOST\tPORT\tTAGS\n")
	for _, conn := range connections {
		host := conn.Host
		if user := conn.EffectiveUser(); user != "" {
			host = user + "@" + conn.Host
		}
		tags := strings.Join(conn.Tags, ", ")
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", conn.ID, host, conn.Port, tags)
	}
	return w.Flush()
}
