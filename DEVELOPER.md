# Developer Guide

This guide is for developers who want to contribute to or modify the AWS Profile Manager project.

## Quick Start

### Prerequisites
- Go 1.25.10+ (for building and running)
- Node.js/npm (for convenience scripts)
- System dependencies for GUI (automatically installed via make)

### Setup
```bash
# Clone the repository
git clone <repository-url>
cd aws-profile-manager

# First-time setup (installs dependencies and GUI system libraries)
npm run setup
```

### Environment Configuration
The project uses a `.env` file for environment variables. You may need to customize it for your system:

```bash
# .env file - customize these values if needed:
AWS_PROFILE_MANAGER_DEBUG=1                                     # Enable debug logging (1=on, 0=off)
AWS_PROFILE_MANAGER_CONFIG_DIR=./.dev/config   # Development config directory
AWS_PROFILE_MANAGER_THROTTLE=100ms             # Optional: Throttle for testing cancellation (default: off)
```

**Note**: Tests automatically use temporary directories and ignore these settings for isolation.

### Development Data Setup

To test with realistic AWS CLI configurations and SSO sessions, sync your existing AWS CLI data to the development directories:

#### Quick Sync (Recommended)
```bash
# Use the convenient sync script
./scripts/sync-dev-data.sh

# Or via npm script
npm run sync-dev-data
```

#### Manual Setup
```bash
# Create development directories
mkdir -p .dev/aws/sso/cache

# Copy your AWS CLI configuration (profiles and SSO sessions)
cp ~/.aws/config .dev/aws/config
cp ~/.aws/credentials .dev/aws/credentials  # Optional: only if you have non-SSO profiles

# Copy SSO cache files for active sessions (enables session status testing)
cp ~/.aws/sso/cache/* .dev/aws/sso/cache/ 2>/dev/null || echo "No SSO cache files found"

# Verify the setup
ls -la .dev/aws/
ls -la .dev/aws/sso/cache/
```

**Benefits of Using Real Data:**
- ✅ **Test SSO session status** - See active/expired sessions in GUI
- ✅ **Realistic profile filtering** - Test with your actual AWS accounts and roles  
- ✅ **Session management testing** - Verify SSO login/logout functionality
- ✅ **Profile extraction testing** - Test AWS CLI config parsing with real data

**Development Isolation:**
- Development uses `.dev/` directories (never touches `~/.aws/`)
- Tests use `t.TempDir()` (completely isolated)
- Production uses standard `~/.aws/` directories

**Updating Development Data:**
```bash
# Quick sync (recommended)
./scripts/sync-dev-data.sh

# Or manual sync when your AWS CLI config changes
cp ~/.aws/config .dev/aws/config

# Refresh SSO cache after new logins
cp ~/.aws/sso/cache/* .dev/aws/sso/cache/
```

**Note**: The `.dev/` directories are git-ignored and won't be committed to the repository.


## Development Tools

This project leverages modern development tools to accelerate development while maintaining high quality standards:

### AI-Assisted Development

**Tools Used:**
- **GitHub Copilot** - Code completion and generation
- **Claude/ChatGPT** - Architecture discussions, documentation, and testing strategies
- **AI Code Review** - Additional validation layer

**Quality Assurance:**
- ✅ All AI-generated code is human-reviewed
- ✅ Comprehensive test coverage (95%+) validates functionality
- ✅ Vulnerability scanning (`govulncheck`) catches known CVEs in dependencies
- ✅ Linting enforces code standards (100% clean)
- ✅ Manual testing for GUI and CLI workflows

**How AI Accelerates Development:**
- Faster implementation of boilerplate and common patterns
- Comprehensive test case generation
- Documentation consistency
- Security best practices enforcement

**Development Philosophy:**
AI is a tool, like an IDE or linter. The code is still written, reviewed, tested, and validated by humans. Quality and security are never compromised for speed.


## Development Workflow

**All development happens on the host system for optimal GUI support and performance.**

### Daily Development
```bash
# Recommended command hierarchy (npm scripts first, then make, then go direct):

# 1. PRIMARY: Use npm scripts for common tasks
npm run make:test-coverage  # Run tests with coverage (most used)
npm run make:fmt           # Format code  
npm run make:lint          # Lint code
npm run make:vuln          # Run vulnerability scan
npm run make:build         # Build the application

# 2. SECONDARY: Use make for development-specific tasks
make deps-system          # Install GUI system dependencies
make help                 # See all available commands

# 3. DIRECT: Use go commands for specific development tasks
go run ./cmd/aws-profile-manager/main.go gui    # Run GUI directly for testing
go run ./cmd/aws-profile-manager/main.go --help # CLI help
./bin/aws-profile-manager # Use built binary
```

**💡 Tip**: For realistic testing with SSO sessions and profiles, sync your AWS CLI data to `.dev/aws/` (see Development Data Setup above).

### Available npm Scripts
- **`npm run setup`** - One-time setup (install dependencies + GUI libraries)
- **`npm run sync-dev-data`** - Sync AWS CLI config/cache to development directories
- **`npm run reset-dev`** - Reset development data (clears AWS config, cache, cheat sheet)
- **`npm run make:build`** - Build application binary
- **`npm run make:build:all`** - Build for all supported platforms
- **`npm run make:test`** - Run tests
- **`npm run make:test-verbose`** - Run tests with verbose output
- **`npm run make:test-coverage`** - Run tests with coverage report (most used)
- **`npm run make:fmt`** - Format Go code
- **`npm run make:lint`** - Run code linter
- **`npm run make:vet`** - Run go vet static analysis
- **`npm run make:vuln`** - Run vulnerability scanner (govulncheck)
- **`npm run make:clean`** - Clean build artifacts
- **`npm run make:help`** - Show all make targets
- **`npm run run:go:gui`** - Run GUI directly via `go run` (no build step)
- **`npm run run:go:gui:debug`** - Run GUI via `go run` with debug logging
- **`npm run run:local:gui`** - Run the built local binary GUI
- **`npm run run:local:gui:debug`** - Run the built local binary GUI with debug logging

## Architecture Overview

### Project Structure
```
├── cmd/
│   └── aws-profile-manager/
│       ├── main.go              # Application entry point
│       └── main_test.go
├── internal/
│   ├── awscli/          # AWS CLI integration (extraction, caching, filtering, sessions)
│   ├── backup/          # Export/import functionality
│   ├── cli/             # Command-line interface (Cobra commands)
│   ├── core/            # Settings, state management, version info
│   ├── profiles/        # AWS config processing and profile generation
│   ├── generators/      # Profile generators (SSO, IAM, AssumeRole)
│   ├── schema/          # Data models
│   ├── sync/            # Config sync system
│   ├── logging/         # Logger and debug systems
│   └── bundled/         # Auto-generated embedded resources
├── bin/                    # Built binaries
├── coverage/              # Test coverage reports  
├── Makefile              # Build system (source of truth for commands)
└── package.json          # npm convenience scripts
```

### Key Components

#### CLI Commands
- **`install`** - Generate AWS CLI profiles from configuration files (primary feature)
- **`sessions`** - List AWS SSO sessions and their status (active/expired)
- **`profiles`** - List and filter existing AWS CLI profiles from ~/.aws/config (supports SSO, IAM, AssumeRole)
- **`version`** - Display application version and build information
- **`gui`** - Launch GUI interface (when available)

#### Core Libraries
- **awscli package** - AWS CLI config parsing with support for SSO, IAM, and AssumeRole profiles
- **profiles package** - Processes AWS configurations and generates profiles
- **generators package** - Profile content generation (SSO, IAM, AssumeRole, Generic)
- **sync package** - Remote configuration fetching (HTTP, S3, Git, Local)
- **backup package** - Export/import functionality
- **logging package** - Structured logging with debug support
- **core package** - Settings and application state management

## Testing

**📖 Complete Testing Guide**: See [`docs/TestingGuide.md`](docs/TestingGuide.md) for comprehensive testing standards, patterns, and best practices.

### Test Isolation (MANDATORY)

All tests MUST use the `test.SetupTestEnvironment()` helper to ensure complete isolation:

```go
import "aws-profile-manager/internal/test"

func TestMyFunction(t *testing.T) {
    _ = test.SetupTestEnvironment(t)  // REQUIRED for file I/O
    
    if err := core.App.Initialize(); err != nil {
        t.Fatalf("Failed to initialize: %v", err)
    }
    
    // All file operations use temp directories automatically
}
```

**Why**: This helper sets all three environment variables to temporary directories, ensuring tests never touch real user files or `.dev/` directories. Cache files are automatically stored in `$CONFIG_DIR/cache`.

### Environment Variables for Testing

| Variable | Purpose | Test Value |
|----------|---------|------------|
| `AWS_PROFILE_MANAGER_CONFIG_DIR` | App settings directory | `/tmp/TestName.../config` |
| `AWS_PROFILE_MANAGER_AWS_DIR` | AWS CLI directory | `/tmp/TestName.../.aws` |
| `AWS_PROFILE_MANAGER_DESKTOP_DIR` | Desktop cheat sheets | `/tmp/TestName.../Desktop` |
| `AWS_PROFILE_MANAGER_THROTTLE` | Optional throttle for testing cancellation | `100ms`, `1s`, etc. (default: off) |

**Cache Directory**: Automatically created in `$CONFIG_DIR/cache` (e.g., `/tmp/TestName.../config/cache`)

### Testing with Throttle

The `AWS_PROFILE_MANAGER_THROTTLE` environment variable allows you to slow down profile generation for manual testing of cancellation and progress reporting:

```bash
# Run with 100ms throttle (fast testing)
AWS_PROFILE_MANAGER_THROTTLE=100ms go test ./internal/generators -v

# Run with 1s throttle (easier to observe cancellation)
AWS_PROFILE_MANAGER_THROTTLE=1s ./bin/aws-profile-manager install

# Run without throttle (full speed - default for tests)
go test ./internal/generators -v
```

**Use Cases**:
- Manual testing of context cancellation
- Verifying progress reporting works correctly
- Testing GUI responsiveness during long operations
- Simulating slower network conditions

**Important**: This is ONLY for manual testing. All automated tests run at full speed by default (no throttle).

### Test Coverage
- **Target**: 95%+ overall coverage
- **Current**: 87.7%
  - CLI: 92.1% ✅
  - Core: 96.7% ✅  
  - Logging: 100% ✅
  - Installer: 93.0% ✅
  - AWS CLI: 76.4%
  - Sync: 65.1%

### Running Tests
```bash
# Run tests with coverage (most common)
npm run make:test-coverage

# Quick test run (no coverage)
npm run make:test

# Test specific package
go test ./internal/core/... -v
go test ./internal/gui/... -v
```

### Test Requirements
- Every `.go` file MUST have corresponding `_test.go` file
- Tests MUST use `test.SetupTestEnvironment(t)` for file I/O
- Both success and error paths must be tested
- Use table-driven tests for multiple scenarios
- 100% clean lint (zero errors) before commit

## Code Standards

### Logging Requirements
**CRITICAL: Always use the logging package - NEVER use fmt.Print\* or log.Print\***

**Two Patterns:**
1. **Key-Value Logging** (non-f functions) - Inline metadata as key=value
2. **Formatted String Logging** (f functions) - String interpolation with placeholders

```go
// ✅ CORRECT: Key-value logging (non-f functions)
logging.Log.Info("Operation completed", "duration", "2s", "files", 10)
logging.Log.Success("Profile created", "name", "prod", "region", "us-east-1")
logging.Log.Warn("Cache miss", "key", cacheKey, "fallback", "network")
logging.Log.Error("Failed to process", "file", filename, "reason", "not found")

// ✅ CORRECT: Formatted string logging (f functions)
logging.Log.Infof("Processing %d files", count)
logging.Log.Successf("Saved to %s", filepath)
logging.Log.Warnf("Retrying in %d seconds", delay)
logging.Log.Errorf("Connection failed: %w", err)

// ✅ CORRECT: Debug logging (hierarchical key-value display)
logging.Debug.Log("Operation details", "status", "active", "count", 5)
// Output:
// 🐞 Operation details
//     🔹 status: active
//     🔹 count: 5

logging.Debug.Logf("Cache hit rate: %.2f%%", hitRate)

// ❌ WRONG: NEVER use these
fmt.Println("message")        // Use logging.Log.Info("message")
log.Println("message")        // Use logging.Log.Info("message")
panic("error")                // Use logging.Log.Error() + return err
logging.Log.Info("Count: %d", count)           // Wrong! Use Infof()
logging.Log.Infof("Processing", "count", count) // Wrong! Use Info()
```

### Naming Conventions
**SSO Session Pattern:** `<organization-alias>-<partition>`
- Examples: `test-org-commercial`, `org2-prod-govcloud`

**Profile Pattern:** `<partition>-<account-alias>-<role>[--<region>]`
- Examples: `commercial-dev-Developer`, `commercial-dev-Developer--us-west-2`

### Host-Based Development

#### Why Host-Based?
- **GUI Support** - Fyne GUI applications require native OpenGL libraries and display access
- **Performance** - Direct execution without container overhead
- **Simplicity** - No container complexity for a single-binary application
- **Native Integration** - Better integration with host system tools and environment

#### System Requirements
- **Go 1.25.10+** - For building and running the application
- **OpenGL libraries** - Automatically installed via `make deps-system`
- **AWS CLI** - For testing session management (optional)

## Building and Distribution

### Development Builds
```bash
# Build for current platform
npm run build
# OR: make build

# Run GUI for testing
npm run run:gui
# OR: go run ./cmd/aws-profile-manager/main.go gui
```

### Production Builds
Production builds are handled by GitHub Actions for cross-platform distribution. The development environment focuses on:
- Consistent development experience
- Fast iteration cycles  
- Comprehensive testing
- Host-native builds for GUI testing

## Platform Considerations

### Linux Developers
- ✅ **Full native support** - Primary development platform
- ✅ **GUI libraries** - OpenGL/X11/Wayland dependencies auto-installed via `make deps-system`
- ✅ **Complete workflow** - Build, test, and run GUI applications natively
- ✅ **Package managers** - Automatic dependency detection (apt, yum, etc.)

### macOS Developers
- ✅ **Native Go support** - Full development capabilities
- ✅ **GUI support** - Fyne works natively with macOS windowing system
- ✅ **Dependencies** - Xcode Command Line Tools provide required libraries
- ⚠️ **Note**: Ensure Xcode CLT installed: `xcode-select --install`

### Windows Developers
- ✅ **Go development** - Full Go toolchain support on Windows
- ✅ **CLI functionality** - Complete CLI development and testing
- ⚠️ **GUI dependencies** - May require manual installation of OpenGL/graphics libraries
- 💡 **Recommendation**: Use WSL2 with Linux workflow for GUI development, or develop on Windows with CI/CD for cross-platform builds

## Debugging

### Debug Logging
Configure debug logging using the `AWS_PROFILE_MANAGER_DEBUG` environment variable:

```bash
# Option 1: Set in .env file (persistent)
AWS_PROFILE_MANAGER_DEBUG=1

# Option 2: Override for single command
AWS_PROFILE_MANAGER_DEBUG=1 npm run dev

# Option 3: When running built binary
AWS_PROFILE_MANAGER_DEBUG=1 ./bin/aws-profile-manager
```

### Interactive Development
```bash
# Direct development commands
go run ./cmd/aws-profile-manager/main.go --help    # Test CLI directly
npm run make:test           # Run tests
npm run make:fmt            # Format code

# GUI development and testing
npm run run:gui             # Build and launch GUI
go run ./cmd/aws-profile-manager/main.go gui      # Launch GUI directly
```

## Contributing

### Before Submitting PRs
1. Format code: `npm run make:fmt`
2. Run static analysis: `npm run make:vet`
3. Run linter: `npm run make:lint`
4. Run vulnerability scan: `npm run make:vuln`
5. Run full test suite: `npm run make:test-coverage`
6. Verify build succeeds: `npm run make:build`
7. Update tests for new functionality
8. Maintain or improve test coverage

### Development Best Practices
- Use npm scripts as the primary interface for development tasks
- Test both CLI and GUI functionality before submitting
- Follow existing naming conventions for profiles and sessions
- Use structured logging with appropriate log levels
- Write comprehensive tests with proper isolation (temp directories)
- Install GUI dependencies with `make deps-system` when needed

## Future Development

### Planned Features
- Enhanced GUI implementation (additional views: installer, sync, profiles, sessions)
- Advanced AWS session management
- Cross-platform installers

### Architecture Goals
- Maintain thread-safe singleton patterns
- Keep clean separation between packages
- Ensure cross-platform compatibility
- Maintain high test coverage standards

## Troubleshooting

### Development Data Issues

**No sessions showing in GUI:**
```bash
# Check if SSO cache was synced correctly
ls -la .dev/aws/sso/cache/
# Should show .json files if you have active sessions

# Refresh from your current AWS CLI cache
./scripts/sync-dev-data.sh
```

**Profiles not loading:**
```bash
# Verify AWS config was copied
cat .dev/aws/config
# Should show your AWS CLI profiles

# Re-sync if needed
./scripts/sync-dev-data.sh
```

**SSO sessions showing as expired:**
```bash
# Refresh your actual AWS SSO sessions first
aws sso login --profile <your-profile-name>

# Then sync the updated cache
./scripts/sync-dev-data.sh
```

**Debug mode not working:**
```bash
# Ensure AWS_PROFILE_MANAGER_DEBUG is set to 1 in .env file
echo "AWS_PROFILE_MANAGER_DEBUG=1" >> .env

# Or set temporarily
AWS_PROFILE_MANAGER_DEBUG=1 go run ./cmd/aws-profile-manager/main.go gui
```

### Common Build Issues

**GUI won't launch:**
```bash
# Install GUI system dependencies
make deps-system

# Verify OpenGL libraries are available
# On Ubuntu/Debian: apt list --installed | grep -i opengl
```

**Test failures:**
```bash
# Run tests with verbose output to see specific failures
go test -v ./internal/...

# Check that each .go file has corresponding _test.go
find internal -name "*.go" -not -name "*_test.go" | while read f; do
    test_file="${f%%.go}_test.go"
    [ ! -f "$test_file" ] && echo "Missing test file: $test_file"
done
```
- Keep high test coverage (95%+ target)
- Ensure cross-platform compatibility
- Preserve clean separation between CLI and GUI components


## Links

- Theme Icons: https://docs.fyne.io/explore/icons/
- Emojis: https://emojipedia.org/card-index
