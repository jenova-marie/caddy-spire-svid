# 🌸 Caddy SPIRE Client

[![CI Pipeline](https://github.com/jenova-marie/caddy-spire-svid/actions/workflows/ci.yml/badge.svg)](https://github.com/jenova-marie/caddy-spire-svid/actions/workflows/ci.yml)
[![Release Workflow](https://github.com/jenova-marie/caddy-spire-svid/actions/workflows/release.yml/badge.svg)](https://github.com/jenova-marie/caddy-spire-svid/actions/workflows/release.yml)
[![Release](https://img.shields.io/github/v/release/jenova-marie/caddy-spire-svid?sort=semver)](https://github.com/jenova-marie/caddy-spire-svid/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/jenova-marie/caddy-spire-svid)](https://goreportcard.com/report/github.com/jenova-marie/caddy-spire-svid)
[![codecov](https://codecov.io/gh/jenova-marie/caddy-spire-svid/branch/main/graph/badge.svg)](https://codecov.io/gh/jenova-marie/caddy-spire-svid)
[![Go Reference](https://pkg.go.dev/badge/github.com/jenova-marie/caddy-spire-svid.svg)](https://pkg.go.dev/github.com/jenova-marie/caddy-spire-svid)
[![Container](https://img.shields.io/badge/ghcr.io-caddy--spire--client-blue?logo=github)](https://github.com/jenova-marie/caddy-spire-svid/pkgs/container/caddy-spire-svid)
[![License: MIT](https://img.shields.io/badge/License-MIT-pink.svg)](https://opensource.org/licenses/MIT)


> ✨ A sparkly Go library and CLI tool for integrating SPIFFE/SPIRE workload API with Caddy server's TLS configuration.

## 🎯 Overview

The **Caddy SPIRE Client** provides seamless integration between [SPIFFE/SPIRE](https://spiffe.io/) identity framework and [Caddy](https://caddyserver.com/) web server with Layer 4 (TCP/UDP) proxy capabilities. It automatically fetches, refreshes, and manages SPIFFE X.509 SVIDs (SPIFFE Verifiable Identity Documents) for use as TLS certificates in both HTTP/HTTPS and Layer 4 proxy scenarios.

### ✨ Key Features

- 🔄 **Automatic Certificate Rotation**: Continuously refreshes SPIFFE certificates before expiry
- 🔒 **Zero-Trust Security**: Leverages SPIFFE's cryptographic identity for service authentication
- 🚀 **High Performance**: Thread-safe, efficient certificate management with minimal overhead
- 🌐 **Layer 4 Proxy**: TCP/UDP proxying with SPIFFE-secured TLS termination
- 🛠️ **Easy Integration**: Simple Go API and CLI tool for quick setup
- 🐳 **Container Ready**: Docker support with multi-arch builds
- 💖 **Production Ready**: Comprehensive testing, logging, and error handling

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   SPIRE Agent   │◄──►│ Caddy SPIRE     │◄──►│   Caddy Server  │
│                 │    │    Client       │    │                 │
│ • Issues SVIDs  │    │ • Fetches certs │    │ • Serves HTTPS  │
│ • Validates IDs │    │ • Auto-refresh  │    │ • Uses SVIDs    │
│ • Workload API  │    │ • TLS Config    │    │ • Zero-trust    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 📦 Installation

### Pre-built Binaries

Download the latest release for your platform:

```bash
# Linux (amd64)
curl -L https://github.com/jenova-marie/caddy-spire-svid/releases/latest/download/caddy-spire-svid-linux-amd64.tar.gz | tar xz

# macOS (arm64)
curl -L https://github.com/jenova-marie/caddy-spire-svid/releases/latest/download/caddy-spire-svid-darwin-arm64.tar.gz | tar xz

# Install globally
sudo mv caddy-spire-svid /usr/local/bin/
```

### From Source

```bash
# Clone the repository
git clone https://github.com/jenova-marie/caddy-spire-svid.git
cd caddy-spire-svid

# Build and install
make install

# Or just build
make build
```

### Docker

#### CLI Tool
```bash
# Pull from GitHub Container Registry
docker pull ghcr.io/jenova-marie/caddy-spire-svid:latest

# Run the container
docker run --rm -v /tmp/spire-agent/public:/tmp/spire-agent/public \
  ghcr.io/jenova-marie/caddy-spire-svid:latest
```

#### Custom Caddy with SPIRE Module
```bash
# Pull the custom Caddy image
docker pull ghcr.io/jenova-marie/caddy-spire-svid-caddy:latest

# Run custom Caddy
docker run --rm -p 80:80 -p 443:443 -p 8443:8443 \
  -v /tmp/spire-agent/public:/tmp/spire-agent/public \
  ghcr.io/jenova-marie/caddy-spire-svid-caddy:latest

# Or build locally and use Docker Compose
make docker-build-caddy
docker-compose -f docker-compose.caddy.yml up --build
```

### Go Module

```bash
go get github.com/jenova-marie/caddy-spire-svid
```

## 🚀 Quick Start

### CLI Usage

```bash
# Start with default settings
caddy-spire-svid

# Custom socket path and refresh interval
caddy-spire-svid -socket /custom/path/api.sock -refresh 1m

# Show help
caddy-spire-svid -help

# Show version
caddy-spire-svid -version
```

### Programmatic Usage

```go
package main

import (
    "log"
    "time"
    
    "github.com/jenova-marie/caddy-spire-svid/pkg/spire"
)

func main() {
    // Create SPIRE client
    config := spire.Config{
        SocketPath:      "/tmp/spire-agent/public/api.sock",
        RefreshInterval: 30 * time.Second,
    }
    
    client, err := spire.NewClient(config)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    // Get TLS config for use with HTTP servers
    tlsConfig := client.GetTLSConfig()
    
    // Use tlsConfig with your HTTP server
    // server.TLSConfig = tlsConfig
}
```

## 🛠️ Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SPIRE_SOCKET_PATH` | `/tmp/spire-agent/public/api.sock` | Path to SPIRE agent socket |
| `SPIRE_REFRESH_INTERVAL` | `30s` | Certificate refresh interval |

### CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-socket` | `/tmp/spire-agent/public/api.sock` | SPIRE agent socket path |
| `-refresh` | `30s` | SVID refresh interval |
| `-version` | - | Show version information |
| `-help` | - | Show help message |

## 🔧 Caddy Integration Methods

This project offers multiple ways to integrate SPIFFE/SPIRE with Caddy, each with distinct advantages and use cases:

### 🎯 **Method 1: Caddy Module/Plugin** ⭐ *Recommended*

**Description**: Create a first-class Caddy module that integrates directly with Caddy's TLS automation system as a certificate issuer.

**✅ Pros:**
- **Native Integration**: Works seamlessly with Caddy's built-in TLS automation
- **Zero-Code Changes**: No application code changes required
- **Production Ready**: Leverages Caddy's robust certificate management
- **Automatic Renewal**: Certificates rotate automatically with SPIRE
- **Configuration-Driven**: Simple Caddyfile or JSON configuration
- **Performance**: No additional proxies or overhead

**❌ Cons:**
- **Custom Build Required**: Need to build Caddy with the module included
- **Go Knowledge**: Module development requires Go programming skills
- **Maintenance**: Must keep module updated with Caddy releases

**Use Cases**: Production deployments, long-term projects, teams comfortable with custom Caddy builds

**📚 Implementation**: See [`examples/caddy-module-integration.md`](examples/caddy-module-integration.md) for detailed integration guide

**Configuration Example**:
```caddyfile
example.com {
    tls {
        issuer spire {
            socket_path /tmp/spire-agent/public/api.sock
            refresh_interval 30s
        }
    }
    respond "Secured with SPIFFE! 🌸"
}
```

### 🎯 **Method 2: External Certificate Provider**

**Description**: Run `caddy-spire-svid` as a separate service that provides certificates to Caddy via shared storage or API.

**✅ Pros:**
- **Standard Caddy**: Uses unmodified Caddy binary
- **Service Separation**: SPIRE logic isolated from web server
- **Flexibility**: Can serve multiple Caddy instances
- **No Rebuilds**: Easy to update independently

**❌ Cons:**
- **Additional Complexity**: Requires certificate sharing mechanism
- **Storage Management**: Need to manage certificate files/storage
- **Synchronization**: Manual coordination between services
- **Security**: File-based certificate sharing has security implications

**Use Cases**: Environments where custom Caddy builds aren't feasible, legacy deployments

### 🎯 **Method 3: Sidecar Proxy Pattern**

**Description**: Use SPIFFE-aware proxy (like Envoy) alongside Caddy to handle mTLS and SPIRE integration.

**✅ Pros:**
- **Zero Application Changes**: Caddy remains completely unmodified
- **Proven Pattern**: Well-established in service mesh architectures
- **Rich Features**: Proxies offer extensive traffic management
- **Ecosystem Integration**: Works with Istio, Consul Connect, etc.

**❌ Cons:**
- **Resource Overhead**: Additional proxy consumes memory/CPU
- **Operational Complexity**: More components to manage and monitor
- **Latency**: Additional network hop may impact performance
- **Learning Curve**: Requires proxy configuration knowledge

**Use Cases**: Service mesh environments, microservices architectures, teams already using proxies

### 🎯 **Method 4: Application-Level Integration**

**Description**: Directly integrate SPIRE client into your application code alongside Caddy server.

**✅ Pros:**
- **Full Control**: Complete programmatic control over SPIRE integration
- **Custom Logic**: Can implement application-specific certificate handling
- **Direct API Access**: Native access to SPIRE workload API
- **Flexibility**: Unlimited customization possibilities

**❌ Cons:**
- **Code Changes Required**: Must modify application to integrate SPIRE
- **Development Effort**: Significant coding and testing required
- **Maintenance Burden**: Need to handle SPIRE updates and edge cases
- **Complexity**: More complex than configuration-based approaches

**Use Cases**: Applications requiring custom certificate logic, embedded systems, specialized requirements

### 🎯 **Method 5: Helper/Wrapper Tools** ✨ *NEW!*

**Description**: Use our custom SPIRE Helper tool to fetch certificates and write them to files for use with standard Caddy.

**✅ Pros:**
- **Simple Setup**: Minimal configuration required, works with any application
- **Production Ready**: Daemon mode, systemd service, Docker support
- **Legacy Friendly**: Works with existing certificate-based workflows
- **Atomic Writes**: Safe certificate updates with proper file permissions
- **Signal Handling**: SIGHUP for immediate refresh, graceful shutdown
- **Standard Caddy**: No custom builds required

**❌ Cons:**
- **File-based Security**: Certificates stored on disk (but with proper permissions)
- **Additional Process**: One more component to manage
- **Limited Automation**: Less integrated than native approaches

**Use Cases**: Legacy systems, standard Caddy deployments, environments requiring file-based certificates

**📚 Implementation**: See [`examples/helper-tools/README.md`](examples/helper-tools/README.md) for complete setup guide and examples

## 🚀 **Integration Examples**

### Standard HTTP Server Integration

```go
package main

import (
    "net/http"
    "github.com/jenova-marie/caddy-spire-svid/pkg/spire"
)

func main() {
    client, err := spire.NewClient(spire.Config{})
    if err != nil {
        panic(err)
    }
    defer client.Close()

    server := &http.Server{
        Addr:      ":8443",
        TLSConfig: client.GetTLSConfig(),
        Handler:   http.DefaultServeMux,
    }

    server.ListenAndServeTLS("", "")
}
```

### Caddy Module Implementation

See [`examples/caddy-module-integration.md`](examples/caddy-module-integration.md) for detailed instructions on creating a production-ready Caddy module.

### Docker Compose Setup

See [`examples/docker-compose.yml`](examples/docker-compose.yml) for a complete setup with SPIRE server and agent.

## 🎯 **Recommendation Matrix**

| Use Case | Recommended Method | Alternative |
|----------|-------------------|-------------|
| **Production Web Server** | Caddy Module | External Provider |
| **Microservices/K8s** | Sidecar Proxy | Caddy Module |
| **Development/Testing** | Helper Tools | Application Integration |
| **Legacy Systems** | External Provider | Helper Tools |
| **Custom Applications** | Application Integration | Caddy Module |
| **Service Mesh** | Sidecar Proxy | External Provider |

Choose the method that best fits your deployment model, security requirements, and operational constraints.

## 🌈 Platform Support

This project provides **comprehensive multi-architecture support** for all major platforms:

### 📱 **Supported Platforms**
| OS | amd64 | arm64 | arm | 386 | Total |
|----|-------|-------|-----|-----|-------|
| **Linux** | ✅ | ✅ | ✅ | ✅ | 4 |
| **macOS** | ✅ | ✅ | - | - | 2 |
| **Windows** | ✅ | ✅ | ✅ | ✅ | 4 |
| **FreeBSD** | ✅ | ✅ | - | ✅ | 3 |

**Total: 13 platforms × 4 binaries = 52 release artifacts per version!**

### 🐳 **Docker Multi-Arch**
- **linux/amd64** - Intel/AMD 64-bit
- **linux/arm64** - ARM 64-bit (Apple Silicon, AWS Graviton)  
- **linux/arm/v7** - 32-bit ARM (Raspberry Pi, embedded)

### 🚀 **Quick Platform Build**
