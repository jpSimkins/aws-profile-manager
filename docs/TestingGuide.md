# Testing Guide

This guide explains how to run, analyze, and troubleshoot tests in the AWS Profile Manager project.

## Table of Contents
- [Quick Start](#quick-start)
- [Running Tests](#running-tests)
- [Test Performance Analysis](#test-performance-analysis)
- [Test Coverage](#test-coverage)
- [Troubleshooting Tests](#troubleshooting-tests)
- [Test Categories](#test-categories)
- [Best Practices](#best-practices)

---

## Quick Start

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific package tests
go test ./internal/sync/...

# Run specific test
go test ./internal/sync -run TestSync_Success -v
```

---

## Running Tests

### Using Make Targets (Recommended)

```bash
# Run all tests
make test

# Run tests with coverage report
make test-coverage

# Run tests with race detector
go test -race ./...

# Run tests with verbose output
go test -v ./...
```

### Using NPM Scripts

```bash
# Run all tests (wraps Make)
npm run make:test

# Run with coverage (wraps Make)
npm run make:test-coverage
```

### Using Go Commands Directly

```bash
# All tests
go test ./...

# Specific package
go test ./internal/sync

# Specific test function
go test ./internal/sync -run TestSync_Success

# Multiple packages
go test ./internal/sync ./internal/task

# Verbose output
go test -v ./internal/sync

# Short mode (skip long-running tests)
go test -short ./...
```

### Package-Specific Tests

```bash
# Sync package
go test ./internal/sync/...

# Task package
go test ./internal/task/...

# CLI commands
go test ./internal/cli/...

# GUI components
go test ./internal/gui/...

# All internal packages
go test ./internal/...
```

---

## Test Performance Analysis

### Timing Tests

**Measure total test time**:
```bash
time go test ./internal/sync/...
```

Output:
```
ok  aws-profile-manager/internal/sync  1.043s

real    0m1.418s
user    0m0.678s
sys     0m0.308s
```

**Measure specific test**:
```bash
time go test ./internal/sync -run TestSync_Success -v
```

### Finding Slow Tests

**Find slowest tests in a package**:
```bash
go test ./internal/sync -v 2>&1 | \
  grep -E "^(--- PASS:|--- FAIL:)" | \
  grep -E "[0-9]+\.[0-9]+s" | \
  sort -t'(' -k2 -rn | \
  head -10
```

Output:
```
--- PASS: TestHttpFetcher_Timeout (0.20s)
--- PASS: TestHttpFetcher_ContextCancellation (0.05s)
--- PASS: TestCache_Expiration (0.15s)
```

**Find slowest tests across all packages**:
```bash
go test ./internal/... -v 2>&1 | \
  grep -E "^(--- PASS:|--- FAIL:)" | \
  grep -E "[0-9]+\.[0-9]+s" | \
  sort -t'(' -k2 -rn | \
  head -20
```

### Clearing Test Cache

Go caches successful test results. Clear cache to force re-run:

```bash
# Clear test cache
go clean -testcache

# Clear and run tests
go clean -testcache && go test ./...

# Clear cache and time tests
go clean -testcache && time go test ./internal/sync/...
```

**Why clear cache?**
- Verify tests still pass after changes
- Get accurate timing measurements
- Debug flaky tests
- Ensure tests aren't passing due to stale cache

### Benchmarking Tests

**Run benchmarks**:
```bash
# All benchmarks
go test -bench=. ./...

# Specific package
go test -bench=. ./internal/sync

# Specific benchmark
go test -bench=BenchmarkCache ./internal/sync

# With memory stats
go test -bench=. -benchmem ./internal/sync
```

### CPU and Memory Profiling

**CPU profiling**:
```bash
go test -cpuprofile=cpu.prof ./internal/sync
go tool pprof cpu.prof
```

**Memory profiling**:
```bash
go test -memprofile=mem.prof ./internal/sync
go tool pprof mem.prof
```

**Interactive profiling session**:
```
(pprof) top10          # Show top 10 functions
(pprof) list FuncName  # Show source for specific function
(pprof) web            # Open browser visualization
```

---

## Test Coverage

### Basic Coverage

**Run tests with coverage**:
```bash
go test -cover ./internal/sync
```

Output:
```
ok  aws-profile-manager/internal/sync  1.043s  coverage: 95.2% of statements
```

### Detailed Coverage Report

**Generate HTML coverage report**:
```bash
# Single package
go test -coverprofile=coverage.out ./internal/sync
go tool cover -html=coverage.out

# All packages
go test -coverprofile=coverage.out ./internal/...
go tool cover -html=coverage.out
```

**Generate coverage summary**:
```bash
go test -coverprofile=coverage.out ./internal/...
go tool cover -func=coverage.out
```

Output:
```
aws-profile-manager/internal/sync/cache.go:15:     NewCache        100.0%
aws-profile-manager/internal/sync/cache.go:23:     Get             87.5%
aws-profile-manager/internal/sync/cache.go:45:     Set             100.0%
```

### Package Coverage Analysis

**Coverage by package**:
```bash
go test -cover ./internal/... | grep -E "^ok|coverage"
```

Output:
```
ok  aws-profile-manager/internal/sync      1.043s  coverage: 95.2%
ok  aws-profile-manager/internal/task      0.661s  coverage: 98.1%
ok  aws-profile-manager/internal/cli       6.628s  coverage: 75.0%
```

**Find packages below coverage threshold**:
```bash
go test -cover ./internal/... 2>&1 | \
  grep "coverage:" | \
  awk '{if ($5 < 80) print $2, $5}'
```

### Coverage Trends

**Track coverage over time**:
```bash
# Generate coverage report with date
DATE=$(date +%Y%m%d)
go test -coverprofile=coverage_$DATE.out ./internal/...
go tool cover -func=coverage_$DATE.out > coverage_$DATE.txt

# Compare coverage
diff coverage_20250101.txt coverage_20250201.txt
```

---

## Troubleshooting Tests

### Verbose Test Output

**Run tests with full output**:
```bash
go test -v ./internal/sync
```

**Run with test names only**:
```bash
go test -v ./internal/sync 2>&1 | grep "RUN\|PASS\|FAIL"
```

### Filtering Tests

**Run tests matching pattern**:
```bash
# Run all sync-related tests
go test ./internal/sync -run Sync

# Run all cache tests
go test ./internal/sync -run Cache

# Run specific test
go test ./internal/sync -run TestSync_Success

# Run tests with regex
go test ./internal/sync -run "Test.*Cache"
```

### Running Failed Tests Only

**Run only failed tests**:
```bash
# First run to find failures
go test ./... | tee test_output.log

# Extract failed test names
grep "FAIL:" test_output.log

# Re-run specific failed test
go test ./internal/sync -run TestFailedFunction -v
```

### Debug Test Output

**Print detailed error information**:
```bash
go test -v ./internal/sync 2>&1 | grep -A 10 "FAIL:"
```

**Show test output even for passing tests**:
```bash
go test -v ./internal/sync | grep -v "^==="
```

### Race Condition Detection

**Run tests with race detector**:
```bash
# All tests
go test -race ./...

# Specific package
go test -race ./internal/sync

# With verbose output
go test -race -v ./internal/sync
```

**Note**: Race detector adds significant overhead (~10x slower)

### Parallel Test Execution

**Control parallelism**:
```bash
# Default (GOMAXPROCS)
go test ./...

# Limit to 1 (sequential)
go test -p 1 ./...

# Limit to 4 packages in parallel
go test -p 4 ./...

# Limit CPU cores per test
go test -parallel 2 ./...
```

### Test Timeouts

**Set test timeout**:
```bash
# Default timeout: 10 minutes
go test ./...

# Custom timeout
go test -timeout 30s ./internal/sync

# Long-running tests
go test -timeout 30m ./...
```

### Debugging Failing Tests

**Step 1: Run with verbose output**
```bash
go test -v ./internal/sync -run TestFailingTest
```

**Step 2: Check for race conditions**
```bash
go test -race ./internal/sync -run TestFailingTest
```

**Step 3: Run with debug logging**
```bash
AWS_PROFILE_MANAGER_DEBUG=1 go test -v ./internal/sync -run TestFailingTest
```

**Step 4: Run test multiple times**
```bash
# Run 10 times to catch flaky tests
for i in {1..10}; do
  go test ./internal/sync -run TestFailingTest || break
done
```

**Step 5: Profile the test**
```bash
go test -cpuprofile=cpu.prof ./internal/sync -run TestFailingTest
go tool pprof cpu.prof
```

---

## Test Categories

### Unit Tests

Fast, isolated tests with no external dependencies.

**Run unit tests**:
```bash
# All unit tests (default)
go test ./internal/sync/...
```

**Characteristics**:
- Use mocks (httptest, local files)
- Fast (<1s per package)
- No network calls
- No external services

### E2E Tests

End-to-end tests with real external services (opt-in only).

**Run E2E tests**:
```bash
# Configure environment
export SYNC_E2E_HTTP_URL="https://example.com/config.json"
export SYNC_E2E_S3_BUCKET="my-bucket"
export SYNC_E2E_S3_KEY="config.json"

# Run with e2e tag
go test -tags=e2e ./internal/sync/... -v

# Run specific E2E test
go test -tags=e2e ./internal/sync -run TestHttpFetcher_RealHttp -v
```

**Characteristics**:
- Require build tag: `-tags=e2e`
- Make real network calls
- Require configuration via env vars
- Auto-skip if env vars not set
- Slow (network-dependent)

**See**: [`docs/packages/sync/E2eTesting.md`](packages/sync/E2eTesting.md) for complete E2E testing guide.

### Integration Tests

Tests that verify package interactions but use mocks for external services.

**Run integration tests**:
```bash
go test ./internal/cli/... -v
```

**Characteristics**:
- Test multiple packages together
- Use test fixtures
- May involve file I/O
- Slower than unit tests (~0.5s per test)

---

## Best Practices

### Before Committing

Run this sequence:

```bash
# 1. Format code
make fmt

# 2. Run linter (must be clean)
make lint

# 3. Clear cache and run tests
go clean -testcache && make test

# 4. Check coverage
make test-coverage

# 5. Build to ensure no compile errors
make build
```

**One-liner**:
```bash
make fmt && make vet && make lint && go clean -testcache && make test-coverage && make build
```

### Writing Fast Tests

**DO**:
- ✅ Use `httptest.NewServer()` for HTTP tests
- ✅ Use local git repos for git tests
- ✅ Use `test.SetupTestEnvironment(t)` for isolation
- ✅ Use short timeouts (50-200ms) for timeout tests
- ✅ Mock external services

**DON'T**:
- ❌ Make real network calls in unit tests
- ❌ Use `time.Sleep()` longer than necessary
- ❌ Access real AWS services
- ❌ Depend on external state

### Debugging Flaky Tests

**Technique 1: Run multiple times**
```bash
for i in {1..100}; do
  go test ./internal/sync -run TestFlakyTest || break
done
```

**Technique 2: Run with race detector**
```bash
go test -race ./internal/sync -run TestFlakyTest -count=100
```

**Technique 3: Add verbose logging**
```bash
AWS_PROFILE_MANAGER_DEBUG=1 go test -v ./internal/sync -run TestFlakyTest
```

**Technique 4: Run in parallel**
```bash
go test -parallel 10 ./internal/sync -run TestFlakyTest -count=100
```

### Test Isolation

**Always use test helpers**:
```bash
# ✅ Good: Isolated environment
func TestMyFunction(t *testing.T) {
    test.SetupTestEnvironment(t)  // Creates temp dirs
    // Test uses temp directories automatically
}

# ❌ Bad: Uses real paths
func TestMyFunction(t *testing.T) {
    os.WriteFile("~/.aws/config", data, 0644)  // Writes to real file!
}
```

### Test Coverage Goals

**Target coverage by package**:
- `internal/sync`: 95%+
- `internal/task`: 95%+
- `internal/generators`: 95%+
- `internal/awscli`: 85%+
- `internal/cli`: 75%+

**Check current coverage**:
```bash
go test -cover ./internal/... | grep coverage
```

---

## Common Issues and Solutions

### Issue: Tests Are Slow

**Diagnose**:
```bash
go test -v ./internal/cli 2>&1 | \
  grep -E "^(--- PASS:|--- FAIL:)" | \
  grep -E "[0-9]+\.[0-9]+s" | \
  sort -t'(' -k2 -rn | \
  head -10
```

**Solutions**:
- Reduce `time.Sleep()` durations
- Use mocks instead of real services
- Check for unnecessary file I/O
- Run tests in parallel

### Issue: Test Cache Hides Failures

**Solution**:
```bash
# Always clear cache before important test runs
go clean -testcache && go test ./...
```

### Issue: Race Conditions

**Diagnose**:
```bash
go test -race ./internal/sync
```

**Common causes**:
- Concurrent access to shared variables
- Missing mutex locks
- Channel operations without synchronization

### Issue: Tests Pass Locally, Fail in CI

**Possible causes**:
1. Different Go version
2. Different environment variables
3. Cached test results
4. Timing-dependent tests

**Debug**:
```bash
# Match CI environment
go clean -testcache
GO_VERSION=1.21 go test ./...

# Check for environment dependencies
env -i go test ./...
```

### Issue: Cannot Reproduce Test Failure

**Techniques**:
```bash
# Run many times
go test -count=1000 ./internal/sync -run TestFlaky

# Run with race detector
go test -race -count=100 ./internal/sync -run TestFlaky

# Run with shuffle
go test -shuffle=on ./internal/sync -run TestFlaky
```

---

## Advanced Testing

### Test Shuffling

**Randomize test execution order**:
```bash
# Run tests in random order
go test -shuffle=on ./internal/sync

# Use specific seed for reproducibility
go test -shuffle=1234567890 ./internal/sync
```

### Subtest Filtering

**Run specific subtests**:
```bash
# Run all subtests of TestValidate
go test ./internal/sync -run TestValidate

# Run specific subtest
go test ./internal/sync -run TestValidate/valid_config

# Run multiple subtests
go test ./internal/sync -run "TestValidate/(valid|invalid)"
```

### JSON Output

**Generate JSON test output**:
```bash
go test -json ./internal/sync > test_output.json

# Parse with jq
go test -json ./internal/sync | jq -r 'select(.Action=="fail")'
```

### Coverage Analysis Tools

**Using go-cover-treemap**:
```bash
# Install
go install github.com/nikolaydubina/go-cover-treemap@latest

# Generate treemap
go test -coverprofile=coverage.out ./internal/...
go-cover-treemap -coverprofile coverage.out > coverage.svg
```

**Using gocov**:
```bash
# Install
go install github.com/axw/gocov/gocov@latest

# Generate HTML report
go test -coverprofile=coverage.out ./internal/...
gocov convert coverage.out | gocov-html > coverage.html
```

---

## Quick Reference

### Most Common Commands

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Clear cache and run
go clean -testcache && go test ./...

# Time tests
time go test ./internal/sync/...

# Find slow tests
go test -v ./internal/sync 2>&1 | grep -E "PASS.*[0-9]" | sort -t'(' -k2 -rn

# Run specific test
go test ./internal/sync -run TestSync_Success -v

# Run E2E tests
go test -tags=e2e ./internal/sync/... -v

# Race detection
go test -race ./...

# Generate coverage report
go test -coverprofile=coverage.out ./internal/...
go tool cover -html=coverage.out
```

### Test Flags Quick Reference

| Flag | Description | Example |
|------|-------------|---------|
| `-v` | Verbose output | `go test -v ./...` |
| `-run` | Run specific tests | `go test -run TestSync` |
| `-cover` | Show coverage | `go test -cover ./...` |
| `-coverprofile` | Coverage profile | `go test -coverprofile=c.out` |
| `-race` | Race detection | `go test -race ./...` |
| `-timeout` | Test timeout | `go test -timeout 30s` |
| `-parallel` | Parallel tests | `go test -parallel 4` |
| `-count` | Run N times | `go test -count 10` |
| `-shuffle` | Randomize order | `go test -shuffle=on` |
| `-short` | Skip long tests | `go test -short ./...` |
| `-tags` | Build tags | `go test -tags=e2e` |
| `-json` | JSON output | `go test -json ./...` |

---

## See Also

- [Contributing Guide](../CONTRIBUTING.md) - Development workflow
- [Copilot Instructions](../.github/copilot-instructions.md) - Testing standards
- [E2E Testing Guide](packages/sync/E2eTesting.md) - End-to-end test setup
- [Test Strategy](packages/sync/TestStrategy.md) - Unit vs E2E approach
