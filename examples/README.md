# 🌸 Caddy SPIRE Examples

This directory contains standardized examples for integrating SPIFFE/SPIRE with Caddy server.

## Examples Overview

### 1. Basic Configuration (`01-basic.caddyfile`)
- **Purpose**: Simple localhost HTTPS server with SPIFFE certificates
- **Features**: 
  - Basic SPIRE integration
  - HTTPS and HTTP endpoints
  - Default socket path
  - 30-second refresh interval

### 2. Medium Configuration (`02-medium.caddyfile`)
- **Purpose**: Multiple sites with different SPIRE settings
- **Features**:
  - Multiple domains with different refresh intervals
  - Reverse proxy configuration
  - Security headers
  - Static file serving
  - Health check endpoints

### 3. Advanced Configuration (`03-advanced.caddyfile`)
- **Purpose**: Production-ready setup with advanced features
- **Features**:
  - Load balancing with health checks
  - Multiple environments (production/staging)
  - Rate limiting
  - Mutual TLS authentication
  - WebSocket support
  - Comprehensive security headers
  - Monitoring and metrics

### 4. Features Showcase (`04-features.json`)
- **Purpose**: Demonstrates all SPIRE module capabilities
- **Features**:
  - JSON configuration format
  - Percentage-based refresh (smart refresh)
  - Interval-based refresh (legacy)
  - Custom socket paths
  - Custom timeout settings
  - All configuration options

### 5. Layer 4 Proxy (`05-layer4.caddyfile`)
- **Purpose**: TCP/UDP proxying with SPIFFE-secured TLS termination
- **Features**:
  - Layer 4 TCP proxy configuration
  - TLS termination with SPIRE certificates
  - Backend proxying after TLS termination
  - Mixed TLS/non-TLS traffic handling
  - Demonstration backend server

### 6. Programmatic Integration (`basic-integration.go`)
- **Purpose**: Shows direct use of the SPIRE client in Go applications
- **Features**:
  - Direct API usage
  - Health checks
  - SPIRE info endpoints
  - Graceful shutdown
  - Error handling

## Quick Start

### Prerequisites
1. SPIRE agent running locally
2. Workload registered in SPIRE
3. Caddy with SPIRE module built

### Running the Examples

#### Basic Example
```bash
# Start with basic configuration
caddy run --config examples/01-basic.caddyfile

# Test the endpoints
curl -k https://localhost:8443
curl http://localhost:8080
```

#### Medium Example
```bash
# Update /etc/hosts first
echo "127.0.0.1 app.example.com api.example.com static.example.com" >> /etc/hosts

# Start Caddy
caddy run --config examples/02-medium.caddyfile

# Test endpoints
curl -k https://app.example.com
curl -k https://api.example.com/health
```

#### Advanced Example
```bash
# Update /etc/hosts for all domains
echo "127.0.0.1 production.example.com staging.example.com internal.example.com metrics.example.com ws.example.com" >> /etc/hosts

# Start Caddy
caddy run --config examples/03-advanced.caddyfile
```

#### Features Showcase (JSON)
```bash
# Run with JSON configuration
caddy run --config examples/04-features.json

# Update /etc/hosts
echo "127.0.0.1 features.example.com percentage-refresh.example.com interval-refresh.example.com custom-socket.example.com timeout-demo.example.com" >> /etc/hosts
```

#### Layer 4 Proxy Example
```bash
# Run with Layer 4 configuration
caddy run --config examples/05-layer4.caddyfile

# Test TLS termination on port 1443
curl -k https://localhost:1443

# Test standard HTTPS on port 8443
curl -k https://localhost:8443

# Test backend directly
curl http://localhost:8080
```

#### Programmatic Integration
```bash
# Build and run the Go example
go run examples/basic-integration.go

# Test endpoints
curl -k https://localhost:8443/
curl -k https://localhost:8443/health
curl -k https://localhost:8443/spire
```

## Configuration Options

### SPIRE Module Settings

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `socket_path` | string | `/tmp/spire-agent/public/api.sock` | Path to SPIRE agent socket |
| `refresh_interval` | duration | - | Fixed refresh interval (mutually exclusive with `refresh_at_percent`) |
| `refresh_at_percent` | int | `65` | Refresh when certificate reaches this percentage of lifetime |
| `refresh_timeout` | duration | `10s` | Timeout for SPIRE operations |

### Caddyfile Syntax
```caddyfile
example.com {
    tls {
        issuer spire {
            socket_path /tmp/spire-agent/public/api.sock
            refresh_at_percent 65
            refresh_timeout 10s
        }
    }
}
```

### JSON Syntax
```json
{
  "subjects": ["example.com"],
  "issuers": [{
    "module": "spire",
    "socket_path": "/tmp/spire-agent/public/api.sock",
    "refresh_at_percent": 65,
    "refresh_timeout": "10s"
  }]
}
```

## SPIRE Setup

### 1. Start SPIRE Agent
```bash
# Example SPIRE agent configuration
sudo spire-agent run -config /etc/spire/agent.conf
```

### 2. Register Workload
```bash
# Register your Caddy workload
spire-server entry create \
    -spiffeID spiffe://example.com/caddy \
    -parentID spiffe://example.com/node \
    -selector unix:user:caddy
```

### 3. Verify Registration
```bash
# Check entries
spire-server entry show

# Test workload API access
/opt/spire/bin/spire-agent api fetch -socketPath /tmp/spire-agent/public/api.sock
```

## Troubleshooting

### Common Issues

1. **"failed to create workload API source"**
   - Check SPIRE agent is running
   - Verify socket path is correct
   - Ensure workload is registered

2. **"timeout waiting for SPIFFE certificate"**
   - Check workload registration
   - Verify attestation is working
   - Increase `refresh_timeout`

3. **Certificate not refreshing**
   - Check refresh settings
   - Verify SPIRE agent connectivity
   - Monitor Caddy logs

### Debug Commands
```bash
# Check SPIRE agent status
sudo systemctl status spire-agent

# Test workload API directly
spire-agent api fetch -socketPath /tmp/spire-agent/public/api.sock

# Check Caddy configuration
caddy validate --config your-config.caddyfile

# Enable debug logging
caddy run --config your-config.caddyfile --adapter caddyfile -v
```

## Security Notes

- Always use HTTPS in production
- Rotate SPIRE root keys regularly
- Monitor certificate expiration
- Use mutual TLS for internal services
- Implement proper access controls

## 🐳 Docker Deployment

For containerized deployment, see [DOCKER.md](./DOCKER.md) which includes:
- Simplified Dockerfile using our custom Caddy SPIRE image
- Docker Compose with proper SPIRE socket mapping
- Container-optimized configurations
- Debugging and troubleshooting guides

## Next Steps

1. Explore the [Caddy SPIRE Module Documentation](../README.md)
2. Try [Docker deployment](./DOCKER.md) for containerized environments
3. Set up [SPIRE in production](https://spiffe.io/docs/latest/deploying/)
4. Configure [workload attestation](https://spiffe.io/docs/latest/spiffe-about/spiffe-concepts/#workload-attestation)
5. Implement [SPIFFE Federation](https://spiffe.io/docs/latest/spiffe-about/federation/)

---

💖 **Happy SPIFFE-ing with Caddy!** ✨
