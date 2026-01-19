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
