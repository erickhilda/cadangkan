# Cadangkan

> **Sleep soundly knowing your databases are backed up**

Cadangkan (Indonesian for "backup") is a universal database backup and synchronization tool that makes database protection effortless, affordable, and accessible to every developer.

## 🚧 Development Status

**Current Phase:** Phase 0 - Project Setup

This project is in early development. We're currently setting up the foundation and working towards our MVP release (v0.1.0) which will support MySQL backup and restore functionality.

## 🎯 Vision

DB-Shield aims to provide:

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
- MySQL client tools (for `mysqldump` utility)

### Installing mysqldump

Cadangkan uses `mysqldump` to create database backups. You need to install MySQL client tools, **not the full MySQL server**.

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
- A single `mysqldump` version can backup multiple MySQL server versions
- Newer `mysqldump` can backup older MySQL servers without issues

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

## 📖 Documentation

For detailed product specifications and roadmap, see:
- [Product Specifications](docs/product-sepcifications.md)
- [Architecture Decision Records (ADRs)](docs/adr/README.md) - Important architectural decisions and their context

## 🗺️ Roadmap

- **Phase 0** (Current): Project setup and CI/CD
- **Phase 1** (Weeks 1-3): MySQL MVP - Backup, restore, and automation
- **Phase 2** (Weeks 4-5): PostgreSQL support
- **Phase 3** (Week 6): Cloud storage integration
- **Phase 4** (Week 7): Clone and notifications
- **Phase 5** (Weeks 8-10): Web interface

## 🤝 Contributing

We welcome contributions! Here's how you can help:

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request to the `develop` branch

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🌟 Project Goals

- **6-Month Goals:**
  - 500+ active users
  - 200+ GitHub stars
  - >99% backup success rate
  - $150K+ in collective cost savings for users

## 📬 Contact

- GitHub: [@erickhilda](https://github.com/erickhilda)
- Project Link: [https://github.com/erickhilda/cadangkan](https://github.com/erickhilda/cadangkan)

---

**Note:** This is an early-stage project. APIs and commands may change before the v1.0 release.
