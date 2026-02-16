package postgres

import (
	"context"
	"fmt"
	"time"

	"superaib/internal/core/config"
	"superaib/internal/core/logger"
	"superaib/internal/models"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// DB wraps *gorm.DB
type DB struct {
	*gorm.DB
}

// NewDB creates a new database connection
func NewDB(cfg *config.Config) (*DB, error) {
	dsn := cfg.DatabaseURL
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: NewGormLogger(logger.Log), // custom Logrus logger
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	logger.Log.Info("✅ Database connection established successfully")
	return &DB{db}, nil
}

// AutoMigrate runs GORM auto migrations
func (db *DB) AutoMigrate() error {
	logger.Log.Info("🔄 Running GORM auto-migrations...")
	return db.Migrator().AutoMigrate(
		&models.User{}, // Add more models here as needed
	)
}

// Close closes the database connection
func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}
	logger.Log.Info("🛑 Closing database connection")
	return sqlDB.Close()
}

// -------------------------
// GORM Logger Adapter for Logrus
// -------------------------
type GormLogger struct {
	log *logrus.Logger
}

// NewGormLogger returns a gormlogger.Interface compatible with GORM v2
func NewGormLogger(l *logrus.Logger) gormlogger.Interface {
	return &GormLogger{log: l}
}

// LogMode implements gormlogger.Interface
func (l *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	// Optional: map levels if needed
	return l
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.log.Infof(msg, data...)
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.log.Warnf(msg, data...)
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.log.Errorf(msg, data...)
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	sql, rows := fc()
	elapsed := time.Since(begin)

	entry := l.log.WithFields(logrus.Fields{
		"duration_ms": elapsed.Milliseconds(),
		"rows":        rows,
		"sql":         sql,
		"error":       err,
	})

	if err != nil {
		entry.Error("GORM SQL Error")
	} else {
		entry.Debug("GORM SQL Query")
	}
}
