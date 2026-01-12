package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

const (
	AppName    = "cadangkan"
	AppVersion = "0.1.0"
	AppUsage   = "Database backup and restore tool"
)

func main() {
	app := &cli.App{
		Name:    AppName,
		Version: AppVersion,
		Usage:   AppUsage,
		Commands: []*cli.Command{
			// Database management
			addCommand(),
			listCommand(),
			testCommand(),
			removeCommand(),
			editCommand(),
			// Backup operations
			backupCommand(),
			backupListCommand(),
			restoreCommand(),
			// Status & monitoring
			statusCommand(),
			healthCommand(),
			storageCommand(),
			// Future: scheduleCommand()
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
