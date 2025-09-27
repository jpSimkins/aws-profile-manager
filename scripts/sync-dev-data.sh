#!/bin/bash

# Sync Development Data Script
# Syncs real AWS CLI configuration and cache to development directories
# for realistic testing without affecting production AWS CLI setup
# 
# Run this script whenever your AWS CLI config changes or SSO sessions expire

set -e

echo "� Syncing development data for AWS Profile Manager"
echo ""

# Check if AWS CLI config exists
if [ ! -f "$HOME/.aws/config" ]; then
    echo "❌ No AWS CLI configuration found at ~/.aws/config"
    echo "   Please configure AWS CLI first with 'aws configure' or AWS SSO"
    exit 1
fi

# Create development directories
echo "📁 Creating development directories..."
mkdir -p .dev/aws/sso/cache
mkdir -p .dev/config

# Copy AWS CLI configuration
echo "📋 Copying AWS CLI configuration..."
cp "$HOME/.aws/config" .dev/aws/config
echo "   ✅ Copied ~/.aws/config → .dev/aws/config"

# Copy credentials if they exist
if [ -f "$HOME/.aws/credentials" ]; then
    cp "$HOME/.aws/credentials" .dev/aws/credentials
    echo "   ✅ Copied ~/.aws/credentials → .dev/aws/credentials"
else
    echo "   ℹ️  No ~/.aws/credentials file found (SSO-only setup)"
fi

# Copy SSO cache files if they exist
if [ -d "$HOME/.aws/sso/cache" ] && [ "$(ls -A "$HOME/.aws/sso/cache" 2>/dev/null)" ]; then
    cp "$HOME/.aws/sso/cache"/* .dev/aws/sso/cache/ 2>/dev/null || true
    cache_count=$(ls -1 .dev/aws/sso/cache/ 2>/dev/null | wc -l)
    echo "   ✅ Copied $cache_count SSO cache files → .dev/aws/sso/cache/"
else
    echo "   ℹ️  No SSO cache files found (no active sessions)"
fi

# Show what was copied
echo ""
echo "📊 Development data summary:"
echo "   AWS Config: $(wc -l < .dev/aws/config) lines"
if [ -f ".dev/aws/credentials" ]; then
    echo "   Credentials: $(wc -l < .dev/aws/credentials) lines"
fi
echo "   SSO Cache: $(ls -1 .dev/aws/sso/cache/ 2>/dev/null | wc -l) files"

echo ""
echo "✅ Development data sync complete!"
echo ""
echo "🎯 Benefits:"
echo "   • Test with realistic AWS profiles and accounts"
echo "   • See actual SSO session status (active/expired)"
echo "   • Verify session management functionality"
echo "   • Test profile filtering with real data"
echo ""
echo "🔄 To sync again after AWS CLI changes:"
echo "   ./scripts/sync-dev-data.sh"
echo ""
echo "🚀 Ready to develop! Try:"
echo "   AWS_PROFILE_MANAGER_DEBUG=1 go run src/main.go gui"