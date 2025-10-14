#!/bin/sh
set -e

# Run migrations
echo "Running migrations..."
./main migrate up

# Start the application
exec "$@"
