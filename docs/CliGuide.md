# CLI Command Reference

Complete reference for all AWS Profile Manager CLI commands with detailed examples and use cases.

## Table of Contents
- [Global Options](#global-options)
- [install](#install)
- [export](#export)
- [import](#import)
- [profiles](#profiles)
- [sessions](#sessions)
- [sync](#sync)
- [gui](#gui)
- [version](#version)
- [Integration Examples](#integration-examples)
- [Troubleshooting](#troubleshooting)

---

## Global Options

These options work with any command:

```bash
--config FILE    Specify configuration file path (overrides sync cache)
--help           Show command help
--version        Show version information
```

**Example:**
```bash
# Use a specific config file instead of sync cache
aws-profile-manager install --config /path/to/custom-config.json

# Show help for any command
aws-profile-manager install --help
```

---

## install

Generate AWS CLI profiles from configuration and write them to `~/.aws/config`.

### Usage

```bash
aws-profile-manager install [OPTIONS]
```

### Options

| Flag | Type | Description |
|------|------|-------------|
| `--organizations` | strings | Filter by organization names (comma-separated) |
| `--partitions` | strings | Filter by partitions: `commercial`, `govcloud` (comma-separated) |
| `--accounts` | strings | Filter by account aliases (comma-separated) |
| `--roles` | strings | Filter by role names (comma-separated) |
| `--regions` | strings | Filter by specific regions (comma-separated) |
| `--all-regions` | flag | Include all available regions for each account (overrides `--regions`) |
| `--output` | string | AWS config file path (defaults to `~/.aws/config`) |
| `--dry-run` | flag | Show what would be installed without making changes |
| `--remove` | flag | Remove installed profiles and cheat sheet |
| `--cheat-sheet` | string | Generate Markdown cheat sheet (optionally specify path) |
| `--cheat-sheet-only` | flag | Skip updating AWS config; only generate cheat sheet |
| `--verbose` | flag | Enable verbose output |

### Examples

**Install all profiles:**
```bash
aws-profile-manager install
```

**Install profiles for specific organizations:**
```bash
aws-profile-manager install --organizations=org1,org2
```

**Install only development accounts:**
```bash
aws-profile-manager install --accounts=dev,staging
```

**Install specific roles across all accounts:**
```bash
aws-profile-manager install --roles=Developer,ReadOnly
```

**Install with all regions (not just default):**
```bash
aws-profile-manager install --all-regions
```

**Combine multiple filters:**
```bash
aws-profile-manager install \
  --organizations=my-org \
  --partitions=commercial \
  --roles=Developer \
  --regions=us-east-1,us-west-2
```

**Preview changes without applying:**
```bash
aws-profile-manager install --dry-run
```

**Generate cheat sheet only:**
```bash
aws-profile-manager install --cheat-sheet-only
```

**Remove installed profiles:**
```bash
aws-profile-manager install --remove
```

**Use custom config file:**
```bash
aws-profile-manager install --config /path/to/config.json
```

**Output to custom location:**
```bash
aws-profile-manager install --output /custom/path/config
```

---

## export

Export AWS profiles from `~/.aws/config` to backup JSON format.

### Usage

```bash
aws-profile-manager export [OPTIONS]
```

### Options

| Flag | Type | Description |
|------|------|-------------|
| `-o, --output` | string | Output JSON file path (required) |
| `--include-managed` | flag | Include managed profiles (between markers) |
| `--include-above` | flag | Include personal profiles above managed section |
| `--include-below` | flag | Include personal profiles below managed section |
| `--description` | string | Add description metadata to export |
| `--exclude-settings` | flag | Exclude application settings from backup |
| `--verbose` | flag | Show detailed export information |

**Default:** If no `--include-*` flags are provided, exports everything (full backup).

### Examples

**Full backup (default):**
```bash
aws-profile-manager export --output backup.json
```

**Managed profiles only (for installer config):**
```bash
aws-profile-manager export --include-managed --output installer.json
```

**Personal profiles only:**
```bash
aws-profile-manager export --include-above --include-below --output personal.json
```

**With description metadata:**
```bash
aws-profile-manager export \
  --output backup.json \
  --description "Pre-OS-wipe backup"
```

**Without application settings:**
```bash
aws-profile-manager export --output backup.json --exclude-settings
```

### Use Cases

1. **Company admins creating installer configs** - Export managed profiles only
2. **Personal backup before OS reinstall** - Full backup with all sections
3. **Disaster recovery** - Regular backups stored in Git or S3
4. **Configuration migration** - Move profiles between machines

---

## import

Import AWS profiles from a backup JSON file to `~/.aws/config`.

### Usage

```bash
aws-profile-manager import [backup-file] [OPTIONS]
```

### Options

| Flag | Type | Description |
|------|------|-------------|
| `--backup` | string | Backup JSON file to import (alternative to positional arg) |
| `--include-managed` | flag | Include managed profiles (between markers) |
| `--include-above` | flag | Include personal profiles above managed section |
| `--include-below` | flag | Include personal profiles below managed section |
| `--dry-run` | flag | Preview import without making changes |
| `--ignore-settings` | flag | Don't restore application settings from backup |
| `--backup-current-settings` | flag | Backup current settings before restoring (default: true) |
| `--cheat-sheet` | flag | Generate cheat sheet after import |
| `--verbose` | flag | Show detailed import information |

**Default:** If no `--include-*` flags are provided, imports everything (full restore).

### Examples

**Import full backup:**
```bash
aws-profile-manager import backup.json
```

**Import managed profiles only:**
```bash
aws-profile-manager import --include-managed backup.json
```

**Import personal profiles only:**
```bash
aws-profile-manager import --include-above --include-below personal.json
```

**Preview import without making changes:**
```bash
aws-profile-manager import --dry-run backup.json
```

**Import without restoring settings:**
```bash
aws-profile-manager import backup.json --ignore-settings
```

**Backup current settings before restoring:**
```bash
aws-profile-manager import backup.json --backup-current-settings
```

---

## profiles

List and filter AWS CLI profiles from `~/.aws/config`.

Supports all profile types: SSO, IAM, and AssumeRole.

### Usage

```bash
aws-profile-manager profiles [OPTIONS]
```

### Options

| Flag | Type | Description |
|------|------|-------------|
| `--account-id` | string | Filter profiles by AWS account ID |
| `--role` | string | Filter profiles by role name |
| `--region` | string | Filter profiles by region |
| `--session` | string | Filter profiles by SSO session name |
| `--pattern` | string | Filter profiles by name pattern (regex) |
| `--verbose` | flag | Show detailed profile information |

### Examples

**List all profiles:**
```bash
aws-profile-manager profiles
```

**Filter by account ID:**
```bash
aws-profile-manager profiles --account-id 123456789012
```

**Filter by role name:**
```bash
aws-profile-manager profiles --role Developer
```

**Filter by region:**
```bash
aws-profile-manager profiles --region us-east-1
```

**Filter by SSO session:**
```bash
aws-profile-manager profiles --session my-org-commercial
```

**Regex pattern matching:**
```bash
# All production profiles
aws-profile-manager profiles --pattern "prod.*"

# All Developer roles
aws-profile-manager profiles --pattern ".*-Developer"
```

**Verbose output with details:**
```bash
aws-profile-manager profiles --verbose
```

### Output Format

**Standard:**
```
Found 3 profiles:

commercial-dev-Developer
commercial-prod-Developer
commercial-staging-Developer
```

**Verbose:**
```
Profile: commercial-dev-Developer
  Type: SSO
  Account: 123456789012
  Role: Developer
  Region: us-east-1
  Session: my-org-commercial
```

---

## sessions

List AWS SSO sessions and their status (active/expired).

### Usage

```bash
aws-profile-manager sessions [OPTIONS]
```

### Options

| Flag | Type | Description |
|------|------|-------------|
| `--verbose` | flag | Show detailed session information |
| `--refresh` | flag | Force refresh session status |

### Examples

**Check session status:**
```bash
aws-profile-manager sessions
```

**Verbose details:**
```bash
aws-profile-manager sessions --verbose
```

**Force refresh:**
```bash
aws-profile-manager sessions --refresh
```

### Output Format

**Standard:**
```
🟢 my-org-commercial (Valid until 2025-10-06 14:30:00)
🔴 my-org-govcloud (Expired)
⚪ other-org-commercial (No session)
```

**Verbose:**
```
Session: my-org-commercial
  Status: Active ✓
  Start URL: https://my-org.awsapps.com/start
  Region: us-east-1
  Expires: 2025-10-06 14:30:00 UTC
  Time Remaining: 7 hours 23 minutes

Session: my-org-govcloud
  Status: Expired ✗
  Start URL: https://start.us-gov-home.awsapps.com/directory/my-org
  Region: us-gov-west-1
  Expired: 2025-10-05 08:15:00 UTC
```

---

## sync

Manage AWS configuration synchronization from remote sources.

### Subcommands

- `fetch` - Fetch configuration from remote source
- `status` - Show sync configuration and cache status
- `clear-cache` - Remove locally cached configuration
- `setup` - Show setup instructions for new hires

### sync fetch

Fetch AWS configuration from the configured remote source.

```bash
aws-profile-manager sync fetch [OPTIONS]
```

**Options:**
- `-f, --force` - Force fetch even if cache is recent
- `-v, --verbose` - Show detailed fetch information

**Examples:**
```bash
# Fetch latest configuration
aws-profile-manager sync fetch

# Force fetch (ignore cache)
aws-profile-manager sync fetch --force

# Verbose output
aws-profile-manager sync fetch --verbose
```

### sync status

Display current sync configuration and cache information.

```bash
aws-profile-manager sync status [OPTIONS]
```

**Options:**
- `-v, --verbose` - Show detailed status information

**Examples:**
```bash
# Check sync status
aws-profile-manager sync status

# Detailed status
aws-profile-manager sync status --verbose
```

**Output:**
```
Sync Status:
  Strategy: HTTP
  URL: https://internal.example.com/aws-config.json
  Cache Age: 2 hours 15 minutes
  Last Sync: 2025-10-05 12:45:00 UTC
  Status: ✓ Up to date
```

### sync clear-cache

Remove the locally cached configuration.

```bash
aws-profile-manager sync clear-cache [OPTIONS]
```

**Options:**
- `-y, --yes` - Skip confirmation prompt

**Examples:**
```bash
# Clear cache (with confirmation)
aws-profile-manager sync clear-cache

# Skip confirmation
aws-profile-manager sync clear-cache --yes
```

### sync setup

Display bootstrap instructions for new team members.

```bash
aws-profile-manager sync setup
```

Shows step-by-step guidance based on your organization's sync strategy (HTTP, S3 SSO, S3 public).

---

## gui

Launch the graphical user interface.

### Usage

```bash
aws-profile-manager gui
```

No additional options. The GUI provides interactive access to all CLI functionality with visual controls for:
- Installing profiles with filters and presets
- Viewing session status
- Listing and searching profiles
- Managing sync configuration
- Exporting and importing backups

---

## version

Display version and build information.

### Usage

```bash
aws-profile-manager version
```

**Output:**
```
AWS Profile Manager
Version: 1.0.0
Git Commit: a1b2c3d
Build Date: 2025-10-01 14:30:00 UTC
Go Version: go1.22.1
```

---

## Integration Examples

### CI/CD Pipeline

```bash
#!/bin/bash
# Deploy script for CI/CD pipeline

# Fetch latest config
aws-profile-manager sync fetch

# Install production profiles only
aws-profile-manager install \
  --organizations=production \
  --partitions=commercial

# Verify profile was created
aws-profile-manager profiles --pattern "commercial-prod.*"

# Use profile for deployment
aws --profile commercial-prod-Developer s3 sync ./dist/ s3://my-bucket/
```

### Development Workflow

```bash
#!/bin/bash
# Daily development setup

# Check SSO sessions
if ! aws-profile-manager sessions | grep -q "🟢"; then
    echo "No active sessions. Please log in:"
    aws sso login --profile commercial-dev-Developer
fi

# Fetch latest profiles
aws-profile-manager sync fetch

# Install dev profiles only
aws-profile-manager install \
  --accounts=dev,staging \
  --roles=Developer

# List available profiles
aws-profile-manager profiles --pattern "dev.*"
```

### Profile Maintenance

```bash
#!/bin/bash
# Weekly profile audit and cleanup

# Export current profiles for backup
aws-profile-manager export \
  --output "backup-$(date +%Y%m%d).json" \
  --description "Weekly backup"

# Audit existing profiles
aws-profile-manager profiles --verbose > profile-audit.txt

# Clean slate - remove and reinstall
aws-profile-manager install --remove
aws-profile-manager sync fetch
aws-profile-manager install

echo "Profile maintenance complete"
```

### Multi-Environment Setup

```bash
#!/bin/bash
# Install profiles for multiple environments

# Production: Admin access only
aws-profile-manager install \
  --organizations=prod-org \
  --roles=SystemAdmin \
  --output ~/.aws/config-production

# Development: All roles
aws-profile-manager install \
  --organizations=dev-org \
  --output ~/.aws/config-development

# Use with AWS_CONFIG_FILE
export AWS_CONFIG_FILE=~/.aws/config-production
aws s3 ls  # Uses production config
```

---

## Troubleshooting

### Debug Mode

Enable detailed logging for troubleshooting:

```bash
# Temporary (single command)
AWS_PROFILE_MANAGER_DEBUG=1 aws-profile-manager install

# Persistent (set environment variable)
export AWS_PROFILE_MANAGER_DEBUG=1
aws-profile-manager install
```

### Common Issues

**Profiles not appearing:**
```bash
# Verify config file location
aws-profile-manager profiles --verbose

# Check if profiles were actually written
cat ~/.aws/config | grep "profile commercial"

# Try dry-run to see what would be installed
aws-profile-manager install --dry-run
```

**Sync fetch fails:**
```bash
# Check sync status first
aws-profile-manager sync status

# Try with debug mode
AWS_PROFILE_MANAGER_DEBUG=1 aws-profile-manager sync fetch --verbose

# Clear cache and retry
aws-profile-manager sync clear-cache --yes
aws-profile-manager sync fetch
```

**SSO sessions not showing:**
```bash
# Refresh session status
aws-profile-manager sessions --refresh

# Check if SSO cache exists
ls -la ~/.aws/sso/cache/

# Log in again
aws sso login --profile commercial-dev-Developer
```

**Filters not working:**
```bash
# Use verbose mode to see what's being filtered
aws-profile-manager install --dry-run --verbose \
  --organizations=my-org \
  --roles=Developer
```

### Getting Help

For specific command help:
```bash
aws-profile-manager <command> --help
```

For issues and feature requests, visit the [GitHub repository](https://github.com/jpSimkins/aws-profile-manager).
