// Package spire provides SPIFFE/SPIRE workload API client functionality
// for integrating with Caddy server's TLS configuration.
package spire

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
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
}

// Config holds configuration for the SPIRE client
type Config struct {
	SocketPath       string
	RefreshInterval  time.Duration
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
