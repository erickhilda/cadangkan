package backup

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/erickhilda/cadangkan/pkg/database/mysql"
)

const (
	// MetadataVersion is the current version of the metadata format
	MetadataVersion = "1.0"

	// ToolName is the name of this tool
	ToolName = "cadangkan"

	// ToolVersion is the version of this tool
	ToolVersion = "0.1.0"
)

// MetadataGenerator creates backup metadata.
type MetadataGenerator struct {
	client mysql.DatabaseClient
}

// NewMetadataGenerator creates a new MetadataGenerator.
func NewMetadataGenerator(client mysql.DatabaseClient) *MetadataGenerator {
	return &MetadataGenerator{
		client: client,
	}
}

// Generate generates complete backup metadata.
func (g *MetadataGenerator) Generate(
	backupID string,
	dbConfig *mysql.Config,
	result *BackupResult,
	options *BackupOptions,
	mysqldumpVersion string,
) (*BackupMetadata, error) {
	// Get database version if client is available and connected
	var dbVersion string
	if g.client != nil && g.client.IsConnected() {
		version, err := g.client.GetVersion()
		if err == nil {
			dbVersion = version
		}
	}

	// Get file name from path
	fileName := filepath.Base(result.FilePath)

	// Create metadata
	metadata := &BackupMetadata{
		Version:  MetadataVersion,
		BackupID: backupID,
		Database: DatabaseInfo{
			Type:     "mysql",
			Host:     dbConfig.Host,
			Port:     dbConfig.Port,
			Database: options.Database,
			Version:  dbVersion,
		},
		CreatedAt:       result.StartedAt,
		CompletedAt:     result.CompletedAt,
		DurationSeconds: int64(result.Duration.Seconds()),
		Status:          result.Status,
		Backup: BackupFileInfo{
			File:        fileName,
			SizeBytes:   result.SizeBytes,
			SizeHuman:   FormatBytes(result.SizeBytes),
			Compression: options.Compression,
			Checksum:    result.Checksum,
		},
		Options: BackupOptionsInfo{
			SchemaOnly:    options.SchemaOnly,
			Tables:        options.Tables,
			ExcludeTables: options.ExcludeTables,
		},
		Tool: ToolInfo{
			Name:             ToolName,
			Version:          ToolVersion,
			MySQLDumpVersion: mysqldumpVersion,
		},
	}

	// Set error if backup failed
	if result.Status == StatusFailed && result.Error != nil {
		metadata.Error = result.Error.Error()
	}

	return metadata, nil
}

// GenerateSimple generates metadata without database client (for testing).
func GenerateSimple(
	backupID string,
	database string,
	host string,
	port int,
	filePath string,
	sizeBytes int64,
	duration time.Duration,
	checksum string,
	compression string,
	status string,
) *BackupMetadata {
	fileName := filepath.Base(filePath)
	now := time.Now()

	return &BackupMetadata{
		Version:  MetadataVersion,
		BackupID: backupID,
		Database: DatabaseInfo{
			Type:     "mysql",
			Host:     host,
			Port:     port,
			Database: database,
			Version:  "", // Unknown
		},
		CreatedAt:       now.Add(-duration),
		CompletedAt:     now,
		DurationSeconds: int64(duration.Seconds()),
		Status:          status,
		Backup: BackupFileInfo{
			File:        fileName,
			SizeBytes:   sizeBytes,
			SizeHuman:   FormatBytes(sizeBytes),
			Compression: compression,
			Checksum:    checksum,
		},
		Options: BackupOptionsInfo{
			SchemaOnly:    false,
			Tables:        []string{},
			ExcludeTables: []string{},
		},
		Tool: ToolInfo{
			Name:    ToolName,
			Version: ToolVersion,
		},
	}
}

// UpdateMetadata updates an existing metadata with final information.
func UpdateMetadata(metadata *BackupMetadata, result *BackupResult) {
	metadata.CompletedAt = result.CompletedAt
	metadata.DurationSeconds = int64(result.Duration.Seconds())
	metadata.Status = result.Status
	metadata.Backup.SizeBytes = result.SizeBytes
	metadata.Backup.SizeHuman = FormatBytes(result.SizeBytes)
	metadata.Backup.Checksum = result.Checksum

	if result.Error != nil {
		metadata.Error = result.Error.Error()
	}
}

// CreateInitialMetadata creates initial metadata at the start of a backup.
func CreateInitialMetadata(
	backupID string,
	database string,
	dbConfig *mysql.Config,
	options *BackupOptions,
) *BackupMetadata {
	now := time.Now()

	return &BackupMetadata{
		Version:  MetadataVersion,
		BackupID: backupID,
		Database: DatabaseInfo{
			Type:     "mysql",
			Host:     dbConfig.Host,
			Port:     dbConfig.Port,
			Database: database,
		},
		CreatedAt:       now,
		DurationSeconds: 0,
		Status:          StatusRunning,
		Options: BackupOptionsInfo{
			SchemaOnly:    options.SchemaOnly,
			Tables:        options.Tables,
			ExcludeTables: options.ExcludeTables,
		},
		Tool: ToolInfo{
			Name:    ToolName,
			Version: ToolVersion,
		},
	}
}

// MarkFailed marks metadata as failed with an error message.
func MarkFailed(metadata *BackupMetadata, err error) {
	metadata.Status = StatusFailed
	metadata.CompletedAt = time.Now()
	metadata.DurationSeconds = int64(metadata.CompletedAt.Sub(metadata.CreatedAt).Seconds())
	if err != nil {
		metadata.Error = err.Error()
	}
}

// MarkCompleted marks metadata as completed.
func MarkCompleted(metadata *BackupMetadata) {
	metadata.Status = StatusCompleted
	metadata.CompletedAt = time.Now()
	metadata.DurationSeconds = int64(metadata.CompletedAt.Sub(metadata.CreatedAt).Seconds())
}

// ValidateMetadata validates metadata structure.
func ValidateMetadata(metadata *BackupMetadata) error {
	if metadata.BackupID == "" {
		return &MetadataError{
			Message: "backup ID is required",
		}
	}

	if metadata.Database.Database == "" {
		return &MetadataError{
			BackupID: metadata.BackupID,
			Message:  "database name is required",
		}
	}

	if metadata.Status == "" {
		return &MetadataError{
			BackupID: metadata.BackupID,
			Message:  "status is required",
		}
	}

	return nil
}

// GetBackupAge returns the age of a backup in duration.
func GetBackupAge(metadata *BackupMetadata) time.Duration {
	return time.Since(metadata.CreatedAt)
}

// IsBackupOlderThan checks if a backup is older than the specified duration.
func IsBackupOlderThan(metadata *BackupMetadata, duration time.Duration) bool {
	return GetBackupAge(metadata) > duration
}

// FormatDuration formats a duration in a human-readable way.
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) - (minutes * 60)
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) - (hours * 60)
	return fmt.Sprintf("%dh %dm", hours, minutes)
}

// GetMySQLDumpVersion gets the mysqldump version.
func GetMySQLDumpVersion() string {
	version, err := CheckMySQLDump()
	if err != nil {
		return "unknown"
	}
	return version
}
