package components

import (
	"aws-profile-manager/internal/test"
	"testing"

	fyneTest "fyne.io/fyne/v2/test"
)

func TestNewHeader(t *testing.T) {
	// Create Fyne test app (headless)
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	// Initialize core app state
	test.SetupTestEnvironment(t)

	// Create header
	header := NewHeader()

	if header == nil {
		t.Fatal("NewHeader() should not return nil")
	}

	// Trigger renderer creation to ensure logo is initialized
	_ = fyneTest.WidgetRenderer(header)

	if header.GetLogo() == nil {
		t.Error("Header logo should not be nil after renderer creation")
	}
}

func TestHeader_Renderer(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	test.SetupTestEnvironment(t)

	header := NewHeader()
	renderer := fyneTest.WidgetRenderer(header)

	if renderer == nil {
		t.Fatal("CreateRenderer() should not return nil")
	}

	// Verify renderer has objects
	if len(renderer.Objects()) == 0 {
		t.Error("Header renderer should have objects")
	}
}

func TestHeader_GetLogo(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	test.SetupTestEnvironment(t)

	header := NewHeader()

	// Trigger renderer creation to initialize logo
	_ = fyneTest.WidgetRenderer(header)

	logo := header.GetLogo()

	if logo == nil {
		t.Fatal("GetLogo() should not return nil after renderer creation")
	}

	if logo.Resource == nil {
		t.Error("Logo resource should not be nil")
	}
}

func TestHeader_Refresh(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	test.SetupTestEnvironment(t)

	header := NewHeader()

	// Get renderer to initialize logo
	renderer := fyneTest.WidgetRenderer(header)

	// Get initial logo resource
	initialResource := header.GetLogo().Resource

	if initialResource == nil {
		t.Fatal("Initial logo resource should not be nil")
	}

	// Call refresh (simulates theme change)
	renderer.Refresh()

	// Logo should still have a resource after refresh
	if header.GetLogo().Resource == nil {
		t.Error("Logo resource should not be nil after Refresh()")
	}
}

func TestHeader_getLogoResourceForTheme(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	test.SetupTestEnvironment(t)

	header := NewHeader()

	// Get logo resource
	resource := header.getLogoResourceForTheme()

	if resource == nil {
		t.Error("getLogoResourceForTheme() should not return nil")
	}

	// Verify it's a valid resource with content
	if resource.Name() == "" {
		t.Error("Logo resource should have a name")
	}
}
