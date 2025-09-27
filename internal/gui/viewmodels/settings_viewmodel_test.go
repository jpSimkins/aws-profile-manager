package viewmodels

import (
	"testing"

	"aws-profile-manager/internal/core"
	"aws-profile-manager/internal/settings"
	"aws-profile-manager/internal/test"
)

func TestNewSettingsViewModel(t *testing.T) {
	test.SetupTestEnvironment(t)

	vm := NewSettingsViewModel()
	defer vm.Cleanup()

	if vm == nil {
		t.Fatal("NewSettingsViewModel returned nil")
	}

	state := core.App.GetState("settings-view")
	if state == nil {
		t.Error("ViewModel not registered in core state")
	}

	if vm.GetIsDirty() {
		t.Error("New ViewModel should not be dirty")
	}
}

func TestSettingsViewModel_MarkDirty(t *testing.T) {
	test.SetupTestEnvironment(t)

	vm := NewSettingsViewModel()
	defer vm.Cleanup()

	if vm.GetIsDirty() {
		t.Error("New ViewModel should not be dirty")
	}

	vm.MarkDirty()

	if !vm.GetIsDirty() {
		t.Error("ViewModel should be dirty after MarkDirty()")
	}
}

func TestSettingsViewModel_SaveSettings(t *testing.T) {
	test.SetupTestEnvironment(t)

	vm := NewSettingsViewModel()
	defer vm.Cleanup()

	// Set up explicit test settings
	currentSettings := settings.Get()
	currentSettings.Logging.EnableDebug = false
	if err := settings.Set(currentSettings); err != nil {
		t.Fatalf("Failed to set settings: %v", err)
	}

	vm.MarkDirty()
	if !vm.GetIsDirty() {
		t.Error("ViewModel should be dirty after MarkDirty()")
	}

	vm.InitializeValues()
	if _, err := vm.SaveSettings(nil); err != nil {
		t.Errorf("SaveSettings failed: %v", err)
	}

	if vm.GetIsDirty() {
		t.Error("ViewModel should not be dirty after successful save")
	}
}

func TestSettingsViewModel_Cleanup(t *testing.T) {
	test.SetupTestEnvironment(t)

	vm := NewSettingsViewModel()

	state := core.App.GetState("settings-view")
	if state == nil {
		t.Error("ViewModel should be registered before cleanup")
	}

	vm.Cleanup()

	state = core.App.GetState("settings-view")
	if state != nil {
		t.Error("ViewModel should be unregistered after cleanup")
	}
}

func TestSettingsViewModel_ThreadSafety(t *testing.T) {
	test.SetupTestEnvironment(t)

	vm := NewSettingsViewModel()
	defer vm.Cleanup()

	done := make(chan bool, 3)

	go func() {
		for i := 0; i < 100; i++ {
			vm.MarkDirty()
			vm.InitializeValues()
			_, _ = vm.SaveSettings(nil)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			vm.MarkDirty()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			_ = vm.GetIsDirty()
		}
		done <- true
	}()

	// Wait for all goroutines
	<-done
	<-done
	<-done
}
