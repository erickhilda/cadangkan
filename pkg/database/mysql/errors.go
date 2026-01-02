// Package mysql provides MySQL database client functionality.
package mysql

import (
	"errors"
	"fmt"
)

// Common sentinel errors for the MySQL client.
var (
	// ErrNotConnected indicates the client is not connected to the database.
	ErrNotConnected = errors.New("mysql: not connected to database")

	// ErrAlreadyConnected indicates the client is already connected.
	ErrAlreadyConnected = errors.New("mysql: already connected to database")

	// ErrInvalidConfig indicates the configuration is invalid.
	ErrInvalidConfig = errors.New("mysql: invalid configuration")

	// ErrEmptyResult indicates the query returned no results.
	ErrEmptyResult = errors.New("mysql: query returned no results")
)

// ConnectionError represents a database connection error.
type ConnectionError struct {
	Host    string
	Port    int
	Message string
	Err     error
}

// Error returns the error message.
func (e *ConnectionError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("mysql connection error to %s:%d: %s: %v", e.Host, e.Port, e.Message, e.Err)
	}
	return fmt.Sprintf("mysql connection error to %s:%d: %s", e.Host, e.Port, e.Message)
}

// Unwrap returns the underlying error.
func (e *ConnectionError) Unwrap() error {
	return e.Err
}

// QueryError represents a database query error.
type QueryError struct {
	Query   string
	Message string
	Err     error
}

// Error returns the error message.
func (e *QueryError) Error() string {
	// Truncate query if too long for error message
	query := e.Query
	if len(query) > 100 {
		query = query[:100] + "..."
	}
	if e.Err != nil {
		return fmt.Sprintf("mysql query error [%s]: %s: %v", query, e.Message, e.Err)
	}
	return fmt.Sprintf("mysql query error [%s]: %s", query, e.Message)
}

// Unwrap returns the underlying error.
func (e *QueryError) Unwrap() error {
	return e.Err
}

// TimeoutError represents a database operation timeout.
type TimeoutError struct {
	Operation string
	Duration  string
	Err       error
}

// Error returns the error message.
func (e *TimeoutError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("mysql timeout during %s after %s: %v", e.Operation, e.Duration, e.Err)
	}
	return fmt.Sprintf("mysql timeout during %s after %s", e.Operation, e.Duration)
}

// Unwrap returns the underlying error.
func (e *TimeoutError) Unwrap() error {
	return e.Err
}

// ConfigError represents a configuration error.
type ConfigError struct {
	Field   string
	Message string
}

// Error returns the error message.
func (e *ConfigError) Error() string {
	return fmt.Sprintf("mysql config error: %s: %s", e.Field, e.Message)
}

// IsConnectionError checks if the error is a ConnectionError.
func IsConnectionError(err error) bool {
	var connErr *ConnectionError
	return errors.As(err, &connErr)
}

// IsQueryError checks if the error is a QueryError.
func IsQueryError(err error) bool {
	var queryErr *QueryError
	return errors.As(err, &queryErr)
}

// IsTimeoutError checks if the error is a TimeoutError.
func IsTimeoutError(err error) bool {
	var timeoutErr *TimeoutError
	return errors.As(err, &timeoutErr)
}

// IsConfigError checks if the error is a ConfigError.
func IsConfigError(err error) bool {
	var configErr *ConfigError
	return errors.As(err, &configErr)
}

// WrapConnectionError wraps an error as a ConnectionError.
func WrapConnectionError(host string, port int, message string, err error) error {
	return &ConnectionError{
		Host:    host,
		Port:    port,
		Message: message,
		Err:     err,
	}
}

// WrapQueryError wraps an error as a QueryError.
func WrapQueryError(query, message string, err error) error {
	return &QueryError{
		Query:   query,
		Message: message,
		Err:     err,
	}
}

// WrapTimeoutError wraps an error as a TimeoutError.
func WrapTimeoutError(operation, duration string, err error) error {
	return &TimeoutError{
		Operation: operation,
		Duration:  duration,
		Err:       err,
	}
}
