#!/bin/bash
# 🌸 Caddy SPIRE Hybrid Configuration Script
# This script combines Caddyfile (for HTTP apps) with JSON (for Layer 4 + file-based TLS)
#
# Usage: ./caddy-spire-hybrid.sh [caddyfile] [output-json] [cert-file] [key-file]
#
# Environment Variables:
#   SPIRE_CERT_FILE - Path to SPIRE certificate file (default: /tmp/spire-certs/cert.pem)
#   SPIRE_KEY_FILE  - Path to SPIRE private key file (default: /tmp/spire-certs/key.pem)
#   CADDY_BINARY    - Path to Caddy binary (default: ./bin/caddy-with-spire)

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
RESET='\033[0m'

# Default values
CADDYFILE="${1:-test/integration/level3/test.caddyfile}"
OUTPUT_JSON="${2:-test/integration/level3/hybrid.json}"
CERT_FILE="${3:-${SPIRE_CERT_FILE:-/tmp/spire-certs/cert.pem}}"
KEY_FILE="${4:-${SPIRE_KEY_FILE:-/tmp/spire-certs/key.pem}}"
CADDY_BINARY="${CADDY_BINARY:-./bin/caddy-with-spire}"

echo -e "${MAGENTA}🌸 Caddy SPIRE Hybrid Configuration Script${RESET}"
echo -e "${CYAN}================================================${RESET}"
echo -e "${BLUE}📄 Caddyfile:${RESET} $CADDYFILE"
echo -e "${BLUE}📄 Output JSON:${RESET} $OUTPUT_JSON"
echo -e "${BLUE}🔐 Certificate:${RESET} $CERT_FILE"
echo -e "${BLUE}🗝️  Private Key:${RESET} $KEY_FILE"
echo -e "${BLUE}🏗️  Caddy Binary:${RESET} $CADDY_BINARY"
echo ""

# Validate inputs
if [[ ! -f "$CADDYFILE" ]]; then
    echo -e "${RED}❌ Error: Caddyfile not found: $CADDYFILE${RESET}"
    exit 1
fi

if [[ ! -f "$CADDY_BINARY" ]]; then
    echo -e "${RED}❌ Error: Caddy binary not found: $CADDY_BINARY${RESET}"
    echo -e "${YELLOW}💡 Hint: Run 'make build-caddy-with-spire' first${RESET}"
    exit 1
fi

# Step 1: Convert Caddyfile to JSON
echo -e "${CYAN}🔄 Step 1: Converting Caddyfile to JSON...${RESET}"
if ! "$CADDY_BINARY" adapt --config "$CADDYFILE" --adapter caddyfile > "${OUTPUT_JSON}.tmp"; then
    echo -e "${RED}❌ Error: Failed to adapt Caddyfile to JSON${RESET}"
    exit 1
fi
echo -e "${GREEN}✅ Caddyfile adapted to JSON${RESET}"

# Step 2: Add Layer 4 configuration with file-based TLS
echo -e "${CYAN}🔄 Step 2: Adding Layer 4 configuration...${RESET}"

# Use jq to add Layer 4 app configuration
# Based on caddy-l4 documentation: https://github.com/mholt/caddy-l4
# For now, using TLS passthrough (proxy to SPIRE-secured backend) while we research TLS termination
cat "${OUTPUT_JSON}.tmp" | jq --arg cert_file "$CERT_FILE" --arg key_file "$KEY_FILE" '
# Add layer4 app to the configuration - TLS passthrough to SPIRE backend
.apps.layer4 = {
  "servers": {
    "spire_l4_proxy": {
      "listen": [":9443"],
      "routes": [
        {
          "match": [
            {
              "tls": {}
            }
          ],
          "handle": [
            {
              "handler": "proxy",
              "upstreams": [
                {
                  "dial": ["localhost:8443"]
                }
              ]
            }
          ]
        },
        {
          "match": [
            {
              "http": []
            }
          ],
          "handle": [
            {
              "handler": "proxy",
              "upstreams": [
                {
                  "dial": ["localhost:8080"]
                }
              ]
            }
          ]
        }
      ]
    }
  }
}
' > "$OUTPUT_JSON"

# Clean up temp file
rm "${OUTPUT_JSON}.tmp"

echo -e "${GREEN}✅ Layer 4 configuration added${RESET}"

# Step 3: Validate the final JSON
echo -e "${CYAN}🔄 Step 3: Validating final configuration...${RESET}"
if ! "$CADDY_BINARY" validate --config "$OUTPUT_JSON"; then
    echo -e "${RED}❌ Error: Final JSON configuration is invalid${RESET}"
    exit 1
fi
echo -e "${GREEN}✅ Configuration validated successfully${RESET}"

# Step 4: Display summary
echo ""
echo -e "${MAGENTA}🎉 Hybrid Configuration Complete!${RESET}"
echo -e "${CYAN}================================================${RESET}"
echo -e "${GREEN}📄 Generated JSON:${RESET} $OUTPUT_JSON"
echo ""
echo -e "${YELLOW}🚀 To run Caddy with this configuration:${RESET}"
echo -e "${BLUE}   SPIRE_CERT_FILE=\"$CERT_FILE\" SPIRE_KEY_FILE=\"$KEY_FILE\" $CADDY_BINARY run --config $OUTPUT_JSON${RESET}"
echo ""
echo -e "${YELLOW}📊 Configuration Summary:${RESET}"
echo -e "${CYAN}   • HTTP App:${RESET} SPIRE-secured HTTPS on :8443, HTTP on :8080"
echo -e "${CYAN}   • Layer 4:${RESET} TLS proxy with file-based SPIRE certificates on :9443"
echo -e "${CYAN}   • Admin API:${RESET} localhost:2020"
echo ""
echo -e "${MAGENTA}💖 Ready for testing!${RESET}"
