// caddy-with-spire builds a custom Caddy server with SPIFFE/SPIRE support.
// This is a complete Caddy server with the SPIRE certificate issuer module included.
package main

import (
	caddycmd "github.com/caddyserver/caddy/v2/cmd"

	// Standard Caddy modules
	_ "github.com/caddyserver/caddy/v2/modules/standard"

	// Brotli compression encoder (pure Go implementation)
	_ "github.com/ueffel/caddy-brotli"

	// CrowdSec bouncer for blocking malicious traffic
	_ "github.com/hslatman/caddy-crowdsec-bouncer/appsec"
	_ "github.com/hslatman/caddy-crowdsec-bouncer/http"
	_ "github.com/hslatman/caddy-crowdsec-bouncer/layer4"

	// Layer 4 (TCP/UDP) proxy module
	_ "github.com/mholt/caddy-l4"

	// Rate limiting module
	_ "github.com/mholt/caddy-ratelimit"

	// Our SPIRE certificate issuer module
	_ "github.com/jenova-marie/caddy-spire-svid/pkg/caddyspire"
)

func main() {
	caddycmd.Main()
}
