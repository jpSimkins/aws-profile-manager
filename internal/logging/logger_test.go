package logging

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
)

// newTestLogger creates a new logger with silenced output for testing
func newTestLogger(t *testing.T) *Logger {
	t.Helper()

	// Save original value
	originalValue := os.Getenv("AWS_PROFILE_MANAGER_SILENCE_LOGGER")

	// Restore after test
	t.Cleanup(func() {
		if originalValue != "" {
			_ = os.Setenv("AWS_PROFILE_MANAGER_SILENCE_LOGGER", originalValue)
		} else {
			_ = os.Unsetenv("AWS_PROFILE_MANAGER_SILENCE_LOGGER")
		}
	})

	// Silence logger output during tests
	_ = os.Setenv("AWS_PROFILE_MANAGER_SILENCE_LOGGER", "1")
	return NewLogger()
}

func TestNewLogger(t *testing.T) {
	logger := newTestLogger(t)

	if logger == nil {
		t.Fatal("NewLogger() returned nil")
	}

	if logger.errorColor == nil {
		t.Error("errorColor not initialized")
	}

	if logger.successColor == nil {
		t.Error("successColor not initialized")
	}

	if logger.infoColor == nil {
		t.Error("infoColor not initialized")
	}

	if logger.valueColor == nil {
		t.Error("valueColor not initialized")
	}

	if logger.warnColor == nil {
		t.Error("warnColor not initialized")
	}

	if logger.currentLevel != LogLevelInfo {
		t.Errorf("Expected default log level to be %s, got %s", LogLevelInfo, logger.currentLevel)
	}
}

func TestLoggerSingleton(t *testing.T) {
	// Test that GetLogger returns the same instance
	logger1 := GetLogger()
	logger2 := GetLogger()

	if logger1 != logger2 {
		t.Error("GetLogger() should return the same singleton instance")
	}
}

func TestLogLevels(t *testing.T) {
	logger := newTestLogger(t)

	// Test setting and getting log levels
	testLevels := []string{LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError}

	for _, level := range testLevels {
		logger.SetLogLevel(level)
		if logger.GetLogLevel() != level {
			t.Errorf("Expected log level %s, got %s", level, logger.GetLogLevel())
		}
	}
}

func TestShouldLog(t *testing.T) {
	logger := newTestLogger(t)

	tests := []struct {
		currentLevel string
		messageLevel string
		shouldLog    bool
		description  string
	}{
		{LogLevelDebug, LogLevelDebug, true, "debug message at debug level"},
		{LogLevelDebug, LogLevelInfo, true, "info message at debug level"},
		{LogLevelDebug, LogLevelWarn, true, "warn message at debug level"},
		{LogLevelDebug, LogLevelError, true, "error message at debug level"},

		{LogLevelInfo, LogLevelDebug, false, "debug message at info level"},
		{LogLevelInfo, LogLevelInfo, true, "info message at info level"},
		{LogLevelInfo, LogLevelWarn, true, "warn message at info level"},
		{LogLevelInfo, LogLevelError, true, "error message at info level"},

		{LogLevelWarn, LogLevelDebug, false, "debug message at warn level"},
		{LogLevelWarn, LogLevelInfo, false, "info message at warn level"},
		{LogLevelWarn, LogLevelWarn, true, "warn message at warn level"},
		{LogLevelWarn, LogLevelError, true, "error message at warn level"},

		{LogLevelError, LogLevelDebug, false, "debug message at error level"},
		{LogLevelError, LogLevelInfo, false, "info message at error level"},
		{LogLevelError, LogLevelWarn, false, "warn message at error level"},
		{LogLevelError, LogLevelError, true, "error message at error level"},

		{"unknown", LogLevelInfo, true, "unknown current level should allow logging"},
		{LogLevelInfo, "unknown", true, "unknown message level should allow logging"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			logger.SetLogLevel(tt.currentLevel)
			result := shouldLog(logger, tt.messageLevel)
			if result != tt.shouldLog {
				t.Errorf("shouldLog(%s) with level %s: expected %v, got %v",
					tt.messageLevel, tt.currentLevel, tt.shouldLog, result)
			}
		})
	}
}

func TestErrorMethods(t *testing.T) {
	logger := newTestLogger(t)

	// Test Error method returns error
	err := logger.Error("test error")
	if err == nil {
		t.Error("Error() should return an error")
	}
	if err.Error() != "test error" {
		t.Errorf("Expected error message 'test error', got '%s'", err.Error())
	}

	// Test Error with format string
	err = logger.Error("test error with values: %s and %d", "value1", 123)
	if err == nil {
		t.Error("Error() with values should return an error")
	}
	if !strings.Contains(err.Error(), "test error with values") {
		t.Errorf("Error message should contain 'test error with values', got '%s'", err.Error())
	}

	// Test Errorf method
	err = logger.Errorf("formatted error: %s %d", "test", 42)
	if err == nil {
		t.Error("Errorf() should return an error")
	}
	expectedMsg := "formatted error: test 42"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestInfoMethods(t *testing.T) {
	logger := newTestLogger(t)

	// Test Info respects log level
	logger.SetLogLevel(LogLevelWarn) // Info should be filtered out
	logger.Info("should not log")

	logger.SetLogLevel(LogLevelInfo) // Info should log
	logger.Info("should log")
	logger.Info("should log with values", "value1", 123)
}

func TestSuccessMethods(t *testing.T) {
	logger := newTestLogger(t)

	// Test Success respects log level (treated as info)
	logger.SetLogLevel(LogLevelWarn) // Success should be filtered out
	logger.Success("should not log")

	logger.SetLogLevel(LogLevelInfo) // Success should log
	logger.Success("should log")
	logger.Success("should log with values", "value1", 123)
}

func TestWarnMethods(t *testing.T) {
	logger := newTestLogger(t)

	// Test Warn respects log level
	logger.SetLogLevel(LogLevelError) // Warn should be filtered out
	logger.Warn("should not log")
	logger.Warnf("should not log: %s", "test")

	logger.SetLogLevel(LogLevelWarn) // Warn should log
	logger.Warn("should log")
	logger.Warn("should log with values", "value1", 123)
	logger.Warnf("formatted warning: %s", "test")
}

func TestErrorWithDetails(t *testing.T) {
	logger := newTestLogger(t)

	// Test with regular error
	err := fmt.Errorf("simple error")
	_ = logger.ErrorWithDetails("Test Error", err)

	// Test with wrapped error
	innerErr := fmt.Errorf("inner error")
	wrappedErr := fmt.Errorf("outer error: %w", innerErr)
	_ = logger.ErrorWithDetails("Wrapped Error", wrappedErr)

	// Test with complex error chain
	complexErr := fmt.Errorf("failed to read file: open /path/to/file: permission denied")
	_ = logger.ErrorWithDetails("Complex Error", complexErr)

	// Test with system call errors that should be merged
	systemErr := fmt.Errorf("mkdir /some/path: permission denied")
	_ = logger.ErrorWithDetails("System Error", systemErr)

	// Test with non-error details
	_ = logger.ErrorWithDetails("Non-Error Details", "simple string")
	_ = logger.ErrorWithDetails("Non-Error Details", 12345)
}

func TestErrorfWithDetails(t *testing.T) {
	logger := newTestLogger(t)

	err := fmt.Errorf("test error")
	_ = logger.ErrorfWithDetails("Formatted title: %s", err, "test")
}

func TestInfoWithDetails(t *testing.T) {
	logger := newTestLogger(t)

	logger.InfoWithDetails("Info Title", "some details")
	logger.InfoWithDetails("Info Title", 12345)
}

func TestSuccessWithDetails(t *testing.T) {
	logger := newTestLogger(t)

	logger.SuccessWithDetails("Success Title", "operation completed")
	logger.SuccessWithDetails("Success Title", map[string]string{"key": "value"})
}

func TestThreadSafety(t *testing.T) {
	logger := newTestLogger(t)

	var wg sync.WaitGroup

	// Test concurrent access to log level
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(level string) {
			defer wg.Done()
			logger.SetLogLevel(level)
			_ = logger.GetLogLevel()
		}(fmt.Sprintf("level-%d", i%4))
	}

	// Test concurrent logging
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			logger.Info("concurrent message %d", i)
			_ = logger.Error("concurrent error %d", i)
			logger.Warn("concurrent warning %d", i)
		}(i)
	}

	wg.Wait()
}

func TestUpdateLoggerFromSettings(t *testing.T) {
	// Test the package-level function
	UpdateLoggerFromSettings(LogLevelDebug)
	if Log.GetLogLevel() != LogLevelDebug {
		t.Errorf("Expected log level %s, got %s", LogLevelDebug, Log.GetLogLevel())
	}

	UpdateLoggerFromSettings(LogLevelError)
	if Log.GetLogLevel() != LogLevelError {
		t.Errorf("Expected log level %s, got %s", LogLevelError, Log.GetLogLevel())
	}
}

func TestGlobalLogInstance(t *testing.T) {
	// Test that the global Log instance works
	if Log == nil {
		t.Fatal("Global Log instance is nil")
	}

	// Test that it's the same as GetLogger()
	if Log != GetLogger() {
		t.Error("Global Log instance should be the same as GetLogger()")
	}

	// Test using the global instance
	err := Log.Error("global test error")
	if err == nil {
		t.Error("Global Log.Error() should return an error")
	}
}

func TestLogLevelConstants(t *testing.T) {
	// Test that all log level constants are defined
	levels := []string{LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError}
	expectedLevels := []string{"debug", "info", "warn", "error"}

	for i, level := range levels {
		if level != expectedLevels[i] {
			t.Errorf("Expected log level constant %s, got %s", expectedLevels[i], level)
		}
	}

	// Test that all levels have priorities
	for _, level := range levels {
		if _, exists := logLevelPriority[level]; !exists {
			t.Errorf("Log level %s missing from priority map", level)
		}
	}
}

func TestLogLevelPriorities(t *testing.T) {
	// Test that priorities are in correct order
	debugPriority := logLevelPriority[LogLevelDebug]
	infoPriority := logLevelPriority[LogLevelInfo]
	warnPriority := logLevelPriority[LogLevelWarn]
	errorPriority := logLevelPriority[LogLevelError]

	if debugPriority >= infoPriority {
		t.Error("Debug priority should be less than Info priority")
	}
	if infoPriority >= warnPriority {
		t.Error("Info priority should be less than Warn priority")
	}
	if warnPriority >= errorPriority {
		t.Error("Warn priority should be less than Error priority")
	}
}

// Test coverage for edge cases
func TestEdgeCases(t *testing.T) {
	logger := newTestLogger(t)

	// Test empty messages
	err := logger.Error("")
	if err == nil {
		t.Error("Empty error message should still return error")
	}

	logger.Info("")
	logger.Success("")
	logger.Warn("")
	logger.Warnf("")

	// Test with nil values (should not panic)
	logger.Info("test", nil)
	logger.Success("test", nil)
	logger.Warn("test", nil)

	// Test ErrorWithDetails with empty error
	emptyErr := fmt.Errorf("")
	_ = logger.ErrorWithDetails("Empty Error", emptyErr)
}

func TestIsSilenced(t *testing.T) {
	// Test cases for isSilenced function
	tests := []struct {
		name     string
		envValue string
		expected bool
	}{
		{"empty string", "", false},
		{"true lowercase", "true", true},
		{"true uppercase", "TRUE", true},
		{"true mixed case", "TrUe", true},
		{"1 as string", "1", true},
		{"false lowercase", "false", false},
		{"false uppercase", "FALSE", false},
		{"0 as string", "0", false},
		{"random string", "random", false},
		{"whitespace true", "  true  ", true},
		{"whitespace 1", "  1  ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set test environment variable
			if tt.envValue == "" {
				_ = os.Unsetenv("AWS_PROFILE_MANAGER_SILENCE_LOGGER")
			} else {
				_ = os.Setenv("AWS_PROFILE_MANAGER_SILENCE_LOGGER", tt.envValue)
			}

			result := isSilenced()
			if result != tt.expected {
				t.Errorf("isSilenced() with env=%q: expected %v, got %v", tt.envValue, tt.expected, result)
			}
		})
	}
}

func TestLoggerSilenced(t *testing.T) {
	// Enable silence
	_ = os.Setenv("AWS_PROFILE_MANAGER_SILENCE_LOGGER", "1")

	logger := newTestLogger(t)

	// All these methods should not panic when silenced
	// and should return appropriately

	// Test Error - should still return error even when silenced
	err := logger.Error("silenced error")
	if err == nil {
		t.Error("Error() should return error even when silenced")
	}
	if err.Error() != "silenced error" {
		t.Errorf("Expected error message 'silenced error', got '%s'", err.Error())
	}

	// Test Errorf - should still return error even when silenced
	err = logger.Errorf("silenced error: %s", "details")
	if err == nil {
		t.Error("Errorf() should return error even when silenced")
	}

	// Test ErrorWithDetails - should still return error even when silenced
	testErr := fmt.Errorf("test error")
	err = logger.ErrorWithDetails("Silenced error with details", testErr)
	if err == nil {
		t.Error("ErrorWithDetails() should return error even when silenced")
	}

	// Test ErrorfWithDetails - should still return error even when silenced
	err = logger.ErrorfWithDetails("Silenced error: %s", testErr, "arg")
	if err == nil {
		t.Error("ErrorfWithDetails() should return error even when silenced")
	}

	// Test non-error methods - should not panic when silenced
	logger.Info("silenced info")
	logger.Success("silenced success")
	logger.Warn("silenced warn")
	logger.Warnf("silenced warnf: %s", "details")
	logger.InfoWithDetails("silenced info details", "details")
	logger.SuccessWithDetails("silenced success details", "details")

	// Verify silence is actually enabled
	if !isSilenced() {
		t.Error("isSilenced() should return true when AWS_PROFILE_MANAGER_SILENCE_LOGGER=1")
	}
}

func TestLoggerSilencedWithValues(t *testing.T) {
	// Enable silence
	_ = os.Setenv("AWS_PROFILE_MANAGER_SILENCE_LOGGER", "true")

	logger := newTestLogger(t)

	// Test all methods with format strings
	err := logger.Error("error with values: %s and %d", "value1", 123)
	if err == nil {
		t.Error("Error() with values should return error even when silenced")
	}

	logger.Info("info with values: %s and %d", "value1", 123)
	logger.Success("success with values: %s and %d", "value1", 123)
	logger.Warn("warn with values: %s and %d", "value1", 123)
}
