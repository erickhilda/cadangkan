package mysql

import (
	"database/sql"
	"sync"
)

// MockClient is a mock implementation of DatabaseClient for testing.
type MockClient struct {
	mu sync.RWMutex

	// Connection state
	connected bool

	// Configurable responses
	ConnectErr   error
	PingErr      error
	CloseErr     error
	Version      string
	VersionErr   error
	Databases    []string
	DatabasesErr error
	Tables       map[string][]string // database -> tables
	TablesErr    error
	TableSizes   map[string]map[string]int64 // database -> table -> size
	TableSizeErr error
	RowCounts    map[string]map[string]int64 // database -> table -> count
	RowCountErr  error
	DBSizes      map[string]int64 // database -> size
	DBSizeErr    error
	TableInfos   map[string]map[string]*TableInfo // database -> table -> info
	TableInfoErr error
	DBInfos      map[string]*DatabaseInfo // database -> info
	DBInfoErr    error

	// Query responses
	QueryRows   *sql.Rows
	QueryErr    error
	ExecResult  sql.Result
	ExecErr     error

	// Call tracking
	Calls []MockCall
}

// MockCall records a method call for verification.
type MockCall struct {
	Method string
	Args   []interface{}
}

// NewMockClient creates a new mock client for testing.
func NewMockClient() *MockClient {
	return &MockClient{
		Tables:     make(map[string][]string),
		TableSizes: make(map[string]map[string]int64),
		RowCounts:  make(map[string]map[string]int64),
		DBSizes:    make(map[string]int64),
		TableInfos: make(map[string]map[string]*TableInfo),
		DBInfos:    make(map[string]*DatabaseInfo),
		Calls:      []MockCall{},
	}
}

// recordCall records a method call for verification.
func (m *MockClient) recordCall(method string, args ...interface{}) {
	m.Calls = append(m.Calls, MockCall{Method: method, Args: args})
}

// GetCalls returns all recorded calls.
func (m *MockClient) GetCalls() []MockCall {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Calls
}

// GetCallCount returns the number of times a method was called.
func (m *MockClient) GetCallCount(method string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, call := range m.Calls {
		if call.Method == method {
			count++
		}
	}
	return count
}

// ResetCalls clears all recorded calls.
func (m *MockClient) ResetCalls() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = []MockCall{}
}

// Connect simulates connecting to the database.
func (m *MockClient) Connect() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.recordCall("Connect")

	if m.ConnectErr != nil {
		return m.ConnectErr
	}

	m.connected = true
	return nil
}

// Ping simulates pinging the database.
func (m *MockClient) Ping() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.recordCall("Ping")

	if !m.connected {
		return ErrNotConnected
	}

	return m.PingErr
}

// Close simulates closing the database connection.
func (m *MockClient) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.recordCall("Close")

	if m.CloseErr != nil {
		return m.CloseErr
	}

	m.connected = false
	return nil
}

// IsConnected returns the mock connection state.
func (m *MockClient) IsConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.recordCall("IsConnected")
	return m.connected
}

// ExecuteQuery simulates executing a query.
func (m *MockClient) ExecuteQuery(query string) (*sql.Rows, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.recordCall("ExecuteQuery", query)

	if !m.connected {
		return nil, ErrNotConnected
	}

	if m.QueryErr != nil {
		return nil, m.QueryErr
	}

	return m.QueryRows, nil
}

// ExecuteQueryArgs simulates executing a query with arguments.
func (m *MockClient) ExecuteQueryArgs(query string, args ...interface{}) (*sql.Rows, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	callArgs := append([]interface{}{query}, args...)
	m.recordCall("ExecuteQueryArgs", callArgs...)

	if !m.connected {
		return nil, ErrNotConnected
	}

	if m.QueryErr != nil {
		return nil, m.QueryErr
	}

	return m.QueryRows, nil
}

// Execute simulates executing a non-SELECT query.
func (m *MockClient) Execute(query string, args ...interface{}) (sql.Result, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	callArgs := append([]interface{}{query}, args...)
	m.recordCall("Execute", callArgs...)

	if !m.connected {
		return nil, ErrNotConnected
	}

	if m.ExecErr != nil {
		return nil, m.ExecErr
	}

	return m.ExecResult, nil
}

// GetVersion returns the mock version.
func (m *MockClient) GetVersion() (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.recordCall("GetVersion")

	if !m.connected {
		return "", ErrNotConnected
	}

	if m.VersionErr != nil {
		return "", m.VersionErr
	}

	return m.Version, nil
}

// GetDatabases returns the mock database list.
func (m *MockClient) GetDatabases() ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.recordCall("GetDatabases")

	if !m.connected {
		return nil, ErrNotConnected
	}

	if m.DatabasesErr != nil {
		return nil, m.DatabasesErr
	}

	return m.Databases, nil
}

// GetTables returns the mock table list for a database.
func (m *MockClient) GetTables(database string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.recordCall("GetTables", database)

	if !m.connected {
		return nil, ErrNotConnected
	}

	if m.TablesErr != nil {
		return nil, m.TablesErr
	}

	tables, ok := m.Tables[database]
	if !ok {
		return []string{}, nil
	}

	return tables, nil
}

// GetTableSize returns the mock table size.
func (m *MockClient) GetTableSize(database, table string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.recordCall("GetTableSize", database, table)

	if !m.connected {
		return 0, ErrNotConnected
	}

	if m.TableSizeErr != nil {
		return 0, m.TableSizeErr
	}

	if dbTables, ok := m.TableSizes[database]; ok {
		if size, ok := dbTables[table]; ok {
			return size, nil
		}
	}

	return 0, ErrEmptyResult
}

// GetTableRowCount returns the mock row count.
func (m *MockClient) GetTableRowCount(database, table string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.recordCall("GetTableRowCount", database, table)

	if !m.connected {
		return 0, ErrNotConnected
	}

	if m.RowCountErr != nil {
		return 0, m.RowCountErr
	}

	if dbTables, ok := m.RowCounts[database]; ok {
		if count, ok := dbTables[table]; ok {
			return count, nil
		}
	}

	return 0, ErrEmptyResult
}

// GetDatabaseSize returns the mock database size.
func (m *MockClient) GetDatabaseSize(database string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.recordCall("GetDatabaseSize", database)

	if !m.connected {
		return 0, ErrNotConnected
	}

	if m.DBSizeErr != nil {
		return 0, m.DBSizeErr
	}

	if size, ok := m.DBSizes[database]; ok {
		return size, nil
	}

	return 0, nil
}

// GetTableInfo returns the mock table info.
func (m *MockClient) GetTableInfo(database, table string) (*TableInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.recordCall("GetTableInfo", database, table)

	if !m.connected {
		return nil, ErrNotConnected
	}

	if m.TableInfoErr != nil {
		return nil, m.TableInfoErr
	}

	if dbTables, ok := m.TableInfos[database]; ok {
		if info, ok := dbTables[table]; ok {
			return info, nil
		}
	}

	return nil, ErrEmptyResult
}

// GetDatabaseInfo returns the mock database info.
func (m *MockClient) GetDatabaseInfo(database string) (*DatabaseInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.recordCall("GetDatabaseInfo", database)

	if !m.connected {
		return nil, ErrNotConnected
	}

	if m.DBInfoErr != nil {
		return nil, m.DBInfoErr
	}

	if info, ok := m.DBInfos[database]; ok {
		return info, nil
	}

	return &DatabaseInfo{Name: database}, nil
}

// SetConnected allows setting the connection state directly.
func (m *MockClient) SetConnected(connected bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connected = connected
}

// SetTables sets the mock tables for a database.
func (m *MockClient) SetTables(database string, tables []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Tables[database] = tables
}

// SetTableSize sets the mock size for a table.
func (m *MockClient) SetTableSize(database, table string, size int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.TableSizes[database] == nil {
		m.TableSizes[database] = make(map[string]int64)
	}
	m.TableSizes[database][table] = size
}

// SetRowCount sets the mock row count for a table.
func (m *MockClient) SetRowCount(database, table string, count int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.RowCounts[database] == nil {
		m.RowCounts[database] = make(map[string]int64)
	}
	m.RowCounts[database][table] = count
}

// SetDatabaseSize sets the mock size for a database.
func (m *MockClient) SetDatabaseSize(database string, size int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.DBSizes[database] = size
}

// SetTableInfo sets the mock info for a table.
func (m *MockClient) SetTableInfo(database, table string, info *TableInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.TableInfos[database] == nil {
		m.TableInfos[database] = make(map[string]*TableInfo)
	}
	m.TableInfos[database][table] = info
}

// SetDatabaseInfo sets the mock info for a database.
func (m *MockClient) SetDatabaseInfo(database string, info *DatabaseInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.DBInfos[database] = info
}

// MockResult implements sql.Result for testing.
type MockResult struct {
	LastID   int64
	Affected int64
}

// LastInsertId returns the mock last insert ID.
func (r *MockResult) LastInsertId() (int64, error) {
	return r.LastID, nil
}

// RowsAffected returns the mock rows affected.
func (r *MockResult) RowsAffected() (int64, error) {
	return r.Affected, nil
}

// Ensure MockClient implements DatabaseClient interface.
var _ DatabaseClient = (*MockClient)(nil)
