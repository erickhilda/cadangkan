# 0008. CLI Architecture and User Interface Design

**Status:** Accepted  
**Date:** 2025-01-03

## Context

Cadangkan needs a command-line interface (CLI) to make database backups accessible and user-friendly. The CLI is the primary interface for the MVP and must balance simplicity, extensibility, and professional user experience.

**Requirements:**
- Simple, intuitive command structure for common operations
- Support for MySQL backups with various options (tables, compression, schema-only)
- Extensible to support multiple database types (PostgreSQL, MongoDB in future)
- Professional output with progress indicators and colored feedback
- Clear error messages to guide users
- Support both flag-based and config-based workflows
- Cross-platform compatibility (Linux, macOS, Windows)

**User Experience Goals:**
- One-line backup command: `cadangkan backup --host=... --database=...`
- Instant visual feedback with spinners and colored output
- Clear success/failure indication
- Informative result display (size, duration, checksum)
- Helpful error messages with installation instructions

**Technical Constraints:**
- Must integrate with existing backup service (internal/backup)
- Should prepare for config file support (Week 2)
- Need to check mysqldump availability before attempting backup
- Must work in terminals with and without color support

## Decision

We will implement the CLI using **urfave/cli/v2** with a **direct command structure** and **ANSI-colored output**.

### 1. CLI Framework: urfave/cli/v2

Use `github.com/urfave/cli/v2` for command parsing and application structure.

**Rationale:**
- Middle ground between simplicity and features
- Good flag handling with required/optional flags
- Built-in help generation
- Command and subcommand support
- Smaller than Cobra, more structured than standard flag package

### 2. Command Structure: Direct Commands with Type Flag

**Command Pattern:**
```bash
cadangkan backup --type=mysql --host=... --database=...
```

Not:
```bash
cadangkan backup mysql --host=... --database=...  # Rejected
```

**Structure:**
- Main command: `cadangkan`
- Direct command: `backup`
- Database type via flag: `--type=mysql` (default: mysql)
- Connection flags: `--host`, `--port`, `--user`, `--password`, `--database`
- Backup options: `--tables`, `--exclude-tables`, `--schema-only`, `--compression`, `--output`

**Rationale:**
- Simpler command structure (fewer levels of nesting)
- Easier to add new database types (just another value for --type)
- Cleaner help output
- Natural for users familiar with standard tools (git, docker)
- Prepares for config file: `cadangkan backup [connection-name]`

### 3. User Interface Design

**Color-Coded Output:**
- Success: Green checkmark (✓)
- Error: Red X (✗)
- Info: Blue info icon (ℹ)
- Warning: Yellow warning icon (⚠)
- Highlights: Cyan for labels

**Progress Indicators:**
- Spinner animation during backup: ⠋ ⠙ ⠹ ⠸ ⠼ ⠴ ⠦ ⠧ ⠇ ⠏
- Simple, non-intrusive
- Automatically cleared on completion

**Result Display:**
```
✓ Backup completed!

  Backup ID:   2026-01-03-190000
  Database:    northwind
  File:        ~/.cadangkan/backups/northwind/2026-01-03-190000.sql.gz
  Size:        2.4 MB
  Duration:    3.2s
  Checksum:    sha256:abc123...

Backup saved to: ~/.cadangkan/backups/northwind/2026-01-03-190000.sql.gz
```

**Error Messages:**
- Clear problem statement
- Actionable next steps
- Installation instructions when tools missing

### 4. Configuration Hierarchy (Phase 1 + Prep for Phase 2)

**Priority Order:**
1. Command-line flags (highest priority)
2. Config file entries (Week 2) - prepared but not implemented
3. Environment variables (future)
4. Defaults

**Week 1 (MVP):** Flags only
**Week 2:** Add config file support at `~/.cadangkan/config.yaml`

### 5. Code Organization

**Files:**
- `cmd/cadangkan/main.go` - Application entry point, version, commands
- `cmd/cadangkan/backup.go` - Backup command implementation
- `cmd/cadangkan/shared.go` - Shared utilities (colors, formatting, progress)

**Utilities Approach:**
- Reuse existing functions from `internal/backup` (FormatBytes, FormatDuration, CheckMySQLDump)
- Only create CLI-specific utilities (colored output, progress spinner)
- Keep CLI layer thin - business logic stays in internal packages

### 6. Host Configuration

**Default host:** `127.0.0.1` (not `localhost`)

**Rationale:**
- `localhost` triggers Unix socket connection in mysqldump
- Docker containers don't expose Unix sockets to host
- `127.0.0.1` forces TCP/IP connection
- Works reliably with Docker and local MySQL
- Document this clearly in help text

## Consequences

### Positive

- **Simple User Experience:** One-line backup command with clear output
- **Professional Appearance:** Colored output and spinners match modern CLI tools
- **Extensible:** Easy to add new database types with --type flag
- **Discoverable:** Built-in help (`--help`) at every level
- **Clear Feedback:** Users immediately see success/failure with visual indicators
- **Thin CLI Layer:** Business logic in internal packages, CLI is just interface
- **Future-Ready:** Structure prepared for config file support (Week 2)
- **Cross-Platform:** Works on Linux, macOS, Windows
- **Standard Conventions:** Follows Unix CLI patterns (flags, exit codes)

### Negative

- **External Dependency:** Requires urfave/cli/v2 package
- **ANSI Colors:** May not work in all terminals (PowerShell, some CI environments)
- **Learning Curve:** Developers need to learn urfave/cli patterns
- **Less Flexible:** Can't easily change command structure later without breaking changes
- **Spinner Complexity:** Progress indicator requires goroutine management

### Risks

- **Terminal Compatibility:** ANSI colors fail in some environments
- **Unicode Characters:** Spinner/emoji might not render on all systems
- **Framework Changes:** urfave/cli/v2 API changes could break compatibility
- **Command Structure:** If we need hierarchical subcommands later, current structure may be limiting

**Mitigations:**
- Detect color support and fall back to plain text
- Provide `--no-color` flag for CI/CD environments
- Version lock urfave/cli dependency
- Current structure can be extended with additional commands without breaking existing ones
- Document command structure decisions in this ADR

## Alternatives Considered

### Alternative 1: Cobra CLI Framework

**Description:** Use `github.com/spf13/cobra` - the most popular Go CLI framework

**Pros:**
- Most popular (used by kubectl, docker, hugo)
- Excellent documentation and community
- Powerful command hierarchy support
- Built-in generators for commands
- Persistent flags across commands

**Cons:**
- Heavier dependency (~20+ packages)
- More complex than needed for MVP
- Steeper learning curve
- Over-engineered for simple tool
- Verbose command setup

**Why not chosen:** Too heavy for our needs. Cadangkan's MVP is simple (backup, restore, list), and Cobra's advanced features (nested subcommands, plugins) aren't required. urfave/cli provides 80% of Cobra's benefits with 20% of the complexity.

### Alternative 2: Standard flag Package

**Description:** Use Go's built-in `flag` package with custom command routing

**Pros:**
- No external dependencies
- Simple and lightweight
- Complete control over parsing
- Well-documented (stdlib)

**Cons:**
- Manual help text generation
- No built-in command/subcommand support
- Manual validation and error handling
- Less discoverable (no standard help patterns)
- More boilerplate code

**Why not chosen:** Too much manual work for common patterns. Would need to implement command routing, help generation, validation, etc. Time better spent on features. urfave/cli provides these for free.

### Alternative 3: Subcommand Structure (mysql as subcommand)

**Description:** Use `cadangkan backup mysql` instead of `cadangkan backup --type=mysql`

**Pros:**
- More explicit database type
- Natural grouping: `cadangkan backup mysql`, `cadangkan backup postgres`
- Follows pattern of tools like `docker` and `kubectl`
- Each database type can have specific flags

**Cons:**
- Deeper command hierarchy (3 levels)
- More typing for simple operations
- Doesn't work well with named connections (config file)
- Less consistent with config-based approach
- Example: `cadangkan backup mysql production` vs `cadangkan backup production`

**Why not chosen:** Conflicts with config file approach. When using saved connections (Week 2), users want `cadangkan backup production`, not `cadangkan backup mysql production`. The --type flag can be inferred from config or default to mysql. Makes the tool more flexible for both flag-based and config-based workflows.

### Alternative 4: Positional Arguments Instead of Flags

**Description:** Use `cadangkan backup host user database` instead of flags

**Pros:**
- Shorter commands
- Follows some Unix tools (rsync, scp)
- Less typing

**Cons:**
- Order matters (error-prone)
- Not self-documenting
- Hard to add optional parameters
- Doesn't scale with more options
- Poor discoverability

**Why not chosen:** Flags are more explicit, self-documenting, and flexible. Order doesn't matter, optional parameters are natural, and help text clearly shows what each flag does. Better user experience.

### Alternative 5: Interactive Prompts

**Description:** Prompt user for connection details instead of flags

**Pros:**
- Guided experience for beginners
- No need to remember flags
- Can validate inputs interactively
- Nice for first-time setup

**Cons:**
- Not scriptable (breaks automation)
- Slower for experienced users
- Doesn't work in CI/CD pipelines
- Extra code complexity
- Still need flag support for automation

**Why not chosen:** Breaks automation. Backups need to run in cron jobs, CI/CD, scripts. Interactive prompts prevent this. Flags are universal, work everywhere. Can add interactive mode as optional feature later without breaking flag-based approach.

### Alternative 6: No Colors or Spinners (Plain Text)

**Description:** Use plain text output without ANSI colors or animations

**Pros:**
- Universal compatibility (all terminals)
- Simpler implementation
- No terminal capability detection needed
- Works in all CI/CD systems
- Easier to test

**Cons:**
- Less visually appealing
- Harder to scan output
- No instant visual feedback
- Feels dated compared to modern CLIs
- Less professional appearance

**Why not chosen:** Modern users expect colored, animated CLIs (npm, cargo, docker all use colors). Professional appearance matters for adoption. Can detect color support and fall back to plain text. Benefits outweigh complexity.

## Related Decisions

- ADR-0001: Use Go for Implementation (Go's CLI ecosystem)
- ADR-0007: mysqldump Backup Strategy (CLI calls backup service)
- Future ADR: Configuration File Format (Week 2)

## Notes

### Terminal Compatibility

**Color Support Detection:**
- Check `TERM` environment variable
- Look for `NO_COLOR` environment variable
- Provide `--no-color` flag override
- Graceful fallback to plain text

**Unicode Support:**
- Spinner characters (⠋ ⠙ etc.) are Braille patterns
- Checkmarks (✓ ✗) are Unicode symbols
- May not render on all systems
- Consider ASCII fallbacks for Windows CMD

### Implementation Details

**File Structure:**
```
cmd/cadangkan/
├── main.go     # App setup, version, command registration
├── backup.go   # Backup command implementation
└── shared.go   # CLI utilities (colors, spinner, formatting)
```

**Shared Utilities (Reuse from internal/backup):**
- ✓ `backup.FormatBytes()` - Already exists
- ✓ `backup.FormatDuration()` - Already exists  
- ✓ `backup.CheckMySQLDump()` - Already exists

**New CLI-Specific Utilities:**
- `printSuccess()`, `printError()`, `printInfo()` - Colored output
- `showSpinner()` - Progress animation
- `formatBackupResult()` - Result display
- `getConfigPath()` - Config file path (prep for Week 2)

### Future Enhancements

**Week 2 (Config File Support):**
- Add `cadangkan add mysql production --host=...` to save connections
- Support `cadangkan backup production` using saved config
- Encrypt credentials in config file

**Week 3 (More Commands):**
- `cadangkan list` - List saved backups
- `cadangkan restore` - Restore from backup
- `cadangkan schedule` - Set up cron jobs

**Future (Advanced):**
- `--quiet` flag for script usage
- `--json` flag for machine-readable output
- `--progress` flag for detailed progress (table count, data size)
- Interactive mode with prompts
- Shell completion (bash, zsh, fish)

### Testing Strategy

**Manual Testing:**
- ✓ Test on Linux, macOS, Windows
- ✓ Test with color terminals (iTerm, GNOME Terminal)
- ✓ Test without colors (CI, pipe to file)
- ✓ Test help text at all levels
- ✓ Test error handling (missing mysqldump, bad credentials)

**Automated Testing:**
- Integration tests for CLI commands
- Mock backup service for fast tests
- Capture stdout/stderr for validation
- Test exit codes (0 = success, 1 = error)

### Command Examples

**Basic Backup:**
```bash
cadangkan backup --host=127.0.0.1 --user=root --password=secret --database=mydb
```

**With Options:**
```bash
# Specific tables
cadangkan backup --host=127.0.0.1 --user=root --password=secret \
  --database=mydb --tables=users,orders

# Schema only
cadangkan backup --host=127.0.0.1 --user=root --password=secret \
  --database=mydb --schema-only

# Custom output
cadangkan backup --host=127.0.0.1 --user=root --password=secret \
  --database=mydb --output=/backups/manual
```

**Future (Named Connections):**
```bash
# Save connection (Week 2)
cadangkan add mysql production --host=db.prod.com --user=backup --password=...

# Use saved connection
cadangkan backup production
```

### References

- [urfave/cli Documentation](https://cli.urfave.org/)
- [The Art of Command Line](https://github.com/jlevy/the-art-of-command-line)
- [12 Factor CLI Apps](https://medium.com/@jdxcode/12-factor-cli-apps-dd3c227a0e46)
- [CLI Guidelines](https://clig.dev/)
- [ANSI Escape Codes](https://en.wikipedia.org/wiki/ANSI_escape_code)
