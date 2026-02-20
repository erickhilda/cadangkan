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

// formatTimeAgo formats a time as "X ago" (e.g., "2 hours ago", "3 days ago")
func formatTimeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "just now"
	} else if diff < time.Hour {
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else if diff < 30*24*time.Hour {
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	} else if diff < 365*24*time.Hour {
		months := int(diff.Hours() / (24 * 30))
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	} else {
		years := int(diff.Hours() / (24 * 365))
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}

// formatNextRun formats a time as "in X" (e.g., "in 22 hours", "in 6 days")
func formatNextRun(t time.Time) string {
	now := time.Now()
	diff := t.Sub(now)

	if diff < 0 {
		return "overdue"
	} else if diff < time.Hour {
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "in 1 minute"
		}
		return fmt.Sprintf("in %d minutes", minutes)
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "in 1 hour"
		}
		return fmt.Sprintf("in %d hours", hours)
	} else {
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "in 1 day"
		}
		return fmt.Sprintf("in %d days", days)
	}
}

// getStatusIndicator returns a status indicator symbol based on status string
func getStatusIndicator(status string) string {
	switch status {
	case "healthy":
		return fmt.Sprintf("%s✓%s", colorGreen, colorReset)
	case "warning":
		return fmt.Sprintf("%s⚠%s", colorYellow, colorReset)
	case "critical":
		return fmt.Sprintf("%s✗%s", colorRed, colorReset)
	default:
		return "?"
	}
}

// formatStorageUsage formats storage usage as "X / Y (Z%)"
func formatStorageUsage(used int64, total uint64) string {
	if total == 0 {
		return backup.FormatBytes(used)
	}
	percentage := (float64(used) / float64(total)) * 100.0
	return fmt.Sprintf("%s / %s (%.1f%%)", backup.FormatBytes(used), backup.FormatBytes(int64(total)), percentage)
}

// formatAge formats a time as age string (e.g., "2 hours", "3 days")
func formatAge(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "< 1 minute"
	} else if diff < time.Hour {
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute"
		}
		return fmt.Sprintf("%d minutes", minutes)
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", hours)
	} else if diff < 30*24*time.Hour {
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day"
		}
		return fmt.Sprintf("%d days", days)
	} else if diff < 365*24*time.Hour {
		months := int(diff.Hours() / (24 * 30))
		if months == 1 {
			return "1 month"
		}
		return fmt.Sprintf("%d months", months)
	} else {
		years := int(diff.Hours() / (24 * 365))
		if years == 1 {
			return "1 year"
		}
		return fmt.Sprintf("%d years", years)
	}
}

