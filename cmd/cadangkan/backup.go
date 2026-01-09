package main

import (
	"fmt"
	"time"

	"github.com/erickhilda/cadangkan/internal/backup"
	"github.com/erickhilda/cadangkan/internal/config"
	"github.com/erickhilda/cadangkan/internal/storage"
	"github.com/erickhilda/cadangkan/pkg/database/mysql"
	"github.com/urfave/cli/v2"
)

func backupCommand() *cli.Command {
	return &cli.Command{
		Name:      "backup",
		Usage:     "Create database backup",
		ArgsUsage: "[name]",
		Description: `Create a backup of a database.

   USAGE MODES:
     1. Named mode (from config):
        cadangkan backup <name>
        
     2. Direct mode (with flags):
        cadangkan backup --host=<host> --user=<user> --database=<db> --password=<pass>

   Flags can override config values when using named mode.`,
		Flags: []cli.Flag{
			// Database type
			&cli.StringFlag{
				Name:  "type",
				Value: "mysql",
				Usage: "Database type (mysql)",
			},

			// Connection flags (now optional for named mode)
			&cli.StringFlag{
				Name:  "host",
				Usage: "Database host (overrides config)",
			},
			&cli.IntFlag{
				Name:  "port",
				Usage: "Database port (overrides config)",
			},
			&cli.StringFlag{
				Name:  "user",
				Usage: "Database user (overrides config)",
			},
			&cli.StringFlag{
				Name:  "password",
				Usage: "Database password (overrides config)",
			},
			&cli.StringFlag{
				Name:  "database",
				Usage: "Database name (overrides config)",
			},

			// Backup options
			&cli.StringSliceFlag{
				Name:  "tables",
				Usage: "Specific tables to backup (comma-separated)",
			},
			&cli.StringSliceFlag{
				Name:  "exclude-tables",
				Usage: "Tables to exclude (comma-separated)",
			},
			&cli.BoolFlag{
				Name:  "schema-only",
				Usage: "Backup schema only (no data)",
			},
			&cli.StringFlag{
				Name:  "compression",
				Value: "gzip",
				Usage: "Compression type (gzip|none)",
			},
			&cli.StringFlag{
				Name:  "output",
				Value: "",
				Usage: "Output directory (default: ~/.cadangkan/backups)",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Show verbose output including mysqldump command",
			},
		},
		Action: runBackup,
	}
}

func runBackup(c *cli.Context) error {
	var host, user, password, database, configName string
	var port int
	var usingConfig bool

	// Check if using named mode (config) or direct mode (flags)
	if c.NArg() > 0 {
		// Named mode - load from config
		name := c.Args().Get(0)
		configName = name
		usingConfig = true

		mgr, err := config.NewManager()
		if err != nil {
			return fmt.Errorf("failed to create config manager: %w", err)
		}

		dbConfig, err := mgr.GetDatabase(name)
		if err != nil {
			printError(fmt.Sprintf("Database '%s' not found in config", name))
			fmt.Println()
			fmt.Printf("Available databases: run %scadangkan list%s\n", colorCyan, colorReset)
			fmt.Printf("Add a database:      run %scadangkan add mysql %s%s\n", colorCyan, name, colorReset)
			return err
		}

		// Load config values
		host = dbConfig.Host
		port = dbConfig.Port
		user = dbConfig.User
		database = dbConfig.Database

		// Decrypt password
		password, err = config.DecryptPassword(dbConfig.PasswordEncrypted)
		if err != nil {
			return fmt.Errorf("failed to decrypt password: %w", err)
		}

		printInfo(fmt.Sprintf("Using configuration for '%s'", name))
	} else {
		// Direct mode - use flags
		host = c.String("host")
		port = c.Int("port")
		user = c.String("user")
		password = c.String("password")
		database = c.String("database")

		// Validate required flags for direct mode
		if host == "" {
			return fmt.Errorf("--host is required when not using named mode")
		}
		if user == "" {
			return fmt.Errorf("--user is required when not using named mode")
		}
		if database == "" {
			return fmt.Errorf("--database is required when not using named mode")
		}
		if port == 0 {
			port = 3306 // Default port
		}
	}

	// Allow flags to override config values
	if c.IsSet("host") && usingConfig {
		host = c.String("host")
	}
	if c.IsSet("port") && usingConfig {
		port = c.Int("port")
	}
	if c.IsSet("user") && usingConfig {
		user = c.String("user")
	}
	if c.IsSet("password") && usingConfig {
		password = c.String("password")
	}
	if c.IsSet("database") && usingConfig {
		database = c.String("database")
	}

	// Parse backup options
	tables := c.StringSlice("tables")
	excludeTables := c.StringSlice("exclude-tables")
	schemaOnly := c.Bool("schema-only")
	compression := c.String("compression")
	outputDir := c.String("output")

	// Validate database type
	dbType := c.String("type")
	if dbType != "mysql" {
		return fmt.Errorf("unsupported database type: %s (only 'mysql' is supported)", dbType)
	}

	// 2. Check for mysqldump availability
	printInfo("Checking mysqldump availability...")
	version, err := backup.CheckMySQLDump()
	if err != nil {
		printError("mysqldump not found")
		fmt.Println("\nPlease install MySQL client tools:")
		fmt.Println("  Ubuntu/Debian: sudo apt-get install mysql-client")
		fmt.Println("  RHEL/CentOS:   sudo yum install mysql")
		fmt.Println("  macOS:         brew install mysql-client")
		return err
	}
	printSuccess(fmt.Sprintf("Found %s", version))

	// 3. Create MySQL config
	config := &mysql.Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Database: database,
		Timeout:  10 * time.Second,
	}

	// 4. Create client and connect
	printInfo(fmt.Sprintf("Connecting to %s@%s:%d...", user, host, port))
	client, err := mysql.NewClient(config)
	if err != nil {
		printError("Failed to create MySQL client")
		return err
	}

	if err := client.Connect(); err != nil {
		printError("Connection failed")
		return err
	}
	defer client.Close()

	// Get database version
	dbVersion, err := client.GetVersion()
	if err != nil {
		dbVersion = "unknown"
	}
	printSuccess(fmt.Sprintf("Connected to database (MySQL %s)", dbVersion))

	// 5. Create storage
	var localStorage *storage.LocalStorage
	if outputDir != "" {
		localStorage, err = storage.NewLocalStorage(outputDir)
	} else {
		localStorage, err = storage.NewLocalStorage("")
	}
	if err != nil {
		printError("Failed to create storage")
		return err
	}

	// 6. Create backup service
	service := backup.NewService(client, localStorage, config)

	// Enable verbose mode if requested
	verbose := c.Bool("verbose")
	if verbose {
		service.SetVerbose(true)
	}

	// 7. Execute backup with progress
	printInfo("Starting backup...")

	options := &backup.BackupOptions{
		Database:      database,
		ConfigName:    configName,
		Tables:        tables,
		ExcludeTables: excludeTables,
		SchemaOnly:    schemaOnly,
		Compression:   compression,
	}

	// Show a simple progress indicator
	done := make(chan bool)
	go showSpinner(done)

	result, err := service.Backup(options)
	done <- true

	if err != nil {
		printError("Backup failed")
		return err
	}

	// 8. Display results
	printSuccess("Backup completed!")
	fmt.Println()
	formatBackupResult(result, database)

	return nil
}
