# 0001. Use Go for Implementation

**Status:** Accepted  
**Date:** 2025-01-02

## Context

Cadangkan is a universal database backup and sync tool that needs to:
- Run on multiple platforms (Linux, macOS, Windows)
- Execute system commands (mysqldump, pg_dump, etc.)
- Handle concurrent backup operations efficiently
- Be distributed as a single binary for ease of installation
- Support CLI interfaces with rich features
- Interface with databases using native drivers
- Process large amounts of data efficiently

The language choice fundamentally affects:
- Development velocity and maintainability
- Performance characteristics for backup operations
- Ease of distribution and deployment
- Available ecosystem and libraries
- Team expertise and hiring considerations

## Decision

We will use **Go (Golang) version 1.21 or higher** as the primary implementation language for Cadangkan.

Specific choices:
- Go 1.21+ for latest language features and security updates
- Standard library for most operations
- Minimal external dependencies where possible
- Leverage Go's concurrency primitives (goroutines, channels)

## Consequences

### Positive

- **Single Binary Distribution:** Go compiles to a single statically-linked binary, making installation trivial (no dependencies, no runtime)
- **Cross-Platform:** Easy cross-compilation for Linux, macOS, Windows from a single codebase
- **Excellent Standard Library:** Built-in packages for file I/O, networking, compression, JSON, etc.
- **Native Performance:** Compiled language with performance close to C/C++, suitable for data-intensive backup operations
- **Built-in Concurrency:** Goroutines and channels enable easy parallel backups without complex threading
- **Strong CLI Ecosystem:** Libraries like cobra, urfave/cli provide excellent CLI building blocks
- **Database Drivers:** Mature drivers for MySQL, PostgreSQL, and other databases
- **Fast Compilation:** Quick build times improve development iteration
- **Memory Safety:** Garbage collected but predictable, avoiding common bugs
- **Growing Ecosystem:** Large and active community, plenty of libraries for backup/storage operations

### Negative

- **Garbage Collection:** GC pauses could affect performance in extreme cases (though rare for backup workloads)
- **Binary Size:** Go binaries are larger than C/C++ equivalents (but acceptable for CLI tools)
- **Learning Curve:** Team members unfamiliar with Go will need to learn:
  - Go idioms and conventions
  - Concurrency patterns
  - Error handling patterns (no exceptions)
- **No Generics (pre-1.18):** Older Go versions lack generics, but Go 1.21+ has them
- **Dependency Management:** Go modules are good but can have version conflicts
- **Limited OOP:** No traditional classes/inheritance (uses composition and interfaces instead)

### Risks

- **Ecosystem Changes:** Go language and ecosystem could introduce breaking changes (mitigated by pinning versions)
- **Performance Edge Cases:** GC might cause issues with extremely large backups (can be tuned)
- **Team Adoption:** If team is primarily experienced with other languages, productivity may be initially slower

## Alternatives Considered

### Alternative 1: Python

**Description:** Python with libraries like SQLAlchemy, paramiko, boto3

**Pros:**
- Rapid development with concise syntax
- Huge ecosystem of libraries
- Easy to learn and read
- Strong database support
- Excellent for scripting and system automation

**Cons:**
- Requires Python runtime installation
- Significantly slower for CPU-intensive operations
- GIL (Global Interpreter Lock) limits true concurrency
- Packaging/distribution is complex (virtualenv, pip, dependencies)
- Version fragmentation (Python 2 vs 3, different minor versions)

**Why not chosen:** Distribution complexity and performance concerns make it less suitable for a tool that users should be able to download and run immediately. Backup operations can be CPU-intensive (compression, encryption), where Python's performance would be a limitation.

### Alternative 2: Rust

**Description:** Systems programming language with memory safety guarantees

**Pros:**
- Best-in-class performance and memory efficiency
- Memory safety without garbage collection
- Growing ecosystem
- Excellent for system tools
- Strong type system prevents many bugs

**Cons:**
- Steeper learning curve (ownership, borrowing, lifetimes)
- Slower development velocity for most teams
- Smaller ecosystem compared to Go (especially for database drivers)
- Longer compilation times
- More complex error handling

**Why not chosen:** While Rust offers superior performance, the development velocity and ecosystem maturity of Go provide better trade-offs for a database backup tool. The learning curve for Rust would slow down development, and Go's performance is more than adequate for backup operations.

### Alternative 3: Shell Scripts (Bash)

**Description:** Traditional Unix shell scripts wrapping mysqldump, pg_dump, etc.

**Pros:**
- Simple for basic operations
- Universal availability on Unix systems
- Easy to read and modify
- Direct system command execution

**Cons:**
- Poor error handling
- No cross-platform support (Windows)
- Limited data structures and logic
- Hard to test
- Poor performance for complex operations
- Difficult to maintain as complexity grows

**Why not chosen:** While shell scripts work for simple backup tasks, Cadangkan requires robust error handling, cross-platform support, configuration management, and advanced features (scheduling, retention policies, cloud upload) that are painful to implement in shell scripts.

### Alternative 4: Node.js

**Description:** JavaScript runtime with npm ecosystem

**Pros:**
- Large ecosystem (npm)
- Familiar to web developers
- Good async I/O performance
- Cross-platform

**Cons:**
- Requires Node.js runtime
- Less suitable for CPU-intensive operations
- Callback hell / async complexity
- npm dependency management issues
- V8 memory limits for large operations

**Why not chosen:** Similar distribution challenges to Python, and JavaScript's asynchronous model adds complexity for sequential backup operations. The ecosystem is more focused on web development than system tools.

## Related Decisions

- ADR-0002: MySQL Client Architecture (builds on Go's standard library patterns)
- ADR-0003: Use go-sqlmock for Testing (Go-specific testing approach)

## Notes

- Go version should be kept reasonably current to get security updates and language improvements
- The decision to use Go 1.21+ ensures we have generics available if needed
- We may need to use CGO for some native library integrations, but should minimize its use
- Consider using `-ldflags` to reduce binary size in production builds

### References

- [Go Official Website](https://golang.org/)
- [Go Standard Library](https://pkg.go.dev/std)
- [Effective Go](https://golang.org/doc/effective_go)
- [Go Proverbs](https://go-proverbs.github.io/)
