package config

// Config represents the main configuration file.
type Config struct {
	Version   string                     `yaml:"version"`
	Databases map[string]*DatabaseConfig `yaml:"databases"`
}

// DatabaseConfig represents a database configuration.
type DatabaseConfig struct {
	Name              string `yaml:"-"` // Not stored in YAML, derived from map key
	Type              string `yaml:"type"`
	Host              string `yaml:"host"`
	Port              int    `yaml:"port"`
	Database          string `yaml:"database"`
	User              string `yaml:"user"`
	PasswordEncrypted string `yaml:"password_encrypted,omitempty"`
}

// NewConfig creates a new Config with default values.
func NewConfig() *Config {
	return &Config{
		Version:   "1.0",
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
