package hopmcp

// Output limits
const (
	MaxBytesPerHost    = 64 * 1024 // 64KB per host output
	MaxHostsPerRequest = 50
)

// ListConnectionsInput filters connections by project, env, or tag.
type ListConnectionsInput struct {
	Project string `json:"project,omitempty" jsonschema:"Filter by project name"`
	Env     string `json:"env,omitempty" jsonschema:"Filter by environment (e.g. prod, staging, dev)"`
	Tag     string `json:"tag,omitempty" jsonschema:"Filter by tag"`
}

// SearchConnectionsInput searches connections by fuzzy query.
type SearchConnectionsInput struct {
	Query string `json:"query" jsonschema:"Search query (fuzzy matched against IDs, hosts, tags, projects)"`
}

// GetConnectionInput retrieves a single connection by ID.
type GetConnectionInput struct {
	ID string `json:"id" jsonschema:"Connection ID"`
}

// ExecCommandInput executes a command on matched connections.
type ExecCommandInput struct {
	Target   string `json:"target" jsonschema:"Target pattern (group name, project-env, glob, or fuzzy match)"`
	Command  string `json:"command" jsonschema:"Shell command to execute on remote hosts"`
	Tag      string `json:"tag,omitempty" jsonschema:"Filter matched connections by tag"`
	Parallel int    `json:"parallel,omitempty" jsonschema:"Max concurrent connections (default: 10)"`
	Timeout  string `json:"timeout,omitempty" jsonschema:"Command timeout (e.g. 30s, 5m)"`
}

// ResolveTargetInput resolves a target to connections.
type ResolveTargetInput struct {
	Target string `json:"target" jsonschema:"Target pattern to resolve (group name, project-env, glob, or fuzzy match)"`
	Tag    string `json:"tag,omitempty" jsonschema:"Filter resolved connections by tag"`
}

// ListGroupsInput lists all defined groups (no params needed).
type ListGroupsInput struct{}

// GetHistoryInput retrieves connection usage history.
type GetHistoryInput struct {
	Limit  int    `json:"limit,omitempty" jsonschema:"Maximum entries to return (default: 10)"`
	SortBy string `json:"sort_by,omitempty" jsonschema:"Sort order: recent (default) or frequent"`
}

// BuildSSHCommandInput builds an SSH command string.
type BuildSSHCommandInput struct {
	ID       string `json:"id" jsonschema:"Connection ID"`
	Command  string `json:"command,omitempty" jsonschema:"Remote command to include in the SSH command"`
	ForceTTY bool   `json:"force_tty,omitempty" jsonschema:"Force TTY allocation (-t flag)"`
}

// ConnectionInfo is the output representation of a connection.
// IdentityFile is intentionally omitted for security.
type ConnectionInfo struct {
	ID           string            `json:"id"`
	Host         string            `json:"host"`
	User         string            `json:"user,omitempty"`
	Port         int               `json:"port"`
	Project      string            `json:"project,omitempty"`
	Env          string            `json:"env,omitempty"`
	ProxyJump    string            `json:"proxy_jump,omitempty"`
	ForwardAgent bool              `json:"forward_agent,omitempty"`
	Tags         []string          `json:"tags,omitempty"`
	Options      map[string]string `json:"options,omitempty"`
}
