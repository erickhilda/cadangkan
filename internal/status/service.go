package status

import (
	"fmt"
	"sort"
	"time"

	"github.com/erickhilda/cadangkan/internal/backup"
	"github.com/erickhilda/cadangkan/internal/config"
	"github.com/erickhilda/cadangkan/internal/storage"
)

// Service provides status and health monitoring functionality.
type Service struct {
	configManager config.Manager
	storage       *storage.LocalStorage
}

// NewService creates a new status service.
func NewService(configManager config.Manager, stor *storage.LocalStorage) *Service {
	return &Service{
		configManager: configManager,
		storage:       stor,
	}
}

// GetOverallStatus returns the overall status across all databases.
func (s *Service) GetOverallStatus() (*OverallStatus, error) {
	// Load configuration
	cfg, err := s.configManager.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	overall := &OverallStatus{
		ServiceStatus:    "Not running", // Placeholder until service is implemented
		DatabaseCount:    len(cfg.Databases),
		ActiveCount:      0,
		TotalBackups:     0,
		StorageUsed:      0,
		StorageAvailable: 0,
		LastBackup:       nil,
		Databases:        []DatabaseStatus{},
		HealthSummary:    []string{},
	}

	// Get available disk space
	available, err := s.storage.CheckDiskSpace()
	if err == nil {
		overall.StorageAvailable = available
	}

	// Process each database
	var latestBackupTime *time.Time
	dbNames := make([]string, 0, len(cfg.Databases))
	for name := range cfg.Databases {
		dbNames = append(dbNames, name)
	}
	sort.Strings(dbNames)

	for _, dbName := range dbNames {
		dbConfig := cfg.Databases[dbName]
		dbStatus, err := s.GetDatabaseStatus(dbName)
		if err != nil {
			// Skip databases with errors but continue processing others
			continue
		}

		// Set type from config
		dbStatus.Type = dbConfig.Type

		overall.Databases = append(overall.Databases, *dbStatus)
		overall.TotalBackups += dbStatus.BackupCount
		overall.StorageUsed += dbStatus.StorageUsed

		if dbStatus.BackupCount > 0 {
			overall.ActiveCount++
		}

		// Track latest backup across all databases
		if dbStatus.LastBackup != nil {
			if latestBackupTime == nil || dbStatus.LastBackup.After(*latestBackupTime) {
				latestBackupTime = dbStatus.LastBackup
			}
		}
	}

	overall.LastBackup = latestBackupTime

	// Generate health summary
	overall.HealthSummary = s.generateHealthSummary(overall.Databases)

	return overall, nil
}

// GetDatabaseStatus returns detailed status for a specific database.
func (s *Service) GetDatabaseStatus(dbName string) (*DatabaseStatus, error) {
	// Load configuration
	cfg, err := s.configManager.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	dbConfig, exists := cfg.Databases[dbName]
	if !exists {
		return nil, fmt.Errorf("database '%s' not found", dbName)
	}

	status := &DatabaseStatus{
		Name:          dbName,
		Type:          dbConfig.Type,
		NextBackup:    "Not scheduled", // Placeholder until scheduling is implemented
		RecentBackups: []backup.BackupListEntry{},
	}

	// Get all backups for this database
	backups, err := s.storage.ListBackups(dbName)
	if err != nil {
		// If no backups exist, return empty status
		status.Status = "critical"
		return status, nil
	}

	status.BackupCount = len(backups)

	// Count successful vs failed backups
	successfulCount := 0
	failedCount := 0
	var totalSize int64

	for _, b := range backups {
		if b.Status == backup.StatusCompleted || b.Status == "" {
			successfulCount++
		} else if b.Status == backup.StatusFailed {
			failedCount++
		}
		totalSize += b.SizeBytes
	}

	status.SuccessfulCount = successfulCount
	status.FailedCount = failedCount
	status.StorageUsed = totalSize

	// Get latest backup
	if len(backups) > 0 {
		latest := backups[0] // Already sorted newest first
		status.LastBackup = &latest.CreatedAt
		status.LastBackupID = latest.BackupID

		// Get recent backups (last 5)
		maxRecent := 5
		if len(backups) < maxRecent {
			maxRecent = len(backups)
		}
		status.RecentBackups = convertBackupListEntries(backups[:maxRecent])
	}

	// Calculate health score to determine status
	backupEntries := convertBackupListEntries(backups)
	healthScore := CalculateHealthScore(backupEntries)
	status.Status = GetHealthStatus(healthScore.TotalScore)

	return status, nil
}

// GetStorageUsage returns storage usage information.
func (s *Service) GetStorageUsage() (*StorageUsage, error) {
	// Load configuration
	cfg, err := s.configManager.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	usage := &StorageUsage{
		ByDatabase:     []DatabaseStorage{},
		LargestBackups: []backup.BackupListEntry{},
	}

	// Get available disk space
	available, err := s.storage.CheckDiskSpace()
	if err == nil {
		usage.TotalAvailable = available
	}

	// Collect all backups across all databases
	var allBackups []backup.BackupListEntry
	dbNames := make([]string, 0, len(cfg.Databases))
	for name := range cfg.Databases {
		dbNames = append(dbNames, name)
	}
	sort.Strings(dbNames)

	for _, dbName := range dbNames {
		backups, err := s.storage.ListBackups(dbName)
		if err != nil {
			continue
		}

		var dbTotalSize int64
		for _, b := range backups {
			dbTotalSize += b.SizeBytes
			backupEntry := convertBackupListEntry(b)
			allBackups = append(allBackups, backupEntry)
		}

		usage.ByDatabase = append(usage.ByDatabase, DatabaseStorage{
			Database:    dbName,
			BackupCount: len(backups),
			SizeBytes:   dbTotalSize,
		})
	}

	// Calculate total used
	for _, dbStorage := range usage.ByDatabase {
		usage.TotalUsed += dbStorage.SizeBytes
	}

	// Calculate percentages
	if usage.TotalUsed > 0 {
		for i := range usage.ByDatabase {
			usage.ByDatabase[i].Percentage = (float64(usage.ByDatabase[i].SizeBytes) / float64(usage.TotalUsed)) * 100.0
		}
	}

	// Sort databases by size (largest first)
	sort.Slice(usage.ByDatabase, func(i, j int) bool {
		return usage.ByDatabase[i].SizeBytes > usage.ByDatabase[j].SizeBytes
	})

	// Get largest backups (top 5)
	sort.Slice(allBackups, func(i, j int) bool {
		return allBackups[i].SizeBytes > allBackups[j].SizeBytes
	})
	maxLargest := 5
	if len(allBackups) < maxLargest {
		maxLargest = len(allBackups)
	}
	usage.LargestBackups = allBackups[:maxLargest]

	return usage, nil
}

// generateHealthSummary generates a summary of health status across databases.
func (s *Service) generateHealthSummary(databases []DatabaseStatus) []string {
	var summary []string

	healthyCount := 0
	warningCount := 0
	criticalCount := 0
	neverBackedUp := 0

	for _, db := range databases {
		switch db.Status {
		case "healthy":
			healthyCount++
		case "warning":
			warningCount++
		case "critical":
			criticalCount++
		}

		if db.BackupCount == 0 {
			neverBackedUp++
		}
	}

	if healthyCount > 0 && warningCount == 0 && criticalCount == 0 {
		summary = append(summary, fmt.Sprintf("✓ All databases backed up successfully"))
	} else if healthyCount > 0 {
		summary = append(summary, fmt.Sprintf("✓ %d database(s) healthy", healthyCount))
	}

	if warningCount > 0 {
		summary = append(summary, fmt.Sprintf("⚠ %d database(s) need attention", warningCount))
	}

	if criticalCount > 0 {
		summary = append(summary, fmt.Sprintf("✗ %d database(s) have critical issues", criticalCount))
	}

	if neverBackedUp > 0 {
		summary = append(summary, fmt.Sprintf("⚠ %d database(s) never backed up", neverBackedUp))
	}

	// Check for failed backups
	totalFailed := 0
	for _, db := range databases {
		totalFailed += db.FailedCount
	}
	if totalFailed > 0 {
		summary = append(summary, fmt.Sprintf("⚠ %d failed backup(s) in last 30 days", totalFailed))
	}

	if len(summary) == 0 {
		summary = append(summary, "No databases configured")
	}

	return summary
}

// convertBackupListEntry converts storage.BackupListEntry to backup.BackupListEntry
func convertBackupListEntry(entry storage.BackupListEntry) backup.BackupListEntry {
	return backup.BackupListEntry{
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

// convertBackupListEntries converts a slice of storage.BackupListEntry to backup.BackupListEntry
func convertBackupListEntries(entries []storage.BackupListEntry) []backup.BackupListEntry {
	result := make([]backup.BackupListEntry, len(entries))
	for i, entry := range entries {
		result[i] = convertBackupListEntry(entry)
	}
	return result
}
