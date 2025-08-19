#!/bin/bash
# 🌸 Register test workload with local SPIRE agent
# This script registers the Go test process as a valid SPIRE workload

set -e

echo "🌸 Registering workloads with local SPIRE agent..."

# Check if spire-server binary is available
if ! command -v spire-server &> /dev/null; then
    echo "❌ spire-server binary not found in PATH"
    echo "Please install SPIRE server or add it to your PATH"
    exit 1
fi

# Get current user info
CURRENT_UID=$(id -u)
CURRENT_USER=$(whoami)
echo "📋 Current user: $CURRENT_USER (UID: $CURRENT_UID)"

# Check SPIRE server socket
SPIRE_SERVER_SOCKET="/tmp/spire-server/private/api.sock"
if [ ! -S "$SPIRE_SERVER_SOCKET" ]; then
    echo "❌ SPIRE server socket not found at $SPIRE_SERVER_SOCKET"
    echo "Please ensure SPIRE server is running"
    exit 1
fi

echo "🔧 Registering agent node attestor..."
# First, ensure we have a proper node attestor
spire-server entry create \
    -node \
    -spiffeID spiffe://example.org/spire/agent/local \
    -selector unix:uid:$CURRENT_UID \
    -socketPath "$SPIRE_SERVER_SOCKET" 2>/dev/null || echo "Node entry may already exist"

echo "🧪 Registering Go test workload..."
# Register the Go test workload entry
spire-server entry create \
    -spiffeID spiffe://example.org/go-test-workload \
    -parentID spiffe://example.org/spire/agent/local \
    -selector unix:uid:$CURRENT_UID \
    -socketPath "$SPIRE_SERVER_SOCKET" || {
    
    echo "⚠️  Failed to register with unix UID selector, trying process path..."
    
    # Try with process selector if UID doesn't work
    spire-server entry create \
        -spiffeID spiffe://example.org/go-test-workload \
        -parentID spiffe://example.org/spire/agent/local \
        -selector unix:path:$(which go) \
        -socketPath "$SPIRE_SERVER_SOCKET" || {
        
        echo "❌ Failed to register workload entry"
        echo "Please check your SPIRE server configuration and agent setup"
        exit 1
    }
}

echo "🐳 Registering Docker Caddy workload (UID 1000)..."
# Register Docker Caddy workload
spire-server entry create \
    -spiffeID spiffe://example.org/caddy \
    -parentID spiffe://example.org/spire/agent/local \
    -selector unix:uid:1000 \
    -socketPath "$SPIRE_SERVER_SOCKET" 2>/dev/null || echo "Caddy entry may already exist"

echo ""
echo "✨ Workload entries registered successfully!"
echo ""
echo "🔍 Verify registrations:"
echo "   spire-server entry show -socketPath $SPIRE_SERVER_SOCKET"
echo ""
echo "🧪 Run integration tests:"
echo "   go test ./pkg/spire -v -run TestRealSpire"
echo ""
echo "🐳 Test Docker Caddy:"
echo "   docker-compose -f docker-compose.caddy.yml up --build"
