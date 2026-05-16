# Sync Package — E2E Testing Guide

E2E tests run against real external services and are skipped by default. This guide explains how to configure and run each category of E2E test.

## Running E2E Tests

All E2E tests use the `e2e` build tag. They are never included in a normal `go test` run.

```bash
# Run all sync E2E tests (requires all env vars to be set, or tests will be skipped)
go test -tags=e2e -v ./internal/sync/...

# Run only HTTP E2E tests
go test -tags=e2e -v ./internal/sync -run TestHttpFetcher

# Run only S3 E2E tests
go test -tags=e2e -v ./internal/sync -run TestS3Fetcher

# Run only Git E2E tests
go test -tags=e2e -v ./internal/sync -run TestGitFetcher
```

Tests that don't have their required env vars set are skipped, not failed.

---

## HTTP E2E Tests

**File**: `http_fetcher_e2e_test.go`

### Basic fetch

```bash
export SYNC_E2E_HTTP_URL="https://example.com/aws-config.json"
go test -tags=e2e -v ./internal/sync -run TestHttpFetcher_RealHTTP
```

The URL must return a valid `aws-profile-manager` config JSON. The test verifies the fetch succeeds and returns non-empty data.

### Fetch with auth header

```bash
export SYNC_E2E_HTTP_URL="https://internal.example.com/aws-config.json"
export SYNC_E2E_HTTP_HEADER_Authorization="Bearer your-token"
go test -tags=e2e -v ./internal/sync -run TestHttpFetcher_RealHTTPWithAuth
```

Any env var prefixed with `SYNC_E2E_HTTP_HEADER_` becomes a request header. The part after the prefix is the header name.

### Environment variables

| Variable | Required | Description |
|---|---|---|
| `SYNC_E2E_HTTP_URL` | Yes | HTTPS URL to a valid config JSON file |
| `SYNC_E2E_HTTP_HEADER_<Name>` | No | Custom request headers (e.g., `SYNC_E2E_HTTP_HEADER_Authorization`) |

---

## S3 E2E Tests

**File**: `s3_fetcher_e2e_test.go`

Requires the AWS CLI to be installed and configured.

### Basic fetch (SSO)

```bash
# Log in first
aws sso login --profile your-sso-profile

export SYNC_E2E_S3_BUCKET="my-config-bucket"
export SYNC_E2E_S3_KEY="aws-config.json"
export SYNC_E2E_S3_REGION="us-east-1"
export SYNC_E2E_S3_PROFILE="your-sso-profile"

go test -tags=e2e -v ./internal/sync -run TestS3Fetcher_RealS3
```

### Basic fetch (IAM credentials in environment)

```bash
export AWS_ACCESS_KEY_ID="AKIAIOSFODNN7EXAMPLE"
export AWS_SECRET_ACCESS_KEY="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
export AWS_DEFAULT_REGION="us-east-1"

export SYNC_E2E_S3_BUCKET="my-config-bucket"
export SYNC_E2E_S3_KEY="aws-config.json"
export SYNC_E2E_S3_REGION="us-east-1"

go test -tags=e2e -v ./internal/sync -run TestS3Fetcher_RealS3
```

### Environment variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `SYNC_E2E_S3_BUCKET` | Yes | — | S3 bucket name |
| `SYNC_E2E_S3_KEY` | Yes | — | Object key (path within bucket) |
| `SYNC_E2E_S3_REGION` | No | `us-east-1` | AWS region where bucket lives |
| `SYNC_E2E_S3_PROFILE` | No | — | AWS CLI profile name for authentication |

---

## Git E2E Tests

**File**: `git_fetcher_e2e_test.go`

Requires `git` to be installed and accessible on PATH.

### HTTPS (public repository)

```bash
export SYNC_E2E_GIT_REPO_URL="https://github.com/your-org/aws-config.git"
export SYNC_E2E_GIT_FILE_PATH="aws-config.json"
export SYNC_E2E_GIT_BRANCH="main"  # optional, defaults to "main"

go test -tags=e2e -v ./internal/sync -run TestGitFetcher_RealGitHTTPS
```

### HTTPS (private repository with token)

Embed the token in the URL — Git handles credential extraction:

```bash
export SYNC_E2E_GIT_REPO_URL="https://token:ghp_your_token@github.com/your-org/aws-config.git"
export SYNC_E2E_GIT_FILE_PATH="aws-config.json"

go test -tags=e2e -v ./internal/sync -run TestGitFetcher_RealGitHTTPS
```

### SSH

Requires an SSH key configured in your SSH agent or `~/.ssh/`:

```bash
export SYNC_E2E_GIT_REPO_URL="git@github.com:your-org/aws-config.git"
export SYNC_E2E_GIT_FILE_PATH="aws-config.json"

go test -tags=e2e -v ./internal/sync -run TestGitFetcher_RealGitSSH
```

### Environment variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `SYNC_E2E_GIT_REPO_URL` | Yes | — | Git repository URL (HTTPS or SSH) |
| `SYNC_E2E_GIT_FILE_PATH` | Yes | — | Path to config file within the repo |
| `SYNC_E2E_GIT_BRANCH` | No | `main` | Branch, tag, or commit ref |

---

## Expected Test Output

A passing E2E test looks like:

```
=== RUN   TestHttpFetcher_RealHTTP
    http_fetcher_e2e_test.go:26: Testing HTTP fetch from: https://example.com/aws-config.json
    http_fetcher_e2e_test.go:50: Fetched 4821 bytes successfully
--- PASS: TestHttpFetcher_RealHTTP (0.83s)
```

A skipped test (env var not set) looks like:

```
=== RUN   TestHttpFetcher_RealHTTP
--- SKIP: TestHttpFetcher_RealHTTP (0.00s)
    http_fetcher_e2e_test.go:21: SYNC_E2E_HTTP_URL not set - skipping E2E test
```

A failing test shows the error from the fetcher:

```
=== RUN   TestS3Fetcher_RealS3
    s3_fetcher_e2e_test.go:XX: Testing S3 fetch from: s3://my-bucket/config.json
--- FAIL: TestS3Fetcher_RealS3 (2.14s)
    s3_fetcher_e2e_test.go:XX: Fetch failed: subprocess failed: exit status 255
```

In the failure case, run the equivalent `aws s3 cp` command manually to diagnose the credential or permission issue.
