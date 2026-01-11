package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/erickhilda/cadangkan/internal/backup"
	"github.com/erickhilda/cadangkan/internal/config"
	"github.com/erickhilda/cadangkan/internal/storage"
	"github.com/urfave/cli/v2"
)

// databaseBackups holds backups for a single database
type databaseBackups struct {
	database string
	backups  []backup.BackupListEntry
}

func backupListCommand() *cli.Command {
	return &cli.Command{
		Name:    "backup-list",
		Aliases: []string{"backups"},
		Usage:   "List all backups for configured databases",
		Description: `List backups for one or all configured databases.

   USAGE:
     cadangkan backup-list                    # List backups for all databases
     cadangkan backup-list <database-name>    # List backups for specific database
     cadangkan backup-list --format=json      # Output in JSON format`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "format",
				Value: "table",
				Usage: "Output format: table (default) or json",
			},
		},
		Action: runBackupList,
	}
}

func runBackupList(c *cli.Context) error {
	format := c.String("format")
	if format != "table" && format != "json" {
		return fmt.Errorf("invalid format: %s (must be 'table' or 'json')", format)
	}

	// Create storage and backup service
	storageInstance, err := storage.NewLocalStorage("")
	if err != nil {
		return fmt.Errorf("failed to create storage: %w", err)
	}

	// Create config manager
	mgr, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create config manager: %w", err)
	}

	// Get database name from args (optional)
	var targetDatabase string
	if c.NArg() > 0 {
		targetDatabase = c.Args().Get(0)
	}

	// Load config to get all database names
	cfg, err := mgr.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// If specific database requested, verify it exists
	if targetDatabase != "" {
		if _, exists := cfg.Databases[targetDatabase]; !exists {
			return fmt.Errorf("database '%s' not found in configuration", targetDatabase)
		}
	}

	// Collect all backups
	var allBackups []databaseBackups

	if targetDatabase != "" {
		// List backups for specific database
		backups, err := storageInstance.ListBackups(targetDatabase)
		if err != nil {
			return fmt.Errorf("failed to list backups for '%s': %w", targetDatabase, err)
		}

		// Convert storage.BackupListEntry to backup.BackupListEntry
		backupEntries := make([]backup.BackupListEntry, len(backups))
		for i, entry := range backups {
			backupEntries[i] = backup.BackupListEntry{
				BackupID:     entry.BackupID,
				Database:     entry.Database,
				CreatedAt:    entry.CreatedAt,
				SizeBytes:    entry.SizeBytes,
				SizeHuman:    entry.SizeHuman,
				Status:       entry.Status,
				FilePath:     entry.FilePath,
				MetadataPath: entry.MetadataPath,
			}
		}

		if len(backupEntries) > 0 {
			allBackups = append(allBackups, databaseBackups{
				database: targetDatabase,
				backups:  backupEntries,
			})
		}
	} else {
		// List backups for all databases
		dbNames := make([]string, 0, len(cfg.Databases))
		for name := range cfg.Databases {
			dbNames = append(dbNames, name)
		}
		sort.Strings(dbNames)

		for _, dbName := range dbNames {
			backups, err := storageInstance.ListBackups(dbName)
			if err != nil {
				// Log error but continue with other databases
				fmt.Fprintf(os.Stderr, "Warning: failed to list backups for '%s': %v\n", dbName, err)
				continue
			}

			// Convert storage.BackupListEntry to backup.BackupListEntry
			backupEntries := make([]backup.BackupListEntry, len(backups))
			for i, entry := range backups {
				backupEntries[i] = backup.BackupListEntry{
					BackupID:     entry.BackupID,
					Database:     entry.Database,
					CreatedAt:    entry.CreatedAt,
					SizeBytes:    entry.SizeBytes,
					SizeHuman:    entry.SizeHuman,
					Status:       entry.Status,
					FilePath:     entry.FilePath,
					MetadataPath: entry.MetadataPath,
				}
			}

			if len(backupEntries) > 0 {
				allBackups = append(allBackups, databaseBackups{
					database: dbName,
					backups:  backupEntries,
				})
			}
		}
	}

	// Output results
	if format == "json" {
		return outputBackupsJSON(allBackups)
	}

	return outputBackupsTable(allBackups, targetDatabase)
}

func outputBackupsTable(allBackups []databaseBackups, targetDatabase string) error {
	if len(allBackups) == 0 {
		if targetDatabase != "" {
			printInfo(fmt.Sprintf("No backups found for database '%s'", targetDatabase))
		} else {
			printInfo("No backups found for any configured database")
		}
		fmt.Println()
		fmt.Printf("Create a backup with: %scadangkan backup <name>%s\n", colorCyan, colorReset)
		return nil
	}

	totalBackups := 0
	for _, dbBackups := range allBackups {
		totalBackups += len(dbBackups.backups)
	}

	// If listing all databases, show each database's backups separately
	if targetDatabase == "" && len(allBackups) > 1 {
		for i, dbBackups := range allBackups {
			if i > 0 {
				fmt.Println()
			}
			printBackupsForDatabase(dbBackups.database, dbBackups.backups)
		}
		fmt.Println()
		fmt.Printf("Total: %d backup(s) across %d database(s)\n", totalBackups, len(allBackups))
	} else {
		// Single database view
		dbBackups := allBackups[0]
		printBackupsForDatabase(dbBackups.database, dbBackups.backups)
	}

	return nil
}

func printBackupsForDatabase(database string, backups []backup.BackupListEntry) {
	fmt.Printf("\n%sBackups for %s%s\n", colorCyan, colorReset, database)
	fmt.Println(strings.Repeat("=", 100))
	fmt.Printf("%-20s %-20s %-12s %-12s\n", "BACKUP ID", "DATE", "SIZE", "STATUS")
	fmt.Println(strings.Repeat("-", 100))

	for _, b := range backups {
		dateStr := b.CreatedAt.Format("2006-01-02 15:04:05")
		sizeStr := b.SizeHuman
		if sizeStr == "" {
			sizeStr = backup.FormatBytes(b.SizeBytes)
		}

		statusStr := b.Status
		if statusStr == "" {
			statusStr = "completed"
		}

		fmt.Printf("%-20s %-20s %-12s %-12s\n", b.BackupID, dateStr, sizeStr, statusStr)
	}

	fmt.Println()
	fmt.Printf("Total: %d backup(s)\n", len(backups))
}

func outputBackupsJSON(allBackups []databaseBackups) error {
	// Simple JSON output (could be enhanced with proper JSON marshaling)
	fmt.Println("{")
	fmt.Println(`  "backups": [`)

	for i, dbBackups := range allBackups {
		for j, b := range dbBackups.backups {
			if i > 0 || j > 0 {
				fmt.Println(",")
			}
			dateStr := b.CreatedAt.Format(time.RFC3339)
			sizeStr := b.SizeHuman
			if sizeStr == "" {
				sizeStr = backup.FormatBytes(b.SizeBytes)
			}

			fmt.Printf(`    {
      "backup_id": "%s",
      "database": "%s",
      "created_at": "%s",
      "size_bytes": %d,
      "size_human": "%s",
      "status": "%s",
      "file_path": "%s"
    }`, b.BackupID, b.Database, dateStr, b.SizeBytes, sizeStr, b.Status, b.FilePath)
		}
	}

	fmt.Println()
	fmt.Println("  ]")
	fmt.Println("}")
	return nil
}
