package backup

import "time"

// BackupOptions defines configuration for a backup operation.
type BackupOptions struct {
	// Database name to backup
	Database string

	// ConfigName is the configuration name (used for storage paths)
	// If empty, falls back to using Database name
	ConfigName string

	// Tables to include (empty means all tables)
	Tables []string

	// Tables to exclude
	ExcludeTables []string

	// SchemaOnly backs up only the schema, not data
	SchemaOnly bool

	// Compression method: "gzip", "zstd", "none"
	Compression string

	// OutputPath is the directory where backup will be stored
	// If empty, uses default location (~/.cadangkan/backups/{database}/)
	OutputPath string
}

// BackupResult contains the result of a backup operation.
type BackupResult struct {
	// BackupID is the unique identifier for this backup (timestamp-based)
	BackupID string

	// FilePath is the full path to the backup file
	FilePath string

	// MetadataPath is the full path to the metadata file
	MetadataPath string

	// SizeBytes is the size of the backup file in bytes
	SizeBytes int64

	// Duration is how long the backup took
	Duration time.Duration

	// Checksum is the SHA-256 checksum of the backup file
	Checksum string

	// Status indicates the backup outcome
	Status string

	// StartedAt is when the backup started
	StartedAt time.Time

	// CompletedAt is when the backup completed
	CompletedAt time.Time

	// Error contains any error that occurred
	Error error
}

// BackupMetadata represents metadata stored with each backup.
type BackupMetadata struct {
	// Version of the metadata format
	Version string `json:"version"`

	// BackupID is the unique identifier for this backup
	BackupID string `json:"backup_id"`

	// Database information
	Database DatabaseInfo `json:"database"`

	// CreatedAt is when the backup started
	CreatedAt time.Time `json:"created_at"`

	// CompletedAt is when the backup finished
	CompletedAt time.Time `json:"completed_at"`

	// Duration in seconds
	DurationSeconds int64 `json:"duration_seconds"`

	// Status of the backup: "completed", "failed", "partial"
	Status string `json:"status"`

	// Backup file information
	Backup BackupFileInfo `json:"backup"`

	// Options used for this backup
	Options BackupOptionsInfo `json:"options"`

	// Tool information
	Tool ToolInfo `json:"tool"`

	// Error message if backup failed
	Error string `json:"error,omitempty"`
}

// DatabaseInfo contains information about the backed up database.
type DatabaseInfo struct {
	// Type of database: "mysql", "postgresql", etc.
	Type string `json:"type"`

	// Host where the database is located
	Host string `json:"host"`

	// Port the database is listening on
	Port int `json:"port"`

	// Database name
	Database string `json:"database"`

	// Version of the database server
	Version string `json:"version"`
}

// BackupFileInfo contains information about the backup file.
type BackupFileInfo struct {
	// File name (relative to backup directory)
	File string `json:"file"`

	// Size in bytes
	SizeBytes int64 `json:"size_bytes"`

	// Human-readable size
	SizeHuman string `json:"size_human"`

	// Compression method used
	Compression string `json:"compression"`

	// Checksum of the backup file (format: "sha256:...")
	Checksum string `json:"checksum"`
}

// BackupOptionsInfo contains the options used for the backup.
type BackupOptionsInfo struct {
	// SchemaOnly indicates if only schema was backed up
	SchemaOnly bool `json:"schema_only"`

	// Tables that were included (empty means all)
	Tables []string `json:"tables"`

	// Tables that were excluded
	ExcludeTables []string `json:"exclude_tables"`
}

// ToolInfo contains information about the tool that created the backup.
type ToolInfo struct {
	// Name of the tool
	Name string `json:"name"`

	// Version of the tool
	Version string `json:"version"`

	// MySQLDump version used (if applicable)
	MySQLDumpVersion string `json:"mysqldump_version,omitempty"`
}

// BackupProgress tracks the progress of an ongoing backup.
type BackupProgress struct {
	// Phase of backup: "connecting", "dumping", "compressing", "finalizing"
	Phase string

	// BytesWritten is the number of bytes written so far
	BytesWritten int64

	// Message is a human-readable progress message
	Message string

	// StartedAt is when this backup started
	StartedAt time.Time
}

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

// Constants for backup status
const (
	StatusCompleted = "completed"
	StatusFailed    = "failed"
	StatusPartial   = "partial"
	StatusRunning   = "running"
)

// Constants for compression types
const (
	CompressionGzip = "gzip"
	CompressionZstd = "zstd"
	CompressionNone = "none"
)

// Constants for backup phases
const (
	PhaseConnecting  = "connecting"
	PhaseDumping     = "dumping"
	PhaseCompressing = "compressing"
	PhaseFinalizing  = "finalizing"
)

// DefaultOptions returns BackupOptions with sensible defaults.
func DefaultOptions() *BackupOptions {
	return &BackupOptions{
		Compression:   CompressionGzip,
		SchemaOnly:    false,
		Tables:        []string{},
		ExcludeTables: []string{},
	}
}
