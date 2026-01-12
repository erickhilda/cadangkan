package main

import (
	"fmt"
	"strings"

	"github.com/erickhilda/cadangkan/internal/backup"
	"github.com/erickhilda/cadangkan/internal/config"
	"github.com/erickhilda/cadangkan/internal/status"
	"github.com/erickhilda/cadangkan/internal/storage"
	"github.com/urfave/cli/v2"
)

func storageCommand() *cli.Command {
	return &cli.Command{
		Name:  "storage",
		Usage: "Show storage usage breakdown",
		Description: `Display storage usage across all databases.

   Shows total storage used, available disk space, breakdown by database,
   and largest backups.

   USAGE:
     cadangkan storage   # Show storage usage breakdown`,
		Action: runStorage,
	}
}

func runStorage(c *cli.Context) error {
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

	// Get storage usage
	usage, err := statusService.GetStorageUsage()
	if err != nil {
		return fmt.Errorf("failed to get storage usage: %w", err)
	}

	return showStorageUsage(usage)
}

func showStorageUsage(usage *status.StorageUsage) error {
	fmt.Printf("\n%sStorage Usage%s\n", colorCyan, colorReset)
	fmt.Println(strings.Repeat("=", 80))

	// Total storage
	fmt.Println("Total Storage:")
	if usage.TotalAvailable > 0 {
		fmt.Printf("  Used:      %s\n", formatStorageUsage(usage.TotalUsed, usage.TotalAvailable))
		fmt.Printf("  Available: %s\n", backup.FormatBytes(int64(usage.TotalAvailable)))
	} else {
		fmt.Printf("  Used:      %s\n", backup.FormatBytes(usage.TotalUsed))
		fmt.Printf("  Available: %sUnknown%s\n", colorYellow, colorReset)
	}
	fmt.Println()

	// Storage by database
	if len(usage.ByDatabase) > 0 {
		fmt.Println("Storage by Database:")
		fmt.Printf("%-20s %-12s %-15s %-10s\n", "DATABASE", "BACKUPS", "SIZE", "PERCENTAGE")
		fmt.Println(strings.Repeat("-", 80))

		for _, dbStorage := range usage.ByDatabase {
			fmt.Printf("%-20s %-12d %-15s %-10.1f%%\n",
				dbStorage.Database,
				dbStorage.BackupCount,
				backup.FormatBytes(dbStorage.SizeBytes),
				dbStorage.Percentage,
			)
		}
		fmt.Println()
	} else {
		fmt.Println("Storage by Database:")
		fmt.Printf("  %sNo backups found%s\n", colorYellow, colorReset)
		fmt.Println()
	}

	// Largest backups
	if len(usage.LargestBackups) > 0 {
		fmt.Println("Largest Backups:")
		fmt.Printf("%-20s %-20s %-15s %-12s\n", "BACKUP ID", "DATABASE", "DATE", "SIZE")
		fmt.Println(strings.Repeat("-", 80))

		for _, b := range usage.LargestBackups {
			dateStr := b.CreatedAt.Format("2006-01-02 15:04:05")
			sizeStr := b.SizeHuman
			if sizeStr == "" {
				sizeStr = backup.FormatBytes(b.SizeBytes)
			}

			fmt.Printf("%-20s %-20s %-15s %-12s\n",
				b.BackupID,
				b.Database,
				dateStr,
				sizeStr,
			)
		}
		fmt.Println()
	}

	return nil
}
