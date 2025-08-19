// Package main demonstrates basic integration of caddy-spire-client
// with a simple HTTP server using SPIFFE/SPIRE certificates.
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jenova-marie/caddy-spire-client/pkg/spire"
)

func main() {
	// Create SPIRE client configuration
	config := spire.Config{
		SocketPath:      "/tmp/spire-agent/public/api.sock",
		RefreshInterval: 30 * time.Second,
	}

	// Initialize SPIRE client
	client, err := spire.NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create SPIRE client: %v", err)
	}
	defer client.Close()

	// Get TLS configuration from SPIRE
	tlsConfig := client.GetTLSConfig()

	// Create HTTP server with SPIFFE certificates
	server := &http.Server{
		Addr:      ":8443",
		TLSConfig: tlsConfig,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "🌸 Hello from SPIFFE-secured server!\n")
			fmt.Fprintf(w, "Request from: %s\n", r.RemoteAddr)
			fmt.Fprintf(w, "Secured with SPIRE certificates! ✨\n")

			// Show SVID info
			svid, err := client.GetCurrentSVID()
			if err == nil && len(svid.Certificates) > 0 {
				fmt.Fprintf(w, "Server SPIFFE ID: %s\n", svid.ID)
				fmt.Fprintf(w, "Certificate expires: %v\n", svid.Certificates[0].NotAfter)
			}
		}),
	}

	log.Println("🚀 Starting SPIFFE-secured HTTPS server on :8443")
	log.Println("💖 Server secured with SPIRE certificates")

	// Start server with TLS
	if err := server.ListenAndServeTLS("", ""); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
