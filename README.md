# AWS Profile Manager

A powerful CLI and GUI application that simplifies AWS profile management by automatically generating AWS CLI profiles from centralized configuration files.

## What is AWS Profile Manager?

AWS Profile Manager streamlines AWS development workflows by:

- **Centralized Configuration**: Define all your AWS accounts, roles, and environments in a single JSON file
- **Automatic Profile Generation**: Generate hundreds of AWS CLI profiles instantly 
- **Multiple Auth Methods**: Built-in support for AWS SSO, IAM users, and AssumeRole profiles
- **Profile Discovery**: List and filter existing AWS CLI profiles (all authentication types)
- **Session Management**: Monitor AWS SSO session status (active/expired)
- **Cross-Platform**: Works on Linux, macOS, and Windows
- **Dual Interface**: Command-line for automation, GUI for interactive use

## Quick Start

### Installation

Download the latest release for your platform from the [releases page](../../releases).

**Linux/macOS:**
```bash
# Download and make executable
chmod u+x aws-profile-manager
sudo mv aws-profile-manager /usr/local/bin/
```

**Windows:**
```cmd
# Download aws-profile-manager.exe
# Add to your PATH or run from current directory
```

### Basic Usage

```bash
# Generate AWS CLI profiles from configuration
aws-profile-manager install --config aws-config.json

# List AWS SSO sessions and their status
aws-profile-manager sessions

# List and filter AWS CLI profiles  
aws-profile-manager profiles --account 123456789012

# Launch GUI interface
aws-profile-manager gui

# Show help
aws-profile-manager --help
```

## Features

### 🚀 Profile Generation
Generate AWS CLI profiles automatically from a centralized configuration:

```bash
# Generate profiles from your organization's config
aws-profile-manager install --config company-aws-config.json

# Output: Creates profiles like:
# - commercial-prod-Developer
# - commercial-prod-SystemAdmin  
# - govcloud-staging-PowerUser
# - commercial-dev-ReadOnly--us-west-2
```

### 📊 Session Management
Monitor your AWS SSO sessions:

```bash
# Check session status
aws-profile-manager sessions

# Example output:
# 🟢 my-org-commercial (Valid until 2025-10-06 14:30:00)
# 🔴 my-org-govcloud (Expired)
# ⚪ other-org-commercial (No session)

# Verbose details
aws-profile-manager sessions --verbose

# Force refresh session status
aws-profile-manager sessions --refresh
```

### 🔍 Profile Discovery
List and filter existing AWS CLI profiles (supports SSO, IAM, and AssumeRole profiles):

```bash
# List all profiles (all types)
aws-profile-manager profiles

# Filter by account ID
aws-profile-manager profiles --account 123456789012

# Filter by role name
aws-profile-manager profiles --role PowerUser

# Filter by region
aws-profile-manager profiles --region us-east-1

# Filter by SSO session (SSO profiles only)
aws-profile-manager profiles --session my-org-commercial

# Regex pattern matching
aws-profile-manager profiles --pattern "prod.*"

# Verbose output with details (shows profile type)
aws-profile-manager profiles --verbose
```

### 🔄 Configuration Sync
Automatically download AWS configuration from centralized sources:

```bash
# Fetch latest configuration from remote
aws-profile-manager sync fetch

# Check sync status and cache age
aws-profile-manager sync status

# View setup instructions for your organization
aws-profile-manager sync setup

# Clear cached configuration
aws-profile-manager sync clear-cache
```

**Supported Sources:**
- **HTTP/HTTPS**: Public URLs, CDNs, internal web servers
- **S3 Buckets**: Public or authenticated via AWS SSO/IAM
- **Git** (planned): Version-controlled configurations

**Benefits:**
- ✅ Always have the latest AWS accounts and roles
- ✅ Works offline with automatic caching
- ✅ Manual override with `--config` flag when needed
- ✅ See [docs/SYNC-GUIDE.md](docs/SYNC-GUIDE.md) for setup

### 🎨 GUI Interface
Launch the graphical interface for interactive management:

```bash
aws-profile-manager gui --config aws-config.json
```

## Configuration

AWS Profile Manager uses a JSON configuration file to define your AWS environments:

```json
{
  "version": "1.0",
  "environments": {
    "production": {
      "partitions": {
        "commercial": {
          "accounts": [
            {
              "alias": "prod-web",
              "name": "Production Web Services",
              "id": "123456789012"
            }
          ],
          "regions": ["us-east-1", "us-west-2"],
          "default_region": "us-east-1",
          "roles": ["Developer", "SystemAdmin"],
          "sso_start_url": "https://my-org.awsapps.com/start",
          "sso_region": "us-east-1"
        }
      }
    }
  }
}
```

See [Configuration Guide](docs/CONFIGURATION.md) for detailed configuration options.

## Command Reference

### Global Options
- `--config FILE` - Specify configuration file path
- `--help` - Show command help
- `--version` - Show version information

### Commands

#### `install`
Generate AWS CLI profiles from configuration.

```bash
aws-profile-manager install [OPTIONS]

Options:
  --config FILE     Configuration file path
  --environment ENV Filter by environment (e.g., production, staging)
  --partition PART  Filter by partition (commercial, govcloud)
  --dry-run        Show what would be generated without writing
```

#### `sessions`
List AWS SSO sessions and their status.

```bash
aws-profile-manager sessions [OPTIONS]

Options:
  --verbose        Show detailed session information
  --refresh        Force refresh session status
```

#### `profiles`
List and filter existing AWS CLI profiles.

```bash
aws-profile-manager profiles [OPTIONS]

Options:
  --account ID     Filter by AWS account ID
  --role NAME      Filter by IAM role name
  --region NAME    Filter by AWS region
  --session NAME   Filter by SSO session name
  --pattern REGEX  Filter by regex pattern
  --verbose        Show detailed profile information
```

#### `version`
Display version and build information.

```bash
aws-profile-manager version
```

#### `gui`
Launch the graphical user interface.

```bash
aws-profile-manager gui [OPTIONS]

Options:
  --config FILE    Configuration file path
```

## Integration Examples

### CI/CD Pipeline
```bash
# Generate profiles for deployment
aws-profile-manager install --config production-config.json --environment production

# Use generated profile
aws --profile commercial-prod-Developer s3 ls
```

### Development Workflow
```bash
# Check if SSO session is valid
aws-profile-manager sessions | grep "🟢"

# Generate development profiles
aws-profile-manager install --config dev-config.json --environment development

# List available profiles for current project
aws-profile-manager profiles --pattern "dev.*"
```

### Profile Maintenance
```bash
# Audit existing profiles
aws-profile-manager profiles --verbose > profile-audit.txt

# Clean up and regenerate
aws configure list-profiles | grep "old-" | xargs -I {} aws configure set --profile {} region ""
aws-profile-manager install --config updated-config.json
```

## Troubleshooting

### Debug Mode
Enable detailed logging for troubleshooting:

```bash
# Enable debug output
AWS_PROFILE_MANAGER_DEBUG=1 aws-profile-manager install --config config.json

# Or set environment variable
export AWS_PROFILE_MANAGER_DEBUG=1
aws-profile-manager sessions --verbose
```

### Common Issues

**"No SSO session found"**
- Run `aws sso login --sso-session <session-name>` to authenticate
- Check that your SSO session name matches the configuration

**"Permission denied" errors**
- Ensure your AWS CLI configuration directory (~/.aws) is writable
- Check that you have permission to modify AWS CLI config files

**GUI won't launch**
- Ensure you're running on a system with GUI support
- For remote systems, verify X11 forwarding or similar is configured

## Documentation

- **[Configuration Sync Guide](docs/SYNC-GUIDE.md)**: User guide for setting up and using config sync
- **[Sync Reference](docs/SYNC-REFERENCE.md)**: Admin/DevOps guide for sync configuration and deployment
- **[Configuration Guide](docs/CONFIGURATION.md)**: Detailed AWS config file format
- **[Developer Guide](DEVELOPER.md)**: Contributing and development setup
- **[Security](docs/SECURITY.md)**: Security roadmap and future enhancements

## Support

- **Issues**: Report bugs and feature requests on [GitHub Issues](../../issues)
- **Documentation**: See the [docs/](docs/) directory for detailed guides
- **Development**: See [DEVELOPER.md](DEVELOPER.md) for contribution guidelines

## License

[License information here]