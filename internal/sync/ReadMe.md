# Sync Package

Centralized configuration synchronization from remote sources (HTTP, S3, Git, local files).

## Documentation

Full documentation is available in the [`docs/packages/sync/`](../../docs/packages/sync/) directory:

- **[ReadMe.md](../../docs/packages/sync/ReadMe.md)** - Complete package documentation
- **[TestingE2e.md](../../docs/packages/sync/TestingE2e.md)** - End-to-end testing guide
- **[TestingStrategy.md](../../docs/packages/sync/TestingStrategy.md)** - Unit vs E2E test strategy

## Quick Start

```go
import (
    "context"
    "aws-profile-manager/internal/sync"
    "aws-profile-manager/internal/settings"
    "aws-profile-manager/internal/task"
)

// Convert settings to config
cfg := sync.ConfigFromSettings(&settings.Get().Sync)

// Sync with progress reporting
result, err := sync.Sync(context.Background(), cfg, sync.Options{}, task.CliReporter{})
if err != nil {
    log.Fatal(err)
}

// Use result.Data (*schema.Schema)
```

## Running Tests

```bash
# Unit tests (fast, no external dependencies)
go test ./internal/sync/...

# E2E tests (requires configuration)
export SYNC_E2E_HTTP_URL="https://example.com/config.json"
go test -tags=e2e ./internal/sync/...
```

See [TestingE2e.md](../../docs/packages/sync/TestingE2e.md) for complete E2E test documentation.
