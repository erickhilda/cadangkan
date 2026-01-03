package backup

import (
	"testing"
	"time"

	"github.com/erickhilda/cadangkan/pkg/database/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateSimple(t *testing.T) {
	backupID := "2025-01-02-143022"
	database := "mydb"
	host := "localhost"
	port := 3306
	filePath := "/backups/mydb/2025-01-02-143022.sql.gz"
	sizeBytes := int64(1024 * 1024 * 100) // 100MB
	duration := 2 * time.Minute
	checksum := "sha256:abc123"
	compression := CompressionGzip

	metadata := GenerateSimple(
		backupID, database, host, port, filePath,
		sizeBytes, duration, checksum, compression, StatusCompleted,
	)

	assert.Equal(t, MetadataVersion, metadata.Version)
	assert.Equal(t, backupID, metadata.BackupID)
	assert.Equal(t, database, metadata.Database.Database)
	assert.Equal(t, host, metadata.Database.Host)
	assert.Equal(t, port, metadata.Database.Port)
	assert.Equal(t, sizeBytes, metadata.Backup.SizeBytes)
	assert.Equal(t, checksum, metadata.Backup.Checksum)
	assert.Equal(t, compression, metadata.Backup.Compression)
	assert.Equal(t, StatusCompleted, metadata.Status)
	assert.Equal(t, int64(120), metadata.DurationSeconds) // 2 minutes
}

func TestCreateInitialMetadata(t *testing.T) {
	backupID := GenerateBackupID()
	database := "testdb"
	config := mysql.NewConfig().
		WithHost("localhost").
		WithPort(3306).
		WithDatabase(database)

	options := DefaultOptions()
	options.Database = database
	options.SchemaOnly = true

	metadata := CreateInitialMetadata(backupID, database, config, options)

	assert.Equal(t, MetadataVersion, metadata.Version)
	assert.Equal(t, backupID, metadata.BackupID)
	assert.Equal(t, database, metadata.Database.Database)
	assert.Equal(t, "localhost", metadata.Database.Host)
	assert.Equal(t, 3306, metadata.Database.Port)
	assert.Equal(t, StatusRunning, metadata.Status)
	assert.True(t, metadata.Options.SchemaOnly)
	assert.Equal(t, ToolName, metadata.Tool.Name)
}

func TestUpdateMetadata(t *testing.T) {
	metadata := &BackupMetadata{
		BackupID: "test-backup",
		Status:   StatusRunning,
	}

	result := &BackupResult{
		BackupID:    "test-backup",
		SizeBytes:   1024000,
		Checksum:    "sha256:test",
		Duration:    30 * time.Second,
		Status:      StatusCompleted,
		StartedAt:   time.Now().Add(-30 * time.Second),
		CompletedAt: time.Now(),
	}

	UpdateMetadata(metadata, result)

	assert.Equal(t, StatusCompleted, metadata.Status)
	assert.Equal(t, int64(1024000), metadata.Backup.SizeBytes)
	assert.Equal(t, "sha256:test", metadata.Backup.Checksum)
	assert.Equal(t, int64(30), metadata.DurationSeconds)
}

func TestMarkFailed(t *testing.T) {
	metadata := &BackupMetadata{
		BackupID:  "test-backup",
		Status:    StatusRunning,
		CreatedAt: time.Now().Add(-1 * time.Minute),
	}

	err := &DumpError{
		Database: "testdb",
		Command:  "mysqldump",
		Stderr:   "dump failed",
		ExitCode: 1,
	}

	MarkFailed(metadata, err)

	assert.Equal(t, StatusFailed, metadata.Status)
	assert.NotZero(t, metadata.CompletedAt)
	assert.Greater(t, metadata.DurationSeconds, int64(0))
	assert.Contains(t, metadata.Error, "dump failed")
}

func TestMarkCompleted(t *testing.T) {
	metadata := &BackupMetadata{
		BackupID:  "test-backup",
		Status:    StatusRunning,
		CreatedAt: time.Now().Add(-2 * time.Minute),
	}

	MarkCompleted(metadata)

	assert.Equal(t, StatusCompleted, metadata.Status)
	assert.NotZero(t, metadata.CompletedAt)
	assert.GreaterOrEqual(t, metadata.DurationSeconds, int64(120)) // At least 2 minutes
}

func TestValidateMetadata(t *testing.T) {
	t.Run("valid metadata", func(t *testing.T) {
		metadata := &BackupMetadata{
			BackupID: "test-backup",
			Database: DatabaseInfo{
				Database: "testdb",
			},
			Status: StatusCompleted,
		}

		err := ValidateMetadata(metadata)
		assert.NoError(t, err)
	})

	t.Run("missing backup ID", func(t *testing.T) {
		metadata := &BackupMetadata{
			Database: DatabaseInfo{
				Database: "testdb",
			},
			Status: StatusCompleted,
		}

		err := ValidateMetadata(metadata)
		assert.Error(t, err)
		assert.True(t, IsMetadataError(err))
	})

	t.Run("missing database", func(t *testing.T) {
		metadata := &BackupMetadata{
			BackupID: "test-backup",
			Status:   StatusCompleted,
		}

		err := ValidateMetadata(metadata)
		assert.Error(t, err)
		assert.True(t, IsMetadataError(err))
	})

	t.Run("missing status", func(t *testing.T) {
		metadata := &BackupMetadata{
			BackupID: "test-backup",
			Database: DatabaseInfo{
				Database: "testdb",
			},
		}

		err := ValidateMetadata(metadata)
		assert.Error(t, err)
		assert.True(t, IsMetadataError(err))
	})
}

func TestGetBackupAge(t *testing.T) {
	metadata := &BackupMetadata{
		CreatedAt: time.Now().Add(-2 * time.Hour),
	}

	age := GetBackupAge(metadata)
	assert.GreaterOrEqual(t, age, 2*time.Hour)
	assert.Less(t, age, 3*time.Hour)
}

func TestIsBackupOlderThan(t *testing.T) {
	metadata := &BackupMetadata{
		CreatedAt: time.Now().Add(-2 * time.Hour),
	}

	// Should be older than 1 hour
	assert.True(t, IsBackupOlderThan(metadata, 1*time.Hour))

	// Should not be older than 3 hours
	assert.False(t, IsBackupOlderThan(metadata, 3*time.Hour))
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{30 * time.Second, "30s"},
		{90 * time.Second, "1m 30s"},
		{5 * time.Minute, "5m 0s"},
		{65 * time.Minute, "1h 5m"},
		{125 * time.Minute, "2h 5m"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := FormatDuration(tt.duration)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetMySQLDumpVersion(t *testing.T) {
	version := GetMySQLDumpVersion()
	// Should return either a version string or "unknown"
	assert.NotEmpty(t, version)
}

func TestMetadataGenerator(t *testing.T) {
	// Create a mock client
	mockClient := mysql.NewMockClient()
	mockClient.SetConnected(true)
	mockClient.Version = "8.0.35"

	generator := NewMetadataGenerator(mockClient)
	assert.NotNil(t, generator)

	config := mysql.NewConfig().
		WithHost("localhost").
		WithPort(3306).
		WithDatabase("testdb")

	options := DefaultOptions()
	options.Database = "testdb"

	result := &BackupResult{
		BackupID:    "2025-01-02-143022",
		FilePath:    "/backups/testdb/2025-01-02-143022.sql.gz",
		SizeBytes:   1024000,
		Checksum:    "sha256:abc123",
		Duration:    2 * time.Minute,
		Status:      StatusCompleted,
		StartedAt:   time.Now().Add(-2 * time.Minute),
		CompletedAt: time.Now(),
	}

	metadata, err := generator.Generate(
		result.BackupID,
		config,
		result,
		options,
		"mysqldump 8.0.35",
	)

	require.NoError(t, err)
	assert.Equal(t, result.BackupID, metadata.BackupID)
	assert.Equal(t, "testdb", metadata.Database.Database)
	assert.Equal(t, "8.0.35", metadata.Database.Version)
	assert.Equal(t, result.SizeBytes, metadata.Backup.SizeBytes)
	assert.Equal(t, result.Checksum, metadata.Backup.Checksum)
	assert.Equal(t, "mysqldump 8.0.35", metadata.Tool.MySQLDumpVersion)
}
