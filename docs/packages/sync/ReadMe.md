# Sync Package

The sync package retrieves a JSON configuration file from a remote source and returns it as a validated `*schema.Schema`. It supports multiple fetch strategies (HTTP, S3, Git, local), TTL-based caching, SSRF protection, and context cancellation.

## Table of Contents
- [Architecture](#architecture)
- [Strategies](#strategies)
- [API](#api)
- [Configuration](#configuration)
- [Caching](#caching)
- [Validation and Security](#validation-and-security)
- [Adding a New Strategy](#adding-a-new-strategy)

---

## Architecture

```
Sync()
  │
  ├── ValidateConfig()        ← rejects bad configuration before any I/O
  │
  ├── Cache.Get()             ← return cached data if TTL not expired
  │
  ├── createFetcher()         ← selects fetcher based on Strategy field
  │   ├── HttpFetcher
  │   ├── S3Fetcher
  │   ├── GitFetcher
  │   └── LocalFetcher
  │
  ├── fetcher.Validate()      ← fetcher-level validation (URL format, bucket name, etc.)
  │
  ├── fetcher.Fetch()         ← retrieves raw JSON bytes
  │
  ├── json.Unmarshal()        ← parse into schema.Schema
  │
  ├── schema.Validate()       ← ALWAYS validates before caching
  │
  └── Cache.Set()             ← write validated schema to disk cache
```

The `Fetcher` interface is the extension point. Each strategy implements:

```go
type Fetcher interface {
    Fetch(ctx context.Context, reporter task.Reporter) ([]byte, error)
    Validate() error
    String() string
}
```

---

## Strategies

### HTTP (`StrategyHTTP`)

Fetches via HTTP/HTTPS with retry logic, custom headers, and TLS verification.

**Implemented in**: `http_fetcher.go`

Key behaviour:
- Retries on transient errors (configurable count and delay)
- SSRF protection: rejects private IP ranges and localhost by default
- TLS certificate verification on by default
- Custom headers for token-based auth

### S3 (`StrategyS3`)

Fetches from an S3 bucket using the AWS CLI subprocess.

**Implemented in**: `s3_fetcher.go`

Key behaviour:
- Invokes `aws s3 cp s3://<bucket>/<key> -` via `task.SubprocessTask`
- Supports authentication via an existing AWS CLI profile
- Explicit environment: only passes the vars needed (no `os.Environ()` inheritance)
- Compatible with SSO-authenticated profiles

### Git (`StrategyGit`)

Clones or fetches a Git repository and reads a file from it.

**Implemented in**: `git_fetcher.go`

Key behaviour:
- Invokes `git clone --depth 1 --branch <branch>` via `task.SubprocessTask`
- Supports SSH and HTTPS remotes
- Uses a configurable working directory for the clone

### Local (`StrategyLocal`)

Reads a file directly from disk. Intended for development and testing.

**Implemented in**: `local_fetcher.go`

---

## API

### `Sync()`

The main entry point. Takes configuration, options, and a reporter; returns a validated schema.

```go
cfg := sync.ConfigFromSettings(&settings.Get().Sync)

result, err := sync.Sync(
    ctx,
    cfg,
    sync.Options{ForceRefresh: false},
    task.CliReporter{},
)
if err != nil {
    return fmt.Errorf("sync failed: %w", err)
}

// result.Data is *schema.Schema
// result.Source is "cache" or the fetcher string
// result.CacheHit is true if served from cache
```

### `ConfigFromSettings()`

Converts `settings.SyncSettings` into a `SyncConfig`. Call this in CLI/GUI layers — never in business logic.

```go
syncSettings := settings.Get().Sync
cfg := sync.ConfigFromSettings(&syncSettings)
```

### `ValidateConfig()`

Validates a `SyncConfig` without fetching. Returns the first validation error, or nil.

```go
if err := sync.ValidateConfig(cfg); err != nil {
    return fmt.Errorf("invalid sync config: %w", err)
}
```

---

## Configuration

`SyncConfig` is populated by `ConfigFromSettings()` in the CLI/GUI layer. The package never imports the settings package directly.

```go
type SyncConfig struct {
    Strategy Strategy  // "http", "s3", "git", "local"

    // HTTP
    HTTPUrl        string
    HTTPHeaders    map[string]string
    HTTPTimeout    time.Duration
    HTTPRetries    int
    HTTPRetryDelay time.Duration
    HTTPTLSVerify  bool

    // S3
    S3Bucket  string
    S3Key     string
    S3Region  string
    S3Profile string   // AWS CLI profile to use for authentication
    S3AWSPath string   // path to aws binary (default: "aws")
    S3AWSEnv  map[string]string

    // Git
    GitRepoURL  string
    GitBranch   string
    GitFilePath string
    GitWorkDir  string
    GitPath     string  // path to git binary (default: "git")
    GitEnv      map[string]string

    // Local
    LocalPath string

    // Cache
    CacheTTL time.Duration
    CacheDir string
}
```

---

## Caching

The cache stores a validated `*schema.Schema` to disk as JSON. It is TTL-aware: entries older than `CacheTTL` are ignored and a fresh fetch is performed.

- **Cache file**: `<CacheDir>/sync_cache.json`
- **TTL**: Configurable via `SyncConfig.CacheTTL` (0 = disabled)
- **Force refresh**: Set `Options.ForceRefresh = true` to skip cache
- **Validation gate**: Only validated schemas are written to cache. If the remote returns an invalid config, the existing cache is preserved and an error is returned.

Cache behaviour on errors:
- Remote unreachable + valid cache → returns cached data with a warning
- Remote returns invalid schema → returns error, cache unchanged
- Cache file corrupt → treated as a cache miss, proceeds to fetch

---

## Validation and Security

### Config validation (`ValidateConfig`)

Runs before any I/O. Checks:
- HTTP: URL is parseable, scheme is `http` or `https`, SSRF check passes
- S3: bucket and key are non-empty
- Local: path is non-empty
- Unknown strategy: returns error immediately

### SSRF protection

HTTP fetcher rejects URLs that resolve to private IP ranges (`10.x.x.x`, `172.16–31.x.x`, `192.168.x.x`) and loopback addresses. This prevents the tool from being used to probe internal infrastructure.

Override with `HTTPBypassSSRF = true` — but only for local development or trusted internal networks.

### Schema validation

`schema.Validate()` is called on every freshly fetched payload before it is cached or returned. An invalid schema from the remote is never cached or used. This protects against:
- Misconfigured remote configs breaking all users
- Partial uploads or truncated responses
- Malformed JSON

---

## Adding a New Strategy

1. **Add a constant** to `types.go`:

```go
const StrategyMyNew Strategy = "mynew"
```

2. **Create a fetcher file** `mynew_fetcher.go`:

```go
type MyNewFetcher struct {
    // config fields
}

func NewMyNewFetcher(/* params */) *MyNewFetcher { ... }

func (f *MyNewFetcher) Fetch(ctx context.Context, reporter task.Reporter) ([]byte, error) {
    // implement fetch, respect ctx.Done(), report progress
}

func (f *MyNewFetcher) Validate() error {
    // validate configuration fields
}

func (f *MyNewFetcher) String() string {
    return "mynew: <description>"
}
```

3. **Register in `createFetcher()`** in `sync.go`:

```go
case StrategyMyNew:
    return NewMyNewFetcher(cfg.MyNewField, ...), nil
```

4. **Add validation** in `ValidateConfig()` in `validation.go`:

```go
case StrategyMyNew:
    return validateMyNewConfig(cfg)
```

5. **Add settings fields** in `internal/settings/sync.go` and `ConfigFromSettings()`.

6. **Write tests** — both unit tests and an E2E test file following the existing pattern. See [TestingStrategy.md](TestingStrategy.md).
