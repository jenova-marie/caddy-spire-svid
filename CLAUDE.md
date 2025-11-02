# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Caddy SPIRE Client is a comprehensive Go library and CLI tool for integrating SPIFFE/SPIRE workload API with Caddy server's TLS configuration. The project provides 5 different integration methods ranging from native Caddy modules to external certificate providers.

## Core Architecture

### Main Components

- **`pkg/spire/client.go`** - Core SPIRE client library providing thread-safe certificate management with automatic rotation
- **`pkg/caddyspire/module.go`** - Caddy TLS automation module implementing the GetCertificate interface
- **`cmd/caddy-spire-svid/`** - Standalone CLI tool for SPIRE certificate management
- **`cmd/caddy-with-spire/`** - Custom Caddy build with SPIRE module included
- **`cmd/caddy-spire-file-writer/`** - Helper tool for file-based certificate management (Layer 4 integration)

### Key Design Patterns

1. **Lazy Initialization**: SPIRE connections are established on first certificate request to avoid blocking Caddy startup
2. **Smart Refresh**: Certificates refresh at configurable percentage of lifetime (default 65%) rather than fixed intervals
3. **Multi-attestation Support**: Automatic SVID selection based on DNS names or explicit SPIFFE ID specification
4. **Thread-safe Operations**: All certificate operations use proper synchronization for concurrent access
5. **File-based Integration**: Atomic certificate file writes for Layer 4 and legacy system integration

## Build System & Commands

### Primary Development Commands

- `make build` - Build all main project binaries (caddy-spire-svid, caddy-with-spire, caddy-spire-file-writer)
- `make install` - Build and install binaries to $GOPATH/bin
- `make release` - Build multi-platform release artifacts with proper naming

### Testing Framework (5-Level Progressive Testing)

- `make test` - Quick CI-safe tests (Level 1 + 2 unit tests only)
- `make test-all` - Complete testing suite (all 5 levels, requires SPIRE agent + Docker)
- `make test-level1` - Mock-based unit tests (fastest, no dependencies)
- `make test-level2` - Real SPIRE agent integration tests (requires local spire-agent)
- `make test-level3` - Comprehensive Caddy + SPIRE integration testing
- `make test-level4` - Container SPIRE socket connectivity diagnostic
- `make test-level5` - Containerized comprehensive testing

### Docker Multi-Architecture

- `make docker-setup` - Configure Docker buildx for multi-arch builds
- `make docker-build` - Build multi-arch image (linux/amd64,linux/arm64,linux/arm/v7)
- `make docker-push` - Build and push to GitHub Container Registry

### Code Quality

- Linting: `golangci-lint run` (configuration in `.golangci.yml`)
- Coverage: `make coverage` (generates `coverage.html`)
- Go version: 1.24.2 (specified in go.mod and .golangci.yml)

## Configuration Options

### SPIRE Client Configuration

```go
type Config struct {
    SocketPath       string        // Default: "/tmp/spire-agent/public/api.sock"
    RefreshInterval  time.Duration // Fixed interval refresh (mutually exclusive with RefreshAtPercent)
    RefreshAtPercent int          // Percentage-based refresh (default: 65%, mutually exclusive with RefreshInterval)
    CertFile         string       // Path for certificate chain output (Layer 4 integration)
    KeyFile          string       // Path for private key output (Layer 4 integration)
    FileMode         os.FileMode  // File permissions (default: 0600)
    RefreshTimeout   time.Duration // Timeout for certificate operations (default: 10s)
}
```

### Caddy Module Configuration

**Caddyfile format:**
```
example.com {
    tls {
        issuer spire {
            socket_path /tmp/spire-agent/public/api.sock
            refresh_at_percent 65
            spiffe_id spiffe://domain/workload  # Optional: explicit SPIFFE ID selection
        }
    }
}
```

**JSON format:**
```json
{
    "apps": {
        "tls": {
            "automation": {
                "policies": [{
                    "issuers": [{
                        "module": "spire",
                        "socket_path": "/tmp/spire-agent/public/api.sock",
                        "refresh_at_percent": 65,
                        "spiffe_id": "spiffe://domain/workload"
                    }]
                }]
            }
        }
    }
}
```

## Integration Methods

1. **Caddy Module** (Recommended) - Native TLS automation via `pkg/caddyspire/module.go`
2. **External Certificate Provider** - Separate service writing certificates for standard Caddy
3. **Sidecar Proxy Pattern** - SPIFFE-aware proxy alongside Caddy  
4. **Application Integration** - Direct SPIRE client usage in application code
5. **Helper Tools** - File-based certificate management for legacy systems

## Testing Dependencies

### Level 1 Tests (CI-Safe)
- No external dependencies
- Mock-based testing in `pkg/spire/*_test.go`

### Level 2+ Tests (Local Development)
- **SPIRE Agent**: Must be running locally at configured socket path
- **Docker**: Required for Level 4+ container tests
- **Workload Registration**: SPIRE entries must exist for test workloads

### SPIRE Agent Setup Example
```bash
# Start SPIRE agent (adjust config path as needed)
sudo spire-agent run -config ~/.ssh/spire-agent.conf
```

## Key Files to Understand

- **`pkg/spire/client.go:278`** - `getSVIDWithTimeout()` handles timeouts for all certificate operations
- **`pkg/spire/client.go:460`** - `GetSVIDByID()` enables explicit SPIFFE ID selection
- **`pkg/spire/client.go:485`** - `GetSVIDForServerName()` provides multi-attestation support
- **`pkg/caddyspire/module.go:147`** - `GetCertificate()` implements Caddy TLS automation interface
- **`Makefile`** - Comprehensive build system with 15+ targets for development workflow

## Development Workflow

1. **Start with Level 1 tests**: `make test-level1` for fast feedback during development
2. **Build binaries**: `make build` to create development binaries
3. **Integration testing**: `make test-level2` with local SPIRE agent for real validation
4. **Container testing**: `make test-level4` and `make test-level5` for deployment validation
5. **Code quality**: `golangci-lint run` before committing

## Security Considerations

- All certificate files written with 0600 permissions (owner read/write only)
- Atomic file writes prevent partial certificate exposure
- Timeout protections prevent hanging operations
- No long-lived secrets - certificates auto-rotate based on SPIRE policies
- TLS 1.3 by default with strong cipher suites

## Multi-Platform Support

The project builds for 13 platform combinations:
- Linux: amd64, arm64, arm, 386
- macOS: amd64, arm64  
- Windows: amd64, arm64, arm, 386
- FreeBSD: amd64, arm64, 386

Docker images support: linux/amd64, linux/arm64, linux/arm/v7

## Performance Characteristics

- Certificate operations complete in <10s (configurable timeout)
- HTTPS response times typically <3ms with SPIRE certificates
- Thread-safe operations support high concurrency
- Smart refresh reduces unnecessary certificate fetches

## Error Handling Patterns

- Timeout-protected operations with context cancellation
- Graceful degradation when SPIRE agent unavailable
- Comprehensive error logging with structured fields
- Retry logic for transient failures during certificate refresh