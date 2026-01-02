// Package mysql provides a MySQL database client for Cadangkan.
//
// This package offers a simple, unified interface for connecting to MySQL databases,
// executing queries, and retrieving database metadata. It is designed to support
// the backup and restore operations of the Cadangkan tool.
//
// # Features
//
//   - Connection management with configurable pooling
//   - Query execution for SELECT and non-SELECT statements
//   - Database introspection (list databases, tables, sizes)
//   - Thread-safe concurrent access
//   - Comprehensive error handling with custom error types
//   - Mock client for testing
//
// # Quick Start
//
// Create a new client and connect to a database:
//
//	config := mysql.NewConfig().
//		WithHost("localhost").
//		WithPort(3306).
//		WithUser("root").
//		WithPassword("password").
//		WithDatabase("mydb")
//
//	client, err := mysql.NewClient(config)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer client.Close()
//
//	if err := client.Connect(); err != nil {
//		log.Fatal(err)
//	}
//
//	// Get MySQL version
//	version, err := client.GetVersion()
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println("MySQL version:", version)
//
// # Configuration
//
// The Config struct provides comprehensive options for configuring the MySQL connection:
//
//   - Host: Database server hostname or IP
//   - Port: Database server port (default: 3306)
//   - User: Database username
//   - Password: Database password
//   - Database: Database name to connect to
//   - Timeout: Connection timeout (default: 10s)
//   - MaxOpenConns: Maximum open connections (default: 25)
//   - MaxIdleConns: Maximum idle connections (default: 10)
//   - ConnMaxLifetime: Maximum connection lifetime (default: 5m)
//   - ConnMaxIdleTime: Maximum idle time (default: 30s)
//
// # Database Introspection
//
// The client provides methods to inspect database structure:
//
//	// List all databases
//	databases, err := client.GetDatabases()
//
//	// List tables in a database
//	tables, err := client.GetTables("mydb")
//
//	// Get table size
//	size, err := client.GetTableSize("mydb", "users")
//
//	// Get row count
//	count, err := client.GetTableRowCount("mydb", "users")
//
//	// Get complete database info
//	info, err := client.GetDatabaseInfo("mydb")
//	fmt.Printf("Database %s has %d tables, total size: %d bytes\n",
//		info.Name, info.TableCount, info.TotalSize)
//
// # Error Handling
//
// The package provides custom error types for better error handling:
//
//   - ConnectionError: Database connection failures
//   - QueryError: Query execution failures
//   - TimeoutError: Operation timeouts
//   - ConfigError: Configuration validation errors
//
// Use the helper functions to check error types:
//
//	err := client.Connect()
//	if mysql.IsConnectionError(err) {
//		// Handle connection error
//	}
//
// # Testing
//
// For testing, use the MockClient which implements the DatabaseClient interface:
//
//	mock := mysql.NewMockClient()
//	mock.SetConnected(true)
//	mock.Version = "8.0.35"
//	mock.Databases = []string{"db1", "db2"}
//	mock.SetTables("db1", []string{"users", "orders"})
//
//	// Use mock in tests
//	version, err := mock.GetVersion()
//	// version == "8.0.35"
//
// The mock also tracks method calls for verification:
//
//	mock.GetVersion()
//	mock.GetDatabases()
//	assert.Equal(t, 1, mock.GetCallCount("GetVersion"))
//	assert.Equal(t, 1, mock.GetCallCount("GetDatabases"))
//
// # Thread Safety
//
// The Client is thread-safe and can be used concurrently from multiple goroutines.
// The underlying sql.DB connection pool handles concurrent access automatically.
package mysql
