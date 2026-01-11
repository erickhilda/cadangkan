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

// hasPasswordFlag checks if --password flag appears in command line arguments
// This handles the case where --password is used without a value
func hasPasswordFlag() bool {
	for _, arg := range os.Args {
		// Check for --password as a standalone flag (but not --password-stdin)
		if arg == "--password" {
			return true
		}
		// Check for --password=value format
		if strings.HasPrefix(arg, "--password=") {
			return true
		}
	}
	return false
}

func editCommand() *cli.Command {
	return &cli.Command{
		Name:      "edit",
		Usage:     "Edit a database configuration",
		ArgsUsage: "[flags] <name>",
		Description: `Edit an existing database configuration.

   You can update individual fields without affecting others. Only the fields
   specified via flags will be updated. All other fields remain unchanged.

   IMPORTANT: Flags come first, followed by the database name.

   EXAMPLES:
     cadangkan edit --host=newhost.example.com production
     cadangkan edit --port=3307 production
     cadangkan edit --password production  # Interactive password prompt
     cadangkan edit --password=mypassword production  # Direct password (not recommended)
     cadangkan edit --host=newhost --port=3307 --skip-test production`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "host",
				Usage: "Update database host",
			},
			&cli.IntFlag{
				Name:  "port",
				Usage: "Update database port",
			},
			&cli.StringFlag{
				Name:  "user",
				Usage: "Update database user",
			},
			&cli.StringFlag{
				Name:  "database",
				Usage: "Update database name",
			},
			&cli.StringFlag{
				Name:  "password",
				Usage: "Update password (prefer --password-stdin or interactive prompt)",
			},
			&cli.BoolFlag{
				Name:  "password-stdin",
				Usage: "Read password from stdin",
			},
			&cli.BoolFlag{
				Name:  "skip-test",
				Usage: "Skip connection test after update",
			},
		},
		Action: runEdit,
	}
}

func runEdit(c *cli.Context) error {
	// Parse arguments - database name is the last argument
	// Handle case where --password without value consumes the database name
	name := ""
	password := c.String("password")
	passwordFlagProvided := hasPasswordFlag() || c.IsSet("password")

	if c.NArg() > 0 {
		// Normal case: database name is the last argument
		name = c.Args().Get(c.NArg() - 1)

		// If password equals the database name and --password flag was provided,
		// it means --password consumed the db name, so reset password to prompt
		if password == name && passwordFlagProvided {
			password = "" // Reset password, will prompt later
		}
	} else if password != "" && passwordFlagProvided {
		// Edge case: --password consumed the database name
		// Use password value as the database name and reset password to prompt
		name = password
		password = "" // Will prompt for password later
	} else {
		return fmt.Errorf("usage: cadangkan edit [flags] <name>")
	}

	// Sanitize name
	name = config.SanitizeName(name)
	if name == "" {
		return fmt.Errorf("invalid database name")
	}

	// Create config manager
	mgr, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create config manager: %w", err)
	}

	// Load existing config
	printInfo(fmt.Sprintf("Loading configuration for '%s'...", name))
	dbConfig, err := mgr.GetDatabase(name)
	if err != nil {
		printError("Database not found")
		return err
	}

	// Track if any changes were made
	hasChanges := false
	passwordChanged := false

	// Update host if provided
	if host := c.String("host"); host != "" {
		if host != dbConfig.Host {
			dbConfig.Host = host
			hasChanges = true
		}
	}

	// Update port if provided
	if port := c.Int("port"); port > 0 {
		if port != dbConfig.Port {
			dbConfig.Port = port
			hasChanges = true
		}
	}

	// Update user if provided
	if user := c.String("user"); user != "" {
		if user != dbConfig.User {
			dbConfig.User = user
			hasChanges = true
		}
	}

	// Update database name if provided
	if database := c.String("database"); database != "" {
		if database != dbConfig.Database {
			dbConfig.Database = database
			hasChanges = true
		}
	}

	// Handle password update
	passwordStdin := c.Bool("password-stdin")

	// If password is empty but flag was provided, get it from context
	// (it might have been set via --password=value)
	if password == "" && !passwordFlagProvided {
		password = c.String("password")
	}

	// If --password flag is set (with or without value) or --password-stdin is set
	if passwordFlagProvided || passwordStdin {
		// Get password if not provided via flag value
		// If password is still empty, prompt for it
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
				// Interactive prompt (when --password is used without value)
				fmt.Print("Enter new password: ")
				passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
				fmt.Println() // New line after password input
				if err != nil {
					return fmt.Errorf("failed to read password: %w", err)
				}
				password = string(passwordBytes)
			}
		}

		if password == "" {
			return fmt.Errorf("password cannot be empty")
		}

		passwordChanged = true
		hasChanges = true
	}

	// Check if any changes were made
	if !hasChanges {
		printInfo("No changes specified. Use flags to update fields.")
		fmt.Println()
		fmt.Printf("Current configuration:\n")
		fmt.Printf("  Host:     %s\n", dbConfig.Host)
		fmt.Printf("  Port:     %d\n", dbConfig.Port)
		fmt.Printf("  User:     %s\n", dbConfig.User)
		fmt.Printf("  Database: %s\n", dbConfig.Database)
		return nil
	}

	// Test connection (unless skipped)
	skipTest := c.Bool("skip-test")
	if !skipTest {
		// For connection test, we need the password
		testPassword := password
		if !passwordChanged {
			// Decrypt existing password for connection test
			decryptedPassword, err := config.DecryptPassword(dbConfig.PasswordEncrypted)
			if err != nil {
				printWarning("Failed to decrypt existing password for connection test")
				printInfo("Skipping connection test. Use --skip-test to suppress this warning.")
				testPassword = ""
			} else {
				testPassword = decryptedPassword
			}
		}

		if testPassword != "" {
			printInfo(fmt.Sprintf("Testing connection to %s@%s:%d...", dbConfig.User, dbConfig.Host, dbConfig.Port))

			mysqlConfig := &mysql.Config{
				Host:     dbConfig.Host,
				Port:     dbConfig.Port,
				User:     dbConfig.User,
				Password: testPassword,
				Database: dbConfig.Database,
				Timeout:  10 * time.Second,
			}

			client, err := mysql.NewClient(mysqlConfig)
			if err != nil {
				printError("Failed to create MySQL client")
				return err
			}

			if err := client.Connect(); err != nil {
				printError("Connection test failed")
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
	}

	// Encrypt password if it was changed
	if passwordChanged {
		printInfo("Encrypting password...")
		encryptedPassword, err := config.EncryptPassword(password)
		if err != nil {
			printError("Failed to encrypt password")
			return err
		}
		dbConfig.PasswordEncrypted = encryptedPassword
	}

	// Save updated config
	printInfo("Saving configuration...")
	if err := mgr.AddDatabase(name, dbConfig); err != nil {
		printError("Failed to save configuration")
		return err
	}

	printSuccess(fmt.Sprintf("Database '%s' updated successfully!", name))
	fmt.Println()
	fmt.Printf("You can test the connection with: %scadangkan test %s%s\n", colorCyan, name, colorReset)

	return nil
}
