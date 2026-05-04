package core

import (
	"testing"

	"aws-profile-manager/internal/settings"
	"aws-profile-manager/internal/test"
)

func TestAppStateInitialize(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := App.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Verify settings were loaded
	currentSettings := settings.Get()
	if currentSettings == nil {
		t.Fatal("Settings should not be nil after initialization")
	}
	if currentSettings.Version == "" {
		t.Error("Settings version should be set")
	}
}

func TestAppStateRegistry(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := App.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	testState := map[string]interface{}{"test": "value"}
	App.RegisterState("test-key", testState)

	retrieved := App.GetState("test-key")
	if retrieved == nil {
		t.Error("Should retrieve registered state")
	}

	App.UnregisterState("test-key")

	retrieved = App.GetState("test-key")
	if retrieved != nil {
		t.Error("State should be nil after unregister")
	}
}

func TestGlobalAppInstance(t *testing.T) {
	if App == nil {
		t.Fatal("Global App instance should not be nil")
	}
}
