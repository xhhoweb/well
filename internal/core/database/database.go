package database

import (
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"well_go/internal/core/config"
	"well_go/internal/core/logger"
)

var db *sqlx.DB

// Init Initialize database connection
func Init(cfg *config.DatabaseConfig) error {
	var err error

	db, err = sqlx.Connect("mysql", cfg.GetDSN())
	if err != nil {
		logger.Error("failed to connect database", logger.String("error", err.Error()))
		return err
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	logger.Info("database initialized successfully",
		logger.String("host", cfg.Host),
		logger.Int("port", cfg.Port),
		logger.String("database", cfg.Name))

	return nil
}

// Get Get database instance
func Get() *sqlx.DB {
	return db
}

// Close Close database connection
func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// Ping Check database connection
func Ping() error {
	if db == nil {
		return nil
	}
	return db.Ping()
}
