package main

import (
	"fmt"
	"time"

	"github.com/erickhilda/cadangkan/internal/backup"
	"github.com/erickhilda/cadangkan/internal/storage"
	"github.com/erickhilda/cadangkan/pkg/database/mysql"
	"github.com/urfave/cli/v2"
)

func backupCommand() *cli.Command {
	return &cli.Command{
		Name:      "backup",
		Usage:     "Create database backup",
		ArgsUsage: "[name]",
		Flags: []cli.Flag{
			// Database type
			&cli.StringFlag{
				Name:  "type",
				Value: "mysql",
				Usage: "Database type (mysql)",
			},

			// Connection flags
			&cli.StringFlag{
				Name:  "host",
				Value: "127.0.0.1",
				Usage: "Database host (use 127.0.0.1 for Docker)",
			},
			&cli.IntFlag{
				Name:  "port",
				Value: 3306,
				Usage: "Database port",
			},
			&cli.StringFlag{
				Name:     "user",
				Required: true,
				Usage:    "Database user",
			},
			&cli.StringFlag{
				Name:  "password",
				Usage: "Database password",
			},
			&cli.StringFlag{
				Name:     "database",
				Required: true,
				Usage:    "Database name",
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
		},
		Action: runBackup,
	}
}

func runBackup(c *cli.Context) error {
	// 1. Parse and validate inputs
	dbType := c.String("type")
	if dbType != "mysql" {
		return fmt.Errorf("unsupported database type: %s (only 'mysql' is supported)", dbType)
	}

	host := c.String("host")
	port := c.Int("port")
	user := c.String("user")
	password := c.String("password")
	database := c.String("database")
	tables := c.StringSlice("tables")
	excludeTables := c.StringSlice("exclude-tables")
	schemaOnly := c.Bool("schema-only")
	compression := c.String("compression")
	outputDir := c.String("output")

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

	// 7. Execute backup with progress
	printInfo("Starting backup...")

	options := &backup.BackupOptions{
		Database:      database,
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
