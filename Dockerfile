# Build stage
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 go build -o avrnpo .

# Runtime stage
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/avrnpo /app/avrnpo

# Copy migrations (required for schema setup)
COPY --from=builder /app/migrations /app/migrations

# Copy public assets (CSS, JS, images)
COPY --from=builder /app/pb_public /app/pb_public

# Create data directory for PocketBase database
RUN mkdir -p /app/pb_data

# Expose PocketBase port
EXPOSE 8090

# Volume for persistent data (mount externally in production)
VOLUME ["/app/pb_data"]

# Start PocketBase server
CMD ["./avrnpo", "serve", "--http=0.0.0.0:8090"]
