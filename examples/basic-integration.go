// Package main demonstrates basic programmatic integration of caddy-spire-svid
// Shows how to use the SPIRE client directly in Go applications
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jenova-marie/caddy-spire-svid/pkg/spire"
)

func main() {
	log.Println("🌸 Starting SPIRE Integration Demo")

	// Create SPIRE client configuration with percentage-based refresh
	config := spire.Config{
		SocketPath:       "/tmp/spire-agent/public/api.sock",
		RefreshAtPercent: 65, // Refresh at 65% of certificate lifetime
		RefreshTimeout:   10 * time.Second,
	}

	// Initialize SPIRE client
	log.Printf("🔗 Connecting to SPIRE agent at %s", config.SocketPath)
	client, err := spire.NewClient(config)
	if err != nil {
		log.Fatalf("❌ Failed to create SPIRE client: %v", err)
	}
	defer client.Close()

	// Get TLS configuration from SPIRE
	tlsConfig := client.GetTLSConfig()
	if tlsConfig == nil {
		log.Fatal("❌ Failed to get TLS configuration from SPIRE")
	}

	// Create HTTP handlers
	mux := http.NewServeMux()

	// Main handler
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "🌸 Hello from SPIFFE-secured server!\n")
		fmt.Fprintf(w, "Request from: %s\n", r.RemoteAddr)
		fmt.Fprintf(w, "Method: %s %s\n", r.Method, r.URL.Path)
		fmt.Fprintf(w, "Secured with SPIRE certificates! ✨\n\n")

		// Show current SVID information
		svid, err := client.GetCurrentSVID()
		if err != nil {
			fmt.Fprintf(w, "⚠️  Could not get SVID info: %v\n", err)
		} else if len(svid.Certificates) > 0 {
			cert := svid.Certificates[0]
			fmt.Fprintf(w, "📋 Certificate Info:\n")
			fmt.Fprintf(w, "   SPIFFE ID: %s\n", svid.ID)
			fmt.Fprintf(w, "   Subject: %s\n", cert.Subject)
			fmt.Fprintf(w, "   Valid from: %v\n", cert.NotBefore)
			fmt.Fprintf(w, "   Expires at: %v\n", cert.NotAfter)
			fmt.Fprintf(w, "   Remaining: %v\n", time.Until(cert.NotAfter))
		}
	})

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Check SPIRE connectivity
		svid, err := client.GetCurrentSVID()
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, `{"status": "unhealthy", "spire": "disconnected", "error": "%v"}`, err)
			return
		}

		status := "healthy"
		if len(svid.Certificates) > 0 {
			cert := svid.Certificates[0]
			if time.Until(cert.NotAfter) < 24*time.Hour {
				status = "warning"
			}
		}

		fmt.Fprintf(w, `{
  "status": "%s",
  "spire": "connected",
  "spiffe_id": "%s",
  "certificate_expires": "%v"
}`, status, svid.ID, svid.Certificates[0].NotAfter)
	})

	// SPIRE info endpoint
	mux.HandleFunc("/spire", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		svid, err := client.GetCurrentSVID()
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, `{"error": "Cannot get SVID: %v"}`, err)
			return
		}

		if len(svid.Certificates) == 0 {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprint(w, `{"error": "No certificates available"}`)
			return
		}

		cert := svid.Certificates[0]
		fmt.Fprintf(w, `{
  "spiffe_id": "%s",
  "certificate": {
    "subject": "%s",
    "serial": "%s",
    "not_before": "%v",
    "not_after": "%v",
    "dns_names": %q,
    "ip_addresses": "%v"
  },
  "socket_path": "%s"
}`, svid.ID, cert.Subject, cert.SerialNumber,
			cert.NotBefore, cert.NotAfter, cert.DNSNames,
			cert.IPAddresses, config.SocketPath)
	})

	// Create HTTP server with SPIFFE certificates
	server := &http.Server{
		Addr:      ":8443",
		TLSConfig: tlsConfig,
		Handler:   mux,

		// Reasonable timeouts
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("🛑 Received shutdown signal, gracefully stopping...")

		// Give the server 30 seconds to finish existing requests
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("❌ Server shutdown error: %v", err)
		}
		cancel()
	}()

	log.Println("🚀 Starting SPIFFE-secured HTTPS server on :8443")
	log.Println("💖 Server secured with SPIRE certificates")
	log.Println("🌐 Endpoints:")
	log.Println("   https://localhost:8443/        - Main page")
	log.Println("   https://localhost:8443/health  - Health check")
	log.Println("   https://localhost:8443/spire   - SPIRE info")

	// Start server with TLS (certificates come from SPIRE)
	if err := server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ Server failed: %v", err)
	}

	// Wait for graceful shutdown
	<-ctx.Done()
	log.Println("💖 Server shutdown complete")
}
