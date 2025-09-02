# Use official Buffalo Docker image
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

# Run database migrations (will use DATABASE_URL from environment)
RUN soda migrate up

# Set the default command
CMD ./bin/app