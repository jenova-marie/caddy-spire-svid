// Package caddyspire implements a Caddy module for SPIFFE/SPIRE certificate automation.
// This module integrates with Caddy's TLS automation system to provide automatic
// certificate management using SPIFFE identities from a SPIRE agent.
package caddyspire

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
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
	// Defaults to 30 seconds.
	RefreshInterval caddy.Duration `json:"refresh_interval,omitempty"`

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
	if sm.RefreshInterval == 0 {
		sm.RefreshInterval = caddy.Duration(30 * time.Second)
	}

	// Check environment variables for file-based certificate management
	certFile := os.Getenv("SPIRE_CERT_FILE")
	keyFile := os.Getenv("SPIRE_KEY_FILE")

	sm.logger.Info("🌸 Provisioning SPIRE certificate manager",
		zap.String("socket_path", sm.SocketPath),
		zap.Duration("refresh_interval", time.Duration(sm.RefreshInterval)),
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
		zap.String("server_name", hello.ServerName))

	// Lazy initialization: create SPIRE client on first certificate request
	if sm.client == nil {
		sm.logger.Info("🔗 Connecting to SPIRE agent for first time",
			zap.String("socket_path", sm.SocketPath))

		config := spire.Config{
			SocketPath:      sm.SocketPath,
			RefreshInterval: time.Duration(sm.RefreshInterval),
			CertFile:        os.Getenv("SPIRE_CERT_FILE"),
			KeyFile:         os.Getenv("SPIRE_KEY_FILE"),
		}

		client, err := spire.NewClient(config)
		if err != nil {
			sm.logger.Error("Failed to connect to SPIRE agent", zap.Error(err))
			return nil, fmt.Errorf("failed to create SPIRE client: %w", err)
		}

		sm.client = client
		sm.logger.Info("🎉 Successfully connected to SPIRE agent")
	}

	// Get certificate with full chain (SVID + intermediates + CA)
	cert, err := sm.client.GetCertificateWithChain()
	if err != nil {
		sm.logger.Error("Failed to get certificate chain from SPIRE", zap.Error(err))
		return nil, fmt.Errorf("failed to get certificate chain from SPIRE: %w", err)
	}

	// Log certificate chain information for debugging
	sm.logger.Debug("🎉 Successfully got certificate chain via SPIRE",
		zap.String("server_name", hello.ServerName),
		zap.Int("chain_length", len(cert.Certificate)),
		zap.String("leaf_subject", cert.Leaf.Subject.String()))

	return cert, nil
}

// Interface guards to ensure we implement the required interfaces
var (
	_ caddy.Provisioner     = (*SpireManager)(nil)
	_ caddy.CleanerUpper    = (*SpireManager)(nil)
	_ caddyfile.Unmarshaler = (*SpireManager)(nil)
)
