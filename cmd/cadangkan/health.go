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

func healthCommand() *cli.Command {
	return &cli.Command{
		Name:  "health",
		Usage: "Show health score for a database",
		Description: `Calculate and display health score for a specific database.

   The health score is calculated based on:
   - Success Rate (50%): Percentage of successful backups
   - Recency (30%): How recent the last backup is
   - Consistency (20%): Regularity of backup intervals

   USAGE:
     cadangkan health <database>   # Show health score for a database`,
		Action: runHealth,
	}
}

func runHealth(c *cli.Context) error {
	if c.NArg() == 0 {
		return fmt.Errorf("database name is required")
	}

	dbName := c.Args().Get(0)

	// Create storage and config manager
	storageInstance, err := storage.NewLocalStorage("")
	if err != nil {
		return fmt.Errorf("failed to create storage: %w", err)
	}

	configManager, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create config manager: %w", err)
	}

	// Verify database exists
	cfg, err := configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if _, exists := cfg.Databases[dbName]; !exists {
		return fmt.Errorf("database '%s' not found", dbName)
	}

	// Get all backups for health calculation
	storageBackups, err := storageInstance.ListBackups(dbName)
	if err != nil {
		return fmt.Errorf("failed to list backups: %w", err)
	}

	// Convert storage.BackupListEntry to backup.BackupListEntry
	backups := make([]backup.BackupListEntry, len(storageBackups))
	for i, b := range storageBackups {
		backups[i] = backup.BackupListEntry{
			BackupID:     b.BackupID,
			Database:     b.Database,
			CreatedAt:    b.CreatedAt,
			SizeBytes:    b.SizeBytes,
			SizeHuman:    b.SizeHuman,
			Status:       b.Status,
			FilePath:     b.FilePath,
			MetadataPath: b.MetadataPath,
		}
	}

	// Calculate health score
	healthScore := status.CalculateHealthScore(backups)

	// Display health score
	return showHealthScore(dbName, healthScore)
}

func showHealthScore(dbName string, score status.HealthScore) error {
	fmt.Printf("\n%sHealth Score for %s%s\n", colorCyan, colorReset, dbName)
	fmt.Println(strings.Repeat("=", 80))

	// Overall score with color coding
	scoreColor := colorRed
	scoreLabel := "Critical"
	if score.TotalScore >= 80.0 {
		scoreColor = colorGreen
		scoreLabel = "Healthy"
	} else if score.TotalScore >= 50.0 {
		scoreColor = colorYellow
		scoreLabel = "Warning"
	}

	fmt.Printf("Overall Score: %s%.1f / 100.0%s (%s)\n", scoreColor, score.TotalScore, colorReset, scoreLabel)
	fmt.Println()

	// Score breakdown
	fmt.Println("Score Breakdown:")
	fmt.Printf("  Success Rate:      %.1f / 50.0  (%.1f%%)\n",
		score.SuccessRate,
		(score.SuccessRate/50.0)*100.0,
	)
	fmt.Printf("  Recency:           %.1f / 30.0  (%.1f%%)\n",
		score.RecencyScore,
		(score.RecencyScore/30.0)*100.0,
	)
	fmt.Printf("  Consistency:       %.1f / 20.0  (%.1f%%)\n",
		score.ConsistencyScore,
		(score.ConsistencyScore/20.0)*100.0,
	)
	fmt.Println()

	// Recommendations
	if len(score.Recommendations) > 0 {
		fmt.Println("Recommendations:")
		for _, rec := range score.Recommendations {
			fmt.Printf("  %s⚠%s %s\n", colorYellow, colorReset, rec)
		}
		fmt.Println()
	} else {
		fmt.Println("Recommendations:")
		fmt.Printf("  %s✓%s No issues detected. Keep up the good work!\n", colorGreen, colorReset)
		fmt.Println()
	}

	// Recent backup history
	if len(score.RecentBackups) > 0 {
		fmt.Println("Recent Backup History (last 10):")
		fmt.Printf("%-20s %-20s %-12s %-12s\n", "BACKUP ID", "DATE", "SIZE", "STATUS")
		fmt.Println(strings.Repeat("-", 80))

		maxRecent := 10
		if len(score.RecentBackups) < maxRecent {
			maxRecent = len(score.RecentBackups)
		}

		for i := 0; i < maxRecent; i++ {
			b := score.RecentBackups[i]
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
	} else {
		fmt.Println("Recent Backup History:")
		fmt.Printf("  %sNo backups found%s\n", colorYellow, colorReset)
		fmt.Println()
	}

	return nil
}
