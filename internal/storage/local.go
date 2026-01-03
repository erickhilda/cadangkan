package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// LocalStorage manages local file system storage for backups.
type LocalStorage struct {
	// basePath is the base directory for all backups
	// Default: ~/.cadangkan/backups
	basePath string
}

// NewLocalStorage creates a new LocalStorage instance.
// If basePath is empty, uses ~/.cadangkan/backups
func NewLocalStorage(basePath string) (*LocalStorage, error) {
	if basePath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		basePath = filepath.Join(homeDir, ".cadangkan", "backups")
	}

	return &LocalStorage{
		basePath: basePath,
	}, nil
}

// GetBasePath returns the base path for backups.
func (s *LocalStorage) GetBasePath() string {
	return s.basePath
}

// GetDatabasePath returns the directory path for a specific database.
func (s *LocalStorage) GetDatabasePath(database string) string {
	return filepath.Join(s.basePath, database)
}

// EnsureDatabaseDir ensures the backup directory for a database exists.
func (s *LocalStorage) EnsureDatabaseDir(database string) error {
	dbPath := s.GetDatabasePath(database)
	if err := os.MkdirAll(dbPath, 0755); err != nil {
		return &StorageError{
			Path:    dbPath,
			Op:      "create",
			Message: "failed to create database directory",
			Err:     err,
		}
	}
	return nil
}

// CheckDiskSpace checks available disk space at the base path.
func (s *LocalStorage) CheckDiskSpace() (uint64, error) {
	available, err := checkDiskSpace(s.basePath)
	if err != nil {
		return 0, &StorageError{
			Path:    s.basePath,
			Op:      "check",
			Message: "failed to check disk space",
			Err:     err,
		}
	}
	return available, nil
}

// HasEnoughSpace checks if there's enough space for estimated backup size.
func (s *LocalStorage) HasEnoughSpace(estimatedSize int64) (bool, error) {
	available, err := s.CheckDiskSpace()
	if err != nil {
		return false, err
	}

	// Require estimated size plus 20% buffer
	requiredSize := uint64(float64(estimatedSize) * 1.2)
	return available >= requiredSize, nil
}

// GetBackupPath returns the full path for a backup file.
func (s *LocalStorage) GetBackupPath(database, backupID, compression string) string {
	dbPath := s.GetDatabasePath(database)
	return getBackupFilePath(dbPath, backupID, compression)
}

// GetMetadataPath returns the full path for a metadata file.
func (s *LocalStorage) GetMetadataPath(database, backupID string) string {
	dbPath := s.GetDatabasePath(database)
	return filepath.Join(dbPath, backupID+".meta.json")
}

// ListBackups lists all backups for a database.
func (s *LocalStorage) ListBackups(database string) ([]BackupListEntry, error) {
	dbPath := s.GetDatabasePath(database)

	// Check if directory exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return []BackupListEntry{}, nil
	}

	// Read directory
	entries, err := os.ReadDir(dbPath)
	if err != nil {
		return nil, &StorageError{
			Path:    dbPath,
			Op:      "read",
			Message: "failed to read backup directory",
			Err:     err,
		}
	}

	// Find all metadata files
	var backups []BackupListEntry
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".meta.json") {
			continue
		}

		// Parse metadata
		metaPath := filepath.Join(dbPath, name)
		var meta MetadataStub
		err := s.LoadMetadata(database, strings.TrimSuffix(name, ".meta.json"), &meta)
		if err != nil {
			// Skip invalid metadata files
			continue
		}

		// Find the backup file
		backupPath := filepath.Join(dbPath, meta.Backup.File)
		fileInfo, err := os.Stat(backupPath)
		if err != nil {
			// Backup file missing, skip
			continue
		}

		backups = append(backups, BackupListEntry{
			BackupID:     meta.BackupID,
			Database:     database,
			CreatedAt:    meta.CreatedAt,
			SizeBytes:    fileInfo.Size(),
			SizeHuman:    meta.Backup.SizeHuman,
			Status:       meta.Status,
			FilePath:     backupPath,
			MetadataPath: metaPath,
		})
	}

	// Sort by creation time (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt.After(backups[j].CreatedAt)
	})

	return backups, nil
}

// SaveMetadata saves backup metadata to a JSON file.
// metadata should be a struct that can be marshaled to JSON.
func (s *LocalStorage) SaveMetadata(database string, backupID string, metadata interface{}) error {
	metaPath := s.GetMetadataPath(database, backupID)

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return &MetadataError{
			BackupID: backupID,
			Message:  "failed to marshal metadata",
			Err:      err,
		}
	}

	if err := os.WriteFile(metaPath, data, 0644); err != nil {
		return &StorageError{
			Path:    metaPath,
			Op:      "write",
			Message: "failed to write metadata file",
			Err:     err,
		}
	}

	return nil
}

// LoadMetadata loads backup metadata from a JSON file into the provided struct.
// result should be a pointer to a struct that can be unmarshaled from JSON.
func (s *LocalStorage) LoadMetadata(database, backupID string, result interface{}) error {
	metaPath := s.GetMetadataPath(database, backupID)

	data, err := os.ReadFile(metaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrBackupNotFound
		}
		return &StorageError{
			Path:    metaPath,
			Op:      "read",
			Message: "failed to read metadata file",
			Err:     err,
		}
	}

	if err := json.Unmarshal(data, result); err != nil {
		return &MetadataError{
			BackupID: backupID,
			Message:  "failed to unmarshal metadata",
			Err:      err,
		}
	}

	return nil
}

// DeleteBackup deletes a backup and its metadata.
func (s *LocalStorage) DeleteBackup(database, backupID string) error {
	// Load metadata to get backup file name
	var meta MetadataStub
	err := s.LoadMetadata(database, backupID, &meta)
	if err != nil {
		return err
	}

	// Delete backup file
	backupPath := filepath.Join(s.GetDatabasePath(database), meta.Backup.File)
	if err := os.Remove(backupPath); err != nil && !os.IsNotExist(err) {
		return &StorageError{
			Path:    backupPath,
			Op:      "delete",
			Message: "failed to delete backup file",
			Err:     err,
		}
	}

	// Delete metadata file
	metaPath := s.GetMetadataPath(database, backupID)
	if err := os.Remove(metaPath); err != nil && !os.IsNotExist(err) {
		return &StorageError{
			Path:    metaPath,
			Op:      "delete",
			Message: "failed to delete metadata file",
			Err:     err,
		}
	}

	return nil
}

// CleanupPartialBackup removes a partial backup (both file and metadata if they exist).
func (s *LocalStorage) CleanupPartialBackup(database, backupID, compression string) error {
	// Try to delete backup file
	backupPath := s.GetBackupPath(database, backupID, compression)
	if err := os.Remove(backupPath); err != nil && !os.IsNotExist(err) {
		// Log but don't fail on cleanup errors
		fmt.Fprintf(os.Stderr, "Warning: failed to cleanup backup file %s: %v\n", backupPath, err)
	}

	// Try to delete metadata file
	metaPath := s.GetMetadataPath(database, backupID)
	if err := os.Remove(metaPath); err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Warning: failed to cleanup metadata file %s: %v\n", metaPath, err)
	}

	return nil
}

// GetLatestBackup returns the most recent backup for a database.
func (s *LocalStorage) GetLatestBackup(database string) (*BackupListEntry, error) {
	backups, err := s.ListBackups(database)
	if err != nil {
		return nil, err
	}

	if len(backups) == 0 {
		return nil, ErrBackupNotFound
	}

	return &backups[0], nil
}

// Helper functions

func getBackupFilePath(backupDir, backupID, compression string) string {
	var ext string
	switch compression {
	case CompressionGzip:
		ext = ".sql.gz"
	case CompressionZstd:
		ext = ".sql.zst"
	case CompressionNone:
		ext = ".sql"
	default:
		ext = ".sql.gz"
	}

	return filepath.Join(backupDir, backupID+ext)
}

func checkDiskSpace(path string) (uint64, error) {
	// Try to stat the path
	_, err := os.Stat(path)
	if err != nil {
		// If path doesn't exist, try parent directory
		if os.IsNotExist(err) {
			parentPath := filepath.Dir(path)
			_, err = os.Stat(parentPath)
			if err != nil {
				return 0, err
			}
			// Use parent path for disk space check
			return getDiskSpace(parentPath)
		}
		return 0, err
	}

	// Get filesystem stats using syscall
	// Note: This is platform-specific (Linux/Unix)
	// For cross-platform support, we'd need build tags
	return getDiskSpace(path)
}
