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
			// Backup operations
			backupCommand(),
			// Future: restoreCommand(), scheduleCommand()
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
