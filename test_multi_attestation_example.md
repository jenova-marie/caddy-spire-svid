# Multi-Attestation Test Scenario

## SPIRE Entries Configuration

Based on your provided SPIRE entries, the system now supports automatic certificate selection:

```json
{
    "spiffe_id": "spiffe://recoverysky.org/prod/mnemosyne/caddy-vault",
    "dns_names": ["vault.rso"]
},
{
    "spiffe_id": "spiffe://recoverysky.org/prod/mnemosyne/caddy-neo4j", 
    "dns_names": ["neo4j.rso"]
},
{
    "spiffe_id": "spiffe://recoverysky.org/prod/mnemosyne/caddy-oidc",
    "dns_names": ["oidc.rso"]
},
{
    "spiffe_id": "spiffe://recoverysky.org/prod/mnemosyne/caddy-ip",
    "dns_names": ["10.1.1.11"]
}
```

## How It Works Now

1. **Caddy Request**: When a client connects to `https://vault.rso`, Caddy calls `GetCertificate()` with `hello.ServerName = "vault.rso"`

2. **SVID Selection**: The new `GetSVIDForServerName("vault.rso")` method:
   - Fetches all available SVIDs from SPIRE agent
   - Searches through each SVID's certificate DNS names
   - Finds the SVID with `dns_names: ["vault.rso"]`
   - Returns the `spiffe://recoverysky.org/prod/mnemosyne/caddy-vault` SVID

3. **Certificate Chain**: The system builds a complete TLS certificate chain with:
   - Leaf certificate (SPIRE SVID for vault.rso)
   - Intermediate certificates (if any)
   - CA certificates from trust bundle

## Expected Log Output

```
🔍 Selecting SVID for server name: vault.rso from 4 available SVIDs
   SVID spiffe://recoverysky.org/prod/mnemosyne/caddy-vault has DNS names: [vault.rso]
   SVID spiffe://recoverysky.org/prod/mnemosyne/caddy-neo4j has DNS names: [neo4j.rso]
   SVID spiffe://recoverysky.org/prod/mnemosyne/caddy-oidc has DNS names: [oidc.rso]
   SVID spiffe://recoverysky.org/prod/mnemosyne/caddy-ip has DNS names: [10.1.1.11]
✅ Found matching SVID for vault.rso: SPIFFE ID spiffe://recoverysky.org/prod/mnemosyne/caddy-vault
🎉 Successfully got certificate chain via SPIRE for server name vault.rso
```

## Automatic Mapping

The system automatically maps:
- `vault.rso` → `spiffe://recoverysky.org/prod/mnemosyne/caddy-vault`
- `neo4j.rso` → `spiffe://recoverysky.org/prod/mnemosyne/caddy-neo4j`  
- `oidc.rso` → `spiffe://recoverysky.org/prod/mnemosyne/caddy-oidc`
- `10.1.1.11` → `spiffe://recoverysky.org/prod/mnemosyne/caddy-ip`

No additional configuration needed! 🎉
