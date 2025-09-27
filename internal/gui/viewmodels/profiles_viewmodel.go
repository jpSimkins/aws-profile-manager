package viewmodels

import (
	"context"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"aws-profile-manager/internal/core"
	"aws-profile-manager/internal/logging"
	"aws-profile-manager/internal/profiles"
	"aws-profile-manager/internal/schema"
	"aws-profile-manager/internal/settings"
	"aws-profile-manager/internal/task"
)

// ProfilesViewModel manages all business logic for the Profiles view.
//
// Owns schema loading, normalization, and async orchestration.
// The view layer only handles UI wiring and rendering.
type ProfilesViewModel struct {
	IsLoading bool
	mu        sync.RWMutex
}

// NewProfilesViewModel creates and registers a profiles view model.
func NewProfilesViewModel() *ProfilesViewModel {
	logging.Debug.Log("\t🔹 Creating profiles view model")

	vm := &ProfilesViewModel{
		IsLoading: false,
	}
	core.App.RegisterState("profiles-view", vm)

	logging.Debug.Log("\t🔹 Profiles view model created")
	return vm
}

// ConfigPath returns the AWS config file path.
func (vm *ProfilesViewModel) ConfigPath() string {
	return filepath.Join(settings.GetAwsDir(), "config")
}

// FormatLoadedAt formats a timestamp for display.
func (vm *ProfilesViewModel) FormatLoadedAt(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// LoadDisplaySchema loads and normalizes schema from AWS config using task.Run.
//
// Parameters:
//   - ctx: Context for cancellation
//   - reporter: Task reporter for progress updates
//
// Returns: (displaySchema, configPath, error)
func (vm *ProfilesViewModel) LoadDisplaySchema(ctx context.Context, reporter task.Reporter) (*schema.Schema, string, error) {
	configPath := vm.ConfigPath()
	currentSettings := settings.Get()
	appSettings := currentSettings.Application

	loaderConfig := profiles.Config{
		ConfigPath:  configPath,
		AwsDir:      settings.GetAwsDir(),
		StartMarker: appSettings.GetFormattedStartMarker(),
		EndMarker:   appSettings.GetFormattedEndMarker(),
	}

	var displaySchema *schema.Schema

	logging.Debug.Log("Profiles view model: Loading display schema",
		"configPath", configPath,
	)

	loadTask := &task.FunctionTask{
		Name: "load-profiles-schema",
		Fn: func(runCtx context.Context, runReporter task.Reporter) ([]byte, error) {
			runReporter.ReportStatus("Reading AWS config")

			schemaReader := profiles.NewSchemaReader(loaderConfig)
			loadedSchema, err := schemaReader.Read(
				runCtx,
				profiles.SchemaReadOptions{
					IncludeManaged:        true,
					IncludeUnmanagedAbove: true,
					IncludeUnmanagedBelow: true,
				},
				runReporter,
			)
			if err != nil {
				return nil, err
			}

			if loadedSchema == nil {
				displaySchema = vm.emptyDisplaySchema()
			} else {
				displaySchema = vm.normalizeForDisplay(loadedSchema)
			}

			runReporter.ReportStatus("Schema loaded")
			return []byte("ok"), nil
		},
	}

	if _, err := task.Run(ctx, loadTask, reporter); err != nil {
		logging.Debug.Log("Profiles view model: Schema load failed",
			"configPath", configPath,
			"error", err,
		)
		return nil, configPath, err
	}

	logging.Debug.Log("Profiles view model: Schema loaded successfully",
		"configPath", configPath,
	)
	return displaySchema, configPath, nil
}

// StartLoad loads schema asynchronously with callback.
//
// Parameters:
//   - ctx: Context for cancellation
//   - reporter: Task reporter for progress updates
//   - onLoadComplete: Callback receives (displaySchema, configPath, error, loadedAt)
func (vm *ProfilesViewModel) StartLoad(
	ctx context.Context,
	reporter task.Reporter,
	onLoadComplete func(*schema.Schema, string, error, time.Time),
) {
	logging.Debug.Log("Profiles view model: StartLoad triggered")

	vm.mu.Lock()
	vm.IsLoading = true
	vm.mu.Unlock()

	var displaySchema *schema.Schema
	var configPath string

	asyncTask := &task.FunctionTask{
		Name: "load-profiles-schema-async",
		Fn: func(runCtx context.Context, runReporter task.Reporter) ([]byte, error) {
			var err error
			displaySchema, configPath, err = vm.LoadDisplaySchema(runCtx, runReporter)
			if err != nil {
				return nil, err
			}
			return []byte("ok"), nil
		},
	}

	task.RunAsync(ctx, asyncTask, reporter, func(_ *task.Result, err error) {
		vm.mu.Lock()
		vm.IsLoading = false
		vm.mu.Unlock()

		logging.Debug.Log("Profiles view model: StartLoad complete",
			"error", err,
		)

		if onLoadComplete != nil {
			onLoadComplete(displaySchema, configPath, err, time.Now())
		}
	})
}

func (vm *ProfilesViewModel) emptyDisplaySchema() *schema.Schema {
	return &schema.Schema{
		Managed: &schema.ProfileCollection{
			Organizations: map[string]*schema.Organization{},
		},
	}
}

func (vm *ProfilesViewModel) normalizeForDisplay(sourceSchema *schema.Schema) *schema.Schema {
	if sourceSchema == nil {
		return vm.emptyDisplaySchema()
	}

	normalized := vm.emptyDisplaySchema()

	if sourceSchema.Managed != nil {
		vm.mergeIntoManaged(normalized.Managed, sourceSchema.Managed)
	}

	if sourceSchema.Unmanaged != nil {
		if sourceSchema.Unmanaged.Above != nil {
			vm.mergeIntoManaged(normalized.Managed, sourceSchema.Unmanaged.Above)
		}
		if sourceSchema.Unmanaged.Below != nil {
			vm.mergeIntoManaged(normalized.Managed, sourceSchema.Unmanaged.Below)
		}
	}

	return normalized
}

func (vm *ProfilesViewModel) mergeIntoManaged(target *schema.ProfileCollection, source *schema.ProfileCollection) {
	if target == nil || source == nil {
		return
	}

	if target.Organizations == nil {
		target.Organizations = map[string]*schema.Organization{}
	}

	for orgAlias, sourceOrg := range source.Organizations {
		if sourceOrg == nil {
			continue
		}

		targetOrg, exists := target.Organizations[orgAlias]
		if !exists || targetOrg == nil {
			targetOrg = &schema.Organization{
				Name:        sourceOrg.Name,
				Description: sourceOrg.Description,
				Partitions:  map[string]schema.Partition{},
			}
			target.Organizations[orgAlias] = targetOrg
		}

		if targetOrg.Name == "" {
			targetOrg.Name = sourceOrg.Name
		}
		if targetOrg.Description == "" {
			targetOrg.Description = sourceOrg.Description
		}

		if targetOrg.Partitions == nil {
			targetOrg.Partitions = map[string]schema.Partition{}
		}

		for partitionName, sourcePartition := range sourceOrg.Partitions {
			targetPartition, partitionExists := targetOrg.Partitions[partitionName]
			if !partitionExists {
				targetOrg.Partitions[partitionName] = sourcePartition
				continue
			}

			targetPartition = vm.mergePartition(targetPartition, sourcePartition)
			targetOrg.Partitions[partitionName] = targetPartition
		}
	}
}

func (vm *ProfilesViewModel) mergePartition(target schema.Partition, source schema.Partition) schema.Partition {
	if target.URL == "" {
		target.URL = source.URL
	}
	if target.DefaultRegion == "" {
		target.DefaultRegion = source.DefaultRegion
	}

	target.Regions = vm.mergeStrings(target.Regions, source.Regions)
	target.Roles = vm.mergeStrings(target.Roles, source.Roles)
	target.Accounts = vm.mergeAccounts(target.Accounts, source.Accounts)

	return target
}

func (vm *ProfilesViewModel) mergeStrings(existing []string, incoming []string) []string {
	merged := append([]string{}, existing...)
	seen := make(map[string]struct{})

	for _, s := range merged {
		seen[s] = struct{}{}
	}

	for _, s := range incoming {
		if _, exists := seen[s]; !exists {
			seen[s] = struct{}{}
			merged = append(merged, s)
		}
	}

	sort.Strings(merged)
	return merged
}

func (vm *ProfilesViewModel) mergeAccounts(existing []schema.Account, incoming []schema.Account) []schema.Account {
	merged := append([]schema.Account{}, existing...)
	seen := make(map[string]struct{})

	for _, a := range merged {
		seen[vm.accountKey(a)] = struct{}{}
	}

	for _, a := range incoming {
		key := vm.accountKey(a)
		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}
			merged = append(merged, a)
		}
	}

	return merged
}

func (vm *ProfilesViewModel) accountKey(a schema.Account) string {
	return a.Alias + "|" + a.ID
}

// EmptyDisplaySchema returns an empty schema for initial display.
func (vm *ProfilesViewModel) EmptyDisplaySchema() *schema.Schema {
	return &schema.Schema{
		Managed: &schema.ProfileCollection{
			Organizations: map[string]*schema.Organization{},
		},
	}
}
