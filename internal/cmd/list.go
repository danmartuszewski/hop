package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/hop-cli/hop/internal/config"
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
	for _, conn := range connections {
		user := conn.EffectiveUser()
		if user != "" {
			fmt.Printf("%s\t%s@%s\n", conn.ID, user, conn.Host)
		} else {
			fmt.Printf("%s\t%s\n", conn.ID, conn.Host)
		}
	}
	return nil
}

func printGrouped(connections []config.Connection) error {
	groups := groupConnections(connections)
	projectOrder := sortedKeys(groups)

	for _, project := range projectOrder {
		envs := groups[project]
		envOrder := sortedKeys(envs)

		if project == "" {
			for _, env := range envOrder {
				for _, conn := range envs[env] {
					printConnection(conn, "")
				}
			}
			continue
		}

		fmt.Printf("%s\n", project)
		for _, env := range envOrder {
			conns := envs[env]
			if env != "" {
				fmt.Printf("  %s\n", env)
				for _, conn := range conns {
					printConnection(conn, "    ")
				}
			} else {
				for _, conn := range conns {
					printConnection(conn, "  ")
				}
			}
		}
		fmt.Println()
	}

	return nil
}

func printConnection(conn config.Connection, indent string) {
	user := conn.EffectiveUser()
	name := conn.ID
	if conn.Project != "" {
		name = strings.TrimPrefix(name, conn.Project+"-")
	}
	if conn.Env != "" {
		name = strings.TrimPrefix(name, conn.Env+"-")
	}

	if user != "" {
		fmt.Printf("%s%s\t%s\t%s@%s\n", indent, name, conn.ID, user, conn.Host)
	} else {
		fmt.Printf("%s%s\t%s\t%s\n", indent, name, conn.ID, conn.Host)
	}
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
