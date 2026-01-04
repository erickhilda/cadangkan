package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/erickhilda/cadangkan/internal/config"
	"github.com/erickhilda/cadangkan/pkg/database/mysql"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"
)

func addCommand() *cli.Command {
	return &cli.Command{
		Name:      "add",
		Usage:     "Add a database configuration",
		ArgsUsage: "mysql <name>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "host",
				Usage:    "Database host",
				Required: true,
			},
			&cli.IntFlag{
				Name:  "port",
				Usage: "Database port",
				Value: 3306,
			},
			&cli.StringFlag{
				Name:     "user",
				Usage:    "Database user",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "database",
				Usage:    "Database name",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "password",
				Usage: "Database password (prefer --password-stdin or interactive prompt)",
			},
			&cli.BoolFlag{
				Name:  "password-stdin",
				Usage: "Read password from stdin",
			},
			&cli.BoolFlag{
				Name:  "skip-test",
				Usage: "Skip connection test",
			},
		},
		Action: runAdd,
	}
}

func runAdd(c *cli.Context) error {
	// Parse arguments
	if c.NArg() < 2 {
		return fmt.Errorf("usage: cadangkan add mysql <name>")
	}

	dbType := c.Args().Get(0)
	name := c.Args().Get(1)

	if dbType != "mysql" {
		return fmt.Errorf("unsupported database type: %s (only 'mysql' is supported)", dbType)
	}

	// Sanitize name
	name = config.SanitizeName(name)
	if name == "" {
		return fmt.Errorf("invalid database name")
	}

	// Parse flags
	host := c.String("host")
	port := c.Int("port")
	user := c.String("user")
	database := c.String("database")
	password := c.String("password")
	passwordStdin := c.Bool("password-stdin")
	skipTest := c.Bool("skip-test")

	// Get password if not provided
	if password == "" {
		if passwordStdin {
			// Read from stdin
			reader := bufio.NewReader(os.Stdin)
			passwordBytes, err := io.ReadAll(reader)
			if err != nil {
				return fmt.Errorf("failed to read password from stdin: %w", err)
			}
			password = strings.TrimSpace(string(passwordBytes))
		} else {
			// Interactive prompt
			fmt.Print("Enter password: ")
			passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
			fmt.Println() // New line after password input
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}
			password = string(passwordBytes)
		}
	}

	if password == "" {
		return fmt.Errorf("password is required")
	}

	// Create config manager
	mgr, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create config manager: %w", err)
	}

	// Check if database already exists
	exists, err := mgr.DatabaseExists(name)
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	if exists {
		printWarning(fmt.Sprintf("Database '%s' already exists, it will be overwritten", name))
	}

	// Test connection (unless skipped)
	if !skipTest {
		printInfo(fmt.Sprintf("Testing connection to %s@%s:%d...", user, host, port))

		mysqlConfig := &mysql.Config{
			Host:     host,
			Port:     port,
			User:     user,
			Password: password,
			Database: database,
			Timeout:  10 * time.Second,
		}

		client, err := mysql.NewClient(mysqlConfig)
		if err != nil {
			printError("Failed to create MySQL client")
			return err
		}

		if err := client.Connect(); err != nil {
			printError("Connection failed")
			return fmt.Errorf("connection test failed: %w", err)
		}

		// Get database version
		dbVersion, err := client.GetVersion()
		if err != nil {
			dbVersion = "unknown"
		}

		client.Close()
		printSuccess(fmt.Sprintf("Connected successfully (MySQL %s)", dbVersion))
	}

	// Encrypt password
	printInfo("Encrypting password...")
	encryptedPassword, err := config.EncryptPassword(password)
	if err != nil {
		printError("Failed to encrypt password")
		return err
	}

	// Create database config
	dbConfig := &config.DatabaseConfig{
		Type:              "mysql",
		Host:              host,
		Port:              port,
		Database:          database,
		User:              user,
		PasswordEncrypted: encryptedPassword,
	}

	// Save to config
	printInfo("Saving configuration...")
	if err := mgr.AddDatabase(name, dbConfig); err != nil {
		printError("Failed to save configuration")
		return err
	}

	printSuccess(fmt.Sprintf("Database '%s' added successfully!", name))
	fmt.Println()
	fmt.Printf("You can now run: %scadangkan backup %s%s\n", colorCyan, name, colorReset)

	return nil
}
