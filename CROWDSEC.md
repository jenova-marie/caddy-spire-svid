# CrowdSec Integration Guide

This guide covers installing CrowdSec, configuring the bouncer, and integrating it with your custom Caddy build.

## Overview

CrowdSec is a collaborative security engine that analyzes logs, detects attacks, and shares threat intelligence. The caddy-crowdsec-bouncer module blocks malicious traffic based on CrowdSec decisions.

**Components included in this build:**
- `http.handlers.crowdsec` - Block IPs at HTTP layer
- `http.handlers.appsec` - WAF-like request inspection
- `layer4.matchers.crowdsec` - Block at TCP/UDP layer

## Prerequisites

- CrowdSec Security Engine installed and running
- Access to the CrowdSec Local API (default: `http://localhost:8080`)
- A registered bouncer API key

## Installing CrowdSec

### macOS (Homebrew)

```bash
brew install crowdsec
brew services start crowdsec
```

### Debian/Ubuntu

```bash
curl -s https://install.crowdsec.net | sudo sh
sudo apt install crowdsec
sudo systemctl enable crowdsec
sudo systemctl start crowdsec
```

### RHEL/CentOS/Fedora

```bash
curl -s https://install.crowdsec.net | sudo sh
sudo dnf install crowdsec
sudo systemctl enable crowdsec
sudo systemctl start crowdsec
```

### Docker

```bash
docker run -d \
  --name crowdsec \
  -v /var/log:/var/log:ro \
  -v crowdsec_db:/var/lib/crowdsec/data \
  -v crowdsec_config:/etc/crowdsec \
  -p 8080:8080 \
  crowdsecurity/crowdsec
```

## Generating a Bouncer API Key

Register a new bouncer with CrowdSec to obtain an API key:

```bash
# On the CrowdSec host
sudo cscli bouncers add caddy-bouncer
```

Output:
```
API key for 'caddy-bouncer':

   ********************************

Please keep this key since you will not be able to retrieve it!
```

Save this API key securely - you'll need it for Caddy configuration.

### Managing Bouncers

```bash
# List registered bouncers
sudo cscli bouncers list

# Remove a bouncer
sudo cscli bouncers delete caddy-bouncer
```

## Caddy Configuration

### Global CrowdSec Block (Required)

Add the CrowdSec app configuration to your Caddyfile's global options:

```caddyfile
{
    crowdsec {
        api_url http://localhost:8080
        api_key YOUR_BOUNCER_API_KEY
        ticker_interval 15s
    }
}
```

### Configuration Options

| Option | Required | Default | Description |
|--------|----------|---------|-------------|
| `api_url` | Yes | - | CrowdSec Local API endpoint |
| `api_key` | Yes | - | Bouncer API key from `cscli bouncers add` |
| `ticker_interval` | No | `15s` | How often to poll for decision updates |
| `appsec_url` | No | - | AppSec component URL for WAF features |
| `disable_streaming` | No | `false` | Use polling instead of streaming mode |
| `enable_hard_fails` | No | `false` | Block all traffic if API is unavailable |

### HTTP Handler

Add the `crowdsec` directive to block banned IPs at the HTTP layer:

```caddyfile
example.com {
    route {
        crowdsec
        reverse_proxy localhost:8080
    }
}
```

### AppSec Handler (WAF)

For application-level security rules, add the `appsec` directive:

```caddyfile
{
    crowdsec {
        api_url http://localhost:8080
        api_key YOUR_BOUNCER_API_KEY
        appsec_url http://localhost:7422
    }
}

example.com {
    route {
        appsec
        crowdsec
        reverse_proxy localhost:8080
    }
}
```

### Layer 4 Blocking

Block connections at the TCP/UDP layer before TLS handshake:

```caddyfile
{
    crowdsec {
        api_url http://localhost:8080
        api_key YOUR_BOUNCER_API_KEY
    }

    layer4 {
        :443 {
            @banned crowdsec
            route @banned {
                # Banned IPs are dropped here
            }
            route {
                tls
                proxy localhost:8443
            }
        }
    }
}
```

## Complete Example

A full Caddyfile with SPIRE TLS and CrowdSec protection:

```caddyfile
{
    crowdsec {
        api_url http://localhost:8080
        api_key {$CROWDSEC_API_KEY}
        ticker_interval 15s
    }
}

example.com {
    tls {
        issuer spire {
            socket_path /tmp/spire-agent/public/api.sock
        }
    }

    encode br gzip

    route {
        crowdsec
        reverse_proxy localhost:8080
    }
}
```

## Environment Variables

For security, use environment variables for the API key:

```bash
export CROWDSEC_API_KEY="your-api-key-here"
caddy run --config Caddyfile
```

Reference in Caddyfile:
```caddyfile
{
    crowdsec {
        api_url http://localhost:8080
        api_key {$CROWDSEC_API_KEY}
    }
}
```

## Verifying the Integration

### Check Bouncer Connection

```bash
# Verify bouncer is connected
sudo cscli bouncers list
```

Look for your bouncer with a recent "Last API Pull" timestamp.

### Check Caddy Modules

```bash
./bin/caddy-with-spire list-modules | grep crowdsec
```

Expected output:
```
admin.api.crowdsec
crowdsec
http.handlers.crowdsec
layer4.matchers.crowdsec
```

### Test with a Banned IP

```bash
# Add a test ban
sudo cscli decisions add --ip 192.168.1.100 --duration 1h --reason "test ban"

# Verify the decision
sudo cscli decisions list

# Test from the banned IP (should be blocked)
curl -v https://example.com

# Remove the test ban
sudo cscli decisions delete --ip 192.168.1.100
```

## Operational Modes

### Stream Mode (Default)

The bouncer polls CrowdSec at `ticker_interval` for decision updates. Efficient for high-traffic sites.

### Live Mode

Query CrowdSec for every request. More resource-intensive but provides real-time blocking:

```caddyfile
{
    crowdsec {
        api_url http://localhost:8080
        api_key {$CROWDSEC_API_KEY}
        disable_streaming
    }
}
```

## Troubleshooting

### Bouncer Not Connecting

1. Verify CrowdSec is running:
   ```bash
   sudo systemctl status crowdsec
   ```

2. Check Local API is accessible:
   ```bash
   curl http://localhost:8080/health
   ```

3. Verify API key is correct:
   ```bash
   sudo cscli bouncers list
   ```

### Decisions Not Being Applied

1. Check decision list:
   ```bash
   sudo cscli decisions list
   ```

2. Verify ticker interval hasn't elapsed yet (default 15s)

3. Check Caddy logs for CrowdSec-related errors

### AppSec Not Working

1. Ensure AppSec component is installed and running
2. Verify `appsec_url` is correctly configured
3. Check AppSec rules are loaded:
   ```bash
   sudo cscli appsec-rules list
   ```

## Security Considerations

- Store API keys in environment variables, not in config files
- Use `enable_hard_fails` in high-security environments to block traffic when CrowdSec is unavailable
- Regularly update CrowdSec and its collections for latest threat intelligence
- Monitor bouncer connectivity in production

## Resources

- [CrowdSec Documentation](https://docs.crowdsec.net/)
- [caddy-crowdsec-bouncer GitHub](https://github.com/hslatman/caddy-crowdsec-bouncer)
- [CrowdSec Hub - Collections & Parsers](https://hub.crowdsec.net/)
