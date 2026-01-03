# 0007. mysqldump Backup Strategy

**Status:** Accepted  
**Date:** 2025-01-02

## Context

Cadangkan needs to create backups of MySQL databases. There are several approaches to implement database backups:

**Options Available:**
1. Use mysqldump command-line utility
2. Implement native Go MySQL backup (read tables, write SQL)
3. Use MySQL binary log (binlog) based backups
4. Use filesystem-level snapshots (LVM, ZFS)
5. Use third-party Go libraries for backup

**Requirements:**
- Must work with any MySQL server (5.7+, 8.0+)
- Must handle databases from 100MB to 100GB+
- Low memory footprint (<100MB regardless of database size)
- Compressed backups for storage efficiency
- Consistent snapshots (point-in-time)
- Include all database objects (tables, views, procedures, triggers, events)
- Compatible with restore operations
- No MySQL server plugins or special privileges required

**Challenges:**
- Large databases can't fit in memory
- Need to maintain consistency during backup
- Must handle various MySQL versions and configurations
- Need to preserve all database structure and data
- Performance is critical

## Decision

We will use **mysqldump with stream-based processing** for MySQL backups.

**Key Implementation Decisions:**

### 1. Use mysqldump Command-Line Tool

Execute `mysqldump` as an external command rather than implementing native Go backup.

**Optimal mysqldump Flags:**
```bash
mysqldump \
  --single-transaction \
  --quick \
  --routines \
  --triggers \
  --events \
  --no-tablespaces \
  --skip-lock-tables \
  --set-gtid-purged=OFF \
  {database}
```

### 2. Stream-Based Processing

Stream mysqldump output directly to compression without intermediate storage:

```
mysqldump → stdout → gzip compression → file
```

**Implementation:**
```go
cmd := exec.Command("mysqldump", args...)
stdout, _ := cmd.StdoutPipe()
gzWriter := gzip.NewWriter(file)
io.Copy(gzWriter, stdout)
```

### 3. Real-Time Compression

Compress data as it streams from mysqldump using gzip compression (level: default).

### 4. Checksum Calculation

Calculate SHA-256 checksum while compressing using `io.TeeReader`:

```go
hasher := sha256.New()
checksumReader := io.TeeReader(reader, hasher)
// Process checksumReader instead of reader
```

### 5. Metadata Storage

Store comprehensive metadata alongside each backup in JSON format:
- Backup ID (timestamp: YYYY-MM-DD-HHMMSS)
- Database information (host, port, version)
- File information (size, checksum, compression)
- Timing information (duration, timestamps)
- Tool version and mysqldump version
- Options used

### 6. Error Handling

- Capture stderr from mysqldump for error messages
- Retry on transient failures (3 attempts with exponential backoff)
- Clean up partial backups on failure
- Detailed error types for different failure modes

## Consequences

### Positive

- **Battle-Tested:** mysqldump is mature, well-tested, used by millions
- **Compatible:** Works with all MySQL versions (5.7, 8.0, MariaDB)
- **Complete:** Captures all database objects (tables, views, procedures, triggers, events)
- **Consistent:** `--single-transaction` provides point-in-time snapshot without locking
- **Memory Efficient:** Stream processing uses <100MB regardless of database size
- **Fast:** Direct streaming avoids intermediate I/O
- **Standard Format:** SQL dump format is universally compatible
- **Restore-Friendly:** SQL dumps can be restored with standard `mysql` command
- **Portable:** Backups work across different MySQL installations
- **No Special Privileges:** Doesn't require MySQL server plugins or FILE privilege
- **Proven Flags:** Our flag combination is well-documented and widely used
- **Compression:** gzip provides good balance of speed and compression ratio (~70%)
- **Integrity:** SHA-256 checksum ensures backup integrity

### Negative

- **External Dependency:** Requires mysqldump binary to be installed (MySQL client tools only, not full server)
- **Version Dependency:** mysqldump behavior varies slightly across versions (mitigated by using MySQL 8.0 client)
- **Not Binary Format:** SQL dumps are less efficient than binary formats
- **Single-Threaded:** mysqldump processes tables sequentially
- **Full Backups Only:** No built-in incremental backup support (initially)
- **Large Databases:** Very large databases (100GB+) take significant time
- **Network Overhead:** Remote databases transfer all data over network
- **Text Format:** Larger than binary formats before compression

### Risks

- **mysqldump Not Available:** User's system might not have mysqldump
- **Version Incompatibility:** Older mysqldump versions might not support all flags
- **Performance:** Very large databases might take too long
- **Memory Issues:** Extremely large single tables might cause issues

**Mitigations:**
- Check for mysqldump availability and version at startup
- Document mysqldump installation requirements clearly
- Recommend MySQL 8.0 client tools for best compatibility
- Provide clear error messages if mysqldump missing
- Future: Add support for alternative backup methods
- Future: Implement parallel table dumps for very large databases

### mysqldump Installation Requirements

**Important:** Users need MySQL client tools installed, **not the full MySQL server**.

**Installation by Platform:**

- **Linux (Debian/Ubuntu):** `sudo apt-get install mysql-client`
- **Linux (RHEL/CentOS/Fedora):** `sudo yum install mysql` or `sudo dnf install mysql`
- **macOS:** `brew install mysql-client` (requires PATH configuration)
- **Windows:** MySQL Installer (select "MySQL Client" component only)

**Version Compatibility:**

- **Recommended:** Install MySQL 8.0 client tools for best compatibility
- **Forward Compatible:** MySQL 8.0 `mysqldump` can backup both MySQL 5.7 and 8.0 servers
- **Backward Compatible:** MySQL 5.7 `mysqldump` may work with MySQL 8.0 servers but can miss features
- **Cross-Version:** A single `mysqldump` version can handle multiple server versions
- **Best Practice:** Use `mysqldump` version matching or newer than the source MySQL server

**Why This Matters:**
- Lighter footprint than full MySQL server installation
- Can backup remote MySQL servers without local server
- Simpler deployment and maintenance
- No MySQL server process overhead

## Alternatives Considered

### Alternative 1: Native Go MySQL Backup

**Description:** Implement backup entirely in Go by reading tables and generating SQL

**Pros:**
- No external dependency
- Full control over process
- Could optimize for specific use cases
- Pure Go solution

**Cons:**
- Significant development effort
- Need to handle all MySQL data types correctly
- Need to handle all database objects (views, procedures, triggers, events)
- Need to maintain compatibility with MySQL versions
- Likely to have bugs that mysqldump doesn't have
- Would reinvent a well-tested wheel

**Why not chosen:** mysqldump is battle-tested and maintained by MySQL team. Reimplementing it would take months and likely have issues that mysqldump solved years ago.

### Alternative 2: MySQL Binary Log (binlog) Backups

**Description:** Use binary logs for incremental backups

**Pros:**
- Very efficient for incremental backups
- Point-in-time recovery
- Used by MySQL replication

**Cons:**
- Requires binary logging enabled (not default)
- Complex to implement
- Requires baseline full backup
- Harder to restore
- Binary format less portable
- Requires more MySQL privileges

**Why not chosen:** Too complex for MVP. Binary logs are great for advanced use cases but overkill for initial version. Can be added later for incremental backups.

### Alternative 3: Filesystem Snapshots

**Description:** Use LVM/ZFS snapshots of MySQL data directory

**Pros:**
- Very fast (snapshot is instant)
- Binary format (exact copy)
- Can backup while running

**Cons:**
- Requires specific filesystem setup
- Requires MySQL data directory access
- Not portable across systems
- Requires careful handling of InnoDB
- Doesn't work with remote databases
- Complex restore process

**Why not chosen:** Too many prerequisites. Most users don't have LVM/ZFS or direct filesystem access (cloud databases). Not suitable for general-purpose tool.

### Alternative 4: Third-Party Go Libraries

**Description:** Use existing Go libraries for MySQL backup (e.g., go-mydumper)

**Pros:**
- Pure Go
- Some performance optimizations
- No external binary

**Cons:**
- Less mature than mysqldump
- Smaller community
- May have compatibility issues
- Additional dependency to maintain
- May not support all MySQL features

**Why not chosen:** mysqldump is more reliable and widely used. Third-party libraries are less mature and may have compatibility issues. Better to use the official tool.

### Alternative 5: MySQL Enterprise Backup

**Description:** Use MySQL Enterprise Backup (commercial tool)

**Pros:**
- Official MySQL solution
- Fast binary backups
- Incremental backup support
- Hot backup without downtime

**Cons:**
- Commercial license required
- Not free/open source
- Not available to all users
- Violates our "free" principle

**Why not chosen:** Cadangkan must be free and open source. Enterprise Backup is commercial and not available to all users.

## Related Decisions

- ADR-0001: Use Go for Implementation (enables exec.Command)
- ADR-0002: MySQL Client Architecture (client provides database info)
- ADR-0006: Connection Pool Configuration (used for metadata collection)

## Notes

### mysqldump Flag Details

**`--single-transaction`**
- Creates consistent snapshot using InnoDB transaction
- No table locking required
- Only works with transactional storage engines (InnoDB)
- Essential for production backups

**`--quick`**
- Retrieves rows one at a time rather than buffering entire result
- Critical for low memory usage
- Enables streaming

**`--routines`**
- Includes stored procedures and functions
- Important for complete database backup

**`--triggers`**
- Includes triggers
- Default in newer versions but explicit for clarity

**`--events`**
- Includes event scheduler events
- Not included by default

**`--no-tablespaces`**
- Avoids tablespace commands that often cause permission errors
- Makes backup more portable

**`--skip-lock-tables`**
- Don't lock tables (use --single-transaction instead)
- Prevents blocking other operations

**`--set-gtid-purged=OFF`**
- Don't include GTID information
- Prevents issues when restoring to different servers

### Performance Characteristics

Based on testing and typical usage:

**Throughput:**
- ~50-100 MB/s for local databases
- ~10-50 MB/s for remote databases
- Depends on disk I/O and network speed

**Compression:**
- gzip provides ~60-70% compression
- 1GB database → ~300-400MB compressed
- Text data compresses better than binary data

**Memory Usage:**
- Cadangkan: <50MB
- mysqldump: <100MB
- Total: <150MB regardless of database size

**Duration:**
- 1GB database: 1-2 minutes
- 10GB database: 10-20 minutes
- 100GB database: 1.5-3 hours

### Streaming Pipeline

The complete streaming pipeline:

```
MySQL Server
    ↓ (TCP socket)
mysqldump process
    ↓ (stdout pipe)
io.TeeReader (for checksum)
    ↓
gzip.Writer (compression)
    ↓
os.File (output file)
```

All steps happen concurrently with minimal buffering.

### Future Enhancements

Potential improvements (not in MVP):

1. **Parallel Dumps:** Dump multiple tables concurrently for large databases
2. **Incremental Backups:** Use binary logs for incremental backups
3. **Compression Options:** Support zstd, lz4 for better compression/speed trade-offs
4. **Progress Tracking:** Parse mysqldump stderr for progress information
5. **Selective Backup:** Backup only changed tables
6. **Split Backups:** Split large backups into multiple files
7. **Encryption:** Encrypt backups at rest
8. **Verification:** Automatic backup verification after creation

### Testing Strategy

**Unit Tests:**
- Mock exec.Command for testing without mysqldump
- Test argument generation
- Test error handling
- Test stream processing

**Integration Tests:**
- Real mysqldump with test database
- Verify backup can be restored
- Test with different MySQL versions
- Performance benchmarks

### References

- [mysqldump Documentation](https://dev.mysql.com/doc/refman/8.0/en/mysqldump.html)
- [MySQL Backup and Recovery](https://dev.mysql.com/doc/refman/8.0/en/backup-and-recovery.html)
- [mysqldump Best Practices](https://dev.mysql.com/doc/refman/8.0/en/mysqldump-sql-format.html)
- [InnoDB and ACID Model](https://dev.mysql.com/doc/refman/8.0/en/mysql-acid.html)
- [MySQL Server-Tool Compatibility Matrix](https://dev.mysql.com/doc/mysql-compat-matrix/en/)
- [MySQL 8.0 Upgrade Guide](https://dev.mysql.com/doc/refman/8.0/en/upgrading-from-previous-series.html)
