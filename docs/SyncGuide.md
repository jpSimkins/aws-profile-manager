# Sync Guide

Sync lets your team always pull the latest AWS configuration from a central source instead of distributing config files manually. Set it up once, and every team member can run `aws-profile-manager sync fetch` to get current profiles.

## Table of Contents
- [How It Works](#how-it-works)
- [Choosing a Strategy](#choosing-a-strategy)
- [HTTP Setup](#http-setup)
- [S3 Setup](#s3-setup)
- [Git Setup](#git-setup)
- [Local Setup](#local-setup)
- [Configuring the Application](#configuring-the-application)
- [Using Sync](#using-sync)
- [Caching](#caching)
- [Troubleshooting](#troubleshooting)

---

## How It Works

1. An admin hosts the JSON config file somewhere (HTTP server, S3 bucket, Git repo).
2. Each user configures the app to point at that location (once, via settings).
3. When users run `sync fetch` (or when auto-update fires on startup), the app downloads the latest config, validates it, and caches it locally.
4. `install` uses the cached config to write profiles to `~/.aws/config`.

The cached config persists locally so the tool works offline. The cache is refreshed whenever you explicitly `sync fetch` or when the TTL expires (if auto-update is on).

---

## Choosing a Strategy

| Strategy | Best For | Auth Required |
|----------|----------|---------------|
| **HTTP** | Public URLs, internal web servers, CDNs | Optional (basic auth or headers) |
| **S3** | Company-internal configs, fine-grained access control | Yes (SSO or IAM) |
| **Git** | Version-controlled configs, audit trail | Depends on repo visibility |
| **Local** | Development and testing | No |

**Rule of thumb**: Use S3 if the config is sensitive (contains account IDs or role names you don't want public). Use HTTP if the config can be public or if you already have an internal CDN.

---

## HTTP Setup

### Hosting the Config

Upload your `aws-config.json` to any HTTP-accessible location:

- A static web server or CDN (nginx, CloudFront, GitHub Pages)
- An internal web server behind a VPN
- Any URL that returns the JSON file with `Content-Type: application/json`

The URL must be HTTPS in production. The app rejects plain HTTP for remote URLs by default (SSRF and eavesdropping protection).

### User Configuration (GUI)

1. Open **Settings** → **Sync**
2. Set **Strategy** to `HTTP`
3. Enter the **URL** pointing to your config file
4. Optionally add custom headers (e.g., for token-based auth on a private URL)
5. Enable **Auto Update** if you want fresh config on every startup

### User Configuration (settings.json)

```json
{
  "sync": {
    "enabled": true,
    "auto_update": true,
    "strategy": "http",
    "http": {
      "url": "https://internal.example.com/aws-config.json"
    }
  }
}
```

With a custom auth header:

```json
{
  "sync": {
    "enabled": true,
    "strategy": "http",
    "http": {
      "url": "https://internal.example.com/aws-config.json",
      "headers": {
        "Authorization": "Bearer my-token"
      }
    }
  }
}
```

---

## S3 Setup

### Creating the Bucket

```bash
# Create a private bucket
aws s3api create-bucket \
  --bucket my-company-aws-config \
  --region us-east-1

# Block all public access (recommended)
aws s3api put-public-access-block \
  --bucket my-company-aws-config \
  --public-access-block-configuration \
    "BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true"

# Upload the config
aws s3 cp aws-config.json s3://my-company-aws-config/aws-config.json
```

### Granting Access

#### Option A: SSO (Recommended)

Create an IAM policy that allows reading the config object, and attach it to the IAM Identity Center permission set your team uses:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["s3:GetObject"],
      "Resource": "arn:aws:s3:::my-company-aws-config/aws-config.json"
    },
    {
      "Effect": "Allow",
      "Action": ["s3:ListBucket"],
      "Resource": "arn:aws:s3:::my-company-aws-config",
      "Condition": {
        "StringLike": {
          "s3:prefix": ["aws-config.json"]
        }
      }
    }
  ]
}
```

Users authenticate using an existing SSO profile. Specify that profile in the sync settings (see below).

#### Option B: IAM User

Create an IAM user with the same policy above, generate access keys, and configure the user's `~/.aws/credentials` file with a named profile. Then reference that profile in sync settings.

#### Option C: Bucket Policy (Read-Only, Not Recommended)

For low-sensitivity configs, you can make the object readable without credentials:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:GetObject",
      "Resource": "arn:aws:s3:::my-company-aws-config/aws-config.json"
    }
  ]
}
```

> **Note**: A public bucket policy exposes the config file to the internet. Avoid this if the config contains sensitive account IDs or role names.

### User Configuration (GUI)

1. Open **Settings** → **Sync**
2. Set **Strategy** to `S3`
3. Enter the **Bucket** name
4. Enter the **Key** (object path, e.g., `aws-config.json`)
5. Set the **Region** where the bucket lives
6. Set **Profile** to an existing AWS CLI profile that has access (e.g., `commercial-hub-Developer`)

### User Configuration (settings.json)

```json
{
  "sync": {
    "enabled": true,
    "auto_update": true,
    "strategy": "s3",
    "s3": {
      "bucket": "my-company-aws-config",
      "key": "aws-config.json",
      "region": "us-east-1",
      "profile": "commercial-hub-Developer",
      "use_sso": true
    }
  }
}
```

---

## Git Setup

> **Note**: Git sync is available but experimental. It clones or fetches a repository and reads the config file from it.

### Requirements

- `git` must be installed and on your PATH
- The repository must be accessible (public, or SSH/HTTPS with credentials configured)

### User Configuration (settings.json)

```json
{
  "sync": {
    "enabled": true,
    "strategy": "git",
    "git": {
      "repo_url": "git@github.com:my-company/aws-config.git",
      "branch": "main",
      "file_path": "aws-config.json"
    }
  }
}
```

For HTTPS with a private repo, configure Git credentials in your system credential store before using this strategy.

---

## Local Setup

Local sync reads a config file from disk. It's intended for development and testing, not team distribution.

### User Configuration (settings.json)

```json
{
  "sync": {
    "enabled": true,
    "strategy": "local",
    "local": {
      "path": "/path/to/aws-config.json"
    }
  }
}
```

---

## Configuring the Application

### Settings Reference

All sync settings live under the `sync` key in the application settings file (`~/.config/aws-profile-manager/settings.json`).

| Setting | Type | Description |
|---------|------|-------------|
| `enabled` | bool | Enable sync. When false, sync commands do nothing. |
| `auto_update` | bool | Fetch fresh config automatically on startup. |
| `update_on_read` | bool | Check for updates when the app reads cached config. |
| `strategy` | string | Active strategy: `"http"`, `"s3"`, `"git"`, `"local"`. |

The easiest way to configure sync is through the GUI: **Settings → Sync**.

### Bootstrap for New Team Members

The `sync setup` command prints personalized setup instructions based on your organization's configuration:

```bash
aws-profile-manager sync setup
```

Organizations can put setup instructions in the config file. If no instructions are configured, the command prints generic guidance.

---

## Using Sync

### Fetching Configuration

```bash
# Fetch and cache the latest config
aws-profile-manager sync fetch

# Bypass cache and force a fresh download
aws-profile-manager sync fetch --force

# Check what's cached without fetching
aws-profile-manager sync status
```

### Checking Status

```bash
aws-profile-manager sync status
```

Example output:
```
Sync Status
  Strategy:   s3
  Source:     s3://my-company-aws-config/aws-config.json
  Cache age:  2h 14m
  Cache TTL:  24h
  Last fetch: 2026-05-15 09:41:00
```

### Clearing the Cache

```bash
aws-profile-manager sync clear-cache
```

This forces a fresh download on the next `sync fetch` or `install`.

### Full Workflow

```bash
# 1. Fetch the latest config from your org's source
aws-profile-manager sync fetch

# 2. Install the profiles to ~/.aws/config
aws-profile-manager install

# Or do both in one step (install fetches automatically if sync is configured)
aws-profile-manager install
```

---

## Caching

The app caches the downloaded config locally so it works offline and doesn't fetch on every command.

- **Cache location**: `~/.config/aws-profile-manager/cache/`
- **Cache TTL**: Configurable in settings (default: 24 hours)
- **Validation**: The config is validated before it's written to cache. An invalid config from the remote source is rejected and the previous cache is kept.
- **Force refresh**: `sync fetch --force` ignores the TTL and downloads fresh data.

If the cache is expired and the remote source is unreachable, the app falls back to the expired cache and logs a warning.

---

## Troubleshooting

### "Fetch failed: connection refused" or "no such host"

- Check that the URL or bucket name in settings is correct.
- Verify network connectivity to the source (VPN, firewall rules).
- For S3: confirm the AWS CLI profile specified in settings has `s3:GetObject` access.

### "Schema validation failed"

The downloaded file failed schema validation. This means the remote config has an error (missing required field, wrong version, etc.). The previous cache is kept. Report the error to whoever maintains the config file.

### "Invalid configuration: strategy not configured"

Sync is enabled but no strategy is selected, or the selected strategy is missing required fields (e.g., HTTP strategy with no URL). Open **Settings → Sync** and complete the configuration.

### S3: "NoCredentialsError" or "AccessDenied"

- The AWS CLI profile specified in the S3 settings is not valid or lacks permissions.
- Run `aws s3 ls s3://your-bucket/your-key --profile your-profile` in a terminal to test access.
- Ensure the SSO session for that profile is active: `aws-profile-manager sessions`.

### HTTP: "TLS certificate verification failed"

- The server's TLS certificate is invalid or self-signed.
- For internal servers with self-signed certs, either add the CA to your system trust store or contact your admin to use a proper certificate. Do not disable TLS verification in production.
