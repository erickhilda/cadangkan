package backup

import (
	"fmt"
	"os"
	"time"

	"github.com/erickhilda/cadangkan/internal/storage"
	"github.com/erickhilda/cadangkan/pkg/database/mysql"
)

// RestoreService orchestrates restore operations.
type RestoreService struct {
	client  mysql.DatabaseClient
	storage *storage.LocalStorage
	config  *mysql.Config
	verbose bool
}

// NewRestoreService creates a new restore service.
func NewRestoreService(client mysql.DatabaseClient, stor *storage.LocalStorage, config *mysql.Config) *RestoreService {
	return &RestoreService{
		client:  client,
		storage: stor,
		config:  config,
		verbose: false,
	}
}

// SetVerbose enables or disables verbose logging.
func (s *RestoreService) SetVerbose(verbose bool) {
	s.verbose = verbose
}

// Restore performs a complete restore operation.
func (s *RestoreService) Restore(options *RestoreOptions) (*RestoreResult, error) {
	if options == nil {
		return nil, WrapRestoreError("", "restore options are required", fmt.Errorf("nil options"))
	}

	startTime := time.Now()

	// Determine target database
	targetDatabase := options.Database
	if options.TargetDatabase != "" {
		targetDatabase = options.TargetDatabase
	}

	if targetDatabase == "" {
		return nil, WrapRestoreError("", "target database is required", fmt.Errorf("empty database name"))
	}

	// Initialize result
	result := &RestoreResult{
		TargetDatabase: targetDatabase,
		StartedAt:      startTime,
		Status:         RestoreStatusFailed,
	}

	// Get storage name (config name if available, otherwise database name)
	storageName := getStorageNameForRestore(options)

	// Load backup metadata
	backupEntry, err := s.loadBackupMetadata(storageName, options.BackupID)
	if err != nil {
		result.Error = err
		return nil, err
	}

	result.BackupID = backupEntry.BackupID

	// Load full metadata to get compression info
	var metadata BackupMetadata
	err = s.storage.LoadMetadata(storageName, backupEntry.BackupID, &metadata)
	if err != nil {
		result.Error = WrapRestoreError(targetDatabase, "failed to load backup metadata", err)
		return nil, result.Error
	}

	// Validate backup file exists
	backupPath := backupEntry.FilePath
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		result.Error = &BackupNotFoundError{
			BackupID: backupEntry.BackupID,
			Database: storageName,
		}
		return nil, result.Error
	}

	// Verify checksum if available
	if metadata.Backup.Checksum != "" {
		valid, err := VerifyChecksum(backupPath, metadata.Backup.Checksum)
		if err != nil {
			result.Error = WrapRestoreError(targetDatabase, "failed to verify checksum", err)
			return nil, result.Error
		}
		if !valid {
			result.Error = &ChecksumMismatchError{
				BackupID:         backupEntry.BackupID,
				ExpectedChecksum: metadata.Backup.Checksum,
				ActualChecksum:   "", // We don't recalculate, just report mismatch
			}
			return nil, result.Error
		}
	}

	// Check if database exists
	dbExists, err := s.client.DatabaseExists(targetDatabase)
	if err != nil {
		result.Error = WrapRestoreError(targetDatabase, "failed to check if database exists", err)
		return nil, result.Error
	}

	// Create database if needed
	if !dbExists {
		if options.CreateDatabase {
			if s.verbose {
				fmt.Printf("[DEBUG] Creating database %s\n", targetDatabase)
			}
			if err := s.client.CreateDatabase(targetDatabase); err != nil {
				result.Error = WrapRestoreError(targetDatabase, "failed to create database", err)
				return nil, result.Error
			}
		} else {
			result.Error = WrapRestoreError(targetDatabase, "database does not exist", fmt.Errorf("use --create-db to create it"))
			return nil, result.Error
		}
	}

	// If backup-first and database exists, create backup
	if options.BackupFirst && dbExists {
		// Note: This would require creating a backup service instance
		// For now, we'll skip this and let the CLI handle it
		// TODO: Implement backup-first functionality
	}

	// Dry-run: validate without executing
	if options.DryRun {
		result.Status = RestoreStatusCompleted
		result.CompletedAt = time.Now()
		result.Duration = result.CompletedAt.Sub(result.StartedAt)
		return result, nil
	}

	// Decompress and restore
	compression := metadata.Backup.Compression
	if compression == "" {
		compression = CompressionGzip // Default
	}

	// Open backup file
	backupFile, err := os.Open(backupPath)
	if err != nil {
		result.Error = WrapRestoreError(targetDatabase, "failed to open backup file", err)
		return nil, result.Error
	}
	defer backupFile.Close()

	// Create decompressor
	decompressor := NewDecompressor(compression)

	// Create MySQL restorer with config that includes target database
	// The restorer needs the database name for the mysql command
	restorerConfig := &mysql.Config{
		Host:     s.config.Host,
		Port:     s.config.Port,
		User:     s.config.User,
		Password: s.config.Password,
		Database: targetDatabase, // Target database for restore command
		Timeout:  s.config.Timeout,
	}
	restorer := NewMySQLRestorer(restorerConfig)

	// Restore with decompression
	var cmdLogger func(string)
	if s.verbose {
		cmdLogger = func(cmd string) {
			fmt.Printf("[DEBUG] %s\n", cmd)
		}
	}

	// Create a pipe: decompressor -> restorer
	// We'll use a temporary approach: decompress to a pipe reader
	decompressedReader, err := decompressor.DecompressToReader(backupFile)
	if err != nil {
		result.Error = WrapRestoreError(targetDatabase, "failed to decompress backup", err)
		return nil, result.Error
	}
	defer decompressedReader.Close()

	// Execute restore
	if err := restorer.RestoreWithCommand(targetDatabase, decompressedReader, cmdLogger); err != nil {
		result.Error = WrapRestoreError(targetDatabase, "restore failed", err)
		return nil, result.Error
	}

	// Success
	result.Status = RestoreStatusCompleted
	result.CompletedAt = time.Now()
	result.Duration = result.CompletedAt.Sub(result.StartedAt)

	return result, nil
}

// loadBackupMetadata loads backup metadata (latest or specific).
func (s *RestoreService) loadBackupMetadata(storageName, backupID string) (*storage.BackupListEntry, error) {
	if backupID == "" {
		// Get latest backup
		entry, err := s.storage.GetLatestBackup(storageName)
		if err != nil {
			return nil, &BackupNotFoundError{
				BackupID: "latest",
				Database: storageName,
			}
		}
		return entry, nil
	}

	// Get specific backup
	backups, err := s.storage.ListBackups(storageName)
	if err != nil {
		return nil, WrapRestoreError(storageName, "failed to list backups", err)
	}

	for _, backup := range backups {
		if backup.BackupID == backupID {
			return &backup, nil
		}
	}

	return nil, &BackupNotFoundError{
		BackupID: backupID,
		Database: storageName,
	}
}

// getStorageNameForRestore returns the storage name for restore operations.
func getStorageNameForRestore(options *RestoreOptions) string {
	if options.ConfigName != "" {
		return options.ConfigName
	}
	return options.Database
}
