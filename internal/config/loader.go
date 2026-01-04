package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Manager handles configuration loading and saving.
type Manager struct {
	configPath string
}

// NewManager creates a new config manager.
func NewManager() (*Manager, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	return &Manager{
		configPath: configPath,
	}, nil
}

// Load loads the configuration from disk.
// If the config file doesn't exist, returns an empty config.
func (m *Manager) Load() (*Config, error) {
	// Check if config file exists
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		// Return empty config
		return NewConfig(), nil
	}

	// Read config file
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Ensure databases map is initialized
	if config.Databases == nil {
		config.Databases = make(map[string]*DatabaseConfig)
	}

	// Set database names from map keys
	for name, db := range config.Databases {
		db.Name = name
	}

	return &config, nil
}

// Save saves the configuration to disk.
func (m *Manager) Save(config *Config) error {
	// Validate config before saving
	if err := config.Validate(); err != nil {
		return err
	}

	// Ensure config directory exists
	configDir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file with restricted permissions
	if err := os.WriteFile(m.configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetDatabase retrieves a specific database configuration.
func (m *Manager) GetDatabase(name string) (*DatabaseConfig, error) {
	config, err := m.Load()
	if err != nil {
		return nil, err
	}

	db, exists := config.Databases[name]
	if !exists {
		return nil, &DatabaseNotFoundError{Name: name}
	}

	db.Name = name
	return db, nil
}

// AddDatabase adds or updates a database configuration.
func (m *Manager) AddDatabase(name string, db *DatabaseConfig) error {
	config, err := m.Load()
	if err != nil {
		return err
	}

	// Set the name
	db.Name = name

	// Validate the database config
	if err := db.Validate(); err != nil {
		return err
	}

	// Add to config
	config.Databases[name] = db

	// Save config
	return m.Save(config)
}

// RemoveDatabase removes a database configuration.
func (m *Manager) RemoveDatabase(name string) error {
	config, err := m.Load()
	if err != nil {
		return err
	}

	// Check if database exists
	if _, exists := config.Databases[name]; !exists {
		return &DatabaseNotFoundError{Name: name}
	}

	// Remove from config
	delete(config.Databases, name)

	// Save config
	return m.Save(config)
}

// ListDatabases returns a list of all configured database names.
func (m *Manager) ListDatabases() ([]string, error) {
	config, err := m.Load()
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(config.Databases))
	for name := range config.Databases {
		names = append(names, name)
	}

	return names, nil
}

// DatabaseExists checks if a database configuration exists.
func (m *Manager) DatabaseExists(name string) (bool, error) {
	config, err := m.Load()
	if err != nil {
		return false, err
	}

	_, exists := config.Databases[name]
	return exists, nil
}

// GetConfigPath returns the path to the config file.
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".cadangkan", "config.yaml"), nil
}

// EnsureConfigDir ensures the config directory exists.
func EnsureConfigDir() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	configDir := filepath.Join(homeDir, ".cadangkan")
	return os.MkdirAll(configDir, 0700)
}
