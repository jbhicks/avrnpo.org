#!/bin/bash

# create-admin.sh
# Script to create an admin user during deployment

set -e  # Exit on any error

echo "🔧 Creating Admin User for AVRNPO"
echo "================================="

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    echo "❌ Error: Please run this script from the project root directory"
    exit 1
fi

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check for required tools
if ! command_exists buffalo; then
    echo "❌ Error: Buffalo CLI not found. Please install Buffalo first."
    exit 1
fi

# Set environment if not already set
if [ -z "$GO_ENV" ]; then
    export GO_ENV="production"
    echo "ℹ️  GO_ENV not set, using 'production'"
fi

echo "🔍 Environment: $GO_ENV"

# Method 1: Use environment variables (recommended for production)
if [ -n "$ADMIN_EMAIL" ] && [ -n "$ADMIN_PASSWORD" ]; then
    echo "📧 Creating admin user from environment variables..."
    echo "   Email: $ADMIN_EMAIL"
    buffalo task db:create_admin
    echo "✅ Admin user created successfully!"
    exit 0
fi

# Method 2: Interactive mode (for development/manual setup)
echo "📝 No ADMIN_EMAIL/ADMIN_PASSWORD found in environment."
echo "   Would you like to create an admin user interactively? (y/n)"
read -r response

if [ "$response" = "y" ] || [ "$response" = "Y" ]; then
    buffalo task db:create_admin_interactive
    echo "✅ Admin user created successfully!"
else
    echo "ℹ️  Skipping admin user creation."
    echo ""
    echo "To create an admin user later, you can:"
    echo "  1. Set environment variables and run:"
    echo "     export ADMIN_EMAIL=your-email@example.com"
    echo "     export ADMIN_PASSWORD=your-secure-password"
    echo "     buffalo task db:create_admin"
    echo ""
    echo "  2. Or run interactively:"
    echo "     buffalo task db:create_admin_interactive"
    echo ""
    echo "  3. Or promote the first user who signs up:"
    echo "     buffalo task db:promote_admin"
fi