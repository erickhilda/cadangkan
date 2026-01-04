package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/erickhilda/cadangkan/internal/config"
	"github.com/urfave/cli/v2"
)

func removeCommand() *cli.Command {
	return &cli.Command{
		Name:      "remove",
		Aliases:   []string{"rm"},
		Usage:     "Remove a database configuration",
		ArgsUsage: "<name>",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Skip confirmation prompt",
			},
		},
		Action: runRemove,
	}
}

func runRemove(c *cli.Context) error {
	// Parse arguments
	if c.NArg() < 1 {
		return fmt.Errorf("usage: cadangkan remove <name>")
	}

	name := c.Args().Get(0)
	force := c.Bool("force")

	// Create config manager
	mgr, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create config manager: %w", err)
	}

	// Check if database exists
	dbConfig, err := mgr.GetDatabase(name)
	if err != nil {
		printError("Database not found")
		return err
	}

	// Confirm deletion (unless --force)
	if !force {
		fmt.Printf("\n%sWarning:%s You are about to remove the database configuration:\n\n", colorYellow, colorReset)
		fmt.Printf("  Name:     %s\n", name)
		fmt.Printf("  Type:     %s\n", dbConfig.Type)
		fmt.Printf("  Host:     %s:%d\n", dbConfig.Host, dbConfig.Port)
		fmt.Printf("  Database: %s\n\n", dbConfig.Database)
		fmt.Printf("%sNote:%s This will only remove the configuration, not the actual database or backups.\n\n", colorYellow, colorReset)

		fmt.Print("Are you sure? (yes/no): ")
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "yes" && response != "y" {
			printInfo("Cancelled")
			return nil
		}
	}

	// Remove database
	printInfo("Removing configuration...")
	if err := mgr.RemoveDatabase(name); err != nil {
		printError("Failed to remove configuration")
		return err
	}

	printSuccess(fmt.Sprintf("Database '%s' removed successfully!", name))

	return nil
}
