package config

// Manager defines the interface for configuration storage backends.
// This abstraction allows swapping YAML for SQLite or other backends
// in the future without changing the public API.
type Manager interface {
	Load() (*Config, error)
	Save(*Config) error
	GetDatabase(name string) (*DatabaseConfig, error)
	AddDatabase(name string, db *DatabaseConfig) error
	RemoveDatabase(name string) error
	ListDatabases() ([]string, error)
	DatabaseExists(name string) (bool, error)
}
