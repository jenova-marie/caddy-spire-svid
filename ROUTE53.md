# AWS Route53 DNS Integration Guide

This guide covers configuring AWS Route53 as a DNS-01 ACME challenge provider for automatic TLS certificate issuance with your custom Caddy build.

## Overview

DNS-01 challenges prove domain ownership by creating a TXT record in your DNS zone. Unlike HTTP-01 challenges, DNS-01 supports:

- **Wildcard certificates** (`*.example.com`)
- **Servers without public HTTP access** (internal services, firewalled hosts)
- **Port 80/443 not required** during certificate issuance

The `dns.providers.route53` module handles Route53 record creation/cleanup automatically during ACME certificate issuance and renewal.

**Module included in this build:**
- `dns.providers.route53` - AWS Route53 DNS challenge solver

## Prerequisites

- An AWS account with Route53 hosted zone(s)
- IAM credentials with Route53 permissions
- Domain(s) managed in Route53

## AWS IAM Policy

### Minimal Permissions

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "route53:ListHostedZones",
        "route53:ListHostedZonesByName",
        "route53:GetChange",
        "route53:ChangeResourceRecordSets",
        "route53:ListResourceRecordSets"
      ],
      "Resource": "*"
    }
  ]
}
```

### Scoped to Specific Hosted Zone

For tighter security, restrict to a specific hosted zone:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "route53:GetChange",
        "route53:ChangeResourceRecordSets",
        "route53:ListResourceRecordSets"
      ],
      "Resource": [
        "arn:aws:route53:::hostedzone/Z0123456789ABCDEFGHIJ",
        "arn:aws:route53:::change/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "route53:ListHostedZones",
        "route53:ListHostedZonesByName"
      ],
      "Resource": "*"
    }
  ]
}
```

## Authentication Methods

The module uses the AWS SDK for Go v2 credential chain. Methods are tried in order:

### 1. Environment Variables (Recommended for containers)

```bash
export AWS_ACCESS_KEY_ID="AKIAIOSFODNN7EXAMPLE"
export AWS_SECRET_ACCESS_KEY="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
export AWS_REGION="us-east-1"

# Optional: for temporary credentials (STS)
export AWS_SESSION_TOKEN="..."
```

### 2. Shared Credentials File

```bash
# ~/.aws/credentials
[default]
aws_access_key_id = AKIAIOSFODNN7EXAMPLE
aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY

# Use a named profile
export AWS_PROFILE="production"
```

### 3. EC2 Instance Role / IAM Role (Recommended for production on AWS)

No configuration needed. The SDK automatically retrieves credentials from the instance metadata service or ECS task role.

### 4. Explicit Caddyfile Configuration (Development only)

```caddyfile
tls {
    dns route53 {
        access_key_id AKIAIOSFODNN7EXAMPLE
        secret_access_key wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
        region us-east-1
    }
}
```

**Warning:** Avoid hardcoding credentials in config files. Use environment variable references instead:

```caddyfile
tls {
    dns route53 {
        access_key_id {$AWS_ACCESS_KEY_ID}
        secret_access_key {$AWS_SECRET_ACCESS_KEY}
        region {$AWS_REGION}
    }
}
```

## Caddy Configuration

### Caddyfile Syntax

```caddyfile
tls {
    dns route53 {
        access_key_id <key>                   # AWS access key (optional with IAM roles)
        secret_access_key <secret>            # AWS secret key (optional with IAM roles)
        session_token <token>                 # STS session token (optional)
        region <region>                       # AWS region (default: us-east-1)
        profile <name>                        # AWS profile name (optional)
        max_retries <n>                       # Max API retries (optional)
        hosted_zone_id <id>                   # Explicit hosted zone ID (optional)
        wait_for_route53_sync <bool>          # Wait for DNS propagation (default: true)
        route53_max_wait <duration>           # Max wait for DNS sync (optional)
        skip_route53_sync_on_delete <bool>    # Skip sync on cleanup (default: true)
    }
}
```

### Configuration Options

| Option | Required | Default | Description |
|--------|----------|---------|-------------|
| `access_key_id` | No | From env/role | AWS access key ID |
| `secret_access_key` | No | From env/role | AWS secret access key |
| `session_token` | No | - | STS temporary session token |
| `region` | No | `us-east-1` | AWS region |
| `profile` | No | `default` | AWS credentials profile name |
| `max_retries` | No | SDK default | Maximum API retry attempts |
| `hosted_zone_id` | No | Auto-detected | Explicit Route53 hosted zone ID |
| `wait_for_route53_sync` | No | `true` | Wait for DNS change to propagate |
| `route53_max_wait` | No | SDK default | Maximum wait time for DNS propagation |
| `skip_route53_sync_on_delete` | No | `true` | Skip waiting when cleaning up TXT records |

## Examples

### Basic Route53 ACME

Using environment variable credentials (simplest setup):

```caddyfile
example.com {
    tls {
        dns route53 {
            region us-east-1
        }
    }
    reverse_proxy localhost:8080
}
```

### Wildcard Certificate

```caddyfile
*.example.com {
    tls {
        dns route53 {
            region us-east-1
            hosted_zone_id Z0123456789ABCDEFGHIJ
        }
    }

    @app host app.example.com
    handle @app {
        reverse_proxy localhost:3000
    }

    @api host api.example.com
    handle @api {
        reverse_proxy localhost:8080
    }
}
```

### Hybrid: Route53 ACME + SPIRE

Public-facing services get certificates from a public CA via Route53 DNS challenge. Internal services use SPIRE for zero-trust mTLS.

```caddyfile
{
    email admin@example.com
}

# Public service: Route53 ACME for public TLS
public.example.com {
    tls {
        dns route53 {
            region us-east-1
        }
    }

    encode br gzip
    reverse_proxy localhost:3000
}

# Internal service: SPIRE for zero-trust internal TLS
internal.example.com {
    tls {
        issuer spire {
            socket_path /tmp/spire-agent/public/api.sock
            refresh_at_percent 65
        }
    }
    reverse_proxy localhost:9000
}
```

### Multiple Domains with Explicit Credentials

```caddyfile
{
    email admin@example.com
}

app.example.com, api.example.com {
    tls {
        dns route53 {
            access_key_id {$AWS_ACCESS_KEY_ID}
            secret_access_key {$AWS_SECRET_ACCESS_KEY}
            region {$AWS_REGION}
        }
    }

    @app host app.example.com
    handle @app {
        reverse_proxy localhost:3000
    }

    @api host api.example.com
    handle @api {
        reverse_proxy localhost:8080
    }
}
```

## Environment Variables

For security, always use environment variables for credentials:

```bash
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"
export AWS_REGION="us-east-1"
caddy run --config Caddyfile
```

Reference in Caddyfile with `{$VAR_NAME}` syntax:

```caddyfile
tls {
    dns route53 {
        access_key_id {$AWS_ACCESS_KEY_ID}
        secret_access_key {$AWS_SECRET_ACCESS_KEY}
        region {$AWS_REGION}
    }
}
```

## Verifying the Integration

### Check Module Registration

```bash
./bin/caddy-with-spire list-modules | grep route53
```

Expected output:
```
dns.providers.route53
```

### Test Configuration

```bash
# Validate a Caddyfile that uses Route53
./bin/caddy-with-spire validate --config Caddyfile
```

### Verify DNS Record Creation

During certificate issuance, check Route53 for the `_acme-challenge` TXT record:

```bash
aws route53 list-resource-record-sets \
  --hosted-zone-id Z0123456789ABCDEFGHIJ \
  --query "ResourceRecordSets[?Name=='_acme-challenge.example.com.']"
```

## Troubleshooting

### IAM Permission Errors

```
Error: AccessDenied: User is not authorized to perform: route53:ChangeResourceRecordSets
```

Verify your IAM policy includes all required actions. The `ListHostedZones` permission is needed for zone auto-detection even if you specify `hosted_zone_id`.

### Region Configuration

Route53 is a global service, but the SDK still requires a region. If you get credential errors, ensure `AWS_REGION` is set or `region` is configured in the Caddyfile.

### DNS Propagation Timeouts

If certificate issuance fails with propagation timeouts, increase the wait time:

```caddyfile
dns route53 {
    route53_max_wait 5m
}
```

### Wrong Hosted Zone Selected

If you have multiple hosted zones for the same domain, specify the zone explicitly:

```caddyfile
dns route53 {
    hosted_zone_id Z0123456789ABCDEFGHIJ
}
```

Find your hosted zone ID:
```bash
aws route53 list-hosted-zones --query "HostedZones[*].[Id,Name]" --output table
```

### Credential Chain Issues

Enable debug logging to see which credential source the SDK is using:

```bash
AWS_SDK_LOG_LEVEL=debug caddy run --config Caddyfile
```

## Security Considerations

- **Never commit AWS credentials** to version control
- **Use IAM roles** in production on AWS (EC2, ECS, EKS) instead of static credentials
- **Scope IAM policies** to specific hosted zones when possible
- **Rotate access keys** regularly if using static credentials
- **Use STS temporary credentials** for short-lived operations
- **Enable CloudTrail** to audit Route53 API calls

## Resources

- [caddy-dns/route53 GitHub](https://github.com/caddy-dns/route53)
- [Caddy DNS Challenge Documentation](https://caddyserver.com/docs/automatic-https#dns-challenge)
- [AWS Route53 Documentation](https://docs.aws.amazon.com/Route53/)
- [AWS IAM Best Practices](https://docs.aws.amazon.com/IAM/latest/UserGuide/best-practices.html)
