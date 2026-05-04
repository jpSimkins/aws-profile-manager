// Package core provides core application utilities including state management,
// version information, and time utilities.
//
// The core package contains fundamental functionality used throughout the application:
//   - AppState: Global application state with thread-safe registry
//   - Version: Application version and build information
//   - Time Utilities: Time formatting and duration helpers
//
// State Management:
//
//	The AppState singleton (core.App) provides centralized state management
//	for the application, including settings initialization and a thread-safe
//	registry for storing named state objects (e.g., GUI ViewModels).
//
// Example Usage:
//
//	// Initialize application state and load settings
//	if err := core.App.Initialize(); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Register a state object
//	core.App.RegisterState("myViewModel", viewModel)
//
//	// Retrieve a state object
//	vm := core.App.GetState("myViewModel")
package core

import (
	"fmt"
	"path/filepath"
	"sync"

	"aws-profile-manager/internal/logging"
	"aws-profile-manager/internal/settings"
)

// AppState holds global application state with thread-safe access.
//
// AppState provides a centralized location for application-wide state management,
// including settings initialization and a flexible registry for storing named
// state objects. The registry is commonly used for GUI ViewModels but can store
// any type of state object.
//
// Thread Safety:
//
//	All methods are thread-safe and use read/write locks for optimal performance.
//	Multiple readers can access state concurrently, while writes are exclusive.
type AppState struct {
	registry map[string]interface{} // Named state objects (GUI ViewModels, etc.)
	mu       sync.RWMutex           // Thread-safe access to registry
}

// App is the global application state instance.
//
// This singleton provides application-wide access to state management.
// Initialize it early in application startup with App.Initialize().
var App = newAppState()

// newAppState creates a new application state instance.
//
// This function is called once during package initialization to create
// the global App singleton.
//
// Returns:
//   - *AppState: Initialized application state with empty registry
func newAppState() *AppState {
	return &AppState{
		registry: make(map[string]interface{}),
	}
}

// Initialize loads application settings and configures the application.
//
// This method must be called early in application startup, before accessing
// settings or using other application features. It loads settings from disk,
// creating the file with defaults if it doesn't exist, and applies logging
// configuration.
//
// Process:
//  1. Determine config directory path (from environment or default)
//  2. Load settings from $CONFIG_DIR/settings.json
//  3. Create settings file with defaults if it doesn't exist
//  4. Apply logging configuration (level, debug mode)
//
// Returns:
//   - error: Any error encountered during initialization
//
// Example:
//
//	if err := core.App.Initialize(); err != nil {
//	    log.Fatalf("Failed to initialize application: %v", err)
//	}
func (a *AppState) Initialize() error {
	logging.Debug.Log("AppState initializing")

	// Get config directory and settings path
	configDir, err := settings.GetConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	settingsPath := filepath.Join(configDir, "settings.json")
	logging.Debug.Logf("\t🔹 Loading settings from: %s", settingsPath)

	// Load settings (creates file with defaults if it doesn't exist)
	if err := settings.Load(settingsPath); err != nil {
		return fmt.Errorf("failed to load settings: %w", err)
	}

	// Apply logging settings
	currentSettings := settings.Get()
	logging.Debug.SetEnabled(currentSettings.Logging.EnableDebug)
	logging.UpdateLoggerFromSettings(currentSettings.Logging.LogLevel)
	logging.Debug.Logf("\t\t🔸 Log level: %s", currentSettings.Logging.LogLevel)
	logging.Debug.Logf("\t\t🔸 Debug enabled: %v", currentSettings.Logging.EnableDebug)

	logging.Debug.Log("AppState initialization complete")
	return nil
}

// RegisterState registers a named state object in the global registry.
//
// This method allows different parts of the application to store state objects
// that can be retrieved by other components. Commonly used for GUI ViewModels
// but can store any type of object.
//
// Thread Safety:
//
//	Uses write lock for exclusive access during registration.
//
// Parameters:
//   - key: Unique identifier for the state object
//   - value: State object to register (any type)
//
// Example:
//
//	viewModel := NewSettingsViewModel()
//	core.App.RegisterState("settingsViewModel", viewModel)
func (a *AppState) RegisterState(key string, value interface{}) {
	a.mu.Lock()
	defer a.mu.Unlock()

	logging.Debug.Logf("\t🔹 Registering state: %s (type: %T)", key, value)
	a.registry[key] = value
}

// GetState retrieves a named state object from the registry.
//
// This method returns the state object associated with the given key.
// Returns nil if the key doesn't exist - callers must check for nil.
//
// Thread Safety:
//
//	Uses read lock for concurrent access by multiple readers.
//
// Parameters:
//   - key: Unique identifier for the state object
//
// Returns:
//   - interface{}: State object, or nil if key doesn't exist
//
// Example:
//
//	if vm := core.App.GetState("settingsViewModel"); vm != nil {
//	    settingsVM := vm.(*SettingsViewModel)
//	    // Use settingsVM...
//	}
func (a *AppState) GetState(key string) interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.registry[key]
}

// UnregisterState removes a named state object from the registry.
//
// This method is useful for cleanup when components are destroyed or no longer
// needed. Safe to call even if the key doesn't exist.
//
// Thread Safety:
//
//	Uses write lock for exclusive access during removal.
//
// Parameters:
//   - key: Unique identifier for the state object to remove
//
// Example:
//
//	core.App.UnregisterState("settingsViewModel")
func (a *AppState) UnregisterState(key string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	logging.Debug.Logf("\t🔹 Unregistering state: %s", key)
	delete(a.registry, key)
}

// HasState checks if a named state object exists in the registry.
//
// This method performs a thread-safe existence check without retrieving
// the value, which is more efficient than GetState when you only need
// to know if a key exists.
//
// Thread Safety:
//
//	Uses read lock for concurrent access.
//
// Parameters:
//   - key: Unique identifier to check
//
// Returns:
//   - bool: true if key exists, false otherwise
//
// Example:
//
//	if core.App.HasState("settingsViewModel") {
//	    vm := core.App.GetState("settingsViewModel")
//	    // Use vm...
//	}
func (a *AppState) HasState(key string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	_, exists := a.registry[key]
	return exists
}
