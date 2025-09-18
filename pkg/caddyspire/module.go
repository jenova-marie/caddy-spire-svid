// Package caddyspire implements a Caddy module for SPIFFE/SPIRE certificate automation.
// This module integrates with Caddy's TLS automation system to provide automatic
// certificate management using SPIFFE identities from a SPIRE agent.
package caddyspire

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"strconv"
	"time"

	caddy "github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"go.uber.org/zap"

	"github.com/jenova-marie/caddy-spire-svid/pkg/spire"
)

func init() {
	caddy.RegisterModule(SpireManager{})
}

// SpireManager implements a Caddy certificate manager for SPIFFE/SPIRE integration.
// It provides automatic certificate management using SPIFFE identities.
// File-based certificate management is enabled via environment variables:
// SPIRE_CERT_FILE and SPIRE_KEY_FILE for Layer 4 integration.
type SpireManager struct {
	// SocketPath is the path to the SPIRE agent socket.
	// Defaults to "/tmp/spire-agent/public/api.sock"
	SocketPath string `json:"socket_path,omitempty"`

	// RefreshInterval controls how often to check for certificate updates.
	// Defaults to 30 seconds. Mutually exclusive with RefreshAtPercent.
	RefreshInterval caddy.Duration `json:"refresh_interval,omitempty"`

	// RefreshAtPercent controls when to refresh certificates based on lifetime percentage.
	// For example, 65 means refresh when certificate has reached 65% of its lifetime.
	// Defaults to 65%. Mutually exclusive with RefreshInterval.
	RefreshAtPercent int `json:"refresh_at_percent,omitempty"`

	// SpiffeID explicitly specifies which SPIFFE ID to use for this server.
	// When set, only the SVID matching this SPIFFE ID will be used.
	// When empty, the system automatically selects based on DNS names.
	// Example: "spiffe://recoverysky.org/prod/metis/caddy-loki"
	SpiffeID string `json:"spiffe_id,omitempty"`

	// Logger for this module
	logger *zap.Logger

	// SPIRE client instance
	client *spire.Client
}

// CaddyModule returns the Caddy module information.
func (SpireManager) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "tls.get_certificate.spire",
		New: func() caddy.Module { return new(SpireManager) },
	}
}

// Provision sets up the SPIRE manager module.
func (sm *SpireManager) Provision(ctx caddy.Context) error {
	sm.logger = ctx.Logger(sm)

	// Set defaults
	if sm.SocketPath == "" {
		sm.SocketPath = spire.DefaultSpireSocketPath
	}
	// Only set default RefreshInterval if neither interval nor percentage is specified
	if sm.RefreshInterval == 0 && sm.RefreshAtPercent == 0 {
		sm.RefreshInterval = caddy.Duration(30 * time.Second)
	}

	// Check environment variables for file-based certificate management
	certFile := os.Getenv("SPIRE_CERT_FILE")
	keyFile := os.Getenv("SPIRE_KEY_FILE")

	sm.logger.Info("🌸 Provisioning SPIRE certificate manager",
		zap.String("socket_path", sm.SocketPath),
		zap.Duration("refresh_interval", time.Duration(sm.RefreshInterval)),
		zap.Int("refresh_at_percent", sm.RefreshAtPercent),
		zap.String("spiffe_id", sm.SpiffeID),
		zap.String("cert_file_env", certFile),
		zap.String("key_file_env", keyFile))

	// Log file-based management status
	if certFile != "" && keyFile != "" {
		sm.logger.Info("📁 File-based certificate management enabled for Layer 4 integration",
			zap.String("cert_file", certFile),
			zap.String("key_file", keyFile))
	} else if certFile != "" || keyFile != "" {
		sm.logger.Warn("⚠️ Incomplete file-based configuration: both SPIRE_CERT_FILE and SPIRE_KEY_FILE must be set")
	}

	// Don't create SPIRE client during provisioning to avoid blocking Caddy startup
	// The client will be created lazily when first certificate is requested
	sm.logger.Info("✨ SPIRE certificate manager ready (client will connect on first certificate request)")

	return nil
}

// Cleanup shuts down the SPIRE manager module.
func (sm *SpireManager) Cleanup() error {
	if sm.client != nil {
		sm.logger.Info("💖 Shutting down SPIRE certificate manager")
		return sm.client.Close()
	}
	return nil
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler to enable Caddyfile configuration.
func (sm *SpireManager) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		for d.NextBlock(0) {
			switch d.Val() {
			case "socket_path":
				if !d.NextArg() {
					return d.ArgErr()
				}
				sm.SocketPath = d.Val()

			case "refresh_interval":
				if !d.NextArg() {
					return d.ArgErr()
				}
				dur, err := time.ParseDuration(d.Val())
				if err != nil {
					return d.Errf("invalid duration: %v", err)
				}
				sm.RefreshInterval = caddy.Duration(dur)

			case "refresh_at_percent":
				if !d.NextArg() {
					return d.ArgErr()
				}
				var err error
				sm.RefreshAtPercent, err = strconv.Atoi(d.Val())
				if err != nil || sm.RefreshAtPercent < 1 || sm.RefreshAtPercent > 99 {
					return d.Errf("refresh_at_percent must be between 1 and 99, got: %s", d.Val())
				}

			case "spiffe_id":
				if !d.NextArg() {
					return d.ArgErr()
				}
				sm.SpiffeID = d.Val()

			default:
				return d.Errf("unknown subdirective: %s", d.Val())
			}
		}
	}
	return nil
}

// GetCertificate implements the GetCertificate interface for Caddy's TLS automation.
// This function is called by Caddy when it needs a certificate.
func (sm *SpireManager) GetCertificate(ctx context.Context, hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	sm.logger.Debug("🔐 Getting certificate via SPIRE",
		zap.String("server_name", hello.ServerName),
		zap.String("configured_spiffe_id", sm.SpiffeID))

	// Lazy initialization: create SPIRE client on first certificate request
	if sm.client == nil {
		sm.logger.Info("🔗 Connecting to SPIRE agent for first time",
			zap.String("socket_path", sm.SocketPath))

		config := spire.Config{
			SocketPath:       sm.SocketPath,
			RefreshInterval:  time.Duration(sm.RefreshInterval),
			RefreshAtPercent: sm.RefreshAtPercent,
			CertFile:         os.Getenv("SPIRE_CERT_FILE"),
			KeyFile:          os.Getenv("SPIRE_KEY_FILE"),
		}

		client, err := spire.NewClient(config)
		if err != nil {
			sm.logger.Error("Failed to connect to SPIRE agent", zap.Error(err))
			return nil, fmt.Errorf("failed to create SPIRE client: %w", err)
		}

		sm.client = client
		sm.logger.Info("🎉 Successfully connected to SPIRE agent")
	}

	var cert *tls.Certificate
	var err error

	// If a specific SPIFFE ID is configured, use it exclusively
	if sm.SpiffeID != "" {
		sm.logger.Debug("Using explicitly configured SPIFFE ID",
			zap.String("spiffe_id", sm.SpiffeID))
		cert, err = sm.client.GetCertificateWithChainByID(sm.SpiffeID)
		if err != nil {
			sm.logger.Error("Failed to get certificate chain from SPIRE for configured SPIFFE ID",
				zap.String("spiffe_id", sm.SpiffeID),
				zap.Error(err))
			return nil, fmt.Errorf("failed to get certificate chain from SPIRE for SPIFFE ID %s: %w", sm.SpiffeID, err)
		}
	} else {
		// Fall back to automatic selection based on server name (multi-attestation support)
		// This enables automatic SVID selection based on DNS names in SPIRE entries
		cert, err = sm.client.GetCertificateWithChainForServerName(hello.ServerName)
		if err != nil {
			sm.logger.Error("Failed to get certificate chain from SPIRE for server name",
				zap.String("server_name", hello.ServerName),
				zap.Error(err))
			return nil, fmt.Errorf("failed to get certificate chain from SPIRE for %s: %w", hello.ServerName, err)
		}
	}

	// Log certificate chain information for debugging
	sm.logger.Debug("🎉 Successfully got certificate chain via SPIRE",
		zap.String("server_name", hello.ServerName),
		zap.String("used_spiffe_id", sm.SpiffeID),
		zap.Int("chain_length", len(cert.Certificate)),
		zap.String("leaf_subject", cert.Leaf.Subject.String()),
		zap.Strings("leaf_dns_names", cert.Leaf.DNSNames))

	return cert, nil
}

// Interface guards to ensure we implement the required interfaces
var (
	_ caddy.Provisioner     = (*SpireManager)(nil)
	_ caddy.CleanerUpper    = (*SpireManager)(nil)
	_ caddyfile.Unmarshaler = (*SpireManager)(nil)
)
