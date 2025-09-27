package components

import (
	"testing"

	"fyne.io/fyne/v2"

	"aws-profile-manager/internal/settings"
	"aws-profile-manager/internal/test"
)

func TestCreateMainMenu_SyncDisabled_HidesSyncNow(t *testing.T) {
	test.SetupTestEnvironment(t)

	cfg := settings.GetDefaults()
	cfg.Sync.Enabled = false
	if err := settings.Set(cfg); err != nil {
		t.Fatalf("failed to set settings: %v", err)
	}

	menu := CreateMainMenu(MenuCallbacks{})
	if menu == nil {
		t.Fatal("CreateMainMenu should not return nil")
	}

	fileMenu := findMenu(menu, "File")
	if fileMenu == nil {
		t.Fatal("expected File menu")
	}

	if hasMenuItem(fileMenu, "Sync Now") {
		t.Fatal("expected Sync Now hidden when sync is disabled")
	}
}

func TestCreateMainMenu_SyncEnabled_ShowsSyncNow(t *testing.T) {
	test.SetupTestEnvironment(t)

	cfg := settings.GetDefaults()
	cfg.Sync.Enabled = true
	cfg.Sync.Strategy = "local"
	cfg.Sync.Local.Path = "/tmp/config.json"
	if err := settings.Set(cfg); err != nil {
		t.Fatalf("failed to set settings: %v", err)
	}

	menu := CreateMainMenu(MenuCallbacks{})
	fileMenu := findMenu(menu, "File")
	if fileMenu == nil {
		t.Fatal("expected File menu")
	}

	if !hasMenuItem(fileMenu, "Sync Now") {
		t.Fatal("expected Sync Now shown when sync is enabled")
	}
}

func TestCreateMainMenu_ExitItemIsQuit(t *testing.T) {
	test.SetupTestEnvironment(t)

	menu := CreateMainMenu(MenuCallbacks{})
	fileMenu := findMenu(menu, "File")
	if fileMenu == nil {
		t.Fatal("expected File menu")
	}

	exitItem := findMenuItem(fileMenu, "Exit")
	if exitItem == nil {
		t.Fatal("expected Exit item")
	}
	if !exitItem.IsQuit {
		t.Fatal("expected Exit item IsQuit=true")
	}
}

func findMenu(menu *fyne.MainMenu, label string) *fyne.Menu {
	if menu == nil {
		return nil
	}
	for _, m := range menu.Items {
		if m != nil && m.Label == label {
			return m
		}
	}
	return nil
}

func hasMenuItem(menu *fyne.Menu, label string) bool {
	return findMenuItem(menu, label) != nil
}

func findMenuItem(menu *fyne.Menu, label string) *fyne.MenuItem {
	if menu == nil {
		return nil
	}
	for _, item := range menu.Items {
		if item != nil && item.Label == label {
			return item
		}
	}
	return nil
}
