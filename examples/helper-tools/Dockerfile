# 🌸 Dockerfile for SPIRE Helper Tool
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the helper tool
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o spire-helper-wrapper \
    ./examples/helper-tools/spire-helper-wrapper.go

# Final stage
FROM scratch

# Copy CA certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary
COPY --from=builder /app/spire-helper-wrapper /spire-helper-wrapper

# Set timezone
ENV TZ=UTC

# Create non-root user (using numeric IDs for scratch image)
USER 65534:65534

# Default socket path
ENV SPIRE_SOCKET_PATH=/tmp/spire-agent/public/api.sock

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["/spire-helper-wrapper", "-once", "-verbose"]

# Default command
ENTRYPOINT ["/spire-helper-wrapper"]
CMD ["-daemon", "-verbose"]
