package logging

import (
	"os"
	"sync"
	"testing"
)

func TestNewDebugLogger(t *testing.T) {
	debug := NewDebugLogger()

	if debug == nil {
		t.Fatal("NewDebugLogger() returned nil")
	}

	if debug.debugColor == nil {
		t.Error("debugColor not initialized")
	}

	// Should default to environment variable check
	expectedEnabled := parseDebugEnv()
	if debug.enabled != expectedEnabled {
		t.Errorf("Expected enabled to be %v (from env), got %v", expectedEnabled, debug.enabled)
	}
}

func TestParseDebugEnv(t *testing.T) {
	// Save original env value
	originalValue := os.Getenv("AWS_PROFILE_MANAGER_DEBUG")
	defer func() {
		if originalValue == "" {
			os.Unsetenv("AWS_PROFILE_MANAGER_DEBUG")
		} else {
			os.Setenv("AWS_PROFILE_MANAGER_DEBUG", originalValue)
		}
	}()

	tests := []struct {
		envValue    string
		expected    bool
		description string
	}{
		{"", false, "empty env var should return false"},
		{"true", true, "true should return true"},
		{"TRUE", true, "TRUE should return true (case insensitive)"},
		{"True", true, "True should return true (case insensitive)"},
		{"1", true, "1 should return true"},
		{"false", false, "false should return false"},
		{"FALSE", false, "FALSE should return false (case insensitive)"},
		{"False", false, "False should return false (case insensitive)"},
		{"0", false, "0 should return false"},
		{"invalid", false, "invalid value should return false"},
		{"yes", false, "yes should return false (not supported)"},
		{"no", false, "no should return false"},
		{"  true  ", true, "true with whitespace should return true"},
		{"  false  ", false, "false with whitespace should return false"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			if tt.envValue == "" {
				os.Unsetenv("AWS_PROFILE_MANAGER_DEBUG")
			} else {
				os.Setenv("AWS_PROFILE_MANAGER_DEBUG", tt.envValue)
			}

			result := parseDebugEnv()
			if result != tt.expected {
				t.Errorf("parseDebugEnv() with AWS_PROFILE_MANAGER_DEBUG='%s': expected %v, got %v",
					tt.envValue, tt.expected, result)
			}
		})
	}
}

func TestIsEnvOverrideActive(t *testing.T) {
	// Save original env value
	originalValue := os.Getenv("AWS_PROFILE_MANAGER_DEBUG")
	defer func() {
		if originalValue == "" {
			os.Unsetenv("AWS_PROFILE_MANAGER_DEBUG")
		} else {
			os.Setenv("AWS_PROFILE_MANAGER_DEBUG", originalValue)
		}
	}()

	// Test with no env var
	os.Unsetenv("AWS_PROFILE_MANAGER_DEBUG")
	if IsEnvOverrideActive() {
		t.Error("IsEnvOverrideActive() should return false when AWS_PROFILE_MANAGER_DEBUG is not set")
	}

	// Test with env var set
	os.Setenv("AWS_PROFILE_MANAGER_DEBUG", "true")
	if !IsEnvOverrideActive() {
		t.Error("IsEnvOverrideActive() should return true when AWS_PROFILE_MANAGER_DEBUG is set")
	}

	// Test with empty env var (still considered set)
	os.Setenv("AWS_PROFILE_MANAGER_DEBUG", "")
	if IsEnvOverrideActive() {
		t.Error("IsEnvOverrideActive() should return false when AWS_PROFILE_MANAGER_DEBUG is empty")
	}

	// Test with any value
	os.Setenv("AWS_PROFILE_MANAGER_DEBUG", "anything")
	if !IsEnvOverrideActive() {
		t.Error("IsEnvOverrideActive() should return true when AWS_PROFILE_MANAGER_DEBUG has any non-empty value")
	}
}

func TestDebugLoggerSingleton(t *testing.T) {
	// Test that GetDebugLogger returns the same instance
	debug1 := GetDebugLogger()
	debug2 := GetDebugLogger()

	if debug1 != debug2 {
		t.Error("GetDebugLogger() should return the same singleton instance")
	}
}

func TestSetEnabledWithoutEnvOverride(t *testing.T) {
	// Save original env value
	originalValue := os.Getenv("AWS_PROFILE_MANAGER_DEBUG")
	defer func() {
		if originalValue == "" {
			os.Unsetenv("AWS_PROFILE_MANAGER_DEBUG")
		} else {
			os.Setenv("AWS_PROFILE_MANAGER_DEBUG", originalValue)
		}
	}()

	// Ensure no env override
	os.Unsetenv("AWS_PROFILE_MANAGER_DEBUG")

	debug := NewDebugLogger()

	// Test setting enabled to true
	debug.SetEnabled(true)
	if !debug.IsEnabled() {
		t.Error("SetEnabled(true) should enable debug logging")
	}

	// Test setting enabled to false
	debug.SetEnabled(false)
	if debug.IsEnabled() {
		t.Error("SetEnabled(false) should disable debug logging")
	}
}

func TestSetEnabledWithEnvOverride(t *testing.T) {
	// Save original env value
	originalValue := os.Getenv("AWS_PROFILE_MANAGER_DEBUG")
	defer func() {
		if originalValue == "" {
			os.Unsetenv("AWS_PROFILE_MANAGER_DEBUG")
		} else {
			os.Setenv("AWS_PROFILE_MANAGER_DEBUG", originalValue)
		}
	}()

	// Set env override to true
	os.Setenv("AWS_PROFILE_MANAGER_DEBUG", "true")

	debug := NewDebugLogger()

	// Even if we try to disable, env override should keep it enabled
	debug.SetEnabled(false)
	if !debug.IsEnabled() {
		t.Error("SetEnabled(false) should be overridden by AWS_PROFILE_MANAGER_DEBUG=true")
	}

	// Set env override to false
	os.Setenv("AWS_PROFILE_MANAGER_DEBUG", "false")

	// Even if we try to enable, env override should keep it disabled
	debug.SetEnabled(true)
	if debug.IsEnabled() {
		t.Error("SetEnabled(true) should be overridden by AWS_PROFILE_MANAGER_DEBUG=false")
	}
}

func TestDebugLogMethods(t *testing.T) {
	// Save original env value
	originalValue := os.Getenv("AWS_PROFILE_MANAGER_DEBUG")
	defer func() {
		if originalValue == "" {
			os.Unsetenv("AWS_PROFILE_MANAGER_DEBUG")
		} else {
			os.Setenv("AWS_PROFILE_MANAGER_DEBUG", originalValue)
		}
	}()

	// Ensure no env override for this test
	os.Unsetenv("AWS_PROFILE_MANAGER_DEBUG")

	debug := NewDebugLogger()

	// Test Log method when enabled
	debug.SetEnabled(true)
	debug.Log("test message")
	debug.Log("test message with values", "value1", 123)

	// Test Logf method when enabled
	debug.Logf("formatted message: %s %d", "test", 42)

	// Test methods when disabled (should not output anything)
	debug.SetEnabled(false)
	debug.Log("should not appear")
	debug.Log("should not appear", "value1", 123)
	debug.Logf("should not appear: %s", "test")
}

func TestPackageLevelFunctions(t *testing.T) {
	// Save original env value
	originalValue := os.Getenv("AWS_PROFILE_MANAGER_DEBUG")
	defer func() {
		if originalValue == "" {
			os.Unsetenv("AWS_PROFILE_MANAGER_DEBUG")
		} else {
			os.Setenv("AWS_PROFILE_MANAGER_DEBUG", originalValue)
		}
	}()

	// Ensure no env override for this test
	os.Unsetenv("AWS_PROFILE_MANAGER_DEBUG")

	// Test SetDebugEnabled
	SetDebugEnabled(true)
	if !IsDebugEnabled() {
		t.Error("SetDebugEnabled(true) should enable debug logging")
	}

	SetDebugEnabled(false)
	if IsDebugEnabled() {
		t.Error("SetDebugEnabled(false) should disable debug logging")
	}

	// Test UpdateDebugFromSettings
	UpdateDebugFromSettings(true)
	if !IsDebugEnabled() {
		t.Error("UpdateDebugFromSettings(true) should enable debug logging")
	}

	UpdateDebugFromSettings(false)
	if IsDebugEnabled() {
		t.Error("UpdateDebugFromSettings(false) should disable debug logging")
	}
}

func TestGlobalDebugInstance(t *testing.T) {
	// Save original env value and temporarily unset to allow programmatic control
	originalValue := os.Getenv("AWS_PROFILE_MANAGER_DEBUG")
	os.Unsetenv("AWS_PROFILE_MANAGER_DEBUG")
	defer func() {
		if originalValue == "" {
			os.Unsetenv("AWS_PROFILE_MANAGER_DEBUG")
		} else {
			os.Setenv("AWS_PROFILE_MANAGER_DEBUG", originalValue)
		}
	}()

	// Test that the global Debug instance works
	if Debug == nil {
		t.Fatal("Global Debug instance is nil")
	}

	// Test that it's the same as GetDebugLogger()
	if Debug != GetDebugLogger() {
		t.Error("Global Debug instance should be the same as GetDebugLogger()")
	}

	// Test using the global instance (should work now without env override)
	Debug.SetEnabled(true)
	if !Debug.IsEnabled() {
		t.Error("Global Debug.SetEnabled(true) should work")
	}
}

func TestThreadSafetyDebug(t *testing.T) {
	// Save original env value
	originalValue := os.Getenv("AWS_PROFILE_MANAGER_DEBUG")
	defer func() {
		if originalValue == "" {
			os.Unsetenv("AWS_PROFILE_MANAGER_DEBUG")
		} else {
			os.Setenv("AWS_PROFILE_MANAGER_DEBUG", originalValue)
		}
	}()

	// Ensure no env override for this test
	os.Unsetenv("AWS_PROFILE_MANAGER_DEBUG")

	debug := NewDebugLogger()

	var wg sync.WaitGroup

	// Test concurrent access to enabled state
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(enabled bool) {
			defer wg.Done()
			debug.SetEnabled(enabled)
			_ = debug.IsEnabled()
		}(i%2 == 0)
	}

	// Test concurrent logging
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			debug.Log("concurrent debug message", i)
			debug.Logf("concurrent formatted message: %d", i)
		}(i)
	}

	wg.Wait()
}

func TestDebugLoggerInitialization(t *testing.T) {
	// Save original env value
	originalValue := os.Getenv("AWS_PROFILE_MANAGER_DEBUG")
	defer func() {
		if originalValue == "" {
			os.Unsetenv("AWS_PROFILE_MANAGER_DEBUG")
		} else {
			os.Setenv("AWS_PROFILE_MANAGER_DEBUG", originalValue)
		}
	}()

	// Test initialization with env var set to true
	os.Setenv("AWS_PROFILE_MANAGER_DEBUG", "1")
	debug1 := NewDebugLogger()
	if !debug1.IsEnabled() {
		t.Error("NewDebugLogger() should initialize as enabled when AWS_PROFILE_MANAGER_DEBUG=1")
	}

	// Test initialization with env var set to false
	os.Setenv("AWS_PROFILE_MANAGER_DEBUG", "0")
	debug2 := NewDebugLogger()
	if debug2.IsEnabled() {
		t.Error("NewDebugLogger() should initialize as disabled when AWS_PROFILE_MANAGER_DEBUG=0")
	}

	// Test initialization with no env var
	os.Unsetenv("AWS_PROFILE_MANAGER_DEBUG")
	debug3 := NewDebugLogger()
	if debug3.IsEnabled() {
		t.Error("NewDebugLogger() should initialize as disabled when AWS_PROFILE_MANAGER_DEBUG is not set")
	}
}

func TestLogFormats(t *testing.T) {
	// Save original env value
	originalValue := os.Getenv("AWS_PROFILE_MANAGER_DEBUG")
	defer func() {
		if originalValue == "" {
			os.Unsetenv("AWS_PROFILE_MANAGER_DEBUG")
		} else {
			os.Setenv("AWS_PROFILE_MANAGER_DEBUG", originalValue)
		}
	}()

	// Ensure no env override for this test
	os.Unsetenv("AWS_PROFILE_MANAGER_DEBUG")

	debug := NewDebugLogger()
	debug.SetEnabled(true)

	// Test different value types
	debug.Log("test string")
	debug.Log("test with int", 42)
	debug.Log("test with multiple values", "str", 123, true)
	debug.Log("test with nil", nil)
	debug.Log("test with empty string", "")

	// Test different format strings
	debug.Logf("simple format")
	debug.Logf("format with string: %s", "test")
	debug.Logf("format with int: %d", 42)
	debug.Logf("format with multiple: %s %d %t", "test", 42, true)
}

func TestEdgeCasesDebug(t *testing.T) {
	// Save original env value
	originalValue := os.Getenv("AWS_PROFILE_MANAGER_DEBUG")
	defer func() {
		if originalValue == "" {
			os.Unsetenv("AWS_PROFILE_MANAGER_DEBUG")
		} else {
			os.Setenv("AWS_PROFILE_MANAGER_DEBUG", originalValue)
		}
	}()

	// Ensure no env override for this test
	os.Unsetenv("AWS_PROFILE_MANAGER_DEBUG")

	debug := NewDebugLogger()

	// Test empty messages when enabled
	debug.SetEnabled(true)
	debug.Log("")
	debug.Log("", "")
	debug.Logf("")

	// Test empty messages when disabled
	debug.SetEnabled(false)
	debug.Log("")
	debug.Log("", "")
	debug.Logf("")

	// Test with various nil combinations
	debug.SetEnabled(true)
	debug.Log("test", nil, nil)
	debug.Log("", nil)
}

// Benchmark tests for performance
func BenchmarkDebugLogEnabled(b *testing.B) {
	// Save original env value
	originalValue := os.Getenv("AWS_PROFILE_MANAGER_DEBUG")
	defer func() {
		if originalValue == "" {
			os.Unsetenv("AWS_PROFILE_MANAGER_DEBUG")
		} else {
			os.Setenv("AWS_PROFILE_MANAGER_DEBUG", originalValue)
		}
	}()

	// Ensure no env override for this test
	os.Unsetenv("AWS_PROFILE_MANAGER_DEBUG")

	debug := NewDebugLogger()
	debug.SetEnabled(true)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		debug.Log("benchmark message", i)
	}
}

func BenchmarkDebugLogDisabled(b *testing.B) {
	// Save original env value
	originalValue := os.Getenv("AWS_PROFILE_MANAGER_DEBUG")
	defer func() {
		if originalValue == "" {
			os.Unsetenv("AWS_PROFILE_MANAGER_DEBUG")
		} else {
			os.Setenv("AWS_PROFILE_MANAGER_DEBUG", originalValue)
		}
	}()

	// Ensure no env override for this test
	os.Unsetenv("AWS_PROFILE_MANAGER_DEBUG")

	debug := NewDebugLogger()
	debug.SetEnabled(false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		debug.Log("benchmark message", i)
	}
}

func BenchmarkParseDebugEnv(b *testing.B) {
	os.Setenv("AWS_PROFILE_MANAGER_DEBUG", "true")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parseDebugEnv()
	}
}

func TestIsSilencedDebug(t *testing.T) {
	// Test cases for isSilencedDebug function
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
			// Save original environment variable
			originalValue := os.Getenv("AWS_PROFILE_MANAGER_SILENCE_LOGGER")
			defer func() {
				if originalValue != "" {
					_ = os.Setenv("AWS_PROFILE_MANAGER_SILENCE_LOGGER", originalValue)
				} else {
					_ = os.Unsetenv("AWS_PROFILE_MANAGER_SILENCE_LOGGER")
				}
			}()

			// Set test environment variable
			if tt.envValue == "" {
				_ = os.Unsetenv("AWS_PROFILE_MANAGER_SILENCE_LOGGER")
			} else {
				_ = os.Setenv("AWS_PROFILE_MANAGER_SILENCE_LOGGER", tt.envValue)
			}

			result := isSilencedDebug()
			if result != tt.expected {
				t.Errorf("isSilencedDebug() with env=%q: expected %v, got %v", tt.envValue, tt.expected, result)
			}
		})
	}
}

func TestDebugLoggerSilenced(t *testing.T) {
	// Save original environment variables
	originalDebug := os.Getenv("AWS_PROFILE_MANAGER_DEBUG")
	originalSilence := os.Getenv("AWS_PROFILE_MANAGER_SILENCE_LOGGER")
	defer func() {
		if originalDebug != "" {
			_ = os.Setenv("AWS_PROFILE_MANAGER_DEBUG", originalDebug)
		} else {
			_ = os.Unsetenv("AWS_PROFILE_MANAGER_DEBUG")
		}
		if originalSilence != "" {
			_ = os.Setenv("AWS_PROFILE_MANAGER_SILENCE_LOGGER", originalSilence)
		} else {
			_ = os.Unsetenv("AWS_PROFILE_MANAGER_SILENCE_LOGGER")
		}
	}()

	// Enable debug but also enable silence
	_ = os.Setenv("AWS_PROFILE_MANAGER_DEBUG", "true")
	_ = os.Setenv("AWS_PROFILE_MANAGER_SILENCE_LOGGER", "1")

	debug := NewDebugLogger()

	// Debug should be enabled but silenced
	if !debug.IsEnabled() {
		t.Error("Debug should be enabled when AWS_PROFILE_MANAGER_DEBUG=true")
	}

	if !isSilencedDebug() {
		t.Error("isSilencedDebug() should return true when AWS_PROFILE_MANAGER_SILENCE_LOGGER=1")
	}

	// These methods should not panic when silenced
	debug.Log("silenced debug message")
	debug.Log("silenced debug", "value1", 123, "value2")
	debug.Logf("silenced debug formatted: %s %d", "text", 456)
}

func TestDebugLoggerNotSilenced(t *testing.T) {
	// Save original environment variables
	originalDebug := os.Getenv("AWS_PROFILE_MANAGER_DEBUG")
	originalSilence := os.Getenv("AWS_PROFILE_MANAGER_SILENCE_LOGGER")
	defer func() {
		if originalDebug != "" {
			_ = os.Setenv("AWS_PROFILE_MANAGER_DEBUG", originalDebug)
		} else {
			_ = os.Unsetenv("AWS_PROFILE_MANAGER_DEBUG")
		}
		if originalSilence != "" {
			_ = os.Setenv("AWS_PROFILE_MANAGER_SILENCE_LOGGER", originalSilence)
		} else {
			_ = os.Unsetenv("AWS_PROFILE_MANAGER_SILENCE_LOGGER")
		}
	}()

	// Enable debug and disable silence
	_ = os.Setenv("AWS_PROFILE_MANAGER_DEBUG", "true")
	_ = os.Unsetenv("AWS_PROFILE_MANAGER_SILENCE_LOGGER")

	debug := NewDebugLogger()

	// Debug should be enabled and not silenced
	if !debug.IsEnabled() {
		t.Error("Debug should be enabled when AWS_PROFILE_MANAGER_DEBUG=true")
	}

	if isSilencedDebug() {
		t.Error("isSilencedDebug() should return false when AWS_PROFILE_MANAGER_SILENCE_LOGGER is not set")
	}

	// These methods should execute normally (not panic)
	debug.Log("normal debug message")
	debug.Log("normal debug", "value1", 123, "value2")
	debug.Logf("normal debug formatted: %s %d", "text", 456)
}
