# 0006. Connection Pool Configuration

**Status:** Accepted  
**Date:** 2025-01-02

## Context

The MySQL client uses Go's `database/sql` package, which provides connection pooling. The pool needs proper configuration to balance:

**Performance Needs:**
- Handle concurrent backup operations
- Minimize connection overhead
- Avoid connection exhaustion
- Efficient resource usage

**Reliability Needs:**
- Prevent connection leaks
- Handle stale connections
- Recover from network issues
- Avoid overwhelming database server

**Resource Constraints:**
- Backup tool isn't a high-concurrency web server
- Typically 1-10 concurrent backups, not 1000s
- Database server may have connection limits
- Client machine has memory/CPU limits

**Configuration Parameters:**
Go's `database/sql` provides these knobs:
- `MaxOpenConns` - Maximum number of open connections
- `MaxIdleConns` - Maximum number of idle connections in pool
- `ConnMaxLifetime` - Maximum lifetime of a connection
- `ConnMaxIdleTime` - Maximum time a connection can be idle

**Questions to Answer:**
1. How many connections should we allow?
2. How long should connections live?
3. What are sensible defaults for a backup tool?
4. Should these be configurable by users?

## Decision

We will use the following **default connection pool settings**:

```go
MaxOpenConns:    25    // Maximum open connections
MaxIdleConns:    10    // Maximum idle connections
ConnMaxLifetime: 5m    // Connection lifetime (5 minutes)
ConnMaxIdleTime: 30s   // Idle connection timeout (30 seconds)
```

**Rationale for Each Setting:**

### MaxOpenConns: 25
- Backup operations are typically sequential per database
- Allows for ~10 concurrent database backups with headroom
- Far below typical MySQL max_connections (151 default, often 500+)
- Prevents overwhelming the database server
- Higher than typical connection needs, leaves room for growth

### MaxIdleConns: 10
- Keeps connections warm for quick reuse
- Balance between memory usage and performance
- About 40% of MaxOpenConns (common ratio)
- Enough for typical concurrent operations
- Not so many that we hold connections unnecessarily

### ConnMaxLifetime: 5 minutes
- Prevents stale connections
- Long enough for backup operations to complete
- Short enough to recover from network changes
- Aligns with typical network timeout values
- Prevents connection leaks in long-running processes

### ConnMaxIdleTime: 30 seconds
- Aggressive cleanup of unused connections
- Returns connections to server quickly
- Appropriate for batch operations (backups happen then idle)
- Low memory footprint when not in use
- Still fast enough for typical backup intervals

**User Configurability:**
- Settings exposed in Config struct
- Users can override defaults if needed
- Defaults suitable for 90%+ of use cases

## Consequences

### Positive

- **Good Defaults:** Works well for typical backup scenarios
- **Resource Efficient:** Doesn't hold excessive connections
- **Scalable:** Handles concurrent backups well
- **Reliable:** Automatic cleanup prevents issues
- **Server Friendly:** Won't overwhelm database server
- **Tunable:** Advanced users can adjust if needed
- **Memory Efficient:** Idle connections cleaned up quickly
- **Performance:** Warm pool provides fast connections

### Negative

- **Not One-Size-Fits-All:** Defaults may not suit all scenarios
- **Configuration Complexity:** Users need to understand pool settings
- **Potential Under-Utilization:** Some use cases might need more connections
- **Idle Timeout Tuning:** May need adjustment for different workloads

### Risks

- **Connection Exhaustion:** If many concurrent operations, might hit limit
- **Idle Cleanup:** Aggressive idle timeout might reconnect too often
- **Lifetime vs Operation Time:** Long operations might exceed lifetime

**Mitigations:**
- Document when to adjust settings
- Provide examples for different scenarios
- Monitor connection usage in logs
- Add metrics if needed

## Alternatives Considered

### Alternative 1: Unlimited Connections

**Description:** Set MaxOpenConns to 0 (unlimited)

**Pros:**
- Never blocks on connection
- Maximum parallelism

**Cons:**
- Can overwhelm database server
- Memory exhaustion risk
- Poor resource management
- Can hit database connection limits

**Why not chosen:** Too risky. Backup operations could spawn many goroutines and exhaust connections. Need explicit limits.

### Alternative 2: Very Conservative (5/2/1m/10s)

**Description:** MaxOpenConns=5, MaxIdleConns=2, ConnMaxLifetime=1m, ConnMaxIdleTime=10s

**Pros:**
- Very resource efficient
- Safe for small servers
- Low memory footprint

**Cons:**
- Limits concurrent operations
- More connection overhead (frequent reconnection)
- Slower for parallel backups
- Overly conservative for most use cases

**Why not chosen:** Too restrictive. Modern databases can handle more connections. Backup operations benefit from some parallelism.

### Alternative 3: Web Server Defaults (100/50/5m/5m)

**Description:** High connection count typical for web apps

**Pros:**
- Handles high concurrency
- Fast connection availability
- Rarely hits limits

**Cons:**
- Overkill for backup workload
- Wastes database resources
- Higher memory usage
- Not appropriate for tool usage pattern

**Why not chosen:** Backup tool isn't a high-concurrency web server. These settings optimize for thousands of short requests, not dozens of long operations.

### Alternative 4: No Idle Connections (25/0/5m/0)

**Description:** Don't keep idle connections

**Pros:**
- Minimal resource usage
- Returns connections immediately

**Cons:**
- Connection overhead on every operation
- Slower performance
- More database load (constant connect/disconnect)

**Why not chosen:** Backup operations often come in bursts (back up multiple databases). Keeping some idle connections improves performance.

### Alternative 5: Dynamic Configuration

**Description:** Automatically adjust pool based on workload

**Pros:**
- Optimizes for actual usage
- No manual tuning needed

**Cons:**
- Complex to implement
- Unpredictable behavior
- Hard to debug
- Overkill for backup tool

**Why not chosen:** Over-engineering. Static configuration with good defaults is simpler and more predictable for a backup tool.

## Related Decisions

- ADR-0002: MySQL Client Architecture (pool is part of client)
- ADR-0001: Use Go for Implementation (uses database/sql pooling)

## Notes

### Pool Behavior

**Connection Lifecycle:**
1. Client calls a query method
2. Pool checks for idle connection
3. If available: reuse. If not and under MaxOpenConns: create new
4. After use: return to idle pool (if under MaxIdleConns) or close
5. Idle connections cleaned up after ConnMaxIdleTime
6. All connections closed after ConnMaxLifetime

**Blocking Behavior:**
If all MaxOpenConns are in use, requests block until a connection is available.

### Tuning Guidelines

**When to Increase MaxOpenConns:**
- Backing up 10+ databases concurrently
- Large database server with high connection limit
- Operations timing out waiting for connections

**When to Decrease MaxOpenConns:**
- Small database server (low max_connections)
- Memory constraints
- Single-threaded backup operations

**When to Adjust Lifetimes:**
- Longer ConnMaxLifetime: for very long backup operations (>5 minutes)
- Shorter ConnMaxIdleTime: to free resources faster when idle
- Longer ConnMaxIdleTime: if backups run frequently (e.g., every minute)

### Example Configurations

**Small Server (Raspberry Pi, 1GB RAM):**
```go
config.MaxOpenConns = 10
config.MaxIdleConns = 5
config.ConnMaxLifetime = 3 * time.Minute
config.ConnMaxIdleTime = 20 * time.Second
```

**Large Server (Dedicated backup server):**
```go
config.MaxOpenConns = 50
config.MaxIdleConns = 20
config.ConnMaxLifetime = 10 * time.Minute
config.ConnMaxIdleTime = 1 * time.Minute
```

**Low-Frequency Backups (once per hour):**
```go
config.MaxOpenConns = 25
config.MaxIdleConns = 5  // Lower since long idle periods
config.ConnMaxLifetime = 5 * time.Minute
config.ConnMaxIdleTime = 10 * time.Second  // Cleanup quickly
```

**High-Frequency Backups (every 5 minutes):**
```go
config.MaxOpenConns = 25
config.MaxIdleConns = 15  // Keep more warm
config.ConnMaxLifetime = 10 * time.Minute
config.ConnMaxIdleTime = 2 * time.Minute  // Keep longer
```

### Monitoring

To verify pool settings are appropriate, monitor:
- Wait time for connections (if high, increase MaxOpenConns)
- Connection creation rate (if high, increase MaxIdleConns or idle time)
- Memory usage (if high, decrease idle connections)
- Database connection count (ensure not hitting max_connections)

### Database Server Considerations

**MySQL max_connections:**
- Default: 151
- Common: 500-1000
- Check: `SHOW VARIABLES LIKE 'max_connections';`
- Our 25 connections use 5-17% of typical limits (safe)

**Connection Overhead:**
- Each connection uses ~256KB of memory
- 25 connections ≈ 6.4MB (negligible)
- Idle connections use minimal CPU

### Code Implementation

```go
// Configure connection pool
db.SetMaxOpenConns(c.config.MaxOpenConns)
db.SetMaxIdleConns(c.config.MaxIdleConns)
db.SetConnMaxLifetime(c.config.ConnMaxLifetime)
db.SetConnMaxIdleTime(c.config.ConnMaxIdleTime)
```

### Testing Considerations

In tests with sqlmock:
- Pool settings don't affect mock behavior
- Integration tests should verify pool works correctly
- Can test pool exhaustion scenarios by limiting MaxOpenConns

### Future Enhancements

Potential improvements:
- Connection pool metrics/observability
- Dynamic adjustment based on workload
- Per-database pool configuration
- Connection health checks
- Pool statistics API

### References

- [database/sql Package](https://pkg.go.dev/database/sql)
- [Configuring sql.DB for Better Performance](https://www.alexedwards.net/blog/configuring-sqldb)
- [Go database/sql Tutorial](http://go-database-sql.org/connection-pool.html)
- [MySQL Connection Management](https://dev.mysql.com/doc/refman/8.0/en/connection-management.html)
