# 0004. Custom Error Types

**Status:** Accepted  
**Date:** 2025-01-02

## Context

The MySQL client needs robust error handling to:
- Provide clear, actionable error messages to users
- Enable programmatic error handling (check error types, not strings)
- Include context about what operation failed
- Support error wrapping and unwrapping (Go 1.13+ error chains)
- Differentiate between error categories (connection, query, timeout, config)
- Mask sensitive information in error messages (passwords)
- Provide debugging information without exposing secrets

**Current Pain Points with Generic Errors:**
- Can't distinguish between connection errors and query errors programmatically
- Error messages lack context (which host? which query?)
- Hard to handle errors appropriately (retry vs fail)
- Testing error conditions is awkward
- User-facing error messages are unclear

**Requirements:**
- Type-safe error checking
- Rich context in error messages
- Support for error wrapping
- Backward compatible with error interface
- Easy to use and test

## Decision

We will implement **custom error types** for different error categories in the MySQL client:

**Error Types:**
1. **ConnectionError** - Database connection failures
2. **QueryError** - Query execution failures  
3. **TimeoutError** - Operation timeouts
4. **ConfigError** - Configuration validation errors

**Design Principles:**
- Each error type is a struct implementing the `error` interface
- Include relevant context fields (host, port, query, etc.)
- Implement `Unwrap()` to support error chains
- Provide helper functions for error type checking
- Helper functions to wrap errors with context

**Sentinel Errors:**
- `ErrNotConnected` - Client is not connected
- `ErrAlreadyConnected` - Client is already connected
- `ErrInvalidConfig` - Invalid configuration
- `ErrEmptyResult` - Query returned no results

## Consequences

### Positive

- **Type-Safe Error Handling:** Use `errors.As()` instead of string matching
- **Rich Context:** Errors include relevant information (host, port, query)
- **Better User Experience:** Clear, actionable error messages
- **Debugging-Friendly:** Easy to understand what went wrong
- **Testable:** Easy to create and check error types in tests
- **Error Chains:** Support wrapping with `fmt.Errorf("%w", err)`
- **Programmatic Handling:** Code can respond differently to different errors
- **Security:** Can mask sensitive data in error strings
- **Go Idioms:** Follows Go 1.13+ error handling patterns

### Negative

- **More Code:** More verbose than returning string errors
- **Learning Curve:** Team needs to understand when to use which error type
- **Consistency Burden:** Need discipline to use correct error types
- **Breaking Changes:** Adding fields to error structs could break consumers

### Risks

- **Over-Engineering:** Could create too many error types
- **Misuse:** Developers might use wrong error type
- **Testing Overhead:** More error types = more test cases

**Mitigations:**
- Keep error types minimal and focused
- Document when to use each error type
- Provide helper functions to make usage easy
- Code review for proper error type usage

## Alternatives Considered

### Alternative 1: Generic errors.New()

**Description:** Use standard `errors.New()` and `fmt.Errorf()` everywhere

**Pros:**
- Simple and familiar
- No custom code needed
- Works with standard library

**Cons:**
- Can't distinguish error types programmatically
- Must parse error strings to understand type
- No structured context
- Hard to test
- Poor user experience

**Why not chosen:** Too limited for a production library. Users need to handle different error scenarios differently (retry connection errors, don't retry config errors).

### Alternative 2: Error Codes (Integer/String)

**Description:** Return error with code field like `Error{Code: "CONN_001", Message: "..."}`

**Pros:**
- Can check error codes programmatically
- Easy to internationalize
- Common in other languages

**Cons:**
- Need to maintain error code registry
- Less idiomatic in Go
- Doesn't use Go's error wrapping
- String/int comparisons are fragile

**Why not chosen:** Not idiomatic Go. Error types are more type-safe and Go-like.

### Alternative 3: pkg/errors (Third-Party)

**Description:** Use `github.com/pkg/errors` for stack traces and wrapping

**Pros:**
- Stack traces for debugging
- Good wrapping support
- Widely used

**Cons:**
- External dependency
- Go 1.13+ has built-in wrapping
- Heavier than needed
- Stack traces can be verbose

**Why not chosen:** Go 1.13+ provides error wrapping natively. Stack traces are useful but can be added later if needed. Prefer standard library.

### Alternative 4: Single Error Type with Category Field

**Description:** One error struct with a `Category` enum field

```go
type Error struct {
    Category ErrorCategory // CONNECTION, QUERY, TIMEOUT, CONFIG
    Message  string
    Err      error
}
```

**Pros:**
- Single type to learn
- Easy to extend categories

**Cons:**
- Switch statements everywhere
- Can't use `errors.As()` to distinguish types
- Less type-safe
- Can't have type-specific fields

**Why not chosen:** Loses type safety benefits. Better to use Go's type system.

### Alternative 5: Error Interface Hierarchy

**Description:** Create interfaces like `ConnectionError interface { error; IsConnectionError() }`

**Pros:**
- Very flexible
- Can have multiple implementations

**Cons:**
- More complex
- Overhead for simple cases
- Harder to use

**Why not chosen:** Overly complex for our needs. Struct types with `errors.As()` are simpler.

## Related Decisions

- ADR-0002: MySQL Client Architecture (part of client design)
- ADR-0003: Use go-sqlmock for Testing (testing errors)

## Notes

### Error Type Implementations

```go
// ConnectionError for connection failures
type ConnectionError struct {
    Host    string
    Port    int
    Message string
    Err     error
}

func (e *ConnectionError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("mysql connection error to %s:%d: %s: %v",
            e.Host, e.Port, e.Message, e.Err)
    }
    return fmt.Sprintf("mysql connection error to %s:%d: %s",
        e.Host, e.Port, e.Message)
}

func (e *ConnectionError) Unwrap() error {
    return e.Err
}
```

### Usage Examples

**Creating Errors:**
```go
// Using helper function
err := WrapConnectionError("localhost", 3306, "connection refused", underlyingErr)

// Direct creation
err := &QueryError{
    Query:   "SELECT * FROM users",
    Message: "syntax error",
    Err:     underlyingErr,
}
```

**Checking Error Types:**
```go
// Check if error is a connection error
if IsConnectionError(err) {
    // Retry connection
}

// Get specific error type
var connErr *ConnectionError
if errors.As(err, &connErr) {
    log.Printf("Failed to connect to %s:%d", connErr.Host, connErr.Port)
}

// Check sentinel errors
if errors.Is(err, ErrNotConnected) {
    // Handle not connected
}
```

### Error Type Matrix

| Error Type | Use When | Fields | Example |
|------------|----------|--------|---------|
| ConnectionError | Connection fails | Host, Port, Message, Err | Can't reach database |
| QueryError | Query fails | Query, Message, Err | SQL syntax error |
| TimeoutError | Operation times out | Operation, Duration, Err | Query took too long |
| ConfigError | Config invalid | Field, Message | Missing required field |

### Helper Functions

```go
// Type checking
IsConnectionError(err) bool
IsQueryError(err) bool
IsTimeoutError(err) bool
IsConfigError(err) bool

// Error wrapping
WrapConnectionError(host, port, message, err) error
WrapQueryError(query, message, err) error
WrapTimeoutError(operation, duration, err) error
```

### Security Considerations

- Never include passwords in error messages
- Truncate long queries in QueryError (max 100 chars)
- Use masked DSN for logging connections
- Be careful with error messages exposed to end users

### Testing

Custom errors make testing easier:
```go
func TestConnectionError(t *testing.T) {
    err := &ConnectionError{
        Host: "localhost",
        Port: 3306,
        Message: "test",
    }
    
    assert.True(t, IsConnectionError(err))
    assert.Contains(t, err.Error(), "localhost:3306")
}
```

### Future Enhancements

Potential additions:
- Retry-able error interface
- Temporary error detection
- Error codes for internationalization
- Structured logging integration
- Telemetry/metrics integration

### References

- [Go Error Handling](https://blog.golang.org/error-handling-and-go)
- [Working with Errors in Go 1.13](https://blog.golang.org/go1.13-errors)
- [errors package](https://pkg.go.dev/errors)
- [Effective Error Handling in Go](https://earthly.dev/blog/golang-errors/)
