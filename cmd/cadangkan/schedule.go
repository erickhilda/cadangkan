package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/erickhilda/cadangkan/internal/config"
	"github.com/robfig/cron/v3"
	"github.com/urfave/cli/v2"
)

func scheduleCommand() *cli.Command {
	return &cli.Command{
		Name:  "schedule",
		Usage: "Manage backup schedules",
		Subcommands: []*cli.Command{
			scheduleSetCommand(),
			scheduleEnableCommand(),
			scheduleDisableCommand(),
			scheduleListCommand(),
			scheduleNextCommand(),
		},
	}
}

func scheduleSetCommand() *cli.Command {
	return &cli.Command{
		Name:      "set",
		Usage:     "Set backup schedule for a database",
		ArgsUsage: "<name>",
		Description: `Set a backup schedule for a database using cron syntax.

   EXAMPLES:
     Daily at 2 AM:
       cadangkan schedule set production --daily --time=02:00

     Weekly on Sunday at 3 AM:
       cadangkan schedule set production --weekly --day=sunday --time=03:00

     Custom cron expression (every 6 hours):
       cadangkan schedule set production --cron="0 */6 * * *"

   CRON FORMAT: minute hour day month weekday
     - minute: 0-59
     - hour: 0-23
     - day: 1-31
     - month: 1-12
     - weekday: 0-6 (0 = Sunday)`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "daily",
				Usage: "Schedule daily backup",
			},
			&cli.BoolFlag{
				Name:  "weekly",
				Usage: "Schedule weekly backup",
			},
			&cli.StringFlag{
				Name:  "time",
				Usage: "Time to run backup (HH:MM format, e.g., 02:00)",
			},
			&cli.StringFlag{
				Name:  "day",
				Usage: "Day of week for weekly backup (e.g., sunday, monday)",
			},
			&cli.StringFlag{
				Name:  "cron",
				Usage: "Custom cron expression (e.g., '0 2 * * *')",
			},
		},
		Action: runScheduleSet,
	}
}

func runScheduleSet(c *cli.Context) error {
	if c.NArg() == 0 {
		return fmt.Errorf("database name is required\n\nUsage: cadangkan schedule set <name> [flags]")
	}

	name := c.Args().Get(0)

	// Load configuration
	mgr, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create config manager: %w", err)
	}

	// Check if database exists
	dbConfig, err := mgr.GetDatabase(name)
	if err != nil {
		printError(fmt.Sprintf("Database '%s' not found", name))
		return err
	}

	// Determine cron expression
	var cronExpr string

	if c.IsSet("cron") {
		// Custom cron expression
		cronExpr = c.String("cron")
	} else if c.Bool("daily") {
		// Daily schedule
		timeStr := c.String("time")
		if timeStr == "" {
			timeStr = "02:00" // Default to 2 AM
		}
		cronExpr, err = parseDailyCron(timeStr)
		if err != nil {
			return err
		}
	} else if c.Bool("weekly") {
		// Weekly schedule
		timeStr := c.String("time")
		if timeStr == "" {
			timeStr = "03:00" // Default to 3 AM
		}
		dayStr := c.String("day")
		if dayStr == "" {
			dayStr = "sunday" // Default to Sunday
		}
		cronExpr, err = parseWeeklyCron(timeStr, dayStr)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("must specify --daily, --weekly, or --cron")
	}

	// Validate cron expression
	_, err = cron.ParseStandard(cronExpr)
	if err != nil {
		return fmt.Errorf("invalid cron expression '%s': %w", cronExpr, err)
	}

	// Update database config
	if dbConfig.Schedule == nil {
		dbConfig.Schedule = &config.ScheduleConfig{}
	}
	dbConfig.Schedule.Cron = cronExpr
	dbConfig.Schedule.Enabled = true

	// Save configuration
	if err := mgr.AddDatabase(name, dbConfig); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Calculate next run
	sched, _ := cron.ParseStandard(cronExpr)
	nextRun := sched.Next(time.Now())

	printSuccess(fmt.Sprintf("Schedule configured for '%s'", name))
	fmt.Println()
	fmt.Printf("  %sSchedule:%s  %s\n", colorCyan, colorReset, cronExpr)
	fmt.Printf("  %sNext run:%s  %s (%s)\n", colorCyan, colorReset, nextRun.Format("2006-01-02 15:04:05"), formatNextRun(nextRun))
	fmt.Printf("  %sStatus:%s    %sEnabled%s\n", colorCyan, colorReset, colorGreen, colorReset)
	fmt.Println()
	fmt.Println("The schedule will be active when the Cadangkan service is running.")
	fmt.Println()
	printInfo("To start the service:")
	fmt.Printf("  %scadangkan daemon%s\n", colorCyan, colorReset)

	return nil
}

func scheduleEnableCommand() *cli.Command {
	return &cli.Command{
		Name:      "enable",
		Usage:     "Enable backup schedule for a database",
		ArgsUsage: "<name>",
		Action:    runScheduleEnable,
	}
}

func runScheduleEnable(c *cli.Context) error {
	if c.NArg() == 0 {
		return fmt.Errorf("database name is required\n\nUsage: cadangkan schedule enable <name>")
	}

	name := c.Args().Get(0)

	mgr, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create config manager: %w", err)
	}

	dbConfig, err := mgr.GetDatabase(name)
	if err != nil {
		printError(fmt.Sprintf("Database '%s' not found", name))
		return err
	}

	if dbConfig.Schedule == nil || dbConfig.Schedule.Cron == "" {
		return fmt.Errorf("no schedule configured for '%s'\n\nSet a schedule first: cadangkan schedule set %s --daily", name, name)
	}

	dbConfig.Schedule.Enabled = true

	if err := mgr.AddDatabase(name, dbConfig); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	printSuccess(fmt.Sprintf("Schedule enabled for '%s'", name))
	return nil
}

func scheduleDisableCommand() *cli.Command {
	return &cli.Command{
		Name:      "disable",
		Usage:     "Disable backup schedule for a database",
		ArgsUsage: "<name>",
		Action:    runScheduleDisable,
	}
}

func runScheduleDisable(c *cli.Context) error {
	if c.NArg() == 0 {
		return fmt.Errorf("database name is required\n\nUsage: cadangkan schedule disable <name>")
	}

	name := c.Args().Get(0)

	mgr, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create config manager: %w", err)
	}

	dbConfig, err := mgr.GetDatabase(name)
	if err != nil {
		printError(fmt.Sprintf("Database '%s' not found", name))
		return err
	}

	if dbConfig.Schedule == nil {
		printInfo(fmt.Sprintf("No schedule configured for '%s'", name))
		return nil
	}

	dbConfig.Schedule.Enabled = false

	if err := mgr.AddDatabase(name, dbConfig); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	printSuccess(fmt.Sprintf("Schedule disabled for '%s'", name))
	return nil
}

func scheduleListCommand() *cli.Command {
	return &cli.Command{
		Name:   "list",
		Usage:  "List all backup schedules",
		Action: runScheduleList,
	}
}

func runScheduleList(c *cli.Context) error {
	mgr, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create config manager: %w", err)
	}

	cfg, err := mgr.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Collect scheduled databases
	type scheduleEntry struct {
		name     string
		config   *config.DatabaseConfig
		nextRun  time.Time
		schedule cron.Schedule
	}

	var entries []scheduleEntry
	for name, dbConfig := range cfg.Databases {
		if dbConfig.Schedule != nil && dbConfig.Schedule.Cron != "" {
			sched, err := cron.ParseStandard(dbConfig.Schedule.Cron)
			if err != nil {
				continue
			}
			entries = append(entries, scheduleEntry{
				name:     name,
				config:   dbConfig,
				nextRun:  sched.Next(time.Now()),
				schedule: sched,
			})
		}
	}

	if len(entries) == 0 {
		printInfo("No schedules configured")
		fmt.Println()
		fmt.Println("To add a schedule:")
		fmt.Printf("  %scadangkan schedule set <name> --daily --time=02:00%s\n", colorCyan, colorReset)
		return nil
	}

	// Sort by next run time
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].nextRun.Before(entries[j].nextRun)
	})

	// Display schedules
	fmt.Println()
	fmt.Printf("Backup Schedules (%d)\n", len(entries))
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()

	for _, entry := range entries {
		status := colorRed + "Disabled" + colorReset
		if entry.config.Schedule.Enabled {
			status = colorGreen + "Enabled" + colorReset
		}

		fmt.Printf("%s%-20s%s  %s\n", colorCyan, entry.name, colorReset, status)
		fmt.Printf("  Schedule:  %s\n", entry.config.Schedule.Cron)
		if entry.config.Schedule.Enabled {
			fmt.Printf("  Next run:  %s (%s)\n", entry.nextRun.Format("2006-01-02 15:04:05"), formatNextRun(entry.nextRun))
		}
		fmt.Println()
	}

	fmt.Println("To start scheduled backups:")
	fmt.Printf("  %scadangkan daemon%s\n", colorCyan, colorReset)
	fmt.Println()

	return nil
}

func scheduleNextCommand() *cli.Command {
	return &cli.Command{
		Name:   "next",
		Usage:  "Show next scheduled backup runs",
		Action: runScheduleNext,
	}
}

func runScheduleNext(c *cli.Context) error {
	mgr, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create config manager: %w", err)
	}

	cfg, err := mgr.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Collect enabled schedules
	type nextEntry struct {
		name    string
		nextRun time.Time
		cron    string
	}

	var entries []nextEntry
	for name, dbConfig := range cfg.Databases {
		if dbConfig.Schedule != nil && dbConfig.Schedule.Enabled {
			sched, err := cron.ParseStandard(dbConfig.Schedule.Cron)
			if err != nil {
				continue
			}
			entries = append(entries, nextEntry{
				name:    name,
				nextRun: sched.Next(time.Now()),
				cron:    dbConfig.Schedule.Cron,
			})
		}
	}

	if len(entries) == 0 {
		printInfo("No enabled schedules")
		return nil
	}

	// Sort by next run time
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].nextRun.Before(entries[j].nextRun)
	})

	// Display next runs
	fmt.Println()
	fmt.Println("Next Scheduled Backups")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()

	for _, entry := range entries {
		timeUntil := formatNextRun(entry.nextRun)
		fmt.Printf("%-20s  %s  %s(%s)%s\n",
			entry.name,
			entry.nextRun.Format("2006-01-02 15:04:05"),
			colorCyan,
			timeUntil,
			colorReset,
		)
	}
	fmt.Println()

	return nil
}

// parseDailyCron converts a time string (HH:MM) to a daily cron expression.
func parseDailyCron(timeStr string) (string, error) {
	t, err := time.Parse("15:04", timeStr)
	if err != nil {
		return "", fmt.Errorf("invalid time format '%s', use HH:MM (e.g., 02:00)", timeStr)
	}
	return fmt.Sprintf("%d %d * * *", t.Minute(), t.Hour()), nil
}

// parseWeeklyCron converts a time and day to a weekly cron expression.
func parseWeeklyCron(timeStr, dayStr string) (string, error) {
	t, err := time.Parse("15:04", timeStr)
	if err != nil {
		return "", fmt.Errorf("invalid time format '%s', use HH:MM (e.g., 03:00)", timeStr)
	}

	dayMap := map[string]int{
		"sunday":    0,
		"monday":    1,
		"tuesday":   2,
		"wednesday": 3,
		"thursday":  4,
		"friday":    5,
		"saturday":  6,
		"sun":       0,
		"mon":       1,
		"tue":       2,
		"wed":       3,
		"thu":       4,
		"fri":       5,
		"sat":       6,
	}

	day, ok := dayMap[strings.ToLower(dayStr)]
	if !ok {
		return "", fmt.Errorf("invalid day '%s', use full day name (e.g., sunday, monday)", dayStr)
	}

	return fmt.Sprintf("%d %d * * %d", t.Minute(), t.Hour(), day), nil
}
