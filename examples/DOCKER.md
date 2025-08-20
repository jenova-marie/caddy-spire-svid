# 🐳 Docker Examples for Caddy SPIRE

This directory contains simplified Docker examples showing how to run Caddy with SPIRE integration.

## 🔑 Key Concept: Host Socket Mapping

The most important aspect of running Caddy SPIRE in Docker is properly mapping the host SPIRE agent socket:

```yaml
volumes:
  # 🔑 CRITICAL: Map host SPIRE agent socket into container
  - /tmp/spire-agent/public/api.sock:/tmp/spire-agent/public/api.sock:ro
```

This allows the containerized Caddy to communicate with the SPIRE agent running on the host.

## 📋 Prerequisites

- SPIRE agent running on the host system
- Workload registered in SPIRE for Caddy
- SPIRE socket accessible at `/tmp/spire-agent/public/api.sock`

## 🚀 Quick Start

### 1. Build Caddy SPIRE Image
```bash
# Make sure you have the multi-arch image available
docker pull ghcr.io/jenova-marie/caddy-spire-svid:latest

# Or build locally
make docker-build
```

### 2. Start SPIRE Agent (on host)
```bash
# Make sure SPIRE agent is running on your host
sudo spire-agent run -config /etc/spire/agent.conf

# Verify socket exists
ls -la /tmp/spire-agent/public/api.sock
```

### 3. Run with Docker Compose
```bash
cd examples/
docker-compose up -d

# Check logs
docker-compose logs -f caddy-spire

# Test endpoints
curl http://localhost:8080
curl -k https://localhost:8443
```

### 4. Stop Services
```bash
docker-compose down
```

## 📁 File Structure

```
examples/
├── Dockerfile              # Simplified Caddy SPIRE container
├── docker-compose.yml      # Complete stack with SPIRE agent
├── Caddyfile               # Container-optimized Caddy config
├── DOCKER.md               # This documentation
├── html/                   # Static web content
│   ├── index.html          # Main page
│   └── backend/            # Backend service content
├── sites/                  # Additional site configs (optional)
├── headers/                # Shared header configs (optional)
└── shared/                 # Shared Caddy configs (optional)
```

## 🐳 Container Architecture

```
┌─────────────────────────────────────────────────────────────┐
│ Docker Host                                                 │
│                                                             │
│  ┌─────────────────┐    ┌─────────────────────────────────┐ │
│  │ spire-agent     │    │ caddy-spire (container)         │ │
│  │ (host process)  │    │                                 │ │
│  │                 │    │ • Caddy with SPIRE module       │ │
│  │ • Issues SVIDs  │    │ • Reads from mounted socket     │ │
│  │ • Socket API    │◄───┤ • Serves HTTPS traffic          │ │
│  │                 │    │                                 │ │
│  └─────────────────┘    └─────────────────────────────────┘ │
│           │                            │                    │
│           │                            │                    │
│     ┌─────▼──────┐                ┌────▼────┐               │
│     │ /tmp/spire-│                │ ports   │               │
│     │ agent/     │                │ 80, 443 │               │
│     │ public/    │                │ 8080,   │               │
│     │ api.sock   │                │ 8443    │               │
│     │ (mounted)  │                │ 2019    │               │
│     └────────────┘                └─────────┘               │
└─────────────────────────────────────────────────────────────┘
```

## 🔧 Configuration Details

### Dockerfile Features
- **Base Image**: Uses our custom `ghcr.io/jenova-marie/caddy-spire-svid:latest`
- **Debug Tools**: Includes `bind-tools` for DNS debugging
- **Directory Structure**: Creates standard Caddy directories
- **Health Check**: Built-in HTTP health check endpoint
- **Static Content**: Supports serving static files

### Docker Compose Features
- **Socket Mapping**: Proper SPIRE socket volume mapping
- **Service Dependencies**: Ensures SPIRE agent starts first
- **Health Checks**: Container health monitoring
- **Log Management**: Persistent log volumes
- **Network Isolation**: Dedicated network for SPIRE communication

### Caddyfile Configuration
- **Container-Optimized**: Designed for container deployment
- **Health Endpoint**: Built-in health check at `/health`
- **SPIRE Integration**: Uses socket path `/tmp/spire-agent/public/api.sock`
- **Debug Logging**: Enabled for troubleshooting

## 🌐 Endpoints

| Endpoint | Protocol | Purpose |
|----------|----------|---------|
| `localhost:80` | HTTP | Health check and basic HTTP |
| `localhost:8080` | HTTP | Alternative HTTP port |
| `localhost:8443` | HTTPS | SPIRE-secured HTTPS |
| `localhost:443` | HTTPS | Standard HTTPS port |
| `localhost:2019` | HTTP | Caddy Admin API |

## 🔍 Debugging

### Check Container Status
```bash
# View all containers
docker-compose ps

# Check specific container logs
docker-compose logs caddy-spire
docker-compose logs spire-agent

# Follow logs in real-time
docker-compose logs -f caddy-spire
```

### Test SPIRE Connectivity
```bash
# Enter Caddy container
docker-compose exec caddy-spire sh

# Check socket exists
ls -la /tmp/spire-agent/public/api.sock

# Test with curl
curl http://localhost:80/health
curl -k https://localhost:8443
```

### SPIRE Agent Debugging
```bash
# Enter SPIRE agent container
docker-compose exec spire-agent sh

# Check SPIRE agent status
/opt/spire/bin/spire-agent api fetch -socketPath /tmp/spire-agent/public/api.sock
```

## 🚨 Common Issues

### 1. Socket Permission Denied
```bash
# Check socket permissions in both containers
docker-compose exec spire-agent ls -la /tmp/spire-agent/public/api.sock
docker-compose exec caddy-spire ls -la /tmp/spire-agent/public/api.sock
```

### 2. Container Cannot Connect to SPIRE
```bash
# Verify volume mapping
docker volume inspect examples_spire-socket

# Check network connectivity
docker-compose exec caddy-spire ping spire-agent
```

### 3. Certificates Not Loading
```bash
# Check Caddy configuration
docker-compose exec caddy-spire caddy validate --config /etc/caddy/Caddyfile

# View detailed logs
docker-compose logs caddy-spire | grep -i spire
```

## 🔒 Security Considerations

### Socket Security
- **Read-Only**: SPIRE socket mounted as read-only in Caddy container
- **Volume Isolation**: Dedicated volume for socket sharing
- **Container User**: Consider running containers as non-root users

### Network Security
- **Isolated Network**: Services communicate on dedicated Docker network
- **Port Exposure**: Only necessary ports exposed to host
- **TLS**: All SPIRE communication uses secure channels

### Certificate Management
- **Automatic Rotation**: SPIRE handles certificate lifecycle
- **No File Storage**: Certificates stored in memory, not files
- **Mutual TLS**: Can be configured for service-to-service communication

## 🚀 Production Deployment

### Environment Variables
```bash
# Set in docker-compose.yml or .env file
CADDY_DEBUG=0                    # Disable debug in production
SPIRE_SOCKET_PATH=/tmp/spire-agent/public/api.sock
CADDY_ADMIN_LISTEN=localhost:2019
```

### Resource Limits
```yaml
# Add to docker-compose.yml services
deploy:
  resources:
    limits:
      memory: 512M
      cpus: '0.5'
    reservations:
      memory: 256M
      cpus: '0.25'
```

### Health Monitoring
```bash
# Check health status
docker-compose ps
curl http://localhost:80/health

# Monitor with external tools
# - Prometheus metrics
# - Docker health checks
# - Application monitoring
```

## 📚 Related Documentation

- [Basic Examples](./README.md) - Caddyfile configurations
- [Main Project README](../README.md) - Full project documentation
- [SPIRE Documentation](https://spiffe.io/docs/) - SPIRE setup and configuration
- [Caddy Documentation](https://caddyserver.com/docs/) - Caddy server documentation

---

🌸 **Happy containerized SPIFFE-ing!** ✨
