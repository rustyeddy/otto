# Stage 1: Build
FROM golang:1.23.3-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN make build

# Stage 2: Runtime
FROM alpine:latest

# Install runtime dependencies (ca-certificates for MQTT TLS)
RUN apk add --no-cache ca-certificates

# Create non-root user
RUN addgroup -g 1000 otto && \
    adduser -D -u 1000 -G otto otto

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/otto /app/otto

# Create data directory for otto
RUN mkdir -p /app/data && \
    chown -R otto:otto /app

# Switch to non-root user
USER otto

# Expose ports
# 8011 - Web UI and REST API
# 1883 - MQTT (if running internal broker)
EXPOSE 8011 1883

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD /app/otto version || exit 1

# Default command
CMD ["/app/otto", "serve"]
