package backup

import (
	"errors"
	"fmt"
)

// Common sentinel errors for backup operations.
var (
	// ErrInvalidOptions indicates that the provided backup options are invalid.
	ErrInvalidOptions = errors.New("backup: invalid options")

	// ErrDatabaseRequired indicates that a database name is required.
	ErrDatabaseRequired = errors.New("backup: database name is required")

	// ErrBackupNotFound indicates that the requested backup was not found.
	ErrBackupNotFound = errors.New("backup: backup not found")

	// ErrInsufficientSpace indicates that there is not enough disk space.
	ErrInsufficientSpace = errors.New("backup: insufficient disk space")

	// ErrBackupInProgress indicates that a backup is already in progress.
	ErrBackupInProgress = errors.New("backup: backup already in progress")
)

// BackupError represents a general backup error.
type BackupError struct {
	Database string
	Message  string
	Err      error
}

// Error returns the error message.
func (e *BackupError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("backup error for database %s: %s: %v", e.Database, e.Message, e.Err)
	}
	return fmt.Sprintf("backup error for database %s: %s", e.Database, e.Message)
}

// Unwrap returns the underlying error.
func (e *BackupError) Unwrap() error {
	return e.Err
}

// DumpError represents an error during mysqldump execution.
type DumpError struct {
	Database string
	Command  string
	Stderr   string
	ExitCode int
	Err      error
}

// Error returns the error message.
func (e *DumpError) Error() string {
	msg := fmt.Sprintf("mysqldump error for database %s (exit code %d)", e.Database, e.ExitCode)
	if e.Stderr != "" {
		// Truncate stderr if too long
		stderr := e.Stderr
		if len(stderr) > 500 {
			stderr = stderr[:500] + "... (truncated)"
		}
		msg += fmt.Sprintf(": %s", stderr)
	}
	if e.Err != nil {
		msg += fmt.Sprintf(": %v", e.Err)
	}
	return msg
}

// Unwrap returns the underlying error.
func (e *DumpError) Unwrap() error {
	return e.Err
}

// CompressionError represents an error during compression.
type CompressionError struct {
	File    string
	Message string
	Err     error
}

// Error returns the error message.
func (e *CompressionError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("compression error for %s: %s: %v", e.File, e.Message, e.Err)
	}
	return fmt.Sprintf("compression error for %s: %s", e.File, e.Message)
}

// Unwrap returns the underlying error.
func (e *CompressionError) Unwrap() error {
	return e.Err
}

// StorageError represents an error related to storage operations.
type StorageError struct {
	Path    string
	Op      string // Operation: "create", "write", "delete", "check"
	Message string
	Err     error
}

// Error returns the error message.
func (e *StorageError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("storage error (%s) at %s: %s: %v", e.Op, e.Path, e.Message, e.Err)
	}
	return fmt.Sprintf("storage error (%s) at %s: %s", e.Op, e.Path, e.Message)
}

// Unwrap returns the underlying error.
func (e *StorageError) Unwrap() error {
	return e.Err
}

// MetadataError represents an error related to metadata operations.
type MetadataError struct {
	BackupID string
	Message  string
	Err      error
}

// Error returns the error message.
func (e *MetadataError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("metadata error for backup %s: %s: %v", e.BackupID, e.Message, e.Err)
	}
	return fmt.Sprintf("metadata error for backup %s: %s", e.BackupID, e.Message)
}

// Unwrap returns the underlying error.
func (e *MetadataError) Unwrap() error {
	return e.Err
}

// ValidationError represents an error during option validation.
type ValidationError struct {
	Field   string
	Message string
}

// Error returns the error message.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s: %s", e.Field, e.Message)
}

// IsBackupError checks if the error is a BackupError.
func IsBackupError(err error) bool {
	var backupErr *BackupError
	return errors.As(err, &backupErr)
}

// IsDumpError checks if the error is a DumpError.
func IsDumpError(err error) bool {
	var dumpErr *DumpError
	return errors.As(err, &dumpErr)
}

// IsCompressionError checks if the error is a CompressionError.
func IsCompressionError(err error) bool {
	var compErr *CompressionError
	return errors.As(err, &compErr)
}

// IsStorageError checks if the error is a StorageError.
func IsStorageError(err error) bool {
	var storageErr *StorageError
	return errors.As(err, &storageErr)
}

// IsMetadataError checks if the error is a MetadataError.
func IsMetadataError(err error) bool {
	var metaErr *MetadataError
	return errors.As(err, &metaErr)
}

// IsValidationError checks if the error is a ValidationError.
func IsValidationError(err error) bool {
	var valErr *ValidationError
	return errors.As(err, &valErr)
}

// WrapBackupError wraps an error as a BackupError.
func WrapBackupError(database, message string, err error) error {
	return &BackupError{
		Database: database,
		Message:  message,
		Err:      err,
	}
}

// WrapDumpError wraps an error as a DumpError.
func WrapDumpError(database, command, stderr string, exitCode int, err error) error {
	return &DumpError{
		Database: database,
		Command:  command,
		Stderr:   stderr,
		ExitCode: exitCode,
		Err:      err,
	}
}

// WrapCompressionError wraps an error as a CompressionError.
func WrapCompressionError(file, message string, err error) error {
	return &CompressionError{
		File:    file,
		Message: message,
		Err:     err,
	}
}

// WrapStorageError wraps an error as a StorageError.
func WrapStorageError(path, op, message string, err error) error {
	return &StorageError{
		Path:    path,
		Op:      op,
		Message: message,
		Err:     err,
	}
}

// WrapMetadataError wraps an error as a MetadataError.
func WrapMetadataError(backupID, message string, err error) error {
	return &MetadataError{
		BackupID: backupID,
		Message:  message,
		Err:      err,
	}
}

// RestoreError represents a general restore error.
type RestoreError struct {
	Database string
	Message  string
	Err      error
}

// Error returns the error message.
func (e *RestoreError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("restore error for database %s: %s: %v", e.Database, e.Message, e.Err)
	}
	return fmt.Sprintf("restore error for database %s: %s", e.Database, e.Message)
}

// Unwrap returns the underlying error.
func (e *RestoreError) Unwrap() error {
	return e.Err
}

// BackupNotFoundError indicates that the requested backup was not found.
type BackupNotFoundError struct {
	BackupID string
	Database string
}

// Error returns the error message.
func (e *BackupNotFoundError) Error() string {
	if e.Database != "" {
		return fmt.Sprintf("backup %s not found for database %s", e.BackupID, e.Database)
	}
	return fmt.Sprintf("backup %s not found", e.BackupID)
}

// ChecksumMismatchError indicates that the backup checksum doesn't match.
type ChecksumMismatchError struct {
	BackupID         string
	ExpectedChecksum string
	ActualChecksum   string
}

// Error returns the error message.
func (e *ChecksumMismatchError) Error() string {
	return fmt.Sprintf("checksum mismatch for backup %s: expected %s, got %s", e.BackupID, e.ExpectedChecksum, e.ActualChecksum)
}

// IsRestoreError checks if the error is a RestoreError.
func IsRestoreError(err error) bool {
	var restoreErr *RestoreError
	return errors.As(err, &restoreErr)
}

// IsBackupNotFoundError checks if the error is a BackupNotFoundError.
func IsBackupNotFoundError(err error) bool {
	var notFoundErr *BackupNotFoundError
	return errors.As(err, &notFoundErr)
}

// IsChecksumMismatchError checks if the error is a ChecksumMismatchError.
func IsChecksumMismatchError(err error) bool {
	var checksumErr *ChecksumMismatchError
	return errors.As(err, &checksumErr)
}

// WrapRestoreError wraps an error as a RestoreError.
func WrapRestoreError(database, message string, err error) error {
	return &RestoreError{
		Database: database,
		Message:  message,
		Err:      err,
	}
}
