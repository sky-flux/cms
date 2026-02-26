#!/bin/bash
# Production Deployment Setup Script
# This script creates the required secrets directory and password files

set -e

echo "🔧 Sky Flux CMS - Production Environment Setup"
echo "=============================================="
echo ""

# Check if secrets directory exists
if [ -d "secrets" ]; then
    echo "⚠️  Secrets directory already exists"
    read -p "Do you want to overwrite existing secrets? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "❌ Aborted"
        exit 1
    fi
    echo "🗑️  Backing up existing secrets to secrets.backup..."
    mv secrets secrets.backup
fi

# Create secrets directory
mkdir -p secrets
chmod 700 secrets

echo "📝 Creating secret files..."
echo "   IMPORTANT: Use strong, unique passwords for production!"
echo ""

# Function to generate random password
generate_password() {
    openssl rand -base64 32 | tr -d "=+/" | cut -c1-32
}

# Database password
echo "1/8 Database password"
read -p "Generate random password? (Y/n): " -n 1 -r
echo
if [[ $REPLY =~ ^[Nn]$ ]]; then
    read -sp "Enter database password: " DB_PASS
    echo
else
    DB_PASS=$(generate_password)
    echo "✅ Generated: $DB_PASS"
fi
echo "$DB_PASS" > secrets/db_password
chmod 600 secrets/db_password

# Redis password
echo ""
echo "2/8 Redis password"
read -p "Generate random password? (Y/n): " -n 1 -r
echo
if [[ $REPLY =~ ^[Nn]$ ]]; then
    read -sp "Enter Redis password: " REDIS_PASS
    echo
else
    REDIS_PASS=$(generate_password)
    echo "✅ Generated: $REDIS_PASS"
fi
echo "$REDIS_PASS" > secrets/redis_password
chmod 600 secrets/redis_password

# JWT secret
echo ""
echo "3/8 JWT secret (for token signing)"
read -p "Generate random secret? (Y/n): " -n 1 -r
echo
if [[ $REPLY =~ ^[Nn]$ ]]; then
    read -sp "Enter JWT secret (min 32 chars): " JWT_SECRET
    echo
else
    JWT_SECRET=$(generate_password)
    echo "✅ Generated: $JWT_SECRET"
fi
echo "$JWT_SECRET" > secrets/jwt_secret
chmod 600 secrets/jwt_secret

# TOTP encryption key
echo ""
echo "4/8 TOTP encryption key (for encrypting 2FA secrets)"
read -p "Generate random key? (Y/n): " -n 1 -r
echo
if [[ $REPLY =~ ^[Nn]$ ]]; then
    read -sp "Enter TOTP key (32 chars): " TOTP_KEY
    echo
else
    TOTP_KEY=$(generate_password)
    echo "✅ Generated: $TOTP_KEY"
fi
echo "$TOTP_KEY" > secrets/totp_key
chmod 600 secrets/totp_key

# Meilisearch master key
echo ""
echo "5/8 Meilisearch master key"
read -p "Generate random key? (Y/n): " -n 1 -r
echo
if [[ $REPLY =~ ^[Nn]$ ]]; then
    read -sp "Enter Meilisearch master key: " MEILI_KEY
    echo
else
    MEILI_KEY=$(generate_password)
    echo "✅ Generated: $MEILI_KEY"
fi
echo "$MEILI_KEY" > secrets/meili_master_key
chmod 600 secrets/meili_master_key

# RustFS access key
echo ""
echo "6/8 RustFS S3 access key"
read -p "Generate random key? (Y/n): " -n 1 -r
echo
if [[ $REPLY =~ ^[Nn]$ ]]; then
    read -sp "Enter RustFS access key: " RUSTFS_ACCESS
    echo
else
    RUSTFS_ACCESS=$(generate_password | tr '[:upper:]' '[:lower:]')
    echo "✅ Generated: $RUSTFS_ACCESS"
fi
echo "$RUSTFS_ACCESS" > secrets/rustfs_access_key
chmod 600 secrets/rustfs_access_key

# RustFS secret key
echo ""
echo "7/8 RustFS S3 secret key"
read -p "Generate random key? (Y/n): " -n 1 -r
echo
if [[ $REPLY =~ ^[Nn]$ ]]; then
    read -sp "Enter RustFS secret key: " RUSTFS_SECRET
    echo
else
    RUSTFS_SECRET=$(generate_password)
    echo "✅ Generated: $RUSTFS_SECRET"
fi
echo "$RUSTFS_SECRET" > secrets/rustfs_secret_key
chmod 600 secrets/rustfs_secret_key

# Resend API key
echo ""
echo "8/8 Resend API key (get from https://resend.com/api-keys)"
read -p "Enter your Resend API key (re:xxxxxxxxxxxxxx): " RESEND_KEY
if [ -z "$RESEND_KEY" ]; then
    echo "⚠️  Warning: No Resend API key provided. Email features will not work."
    RESEND_KEY="placeholder"
fi
echo "$RESEND_KEY" > secrets/resend_api_key
chmod 600 secrets/resend_api_key

echo ""
echo "✅ Secrets directory created successfully!"
echo ""
echo "📋 Next steps:"
echo "   1. Review generated passwords in secrets/ directory"
echo "   2. Set your DOMAIN environment variable:"
echo "      export DOMAIN=your-domain.com"
echo "   3. Set FRONTEND_URL:"
echo "      export FRONTEND_URL=https://your-domain.com"
echo "   4. Start production containers:"
echo "      make docker-prod-up"
echo ""
echo "⚠️  IMPORTANT: Store these secrets securely!"
echo "   - Do NOT commit secrets/ to git"
echo "   - Add 'secrets/' to .gitignore (already done)"
echo ""
