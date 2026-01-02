package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
)

// Client represents a MySQL database client.
type Client struct {
	config    *Config
	db        *sql.DB
	connected bool
	mu        sync.RWMutex
}

// NewClient creates a new MySQL client with the given configuration.
// It does not establish a connection; call Connect() to connect.
func NewClient(config *Config) (*Client, error) {
	if config == nil {
		return nil, ErrInvalidConfig
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &Client{
		config:    config,
		connected: false,
	}, nil
}

// NewClientWithDB creates a new MySQL client with an existing database connection.
// This is primarily used for testing with mocked connections.
func NewClientWithDB(config *Config, db *sql.DB) (*Client, error) {
	if config == nil {
		return nil, ErrInvalidConfig
	}

	return &Client{
		config:    config,
		db:        db,
		connected: db != nil,
	}, nil
}

// Connect establishes a connection to the MySQL database.
func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return ErrAlreadyConnected
	}

	db, err := sql.Open("mysql", c.config.DSN())
	if err != nil {
		return WrapConnectionError(c.config.Host, c.config.Port, "failed to open connection", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(c.config.MaxOpenConns)
	db.SetMaxIdleConns(c.config.MaxIdleConns)
	db.SetConnMaxLifetime(c.config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(c.config.ConnMaxIdleTime)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return WrapConnectionError(c.config.Host, c.config.Port, "failed to ping database", err)
	}

	c.db = db
	c.connected = true
	return nil
}

// Ping checks if the database connection is still alive.
func (c *Client) Ping() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected || c.db == nil {
		return ErrNotConnected
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
	defer cancel()

	if err := c.db.PingContext(ctx); err != nil {
		return WrapConnectionError(c.config.Host, c.config.Port, "ping failed", err)
	}

	return nil
}

// Close closes the database connection.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected || c.db == nil {
		return nil // Not an error to close an already closed connection
	}

	err := c.db.Close()
	c.db = nil
	c.connected = false

	if err != nil {
		return WrapConnectionError(c.config.Host, c.config.Port, "failed to close connection", err)
	}

	return nil
}

// IsConnected returns true if the client is connected to the database.
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// DB returns the underlying sql.DB instance.
// This should be used with caution and is primarily for advanced use cases.
func (c *Client) DB() *sql.DB {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.db
}

// Config returns a copy of the client configuration.
func (c *Client) Config() Config {
	return *c.config
}

// ExecuteQuery executes a SELECT query and returns the rows.
// The caller is responsible for closing the returned rows.
func (c *Client) ExecuteQuery(query string) (*sql.Rows, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected || c.db == nil {
		return nil, ErrNotConnected
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
	defer cancel()

	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		return nil, WrapQueryError(query, "query execution failed", err)
	}

	return rows, nil
}

// ExecuteQueryArgs executes a SELECT query with arguments and returns the rows.
// The caller is responsible for closing the returned rows.
func (c *Client) ExecuteQueryArgs(query string, args ...interface{}) (*sql.Rows, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected || c.db == nil {
		return nil, ErrNotConnected
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
	defer cancel()

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, WrapQueryError(query, "query execution failed", err)
	}

	return rows, nil
}

// Execute executes a non-SELECT query (INSERT, UPDATE, DELETE, etc.) and returns the result.
func (c *Client) Execute(query string, args ...interface{}) (sql.Result, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected || c.db == nil {
		return nil, ErrNotConnected
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
	defer cancel()

	result, err := c.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, WrapQueryError(query, "execution failed", err)
	}

	return result, nil
}

// GetVersion returns the MySQL server version.
func (c *Client) GetVersion() (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected || c.db == nil {
		return "", ErrNotConnected
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
	defer cancel()

	var version string
	err := c.db.QueryRowContext(ctx, "SELECT VERSION()").Scan(&version)
	if err != nil {
		return "", WrapQueryError("SELECT VERSION()", "failed to get version", err)
	}

	return version, nil
}

// GetDatabases returns a list of all databases on the server.
func (c *Client) GetDatabases() ([]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected || c.db == nil {
		return nil, ErrNotConnected
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
	defer cancel()

	rows, err := c.db.QueryContext(ctx, "SHOW DATABASES")
	if err != nil {
		return nil, WrapQueryError("SHOW DATABASES", "failed to list databases", err)
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var db string
		if err := rows.Scan(&db); err != nil {
			return nil, WrapQueryError("SHOW DATABASES", "failed to scan database name", err)
		}
		databases = append(databases, db)
	}

	if err := rows.Err(); err != nil {
		return nil, WrapQueryError("SHOW DATABASES", "error iterating rows", err)
	}

	return databases, nil
}

// GetTables returns a list of all tables in the specified database.
func (c *Client) GetTables(database string) ([]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected || c.db == nil {
		return nil, ErrNotConnected
	}

	if database == "" {
		return nil, &ConfigError{Field: "database", Message: "database name is required"}
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
	defer cancel()

	query := fmt.Sprintf("SHOW TABLES FROM `%s`", database)
	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		return nil, WrapQueryError(query, "failed to list tables", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, WrapQueryError(query, "failed to scan table name", err)
		}
		tables = append(tables, table)
	}

	if err := rows.Err(); err != nil {
		return nil, WrapQueryError(query, "error iterating rows", err)
	}

	return tables, nil
}

// GetTableSize returns the size of the specified table in bytes.
func (c *Client) GetTableSize(database, table string) (int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected || c.db == nil {
		return 0, ErrNotConnected
	}

	if database == "" {
		return 0, &ConfigError{Field: "database", Message: "database name is required"}
	}
	if table == "" {
		return 0, &ConfigError{Field: "table", Message: "table name is required"}
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
	defer cancel()

	query := `
		SELECT COALESCE(data_length + index_length, 0) AS size
		FROM information_schema.TABLES
		WHERE table_schema = ? AND table_name = ?
	`

	var size int64
	err := c.db.QueryRowContext(ctx, query, database, table).Scan(&size)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrEmptyResult
		}
		return 0, WrapQueryError(query, "failed to get table size", err)
	}

	return size, nil
}

// GetTableRowCount returns the approximate row count for the specified table.
// Note: This uses information_schema which may not be exact for InnoDB tables.
func (c *Client) GetTableRowCount(database, table string) (int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected || c.db == nil {
		return 0, ErrNotConnected
	}

	if database == "" {
		return 0, &ConfigError{Field: "database", Message: "database name is required"}
	}
	if table == "" {
		return 0, &ConfigError{Field: "table", Message: "table name is required"}
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
	defer cancel()

	query := `
		SELECT COALESCE(table_rows, 0) AS row_count
		FROM information_schema.TABLES
		WHERE table_schema = ? AND table_name = ?
	`

	var rowCount int64
	err := c.db.QueryRowContext(ctx, query, database, table).Scan(&rowCount)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrEmptyResult
		}
		return 0, WrapQueryError(query, "failed to get row count", err)
	}

	return rowCount, nil
}

// GetDatabaseSize returns the total size of the specified database in bytes.
func (c *Client) GetDatabaseSize(database string) (int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected || c.db == nil {
		return 0, ErrNotConnected
	}

	if database == "" {
		return 0, &ConfigError{Field: "database", Message: "database name is required"}
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
	defer cancel()

	query := `
		SELECT COALESCE(SUM(data_length + index_length), 0) AS size
		FROM information_schema.TABLES
		WHERE table_schema = ?
	`

	var size int64
	err := c.db.QueryRowContext(ctx, query, database).Scan(&size)
	if err != nil {
		return 0, WrapQueryError(query, "failed to get database size", err)
	}

	return size, nil
}

// GetTableInfo returns detailed information about a table.
type TableInfo struct {
	Name      string
	Engine    string
	RowCount  int64
	DataSize  int64
	IndexSize int64
	TotalSize int64
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

// GetTableInfo returns detailed information about the specified table.
func (c *Client) GetTableInfo(database, table string) (*TableInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected || c.db == nil {
		return nil, ErrNotConnected
	}

	if database == "" {
		return nil, &ConfigError{Field: "database", Message: "database name is required"}
	}
	if table == "" {
		return nil, &ConfigError{Field: "table", Message: "table name is required"}
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
	defer cancel()

	query := `
		SELECT 
			table_name,
			COALESCE(engine, '') AS engine,
			COALESCE(table_rows, 0) AS row_count,
			COALESCE(data_length, 0) AS data_size,
			COALESCE(index_length, 0) AS index_size,
			COALESCE(data_length + index_length, 0) AS total_size,
			create_time,
			update_time
		FROM information_schema.TABLES
		WHERE table_schema = ? AND table_name = ?
	`

	info := &TableInfo{}
	var createdAt, updatedAt sql.NullTime

	err := c.db.QueryRowContext(ctx, query, database, table).Scan(
		&info.Name,
		&info.Engine,
		&info.RowCount,
		&info.DataSize,
		&info.IndexSize,
		&info.TotalSize,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrEmptyResult
		}
		return nil, WrapQueryError(query, "failed to get table info", err)
	}

	if createdAt.Valid {
		info.CreatedAt = &createdAt.Time
	}
	if updatedAt.Valid {
		info.UpdatedAt = &updatedAt.Time
	}

	return info, nil
}

// DatabaseInfo holds database-level statistics.
type DatabaseInfo struct {
	Name       string
	TableCount int
	TotalSize  int64
	Tables     []TableInfo
}

// GetDatabaseInfo returns detailed information about the specified database.
func (c *Client) GetDatabaseInfo(database string) (*DatabaseInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected || c.db == nil {
		return nil, ErrNotConnected
	}

	if database == "" {
		return nil, &ConfigError{Field: "database", Message: "database name is required"}
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
	defer cancel()

	query := `
		SELECT 
			table_name,
			COALESCE(engine, '') AS engine,
			COALESCE(table_rows, 0) AS row_count,
			COALESCE(data_length, 0) AS data_size,
			COALESCE(index_length, 0) AS index_size,
			COALESCE(data_length + index_length, 0) AS total_size,
			create_time,
			update_time
		FROM information_schema.TABLES
		WHERE table_schema = ?
		ORDER BY table_name
	`

	rows, err := c.db.QueryContext(ctx, query, database)
	if err != nil {
		return nil, WrapQueryError(query, "failed to get database info", err)
	}
	defer rows.Close()

	info := &DatabaseInfo{
		Name:   database,
		Tables: []TableInfo{},
	}

	for rows.Next() {
		var tableInfo TableInfo
		var createdAt, updatedAt sql.NullTime

		err := rows.Scan(
			&tableInfo.Name,
			&tableInfo.Engine,
			&tableInfo.RowCount,
			&tableInfo.DataSize,
			&tableInfo.IndexSize,
			&tableInfo.TotalSize,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, WrapQueryError(query, "failed to scan table info", err)
		}

		if createdAt.Valid {
			tableInfo.CreatedAt = &createdAt.Time
		}
		if updatedAt.Valid {
			tableInfo.UpdatedAt = &updatedAt.Time
		}

		info.Tables = append(info.Tables, tableInfo)
		info.TotalSize += tableInfo.TotalSize
	}

	if err := rows.Err(); err != nil {
		return nil, WrapQueryError(query, "error iterating rows", err)
	}

	info.TableCount = len(info.Tables)

	return info, nil
}
