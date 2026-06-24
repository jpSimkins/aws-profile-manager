.PHONY: help build build-all build-local run clean test test-coverage deps deps-dev lint fmt vet vuln vuln-verbose package-desktop install install-desktop uninstall-desktop dev run-help embed-config embed-logo embed-all clean-embed fyne-tool icon generate-logos

# -----------------------------------------------------------------------------
# Variables
# -----------------------------------------------------------------------------
BINARY_NAME := aws-profile-manager
MAIN_PATH := ./cmd/aws-profile-manager
BUILD_DIR := ./bin
SRC_DIR := ./internal
CMD_DIR := ./cmd
APP_NAME := AWS Profile Manager
APP_ID := com.son9ne.aws-profile-manager

# Version information (can be overridden)
# Priority: explicit VERSION= override → git tag → version.go constant → fallback
_GIT_TAG  := $(shell git describe --tags --exact-match 2>/dev/null | sed 's/^v//')
_VGO_TAG  := $(shell grep -m1 'AppVersion\s*=' internal/core/version.go | sed 's/.*"\(.*\)".*/\1/' 2>/dev/null)
VERSION   ?= $(or $(_GIT_TAG),$(_VGO_TAG),0.0.0)
APP_VERSION := $(patsubst v%,%,$(VERSION))
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "")
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Strip debug info and remove absolute paths from binaries
GOFLAGS := -trimpath
LDFLAGS := -s -w -X aws-profile-manager/internal/core.Version=$(VERSION) -X aws-profile-manager/internal/core.Commit=$(COMMIT) -X aws-profile-manager/internal/core.Date=$(DATE)

# Tool locations / versions
GOPATH_BIN := $(shell go env GOPATH)/bin
GOLANGCI_LINT_VERSION ?= latest

# Host OS/arch helpers and naming
HOST_OS   := $(shell go env GOOS)
HOST_ARCH := $(shell go env GOARCH)
SUFFIX    := $(HOST_OS)-$(HOST_ARCH)

FYNE_SOURCE_ARG :=
FYNE_ICON_ARG :=
FYNE_APP_ID_ARG :=
ifeq ($(HOST_OS),windows)
FYNE_SOURCE_ARG := --source-dir $(MAIN_PATH)
FYNE_ICON_ARG := --icon Icon.png
FYNE_APP_ID_ARG := --app-id $(APP_ID)
endif

# Proper extension for current host (used by build)
EXE_EXT :=
ifeq ($(HOST_OS),windows)
EXE_EXT := .exe
endif

# -----------------------------------------------------------------------------
# Default: Show help
# -----------------------------------------------------------------------------
help: ## Show available targets
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z0-9_.-]+:.*##/ { printf "  %-30s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

clean-dev:
	@echo "Cleaning Development Environment..."
	rm -f ./.dev/aws/config
	rm -rf ./.dev/aws/sso/cache
	rm -f ./.dev/config/settings*.json
	rm -f ./.dev/desktop/*.md

# -----------------------------------------------------------------------------
# Download and tidy Go module dependencies
# -----------------------------------------------------------------------------
deps: ## Download and tidy dependencies
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# -----------------------------------------------------------------------------
# Install system dependencies required for Fyne (host architecture only)
# -----------------------------------------------------------------------------
deps-system: ## Install system dependencies (host arch)
	@echo "Installing system dependencies for Fyne GUI framework..."
	@if command -v apt-get >/dev/null 2>&1; then \
		echo "Detected Debian/Ubuntu system"; \
		sudo apt-get update; \
		sudo apt-get install -y gcc libgl1-mesa-dev xorg-dev libxkbcommon-dev pkg-config; \
	elif command -v yum >/dev/null 2>&1; then \
		echo "Detected RHEL/CentOS system"; \
		sudo yum install -y gcc mesa-libGL-devel libX11-devel libXrandr-devel libXinerama-devel libXcursor-devel libXi-devel libXxf86vm-devel pkgconf-pkg-config; \
	elif command -v brew >/dev/null 2>&1; then \
		echo "Detected macOS system (use Xcode CLT / brew as needed)"; \
	else \
		echo "Unknown system; please install C toolchain, OpenGL, X11/Wayland, and pkg-config"; \
		echo "Docs: https://docs.fyne.io/started/"; \
	fi

# -----------------------------------------------------------------------------
# Install developer tools (linters, etc.)
# -----------------------------------------------------------------------------
deps-dev: ## Install developer tools (linters, debugger)
	@echo "Installing developer tools..."
	@echo "• golangci-lint $(GOLANGCI_LINT_VERSION)"
	@GOBIN="$(GOPATH_BIN)" go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	@echo "• dlv (Delve debugger)"
	@GONOSUMDB=* GOBIN="$(GOPATH_BIN)" go install github.com/go-delve/delve/cmd/dlv@latest
	@echo "• govulncheck"
	@GOBIN="$(GOPATH_BIN)" go install golang.org/x/vuln/cmd/govulncheck@latest

# -----------------------------------------------------------------------------
# Convenience: Install both Go and system deps
# -----------------------------------------------------------------------------
deps-all: deps deps-system deps-dev

# -----------------------------------------------------------------------------
# Development commands
# -----------------------------------------------------------------------------
version: ## Show version information
	@echo "Version: $(VERSION)"
	@echo "Commit:  $(COMMIT)"
	@echo "Date:    $(DATE)"

# -----------------------------------------------------------------------------
# Build for current platform
# -----------------------------------------------------------------------------
build: deps ## Build for current platform
	@echo "Building $(BINARY_NAME) for $(HOST_OS)/$(HOST_ARCH)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 go build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o $(BUILD_DIR)/$(BINARY_NAME)$(EXE_EXT) $(MAIN_PATH)
	@echo "Built $(BUILD_DIR)/$(BINARY_NAME)$(EXE_EXT)"

# -----------------------------------------------------------------------------
# Build for current platform with an explicit suffix
# -----------------------------------------------------------------------------
build-local: deps ## Build for current platform (with explicit suffix)
	@mkdir -p $(BUILD_DIR)
	@echo "Building for $(HOST_OS)/$(HOST_ARCH)..."
	CGO_ENABLED=1 go build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o $(BUILD_DIR)/$(BINARY_NAME)-$(HOST_OS)-$(HOST_ARCH)$(EXE_EXT) $(MAIN_PATH)
	@echo "Built $(BUILD_DIR)/$(BINARY_NAME)-$(HOST_OS)-$(HOST_ARCH)$(EXE_EXT)"

# -----------------------------------------------------------------------------
# Explicit, single-platform build targets (call these from CI)
# -----------------------------------------------------------------------------
build-linux-amd64: ## Build linux/amd64
	@mkdir -p $(BUILD_DIR)
	@echo "Building linux/amd64..."
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)

build-linux-arm64: ## Build linux/arm64
	@mkdir -p $(BUILD_DIR)
	@echo "Building linux/arm64..."
	CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)

build-darwin-amd64: ## Build macOS/amd64 (Intel)
	@mkdir -p $(BUILD_DIR)
	@echo "Building darwin/amd64..."
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)

build-darwin-arm64: ## Build macOS/arm64 (Apple Silicon)
	@mkdir -p $(BUILD_DIR)
	@echo "Building darwin/arm64..."
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)

build-windows-amd64: ## Build windows/amd64
	@mkdir -p $(BUILD_DIR)
	@echo "Building windows/amd64..."
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

# -----------------------------------------------------------------------------
# Host-aware build-all: builds only for the current host OS/arch
# NOTE: Cross-compilation with CGO_ENABLED=1 requires platform-specific toolchains
# NOTE: Run 'make embed-all' first to generate embedded resources
# -----------------------------------------------------------------------------
build-all: deps ## Build for this host only (cross-platform builds require host-specific tooling)
	@mkdir -p $(BUILD_DIR)
	@echo "Host detected: $(HOST_OS)/$(HOST_ARCH)"
	@if [ "$(HOST_OS)" = "linux" ]; then \
		$(MAKE) build-linux-$(HOST_ARCH); \
	elif [ "$(HOST_OS)" = "darwin" ]; then \
		$(MAKE) build-darwin-$(HOST_ARCH); \
	elif [ "$(HOST_OS)" = "windows" ]; then \
		$(MAKE) build-windows-amd64; \
	else \
		echo "Unsupported host OS: $(HOST_OS)"; exit 1; \
	fi
	@echo "Done building for host $(HOST_OS)/$(HOST_ARCH)"

# -----------------------------------------------------------------------------
# Testing and quality
# -----------------------------------------------------------------------------
test: ## Run tests
	@echo "Running tests..."
	go test $(CMD_DIR)/... $(SRC_DIR)/...

test-verbose: ## Run tests
	@echo "Running tests..."
	go test -v $(CMD_DIR)/... $(SRC_DIR)/...

test-coverage: ## Run tests with coverage
	@mkdir -p coverage
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage/coverage.out $(CMD_DIR)/... $(SRC_DIR)/...
	go tool cover -html=coverage/coverage.out -o coverage/coverage.html
	go tool cover -func=coverage/coverage.out > coverage/coverage-summary.txt
	@echo "Coverage reports generated in coverage/:"
	@echo "  - coverage/coverage.html"
	@echo "  - coverage/coverage-summary.txt"

bench: ## Run benchmarks
	@echo "Running benchmarks..."
	go test -bench=. $(CMD_DIR)/... $(SRC_DIR)/...

fmt: ## Format Go code
	@echo "Formatting code..."
	go fmt $(CMD_DIR)/... $(SRC_DIR)/...

lint: ## Run linter
	@command -v "$(GOPATH_BIN)/golangci-lint" >/dev/null 2>&1 || $(MAKE) deps-dev
	"$(GOPATH_BIN)/golangci-lint" run $(CMD_DIR)/... $(SRC_DIR)/...

vet: ## Run go vet
	@echo "Running go vet..."
	go vet $(CMD_DIR)/... $(SRC_DIR)/...

vuln: ## Run vulnerability scanner (govulncheck)
	@command -v "$(GOPATH_BIN)/govulncheck" >/dev/null 2>&1 || $(MAKE) deps-dev
	@echo "Running vulnerability scan..."
	"$(GOPATH_BIN)/govulncheck" ./...

vuln-verbose: ## Run vulnerability scanner with full details (includes non-called vulns)
	@command -v "$(GOPATH_BIN)/govulncheck" >/dev/null 2>&1 || $(MAKE) deps-dev
	@echo "Running vulnerability scan (verbose)..."
	"$(GOPATH_BIN)/govulncheck" -show verbose ./...

# -----------------------------------------------------------------------------
# Installation (Linux only for now)
# -----------------------------------------------------------------------------
install: build ## Install binary to ~/.local/bin
	@echo "Installing $(BINARY_NAME) to ~/.local/bin..."
	@mkdir -p ~/.local/bin
	@cp $(BUILD_DIR)/$(BINARY_NAME)$(EXE_EXT) ~/.local/bin/$(BINARY_NAME)
	@chmod +x ~/.local/bin/$(BINARY_NAME)
	@echo "Installed! Make sure ~/.local/bin is in your PATH"
	@echo "Run with: $(BINARY_NAME)"

uninstall: ## Uninstall binary from ~/.local/bin
	@echo "Uninstalling $(BINARY_NAME)..."
	@rm -f ~/.local/bin/$(BINARY_NAME)
	@echo "Uninstalled."

# -----------------------------------------------------------------------------
# Utility commands
# -----------------------------------------------------------------------------
clean: clean-embed ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)/aws-profile-manager*
	rm -rf coverage
	rm -f Icon.png
	go clean

# -----------------------------------------------------------------------------
# Embedded Resources
# -----------------------------------------------------------------------------

# Ensure the Fyne CLI tool is installed (used by bundle/package steps)
fyne-tool:
	@command -v "$(GOPATH_BIN)/fyne" >/dev/null 2>&1 || { \
	  echo "Installing fyne tool..."; \
	  GOBIN="$(GOPATH_BIN)" go install fyne.io/tools/cmd/fyne@latest; \
	}

# Convert Icon.svg to Icon.png for desktop packaging
icon: ## Generate Icon.png from Icon.svg
	@echo "Converting Icon.svg to Icon.png (512x512)..."
	@# Icon.svg already has optimized viewBox with tight cropping
	@# Ensure transparent background for all converters
	@if command -v rsvg-convert >/dev/null 2>&1; then \
		rsvg-convert -w 512 -h 512 --background-color=transparent Icon.svg -o Icon.png; \
		echo "✓ Icon.png generated successfully using rsvg-convert"; \
	elif command -v magick >/dev/null 2>&1; then \
		magick Icon.svg -background none -transparent white  -resize 512x512 Icon.png; \
		echo "✓ Icon.png generated successfully using ImageMagick (magick)"; \
	elif command -v convert >/dev/null 2>&1; then \
		convert -background none -transparent white  -resize 512x512 Icon.svg Icon.png; \
		echo "✓ Icon.png generated successfully using ImageMagick (convert)"; \
	elif command -v inkscape >/dev/null 2>&1; then \
		inkscape Icon.svg --export-filename=Icon.png --export-width=512 --export-height=512 --export-background-opacity=0; \
		echo "✓ Icon.png generated successfully using Inkscape"; \
	else \
		echo "ERROR: No SVG converter found. Please install one of:"; \
		echo "  - librsvg (rsvg-convert)"; \
		echo "  - ImageMagick (convert)"; \
		echo "  - Inkscape"; \
		exit 1; \
	fi

# Generate logo variants from Icon.svg (single source of truth)
generate-logos: ## Generate logo.svg and logo-dark-mode.svg from Icon.svg
	@echo "Generating logo variants from Icon.svg..."
	@./scripts/generate-logos.sh

# Generate embedded logo from assets using fyne bundle
embed-logo: fyne-tool generate-logos ## Generate embedded logo resources
	@echo "Embedding logos from assets/..."
	@mkdir -p internal/bundled
	@cd internal/bundled && "$(GOPATH_BIN)/fyne" bundle \
		--package bundled \
		--name ResourceLogo \
		-o logo_resource.go \
		assets/logo.png
	@cd internal/bundled && "$(GOPATH_BIN)/fyne" bundle \
		--package bundled \
		--name ResourceLogoDarkMode \
		-a -o logo_resource.go \
		assets/logo-dark-mode.png
	@echo "Embedded logo resources generated successfully"

# Generate all embedded resources
embed-all: icon embed-logo ## Generate all embedded resources
	@echo "All embedded resources generated successfully"

# Clean embedded resources
clean-embed: ## Clean generated embedded resources
	@echo "Cleaning embedded resources..."
	@rm -f internal/bundled/logo_resource.go
	@rm -f internal/bundled/assets/logo.png
	@rm -f internal/bundled/assets/logo-dark-mode.png
	@echo "Embedded resources cleaned"

setup: deps deps-dev ## Setup development environment
	@echo "Development environment setup complete!"

# -----------------------------------------------------------------------------
# Package desktop application for distribution (uses fyne package)
# Produces platform-specific artifacts in bin/:
#   Linux:   "AWS Profile Manager-linux-<arch>.tar.xz"
#   macOS:   "AWS Profile Manager-darwin-<arch>.zip"
#   Windows: "AWS Profile Manager-windows-<arch>.exe"
# -----------------------------------------------------------------------------
package-desktop: build fyne-tool ## Package desktop application for distribution
	@echo "Packaging $(APP_NAME) for $(HOST_OS)/$(HOST_ARCH)..."
	@mkdir -p $(BUILD_DIR)
	@echo "Stamping FyneApp.toml with app version $(APP_VERSION) (from $(VERSION))..."
	@tmp_file=$$(mktemp); \
		sed 's/^Version = .*/Version = "$(APP_VERSION)"/' FyneApp.toml > "$$tmp_file" && \
		mv "$$tmp_file" FyneApp.toml
	"$(GOPATH_BIN)/fyne" package \
		--release \
		$(FYNE_SOURCE_ARG) \
		$(FYNE_ICON_ARG) \
		$(FYNE_APP_ID_ARG) \
		--executable $(BUILD_DIR)/$(BINARY_NAME)$(EXE_EXT)
	@if [ "$(HOST_OS)" = "linux" ]; then \
		mv "$(APP_NAME).tar.xz" "$(BUILD_DIR)/$(APP_NAME)-$(SUFFIX).tar.xz"; \
		echo "Created $(BUILD_DIR)/$(APP_NAME)-$(SUFFIX).tar.xz"; \
	elif [ "$(HOST_OS)" = "darwin" ]; then \
		zip -r "$(BUILD_DIR)/$(APP_NAME)-$(SUFFIX).zip" "$(APP_NAME).app"; \
		rm -rf "$(APP_NAME).app"; \
		echo "Created $(BUILD_DIR)/$(APP_NAME)-$(SUFFIX).zip"; \
	elif [ "$(HOST_OS)" = "windows" ]; then \
		mv "$(APP_NAME) Setup.exe" "$(BUILD_DIR)/$(APP_NAME)-$(SUFFIX).exe" 2>/dev/null || \
		mv "$(APP_NAME).exe" "$(BUILD_DIR)/$(APP_NAME)-$(SUFFIX).exe" 2>/dev/null || true; \
		echo "Created $(BUILD_DIR)/$(APP_NAME)-$(SUFFIX).exe"; \
	fi

# -----------------------------------------------------------------------------
# CI/CD helpers
# -----------------------------------------------------------------------------
ci-test: deps vet lint test ## Run all CI tests

release: clean build-all ## Create release builds
	@echo "Release builds created in $(BUILD_DIR)/ directory"