package backup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateBackupID(t *testing.T) {
	id := GenerateBackupID()

	// Should be in format YYYY-MM-DD-HHMMSS (17 chars)
	assert.Len(t, id, 17) // 2026-01-02-143022
	assert.Contains(t, id, "-")

	// Should parse back to a time without error
	parsed, err := ParseBackupID(id)
	assert.NoError(t, err)
	assert.NotZero(t, parsed)
	
	// Year should be reasonable (this century)
	assert.GreaterOrEqual(t, parsed.Year(), 2020)
	assert.LessOrEqual(t, parsed.Year(), 2100)
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
		{1024 * 1024 * 1024 * 1024, "1.0 TB"},
		{245760000, "234.4 MB"}, // Actual calculated value
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := FormatBytes(tt.bytes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateChecksum(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")

	content := []byte("Hello, World!")
	err := os.WriteFile(filePath, content, 0644)
	require.NoError(t, err)

	// Calculate checksum
	checksum, err := CalculateChecksum(filePath)
	require.NoError(t, err)

	// Should start with "sha256:"
	assert.True(t, strings.HasPrefix(checksum, "sha256:"))
	assert.Len(t, checksum, 71) // "sha256:" + 64 hex chars

	// Same file should have same checksum
	checksum2, err := CalculateChecksum(filePath)
	require.NoError(t, err)
	assert.Equal(t, checksum, checksum2)
}

func TestCalculateChecksumNonExistent(t *testing.T) {
	_, err := CalculateChecksum("/nonexistent/file.txt")
	assert.Error(t, err)
}

func TestCheckDiskSpace(t *testing.T) {
	tmpDir := t.TempDir()

	available, err := CheckDiskSpace(tmpDir)
	require.NoError(t, err)
	assert.Greater(t, available, uint64(0))
}

func TestHasEnoughDiskSpace(t *testing.T) {
	tmpDir := t.TempDir()

	// Should have enough space for 1KB
	hasSpace, err := HasEnoughDiskSpace(tmpDir, 1024)
	require.NoError(t, err)
	assert.True(t, hasSpace)

	// Should not have enough space for exabytes
	hasSpace, err = HasEnoughDiskSpace(tmpDir, 1024*1024*1024*1024*1024)
	require.NoError(t, err)
	assert.False(t, hasSpace)
}

func TestEnsureDir(t *testing.T) {
	tmpDir := t.TempDir()
	newDir := filepath.Join(tmpDir, "test", "nested", "dir")

	// Directory should not exist
	assert.False(t, PathExists(newDir))

	// Create it
	err := EnsureDir(newDir)
	require.NoError(t, err)

	// Should now exist
	assert.True(t, PathExists(newDir))

	// Calling again should be fine
	err = EnsureDir(newDir)
	assert.NoError(t, err)
}

func TestPathExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Temp dir should exist
	assert.True(t, PathExists(tmpDir))

	// Non-existent path should not exist
	assert.False(t, PathExists(filepath.Join(tmpDir, "nonexistent")))
}

func TestGetFileSize(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")

	content := []byte("Hello, World!")
	err := os.WriteFile(filePath, content, 0644)
	require.NoError(t, err)

	size, err := GetFileSize(filePath)
	require.NoError(t, err)
	assert.Equal(t, int64(len(content)), size)
}

func TestCleanupFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")

	// Create file
	err := os.WriteFile(filePath, []byte("test"), 0644)
	require.NoError(t, err)
	assert.True(t, PathExists(filePath))

	// Cleanup
	err = CleanupFile(filePath)
	require.NoError(t, err)
	assert.False(t, PathExists(filePath))

	// Cleanup non-existent file should not error
	err = CleanupFile(filePath)
	assert.NoError(t, err)
}

func TestGetBackupDir(t *testing.T) {
	t.Run("with base path", func(t *testing.T) {
		dir, err := GetBackupDir("mydb", "/tmp/backups")
		require.NoError(t, err)
		assert.Equal(t, "/tmp/backups/mydb", dir)
	})

	t.Run("without base path", func(t *testing.T) {
		dir, err := GetBackupDir("mydb", "")
		require.NoError(t, err)
		assert.Contains(t, dir, ".cadangkan/backups/mydb")
	})

	t.Run("empty database", func(t *testing.T) {
		_, err := GetBackupDir("", "/tmp/backups")
		assert.Equal(t, ErrDatabaseRequired, err)
	})
}

func TestGetBackupFilePath(t *testing.T) {
	tests := []struct {
		compression string
		expected    string
	}{
		{CompressionGzip, "2025-01-02-143022.sql.gz"},
		{CompressionNone, "2025-01-02-143022.sql"},
		{CompressionZstd, "2025-01-02-143022.sql.zst"},
		{"invalid", "2025-01-02-143022.sql.gz"}, // Defaults to gzip
	}

	for _, tt := range tests {
		t.Run(tt.compression, func(t *testing.T) {
			path := GetBackupFilePath("/backups", "2025-01-02-143022", tt.compression)
			assert.Equal(t, filepath.Join("/backups", tt.expected), path)
		})
	}
}

func TestGetMetadataFilePath(t *testing.T) {
	path := GetMetadataFilePath("/backups", "2025-01-02-143022")
	assert.Equal(t, "/backups/2025-01-02-143022.meta.json", path)
}

func TestSanitizeDatabaseName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"mydb", "mydb"},
		{"my-db", "my-db"},
		{"my_db", "my_db"},
		{"my.db", "my_db"},
		{"my db", "my_db"},
		{"my@db#123", "my_db_123"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := SanitizeDatabaseName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEstimateBackupSize(t *testing.T) {
	dbSize := int64(1024 * 1024 * 1024) // 1GB

	// With compression, should be ~35% of original
	compressed := EstimateBackupSize(dbSize, CompressionGzip)
	assert.InDelta(t, float64(dbSize)*0.35, float64(compressed), float64(dbSize)*0.1)

	// Without compression, should be same as original
	uncompressed := EstimateBackupSize(dbSize, CompressionNone)
	assert.Equal(t, dbSize, uncompressed)
}

func TestGetBackupIDFromTime(t *testing.T) {
	testTime := time.Date(2025, 1, 2, 14, 30, 22, 0, time.UTC)
	id := GetBackupIDFromTime(testTime)
	assert.Equal(t, "2025-01-02-143022", id)
}

func TestParseBackupID(t *testing.T) {
	t.Run("valid ID", func(t *testing.T) {
		id := "2025-01-02-143022"
		parsed, err := ParseBackupID(id)
		require.NoError(t, err)
		assert.Equal(t, 2025, parsed.Year())
		assert.Equal(t, time.January, parsed.Month())
		assert.Equal(t, 2, parsed.Day())
		assert.Equal(t, 14, parsed.Hour())
		assert.Equal(t, 30, parsed.Minute())
		assert.Equal(t, 22, parsed.Second())
	})

	t.Run("invalid ID", func(t *testing.T) {
		_, err := ParseBackupID("invalid-id")
		assert.Error(t, err)
	})
}

func TestParseBackupIDRoundTrip(t *testing.T) {
	// Use UTC to avoid timezone issues
	original := time.Now().UTC().Truncate(time.Second)
	id := GetBackupIDFromTime(original)
	parsed, err := ParseBackupID(id)
	require.NoError(t, err)
	// ParseBackupID returns time in UTC, so compare timestamps
	assert.Equal(t, original.Unix(), parsed.Unix())
}
