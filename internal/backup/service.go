package backup

import (
	"fmt"
	"time"

	"github.com/erickhilda/cadangkan/internal/storage"
	"github.com/erickhilda/cadangkan/pkg/database/mysql"
)

// Service orchestrates backup operations.
type Service struct {
	client   mysql.DatabaseClient
	storage  *storage.LocalStorage
	config   *mysql.Config
}

// NewService creates a new backup service.
func NewService(client mysql.DatabaseClient, stor *storage.LocalStorage, config *mysql.Config) *Service {
	return &Service{
		client:  client,
		storage: stor,
		config:  config,
	}
}

// Backup performs a complete backup operation.
func (s *Service) Backup(options *BackupOptions) (*BackupResult, error) {
	if options == nil {
		options = DefaultOptions()
	}

	// Validate options
	if err := s.validateOptions(options); err != nil {
		return nil, err
	}

	// Generate backup ID
	backupID := GenerateBackupID()
	startTime := time.Now()

	// Initialize result
	result := &BackupResult{
		BackupID:  backupID,
		StartedAt: startTime,
		Status:    StatusRunning,
	}

	// Ensure database directory exists
	if err := s.storage.EnsureDatabaseDir(options.Database); err != nil {
		return nil, err
	}

	// Check disk space
	if err := s.checkDiskSpace(options); err != nil {
		return nil, err
	}

	// Get file paths
	result.FilePath = s.storage.GetBackupPath(options.Database, backupID, options.Compression)
	result.MetadataPath = s.storage.GetMetadataPath(options.Database, backupID)

	// Create initial metadata
	metadata := CreateInitialMetadata(backupID, options.Database, s.config, options)

	// Perform backup with cleanup on failure
	err := s.performBackup(options, result)
	if err != nil {
		// Clean up partial backup
		s.storage.CleanupPartialBackup(options.Database, backupID, options.Compression)
		
		// Mark metadata as failed
		MarkFailed(metadata, err)
		s.storage.SaveMetadata(options.Database, backupID, metadata)
		
		return nil, err
	}

	// Calculate final result
	result.CompletedAt = time.Now()
	result.Duration = result.CompletedAt.Sub(result.StartedAt)
	result.Status = StatusCompleted

	// Get mysqldump version
	mysqldumpVersion := GetMySQLDumpVersion()

	// Generate final metadata
	metaGen := NewMetadataGenerator(s.client)
	finalMetadata, err := metaGen.Generate(backupID, s.config, result, options, mysqldumpVersion)
	if err != nil {
		return nil, WrapMetadataError(backupID, "failed to generate metadata", err)
	}

	// Save metadata
	if err := s.storage.SaveMetadata(options.Database, backupID, finalMetadata); err != nil {
		return nil, err
	}

	return result, nil
}

// performBackup executes the actual backup process.
func (s *Service) performBackup(options *BackupOptions, result *BackupResult) error {
	// Create mysqldump options
	dumpOpts := &DumpOptions{
		Tables:        options.Tables,
		ExcludeTables: options.ExcludeTables,
		SchemaOnly:    options.SchemaOnly,
		Routines:      true,
		Triggers:      true,
		Events:        true,
	}

	// Create dumper
	dumper := NewMySQLDumper(s.config)

	// Get dump reader
	dumpReader, err := dumper.Dump(options.Database, dumpOpts)
	if err != nil {
		return WrapBackupError(options.Database, "failed to start dump", err)
	}
	defer dumpReader.Close()

	// Create compressor
	compressor := NewCompressor(options.Compression)

	// Stream dump to compressed file with checksum
	compressResult, err := compressor.StreamCompress(dumpReader, result.FilePath)
	if err != nil {
		return WrapBackupError(options.Database, "failed to compress backup", err)
	}

	// Update result with compression info
	result.SizeBytes = compressResult.BytesWritten
	result.Checksum = compressResult.Checksum

	return nil
}

// validateOptions validates backup options.
func (s *Service) validateOptions(options *BackupOptions) error {
	if options.Database == "" {
		return ErrDatabaseRequired
	}

	// Validate compression type
	switch options.Compression {
	case CompressionGzip, CompressionNone:
		// Valid
	case CompressionZstd:
		return &ValidationError{
			Field:   "Compression",
			Message: "zstd compression not yet implemented",
		}
	default:
		return &ValidationError{
			Field:   "Compression",
			Message: fmt.Sprintf("invalid compression type: %s", options.Compression),
		}
	}

	// Validate tables and exclude tables don't overlap
	if len(options.Tables) > 0 && len(options.ExcludeTables) > 0 {
		return &ValidationError{
			Field:   "Tables",
			Message: "cannot specify both tables and exclude_tables",
		}
	}

	return nil
}

// checkDiskSpace verifies there is enough disk space for the backup.
func (s *Service) checkDiskSpace(options *BackupOptions) error {
	// Try to estimate database size if client is connected
	var estimatedSize int64 = 1024 * 1024 * 1024 // Default 1GB

	if s.client != nil && s.client.IsConnected() {
		size, err := s.client.GetDatabaseSize(options.Database)
		if err == nil && size > 0 {
			// Estimate compressed size (typically 30-40% of original)
			estimatedSize = EstimateBackupSize(size, options.Compression)
		}
	}

	// Check if we have enough space
	hasSpace, err := s.storage.HasEnoughSpace(estimatedSize)
	if err != nil {
		return WrapStorageError(s.storage.GetBasePath(), "check", "failed to check disk space", err)
	}

	if !hasSpace {
		available, _ := s.storage.CheckDiskSpace()
		return &StorageError{
			Path:    s.storage.GetBasePath(),
			Op:      "check",
			Message: fmt.Sprintf("insufficient disk space: need ~%s, have %s", FormatBytes(estimatedSize), FormatBytes(int64(available))),
		}
	}

	return nil
}

// ListBackups lists all backups for a database.
func (s *Service) ListBackups(database string) ([]BackupListEntry, error) {
	storageList, err := s.storage.ListBackups(database)
	if err != nil {
		return nil, err
	}

	// Convert storage.BackupListEntry to backup.BackupListEntry
	backupList := make([]BackupListEntry, len(storageList))
	for i, entry := range storageList {
		backupList[i] = BackupListEntry{
			BackupID:     entry.BackupID,
			Database:     entry.Database,
			CreatedAt:    entry.CreatedAt,
			SizeBytes:    entry.SizeBytes,
			SizeHuman:    entry.SizeHuman,
			Status:       entry.Status,
			FilePath:     entry.FilePath,
			MetadataPath: entry.MetadataPath,
		}
	}

	return backupList, nil
}

// GetBackup retrieves metadata for a specific backup.
func (s *Service) GetBackup(database, backupID string) (*BackupMetadata, error) {
	var metadata BackupMetadata
	err := s.storage.LoadMetadata(database, backupID, &metadata)
	if err != nil {
		return nil, err
	}
	return &metadata, nil
}

// GetLatestBackup retrieves the most recent backup for a database.
func (s *Service) GetLatestBackup(database string) (*BackupListEntry, error) {
	storageEntry, err := s.storage.GetLatestBackup(database)
	if err != nil {
		return nil, err
	}

	// Convert storage.BackupListEntry to backup.BackupListEntry
	return &BackupListEntry{
		BackupID:     storageEntry.BackupID,
		Database:     storageEntry.Database,
		CreatedAt:    storageEntry.CreatedAt,
		SizeBytes:    storageEntry.SizeBytes,
		SizeHuman:    storageEntry.SizeHuman,
		Status:       storageEntry.Status,
		FilePath:     storageEntry.FilePath,
		MetadataPath: storageEntry.MetadataPath,
	}, nil
}

// DeleteBackup deletes a backup and its metadata.
func (s *Service) DeleteBackup(database, backupID string) error {
	return s.storage.DeleteBackup(database, backupID)
}

// VerifyBackup verifies a backup's integrity by checking its checksum.
func (s *Service) VerifyBackup(database, backupID string) (bool, error) {
	// Load metadata
	var metadata BackupMetadata
	err := s.storage.LoadMetadata(database, backupID, &metadata)
	if err != nil {
		return false, err
	}

	// Get backup file path
	backupPath := s.storage.GetBackupPath(database, backupID, metadata.Backup.Compression)

	// Verify checksum
	valid, err := VerifyChecksum(backupPath, metadata.Backup.Checksum)
	if err != nil {
		return false, WrapBackupError(database, "failed to verify checksum", err)
	}

	return valid, nil
}

// GetBackupSize returns the size of a backup in bytes.
func (s *Service) GetBackupSize(database, backupID string) (int64, error) {
	// Load metadata
	var metadata BackupMetadata
	err := s.storage.LoadMetadata(database, backupID, &metadata)
	if err != nil {
		return 0, err
	}

	return metadata.Backup.SizeBytes, nil
}

// QuickBackup performs a backup with default options.
func (s *Service) QuickBackup(database string) (*BackupResult, error) {
	options := DefaultOptions()
	options.Database = database
	return s.Backup(options)
}

// SchemaBackup performs a schema-only backup.
func (s *Service) SchemaBackup(database string) (*BackupResult, error) {
	options := DefaultOptions()
	options.Database = database
	options.SchemaOnly = true
	return s.Backup(options)
}

// TableBackup backs up specific tables.
func (s *Service) TableBackup(database string, tables []string) (*BackupResult, error) {
	options := DefaultOptions()
	options.Database = database
	options.Tables = tables
	return s.Backup(options)
}

// BackupWithProgress performs a backup with progress callback.
// The callback receives progress updates during the backup.
type ProgressCallback func(progress *BackupProgress)

func (s *Service) BackupWithProgress(options *BackupOptions, callback ProgressCallback) (*BackupResult, error) {
	// For now, just call regular Backup
	// Progress tracking will be implemented in a future version
	if callback != nil {
		callback(&BackupProgress{
			Phase:   PhaseConnecting,
			Message: "Starting backup...",
		})
	}

	result, err := s.Backup(options)

	if callback != nil {
		if err != nil {
			callback(&BackupProgress{
				Phase:   PhaseFinalizing,
				Message: fmt.Sprintf("Backup failed: %v", err),
			})
		} else {
			callback(&BackupProgress{
				Phase:        PhaseFinalizing,
				Message:      "Backup completed successfully",
				BytesWritten: result.SizeBytes,
			})
		}
	}

	return result, err
}

// EstimateDatabaseSize estimates the size of a database.
func (s *Service) EstimateDatabaseSize(database string) (int64, error) {
	if s.client == nil || !s.client.IsConnected() {
		return 0, fmt.Errorf("client not connected")
	}

	return s.client.GetDatabaseSize(database)
}

// CheckConnectivity verifies connection to the database.
func (s *Service) CheckConnectivity() error {
	if s.client == nil {
		return fmt.Errorf("client not initialized")
	}

	if !s.client.IsConnected() {
		return fmt.Errorf("client not connected")
	}

	return s.client.Ping()
}
