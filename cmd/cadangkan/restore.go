package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/erickhilda/cadangkan/internal/backup"
	"github.com/erickhilda/cadangkan/internal/config"
	"github.com/erickhilda/cadangkan/internal/storage"
	"github.com/erickhilda/cadangkan/pkg/database/mysql"
	"github.com/urfave/cli/v2"
)

func restoreCommand() *cli.Command {
	return &cli.Command{
		Name:      "restore",
		Usage:     "Restore database backup",
		ArgsUsage: "[name]",
		Description: `Restore a database backup.

   USAGE MODES:
     1. Named mode (from config):
        cadangkan restore <name>
        
     2. Direct mode (with flags):
        cadangkan restore --host=<host> --user=<user> --database=<db> --password=<pass>

   Flags can override config values when using named mode.`,
		Flags: []cli.Flag{
			// Database type
			&cli.StringFlag{
				Name:  "type",
				Value: "mysql",
				Usage: "Database type (mysql)",
			},

			// Backup selection
			&cli.StringFlag{
				Name:  "from",
				Usage: "Specific backup ID to restore (default: latest)",
			},

			// Target database
			&cli.StringFlag{
				Name:  "to",
				Usage: "Target database name (overrides config database)",
			},

			// Database creation
			&cli.BoolFlag{
				Name:  "create-db",
				Usage: "Create database if it doesn't exist",
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

			// Safety options
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Validate restore without executing",
			},
			&cli.BoolFlag{
				Name:  "backup-first",
				Usage: "Backup target database before restore (only if DB exists)",
			},
			&cli.BoolFlag{
				Name:    "yes",
				Aliases: []string{"y"},
				Usage:   "Skip confirmation prompt",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Show verbose output including mysql command",
			},
		},
		Action: runRestore,
	}
}

func runRestore(c *cli.Context) error {
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

	// Get target database (--to overrides)
	targetDatabase := database
	if c.IsSet("to") {
		targetDatabase = c.String("to")
	}

	// Validate database type
	dbType := c.String("type")
	if dbType != "mysql" {
		return fmt.Errorf("unsupported database type: %s (only 'mysql' is supported)", dbType)
	}

	// Check for mysql availability
	printInfo("Checking mysql availability...")
	version, err := backup.CheckMySQL()
	if err != nil {
		printError("mysql not found")
		fmt.Println("\nPlease install MySQL client tools:")
		fmt.Println("  Ubuntu/Debian: sudo apt-get install mysql-client")
		fmt.Println("  RHEL/CentOS:   sudo yum install mysql")
		fmt.Println("  macOS:         brew install mysql-client")
		return err
	}
	printSuccess(fmt.Sprintf("Found %s", version))

	// Create MySQL config
	// Connect without specifying database so we can create/restore into any database
	mysqlConfig := &mysql.Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Database: "", // Empty - connect to server, not specific database
		Timeout:  10 * time.Second,
	}

	// Create client and connect
	printInfo(fmt.Sprintf("Connecting to %s@%s:%d...", user, host, port))
	client, err := mysql.NewClient(mysqlConfig)
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

	// Create storage
	localStorage, err := storage.NewLocalStorage("")
	if err != nil {
		printError("Failed to create storage")
		return err
	}

	// Create restore service
	service := backup.NewRestoreService(client, localStorage, mysqlConfig)

	// Enable verbose mode if requested
	verbose := c.Bool("verbose")
	if verbose {
		service.SetVerbose(true)
	}

	// Get backup ID
	backupID := c.String("from")

	// Load backup metadata for preview
	storageName := configName
	if storageName == "" {
		storageName = database
	}

	var backupEntry *storage.BackupListEntry
	if backupID == "" {
		// Get latest backup
		entry, err := localStorage.GetLatestBackup(storageName)
		if err != nil {
			printError(fmt.Sprintf("No backups found for '%s'", storageName))
			return err
		}
		backupEntry = entry
		backupID = entry.BackupID
	} else {
		// Get specific backup
		backups, err := localStorage.ListBackups(storageName)
		if err != nil {
			return fmt.Errorf("failed to list backups: %w", err)
		}

		found := false
		for _, backup := range backups {
			if backup.BackupID == backupID {
				backupEntry = &backup
				found = true
				break
			}
		}

		if !found {
			printError(fmt.Sprintf("Backup '%s' not found", backupID))
			return fmt.Errorf("backup not found")
		}
	}

	// Load full metadata
	var metadata backup.BackupMetadata
	err = localStorage.LoadMetadata(storageName, backupID, &metadata)
	if err != nil {
		return fmt.Errorf("failed to load backup metadata: %w", err)
	}

	// Check if target database exists
	dbExists, err := client.DatabaseExists(targetDatabase)
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	// Show restore preview
	fmt.Println()
	printWarning("WARNING: This will restore the database")
	if dbExists {
		printWarning(fmt.Sprintf("Current data in '%s' will be overwritten!", targetDatabase))
	} else {
		printInfo(fmt.Sprintf("Database '%s' does not exist", targetDatabase))
		if !c.Bool("create-db") {
			printError("Use --create-db to create the database")
			return fmt.Errorf("database does not exist")
		}
	}
	fmt.Println()

	fmt.Printf("Backup to restore:\n")
	fmt.Printf("  %sID:%s        %s\n", colorCyan, colorReset, backupEntry.BackupID)
	fmt.Printf("  %sCreated:%s    %s\n", colorCyan, colorReset, backupEntry.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("  %sSize:%s       %s\n", colorCyan, colorReset, backupEntry.SizeHuman)
	fmt.Printf("  %sDatabase:%s   %s\n", colorCyan, colorReset, metadata.Database.Database)
	fmt.Println()

	fmt.Printf("Target database:\n")
	fmt.Printf("  %sName:%s       %s\n", colorCyan, colorReset, targetDatabase)
	fmt.Printf("  %sHost:%s       %s:%d\n", colorCyan, colorReset, host, port)
	if dbExists {
		printInfo("Database exists - data will be overwritten")
	} else {
		printInfo("Database will be created")
	}
	fmt.Println()

	// Dry-run mode
	if c.Bool("dry-run") {
		printInfo("Dry-run mode: Validation only, no changes will be made")
		fmt.Println()
		printSuccess("Validation passed! Use without --dry-run to restore.")
		return nil
	}

	// Confirmation prompt
	if !c.Bool("yes") {
		fmt.Print("Continue? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			printInfo("Restore cancelled")
			return nil
		}
		fmt.Println()
	}

	// Backup-first option
	if c.Bool("backup-first") && dbExists {
		printInfo("Creating backup of target database before restore...")
		// TODO: Implement backup-first functionality
		printWarning("backup-first not yet implemented, skipping")
	}

	// Execute restore
	printInfo("Starting restore...")

	options := &backup.RestoreOptions{
		BackupID:         backupID,
		Database:         database,
		ConfigName:       configName,
		TargetDatabase:   targetDatabase,
		CreateDatabase:   c.Bool("create-db"),
		DryRun:           c.Bool("dry-run"),
		BackupFirst:      c.Bool("backup-first"),
		SkipConfirmation: c.Bool("yes"),
	}

	// Show spinner during restore
	done := make(chan bool)
	go showRestoreSpinner(done)

	result, err := service.Restore(options)
	done <- true

	if err != nil {
		printError("Restore failed")
		return err
	}

	// Display results
	printSuccess("Restore completed!")
	fmt.Println()
	formatRestoreResult(result, targetDatabase)

	return nil
}

// showRestoreSpinner displays a spinner during restore
func showRestoreSpinner(done chan bool) {
	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	i := 0
	for {
		select {
		case <-done:
			fmt.Print("\r") // Clear the spinner line
			return
		default:
			fmt.Printf("\r%s Restoring... ", spinner[i%len(spinner)])
			i++
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// formatRestoreResult formats and displays the restore result
func formatRestoreResult(result *backup.RestoreResult, database string) {
	fmt.Printf("  %sBackup ID:%s       %s\n", colorCyan, colorReset, result.BackupID)
	fmt.Printf("  %sTarget Database:%s %s\n", colorCyan, colorReset, database)
	fmt.Printf("  %sDuration:%s        %s\n", colorCyan, colorReset, backup.FormatDuration(result.Duration))
	fmt.Println()
	fmt.Printf("Database '%s' has been restored successfully.\n", database)
}
