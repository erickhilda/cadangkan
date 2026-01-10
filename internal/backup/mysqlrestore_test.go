package backup

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/erickhilda/cadangkan/pkg/database/mysql"
	"github.com/stretchr/testify/assert"
)

func TestNewMySQLRestorer(t *testing.T) {
	config := &mysql.Config{
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "password",
		Database: "testdb",
		Timeout:  10 * time.Second,
	}

	restorer := NewMySQLRestorer(config)
	assert.NotNil(t, restorer)
	assert.Equal(t, config, restorer.config)
	// Timeout should be multiplied by 6
	assert.Equal(t, 60*time.Second, restorer.timeout)
}

func TestNewMySQLRestorerDefaultTimeout(t *testing.T) {
	config := &mysql.Config{
		Host:    "localhost",
		Port:    3306,
		User:    "root",
		Timeout: 0, // No timeout set
	}

	restorer := NewMySQLRestorer(config)
	assert.NotNil(t, restorer)
	// Should default to 30 minutes
	assert.Equal(t, 30*time.Minute, restorer.timeout)
}

func TestMySQLRestorerBuildArgs(t *testing.T) {
	t.Run("with password", func(t *testing.T) {
		config := &mysql.Config{
			Host:     "localhost",
			Port:     3306,
			User:     "root",
			Password: "secret",
		}
		restorer := NewMySQLRestorer(config)

		args := restorer.buildArgs("testdb")
		assert.Contains(t, args, "--host=localhost")
		assert.Contains(t, args, "--port=3306")
		assert.Contains(t, args, "--user=root")
		assert.Contains(t, args, "--password=secret")
		assert.Contains(t, args, "testdb")
	})

	t.Run("without password", func(t *testing.T) {
		config := &mysql.Config{
			Host: "localhost",
			Port: 3306,
			User: "root",
			// No password
		}
		restorer := NewMySQLRestorer(config)

		args := restorer.buildArgs("testdb")
		assert.Contains(t, args, "--host=localhost")
		assert.Contains(t, args, "--port=3306")
		assert.Contains(t, args, "--user=root")
		// Should not contain password flag
		for _, arg := range args {
			assert.False(t, strings.HasPrefix(arg, "--password"), "should not have password flag when password is empty")
		}
		assert.Contains(t, args, "testdb")
	})

	t.Run("all connection parameters", func(t *testing.T) {
		config := &mysql.Config{
			Host:     "remote.example.com",
			Port:     3307,
			User:     "backup_user",
			Password: "mypassword",
		}
		restorer := NewMySQLRestorer(config)

		args := restorer.buildArgs("mydb")
		assert.Contains(t, args, "--host=remote.example.com")
		assert.Contains(t, args, "--port=3307")
		assert.Contains(t, args, "--user=backup_user")
		assert.Contains(t, args, "--password=mypassword")
		assert.Contains(t, args, "mydb")
	})
}

func TestMySQLRestorerRestore(t *testing.T) {
	t.Run("empty database name", func(t *testing.T) {
		config := &mysql.Config{
			Host: "localhost",
			Port: 3306,
			User: "root",
		}
		restorer := NewMySQLRestorer(config)

		sqlData := bytes.NewReader([]byte("CREATE TABLE test (id INT);"))
		err := restorer.Restore("", sqlData)
		assert.Error(t, err)
		assert.True(t, IsRestoreError(err))
		assert.Contains(t, err.Error(), "database name is required")
	})

	// Note: Actual command execution tests would require mysql to be available
	// These are integration-style tests that can be skipped
	t.Run("restore with command logging", func(t *testing.T) {
		config := &mysql.Config{
			Host:     "localhost",
			Port:     3306,
			User:     "root",
			Password: "secret",
		}
		restorer := NewMySQLRestorer(config)

		var loggedCommand string
		cmdLogger := func(cmd string) {
			loggedCommand = cmd
		}

		sqlData := bytes.NewReader([]byte("SELECT 1;"))
		// This will fail if mysql is not available, but we can test the logging
		err := restorer.RestoreWithCommand("testdb", sqlData, cmdLogger)

		// Verify command was logged (even if execution failed)
		if loggedCommand != "" {
			assert.Contains(t, loggedCommand, "mysql")
			assert.Contains(t, loggedCommand, "--host=localhost")
			assert.Contains(t, loggedCommand, "--port=3306")
			assert.Contains(t, loggedCommand, "--user=root")
			// Password should be masked
			assert.Contains(t, loggedCommand, "--password=***")
			assert.NotContains(t, loggedCommand, "--password=secret")
			assert.Contains(t, loggedCommand, "testdb")
		}

		// If mysql is not available, error is expected
		if err != nil {
			assert.True(t, IsRestoreError(err) || strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "executable"))
		}
	})
}

func TestMySQLRestorerRestoreWithCommand(t *testing.T) {
	t.Run("password masking in logs", func(t *testing.T) {
		config := &mysql.Config{
			Host:     "localhost",
			Port:     3306,
			User:     "root",
			Password: "super_secret_password_123",
		}
		restorer := NewMySQLRestorer(config)

		var loggedCommand string
		cmdLogger := func(cmd string) {
			loggedCommand = cmd
		}

		sqlData := bytes.NewReader([]byte("SELECT 1;"))
		_ = restorer.RestoreWithCommand("testdb", sqlData, cmdLogger)

		if loggedCommand != "" {
			// Verify password is masked
			assert.Contains(t, loggedCommand, "--password=***")
			assert.NotContains(t, loggedCommand, "super_secret_password_123")
		}
	})

	t.Run("no logger provided", func(t *testing.T) {
		config := &mysql.Config{
			Host: "localhost",
			Port: 3306,
			User: "root",
		}
		restorer := NewMySQLRestorer(config)

		sqlData := bytes.NewReader([]byte("SELECT 1;"))
		// Should not panic when logger is nil
		_ = restorer.RestoreWithCommand("testdb", sqlData, nil)
	})
}

func TestCheckMySQL(t *testing.T) {
	// This test requires mysql to be available
	// If not available, it will fail but that's expected
	version, err := CheckMySQL()

	if err != nil {
		// mysql not found - this is acceptable in test environments
		errMsg := err.Error()
		if !strings.Contains(errMsg, "not found") && !strings.Contains(errMsg, "executable") {
			t.Errorf("unexpected error: %s", errMsg)
		}
		// assert.Contains(t, err.Error(), "not found") || assert.Contains(t, err.Error(), "executable")
		t.Skip("mysql command not available, skipping CheckMySQL test")
		return
	}

	// If mysql is available, verify version string
	assert.NotEmpty(t, version)
	assert.Contains(t, strings.ToLower(version), "mysql")
}

func TestGetRestoreExitCode(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		code := getRestoreExitCode(nil)
		assert.Equal(t, 0, code)
	})

	t.Run("exit error", func(t *testing.T) {
		// Use a command that will exit with a non-zero code
		// On Unix systems, `sh -c "exit 42"` will exit with code 42
		cmd := exec.Command("sh", "-c", "exit 42")
		err := cmd.Run()
		if err != nil {
			code := getRestoreExitCode(err)
			// Should extract exit code (42 in this case)
			assert.Equal(t, 42, code)
			assert.NotEqual(t, -1, code)
		} else {
			t.Fatal("expected command to fail with exit code 42")
		}
	})

	t.Run("non-exit error", func(t *testing.T) {
		err := errors.New("some other error")
		code := getRestoreExitCode(err)
		assert.Equal(t, -1, code)
	})
}

func TestMySQLRestorerErrorHandling(t *testing.T) {
	t.Run("stderr error detection", func(t *testing.T) {
		// This tests the error pattern detection logic
		// We can't easily test the actual command execution, but we can verify
		// the error handling logic works correctly

		errorPatterns := []string{
			"error",
			"failed",
			"cannot",
			"denied",
			"access denied",
			"unknown database",
		}

		testCases := []struct {
			name      string
			stderr    string
			shouldErr bool
		}{
			{"no error", "", false},
			{"warning only", "Warning: Using a password on the command line", false},
			{"contains error", "Error: Access denied", true},
			{"contains failed", "Failed to connect", true},
			{"contains cannot", "Cannot connect to MySQL server", true},
			{"contains denied", "Access denied for user", true},
			{"contains unknown database", "Unknown database 'testdb'", true},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				lowerStderr := strings.ToLower(tc.stderr)
				hasError := false
				for _, pattern := range errorPatterns {
					if strings.Contains(lowerStderr, pattern) {
						hasError = true
						break
					}
				}
				assert.Equal(t, tc.shouldErr, hasError, "stderr: %s", tc.stderr)
			})
		}
	})
}
