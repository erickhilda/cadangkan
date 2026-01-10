package mysql

import "database/sql"

// DatabaseClient defines the interface for MySQL database operations.
// This interface enables mocking for unit tests.
type DatabaseClient interface {
	// Connection management
	Connect() error
	Ping() error
	Close() error
	IsConnected() bool

	// Query execution
	ExecuteQuery(query string) (*sql.Rows, error)
	ExecuteQueryArgs(query string, args ...interface{}) (*sql.Rows, error)
	Execute(query string, args ...interface{}) (sql.Result, error)

	// Introspection methods
	GetVersion() (string, error)
	GetDatabases() ([]string, error)
	GetTables(database string) ([]string, error)
	GetTableSize(database, table string) (int64, error)
	GetTableRowCount(database, table string) (int64, error)
	GetDatabaseSize(database string) (int64, error)
	GetTableInfo(database, table string) (*TableInfo, error)
	GetDatabaseInfo(database string) (*DatabaseInfo, error)
	CreateDatabase(database string) error
	DatabaseExists(database string) (bool, error)
}

// Ensure Client implements DatabaseClient interface.
var _ DatabaseClient = (*Client)(nil)
