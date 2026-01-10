package backup

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/erickhilda/cadangkan/pkg/database/mysql"
)

// MySQLRestorer executes mysql command to restore database backups.
type MySQLRestorer struct {
	config  *mysql.Config
	timeout time.Duration
}

// NewMySQLRestorer creates a new MySQLRestorer.
func NewMySQLRestorer(config *mysql.Config) *MySQLRestorer {
	timeout := 30 * time.Minute // Default 30 minute timeout
	if config.Timeout > 0 {
		timeout = config.Timeout * 6 // Multiply by 6 for restore operations
	}

	return &MySQLRestorer{
		config:  config,
		timeout: timeout,
	}
}

// Restore executes mysql command with SQL input from reader.
func (r *MySQLRestorer) Restore(database string, sqlReader io.Reader) error {
	return r.RestoreWithCommand(database, sqlReader, nil)
}

// RestoreWithCommand executes mysql command with SQL input from reader.
// If cmdLogger is provided, it will be called with the full command for debugging.
func (r *MySQLRestorer) RestoreWithCommand(database string, sqlReader io.Reader, cmdLogger func(string)) error {
	if database == "" {
		return WrapRestoreError("", "database name is required", fmt.Errorf("empty database name"))
	}

	// Build mysql command arguments
	args := r.buildArgs(database)

	// Log command if logger provided (for debugging)
	if cmdLogger != nil {
		// Mask password in logged command
		logArgs := make([]string, len(args))
		copy(logArgs, args)
		for i, arg := range logArgs {
			if strings.HasPrefix(arg, "--password=") {
				logArgs[i] = "--password=***"
			}
		}
		cmdStr := fmt.Sprintf("mysql %s", strings.Join(logArgs, " "))
		cmdLogger(cmdStr)
	}

	// Create command with context for timeout
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "mysql", args...)

	// Set stdin to read from sqlReader
	cmd.Stdin = sqlReader

	// Capture stderr to detect errors
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	// Execute command
	if err := cmd.Run(); err != nil {
		stderr := stderrBuf.String()
		exitCode := getRestoreExitCode(err)
		return WrapRestoreError(database, fmt.Sprintf("mysql restore failed (exit code %d)", exitCode), fmt.Errorf("stderr: %s", stderr))
	}

	// Check for warnings/errors in stderr even if exit code is 0
	stderr := stderrBuf.String()
	if stderr != "" {
		// Check for common error patterns
		errorPatterns := []string{
			"error",
			"failed",
			"cannot",
			"denied",
			"access denied",
			"unknown database",
		}

		lowerStderr := strings.ToLower(stderr)
		for _, pattern := range errorPatterns {
			if strings.Contains(lowerStderr, pattern) {
				return WrapRestoreError(database, "mysql restore completed but reported errors", fmt.Errorf("stderr: %s", stderr))
			}
		}
	}

	return nil
}

// buildArgs builds the mysql command arguments.
func (r *MySQLRestorer) buildArgs(database string) []string {
	args := []string{
		fmt.Sprintf("--host=%s", r.config.Host),
		fmt.Sprintf("--port=%d", r.config.Port),
		fmt.Sprintf("--user=%s", r.config.User),
	}

	// Add password if provided
	if r.config.Password != "" {
		args = append(args, fmt.Sprintf("--password=%s", r.config.Password))
	}

	// Add database name
	args = append(args, database)

	return args
}

// CheckMySQL checks if mysql command is available and returns its version.
func CheckMySQL() (string, error) {
	cmd := exec.Command("mysql", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("mysql not found or not executable: %w", err)
	}

	version := strings.TrimSpace(string(output))
	return version, nil
}

// getRestoreExitCode extracts exit code from command error.
func getRestoreExitCode(err error) int {
	if err == nil {
		return 0
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.ExitCode()
	}

	return -1
}
