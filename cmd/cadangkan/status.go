package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/erickhilda/cadangkan/internal/backup"
	"github.com/erickhilda/cadangkan/internal/config"
	"github.com/erickhilda/cadangkan/internal/status"
	"github.com/erickhilda/cadangkan/internal/storage"
	"github.com/urfave/cli/v2"
)

func statusCommand() *cli.Command {
	return &cli.Command{
		Name:    "status",
		Aliases: []string{"st"},
		Usage:   "Show backup status for all databases or a specific database",
		Description: `Show overall backup status or detailed status for a specific database.

   USAGE:
     cadangkan status              # Show overall status for all databases
     cadangkan status <database>   # Show detailed status for a specific database`,
		Action: runStatus,
	}
}

func runStatus(c *cli.Context) error {
	// Create storage and config manager
	storageInstance, err := storage.NewLocalStorage("")
	if err != nil {
		return fmt.Errorf("failed to create storage: %w", err)
	}

	configManager, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create config manager: %w", err)
	}

	// Create status service
	statusService := status.NewService(configManager, storageInstance)

	// Check if specific database requested
	if c.NArg() > 0 {
		dbName := c.Args().Get(0)
		return showDatabaseStatus(statusService, dbName)
	}

	return showOverallStatus(statusService)
}

func showOverallStatus(svc *status.Service) error {
	overall, err := svc.GetOverallStatus()
	if err != nil {
		return fmt.Errorf("failed to get overall status: %w", err)
	}

	fmt.Printf("\n%sCadangkan Status%s\n", colorCyan, colorReset)
	fmt.Println(strings.Repeat("=", 80))

	// Service status
	serviceStatus := overall.ServiceStatus
	if serviceStatus == "Not running" {
		fmt.Printf("Service: %s%s%s\n", colorYellow, serviceStatus, colorReset)
	} else {
		fmt.Printf("Service: %s%s%s\n", colorGreen, serviceStatus, colorReset)
	}

	// Database count
	fmt.Printf("Databases: %d", overall.DatabaseCount)
	if overall.ActiveCount > 0 {
		fmt.Printf(" (%d active)", overall.ActiveCount)
	}
	fmt.Println()

	// Total backups
	fmt.Printf("Total Backups: %d\n", overall.TotalBackups)

	// Storage usage
	if overall.StorageAvailable > 0 {
		fmt.Printf("Storage Used: %s\n", formatStorageUsage(overall.StorageUsed, overall.StorageAvailable))
	} else {
		fmt.Printf("Storage Used: %s\n", backup.FormatBytes(overall.StorageUsed))
	}

	// Last backup
	if overall.LastBackup != nil {
		fmt.Printf("Last Backup: %s %s\n", formatTimeAgo(*overall.LastBackup), getStatusIndicator("healthy"))
	} else {
		fmt.Printf("Last Backup: %sNever%s\n", colorYellow, colorReset)
	}

	fmt.Println()

	// Database table
	if len(overall.Databases) > 0 {
		fmt.Printf("%-20s %-10s %-8s %-20s %-15s\n", "DATABASE", "TYPE", "STATUS", "LAST BACKUP", "NEXT BACKUP")
		fmt.Println(strings.Repeat("-", 80))

		for _, db := range overall.Databases {
			statusInd := getStatusIndicator(db.Status)
			lastBackupStr := "Never"
			if db.LastBackup != nil {
				lastBackupStr = formatTimeAgo(*db.LastBackup)
			}
			nextBackupStr := db.NextBackup

			fmt.Printf("%-20s %-10s %-8s %-20s %-15s\n",
				db.Name,
				db.Type,
				statusInd,
				lastBackupStr,
				nextBackupStr,
			)
		}
		fmt.Println()
	}

	// Health summary
	if len(overall.HealthSummary) > 0 {
		fmt.Println("Health Summary:")
		for _, summary := range overall.HealthSummary {
			fmt.Printf("  %s\n", summary)
		}
		fmt.Println()
	}

	// Helpful commands
	fmt.Println("Commands:")
	fmt.Printf("  %scadangkan status <database>%s  # Detailed status\n", colorCyan, colorReset)
	fmt.Printf("  %scadangkan health <database>%s  # Health score\n", colorCyan, colorReset)
	fmt.Printf("  %scadangkan storage%s            # Storage breakdown\n", colorCyan, colorReset)

	return nil
}

func showDatabaseStatus(svc *status.Service, dbName string) error {
	dbStatus, err := svc.GetDatabaseStatus(dbName)
	if err != nil {
		return fmt.Errorf("failed to get database status: %w", err)
	}

	// Load config to get connection details
	configManager, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create config manager: %w", err)
	}

	cfg, err := configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	dbConfig, exists := cfg.Databases[dbName]
	if !exists {
		return fmt.Errorf("database '%s' not found", dbName)
	}

	fmt.Printf("\n%sStatus for %s%s\n", colorCyan, colorReset, dbName)
	fmt.Println(strings.Repeat("=", 80))

	// Database info
	fmt.Printf("Type:     %s\n", dbStatus.Type)
	fmt.Printf("Host:     %s:%d\n", dbConfig.Host, dbConfig.Port)
	fmt.Printf("Database: %s\n", dbConfig.Database)
	fmt.Printf("User:     %s\n", dbConfig.User)
	fmt.Println()

	// Status indicator
	statusInd := getStatusIndicator(dbStatus.Status)
	fmt.Printf("Status: %s %s\n", statusInd, dbStatus.Status)
	fmt.Println()

	// Backup statistics
	fmt.Println("Backup Statistics:")
	fmt.Printf("  Total Backups:     %d\n", dbStatus.BackupCount)
	fmt.Printf("  Successful:        %d\n", dbStatus.SuccessfulCount)
	if dbStatus.FailedCount > 0 {
		fmt.Printf("  Failed:            %s%d%s\n", colorRed, dbStatus.FailedCount, colorReset)
	} else {
		fmt.Printf("  Failed:            %d\n", dbStatus.FailedCount)
	}
	fmt.Printf("  Storage Used:      %s\n", backup.FormatBytes(dbStatus.StorageUsed))
	fmt.Println()

	// Last backup details
	if dbStatus.LastBackup != nil {
		fmt.Println("Last Backup:")
		fmt.Printf("  ID:       %s\n", dbStatus.LastBackupID)
		fmt.Printf("  Time:     %s (%s)\n", dbStatus.LastBackup.Format(time.RFC3339), formatTimeAgo(*dbStatus.LastBackup))
		fmt.Println()
	} else {
		fmt.Println("Last Backup: Never")
		fmt.Println()
	}

	// Next scheduled backup
	fmt.Printf("Next Scheduled Backup: %s\n", dbStatus.NextBackup)
	fmt.Println()

	// Recent backups
	if len(dbStatus.RecentBackups) > 0 {
		fmt.Println("Recent Backups:")
		fmt.Printf("%-20s %-20s %-12s %-12s\n", "BACKUP ID", "DATE", "SIZE", "STATUS")
		fmt.Println(strings.Repeat("-", 80))

		for _, b := range dbStatus.RecentBackups {
			dateStr := b.CreatedAt.Format("2006-01-02 15:04:05")
			sizeStr := b.SizeHuman
			if sizeStr == "" {
				sizeStr = backup.FormatBytes(b.SizeBytes)
			}

			statusStr := b.Status
			if statusStr == "" {
				statusStr = "completed"
			}

			statusColor := colorGreen
			if statusStr == backup.StatusFailed {
				statusColor = colorRed
			}

			fmt.Printf("%-20s %-20s %-12s %s%-12s%s\n",
				b.BackupID,
				dateStr,
				sizeStr,
				statusColor,
				statusStr,
				colorReset,
			)
		}
		fmt.Println()
	}

	return nil
}
