package config

import (
	"fmt"
	"strings"
)

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	var msgs []string
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "\n")
}

// CheckSafety reports whether any field that becomes a standalone token in the
// ssh/mosh argv could be misinterpreted as a command-line option. A value
// beginning with "-" (e.g. "-oProxyCommand=touch /tmp/pwned") would be parsed by
// ssh as an option and execute an arbitrary command on the local machine
// (CWE-88). It is enforced on config load, at import time, and again in
// ssh.Connect immediately before a connection is launched.
func (c *Connection) CheckSafety() error {
	fields := []struct {
		name, value string
	}{
		{"host", c.Host},
		{"user", c.User},
		{"proxy_jump", c.ProxyJump},
	}
	for _, f := range fields {
		if strings.HasPrefix(f.value, "-") {
			return ValidationError{
				Field:   f.name,
				Message: fmt.Sprintf("must not start with '-' (value %q could be interpreted as an ssh option)", f.value),
			}
		}
	}
	return nil
}

func (c *Config) Validate() error {
	var errs ValidationErrors

	if c.Version != 1 {
		errs = append(errs, ValidationError{
			Field:   "version",
			Message: "must be 1",
		})
	}

	seenIDs := make(map[string]bool)
	for i, conn := range c.Connections {
		prefix := fmt.Sprintf("connections[%d]", i)

		if conn.ID == "" {
			errs = append(errs, ValidationError{
				Field:   prefix + ".id",
				Message: "is required",
			})
		} else if seenIDs[conn.ID] {
			errs = append(errs, ValidationError{
				Field:   prefix + ".id",
				Message: fmt.Sprintf("duplicate id '%s'", conn.ID),
			})
		} else {
			seenIDs[conn.ID] = true
		}

		if conn.Host == "" {
			errs = append(errs, ValidationError{
				Field:   prefix + ".host",
				Message: "is required",
			})
		}

		if err := conn.CheckSafety(); err != nil {
			if ve, ok := err.(ValidationError); ok {
				ve.Field = prefix + "." + ve.Field
				errs = append(errs, ve)
			} else {
				errs = append(errs, ValidationError{Field: prefix, Message: err.Error()})
			}
		}
	}

	for name, members := range c.Groups {
		for _, member := range members {
			if !seenIDs[member] {
				errs = append(errs, ValidationError{
					Field:   fmt.Sprintf("groups.%s", name),
					Message: fmt.Sprintf("references unknown connection '%s'", member),
				})
			}
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}
