#!/bin/sh

# deploy.sh - Production deployment script for Coolify
set -e

echo "🚀 AVRNPO Deployment Starting..."
echo "================================"

# Ensure we're in the right directory
cd /app

echo "📊 Running database migrations..."
soda migrate up

echo "👤 Setting up admin user..."
if [ -n "$ADMIN_EMAIL" ] && [ -n "$ADMIN_PASSWORD" ]; then
    echo "   Creating admin user from environment variables..."
    echo "   Email: $ADMIN_EMAIL"
    ./bin/app task db:create_admin
    echo "✅ Admin user setup completed!"
else
    echo "⚠️  No ADMIN_EMAIL/ADMIN_PASSWORD found."
    echo "   Admin user will need to be created manually."
    echo "   You can promote the first user with:"
    echo "   ./bin/app task db:promote_admin"
fi

echo "🌐 Starting web server..."
echo "   Listening on port ${PORT:-3001}"
echo "   Environment: ${GO_ENV:-production}"
echo "================================"

# Start the application
exec ./bin/app