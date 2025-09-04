# Use Buffalo image with Go 1.22 support
FROM gobuffalo/buffalo:latest

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN buffalo build -o bin/app

# Install soda for migrations
RUN go install github.com/gobuffalo/pop/v6/soda@latest

# Run database migrations (will use DATABASE_URL from environment)
RUN soda migrate up

# Set the default command
CMD ./bin/app