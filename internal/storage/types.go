package storage

import (
	"errors"
	"fmt"
	"time"
)

// BackupListEntry represents a single backup in a list of backups.
type BackupListEntry struct {
	// BackupID is the unique identifier
	BackupID string

	// Database name
	Database string

	// CreatedAt is when the backup was created
	CreatedAt time.Time

	// SizeBytes is the size of the backup file
	SizeBytes int64

	// SizeHuman is the human-readable size
	SizeHuman string

	// Status of the backup
	Status string

	// FilePath is the full path to the backup file
	FilePath string

	// MetadataPath is the full path to the metadata file
	MetadataPath string
}

// MetadataStub is a minimal representation of metadata for listing.
type MetadataStub struct {
	BackupID  string    `json:"backup_id"`
	CreatedAt time.Time `json:"created_at"`
	Status    string    `json:"status"`
	Backup    struct {
		File      string `json:"file"`
		SizeHuman string `json:"size_human"`
	} `json:"backup"`
}

// Constants for compression types
const (
	CompressionGzip = "gzip"
	CompressionZstd = "zstd"
	CompressionNone = "none"
)

// Common errors
var (
	ErrBackupNotFound = errors.New("backup not found")
)

// StorageError represents a storage operation error.
type StorageError struct {
	Path    string
	Op      string
	Message string
	Err     error
}

func (e *StorageError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("storage error (%s) at %s: %s: %v", e.Op, e.Path, e.Message, e.Err)
	}
	return fmt.Sprintf("storage error (%s) at %s: %s", e.Op, e.Path, e.Message)
}

func (e *StorageError) Unwrap() error {
	return e.Err
}

// MetadataError represents a metadata operation error.
type MetadataError struct {
	BackupID string
	Message  string
	Err      error
}

func (e *MetadataError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("metadata error for backup %s: %s: %v", e.BackupID, e.Message, e.Err)
	}
	return fmt.Sprintf("metadata error for backup %s: %s", e.BackupID, e.Message)
}

func (e *MetadataError) Unwrap() error {
	return e.Err
}
