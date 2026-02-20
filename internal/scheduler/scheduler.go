package scheduler

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/erickhilda/cadangkan/internal/backup"
	"github.com/erickhilda/cadangkan/internal/config"
	"github.com/erickhilda/cadangkan/internal/storage"
	"github.com/erickhilda/cadangkan/pkg/database/mysql"
	"github.com/robfig/cron/v3"
)

// Scheduler manages scheduled backup jobs.
type Scheduler struct {
	cron      *cron.Cron
	jobs      map[string]cron.EntryID // database name -> cron entry ID
	config    *config.Config
	storage   *storage.LocalStorage
	mu        sync.RWMutex
	logger    *log.Logger
	verbose   bool
}

// New creates a new scheduler instance.
func New(cfg *config.Config, stor *storage.LocalStorage) *Scheduler {
	return &Scheduler{
		cron:    cron.New(cron.WithLocation(time.Local)),
		jobs:    make(map[string]cron.EntryID),
		config:  cfg,
		storage: stor,
		logger:  log.New(log.Writer(), "[scheduler] ", log.LstdFlags),
	}
}

// SetVerbose enables or disables verbose logging.
func (s *Scheduler) SetVerbose(verbose bool) {
	s.verbose = verbose
}

// Start starts the scheduler.
func (s *Scheduler) Start() {
	s.cron.Start()
	if s.verbose {
		s.logger.Println("Scheduler started")
	}
}

// Stop stops the scheduler.
func (s *Scheduler) Stop() {
	s.cron.Stop()
	if s.verbose {
		s.logger.Println("Scheduler stopped")
	}
}

// LoadSchedules loads all schedules from config and registers them.
func (s *Scheduler) LoadSchedules() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear existing jobs
	for dbName, entryID := range s.jobs {
		s.cron.Remove(entryID)
		delete(s.jobs, dbName)
	}

	// Register all enabled schedules
	for dbName, dbConfig := range s.config.Databases {
		if dbConfig.Schedule != nil && dbConfig.Schedule.Enabled {
			if err := s.addSchedule(dbName, dbConfig); err != nil {
				s.logger.Printf("Failed to add schedule for %s: %v", dbName, err)
				continue
			}
		}
	}

	return nil
}

// addSchedule adds a schedule for a database (internal, assumes lock is held).
func (s *Scheduler) addSchedule(dbName string, dbConfig *config.DatabaseConfig) error {
	if dbConfig.Schedule == nil || dbConfig.Schedule.Cron == "" {
		return fmt.Errorf("no schedule configured")
	}

	// Validate cron expression
	_, err := cron.ParseStandard(dbConfig.Schedule.Cron)
	if err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}

	// Create backup job
	job := s.createBackupJob(dbName, dbConfig)

	// Add to cron
	entryID, err := s.cron.AddFunc(dbConfig.Schedule.Cron, job)
	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	s.jobs[dbName] = entryID

	if s.verbose {
		s.logger.Printf("Added schedule for %s: %s", dbName, dbConfig.Schedule.Cron)
	}

	return nil
}

// createBackupJob creates a backup job function for a database.
func (s *Scheduler) createBackupJob(dbName string, dbConfig *config.DatabaseConfig) func() {
	return func() {
		s.logger.Printf("Running scheduled backup for %s", dbName)

		// Decrypt password
		password, err := config.DecryptPassword(dbConfig.PasswordEncrypted)
		if err != nil {
			s.logger.Printf("Failed to decrypt password for %s: %v", dbName, err)
			return
		}

		// Create MySQL client
		mysqlConfig := &mysql.Config{
			Host:     dbConfig.Host,
			Port:     dbConfig.Port,
			User:     dbConfig.User,
			Password: password,
			Database: dbConfig.Database,
			Timeout:  10 * time.Second,
		}

		client, err := mysql.NewClient(mysqlConfig)
		if err != nil {
			s.logger.Printf("Failed to create client for %s: %v", dbName, err)
			return
		}

		if err := client.Connect(); err != nil {
			s.logger.Printf("Failed to connect to %s: %v", dbName, err)
			return
		}
		defer client.Close()

		// Create backup service
		backupService := backup.NewService(client, s.storage, mysqlConfig)
		if s.verbose {
			backupService.SetVerbose(true)
		}

		// Backup options
		backupOptions := &backup.BackupOptions{
			Database:      dbConfig.Database,
			ConfigName:    dbName,
			Compression:   backup.CompressionGzip,
			Tables:        nil,
			ExcludeTables: nil,
			SchemaOnly:    false,
		}

		// Execute backup
		result, err := backupService.Backup(backupOptions)
		if err != nil {
			s.logger.Printf("Backup failed for %s: %v", dbName, err)
			return
		}

		s.logger.Printf("Backup completed for %s: %s (%s)", dbName, result.BackupID, backup.FormatBytes(result.SizeBytes))

		// Apply retention policy if configured
		if dbConfig.Retention != nil && !dbConfig.Retention.KeepAll {
			retentionService := backup.NewRetentionService(s.storage)
			cleanupResult, err := retentionService.ApplyRetentionPolicy(dbName, dbConfig.Retention, false)
			if err != nil {
				s.logger.Printf("Retention cleanup failed for %s: %v", dbName, err)
			} else if len(cleanupResult.ToDelete) > 0 {
				s.logger.Printf("Cleaned up %d old backup(s) for %s", len(cleanupResult.ToDelete), dbName)
			}
		}
	}
}

// GetNextRun returns the next run time for a database schedule.
func (s *Scheduler) GetNextRun(dbName string) (time.Time, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entryID, exists := s.jobs[dbName]
	if !exists {
		return time.Time{}, fmt.Errorf("no active schedule for %s", dbName)
	}

	entry := s.cron.Entry(entryID)
	if entry.ID == 0 {
		return time.Time{}, fmt.Errorf("schedule not found")
	}

	return entry.Next, nil
}

// ListSchedules returns information about all active schedules.
func (s *Scheduler) ListSchedules() []ScheduleInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var schedules []ScheduleInfo
	for dbName, entryID := range s.jobs {
		entry := s.cron.Entry(entryID)
		if entry.ID == 0 {
			continue
		}

		dbConfig := s.config.Databases[dbName]
		schedules = append(schedules, ScheduleInfo{
			Database: dbName,
			Cron:     dbConfig.Schedule.Cron,
			Enabled:  dbConfig.Schedule.Enabled,
			NextRun:  entry.Next,
			PrevRun:  entry.Prev,
		})
	}

	return schedules
}

// ScheduleInfo contains information about a scheduled backup.
type ScheduleInfo struct {
	Database string
	Cron     string
	Enabled  bool
	NextRun  time.Time
	PrevRun  time.Time
}
