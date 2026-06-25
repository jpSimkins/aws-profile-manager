# AI Assistant Instructions - AWS Profile Manager

## 📋 Table of Contents
1. [Project Overview](#project-overview)
2. [Critical Rules](#critical-rules)
3. [Project Structure](#project-structure)
4. [Development Workflow](#development-workflow)
5. [Testing Standards](#testing-standards)
6. [Architecture Patterns](#architecture-patterns)
7. [Package Documentation](#package-documentation)

---

## 🎯 Project Overview

AWS Profile Manager is a Go CLI/GUI tool for managing AWS CLI profiles with centralized configuration distribution.

### Quick Facts
- **Language**: Go 1.26+
- **Architecture**: [Standard Go project layout](https://github.com/golang-standards/project-layout) (`cmd/`, `internal/`)
- **Frameworks**: Cobra (CLI), Fyne 2.7.0 (GUI)
- **Test Coverage**: 71.5% (target: 95%+)
- **Platform**: Cross-platform (Linux, macOS, Windows)

### Core Features
- ✅ **Profile Installation**: Generate AWS profiles from centralized configs
- ✅ **Profile Export/Import**: Backup and restore AWS configurations (managed & personal profiles)
- ✅ **Settings Export/Import**: Backup and restore application settings
- ✅ **Config Sync**: HTTP/S3 remote configuration fetching
- ✅ **Profile Introspection**: List and filter existing AWS profiles
- ✅ **SSO Session Management**: Track SSO session status
- ✅ **GUI Interface**: MVVM architecture with Fyne (Phase 1 complete)

### Main Commands
```bash
aws-profile-manager install          # Install profiles from config
aws-profile-manager export           # Export AWS profiles to JSON
aws-profile-manager import           # Import profiles from backup
aws-profile-manager profiles         # List/filter AWS profiles
aws-profile-manager sessions         # Check SSO session status
aws-profile-manager sync fetch       # Fetch remote configuration
aws-profile-manager gui              # Launch GUI interface
```

---

## ⚠️ Critical Rules

### � BANNED PATTERNS — Never Generate These

The following patterns are **absolutely forbidden** in this codebase. Before writing any code, check this table.

| ❌ Banned | ✅ Use Instead | Why |
|---|---|---|
| `exec.Command(...)` | `task.SubprocessTask{...}` | No context, no progress, no env isolation |
| `exec.CommandContext(...)` | `task.SubprocessTask{...}` | Use the task system, not raw exec |
| `go func() { doWork() }()` in business logic | `task.FunctionTask{...}` | Business logic must be testable and reporter-aware |
| `go func() { doWork() }()` in GUI without `fyne.Do()` | `go func() { ...; fyne.Do(func(){...}) }()` | GUI updates must be on main thread |
| `fmt.Println(...)` / `log.Print*(...)` | `logging.Log.Info(...)` | Only the logging package produces output |
| `panic(...)` | `return logging.Log.ErrorfWithDetails(...)` | Never panic |
| `os.Getenv("AWS_PROFILE_MANAGER_*")` | `settings.GetAwsDir()` / `settings.GetCacheDir()` | Always use settings helpers |
| `import "aws-profile-manager/internal/settings"` in business logic | Accept a config struct parameter | Dependency injection rule |
| Hardcoded `# START - Managed by...` strings | `settings.GetApplication().GetFormattedStartMarker()` | Single source of truth |
| Inline test schemas | `schematest.New*()` functions | Use centralized fixtures |

### �🚨 MANDATORY - Read Before Making ANY Changes

#### 1. Logging Package (REQUIRED)
**NEVER use `fmt.Print*`, `log.Print*`, or `panic()` - ALWAYS use the logging package.**

```go
// ✅ CORRECT: Key-value logging (ALWAYS use multi-line for clarity)
logging.Log.Info("Operation completed",
	"duration", "2s",
	"files", 10,
)
logging.Log.Success("Profile created",
	"name", "prod",
)
logging.Log.Warn("Cache miss",
	"key", cacheKey,
)
logging.Log.Error("Failed to process",
	"error", err,
)

// ✅ CORRECT: Formatted string logging
logging.Log.Infof("Processing %d files", count)
logging.Log.Successf("Saved to %s", filepath)

// ✅ CORRECT: Debug logging (only when AWS_PROFILE_MANAGER_DEBUG=1)
logging.Debug.Log("Details",
    "status", "active",
    "count", 5,
)
logging.Debug.Logf("Cache hit rate: %.2f%%", hitRate)

// ❌ WRONG: Direct output
fmt.Println("message")      // Use logging.Log.Info()
log.Printf("error: %v", e)  // Use logging.Log.Error()
panic("error")              // Use logging.Log.Error() + return err
```

**Rules**:
- Key-value pairs: Use non-f functions (`Info`, `Success`, `Warn`, `Error`)
- Format strings: Use f functions (`Infof`, `Successf`, `Warnf`, `Errorf`)
- **Multi-line formatting**: Always put key-value pairs on separate lines for clarity
- Never mix patterns (format string in non-f or key-value in f function)

**Error Handling Pattern - The Golden Rule**:

**If a function RETURNS an error → use `fmt.Errorf()`**
- All internal functions that return errors use `fmt.Errorf()` with `%w`
- Never use `logging.Log.Error*()` in functions that return errors
- Errors bubble up the call stack, building context

**If a function PRESENTS error to user → use `logging.Log.Error*()`**
- Only at final consumers (CLI commands, GUI handlers, entry points)
- Where error stops being returned and becomes user-visible
- This is the ONLY place user sees error output

**Simple Decision Tree**:
```
Does this function return an error?
├─ YES → Use fmt.Errorf("message: %w", err)
└─ NO  → Use logging.Log.Error*() if showing to user
```

**Examples**:
```go
// ✅ CORRECT: Internal function returns error
func (c *Cache) Get() (*CacheEntry, error) {
    data, err := os.ReadFile(cacheFile)
    if err != nil {
        return nil, fmt.Errorf("failed to read cache file: %w", err)  // Returns, uses fmt
    }
    // ... processing
}

// ✅ CORRECT: Internal validation returns error
func (s *Settings) Validate() error {
    if s.Field == "" {
        return fmt.Errorf("field is required")  // Returns, uses fmt
    }
    return nil
}

// ✅ CORRECT: Consumer (CLI) shows error to user
func runCommand(cmd *cobra.Command, args []string) error {
    result, err := mypackage.DoWork(ctx, opts, reporter)
    if err != nil {
        return logging.Log.ErrorfWithDetails("operation failed", err)  // Final consumer, logs
    }
    displayResults(result)
}

// ✅ CORRECT: Consumer (entry point) shows error to user
func Load(path string) error {
    if err := s.Validate(); err != nil {
        return logging.Log.ErrorfWithDetails("validation failed", err)  // Final consumer, logs
    }
}

// ❌ WRONG: Internal function logs instead of returning
func (c *Cache) Get() (*CacheEntry, error) {
    data, err := os.ReadFile(cacheFile)
    if err != nil {
        logging.Log.Error("failed to read", "error", err)  // DON'T LOG HERE
        return nil, err
    }
}

// ❌ WRONG: Consumer doesn't log (user sees nothing)
func runCommand(cmd *cobra.Command, args []string) error {
    result, err := mypackage.DoWork(ctx)
    if err != nil {
        return err  // User sees nothing! Must log here
    }
}
```

**Key Principle**: 
- **Functions that return errors**: Use `fmt.Errorf()` with `%w` to build error context
- **Final consumers only**: Use `logging.Log.Error*()` to present errors to user
- **User only sees output from logging package** - never from `fmt` or `log`
- **One logging point per error path** - at the consumer, not in internal functions

#### 2. Naming Conventions (REQUIRED)
**ALWAYS use PascalCase for multi-word names - NEVER use ALL_CAPS or snake_case.**

```go
// ✅ CORRECT: PascalCase for variables, functions, types
type HttpClient struct {        // HttpClient, not HTTPClient
    BaseUrl string              // BaseUrl, not BaseURL
}

func NewHttpFetcher() *HttpFetcher { ... }     // HttpFetcher, not HttpFetcher
func GetSsoSession() *SsoSession { ... }       // SsoSession, not SSOSession
func ParseAwsConfig() error { ... }            // AwsConfig, not AWSConfig

var cliReporter = task.CliReporter{}           // CliReporter, not CliReporter
var s3Bucket = "my-bucket"                     // s3Bucket, not S3Bucket
var apiKey = "secret"                          // apiKey, not APIKey

// ✅ CORRECT: PascalCase for file names
TestHttpFetcher.md        // Not TEST_HTTP_FETCHER.md or test-http-fetcher.md
E2eTesting.md            // Not E2E_TESTING.md
HttpClient.go            // Not HTTPClient.go or http_client.go

// ❌ WRONG: ALL_CAPS (looks like screaming)
type HTTPClient struct { ... }           // Use HttpClient
var CliReporter task.Reporter            // Use CliReporter
E2E_TESTING.md                          // Use E2eTesting.md
TEST_STRATEGY.md                        // Use TestStrategy.md

// ❌ WRONG: Snake_case (not Go style)
http_fetcher.go                         // Use httpFetcher.go (private) or HttpFetcher.go (public)
cli_reporter                            // Use cliReporter

// ❌ WRONG: kebab-case (not Go style)
http-fetcher.go                         // Use httpFetcher.go or HttpFetcher.go
```

**Acronym Rules**:
- **Start of name**: Lowercase first letter, then PascalCase: `httpClient`, `ssoSession`, `awsConfig`
- **Middle/end of name**: PascalCase: `NewHttpClient`, `GetSsoSession`, `ParseAwsConfig`
- **Exported types**: First letter uppercase: `HttpClient`, `SsoSession`, `AwsConfig`

**File Naming**:
- **Go files**: Follow Go convention (`http_fetcher.go` for private, `HttpFetcher.go` would be unusual)
- **Documentation**: PascalCase (`TestStrategy.md`, `E2eTesting.md`)
- **NO ALL_CAPS**: Never use `README.md`, `E2E_TESTING.md`, `TEST_STRATEGY.md`

**Why PascalCase**:
- ✅ Easier to read (not screaming)
- ✅ Consistent with Go naming conventions
- ✅ Professional appearance
- ✅ Better for parsing (clear word boundaries)

#### 3. Test Isolation (REQUIRED)
**ALL tests with file I/O MUST use `test.SetupTestEnvironment(t)`**

```go
import (
    "aws-profile-manager/internal/test"
    schematest "aws-profile-manager/internal/schema/test"
)

// ✅ CORRECT: Standard test pattern
func TestMyFunction(t *testing.T) {
    test.SetupTestEnvironment(t)  // Creates temp dirs, sets env vars
    
    // Use centralized test schemas
    schema := schematest.NewManagedOnly()
    
    // All file operations now use isolated temp directories
    // Automatic cleanup via t.Cleanup()
}

// ✅ CORRECT: Tests needing app initialization
func TestWithApp(t *testing.T) {
    test.SetupTestEnvironment(t)
    
    // Initialize app after environment setup
    if err := core.App.Initialize(nil); err != nil {
        t.Fatalf("Failed to initialize: %v", err)
    }
    
    // Now safe to use settings, state, etc.
}

// ✅ CORRECT: Tests requiring valid settings
func TestWithSettings(t *testing.T) {
    test.SetupTestEnvironment(t)
    
    if err := core.App.Initialize(nil); err != nil {
        t.Fatalf("Failed to initialize: %v", err)
    }
    
    // Settings validate on save - must use proper values
    currentSettings := settings.Get()
    currentSettings.Sync.Strategy = "http"  // Valid strategy
    currentSettings.Sync.HttpUrl = "https://example.com/config.json"  // Valid URL
    settings.Set(currentSettings)  // Validation happens here
    
    // Now test with valid settings
}

// ❌ WRONG: Using real paths
func TestBad(t *testing.T) {
    os.WriteFile("~/.aws/config", data, 0644)  // Writes to real files!
}

// ❌ WRONG: Invalid settings (will fail validation)
func TestBadSettings(t *testing.T) {
    settings := settings.Get()
    settings.Sync.Strategy = "invalid"  // Invalid strategy
    settings.Set(settings)  // Will fail validation!
}
```

**What `SetupTestEnvironment` does**:
- Creates temp directories: `config/`, `config/cache/`, `.aws/`, `Desktop/`
- Sets environment variables to point to temp directories
- Ensures complete test isolation
- Automatic cleanup on test completion

**Settings Validation**:
- Settings validate on `Set()` - must use proper values
- Invalid settings (wrong strategy, malformed URLs, etc.) will fail
- Always use valid enum values and proper data types in tests

**Test Helper Functions**:
```go
test.GetTestConfigDir(t)      // $CONFIG_DIR
test.GetTestCacheDir(t)       // $CONFIG_DIR/cache
test.GetTestAwsDir(t)         // $AWS_DIR
test.GetTestDesktopDir(t)     // $DESKTOP_DIR
test.GetTestAwsConfigPath(t)  // $AWS_DIR/config
```

#### 3. CLI/GUI Architecture (REQUIRED)
**CLI and GUI are thin presentation layers. ALL business logic lives in packages.**

```go
// ❌ WRONG: CLI orchestrating business logic
func runCommand(cmd *cobra.Command, args []string) error {
    extractor := awscli.NewExtractor()        // ❌ Creating objects
    data, _ := extractor.ExtractFromFile()    // ❌ Loading data
    filtered := filter.Apply(data)            // ❌ Processing data
    display(filtered)                         // ✅ Display OK
}

// ✅ CORRECT: CLI calls package API
func runCommand(cmd *cobra.Command, args []string) error {
    // Step 1: Parse flags (CLI responsibility)
    criteria := awscli.FilterCriteria{
        AccountIDs: getStringSlice(cmd, "account-id"),
        Regions:    getStringSlice(cmd, "region"),
    }
    
    // Step 2: Call ONE package function (all logic in package)
    result, err := awscli.ListProfiles(criteria)
    if err != nil {
        return err
    }
    
    // Step 3: Display results (CLI responsibility)
    return displayProfiles(result)
}
```

**Available Package APIs**:
```go
// awscli package (internal/awscli/api.go)
result, err := awscli.ListProfiles(criteria)
result, err := awscli.GetSessionStatus(options)

// profiles package (internal/profiles)
installer := profiles.NewInstaller(config)
result, err := installer.Install(ctx, options, reporter)
remover := profiles.NewRemover(config)
result, err := remover.Remove(ctx, options, reporter)

// backup package (internal/backup/)
result, err := backup.ExportProfiles(ctx, cfg, opts, reporter)
result, err := backup.ImportProfiles(ctx, cfg, opts, reporter)

// sync package (internal/sync/)
cfg := sync.ConfigFromSettings(&settings.Get().Sync)
result, err := sync.Sync(ctx, cfg, opts, reporter)
```

**Layer Responsibilities**:
- **CLI/GUI**: Parse input → Call ONE package function → Display results
- **Packages**: Create objects → Load data → Process → Return result struct
- **NEVER**: CLI/GUI creates extractors/filters/managers or orchestrates multiple calls

#### 4. Task Package Integration (REQUIRED)
**NEVER use `exec.Command`, `exec.CommandContext`, or raw `go func()` goroutines in business logic. ALWAYS use the task package.**

This is the most commonly violated rule. Every AI-generated goroutine in business logic and every `exec.Command` call is wrong by default.

**The substitution rule is simple**:
- External command → `task.SubprocessTask{}`
- Long-running Go work → `task.FunctionTask{}`  
- Goroutine in GUI handler → `go func() { ...; fyne.Do(func(){...}) }()`

```go
// ❌ WRONG: exec.Command — no context, no progress, no env isolation
func DoWork(opts Options) error {
    cmd := exec.Command("aws", "s3", "cp", "s3://bucket/key", "-")
    output, err := cmd.Output()
    return err
}

// ❌ WRONG: exec.CommandContext — still bypasses the task system
func DoWork(ctx context.Context, opts Options) error {
    cmd := exec.CommandContext(ctx, "aws", "s3", "cp", "s3://bucket/key", "-")
    output, err := cmd.Output()
    return err
}

// ❌ WRONG: Raw goroutine in business logic — untestable, no reporter
func DoWork(opts Options) {
    go func() {
        result := expensiveOperation()
        // No way to report progress or cancel
    }()
}

// ✅ CORRECT: Business logic uses SubprocessTask for external commands
func DoWork(ctx context.Context, cfg WorkConfig, reporter task.Reporter) (*Result, error) {
    reporter.ReportStatus("Fetching configuration...")

    t := &task.SubprocessTask{
        Name:    "aws",
        Args:    []string{"s3", "cp", "s3://" + cfg.Bucket + "/" + cfg.Key, "-"},
        Env:     map[string]string{"AWS_PROFILE": cfg.Profile},
        Timeout: 30 * time.Second,
    }

    result, err := task.Run(ctx, t, reporter)
    if err != nil {
        return nil, fmt.Errorf("fetch failed: %w", err)
    }

    reporter.ReportStatus("Fetch complete")
    return &Result{Data: result.Output}, nil
}

// ✅ CORRECT: Business logic uses FunctionTask for long-running Go work
func ProcessProfiles(ctx context.Context, cfg WorkConfig, reporter task.Reporter) (*Result, error) {
    reporter.ReportStatus("Processing profiles...")

    t := &task.FunctionTask{
        Name: "process-profiles",
        Fn: func(ctx context.Context, r task.Reporter) ([]byte, error) {
            for i, p := range cfg.Profiles {
                select {
                case <-ctx.Done():
                    return nil, ctx.Err()  // MUST check for cancellation
                default:
                }
                r.ReportProgress(int64(i+1), int64(len(cfg.Profiles)))
                if err := process(p); err != nil {
                    return nil, fmt.Errorf("profile %s: %w", p.Name, err)
                }
            }
            return nil, nil
        },
    }

    _, err := task.Run(ctx, t, reporter)
    if err != nil {
        return nil, fmt.Errorf("processing failed: %w", err)
    }
    return &Result{}, nil
}

// ✅ CORRECT: CLI passes CliReporter — output is automatic
func runMyCommand(cmd *cobra.Command, args []string) error {
    cfg := buildConfig()
    result, err := mypackage.DoWork(ctx, cfg, task.CliReporter{})
    if err != nil {
        return logging.Log.ErrorfWithDetails("operation failed", err)
    }
    displayResults(result)
    return nil
}

// ✅ CORRECT: GUI goroutine with fyne.Do for UI updates
func (v *MyView) runWork() {
    go func() {
        result, err := mypackage.DoWork(ctx, cfg, task.NoOpReporter{})

        fyne.Do(func() {  // ALL GUI updates MUST be inside fyne.Do()
            v.progressDialog.Hide()
            if err != nil {
                dialog.ShowError(err, v.window)
                return
            }
            v.showResult(result)
        })
    }()
}
```

**Reporter selection**:
- **Business logic**: Accept `task.Reporter` as parameter — never hard-code a reporter type
- **CLI commands**: Pass `task.CliReporter{}` — prints status/progress to console automatically
- **GUI handlers**: Pass `task.NoOpReporter{}` when managing your own progress UI, or `task.ChannelReporter` to pipe status into a label
- **Tests**: Pass `task.NoOpReporter{}` — discard all output

**Context cancellation in FunctionTask**: The `Fn` function MUST check `ctx.Done()` inside any loop or blocking call. A function that ignores context will block indefinitely after the user cancels.

#### 5. Dependency Injection (REQUIRED)
**Use dependency injection - NEVER import settings package in business logic.**

```go
// ✅ CORRECT: Accept configuration as parameters
func DoWork(ctx context.Context, cfg WorkConfig, reporter task.Reporter) (*Result, error) {
    // Use cfg.SomeValue, not settings.Get().SomeValue
    reporter.ReportStatus(fmt.Sprintf("Using config: %s", cfg.SomeValue))
    // ... work ...
}

// ✅ CORRECT: CLI/GUI builds config from settings
func runMyCommand(cmd *cobra.Command, args []string) error {
    // Only CLI/GUI layers access settings
    currentSettings := settings.Get()
    
    // Build configuration object
    cfg := mypackage.WorkConfig{
        SomeValue: currentSettings.MySection.SomeValue,
        OtherValue: currentSettings.MySection.OtherValue,
    }

    // Pass config to business logic
    reporter := task.CliReporter{}
    result, err := mypackage.DoWork(ctx, cfg, reporter)
    // ...
}

// ✅ CORRECT: Package defines its own config struct
type WorkConfig struct {
    SomeValue  string
    OtherValue int
    Timeout    time.Duration
}

// ❌ WRONG: Business logic imports and uses settings directly
func DoWork(ctx context.Context, reporter task.Reporter) error {
    settings := settings.Get()  // ❌ Tight coupling!
    value := settings.MySection.SomeValue
    // ...
}

// ❌ WRONG: Business logic depends on settings package
import (
    "aws-profile-manager/internal/settings"  // ❌ In business logic!
)
```

**Why**:
- ✅ **Testability**: Easy to test with mock configurations
- ✅ **Reusability**: Package can be used outside this app
- ✅ **Clarity**: Dependencies are explicit in function signatures
- ✅ **Decoupling**: Business logic doesn't know about settings structure

**Where settings CAN be used**:
- ✅ `internal/cli/` - CLI commands build config from settings
- ✅ `internal/gui/` - GUI views build config from settings
- ✅ `cmd/` - Main entry point initializes settings
- ❌ `internal/awscli/`, `internal/sync/`, `internal/profiles/`, etc. - NO!

**Example: sync package**
```go
// ✅ CORRECT: sync.go defines SyncConfig
type SyncConfig struct {
    Strategy   Strategy
    HTTPUrl    string
    S3Bucket   string
    // ... all needed config
}

// ✅ CORRECT: Helper in sync package converts settings → config
func ConfigFromSettings(syncSettings *settings.SyncSettings) SyncConfig {
	cacheDir, _ := settings.GetCacheDir() // Ignore error, will be validated later
    return SyncConfig{
        Strategy: Strategy(syncSettings.Strategy),
        HTTPUrl:  syncSettings.HTTP.URL,
        S3Bucket: syncSettings.S3.Bucket,
		CacheDir: cacheDir,
        // ...
    }
}

// ✅ CORRECT: Main API accepts config, not settings
func Sync(ctx context.Context, cfg SyncConfig, opts Options, reporter task.Reporter) (*Result, error)

// ✅ CORRECT: CLI builds config and passes it
func runSyncFetch(cmd *cobra.Command, args []string) error {
    syncSettings := settings.Get().Sync
    cfg := sync.ConfigFromSettings(&syncSettings)  // Convert here
    result, err := sync.Sync(ctx, cfg, opts, reporter)
    // ...
}
```

#### 6. Marker Handling (REQUIRED)
**NEVER hardcode markers - always use settings helpers**

```go
// ✅ CORRECT: Use settings helpers

```go
// ✅ CORRECT: Use settings helpers
appSettings := settings.GetApplication()
startMarker := appSettings.GetFormattedStartMarker()  // "# START - Managed..."
endMarker := appSettings.GetFormattedEndMarker()      // "# END - Managed..."

// ❌ WRONG: Hardcoded markers
content := "# START - Managed by AWS Profile Manager\n..."

// ❌ WRONG: Manual formatting
marker := "# " + appSettings.ManagedSectionStart  // Helper does this!
```

**Why**: Single source of truth, easy to change globally, tests validate actual behavior.

#### 7. Error Handling in Tests
**ALL errors in tests must be explicitly handled**

```go
// ✅ CORRECT: Explicitly ignore errors
_ = os.WriteFile(path, data, 0644)
_ = cmd.Flags().Set("verbose", "true")

// ✅ CORRECT: Check errors when needed
if err := os.MkdirAll(dir, 0755); err != nil {
    t.Fatalf("Setup failed: %v", err)
}

// ✅ CORRECT: Use t.Fatal for nil checks (prevents SA5011)
if ptr == nil {
    t.Fatal("Pointer should not be nil")
}
ptr.SomeMethod()  // Safe - linter knows ptr is not nil

// ❌ WRONG: Unchecked errors
os.WriteFile(path, data, 0644)  // Lint error!
```

#### 8. Code Documentation (REQUIRED)
**ALL code MUST be thoroughly commented using godoc format.**

```go
// ✅ CORRECT: Package documentation
// Package awscli provides AWS CLI profile extraction and management.
//
// This package handles reading AWS CLI configuration files, extracting
// profile information, and caching results for performance.
package awscli

// ✅ CORRECT: Exported function documentation
// ListProfiles retrieves and filters AWS CLI profiles from configuration.
//
// The function reads profiles from ~/.aws/config, applies the specified
// filter criteria, and returns matching profiles with their metadata.
//
// Parameters:
//   - criteria: Filter criteria for profile selection
//
// Returns:
//   - *ProfilesResult: Filtered profiles with session information
//   - error: Any error encountered during extraction or filtering
//
// Example:
//
//	result, err := awscli.ListProfiles(awscli.FilterCriteria{
//	    AccountIDs: []string{"123456789012"},
//	})
func ListProfiles(criteria FilterCriteria) (*ProfilesResult, error) {
    // Implementation
}

// ✅ CORRECT: Struct documentation
// ProfilesResult contains the results of profile listing operation.
//
// This struct provides all information needed to display profile
// details, including SSO session status and configuration paths.
type ProfilesResult struct {
    // Profiles contains all matching AWS CLI profiles
    Profiles []Profile
    
    // SsoSessions contains SSO session configurations
    SsoSessions []SsoSession
    
    // SessionStatus provides SSO session validity information
    SessionStatus map[string]bool
    
    // ConfigPath is the path to the AWS config file
    ConfigPath string
}

// ❌ WRONG: Missing or inadequate comments
func DoSomething(x int) error {  // What does this do?
    // ...
}

type MyStruct struct {  // No documentation
    Field string        // What is this field?
}
```

**Documentation Requirements**:
- **Package-level**: Every package needs a package doc comment
- **Exported symbols**: All exported functions, types, constants, variables MUST be documented
- **Parameters**: Document all parameters and return values
- **Examples**: Include usage examples for complex functions
- **Why, not what**: Explain the purpose and reasoning, not just what the code does
- **godoc format**: Use standard Go documentation conventions

**godoc Conventions**:
- First sentence is a summary (shows in package listings)
- Start with the name of the item being documented
- Use proper punctuation and complete sentences
- Add blank line before parameter/return documentation
- Use code blocks with proper indentation for examples

#### 9. Pre-Commit Checklist
**Run these BEFORE every commit (IN ORDER):**

```bash
make fmt                      # Format code
make vet                      # Run go vet
make lint                     # MUST be 100% clean (zero errors)
make test-coverage            # Verify tests pass (target: 95%+)
make build                    # Verify build succeeds
```

**Critical Requirements**:
- **Documentation MUST be thorough** - all exported symbols documented
- **Lint MUST show zero errors** - non-negotiable
- **Test coverage target: 95%+** - strive for comprehensive testing
- **Follow the sequence** - each step validates the previous work

---

## 🏗️ Project Structure

### Standard Go Layout

This project follows the [golang-standards/project-layout](https://github.com/golang-standards/project-layout) conventions:

```
aws-profile-manager/
├── cmd/                        # Application entry points
│   └── aws-profile-manager/
│       ├── main.go            # Main application
│       └── main_test.go
│
├── internal/                   # Private application code (compiler-protected)
│   ├── awscli/                # AWS CLI integration
│   ├── backup/                # Export/import functionality
│   ├── bundled/               # Embedded resources
│   ├── cli/                   # CLI commands
│   ├── core/                  # Core utilities (state, version, timeutil)
│   ├── generators/            # Profile generators (SSO, IAM, AssumeRole)
│   ├── gui/                   # GUI application (MVVM pattern)
│   ├── logging/               # Structured logging system
│   ├── schema/                # Data models
│   ├── settings/              # Settings management
│   ├── sync/                  # Config sync system
│   └── test/          # Test utilities
│
├── configs/                    # Configuration templates
│
├── docs/                       # Documentation
├── scripts/                    # Build scripts
├── bin/                        # Compiled binaries (gitignored)
├── Makefile                    # Build system
└── package.json               # NPM task runner
```

### Key Directories

**`cmd/`** - Application entry points
- `main.go`: Root command setup, provider registration
- One binary per subdirectory pattern

**`internal/`** - Protected code (Go compiler enforces privacy)
- Cannot be imported by external projects
- All business logic lives here

**`configs/`** - Configuration templates
- Example configurations for users

**`bin/`** - Build outputs (NOT in git)
- Compiled binaries
- Platform-specific builds

### Environment Variables

The application uses three environment variables for file locations:

1. **`AWS_PROFILE_MANAGER_CONFIG_DIR`** - Application settings
   - Default: `~/.config/aws-profile-manager/` (Linux/macOS)
   - Stores: `settings.json`, `cache/` subdirectory

2. **`AWS_PROFILE_MANAGER_AWS_DIR`** - AWS CLI directory
   - Default: `~/.aws/`
   - Stores: `config`, `credentials`, `sso/cache/`

3. **`AWS_PROFILE_MANAGER_DESKTOP_DIR`** - Desktop directory
   - Default: `~/Desktop/`
   - Used for: Cheat sheet output

**CRITICAL**: Never use `os.Getenv()` directly in production code. Always use settings helpers:

```go
awsDir := settings.GetAwsDir()              // Correct
desktopDir := settings.GetDesktopDir()      // Correct
cacheDir, _ := settings.GetCacheDir()       // Correct

awsDir := os.Getenv("AWS_PROFILE_MANAGER_AWS_DIR")  // WRONG!
```

**Exceptions** (bootstrap only):
- `internal/logging/debug.go` - AWS_PROFILE_MANAGER_DEBUG flag
- `internal/core/state.go` - CONFIG_DIR for initial settings load

### Development Setup

**.env file** (for local development):
```bash
AWS_PROFILE_MANAGER_DEBUG=1                                     # Enable debug logging
AWS_PROFILE_MANAGER_CONFIG_DIR=./.dev/config   # Dev config directory
```

**Tests**: Automatically isolated via `test.SetupTestEnvironment(t)` - no manual setup needed.

---

## 🔧 Development Workflow

### Quick Start
```bash
# One-time setup
npm run setup                 # Install dependencies

# Daily development (IN ORDER)
make fmt                      # Format code
make vet                      # Run go vet
make lint                     # Run linter (MUST be clean)
make test-coverage            # Run tests with coverage (target: 95%+)
make build                    # Build application

# Run application
./bin/aws-profile-manager --help
./bin/aws-profile-manager gui
```

### Command Priority
1. **NPM scripts** (preferred): `npm run make:test-coverage`
2. **Make targets** (when npm unavailable): `make test-coverage`
3. **Direct Go commands** (custom operations only): `go test ./internal/...`

### Build System

**Key Make targets**:
```bash
make build                 # Build for current platform
make build-all            # Build for all platforms
make test                 # Run tests
make test-coverage        # Tests with coverage report
make fmt                  # Format code
make lint                 # Run linter (MUST be clean)
make vet                  # Run go vet
make clean                # Clean build artifacts
make deps                 # Download dependencies
make deps-system          # Install GUI system dependencies (OpenGL)
```

**NPM scripts** (wraps Make targets):
```bash
npm run make:build
npm run make:test
npm run make:test-coverage
npm run make:fmt
npm run make:lint
```

### Version Information

Version is extracted from `internal/core/version.go`:
```go
const AppVersion = "0.0.1"
```

Build includes: version, git commit, build date (injected via ldflags).

---

## 🧪 Testing Standards

### Coverage Requirements
- **Target**: 95%+ overall coverage
- **Current**: 90%+ for core packages
- **Minimum**: Every `.go` file has a `_test.go` companion
- **Lint**: 100% clean (ZERO errors) - non-negotiable

### Test Organization

**File naming**: `package_test.go` for each `package.go`

**Package coverage** (current):
- generators: 98.1% ✅
- backup: 91.0% ✅
- installer: 90.3% ✅
- awscli: 86.0% ✅
- schema: 87.1% ✅
- schema/test: 100% ✅
- cli: 75.0%
- sync: 53.3%

### Standard Test Pattern

```go
package mypackage

import (
    "testing"
    "aws-profile-manager/internal/test"
    schematest "aws-profile-manager/internal/schema/test"
)

func TestMyFunction(t *testing.T) {
    // Step 1: Setup isolated environment
    test.SetupTestEnvironment(t)
    
    // Step 2: Use test schemas from schema/test package
    schema := schematest.NewManagedSsoSingle()  // Single SSO profile
    // OR: schematest.NewManagedAll(), NewMixedSimple(), NewEmpty(), etc.
    
    // Step 3: Initialize app if needed
    // (skip if testing pure functions)
    
    // Step 4: Test implementation
    result := MyFunction(schema)
    
    // Step 5: Assertions
    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }
}
```

### Test Schema Usage (REQUIRED)

**ALWAYS use the centralized test schemas from `internal/schema/test`:**

The schema test package is organized into 4 files:
- **`managed.go`** - Tool-managed profiles (SSO, IAM, AssumeRole, Generic)
- **`unmanaged.go`** - Personal/user profiles (can contain any profile type)
- **`mixed.go`** - Both managed and unmanaged sections
- **`specialized.go`** - Empty, invalid, large-scale, edge cases

```go
import schematest "aws-profile-manager/internal/schema/test"

// === MANAGED PROFILES (managed.go) ===
// SSO Profiles
schematest.NewManagedSsoSingle()          // 1 org, 1 account, 1 role (most common)
schematest.NewManagedSsoMultiAccount()    // 1 org, 3 accounts
schematest.NewManagedSsoMultiOrg()        // 2 orgs, mixed partitions
schematest.NewManagedSsoComplex()         // 3 orgs, 10+ accounts

// IAM Profiles
schematest.NewManagedIamSingle()          // 1 IAM user
schematest.NewManagedIamMulti()           // 3 IAM users

// AssumeRole Profiles
schematest.NewManagedAssumeRoleSingle()   // 1 role chain
schematest.NewManagedAssumeRoleMulti()    // 3 role chains

// Generic Profiles
schematest.NewManagedGenericSingle()      // 1 generic profile
schematest.NewManagedGenericMulti()       // 3 generic profiles

// Combined Types
schematest.NewManagedAll()                // All profile types
schematest.NewManagedSsoIam()             // SSO + IAM only
schematest.NewManagedEmpty()              // Empty managed section

// === PERSONAL PROFILES (unmanaged.go) ===
// Personal profiles can contain ANY profile type (SSO, IAM, AssumeRole, Generic)
schematest.NewUnmanagedSsoSingle()        // 1 personal SSO org
schematest.NewUnmanagedSsoMulti()         // 2 personal SSO orgs
schematest.NewUnmanagedIamSingle()        // 1 personal IAM user
schematest.NewUnmanagedIamMulti()         // 3 personal IAM users
schematest.NewUnmanagedAssumeRoleSingle() // 1 personal role chain
schematest.NewUnmanagedAssumeRoleMulti()  // 3 personal role chains
schematest.NewUnmanagedGenericSingle()    // 1 personal generic profile
schematest.NewUnmanagedGenericMulti()     // 5 personal generic profiles
schematest.NewUnmanagedAll()              // All personal profile types
schematest.NewUnmanagedMixed()            // SSO + IAM + Generic personal
schematest.NewUnmanagedAboveBelow()       // Profiles Above and Below managed section
schematest.NewUnmanagedEmpty()            // Empty unmanaged section

// === MIXED (BOTH SECTIONS) (mixed.go) ===
schematest.NewMixedSsoSso()               // Work SSO + Personal SSO
schematest.NewMixedIamIam()               // Work IAM + Personal IAM
schematest.NewMixedAllAll()               // All managed + All personal
schematest.NewMixedSsoIam()               // Managed SSO + Personal IAM
schematest.NewMixedAllSso()               // All managed + Personal SSO
schematest.NewMixedSsoGeneric()           // Managed SSO + Personal generic
schematest.NewMixedSimple()               // Basic work + personal (common)
schematest.NewMixedComplex()              // Comprehensive stress test

// === SPECIALIZED (specialized.go) ===
// Empty Schemas
schematest.NewEmpty()                     // Completely empty
schematest.NewManagedOnlyEmpty()          // Only empty managed section
schematest.NewUnmanagedOnlyEmpty()        // Only empty unmanaged section

// Invalid Schemas (error testing)
schematest.NewInvalid()                   // Invalid data (all types)
schematest.NewInvalidMissingRequired()    // Missing required fields

// Partial/Missing Data Schemas (error testing)
schematest.NewPartialSsoMissingUrl()      // SSO org missing URL
schematest.NewPartialSsoMissingRegion()   // SSO org missing default region
schematest.NewPartialSsoEmptyAccounts()   // SSO org with no accounts
schematest.NewPartialSsoEmptyRoles()      // SSO org with no roles
schematest.NewPartialIamMissingCreds()    // IAM user missing credentials
schematest.NewPartialIamMissingRegion()   // IAM user missing region
schematest.NewPartialAssumeRoleMissingArn()        // AssumeRole missing ARN
schematest.NewPartialAssumeRoleMissingSource()     // AssumeRole missing source
schematest.NewPartialGenericEmptyProperties()      // Generic with empty properties
schematest.NewPartialGenericEmptyName()            // Generic with empty name

// Large Scale Schemas (performance testing)
schematest.NewLargeScale()                // 2100+ managed profiles (10 orgs, 100 accounts)
schematest.NewLargeScaleUnmanaged()       // 100 personal profiles

// Edge Cases
schematest.NewSingleProfileAllTypes()     // 1 of each type
schematest.NewMinimal()                   // Smallest valid schema
```

**Why**:
- ✅ 48 pre-built schemas covering all scenarios
- ✅ Consistency across all tests
- ✅ Easy to maintain (update once, fixes everywhere)
- ✅ Well-documented and validated (100% test coverage)
- ✅ Covers simple to complex scenarios

**Rules**:
- ❌ NEVER hardcode schemas in individual tests
- ❌ NEVER create inline test schemas
- ✅ ALWAYS use `schematest.*` functions
- ✅ Add new test schemas to appropriate file (managed/unmanaged/mixed/specialized) if needed

**Naming Convention**:
- Format: `New{Section}{ProfileType}{Complexity}()`
- Section: `Managed` | `Unmanaged` | `Mixed`
- ProfileType: `Sso` | `Iam` | `AssumeRole` | `Generic` | `All` | (empty for specialized)
- Complexity: `Single` | `Multi` | `Complex` | `Simple` | `Empty` | `Invalid`

### Table-Driven Tests

Use for multiple test cases:

```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"empty input", "", ""},
        {"valid input", "test", "TEST"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Function(tt.input)
            if got != tt.expected {
                t.Errorf("got %v, want %v", got, tt.expected)
            }
        })
    }
}
```

### Testing with App State

```go
func TestWithState(t *testing.T) {
    test.SetupTestEnvironment(t)
    
    // Initialize app with default providers
    if err := core.App.Initialize(nil); err != nil {
        t.Fatalf("Failed to initialize: %v", err)
    }
    
    // Access settings, state, etc.
    settings := settings.Get()
    // ... test with settings
}
```

### GUI Testing

Use Fyne's test framework:

```go
import fyneTest "fyne.io/fyne/v2/test"

func TestGUIComponent(t *testing.T) {
    testApp := fyneTest.NewApp()  // Headless test app
    defer testApp.Quit()
    
    view := NewMyView()
    
    // Test widget interactions
    button := view.GetButton()
    test.Tap(button)
    
    // Assertions...
}
```

### Running Tests

```bash
# All tests
make test

# With coverage
npm run make:test-coverage

# Specific package
go test ./internal/awscli -v

# Single test
go test ./internal/awscli -run TestExtractor -v
```

---

## 🏛️ Architecture Patterns

### CLI/GUI Separation

**Principle**: Presentation layer vs Business logic layer

```
┌─────────────────────────────────────┐
│   CLI Commands (internal/cli/)      │  Parse flags → Call API → Display
│   GUI Views (internal/gui/views/)   │  Parse input → Call API → Render
└─────────────────────────────────────┘
              ↓ Call API
┌─────────────────────────────────────┐
│   Package APIs                    │  Orchestrate → Process → Return
│   - awscli.ListProfiles()          │
│   - profiles.NewInstaller(cfg)     │
│   - sync.Sync(ctx, cfg, ...)       │
└─────────────────────────────────────┘
```

**Benefits**:
- CLI and GUI call identical functions
- Zero code duplication
- Business logic tested independently
- Easy to add new interfaces

### Config Priority Chain

The application follows a strict priority order for configuration:

```
1. Explicit --config flag    (Highest priority - user override)
   ↓ (if not provided)
2. Sync (remote fetch)        (Remote configuration if enabled)
   ↓ (if sync fails/disabled)
3. Cache (local copy)         (Offline fallback)
```

**Implementation**: CLI loader logic plus `internal/profiles/` components

### Settings System

**Architecture**: Unified settings in single package with schemas for validation/UI generation.

**Sections**:
- **Application**: Markers, metadata
- **Logging**: Log level, debug mode
- **GUI**: Theme, window size, dialog dimensions
- **Sync**: Strategy, URLs, credentials
- **AwsCLI**: Auto-refresh, cache intervals

**Access pattern**:
```go
// Get full settings object (for modification)
currentSettings := settings.Get()
currentSettings.GUI.Theme = "dark"
settings.Set(currentSettings)

// Get individual section (read-only)
guiSettings := settings.GetGUI()
theme := guiSettings.Theme

// Get environment-based paths
awsDir := settings.GetAwsDir()
cacheDir, _ := settings.GetCacheDir()
```

**Schema-driven UI**: Settings dialog auto-generates from field schemas - add fields without changing UI code.

### Naming Conventions

**CRITICAL** - These patterns are used throughout the system:

**SSO Sessions**: `<organization-alias>-<partition>`
- Example: `test-org-commercial`, `org2-govcloud`

**Profiles**: `<partition>-<account-alias>-<role>[--<region>]`
- Default region: `commercial-dev-Developer`
- Non-default region: `commercial-dev-Developer--us-west-2`
- GovCloud: `govcloud-prod-SystemAdmin--us-gov-east-1`

**Why**: Hierarchical naming enables filtering, recognition, and SSO session sharing.

**Implementation**: `internal/schema/naming.go`

### GUI Architecture (MVVM Pattern)

**Pattern**: Model-View-ViewModel separation

```
internal/gui/
├── gui.go                    # Main app, window setup
├── components/               # Reusable UI components
│   ├── header.go            # Header with logo
│   ├── footer.go            # Status bar
│   └── dialog.go            # Standard dialogs
├── viewmodels/              # View-specific state
│   ├── settings_viewmodel.go
│   ├── installer_viewmodel.go
│   └── ...
└── views/                   # UI components (Fyne widgets)
    ├── settings_view.go
    ├── installer_view.go
    └── ...
```

**Responsibilities**:
- **Models**: Existing packages (awscli, installer, sync, etc.)
- **ViewModels**: View state, UI logic, thread-safe operations
- **Views**: Pure Fyne widgets, zero business logic

**GUI Standards**:

1. **Text Styling** - Use Markdown for ALL formatted text:
```go
// ✅ CORRECT
title := widget.NewRichTextFromMarkdown("# AWS Profile Manager")
section := widget.NewRichTextFromMarkdown("## Configuration")
plain := widget.NewLabel("Status: Ready")

// ❌ WRONG
text := canvas.NewText("Title", theme.ForegroundColor())
text.TextSize = 24
```

2. **Dialogs** - ALWAYS use `components.ShowCustomDialog()`:
```go
// ✅ CORRECT
dialog := components.ShowCustomDialog(components.DialogOptions{
    Title:       "Settings",
    Content:     form,
    Buttons:     buttons,
    Window:      window,
    Scrollable:  true,
    UseSettings: true,
})

// ❌ WRONG
dialog := dialog.NewCustom("Settings", "OK", content, window)
```

3. **Thread Safety** - Use `fyne.Do()` for GUI updates from goroutines:
```go
go func() {
    result := doWork()
    fyne.Do(func() {
        label.SetText(result)  // Safe GUI update
    })
}()
```

---

## 📦 Package Documentation

All documentation should use PascalCase for filenames.

### Core Packages Overview

#### internal/awscli
**Purpose**: AWS CLI introspection and profile management

**Features**:
- Extract profiles from `~/.aws/config`
- Cache profiles for performance
- Filter profiles (regex, account ID, region, etc.)
- SSO session status tracking

**API**:
```go
result, err := awscli.ListProfiles(criteria)
status, err := awscli.GetSessionStatus(options)
```

**Files**:
- `api.go` - High-level functions
- `extractor.go` - Config file parsing
- `extractor_sso.go` - SSO profile extraction
- `extractor_iam.go` - IAM profile extraction
- `cache.go` - Profile caching
- `filters.go` - Advanced filtering
- `sessions.go` - SSO session management
- `types.go` - Data structures

**Coverage**: 86.0%

#### internal/backup
**Purpose**: Export/import AWS profiles and application settings

**Features**:
- Export: AWS config → Schema JSON (managed & unmanaged sections)
- Import: Schema JSON → AWS config (with duplicate detection)
- Settings backup: Export/import application settings
- Personal profile preservation: Maintains user's custom profiles
- Managed/unmanaged section handling

**API**:
```go
result, err := backup.ExportProfiles(options)
result, err := backup.ImportProfiles(options)
```

**Use Cases**:
- Disaster recovery (backup before OS reinstall)
- Configuration migration (move profiles between machines)
- Settings backup (preserve app configuration)
- Personal profile backup (save custom profiles not in managed section)
- Creating installer configs (export as template for team distribution)

**Files**:
- `api.go` - High-level functions
- `exporter.go` - Export logic
- `importer.go` - Import logic
- `merger.go` - Duplicate detection
- `schema.go` - Export schema definition

**Coverage**: 91.0%

#### internal/profiles
**Purpose**: Unified profile installation, export, import, removal, and cheat sheet generation

**Features**:
- Write profiles to `~/.aws/config`
- Managed section with markers
- Profile filtering
- Cheat sheet generation
- Profile removal
- JSON export/import with duplicate handling

**API**:
```go
installer := profiles.NewInstaller(config)
result, err := installer.Install(ctx, opts, reporter)
remover := profiles.NewRemover(config)
result, err := remover.Remove(ctx, opts, reporter)
exporter := profiles.NewExporter(config)
result, err := exporter.Export(ctx, opts, reporter)
```

#### internal/sync
**Purpose**: Remote configuration fetching

**Strategies**:
- HTTP/HTTPS: Public URLs, CDNs
- S3: AWS buckets (SSO/IAM/public-read)
- Local: File system (for testing)
- Git: Available (experimental)

**API**:
```go
// Convert settings → config (CLI/GUI only)
cfg := sync.ConfigFromSettings(&settings.Get().Sync)

// Fetch (business logic entry point)
result, err := sync.Sync(ctx, cfg, sync.Options{}, reporter)
// result.Data is *schema.Schema
// result.CacheHit is true if served from local cache
```

**Files**:
- `sync.go` - Main `Sync()` entry point
- `fetcher.go` - Fetcher interface
- `http_fetcher.go` - HTTP/HTTPS fetching
- `s3_fetcher.go` - S3 fetching
- `git_fetcher.go` - Git fetching
- `local_fetcher.go` - Local file system
- `cache.go` - TTL-aware caching
- `validation.go` - Config and SSRF validation
- `types.go` - Data structures

**Coverage**: 53.3%

#### internal/schema
**Purpose**: Data model for AWS configurations

**Types**:
- **Schema**: Top-level with managed/unmanaged sections
- **ProfileCollection**: Organizations, accounts, roles
- **Organization**: AWS org with partitions
- **Partition**: Commercial/GovCloud with accounts
- **Account**: Individual AWS account

**Features**:
- Validation
- Filtering
- JSON serialization
- Naming conventions

**Files**:
- `types.go` - Data structures
- `naming.go` - Profile/session naming
- `validation.go` - Schema validation
- `filtering.go` - Filter operations
- `json.go` - JSON handling

**Coverage**: 87.1%

#### internal/schema/test
**Purpose**: Centralized test schema fixtures for all tests

**Organization**: 4 files by section type
- **`managed.go`** - Tool-managed profiles (14 schemas)
- **`unmanaged.go`** - Personal/user profiles (13 schemas)
- **`mixed.go`** - Both sections (8 schemas)
- **`specialized.go`** - Special cases (21 schemas)

**Total**: 56 test schemas covering all scenarios

**Schema Categories**:

**Managed Profiles (managed.go)**:
```go
// SSO Profiles
schematest.NewManagedSsoSingle()          // 1 org, 1 account, 1 role (most common)
schematest.NewManagedSsoMultiAccount()    // 1 org, 3 accounts
schematest.NewManagedSsoMultiOrg()        // 2 orgs, mixed partitions
schematest.NewManagedSsoComplex()         // 3 orgs, 10+ accounts

// IAM, AssumeRole, Generic
schematest.NewManagedIamSingle()          // 1 IAM user
schematest.NewManagedIamMulti()           // 3 IAM users
schematest.NewManagedAssumeRoleSingle()   // 1 role chain
schematest.NewManagedAssumeRoleMulti()    // 3 role chains
schematest.NewManagedGenericSingle()      // 1 generic profile
schematest.NewManagedGenericMulti()       // 3 generic profiles

// Combined
schematest.NewManagedAll()                // All profile types
schematest.NewManagedSsoIam()             // SSO + IAM only
schematest.NewManagedEmpty()              // Empty managed section
```

**Unmanaged Profiles (unmanaged.go)**:
```go
// Personal profiles (any type: SSO, IAM, AssumeRole, Generic)
schematest.NewUnmanagedSsoSingle()        // 1 personal SSO org
schematest.NewUnmanagedSsoMulti()         // 2 personal SSO orgs
schematest.NewUnmanagedIamSingle()        // 1 personal IAM user
schematest.NewUnmanagedIamMulti()         // 3 personal IAM users
schematest.NewUnmanagedAssumeRoleSingle() // 1 personal role chain
schematest.NewUnmanagedAssumeRoleMulti()  // 3 personal role chains
schematest.NewUnmanagedGenericSingle()    // 1 personal generic
schematest.NewUnmanagedGenericMulti()     // 5 personal generic
schematest.NewUnmanagedAll()              // All personal types
schematest.NewUnmanagedMixed()            // SSO + IAM + Generic
schematest.NewUnmanagedAboveBelow()       // Profiles in both Above and Below
schematest.NewUnmanagedEmpty()            // Empty unmanaged section
```

**Mixed Schemas (mixed.go)**:
```go
// Both managed and unmanaged sections
schematest.NewMixedSsoSso()               // Work SSO + Personal SSO
schematest.NewMixedIamIam()               // Work IAM + Personal IAM
schematest.NewMixedAllAll()               // All managed + All personal
schematest.NewMixedSsoIam()               // Managed SSO + Personal IAM
schematest.NewMixedAllSso()               // All managed + Personal SSO
schematest.NewMixedSsoGeneric()           // Managed SSO + Personal generic
schematest.NewMixedSimple()               // Basic work + personal (common)
schematest.NewMixedComplex()              // Comprehensive stress test
```

**Specialized Schemas (specialized.go)**:
```go
// Empty Schemas
schematest.NewEmpty()                     // Completely empty
schematest.NewManagedOnlyEmpty()          // Only empty managed
schematest.NewUnmanagedOnlyEmpty()        // Only empty unmanaged

// Invalid Schemas (error testing)
schematest.NewInvalid()                   // Invalid data (all types)
schematest.NewInvalidMissingRequired()    // Missing required fields

// Partial/Missing Data Schemas (error testing)
schematest.NewPartialSsoMissingUrl()      // SSO org missing URL
schematest.NewPartialSsoMissingRegion()   // SSO org missing default region
schematest.NewPartialSsoEmptyAccounts()   // SSO org with no accounts
schematest.NewPartialSsoEmptyRoles()      // SSO org with no roles
schematest.NewPartialIamMissingCreds()    // IAM user missing credentials
schematest.NewPartialIamMissingRegion()   // IAM user missing region
schematest.NewPartialAssumeRoleMissingArn()        // AssumeRole missing ARN
schematest.NewPartialAssumeRoleMissingSource()     // AssumeRole missing source
schematest.NewPartialGenericEmptyProperties()      // Generic empty properties
schematest.NewPartialGenericEmptyName()            // Generic empty name

// Large Scale Schemas (performance testing)
schematest.NewLargeScale()                // 2100+ managed profiles
schematest.NewLargeScaleUnmanaged()       // 100 personal profiles

// Edge Cases
schematest.NewSingleProfileAllTypes()     // 1 of each type
schematest.NewMinimal()                   // Smallest valid schema
```

**Naming Convention**:
- Format: `New{Section}{ProfileType}{Complexity}()`
- Section: `Managed` | `Unmanaged` | `Mixed` | (empty for specialized)
- ProfileType: `Sso` | `Iam` | `AssumeRole` | `Generic` | `All` | `Partial` | `Invalid`
- Complexity: `Single` | `Multi` | `Complex` | `Simple` | `Empty`

**Usage**:
```go
import schematest "aws-profile-manager/internal/schema/test"

func TestInstaller(t *testing.T) {
    schema := schematest.NewManagedSsoSingle()  // Clear what it contains
    // ... test with schema
}

func TestErrorHandling(t *testing.T) {
    schema := schematest.NewPartialSsoMissingUrl()  // Test validation
    // ... test error handling
}
```

**Rules**:
- ✅ ALWAYS use these schemas in tests
- ❌ NEVER hardcode or create inline schemas
- ✅ Add new schemas to appropriate file if needed
- ✅ Use Partial* schemas for error testing
- ✅ Use Invalid* schemas for validation testing

**Files**:
- `managed.go` / `managed_test.go` (14 schemas)
- `unmanaged.go` / `unmanaged_test.go` (13 schemas)
- `mixed.go` / `mixed_test.go` (8 schemas)
- `specialized.go` / `specialized_test.go` (21 schemas)

**Coverage**: 100%

#### internal/generators
**Purpose**: Generate AWS config file content

**Generators**:
- **SSO**: SSO profiles and sessions
- **IAM**: IAM user profiles (credential_process)
- **AssumeRole**: Role assumption profiles
- **Generic**: Catch-all profiles

**API**:
```go
content, stats := generators.GenerateSsoProfiles(profiles)
content, stats := generators.GenerateIamProfiles(profiles)
content, stats := generators.GenerateAssumeRoleProfiles(profiles)
content, stats := generators.GenerateGenericProfiles(profiles)
```

**Features**:
- Pure content generation (no file I/O)
- Section-agnostic (works with any ProfileCollection)
- Shared by installer and backup
- Returns content + statistics

**Coverage**: 98.1%

#### internal/logging
**Purpose**: Structured logging system

**Features**:
- Color-coded output
- Key-value and formatted logging
- Debug mode (AWS_PROFILE_MANAGER_DEBUG flag)
- Hierarchical debug output

**Usage**:
```go
logging.Log.Info("message", "key", value)
logging.Log.Infof("message %s", value)
logging.Debug.Log("debug info", "detail", data)
```

**Coverage**: 100%

#### internal/settings
**Purpose**: Unified application settings

**Architecture**: Schema-driven with automatic UI generation

**Sections**:
- Application, Logging, GUI, Sync, AwsCLI

**Features**:
- Type-safe access
- Schema-driven UI
- Dependency management
- Field validation

**Coverage**: Well-tested

#### internal/gui
**Purpose**: Graphical user interface

**Architecture**: MVVM with Fyne framework

**Status**: Phase 1 complete (layout, theme, settings)

**Planned**: Installer, sync, profiles, sessions views

**Coverage**: Good test coverage for completed components

#### internal/cli
**Purpose**: Command-line interface

**Commands**:
- install, export, import, profiles, sessions, sync, gui, version

**Pattern**: Parse flags → Call package API → Display results

**Coverage**: 75.0%

#### internal/core
**Purpose**: Core utilities

**Contains**:
- `state.go` - Application state management
- `version.go` - Version information
- `timeutil.go` - Time utilities

**Coverage**: High

#### internal/test
**Purpose**: Test isolation utilities

**Key function**: `SetupTestEnvironment(t)` - Creates isolated test environment

**Helpers**: Path utilities for test directories

**Coverage**: 100%

#### internal/bundled
**Purpose**: Embedded resources

**Contains**:
- Logo assets (generated at build time)

**Build**: Generated by `make embed-all`

---

## 🔍 Decision Trees for AI Assistants

### When Adding/Modifying Package Functions
1. ✅ Do I create high-level API function? → YES
2. ✅ Do I accept `task.Reporter` parameter? → YES (if long-running or subprocess)
3. ✅ Do I use `task.SubprocessTask` for external commands? → YES
4. ✅ Do I call `reporter.ReportStatus()` for progress? → YES
5. ✅ Do I accept config structs (NOT settings)? → YES (dependency injection)
6. ✅ Do I create objects internally? → YES
7. ✅ Do I apply business logic? → YES
8. ✅ Do I return result struct? → YES
9. ❌ Do I import settings package? → NO (except for helper functions)
10. ❌ Do I parse CLI flags? → NO
11. ❌ Do I format display output? → NO

### When Adding/Modifying CLI Commands
1. ❌ Do I create Extractor/Filter in CLI? → NO
2. ✅ Do I parse CLI flags? → YES
3. ✅ Do I build config from settings? → YES (convert settings → config struct)
4. ✅ Do I pass `task.CliReporter{}`? → YES (to package functions)
5. ✅ Do I call ONE package function? → YES
6. ✅ Do I display results only? → YES
7. ✅ Do I use `logging.Log.ErrorfWithDetails()` for errors? → YES

### When Adding/Modifying GUI Views
1. ❌ Do I put business logic in GUI? → NO
2. ✅ Do I build config from settings? → YES (convert settings → config struct)
3. ✅ Do I run work in goroutine? → YES (keep UI responsive)
4. ✅ Do I pass `task.NoOpReporter{}` or `task.ChannelReporter`? → YES
5. ✅ Do I use `fyne.Do()` for UI updates? → YES (from goroutine)
6. ✅ Do I use `dialog.ShowError()` for errors? → YES
7. ✅ Do I hide progress dialog before showing error? → YES

### When Writing Tests
1. ✅ Do I call `test.SetupTestEnvironment(t)`? → YES (for file I/O tests)
2. ✅ Do I initialize app if needed? → YES (`core.App.Initialize(nil)`)
3. ✅ Do I use valid settings values? → YES (settings validate on `Set()`)
4. ✅ Do I test error paths? → YES (required for coverage)
5. ✅ Do I create `_test.go` file? → YES (for every `.go` file)
6. ✅ Do I handle ALL errors explicitly? → YES (use `_` or check)
7. ✅ Do I use settings for markers? → YES (never hardcode)

### When Writing Go Code
1. ❌ Do I use `fmt.Println()`? → NO (use `logging.Log.Info()`)
2. ❌ Do I use `os.Getenv()` for app paths? → NO (use `settings.GetAwsDir()`)
3. ❌ Do I use ALL_CAPS names? → NO (use PascalCase: HttpClient, CliReporter)
4. ✅ Do I use `logging` package? → YES (always)
5. ✅ Do I use PascalCase for multi-word names? → YES (HttpClient, not HTTPClient)
6. ✅ Do I document all exported symbols? → YES (godoc format)
7. ✅ Do I include usage examples? → YES (for complex functions)

### When Working with GUI
1. ✅ Do I use Markdown for text? → YES (`widget.NewRichTextFromMarkdown()`)
2. ✅ Do I use `ShowCustomDialog()`? → YES (never manual dialogs)
3. ✅ Do I use `fyne.Do()` for background updates? → YES
4. ❌ Do I use manual text styling? → NO (breaks theming)

### Before Committing
1. ✅ Run `make fmt`? → YES
2. ✅ Run `make vet`? → YES
3. ✅ Run `make lint` (zero errors)? → YES
4. ✅ Run `make test-coverage` (target: 95%+)? → YES
5. ✅ Run `make build`? → YES
6. ✅ Check documentation complete? → YES (all exported symbols)

---

## 📝 Important Reminders

1. **Standard Go Layout**: `cmd/` for entry points, `internal/` for protected code
2. **Logging**: ALWAYS use logging package, NEVER fmt.Print* or log.Print*
3. **Naming Conventions**: ALWAYS use PascalCase (HttpClient, CliReporter, E2eTesting.md), NEVER ALL_CAPS or snake_case
4. **Test Isolation**: ALWAYS use `test.SetupTestEnvironment(t)` for file I/O tests
5. **CLI/GUI Pattern**: Thin presentation layer, call ONE package API function
6. **Task Package**: Business logic accepts `task.Reporter`, CLI uses `task.CliReporter{}`, GUI uses `task.NoOpReporter{}`
7. **Dependency Injection**: Business logic accepts config structs, NEVER imports settings package
8. **Markers**: ALWAYS use settings helpers, NEVER hardcode
9. **Documentation**: ALL exported symbols MUST be documented (godoc format)
10. **Lint**: MUST be 100% clean (zero errors) before commit
11. **Coverage**: Target 95%+, every .go file has _test.go
12. **Error Handling**: Explicitly handle ALL errors in tests
13. **GUI**: Markdown for text, ShowCustomDialog for dialogs, fyne.Do for updates
14. **Paths**: Use settings helpers, NEVER os.Getenv() directly
15. **Settings Validation**: Settings validate on `Set()` - use proper values in tests
16. **Test Schemas**: ALWAYS use `schematest.*` functions, NEVER hardcode schemas
17. **Workflow**: fmt → vet → lint → test-coverage → build (IN ORDER)

---

**Last Updated**: May 15, 2026  
**Go Version**: 1.22+  
**Test Coverage**: 90%+ (target: 95%+)  
**Architecture**: Standard Go project layout with internal/ protection  
**Status**: Production-ready with active GUI development
