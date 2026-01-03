# Cadangkan Product Specification
**Version 1.0**  
**Last Updated:** January 2025

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Product Overview](#2-product-overview)
3. [Market Analysis](#3-market-analysis)
4. [Target Users](#4-target-users)
5. [Core Features](#5-core-features)
6. [Technical Specification](#6-technical-specification)
7. [User Experience](#7-user-experience)
8. [Development Roadmap](#8-development-roadmap)
9. [Success Metrics](#9-success-metrics)
10. [Go-to-Market Strategy](#10-go-to-market-strategy)
11. [Risk Assessment](#11-risk-assessment)
12. [Appendices](#12-appendices)

---

## 1. Executive Summary

### 1.1 Product Vision

**Cadangkan** is a universal database backup and sync tool that makes database protection effortless, affordable, and accessible to every developer.

**Tagline:** *"Sleep soundly knowing your databases are backed up"*

### 1.2 Problem Statement

Developers and teams face significant challenges with database backups:

**Cost Barriers:**
- Cloud provider backup features cost $25-500+ per month per database
- Small projects and hobby developers can't afford professional backup solutions
- Managed database providers charge premiums for automated backups

**Complexity Challenges:**
- Manual backup processes are error-prone and time-consuming
- Each database provider has different backup mechanisms
- No unified tool works across MySQL, PostgreSQL, and other databases
- Setting up automated backups requires DevOps expertise

**Development Workflow Issues:**
- Cloning production databases to local development is complex
- Multiple manual steps required for environment setup
- Data sanitization for development is tedious
- No consistent process across different database types

### 1.3 Solution Overview

Cadangkan provides:

1. **Universal Backup Automation**
   - Works with MySQL, PostgreSQL, MongoDB (future), and more
   - Single command to backup any database
   - Automated scheduling with retention policies
   - Local and cloud storage support

2. **Simplified Database Cloning**
   - One-command cloning to local development
   - Automatic Docker setup for local databases
   - Optional data sanitization
   - Consistent workflow across database types

3. **Free and Open Source**
   - No subscription costs
   - Community-driven development
   - Self-hosted, full control over data
   - Enterprise features available for all

### 1.4 Key Differentiators

| Feature | Cadangkan | Cloud Provider Backups | Manual Scripts |
|---------|-----------|------------------------|----------------|
| **Cost** | Free | $25-500/month | Free |
| **Setup Time** | 5 minutes | Instant (if paid) | Hours |
| **Multi-Database** | ✓ MySQL, PostgreSQL+ | Single provider | Per-database scripts |
| **Automation** | ✓ Built-in | ✓ Built-in | Manual cron setup |
| **Clone to Local** | ✓ One command | ✗ Complex | ✗ Manual |
| **Cloud Storage** | ✓ S3, GCS, R2 | ✓ Provider storage | ✗ Manual upload |
| **Self-Hosted** | ✓ Yes | ✗ No | ✓ Yes |
| **Retention Policies** | ✓ Configurable | ✓ Fixed by plan | Manual cleanup |

### 1.5 Success Metrics (6-Month Goals)

- **Adoption:** 500+ active users
- **Community:** 200+ GitHub stars
- **Reliability:** >99% backup success rate
- **Performance:** <5min backup for 1GB database
- **Documentation:** 95%+ satisfaction
- **Market Impact:** $150K+ in collective cost savings for users

---

## 2. Product Overview

### 2.1 Product Description

Cadangkan is a command-line tool written in Go that provides comprehensive database backup, restore, and cloning capabilities across multiple database systems. It abstracts the complexity of database-specific backup tools (mysqldump, pg_dump, etc.) behind a simple, unified interface.

### 2.2 Core Value Propositions

**For Individual Developers:**
- "Back up your databases in 5 minutes, not 5 hours"
- "Never worry about losing your data again"
- "Clone production to development with one command"

**For Small Teams:**
- "One tool for all your databases"
- "Automated backups that just work"
- "Save $300-7,000/year on backup costs"

**For Freelancers/Agencies:**
- "Manage backups for 20+ client databases from one place"
- "Consistent backup strategy across all projects"
- "Professional backup solution without enterprise costs"

### 2.3 Product Principles

1. **Simplicity First:** Complex operations should be simple commands
2. **Database Agnostic:** Same interface for all database types
3. **Safety by Default:** Confirmations for destructive operations
4. **Transparency:** Clear error messages, detailed logging
5. **Performance:** Fast backups, minimal resource usage
6. **Reliability:** Backups must work, every time

### 2.4 Product Positioning

**Market Position:** Open-source alternative to paid managed backup solutions

**Positioning Statement:**
> "For developers and teams who need reliable database backups without enterprise costs, Cadangkan is an open-source CLI tool that provides automated backup, restore, and cloning for MySQL, PostgreSQL, and more. Unlike managed backup services that cost hundreds per month or manual scripts that require DevOps expertise, Cadangkan offers a simple, unified interface that works with any database in just minutes."

---

## 3. Market Analysis

### 3.1 Market Size

**Primary Market: MySQL Users**
- Global MySQL users: 8+ million developers
- Organizations using MySQL: 100,000+
- Managed MySQL services growing 25% YoY

**Secondary Market: PostgreSQL Users**
- Global PostgreSQL users: 5+ million developers
- Rapid growth in managed PostgreSQL (Supabase, Neon, etc.)
- Enterprise adoption increasing

**Tertiary Markets:**
- MongoDB users: 3+ million developers
- SQLite users: Ubiquitous in mobile/edge
- Other SQL databases: SQL Server, MariaDB, etc.

### 3.2 Market Trends

**1. Managed Database Growth**
- Cloud databases growing 30% annually
- More developers using managed services
- PlanetScale, Supabase, Railway, Neon gaining traction
- Free tiers becoming limited or removed

**2. Cost Optimization**
- Startups cutting infrastructure costs
- Developers seeking alternatives to paid features
- DIY DevOps becoming mainstream
- Open-source tools gaining adoption

**3. Developer Experience Focus**
- CLI tools gaining popularity
- Infrastructure-as-code becoming standard
- Developers want ownership and control
- Self-hosted solutions preferred for cost and privacy

### 3.3 Competitive Landscape

**Direct Competitors:**

1. **Cloud Provider Native Backups**
   - AWS RDS Automated Backups
   - Google Cloud SQL Backups
   - Azure Database Backups
   - **Weakness:** Expensive, vendor lock-in

2. **Backup SaaS Tools**
   - BackupNinja
   - DBacked
   - SimpleBackups
   - **Weakness:** Subscription costs, limited database support

3. **Open Source Tools**
   - pg_dump/mysqldump (manual)
   - pgbackrest (PostgreSQL only)
   - mydumper (MySQL only)
   - **Weakness:** Database-specific, manual setup

**Indirect Competitors:**
- Custom scripts (time-consuming)
- Manual backups (unreliable)
- No backups (risky)

**Competitive Advantages:**
- ✓ Free and open source
- ✓ Multi-database support
- ✓ Simple, unified interface
- ✓ Automated scheduling
- ✓ Clone functionality
- ✓ Active development and community

### 3.4 Market Opportunity

**Recent Market Events:**

1. **PlanetScale Free Tier Removal (November 2024)**
   - Thousands of developers need alternatives
   - Community actively searching for solutions
   - Perfect timing for Cadangkan launch

2. **Supabase Growth**
   - 500K+ users, mostly on free tier
   - Backup features only on paid plans ($25+/month)
   - Large market of cost-conscious developers

3. **Startup Cost Cutting**
   - Economic uncertainty driving optimization
   - Infrastructure costs under scrutiny
   - Self-hosted solutions gaining favor

**Total Addressable Market (TAM):**
- MySQL/PostgreSQL developers needing backups: 10M+
- Managed database users seeking alternatives: 1M+
- Immediate addressable market: 100K+ developers

**Serviceable Addressable Market (SAM):**
- Developers willing to use CLI tools: 20%
- Cost-conscious or hobby developers: 30%
- SAM: 6M+ developers

**Serviceable Obtainable Market (SOM):**
- Realistic 2-year goal: 0.01% = 10,000 users
- Conservative 1-year goal: 1,000 active users

---

## 4. Target Users

### 4.1 Primary Personas

#### Persona 1: "Alex" - The Solo Developer

**Demographics:**
- Age: 25-35
- Experience: 3-5 years
- Role: Full-stack developer / Indie hacker
- Location: Global, often in lower-cost regions

**Context:**
- Working on 2-3 side projects
- Using free/hobby tier databases
- Budget: $0-50/month for infrastructure
- Technical: Comfortable with CLI, basic DevOps

**Pain Points:**
- Can't afford $25/month for backups on each project
- Manual backups are tedious and forgotten
- Lost data on a project once, scared it will happen again
- Wants professional setup without enterprise costs

**Goals:**
- Automated backups without monthly costs
- Simple setup and maintenance
- Peace of mind about data safety
- Clone prod to dev for testing

**Cadangkan Value:**
- Free, automated backups
- 5-minute setup
- Reliable and tested
- One tool for all projects

**Quote:** *"I'm building my SaaS and can't afford to pay for backups on top of hosting. I need something simple that just works."*

---

#### Persona 2: "Sarah" - The Startup CTO

**Demographics:**
- Age: 30-40
- Experience: 8-12 years
- Role: CTO / Lead Engineer
- Team: 2-5 developers
- Location: Major tech hub or remote

**Context:**
- Running 5-10 databases (prod, staging, dev)
- Using managed databases (AWS RDS, Railway, etc.)
- Budget: Tight, every dollar counts
- Technical: Strong DevOps, but time-constrained

**Pain Points:**
- Backup costs adding up: $200-500/month
- Different backup process for each provider
- Team needs consistent dev environments
- Compliance requirements for data protection
- Time spent managing backup infrastructure

**Goals:**
- Reduce infrastructure costs
- Standardize backup process
- Easy dev environment setup for team
- Reliable disaster recovery plan
- Audit trail for compliance

**Cadangkan Value:**
- Save $200-500/month
- Unified backup strategy
- Easy onboarding for new developers
- Professional features without cost
- Open source = audit trail

**Quote:** *"We're spending $400/month on database backups alone. I need a solution that scales with our team without scaling costs."*

---

#### Persona 3: "Marcus" - The Freelance Developer

**Demographics:**
- Age: 28-45
- Experience: 5-10 years
- Role: Freelancer / Agency owner
- Clients: 10-30 active clients
- Location: Global

**Context:**
- Managing databases for multiple clients
- Mix of MySQL, PostgreSQL, etc.
- Each client has different hosting
- Responsible for client data protection
- Budget: Per-client, needs cost efficiency

**Pain Points:**
- Managing backups for 20+ databases manually
- Each client project has different setup
- No consistent backup strategy
- Worried about liability if data is lost
- Time spent on backup management not billable

**Goals:**
- Single tool for all client databases
- Automated, reliable backups
- Easy to show clients backup status
- Professional appearance
- Minimize non-billable work

**Cadangkan Value:**
- One tool for all clients
- Automated, hands-off operation
- Professional backup reports
- Free = higher margins
- Reduces liability risk

**Quote:** *"I manage databases for 25 clients. I need one tool that works reliably for all of them so I can focus on billable work."*

---

### 4.2 Secondary Personas

#### Persona 4: "Dev Team Lead" at Mid-Size Company
- Manages 20+ databases
- Needs audit compliance
- Budget for tools but wants efficiency
- Values: Reliability, standardization, reporting

#### Persona 5: "CS Student / Bootcamp Graduate"
- Learning database management
- Building portfolio projects
- Zero budget
- Values: Free, educational, simple

#### Persona 6: "DevOps Engineer"
- Managing backup infrastructure
- Looking to simplify/modernize
- Wants infrastructure-as-code
- Values: Automation, monitoring, integration

### 4.3 User Journey Map

**Discovery Phase:**
```
Problem Recognition → Search for Solutions → Evaluate Options
↓
"I need backups" → "backup tool mysql" → Compare tools
                     "planetscale alternative"
                     "free database backup"
```

**Evaluation Phase:**
```
Find Cadangkan → Read Docs → Try Demo → Check GitHub
↓
Landing page → Quick start → Local test → Review code
```

**Adoption Phase:**
```
Install → First Backup → Schedule → Multiple Databases
↓
5 minutes → Success! → Automate → Standardize
```

**Retention Phase:**
```
Daily Use → Recommend → Contribute → Advocate
↓
"It just works" → Tell colleagues → Open PRs → Blog post
```

---

## 5. Core Features

### 5.1 Feature Overview

Cadangkan is organized into six core feature areas:

1. **Database Management** - Add, configure, and manage databases
2. **Backup Operations** - Create, manage, and validate backups
3. **Restore Operations** - Restore backups to any target
4. **Clone Operations** - Copy databases to local development
5. **Automation** - Schedule backups and retention policies
6. **Monitoring** - Track backup health and storage

### 5.2 MVP Features (Phase 1 - MySQL)

#### 5.2.1 Database Connection & Management

**Feature:** Add and manage MySQL database configurations

**User Stories:**
- As a developer, I want to add my MySQL database credentials so that I can start backing it up
- As a user, I want to test my database connection before saving it so that I know my credentials are correct
- As a user, I want to manage multiple databases so that I can backup all my projects

**Acceptance Criteria:**
- ✓ User can add MySQL database via CLI command
- ✓ User can add database via YAML config file
- ✓ Credentials are encrypted when stored
- ✓ Connection is tested before saving
- ✓ User receives clear error messages on failure
- ✓ User can list all configured databases
- ✓ User can remove databases
- ✓ Supports environment variables for credentials

**Commands:**
```bash
# Add database interactively
cadangkan add mysql production

# Add database with flags
cadangkan add mysql production \
  --host=mysql.example.com \
  --port=3306 \
  --database=myapp \
  --user=backup_user \
  --password-stdin

# List databases
cadangkan list

# Test connection
cadangkan test production

# Remove database
cadangkan remove production
```

**Technical Implementation:**
- Use `github.com/go-sql-driver/mysql` for MySQL driver
- Implement connection pooling
- Timeout handling (10s default)
- Credential encryption with AES-256
- Store config in `~/.cadangkan/config.yaml`
- Store encrypted credentials separately

**Priority:** P0 (Must Have for MVP)

---

#### 5.2.2 Manual Backup Creation

**Feature:** Create on-demand backups of MySQL databases

**User Stories:**
- As a developer, I want to backup my database with one command so that I don't have to remember mysqldump syntax
- As a user, I want to see progress while backing up so that I know it's working
- As a user, I want backups to be compressed so that they use less disk space
- As a developer, I want to backup only specific tables so that I can backup faster

**Acceptance Criteria:**
- ✓ User can create full database backup
- ✓ User can backup specific tables
- ✓ User can exclude specific tables
- ✓ Backup is compressed (gzip by default)
- ✓ Progress is shown during backup
- ✓ Backup includes metadata (size, duration, checksum)
- ✓ User sees clear success/failure message
- ✓ Backup time is reasonable (<5min for 1GB database)

**Commands:**
```bash
# Full backup
cadangkan backup production

# Specific tables
cadangkan backup production --tables=users,orders

# Exclude tables
cadangkan backup production --exclude-tables=logs,sessions

# Schema only
cadangkan backup production --schema-only

# Custom compression
cadangkan backup production --compression=zstd
```

**User Experience:**
```
$ cadangkan backup production

✓ Connected to database (MySQL 8.0.35)
✓ Starting backup...
⠋ Backing up... 234/500 tables (47%) [2m 15s]
✓ Compressing...
✓ Generating metadata...
✓ Backup complete!

Backup ID: 2025-01-15-143022
Size: 234 MB (compressed from 890 MB)
Duration: 2m 34s
Tables: 500
Location: ~/.cadangkan/backups/production/2025-01-15-143022.sql.gz

Next steps:
  - Restore: cadangkan restore production --from=2025-01-15-143022
  - Schedule: cadangkan schedule production --daily
```

**Technical Implementation:**
- Execute `mysqldump` with optimal flags:
  - `--single-transaction` for consistency
  - `--quick` for memory efficiency
  - `--routines` for stored procedures
  - `--triggers` for triggers
  - `--events` for events
- Stream output directly to compression
- Calculate SHA-256 checksum
- Generate metadata JSON file
- Organized directory structure:
  ```
  ~/.cadangkan/backups/production/
    2025-01-15-143022.sql.gz
    2025-01-15-143022.meta.json
  ```

**Performance Requirements:**
- 1 GB database: <5 minutes
- 10 GB database: <30 minutes
- Memory usage: <500 MB
- CPU usage: 1-2 cores

**Error Handling:**
- Connection failures: Retry 3 times with backoff
- Disk full: Check space before backup
- Process killed: Clean up partial backup
- MySQL errors: Show clear error message

**Priority:** P0 (Must Have for MVP)

---

#### 5.2.3 Backup Management

**Feature:** List, view, and manage existing backups

**User Stories:**
- As a user, I want to see all my backups so that I can choose which one to restore
- As a user, I want to see backup details so that I know what's in each backup
- As a user, I want to delete old backups so that I can free up disk space
- As a user, I want to validate backups so that I know they're not corrupted

**Acceptance Criteria:**
- ✓ User can list all backups for a database
- ✓ List shows ID, date, size, type
- ✓ User can view detailed backup information
- ✓ User can delete individual backups
- ✓ Deletion requires confirmation
- ✓ User can validate backup integrity
- ✓ List supports filtering and sorting

**Commands:**
```bash
# List all backups
cadangkan backups list production

# Show backup details
cadangkan backups info production 2025-01-15-143022

# Delete backup
cadangkan backups delete production 2025-01-15-143022

# Validate backup
cadangkan backups validate production 2025-01-15-143022
```

**User Experience:**
```
$ cadangkan backups list production

Backups for production (MySQL)
════════════════════════════════════════════════════════

ID                  DATE                SIZE      TYPE     STATUS
2025-01-15-143022  2025-01-15 14:30    234 MB    manual   ✓
2025-01-15-020000  2025-01-15 02:00    230 MB    daily    ✓
2025-01-14-020000  2025-01-14 02:00    228 MB    daily    ✓
2025-01-13-020000  2025-01-13 02:00    225 MB    daily    ✓

Total: 4 backups, 917 MB

Commands:
  info     cadangkan backups info production <id>
  restore  cadangkan restore production --from=<id>
  delete   cadangkan backups delete production <id>
```

**Technical Implementation:**
- Read backup directory
- Parse metadata files
- Sort by date (newest first)
- Calculate total storage
- Color-code status (green = valid, red = corrupted)
- Pagination for large lists

**Priority:** P0 (Must Have for MVP)

---

#### 5.2.4 Restore Operations

**Feature:** Restore backups to MySQL databases

**User Stories:**
- As a user, I want to restore my latest backup so that I can recover from data loss
- As a user, I want to restore to a different database so that I can test without affecting production
- As a developer, I want a dry-run option so that I can verify before restoring
- As a user, I want confirmation before restoring so that I don't accidentally overwrite data

**Acceptance Criteria:**
- ✓ User can restore latest backup
- ✓ User can restore specific backup by ID
- ✓ User can restore to different database
- ✓ Restore shows progress
- ✓ User must confirm before restore
- ✓ Dry-run mode validates without applying
- ✓ Option to backup target before restore
- ✓ Clear error messages on failure

**Commands:**
```bash
# Restore latest backup
cadangkan restore production

# Restore specific backup
cadangkan restore production --from=2025-01-15-143022

# Restore to different database
cadangkan restore production \
  --from=2025-01-15-143022 \
  --to=staging

# Dry run
cadangkan restore production --dry-run

# Backup before restore
cadangkan restore production --backup-first
```

**User Experience:**
```
$ cadangkan restore production --from=2025-01-15-143022

⚠  WARNING: This will restore production database
⚠  Current data will be overwritten!

Backup to restore:
  ID: 2025-01-15-143022
  Created: 2025-01-15 14:30 (2 hours ago)
  Size: 234 MB
  Tables: 500

Target database:
  Name: production
  Host: mysql.example.com
  Current size: 890 MB

Continue? [y/N]: y

✓ Downloading backup...
✓ Decompressing...
✓ Validating target connection...
⠋ Restoring... (47%) [1m 15s]
✓ Restore complete!

Duration: 2m 45s
Tables restored: 500

Database is now ready to use.
```

**Technical Implementation:**
- Decompress backup file
- Verify checksum
- Test target database connection
- Execute `mysql` command with streaming
- Progress tracking via line counting
- Transaction-based restore when possible
- Rollback on error

**Safety Features:**
- Confirmation prompt
- Dry-run validation
- Pre-restore backup option
- Target database checks
- Clear warnings

**Priority:** P0 (Must Have for MVP)

---

#### 5.2.5 Scheduled Backups

**Feature:** Automate backups with cron-based scheduling

**User Stories:**
- As a user, I want to schedule daily backups so that I don't have to remember to run them
- As a developer, I want flexible scheduling so that I can backup at optimal times
- As a user, I want to enable/disable schedules so that I can control when backups run
- As a user, I want to see when the next backup will run so that I can plan accordingly

**Acceptance Criteria:**
- ✓ User can schedule backups with cron syntax
- ✓ User can set daily/weekly shortcuts
- ✓ User can enable/disable schedules
- ✓ Schedule persists across restarts
- ✓ User can see next scheduled run
- ✓ Failed backups retry automatically
- ✓ Schedules run even when terminal is closed

**Commands:**
```bash
# Daily backup at 2 AM
cadangkan schedule set production --daily --time=02:00

# Weekly backup on Sunday at 3 AM
cadangkan schedule set production --weekly --day=sunday --time=03:00

# Custom cron expression
cadangkan schedule set production --cron="0 2 * * *"

# Enable/disable
cadangkan schedule enable production
cadangkan schedule disable production

# List all schedules
cadangkan schedule list

# Show next runs
cadangkan schedule next
```

**User Experience:**
```
$ cadangkan schedule set production --daily --time=02:00

✓ Schedule configured for production

Schedule: Daily at 02:00 (0 2 * * *)
Next run: Tomorrow at 02:00 (in 9 hours)
Status: Enabled

The backup will run automatically in the background.

To start the scheduler:
  cadangkan service install  # Install as system service
  cadangkan service start    # Start the service
```

**Technical Implementation:**
- Use `github.com/robfig/cron/v3` for scheduling
- Cron job registry in database
- Persistent schedule storage
- Retry logic: 3 attempts with exponential backoff
- Logging of all scheduled runs
- Next run calculation and display

**Priority:** P0 (Must Have for MVP)

---

#### 5.2.6 System Service Integration

**Feature:** Run Cadangkan as a background service

**User Stories:**
- As a user, I want backups to run automatically even when I'm not logged in
- As a server admin, I want Cadangkan to start on system boot
- As a user, I want to easily check if the service is running
- As a user, I want to view service logs to troubleshoot issues

**Acceptance Criteria:**
- ✓ Service can be installed as systemd unit (Linux)
- ✓ Service can be installed as launchd job (macOS)
- ✓ Service starts on system boot
- ✓ Service can be started/stopped/restarted
- ✓ User can check service status
- ✓ User can view service logs
- ✓ Service runs as non-root user (Linux)

**Commands:**
```bash
# Install service
sudo cadangkan service install

# Start service
sudo cadangkan service start

# Stop service
sudo cadangkan service stop

# Restart service
sudo cadangkan service restart

# Check status
cadangkan service status

# View logs
cadangkan service logs

# Follow logs
cadangkan service logs --follow

# Uninstall service
sudo cadangkan service uninstall
```

**User Experience:**
```
$ sudo cadangkan service install

Installing Cadangkan as system service...

✓ Created systemd unit file: /etc/systemd/system/cadangkan.service
✓ Created user: dbshield
✓ Set permissions
✓ Reloaded systemd daemon
✓ Service installed successfully

Next steps:
  sudo cadangkan service start    # Start the service
  cadangkan service status        # Check status

The service will start automatically on system boot.
```

**Technical Implementation:**

**Linux (systemd):**
```ini
# /etc/systemd/system/cadangkan.service
[Unit]
Description=Cadangkan Database Backup Service
After=network.target

[Service]
Type=simple
User=dbshield
Group=dbshield
ExecStart=/usr/local/bin/cadangkan daemon
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

**macOS (launchd):**
```xml
<!-- ~/Library/LaunchAgents/com.dbshield.daemon.plist -->
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.dbshield.daemon</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/cadangkan</string>
        <string>daemon</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
</dict>
</plist>
```

**Priority:** P0 (Must Have for MVP)

---

#### 5.2.7 Retention Policies

**Feature:** Automatically manage backup retention

**User Stories:**
- As a user, I want old backups to be automatically deleted so that I don't run out of disk space
- As a user, I want to keep different numbers of daily/weekly/monthly backups so that I can have granular recent backups and long-term archive
- As a user, I want to configure retention policies so that I can control how long backups are kept

**Acceptance Criteria:**
- ✓ Backups are categorized as daily/weekly/monthly
- ✓ Retention policy is configurable per database
- ✓ Old backups are automatically deleted
- ✓ User can manually trigger cleanup
- ✓ Important backups can be marked for preservation
- ✓ Cleanup shows what will be deleted (dry-run)

**Configuration:**
```yaml
# ~/.cadangkan/config.yaml
databases:
  production:
    type: mysql
    # ... connection details ...
    retention:
      daily: 7      # Keep last 7 daily backups
      weekly: 4     # Keep last 4 weekly backups
      monthly: 12   # Keep last 12 monthly backups
```

**Commands:**
```bash
# Manual cleanup
cadangkan cleanup production

# Dry-run (show what would be deleted)
cadangkan cleanup production --dry-run

# Set retention policy
cadangkan config edit production
```

**Retention Logic:**
- **Daily:** Most recent backup each day for last N days
- **Weekly:** Most recent backup each week (Sunday) for last N weeks
- **Monthly:** Most recent backup each month (1st) for last N months
- Cleanup runs after each backup
- Preserved backups are never deleted

**User Experience:**
```
$ cadangkan cleanup production --dry-run

Cleanup preview for production
════════════════════════════════════════════════════════

Retention policy:
  Daily: Keep last 7 days
  Weekly: Keep last 4 weeks
  Monthly: Keep last 12 months

Backups to delete:
  2025-01-01-020000  (15 days old)  234 MB
  2024-12-28-020000  (18 days old)  230 MB
  2024-12-25-020000  (21 days old)  228 MB

Total to delete: 3 backups, 692 MB

Backups to keep:
  7 daily backups (last 7 days)
  4 weekly backups (last 4 weeks)
  12 monthly backups (last 12 months)

Run without --dry-run to delete these backups.
```

**Priority:** P1 (Should Have for MVP)

---

#### 5.2.8 Status & Health Monitoring

**Feature:** Monitor backup status and health

**User Stories:**
- As a user, I want to see an overview of all my databases so that I know everything is backed up
- As a user, I want to see backup health score so that I can identify problems
- As a user, I want to see storage usage so that I can plan capacity
- As a user, I want to know when the next backup will run

**Acceptance Criteria:**
- ✓ Overall status dashboard
- ✓ Per-database status
- ✓ Last backup time
- ✓ Next scheduled backup
- ✓ Backup success rate
- ✓ Storage usage breakdown
- ✓ Health score calculation

**Commands:**
```bash
# Overall status
cadangkan status

# Database-specific status
cadangkan status production

# Storage usage
cadangkan storage

# Health check
cadangkan health production
```

**User Experience:**
```
$ cadangkan status

Cadangkan Status
════════════════════════════════════════════════════════
Service: Running ✓
Databases: 3 (2 active)
Total Backups: 42
Storage Used: 12.4 GB / 100 GB (12%)
Last Backup: 2 hours ago ✓

DATABASE      TYPE    STATUS  LAST BACKUP   NEXT BACKUP
production    mysql   ✓       2 hours ago   in 22 hours
staging       mysql   ✓       1 day ago     in 6 days
development   mysql   ✗       Never         -

Health Summary:
  ✓ All scheduled databases backed up recently
  ✓ No failed backups in last 7 days
  ⚠  1 database never backed up (development)

Commands:
  cadangkan status <database>  # Detailed status
  cadangkan health <database>  # Health score
  cadangkan storage            # Storage breakdown
```

**Health Score Calculation:**
```
Health Score = (
  Success Rate (50%) +
  Recency (30%) +
  Consistency (20%)
) * 100

Success Rate: Successful backups / Total attempts (last 30 days)
Recency: Days since last backup (0 days = 100%, 7+ days = 0%)
Consistency: Standard deviation of backup intervals
```

**Priority:** P1 (Should Have for MVP)

---

### 5.3 Phase 2 Features (PostgreSQL Support)

#### 5.3.1 PostgreSQL Database Support

**Feature:** Full support for PostgreSQL databases

**Additional Complexity:**
- Schema-aware backups
- Extension handling
- Roles and permissions
- Large objects
- Different dump formats (plain, custom, directory)

**Commands remain the same:**
```bash
# Add PostgreSQL database
cadangkan add postgres production \
  --host=pg.example.com \
  --database=myapp

# Backup works identically
cadangkan backup production
```

**Implementation:**
- Use `github.com/lib/pq` for PostgreSQL driver
- Execute `pg_dump` for backups
- Execute `psql` for restores
- Handle extensions (uuid-ossp, pgcrypto, etc.)
- Preserve schemas and permissions

**Timeline:** Week 4-5

---

### 5.4 Phase 3 Features (Cloud Storage)

#### 5.4.1 Cloud Storage Integration

**Feature:** Upload backups to cloud storage

**Supported Providers:**
- Amazon S3
- Google Cloud Storage
- Cloudflare R2
- Backblaze B2

**Configuration:**
```yaml
databases:
  production:
    storage:
      local:
        enabled: true
        path: ~/.cadangkan/backups/production
      
      cloud:
        enabled: true
        provider: s3
        bucket: my-db-backups
        region: us-east-1
        path: production/
```

**Commands:**
```bash
# Configure cloud storage
cadangkan storage add s3 \
  --bucket=my-backups \
  --region=us-east-1

# Upload specific backup
cadangkan storage upload production --backup-id=xxx

# Download from cloud
cadangkan storage download production --backup-id=xxx

# List cloud backups
cadangkan storage list production --cloud
```

**Features:**
- Automatic upload after local backup
- Upload only backups older than N hours/days
- Download on-demand for restore
- Cost optimization (upload less frequently)
- Encryption at rest

**Timeline:** Week 6

---

### 5.5 Phase 4 Features (Advanced)

#### 5.5.1 Database Cloning

**Feature:** One-command database cloning to local environment

**Commands:**
```bash
# Clone to local Docker
cadangkan clone production --to-docker

# Clone to existing database
cadangkan clone production \
  --to=mysql://localhost:3306/dev_db

# Clone with custom Docker config
cadangkan clone production \
  --to-docker \
  --docker-port=3307 \
  --docker-name=myapp-dev
```

**Implementation:**
- Create Docker container automatically
- Restore latest backup to local database
- Update local config files (optional)
- Data sanitization (optional)

**Timeline:** Week 7

---

#### 5.5.2 Notifications

**Feature:** Alerts on backup success/failure

**Channels:**
- Email (SMTP)
- Slack webhook
- Discord webhook
- Telegram bot
- Custom webhook

**Configuration:**
```yaml
databases:
  production:
    notifications:
      on_success: false
      on_failure: true
      channels:
        - type: email
          to: admin@example.com
        - type: slack
          webhook_url: https://hooks.slack.com/...
```

**Timeline:** Week 7

---

#### 5.5.3 Web Interface

**Feature:** Visual dashboard for backup management

**Technology:**
- React + TypeScript
- Tailwind CSS
- D3.js for visualizations
- WebSocket for real-time updates

**Features:**
- Visual backup timeline
- One-click backup/restore
- Real-time progress
- Storage usage charts
- Configuration editor
- Backup comparison

**Timeline:** Week 8-10

---

## 6. Technical Specification

### 6.1 Technology Stack

**Primary Language:** Go 1.21+

**Core Dependencies:**
```go
// CLI & Configuration
github.com/spf13/cobra v1.8.0        // CLI framework
github.com/spf13/viper v1.18.2       // Configuration management

// Database Drivers
github.com/go-sql-driver/mysql v1.7.1  // MySQL driver
github.com/lib/pq v1.10.9              // PostgreSQL driver (Phase 2)

// Compression
github.com/klauspost/compress v1.17.4  // Zstd compression

// Scheduling & Automation
github.com/robfig/cron/v3 v3.0.1      // Cron scheduling

// Cloud Storage (Phase 3)
github.com/aws/aws-sdk-go v1.49.0           // AWS S3
cloud.google.com/go/storage v1.36.0         // Google Cloud Storage

// Docker (Phase 4)
github.com/docker/docker v24.0.7+incompatible
github.com/docker/go-connections v0.4.0

// Utilities
github.com/fatih/color v1.16.0                  // Colored output
github.com/schollz/progressbar/v3 v3.14.1      // Progress bars
go.uber.org/zap v1.26.0                        // Structured logging
github.com/joho/godotenv v1.5.1                // .env support
```

**Development Tools:**
- Make (build automation)
- GitHub Actions (CI/CD)
- golangci-lint (code quality)

### 6.2 System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      CLI Layer                               │
│              (Cobra Commands & Flags)                        │
└────────────┬────────────────────────────────────────────────┘
             │
┌────────────▼────────────────────────────────────────────────┐
│                  Application Layer                           │
│  ┌────────────────────────────────────────────────────┐     │
│  │  Command Handlers                                  │     │
│  │  - backup, restore, schedule, etc.                │     │
│  └────────────────────────────────────────────────────┘     │
└────────────┬────────────────────────────────────────────────┘
             │
┌────────────▼────────────────────────────────────────────────┐
│                    Service Layer                             │
│  ┌──────────────┐  ┌──────────────┐  ┌─────────────┐       │
│  │   Backup     │  │   Config     │  │  Scheduler  │       │
│  │   Service    │  │   Service    │  │   Service   │       │
│  └──────────────┘  └──────────────┘  └─────────────┘       │
└────────────┬────────────────────────────────────────────────┘
             │
┌────────────▼────────────────────────────────────────────────┐
│               Database Abstraction Layer                     │
│  ┌──────────────────────────────────────────────────┐       │
│  │  Database Interface                              │       │
│  │  - Connect(), Backup(), Restore(), etc.         │       │
│  └──────────────┬────────────────┬──────────────────┘       │
│                 │                │                           │
│        ┌────────▼─────┐  ┌──────▼───────┐                  │
│        │    MySQL     │  │  PostgreSQL  │                  │
│        │  Impl        │  │    Impl      │                  │
│        └──────────────┘  └──────────────┘                  │
└─────────────────────────────────────────────────────────────┘
```

### 6.3 Directory Structure

```
cadangkan/
├── cmd/
│   └── cadangkan/
│       ├── main.go                 # Entry point
│       └── cmd/                    # Cobra commands
│           ├── root.go
│           ├── add.go
│           ├── backup.go
│           ├── restore.go
│           ├── schedule.go
│           ├── service.go
│           ├── status.go
│           └── ...
├── internal/
│   ├── backup/
│   │   ├── service.go             # Backup service
│   │   ├── metadata.go            # Metadata handling
│   │   └── executor.go            # Backup execution
│   ├── restore/
│   │   ├── service.go             # Restore service
│   │   └── executor.go            # Restore execution
│   ├── config/
│   │   ├── config.go              # Config structures
│   │   ├── loader.go              # Load/save config
│   │   ├── encryption.go          # Credential encryption
│   │   └── validation.go          # Config validation
│   ├── storage/
│   │   ├── storage.go             # Storage interface
│   │   ├── local.go               # Local storage
│   │   ├── s3.go                  # AWS S3 (Phase 3)
│   │   └── gcs.go                 # Google Cloud Storage (Phase 3)
│   ├── scheduler/
│   │   ├── scheduler.go           # Job scheduler
│   │   ├── retention.go           # Retention policies
│   │   └── executor.go            # Job execution
│   ├── service/
│   │   ├── service.go             # Service interface
│   │   ├── systemd.go             # Linux systemd
│   │   └── launchd.go             # macOS launchd
│   ├── docker/
│   │   └── docker.go              # Docker integration (Phase 4)
│   └── notification/
│       ├── notification.go        # Notification interface (Phase 4)
│       ├── email.go               # Email notifications
│       └── webhook.go             # Webhook notifications
├── pkg/
│   ├── database/
│   │   ├── database.go            # Database interface
│   │   ├── mysql/
│   │   │   ├── client.go          # MySQL client
│   │   │   ├── dump.go            # mysqldump wrapper
│   │   │   └── restore.go         # mysql restore wrapper
│   │   └── postgres/
│   │       ├── client.go          # PostgreSQL client (Phase 2)
│   │       ├── dump.go            # pg_dump wrapper
│   │       └── restore.go         # psql restore wrapper
│   ├── compression/
│   │   ├── compression.go         # Compression interface
│   │   ├── gzip.go                # Gzip compression
│   │   └── zstd.go                # Zstandard compression
│   ├── logger/
│   │   └── logger.go              # Structured logging
│   └── health/
│       └── health.go              # Health monitoring
├── scripts/
│   ├── install.sh                 # Installation script
│   ├── build.sh                   # Build script
│   └── release.sh                 # Release script
├── docs/
│   ├── README.md
│   ├── INSTALLATION.md
│   ├── MYSQL.md
│   ├── POSTGRES.md
│   ├── CONFIGURATION.md
│   ├── COMMANDS.md
│   └── TROUBLESHOOTING.md
├── .github/
│   └── workflows/
│       ├── test.yml               # Test workflow
│       ├── build.yml              # Build workflow
│       └── release.yml            # Release workflow
├── go.mod
├── go.sum
├── Makefile
├── LICENSE
└── README.md
```

### 6.4 Data Models

#### Configuration

```go
// Config represents the main configuration file
type Config struct {
    Version   string                  `yaml:"version"`
    Defaults  Defaults                `yaml:"defaults"`
    Databases map[string]*Database    `yaml:"databases"`
}

// Defaults contains default settings
type Defaults struct {
    BackupDir   string      `yaml:"backup_dir"`
    Compression string      `yaml:"compression"`
    Retention   *Retention  `yaml:"retention"`
}

// Database represents a database configuration
type Database struct {
    Name          string         `yaml:"name"`
    Type          string         `yaml:"type"` // mysql, postgres
    Host          string         `yaml:"host"`
    Port          int            `yaml:"port"`
    Database      string         `yaml:"database"`
    User          string         `yaml:"user"`
    PasswordFile  string         `yaml:"password_file,omitempty"`
    Schedule      *Schedule      `yaml:"schedule,omitempty"`
    Retention     *Retention     `yaml:"retention,omitempty"`
    BackupOptions *BackupOptions `yaml:"backup_options,omitempty"`
    Storage       *Storage       `yaml:"storage,omitempty"`
}

// Schedule represents backup schedule
type Schedule struct {
    Enabled bool   `yaml:"enabled"`
    Cron    string `yaml:"cron"`
}

// Retention represents retention policy
type Retention struct {
    Daily   int  `yaml:"daily"`
    Weekly  int  `yaml:"weekly"`
    Monthly int  `yaml:"monthly"`
    KeepAll bool `yaml:"keep_all"`
}

// BackupOptions represents backup options
type BackupOptions struct {
    Tables         []string `yaml:"tables,omitempty"`
    ExcludeTables  []string `yaml:"exclude_tables,omitempty"`
    SchemaOnly     bool     `yaml:"schema_only,omitempty"`
    DataOnly       bool     `yaml:"data_only,omitempty"`
}

// Storage represents storage configuration
type Storage struct {
    Local *LocalStorage `yaml:"local,omitempty"`
    Cloud *CloudStorage `yaml:"cloud,omitempty"`
}
```

#### Backup Metadata

```go
// BackupMetadata represents backup metadata
type BackupMetadata struct {
    Version     string           `json:"version"`
    BackupID    string           `json:"backup_id"`
    Database    DatabaseInfo     `json:"database"`
    CreatedAt   time.Time        `json:"created_at"`
    CompletedAt time.Time        `json:"completed_at"`
    Duration    int64            `json:"duration_seconds"`
    Status      string           `json:"status"`
    Backup      BackupInfo       `json:"backup"`
    Options     BackupOptions    `json:"options"`
    Tool        ToolInfo         `json:"tool"`
    Error       string           `json:"error,omitempty"`
}

// DatabaseInfo contains database information
type DatabaseInfo struct {
    Type     string `json:"type"`
    Host     string `json:"host"`
    Port     int    `json:"port"`
    Database string `json:"database"`
    Version  string `json:"version"`
}

// BackupInfo contains backup file information
type BackupInfo struct {
    File        string `json:"file"`
    SizeBytes   int64  `json:"size_bytes"`
    SizeHuman   string `json:"size_human"`
    Compression string `json:"compression"`
    Checksum    string `json:"checksum"`
}
```

### 6.5 Key Algorithms

#### Backup Process

```
1. Load Configuration
   - Read config file
   - Decrypt credentials
   - Validate database config

2. Connect to Database
   - Establish connection
   - Verify connectivity
   - Get database info

3. Prepare Backup
   - Generate backup ID (timestamp)
   - Create backup directory
   - Check disk space

4. Execute Backup
   - Build dump command (mysqldump/pg_dump)
   - Start command process
   - Stream output to compression
   - Calculate checksum
   - Track progress

5. Generate Metadata
   - Collect backup information
   - Calculate duration
   - Write metadata JSON

6. Cleanup
   - Close connections
   - Apply retention policy
   - Log completion

7. Post-Backup (if configured)
   - Upload to cloud storage
   - Send notifications
```

#### Retention Policy Algorithm

```
1. List All Backups
   - Get all backup files
   - Parse metadata
   - Sort by date (newest first)

2. Categorize Backups
   - Identify daily backups (most recent per day)
   - Identify weekly backups (most recent per week, Sunday)
   - Identify monthly backups (most recent per month, 1st)

3. Apply Policy
   - Keep last N daily backups
   - Keep last N weekly backups
   - Keep last N monthly backups
   - Mark preserved backups
   - Identify backups to delete

4. Execute Cleanup
   - Delete old backup files
   - Delete metadata files
   - Update storage usage
   - Log deleted backups
```

### 6.6 Performance Requirements

**Backup Performance:**
- 1 GB database: <5 minutes
- 10 GB database: <30 minutes
- 100 GB database: <3 hours

**Resource Usage:**
- Memory: <500 MB during backup
- CPU: 1-2 cores
- Disk I/O: Optimize with streaming

**Reliability:**
- Backup success rate: >99%
- Service uptime: >99.9%
- Crash recovery: Automatic retry

### 6.7 Security Considerations

**Credential Storage:**
- AES-256 encryption for stored passwords
- Support for system keychains (macOS Keychain, Linux Secret Service)
- Environment variable support
- Never log credentials

**Network Security:**
- SSL/TLS for all database connections
- Verify SSL certificates
- Support SSH tunneling (future)

**File Permissions:**
- Config files: 0600 (owner read/write only)
- Backup files: 0600
- Service runs as dedicated user (Linux)

**Backup Encryption:**
- Optional AES-256 encryption (Phase 3)
- Encrypted cloud uploads
- Secure key management

### 6.8 Error Handling

**Connection Errors:**
- Retry 3 times with exponential backoff
- Clear error messages
- Suggest fixes (check host, credentials, firewall)

**Backup Errors:**
- Rollback on failure
- Clean up partial backups
- Detailed error logging
- Preserve previous successful backup

**Restore Errors:**
- Validate before restore
- Transaction-based restore when possible
- Rollback on error
- Backup target before restore (optional)

**Service Errors:**
- Automatic restart on crash
- Error notifications
- Detailed logging
- Health monitoring

---

## 7. User Experience

### 7.1 Installation Experience

**Goal:** From zero to first backup in 5 minutes

**Installation Methods:**

**1. Quick Install (Recommended):**
```bash
curl -sSL https://cadangkan.dev/install.sh | bash
```

**2. Homebrew (macOS):**
```bash
brew tap yourusername/cadangkan
brew install cadangkan
```

**3. apt (Debian/Ubuntu):**
```bash
sudo add-apt-repository ppa:yourusername/cadangkan
sudo apt update
sudo apt install cadangkan
```

**4. Binary Download:**
```bash
# Download from GitHub releases
wget https://github.com/yourusername/cadangkan/releases/download/v1.0.0/cadangkan-linux-amd64
chmod +x cadangkan-linux-amd64
sudo mv cadangkan-linux-amd64 /usr/local/bin/cadangkan
```

**First-Run Experience:**
```
$ cadangkan

Welcome to Cadangkan! 🛡️

It looks like this is your first time running Cadangkan.
Let's get you set up in a few simple steps.

Would you like to:
  1. Add your first database
  2. Read the quick start guide
  3. See available commands

Choose (1-3): 1

Great! Let's add your first database.

What type of database?
  1. MySQL / MariaDB
  2. PostgreSQL
  3. Other (coming soon)

Choose (1-2): 1

MySQL Configuration
═══════════════════════════════════════════════════════

Database name (e.g., production): production
Host: mysql.example.com
Port [3306]: 
Database: myapp
User: backup_user
Password: ••••••••

Testing connection... ✓ Connected successfully!

Would you like to:
  1. Create a backup now
  2. Schedule automatic backups
  3. Configure retention policies

Choose (1-3): 1

Creating backup... ✓ Complete!

✓ Your first backup is ready!

Backup ID: 2025-01-15-143022
Size: 234 MB
Location: ~/.cadangkan/backups/production/

Next steps:
  cadangkan schedule production --daily    # Schedule backups
  cadangkan restore production             # Restore backup
  cadangkan --help                         # See all commands

Happy backing up! 🎉
```

### 7.2 Daily Usage Patterns

**Pattern 1: One-Time Backup Before Deployment**
```bash
# Quick backup before deploying changes
$ cadangkan backup production

✓ Backup complete in 2m 34s
Backup ID: 2025-01-15-143022
```

**Pattern 2: Check Backup Status**
```bash
# Quick status check
$ cadangkan status

Cadangkan Status
════════════════════════════════════════════════════════
✓ All systems operational
✓ Last backup: 2 hours ago
✓ Next backup: in 22 hours
```

**Pattern 3: Clone for Development**
```bash
# Get latest production data locally
$ cadangkan clone production --to-docker

✓ Cloning production to local Docker...
✓ Complete! Database running on localhost:3307
```

**Pattern 4: Restore After Problem**
```bash
# Quick restore of latest backup
$ cadangkan restore production

⚠  This will restore production
Continue? [y/N]: y

✓ Restore complete in 2m 45s
```

### 7.3 CLI Design Principles

**1. Clear and Concise Commands**
- Verb-noun structure: `cadangkan backup production`
- Consistent naming: `add`, `remove`, `list`, `show`
- Short aliases: `cadangkan ls` = `cadangkan list`

**2. Progressive Disclosure**
- Common operations are simple
- Advanced options available via flags
- Help text at every level

**3. Confirmation for Destructive Actions**
- `restore` requires confirmation
- `delete` requires confirmation
- `--yes` flag to skip confirmations

**4. Clear Output**
- Use colors for status (green = success, red = error)
- Progress bars for long operations
- Emoji for visual clarity (✓, ✗, ⚠)
- Human-readable sizes (MB, GB not bytes)

**5. Helpful Error Messages**
```bash
$ cadangkan backup nonexistent

✗ Error: Database 'nonexistent' not found

Available databases:
  - production
  - staging

Did you mean:
  cadangkan backup production
  cadangkan backup staging

Or add a new database:
  cadangkan add mysql nonexistent
```

### 7.4 Configuration Experience

**Simple Configuration (90% of users):**
```yaml
# ~/.cadangkan/config.yaml
databases:
  production:
    type: mysql
    host: mysql.example.com
    database: myapp
    user: backup_user
    schedule:
      enabled: true
      cron: "0 2 * * *"
```

**Advanced Configuration (10% of users):**
```yaml
version: "1.0"

defaults:
  compression: zstd
  retention:
    daily: 14
    weekly: 8
    monthly: 24

databases:
  production:
    type: mysql
    host: mysql.example.com
    port: 3306
    database: myapp
    user: backup_user
    
    schedule:
      enabled: true
      cron: "0 */6 * * *"
    
    backup_options:
      exclude_tables:
        - sessions
        - cache
    
    storage:
      local:
        enabled: true
        path: ~/.cadangkan/backups/production
      cloud:
        enabled: true
        provider: s3
        bucket: my-backups
        region: us-east-1
    
    notifications:
      on_failure: true
      channels:
        - type: email
          to: admin@example.com
        - type: slack
          webhook_url: ${SLACK_WEBHOOK}
```

---

## 8. Development Roadmap

### 8.1 Development Phases

#### Phase 0: Pre-Development (Week 0)
**Duration:** 3-5 days  
**Goal:** Project setup and planning

**Tasks:**
- [x] Finalize product specification
- [x] Create GitHub repository
- [x] Set up project structure
- [x] Initialize Go module
- [x] Set up CI/CD pipeline
- [x] Create development environment

**Deliverables:**
- Project repository
- Initial README
- Development setup guide

---

#### Phase 1: MySQL MVP (Weeks 1-3)
**Duration:** 3 weeks  
**Goal:** Working MySQL backup and restore

**Week 1: Core Backup**
- [x] MySQL client implementation
- [x] Backup execution (mysqldump)
- [x] Local storage
- [x] Compression (gzip)
- [x] Metadata generation
- [ ] CLI: `backup` command

**Week 2: Configuration & Restore**
- [ ] Configuration system (YAML)
- [ ] Credential encryption
- [ ] Multi-database support
- [ ] Restore implementation
- [ ] CLI: `add`, `restore`, `list` commands

**Week 3: Automation & Polish**
- [ ] Cron scheduling
- [ ] System service (systemd/launchd)
- [ ] Retention policies
- [ ] Status/health monitoring
- [ ] CLI: `schedule`, `service`, `status` commands
- [ ] Testing and bug fixes
- [ ] Documentation

**Deliverables:**
- v0.1.0 MVP release
- Working MySQL backup/restore
- Automated scheduling
- Basic documentation

**Success Criteria:**
- [ ] Can backup MySQL database in <5min
- [ ] Can restore backup successfully
- [ ] Scheduled backups work reliably
- [ ] 5+ beta users testing successfully

---

#### Phase 2: PostgreSQL Support (Weeks 4-5)
**Duration:** 2 weeks  
**Goal:** Add PostgreSQL support

**Week 4: PostgreSQL Implementation**
- [ ] PostgreSQL client
- [ ] Backup execution (pg_dump)
- [ ] Restore execution (psql)
- [ ] Schema handling
- [ ] Extension support

**Week 5: Testing & Integration**
- [ ] Test with various PostgreSQL versions
- [ ] Test with cloud providers (Supabase, etc.)
- [ ] Update documentation
- [ ] PostgreSQL examples

**Deliverables:**
- v0.2.0 release
- Full PostgreSQL support
- PostgreSQL documentation

**Success Criteria:**
- [ ] Can backup/restore PostgreSQL
- [ ] Works with Supabase
- [ ] 20+ active users

---

#### Phase 3: Cloud Storage (Week 6)
**Duration:** 1 week  
**Goal:** Add cloud storage support

**Tasks:**
- [ ] S3 integration
- [ ] Google Cloud Storage integration
- [ ] Cloudflare R2 integration
- [ ] Automatic upload logic
- [ ] Cost optimization features
- [ ] Cloud storage documentation

**Deliverables:**
- v0.3.0 release
- Cloud storage support
- Multi-cloud documentation

**Success Criteria:**
- [ ] Can upload to S3/GCS/R2
- [ ] Automatic upload works
- [ ] 50+ active users

---

#### Phase 4: Clone & Notifications (Week 7)
**Duration:** 1 week  
**Goal:** Add clone and notification features

**Tasks:**
- [ ] Docker integration
- [ ] Clone to Docker implementation
- [ ] Email notifications
- [ ] Slack webhook notifications
- [ ] Discord webhook notifications
- [ ] Clone documentation

**Deliverables:**
- v0.4.0 release
- Clone functionality
- Notification system

**Success Criteria:**
- [ ] Can clone to Docker
- [ ] Notifications work
- [ ] 100+ active users

---

#### Phase 5: Web Interface (Weeks 8-10)
**Duration:** 3 weeks  
**Goal:** Build visual dashboard

**Week 8: Backend API**
- [ ] REST API for Cadangkan operations
- [ ] WebSocket for real-time updates
- [ ] Authentication system

**Week 9: Frontend Development**
- [ ] React + TypeScript setup
- [ ] Backup timeline (D3.js)
- [ ] Database management UI
- [ ] Backup/restore operations

**Week 10: Polish & Launch**
- [ ] Storage usage charts
- [ ] Configuration editor
- [ ] User onboarding
- [ ] Documentation

**Deliverables:**
- v1.0.0 release
- Web interface
- Complete feature set

**Success Criteria:**
- [ ] Web UI works reliably
- [ ] 200+ active users
- [ ] 500+ GitHub stars

---

### 8.2 Release Schedule

**v0.1.0 (Week 3)** - MySQL MVP
- Core backup/restore functionality
- Automated scheduling
- Local storage only

**v0.2.0 (Week 5)** - PostgreSQL Support
- Full PostgreSQL support
- Schema-aware backups
- Supabase compatibility

**v0.3.0 (Week 6)** - Cloud Storage
- S3, GCS, R2 support
- Automatic upload
- Cost optimization

**v0.4.0 (Week 7)** - Clone & Notifications
- Docker integration
- Clone to local
- Email/Slack/Discord notifications

**v1.0.0 (Week 10)** - Production Ready
- Web interface
- All core features complete
- Comprehensive documentation
- Production-grade reliability

**v1.1.0 (Month 4)** - Advanced Features
- MongoDB support
- Data sanitization
- Backup testing/verification
- Advanced monitoring

**v2.0.0 (Month 6)** - Enterprise Features
- Multi-region replication
- Incremental backups
- Compliance reporting
- Team collaboration features

---

## 9. Success Metrics

### 9.1 Key Performance Indicators (KPIs)

#### User Adoption Metrics

**Month 1 (MVP Launch):**
- [ ] 50+ GitHub stars
- [ ] 10+ active users
- [ ] 3+ contributors
- [ ] 20+ GitHub issues/discussions

**Month 3:**
- [ ] 200+ GitHub stars
- [ ] 100+ active users
- [ ] 10+ contributors
- [ ] 100+ community messages

**Month 6:**
- [ ] 500+ GitHub stars
- [ ] 500+ active users
- [ ] 20+ contributors
- [ ] Featured in 3+ tech publications

**Month 12:**
- [ ] 1000+ GitHub stars
- [ ] 2000+ active users
- [ ] 50+ contributors
- [ ] 10+ production case studies

#### Technical Metrics

**Reliability:**
- Backup success rate: >99%
- Service uptime: >99.9%
- Average backup time (1GB): <5 minutes
- Crash rate: <0.1%

**Performance:**
- Memory usage: <500 MB
- CPU usage: <50% of 2 cores
- Installation time: <2 minutes
- Time to first backup: <5 minutes

**Quality:**
- Test coverage: >80%
- Documentation completeness: >90%
- User satisfaction: >4.5/5
- Bug fix time: <48 hours (P0), <7 days (P1)

#### Community Metrics

**Engagement:**
- Discord/Slack members: 100+ (Month 6)
- Monthly contributors: 5+ (Month 6)
- Documentation contributions: 20+ (Month 6)
- Blog posts/tutorials: 10+ (Month 12)

**Growth:**
- Month-over-month user growth: 20%
- Weekly active users: 60% of total
- Retention rate (30 days): >70%
- Referral rate: >30%

### 9.2 Business Metrics

**Cost Savings (for users):**
- Individual users: $25-100/month saved
- Small teams: $100-500/month saved
- Agencies: $500-2000/month saved
- Community total: $150K+ saved (Month 6)

**Market Penetration:**
- MySQL users reached: 0.01% (1,000 users)
- PostgreSQL users reached: 0.01% (500 users)
- Market share of backup tools: 1%

**Sustainability:**
- Sponsorship revenue: $500/month (Month 6)
- Consulting inquiries: 5/month
- Enterprise interest: 3+ companies

### 9.3 User Satisfaction Metrics

**Net Promoter Score (NPS):**
- Target: >50 (promoters - detractors)
- Measurement: Monthly survey

**User Feedback:**
- Positive feedback: >80%
- Feature requests addressed: >60%
- Bug reports resolved: >95%

**Common User Quotes (Goal):**
- "Saved me hours of setup time"
- "Just works, no configuration headaches"
- "Saved my team $500/month"
- "Restored my database in minutes during outage"

### 9.4 Measurement Tools

**Usage Analytics (Opt-in):**
```go
// Anonymous telemetry
type TelemetryEvent struct {
    EventType   string    // backup, restore, schedule
    DatabaseType string    // mysql, postgres
    Success     bool
    Duration    int64
    Version     string
    OS          string
}
```

**Analytics Dashboard:**
- GitHub Stars tracking
- npm downloads (if Node.js tool)
- Go module downloads
- Docker pulls
- Website analytics

**Community Tracking:**
- GitHub Issues/PRs
- Discord/Slack activity
- Email list size
- Social media mentions

---

## 10. Go-to-Market Strategy

### 10.1 Target Channels

#### Primary Channels (Month 1-3)

**1. Developer Communities**
- Reddit: r/golang, r/mysql, r/PostgreSQL, r/selfhosted
- Hacker News (Show HN)
- Dev.to
- Hashnode
- Product Hunt

**2. Social Media**
- Twitter/X: #golang, #mysql, #postgresql, #devops
- LinkedIn: Developer groups
- YouTube: Demo videos

**3. Database Communities**
- MySQL forums
- PostgreSQL mailing lists
- Supabase Discord
- PlanetScale community
- Railway Discord

**4. Content Marketing**
- Technical blog posts
- Tutorial videos
- Comparison guides
- Case studies

#### Secondary Channels (Month 4-6)

**5. Partnerships**
- Database provider partnerships (Railway, Render, etc.)
- Tool integrations (listed in marketplaces)
- Conference talks
- Podcast interviews

**6. SEO & Content**
- Documentation SEO optimization
- "How to backup MySQL" guides
- "PlanetScale alternatives" content
- Comparison pages

### 10.2 Launch Plan

#### Pre-Launch (2 weeks before v0.1.0)

**Week -2:**
- [ ] Create landing page (cadangkan.dev)
- [ ] Write launch blog post
- [ ] Create demo video (2-3 minutes)
- [ ] Prepare social media posts
- [ ] Set up analytics
- [ ] Polish README with demo GIF

**Week -1:**
- [ ] Beta testing with 5-10 users
- [ ] Fix critical bugs
- [ ] Finalize documentation
- [ ] Create Show HN draft
- [ ] Prepare email announcement
- [ ] Join relevant Discord servers

#### Launch Day

**Hour 0-1:**
- [ ] Tag v0.1.0 release
- [ ] Publish to package managers
- [ ] Launch landing page
- [ ] Post Show HN
- [ ] Tweet announcement
- [ ] Post to LinkedIn

**Hour 2-6:**
- [ ] Share on Reddit (r/golang, r/mysql)
- [ ] Post in Discord communities
- [ ] Share on Dev.to
- [ ] Email subscribers
- [ ] Monitor and respond to feedback

**Day 2-7:**
- [ ] Respond to all comments/questions
- [ ] Fix reported bugs immediately
- [ ] Write follow-up content
- [ ] Thank early adopters
- [ ] Track metrics

#### Post-Launch (Week 2-4)

**Week 2:**
- [ ] Follow-up blog: "What we learned"
- [ ] User spotlight (interview early user)
- [ ] Technical deep dive post
- [ ] Product Hunt launch

**Week 3:**
- [ ] Tutorial: "Backing up MySQL in 5 minutes"
- [ ] Tutorial: "Migrating from manual backups"
- [ ] Comparison: "Cadangkan vs paid backups"

**Week 4:**
- [ ] v0.1.1 with bug fixes
- [ ] Community update
- [ ] Plan v0.2.0 announcement

### 10.3 Content Strategy

#### Blog Post Ideas

**Educational:**
- "The Complete Guide to Database Backups in 2025"
- "MySQL Backup Best Practices"
- "How to Recover from Database Failure"
- "Understanding Backup Retention Policies"

**Comparative:**
- "Cadangkan vs AWS RDS Automated Backups"
- "Free Database Backup Solutions Compared"
- "Why We Built Cadangkan"

**Technical:**
- "Building a Database Abstraction Layer in Go"
- "Implementing Cron Scheduling in Go"
- "Optimizing Database Backup Performance"

**Case Studies:**
- "How [Company] Saves $500/month with Cadangkan"
- "Restoring 100GB Database in 30 Minutes"
- "Managing Backups for 50 Client Databases"

#### Video Content

**Tutorial Videos:**
- "Getting Started with Cadangkan" (5 min)
- "Setting Up Automated MySQL Backups" (10 min)
- "Cloning Your Database to Local Development" (7 min)
- "Configuring Cloud Storage for Backups" (12 min)

**Demo Videos:**
- "Cadangkan in 60 Seconds" (1 min)
- "Full Feature Walkthrough" (15 min)
- "Disaster Recovery Demo" (8 min)

### 10.4 Messaging Framework

**Value Propositions by Persona:**

**Solo Developer:**
> "Free, automated database backups in 5 minutes. No credit card, no setup headaches."

**Startup CTO:**
> "Save $300-500/month on database backups. One tool for all your databases."

**Freelancer:**
> "Manage backups for all your clients from one place. Professional setup, zero cost."

**DevOps Engineer:**
> "Unified backup strategy across MySQL, PostgreSQL, and more. Infrastructure-as-code ready."

**Key Messages:**
1. **Simple:** "5-minute setup, one-command backups"
2. **Cost-effective:** "Free forever, save hundreds per month"
3. **Reliable:** "Sleep soundly knowing your data is safe"
4. **Universal:** "One tool for all your databases"
5. **Open Source:** "Community-driven, transparent, trustworthy"

### 10.5 Competitive Positioning

**Against Paid Cloud Backups:**
> "Get the same features as AWS RDS backups for $0/month. Keep your data under your control."

**Against Manual Scripts:**
> "Stop maintaining bash scripts. Get a professional backup solution in 5 minutes."

**Against Database-Specific Tools:**
> "One tool for MySQL, PostgreSQL, and more. Learn once, use everywhere."

---

## 11. Risk Assessment

### 11.1 Technical Risks

#### Risk 1: Database Compatibility Issues

**Risk:** Different MySQL/PostgreSQL versions have incompatible backup formats

**Impact:** High - Core functionality broken  
**Probability:** Medium

**Mitigation:**
- Test against multiple database versions
- Document supported versions clearly
- Implement version detection
- Graceful degradation for unsupported features

**Contingency:**
- Quick patch releases
- Clear error messages
- Version compatibility matrix in docs

---

#### Risk 2: Large Database Performance

**Risk:** Backups of 100GB+ databases are too slow or fail

**Impact:** Medium - Limits use cases  
**Probability:** Medium

**Mitigation:**
- Streaming backup implementation
- Parallel dump options (mydumper alternative)
- Progress tracking and resume capability
- Performance testing with large databases

**Contingency:**
- Document performance characteristics
- Recommend parallel backup tools for large DBs
- Implement chunked backups (Phase 2)

---

#### Risk 3: Credential Security Breach

**Risk:** Stored credentials are compromised

**Impact:** Critical - User data at risk  
**Probability:** Low

**Mitigation:**
- AES-256 encryption for stored passwords
- File permissions (0600)
- System keychain integration
- Security audit before launch
- Recommend environment variables

**Contingency:**
- Immediate security patch
- Clear communication to users
- Force credential rotation

---

### 11.2 Market Risks

#### Risk 4: Low User Adoption

**Risk:** Developers don't discover or adopt Cadangkan

**Impact:** High - Project not sustainable  
**Probability:** Medium

**Mitigation:**
- Strong launch strategy (multiple channels)
- Clear value proposition
- Excellent documentation
- Active community engagement
- SEO optimization

**Contingency:**
- Adjust marketing strategy
- Add requested features quickly
- Partner with database providers
- Focus on niche (MySQL first, then expand)

---

#### Risk 5: Competition from Major Players

**Risk:** AWS/Google launch free backup tools

**Impact:** High - Market disruption  
**Probability:** Low

**Mitigation:**
- Focus on multi-database support
- Emphasize self-hosted benefits
- Build loyal community
- Add unique features (clone, sanitization)

**Contingency:**
- Pivot to enterprise features
- Focus on specific databases
- Emphasize open source benefits

---

### 11.3 Resource Risks

#### Risk 6: Maintainer Burnout

**Risk:** Solo developer can't keep up with maintenance

**Impact:** High - Project abandonment  
**Probability:** Medium

**Mitigation:**
- Start with MVP, not perfect
- Build community early
- Document everything
- Accept contributors
- Set realistic expectations

**Contingency:**
- Find co-maintainers
- Reduce scope
- Seek sponsorship
- Hand off to community

---

#### Risk 7: Insufficient Time for Development

**Risk:** Competing priorities delay development

**Impact:** Medium - Delayed launch  
**Probability:** Medium

**Mitigation:**
- Clear 3-week MVP timeline
- Focus on essentials only
- Time-box features
- Use existing libraries
- Skip non-critical features

**Contingency:**
- Extend timeline
- Release smaller MVP
- Focus on core features only
- Get help from community

---

### 11.4 Legal Risks

#### Risk 8: Licensing Issues with Dependencies

**Risk:** Using incompatible licenses for dependencies

**Impact:** Medium - Legal complications  
**Probability:** Low

**Mitigation:**
- Review all dependency licenses
- Use MIT/Apache 2.0 compatible licenses
- Document all dependencies
- License compliance check

**Contingency:**
- Replace problematic dependencies
- Seek legal advice
- Re-license if necessary

---

### 11.5 Risk Mitigation Summary

**High Priority Risks:**
1. Database compatibility → Extensive testing
2. User adoption → Strong launch strategy
3. Credential security → Security-first design

**Medium Priority Risks:**
4. Large database performance → Optimize and document
5. Maintainer burnout → Community building
6. Development time → Realistic scope

**Low Priority Risks:**
7. Competition → Differentiation focus
8. Licensing → Due diligence

---

## 12. Appendices

### 12.1 Example Configuration Files

#### Minimal Configuration

```yaml
# ~/.cadangkan/config.yaml
version: "1.0"

databases:
  production:
    type: mysql
    host: mysql.example.com
    port: 3306
    database: myapp
    user: backup_user
```

#### Standard Configuration

```yaml
version: "1.0"

defaults:
  backup_dir: "~/.cadangkan/backups"
  compression: "gzip"
  retention:
    daily: 7
    weekly: 4
    monthly: 12

databases:
  production:
    type: mysql
    host: mysql.example.com
    port: 3306
    database: myapp
    user: backup_user
    
    schedule:
      enabled: true
      cron: "0 2 * * *"  # Daily at 2 AM
    
    backup_options:
      exclude_tables:
        - sessions
        - cache
```

#### Advanced Configuration

```yaml
version: "1.0"

defaults:
  compression: "zstd"
  retention:
    daily: 14
    weekly: 8
    monthly: 24

databases:
  production:
    type: mysql
    host: mysql.example.com
    port: 3306
    database: myapp
    user: backup_user
    
    schedule:
      enabled: true
      cron: "0 */6 * * *"  # Every 6 hours
    
    backup_options:
      exclude_tables:
        - audit_logs
        - sessions
    
    storage:
      local:
        enabled: true
        path: ~/.cadangkan/backups/production
      
      cloud:
        enabled: true
        provider: s3
        bucket: my-db-backups
        region: us-east-1
        path: production/
        upload_after: 24h  # Upload only after 24 hours
    
    notifications:
      on_success: false
      on_failure: true
      channels:
        - type: email
          to: admin@example.com
          smtp_host: smtp.gmail.com
          smtp_port: 587
          smtp_user: notifications@example.com
        
        - type: slack
          webhook_url: https://hooks.slack.com/services/XXX
        
        - type: webhook
          url: https://myapp.com/webhooks/backup
          headers:
            Authorization: "Bearer ${WEBHOOK_TOKEN}"
```

### 12.2 CLI Command Reference

```bash
# Database Management
cadangkan add mysql <name> --host=<host> --database=<db> --user=<user>
cadangkan add postgres <name> --host=<host> --database=<db> --user=<user>
cadangkan list
cadangkan info <name>
cadangkan test <name>
cadangkan remove <name>

# Backup Operations
cadangkan backup <name>
cadangkan backup <name> --tables=users,orders
cadangkan backup <name> --exclude-tables=logs
cadangkan backup <name> --schema-only
cadangkan backup <name> --compression=zstd

# Backup Management
cadangkan backups list <name>
cadangkan backups info <name> <backup-id>
cadangkan backups delete <name> <backup-id>
cadangkan backups validate <name> <backup-id>

# Restore Operations
cadangkan restore <name>
cadangkan restore <name> --from=<backup-id>
cadangkan restore <name> --from=<backup-id> --to=<target>
cadangkan restore <name> --dry-run
cadangkan restore <name> --backup-first

# Clone Operations (Phase 4)
cadangkan clone <name> --to-docker
cadangkan clone <name> --to=<connection-string>
cadangkan clone <name> --to-docker --sanitize

# Scheduling
cadangkan schedule set <name> --daily --time=02:00
cadangkan schedule set <name> --weekly --day=sunday --time=03:00
cadangkan schedule set <name> --cron="0 2 * * *"
cadangkan schedule enable <name>
cadangkan schedule disable <name>
cadangkan schedule list
cadangkan schedule next

# Service Management
sudo cadangkan service install
sudo cadangkan service start
sudo cadangkan service stop
sudo cadangkan service restart
cadangkan service status
cadangkan service logs
cadangkan service logs --follow
sudo cadangkan service uninstall

# Status & Monitoring
cadangkan status
cadangkan status <name>
cadangkan health <name>
cadangkan storage
cadangkan cleanup <name>
cadangkan cleanup <name> --dry-run

# Configuration
cadangkan config show
cadangkan config show <name>
cadangkan config edit
cadangkan config edit <name>
cadangkan config validate
cadangkan config export > config.yaml
cadangkan config import config.yaml

# Storage (Phase 3)
cadangkan storage add s3 --bucket=<bucket> --region=<region>
cadangkan storage add gcs --bucket=<bucket> --project=<project>
cadangkan storage upload <name> --backup-id=<id>
cadangkan storage download <name> --backup-id=<id>
cadangkan storage list <name> --cloud

# Utilities
cadangkan version
cadangkan --help
cadangkan <command> --help
```

### 12.3 Backup Metadata Example

```json
{
  "version": "1.0",
  "backup_id": "2025-01-15-143022",
  "database": {
    "type": "mysql",
    "host": "mysql.example.com",
    "port": 3306,
    "database": "myapp",
    "version": "8.0.35"
  },
  "created_at": "2025-01-15T14:30:22Z",
  "completed_at": "2025-01-15T14:35:56Z",
  "duration_seconds": 334,
  "status": "success",
  "backup": {
    "file": "2025-01-15-143022.sql.gz",
    "size_bytes": 245678900,
    "size_human": "234 MB",
    "compression": "gzip",
    "checksum": "sha256:a3f5b8c9d2e1f4g7h6i5j8k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6"
  },
  "options": {
    "tables": [],
    "exclude_tables": ["sessions", "cache"],
    "schema_only": false,
    "data_only": false
  },
  "tables_count": 42,
  "rows_estimated": 1500000,
  "tool": {
    "name": "cadangkan",
    "version": "0.1.0",
    "mysqldump_version": "8.0.35"
  }
}
```

### 12.4 systemd Unit File Example

```ini
# /etc/systemd/system/cadangkan.service
[Unit]
Description=Cadangkan Database Backup Service
Documentation=https://cadangkan.dev/docs
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=dbshield
Group=dbshield
WorkingDirectory=/home/dbshield
ExecStart=/usr/local/bin/cadangkan daemon
ExecReload=/bin/kill -HUP $MAINPID
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=cadangkan

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/home/dbshield/.cadangkan

[Install]
WantedBy=multi-user.target
```

### 12.5 Glossary

**Backup:** A copy of database data at a specific point in time  
**Restore:** The process of recovering data from a backup  
**Clone:** Creating a copy of a database in a different environment  
**Retention Policy:** Rules for how long to keep backups  
**Incremental Backup:** Backup containing only changes since last backup  
**Full Backup:** Complete copy of all database data  
**Dump:** Database export in SQL format  
**Metadata:** Information about the backup (size, date, checksums)  
**Cron:** Time-based job scheduler in Unix-like systems  
**systemd:** System and service manager for Linux  
**launchd:** Service management framework for macOS  
**Checksum:** Hash value used to verify file integrity  

---

## Document Information

**Document Version:** 1.0  
**Last Updated:** December 2025  
**Author:** Cadangkan Team  
**Status:** Draft  
**Review Date:** 2026-01-01

**Change Log:**
- v1.0 (2025-12-31): Initial product specification