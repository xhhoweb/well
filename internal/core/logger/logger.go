package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"well_go/internal/core/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

// Init Initialize logger
func Init(cfg *config.LoggingConfig) error {
	// Create log directory if not exists
	logDir := filepath.Dir(cfg.Filename)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	// Set log level
	var level zapcore.Level
	switch cfg.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	// Output to stdout only
	writers := []zapcore.WriteSyncer{
		zapcore.AddSync(os.Stdout),
	}

	// Create encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create encoder
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	// Create core
	core := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(writers...), level)

	// Create logger with caller
	log = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return nil
}

// Get Get logger instance
func Get() *zap.Logger {
	return log
}

// Sync Flush logger buffer
func Sync() {
	if log != nil {
		log.Sync()
	}
}

// Debug Log debug message
func Debug(msg string, fields ...zap.Field) {
	log.Debug(msg, fields...)
}

// Info Log info message
func Info(msg string, fields ...zap.Field) {
	log.Info(msg, fields...)
}

// Warn Log warning message
func Warn(msg string, fields ...zap.Field) {
	log.Warn(msg, fields...)
}

// Error Log error message
func Error(msg string, fields ...zap.Field) {
	log.Error(msg, fields...)
}

// Fatal Log fatal message
func Fatal(msg string, fields ...zap.Field) {
	log.Fatal(msg, fields...)
}

// String String field
func String(key, value string) zap.Field {
	return zap.String(key, value)
}

// Int Int field
func Int(key string, value int) zap.Field {
	return zap.Int(key, value)
}

// Int64 Int64 field
func Int64(key string, value int64) zap.Field {
	return zap.Int64(key, value)
}

// Duration Duration field
func Duration(key string, value time.Duration) zap.Field {
	return zap.Duration(key, value)
}

// ErrorField Error field
func ErrorField(err error) zap.Field {
	return zap.Error(err)
}

// Now Format current time
func Now() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// Print simple print to stdout
func Print(msg string) {
	fmt.Printf("[%s] %s\n", Now(), msg)
}
