package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	// Test loading from project root (go up from pkg/config to project root)
	projectRoot := filepath.Join("..", "..")
	cfg, err := Load(projectRoot)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Verify config.yaml values are loaded
	if cfg.Server.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Server.Port)
	}

	if cfg.Server.Mode != "debug" {
		t.Errorf("expected mode 'debug', got %s", cfg.Server.Mode)
	}

	if cfg.Database.Host != "localhost" {
		t.Errorf("expected database host 'localhost', got %s", cfg.Database.Host)
	}

	if cfg.Redis.Host != "localhost" {
		t.Errorf("expected redis host 'localhost', got %s", cfg.Redis.Host)
	}

	// Verify carriers are loaded from config.yaml
	if len(cfg.Carriers) == 0 {
		t.Error("expected at least one carrier configured")
	}

	// Check first carrier (maersk)
	if len(cfg.Carriers) > 0 {
		if cfg.Carriers[0].Name != "maersk" {
			t.Errorf("expected first carrier name 'maersk', got %s", cfg.Carriers[0].Name)
		}

		if cfg.Carriers[0].Code != "MAEU" {
			t.Errorf("expected carrier code 'MAEU', got %s", cfg.Carriers[0].Code)
		}

		if !cfg.Carriers[0].Enabled {
			t.Error("expected maersk carrier to be enabled")
		}
	}
}

func TestLoadDefaults(t *testing.T) {
	// Test loading without config file - should use defaults
	cfg, err := Load(t.TempDir()) // Empty temp dir, no config file
	if err != nil {
		t.Fatalf("failed to load config with defaults: %v", err)
	}

	// Verify default values
	if cfg.Server.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.Server.Port)
	}

	if cfg.Server.Mode != "debug" {
		t.Errorf("expected default mode 'debug', got %s", cfg.Server.Mode)
	}

	if cfg.Database.Host != "localhost" {
		t.Errorf("expected default database host 'localhost', got %s", cfg.Database.Host)
	}

	if cfg.Redis.Host != "localhost" {
		t.Errorf("expected default redis host 'localhost', got %s", cfg.Redis.Host)
	}
}

func TestLoadFromPath(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_config.yaml")

	configContent := `
server:
  port: 9090
  host: "127.0.0.1"
  mode: release

database:
  host: db.example.com
  port: 5433
  name: testdb

redis:
  host: redis.example.com

carriers: []

logging:
  level: debug
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadFromPath(configPath)
	if err != nil {
		t.Fatalf("failed to load config from path: %v", err)
	}

	// Verify loaded values
	if cfg.Server.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Server.Port)
	}

	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("expected host '127.0.0.1', got %s", cfg.Server.Host)
	}

	if cfg.Server.Mode != "release" {
		t.Errorf("expected mode 'release', got %s", cfg.Server.Mode)
	}

	if cfg.Database.Host != "db.example.com" {
		t.Errorf("expected database host 'db.example.com', got %s", cfg.Database.Host)
	}

	if cfg.Database.Port != 5433 {
		t.Errorf("expected database port 5433, got %d", cfg.Database.Port)
	}

	if cfg.Redis.Host != "redis.example.com" {
		t.Errorf("expected redis host 'redis.example.com', got %s", cfg.Redis.Host)
	}

	if cfg.Logging.Level != "debug" {
		t.Errorf("expected logging level 'debug', got %s", cfg.Logging.Level)
	}
}

func TestEnvironmentVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("HARBORLINK_SERVER_PORT", "3000")
	os.Setenv("HARBORLINK_DATABASE_HOST", "env-db.example.com")
	os.Setenv("HARBORLINK_REDIS_PORT", "6380")
	defer func() {
		os.Unsetenv("HARBORLINK_SERVER_PORT")
		os.Unsetenv("HARBORLINK_DATABASE_HOST")
		os.Unsetenv("HARBORLINK_REDIS_PORT")
	}()

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Verify environment variables override config
	if cfg.Server.Port != 3000 {
		t.Errorf("expected port 3000 from env, got %d", cfg.Server.Port)
	}

	if cfg.Database.Host != "env-db.example.com" {
		t.Errorf("expected database host 'env-db.example.com' from env, got %s", cfg.Database.Host)
	}

	if cfg.Redis.Port != 6380 {
		t.Errorf("expected redis port 6380 from env, got %d", cfg.Redis.Port)
	}
}

func TestDatabaseDSN(t *testing.T) {
	cfg := &DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "secret",
		Name:     "harborlink",
		SSLMode:  "disable",
	}

	dsn := cfg.DSN()
	expected := "host=localhost port=5432 user=postgres password=secret dbname=harborlink sslmode=disable"
	if dsn != expected {
		t.Errorf("expected DSN '%s', got '%s'", expected, dsn)
	}
}

func TestRedisAddr(t *testing.T) {
	cfg := &RedisConfig{
		Host: "127.0.0.1",
		Port: 6379,
	}

	addr := cfg.Addr()
	expected := "127.0.0.1:6379"
	if addr != expected {
		t.Errorf("expected addr '%s', got '%s'", expected, addr)
	}
}

func TestServerAddr(t *testing.T) {
	cfg := &ServerConfig{
		Host: "0.0.0.0",
		Port: 8080,
	}

	addr := cfg.Addr()
	expected := "0.0.0.0:8080"
	if addr != expected {
		t.Errorf("expected addr '%s', got '%s'", expected, addr)
	}
}
