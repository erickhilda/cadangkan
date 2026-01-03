package mysql

import (
	"database/sql"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Config Tests ---

func TestNewConfig(t *testing.T) {
	config := NewConfig()

	assert.Equal(t, DefaultPort, config.Port)
	assert.Equal(t, DefaultTimeout, config.Timeout)
	assert.Equal(t, DefaultMaxOpenConns, config.MaxOpenConns)
	assert.Equal(t, DefaultMaxIdleConns, config.MaxIdleConns)
	assert.Equal(t, DefaultConnMaxLife, config.ConnMaxLifetime)
	assert.Equal(t, DefaultConnMaxIdle, config.ConnMaxIdleTime)
	assert.True(t, config.ParseTime)
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantError bool
		errField  string
	}{
		{
			name: "valid config",
			config: &Config{
				Host: "localhost",
				Port: 3306,
				User: "root",
			},
			wantError: false,
		},
		{
			name: "missing host",
			config: &Config{
				Port: 3306,
				User: "root",
			},
			wantError: true,
			errField:  "Host",
		},
		{
			name: "missing user",
			config: &Config{
				Host: "localhost",
				Port: 3306,
			},
			wantError: true,
			errField:  "User",
		},
		{
			name: "invalid port - zero",
			config: &Config{
				Host: "localhost",
				Port: 0,
				User: "root",
			},
			wantError: true,
			errField:  "Port",
		},
		{
			name: "invalid port - too high",
			config: &Config{
				Host: "localhost",
				Port: 70000,
				User: "root",
			},
			wantError: true,
			errField:  "Port",
		},
		{
			name: "negative timeout",
			config: &Config{
				Host:    "localhost",
				Port:    3306,
				User:    "root",
				Timeout: -1 * time.Second,
			},
			wantError: true,
			errField:  "Timeout",
		},
		{
			name: "negative max open conns",
			config: &Config{
				Host:         "localhost",
				Port:         3306,
				User:         "root",
				MaxOpenConns: -1,
			},
			wantError: true,
			errField:  "MaxOpenConns",
		},
		{
			name: "negative max idle conns",
			config: &Config{
				Host:         "localhost",
				Port:         3306,
				User:         "root",
				MaxIdleConns: -1,
			},
			wantError: true,
			errField:  "MaxIdleConns",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantError {
				assert.Error(t, err)
				var configErr *ConfigError
				if errors.As(err, &configErr) {
					assert.Equal(t, tt.errField, configErr.Field)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigDSN(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		contains []string
	}{
		{
			name: "basic DSN",
			config: &Config{
				Host:     "localhost",
				Port:     3306,
				User:     "root",
				Password: "secret",
				Database: "testdb",
				Timeout:  10 * time.Second,
			},
			contains: []string{
				"root:secret@tcp(localhost:3306)/testdb",
				"timeout=10s",
				"charset=utf8mb4",
			},
		},
		{
			name: "DSN with parseTime",
			config: &Config{
				Host:      "localhost",
				Port:      3306,
				User:      "root",
				Password:  "secret",
				ParseTime: true,
			},
			contains: []string{
				"parseTime=true",
			},
		},
		{
			name: "DSN with TLS",
			config: &Config{
				Host:     "localhost",
				Port:     3306,
				User:     "root",
				Password: "secret",
				TLS:      "skip-verify",
			},
			contains: []string{
				"tls=skip-verify",
			},
		},
		{
			name: "DSN without database",
			config: &Config{
				Host:     "localhost",
				Port:     3306,
				User:     "root",
				Password: "secret",
			},
			contains: []string{
				"root:secret@tcp(localhost:3306)/",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn := tt.config.DSN()
			for _, substr := range tt.contains {
				assert.Contains(t, dsn, substr)
			}
		})
	}
}

func TestConfigDSNMasked(t *testing.T) {
	config := &Config{
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "supersecret",
		Database: "testdb",
	}

	masked := config.DSNMasked()

	assert.Contains(t, masked, "root:***@tcp(localhost:3306)/testdb")
	assert.NotContains(t, masked, "supersecret")
}

func TestConfigChaining(t *testing.T) {
	config := NewConfig().
		WithHost("db.example.com").
		WithPort(3307).
		WithUser("admin").
		WithPassword("pass").
		WithDatabase("mydb").
		WithTimeout(5 * time.Second)

	assert.Equal(t, "db.example.com", config.Host)
	assert.Equal(t, 3307, config.Port)
	assert.Equal(t, "admin", config.User)
	assert.Equal(t, "pass", config.Password)
	assert.Equal(t, "mydb", config.Database)
	assert.Equal(t, 5*time.Second, config.Timeout)
}

// --- Client Tests ---

func TestNewClient(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := NewConfig().WithHost("localhost").WithUser("root")
		client, err := NewClient(config)

		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.False(t, client.IsConnected())
	})

	t.Run("nil config", func(t *testing.T) {
		client, err := NewClient(nil)

		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Equal(t, ErrInvalidConfig, err)
	})

	t.Run("invalid config", func(t *testing.T) {
		config := &Config{} // Missing required fields
		client, err := NewClient(config)

		assert.Error(t, err)
		assert.Nil(t, client)
		assert.True(t, IsConfigError(err))
	})
}

func TestNewClientWithDB(t *testing.T) {
	t.Run("with db", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		config := NewConfig().WithHost("localhost").WithUser("root")
		client, err := NewClientWithDB(config, db)

		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.True(t, client.IsConnected())
	})

	t.Run("nil config", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		client, err := NewClientWithDB(nil, db)
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Equal(t, ErrInvalidConfig, err)
	})

	t.Run("nil db", func(t *testing.T) {
		config := NewConfig().WithHost("localhost").WithUser("root")
		client, err := NewClientWithDB(config, nil)

		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.False(t, client.IsConnected())
	})
}

func TestClientConnect(t *testing.T) {
	t.Run("successful connection", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectPing()

		config := NewConfig().WithHost("localhost").WithUser("root")
		client, _ := NewClientWithDB(config, nil)
		client.db = db
		client.connected = false

		// Since we can't easily test Connect() with sqlmock (it calls sql.Open),
		// we test the behavior through NewClientWithDB
		client2, err := NewClientWithDB(config, db)
		assert.NoError(t, err)
		assert.True(t, client2.IsConnected())
	})

	t.Run("already connected", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		config := NewConfig().WithHost("localhost").WithUser("root")
		client, _ := NewClientWithDB(config, db)

		err = client.Connect()
		assert.Error(t, err)
		assert.Equal(t, ErrAlreadyConnected, err)
	})
}

func TestClientPing(t *testing.T) {
	t.Run("successful ping", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectPing()

		config := NewConfig().WithHost("localhost").WithUser("root").WithTimeout(5 * time.Second)
		client, _ := NewClientWithDB(config, db)

		err = client.Ping()
		assert.NoError(t, err)
	})

	t.Run("ping when not connected", func(t *testing.T) {
		config := NewConfig().WithHost("localhost").WithUser("root")
		client, _ := NewClient(config)

		err := client.Ping()
		assert.Error(t, err)
		assert.Equal(t, ErrNotConnected, err)
	})

	t.Run("ping failure", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectPing().WillReturnError(errors.New("connection lost"))

		config := NewConfig().WithHost("localhost").WithUser("root").WithTimeout(5 * time.Second)
		client, _ := NewClientWithDB(config, db)

		err = client.Ping()
		assert.Error(t, err)
		assert.True(t, IsConnectionError(err))
	})
}

func TestClientClose(t *testing.T) {
	t.Run("successful close", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)

		mock.ExpectClose()

		config := NewConfig().WithHost("localhost").WithUser("root")
		client, _ := NewClientWithDB(config, db)

		err = client.Close()
		assert.NoError(t, err)
		assert.False(t, client.IsConnected())
	})

	t.Run("close when not connected", func(t *testing.T) {
		config := NewConfig().WithHost("localhost").WithUser("root")
		client, _ := NewClient(config)

		err := client.Close()
		assert.NoError(t, err) // Should not error on closing non-connected client
	})
}

func TestClientGetVersion(t *testing.T) {
	t.Run("successful get version", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows([]string{"VERSION()"}).AddRow("8.0.35")
		mock.ExpectQuery("SELECT VERSION()").WillReturnRows(rows)

		config := NewConfig().WithHost("localhost").WithUser("root").WithTimeout(5 * time.Second)
		client, _ := NewClientWithDB(config, db)

		version, err := client.GetVersion()
		assert.NoError(t, err)
		assert.Equal(t, "8.0.35", version)
	})

	t.Run("not connected", func(t *testing.T) {
		config := NewConfig().WithHost("localhost").WithUser("root")
		client, _ := NewClient(config)

		_, err := client.GetVersion()
		assert.Error(t, err)
		assert.Equal(t, ErrNotConnected, err)
	})

	t.Run("query error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery("SELECT VERSION()").WillReturnError(errors.New("query failed"))

		config := NewConfig().WithHost("localhost").WithUser("root").WithTimeout(5 * time.Second)
		client, _ := NewClientWithDB(config, db)

		_, err = client.GetVersion()
		assert.Error(t, err)
		assert.True(t, IsQueryError(err))
	})
}

func TestClientGetDatabases(t *testing.T) {
	t.Run("successful get databases", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows([]string{"Database"}).
			AddRow("information_schema").
			AddRow("mysql").
			AddRow("testdb")
		mock.ExpectQuery("SHOW DATABASES").WillReturnRows(rows)

		config := NewConfig().WithHost("localhost").WithUser("root").WithTimeout(5 * time.Second)
		client, _ := NewClientWithDB(config, db)

		databases, err := client.GetDatabases()
		assert.NoError(t, err)
		assert.Equal(t, []string{"information_schema", "mysql", "testdb"}, databases)
	})

	t.Run("empty database list", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows([]string{"Database"})
		mock.ExpectQuery("SHOW DATABASES").WillReturnRows(rows)

		config := NewConfig().WithHost("localhost").WithUser("root").WithTimeout(5 * time.Second)
		client, _ := NewClientWithDB(config, db)

		databases, err := client.GetDatabases()
		assert.NoError(t, err)
		assert.Nil(t, databases)
	})

	t.Run("not connected", func(t *testing.T) {
		config := NewConfig().WithHost("localhost").WithUser("root")
		client, _ := NewClient(config)

		_, err := client.GetDatabases()
		assert.Error(t, err)
		assert.Equal(t, ErrNotConnected, err)
	})
}

func TestClientGetTables(t *testing.T) {
	t.Run("successful get tables", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows([]string{"Tables_in_testdb"}).
			AddRow("users").
			AddRow("orders").
			AddRow("products")
		mock.ExpectQuery("SHOW TABLES FROM `testdb`").WillReturnRows(rows)

		config := NewConfig().WithHost("localhost").WithUser("root").WithTimeout(5 * time.Second)
		client, _ := NewClientWithDB(config, db)

		tables, err := client.GetTables("testdb")
		assert.NoError(t, err)
		assert.Equal(t, []string{"users", "orders", "products"}, tables)
	})

	t.Run("empty database name", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		config := NewConfig().WithHost("localhost").WithUser("root").WithTimeout(5 * time.Second)
		client, _ := NewClientWithDB(config, db)

		_, err = client.GetTables("")
		assert.Error(t, err)
		assert.True(t, IsConfigError(err))
	})

	t.Run("not connected", func(t *testing.T) {
		config := NewConfig().WithHost("localhost").WithUser("root")
		client, _ := NewClient(config)

		_, err := client.GetTables("testdb")
		assert.Error(t, err)
		assert.Equal(t, ErrNotConnected, err)
	})
}

func TestClientGetTableSize(t *testing.T) {
	t.Run("successful get table size", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows([]string{"size"}).AddRow(1024000)
		mock.ExpectQuery("SELECT COALESCE").
			WithArgs("testdb", "users").
			WillReturnRows(rows)

		config := NewConfig().WithHost("localhost").WithUser("root").WithTimeout(5 * time.Second)
		client, _ := NewClientWithDB(config, db)

		size, err := client.GetTableSize("testdb", "users")
		assert.NoError(t, err)
		assert.Equal(t, int64(1024000), size)
	})

	t.Run("table not found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery("SELECT COALESCE").
			WithArgs("testdb", "nonexistent").
			WillReturnError(sql.ErrNoRows)

		config := NewConfig().WithHost("localhost").WithUser("root").WithTimeout(5 * time.Second)
		client, _ := NewClientWithDB(config, db)

		_, err = client.GetTableSize("testdb", "nonexistent")
		assert.Error(t, err)
		assert.Equal(t, ErrEmptyResult, err)
	})

	t.Run("empty database name", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		config := NewConfig().WithHost("localhost").WithUser("root").WithTimeout(5 * time.Second)
		client, _ := NewClientWithDB(config, db)

		_, err = client.GetTableSize("", "users")
		assert.Error(t, err)
		assert.True(t, IsConfigError(err))
	})

	t.Run("empty table name", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		config := NewConfig().WithHost("localhost").WithUser("root").WithTimeout(5 * time.Second)
		client, _ := NewClientWithDB(config, db)

		_, err = client.GetTableSize("testdb", "")
		assert.Error(t, err)
		assert.True(t, IsConfigError(err))
	})
}

func TestClientGetTableRowCount(t *testing.T) {
	t.Run("successful get row count", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows([]string{"row_count"}).AddRow(5000)
		mock.ExpectQuery("SELECT COALESCE").
			WithArgs("testdb", "users").
			WillReturnRows(rows)

		config := NewConfig().WithHost("localhost").WithUser("root").WithTimeout(5 * time.Second)
		client, _ := NewClientWithDB(config, db)

		count, err := client.GetTableRowCount("testdb", "users")
		assert.NoError(t, err)
		assert.Equal(t, int64(5000), count)
	})

	t.Run("not connected", func(t *testing.T) {
		config := NewConfig().WithHost("localhost").WithUser("root")
		client, _ := NewClient(config)

		_, err := client.GetTableRowCount("testdb", "users")
		assert.Error(t, err)
		assert.Equal(t, ErrNotConnected, err)
	})
}

func TestClientGetDatabaseSize(t *testing.T) {
	t.Run("successful get database size", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows([]string{"size"}).AddRow(10240000)
		mock.ExpectQuery("SELECT COALESCE").
			WithArgs("testdb").
			WillReturnRows(rows)

		config := NewConfig().WithHost("localhost").WithUser("root").WithTimeout(5 * time.Second)
		client, _ := NewClientWithDB(config, db)

		size, err := client.GetDatabaseSize("testdb")
		assert.NoError(t, err)
		assert.Equal(t, int64(10240000), size)
	})

	t.Run("empty database name", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		config := NewConfig().WithHost("localhost").WithUser("root").WithTimeout(5 * time.Second)
		client, _ := NewClientWithDB(config, db)

		_, err = client.GetDatabaseSize("")
		assert.Error(t, err)
		assert.True(t, IsConfigError(err))
	})

	t.Run("not connected", func(t *testing.T) {
		config := NewConfig().WithHost("localhost").WithUser("root")
		client, _ := NewClient(config)

		_, err := client.GetDatabaseSize("testdb")
		assert.Error(t, err)
		assert.Equal(t, ErrNotConnected, err)
	})
}

func TestClientGetTableInfo(t *testing.T) {
	t.Run("successful get table info", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		now := time.Now()
		rows := sqlmock.NewRows([]string{
			"table_name", "engine", "row_count", "data_size",
			"index_size", "total_size", "create_time", "update_time",
		}).AddRow("users", "InnoDB", 1000, 50000, 10000, 60000, now, now)

		mock.ExpectQuery("SELECT").
			WithArgs("testdb", "users").
			WillReturnRows(rows)

		config := NewConfig().WithHost("localhost").WithUser("root").WithTimeout(5 * time.Second)
		client, _ := NewClientWithDB(config, db)

		info, err := client.GetTableInfo("testdb", "users")
		assert.NoError(t, err)
		assert.Equal(t, "users", info.Name)
		assert.Equal(t, "InnoDB", info.Engine)
		assert.Equal(t, int64(1000), info.RowCount)
		assert.Equal(t, int64(50000), info.DataSize)
		assert.Equal(t, int64(10000), info.IndexSize)
		assert.Equal(t, int64(60000), info.TotalSize)
	})

	t.Run("table not found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery("SELECT").
			WithArgs("testdb", "nonexistent").
			WillReturnError(sql.ErrNoRows)

		config := NewConfig().WithHost("localhost").WithUser("root").WithTimeout(5 * time.Second)
		client, _ := NewClientWithDB(config, db)

		_, err = client.GetTableInfo("testdb", "nonexistent")
		assert.Error(t, err)
		assert.Equal(t, ErrEmptyResult, err)
	})
}

func TestClientGetDatabaseInfo(t *testing.T) {
	t.Run("successful get database info", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		now := time.Now()
		rows := sqlmock.NewRows([]string{
			"table_name", "engine", "row_count", "data_size",
			"index_size", "total_size", "create_time", "update_time",
		}).
			AddRow("users", "InnoDB", 1000, 50000, 10000, 60000, now, now).
			AddRow("orders", "InnoDB", 5000, 100000, 20000, 120000, now, nil)

		mock.ExpectQuery("SELECT").
			WithArgs("testdb").
			WillReturnRows(rows)

		config := NewConfig().WithHost("localhost").WithUser("root").WithTimeout(5 * time.Second)
		client, _ := NewClientWithDB(config, db)

		info, err := client.GetDatabaseInfo("testdb")
		assert.NoError(t, err)
		assert.Equal(t, "testdb", info.Name)
		assert.Equal(t, 2, info.TableCount)
		assert.Equal(t, int64(180000), info.TotalSize)
		assert.Len(t, info.Tables, 2)
	})

	t.Run("empty database", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows([]string{
			"table_name", "engine", "row_count", "data_size",
			"index_size", "total_size", "create_time", "update_time",
		})

		mock.ExpectQuery("SELECT").
			WithArgs("emptydb").
			WillReturnRows(rows)

		config := NewConfig().WithHost("localhost").WithUser("root").WithTimeout(5 * time.Second)
		client, _ := NewClientWithDB(config, db)

		info, err := client.GetDatabaseInfo("emptydb")
		assert.NoError(t, err)
		assert.Equal(t, "emptydb", info.Name)
		assert.Equal(t, 0, info.TableCount)
		assert.Equal(t, int64(0), info.TotalSize)
	})
}

func TestClientExecuteQuery(t *testing.T) {
	t.Run("successful query", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "Alice").
			AddRow(2, "Bob")
		mock.ExpectQuery("SELECT").WillReturnRows(rows)

		config := NewConfig().WithHost("localhost").WithUser("root").WithTimeout(5 * time.Second)
		client, _ := NewClientWithDB(config, db)

		result, err := client.ExecuteQuery("SELECT * FROM users")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		result.Close()
	})

	t.Run("not connected", func(t *testing.T) {
		config := NewConfig().WithHost("localhost").WithUser("root")
		client, _ := NewClient(config)

		_, err := client.ExecuteQuery("SELECT 1")
		assert.Error(t, err)
		assert.Equal(t, ErrNotConnected, err)
	})
}

func TestClientExecute(t *testing.T) {
	t.Run("successful execute", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectExec("INSERT INTO").WillReturnResult(sqlmock.NewResult(1, 1))

		config := NewConfig().WithHost("localhost").WithUser("root").WithTimeout(5 * time.Second)
		client, _ := NewClientWithDB(config, db)

		result, err := client.Execute("INSERT INTO users (name) VALUES (?)", "Alice")
		assert.NoError(t, err)
		assert.NotNil(t, result)

		id, _ := result.LastInsertId()
		affected, _ := result.RowsAffected()
		assert.Equal(t, int64(1), id)
		assert.Equal(t, int64(1), affected)
	})

	t.Run("not connected", func(t *testing.T) {
		config := NewConfig().WithHost("localhost").WithUser("root")
		client, _ := NewClient(config)

		_, err := client.Execute("DELETE FROM users")
		assert.Error(t, err)
		assert.Equal(t, ErrNotConnected, err)
	})
}

func TestClientConcurrentAccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Set up expectations for concurrent calls
	for i := 0; i < 10; i++ {
		rows := sqlmock.NewRows([]string{"VERSION()"}).AddRow("8.0.35")
		mock.ExpectQuery("SELECT VERSION()").WillReturnRows(rows)
	}

	config := NewConfig().WithHost("localhost").WithUser("root").WithTimeout(5 * time.Second)
	client, _ := NewClientWithDB(config, db)

	var wg sync.WaitGroup
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := client.GetVersion()
			if err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("concurrent access error: %v", err)
	}
}

func TestClientConfig(t *testing.T) {
	config := NewConfig().WithHost("localhost").WithUser("root")
	client, _ := NewClient(config)

	returnedConfig := client.Config()
	assert.Equal(t, "localhost", returnedConfig.Host)
	assert.Equal(t, "root", returnedConfig.User)
}

// --- Error Tests ---

func TestConnectionError(t *testing.T) {
	err := &ConnectionError{
		Host:    "localhost",
		Port:    3306,
		Message: "failed to connect",
		Err:     errors.New("connection refused"),
	}

	assert.Contains(t, err.Error(), "localhost:3306")
	assert.Contains(t, err.Error(), "failed to connect")
	assert.Contains(t, err.Error(), "connection refused")
	assert.Equal(t, errors.New("connection refused"), err.Unwrap())
}

func TestConnectionErrorWithoutUnderlying(t *testing.T) {
	err := &ConnectionError{
		Host:    "localhost",
		Port:    3306,
		Message: "failed to connect",
	}

	assert.Contains(t, err.Error(), "localhost:3306")
	assert.Contains(t, err.Error(), "failed to connect")
	assert.Nil(t, err.Unwrap())
}

func TestQueryError(t *testing.T) {
	err := &QueryError{
		Query:   "SELECT * FROM users",
		Message: "syntax error",
		Err:     errors.New("near 'FORM'"),
	}

	assert.Contains(t, err.Error(), "SELECT * FROM users")
	assert.Contains(t, err.Error(), "syntax error")
	assert.Equal(t, errors.New("near 'FORM'"), err.Unwrap())
}

func TestQueryErrorLongQuery(t *testing.T) {
	longQuery := "SELECT " + string(make([]byte, 200)) + " FROM users"
	err := &QueryError{
		Query:   longQuery,
		Message: "error",
	}

	// Query should be truncated
	assert.Less(t, len(err.Error()), len(longQuery))
	assert.Contains(t, err.Error(), "...")
}

func TestTimeoutError(t *testing.T) {
	err := &TimeoutError{
		Operation: "connect",
		Duration:  "10s",
		Err:       errors.New("deadline exceeded"),
	}

	assert.Contains(t, err.Error(), "connect")
	assert.Contains(t, err.Error(), "10s")
	assert.Equal(t, errors.New("deadline exceeded"), err.Unwrap())
}

func TestConfigError(t *testing.T) {
	err := &ConfigError{
		Field:   "Host",
		Message: "host is required",
	}

	assert.Contains(t, err.Error(), "Host")
	assert.Contains(t, err.Error(), "host is required")
}

func TestErrorHelpers(t *testing.T) {
	connErr := &ConnectionError{Host: "localhost", Port: 3306, Message: "test"}
	queryErr := &QueryError{Query: "SELECT 1", Message: "test"}
	timeoutErr := &TimeoutError{Operation: "ping", Duration: "5s"}
	configErr := &ConfigError{Field: "Host", Message: "required"}

	assert.True(t, IsConnectionError(connErr))
	assert.False(t, IsConnectionError(queryErr))

	assert.True(t, IsQueryError(queryErr))
	assert.False(t, IsQueryError(connErr))

	assert.True(t, IsTimeoutError(timeoutErr))
	assert.False(t, IsTimeoutError(queryErr))

	assert.True(t, IsConfigError(configErr))
	assert.False(t, IsConfigError(queryErr))
}

func TestWrapErrors(t *testing.T) {
	underlying := errors.New("underlying error")

	connErr := WrapConnectionError("host", 3306, "message", underlying)
	assert.True(t, IsConnectionError(connErr))

	queryErr := WrapQueryError("SELECT 1", "message", underlying)
	assert.True(t, IsQueryError(queryErr))

	timeoutErr := WrapTimeoutError("connect", "10s", underlying)
	assert.True(t, IsTimeoutError(timeoutErr))
}

// --- Mock Client Tests ---

func TestMockClient(t *testing.T) {
	mock := NewMockClient()

	// Test initial state
	assert.False(t, mock.IsConnected())

	// Test Connect
	err := mock.Connect()
	assert.NoError(t, err)
	assert.True(t, mock.IsConnected())

	// Test connect error
	mock2 := NewMockClient()
	mock2.ConnectErr = errors.New("connection refused")
	err = mock2.Connect()
	assert.Error(t, err)
}

func TestMockClientVersion(t *testing.T) {
	mock := NewMockClient()
	mock.SetConnected(true)
	mock.Version = "8.0.35"

	version, err := mock.GetVersion()
	assert.NoError(t, err)
	assert.Equal(t, "8.0.35", version)
}

func TestMockClientDatabases(t *testing.T) {
	mock := NewMockClient()
	mock.SetConnected(true)
	mock.Databases = []string{"db1", "db2", "db3"}

	databases, err := mock.GetDatabases()
	assert.NoError(t, err)
	assert.Equal(t, []string{"db1", "db2", "db3"}, databases)
}

func TestMockClientTables(t *testing.T) {
	mock := NewMockClient()
	mock.SetConnected(true)
	mock.SetTables("testdb", []string{"users", "orders"})

	tables, err := mock.GetTables("testdb")
	assert.NoError(t, err)
	assert.Equal(t, []string{"users", "orders"}, tables)

	// Test non-existent database
	tables, err = mock.GetTables("nonexistent")
	assert.NoError(t, err)
	assert.Empty(t, tables)
}

func TestMockClientTableSize(t *testing.T) {
	mock := NewMockClient()
	mock.SetConnected(true)
	mock.SetTableSize("testdb", "users", 1024000)

	size, err := mock.GetTableSize("testdb", "users")
	assert.NoError(t, err)
	assert.Equal(t, int64(1024000), size)

	// Test non-existent table
	_, err = mock.GetTableSize("testdb", "nonexistent")
	assert.Equal(t, ErrEmptyResult, err)
}

func TestMockClientRowCount(t *testing.T) {
	mock := NewMockClient()
	mock.SetConnected(true)
	mock.SetRowCount("testdb", "users", 5000)

	count, err := mock.GetTableRowCount("testdb", "users")
	assert.NoError(t, err)
	assert.Equal(t, int64(5000), count)
}

func TestMockClientDatabaseSize(t *testing.T) {
	mock := NewMockClient()
	mock.SetConnected(true)
	mock.SetDatabaseSize("testdb", 10240000)

	size, err := mock.GetDatabaseSize("testdb")
	assert.NoError(t, err)
	assert.Equal(t, int64(10240000), size)
}

func TestMockClientTableInfo(t *testing.T) {
	mock := NewMockClient()
	mock.SetConnected(true)

	info := &TableInfo{
		Name:      "users",
		Engine:    "InnoDB",
		RowCount:  1000,
		DataSize:  50000,
		IndexSize: 10000,
		TotalSize: 60000,
	}
	mock.SetTableInfo("testdb", "users", info)

	result, err := mock.GetTableInfo("testdb", "users")
	assert.NoError(t, err)
	assert.Equal(t, info, result)
}

func TestMockClientDatabaseInfo(t *testing.T) {
	mock := NewMockClient()
	mock.SetConnected(true)

	info := &DatabaseInfo{
		Name:       "testdb",
		TableCount: 2,
		TotalSize:  100000,
	}
	mock.SetDatabaseInfo("testdb", info)

	result, err := mock.GetDatabaseInfo("testdb")
	assert.NoError(t, err)
	assert.Equal(t, info, result)
}

func TestMockClientCallTracking(t *testing.T) {
	mock := NewMockClient()
	mock.SetConnected(true)
	mock.Version = "8.0"

	mock.GetVersion()
	mock.GetVersion()
	mock.GetDatabases()

	assert.Equal(t, 2, mock.GetCallCount("GetVersion"))
	assert.Equal(t, 1, mock.GetCallCount("GetDatabases"))
	assert.Equal(t, 0, mock.GetCallCount("GetTables"))

	calls := mock.GetCalls()
	assert.Len(t, calls, 3)

	mock.ResetCalls()
	assert.Empty(t, mock.GetCalls())
}

func TestMockClientNotConnected(t *testing.T) {
	mock := NewMockClient()
	// Not connected

	err := mock.Ping()
	assert.Equal(t, ErrNotConnected, err)

	_, err = mock.GetVersion()
	assert.Equal(t, ErrNotConnected, err)

	_, err = mock.GetDatabases()
	assert.Equal(t, ErrNotConnected, err)

	_, err = mock.ExecuteQuery("SELECT 1")
	assert.Equal(t, ErrNotConnected, err)

	_, err = mock.Execute("INSERT INTO users VALUES (1)")
	assert.Equal(t, ErrNotConnected, err)
}

func TestMockResult(t *testing.T) {
	result := &MockResult{
		LastID:   42,
		Affected: 5,
	}

	id, err := result.LastInsertId()
	assert.NoError(t, err)
	assert.Equal(t, int64(42), id)

	affected, err := result.RowsAffected()
	assert.NoError(t, err)
	assert.Equal(t, int64(5), affected)
}

// --- Additional Coverage Tests ---

func TestClientDB(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	config := NewConfig().WithHost("localhost").WithUser("root")
	client, _ := NewClientWithDB(config, db)

	assert.Equal(t, db, client.DB())
}

func TestClientExecuteQueryArgs(t *testing.T) {
	t.Run("successful query with args", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Alice")
		mock.ExpectQuery("SELECT").WithArgs("Alice").WillReturnRows(rows)

		config := NewConfig().WithHost("localhost").WithUser("root").WithTimeout(5 * time.Second)
		client, _ := NewClientWithDB(config, db)

		result, err := client.ExecuteQueryArgs("SELECT * FROM users WHERE name = ?", "Alice")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		result.Close()
	})

	t.Run("not connected", func(t *testing.T) {
		config := NewConfig().WithHost("localhost").WithUser("root")
		client, _ := NewClient(config)

		_, err := client.ExecuteQueryArgs("SELECT * FROM users WHERE id = ?", 1)
		assert.Error(t, err)
		assert.Equal(t, ErrNotConnected, err)
	})

	t.Run("query error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery("SELECT").WithArgs("test").WillReturnError(errors.New("query failed"))

		config := NewConfig().WithHost("localhost").WithUser("root").WithTimeout(5 * time.Second)
		client, _ := NewClientWithDB(config, db)

		_, err = client.ExecuteQueryArgs("SELECT * FROM users WHERE name = ?", "test")
		assert.Error(t, err)
		assert.True(t, IsQueryError(err))
	})
}

func TestMockClientExecuteQueryArgs(t *testing.T) {
	mock := NewMockClient()
	mock.SetConnected(true)

	// Test when connected
	_, err := mock.ExecuteQueryArgs("SELECT * FROM users WHERE id = ?", 1)
	assert.NoError(t, err)

	// Verify call tracking
	assert.Equal(t, 1, mock.GetCallCount("ExecuteQueryArgs"))
}

func TestMockClientErrors(t *testing.T) {
	t.Run("version error", func(t *testing.T) {
		mock := NewMockClient()
		mock.SetConnected(true)
		mock.VersionErr = errors.New("version error")

		_, err := mock.GetVersion()
		assert.Error(t, err)
	})

	t.Run("databases error", func(t *testing.T) {
		mock := NewMockClient()
		mock.SetConnected(true)
		mock.DatabasesErr = errors.New("databases error")

		_, err := mock.GetDatabases()
		assert.Error(t, err)
	})

	t.Run("tables error", func(t *testing.T) {
		mock := NewMockClient()
		mock.SetConnected(true)
		mock.TablesErr = errors.New("tables error")

		_, err := mock.GetTables("testdb")
		assert.Error(t, err)
	})

	t.Run("table size error", func(t *testing.T) {
		mock := NewMockClient()
		mock.SetConnected(true)
		mock.TableSizeErr = errors.New("table size error")

		_, err := mock.GetTableSize("testdb", "users")
		assert.Error(t, err)
	})

	t.Run("row count error", func(t *testing.T) {
		mock := NewMockClient()
		mock.SetConnected(true)
		mock.RowCountErr = errors.New("row count error")

		_, err := mock.GetTableRowCount("testdb", "users")
		assert.Error(t, err)
	})

	t.Run("database size error", func(t *testing.T) {
		mock := NewMockClient()
		mock.SetConnected(true)
		mock.DBSizeErr = errors.New("db size error")

		_, err := mock.GetDatabaseSize("testdb")
		assert.Error(t, err)
	})

	t.Run("table info error", func(t *testing.T) {
		mock := NewMockClient()
		mock.SetConnected(true)
		mock.TableInfoErr = errors.New("table info error")

		_, err := mock.GetTableInfo("testdb", "users")
		assert.Error(t, err)
	})

	t.Run("database info error", func(t *testing.T) {
		mock := NewMockClient()
		mock.SetConnected(true)
		mock.DBInfoErr = errors.New("db info error")

		_, err := mock.GetDatabaseInfo("testdb")
		assert.Error(t, err)
	})

	t.Run("query error", func(t *testing.T) {
		mock := NewMockClient()
		mock.SetConnected(true)
		mock.QueryErr = errors.New("query error")

		_, err := mock.ExecuteQuery("SELECT 1")
		assert.Error(t, err)
	})

	t.Run("exec error", func(t *testing.T) {
		mock := NewMockClient()
		mock.SetConnected(true)
		mock.ExecErr = errors.New("exec error")

		_, err := mock.Execute("INSERT INTO users VALUES (1)")
		assert.Error(t, err)
	})

	t.Run("close error", func(t *testing.T) {
		mock := NewMockClient()
		mock.SetConnected(true)
		mock.CloseErr = errors.New("close error")

		err := mock.Close()
		assert.Error(t, err)
	})
}

func TestTimeoutErrorWithoutUnderlying(t *testing.T) {
	err := &TimeoutError{
		Operation: "connect",
		Duration:  "10s",
	}

	assert.Contains(t, err.Error(), "connect")
	assert.Contains(t, err.Error(), "10s")
	assert.Nil(t, err.Unwrap())
}

func TestQueryErrorWithoutUnderlying(t *testing.T) {
	err := &QueryError{
		Query:   "SELECT 1",
		Message: "error",
	}

	assert.Contains(t, err.Error(), "SELECT 1")
	assert.Nil(t, err.Unwrap())
}

func TestClientGetTableRowCountError(t *testing.T) {
	t.Run("table not found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery("SELECT COALESCE").
			WithArgs("testdb", "nonexistent").
			WillReturnError(sql.ErrNoRows)

		config := NewConfig().WithHost("localhost").WithUser("root").WithTimeout(5 * time.Second)
		client, _ := NewClientWithDB(config, db)

		_, err = client.GetTableRowCount("testdb", "nonexistent")
		assert.Error(t, err)
		assert.Equal(t, ErrEmptyResult, err)
	})

	t.Run("empty database name", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		config := NewConfig().WithHost("localhost").WithUser("root").WithTimeout(5 * time.Second)
		client, _ := NewClientWithDB(config, db)

		_, err = client.GetTableRowCount("", "users")
		assert.Error(t, err)
		assert.True(t, IsConfigError(err))
	})

	t.Run("empty table name", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		config := NewConfig().WithHost("localhost").WithUser("root").WithTimeout(5 * time.Second)
		client, _ := NewClientWithDB(config, db)

		_, err = client.GetTableRowCount("testdb", "")
		assert.Error(t, err)
		assert.True(t, IsConfigError(err))
	})
}

func TestClientGetTableInfoValidation(t *testing.T) {
	t.Run("empty database name", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		config := NewConfig().WithHost("localhost").WithUser("root").WithTimeout(5 * time.Second)
		client, _ := NewClientWithDB(config, db)

		_, err = client.GetTableInfo("", "users")
		assert.Error(t, err)
		assert.True(t, IsConfigError(err))
	})

	t.Run("empty table name", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		config := NewConfig().WithHost("localhost").WithUser("root").WithTimeout(5 * time.Second)
		client, _ := NewClientWithDB(config, db)

		_, err = client.GetTableInfo("testdb", "")
		assert.Error(t, err)
		assert.True(t, IsConfigError(err))
	})

	t.Run("not connected", func(t *testing.T) {
		config := NewConfig().WithHost("localhost").WithUser("root")
		client, _ := NewClient(config)

		_, err := client.GetTableInfo("testdb", "users")
		assert.Error(t, err)
		assert.Equal(t, ErrNotConnected, err)
	})
}

func TestClientGetDatabaseInfoValidation(t *testing.T) {
	t.Run("empty database name", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		config := NewConfig().WithHost("localhost").WithUser("root").WithTimeout(5 * time.Second)
		client, _ := NewClientWithDB(config, db)

		_, err = client.GetDatabaseInfo("")
		assert.Error(t, err)
		assert.True(t, IsConfigError(err))
	})

	t.Run("not connected", func(t *testing.T) {
		config := NewConfig().WithHost("localhost").WithUser("root")
		client, _ := NewClient(config)

		_, err := client.GetDatabaseInfo("testdb")
		assert.Error(t, err)
		assert.Equal(t, ErrNotConnected, err)
	})
}
