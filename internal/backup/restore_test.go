package backup

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/erickhilda/cadangkan/internal/storage"
	"github.com/erickhilda/cadangkan/pkg/database/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRestoreService(t *testing.T) {
	mockClient := mysql.NewMockClient()
	mockClient.SetConnected(true)

	config := &mysql.Config{
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "password",
		Database: "testdb",
		Timeout:  10 * time.Second,
	}

	tmpDir := t.TempDir()
	localStorage, err := storage.NewLocalStorage(tmpDir)
	require.NoError(t, err)

	service := NewRestoreService(mockClient, localStorage, config)
	assert.NotNil(t, service)
	assert.Equal(t, mockClient, service.client)
	assert.Equal(t, localStorage, service.storage)
	assert.Equal(t, config, service.config)
	assert.False(t, service.verbose)
}

func TestRestoreServiceSetVerbose(t *testing.T) {
	mockClient := mysql.NewMockClient()
	config := &mysql.Config{Host: "localhost", User: "root"}
	tmpDir := t.TempDir()
	localStorage, _ := storage.NewLocalStorage(tmpDir)

	service := NewRestoreService(mockClient, localStorage, config)
	assert.False(t, service.verbose)

	service.SetVerbose(true)
	assert.True(t, service.verbose)

	service.SetVerbose(false)
	assert.False(t, service.verbose)
}

func TestRestoreServiceRestore(t *testing.T) {
	t.Run("nil options", func(t *testing.T) {
		mockClient := mysql.NewMockClient()
		config := &mysql.Config{Host: "localhost", User: "root"}
		tmpDir := t.TempDir()
		localStorage, _ := storage.NewLocalStorage(tmpDir)

		service := NewRestoreService(mockClient, localStorage, config)
		result, err := service.Restore(nil)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, IsRestoreError(err))
	})

	t.Run("empty target database", func(t *testing.T) {
		mockClient := mysql.NewMockClient()
		config := &mysql.Config{Host: "localhost", User: "root"}
		tmpDir := t.TempDir()
		localStorage, _ := storage.NewLocalStorage(tmpDir)

		service := NewRestoreService(mockClient, localStorage, config)
		options := &RestoreOptions{
			Database: "",
		}

		result, err := service.Restore(options)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, IsRestoreError(err))
	})

	t.Run("backup not found", func(t *testing.T) {
		mockClient := mysql.NewMockClient()
		mockClient.SetConnected(true)
		config := &mysql.Config{Host: "localhost", User: "root", Database: "testdb"}
		tmpDir := t.TempDir()
		localStorage, _ := storage.NewLocalStorage(tmpDir)

		service := NewRestoreService(mockClient, localStorage, config)
		options := &RestoreOptions{
			Database:   "testdb",
			BackupID:   "nonexistent",
			ConfigName: "testdb",
		}

		result, err := service.Restore(options)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, IsBackupNotFoundError(err))
	})

	t.Run("database does not exist, CreateDatabase=false", func(t *testing.T) {
		mockClient := mysql.NewMockClient()
		mockClient.SetConnected(true)
		// Database not in list
		mockClient.Databases = []string{"otherdb"}

		config := &mysql.Config{Host: "localhost", User: "root", Database: "testdb"}
		tmpDir := t.TempDir()
		localStorage, _ := storage.NewLocalStorage(tmpDir)

		// Create a backup file and metadata
		backupID := "2025-01-15-143022"
		dbPath := filepath.Join(tmpDir, "testdb")
		require.NoError(t, os.MkdirAll(dbPath, 0755))

		// Create backup file
		backupFile := filepath.Join(dbPath, backupID+".sql.gz")
		createTestBackupFile(t, backupFile, "CREATE TABLE test (id INT);")

		// Create metadata
		metadata := createTestMetadata(backupID, "testdb", backupFile, "gzip")
		metadataPath := filepath.Join(dbPath, backupID+".meta.json")
		saveMetadata(t, metadataPath, metadata)

		service := NewRestoreService(mockClient, localStorage, config)
		options := &RestoreOptions{
			Database:       "testdb",
			BackupID:       backupID,
			ConfigName:     "testdb",
			CreateDatabase: false,
		}

		result, err := service.Restore(options)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "database does not exist")
	})

	t.Run("database does not exist, CreateDatabase=true", func(t *testing.T) {
		mockClient := mysql.NewMockClient()
		mockClient.SetConnected(true)
		// Database not in list initially
		mockClient.Databases = []string{"otherdb"}

		config := &mysql.Config{Host: "localhost", User: "root", Database: "testdb"}
		tmpDir := t.TempDir()
		localStorage, _ := storage.NewLocalStorage(tmpDir)

		// Create a backup file and metadata
		backupID := "2025-01-15-143022"
		dbPath := filepath.Join(tmpDir, "testdb")
		require.NoError(t, os.MkdirAll(dbPath, 0755))

		backupFile := filepath.Join(dbPath, backupID+".sql.gz")
		createTestBackupFile(t, backupFile, "CREATE TABLE test (id INT);")

		metadata := createTestMetadata(backupID, "testdb", backupFile, "gzip")
		metadataPath := filepath.Join(dbPath, backupID+".meta.json")
		saveMetadata(t, metadataPath, metadata)

		service := NewRestoreService(mockClient, localStorage, config)
		options := &RestoreOptions{
			Database:       "testdb",
			BackupID:       backupID,
			ConfigName:     "testdb",
			CreateDatabase: true,
			DryRun:         true, // Use dry-run to avoid actual mysql command
		}

		result, err := service.Restore(options)
		// Dry-run should succeed
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, RestoreStatusCompleted, result.Status)

		// Verify database was created
		exists, _ := mockClient.DatabaseExists("testdb")
		assert.True(t, exists, "database should have been created")
	})

	t.Run("dry-run mode", func(t *testing.T) {
		mockClient := mysql.NewMockClient()
		mockClient.SetConnected(true)
		mockClient.Databases = []string{"testdb"}

		config := &mysql.Config{Host: "localhost", User: "root", Database: "testdb"}
		tmpDir := t.TempDir()
		localStorage, _ := storage.NewLocalStorage(tmpDir)

		backupID := "2025-01-15-143022"
		dbPath := filepath.Join(tmpDir, "testdb")
		require.NoError(t, os.MkdirAll(dbPath, 0755))

		backupFile := filepath.Join(dbPath, backupID+".sql.gz")
		createTestBackupFile(t, backupFile, "CREATE TABLE test (id INT);")

		metadata := createTestMetadata(backupID, "testdb", backupFile, "gzip")
		metadataPath := filepath.Join(dbPath, backupID+".meta.json")
		saveMetadata(t, metadataPath, metadata)

		service := NewRestoreService(mockClient, localStorage, config)
		options := &RestoreOptions{
			Database:   "testdb",
			BackupID:   backupID,
			ConfigName: "testdb",
			DryRun:     true,
		}

		result, err := service.Restore(options)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, RestoreStatusCompleted, result.Status)
		assert.Equal(t, backupID, result.BackupID)
		assert.Equal(t, "testdb", result.TargetDatabase)
	})

	t.Run("checksum mismatch", func(t *testing.T) {
		mockClient := mysql.NewMockClient()
		mockClient.SetConnected(true)
		mockClient.Databases = []string{"testdb"}

		config := &mysql.Config{Host: "localhost", User: "root", Database: "testdb"}
		tmpDir := t.TempDir()
		localStorage, _ := storage.NewLocalStorage(tmpDir)

		backupID := "2025-01-15-143022"
		dbPath := filepath.Join(tmpDir, "testdb")
		require.NoError(t, os.MkdirAll(dbPath, 0755))

		backupFile := filepath.Join(dbPath, backupID+".sql.gz")
		createTestBackupFile(t, backupFile, "CREATE TABLE test (id INT);")

		// Calculate actual checksum first
		actualChecksum, err := CalculateChecksum(backupFile)
		require.NoError(t, err)

		// Create metadata with wrong checksum (different from actual)
		wrongChecksum := "sha256:wrong_checksum_value_that_does_not_match"
		metadata := createTestMetadata(backupID, "testdb", backupID+".sql.gz", "gzip")
		metadata.Backup.Checksum = wrongChecksum
		metadataPath := filepath.Join(dbPath, backupID+".meta.json")
		saveMetadata(t, metadataPath, metadata)

		service := NewRestoreService(mockClient, localStorage, config)
		options := &RestoreOptions{
			Database:   "testdb",
			BackupID:   backupID,
			ConfigName: "testdb",
			DryRun:     true,
		}

		result, err := service.Restore(options)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, IsChecksumMismatchError(err))

		// Verify error includes actual checksum
		var checksumErr *ChecksumMismatchError
		require.ErrorAs(t, err, &checksumErr)
		assert.Equal(t, wrongChecksum, checksumErr.ExpectedChecksum)
		assert.Equal(t, actualChecksum, checksumErr.ActualChecksum, "Error should include actual checksum for debugging")
		assert.Contains(t, err.Error(), actualChecksum, "Error message should include actual checksum")
	})

	t.Run("checksum validation passes", func(t *testing.T) {
		mockClient := mysql.NewMockClient()
		mockClient.SetConnected(true)
		mockClient.Databases = []string{"testdb"}

		config := &mysql.Config{Host: "localhost", User: "root", Database: "testdb"}
		tmpDir := t.TempDir()
		localStorage, _ := storage.NewLocalStorage(tmpDir)

		backupID := "2025-01-15-143022"
		dbPath := filepath.Join(tmpDir, "testdb")
		require.NoError(t, os.MkdirAll(dbPath, 0755))

		backupFile := filepath.Join(dbPath, backupID+".sql.gz")
		createTestBackupFile(t, backupFile, "CREATE TABLE test (id INT);")

		// Calculate actual checksum
		actualChecksum, err := CalculateChecksum(backupFile)
		require.NoError(t, err)

		// Create metadata with correct checksum
		metadata := createTestMetadata(backupID, "testdb", backupID+".sql.gz", "gzip")
		metadata.Backup.Checksum = actualChecksum
		metadataPath := filepath.Join(dbPath, backupID+".meta.json")
		saveMetadata(t, metadataPath, metadata)

		service := NewRestoreService(mockClient, localStorage, config)
		options := &RestoreOptions{
			Database:   "testdb",
			BackupID:   backupID,
			ConfigName: "testdb",
			DryRun:     true,
		}

		result, err := service.Restore(options)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, RestoreStatusCompleted, result.Status)
	})

	t.Run("backup file missing", func(t *testing.T) {
		mockClient := mysql.NewMockClient()
		mockClient.SetConnected(true)
		mockClient.Databases = []string{"testdb"}

		config := &mysql.Config{Host: "localhost", User: "root", Database: "testdb"}
		tmpDir := t.TempDir()
		localStorage, _ := storage.NewLocalStorage(tmpDir)

		backupID := "2025-01-15-143022"
		dbPath := filepath.Join(tmpDir, "testdb")
		require.NoError(t, os.MkdirAll(dbPath, 0755))

		// Create metadata but no backup file
		backupFile := filepath.Join(dbPath, backupID+".sql.gz")
		metadata := createTestMetadata(backupID, "testdb", backupFile, "gzip")
		metadataPath := filepath.Join(dbPath, backupID+".meta.json")
		saveMetadata(t, metadataPath, metadata)

		service := NewRestoreService(mockClient, localStorage, config)
		options := &RestoreOptions{
			Database:   "testdb",
			BackupID:   backupID,
			ConfigName: "testdb",
			DryRun:     true,
		}

		result, err := service.Restore(options)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, IsBackupNotFoundError(err))
	})

	t.Run("target database override", func(t *testing.T) {
		mockClient := mysql.NewMockClient()
		mockClient.SetConnected(true)
		mockClient.Databases = []string{"targetdb"}

		config := &mysql.Config{Host: "localhost", User: "root", Database: "sourcedb"}
		tmpDir := t.TempDir()
		localStorage, _ := storage.NewLocalStorage(tmpDir)

		backupID := "2025-01-15-143022"
		dbPath := filepath.Join(tmpDir, "sourcedb")
		require.NoError(t, os.MkdirAll(dbPath, 0755))

		backupFile := filepath.Join(dbPath, backupID+".sql.gz")
		createTestBackupFile(t, backupFile, "CREATE TABLE test (id INT);")

		metadata := createTestMetadata(backupID, "sourcedb", backupFile, "gzip")
		metadataPath := filepath.Join(dbPath, backupID+".meta.json")
		saveMetadata(t, metadataPath, metadata)

		service := NewRestoreService(mockClient, localStorage, config)
		options := &RestoreOptions{
			Database:       "sourcedb",
			TargetDatabase: "targetdb",
			BackupID:       backupID,
			ConfigName:     "sourcedb",
			DryRun:         true,
		}

		result, err := service.Restore(options)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "targetdb", result.TargetDatabase)
	})
}

func TestRestoreServiceLoadBackupMetadata(t *testing.T) {
	t.Run("latest backup", func(t *testing.T) {
		mockClient := mysql.NewMockClient()
		config := &mysql.Config{Host: "localhost", User: "root"}
		tmpDir := t.TempDir()
		localStorage, _ := storage.NewLocalStorage(tmpDir)

		service := NewRestoreService(mockClient, localStorage, config)

		// Create multiple backups
		dbPath := filepath.Join(tmpDir, "testdb")
		require.NoError(t, os.MkdirAll(dbPath, 0755))

		backup1 := "2025-01-15-100000"
		backup2 := "2025-01-15-143022" // Latest

		createTestBackupFile(t, filepath.Join(dbPath, backup1+".sql.gz"), "SQL1")
		createTestBackupFile(t, filepath.Join(dbPath, backup2+".sql.gz"), "SQL2")

		metadata1 := createTestMetadata(backup1, "testdb", backup1+".sql.gz", "gzip")
		metadata2 := createTestMetadata(backup2, "testdb", backup2+".sql.gz", "gzip")
		// Set different creation times
		metadata1.CreatedAt = time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
		metadata2.CreatedAt = time.Date(2025, 1, 15, 14, 30, 22, 0, time.UTC)

		saveMetadata(t, filepath.Join(dbPath, backup1+".meta.json"), metadata1)
		saveMetadata(t, filepath.Join(dbPath, backup2+".meta.json"), metadata2)

		entry, err := service.loadBackupMetadata("testdb", "")
		assert.NoError(t, err)
		assert.NotNil(t, entry)
		assert.Equal(t, backup2, entry.BackupID) // Should get latest
	})

	t.Run("specific backup", func(t *testing.T) {
		mockClient := mysql.NewMockClient()
		config := &mysql.Config{Host: "localhost", User: "root"}
		tmpDir := t.TempDir()
		localStorage, _ := storage.NewLocalStorage(tmpDir)

		service := NewRestoreService(mockClient, localStorage, config)

		dbPath := filepath.Join(tmpDir, "testdb")
		require.NoError(t, os.MkdirAll(dbPath, 0755))

		backupID := "2025-01-15-143022"
		createTestBackupFile(t, filepath.Join(dbPath, backupID+".sql.gz"), "SQL")
		metadata := createTestMetadata(backupID, "testdb", backupID+".sql.gz", "gzip")
		saveMetadata(t, filepath.Join(dbPath, backupID+".meta.json"), metadata)

		entry, err := service.loadBackupMetadata("testdb", backupID)
		assert.NoError(t, err)
		assert.NotNil(t, entry)
		assert.Equal(t, backupID, entry.BackupID)
	})

	t.Run("backup not found", func(t *testing.T) {
		mockClient := mysql.NewMockClient()
		config := &mysql.Config{Host: "localhost", User: "root"}
		tmpDir := t.TempDir()
		localStorage, _ := storage.NewLocalStorage(tmpDir)

		service := NewRestoreService(mockClient, localStorage, config)

		entry, err := service.loadBackupMetadata("testdb", "nonexistent")
		assert.Error(t, err)
		assert.Nil(t, entry)
		assert.True(t, IsBackupNotFoundError(err))
	})

	t.Run("no backups", func(t *testing.T) {
		mockClient := mysql.NewMockClient()
		config := &mysql.Config{Host: "localhost", User: "root"}
		tmpDir := t.TempDir()
		localStorage, _ := storage.NewLocalStorage(tmpDir)

		service := NewRestoreService(mockClient, localStorage, config)

		entry, err := service.loadBackupMetadata("testdb", "")
		assert.Error(t, err)
		assert.Nil(t, entry)
		assert.True(t, IsBackupNotFoundError(err))
	})
}

func TestGetStorageNameForRestore(t *testing.T) {
	t.Run("with config name", func(t *testing.T) {
		options := &RestoreOptions{
			Database:   "actualdb",
			ConfigName: "configdb",
		}
		name := getStorageNameForRestore(options)
		assert.Equal(t, "configdb", name)
	})

	t.Run("without config name", func(t *testing.T) {
		options := &RestoreOptions{
			Database: "actualdb",
		}
		name := getStorageNameForRestore(options)
		assert.Equal(t, "actualdb", name)
	})
}

// Helper functions for tests

func createTestBackupFile(t *testing.T, filePath, sqlContent string) {
	file, err := os.Create(filePath)
	require.NoError(t, err)
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	_, err = gzWriter.Write([]byte(sqlContent))
	require.NoError(t, err)
	require.NoError(t, gzWriter.Close())
}

func createTestMetadata(backupID, database, file, compression string) BackupMetadata {
	return BackupMetadata{
		Version:  MetadataVersion,
		BackupID: backupID,
		Database: DatabaseInfo{
			Type:     "mysql",
			Host:     "localhost",
			Port:     3306,
			Database: database,
			Version:  "8.0.35",
		},
		CreatedAt:       time.Now(),
		CompletedAt:     time.Now(),
		DurationSeconds: 10,
		Status:          StatusCompleted,
		Backup: BackupFileInfo{
			File:        filepath.Base(file),
			SizeBytes:   1000,
			SizeHuman:   "1.0 KB",
			Compression: compression,
			Checksum:    "", // Will be calculated if needed
		},
		Options: BackupOptionsInfo{
			SchemaOnly: false,
		},
		Tool: ToolInfo{
			Name:    ToolName,
			Version: ToolVersion,
		},
	}
}

func saveMetadata(t *testing.T, path string, metadata BackupMetadata) {
	data, err := json.MarshalIndent(metadata, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0644))
}
