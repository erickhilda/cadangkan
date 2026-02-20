package main

import (
	"fmt"

	"github.com/erickhilda/cadangkan/internal/backup"
	"github.com/erickhilda/cadangkan/internal/config"
	"github.com/erickhilda/cadangkan/internal/storage"
	"github.com/urfave/cli/v2"
)

func cleanupCommand() *cli.Command {
	return &cli.Command{
		Name:      "cleanup",
		Usage:     "Clean up old backups based on retention policy",
		ArgsUsage: "<name>",
		Description: `Apply retention policy to clean up old backups.

   Backups are categorized as:
     - Daily:   Most recent backup each day for last N days
     - Weekly:  Most recent backup each week (Sunday) for last N weeks
     - Monthly: Most recent backup each month (1st) for last N months

   By default, uses retention policy from config:
     daily: 7, weekly: 4, monthly: 12

   Use --dry-run to preview what would be deleted without actually deleting.`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Show what would be deleted without actually deleting",
			},
			&cli.IntFlag{
				Name:  "daily",
				Usage: "Override daily retention (keep last N daily backups)",
			},
			&cli.IntFlag{
				Name:  "weekly",
				Usage: "Override weekly retention (keep last N weekly backups)",
			},
			&cli.IntFlag{
				Name:  "monthly",
				Usage: "Override monthly retention (keep last N monthly backups)",
			},
		},
		Action: runCleanup,
	}
}

func runCleanup(c *cli.Context) error {
	// Require database name
	if c.NArg() == 0 {
		return fmt.Errorf("database name is required\n\nUsage: cadangkan cleanup <name>")
	}

	name := c.Args().Get(0)

	// Load configuration
	mgr, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create config manager: %w", err)
	}

	// Get config to check if database exists
	cfg, err := mgr.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if database exists in config
	if _, exists := cfg.Databases[name]; !exists {
		printError(fmt.Sprintf("Database '%s' not found in config", name))
		fmt.Println()
		fmt.Printf("Available databases: run %scadangkan list%s\n", colorCyan, colorReset)
		return fmt.Errorf("database not found")
	}

	// Get retention policy (from config or overrides)
	policy := cfg.GetEffectiveRetention(name)

	// Apply command-line overrides
	if c.IsSet("daily") {
		policy.Daily = c.Int("daily")
	}
	if c.IsSet("weekly") {
		policy.Weekly = c.Int("weekly")
	}
	if c.IsSet("monthly") {
		policy.Monthly = c.Int("monthly")
	}

	dryRun := c.Bool("dry-run")

	// Create storage
	localStorage, err := storage.NewLocalStorage("")
	if err != nil {
		printError("Failed to create storage")
		return err
	}

	// Create retention service
	retentionService := backup.NewRetentionService(localStorage)

	// Show retention policy
	fmt.Println()
	if dryRun {
		printInfo(fmt.Sprintf("Cleanup preview for '%s' (dry-run mode)", name))
	} else {
		printInfo(fmt.Sprintf("Cleaning up backups for '%s'", name))
	}
	fmt.Println()

	fmt.Printf("Retention policy:\n")
	fmt.Printf("  %sDaily:%s    Keep last %d days\n", colorCyan, colorReset, policy.Daily)
	fmt.Printf("  %sWeekly:%s   Keep last %d weeks\n", colorCyan, colorReset, policy.Weekly)
	fmt.Printf("  %sMonthly:%s  Keep last %d months\n", colorCyan, colorReset, policy.Monthly)
	fmt.Println()

	// Apply retention policy
	result, err := retentionService.ApplyRetentionPolicy(name, policy, dryRun)
	if err != nil {
		printError("Cleanup failed")
		return err
	}

	// Display results
	if len(result.ToDelete) == 0 {
		printSuccess("No backups to delete")
		fmt.Println()
		fmt.Printf("All %d backup(s) match the retention policy.\n", len(result.ToKeep))
		return nil
	}

	// Show backups to delete
	fmt.Printf("Backups to delete: %s%d%s\n", colorYellow, len(result.ToDelete), colorReset)
	fmt.Println()
	for _, backup := range result.ToDelete {
		age := formatAge(backup.CreatedAt)
		fmt.Printf("  %s%-20s%s  %s (%s old)  %s\n",
			colorRed,
			backup.BackupID,
			colorReset,
			backup.SizeHuman,
			age,
			backup.CreatedAt.Format("2006-01-02 15:04:05"),
		)
	}
	fmt.Println()

	// Show backups to keep
	if len(result.ToKeep) > 0 {
		fmt.Printf("Backups to keep: %s%d%s\n", colorGreen, len(result.ToKeep), colorReset)
		fmt.Println()

		// Group by category
		dailyBackups := []backup.CategorizedBackup{}
		weeklyBackups := []backup.CategorizedBackup{}
		monthlyBackups := []backup.CategorizedBackup{}
		keepBackups := []backup.CategorizedBackup{}

		for _, cb := range result.ToKeep {
			switch cb.Category {
			case backup.CategoryDaily:
				dailyBackups = append(dailyBackups, cb)
			case backup.CategoryWeekly:
				weeklyBackups = append(weeklyBackups, cb)
			case backup.CategoryMonthly:
				monthlyBackups = append(monthlyBackups, cb)
			case backup.CategoryKeep:
				keepBackups = append(keepBackups, cb)
			}
		}

		if len(dailyBackups) > 0 {
			fmt.Printf("  %sDaily backups:%s %d\n", colorCyan, colorReset, len(dailyBackups))
		}
		if len(weeklyBackups) > 0 {
			fmt.Printf("  %sWeekly backups:%s %d\n", colorCyan, colorReset, len(weeklyBackups))
		}
		if len(monthlyBackups) > 0 {
			fmt.Printf("  %sMonthly backups:%s %d\n", colorCyan, colorReset, len(monthlyBackups))
		}
		if len(keepBackups) > 0 {
			fmt.Printf("  %sAlways keep:%s %d\n", colorCyan, colorReset, len(keepBackups))
		}
		fmt.Println()
	}

	// Show space reclaimed
	spaceHuman := formatBytes(result.SpaceReclaimed)
	if dryRun {
		fmt.Printf("Space that would be reclaimed: %s%s%s\n", colorYellow, spaceHuman, colorReset)
		fmt.Println()
		printInfo("Run without --dry-run to delete these backups.")
	} else {
		printSuccess(fmt.Sprintf("Deleted %d backup(s)", len(result.ToDelete)))
		fmt.Printf("Space reclaimed: %s%s%s\n", colorGreen, spaceHuman, colorReset)
	}

	return nil
}
