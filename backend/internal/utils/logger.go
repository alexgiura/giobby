package utils

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// LogLevel represents the logging level
type LogLevel string

const (
	DEBUG LogLevel = "DEBUG"
	INFO  LogLevel = "INFO"
	WARN  LogLevel = "WARN"
	ERROR LogLevel = "ERROR"
)

// LogEntry represents a single log entry
type LogEntry struct {
	ID        string    `json:"id"`
	Timestamp string    `json:"timestamp"`
	Method    string    `json:"method"`
	Path      string    `json:"path"`
	Status    int       `json:"status"`
	Duration  int64     `json:"duration"`
	Level     LogLevel  `json:"level"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"-"`
}

// LogsResponse represents the response for logs API
type LogsResponse struct {
	Items []LogEntry `json:"items"`
}

// LogLevelUpdate represents the request for updating log level
type LogLevelUpdate struct {
	Level LogLevel `json:"level"`
}

// Logger is the main logger instance
type Logger struct {
	mu       sync.RWMutex
	entries  []LogEntry
	maxSize  int
	level    LogLevel
	service  string
}

var (
	defaultLogger *Logger
	once         sync.Once
)

// GetLogger returns the default logger instance
func GetLogger(service string) *Logger {
	once.Do(func() {
		defaultLogger = &Logger{
			entries: make([]LogEntry, 0),
			maxSize: 1000, // Keep last 1000 entries
			level:   INFO,
			service: service,
		}
	})
	return defaultLogger
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel returns the current logging level
func (l *Logger) GetLevel() LogLevel {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.level
}

// shouldLog checks if the message should be logged based on current level
func (l *Logger) shouldLog(level LogLevel) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	switch l.level {
	case DEBUG:
		return true
	case INFO:
		return level == INFO || level == WARN || level == ERROR
	case WARN:
		return level == WARN || level == ERROR
	case ERROR:
		return level == ERROR
	default:
		return true
	}
}

// addEntry adds a new log entry
func (l *Logger) addEntry(entry LogEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	// Add new entry
	l.entries = append(l.entries, entry)
	
	// Remove oldest entries if we exceed maxSize
	if len(l.entries) > l.maxSize {
		l.entries = l.entries[len(l.entries)-l.maxSize:]
	}
}

// Log creates a new log entry
func (l *Logger) Log(level LogLevel, method, path string, status int, duration int64, message string) {
	if !l.shouldLog(level) {
		return
	}

	entry := LogEntry{
		ID:        uuid.New().String(),
		Timestamp: fmt.Sprintf("%d", time.Now().UnixMilli()),
		Method:    method,
		Path:      path,
		Status:    status,
		Duration:  duration,
		Level:     level,
		Message:   message,
		CreatedAt: time.Now(),
	}

	l.addEntry(entry)

	// Also log to stdout for debugging
	log.Printf("[%s] %s %s %d %dms - %s", level, method, path, status, duration, message)
}

// GetLogs retrieves logs with optional filtering
func (l *Logger) GetLogs(limit int, from, to *time.Time, search string, status *int, method, level string) LogsResponse {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var filtered []LogEntry

	for _, entry := range l.entries {
		// Apply filters
		if from != nil && entry.CreatedAt.Before(*from) {
			continue
		}
		if to != nil && entry.CreatedAt.After(*to) {
			continue
		}
		if search != "" && !contains(entry.Message, search) {
			continue
		}
		if status != nil && entry.Status != *status {
			continue
		}
		if method != "" && entry.Method != method {
			continue
		}
		if level != "" && string(entry.Level) != level {
			continue
		}

		filtered = append(filtered, entry)
	}

	// Apply limit
	if limit > 0 && len(filtered) > limit {
		filtered = filtered[len(filtered)-limit:]
	}

	return LogsResponse{Items: filtered}
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && 
			(s[:len(substr)] == substr || 
			 s[len(s)-len(substr):] == substr ||
			 containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Debug logs a debug message
func (l *Logger) Debug(method, path string, status int, duration int64, message string) {
	l.Log(DEBUG, method, path, status, duration, message)
}

// Info logs an info message
func (l *Logger) Info(method, path string, status int, duration int64, message string) {
	l.Log(INFO, method, path, status, duration, message)
}

// Warn logs a warning message
func (l *Logger) Warn(method, path string, status int, duration int64, message string) {
	l.Log(WARN, method, path, status, duration, message)
}

// Error logs an error message
func (l *Logger) Error(method, path string, status int, duration int64, message string) {
	l.Log(ERROR, method, path, status, duration, message)
}
