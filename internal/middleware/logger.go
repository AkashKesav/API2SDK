package middleware

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
)

// LogLevel represents different log levels
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// String returns the string representation of log level
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// LoggerConfig defines the config for Logger middleware
type LoggerConfig struct {
	Level         LogLevel
	Format        string
	TimeFormat    string
	Output        *os.File
	DisableColors bool
}

// DefaultLoggerConfig is the default Logger configuration
var DefaultLoggerConfig = LoggerConfig{
	Level:         INFO,
	Format:        "[${time}] ${status} - ${method} ${path} - ${ip} - ${latency}\n",
	TimeFormat:    "2006/01/02 15:04:05",
	Output:        os.Stdout,
	DisableColors: false,
}

// Logger represents a custom logger
type Logger struct {
	config LoggerConfig
	logger *log.Logger
}

// NewLogger creates a new logger instance
func NewLogger(config ...LoggerConfig) *Logger {
	cfg := DefaultLoggerConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	return &Logger{
		config: cfg,
		logger: log.New(cfg.Output, "", 0),
	}
}

// log writes a log message with the given level
func (l *Logger) log(level LogLevel, message string, args ...interface{}) {
	if level < l.config.Level {
		return
	}

	timestamp := time.Now().Format(l.config.TimeFormat)
	levelStr := level.String()

	// Add colors if enabled
	if !l.config.DisableColors && l.config.Output == os.Stdout {
		switch level {
		case DEBUG:
			levelStr = "\033[36m" + levelStr + "\033[0m" // Cyan
		case INFO:
			levelStr = "\033[32m" + levelStr + "\033[0m" // Green
		case WARN:
			levelStr = "\033[33m" + levelStr + "\033[0m" // Yellow
		case ERROR:
			levelStr = "\033[31m" + levelStr + "\033[0m" // Red
		}
	}

	formattedMessage := fmt.Sprintf(message, args...)
	logLine := fmt.Sprintf("[%s] %s: %s", timestamp, levelStr, formattedMessage)
	l.logger.Println(logLine)
}

// Debug logs a debug message
func (l *Logger) Debug(message string, args ...interface{}) {
	l.log(DEBUG, message, args...)
}

// Info logs an info message
func (l *Logger) Info(message string, args ...interface{}) {
	l.log(INFO, message, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, args ...interface{}) {
	l.log(WARN, message, args...)
}

// Error logs an error message
func (l *Logger) Error(message string, args ...interface{}) {
	l.log(ERROR, message, args...)
}

// Global logger instance
var globalLogger = NewLogger()

// SetGlobalLogger sets the global logger instance
func SetGlobalLogger(logger *Logger) {
	globalLogger = logger
}

// Debug logs a debug message using the global logger
func Debug(message string, args ...interface{}) {
	globalLogger.Debug(message, args...)
}

// Info logs an info message using the global logger
func Info(message string, args ...interface{}) {
	globalLogger.Info(message, args...)
}

// Warn logs a warning message using the global logger
func Warn(message string, args ...interface{}) {
	globalLogger.Warn(message, args...)
}

// Error logs an error message using the global logger
func Error(message string, args ...interface{}) {
	globalLogger.Error(message, args...)
}

// LoggerMiddleware creates a logging middleware
func LoggerMiddleware(config ...LoggerConfig) fiber.Handler {
	cfg := DefaultLoggerConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	logger := NewLogger(cfg)

	return func(c fiber.Ctx) error {
		// Start timer
		start := time.Now()

		// Process request
		err := c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get status code
		status := c.Response().StatusCode()

		// Get request info
		method := c.Method()
		path := c.Path()
		ip := c.IP()
		userAgent := c.Get("User-Agent")

		// Format log message
		timestamp := time.Now().Format(cfg.TimeFormat)

		// Replace variables in format
		logMessage := cfg.Format
		logMessage = replaceVar(logMessage, "time", timestamp)
		logMessage = replaceVar(logMessage, "status", fmt.Sprintf("%d", status))
		logMessage = replaceVar(logMessage, "method", method)
		logMessage = replaceVar(logMessage, "path", path)
		logMessage = replaceVar(logMessage, "ip", ip)
		logMessage = replaceVar(logMessage, "latency", latency.String())
		logMessage = replaceVar(logMessage, "user_agent", userAgent)

		// Determine log level based on status code
		var level LogLevel
		switch {
		case status >= 500:
			level = ERROR
		case status >= 400:
			level = WARN
		case status >= 300:
			level = INFO
		default:
			level = DEBUG
		}

		// Add status code coloring
		if !cfg.DisableColors && cfg.Output == os.Stdout {
			statusStr := fmt.Sprintf("%d", status)
			switch {
			case status >= 500:
				statusStr = "\033[31m" + statusStr + "\033[0m" // Red
			case status >= 400:
				statusStr = "\033[33m" + statusStr + "\033[0m" // Yellow
			case status >= 300:
				statusStr = "\033[36m" + statusStr + "\033[0m" // Cyan
			case status >= 200:
				statusStr = "\033[32m" + statusStr + "\033[0m" // Green
			}
			logMessage = replaceVar(logMessage, "status", statusStr)
		}

		// Write log
		if level >= cfg.Level {
			logger.logger.Print(logMessage)
		}

		return err
	}
}

// replaceVar replaces a variable placeholder in the format string
func replaceVar(format, key, value string) string {
	placeholder := "${" + key + "}"
	return fmt.Sprintf("%s", replaceString(format, placeholder, value))
}

// replaceString replaces all occurrences of old with new in s
func replaceString(s, old, new string) string {
	result := ""
	for i := 0; i < len(s); {
		if i <= len(s)-len(old) && s[i:i+len(old)] == old {
			result += new
			i += len(old)
		} else {
			result += string(s[i])
			i++
		}
	}
	return result
}

// RequestLoggerMiddleware creates a request logging middleware with detailed information
func RequestLoggerMiddleware() fiber.Handler {
	return LoggerMiddleware(LoggerConfig{
		Level:         DEBUG,
		Format:        "[${time}] ${status} - ${method} ${path} - ${ip} - ${latency} - \"${user_agent}\"\n",
		TimeFormat:    "2006/01/02 15:04:05",
		Output:        os.Stdout,
		DisableColors: false,
	})
}

// ProductionLoggerMiddleware creates a production logging middleware
func ProductionLoggerMiddleware() fiber.Handler {
	return LoggerMiddleware(LoggerConfig{
		Level:         INFO,
		Format:        "[${time}] ${status} - ${method} ${path} - ${ip} - ${latency}\n",
		TimeFormat:    time.RFC3339,
		Output:        os.Stdout,
		DisableColors: true,
	})
}

// ErrorLoggerMiddleware creates an error-only logging middleware
func ErrorLoggerMiddleware() fiber.Handler {
	return LoggerMiddleware(LoggerConfig{
		Level:         ERROR,
		Format:        "[${time}] ERROR ${status} - ${method} ${path} - ${ip} - ${latency}\n",
		TimeFormat:    time.RFC3339,
		Output:        os.Stderr,
		DisableColors: true,
	})
}
