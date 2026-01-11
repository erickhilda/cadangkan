# Configuration System

The configuration system allows you to save database credentials and manage multiple database connections, eliminating the need to pass connection flags with every backup command.

## Quick Start

### Adding a Database

Add a database configuration with interactive password prompt:

```bash
cadangkan add --host=mysql.example.com --user=backup_user --database=myapp mysql production
```

Or use `--password-stdin` to read password from stdin:

```bash
echo "mypassword" | cadangkan add \
  --host=mysql.example.com \
  --user=backup_user \
  --database=myapp \
  --password-stdin \
  mysql production
```

### Listing Databases

View all configured databases:

```bash
cadangkan list
```

### Testing Connection

Test a database connection:

```bash
cadangkan test production
```

### Using Saved Configuration

Once configured, backup using the saved configuration:

```bash
cadangkan backup production
```

You can still override config values with flags:

```bash
cadangkan backup production --database=other_db
```

### Editing a Database

Edit an existing database configuration:

```bash
# Update host
cadangkan edit --host=newhost.example.com production

# Update password (interactive prompt)
cadangkan edit --password production

# Update multiple fields
cadangkan edit --host=newhost --port=3307 production
```

**Note:** Flags come first, followed by the database name. Only specified fields will be updated.

### Removing a Database

Remove a database configuration:

```bash
cadangkan remove production
```

## Configuration Files

The configuration system uses two files in `~/.cadangkan/`:

- **`config.yaml`** - Stores database configurations with encrypted passwords (0600 permissions)
- **`.key`** - Encryption key for passwords (0600 permissions)

### Example config.yaml

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
  
  staging:
    type: mysql
    host: staging.example.com
    port: 3306
    database: myapp_staging
    user: backup_user
    password_encrypted: "base64-encrypted-string"
```

## Security

### Password Encryption

Passwords are encrypted using **AES-256-GCM** before being stored in the config file. The encryption key is automatically generated on first use and stored in `~/.cadangkan/.key` with restricted permissions (0600).

### File Permissions

- Config file: `0600` (read/write for owner only)
- Encryption key: `0600` (read/write for owner only)
- Config directory: `0700` (full access for owner only)

### Best Practices

1. **Use dedicated backup users** with minimal required privileges
2. **Keep the `.key` file secure** - if lost, you'll need to re-add all databases
3. **Back up your config** to a secure location if managing many databases
4. **Use `--password-stdin`** in scripts instead of `--password` flag

## Usage Modes

The backup command supports two modes:

### Named Mode (Recommended)

Load connection details from saved configuration:

```bash
cadangkan backup production
```

Benefits:
- No need to remember or pass credentials
- Consistent configuration across backups
- More secure (no credentials in shell history)

### Direct Mode

Pass all connection details via flags:

```bash
cadangkan backup \
  --host=mysql.example.com \
  --port=3306 \
  --user=backup_user \
  --password=secret \
  --database=myapp
```

Use this for:
- One-off backups
- Testing different connections
- Automated scripts with environment variables

### Hybrid Mode

Combine both approaches - use saved config but override specific values:

```bash
# Use production config but backup a different database
cadangkan backup production --database=other_db

# Use production config but connect to different host
cadangkan backup production --host=localhost --port=3307
```

## Commands Reference

### add

Add a new database configuration:

```bash
cadangkan add [flags] mysql <name>
```

**Required flags:**
- `--host` - Database host
- `--user` - Database user
- `--database` - Database name

**Optional flags:**
- `--port` - Database port (default: 3306)
- `--password` - Database password (prefer interactive prompt)
- `--password-stdin` - Read password from stdin
- `--skip-test` - Skip connection test

**Examples:**

```bash
# Interactive password prompt (recommended)
cadangkan add --host=db.example.com --user=backup --database=myapp mysql prod

# Read password from stdin
echo "$DB_PASSWORD" | cadangkan add \
  --host=db.example.com --user=backup --database=myapp --password-stdin \
  mysql prod

# Skip connection test (faster, but doesn't verify credentials)
cadangkan add --host=db.example.com --user=backup \
  --database=myapp --skip-test \
  mysql prod
```

### list

List all configured databases:

```bash
cadangkan list
```

Aliases: `ls`

Output example:
```
Configured Databases
================================================================================
NAME                 TYPE       HOST                           DATABASE
production           mysql      mysql.example.com:3306         myapp
staging              mysql      staging.example.com:3306       myapp_staging

Total: 2 database(s)
```

### test

Test connection to a configured database:

```bash
cadangkan test <name>
```

**Example:**

```bash
cadangkan test production
```

Output:
```
ℹ Loading configuration for 'production'...
ℹ Testing connection to backup_user@mysql.example.com:3306...
✓ Connected successfully (MySQL 8.0.35)

  Database: myapp
  Size:     1.2 GB
```

### remove

Remove a database configuration:

```bash
cadangkan remove <name> [flags]
```

Aliases: `rm`

**Optional flags:**
- `--force, -f` - Skip confirmation prompt

**Examples:**

```bash
# With confirmation prompt
cadangkan remove production

# Skip confirmation
cadangkan remove production --force
```

### edit

Edit an existing database configuration:

```bash
cadangkan edit [flags] <name>
```

**Important:** Flags come first, followed by the database name.

**Optional flags:**
- `--host` - Update database host
- `--port` - Update database port
- `--user` - Update database user
- `--database` - Update database name
- `--password` - Update password (triggers interactive prompt if no value provided)
- `--password-stdin` - Read password from stdin
- `--skip-test` - Skip connection test after update

**Behavior:**
- Only the fields specified via flags will be updated
- All other fields remain unchanged (partial update)
- Connection is tested after update unless `--skip-test` is used
- Password is re-encrypted if changed
- When using `--password` without a value, you'll be prompted to enter the password interactively

**Examples:**

```bash
# Update host only
cadangkan edit --host=newhost.example.com production

# Update multiple fields
cadangkan edit --host=newhost --port=3307 production

# Update password (interactive prompt - recommended)
cadangkan edit --password production

# Update password with value (not recommended for security)
cadangkan edit --password=mypassword production

# Update password from stdin
echo "newpassword" | cadangkan edit --password-stdin production

# Update without connection test
cadangkan edit --host=newhost --skip-test production
```

**Note:** For security, prefer using `--password` (interactive prompt) or `--password-stdin` instead of passing the password directly via `--password=value`.

### backup

Create a backup (supports both named and direct mode):

```bash
cadangkan backup [name] [flags]
```

See main documentation for full backup command reference.

### backup-list

List all backups for configured databases:

```bash
cadangkan backup-list [database-name] [flags]
```

Aliases: `backups`

**Optional flags:**
- `--format` - Output format: `table` (default) or `json`

**Behavior:**
- If database name is provided, lists backups for that database only
- If no database name is provided, lists backups for all configured databases
- Backups are sorted by creation date (newest first)
- Shows backup ID, date, size, and status

**Examples:**

```bash
# List all backups for all databases
cadangkan backup-list

# List backups for specific database
cadangkan backup-list production

# Output in JSON format
cadangkan backup-list production --format=json

# Using alias
cadangkan backups production
```

**Output example:**

```
Backups for production
================================================================================
BACKUP ID              DATE                 SIZE      STATUS
2026-01-08-133600      2026-01-08 13:36:00  2.4 MB    completed
2026-01-07-120000      2026-01-07 12:00:00  2.3 MB    completed
2026-01-06-120000      2026-01-06 12:00:00  2.2 MB    completed

Total: 3 backup(s)
```

## Troubleshooting

### "Database not found in config"

```bash
cadangkan backup production
Error: database 'production' not found in config
```

**Solution:** Add the database first with `cadangkan add`

```bash
cadangkan add --host=... --user=... --database=... mysql production
```

### "Failed to decrypt password"

This usually means the encryption key has been lost or corrupted.

**Solution:** Remove and re-add the database:

```bash
cadangkan remove production --force
cadangkan add --host=... --user=... --database=... mysql production
```

### "Connection failed"

```bash
cadangkan test production
Error: connection test failed: dial tcp: connect: connection refused
```

**Possible causes:**
1. Database server is not running
2. Host or port is incorrect
3. Firewall blocking connection
4. Credentials are wrong

**Solution:** Update the configuration using the `edit` command:

```bash
# Update host
cadangkan edit production --host=correct-host

# Or update multiple fields
cadangkan edit production --host=correct-host --port=3307

# Or remove and re-add (if many fields need changing)
cadangkan remove production --force
cadangkan add --host=correct-host --user=... --database=... mysql production
```

### Lost encryption key

If you lose the `.key` file, you'll need to re-add all databases.

**Prevention:** Back up `~/.cadangkan/.key` securely

**Recovery:**
1. Delete old config: `rm ~/.cadangkan/config.yaml ~/.cadangkan/.key`
2. Re-add all databases

## Migration from Direct Mode

If you've been using direct mode (passing all flags), you can migrate to named mode:

**Before:**
```bash
cadangkan backup \
  --host=mysql.example.com \
  --port=3306 \
  --user=backup_user \
  --password=secret \
  --database=myapp
```

**After:**
```bash
# One-time setup
cadangkan add \
  --host=mysql.example.com \
  --user=backup_user \
  --database=myapp \
  mysql production
# Password: [enter interactively]

# Future backups
cadangkan backup production
```

## Environment Variables

You can use environment variables for the add command:

```bash
export DB_HOST="mysql.example.com"
export DB_USER="backup_user"
export DB_NAME="myapp"

cadangkan add \
  --host=$DB_HOST \
  --user=$DB_USER \
  --database=$DB_NAME \
  mysql production
```

## Next Steps

- **Scheduling:** Set up automatic backups (coming in Phase 2)
- **Cloud Storage:** Upload backups to S3/GCS (coming in Phase 3)
- **Retention Policies:** Automatic cleanup of old backups (coming in Phase 2)

## See Also

- [README.md](../README.md) - Main documentation
- [Backup Documentation](./BACKUP.md) - Detailed backup guide
- [Architecture Decision Records](./adr/) - Technical decisions
