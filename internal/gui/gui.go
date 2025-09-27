package gui

import (
	stdsync "sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"

	"aws-profile-manager/internal/gui/components"
	"aws-profile-manager/internal/gui/views"
	"aws-profile-manager/internal/logging"
	"aws-profile-manager/internal/settings"
)

// MenuRefreshCallback is called after GUI closes to refresh menu
var MenuRefreshCallback func()

// Global reference to the Fyne app (set when GUI is created)
var (
	currentFyneApp fyne.App
	fyneAppMu      stdsync.RWMutex
	currentApp     *App
)

// GetCurrentFyneApp returns the current Fyne app if one exists (thread-safe)
func GetCurrentFyneApp() fyne.App {
	fyneAppMu.RLock()
	defer fyneAppMu.RUnlock()
	return currentFyneApp
}

// setCurrentFyneApp stores the Fyne app reference (thread-safe)
func setCurrentFyneApp(app fyne.App) {
	fyneAppMu.Lock()
	defer fyneAppMu.Unlock()
	currentFyneApp = app
}

// GetCurrentApp returns the current App instance
func GetCurrentApp() *App {
	return currentApp
}

// App represents the main GUI application
type App struct {
	fyneApp      fyne.App
	window       fyne.Window
	header       *components.Header
	footer       *components.Footer
	content      fyne.CanvasObject
	overlayStack *fyne.Container // stack that holds tabs + full-screen overlays
}

// NewApp creates and initializes the GUI application
// The settings should already be loaded by main.go
func NewApp() (*App, error) {
	logging.Debug.Log("\t🔹 Creating GUI application")

	// Verify settings are loaded
	currentSettings := settings.Get()
	if currentSettings == nil {
		return nil, logging.Log.Error("Settings not loaded")
	}

	// Create Fyne application
	fyneApp := app.NewWithID("com.son9ne.aws-profile-manager")

	// Store reference for settings updates
	setCurrentFyneApp(fyneApp)

	// Apply theme based on user preference
	// Supports: "light", "dark", or "system" (default)
	guiSettings := &currentSettings.GUI
	ApplyTheme(fyneApp, guiSettings.Theme)
	logging.Debug.Logf("\t🔹 Applied theme: %s", guiSettings.Theme)

	guiApp := &App{
		fyneApp: fyneApp,
	}

	// Store global reference
	currentApp = guiApp

	logging.Debug.Log("GUI application created successfully")
	return guiApp, nil
}

// Run starts the GUI application
func (a *App) Run(configFile string) {
	logging.Debug.Log("Starting GUI application")

	// Get GUI settings - since we're in the same package, we can access the typed settings directly
	currentSettings := settings.Get()
	guiSettings := &currentSettings.GUI

	// Create main window
	a.window = a.fyneApp.NewWindow("AWS Profile Manager")

	// Set window size from settings
	a.window.Resize(fyne.NewSize(
		float32(guiSettings.WindowWidth),
		float32(guiSettings.WindowHeight),
	))

	// Create components
	a.header = components.NewHeader()
	a.footer = components.NewFooter()

	// Create tab content wrapped in a stack so full-screen overlays (e.g.
	// Settings) can be pushed on top without opening a dialog.
	a.content = a.createTabContent()
	a.overlayStack = container.NewStack(a.content)

	// Add padding to header (top padding) - header is now a widget
	paddedHeader := container.NewPadded(a.header)

	// Build main layout
	mainContent := container.NewBorder(
		paddedHeader,            // Top (with padding)
		a.footer.GetContainer(), // Bottom
		nil,                     // Left
		nil,                     // Right
		a.overlayStack,          // Center
	)

	// Create menu using components
	menuCallbacks := components.MenuCallbacks{
		OnExport:   func() { a.handleExport() },
		OnImport:   func() { a.handleImport() },
		OnSettings: func() { a.handleSettings() },
		OnSyncNow:  func() { a.handleSyncNow() },
		OnAbout:    func() { a.handleAbout() },
		OnExit:     func() { a.fyneApp.Quit() },
	}
	a.window.SetMainMenu(components.CreateMainMenu(menuCallbacks))

	// Menu refresh will be triggered by settings save in ViewModel
	// (No callback needed here - settings changes trigger RefreshMenu directly)

	// Set window content
	a.window.SetContent(mainContent)

	// Show initial status
	a.footer.SetStatus("Ready")

	// Center window on screen
	a.window.CenterOnScreen()

	logging.Log.Success("GUI application started")

	// Show and run
	a.window.ShowAndRun()

	logging.Debug.Log("GUI application closed")
}

// createTabContent builds the main tab container with Profiles and Install tabs.
//
// Profiles is the default (first) tab shown when the application starts. Each
// tab's content is lazy-loaded: the view is only constructed when its tab is
// first selected and then cached, so subsequent selections do not rebuild it.
func (a *App) createTabContent() fyne.CanvasObject {
	// Use mutable stack containers as placeholders so content can be injected
	// the first time each tab is selected without rebuilding the tab structure.
	profilesStack := container.NewStack()
	installStack := container.NewStack()

	loaded := map[string]bool{}

	loadView := func(name string, stack *fyne.Container, factory func(fyne.Window) fyne.CanvasObject) {
		if loaded[name] {
			return
		}
		stack.Add(factory(a.window))
		stack.Refresh()
		loaded[name] = true
	}

	profilesTab := container.NewTabItem("Profiles", profilesStack)
	installTab := container.NewTabItem("Install", installStack)

	tabs := container.NewAppTabs(profilesTab, installTab)
	tabs.SetTabLocation(container.TabLocationTop)

	// Lazy-load content when a tab is first selected
	tabs.OnSelected = func(tab *container.TabItem) {
		switch tab {
		case profilesTab:
			loadView("profiles", profilesStack, views.NewProfilesView)
		case installTab:
			loadView("install", installStack, views.NewInstallView)
		}
	}

	// Load the default tab (Profiles) immediately
	loadView("profiles", profilesStack, views.NewProfilesView)

	return tabs
}

// handleSyncNow handles the Sync Now menu action
func (a *App) handleSyncNow() {
	views.ShowSyncDialog(a.window, a.footer)
}

// handleSettings opens the Settings full-screen overlay.
func (a *App) handleSettings() {
	onClose := func() {
		// Pop the settings view — the stack collapses back to tabs.
		if len(a.overlayStack.Objects) > 1 {
			a.overlayStack.Objects = a.overlayStack.Objects[:len(a.overlayStack.Objects)-1]
			a.overlayStack.Refresh()
		}
		if a.footer != nil {
			a.footer.SetStatus("Ready")
		}
	}

	settingsView := views.NewSettingsView(a.window, a.footer, onClose, a.RefreshSettings)
	a.overlayStack.Add(settingsView)
	a.overlayStack.Refresh()
	if a.footer != nil {
		a.footer.SetStatus("Settings")
	}
}

// RefreshSettings refreshes GUI settings after they've been saved
func (a *App) RefreshSettings() {
	// Refresh menu
	a.RefreshMenu()

	// Re-apply theme in case it changed
	currentSettings := settings.Get()
	if currentSettings != nil {
		ApplyTheme(a.fyneApp, currentSettings.GUI.Theme)
		logging.Debug.Logf("Theme re-applied: %s", currentSettings.GUI.Theme)
	}
}

// RefreshMenu updates the main menu with current settings

// handleExport opens the Export Profiles dialog
func (a *App) handleExport() {
	views.ShowExportDialog(a.window, a.footer)
}

// handleImport opens the Import Profiles dialog
func (a *App) handleImport() {
	views.ShowImportDialog(a.window, a.footer)
}

// handleAbout shows the About dialog
func (a *App) handleAbout() {
	views.ShowAboutDialog(a.window, a.footer)
}

// GetWindow returns the main window (useful for testing)
func (a *App) GetWindow() fyne.Window {
	return a.window
}

// GetFyneApp returns the Fyne app instance (useful for testing)
func (a *App) GetFyneApp() fyne.App {
	return a.fyneApp
}

// RefreshMenu rebuilds and updates the main menu based on current settings
func (a *App) RefreshMenu() {
	if a.window == nil {
		return
	}

	menuCallbacks := components.MenuCallbacks{
		OnExport:   func() { a.handleExport() },
		OnImport:   func() { a.handleImport() },
		OnSettings: func() { a.handleSettings() },
		OnSyncNow:  func() { a.handleSyncNow() },
		OnAbout:    func() { a.handleAbout() },
		OnExit:     func() { a.fyneApp.Quit() },
	}
	a.window.SetMainMenu(components.CreateMainMenu(menuCallbacks))
	logging.Debug.Log("Menu refreshed with updated settings")
}
