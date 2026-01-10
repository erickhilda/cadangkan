# Cadangkan

> **Sleep soundly knowing your databases are backed up**

Cadangkan (Indonesian for "backup") is a universal database backup and synchronization tool that makes database protection effortless, affordable, and accessible to every developer.

## 🚧 Development Status

**Current Phase:** Phase 0 - Project Setup

This project is in early development. We're currently setting up the foundation and working towards our MVP release (v0.1.0) which will support MySQL backup and restore functionality.

## 🎯 Vision

Cadangkan aims to provide:

- **Universal Backup Automation** - Works with MySQL, PostgreSQL, MongoDB, and more
- **Free & Open Source** - No subscription costs, full control over your data
- **Simple Interface** - One command to backup any database
- **Automated Scheduling** - Set it and forget it
- **Multi-Storage Support** - Local and cloud storage options

## 🛠️ Development Setup

### Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose (for local testing)
- Git
- MySQL client tools (for `mysqldump` and `mysql` utilities)

### Installing MySQL Client Tools

Cadangkan uses `mysqldump` to create database backups and `mysql` to restore them. You need to install MySQL client tools, **not the full MySQL server**.

**Linux (Debian/Ubuntu):**
```bash
sudo apt-get install mysql-client
```

**Linux (RHEL/CentOS/Fedora):**
```bash
sudo yum install mysql  # or mysql-community-client
# or on newer systems:
sudo dnf install mysql
```

**macOS (Homebrew):**
```bash
brew install mysql-client
```

After installation on macOS, add to your PATH:
```bash
# Intel Macs
echo 'export PATH="/usr/local/opt/mysql-client/bin:$PATH"' >> ~/.bash_profile

# Apple Silicon (M1/M2)
echo 'export PATH="/opt/homebrew/opt/mysql-client/bin:$PATH"' >> ~/.zshrc
```

**Windows:**
- Download [MySQL Installer](https://dev.mysql.com/downloads/installer/) and select only "MySQL Client" component
- Or use Chocolatey: `choco install mysql.utilities`

**Version Compatibility:**
- MySQL client 8.0 is recommended for best compatibility with both MySQL 5.7 and 8.0 servers
- A single MySQL client version can backup and restore multiple MySQL server versions
- Newer MySQL client tools can work with older MySQL servers without issues

### Getting Started

1. **Clone the repository**
   ```bash
   git clone https://github.com/erickhilda/cadangkan.git
   cd cadangkan
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Start MySQL test databases**
   ```bash
   docker-compose up -d
   ```
   
   This will start:
   - MySQL 8.0 on port 3306
   - MySQL 5.7 on port 3307
   
   Test credentials:
   - Root: `root` / `rootpassword`
   - User: `testuser` / `testpassword`
   - Database: `cadangkan_test`

4. **Run tests**
   ```bash
   go test -v ./...
   ```

5. **Build the binary**
   ```bash
   go build -o cadangkan ./cmd/cadangkan
   ```

### Development Workflow

- Create feature branches from `develop`
- Submit Pull Requests to `develop` branch
- Tests will run automatically on PR creation
- All tests must pass before merging

## 📦 Usage

### Building the CLI

```bash
go build -o cadangkan ./cmd/cadangkan
```

### Managing Database Connections

**Add a database configuration:**
```bash
cadangkan add --host=mysql.example.com \
  --user=backup_user \
  --database=myapp \
  mysql production
# Enter password interactively
```

**List configured databases:**
```bash
cadangkan list
```

**Test connection:**
```bash
cadangkan test production
```

**Remove a database:**
```bash
cadangkan remove production
```

### Backup MySQL Database

**Using saved configuration (recommended):**
```bash
cadangkan backup production
```

**Direct mode (passing all flags):**
```bash
cadangkan backup --host=127.0.0.1 --user=root --password=secret --database=mydb
```

**With backup options:**
```bash
# Backup specific tables
cadangkan backup production --tables=users,orders

# Exclude specific tables
cadangkan backup production --exclude-tables=logs,sessions

# Schema only (no data)
cadangkan backup production --schema-only

# Custom output directory
cadangkan backup production --output=/path/to/backups

# Without compression
cadangkan backup production --compression=none
```

**Important:** Use `127.0.0.1` instead of `localhost` when backing up Docker MySQL containers to avoid Unix socket connection issues.

**Backup location:** Backups are stored in `~/.cadangkan/backups/[database]/` by default.

### Restore MySQL Database

**Using saved configuration (restore latest backup):**
```bash
cadangkan restore production
```

**Restore a specific backup:**
```bash
cadangkan restore production --from=2025-01-15-143022
```

**Restore to a different database:**
```bash
# Restore to an existing database
cadangkan restore production --to=production_restored

# Restore to a new database (creates it automatically)
cadangkan restore production --to=new_database --create-db
```

**Direct mode (without saved config):**
```bash
cadangkan restore --host=127.0.0.1 --user=root --password=secret \
  --database=mydb --from=2025-01-15-143022
```

**Safety options:**
```bash
# Dry-run: validate without executing
cadangkan restore production --dry-run

# Backup target database before restoring (if it exists)
cadangkan restore production --backup-first

# Skip confirmation prompt
cadangkan restore production --yes

# Show verbose output (including mysql command)
cadangkan restore production --verbose
```

**Important Notes:**
- By default, restores the **latest backup** if `--from` is not specified
- Use `--create-db` to automatically create the target database if it doesn't exist
- The `--to` flag allows restoring to a different database than the source
- Restore operations require the `mysql` command-line client to be installed
- Backups are automatically decompressed during restore

### Command Options

**Database Management:**
```
cadangkan add [flags] mysql <name>      Add a database configuration
cadangkan list                          List all configured databases
cadangkan test <name>                   Test database connection
cadangkan remove <name>                 Remove a database configuration
```

**Backup:**
```
cadangkan backup [name] [flags]

Flags:
  --type string              Database type (default: "mysql")
  --host string              Database host (overrides config)
  --port int                 Database port (overrides config)
  --user string              Database user (overrides config)
  --password string          Database password (overrides config)
  --database string          Database name (overrides config)
  --tables strings           Specific tables to backup
  --exclude-tables strings   Tables to exclude from backup
  --schema-only              Backup schema only (no data)
  --compression string       Compression type: gzip, none (default: "gzip")
  --output string            Output directory (default: ~/.cadangkan/backups)
```

**Restore:**
```
cadangkan restore [name] [flags]

Flags:
  --type string              Database type (default: "mysql")
  --from string              Specific backup ID to restore (default: latest)
  --to string                Target database name (overrides config database)
  --create-db                Create database if it doesn't exist
  --host string              Database host (overrides config)
  --port int                 Database port (overrides config)
  --user string              Database user (overrides config)
  --password string          Database password (overrides config)
  --database string          Database name (overrides config)
  --dry-run                  Validate restore without executing
  --backup-first             Backup target database before restore (if exists)
  --yes, -y                  Skip confirmation prompt
  --verbose, -v              Show verbose output including mysql command
```

## 📖 Documentation

For detailed information, see:
- [Configuration Guide](docs/CONFIGURATION.md) - Managing database connections
- [Product Specifications](docs/product-sepcifications.md) - Full product vision and roadmap
- [Architecture Decision Records (ADRs)](docs/adr/README.md) - Important architectural decisions and their context

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 📬 Contact

- GitHub: [@erickhilda](https://github.com/erickhilda)
- Project Link: [https://github.com/erickhilda/cadangkan](https://github.com/erickhilda/cadangkan)

---

**Note:** This is an early-stage project. APIs and commands may change before the v1.0 release.
