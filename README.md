# AWS Profile Manager

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![CodeQL](https://github.com/jpSimkins/aws-profile-manager/workflows/Security%20Scans/badge.svg)](https://github.com/jpSimkins/aws-profile-manager/actions/workflows/security.yml)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/jpSimkins/aws-profile-manager/badge)](https://securityscorecards.dev/viewer/?uri=github.com/jpSimkins/aws-profile-manager)
[![Built with AI Assistance](https://img.shields.io/badge/built%20with-AI%20assistance-blue)](https://github.com/features/copilot)
[![Buy Me A Coffee](https://img.shields.io/badge/Buy%20Me%20A%20Coffee-support-orange?logo=buy-me-a-coffee)](https://www.buymeacoffee.com/jpSimkins)

**Stop managing AWS profiles manually. Let your configuration do the work.**

AWS Profile Manager transforms a single JSON file into hundreds of ready-to-use AWS CLI profiles. Whether you're managing 3 accounts or 300, one command sets up your entire AWS environment.

## Why AWS Profile Manager?

**Before:**
```bash
# Manually configure each profile...
aws configure sso --profile prod-developer
# Enter SSO start URL: https://...
# Enter SSO region: us-east-1
# Enter account ID: 123456789012
# Enter role name: Developer
# ...repeat 50+ times for all accounts and roles
```

**After:**
```bash
# One command, all profiles
aws-profile-manager install
# ✓ Generated 247 profiles in 2 seconds
```

### Key Benefits

- ⚡ **Instant Setup** - Generate hundreds of profiles in seconds, not hours
- 🔄 **Always Current** - Pull latest accounts and roles from a central source with auto-sync
- 🎯 **Zero Errors** - No more typos in account IDs or SSO URLs
- 📦 **Team Distribution** - Share one config file instead of setup instructions
- 🔍 **Profile Discovery** - Find and filter existing profiles across all auth types
- 📊 **Session Monitoring** - See which SSO sessions are active or expired
- 🖥️ **Complete GUI** - Full-featured desktop app with all CLI capabilities plus visual workflows

## Quick Start

### 1. Install

**Download from [releases page](../../releases):**

**Linux:**
```bash
# Download and extract
tar -xf aws-profile-manager-linux-amd64.tar.xz
sudo mv aws-profile-manager /usr/local/bin/

# Or use the desktop package
tar -xf aws-profile-manager-linux-amd64.tar.xz
./install.sh  # Adds desktop entry and menu item
```

**macOS:**
```bash
# Download and extract
unzip aws-profile-manager-darwin-amd64.zip

# Move to Applications
mv "AWS Profile Manager.app" /Applications/

# Or use from terminal
chmod u+x aws-profile-manager
sudo mv aws-profile-manager /usr/local/bin/
```

**Windows:**
```bash
# Download and run the installer
aws-profile-manager-windows-amd64.exe

# Or download the CLI binary and add to PATH
```

### 2. Generate Profiles

**Command Line:**
```bash
# Generate all profiles from your organization's config
aws-profile-manager install

# Or start with sync if your org uses it
aws-profile-manager sync fetch
aws-profile-manager install
```

**GUI:**
```bash
# Launch the desktop application
aws-profile-manager gui

# GUI features:
# • Visual profile installation with filters and presets
# • Real-time profile discovery and search
# • SSO session monitoring
# • One-click sync with auto-update on startup
# • Export/import with drag-and-drop
# • Settings management
```

### 3. Start Using AWS CLI

```bash
# List generated profiles
aws-profile-manager profiles

# Use any profile immediately
aws --profile commercial-dev-Developer s3 ls
```

That's it! You now have all your AWS profiles configured and ready to use.

## Core Features

### 🚀 Automatic Profile Generation

Define your AWS environment once in JSON, generate profiles instantly:

```json
{
  "version": "2.0",
  "managed": {
    "organizations": {
      "my-org": {
        "partitions": {
          "commercial": {
            "url": "https://my-org.awsapps.com/start",
            "accounts": [{"alias": "prod", "id": "123456789012"}],
            "roles": ["Developer", "Admin"]
          }
        }
      }
    }
  }
}
```

One `install` command creates every combination of account × role × region.

**Learn more:** [Configuration Guide](docs/ConfigurationGuide.md)

### 🔄 Centralized Sync

Your platform team updates the config once, everyone stays current:

```bash
# Fetch latest from HTTP, S3, or Git
aws-profile-manager sync fetch

# Apply updates
aws-profile-manager install
```

Works offline with smart caching. No more distributing config files manually.

**Learn more:** [Sync Guide](docs/SyncGuide.md)

### 🔍 Profile Discovery

Find what you need across hundreds of profiles:

```bash
# All production profiles
aws-profile-manager profiles --pattern "prod.*"

# Specific account
aws-profile-manager profiles --account-id 123456789012

# Specific role across all accounts
aws-profile-manager profiles --role Developer
```

Supports SSO, IAM, and AssumeRole profiles.

### 📊 Session Management

Know which SSO sessions are active before you try to use them:

```bash
aws-profile-manager sessions
```

Output:
```
🟢 my-org-commercial (Valid until 2025-10-06 14:30:00)
🔴 other-org (Expired)
```

### 🖥️ Full-Featured GUI

**Everything the CLI does, visually:**

```bash
aws-profile-manager gui
```

**GUI Features:**
- **Visual Profile Installation** - Filter by organization, partition, account, role with visual feedback
- **Preset Support** - One-click installation profiles (Developer, Admin, All Regions, etc.)
- **Live Profile Discovery** - Search and filter hundreds of profiles with real-time updates
- **Session Monitoring** - See SSO session status at a glance (active/expired)
- **Smart Sync** - One-click sync with optional auto-update on startup
- **Export/Import** - Full backup/restore with visual progress
- **Settings Management** - Configure sync sources, themes, and preferences

The GUI is not a simplified version—it provides the full power of the CLI with visual controls, progress indicators, and interactive workflows.

## Documentation

- **[CLI Command Reference](docs/CliGuide.md)** - Complete command documentation with examples
- **[Configuration Guide](docs/ConfigurationGuide.md)** - JSON schema reference and patterns
- **[Sync Guide](docs/SyncGuide.md)** - Set up centralized configuration distribution
- **[Developer Guide](DEVELOPER.md)** - Development setup and workflow
- **[Contributing Guide](CONTRIBUTING.md)** - How to contribute to the project

## Common Workflows

**New team member onboarding (CLI):**
```bash
aws-profile-manager sync setup  # Show setup instructions
aws-profile-manager sync fetch  # Download config
aws-profile-manager install     # Generate profiles
```

**New team member onboarding (GUI):**
1. Install desktop app from releases page
2. Launch app → Enable auto-sync in Settings
3. Click "Sync Now" → Select preset → Click "Install"
4. Done! Profiles are ready, sessions are visible

**Daily development (CLI):**
```bash
aws-profile-manager sessions    # Check session status
aws sso login --profile commercial-dev-Developer  # If needed
```

**Daily development (GUI):**
1. Launch app (auto-syncs on startup if enabled)
2. Check Sessions tab for active SSO sessions
3. Use Profiles tab to find and copy profile names
4. Install tab shows current filters and allows quick reinstall

**Platform team update:**
```bash
# Edit config.json (add new account or role)
# Upload to sync source (S3, HTTP server, etc.)
# Team members run: aws-profile-manager sync fetch && install
```

See the [CLI Guide](docs/CliGuide.md) for detailed examples and integration patterns.

## Platform Support

- ✅ **Linux** - Full CLI and GUI support with desktop package (.tar.xz with installer)
- ✅ **macOS** - Full CLI and GUI support with native .app bundle
- ✅ **Windows** - Full CLI and GUI support with native installer (.exe)

**Desktop Integration:**
- Application menu entries
- File associations for config backups
- Native installers for each platform
- Automatic updates support (planned)

## Getting Help

```bash
# Command help
aws-profile-manager --help
aws-profile-manager install --help

# Debug mode
AWS_PROFILE_MANAGER_DEBUG=1 aws-profile-manager install
```

For issues and feature requests, visit the [GitHub repository](https://github.com/jpSimkins/aws-profile-manager).

## Contributing

We welcome contributions from the community! Whether you're fixing bugs, adding features, improving documentation, or sharing ideas, your help is appreciated.

- **[Contributing Guide](CONTRIBUTING.md)** - Guidelines for contributing code, docs, and ideas
- **[Contributors](CONTRIBUTORS.md)** - Recognition of our contributors
- **[Developer Guide](DEVELOPER.md)** - Complete development setup and workflow

Before contributing, please review our coding standards in [`.github/copilot-instructions.md`](.github/copilot-instructions.md).

## Support This Project

If you find AWS Profile Manager useful, consider supporting its development:

[![Buy Me A Coffee](https://img.shields.io/badge/Buy%20Me%20A%20Coffee-support-orange?logo=buy-me-a-coffee&style=for-the-badge)](https://www.buymeacoffee.com/jpSimkins)

Your support helps maintain and improve this tool for the entire community. Thank you! ☕

## License

AWS Profile Manager is open source software licensed under the [MIT License](LICENSE).

Copyright © 2026 Jeremy Simkins (jpSimkins)

### Free Forever

The software is free for everyone - individuals, startups, and enterprises. Use it, modify it, distribute it.

### Future Services (Optional)

While the software will always be free, we may offer optional paid services in the future for teams who need them:
- Hosted configuration sync service
- Priority support and consulting
- Custom integrations and training

**These are optional services, not software licenses.** The tool itself will always remain free and open source under MIT.