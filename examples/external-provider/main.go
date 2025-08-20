// Package main demonstrates Method 2: External Certificate Provider
// This approach runs caddy-spire-svid as a separate service that provides
// certificates to Caddy via shared file storage.
package main

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/jenova-marie/caddy-spire-svid/pkg/spire"
)

const (
	// Default paths for certificate files
	defaultCertDir  = "/tmp/caddy-spire-certs"
	defaultCertFile = "cert.pem"
	defaultKeyFile  = "key.pem"
)

func main() {
	// Create certificate directory
	certDir := os.Getenv("CERT_DIR")
	if certDir == "" {
		certDir = defaultCertDir
	}

	if err := os.MkdirAll(certDir, 0755); err != nil {
		log.Fatalf("Failed to create certificate directory: %v", err)
	}

	log.Printf("🌸 Starting External Certificate Provider")
	log.Printf("📁 Certificate directory: %s", certDir)

	// Create SPIRE client
	config := spire.Config{
		SocketPath:      "/tmp/spire-agent/public/api.sock",
		RefreshInterval: 30 * time.Second,
	}

	client, err := spire.NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create SPIRE client: %v", err)
	}
	defer client.Close()

	// Write initial certificates
	if err := writeCertificates(client, certDir); err != nil {
		log.Fatalf("Failed to write initial certificates: %v", err)
	}

	log.Printf("✨ Initial certificates written successfully")
	log.Printf("💖 External provider running... monitoring for certificate updates")

	// Monitor for certificate updates
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := writeCertificates(client, certDir); err != nil {
				log.Printf("❌ Error updating certificates: %v", err)
			} else {
				log.Printf("🔄 Certificates updated successfully")
			}
		}
	}
}

func writeCertificates(client *spire.Client, certDir string) error {
	// Get current SVID
	svid, err := client.GetCurrentSVID()
	if err != nil {
		return err
	}

	if len(svid.Certificates) == 0 {
		return fmt.Errorf("no certificates available")
	}

	// Write certificate file
	certPath := filepath.Join(certDir, defaultCertFile)
	certFile, err := os.Create(certPath)
	if err != nil {
		return err
	}
	defer certFile.Close()

	// Write certificate chain
	for _, cert := range svid.Certificates {
		if err := pem.Encode(certFile, &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert.Raw,
		}); err != nil {
			return err
		}
	}

	// Write private key file
	keyPath := filepath.Join(certDir, defaultKeyFile)
	keyFile, err := os.Create(keyPath)
	if err != nil {
		return err
	}
	defer keyFile.Close()

	// Encode private key
	keyBytes, err := x509.MarshalPKCS8PrivateKey(svid.PrivateKey)
	if err != nil {
		return err
	}

	if err := pem.Encode(keyFile, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: keyBytes,
	}); err != nil {
		return err
	}

	log.Printf("📝 Certificates written to %s and %s", certPath, keyPath)
	log.Printf("🔐 SPIFFE ID: %s", svid.ID)
	log.Printf("⏰ Expires: %v", svid.Certificates[0].NotAfter)

	return nil
}
