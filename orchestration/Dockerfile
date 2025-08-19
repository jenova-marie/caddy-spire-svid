# SPIRE Docker Diagnostic Tool
# Minimal container to test SPIRE agent connectivity
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY main.go .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o spire-diagnostic .

# Final minimal image
FROM alpine:latest

# Install ca-certificates for TLS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary
COPY --from=builder /app/spire-diagnostic .

# Set default environment variables
ENV SPIRE_SOCKET_PATH=/tmp/spire-agent/public/api.sock

# Run the diagnostic
CMD ["./spire-diagnostic"]
