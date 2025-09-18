# 🔒 Explicit SPIFFE ID Configuration

## Overview

The Caddy SPIRE module now supports explicit SPIFFE ID configuration, allowing administrators to specify exactly which SPIFFE identity should be used for a particular server, rather than relying on automatic DNS-based selection.

## Why Use Explicit SPIFFE ID?

### 1. **Enhanced Security** 🛡️
- Prevents accidental use of wrong SPIFFE identities
- Makes security policies explicit and auditable
- Reduces attack surface by limiting which identities can be used

### 2. **Better Performance** ⚡
- Direct SPIFFE ID lookup instead of searching through all available SVIDs
- Reduces overhead in multi-attestation environments
- Faster certificate retrieval

### 3. **Operational Clarity** 📋
- Configuration clearly shows which identity is used where
- Easier troubleshooting and debugging
- Better alignment with SPIFFE best practices

### 4. **Multi-Attestation Support** 🌐
- Essential for workloads with multiple SPIFFE identities
- Allows precise control over identity selection
- Supports complex deployment scenarios

## Configuration

### Caddyfile

```caddyfile
# Explicit SPIFFE ID selection
loki.rso:3100 {
    tls {
        get_certificate spire {
            socket_path /var/lib/spire/agent/socket/api.sock
            spiffe_id spiffe://recoverysky.org/prod/metis/caddy-loki
            refresh_at_percent 65
        }
    }
    # ... rest of configuration
}

# Automatic selection (fallback behavior)
example.com {
    tls {
        get_certificate spire {
            socket_path /tmp/spire-agent/public/api.sock
            # No spiffe_id - will auto-select based on DNS name
        }
    }
}
```

### JSON Configuration

```json
{
  "apps": {
    "tls": {
      "automation": {
        "policies": [{
          "subjects": ["loki.rso"],
          "issuers": [{
            "module": "spire",
            "socket_path": "/var/lib/spire/agent/socket/api.sock",
            "spiffe_id": "spiffe://recoverysky.org/prod/metis/caddy-loki",
            "refresh_at_percent": 65
          }]
        }]
      }
    }
  }
}
```

## How It Works

1. **With `spiffe_id` configured:**
   - Module directly requests the specified SPIFFE ID
   - Fails if the ID is not available
   - No fallback to other identities

2. **Without `spiffe_id` (default behavior):**
   - Module fetches all available SVIDs
   - Matches based on DNS names in certificates
   - Falls back to first available SVID if no match

## Example Scenarios

### Scenario 1: Service-Specific Identity
```caddyfile
vault.rso {
    tls {
        get_certificate spire {
            spiffe_id spiffe://recoverysky.org/prod/vault/server
        }
    }
}
```

### Scenario 2: Environment-Based Identity
```caddyfile
api.example.com {
    tls {
        get_certificate spire {
            spiffe_id spiffe://example.com/staging/api/server
        }
    }
}
```

### Scenario 3: Workload-Specific Identity
```caddyfile
payments.service {
    tls {
        get_certificate spire {
            spiffe_id spiffe://example.com/prod/payments/processor
        }
    }
}
```

## Troubleshooting

### Common Issues

1. **"SVID with SPIFFE ID not found"**
   - Verify the SPIFFE ID is correctly registered in SPIRE
   - Check workload attestation is working
   - Ensure the workload selector matches

2. **Certificate mismatch**
   - Verify DNS names in SPIRE entry match server name
   - Check SPIFFE ID is spelled correctly
   - Ensure no typos in configuration

### Debug Logging

Enable debug logging to see SPIFFE ID selection:
```caddyfile
{
    debug
}
```

Log output will show:
```
🔍 Looking for specific SPIFFE ID: spiffe://domain/workload
✅ Found matching SVID: spiffe://domain/workload
```

## Best Practices

1. **Use explicit IDs for production** - Always specify SPIFFE IDs in production environments
2. **Match DNS names** - Ensure SPIRE entries have correct DNS names
3. **One ID per service** - Use unique SPIFFE IDs for different services
4. **Document your IDs** - Maintain a registry of SPIFFE IDs and their purposes
5. **Test thoroughly** - Verify identity selection before production deployment

## Migration Guide

To migrate from automatic to explicit selection:

1. List current SPIFFE identities:
   ```bash
   spire-agent api fetch x509 -socketPath /tmp/spire-agent/public/api.sock
   ```

2. Identify which ID is used for each service

3. Update Caddyfile with explicit `spiffe_id`:
   ```caddyfile
   service.example {
       tls {
           get_certificate spire {
               spiffe_id spiffe://example.com/service
           }
       }
   }
   ```

4. Test configuration:
   ```bash
   caddy validate --config Caddyfile
   ```

5. Deploy and monitor logs for successful ID selection

## Related Documentation

- [SPIFFE ID Format](https://spiffe.io/docs/latest/spiffe-about/spiffe-concepts/#spiffe-id)
- [SPIRE Workload Registration](https://spiffe.io/docs/latest/deploying/registering/)
- [Caddy TLS Automation](https://caddyserver.com/docs/automatic-https)
