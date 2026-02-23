package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/erickhilda/cadangkan/internal/backup"
	"github.com/erickhilda/cadangkan/internal/config"
	"github.com/erickhilda/cadangkan/pkg/database/mysql"
	"github.com/urfave/cli/v2"
)

func importCommand() *cli.Command {
	return &cli.Command{
		Name:      "import",
		Usage:     "Import an external SQL dump file into a configured database",
		ArgsUsage: "<config-name>",
		Description: `Import an external SQL dump file (from mysqldump, DBeaver, TablePlus, etc.)
into a database already configured in cadangkan.

   EXAMPLES:
     cadangkan import mydb --file /path/to/dump.sql
     cadangkan import mydb --file /path/to/dump.sql.gz --create-db --yes
     cadangkan import mydb --file /path/to/dump.sql --to other_db`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "file",
				Usage:    "Path to the SQL dump file (.sql or .sql.gz)",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "to",
				Usage: "Target database name (overrides config database)",
			},
			&cli.BoolFlag{
				Name:  "create-db",
				Usage: "Create database if it doesn't exist",
			},
			&cli.BoolFlag{
				Name:    "yes",
				Aliases: []string{"y"},
				Usage:   "Skip confirmation prompt",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Show mysql command being executed",
			},
		},
		Action: runImport,
	}
}

func runImport(c *cli.Context) error {
	// Require config name
	if c.NArg() < 1 {
		return fmt.Errorf("config name is required\n\nUsage: cadangkan import <config-name> --file <path>")
	}
	name := c.Args().Get(0)

	// Validate file exists
	filePath := c.String("file")
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			printError(fmt.Sprintf("File not found: %s", filePath))
			return fmt.Errorf("file not found: %s", filePath)
		}
		return fmt.Errorf("cannot access file: %w", err)
	}
	if fileInfo.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", filePath)
	}

	// Load database config
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

	// Decrypt password
	password, err := config.DecryptPassword(dbConfig.PasswordEncrypted)
	if err != nil {
		return fmt.Errorf("failed to decrypt password: %w", err)
	}

	// Detect compression from file extension
	compression := backup.CompressionNone
	lowerPath := strings.ToLower(filePath)
	if strings.HasSuffix(lowerPath, ".gz") {
		compression = backup.CompressionGzip
	}

	// Check mysql CLI availability
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

	// Determine target database
	targetDatabase := dbConfig.Database
	if c.IsSet("to") {
		targetDatabase = c.String("to")
	}

	// Connect to MySQL server (without specifying database)
	mysqlConfig := &mysql.Config{
		Host:     dbConfig.Host,
		Port:     dbConfig.Port,
		User:     dbConfig.User,
		Password: password,
		Database: "",
		Timeout:  10 * time.Second,
	}

	printInfo(fmt.Sprintf("Connecting to %s@%s:%d...", dbConfig.User, dbConfig.Host, dbConfig.Port))
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

	dbVersion, err := client.GetVersion()
	if err != nil {
		dbVersion = "unknown"
	}
	printSuccess(fmt.Sprintf("Connected to database (MySQL %s)", dbVersion))

	// Check if target database exists
	dbExists, err := client.DatabaseExists(targetDatabase)
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	if !dbExists && !c.Bool("create-db") {
		printError(fmt.Sprintf("Database '%s' does not exist", targetDatabase))
		fmt.Println("Use --create-db to create it automatically")
		return fmt.Errorf("database does not exist")
	}

	// Show confirmation summary
	fmt.Println()
	printWarning("WARNING: This will import data into the database")
	if dbExists {
		printWarning(fmt.Sprintf("Current data in '%s' may be overwritten!", targetDatabase))
	}
	fmt.Println()

	compressionLabel := "none"
	if compression == backup.CompressionGzip {
		compressionLabel = "gzip"
	}

	fmt.Printf("Import file:\n")
	fmt.Printf("  %sFile:%s        %s\n", colorCyan, colorReset, filePath)
	fmt.Printf("  %sSize:%s        %s\n", colorCyan, colorReset, backup.FormatBytes(fileInfo.Size()))
	fmt.Printf("  %sCompression:%s %s\n", colorCyan, colorReset, compressionLabel)
	fmt.Println()

	fmt.Printf("Target database:\n")
	fmt.Printf("  %sName:%s        %s\n", colorCyan, colorReset, targetDatabase)
	fmt.Printf("  %sHost:%s        %s:%d\n", colorCyan, colorReset, dbConfig.Host, dbConfig.Port)
	if dbExists {
		printInfo("Database exists - data may be overwritten")
	} else {
		printInfo("Database will be created")
	}
	fmt.Println()

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
			printInfo("Import cancelled")
			return nil
		}
		fmt.Println()
	}

	// Create database if needed
	if !dbExists {
		printInfo(fmt.Sprintf("Creating database '%s'...", targetDatabase))
		if err := client.CreateDatabase(targetDatabase); err != nil {
			printError(fmt.Sprintf("Failed to create database '%s'", targetDatabase))
			return err
		}
		printSuccess(fmt.Sprintf("Database '%s' created", targetDatabase))
	}

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Decompress if needed
	decompressor := backup.NewDecompressor(compression)
	sqlReader, err := decompressor.DecompressToReader(file)
	if err != nil {
		return fmt.Errorf("failed to decompress file: %w", err)
	}
	defer sqlReader.Close()

	// Execute restore via MySQLRestorer
	printInfo("Starting import...")

	done := make(chan bool)
	go showImportSpinner(done)

	startTime := time.Now()

	// Use a separate config for the restorer without the short connection timeout,
	// so the import gets the default 30-minute timeout instead of 60 seconds.
	restorerConfig := &mysql.Config{
		Host:     dbConfig.Host,
		Port:     dbConfig.Port,
		User:     dbConfig.User,
		Password: password,
		Database: "",
	}
	restorer := backup.NewMySQLRestorer(restorerConfig)

	var cmdLogger func(string)
	if c.Bool("verbose") {
		cmdLogger = func(cmd string) {
			fmt.Printf("\r%sCommand:%s %s\n", colorCyan, colorReset, cmd)
		}
	}

	err = restorer.RestoreWithCommand(targetDatabase, sqlReader, cmdLogger)
	done <- true

	if err != nil {
		printError("Import failed")
		return err
	}

	duration := time.Since(startTime)

	printSuccess("Import completed!")
	fmt.Println()
	fmt.Printf("  %sFile:%s        %s\n", colorCyan, colorReset, filePath)
	fmt.Printf("  %sDatabase:%s    %s\n", colorCyan, colorReset, targetDatabase)
	fmt.Printf("  %sDuration:%s    %s\n", colorCyan, colorReset, backup.FormatDuration(duration))
	fmt.Println()
	fmt.Printf("SQL dump has been imported into '%s' successfully.\n", targetDatabase)

	return nil
}

// showImportSpinner displays a spinner during import
func showImportSpinner(done chan bool) {
	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	i := 0
	for {
		select {
		case <-done:
			fmt.Print("\r")
			return
		default:
			fmt.Printf("\r%s Importing... ", spinner[i%len(spinner)])
			i++
			time.Sleep(100 * time.Millisecond)
		}
	}
}
