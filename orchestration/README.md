# SPIRE Docker Diagnostic Tool

🐳 **Minimal Docker container** for testing SPIRE agent connectivity.

## Purpose

This tool helps debug Docker + SPIRE networking by:
- Connecting to local `spire-agent` from inside a Docker container
- Performing workload attestation 
- Retrieving and validating SVIDs
- Printing detailed diagnostic information

## 🎯 macOS Docker Desktop Solution

**Problem**: Docker Desktop on macOS cannot properly bind mount Unix sockets (`operation not supported`).

**Solution**: Run SPIRE agent in a Docker container too! Both the agent and workloads are in the same Linux VM where Unix sockets work perfectly.

## Quick Start

### Basic Diagnostic (requires host SPIRE agent)
```bash
# Test locally first (no Docker)
make local

# Test in Docker container  
make test

# Clean up
make clean
```

### 🚀 Full Orchestration (macOS Solution)
```bash
# Build everything and run full SPIRE + Caddy orchestration
make full-test

# Test Caddy with SPIRE certificates
curl -k https://localhost:8443

# Clean up everything
make full-clean
```

## SPIRE Server Requirements

You need a workload entry for the diagnostic container:

```json
{
    "spiffe_id": "spiffe://recoverysky.org/development/jenova/docker/spire-diagnostic",
    "parent_id": "spiffe://recoverysky.org/spire/agent/join_token/YOUR-TOKEN",
    "selectors": [
        {
            "type": "docker",
            "value": "container_name:spire-diagnostic"
        }
    ],
    "x509_svid_ttl": 3600,
    "dns_names": ["localhost", "spire-diagnostic"]
}
```

This tool is perfect for isolating SPIRE + Docker networking issues! 🌸
