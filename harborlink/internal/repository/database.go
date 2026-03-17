package repository

import (
	"context"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/yourname/harborlink/pkg/config"
)

// Database wraps the GORM DB connection
type Database struct {
	*gorm.DB
}

// NewDatabase creates a new database connection
func NewDatabase(cfg *config.DatabaseConfig) (*Database, error) {
	// Configure GORM logger
	var gormLogger logger.Interface
	gormLogger = logger.Default.LogMode(logger.Silent)

	var db *gorm.DB
	var err error

	// Use SQLite if host is "sqlite" or empty
	if cfg.Host == "sqlite" || cfg.Host == "" {
		// Use SQLite in-memory database for testing/development
		db, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
			Logger: gormLogger,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to sqlite database: %w", err)
		}
		log.Printf("Connected to SQLite in-memory database")
	} else {
		// Use PostgreSQL
		dsn := cfg.DSN()
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: gormLogger,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to database: %w", err)
		}
	}

	// Get underlying sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Connected to database: %s:%d/%s", cfg.Host, cfg.Port, cfg.Name)

	return &Database{db}, nil
}

// Close closes the database connection
func (db *Database) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Ping tests the database connection
func (db *Database) Ping() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// Migrate runs auto migration for given models
func (db *Database) Migrate(models ...interface{}) error {
	return db.DB.AutoMigrate(models...)
}

// Transaction executes a function within a database transaction
func (db *Database) Transaction(fn func(tx *gorm.DB) error) error {
	return db.DB.Transaction(fn)
}

// WithContext returns a new database instance with the given context
func (db *Database) WithContext(ctx context.Context) *Database {
	return &Database{db.DB.WithContext(ctx)}
}
