package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/erickhilda/cadangkan/internal/config"
	"github.com/erickhilda/cadangkan/internal/scheduler"
	"github.com/erickhilda/cadangkan/internal/storage"
	"github.com/urfave/cli/v2"
)

func daemonCommand() *cli.Command {
	return &cli.Command{
		Name:  "daemon",
		Usage: "Run Cadangkan as a background daemon",
		Description: `Start the Cadangkan daemon to run scheduled backups.

   The daemon will:
     - Load all configured schedules
     - Run backups at the scheduled times
     - Apply retention policies after backups
     - Continue running until stopped (Ctrl+C)

   USAGE:
     cadangkan daemon              Run in foreground
     cadangkan daemon --verbose    Run with verbose logging`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Enable verbose logging",
			},
		},
		Action: runDaemon,
	}
}

func runDaemon(c *cli.Context) error {
	verbose := c.Bool("verbose")

	// Load configuration
	mgr, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create config manager: %w", err)
	}

	cfg, err := mgr.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create storage
	localStorage, err := storage.NewLocalStorage("")
	if err != nil {
		return fmt.Errorf("failed to create storage: %w", err)
	}

	// Create scheduler
	sched := scheduler.New(cfg, localStorage)
	if verbose {
		sched.SetVerbose(true)
	}

	// Load schedules
	if err := sched.LoadSchedules(); err != nil {
		return fmt.Errorf("failed to load schedules: %w", err)
	}

	// Start scheduler
	sched.Start()

	printSuccess("Cadangkan daemon started")
	fmt.Println()

	// List active schedules
	schedules := sched.ListSchedules()
	if len(schedules) == 0 {
		printWarning("No schedules configured")
		fmt.Println()
		fmt.Println("Configure a schedule:")
		fmt.Printf("  %scadangkan schedule set <name> --daily --time=02:00%s\n", colorCyan, colorReset)
	} else {
		fmt.Printf("Active schedules: %s%d%s\n", colorGreen, len(schedules), colorReset)
		fmt.Println()
		for _, info := range schedules {
			fmt.Printf("  %s%-20s%s  Next: %s\n",
				colorCyan,
				info.Database,
				colorReset,
				formatNextRun(info.NextRun),
			)
		}
	}

	fmt.Println()
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println()
	printInfo("Shutting down daemon...")
	sched.Stop()
	printSuccess("Daemon stopped")

	return nil
}
