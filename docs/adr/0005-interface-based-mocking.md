# 0005. Interface-Based Client Design

**Status:** Accepted  
**Date:** 2025-01-02

## Context

The MySQL client will be used by other components in Cadangkan (backup service, restore service, CLI commands). These components need:

**Testability Requirements:**
- Ability to unit test backup logic without a real database
- Mock database operations for edge case testing
- Fast tests that don't require MySQL running
- Deterministic test behavior

**Flexibility Requirements:**
- Potential for multiple database client implementations
- Support for different MySQL client strategies
- Easy dependency injection
- Loose coupling between components

**Design Considerations:**
- Go best practices favor small, focused interfaces
- Interfaces should be defined by consumers, not providers
- Testing should be easy without heavy mocking frameworks
- Clear contracts between components

**Problem Statement:**
Without an interface, consumers of the MySQL client would:
- Be tightly coupled to the concrete Client implementation
- Need a real database for testing
- Be hard to test with edge cases (timeouts, errors)
- Can't easily swap implementations

## Decision

We will define a **DatabaseClient interface** that specifies the contract for MySQL operations.

**Key Decisions:**

### 1. Interface Definition
Create `DatabaseClient` interface in `pkg/database/mysql/interface.go` with all client methods:
- Connection management (Connect, Ping, Close, IsConnected)
- Query execution (ExecuteQuery, ExecuteQueryArgs, Execute)
- Introspection (GetVersion, GetDatabases, GetTables, GetTableSize, etc.)

### 2. Implementations
Provide two implementations:
- **Client** - Real MySQL client using database/sql
- **MockClient** - Testing mock with configurable responses

### 3. Interface Location
Define interface in the same package as the implementation (not in a separate package)

### 4. Consumer Usage
Consumers accept `DatabaseClient` interface, not concrete Client:
```go
type BackupService struct {
    db mysql.DatabaseClient
}
```

### 5. Mock Design
MockClient tracks method calls and allows configurable responses for thorough testing

## Consequences

### Positive

- **Testability:** Consumers can easily test with MockClient
- **Dependency Injection:** Easy to inject different implementations
- **Loose Coupling:** Consumers depend on interface, not concrete type
- **Flexibility:** Can add new implementations without breaking consumers
- **Clear Contract:** Interface documents what operations are available
- **Mocking Without Frameworks:** Built-in mock, no need for gomock/mockery
- **Type Safety:** Interface ensures all implementations have required methods
- **Easy to Understand:** Interface is self-documenting
- **Facilitates TDD:** Can define interface before implementation

### Negative

- **Slight Indirection:** Extra abstraction layer
- **Interface Maintenance:** Changes require updating interface and implementations
- **Runtime Polymorphism:** Minor performance cost (negligible in practice)
- **More Code:** Need to maintain interface and mock implementation
- **Duplication:** Interface lists methods that are already on Client

### Risks

- **Interface Bloat:** Interface could grow too large
- **Breaking Changes:** Adding methods to interface breaks implementers
- **Mock Maintenance:** Mock needs updating when interface changes

**Mitigations:**
- Keep interface focused on essential operations
- Version interface if breaking changes needed
- Use code generation for mock if it grows complex
- Regular review to ensure interface is still appropriate

## Alternatives Considered

### Alternative 1: No Interface (Concrete Client Only)

**Description:** Consumers use *Client directly, no interface

**Pros:**
- Simpler, less abstraction
- Less code to maintain
- No interface maintenance

**Cons:**
- Hard to test consumers
- Tight coupling
- Can't swap implementations
- Need real database for all tests

**Why not chosen:** Makes testing very difficult. Consumers would need sqlmock in every test, defeating the purpose of separation of concerns.

### Alternative 2: Consumer-Defined Interfaces

**Description:** Each consumer defines its own interface with just the methods it needs

```go
// In backup package
type DatabaseReader interface {
    GetDatabases() ([]string, error)
    GetTables(string) ([]string, error)
}
```

**Pros:**
- Follows Go proverb "accept interfaces, return structs"
- Consumers only depend on what they need
- Smaller, focused interfaces
- More idiomatic Go

**Cons:**
- Interface duplication across consumers
- Harder to provide a standard mock
- Each consumer needs its own mock
- More boilerplate

**Why not chosen:** While more idiomatic, the MySQL client is complex enough that most consumers need many of the same methods. A single interface is more practical. If consumers only need 1-2 methods, they can still define their own interface.

### Alternative 3: Mocking Framework (gomock/mockery)

**Description:** Don't provide MockClient, use code generation tools

**Pros:**
- Less manual code
- Automatically updated
- Standard approach

**Cons:**
- External tool dependency
- Generated code in repo or build step
- Learning curve for team
- Harder to customize mock behavior

**Why not chosen:** Our MockClient is straightforward enough to write manually. It also allows custom features like call tracking that are useful for testing. Can always switch to generated mocks later if needed.

### Alternative 4: Separate Interface Package

**Description:** Put interface in `pkg/database/interface` separate from implementation

**Pros:**
- Clear separation
- No import cycles
- Interface stands alone

**Cons:**
- Not idiomatic Go
- Extra package to maintain
- Unnecessary abstraction
- Import complexity

**Why not chosen:** Go idiom is to define interface in the same package as the implementation. Only separate if import cycles force you to.

### Alternative 5: Duck Typing (No Explicit Interface)

**Description:** Don't define interface, rely on structural typing

**Pros:**
- Minimal code
- Maximum flexibility

**Cons:**
- No compile-time checking
- Unclear contract
- Hard to discover what methods are needed
- Poor developer experience

**Why not chosen:** Explicit interfaces provide documentation and compile-time safety. Much better developer experience.

## Related Decisions

- ADR-0002: MySQL Client Architecture (interface is part of client design)
- ADR-0003: Use go-sqlmock for Testing (interface enables easy mocking at higher levels)

## Notes

### Interface Definition

```go
// DatabaseClient defines the interface for MySQL database operations.
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
}

// Ensure Client implements DatabaseClient interface
var _ DatabaseClient = (*Client)(nil)
```

### MockClient Features

The MockClient implementation provides:
- **Configurable Responses:** Set return values for each method
- **Error Simulation:** Configure errors for testing failure scenarios
- **Call Tracking:** Record which methods were called with what arguments
- **Call Verification:** Check how many times methods were called
- **State Management:** Track connection state
- **Helper Methods:** SetTables(), SetTableSize(), etc. for easy test setup

### Usage Example

**Consumer Code:**
```go
type BackupService struct {
    db mysql.DatabaseClient
}

func NewBackupService(db mysql.DatabaseClient) *BackupService {
    return &BackupService{db: db}
}

func (s *BackupService) ListBackupTargets() ([]string, error) {
    return s.db.GetDatabases()
}
```

**Production Usage:**
```go
client, _ := mysql.NewClient(config)
client.Connect()

service := NewBackupService(client)
databases, _ := service.ListBackupTargets()
```

**Test Usage:**
```go
func TestListBackupTargets(t *testing.T) {
    mock := mysql.NewMockClient()
    mock.SetConnected(true)
    mock.Databases = []string{"db1", "db2", "db3"}

    service := NewBackupService(mock)
    databases, err := service.ListBackupTargets()

    assert.NoError(t, err)
    assert.Equal(t, []string{"db1", "db2", "db3"}, databases)
    assert.Equal(t, 1, mock.GetCallCount("GetDatabases"))
}
```

### Interface Evolution

If interface needs to change:

**Adding Methods:**
- Add to interface
- Implement in Client
- Implement in MockClient
- Update tests
- May break consumers (they'll get compile errors)

**Removing Methods:**
- Deprecate first (add comment)
- Give migration period
- Remove from interface
- Clean up implementations

**Versioning:**
If breaking changes needed:
- Create DatabaseClientV2 interface
- Keep V1 for compatibility
- Migrate consumers gradually

### Best Practices

**For Interface Users:**
- Accept `DatabaseClient` in function signatures
- Don't type assert to concrete Client
- Use MockClient for testing
- Only depend on methods you actually use

**For Interface Maintainers:**
- Keep interface focused
- Don't add methods lightly
- Document each method clearly
- Maintain backward compatibility when possible

### Compiler Verification

Use this pattern to ensure implementations satisfy interface:
```go
var _ DatabaseClient = (*Client)(nil)
var _ DatabaseClient = (*MockClient)(nil)
```

This causes compile errors if implementation doesn't match interface.

### Future Considerations

Potential evolutions:
- Context-aware interface (methods take context.Context)
- Streaming interfaces for large results
- Transaction interface
- Batch operation interface
- Separate read/write interfaces

### References

- [Effective Go: Interfaces](https://golang.org/doc/effective_go#interfaces)
- [Go Proverbs](https://go-proverbs.github.io/)
- [Accept Interfaces, Return Structs](https://bryanftan.medium.com/accept-interfaces-return-structs-in-go-d4cab29a301b)
- [Interface Pollution in Go](https://rakyll.org/interface-pollution/)
