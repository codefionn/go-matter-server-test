# Multi-stage build for go-matter-server
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the applications
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o matter-server ./cmd/matter-server
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o example-client ./cmd/example-client

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy built binaries
COPY --from=builder /app/matter-server .
COPY --from=builder /app/example-client .

# Copy configuration files
COPY --from=builder /app/config.example.yaml ./config.yaml
COPY --from=builder /app/.env.example ./.env

# Create data directory
RUN mkdir -p /app/data

# Expose WebSocket port
EXPOSE 5580

# Default command runs the matter-server
CMD ["./matter-server", "--storage-path", "/app/data", "--config", "/app/config.yaml"]