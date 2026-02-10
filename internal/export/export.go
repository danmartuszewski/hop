package export

import (
	"io"

	"github.com/danmartuszewski/hop/internal/config"
	"gopkg.in/yaml.v3"
)

// BuildExportConfig wraps connections in a minimal Config for export.
// Groups and Defaults are omitted since the exported subset is self-contained.
func BuildExportConfig(connections []config.Connection) *config.Config {
	return &config.Config{
		Version:     1,
		Connections: connections,
	}
}

// WriteYAML marshals the config to YAML and writes it to the given writer.
func WriteYAML(w io.Writer, cfg *config.Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}
