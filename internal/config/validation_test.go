package config

import (
	"testing"
)

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Version: "1.0",
				Databases: map[string]*DatabaseConfig{
					"test": {
						Name:     "test",
						Type:     "mysql",
						Host:     "localhost",
						Port:     3306,
						Database: "testdb",
						User:     "testuser",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing version",
			config: &Config{
				Version: "",
				Databases: map[string]*DatabaseConfig{
					"test": {
						Type:     "mysql",
						Host:     "localhost",
						Port:     3306,
						Database: "testdb",
						User:     "testuser",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid database config",
			config: &Config{
				Version: "1.0",
				Databases: map[string]*DatabaseConfig{
					"test": {
						Type: "mysql",
						// Missing required fields
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDatabaseConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *DatabaseConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: &DatabaseConfig{
				Type:     "mysql",
				Host:     "localhost",
				Port:     3306,
				Database: "testdb",
				User:     "testuser",
			},
			wantErr: false,
		},
		{
			name: "missing type",
			config: &DatabaseConfig{
				Type:     "",
				Host:     "localhost",
				Port:     3306,
				Database: "testdb",
				User:     "testuser",
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			config: &DatabaseConfig{
				Type:     "postgres",
				Host:     "localhost",
				Port:     3306,
				Database: "testdb",
				User:     "testuser",
			},
			wantErr: true,
		},
		{
			name: "missing host",
			config: &DatabaseConfig{
				Type:     "mysql",
				Host:     "",
				Port:     3306,
				Database: "testdb",
				User:     "testuser",
			},
			wantErr: true,
		},
		{
			name: "invalid port (too low)",
			config: &DatabaseConfig{
				Type:     "mysql",
				Host:     "localhost",
				Port:     0,
				Database: "testdb",
				User:     "testuser",
			},
			wantErr: true,
		},
		{
			name: "invalid port (too high)",
			config: &DatabaseConfig{
				Type:     "mysql",
				Host:     "localhost",
				Port:     99999,
				Database: "testdb",
				User:     "testuser",
			},
			wantErr: true,
		},
		{
			name: "missing user",
			config: &DatabaseConfig{
				Type:     "mysql",
				Host:     "localhost",
				Port:     3306,
				Database: "testdb",
				User:     "",
			},
			wantErr: true,
		},
		{
			name: "missing database",
			config: &DatabaseConfig{
				Type:     "mysql",
				Host:     "localhost",
				Port:     3306,
				Database: "",
				User:     "testuser",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("DatabaseConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple name",
			input: "production",
			want:  "production",
		},
		{
			name:  "name with spaces",
			input: "my production",
			want:  "my_production",
		},
		{
			name:  "name with dashes",
			input: "my-production",
			want:  "my_production",
		},
		{
			name:  "name with dots",
			input: "my.production",
			want:  "my_production",
		},
		{
			name:  "uppercase name",
			input: "PRODUCTION",
			want:  "production",
		},
		{
			name:  "mixed case with special chars",
			input: "My-Production.DB",
			want:  "my_production_db",
		},
		{
			name:  "name with leading/trailing spaces",
			input: "  production  ",
			want:  "production",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeName(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeName() = %v, want %v", got, tt.want)
			}
		})
	}
}
