# 0002. MySQL Client Architecture

**Status:** Accepted  
**Date:** 2025-01-02

## Context

Cadangkan needs a MySQL client to:
- Connect to MySQL databases with various configurations
- Execute queries and retrieve metadata (databases, tables, sizes)
- Support both local and remote databases
- Handle multiple databases concurrently
- Provide reliable connection management
- Support different MySQL versions (5.7, 8.0+)
- Be testable without requiring a running database
- Be thread-safe for concurrent operations

The client architecture affects:
- Code organization and maintainability
- Testing strategy
- Error handling patterns
- Configuration management
- Performance and resource usage

We need to make decisions about:
1. How to structure the client code
2. How to manage configuration
3. How to handle connections
4. How to expose the API
5. How to ensure thread safety

## Decision

We will implement a **package-based MySQL client** with the following structure:

**Package:** `pkg/database/mysql`

**Core Components:**
1. **Config struct** - Configuration with validation and DSN generation
2. **Client struct** - Main client with connection pool
3. **Custom Error types** - Structured error handling
4. **Interface** - DatabaseClient interface for mocking
5. **Mock implementation** - Testing support

**Key Design Choices:**

### Configuration Management
- Dedicated `Config` struct with all connection parameters
- Builder pattern with `WithHost()`, `WithPort()`, etc. for fluent API
- Validation method to catch errors early
- DSN generation for go-sql-driver/mysql
- Masked DSN for logging (hides passwords)

### Connection Management
- Use `database/sql` package for connection pooling
- Configure pool settings (MaxOpenConns, MaxIdleConns, lifetimes)
- Thread-safe with `sync.RWMutex` for state management
- Explicit `Connect()`, `Ping()`, `Close()` methods
- Context-based timeouts for operations

### API Design
- Clear separation between query methods and introspection methods
- Consistent error handling with custom error types
- Return raw `sql.Rows` for flexibility
- Provide high-level convenience methods (GetDatabases, GetTables, etc.)

### Thread Safety
- All public methods are thread-safe
- Use read locks for queries, write locks for state changes
- Leverage `database/sql`'s built-in thread safety

## Consequences

### Positive

- **Clean Separation of Concerns:** Config, connection, operations are distinct
- **Testable:** Interface allows mocking, client can be tested with go-sqlmock
- **Type Safe:** Strong typing catches errors at compile time
- **Thread Safe:** Can be used from multiple goroutines safely
- **Flexible:** Easy to extend with new methods
- **Well Documented:** Package documentation with examples
- **Ergonomic API:** Builder pattern makes configuration readable
- **Proper Error Handling:** Custom errors provide context
- **Reusable:** Other packages can import and use the client
- **Standard Patterns:** Follows Go idioms and best practices

### Negative

- **More Code:** More comprehensive than a simple wrapper (but better long-term)
- **Learning Curve:** New contributors need to understand the structure
- **Abstraction Overhead:** Slight performance cost vs direct SQL (negligible)
- **Dependency on database/sql:** Tied to standard library patterns

### Risks

- **API Changes:** If we need to change the interface, could affect consumers
- **Performance Tuning:** Connection pool settings might need adjustment per use case
- **MySQL Version Differences:** Some queries may behave differently across versions

## Alternatives Considered

### Alternative 1: Simple Wrapper Functions

**Description:** Package with standalone functions like `Connect(host, port, user, pass)`, `Query(db, sql)`, etc.

**Pros:**
- Simpler to implement initially
- Easy to understand
- Less code

**Cons:**
- No state management
- Connections passed around explicitly
- Hard to test
- No configuration validation
- Thread safety unclear
- Hard to extend

**Why not chosen:** Doesn't scale well as functionality grows. State management becomes messy, and testing is difficult.

### Alternative 2: Singleton Pattern

**Description:** Single global client instance accessed via `GetClient()`

**Pros:**
- Easy access from anywhere
- Single connection pool

**Cons:**
- Global state makes testing hard
- Can't have multiple database connections easily
- Tight coupling across codebase
- Initialization order issues

**Why not chosen:** Singletons are generally considered an anti-pattern in Go. Makes testing difficult and reduces flexibility.

### Alternative 3: ORM Approach (GORM, SQLX)

**Description:** Use a full ORM library for database operations

**Pros:**
- Lots of features out of the box
- Established patterns
- Query builders

**Cons:**
- Heavy dependency
- Unnecessary complexity for our use case
- We mostly need raw SQL and introspection, not object mapping
- Harder to customize for backup-specific needs

**Why not chosen:** Overkill for our needs. We're not building typical CRUD operations; we need database introspection and metadata for backups. Direct SQL with `database/sql` is more appropriate.

### Alternative 4: Direct database/sql Usage Throughout

**Description:** Import `database/sql` directly in each package that needs database access

**Pros:**
- No abstraction
- Direct access to all features
- Simple

**Cons:**
- Code duplication across packages
- No centralized error handling
- Hard to test
- Configuration scattered
- No type safety beyond basic SQL

**Why not chosen:** Would lead to code duplication and inconsistent patterns across the codebase. Better to centralize database logic in a package.

## Related Decisions

- ADR-0001: Use Go for Implementation (enables this design)
- ADR-0003: Use go-sqlmock for Testing (testing strategy)
- ADR-0004: Custom Error Types (error handling)
- ADR-0005: Interface-Based Client Design (mocking strategy)
- ADR-0006: Connection Pool Configuration (specific settings)

## Notes

### Package Structure

```
pkg/database/mysql/
├── client.go      # Main Client implementation
├── config.go      # Config struct and methods
├── errors.go      # Custom error types
├── interface.go   # DatabaseClient interface
├── mock.go        # MockClient for testing
├── doc.go         # Package documentation
└── client_test.go # Comprehensive tests
```

### Example Usage

```go
config := mysql.NewConfig().
    WithHost("localhost").
    WithPort(3306).
    WithUser("root").
    WithPassword("secret").
    WithDatabase("mydb")

client, err := mysql.NewClient(config)
if err != nil {
    log.Fatal(err)
}
defer client.Close()

if err := client.Connect(); err != nil {
    log.Fatal(err)
}

// Get MySQL version
version, err := client.GetVersion()

// List databases
databases, err := client.GetDatabases()

// Get database size
size, err := client.GetDatabaseSize("mydb")
```

### Future Enhancements

Potential extensions without breaking changes:
- Transaction support
- Prepared statement caching
- Query result caching
- Connection retry logic
- Health check endpoints
- Metrics/instrumentation

### References

- [database/sql package](https://pkg.go.dev/database/sql)
- [go-sql-driver/mysql](https://github.com/go-sql-driver/mysql)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
