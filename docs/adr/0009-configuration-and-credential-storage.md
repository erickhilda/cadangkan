# ADR-0009: Configuration Management and Credential Storage

**Status:** Accepted  
**Date:** 2025-01-04

## Context

Cadangkan requires users to provide database credentials (host, port, user, password, database name) for backup operations. Initially, the CLI required all credentials to be passed as flags for every backup command, which created several pain points:

1. **Poor User Experience:** Users had to remember and type multiple flags for each backup
2. **Security Concerns:** Passwords in command-line arguments appear in shell history and process lists
3. **Workflow Friction:** No way to save and reuse database configurations
4. **Error-Prone:** Typing credentials repeatedly increases the chance of mistakes
5. **Script Complexity:** Automated backup scripts became verbose and hard to maintain

We needed a configuration management system that would:
- Securely store database credentials
- Allow users to manage multiple database configurations
- Maintain backward compatibility with the existing flag-based approach
- Be simple enough for MVP while extensible for future enhancements
- Follow security best practices for credential storage

Key constraints:
- Target users range from individual developers to small teams
- Must work cross-platform (Linux, macOS, Windows)
- Should not require external dependencies or services
- Implementation timeline: 1-2 days for MVP
- Must be production-ready with proper testing

## Decision

We will implement a configuration management system with the following architectural decisions:

### 1. Configuration Format: YAML

Store configuration in a human-readable YAML file at `~/.cadangkan/config.yaml`.

**Structure:**
```yaml
version: "1.0"
databases:
  production:
    type: mysql
    host: mysql.example.com
    port: 3306
    database: myapp
    user: backup_user
    password_encrypted: "base64-encrypted-string"
```

### 2. Credential Encryption: AES-256-GCM with Local Key

Encrypt passwords using AES-256-GCM before storing in the config file. The encryption key is:
- Automatically generated on first use
- Stored in `~/.cadangkan/.key` with 0600 permissions
- 32 bytes (256 bits) of cryptographically secure random data
- Used for all password encryption/decryption

### 3. Password Storage: In Main Config File

Store encrypted passwords directly in the config YAML file rather than in a separate credentials file.

### 4. Dual Mode Support: Named and Direct

Support both configuration-based and flag-based backup commands:
- **Named mode:** `cadangkan backup production` (loads from config)
- **Direct mode:** `cadangkan backup --host=... --user=...` (uses flags)
- **Hybrid mode:** Named config with flag overrides

### 5. CLI Commands for Configuration Management

Provide dedicated commands:
- `cadangkan add mysql <name>` - Add database with connection testing
- `cadangkan list` - List all configured databases
- `cadangkan test <name>` - Test database connection
- `cadangkan remove <name>` - Remove database configuration

### 6. Security Measures

Implement multiple security layers:
- Interactive password prompts (no shell history exposure)
- Password-from-stdin support for scripts
- File permissions: 0600 for config and key, 0700 for directory
- Connection testing before saving credentials
- No plaintext passwords anywhere in the codebase

### 7. MVP Scope: MySQL Only, Local Storage

Focus the MVP on:
- MySQL database type only (extensible architecture for future types)
- Local file storage (no cloud sync in MVP)
- Simple AES-256-GCM encryption (no system keyring integration yet)

## Consequences

### Positive

1. **Improved User Experience**
   - One-command backups: `cadangkan backup prod` vs multi-flag commands
   - No need to remember or manage credentials manually
   - Clear, intuitive CLI with `add`, `list`, `test`, `remove` commands

2. **Enhanced Security**
   - Passwords encrypted at rest with industry-standard AES-256-GCM
   - No credentials in shell history or process lists
   - Secure file permissions prevent unauthorized access
   - Interactive prompts keep passwords out of scripts

3. **Backward Compatibility**
   - Existing flag-based workflows continue to work
   - Gradual migration path for users
   - No breaking changes to the CLI

4. **Developer Productivity**
   - Easier to manage multiple database configurations
   - Simplified backup scripts and automation
   - Consistent configuration across environments

5. **Maintainability**
   - Clean separation of concerns (config package)
   - Extensible architecture for future database types
   - Well-tested (68.9% coverage, 28 tests)

6. **Cross-Platform**
   - Works on Linux, macOS, and Windows
   - No external dependencies
   - Pure Go implementation

### Negative

1. **Additional Complexity**
   - More code to maintain (5 implementation files + 3 test files)
   - Two modes of operation to support and document
   - Configuration file format to version and migrate

2. **Key Management Responsibility**
   - Users must protect the `.key` file
   - Lost key means re-adding all databases
   - No automatic key backup or recovery

3. **Local-Only Storage**
   - Config doesn't sync across machines
   - Users must manually backup config files
   - Team collaboration requires manual config sharing

4. **Limited Password Input Methods**
   - No integration with password managers
   - No system keyring support (macOS Keychain, etc.)
   - Manual copy-paste for complex passwords

5. **Single Database Type**
   - MVP only supports MySQL
   - PostgreSQL users must wait for Phase 2
   - Architecture is extensible but needs implementation

### Risks

1. **Key File Loss**
   - Risk: Users lose `.key` file and cannot decrypt passwords
   - Mitigation: Clear documentation on key importance and backup
   - Recovery: Users can re-add databases if key is lost

2. **File Permission Issues**
   - Risk: Incorrect permissions expose credentials
   - Mitigation: Automatic permission setting (0600/0700)
   - Detection: Validate permissions on load

3. **Configuration Corruption**
   - Risk: Manual YAML edits break configuration
   - Mitigation: Validation before saving, clear error messages
   - Recovery: Users can fix YAML or use CLI to recreate

4. **Cross-Platform Differences**
   - Risk: Path or permission handling differs by OS
   - Mitigation: Use Go's standard library abstractions
   - Testing: Verify on Linux, macOS, Windows

## Alternatives Considered

### Alternative 1: System Keyring Integration

**Description:** Use OS-native credential storage (macOS Keychain, Windows Credential Manager, Linux Secret Service)

**Pros:**
- Better security with OS-managed encryption
- Integration with system security policies
- Familiar to users (same as browser passwords)
- Automatic backup on some platforms

**Cons:**
- Complex implementation with platform-specific code
- Multiple dependencies (keyring libraries)
- Harder to test across platforms
- Longer development time (3-5 days vs 1-2 days)
- Some Linux systems lack Secret Service

**Why not chosen:** Too complex for MVP. We can add this as an optional enhancement in a future release while keeping the simple local encryption as default.

### Alternative 2: Environment Variables Only

**Description:** Store credentials in environment variables, no config file

**Pros:**
- Simple implementation
- Common pattern in 12-factor apps
- No file management needed
- Easy for containerized deployments

**Cons:**
- No persistent storage (lost on reboot)
- Difficult to manage multiple databases
- Environment pollution with many configs
- No encryption at rest
- Exposed in process environment

**Why not chosen:** Doesn't solve the core problem of managing multiple database configurations persistently.

### Alternative 3: Separate Credentials File

**Description:** Store encrypted passwords in separate file (`~/.cadangkan/credentials.enc`)

**Pros:**
- Cleaner separation of concerns
- Can set different permissions on credentials
- Easier to backup config without passwords
- Slightly better security through obscurity

**Cons:**
- Two files to manage instead of one
- More complex file operations
- Harder to backup complete configuration
- Minimal security benefit over single file

**Why not chosen:** Added complexity with minimal benefit. Single encrypted config file is simpler to manage and backup.

### Alternative 4: Plaintext Configuration

**Description:** Store passwords in plaintext YAML file

**Pros:**
- Simplest implementation
- No encryption overhead
- Easy debugging and manual editing
- Smaller codebase

**Cons:**
- **Completely insecure** - passwords visible to anyone
- Bad security practice and example
- Violates most security policies
- Cannot recommend for production use
- High risk of credential leakage

**Why not chosen:** Unacceptable security risk. Encryption is a must-have, not optional.

### Alternative 5: Remote Configuration Service

**Description:** Store configs in cloud service (AWS Secrets Manager, HashiCorp Vault)

**Pros:**
- Enterprise-grade security
- Centralized management
- Audit logging
- Rotation policies
- Team collaboration

**Cons:**
- Requires internet connectivity
- External dependency and cost
- Complex setup for individual users
- Overkill for target audience (indie devs, small teams)
- Long development timeline

**Why not chosen:** Too complex and costly for MVP target users. Better suited for Phase 3 enterprise features.

## Related Decisions

- **ADR-0008:** CLI Architecture and User Interface Design - Defines the command structure that the config commands follow
- **ADR-0001:** Use Go for Implementation - Go's standard library made encryption and file handling straightforward
- **ADR-0002:** MySQL Client Architecture - The DatabaseConfig integrates with the MySQL client

## Notes

### Implementation Details

The configuration system was implemented in `internal/config/` package with:
- `types.go` - Config and DatabaseConfig structures
- `loader.go` - YAML loading/saving and database management
- `encryption.go` - AES-256-GCM password encryption
- `validation.go` - Configuration validation logic
- `errors.go` - Custom error types for better UX

Test coverage: 68.9% with 28 comprehensive unit tests covering encryption, validation, and file operations.

### Future Enhancements

Potential improvements for future releases (not in MVP):
1. System keyring integration as optional backend
2. Configuration backup/restore commands
3. Config export/import for team sharing
4. Environment file (.env) support
5. Encryption key rotation
6. PostgreSQL and other database types
7. Cloud configuration sync
8. Configuration templates

### Security Audit Notes

The encryption implementation uses:
- `crypto/aes` - Standard library AES cipher
- `crypto/cipher` - GCM mode for authenticated encryption
- `crypto/rand` - Cryptographically secure random generation
- Random nonce per encryption operation
- Base64 encoding for text storage

This approach is secure for the target use case (local development and small teams) but should be audited before enterprise deployment.

### Migration Path

For users upgrading from pre-config versions:
1. Old flag-based commands continue to work
2. Users can gradually migrate to config-based approach
3. No forced migration or breaking changes
4. Clear documentation guides migration process

### Documentation

Complete documentation provided in:
- `docs/CONFIGURATION.md` - User guide with examples
- `docs/CONFIG_IMPLEMENTATION_SUMMARY.md` - Implementation details
- Updated `README.md` - Quick start guide
- Command help text in CLI

### References

- NIST AES-GCM Specification: https://csrc.nist.gov/publications/detail/sp/800-38d/final
- Go crypto/cipher Documentation: https://pkg.go.dev/crypto/cipher
- YAML v3 Specification: https://yaml.org/spec/1.2/spec.html
- 12-Factor App Config: https://12factor.net/config
