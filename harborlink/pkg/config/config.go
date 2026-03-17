package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	Carriers  []CarrierConfig `mapstructure:"carriers"`
	Logging   LoggingConfig   `mapstructure:"logging"`
	SlotWatch SlotWatchConfig `mapstructure:"slotwatch"`
}

// SlotWatchConfig represents slot watch configuration
type SlotWatchConfig struct {
	NotifyTimeout   time.Duration `mapstructure:"notify_timeout"`   // NOTIFY_CONFIRM超时时间
	LockTimeout     time.Duration `mapstructure:"lock_timeout"`     // 锁定操作超时
	DefaultPriority int           `mapstructure:"default_priority"` // 默认优先级
}

// ServerConfig represents HTTP server configuration
type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	Host         string        `mapstructure:"host"`
	Mode         string        `mapstructure:"mode"` // debug, release, test
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

// DatabaseConfig represents database connection configuration
type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Name            string        `mapstructure:"name"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// RedisConfig represents Redis connection configuration
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// CarrierConfig represents a carrier adapter configuration
type CarrierConfig struct {
	Name      string `mapstructure:"name"`
	Code      string `mapstructure:"code"`
	Adapter   string `mapstructure:"adapter"`
	Enabled   bool   `mapstructure:"enabled"`
	BaseURL   string `mapstructure:"base_url"`
	APIKey    string `mapstructure:"api_key"`
	RateLimit int    `mapstructure:"rate_limit"` // requests per minute

	// Polling configuration for slot monitoring
	PollEnabled   bool          `mapstructure:"poll_enabled"`
	PollInterval  time.Duration `mapstructure:"poll_interval"` // polling interval
	PollTimeout   time.Duration `mapstructure:"poll_timeout"`  // single poll timeout
	BurstLimit    int           `mapstructure:"burst_limit"`   // burst request limit
	MaxRetries    int           `mapstructure:"max_retries"`
	RetryInterval time.Duration `mapstructure:"retry_interval"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level  string `mapstructure:"level"`  // debug, info, warn, error
	Format string `mapstructure:"format"` // json, text
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Configure viper
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	if configPath != "" {
		v.AddConfigPath(configPath)
	} else {
		// Default config paths
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
		v.AddConfigPath("/etc/harborlink")
	}

	// Enable environment variable override
	v.SetEnvPrefix("HARBORLINK")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found, will use defaults and env vars
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &cfg, nil
}

// LoadFromPath loads configuration from a specific path
func LoadFromPath(path string) (*Config, error) {
	v := viper.New()

	setDefaults(v)

	v.SetConfigFile(path)
	v.SetEnvPrefix("HARBORLINK")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.mode", "debug")
	v.SetDefault("server.read_timeout", "10s")
	v.SetDefault("server.write_timeout", "10s")
	v.SetDefault("server.idle_timeout", "60s")

	// Database defaults
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.name", "harborlink")
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.password", "")
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("database.conn_max_lifetime", "5m")

	// Redis defaults
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")

	// Slot watch defaults
	v.SetDefault("slotwatch.notify_timeout", "30s")
	v.SetDefault("slotwatch.lock_timeout", "10s")
	v.SetDefault("slotwatch.default_priority", 5)
}

// DSN returns the database connection string
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

// Addr returns the Redis address
func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// Addr returns the server address
func (c *ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
