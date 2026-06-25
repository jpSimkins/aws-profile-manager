// Package settings provides unified application configuration management.
//
// This package implements a schema-driven settings system with automatic UI
// generation, validation, and thread-safe access. Settings are organized into
// logical sections (Application, Logging, GUI, Sync, AwsCLI) and persisted
// to disk as JSON.
//
// # Architecture
//
// Settings use a schema-driven approach where field metadata (SchemaField)
// defines UI generation, validation rules, and dependencies. This allows the
// GUI to automatically generate settings dialogs without manual UI code.
//
// # Thread Safety
//
// All settings access is thread-safe using read/write locks. The Get() and Set()
// functions provide safe concurrent access to the global settings instance.
//
// # Validation
//
// Settings are validated automatically on Set() and Load(). Invalid settings
// are rejected with descriptive error messages. Each section implements its
// own Validate() method.
//
// # Usage Pattern
//
//	// Load settings from disk
//	if err := settings.Load(settingsPath); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Read settings (thread-safe)
//	currentSettings := settings.Get()
//	theme := currentSettings.GUI.Theme
//
//	// Modify settings (validates automatically)
//	newSettings := settings.Get()
//	newSettings.GUI.Theme = "dark"
//	if err := settings.Set(newSettings); err != nil {
//	    log.Printf("Invalid settings: %v", err)
//	}
//
//	// Save to disk
//	if err := settings.Save(settingsPath); err != nil {
//	    log.Fatal(err)
//	}
package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"aws-profile-manager/internal/logging"
	"aws-profile-manager/internal/security"
)

// Settings is the root configuration structure containing all application settings.
//
// This structure is organized into logical sections, each with its own validation
// and schema. Settings are persisted to disk as JSON and loaded at startup.
type Settings struct {
	Version     string              `json:"version"`     // Schema version for future migrations
	Application ApplicationSettings `json:"application"` // Application metadata and markers
	Logging     LoggingSettings     `json:"logging"`     // Logging configuration
	GUI         GUISettings         `json:"gui"`         // GUI preferences (theme, size, etc.)
	Sync        SyncSettings        `json:"sync"`        // Remote config sync settings
	AwsCLI      AwsCliSettings      `json:"awscli"`      // AWS CLI integration settings
	Terminal    TerminalSettings    `json:"terminal"`    // Terminal launcher configuration
}

// Global settings instance with thread-safe access.
var (
	current *Settings    // Current settings instance
	mu      sync.RWMutex // Protects concurrent access
)

// Get returns the current settings with thread-safe read access.
//
// If no settings have been loaded, this returns default settings. This function
// is safe to call from multiple goroutines.
//
// Returns:
//   - *Settings: Current settings instance (never nil)
func Get() *Settings {
	mu.RLock()
	defer mu.RUnlock()
	if current == nil {
		return GetDefaults()
	}
	return current
}

// Set updates the current settings with thread-safe write access.
//
// This function validates settings before accepting them. Invalid settings
// are rejected and the error is returned. This ensures the application never
// operates with invalid configuration.
//
// Security Note:
//
//	All settings are validated before being accepted to prevent invalid
//	or malicious configuration from affecting application behavior.
//
// Parameters:
//   - s: New settings to set (must be valid)
//
// Returns:
//   - error: Validation error if settings are invalid, nil on success
//
// Example:
//
//	newSettings := settings.Get()
//	newSettings.GUI.Theme = "dark"
//	if err := settings.Set(newSettings); err != nil {
//	    log.Printf("Invalid settings: %v", err)
//	}
func Set(s *Settings) error {
	// Validate before setting
	if err := s.Validate(); err != nil {
		return logging.Log.ErrorfWithDetails("invalid settings provided", err)
	}

	mu.Lock()
	defer mu.Unlock()
	current = s
	return nil
}

// Load reads settings from disk and sets them as current.
//
// This function loads settings from the specified file path. If the file
// doesn't exist, it creates it with default settings. The loaded settings
// are validated and set as the current global settings.
//
// Process:
//  1. Check if file exists
//  2. If not, create with defaults
//  3. Read and parse JSON
//  4. Validate settings
//  5. Set as current
//
// Parameters:
//   - path: File path to load settings from
//
// Returns:
//   - error: Any error encountered during load, validation, or save
//
// Example:
//
//	settingsPath := filepath.Join(configDir, "settings.json")
//	if err := settings.Load(settingsPath); err != nil {
//	    log.Fatalf("Failed to load settings: %v", err)
//	}
func Load(path string) error {
	logging.Debug.Log("Loading settings", "path", path)

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// No file, use defaults and create file
		logging.Log.Infof("Settings file not found, creating with defaults: %s", path)
		logging.Debug.Log("Initializing with default settings")
		if err := Set(GetDefaults()); err != nil {
			return logging.Log.ErrorfWithDetails("failed to set default settings", err)
		}
		return Save(path)
	}

	// Read file
	data, err := security.ReadFile(path, security.ReadOptions{
		AllowedExtensions: []string{".json"},
	})
	if err != nil {
		return logging.Log.ErrorfWithDetails("failed to read settings file", err,
			"path", path)
	}

	// Unmarshal
	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return logging.Log.ErrorfWithDetails("failed to parse settings file", err,
			"path", path)
	}

	// Validate
	if err := s.Validate(); err != nil {
		return logging.Log.ErrorfWithDetails("settings validation failed", err,
			"path", path)
	}

	// Set as current
	if err := Set(&s); err != nil {
		return logging.Log.ErrorfWithDetails("failed to set validated settings", err)
	}

	logging.Debug.Log("Settings loaded successfully",
		"version", s.Version,
	)
	return nil
}

// Save writes current settings to disk.
//
// This function serializes the current settings to JSON with pretty formatting
// and writes them to the specified file path. Settings are validated before
// saving to ensure only valid configuration is persisted.
//
// Process:
//  1. Get current settings
//  2. Validate settings
//  3. Marshal to JSON (with indentation)
//  4. Create parent directory if needed
//  5. Write to file
//
// Parameters:
//   - path: File path to save settings to
//
// Returns:
//   - error: Any error encountered during validation, serialization, or write
//
// Example:
//
//	settingsPath := filepath.Join(configDir, "settings.json")
//	if err := settings.Save(settingsPath); err != nil {
//	    log.Fatalf("Failed to save settings: %v", err)
//	}
func Save(path string) error {
	logging.Debug.Log("Saving settings", "path", path)

	mu.RLock()
	s := current
	mu.RUnlock()

	if s == nil {
		logging.Debug.Log("No current settings, using defaults")
		s = GetDefaults()
	}

	// Validate before saving
	if err := s.Validate(); err != nil {
		return logging.Log.ErrorfWithDetails("cannot save invalid settings", err,
			"path", path)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return logging.Log.ErrorfWithDetails("failed to marshal settings", err,
			"path", path)
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return logging.Log.ErrorfWithDetails("failed to create settings directory", err,
			"dir", dir)
	}

	// Write file
	if err := os.WriteFile(path, data, 0600); err != nil {
		return logging.Log.ErrorfWithDetails("failed to write settings file", err,
			"path", path)
	}

	logging.Log.Success("Settings saved successfully",
		"path", path,
	)

	// Count sections for debug log
	sectionCount := 0
	if current.Application.ManagedSectionStart != "" {
		sectionCount++
	}
	if current.Logging.LogLevel != "" {
		sectionCount++
	}
	if current.GUI.Theme != "" {
		sectionCount++
	}
	if current.Sync.Strategy != "" {
		sectionCount++
	}
	if current.AwsCLI.RefreshIntervalMins > 0 {
		sectionCount++
	}

	logging.Debug.Log("Settings saved",
		"size", len(data),
		"sections", sectionCount)
	return nil
}

// GetDefaults returns default settings for all sections.
//
// # Two-system defaults pattern
//
// Each settings section maintains defaults in two separate places that serve
// different consumers — this is intentional, not redundant:
//
//  1. GetDefault*() functions (assembled here) → produce the runtime *Settings
//     struct used when no settings.json exists yet, or when the app needs a
//     clean state. These are real Go values that the application runs on.
//
//  2. Schema.Fields[x].Default (in each GetSchema()) → UI metadata only.
//     The settings form renderer reads these to populate empty widgets with a
//     sensible starting value. They are never applied to the live settings
//     struct by the framework.
//
// Keep both in sync. For string/numeric defaults, use a shared constant
// (e.g. SectionHeaderText) so a single edit propagates to both consumers.
//
// Returns:
//   - *Settings: Settings instance with default values
func GetDefaults() *Settings {
	return &Settings{
		Version:     "1.0",
		Application: GetDefaultApplication(),
		Logging:     GetDefaultLogging(),
		GUI:         GetDefaultGUI(),
		Sync:        GetDefaultSync(),
		AwsCLI:      GetDefaultAwsCLI(),
		Terminal:    GetDefaultTerminal(),
	}
}

// Validate validates all settings sections.
//
// This method validates the entire settings structure by calling Validate()
// on each section. Any validation error will prevent the settings from being
// accepted or saved.
//
// Validation Order:
//  1. Application settings
//  2. Logging settings
//  3. GUI settings
//  4. Sync settings
//  5. AwsCLI settings
//
// Returns:
//   - error: First validation error encountered, nil if all sections are valid
func (s *Settings) Validate() error {
	logging.Debug.Log("Validating settings", "version", s.Version)

	if err := s.Application.Validate(); err != nil {
		logging.Debug.Log("Application validation failed", "error", err.Error())
		return fmt.Errorf("application: %w", err)
	}
	if err := s.Logging.Validate(); err != nil {
		logging.Debug.Log("Logging validation failed", "error", err.Error())
		return fmt.Errorf("logging: %w", err)
	}
	if err := s.GUI.Validate(); err != nil {
		logging.Debug.Log("GUI validation failed", "error", err.Error())
		return fmt.Errorf("gui: %w", err)
	}
	if err := s.Sync.Validate(); err != nil {
		logging.Debug.Log("Sync validation failed", "error", err.Error())
		return fmt.Errorf("sync: %w", err)
	}
	if err := s.AwsCLI.Validate(); err != nil {
		logging.Debug.Log("AwsCLI validation failed", "error", err.Error())
		return fmt.Errorf("awscli: %w", err)
	}
	if err := s.Terminal.Validate(); err != nil {
		logging.Debug.Log("Terminal validation failed", "error", err.Error())
		return fmt.Errorf("terminal: %w", err)
	}

	logging.Debug.Log("All settings validated successfully")
	return nil
}

// GetAllSchemas returns schemas for all settings sections.
//
// This method aggregates schema definitions from all settings sections, which
// are used by the GUI to dynamically build the settings dialog. Each section's
// schema includes field metadata for automatic UI generation and validation.
//
// Returns:
//   - map[string]Schema: Map of section names to their schema definitions
//
// Example:
//
//	schemas := settings.GetAllSchemas()
//	appSchema := schemas["application"]
//	for fieldName, fieldSchema := range appSchema.Fields {
//	    // Build UI widget based on fieldSchema
//	}
func (s *Settings) GetAllSchemas() map[string]Schema {
	return map[string]Schema{
		"application": s.Application.GetSchema(),
		"logging":     s.Logging.GetSchema(),
		"gui":         s.GUI.GetSchema(),
		"sync":        s.Sync.GetSchema(),
		"awscli":      s.AwsCLI.GetSchema(),
		"terminal":    s.Terminal.GetSchema(),
	}
}
