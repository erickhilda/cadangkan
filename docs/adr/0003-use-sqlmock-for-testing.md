# 0003. Use go-sqlmock for Unit Testing

**Status:** Accepted  
**Date:** 2025-01-02

## Context

The MySQL client needs comprehensive unit tests to ensure reliability. Testing database code presents several challenges:

**Testing Requirements:**
- Test connection logic without a real database
- Test query execution and error handling
- Test all introspection methods (GetVersion, GetDatabases, etc.)
- Test concurrent access scenarios
- Test timeout handling
- Achieve >90% code coverage
- Tests should run quickly in CI/CD
- Tests should be deterministic and not flaky
- No external dependencies for running tests

**Challenges:**
- Real database tests are slow (startup time, cleanup)
- Real databases require Docker or external services
- Hard to simulate specific error conditions
- Database state cleanup between tests is complex
- Network issues can make tests flaky
- CI/CD runners may not have database access

**Options for Testing:**
1. Real MySQL database (via Docker or external)
2. In-memory SQLite (different SQL dialect)
3. SQL mocking library
4. Custom mock implementation
5. Integration tests only (no unit tests)

## Decision

We will use **`github.com/DATA-DOG/go-sqlmock`** for unit testing the MySQL client.

**Testing Strategy:**
- **Unit Tests:** Use go-sqlmock to test client logic without a database
- **Integration Tests:** Separate tests with real MySQL (Docker Compose) for end-to-end validation
- **Target:** >90% code coverage with unit tests
- **Also use:** `github.com/stretchr/testify` for assertions

**Key Aspects:**
- Mock `database/sql` connections with sqlmock
- Define expected queries and results
- Verify query execution and parameters
- Simulate various error conditions
- Test edge cases easily
- Keep tests fast (<1 second for entire suite)

## Consequences

### Positive

- **Fast Tests:** Unit tests run in milliseconds without database startup
- **Deterministic:** Tests always behave the same way, no flakiness
- **No External Dependencies:** CI/CD doesn't need Docker or MySQL
- **Easy Error Simulation:** Can test connection failures, timeouts, query errors easily
- **High Coverage:** Can achieve >90% coverage with comprehensive mocking
- **Parallel Execution:** Tests can run in parallel without interference
- **Clear Expectations:** sqlmock forces you to think about exact queries
- **Good Developer Experience:** Fast feedback loop during development
- **Well Maintained:** go-sqlmock is widely used and actively maintained
- **Works with database/sql:** Direct support for standard library patterns

### Negative

- **Not Testing Real MySQL:** Mocks don't catch MySQL-specific issues (SQL dialect, version differences)
- **Mock Maintenance:** Mocks need updating when queries change
- **False Confidence:** Passing unit tests don't guarantee database compatibility
- **Mock Complexity:** Complex scenarios require detailed mock setup
- **Query String Matching:** Tests can break on minor query format changes
- **Learning Curve:** Team needs to learn sqlmock API

### Risks

- **Divergence from Reality:** Mocked behavior might differ from actual MySQL
- **Regex Matching:** Query matching can be fragile if not done carefully
- **Integration Test Gap:** Need separate integration tests to catch real database issues

**Mitigations:**
- Maintain integration tests with real MySQL in Docker Compose
- Run integration tests in CI/CD regularly
- Document when integration tests should be run
- Use sqlmock's flexible matching (regex, any order for args)

## Alternatives Considered

### Alternative 1: Real MySQL with Docker

**Description:** Use Docker Compose to start MySQL containers for tests

**Pros:**
- Tests against real MySQL behavior
- Catches actual database issues
- No mocking complexity
- Validates SQL syntax and compatibility

**Cons:**
- Slow (10-30 seconds startup time)
- Requires Docker on dev machines and CI/CD
- Tests are harder to write (need cleanup, state management)
- Can be flaky (network issues, port conflicts)
- Resource intensive (memory, CPU)
- Parallel testing is complex

**Why not chosen:** Too slow for unit tests. We'll use this approach for integration tests but not for fast unit tests that run on every save.

### Alternative 2: In-Memory SQLite

**Description:** Use SQLite in-memory mode for tests

**Pros:**
- Fast, in-process database
- No external dependencies
- Easy to reset state

**Cons:**
- Different SQL dialect than MySQL
- Different feature set
- Behavior differences can hide bugs
- False sense of testing "real database"
- MySQL-specific features won't work

**Why not chosen:** Testing against SQLite when deploying to MySQL is risky. Dialect differences mean tests can pass but production fails. Better to mock explicitly than test against wrong database.

### Alternative 3: Custom Mock Implementation

**Description:** Write our own mock implementation of database/sql interfaces

**Pros:**
- Full control over mock behavior
- Can optimize for our specific use cases
- No external dependency

**Cons:**
- Significant development time
- Need to maintain our mock
- Likely to have bugs
- Reinventing the wheel
- Less battle-tested than go-sqlmock

**Why not chosen:** go-sqlmock is well-established, tested, and maintained. Building our own would take significant time and likely have less features and more bugs.

### Alternative 4: Integration Tests Only

**Description:** Skip unit tests, only use real database integration tests

**Pros:**
- Tests real behavior
- No mocking complexity
- Simpler test code

**Cons:**
- Very slow test feedback loop
- Expensive in CI/CD (need database for every test run)
- Harder to test error conditions
- Lower test coverage
- Difficult to test edge cases

**Why not chosen:** Too slow for development workflow. Integration tests are important but shouldn't be the only tests. Unit tests provide fast feedback and high coverage.

### Alternative 5: testcontainers-go

**Description:** Use testcontainers to manage Docker containers in tests

**Pros:**
- Real MySQL in tests
- Automatic container management
- Better than manual Docker setup

**Cons:**
- Still requires Docker
- Still slow (container startup)
- Complex setup
- Heavy dependency

**Why not chosen:** Same fundamental issues as Docker approach - too slow for unit tests. Could be useful for integration tests, but sqlmock is better for unit tests.

## Related Decisions

- ADR-0002: MySQL Client Architecture (what we're testing)
- ADR-0005: Interface-Based Client Design (enables easy mocking at higher levels)

## Notes

### Test Structure

Our test suite has two layers:

**1. Unit Tests (with go-sqlmock)**
- File: `client_test.go`
- Run: `go test ./pkg/database/mysql/...`
- Coverage target: >90%
- Run on every commit
- Fast: <1 second

**2. Integration Tests (with real MySQL)**
- File: `client_integration_test.go` (future)
- Run: `go test -tags=integration ./...`
- Uses Docker Compose
- Run before releases and in CI/CD
- Slower: ~30 seconds

### Example Test with sqlmock

```go
func TestGetVersion(t *testing.T) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close()

    rows := sqlmock.NewRows([]string{"VERSION()"}).
        AddRow("8.0.35")
    mock.ExpectQuery("SELECT VERSION()").
        WillReturnRows(rows)

    config := NewConfig().
        WithHost("localhost").
        WithUser("root").
        WithTimeout(5 * time.Second)
    client, _ := NewClientWithDB(config, db)

    version, err := client.GetVersion()
    assert.NoError(t, err)
    assert.Equal(t, "8.0.35", version)
}
```

### Coverage Achieved

With go-sqlmock, we achieved:
- **90.0% code coverage**
- **70+ test cases**
- **All methods tested** (success and failure paths)
- **Concurrent access tested**
- **Error conditions tested**

### Future Considerations

- Add integration tests with real MySQL in Phase 1, Week 2
- Consider using testcontainers for integration tests
- Document when to run which tests
- Add CI/CD workflow for integration tests

### References

- [go-sqlmock GitHub](https://github.com/DATA-DOG/go-sqlmock)
- [go-sqlmock Documentation](https://pkg.go.dev/github.com/DATA-DOG/go-sqlmock)
- [Testing in Go](https://golang.org/doc/tutorial/add-a-test)
- [Testify GitHub](https://github.com/stretchr/testify)
