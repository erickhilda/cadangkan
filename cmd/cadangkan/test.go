package main

import (
	"fmt"
	"time"

	"github.com/erickhilda/cadangkan/internal/config"
	"github.com/erickhilda/cadangkan/pkg/database/mysql"
	"github.com/urfave/cli/v2"
)

func testCommand() *cli.Command {
	return &cli.Command{
		Name:      "test",
		Usage:     "Test database connection",
		ArgsUsage: "<name>",
		Action:    runTest,
	}
}

func runTest(c *cli.Context) error {
	// Parse arguments
	if c.NArg() < 1 {
		return fmt.Errorf("usage: cadangkan test <name>")
	}

	name := c.Args().Get(0)

	// Create config manager
	mgr, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create config manager: %w", err)
	}

	// Load database config
	printInfo(fmt.Sprintf("Loading configuration for '%s'...", name))
	dbConfig, err := mgr.GetDatabase(name)
	if err != nil {
		printError("Database not found")
		return err
	}

	// Decrypt password
	password, err := config.DecryptPassword(dbConfig.PasswordEncrypted)
	if err != nil {
		printError("Failed to decrypt password")
		return err
	}

	// Test connection
	printInfo(fmt.Sprintf("Testing connection to %s@%s:%d...", dbConfig.User, dbConfig.Host, dbConfig.Port))

	mysqlConfig := &mysql.Config{
		Host:     dbConfig.Host,
		Port:     dbConfig.Port,
		User:     dbConfig.User,
		Password: password,
		Database: dbConfig.Database,
		Timeout:  10 * time.Second,
	}

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

	// Get database version and info
	dbVersion, err := client.GetVersion()
	if err != nil {
		dbVersion = "unknown"
	}

	printSuccess(fmt.Sprintf("Connected successfully (MySQL %s)", dbVersion))

	// Get database size
	size, err := client.GetDatabaseSize(dbConfig.Database)
	if err == nil {
		fmt.Printf("\n  %sDatabase:%s %s\n", colorCyan, colorReset, dbConfig.Database)
		fmt.Printf("  %sSize:%s     %s\n", colorCyan, colorReset, formatBytes(size))
	}

	return nil
}

// formatBytes formats bytes into human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
