# 🔄 Certificate Refresh Mechanisms

## Overview

The Caddy SPIRE module supports two mutually exclusive certificate refresh strategies:

1. **`refresh_interval`** - Fixed interval refresh (legacy approach)
2. **`refresh_at_percent`** - Percentage-based refresh (recommended)

## Comparison

| Feature | `refresh_interval` | `refresh_at_percent` |
|---------|-------------------|---------------------|
| **Approach** | Fixed time intervals | Percentage of certificate lifetime |
| **Default** | 30 seconds | 65% |
| **Efficiency** | Less efficient (polling) | More efficient (calculated) |
| **Best for** | Testing, short-lived certs | Production, variable cert lifetimes |
| **SPIFFE Alignment** | Legacy approach | Aligns with best practices |

## Configuration Examples

### Using `refresh_interval` (Fixed Interval)

```caddyfile
example.com {
    tls {
        get_certificate spire {
            socket_path /tmp/spire-agent/public/api.sock
            refresh_interval 30s  # Check every 30 seconds
        }
    }
}
```

**How it works:**
- Checks certificate status every 30 seconds
- Logs current certificate expiration
- Updates certificate files if configured
- Simple but less efficient

### Using `refresh_at_percent` (Percentage-based) ⭐ Recommended

```caddyfile
example.com {
    tls {
        get_certificate spire {
            socket_path /tmp/spire-agent/public/api.sock
            refresh_at_percent 65  # Refresh at 65% of lifetime
        }
    }
}
```

**How it works:**
- Calculates exact refresh time based on certificate lifetime
- Sleeps until refresh is needed
- More efficient, no unnecessary checks
- Adapts to different certificate lifetimes

## Mutual Exclusivity

⚠️ **Important**: You cannot use both options simultaneously. The module will return an error:

```caddyfile
# ❌ This will fail!
example.com {
    tls {
        get_certificate spire {
            refresh_interval 30s
            refresh_at_percent 65  # ERROR: Mutually exclusive!
        }
    }
}
```

## Default Behavior

If neither option is specified, the module defaults to:
- `refresh_at_percent: 65`

This means certificates will refresh when they reach 65% of their lifetime.

## Examples by Certificate Lifetime

### With `refresh_at_percent: 65` (default)

| Certificate Lifetime | Refresh Time | Time Until Refresh |
|---------------------|--------------|-------------------|
| 1 hour | 39 minutes | After 39 minutes |
| 24 hours | 15.6 hours | After 15 hours 36 min |
| 7 days | 4.55 days | After 4 days 13 hours |
| 30 days | 19.5 days | After 19 days 12 hours |

### With `refresh_interval: 30s`

| Certificate Lifetime | Check Frequency | Total Checks |
|---------------------|----------------|--------------|
| 1 hour | Every 30s | 120 checks |
| 24 hours | Every 30s | 2,880 checks |
| 7 days | Every 30s | 20,160 checks |
| 30 days | Every 30s | 86,400 checks |

## Performance Comparison

### CPU Usage Over 24 Hours

```
refresh_interval (30s):
├── Checks: 2,880
├── CPU wake-ups: 2,880
└── Efficiency: Low ❌

refresh_at_percent (65%):
├── Checks: 1-2
├── CPU wake-ups: 1-2
└── Efficiency: High ✅
```

## Best Practices

### Production Environments

Use `refresh_at_percent`:
```caddyfile
production.example.com {
    tls {
        get_certificate spire {
            refresh_at_percent 75  # Conservative for production
        }
    }
}
```

### Development/Testing

Use `refresh_interval` for predictable behavior:
```caddyfile
test.example.com {
    tls {
        get_certificate spire {
            refresh_interval 10s  # Frequent checks for testing
        }
    }
}
```

### Short-lived Certificates

For certificates < 1 hour, consider lower percentages:
```caddyfile
short-lived.example.com {
    tls {
        get_certificate spire {
            refresh_at_percent 50  # Refresh at halfway point
        }
    }
}
```

## Implementation Details

### `autoRefresh` (Fixed Interval)
```go
func (c *Client) autoRefresh(interval time.Duration) {
    ticker := time.NewTicker(interval)
    for range ticker.C {
        // Check and log certificate status
        // Update files if needed
    }
}
```

### `smartRefresh` (Percentage-based)
```go
func (c *Client) smartRefresh(refreshAtPercent int) {
    for {
        cert := getCurrentCertificate()
        lifetime := cert.NotAfter.Sub(cert.NotBefore)
        refreshTime := lifetime * refreshAtPercent / 100
        
        time.Sleep(refreshTime)
        // Refresh certificate
    }
}
```

## Troubleshooting

### Issue: "RefreshInterval and RefreshAtPercent are mutually exclusive"
**Solution**: Use only one option:
```caddyfile
# Good ✅
tls {
    get_certificate spire {
        refresh_at_percent 65
    }
}

# Also good ✅
tls {
    get_certificate spire {
        refresh_interval 1m
    }
}
```

### Issue: Certificates not refreshing in time
**For `refresh_at_percent`**: Use a lower percentage
```caddyfile
refresh_at_percent 50  # Refresh at 50% instead of 65%
```

**For `refresh_interval`**: Use a shorter interval
```caddyfile
refresh_interval 15s  # Check every 15 seconds instead of 30s
```

### Debug Logging

Enable debug mode to see refresh calculations:
```caddyfile
{
    debug
}
```

Log output shows:
- For `refresh_at_percent`: "Certificate expires at X, will refresh at Y (in Z)"
- For `refresh_interval`: "SVID available, expires at X"

## Migration Guide

### From `refresh_interval` to `refresh_at_percent`

Before:
```caddyfile
tls {
    get_certificate spire {
        refresh_interval 30s
    }
}
```

After:
```caddyfile
tls {
    get_certificate spire {
        refresh_at_percent 65  # Or omit for default
    }
}
```

Benefits:
- ⚡ Reduced CPU usage
- 🎯 More precise refresh timing
- 📈 Better scalability
- 🔧 Adapts to certificate lifetime changes
