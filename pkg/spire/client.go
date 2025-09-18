// Package spire provides SPIFFE/SPIRE workload API client functionality
// for integrating with Caddy server's TLS configuration.
package spire

import (
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/svid/x509svid"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

// DefaultSpireSocketPath is the standard path for SPIRE agent socket
const DefaultSpireSocketPath = "/tmp/spire-agent/public/api.sock"

// Client represents a SPIFFE/SPIRE workload API client that can
// provide TLS certificates for Caddy server
type Client struct {
	socketPath     string
	source         *workloadapi.X509Source
	refreshTicker  *time.Ticker
	refreshTimeout time.Duration
	ctx            context.Context
	cancel         context.CancelFunc

	// File-based certificate management
	certFile  string       // Path to certificate file
	keyFile   string       // Path to private key file
	fileMode  os.FileMode  // File permissions
	fileMutex sync.RWMutex // Protects file operations
}

// Config holds configuration for the SPIRE client
type Config struct {
	SocketPath      string
	RefreshInterval time.Duration

	// File-based certificate management for Layer 4 integration
	CertFile         string      // Path where certificate chain will be written
	KeyFile          string      // Path where private key will be written
	FileMode         os.FileMode // Permission mode for written files (defaults to 0600)
	RefreshAtPercent int
	RefreshTimeout   time.Duration // Timeout for all certificate operations (startup, refresh, etc.)
}

// NewClient creates a new SPIRE client with the given configuration
func NewClient(config Config) (*Client, error) {
	if config.SocketPath == "" {
		config.SocketPath = DefaultSpireSocketPath
	}

	// Set default refresh timeout if not specified
	if config.RefreshTimeout == 0 {
		config.RefreshTimeout = 10 * time.Second
	}

	// Set default file mode if file-based management is enabled
	if config.CertFile != "" && config.FileMode == 0 {
		config.FileMode = 0600 // Read/write for owner only
	}

	// Validate file-based configuration
	if config.CertFile != "" && config.KeyFile == "" {
		return nil, fmt.Errorf("KeyFile must be specified when CertFile is set")
	}
	if config.KeyFile != "" && config.CertFile == "" {
		return nil, fmt.Errorf("CertFile must be specified when KeyFile is set")
	}

	// Make RefreshInterval and RefreshAtPercent mutually exclusive
	// Default to percentage-based refresh at 65% of certificate lifetime
	if config.RefreshInterval == 0 && config.RefreshAtPercent == 0 {
		config.RefreshAtPercent = 65
	}
	if config.RefreshInterval != 0 && config.RefreshAtPercent != 0 {
		return nil, fmt.Errorf("RefreshInterval and RefreshAtPercent are mutually exclusive")
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Create X509Source for workload API
	source, err := workloadapi.NewX509Source(ctx, workloadapi.WithClientOptions(workloadapi.WithAddr("unix:"+config.SocketPath)))
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create workload API source: %w", err)
	}

	spireClient := &Client{
		socketPath:     config.SocketPath,
		source:         source,
		refreshTimeout: config.RefreshTimeout,
		ctx:            ctx,
		cancel:         cancel,
		certFile:       config.CertFile,
		keyFile:        config.KeyFile,
		fileMode:       config.FileMode,
	}

	// Perform initial certificate fetch at startup to catch attestation issues early
	log.Printf("Fetching initial SPIFFE certificate from %s (timeout: %v)", config.SocketPath, config.RefreshTimeout)
	startupCtx, startupCancel := context.WithTimeout(ctx, config.RefreshTimeout)
	defer startupCancel()

	// Try to get initial SVID with timeout
	svid, err := spireClient.getSVIDWithTimeout(startupCtx)
	if err != nil {
		cancel()
		source.Close()
		return nil, fmt.Errorf("failed to fetch initial SPIFFE certificate: %w", err)
	}

	// Log successful certificate details
	if len(svid.Certificates) > 0 {
		cert := svid.Certificates[0]
		log.Printf("✅ Initial SPIFFE certificate obtained successfully")
		log.Printf("   SPIFFE ID: %s", svid.ID)
		log.Printf("   Valid from: %v", cert.NotBefore)
		log.Printf("   Expires at: %v", cert.NotAfter)
		log.Printf("   Lifetime: %v", cert.NotAfter.Sub(cert.NotBefore))
	}

	// Write initial certificate files if file-based management is enabled
	if spireClient.certFile != "" {
		if err := spireClient.writeCertificateFiles(); err != nil {
			cancel()
			source.Close()
			return nil, fmt.Errorf("failed to write initial certificate files: %w", err)
		}
		log.Printf("📁 Certificate files written: %s, %s", spireClient.certFile, spireClient.keyFile)
	}

	// Start appropriate refresh strategy
	if config.RefreshAtPercent > 0 {
		go spireClient.smartRefresh(config.RefreshAtPercent)
	} else {
		go spireClient.autoRefresh(config.RefreshInterval)
	}

	return spireClient, nil
}

// GetTLSConfig returns a tls.Config that uses SPIFFE SVIDs for certificates
func (c *Client) GetTLSConfig() *tls.Config {
	return tlsconfig.TLSServerConfig(c.source)
}

// Close gracefully shuts down the client
func (c *Client) Close() error {
	c.cancel()
	if c.refreshTicker != nil {
		c.refreshTicker.Stop()
	}
	return c.source.Close()
}

// autoRefresh continuously refreshes SVIDs based on fixed interval (legacy approach)
func (c *Client) autoRefresh(interval time.Duration) {
	c.refreshTicker = time.NewTicker(interval)
	defer c.refreshTicker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-c.refreshTicker.C:
			// The X509Source handles automatic refresh internally
			// We just log the current certificate info (with timeout)
			timeoutCtx, timeoutCancel := context.WithTimeout(c.ctx, c.refreshTimeout)
			svid, err := c.getSVIDWithTimeout(timeoutCtx)
			timeoutCancel()

			if err != nil {
				log.Printf("error getting current SVID: %v", err)
			} else if len(svid.Certificates) > 0 {
				log.Printf("SVID available, expires at %v", svid.Certificates[0].NotAfter)

				// Update certificate files if file-based management is enabled
				if c.certFile != "" {
					if err := c.writeCertificateFiles(); err != nil {
						log.Printf("warning: failed to update certificate files: %v", err)
					} else {
						log.Printf("📁 Certificate files updated: %s, %s", c.certFile, c.keyFile)
					}
				}
			}
		}
	}
}

// smartRefresh schedules certificate refresh based on certificate lifetime percentage
// This is more efficient than polling and aligns with SPIFFE best practices
func (c *Client) smartRefresh(refreshAtPercent int) {
	log.Printf("Starting smart refresh at %d%% of certificate lifetime", refreshAtPercent)

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			// Get current certificate and calculate refresh time (with timeout)
			timeoutCtx, timeoutCancel := context.WithTimeout(c.ctx, c.refreshTimeout)
			svid, err := c.getSVIDWithTimeout(timeoutCtx)
			timeoutCancel()

			if err != nil {
				log.Printf("error getting current SVID for smart refresh: %v", err)
				// Fallback to checking again in 30 seconds if we can't get SVID
				time.Sleep(30 * time.Second)
				continue
			}

			if len(svid.Certificates) == 0 {
				log.Printf("no certificates available, retrying in 30 seconds")
				time.Sleep(30 * time.Second)
				continue
			}

			cert := svid.Certificates[0]
			now := time.Now()
			lifetime := cert.NotAfter.Sub(cert.NotBefore)

			// Calculate when to refresh (at specified percentage of lifetime)
			refreshThreshold := cert.NotBefore.Add(lifetime * time.Duration(refreshAtPercent) / 100)

			if now.After(refreshThreshold) {
				log.Printf("Certificate past refresh threshold (%d%%), refreshing now", refreshAtPercent)
				// Force a refresh by getting a new SVID (with timeout)
				refreshCtx, refreshCancel := context.WithTimeout(c.ctx, c.refreshTimeout)
				_, err := c.getSVIDWithTimeout(refreshCtx)
				refreshCancel()

				if err != nil {
					log.Printf("error during certificate refresh: %v", err)
				} else {
					log.Printf("Certificate refresh completed successfully")

					// Update certificate files if file-based management is enabled
					if c.certFile != "" {
						if err := c.writeCertificateFiles(); err != nil {
							log.Printf("warning: failed to update certificate files: %v", err)
						} else {
							log.Printf("📁 Certificate files updated: %s, %s", c.certFile, c.keyFile)
						}
					}
				}
				// After refresh, recalculate timing for the new certificate
				continue
			}

			// Calculate how long to wait until refresh time
			sleepDuration := refreshThreshold.Sub(now)
			log.Printf("Certificate expires at %v, will refresh at %v (in %v)",
				cert.NotAfter, refreshThreshold, sleepDuration)

			// Sleep until refresh time or context cancellation
			select {
			case <-c.ctx.Done():
				return
			case <-time.After(sleepDuration):
				// Time to refresh - the loop will continue and handle it
			}
		}
	}
}

// getSVIDWithTimeout fetches an SVID with context timeout for any certificate operation
func (c *Client) getSVIDWithTimeout(ctx context.Context) (*x509svid.SVID, error) {
	// Channel to receive the result
	resultChan := make(chan struct {
		svid *x509svid.SVID
		err  error
	}, 1)

	// Start the SVID fetch in a goroutine
	go func() {
		svid, err := c.source.GetX509SVID()
		resultChan <- struct {
			svid *x509svid.SVID
			err  error
		}{svid, err}
	}()

	// Wait for either result or timeout
	select {
	case result := <-resultChan:
		if result.err != nil {
			return nil, fmt.Errorf("SPIRE attestation failed: %w", result.err)
		}
		if result.svid == nil || len(result.svid.Certificates) == 0 {
			return nil, fmt.Errorf("SPIRE returned empty certificate")
		}
		return result.svid, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("timeout waiting for SPIFFE certificate (check SPIRE agent connection and workload registration): %w", ctx.Err())
	}
}

// GetCurrentSVID returns the current SVID from the source
func (c *Client) GetCurrentSVID() (*x509svid.SVID, error) {
	return c.source.GetX509SVID()
}

// GetAllSVIDs returns all available SVIDs from the source
// This is used for multi-attestation scenarios where multiple SPIFFE identities exist
func (c *Client) GetAllSVIDs() ([]*x509svid.SVID, error) {
	// Use the workload API client directly to get all SVIDs
	ctx, cancel := context.WithTimeout(c.ctx, c.refreshTimeout)
	defer cancel()

	client, err := workloadapi.New(ctx, workloadapi.WithAddr("unix:"+c.socketPath))
	if err != nil {
		return nil, fmt.Errorf("failed to create workload API client for SVIDs: %w", err)
	}
	defer client.Close()

	// Fetch all X509 SVIDs
	svidResponse, err := client.FetchX509SVIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch X509 SVIDs: %w", err)
	}

	// Convert the response to a slice of SVIDs
	svids := append([]*x509svid.SVID(nil), svidResponse...)

	if len(svids) == 0 {
		return nil, fmt.Errorf("no SVIDs available")
	}

	return svids, nil
}

// GetSVIDByID returns a specific SVID by its SPIFFE ID
// This enables explicit SPIFFE ID selection for enhanced security and control
func (c *Client) GetSVIDByID(spiffeID string) (*x509svid.SVID, error) {
	if spiffeID == "" {
		return nil, fmt.Errorf("SPIFFE ID cannot be empty")
	}

	// Get all available SVIDs
	svids, err := c.GetAllSVIDs()
	if err != nil {
		return nil, fmt.Errorf("failed to get available SVIDs: %w", err)
	}

	log.Printf("🔍 Looking for specific SPIFFE ID: %s from %d available SVIDs", spiffeID, len(svids))

	// Find SVID that matches the requested SPIFFE ID
	for _, svid := range svids {
		if svid.ID.String() == spiffeID {
			log.Printf("✅ Found matching SVID: %s", spiffeID)
			if len(svid.Certificates) > 0 {
				log.Printf("   DNS names: %v", svid.Certificates[0].DNSNames)
			}
			return svid, nil
		}
	}

	// If not found, log available SVIDs for debugging
	log.Printf("❌ SPIFFE ID not found: %s", spiffeID)
	log.Printf("   Available SVIDs:")
	for _, svid := range svids {
		log.Printf("     - %s", svid.ID)
	}

	return nil, fmt.Errorf("SVID with SPIFFE ID %s not found", spiffeID)
}

// GetSVIDForServerName selects the appropriate SVID based on server name (DNS name)
// This enables multi-attestation scenarios where different sites use different SPIFFE identities
func (c *Client) GetSVIDForServerName(serverName string) (*x509svid.SVID, error) {
	if serverName == "" {
		// If no server name specified, return the default SVID
		return c.GetCurrentSVID()
	}

	// Get all available SVIDs
	svids, err := c.GetAllSVIDs()
	if err != nil {
		return nil, fmt.Errorf("failed to get available SVIDs: %w", err)
	}

	log.Printf("🔍 Selecting SVID for server name: %s from %d available SVIDs", serverName, len(svids))

	// Find SVID that matches the server name in DNS names or CN
	for _, svid := range svids {
		if len(svid.Certificates) == 0 {
			continue
		}

		cert := svid.Certificates[0]

		// Check DNS names in the certificate
		for _, dnsName := range cert.DNSNames {
			if dnsName == serverName {
				log.Printf("✅ Found matching SVID for %s: SPIFFE ID %s", serverName, svid.ID)
				return svid, nil
			}
		}

		// Also check Common Name as fallback
		if cert.Subject.CommonName == serverName {
			log.Printf("✅ Found matching SVID for %s via CN: SPIFFE ID %s", serverName, svid.ID)
			return svid, nil
		}

		// Log available DNS names for debugging
		log.Printf("   SVID %s has DNS names: %v", svid.ID, cert.DNSNames)
	}

	// If no exact match found, log available options and return the first SVID as fallback
	log.Printf("⚠️ No exact SVID match found for server name: %s", serverName)
	log.Printf("   Available SVIDs:")
	for _, svid := range svids {
		if len(svid.Certificates) > 0 {
			log.Printf("     - %s (DNS: %v)", svid.ID, svid.Certificates[0].DNSNames)
		}
	}

	log.Printf("   Using first available SVID as fallback: %s", svids[0].ID)
	return svids[0], nil
}

// GetTrustBundle returns the current trust bundle (CA certificates)
func (c *Client) GetTrustBundle() ([]*x509.Certificate, error) {
	// Use the underlying client context to get bundles
	ctx, cancel := context.WithTimeout(c.ctx, c.refreshTimeout)
	defer cancel()

	// Get bundle using workload API client
	client, err := workloadapi.New(ctx, workloadapi.WithAddr("unix:"+c.socketPath))
	if err != nil {
		return nil, fmt.Errorf("failed to create workload API client for bundles: %w", err)
	}
	defer client.Close()

	bundles, err := client.FetchX509Bundles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch X509 bundles: %w", err)
	}

	// Get the bundle for our trust domain
	var allCerts []*x509.Certificate
	for _, bundle := range bundles.Bundles() {
		allCerts = append(allCerts, bundle.X509Authorities()...)
	}

	if len(allCerts) == 0 {
		return nil, fmt.Errorf("no CA certificates found in trust bundle")
	}

	return allCerts, nil
}

// GetCertificateWithChain returns a TLS certificate with the full certificate chain
// including the SVID certificate and CA certificates from the trust bundle
func (c *Client) GetCertificateWithChain() (*tls.Certificate, error) {
	// Get the current SVID
	svid, err := c.GetCurrentSVID()
	if err != nil {
		return nil, fmt.Errorf("failed to get SVID: %w", err)
	}

	return c.buildCertificateChain(svid)
}

// GetCertificateWithChainForServerName returns a TLS certificate with the full certificate chain
// for a specific server name, enabling multi-attestation scenarios
func (c *Client) GetCertificateWithChainForServerName(serverName string) (*tls.Certificate, error) {
	// Get the appropriate SVID for this server name
	svid, err := c.GetSVIDForServerName(serverName)
	if err != nil {
		return nil, fmt.Errorf("failed to get SVID for server name %s: %w", serverName, err)
	}

	return c.buildCertificateChain(svid)
}

// GetCertificateWithChainByID returns a TLS certificate with the full certificate chain
// for a specific SPIFFE ID, enabling explicit identity selection
func (c *Client) GetCertificateWithChainByID(spiffeID string) (*tls.Certificate, error) {
	// Get the SVID by its SPIFFE ID
	svid, err := c.GetSVIDByID(spiffeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get SVID for SPIFFE ID %s: %w", spiffeID, err)
	}

	return c.buildCertificateChain(svid)
}

// buildCertificateChain builds a TLS certificate chain from an SVID and trust bundle
// This is a helper method used by both GetCertificateWithChain methods
func (c *Client) buildCertificateChain(svid *x509svid.SVID) (*tls.Certificate, error) {
	if len(svid.Certificates) == 0 {
		return nil, fmt.Errorf("SVID contains no certificates")
	}

	// Get the trust bundle (CA certificates)
	trustBundle, err := c.GetTrustBundle()
	if err != nil {
		return nil, fmt.Errorf("failed to get trust bundle: %w", err)
	}

	// Build the certificate chain: [leaf SVID, intermediates..., CA]
	var certChain [][]byte

	// Add the SVID certificates (leaf + any intermediates)
	for _, cert := range svid.Certificates {
		certChain = append(certChain, cert.Raw)
	}

	// Add CA certificates from trust bundle
	// Note: We need to be careful about the order and avoid duplicates
	for _, caCert := range trustBundle {
		// Only add if it's not already in the chain
		isDuplicate := false
		for _, existingCert := range svid.Certificates {
			if existingCert.Equal(caCert) {
				isDuplicate = true
				break
			}
		}
		if !isDuplicate {
			certChain = append(certChain, caCert.Raw)
		}
	}

	// Create the TLS certificate with the full chain
	tlsCert := &tls.Certificate{
		Certificate: certChain,
		PrivateKey:  svid.PrivateKey,
		Leaf:        svid.Certificates[0], // First certificate is always the leaf
	}

	return tlsCert, nil
}

// writeCertificateFiles writes the current SVID certificate chain and private key to disk
// for use by Layer 4 or other file-based certificate consumers
func (c *Client) writeCertificateFiles() error {
	if c.certFile == "" || c.keyFile == "" {
		return fmt.Errorf("certificate and key file paths must be configured")
	}

	// Get current SVID with full chain
	cert, err := c.GetCertificateWithChain()
	if err != nil {
		return fmt.Errorf("failed to get certificate chain: %w", err)
	}

	c.fileMutex.Lock()
	defer c.fileMutex.Unlock()

	// Ensure directories exist
	if err := os.MkdirAll(filepath.Dir(c.certFile), 0755); err != nil {
		return fmt.Errorf("failed to create certificate directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(c.keyFile), 0755); err != nil {
		return fmt.Errorf("failed to create key directory: %w", err)
	}

	// Write certificate chain in PEM format
	certData, err := c.encodeCertificateChain(cert.Certificate)
	if err != nil {
		return fmt.Errorf("failed to encode certificate chain: %w", err)
	}

	if err := c.writeAtomicFile(c.certFile, certData); err != nil {
		return fmt.Errorf("failed to write certificate file: %w", err)
	}

	// Write private key in PEM format
	keyData, err := c.encodePrivateKey(cert.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to encode private key: %w", err)
	}

	if err := c.writeAtomicFile(c.keyFile, keyData); err != nil {
		return fmt.Errorf("failed to write key file: %w", err)
	}

	return nil
}

// encodeCertificateChain encodes a certificate chain as PEM data
func (c *Client) encodeCertificateChain(certChain [][]byte) ([]byte, error) {
	var pemData []byte

	for i, certDER := range certChain {
		block := &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: certDER,
		}

		// Add a comment for the first certificate (leaf)
		if i == 0 {
			block.Headers = map[string]string{
				"Subject": "Leaf Certificate (SPIRE SVID)",
			}
		}

		pemData = append(pemData, pem.EncodeToMemory(block)...)
	}

	return pemData, nil
}

// encodePrivateKey encodes a private key as PEM data
func (c *Client) encodePrivateKey(privateKey interface{}) ([]byte, error) {
	switch key := privateKey.(type) {
	case *rsa.PrivateKey:
		return pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		}), nil
	case *ecdsa.PrivateKey:
		keyBytes, err := x509.MarshalECPrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal EC private key: %w", err)
		}
		return pem.EncodeToMemory(&pem.Block{
			Type:  "EC PRIVATE KEY",
			Bytes: keyBytes,
		}), nil
	default:
		// Try PKCS#8 format as fallback
		keyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
		if err != nil {
			return nil, fmt.Errorf("unsupported private key type: %T", privateKey)
		}
		return pem.EncodeToMemory(&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: keyBytes,
		}), nil
	}
}

// writeAtomicFile writes data to a file atomically by writing to a temporary file
// and then renaming it to the target file
func (c *Client) writeAtomicFile(filename string, data []byte) error {
	// Create temporary file in the same directory
	tempFile := filename + ".tmp"

	// Write to temporary file
	if err := os.WriteFile(tempFile, data, c.fileMode); err != nil {
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	// Atomically rename to target file
	if err := os.Rename(tempFile, filename); err != nil {
		os.Remove(tempFile) // Clean up temp file on error
		return fmt.Errorf("failed to rename temporary file: %w", err)
	}

	return nil
}
