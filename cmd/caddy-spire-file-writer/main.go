// caddy-spire-file-writer demonstrates file-based SPIRE certificate management
// This utility fetches SPIRE certificates and writes them to disk for use by Layer 4 or other consumers
package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jenova-marie/caddy-spire-svid/pkg/spire"
)

func main() {
	var (
		socketPath = flag.String("socket", "/tmp/spire-agent/public/api.sock", "SPIRE agent socket path")
		certFile   = flag.String("cert", "/tmp/spire-certs/cert.pem", "Certificate file path")
		keyFile    = flag.String("key", "/tmp/spire-certs/key.pem", "Private key file path")
		interval   = flag.Duration("interval", 30*time.Second, "Refresh interval")
		oneShot    = flag.Bool("oneshot", false, "Write certificates once and exit")
	)
	flag.Parse()

	log.Printf("🌸 Starting SPIRE certificate file writer")
	log.Printf("   Socket: %s", *socketPath)
	log.Printf("   Certificate: %s", *certFile)
	log.Printf("   Key: %s", *keyFile)
	log.Printf("   Interval: %v", *interval)
	log.Printf("   One-shot: %v", *oneShot)

	// Create SPIRE client with file-based configuration
	config := spire.Config{
		SocketPath:      *socketPath,
		RefreshInterval: *interval,
		CertFile:        *certFile,
		KeyFile:         *keyFile,
		FileMode:        0644, // More permissive for Layer 4 to read
	}

	client, err := spire.NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create SPIRE client: %v", err)
	}
	defer client.Close()

	log.Printf("✅ SPIRE client initialized successfully")
	log.Printf("📁 Certificate files will be maintained at:")
	log.Printf("   Cert: %s", *certFile)
	log.Printf("   Key:  %s", *keyFile)

	if *oneShot {
		log.Printf("🏃 One-shot mode: certificates written, exiting")
		return
	}

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("🔄 Running in continuous mode, press Ctrl+C to stop")

	// Block until signal received
	sig := <-sigChan
	log.Printf("Received signal %v, shutting down gracefully", sig)
}
