#!/bin/bash
# Run Buffalo tests without problematic schema dumping

# Clean up any auto-generated SQL files
rm -f db/schema.sql migrations/schema.sql

# Set environment to prevent schema dumping
export SKIP_SCHEMA_DUMP=true

# Run migrations manually for test database
GO_ENV=test soda reset

# Remove any new schema files
rm -f db/schema.sql migrations/schema.sql

# Run tests with Go directly
echo "Running Buffalo tests..."
go test -p 1 -tags development ./actions ./models
