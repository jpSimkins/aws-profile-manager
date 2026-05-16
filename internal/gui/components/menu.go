package components

import (
	"fyne.io/fyne/v2"

	"aws-profile-manager/internal/settings"
)

// MenuCallbacks holds callback functions for menu actions
type MenuCallbacks struct {
	OnExport   func()
	OnImport   func()
	OnSettings func()
	OnSyncNow  func()
	OnAbout    func()
	OnExit     func()
}

// CreateMainMenu creates the application menu with the provided callbacks
func CreateMainMenu(callbacks MenuCallbacks) *fyne.MainMenu {
	// Get sync settings to check if sync is enabled
	currentSettings := settings.Get()
	syncSettings := &currentSettings.Sync

	// File menu
	exportItem := fyne.NewMenuItem("Export", callbacks.OnExport)
	importItem := fyne.NewMenuItem("Import", callbacks.OnImport)
	settingsItem := fyne.NewMenuItem("Settings", callbacks.OnSettings)

	exitItem := fyne.NewMenuItem("Exit", callbacks.OnExit)
	exitItem.IsQuit = true

	// Build menu items - only include Sync if enabled
	var fileMenuItems []*fyne.MenuItem
	fileMenuItems = append(fileMenuItems, exportItem, importItem, fyne.NewMenuItemSeparator())

	// Only add Sync Now if sync is enabled
	if syncSettings.Enabled {
		syncNowItem := fyne.NewMenuItem("Sync Now", callbacks.OnSyncNow)
		syncNowItem.IsQuit = false
		fileMenuItems = append(fileMenuItems, syncNowItem, fyne.NewMenuItemSeparator())
	}

	fileMenuItems = append(fileMenuItems, settingsItem, fyne.NewMenuItemSeparator(), exitItem)

	fileMenu := fyne.NewMenu("File", fileMenuItems...)

	// Help menu
	aboutItem := fyne.NewMenuItem("About", callbacks.OnAbout)
	helpMenu := fyne.NewMenu("Help", aboutItem)

	return fyne.NewMainMenu(fileMenu, helpMenu)
}
