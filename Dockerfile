# ============================================
# Gate API - Multi-stage Dockerfile
# ============================================

# Stage 1: Build
FROM golang:1.22-alpine AS builder

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set timezone
ENV TZ=Asia/Jakarta

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.Version=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')" \
    -o /app/bin/gate-api \
    ./cmd/api

# Stage 2: Runtime
FROM alpine:3.19

# Install dependencies
RUN apk add --no-cache ca-certificates tzdata curl

# Set timezone
ENV TZ=Asia/Jakarta

# Create non-root user
RUN addgroup -g 1000 gate && \
    adduser -u 1000 -G gate -s /bin/sh -D gate

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/gate-api /app/gate-api

# Copy migrations (for embedded migrations if needed)
COPY --from=builder /app/database/migrations /app/database/migrations

# Create directories
# Note: keys directory should be mounted as volume in production for security
# e.g., docker run -v /path/to/keys:/app/keys ...
RUN mkdir -p /app/uploads /app/keys && chown -R gate:gate /app

# Switch to non-root user
USER gate

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Run the application
CMD ["/app/gate-api"]

