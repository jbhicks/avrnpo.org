#!/bin/bash

# Wait for PostgreSQL to be ready
echo "Checking PostgreSQL readiness..."

# Maximum wait time in seconds
MAX_WAIT=30
WAIT_COUNT=0

while [ $WAIT_COUNT -lt $MAX_WAIT ]; do
     # Check if container is running and healthy
     if command -v docker-compose >/dev/null 2>&1; then
         if docker exec my_go_saas_template_postgres pg_isready -U postgres > /dev/null 2>&1; then
             echo "PostgreSQL is ready!"
             exit 0
         fi
     elif command -v podman-compose >/dev/null 2>&1; then
         if podman-compose exec postgres pg_isready -U postgres > /dev/null 2>&1; then
             echo "PostgreSQL is ready!"
             exit 0
         fi
     fi
    echo "Waiting for PostgreSQL... ($((WAIT_COUNT + 1))/$MAX_WAIT)"
    sleep 1
    WAIT_COUNT=$((WAIT_COUNT + 1))
done

echo "ERROR: PostgreSQL failed to become ready within $MAX_WAIT seconds"
echo "Container status:"
if command -v docker-compose >/dev/null 2>&1; then
    docker-compose ps
    echo "Container logs:"
    docker-compose logs postgres 2>/dev/null || echo "Unable to fetch logs"
elif command -v podman-compose >/dev/null 2>&1; then
    podman-compose ps
    echo "Container logs:"
    podman-compose logs postgres --tail 20
fi
exit 1
