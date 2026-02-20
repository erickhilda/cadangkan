package config

// Config represents the main configuration file.
type Config struct {
	Version   string                     `yaml:"version"`
	Defaults  *Defaults                  `yaml:"defaults,omitempty"`
	Databases map[string]*DatabaseConfig `yaml:"databases"`
}

// Defaults contains default settings for all databases.
type Defaults struct {
	Retention *RetentionPolicy `yaml:"retention,omitempty"`
}

// RetentionPolicy defines how long to keep backups.
type RetentionPolicy struct {
	Daily   int  `yaml:"daily"`   // Keep last N daily backups
	Weekly  int  `yaml:"weekly"`  // Keep last N weekly backups (Sunday)
	Monthly int  `yaml:"monthly"` // Keep last N monthly backups (1st of month)
	KeepAll bool `yaml:"keep_all,omitempty"` // Never delete backups
}

// ScheduleConfig defines when backups should run.
type ScheduleConfig struct {
	Enabled bool   `yaml:"enabled"`
	Cron    string `yaml:"cron"` // Cron expression (e.g., "0 2 * * *" for daily at 2 AM)
}

// DatabaseConfig represents a database configuration.
type DatabaseConfig struct {
	Name              string           `yaml:"-"` // Not stored in YAML, derived from map key
	Type              string           `yaml:"type"`
	Host              string           `yaml:"host"`
	Port              int              `yaml:"port"`
	Database          string           `yaml:"database"`
	User              string           `yaml:"user"`
	PasswordEncrypted string           `yaml:"password_encrypted,omitempty"`
	Schedule          *ScheduleConfig  `yaml:"schedule,omitempty"`
	Retention         *RetentionPolicy `yaml:"retention,omitempty"` // Override defaults
}

// NewConfig creates a new Config with default values.
func NewConfig() *Config {
	return &Config{
		Version: "1.0",
		Defaults: &Defaults{
			Retention: DefaultRetentionPolicy(),
		},
		Databases: make(map[string]*DatabaseConfig),
	}
}

// NewDatabaseConfig creates a new DatabaseConfig with default values.
func NewDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Type: "mysql",
		Port: 3306,
	}
}

// DefaultRetentionPolicy returns the default retention policy.
func DefaultRetentionPolicy() *RetentionPolicy {
	return &RetentionPolicy{
		Daily:   7,  // Keep 7 days
		Weekly:  4,  // Keep 4 weeks
		Monthly: 12, // Keep 12 months
		KeepAll: false,
	}
}

// GetEffectiveRetention returns the effective retention policy for a database.
// Database-specific policy overrides defaults.
func (c *Config) GetEffectiveRetention(dbName string) *RetentionPolicy {
	db, exists := c.Databases[dbName]
	if !exists {
		return DefaultRetentionPolicy()
	}

	// Database-specific retention overrides defaults
	if db.Retention != nil {
		return db.Retention
	}

	// Use defaults if available
	if c.Defaults != nil && c.Defaults.Retention != nil {
		return c.Defaults.Retention
	}

	// Fallback to default retention
	return DefaultRetentionPolicy()
}
