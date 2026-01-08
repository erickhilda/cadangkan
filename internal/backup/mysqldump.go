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

// MySQLDumper executes mysqldump to create database backups.
type MySQLDumper struct {
	config  *mysql.Config
	timeout time.Duration
}

// NewMySQLDumper creates a new MySQLDumper.
func NewMySQLDumper(config *mysql.Config) *MySQLDumper {
	timeout := 30 * time.Minute // Default 30 minute timeout
	if config.Timeout > 0 {
		timeout = config.Timeout * 6 // Multiply by 6 for dump operations
	}

	return &MySQLDumper{
		config:  config,
		timeout: timeout,
	}
}

// DumpOptions configures mysqldump execution.
type DumpOptions struct {
	Tables        []string
	ExcludeTables []string
	SchemaOnly    bool
	NoData        bool
	Routines      bool
	Triggers      bool
	Events        bool
}

// DefaultDumpOptions returns optimal default options for mysqldump.
func DefaultDumpOptions() *DumpOptions {
	return &DumpOptions{
		SchemaOnly: false,
		NoData:     false,
		Routines:   true,
		Triggers:   true,
		Events:     true,
	}
}

// DumpResult contains the result of a dump operation.
type DumpResult struct {
	BytesWritten int64
	Duration     time.Duration
	ExitCode     int
	Stderr       string
}

// Dump executes mysqldump and returns a reader for the output.
// The caller is responsible for closing the returned reader.
func (d *MySQLDumper) Dump(database string, options *DumpOptions) (io.ReadCloser, error) {
	return d.DumpWithCommand(database, options, nil)
}

// DumpWithCommand executes mysqldump and returns a reader for the output.
// If cmdLogger is provided, it will be called with the full command for debugging.
func (d *MySQLDumper) DumpWithCommand(database string, options *DumpOptions, cmdLogger func(string)) (io.ReadCloser, error) {
	if options == nil {
		options = DefaultDumpOptions()
	}

	// Build mysqldump command
	args := d.buildArgs(database, options)

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
		cmdStr := fmt.Sprintf("mysqldump %s", strings.Join(logArgs, " "))
		cmdLogger(cmdStr)
	}

	// Create command with context for timeout
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)

	cmd := exec.CommandContext(ctx, "mysqldump", args...)

	// Capture stderr to detect warnings/errors
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	// Get stdout pipe
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, WrapDumpError(database, "mysqldump", "failed to create stdout pipe", 0, err)
	}

	// Start command
	if err := cmd.Start(); err != nil {
		cancel()
		return nil, WrapDumpError(database, "mysqldump", "failed to start mysqldump", 0, err)
	}

	// Return a reader that will handle cleanup
	return &dumpReader{
		reader:   stdout,
		cmd:      cmd,
		cancel:   cancel,
		database: database,
		stderr:   &stderrBuf,
	}, nil
}

// DumpToWriter executes mysqldump and writes output directly to a writer.
func (d *MySQLDumper) DumpToWriter(database string, writer io.Writer, options *DumpOptions) (*DumpResult, error) {
	if options == nil {
		options = DefaultDumpOptions()
	}

	startTime := time.Now()

	// Build mysqldump command
	args := d.buildArgs(database, options)

	// Create command with context for timeout
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "mysqldump", args...)

	// Capture stderr
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	// Get stdout pipe
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, WrapDumpError(database, "mysqldump", "failed to create stdout pipe", 0, err)
	}

	// Start command
	if err := cmd.Start(); err != nil {
		return nil, WrapDumpError(database, "mysqldump", "failed to start mysqldump", 0, err)
	}

	// Copy output to writer
	bytesWritten, err := io.Copy(writer, stdout)
	if err != nil {
		cmd.Process.Kill()
		cmd.Wait()
		return nil, WrapDumpError(database, "mysqldump", "failed to copy output", 0, err)
	}

	// Wait for command to finish
	if err := cmd.Wait(); err != nil {
		stderr := stderrBuf.String()
		exitCode := getExitCode(err)
		return nil, WrapDumpError(database, strings.Join(args, " "), stderr, exitCode, err)
	}

	duration := time.Since(startTime)

	return &DumpResult{
		BytesWritten: bytesWritten,
		Duration:     duration,
		ExitCode:     0,
		Stderr:       stderrBuf.String(),
	}, nil
}

// buildArgs builds the mysqldump command arguments.
func (d *MySQLDumper) buildArgs(database string, options *DumpOptions) []string {
	args := []string{
		fmt.Sprintf("--host=%s", d.config.Host),
		fmt.Sprintf("--port=%d", d.config.Port),
		fmt.Sprintf("--user=%s", d.config.User),
	}

	// Add password if provided
	if d.config.Password != "" {
		args = append(args, fmt.Sprintf("--password=%s", d.config.Password))
	}

	// Optimal flags for consistency and performance
	args = append(args,
		"--single-transaction",  // Consistent snapshot without locking tables
		"--quick",               // Don't buffer entire result in memory
		"--skip-lock-tables",    // Don't lock tables (use single-transaction instead)
		"--no-tablespaces",      // Avoid tablespace issues
		"--set-gtid-purged=OFF", // Don't include GTID info (causes issues with some setups)
	)

	// Add routines, triggers, events if requested
	if options.Routines {
		args = append(args, "--routines")
	}
	if options.Triggers {
		args = append(args, "--triggers")
	}
	if options.Events {
		args = append(args, "--events")
	}

	// Schema-only or no-data
	if options.SchemaOnly || options.NoData {
		args = append(args, "--no-data")
	}

	// Add database name
	args = append(args, database)

	// Specific tables
	if len(options.Tables) > 0 {
		args = append(args, options.Tables...)
	}

	// Exclude tables
	for _, table := range options.ExcludeTables {
		args = append(args, fmt.Sprintf("--ignore-table=%s.%s", database, table))
	}

	return args
}

// CheckMySQLDump checks if mysqldump is available and returns its version.
func CheckMySQLDump() (string, error) {
	cmd := exec.Command("mysqldump", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("mysqldump not found or not executable: %w", err)
	}

	version := strings.TrimSpace(string(output))
	return version, nil
}

// dumpReader wraps the stdout pipe and handles command cleanup.
type dumpReader struct {
	reader   io.ReadCloser
	cmd      *exec.Cmd
	cancel   context.CancelFunc
	database string
	stderr   *bytes.Buffer
	closed   bool
}

// Read implements io.Reader.
func (r *dumpReader) Read(p []byte) (n int, err error) {
	return r.reader.Read(p)
}

// Close implements io.Closer.
func (r *dumpReader) Close() error {
	if r.closed {
		return nil
	}
	r.closed = true

	// Close the reader
	if err := r.reader.Close(); err != nil {
		r.cancel()
		return err
	}

	// Wait for command to finish
	err := r.cmd.Wait()
	stderr := ""
	if r.stderr != nil {
		stderr = r.stderr.String()
	}
	r.cancel()

	if err != nil {
		exitCode := getExitCode(err)
		return WrapDumpError(r.database, "mysqldump", stderr, exitCode, err)
	}

	// Check for warnings in stderr even if exit code is 0
	// mysqldump may succeed but only dump schema if there are permission issues
	if stderr != "" {
		// Check for common warning patterns that indicate problems
		warningPatterns := []string{
			"access denied",
			"got error",
			"warning:",
			"error:",
			"mysqldump:",
			"cannot",
			"failed",
			"denied",
		}

		lowerStderr := strings.ToLower(stderr)
		for _, pattern := range warningPatterns {
			if strings.Contains(lowerStderr, pattern) {
				// Return error with stderr to surface the warning
				return WrapDumpError(r.database, "mysqldump", stderr, 0, fmt.Errorf("mysqldump completed but reported warnings: %s", stderr))
			}
		}
	}

	return nil
}

// getExitCode extracts exit code from command error.
func getExitCode(err error) int {
	if err == nil {
		return 0
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.ExitCode()
	}

	return -1
}

// ExecuteMySQLDump executes mysqldump and streams output to writer with compression.
// This is the main high-level function used by the backup service.
func ExecuteMySQLDump(config *mysql.Config, database string, writer io.Writer, options *DumpOptions) (*DumpResult, error) {
	dumper := NewMySQLDumper(config)
	return dumper.DumpToWriter(database, writer, options)
}

// StreamMySQLDump executes mysqldump and returns a reader for streaming.
// Useful when you need more control over the output processing.
func StreamMySQLDump(config *mysql.Config, database string, options *DumpOptions) (io.ReadCloser, error) {
	dumper := NewMySQLDumper(config)
	return dumper.Dump(database, options)
}

// ValidateMySQLDumpOptions validates dump options.
func ValidateMySQLDumpOptions(options *DumpOptions) error {
	if options == nil {
		return fmt.Errorf("dump options cannot be nil")
	}

	// Check for conflicting options
	if options.SchemaOnly && len(options.Tables) > 0 {
		// This is allowed but warn in logs
	}

	return nil
}
