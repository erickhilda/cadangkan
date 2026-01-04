package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/erickhilda/cadangkan/internal/config"
	"github.com/urfave/cli/v2"
)

func listCommand() *cli.Command {
	return &cli.Command{
		Name:    "list",
		Aliases: []string{"ls"},
		Usage:   "List all configured databases",
		Action:  runList,
	}
}

func runList(c *cli.Context) error {
	// Create config manager
	mgr, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create config manager: %w", err)
	}

	// Load config
	cfg, err := mgr.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if there are any databases
	if len(cfg.Databases) == 0 {
		printInfo("No databases configured")
		fmt.Println()
		fmt.Printf("Add a database with: %scadangkan add mysql <name>%s\n", colorCyan, colorReset)
		return nil
	}

	// Get database names and sort them
	names := make([]string, 0, len(cfg.Databases))
	for name := range cfg.Databases {
		names = append(names, name)
	}
	sort.Strings(names)

	// Print header
	fmt.Printf("\n%sConfigured Databases%s\n", colorCyan, colorReset)
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("%-20s %-10s %-30s %s\n", "NAME", "TYPE", "HOST", "DATABASE")
	fmt.Println(strings.Repeat("-", 80))

	// Print each database
	for _, name := range names {
		db := cfg.Databases[name]
		hostPort := fmt.Sprintf("%s:%d", db.Host, db.Port)
		fmt.Printf("%-20s %-10s %-30s %s\n", name, db.Type, hostPort, db.Database)
	}

	fmt.Println()
	fmt.Printf("Total: %d database(s)\n", len(cfg.Databases))
	fmt.Println()
	fmt.Printf("Backup a database: %scadangkan backup <name>%s\n", colorCyan, colorReset)
	fmt.Printf("Test connection:   %scadangkan test <name>%s\n", colorCyan, colorReset)
	fmt.Printf("Remove database:   %scadangkan remove <name>%s\n", colorCyan, colorReset)

	return nil
}
