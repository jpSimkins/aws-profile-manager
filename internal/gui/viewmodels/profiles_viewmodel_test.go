package viewmodels

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"aws-profile-manager/internal/schema"
	"aws-profile-manager/internal/settings"
	"aws-profile-manager/internal/task"
	"aws-profile-manager/internal/test"
)

func TestProfilesViewModel_EmptyDisplaySchema(t *testing.T) {
	test.SetupTestEnvironment(t)

	vm := NewProfilesViewModel()
	s := vm.EmptyDisplaySchema()
	if s == nil {
		t.Fatal("expected non-nil schema")
	}
	if s.Managed == nil {
		t.Fatal("expected non-nil managed collection")
	}
	if s.Managed.Organizations == nil {
		t.Fatal("expected initialized organizations map")
	}
}

func TestProfilesViewModel_LoadDisplaySchema_MissingConfigReturnsError(t *testing.T) {
	test.SetupTestEnvironment(t)

	vm := NewProfilesViewModel()
	displaySchema, configPath, err := vm.LoadDisplaySchema(context.Background(), task.NoOpReporter{})
	if err == nil {
		t.Fatal("expected error when config file is missing")
	}
	if displaySchema != nil {
		t.Fatalf("expected nil schema on error, got %+v", displaySchema)
	}
	expectedPath := filepath.Join(settings.GetAwsDir(), "config")
	if configPath != expectedPath {
		t.Fatalf("configPath = %q, want %q", configPath, expectedPath)
	}
}

func TestProfilesViewModel_StartLoad_TogglesLoadingAndInvokesCallback(t *testing.T) {
	test.SetupTestEnvironment(t)

	vm := NewProfilesViewModel()
	done := make(chan struct{})

	vm.StartLoad(context.Background(), task.NoOpReporter{}, func(displaySchema *schema.Schema, configPath string, err error, loadedAt time.Time) {
		if err == nil {
			t.Error("expected error because config file is missing")
		}
		if configPath == "" {
			t.Error("expected config path in callback")
		}
		if loadedAt.IsZero() {
			t.Error("expected non-zero loadedAt")
		}
		close(done)
	})

	// Wait for callback completion.
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for StartLoad callback")
	}

	if vm.IsLoading {
		t.Fatal("expected IsLoading=false after callback completion")
	}
}

func TestProfilesViewModel_NormalizeForDisplay_MergesManagedAndUnmanaged(t *testing.T) {
	test.SetupTestEnvironment(t)

	vm := NewProfilesViewModel()
	s := &schema.Schema{
		Managed: &schema.ProfileCollection{
			Organizations: map[string]*schema.Organization{
				"managed-org": {
					Name:       "managed-org",
					Partitions: map[string]schema.Partition{"commercial": {Regions: []string{"us-east-1"}, Roles: []string{"Admin"}, Accounts: []schema.Account{{Alias: "a", ID: "1"}}}},
				},
			},
		},
		Unmanaged: &schema.UnmanagedProfiles{
			Above: &schema.ProfileCollection{
				Organizations: map[string]*schema.Organization{
					"personal-org": {
						Name:       "personal-org",
						Partitions: map[string]schema.Partition{"commercial": {Regions: []string{"us-west-2"}, Roles: []string{"ReadOnly"}, Accounts: []schema.Account{{Alias: "b", ID: "2"}}}},
					},
				},
			},
		},
	}

	normalized := vm.normalizeForDisplay(s)
	if normalized == nil || normalized.Managed == nil {
		t.Fatal("expected normalized managed schema")
	}
	if len(normalized.Managed.Organizations) != 2 {
		t.Fatalf("expected 2 organizations after merge, got %d", len(normalized.Managed.Organizations))
	}
	if _, exists := normalized.Managed.Organizations["managed-org"]; !exists {
		t.Fatal("expected managed-org in normalized schema")
	}
	if _, exists := normalized.Managed.Organizations["personal-org"]; !exists {
		t.Fatal("expected personal-org in normalized schema")
	}
}
