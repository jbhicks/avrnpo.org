# Use official Go image for better version control
FROM golang:1.22-alpine

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

# Run database migrations (will use DATABASE_URL from environment)
RUN soda migrate up

# Set the default command
CMD ./bin/app