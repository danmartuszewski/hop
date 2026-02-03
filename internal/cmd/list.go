package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/danmartuszewski/hop/internal/config"
	"github.com/spf13/cobra"
)

var (
	listJSON bool
	listFlat bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all connections",
	Long:  "Display all configured SSH connections, grouped by project and environment.",
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().BoolVar(&listJSON, "json", false, "output as JSON")
	listCmd.Flags().BoolVar(&listFlat, "flat", false, "flat list without grouping")
}

func runList(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	if listJSON {
		return printJSON(cfg.Connections)
	}

	if listFlat {
		return printFlat(cfg.Connections)
	}

	return printGrouped(cfg.Connections)
}

func printJSON(connections []config.Connection) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(connections)
}

func printFlat(connections []config.Connection) error {
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

func printGrouped(connections []config.Connection) error {
	groups := groupConnections(connections)
	projectOrder := sortedKeys(groups)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	for i, project := range projectOrder {
		envs := groups[project]
		envOrder := sortedKeys(envs)

		if project == "" {
			fmt.Fprintf(w, "ID\tHOST\tPORT\tTAGS\n")
			for _, env := range envOrder {
				for _, conn := range envs[env] {
					printConnection(w, conn)
				}
			}
			continue
		}

		if i > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintf(w, "%s\t\t\t\n", project)
		fmt.Fprintf(w, "ID\tHOST\tPORT\tTAGS\n")
		for _, env := range envOrder {
			conns := envs[env]
			if env != "" {
				fmt.Fprintf(w, "  [%s]\t\t\t\n", env)
			}
			for _, conn := range conns {
				prefix := "  "
				if env != "" {
					prefix = "    "
				}
				printConnectionWithPrefix(w, conn, prefix)
			}
		}
	}

	return w.Flush()
}

func printConnection(w *tabwriter.Writer, conn config.Connection) {
	printConnectionWithPrefix(w, conn, "")
}

func printConnectionWithPrefix(w *tabwriter.Writer, conn config.Connection, prefix string) {
	name := conn.ID
	if conn.Project != "" {
		name = strings.TrimPrefix(name, conn.Project+"-")
	}
	if conn.Env != "" {
		name = strings.TrimPrefix(name, conn.Env+"-")
	}

	host := conn.Host
	if user := conn.EffectiveUser(); user != "" {
		host = user + "@" + conn.Host
	}
	tags := strings.Join(conn.Tags, ", ")
	fmt.Fprintf(w, "%s%s\t%s\t%d\t%s\n", prefix, name, host, conn.Port, tags)
}

func groupConnections(connections []config.Connection) map[string]map[string][]config.Connection {
	groups := make(map[string]map[string][]config.Connection)

	for _, conn := range connections {
		project := conn.Project
		env := conn.Env

		if groups[project] == nil {
			groups[project] = make(map[string][]config.Connection)
		}
		groups[project][env] = append(groups[project][env], conn)
	}

	return groups
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	emptyIdx := -1
	for i, k := range keys {
		if k == "" {
			emptyIdx = i
			break
		}
	}
	if emptyIdx > 0 {
		keys = append(keys[emptyIdx:emptyIdx+1], append(keys[:emptyIdx], keys[emptyIdx+1:]...)...)
	}

	return keys
}
