# Sync Package — Testing Strategy

The sync package has two distinct categories of tests that serve different purposes and run under different conditions.

## Unit Tests (Default)

**File pattern**: `*_test.go` (no build tag)  
**Run with**: `go test ./internal/sync/...` or `make test`  
**Speed**: Fast (milliseconds)  
**Dependencies**: None — no network, no AWS, no git

Unit tests cover all code paths that can be exercised without real external services:

| What's tested | How |
|---|---|
| Config validation | Invalid URLs, missing fields, SSRF-blocked addresses |
| HTTP fetcher | Mock HTTP server (`net/http/httptest`) |
| S3 fetcher | Mocked subprocess output (fake `aws` CLI response) |
| Git fetcher | Mocked subprocess output |
| Local fetcher | Temp directory with test files |
| Cache read/write | Temp directory |
| Cache TTL expiry | Manipulate file modification time |
| `Sync()` orchestration | Mock fetcher injected directly |
| Schema validation gate | Returns invalid JSON from mock fetcher |
| Context cancellation | Cancel before/during fetch |

**What unit tests do NOT cover**:
- Actual network connectivity to HTTP URLs
- Real AWS S3 API calls
- Real git clone operations
- Multi-region S3 latency or auth flows

### When to add a unit test

Add a unit test whenever you can express the scenario without a live external service. If you find yourself wanting to set `SYNC_E2E_*` env vars to test a new code path, ask whether the same coverage is achievable with a mock first.

---

## E2E Tests

**File pattern**: `*_e2e_test.go` with `//go:build e2e`  
**Run with**: `go test -tags=e2e ./internal/sync/...`  
**Speed**: Slow (seconds to minutes, depends on network)  
**Dependencies**: Real credentials, real services, environment variables

E2E tests verify that the fetchers work against real external systems. They are skipped by default and only run when the required environment variables are set.

| File | What it tests |
|---|---|
| `http_fetcher_e2e_test.go` | Real HTTP/HTTPS fetch, custom auth headers |
| `s3_fetcher_e2e_test.go` | Real S3 get-object with SSO or IAM credentials |
| `git_fetcher_e2e_test.go` | Real git clone via HTTPS and SSH |

### When to run E2E tests

- Before merging changes to a fetcher implementation
- When debugging a real-world sync failure that can't be reproduced with mocks
- When adding a new strategy (run E2E to confirm the happy path works)
- Not in CI by default — they require credentials that vary by environment

### When to add an E2E test

Add an E2E test when you need to verify behaviour that depends on the real service's response format, auth flow, or error messages. For example: a new S3 authentication method, a new git host quirk, or an HTTP API that uses non-standard status codes.

Follow the existing pattern in `*_e2e_test.go`: check for a required env var, skip if not set, then test using a real service.

---

## Decision Guide

```
New code path in a fetcher?
├── Can I mock the external dependency? → YES → write a unit test
│                                         NO  → write an E2E test
│
New validation rule?
└── Always unit test — no external deps needed
│
New cache behaviour?
└── Always unit test — use temp directory
│
Suspected production failure?
└── Try to reproduce with a unit test first
    → If you need the real service → write or run the E2E test
```
