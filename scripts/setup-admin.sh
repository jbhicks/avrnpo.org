#!/bin/bash

# Production Admin Setup Script
# This script creates the initial admin user for production deployment

set -e

echo "🔧 Setting up admin user for production..."

# Check if required environment variables are set
if [ -z "$ADMIN_EMAIL" ] || [ -z "$ADMIN_PASSWORD" ]; then
    echo "❌ Error: ADMIN_EMAIL and ADMIN_PASSWORD environment variables must be set"
    echo "   Set these in your Coolify environment variables:"
    echo "   ADMIN_EMAIL=your-admin@avrnpo.org"
    echo "   ADMIN_PASSWORD=your-secure-password"
    exit 1
fi

echo "✅ Environment variables found"
echo "   Admin Email: $ADMIN_EMAIL"
echo "   Admin Name: ${ADMIN_FIRST_NAME:-Admin} ${ADMIN_LAST_NAME:-User}"

# Run the admin creation grift
echo "🚀 Creating admin user..."
buffalo task db:create_admin

echo "✅ Admin setup complete!"
echo ""
echo "🔐 Admin Login Credentials:"
echo "   URL: https://avrnpo.org/admin"
echo "   Email: $ADMIN_EMAIL"
echo "   Password: [Set via ADMIN_PASSWORD env var]"
echo ""
echo "⚠️  IMPORTANT: Change the admin password after first login!"