package config

import "fmt"

// ConfigNotFoundError is returned when the config file doesn't exist.
type ConfigNotFoundError struct {
	Path string
}

func (e *ConfigNotFoundError) Error() string {
	return fmt.Sprintf("config file not found: %s", e.Path)
}

// DatabaseNotFoundError is returned when a database is not found in the config.
type DatabaseNotFoundError struct {
	Name string
}

func (e *DatabaseNotFoundError) Error() string {
	return fmt.Sprintf("database '%s' not found in config", e.Name)
}

// ValidationError is returned when config validation fails.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error [%s]: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// EncryptionError is returned when encryption/decryption fails.
type EncryptionError struct {
	Operation string
	Err       error
}

func (e *EncryptionError) Error() string {
	return fmt.Sprintf("encryption error during %s: %v", e.Operation, e.Err)
}

func (e *EncryptionError) Unwrap() error {
	return e.Err
}
