// Package main demonstrates Method 5: Helper Tools / File-based Legacy Support
// This approach uses file-based certificate management compatible with legacy
// systems and standard certificate workflows.
package main

import (
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/jenova-marie/caddy-spire-client/pkg/spire"
)

const (
	defaultCertDir    = "/etc/ssl/spire"
	defaultCertFile   = "cert.pem"
	defaultKeyFile    = "key.pem"
	defaultBundleFile = "bundle.pem"
)

var (
	certDir    = flag.String("cert-dir", defaultCertDir, "Directory to write certificates")
	certFile   = flag.String("cert-file", defaultCertFile, "Certificate filename")
	keyFile    = flag.String("key-file", defaultKeyFile, "Private key filename")
	bundleFile = flag.String("bundle-file", defaultBundleFile, "Trust bundle filename")
	socketPath = flag.String("socket", "/tmp/spire-agent/public/api.sock", "SPIRE agent socket path")
	interval   = flag.Duration("interval", 30*time.Second, "Certificate refresh interval")
	daemon     = flag.Bool("daemon", false, "Run as daemon (continuous mode)")
	once       = flag.Bool("once", false, "Fetch certificates once and exit")
	verbose    = flag.Bool("verbose", false, "Enable verbose logging")
	help       = flag.Bool("help", false, "Show help")
)

type HelperConfig struct {
	CertDir    string
	CertFile   string
	KeyFile    string
	BundleFile string
	SocketPath string
	Interval   time.Duration
	Daemon     bool
	Once       bool
	Verbose    bool
}

func main() {
	flag.Parse()

	if *help {
		printHelp()
		return
	}

	config := HelperConfig{
		CertDir:    *certDir,
		CertFile:   *certFile,
		KeyFile:    *keyFile,
		BundleFile: *bundleFile,
		SocketPath: *socketPath,
		Interval:   *interval,
		Daemon:     *daemon,
		Once:       *once,
		Verbose:    *verbose,
	}

	if *verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	log.Printf("🌸 SPIRE Helper Tool starting...")
	log.Printf("📁 Certificate directory: %s", config.CertDir)
	log.Printf("🔌 SPIRE socket: %s", config.SocketPath)

	// Create certificate directory
	if err := os.MkdirAll(config.CertDir, 0755); err != nil {
		log.Fatalf("❌ Failed to create certificate directory: %v", err)
	}

	// Create SPIRE client
	spireConfig := spire.Config{
		SocketPath:      config.SocketPath,
		RefreshInterval: config.Interval,
	}

	client, err := spire.NewClient(spireConfig)
	if err != nil {
		log.Fatalf("❌ Failed to create SPIRE client: %v", err)
	}
	defer client.Close()

	// Write initial certificates
	if err := writeCertificates(client, config); err != nil {
		log.Fatalf("❌ Failed to write initial certificates: %v", err)
	}

	log.Printf("✅ Initial certificates written successfully")

	if *once {
		log.Printf("💖 One-time certificate fetch complete")
		return
	}

	if *daemon {
		runDaemon(client, config)
	} else {
		log.Printf("💖 Certificate fetch complete. Use -daemon for continuous updates.")
	}
}

func runDaemon(client *spire.Client, config HelperConfig) {
	log.Printf("🚀 Running in daemon mode, checking every %v", config.Interval)

	// Handle signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	ticker := time.NewTicker(config.Interval)
	defer ticker.Stop()

	for {
		select {
		case sig := <-sigChan:
			switch sig {
			case syscall.SIGHUP:
				log.Printf("🔄 Received SIGHUP, refreshing certificates immediately")
				if err := writeCertificates(client, config); err != nil {
					log.Printf("❌ Error refreshing certificates: %v", err)
				} else {
					log.Printf("✅ Certificates refreshed on signal")
				}
			case syscall.SIGINT, syscall.SIGTERM:
				log.Printf("💖 Received shutdown signal, exiting gracefully...")
				return
			}
		case <-ticker.C:
			if config.Verbose {
				log.Printf("🔄 Checking for certificate updates...")
			}
			if err := writeCertificates(client, config); err != nil {
				log.Printf("❌ Error updating certificates: %v", err)
			} else if config.Verbose {
				log.Printf("✅ Certificates updated successfully")
			}
		}
	}
}

func writeCertificates(client *spire.Client, config HelperConfig) error {
	// Get current SVID
	svid, err := client.GetCurrentSVID()
	if err != nil {
		return fmt.Errorf("failed to get SVID: %w", err)
	}

	if len(svid.Certificates) == 0 {
		return fmt.Errorf("no certificates available")
	}

	// Write certificate file
	certPath := filepath.Join(config.CertDir, config.CertFile)
	if err := writeCertificateFile(certPath, svid.Certificates); err != nil {
		return fmt.Errorf("failed to write certificate file: %w", err)
	}

	// Write private key file
	keyPath := filepath.Join(config.CertDir, config.KeyFile)
	if err := writePrivateKeyFile(keyPath, svid.PrivateKey); err != nil {
		return fmt.Errorf("failed to write private key file: %w", err)
	}

	// Write trust bundle (if available)
	// Note: In a real implementation, you'd get this from SPIRE's trust bundle API
	bundlePath := filepath.Join(config.CertDir, config.BundleFile)
	if err := writeTrustBundle(bundlePath, svid.Certificates); err != nil {
		return fmt.Errorf("failed to write trust bundle: %w", err)
	}

	if config.Verbose {
		cert := svid.Certificates[0]
		log.Printf("📝 Certificate files updated:")
		log.Printf("  🔐 Certificate: %s", certPath)
		log.Printf("  🗝️  Private Key: %s", keyPath)
		log.Printf("  📦 Trust Bundle: %s", bundlePath)
		log.Printf("  🆔 SPIFFE ID: %s", svid.ID)
		log.Printf("  ⏰ Expires: %v", cert.NotAfter)
		log.Printf("  ⌛ Time until expiry: %v", time.Until(cert.NotAfter))
	}

	return nil
}

func writeCertificateFile(path string, certs []*x509.Certificate) error {
	// Create temporary file for atomic write
	tmpPath := path + ".tmp"
	file, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	defer os.Remove(tmpPath) // Clean up on error

	// Write certificate chain
	for _, cert := range certs {
		if err := pem.Encode(file, &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert.Raw,
		}); err != nil {
			file.Close()
			return err
		}
	}

	if err := file.Close(); err != nil {
		return err
	}

	// Atomic move
	return os.Rename(tmpPath, path)
}

func writePrivateKeyFile(path string, privateKey interface{}) error {
	// Create temporary file for atomic write
	tmpPath := path + ".tmp"
	file, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	defer os.Remove(tmpPath) // Clean up on error

	// Set restrictive permissions for private key
	if err := file.Chmod(0600); err != nil {
		file.Close()
		return err
	}

	// Encode private key
	keyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		file.Close()
		return err
	}

	if err := pem.Encode(file, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: keyBytes,
	}); err != nil {
		file.Close()
		return err
	}

	if err := file.Close(); err != nil {
		return err
	}

	// Atomic move
	return os.Rename(tmpPath, path)
}

func writeTrustBundle(path string, certs []*x509.Certificate) error {
	// For demonstration, we'll write the CA certificates from the chain
	// In a real implementation, you'd fetch the trust bundle from SPIRE
	tmpPath := path + ".tmp"
	file, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	defer os.Remove(tmpPath) // Clean up on error

	// Write CA certificates (skip the first one which is the leaf)
	for i := 1; i < len(certs); i++ {
		if err := pem.Encode(file, &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: certs[i].Raw,
		}); err != nil {
			file.Close()
			return err
		}
	}

	if err := file.Close(); err != nil {
		return err
	}

	// Atomic move
	return os.Rename(tmpPath, path)
}

func printHelp() {
	fmt.Printf("🌸 SPIRE Helper Tool - File-based Certificate Management\n\n")
	fmt.Printf("USAGE:\n")
	fmt.Printf("  %s [OPTIONS]\n\n", os.Args[0])
	fmt.Printf("DESCRIPTION:\n")
	fmt.Printf("  Fetches SPIFFE certificates from SPIRE agent and writes them to files\n")
	fmt.Printf("  for use with legacy applications and standard certificate workflows.\n\n")
	fmt.Printf("OPTIONS:\n")
	flag.PrintDefaults()
	fmt.Printf("\nEXAMPLES:\n")
	fmt.Printf("  # Fetch certificates once\n")
	fmt.Printf("  %s -once\n\n", os.Args[0])
	fmt.Printf("  # Run as daemon with custom directory\n")
	fmt.Printf("  %s -daemon -cert-dir /etc/ssl/myapp\n\n", os.Args[0])
	fmt.Printf("  # Fetch with verbose logging\n")
	fmt.Printf("  %s -verbose -once\n\n", os.Args[0])
	fmt.Printf("  # Custom file names\n")
	fmt.Printf("  %s -cert-file server.crt -key-file server.key\n\n", os.Args[0])
	fmt.Printf("SIGNALS (daemon mode):\n")
	fmt.Printf("  SIGHUP  - Refresh certificates immediately\n")
	fmt.Printf("  SIGTERM - Graceful shutdown\n")
	fmt.Printf("  SIGINT  - Graceful shutdown\n\n")
	fmt.Printf("For more information: https://github.com/jenova-marie/caddy-spire-client\n")
}
