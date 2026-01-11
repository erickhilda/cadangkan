package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()

	if cfg.Version != "1.0" {
		t.Errorf("NewConfig() version = %v, want 1.0", cfg.Version)
	}

	if cfg.Databases == nil {
		t.Error("NewConfig() databases map is nil")
	}

	if len(cfg.Databases) != 0 {
		t.Errorf("NewConfig() databases count = %v, want 0", len(cfg.Databases))
	}
}

func TestNewDatabaseConfig(t *testing.T) {
	cfg := NewDatabaseConfig()

	if cfg.Type != "mysql" {
		t.Errorf("NewDatabaseConfig() type = %v, want mysql", cfg.Type)
	}

	if cfg.Port != 3306 {
		t.Errorf("NewDatabaseConfig() port = %v, want 3306", cfg.Port)
	}
}

func TestManagerLoadSave(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	mgr := &YAMLManager{configPath: configPath}

	// Test loading non-existent config (should return empty config)
	cfg, err := mgr.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}

	if len(cfg.Databases) != 0 {
		t.Errorf("Load() empty config databases count = %v, want 0", len(cfg.Databases))
	}

	// Add a database
	cfg.Databases["test"] = &DatabaseConfig{
		Name:     "test",
		Type:     "mysql",
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
		User:     "testuser",
	}

	// Save config
	err = mgr.Save(cfg)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Check file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Save() did not create config file")
	}

	// Load config again
	cfg2, err := mgr.Load()
	if err != nil {
		t.Fatalf("Load() after save error = %v", err)
	}

	// Check database was loaded
	if len(cfg2.Databases) != 1 {
		t.Errorf("Load() databases count = %v, want 1", len(cfg2.Databases))
	}

	db, exists := cfg2.Databases["test"]
	if !exists {
		t.Fatal("Load() database 'test' not found")
	}

	if db.Host != "localhost" {
		t.Errorf("Load() database host = %v, want localhost", db.Host)
	}

	if db.Port != 3306 {
		t.Errorf("Load() database port = %v, want 3306", db.Port)
	}

	if db.Database != "testdb" {
		t.Errorf("Load() database name = %v, want testdb", db.Database)
	}

	if db.User != "testuser" {
		t.Errorf("Load() database user = %v, want testuser", db.User)
	}
}

func TestManagerAddDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	mgr := &YAMLManager{configPath: configPath}

	// Add a database
	db := &DatabaseConfig{
		Type:     "mysql",
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
		User:     "testuser",
	}

	err := mgr.AddDatabase("test", db)
	if err != nil {
		t.Fatalf("AddDatabase() error = %v", err)
	}

	// Load and verify
	cfg, err := mgr.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(cfg.Databases) != 1 {
		t.Errorf("AddDatabase() databases count = %v, want 1", len(cfg.Databases))
	}

	addedDb, exists := cfg.Databases["test"]
	if !exists {
		t.Fatal("AddDatabase() database not found")
	}

	if addedDb.Name != "test" {
		t.Errorf("AddDatabase() name = %v, want test", addedDb.Name)
	}
}

func TestManagerGetDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	mgr := &YAMLManager{configPath: configPath}

	// Add a database
	db := &DatabaseConfig{
		Type:     "mysql",
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
		User:     "testuser",
	}

	err := mgr.AddDatabase("test", db)
	if err != nil {
		t.Fatalf("AddDatabase() error = %v", err)
	}

	// Get database
	retrieved, err := mgr.GetDatabase("test")
	if err != nil {
		t.Fatalf("GetDatabase() error = %v", err)
	}

	if retrieved.Name != "test" {
		t.Errorf("GetDatabase() name = %v, want test", retrieved.Name)
	}

	if retrieved.Host != "localhost" {
		t.Errorf("GetDatabase() host = %v, want localhost", retrieved.Host)
	}

	// Try to get non-existent database
	_, err = mgr.GetDatabase("nonexistent")
	if err == nil {
		t.Error("GetDatabase() non-existent should return error")
	}

	if _, ok := err.(*DatabaseNotFoundError); !ok {
		t.Errorf("GetDatabase() error type = %T, want *DatabaseNotFoundError", err)
	}
}

func TestManagerRemoveDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	mgr := &YAMLManager{configPath: configPath}

	// Add a database
	db := &DatabaseConfig{
		Type:     "mysql",
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
		User:     "testuser",
	}

	err := mgr.AddDatabase("test", db)
	if err != nil {
		t.Fatalf("AddDatabase() error = %v", err)
	}

	// Remove database
	err = mgr.RemoveDatabase("test")
	if err != nil {
		t.Fatalf("RemoveDatabase() error = %v", err)
	}

	// Verify it's gone
	cfg, err := mgr.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(cfg.Databases) != 0 {
		t.Errorf("RemoveDatabase() databases count = %v, want 0", len(cfg.Databases))
	}

	// Try to remove non-existent database
	err = mgr.RemoveDatabase("nonexistent")
	if err == nil {
		t.Error("RemoveDatabase() non-existent should return error")
	}
}

func TestManagerListDatabases(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	mgr := &YAMLManager{configPath: configPath}

	// Add multiple databases
	databases := []string{"db1", "db2", "db3"}
	for _, name := range databases {
		db := &DatabaseConfig{
			Type:     "mysql",
			Host:     "localhost",
			Port:     3306,
			Database: name,
			User:     "testuser",
		}
		err := mgr.AddDatabase(name, db)
		if err != nil {
			t.Fatalf("AddDatabase() error = %v", err)
		}
	}

	// List databases
	names, err := mgr.ListDatabases()
	if err != nil {
		t.Fatalf("ListDatabases() error = %v", err)
	}

	if len(names) != 3 {
		t.Errorf("ListDatabases() count = %v, want 3", len(names))
	}

	// Verify all names are present
	nameMap := make(map[string]bool)
	for _, name := range names {
		nameMap[name] = true
	}

	for _, expected := range databases {
		if !nameMap[expected] {
			t.Errorf("ListDatabases() missing database: %v", expected)
		}
	}
}

func TestManagerDatabaseExists(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	mgr := &YAMLManager{configPath: configPath}

	// Add a database
	db := &DatabaseConfig{
		Type:     "mysql",
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
		User:     "testuser",
	}

	err := mgr.AddDatabase("test", db)
	if err != nil {
		t.Fatalf("AddDatabase() error = %v", err)
	}

	// Check existence
	exists, err := mgr.DatabaseExists("test")
	if err != nil {
		t.Fatalf("DatabaseExists() error = %v", err)
	}

	if !exists {
		t.Error("DatabaseExists() = false, want true")
	}

	// Check non-existent
	exists, err = mgr.DatabaseExists("nonexistent")
	if err != nil {
		t.Fatalf("DatabaseExists() error = %v", err)
	}

	if exists {
		t.Error("DatabaseExists() = true, want false")
	}
}
