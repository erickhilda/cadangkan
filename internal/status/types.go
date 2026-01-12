package status

import (
	"time"

	"github.com/erickhilda/cadangkan/internal/backup"
)

// OverallStatus represents the overall status of all databases.
type OverallStatus struct {
	ServiceStatus    string
	DatabaseCount    int
	ActiveCount      int
	TotalBackups     int
	StorageUsed      int64
	StorageAvailable uint64
	LastBackup       *time.Time
	Databases        []DatabaseStatus
	HealthSummary    []string
}

// DatabaseStatus represents the status of a single database.
type DatabaseStatus struct {
	Name            string
	Type            string
	Status          string // "healthy", "warning", "critical"
	LastBackup      *time.Time
	LastBackupID    string
	NextBackup      string // "Not scheduled" for now
	BackupCount     int
	SuccessfulCount int
	FailedCount     int
	StorageUsed     int64
	RecentBackups   []backup.BackupListEntry
}

// HealthScore represents the health score for a database.
type HealthScore struct {
	TotalScore       float64
	SuccessRate      float64
	RecencyScore     float64
	ConsistencyScore float64
	Recommendations  []string
	RecentBackups    []backup.BackupListEntry
}

// StorageUsage represents storage usage information.
type StorageUsage struct {
	TotalUsed      int64
	TotalAvailable uint64
	ByDatabase     []DatabaseStorage
	LargestBackups []backup.BackupListEntry
}

// DatabaseStorage represents storage usage for a single database.
type DatabaseStorage struct {
	Database    string
	BackupCount int
	SizeBytes   int64
	Percentage  float64
}
