package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/danmartuszewski/hop/internal/config"
	"github.com/danmartuszewski/hop/internal/fuzzy"
	"github.com/spf13/cobra"
)

// getOpts holds the flag values for the `hop get` command.
type getOpts struct {
	NoNewline bool
	Default   string
	JSON      bool
}

// getFieldNames is the canonical, ordered list of supported field names for
// scalar lookups. Order is used in `--help` output and unknown-field errors.
var getFieldNames = []string{
	"id",
	"host",
	"user",
	"port",
	"identity_file",
	"proxy_jump",
	"forward_agent",
	"use_mosh",
	"project",
	"env",
	"tags",
	"options",
}

// getBareFields are the fields included (when non-empty) in the bare
// `hop get <id>` ssh -G style output. tags and options are intentionally
// excluded so each line is a simple "key value" pair.
var getBareFields = []string{
	"id",
	"host",
	"user",
	"port",
	"identity_file",
	"proxy_jump",
	"forward_agent",
	"use_mosh",
	"project",
	"env",
}

// getFieldResolvers returns the value of a single field for a given
// connection as a string. Multi-line values (tags, options) embed newlines.
var getFieldResolvers = map[string]func(*config.Config, *config.Connection) string{
	"id":   func(_ *config.Config, c *config.Connection) string { return c.ID },
	"host": func(_ *config.Config, c *config.Connection) string { return c.Host },
	"user": func(_ *config.Config, c *config.Connection) string { return c.EffectiveUser() },
	"port": func(_ *config.Config, c *config.Connection) string {
		if c.Port == 0 {
			return ""
		}
		return strconv.Itoa(c.Port)
	},
	"identity_file": func(_ *config.Config, c *config.Connection) string { return c.IdentityFile },
	"proxy_jump":    func(_ *config.Config, c *config.Connection) string { return c.ProxyJump },
	"forward_agent": func(_ *config.Config, c *config.Connection) string {
		return strconv.FormatBool(c.ForwardAgent)
	},
	"use_mosh": func(_ *config.Config, c *config.Connection) string {
		return strconv.FormatBool(c.Mosh())
	},
	"project": func(_ *config.Config, c *config.Connection) string { return c.Project },
	"env":     func(_ *config.Config, c *config.Connection) string { return c.Env },
	"tags": func(_ *config.Config, c *config.Connection) string {
		if len(c.Tags) == 0 {
			return ""
		}
		return strings.Join(c.Tags, "\n")
	},
	"options": func(_ *config.Config, c *config.Connection) string {
		if len(c.Options) == 0 {
			return ""
		}
		keys := make([]string, 0, len(c.Options))
		for k := range c.Options {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var b strings.Builder
		for i, k := range keys {
			if i > 0 {
				b.WriteByte('\n')
			}
			b.WriteString(k)
			b.WriteByte('=')
			b.WriteString(c.Options[k])
		}
		return b.String()
	},
}

// getJSONValueResolvers returns the JSON-typed value for a field. Numeric and
// boolean fields stay native types so JSON consumers can use them directly.
var getJSONValueResolvers = map[string]func(*config.Config, *config.Connection) any{
	"id":            func(_ *config.Config, c *config.Connection) any { return c.ID },
	"host":          func(_ *config.Config, c *config.Connection) any { return c.Host },
	"user":          func(_ *config.Config, c *config.Connection) any { return c.EffectiveUser() },
	"port":          func(_ *config.Config, c *config.Connection) any { return c.Port },
	"identity_file": func(_ *config.Config, c *config.Connection) any { return c.IdentityFile },
	"proxy_jump":    func(_ *config.Config, c *config.Connection) any { return c.ProxyJump },
	"forward_agent": func(_ *config.Config, c *config.Connection) any { return c.ForwardAgent },
	"use_mosh":      func(_ *config.Config, c *config.Connection) any { return c.Mosh() },
	"project":       func(_ *config.Config, c *config.Connection) any { return c.Project },
	"env":           func(_ *config.Config, c *config.Connection) any { return c.Env },
	"tags":          func(_ *config.Config, c *config.Connection) any { return c.Tags },
	"options":       func(_ *config.Config, c *config.Connection) any { return c.Options },
}

var (
	getNoNewline bool
	getDefault   string
	getJSON      bool
)

var getCmd = &cobra.Command{
	Use:   "get <id> [field|field1,field2,...]",
	Short: "Print connection field values for scripting",
	Long: `Print one or more field values for a connection in scriptable form.

Looks up a connection by its exact ID and prints field values to stdout, one
value per line by default. Designed for use in shell pipelines and command
substitution (similar in spirit to ` + "`ssh -G`" + `).

ID resolution is exact only — no fuzzy matching. If the ID is not found, the
error message includes "did you mean: ..." with the closest matches.

Supported fields:
  id              Connection ID
  host            Hostname or IP
  user            Effective user (connection > defaults > $USER)
  port            Port number
  identity_file   Path to SSH identity file
  proxy_jump      ProxyJump host
  forward_agent   "true" or "false"
  use_mosh        "true" or "false" (handles unset as false)
  project         Project label
  env             Environment label
  tags            Tags, one per line
  options         SSH options as sorted key=value lines
  options.<key>   Single SSH option value (e.g. options.StrictHostKeyChecking)

Output modes:
  hop get <id> <field>             Print value followed by newline
  hop get <id> f1,f2,f3            Print values tab-separated on one line
  hop get <id>                     Print all non-empty fields as sorted
                                   "key value" lines (ssh -G style); tags
                                   and options are omitted

With --json, a single field becomes {"field": value}; multiple fields become
{"f1": v1, "f2": v2}; bare (no field) emits the full Connection object.

On any error, exits 1 with the message on stderr and nothing on stdout. The
error lists valid field names (for unknown fields) or fuzzy ID suggestions
(for unknown IDs).`,
	Example: `  # Use in command substitution to build an ssh invocation:
  ssh -i "$(hop get prod identity_file)" "$(hop get prod user)@$(hop get prod host)"

  # Print multiple fields tab-separated for parsing with cut/awk:
  hop get prod host,port,user

  # Provide a fallback when a field is empty:
  hop get staging identity_file --default ~/.ssh/id_rsa

  # Show every non-empty scalar field for a connection:
  hop get prod

  # Emit a JSON object for scripting with jq:
  hop get prod host,port --json | jq -r .host

  # Read a single SSH option:
  hop get prod options.StrictHostKeyChecking`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		id := args[0]
		fieldArg := ""
		if len(args) == 2 {
			fieldArg = args[1]
		}

		opts := getOpts{
			NoNewline: getNoNewline,
			Default:   getDefault,
			JSON:      getJSON,
		}

		return silent(runGet(cfg, os.Stdout, os.Stderr, id, fieldArg, opts))
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		switch len(args) {
		case 0:
			return getConnectionCompletions(toComplete)
		case 1:
			return append([]string(nil), getFieldNames...), cobra.ShellCompDirectiveNoFileComp
		default:
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
	},
}

func init() {
	rootCmd.AddCommand(getCmd)

	getCmd.Flags().BoolVarP(&getNoNewline, "no-newline", "n", false, "suppress trailing newline")
	getCmd.Flags().StringVar(&getDefault, "default", "", "value to print if the requested field is empty")
	getCmd.Flags().BoolVar(&getJSON, "json", false, "emit JSON output")
}

// runGet implements `hop get`. It writes only on success: any error path
// leaves stdout untouched. Error messages are also mirrored to stderr so
// callers that capture stderr separately (tests, scripts redirecting 2>) see
// the bad-field/bad-id detail without parsing the returned error.
func runGet(cfg *config.Config, stdout, stderr io.Writer, id string, fieldArg string, opts getOpts) error {
	conn := fuzzy.FindByID(id, cfg.Connections)
	if conn == nil {
		return reportError(stderr, unknownIDError(id, cfg))
	}

	// Bare form: `hop get <id>` with no field argument.
	if fieldArg == "" {
		return runGetBare(cfg, stdout, conn, opts)
	}

	fields := strings.Split(fieldArg, ",")
	for i, f := range fields {
		fields[i] = strings.TrimSpace(f)
	}

	// Validate every field up front so a single bad entry in a comma list
	// produces an error before any output is written.
	for _, f := range fields {
		if !isValidField(f) {
			return reportError(stderr, unknownFieldError(f))
		}
	}

	if opts.JSON {
		return runGetJSON(cfg, stdout, conn, fields, opts)
	}

	return runGetPlain(cfg, stdout, conn, fields, opts)
}

// runGetBare prints all non-empty scalar fields as "key value" lines,
// alphabetically sorted, mimicking `ssh -G`. Tags and options are excluded.
// With opts.JSON, the entire Connection is emitted as a JSON object.
func runGetBare(cfg *config.Config, stdout io.Writer, conn *config.Connection, opts getOpts) error {
	if opts.JSON {
		data, err := json.Marshal(conn)
		if err != nil {
			return fmt.Errorf("marshal connection: %w", err)
		}
		_, err = stdout.Write(append(data, '\n'))
		return err
	}

	type kv struct{ k, v string }
	var pairs []kv
	for _, name := range getBareFields {
		resolver := getFieldResolvers[name]
		v := resolver(cfg, conn)
		if v == "" {
			continue
		}
		pairs = append(pairs, kv{name, v})
	}

	sort.Slice(pairs, func(i, j int) bool { return pairs[i].k < pairs[j].k })

	var b strings.Builder
	for _, p := range pairs {
		b.WriteString(p.k)
		b.WriteByte(' ')
		b.WriteString(p.v)
		b.WriteByte('\n')
	}
	_, err := io.WriteString(stdout, b.String())
	return err
}

// runGetPlain handles single and bulk-comma plain text output.
func runGetPlain(cfg *config.Config, stdout io.Writer, conn *config.Connection, fields []string, opts getOpts) error {
	values := make([]string, len(fields))
	for i, f := range fields {
		values[i] = resolveField(cfg, conn, f, opts.Default)
	}

	var out string
	if len(values) == 1 {
		// Single field: write the value verbatim (may contain newlines for
		// tags/options) and add a trailing newline unless suppressed.
		out = values[0]
	} else {
		// Bulk comma: tab-separated single line.
		out = strings.Join(values, "\t")
	}

	if !opts.NoNewline {
		out += "\n"
	}

	_, err := io.WriteString(stdout, out)
	return err
}

// runGetJSON emits a JSON object keyed by the requested field names.
func runGetJSON(cfg *config.Config, stdout io.Writer, conn *config.Connection, fields []string, opts getOpts) error {
	// Use json.RawMessage entries in an ordered slice so we can preserve the
	// requested field order in the output. encoding/json writes map[string]
	// keys in sorted order, which would lose user-supplied order.
	type entry struct {
		key string
		raw json.RawMessage
	}

	entries := make([]entry, 0, len(fields))
	for _, f := range fields {
		var val any
		if strings.HasPrefix(f, "options.") {
			key := strings.TrimPrefix(f, "options.")
			v := ""
			if conn.Options != nil {
				v = conn.Options[key]
			}
			if v == "" && opts.Default != "" {
				v = opts.Default
			}
			val = v
		} else {
			resolver := getJSONValueResolvers[f]
			v := resolver(cfg, conn)
			if isEmptyJSONValue(v) && opts.Default != "" {
				val = opts.Default
			} else {
				val = v
			}
		}

		raw, err := json.Marshal(val)
		if err != nil {
			return fmt.Errorf("marshal %s: %w", f, err)
		}
		entries = append(entries, entry{key: f, raw: raw})
	}

	var b strings.Builder
	b.WriteByte('{')
	for i, e := range entries {
		if i > 0 {
			b.WriteByte(',')
		}
		keyJSON, _ := json.Marshal(e.key)
		b.Write(keyJSON)
		b.WriteByte(':')
		b.Write(e.raw)
	}
	b.WriteByte('}')

	if !opts.NoNewline {
		b.WriteByte('\n')
	}

	_, err := io.WriteString(stdout, b.String())
	return err
}

// resolveField returns the string value of a single field, applying the
// `--default` fallback when the value is empty. options.<key> is handled
// here as a dynamic lookup outside the static field map.
func resolveField(cfg *config.Config, conn *config.Connection, field, def string) string {
	if strings.HasPrefix(field, "options.") {
		key := strings.TrimPrefix(field, "options.")
		v := ""
		if conn.Options != nil {
			v = conn.Options[key]
		}
		if v == "" && def != "" {
			return def
		}
		return v
	}

	resolver := getFieldResolvers[field]
	v := resolver(cfg, conn)
	if v == "" && def != "" {
		return def
	}
	return v
}

// isValidField returns true for any field name accepted by `hop get`,
// including the dynamic options.<key> form.
func isValidField(name string) bool {
	if name == "" {
		return false
	}
	if strings.HasPrefix(name, "options.") {
		return strings.TrimPrefix(name, "options.") != ""
	}
	_, ok := getFieldResolvers[name]
	return ok
}

// isEmptyJSONValue reports whether a JSON-typed value should be considered
// empty for the purpose of applying the --default fallback. Numbers and
// booleans are never considered empty.
func isEmptyJSONValue(v any) bool {
	switch x := v.(type) {
	case string:
		return x == ""
	case []string:
		return len(x) == 0
	case map[string]string:
		return len(x) == 0
	default:
		return false
	}
}

// reportError writes the error message to stderr (with trailing newline) and
// returns the same error so cobra's error path can also surface it.
func reportError(stderr io.Writer, err error) error {
	if err != nil && stderr != nil {
		fmt.Fprintln(stderr, err.Error())
	}
	return err
}

// unknownIDError formats the "not found" error including up to three fuzzy
// suggestions when any are available.
func unknownIDError(id string, cfg *config.Config) error {
	matches := fuzzy.FindMatches(id, cfg.Connections)
	if len(matches) == 0 {
		return fmt.Errorf("not found: %s", id)
	}
	limit := len(matches)
	if limit > 3 {
		limit = 3
	}
	suggestions := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		suggestions = append(suggestions, matches[i].Connection.ID)
	}
	return fmt.Errorf("not found: %s (did you mean: %s)", id, strings.Join(suggestions, ", "))
}

// unknownFieldError formats the unknown-field error and includes the full
// list of valid field names so callers can correct their invocation.
func unknownFieldError(name string) error {
	return fmt.Errorf("unknown field: %s (valid fields: %s, or options.<key>)",
		name, strings.Join(getFieldNames, ", "))
}
