# Use official Go image for better version control
FROM golang:1.24-alpine

# Install necessary packages for Buffalo
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Install Buffalo CLI
RUN go install github.com/gobuffalo/cli/cmd/buffalo@latest

# Build the application
RUN buffalo build -o bin/app

# Install soda for migrations
RUN go install github.com/gobuffalo/pop/v6/soda@latest

# Expose port
EXPOSE 3001

# Create startup script
RUN echo '#!/bin/sh' > /app/start.sh && \
    echo 'echo "🚀 Starting AVRNPO application..."' >> /app/start.sh && \
    echo 'echo "📊 Running database migrations..."' >> /app/start.sh && \
    echo 'soda migrate up' >> /app/start.sh && \
    echo 'echo "👤 Creating admin user..."' >> /app/start.sh && \
    echo './bin/app task db:create_admin || echo "⚠️  Admin user creation failed or user already exists"' >> /app/start.sh && \
    echo 'echo "🌐 Starting web server..."' >> /app/start.sh && \
    echo 'exec ./bin/app' >> /app/start.sh && \
    chmod +x /app/start.sh

# Use the startup script as the default command
CMD ["/app/start.sh"]