package config

import "strings"

// Validate validates the entire config.
func (c *Config) Validate() error {
	if c.Version == "" {
		return &ValidationError{Field: "version", Message: "version is required"}
	}

	// Validate each database config
	for name, db := range c.Databases {
		db.Name = name // Ensure name is set
		if err := db.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate validates a database configuration.
func (d *DatabaseConfig) Validate() error {
	if d.Type == "" {
		return &ValidationError{Field: "type", Message: "database type is required"}
	}

	if d.Type != "mysql" {
		return &ValidationError{Field: "type", Message: "only 'mysql' type is supported"}
	}

	if d.Host == "" {
		return &ValidationError{Field: "host", Message: "host is required"}
	}

	if d.Port <= 0 || d.Port > 65535 {
		return &ValidationError{Field: "port", Message: "port must be between 1 and 65535"}
	}

	if d.User == "" {
		return &ValidationError{Field: "user", Message: "user is required"}
	}

	if d.Database == "" {
		return &ValidationError{Field: "database", Message: "database name is required"}
	}

	return nil
}

// SanitizeName sanitizes a database name for use as a config key.
func SanitizeName(name string) string {
	// Remove spaces and convert to lowercase
	name = strings.TrimSpace(name)
	name = strings.ToLower(name)

	// Replace invalid characters with underscores
	replacer := strings.NewReplacer(
		" ", "_",
		"-", "_",
		".", "_",
	)

	return replacer.Replace(name)
}
