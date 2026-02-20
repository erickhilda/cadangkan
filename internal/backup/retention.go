package backup

import (
	"fmt"
	"sort"
	"time"

	"github.com/erickhilda/cadangkan/internal/config"
	"github.com/erickhilda/cadangkan/internal/storage"
)

// RetentionService manages backup retention policies.
type RetentionService struct {
	storage *storage.LocalStorage
}

// NewRetentionService creates a new retention service.
func NewRetentionService(stor *storage.LocalStorage) *RetentionService {
	return &RetentionService{
		storage: stor,
	}
}

// BackupCategory represents backup categorization.
type BackupCategory int

const (
	CategoryDaily BackupCategory = iota
	CategoryWeekly
	CategoryMonthly
	CategoryKeep // Always keep
	CategoryDelete
)

// CategorizedBackup represents a backup with its category.
type CategorizedBackup struct {
	Backup   storage.BackupListEntry
	Category BackupCategory
}

// CleanupResult contains the result of a cleanup operation.
type CleanupResult struct {
	ToDelete      []storage.BackupListEntry
	ToKeep        []CategorizedBackup
	SpaceReclaimed int64
	DryRun        bool
}

// ApplyRetentionPolicy applies retention policy and returns backups to delete.
func (s *RetentionService) ApplyRetentionPolicy(databaseName string, policy *config.RetentionPolicy, dryRun bool) (*CleanupResult, error) {
	// Get all backups for this database
	backups, err := s.storage.ListBackups(databaseName)
	if err != nil {
		return nil, fmt.Errorf("failed to list backups: %w", err)
	}

	// If keep_all is true, don't delete anything
	if policy.KeepAll {
		result := &CleanupResult{
			ToDelete:       []storage.BackupListEntry{},
			ToKeep:         []CategorizedBackup{},
			SpaceReclaimed: 0,
			DryRun:         dryRun,
		}
		for _, backup := range backups {
			result.ToKeep = append(result.ToKeep, CategorizedBackup{
				Backup:   backup,
				Category: CategoryKeep,
			})
		}
		return result, nil
	}

	// Sort backups by date (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt.After(backups[j].CreatedAt)
	})

	// Categorize backups
	categorized := s.categorizeBackups(backups, policy)

	// Separate backups to keep and delete
	result := &CleanupResult{
		ToDelete:       []storage.BackupListEntry{},
		ToKeep:         []CategorizedBackup{},
		SpaceReclaimed: 0,
		DryRun:         dryRun,
	}

	for _, cb := range categorized {
		if cb.Category == CategoryDelete {
			result.ToDelete = append(result.ToDelete, cb.Backup)
			result.SpaceReclaimed += cb.Backup.SizeBytes
		} else {
			result.ToKeep = append(result.ToKeep, cb)
		}
	}

	// If not dry-run, delete the backups
	if !dryRun {
		for _, backup := range result.ToDelete {
			if err := s.storage.DeleteBackup(databaseName, backup.BackupID); err != nil {
				return nil, fmt.Errorf("failed to delete backup %s: %w", backup.BackupID, err)
			}
		}
	}

	return result, nil
}

// categorizeBackups categorizes backups based on retention policy.
func (s *RetentionService) categorizeBackups(backups []storage.BackupListEntry, policy *config.RetentionPolicy) []CategorizedBackup {
	result := make([]CategorizedBackup, 0, len(backups))

	// Track what we've seen
	dailyCount := 0
	weeklyCount := 0
	monthlyCount := 0

	// Track dates we've seen
	seenDays := make(map[string]bool)
	seenWeeks := make(map[string]bool)
	seenMonths := make(map[string]bool)

	for _, backup := range backups {
		t := backup.CreatedAt
		dayKey := t.Format("2006-01-02")
		weekKey := getWeekKey(t)
		monthKey := t.Format("2006-01")

		category := CategoryDelete // Default to delete

		// Check if this should be kept as monthly
		if monthlyCount < policy.Monthly && !seenMonths[monthKey] {
			if isFirstOfMonth(t) || !seenMonths[monthKey] {
				category = CategoryMonthly
				seenMonths[monthKey] = true
				monthlyCount++
			}
		}

		// Check if this should be kept as weekly (if not already monthly)
		if category == CategoryDelete && weeklyCount < policy.Weekly && !seenWeeks[weekKey] {
			if isSunday(t) || !seenWeeks[weekKey] {
				category = CategoryWeekly
				seenWeeks[weekKey] = true
				weeklyCount++
			}
		}

		// Check if this should be kept as daily (if not already weekly or monthly)
		if category == CategoryDelete && dailyCount < policy.Daily && !seenDays[dayKey] {
			category = CategoryDaily
			seenDays[dayKey] = true
			dailyCount++
		}

		result = append(result, CategorizedBackup{
			Backup:   backup,
			Category: category,
		})
	}

	return result
}

// getWeekKey returns the week identifier (ISO week format)
func getWeekKey(t time.Time) string {
	year, week := t.ISOWeek()
	return fmt.Sprintf("%d-W%02d", year, week)
}

// isFirstOfMonth checks if the date is the first day of the month
func isFirstOfMonth(t time.Time) bool {
	return t.Day() == 1
}

// isSunday checks if the date is a Sunday
func isSunday(t time.Time) bool {
	return t.Weekday() == time.Sunday
}

// FormatCategory returns a human-readable category name
func FormatCategory(category BackupCategory) string {
	switch category {
	case CategoryDaily:
		return "daily"
	case CategoryWeekly:
		return "weekly"
	case CategoryMonthly:
		return "monthly"
	case CategoryKeep:
		return "keep"
	case CategoryDelete:
		return "delete"
	default:
		return "unknown"
	}
}
