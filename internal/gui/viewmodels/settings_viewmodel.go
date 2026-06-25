package viewmodels

import (
	"encoding/json"
	"fmt"
	"math"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"aws-profile-manager/internal/core"
	"aws-profile-manager/internal/logging"
	"aws-profile-manager/internal/settings"
)

// SettingsViewModel manages the state for the settings view.
//
// It owns the form values map, handles initialization from the current settings,
// and provides logic helpers (field key building, dependency evaluation, schema
// ordering) so that the View layer stays free of business/data logic.
type SettingsViewModel struct {
	IsDirty bool                   // Has the user made changes?
	values  map[string]interface{} // Flat "section.field" → value form state
	mu      sync.RWMutex
}

// NewSettingsViewModel creates a new settings view model and registers it.
func NewSettingsViewModel() *SettingsViewModel {
	logging.Debug.Log("\t🔹 Creating settings view model")

	vm := &SettingsViewModel{
		IsDirty: false,
		values:  make(map[string]interface{}),
	}

	core.App.RegisterState("settings-view", vm)

	logging.Debug.Log("\t🔹 Settings view model created")
	return vm
}

// InitializeValues populates the internal values map from the current settings.
//
// Must be called once after construction before the view builds its form widgets.
func (vm *SettingsViewModel) InitializeValues() {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.values = make(map[string]interface{})
	vm.initializeValuesFromSettings(vm.values, settings.Get())
}

// GetValues returns a direct reference to the internal form-values map.
//
// The view may hold this reference and mutate it via map assignment in widget
// OnChanged callbacks; mutations are immediately visible to the ViewModel.
func (vm *SettingsViewModel) GetValues() map[string]interface{} {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.values
}

// SetValue updates a single form value and marks the ViewModel dirty.
func (vm *SettingsViewModel) SetValue(key string, value interface{}) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.values[key] = value
	vm.IsDirty = true
}

// sectionOrder defines the canonical display order of settings sections.
// It is package-level so both GetSectionOrder and buildForm use the same source.
var sectionOrder = []string{"application", "logging", "gui", "sync", "awscli", "terminal"}

// sectionDisplayNames maps section schema keys to their human-readable names
// shown in the navigation sidebar and as section headings inside the form.
var sectionDisplayNames = map[string]string{
	"application": "Application",
	"logging":     "Logging",
	"gui":         "GUI",
	"sync":        "Sync",
	"awscli":      "AWS CLI",
	"terminal":    "Terminal",
}

// GetSectionOrder returns the ordered list of settings section keys.
//
// The order controls both the navigation sidebar and the form rendering.
func (vm *SettingsViewModel) GetSectionOrder() []string {
	return sectionOrder
}

// GetSectionDisplayName returns the human-readable label for a section key.
//
// Falls back to capitalising the first letter of the key when the key is not
// in the known display-name map (e.g. for future sections added dynamically).
func (vm *SettingsViewModel) GetSectionDisplayName(key string) string {
	if name, ok := sectionDisplayNames[key]; ok {
		return name
	}
	// Fallback: capitalise first letter.
	if key == "" {
		return key
	}
	return strings.ToUpper(key[:1]) + key[1:]
}

// BuildFieldKey constructs the flat dot-notation key for a form field.
//
// Examples:
//   - BuildFieldKey("GUI", "theme", "")       → "gui.theme"
//   - BuildFieldKey("Sync", "bucket", "s3.")  → "sync.s3.bucket"
func (vm *SettingsViewModel) BuildFieldKey(sectionName, fieldName, prefix string) string {
	section := strings.ToLower(strings.ReplaceAll(sectionName, " ", ""))
	if prefix != "" {
		return section + "." + prefix + fieldName
	}
	return section + "." + fieldName
}

// EvaluateDependency checks whether a field's visibility dependency is satisfied
// given the current form values.
//
// A nil dependency always returns true (field is unconditionally visible).
// Unknown operators also return true so unrecognised rules never hide fields.
func (vm *SettingsViewModel) EvaluateDependency(dep *settings.FieldDependency, sectionName string) bool {
	if dep == nil {
		return true
	}
	// Build the flat key for the field this dependency watches, then read its
	// current value from the form-values map.
	depKey := vm.BuildFieldKey(sectionName, dep.Field, "")
	vm.mu.RLock()
	depValue, exists := vm.values[depKey]
	vm.mu.RUnlock()

	logging.Debug.Logf("Evaluating dependency: key=%s, exists=%v, value=%v, expected=%v", depKey, exists, depValue, dep.Value)
	if !exists {
		// If the watched field hasn't been initialized yet, hide the dependent field.
		return false
	}
	switch dep.Operator {
	case "equals":
		// String-compare both sides so bools, ints, and strings all work uniformly.
		result := fmt.Sprintf("%v", depValue) == fmt.Sprintf("%v", dep.Value)
		logging.Debug.Logf("  equals check: %v == %v = %v", depValue, dep.Value, result)
		return result
	case "in":
		// dep.Value must be a []interface{} slice; show the field if any element matches.
		if values, ok := dep.Value.([]interface{}); ok {
			depValueStr := fmt.Sprintf("%v", depValue)
			for _, v := range values {
				if fmt.Sprintf("%v", v) == depValueStr {
					return true
				}
			}
		}
		return false
	default:
		// Unknown operator — fail-open so future operators don't hide fields.
		return true
	}
}

// GetSortedFieldNames returns the field names from the given schema map sorted
// by their Order value (ascending).
//
// Fields with a lower Order value appear first in the rendered form. Fields with
// the same Order value are returned in an unspecified but stable-enough order
// for display purposes.
func (vm *SettingsViewModel) GetSortedFieldNames(fields map[string]settings.FieldSchema) []string {
	type fieldWithOrder struct {
		name  string
		order int
	}
	// Collect all fields with their order numbers.
	list := make([]fieldWithOrder, 0, len(fields))
	for name, schema := range fields {
		list = append(list, fieldWithOrder{name, schema.Order})
	}
	// Simple insertion sort — field counts are small so O(n²) is fine.
	for i := 0; i < len(list); i++ {
		for j := i + 1; j < len(list); j++ {
			if list[i].order > list[j].order {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
	names := make([]string, len(list))
	for i, f := range list {
		names[i] = f.name
	}
	return names
}

// SaveSettings persists the current form values to disk and refreshes the GUI.
//
// Returns (requiresRestart, error). requiresRestart is true when at least one
// changed field has RequiresRestart set in its schema.
func (vm *SettingsViewModel) SaveSettings(refreshCallback func()) (bool, error) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	logging.Debug.Log("Saving settings from view model")

	currentSettings := settings.Get()

	// Snapshot the current on-disk values as a flat map BEFORE mutating. Because
	// settings.Get() returns a pointer, any mutation would corrupt the snapshot
	// if we did this after updateSettingsFromMap.
	oldMap := flattenSettings(currentSettings)

	// Apply the form values onto the settings struct via a JSON round-trip.
	if err := vm.updateSettingsFromMap(currentSettings, vm.values); err != nil {
		return false, fmt.Errorf("failed to update settings: %w", err)
	}

	configDir, err := settings.GetConfigDir()
	if err != nil {
		return false, fmt.Errorf("failed to get config directory: %w", err)
	}
	settingsPath := filepath.Join(configDir, "settings.json")

	// Validate and store in-memory first; this runs the settings validator.
	if err := settings.Set(currentSettings); err != nil {
		return false, fmt.Errorf("invalid settings: %w", err)
	}

	// Write the validated settings to disk.
	if err := settings.Save(settingsPath); err != nil {
		return false, fmt.Errorf("failed to save settings: %w", err)
	}

	// Apply logging settings immediately so the new level/debug flag take
	// effect in the running process without requiring a restart.
	logging.UpdateLoggerFromSettings(currentSettings.Logging.LogLevel)
	logging.Debug.SetEnabled(currentSettings.Logging.EnableDebug)

	// Trigger GUI refresh (e.g. re-apply theme, rebuild menu).
	if refreshCallback != nil {
		refreshCallback()
	}

	vm.IsDirty = false
	logging.Debug.Log("\t🔹 Settings saved from view model")

	// Determine whether the user must restart to pick up any changed fields.
	requiresRestart := vm.checkRequiresRestart(oldMap, currentSettings)
	return requiresRestart, nil
}

// checkRequiresRestart returns true if any restart-required field changed.
//
// It compares the pre-save flat map (oldMap) against the freshly-saved settings
// struct. Only fields marked RequiresRestart in their schema are checked.
func (vm *SettingsViewModel) checkRequiresRestart(oldMap map[string]interface{}, newSettings *settings.Settings) bool {
	if newSettings == nil {
		return false
	}
	newMap := flattenSettings(newSettings)
	for sectionKey, schema := range newSettings.GetAllSchemas() {
		for fieldKey, fieldSchema := range schema.Fields {
			if !fieldSchema.RequiresRestart {
				continue
			}
			flatKey := sectionKey + "." + fieldKey
			// String-compare both values so different types (bool, int) compare safely.
			if fmt.Sprintf("%v", oldMap[flatKey]) != fmt.Sprintf("%v", newMap[flatKey]) {
				logging.Debug.Log("Restart-required field changed", "field", flatKey)
				return true
			}
		}
	}
	return false
}

// flattenSettings converts a Settings pointer to a flat "section.field" → value map.
//
// Each section is JSON-marshalled independently and its keys are prefixed with
// the section name, e.g. "gui.theme". This mirrors the key format used by the
// form-values map so old and new values can be compared directly.
func flattenSettings(s *settings.Settings) map[string]interface{} {
	result := make(map[string]interface{})
	// Map schema section keys to their corresponding struct fields.
	sections := map[string]interface{}{
		"application": s.Application,
		"logging":     s.Logging,
		"gui":         s.GUI,
		"sync":        s.Sync,
		"awscli":      s.AwsCLI,
		"terminal":    s.Terminal,
	}
	for sectionKey, section := range sections {
		// Marshal each section to JSON then back to a plain map so we get the
		// same key names (json tags) that the form uses.
		data, err := json.Marshal(section)
		if err != nil {
			continue
		}
		var m map[string]interface{}
		if err := json.Unmarshal(data, &m); err != nil {
			continue
		}
		for k, v := range m {
			result[sectionKey+"."+k] = v
		}
	}
	return result
}

// updateSettingsFromMap applies the flat form-values map back onto the Settings struct.
//
// The flat map (e.g. {"sync.s3.bucket": "my-bucket"}) is first converted to a
// nested map structure matching the JSON shape of Settings, then unmarshalled
// into the struct. This avoids manual field-by-field mapping and automatically
// handles any new settings fields added in future.
func (vm *SettingsViewModel) updateSettingsFromMap(s *settings.Settings, currentValues map[string]interface{}) error {
	nestedMap := vm.buildNestedMap(currentValues)

	// Debug log to help trace sync.local.path issues during development.
	if syncMap, ok := nestedMap["sync"].(map[string]interface{}); ok {
		if localMap, ok := syncMap["local"].(map[string]interface{}); ok {
			logging.Debug.Logf("Sync local settings: %+v", localMap)
		}
	}

	// JSON round-trip: nested map → JSON bytes → Settings struct.
	jsonData, err := json.Marshal(nestedMap)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, s)
}

// buildNestedMap converts a flat "a.b.c" → value map into a nested map structure.
func (vm *SettingsViewModel) buildNestedMap(flatMap map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range flatMap {
		vm.setNestedValue(result, key, value)
	}
	return result
}

// setNestedValue sets a value inside a nested map using dot-notation key.
//
// Intermediate maps are created automatically. If an intermediate key already
// holds a non-map value the traversal stops silently to avoid overwriting
// existing scalar values with a container.
func (vm *SettingsViewModel) setNestedValue(m map[string]interface{}, key string, value interface{}) {
	parts := splitKey(key)
	if len(parts) == 0 {
		return
	}
	// Walk down to the parent map, creating intermediate maps as needed.
	current := m
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		if _, exists := current[part]; !exists {
			current[part] = make(map[string]interface{})
		}
		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			// A scalar already occupies this key — cannot descend further.
			return
		}
	}
	// Set the leaf value.
	current[parts[len(parts)-1]] = value
}

// splitKey splits a dot-notation key into its component parts.
//
// Example: "sync.s3.bucket" → ["sync", "s3", "bucket"].
// An empty key returns nil. Consecutive dots produce no empty parts.
func splitKey(key string) []string {
	if key == "" {
		return nil
	}
	parts := make([]string, 0)
	var current string
	for _, ch := range key {
		if ch == '.' {
			// Flush the current segment on each dot, ignoring empty segments.
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	// Flush the final segment.
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

// MarkDirty marks the settings as modified.
func (vm *SettingsViewModel) MarkDirty() {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.IsDirty = true
}

// GetIsDirty returns whether settings have been modified.
func (vm *SettingsViewModel) GetIsDirty() bool {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.IsDirty
}

// Cleanup unregisters the view model from the core state.
func (vm *SettingsViewModel) Cleanup() {
	logging.Debug.Log("\t🔹 Cleaning up settings view model")
	core.App.UnregisterState("settings-view")
}

// initializeValuesFromSettings populates currentValues from a Settings struct
// by walking all schema-defined sections via reflection.
//
// The schema map drives which sections exist; reflection is used to look up the
// corresponding Go field by its capitalised name (e.g. "gui" → "GUI").
func (vm *SettingsViewModel) initializeValuesFromSettings(currentValues map[string]interface{}, s *settings.Settings) {
	schemas := s.GetAllSchemas()
	settingsValue := reflect.ValueOf(s).Elem()

	for sectionName := range schemas {
		// Convert the schema key (e.g. "gui") to the struct field name (e.g. "GUI").
		fieldName := capitalizeFirst(sectionName)
		field := settingsValue.FieldByName(fieldName)
		if !field.IsValid() {
			// Schema references a section that doesn't exist in the struct — skip.
			continue
		}
		vm.populateValuesFromStruct(currentValues, sectionName, field.Interface())
	}
}

// populateValuesFromStruct recursively populates currentValues from a struct using reflection.
//
// Nested structs are traversed recursively, building dot-notation keys such as
// "sync.s3.bucket". Field names are taken from the json struct tag so they match
// the JSON serialisation format used elsewhere. Unexported fields are skipped.
func (vm *SettingsViewModel) populateValuesFromStruct(currentValues map[string]interface{}, prefix string, structValue interface{}) {
	val := reflect.ValueOf(structValue)
	typ := reflect.TypeOf(structValue)

	// Dereference pointers so callers don't need to.
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return
		}
		val = val.Elem()
		typ = typ.Elem()
	}
	if val.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Skip unexported fields — they cannot be read via Interface().
		if !field.CanInterface() {
			continue
		}

		// Prefer the json tag name so keys match JSON serialisation.
		// Fall back to snake_case of the Go field name when no tag is present.
		fieldName := fieldType.Tag.Get("json")
		if fieldName == "" || fieldName == "-" {
			fieldName = toSnakeCase(fieldType.Name)
		} else {
			// Strip tag options like "omitempty".
			if idx := strings.Index(fieldName, ","); idx != -1 {
				fieldName = fieldName[:idx]
			}
		}

		fullKey := prefix + "." + fieldName

		switch field.Kind() {
		case reflect.Struct:
			// Recurse into nested structs (e.g. SyncSettings.S3).
			vm.populateValuesFromStruct(currentValues, fullKey, field.Interface())
		case reflect.Bool:
			currentValues[fullKey] = field.Bool()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			// Normalise all integer types to plain int for consistent form handling.
			currentValues[fullKey] = int(field.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			uintValue := field.Uint()
			if uintValue > uint64(math.MaxInt) {
				currentValues[fullKey] = math.MaxInt
			} else {
				currentValues[fullKey] = int(uintValue)
			}
		case reflect.String:
			currentValues[fullKey] = field.String()
		case reflect.Float32, reflect.Float64:
			currentValues[fullKey] = field.Float()
		default:
			// Store other types (slices, maps, etc.) as-is.
			currentValues[fullKey] = field.Interface()
		}
	}
}

// capitalizeFirst returns the section name as the corresponding Settings struct field name.
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	switch s {
	case "gui":
		return "GUI"
	case "awscli":
		return "AwsCLI"
	default:
		return strings.ToUpper(s[:1]) + s[1:]
	}
}

// toSnakeCase converts a CamelCase identifier to snake_case.
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
