// caddy-with-spire builds a custom Caddy server with SPIFFE/SPIRE support.
// This is a complete Caddy server with the SPIRE certificate issuer module included.
package main

import (
	caddycmd "github.com/caddyserver/caddy/v2/cmd"

	// Standard Caddy modules
	_ "github.com/caddyserver/caddy/v2/modules/standard"

	// Our SPIRE certificate issuer module
	_ "github.com/jenova-marie/caddy-spire-svid/pkg/caddyspire"
)

func main() {
	caddycmd.Main()
}
