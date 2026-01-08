# Troubleshooting Guide

## Backup Size Issues

### Problem: Backup is suspiciously small (e.g., 200KB instead of expected 30GB)

This usually indicates that only the database schema (structure) was backed up, not the actual data. This can happen due to:

1. **Missing SELECT permissions** on tables
2. **Empty tables** (schema exists but no data)
3. **Incorrect database name** or connection
4. **mysqldump warnings** that were ignored

### How to Check MySQL User Permissions

Connect to your MySQL server and run these commands:

#### 1. Check current user and permissions

```sql
-- Show current user
SELECT USER(), CURRENT_USER();

-- Show all grants for current user
SHOW GRANTS;

-- Show grants for a specific user
SHOW GRANTS FOR 'your_username'@'your_host';
```

#### 2. Check if user has SELECT permission on the database

```sql
-- Check table-level permissions
SELECT 
    TABLE_SCHEMA,
    TABLE_NAME,
    PRIVILEGE_TYPE
FROM 
    information_schema.TABLE_PRIVILEGES
WHERE 
    GRANTEE = CONCAT('''', USER(), '''')
    AND TABLE_SCHEMA = 'YOUR_DATABASE_NAME'
    AND PRIVILEGE_TYPE = 'SELECT';
```

#### 3. Check database size and table sizes

```sql
-- Check total database size
SELECT 
    table_schema AS 'Database',
    ROUND(SUM(data_length + index_length) / 1024 / 1024 / 1024, 2) AS 'Size (GB)'
FROM 
    information_schema.TABLES 
WHERE 
    table_schema = 'YOUR_DATABASE_NAME'
GROUP BY 
    table_schema;

-- Check individual table sizes
SELECT 
    table_name AS 'Table',
    ROUND(((data_length + index_length) / 1024 / 1024), 2) AS 'Size (MB)',
    table_rows AS 'Rows'
FROM 
    information_schema.TABLES
WHERE 
    table_schema = 'YOUR_DATABASE_NAME'
ORDER BY 
    (data_length + index_length) DESC;
```

#### 4. Check if tables have data

```sql
-- Count rows in each table
SELECT 
    table_name,
    table_rows
FROM 
    information_schema.TABLES
WHERE 
    table_schema = 'YOUR_DATABASE_NAME'
    AND table_type = 'BASE TABLE'
ORDER BY 
    table_rows DESC;
```

### Fixing Permission Issues

#### Grant SELECT permission to user

```sql
-- Grant SELECT on all tables in a database
GRANT SELECT ON your_database_name.* TO 'your_username'@'your_host';

-- Grant SELECT on specific table
GRANT SELECT ON your_database_name.your_table_name TO 'your_username'@'your_host';

-- Grant all necessary permissions for backup
GRANT SELECT, LOCK TABLES, SHOW VIEW, EVENT, TRIGGER ON your_database_name.* 
TO 'your_username'@'your_host';

-- Apply changes
FLUSH PRIVILEGES;
```

#### Required Permissions for mysqldump

For a successful backup, the MySQL user needs:

- **SELECT** - To read table data
- **LOCK TABLES** - For consistent backups (if not using --single-transaction)
- **SHOW VIEW** - To backup views
- **TRIGGER** - To backup triggers
- **EVENT** - To backup scheduled events
- **PROCESS** - To see all processes (optional, for monitoring)

### Testing mysqldump Manually

Test mysqldump directly to see what errors it reports:

```bash
mysqldump \
  --host=YOUR_HOST \
  --port=YOUR_PORT \
  --user=YOUR_USER \
  --password=YOUR_PASSWORD \
  --single-transaction \
  --quick \
  --skip-lock-tables \
  --no-tablespaces \
  --set-gtid-purged=OFF \
  --routines \
  --triggers \
  --events \
  YOUR_DATABASE_NAME > test_backup.sql 2> mysqldump_errors.txt

# Check the error output
cat mysqldump_errors.txt

# Check backup size
ls -lh test_backup.sql
```

### Common Error Messages

- **"Access denied for user"** - User doesn't have required permissions
- **"Got error: 1142 when using LOCK TABLES"** - Missing LOCK TABLES permission
- **"mysqldump: Couldn't execute 'SHOW CREATE TABLE'"** - Missing SELECT permission
- **"Warning: Skipping table"** - Table might be a view or have permission issues

### Verifying Backup Content

Check if your backup contains data or just schema:

```bash
# Decompress if needed
gunzip backup_file.sql.gz

# Check for INSERT statements (data)
grep -c "INSERT INTO" backup_file.sql

# Check for CREATE TABLE statements (schema)
grep -c "CREATE TABLE" backup_file.sql

# If INSERT count is 0 or very low, only schema was backed up
```

## Other Common Issues

### Connection Issues

```bash
# Test connection
mysql -h YOUR_HOST -P YOUR_PORT -u YOUR_USER -p YOUR_DATABASE

# Check if mysqldump is available
mysqldump --version
```

### Disk Space Issues

```bash
# Check available disk space
df -h ~/.cadangkan/backups

# Check backup directory size
du -sh ~/.cadangkan/backups/*
```
