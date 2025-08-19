# 🌸 Method 5: Helper Tools / File-based Legacy Support

This example demonstrates **Method 5** - using file-based certificate management for legacy systems and standard certificate workflows. This approach is perfect for:

- 🔧 **Legacy Applications** that expect certificates as files
- 📁 **Standard Workflows** using traditional certificate paths  
- 🛠️ **Existing Infrastructure** that can't be modified
- 🔄 **Simple Integration** without code changes

## 🎯 **How It Works**

The SPIRE Helper Tool acts as a bridge between SPIRE's workload API and file-based certificate expectations:

```
┌─────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ SPIRE Agent │───►│ Helper Tool     │───►│ Certificate     │
│             │    │                 │    │ Files           │
│ • Issues    │    │ • Fetches SVIDs │    │ • cert.pem      │
│   SVIDs     │    │ • Writes files  │    │ • key.pem       │
│ • Manages   │    │ • Atomic writes │    │ • bundle.pem    │
│   lifecycle │    │ • File watching │    │                 │
└─────────────┘    └─────────────────┘    └─────────────────┘
                            │                        │
                            │                        ▼
                            │               ┌─────────────────┐
                            │               │ Legacy Apps     │
                            │               │ • Caddy         │
                            └──────────────►│ • Nginx         │
                             Daemon Mode    │ • Apache        │
                             Monitoring     │ • Custom Apps   │
                                           └─────────────────┘
```

## 🚀 **Quick Start**

### **1. Build the Helper Tool**

```bash
# From project root
make build-helper-tools  # (we'll add this to Makefile)

# Or build manually
go build -o bin/spire-helper examples/helper-tools/spire-helper-wrapper.go
```

### **2. One-time Certificate Fetch**

```bash
# Fetch certificates once and exit
./bin/spire-helper -once -verbose

# Custom directory and file names
./bin/spire-helper -once \
    -cert-dir /etc/ssl/myapp \
    -cert-file server.crt \
    -key-file server.key
```

### **3. Daemon Mode**

```bash
# Run as daemon with continuous updates
./bin/spire-helper -daemon -verbose

# Custom configuration
./bin/spire-helper -daemon \
    -cert-dir /etc/ssl/spire \
    -interval 30s \
    -socket /var/run/spire/agent.sock
```

### **4. Use with Caddy**

```bash
# Run Caddy with helper-generated certificates
caddy run --config examples/helper-tools/Caddyfile
```

## 📋 **Command Line Options**

| Flag | Default | Description |
|------|---------|-------------|
| `-cert-dir` | `/etc/ssl/spire` | Directory to write certificates |
| `-cert-file` | `cert.pem` | Certificate filename |
| `-key-file` | `key.pem` | Private key filename |
| `-bundle-file` | `bundle.pem` | Trust bundle filename |
| `-socket` | `/tmp/spire-agent/public/api.sock` | SPIRE agent socket path |
| `-interval` | `30s` | Certificate refresh interval |
| `-daemon` | `false` | Run as daemon (continuous mode) |
| `-once` | `false` | Fetch certificates once and exit |
| `-verbose` | `false` | Enable verbose logging |

## 🐳 **Docker Deployment**

### **Docker Compose**

```bash
# Start complete stack
cd examples/helper-tools
docker-compose up

# Check logs
docker-compose logs spire-helper
docker-compose logs caddy
```

### **Standalone Container**

```bash
# Build helper tool image
docker build -t spire-helper -f examples/helper-tools/Dockerfile .

# Run with volume mounts
docker run -d \
    --name spire-helper \
    -v spire-socket:/tmp/spire-agent/public:ro \
    -v cert-storage:/etc/ssl/spire \
    spire-helper -daemon -verbose
```

## 🖥️ **Systemd Service**

### **Install as System Service**

```bash
# Install binary
sudo cp bin/spire-helper /usr/local/bin/spire-helper-wrapper
sudo chmod +x /usr/local/bin/spire-helper-wrapper

# Install service file
sudo cp examples/helper-tools/systemd-service.conf /etc/systemd/system/spire-helper.service

# Create user and directories
sudo useradd -r -s /bin/false spire-helper
sudo mkdir -p /etc/ssl/spire
sudo chown spire-helper:spire-helper /etc/ssl/spire

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable spire-helper
sudo systemctl start spire-helper

# Check status
sudo systemctl status spire-helper
```

### **Service Management**

```bash
# View logs
sudo journalctl -u spire-helper -f

# Reload certificates immediately
sudo systemctl reload spire-helper

# Restart service
sudo systemctl restart spire-helper
```

## 🔧 **Integration Examples**

### **With Standard Caddy**

```caddyfile
example.com {
    tls /etc/ssl/spire/cert.pem /etc/ssl/spire/key.pem {
        reload_time 10s
    }
    respond "Secured with SPIRE!"
}
```

### **With Nginx**

```nginx
server {
    listen 443 ssl http2;
    server_name example.com;
    
    ssl_certificate /etc/ssl/spire/cert.pem;
    ssl_certificate_key /etc/ssl/spire/key.pem;
    ssl_trusted_certificate /etc/ssl/spire/bundle.pem;
    
    # Auto-reload on certificate change
    ssl_session_timeout 1d;
    ssl_session_cache shared:SSL:50m;
}
```

### **With Apache**

```apache
<VirtualHost *:443>
    ServerName example.com
    
    SSLEngine on
    SSLCertificateFile /etc/ssl/spire/cert.pem
    SSLCertificateKeyFile /etc/ssl/spire/key.pem
    SSLCACertificateFile /etc/ssl/spire/bundle.pem
</VirtualHost>
```

## 🔒 **Security Features**

### **File Permissions**
- **Private keys**: Written with `0600` permissions (owner read/write only)
- **Certificates**: Written with `0644` permissions (world readable)
- **Atomic writes**: Temporary files with atomic rename to prevent partial reads

### **Process Security**
- **Non-root execution**: Runs as dedicated user account
- **Signal handling**: Graceful shutdown on SIGTERM/SIGINT
- **SIGHUP reload**: Immediate certificate refresh without restart

### **Certificate Validation**
- **Expiry monitoring**: Logs time until expiry
- **Chain validation**: Ensures complete certificate chain
- **SPIFFE ID tracking**: Logs identity information

## 🚨 **Troubleshooting**

### **Common Issues**

1. **Permission Denied**
   ```bash
   # Check directory permissions
   ls -la /etc/ssl/spire
   
   # Fix permissions
   sudo chown -R spire-helper:spire-helper /etc/ssl/spire
   ```

2. **SPIRE Agent Connection**
   ```bash
   # Check socket exists
   ls -la /tmp/spire-agent/public/api.sock
   
   # Test SPIRE agent
   spire-agent api fetch -socketPath /tmp/spire-agent/public/api.sock
   ```

3. **Certificate Not Updating**
   ```bash
   # Check helper tool logs
   sudo journalctl -u spire-helper -f
   
   # Manually trigger refresh
   sudo systemctl reload spire-helper
   ```

### **Debugging**

```bash
# Run with verbose logging
./bin/spire-helper -daemon -verbose

# One-time fetch for testing
./bin/spire-helper -once -verbose

# Check certificate contents
openssl x509 -in /etc/ssl/spire/cert.pem -text -noout
```

## ✨ **Advantages of This Method**

- **🔧 Legacy Compatibility**: Works with any application that uses file-based certificates
- **📁 Standard Workflows**: Fits existing certificate management processes
- **🛠️ No Code Changes**: Applications don't need modification
- **🔄 Automatic Rotation**: Certificates update automatically with SPIRE
- **🚀 Simple Deployment**: Easy to integrate into existing infrastructure
- **📊 Monitoring Ready**: Systemd integration with logging and health checks

This method provides the **easiest path** to SPIFFE adoption for existing applications! 🌸
