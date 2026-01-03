package mysql

import (
	"fmt"
	"time"
)

// Default configuration values.
const (
	DefaultPort           = 3306
	DefaultTimeout        = 10 * time.Second
	DefaultMaxOpenConns   = 25
	DefaultMaxIdleConns   = 10
	DefaultConnMaxLife    = 5 * time.Minute
	DefaultConnMaxIdle    = 30 * time.Second
)

// Config holds the MySQL connection configuration.
type Config struct {
	// Host is the database server hostname or IP address.
	Host string

	// Port is the database server port (default: 3306).
	Port int

	// User is the database username.
	User string

	// Password is the database password.
	Password string

	// Database is the name of the database to connect to.
	Database string

	// Timeout is the connection timeout duration (default: 10s).
	Timeout time.Duration

	// MaxOpenConns sets the maximum number of open connections (default: 25).
	MaxOpenConns int

	// MaxIdleConns sets the maximum number of idle connections (default: 10).
	MaxIdleConns int

	// ConnMaxLifetime sets the maximum lifetime of a connection (default: 5m).
	ConnMaxLifetime time.Duration

	// ConnMaxIdleTime sets the maximum idle time of a connection (default: 30s).
	ConnMaxIdleTime time.Duration

	// ParseTime enables parsing of DATE and DATETIME values to time.Time.
	ParseTime bool

	// TLS specifies the TLS configuration name (e.g., "true", "false", "skip-verify", or custom).
	TLS string
}

// NewConfig creates a new Config with default values.
func NewConfig() *Config {
	return &Config{
		Port:            DefaultPort,
		Timeout:         DefaultTimeout,
		MaxOpenConns:    DefaultMaxOpenConns,
		MaxIdleConns:    DefaultMaxIdleConns,
		ConnMaxLifetime: DefaultConnMaxLife,
		ConnMaxIdleTime: DefaultConnMaxIdle,
		ParseTime:       true,
	}
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.Host == "" {
		return &ConfigError{Field: "Host", Message: "host is required"}
	}
	if c.User == "" {
		return &ConfigError{Field: "User", Message: "user is required"}
	}
	if c.Port <= 0 || c.Port > 65535 {
		return &ConfigError{Field: "Port", Message: "port must be between 1 and 65535"}
	}
	if c.Timeout < 0 {
		return &ConfigError{Field: "Timeout", Message: "timeout must be non-negative"}
	}
	if c.MaxOpenConns < 0 {
		return &ConfigError{Field: "MaxOpenConns", Message: "max open connections must be non-negative"}
	}
	if c.MaxIdleConns < 0 {
		return &ConfigError{Field: "MaxIdleConns", Message: "max idle connections must be non-negative"}
	}
	return nil
}

// DSN returns the Data Source Name for MySQL connection.
// Format: user:password@tcp(host:port)/database?timeout=10s&parseTime=true
func (c *Config) DSN() string {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/", c.User, c.Password, c.Host, c.Port)

	if c.Database != "" {
		dsn += c.Database
	}

	// Build query parameters
	params := "?"
	first := true

	addParam := func(key, value string) {
		if !first {
			params += "&"
		}
		params += key + "=" + value
		first = false
	}

	if c.Timeout > 0 {
		addParam("timeout", c.Timeout.String())
	}

	if c.ParseTime {
		addParam("parseTime", "true")
	}

	if c.TLS != "" {
		addParam("tls", c.TLS)
	}

	// Add charset for proper encoding
	addParam("charset", "utf8mb4")

	// Add interpolateParams for better performance
	addParam("interpolateParams", "true")

	if !first {
		dsn += params
	}

	return dsn
}

// DSNMasked returns the DSN with the password masked for logging.
func (c *Config) DSNMasked() string {
	masked := fmt.Sprintf("%s:***@tcp(%s:%d)/", c.User, c.Host, c.Port)

	if c.Database != "" {
		masked += c.Database
	}

	return masked
}

// WithHost sets the host and returns the config for chaining.
func (c *Config) WithHost(host string) *Config {
	c.Host = host
	return c
}

// WithPort sets the port and returns the config for chaining.
func (c *Config) WithPort(port int) *Config {
	c.Port = port
	return c
}

// WithUser sets the user and returns the config for chaining.
func (c *Config) WithUser(user string) *Config {
	c.User = user
	return c
}

// WithPassword sets the password and returns the config for chaining.
func (c *Config) WithPassword(password string) *Config {
	c.Password = password
	return c
}

// WithDatabase sets the database and returns the config for chaining.
func (c *Config) WithDatabase(database string) *Config {
	c.Database = database
	return c
}

// WithTimeout sets the timeout and returns the config for chaining.
func (c *Config) WithTimeout(timeout time.Duration) *Config {
	c.Timeout = timeout
	return c
}
