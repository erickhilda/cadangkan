package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/erickhilda/cadangkan/internal/backup"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
)

// printSuccess prints a success message with a green checkmark
func printSuccess(message string) {
	fmt.Printf("%s✓%s %s\n", colorGreen, colorReset, message)
}

// printError prints an error message with a red X
func printError(message string) {
	fmt.Printf("%s✗%s %s\n", colorRed, colorReset, message)
}

// printInfo prints an info message with a blue icon
func printInfo(message string) {
	fmt.Printf("%sℹ%s %s\n", colorBlue, colorReset, message)
}

// printWarning prints a warning message with a yellow icon
func printWarning(message string) {
	fmt.Printf("%s⚠%s %s\n", colorYellow, colorReset, message)
}

// showSpinner displays a simple spinner animation while backup is running
func showSpinner(done chan bool) {
	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	i := 0
	for {
		select {
		case <-done:
			fmt.Print("\r") // Clear the spinner line
			return
		default:
			fmt.Printf("\r%s Backing up... ", spinner[i%len(spinner)])
			i++
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// formatBackupResult formats and displays the backup result
func formatBackupResult(result *backup.BackupResult, database string) {
	// Get home directory for path display
	homeDir, _ := os.UserHomeDir()
	displayPath := result.FilePath
	if homeDir != "" && strings.HasPrefix(result.FilePath, homeDir) {
		displayPath = "~" + strings.TrimPrefix(result.FilePath, homeDir)
	}

	// Display compact checksum (first 16 chars)
	checksum := result.Checksum
	if len(checksum) > 23 { // "sha256:" + 16 chars
		checksum = checksum[:23] + "..."
	}

	fmt.Printf("  %sBackup ID:%s   %s\n", colorCyan, colorReset, result.BackupID)
	fmt.Printf("  %sDatabase:%s    %s\n", colorCyan, colorReset, database)
	fmt.Printf("  %sFile:%s        %s\n", colorCyan, colorReset, displayPath)
	fmt.Printf("  %sSize:%s        %s\n", colorCyan, colorReset, backup.FormatBytes(result.SizeBytes))
	fmt.Printf("  %sDuration:%s    %s\n", colorCyan, colorReset, backup.FormatDuration(result.Duration))
	fmt.Printf("  %sChecksum:%s    %s\n", colorCyan, colorReset, checksum)
	fmt.Println()
	fmt.Printf("Backup saved to: %s\n", displayPath)
}

// getConfigPath returns the path to the config file
func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".cadangkan", "config.yaml"), nil
}

// ensureConfigDir ensures the config directory exists
func ensureConfigDir() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configDir := filepath.Join(homeDir, ".cadangkan")
	return os.MkdirAll(configDir, 0755)
}
