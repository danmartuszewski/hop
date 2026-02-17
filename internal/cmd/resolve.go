package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/danmartuszewski/hop/internal/fuzzy"
	"github.com/danmartuszewski/hop/internal/resolve"
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

func runResolve(cmd *cobra.Command, args []string) error {
	target := args[0]

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	result, err := resolve.ResolveTarget(target, cfg)
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
