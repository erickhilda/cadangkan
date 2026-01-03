package backup

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// GenerateBackupID generates a unique backup ID based on current timestamp.
// Format: YYYY-MM-DD-HHMMSS (e.g., "2025-01-02-143022")
func GenerateBackupID() string {
	return time.Now().Format("2006-01-02-150405")
}

// FormatBytes converts bytes to human-readable format.
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp+1])
}

// CalculateChecksum calculates SHA-256 checksum of a file.
// Returns checksum in format "sha256:hexstring"
func CalculateChecksum(filepath string) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to open file for checksum: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate checksum: %w", err)
	}

	return fmt.Sprintf("sha256:%x", hash.Sum(nil)), nil
}

// CheckDiskSpace checks if there is enough free disk space at the given path.
// Returns available bytes and an error if the check fails.
func CheckDiskSpace(path string) (uint64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, fmt.Errorf("failed to check disk space: %w", err)
	}

	// Available space = Available blocks * Block size
	availableBytes := stat.Bavail * uint64(stat.Bsize)
	return availableBytes, nil
}

// HasEnoughDiskSpace checks if there's enough free space for the estimated backup size.
// It requires at least the estimated size plus 20% buffer.
func HasEnoughDiskSpace(path string, estimatedSize int64) (bool, error) {
	available, err := CheckDiskSpace(path)
	if err != nil {
		return false, err
	}

	// Add 20% buffer to estimated size
	requiredSize := uint64(float64(estimatedSize) * 1.2)

	return available >= requiredSize, nil
}

// EnsureDir ensures that a directory exists, creating it if necessary.
func EnsureDir(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", path, err)
	}
	return nil
}

// PathExists checks if a path exists.
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// GetFileSize returns the size of a file in bytes.
func GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, fmt.Errorf("failed to get file size: %w", err)
	}
	return info.Size(), nil
}

// CleanupFile removes a file if it exists, ignoring errors if file doesn't exist.
func CleanupFile(path string) error {
	if !PathExists(path) {
		return nil
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to cleanup file %s: %w", path, err)
	}
	return nil
}

// GetBackupDir returns the backup directory for a given database.
// If basePath is empty, uses ~/.cadangkan/backups/{database}/
// Otherwise uses {basePath}/{database}/
func GetBackupDir(database string, basePath string) (string, error) {
	if database == "" {
		return "", ErrDatabaseRequired
	}

	var backupDir string
	if basePath != "" {
		backupDir = filepath.Join(basePath, database)
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		backupDir = filepath.Join(homeDir, ".cadangkan", "backups", database)
	}

	return backupDir, nil
}

// GetBackupFilePath returns the full path for a backup file.
func GetBackupFilePath(backupDir, backupID, compression string) string {
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

// GetMetadataFilePath returns the full path for a metadata file.
func GetMetadataFilePath(backupDir, backupID string) string {
	return filepath.Join(backupDir, backupID+".meta.json")
}

// SanitizeDatabaseName sanitizes a database name for use in file paths.
// Replaces special characters with underscores.
func SanitizeDatabaseName(name string) string {
	// Simple sanitization: replace non-alphanumeric characters with underscore
	result := make([]rune, 0, len(name))
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			result = append(result, r)
		} else {
			result = append(result, '_')
		}
	}
	return string(result)
}

// EstimateBackupSize estimates the backup size based on database size.
// This is a rough estimate: compressed size is typically 30-40% of original
func EstimateBackupSize(databaseSize int64, compression string) int64 {
	if compression == CompressionNone {
		return databaseSize
	}

	// Estimate 35% of original size for compression
	return int64(float64(databaseSize) * 0.35)
}

// CalculateChecksumFromReader calculates SHA-256 checksum from a reader.
// Returns checksum in format "sha256:hexstring"
func CalculateChecksumFromReader(reader io.Reader) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return "", fmt.Errorf("failed to calculate checksum: %w", err)
	}

	return fmt.Sprintf("sha256:%x", hash.Sum(nil)), nil
}

// GetBackupIDFromTime generates a backup ID from a specific time.
// Useful for testing or when you need a specific timestamp.
func GetBackupIDFromTime(t time.Time) string {
	return t.Format("2006-01-02-150405")
}

// ParseBackupID parses a backup ID into a time.Time.
// Returns error if the ID is not in the expected format.
func ParseBackupID(backupID string) (time.Time, error) {
	t, err := time.Parse("2006-01-02-150405", backupID)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid backup ID format: %w", err)
	}
	return t, nil
}
