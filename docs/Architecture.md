# Architecture

This document describes how the packages in AWS Profile Manager relate to each other, what each one is responsible for, and how data flows through the main operations.

## Table of Contents
- [Layer Model](#layer-model)
- [Package Responsibilities](#package-responsibilities)
- [Data Flow](#data-flow)
- [Key Design Decisions](#key-design-decisions)
- [Naming Conventions](#naming-conventions)

---

## Layer Model

The codebase is structured in three horizontal layers. Each layer may only depend on layers below it.

```
┌─────────────────────────────────────────────────────┐
│                 Presentation Layer                  │
│  internal/cli/    internal/gui/                     │
│  - Parse flags    - Build Fyne widgets              │
│  - Call ONE API   - Run work in goroutines          │
│  - Display result - Update UI via fyne.Do()         │
└──────────────────────────┬──────────────────────────┘
                           │ calls
┌──────────────────────────▼──────────────────────────┐
│                 Business Logic Layer                │
│  internal/profiles/   internal/sync/                │
│  internal/awscli/     internal/backup/              │
│  internal/generators/ internal/schema/              │
│                                                     │
│  - Implement operations (install, sync, export)     │
│  - Accept config structs (never import settings)    │
│  - Accept task.Reporter for progress                │
│  - Return result structs                            │
└──────────────────────────┬──────────────────────────┘
                           │ uses
┌──────────────────────────▼──────────────────────────┐
│                Infrastructure Layer                 │
│  internal/task/    internal/settings/               │
│  internal/logging/ internal/core/                   │
│                                                     │
│  - Task execution (subprocess + function)           │
│  - Settings read/write                              │
│  - Structured logging                               │
│  - Application state                               │
└─────────────────────────────────────────────────────┘
```

**The rule**: Business logic never imports `internal/cli` or `internal/gui`. Presentation never directly creates extractors, fetchers, or writers — it calls one function and displays the result.

---

## Package Responsibilities

### Presentation

| Package | Responsibility |
|---|---|
| `internal/cli` | Parse Cobra flags, build config from settings, call one package function, display results |
| `internal/gui` | Build Fyne widgets (MVVM), run long work in goroutines, update UI via `fyne.Do()` |
| `cmd/aws-profile-manager` | Entry point: register commands, call `core.App.Initialize()`, run root command |

### Business Logic

| Package | Responsibility |
|---|---|
| `internal/profiles` | All profile file operations: install, remove, export, import, cheat sheet. The central package for anything that touches `~/.aws/config`. |
| `internal/awscli` | Read and parse the existing `~/.aws/config`. Profile extraction, filtering, SSO session status. |
| `internal/sync` | Fetch a config JSON from a remote source (HTTP, S3, Git, local). Caching, validation, strategy routing. |
| `internal/backup` | Export profiles to a JSON backup file. Import from backup. Settings backup/restore. |
| `internal/generators` | Generate AWS config file text from schema types. Pure functions — no file I/O. |
| `internal/schema` | Core data model: `Schema`, `Organization`, `Partition`, `Account`, `IamUser`, `AssumeRoleChain`, `GenericProfile`, `Preset`. Validation, filtering, naming, JSON serialization. |

### Infrastructure

| Package | Responsibility |
|---|---|
| `internal/task` | Execute subprocesses and Go functions with context cancellation and progress reporting. |
| `internal/settings` | Load, save, and validate application settings. Schema-driven (auto-generates GUI fields). |
| `internal/logging` | Structured, color-coded output. Key-value and formatted variants. Debug mode. |
| `internal/core` | Application state (`core.App`), version constants, time utilities. |
| `internal/test` | Test isolation: `SetupTestEnvironment(t)` creates temp directories and sets env vars. |
| `internal/bundled` | Embedded assets (logo) generated at build time. |
| `internal/schema/test` | Centralized test fixtures: 56 pre-built schemas covering all scenarios. |

---

## Data Flow

### Install: config file → `~/.aws/config`

```
User runs: aws-profile-manager install --config config.json

cli/install.go
  │ Reads flags, builds profiles.Config from settings
  │
  ▼
profiles.NewInstaller(config)
  │
  ▼
profiles.Installer.Install(ctx, opts, reporter)
  │
  ├── schema.LoadFromFile(path)          ← parse config.json into *schema.Schema
  │
  ├── schema.Filter(criteria)            ← apply --organizations, --partitions, --accounts, --roles, --regions filters
  │
  ├── generators.GenerateSsoProfiles()   ┐
  ├── generators.GenerateIamProfiles()   ├─ generate text content per profile type
  ├── generators.GenerateAssumeRole()    ┘
  │
  ├── profiles.configWriter.Write()      ← write managed section to ~/.aws/config
  │   ├── read existing file
  │   ├── remove old managed section (between markers)
  │   └── write new managed section + preserve personal profiles
  │
  └── profiles.CheatSheet.Generate()    ← optional markdown guide to ~/Desktop
```

### Sync: remote → cache → install

```
User runs: aws-profile-manager sync fetch

cli/sync.go
  │ Builds sync.SyncConfig from settings
  │
  ▼
sync.Sync(ctx, cfg, opts, reporter)
  │
  ├── sync.ValidateConfig(cfg)           ← SSRF check, field validation
  │
  ├── sync.Cache.Get()                   ← return cached schema if TTL valid
  │
  ├── sync.createFetcher(cfg)            ← select HttpFetcher / S3Fetcher / GitFetcher
  │
  ├── fetcher.Fetch(ctx, reporter)       ← retrieve raw JSON bytes
  │   └── task.Run(ctx, SubprocessTask)  ← for S3/Git: runs aws/git CLI
  │
  ├── schema.Validate()                  ← reject invalid configs before caching
  │
  └── sync.Cache.Set()                   ← write validated schema to disk
```

### Export: `~/.aws/config` → JSON backup

```
User runs: aws-profile-manager export

cli/export.go
  │ Builds backup.Config + ExportOptions from flags/settings
  │
  ▼
backup.ExportProfiles(ctx, cfg, opts, reporter)
  │
  ├── profiles.Exporter.Export()         ← read ~/.aws/config into *schema.Schema
  │   ├── profiles.configReader.Read()   ← parse AWS config file
  │   └── build Schema{Managed, Unmanaged}
  │
  └── backup.writeBackupFile()           ← write JSON to output path
```

### List Profiles: `~/.aws/config` → filtered display

```
User runs: aws-profile-manager profiles --role Developer

cli/profiles.go
  │ Builds awscli.FilterCriteria from flags
  │
  ▼
awscli.ListProfiles(criteria)
  │
  ├── awscli.Extractor.Extract()         ← parse ~/.aws/config
  │   ├── extractor_sso.go               ← SSO profile parsing
  │   ├── extractor_iam.go               ← IAM profile parsing
  │   └── extractor_metadata.go          ← session metadata
  │
  ├── awscli.Filter.Apply(criteria)      ← filter profiles
  │
  ├── awscli.SessionManager.GetStatus()  ← check SSO session validity
  │
  └── return ProfilesResult
```

---

## Key Design Decisions

### 1. Dependency injection — no settings in business logic

Business logic packages (`profiles`, `sync`, `awscli`, `backup`, `generators`) never import `internal/settings`. They accept config structs as parameters. CLI and GUI layers read settings and build those structs before calling the package.

**Why**: Testability. Tests pass a hand-crafted config struct without touching the file system. Reusability. The package can be used by any caller.

### 2. Reporter pattern — same code for CLI and GUI

Every long-running or subprocess-based function accepts a `task.Reporter`. CLI passes `task.CliReporter{}` (prints to console). GUI passes `task.NoOpReporter{}` or `task.ChannelReporter` (sends to UI). Tests pass `task.NoOpReporter{}` (discards).

**Why**: Avoids duplicate business logic — one implementation, three contexts.

### 3. `profiles` is the single owner of `~/.aws/config` writes

All code that modifies `~/.aws/config` goes through `internal/profiles`. The package owns the managed section markers, the merge logic (preserving personal profiles), and the write operations.

**Why**: Prevents multiple packages from racing on the same file and makes behaviour predictable.

### 4. `generators` are pure functions

The generators package produces text content from schema types and returns it as strings. It never reads or writes files. The `profiles` package calls generators and then writes the output.

**Why**: Generators become trivially testable — input schema in, expected string out, no file system mocking needed.

### 5. Schema validation before caching

`sync.Sync()` always calls `schema.Validate()` on freshly fetched data before writing to cache. Invalid schemas are rejected and the existing cache is preserved.

**Why**: A bad config deployed to a remote source should not break every team member's tool. The cache acts as a buffer of last-known-good state.

### 6. Managed section markers

The managed section in `~/.aws/config` is delimited by configurable start and end marker lines (e.g., `# START - Managed by AWS Profile Manager`). Everything between these markers is replaced on install. Everything outside is preserved.

Markers are read from settings helpers — never hardcoded. See `settings.GetApplication().GetFormattedStartMarker()`.

---

## Naming Conventions

### Profile names (SSO)

```
<partition>-<account-alias>-<role>
<partition>-<account-alias>-<role>--<region>
```

The region suffix is only added when the profile's region differs from the partition's `default_region`.

Examples: `commercial-prod-Developer`, `commercial-prod-Developer--us-west-2`, `govcloud-prod-SystemAdmin`

### SSO session names

```
<org-alias>-<partition>
```

Examples: `my-org-commercial`, `my-org-govcloud`

All SSO profiles in the same org+partition share one session, so users authenticate once per session regardless of how many accounts or roles they have.

### Go symbol names

PascalCase for all multi-word identifiers — including those starting with acronyms. `HttpClient` not `HTTPClient`. `SsoSession` not `SSOSession`. `CliReporter` not `CLIReporter`.

### File names (documentation)

PascalCase: `ConfigurationGuide.md`, `TestingStrategy.md`. Never `ALL_CAPS.md` or `kebab-case.md`.
