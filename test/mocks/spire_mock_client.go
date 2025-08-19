// Simple mock client for testing SPIRE functionality without real SPIRE agent
package mocks

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net/url"
	"time"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/svid/x509svid"
)

// MockClient is a simple mock implementation of the SPIRE client for testing
type MockClient struct {
	socketPath      string
	refreshInterval time.Duration
	isClosed        bool
	mockSVID        *x509svid.SVID
	mockTLSConfig   *tls.Config
}

// NewMockClient creates a new mock SPIRE client that doesn't require a real SPIRE agent
func NewMockClient(socketPath string, refreshInterval time.Duration) (*MockClient, error) {
	// Create a mock SVID with test certificate
	mockSVID, err := createMockSVID()
	if err != nil {
		return nil, err
	}

	// Create mock TLS config
	mockTLSConfig := &tls.Config{
		GetCertificate: func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
			// Return the mock certificate
			return &tls.Certificate{
				Certificate: [][]byte{mockSVID.Certificates[0].Raw},
				PrivateKey:  mockSVID.PrivateKey,
			}, nil
		},
	}

	return &MockClient{
		socketPath:      socketPath,
		refreshInterval: refreshInterval,
		isClosed:        false,
		mockSVID:        mockSVID,
		mockTLSConfig:   mockTLSConfig,
	}, nil
}

// GetTLSConfig returns a mock TLS configuration
func (m *MockClient) GetTLSConfig() *tls.Config {
	return m.mockTLSConfig
}

// GetCurrentSVID returns a mock SVID
func (m *MockClient) GetCurrentSVID() (*x509svid.SVID, error) {
	return m.mockSVID, nil
}

// Close marks the mock client as closed
func (m *MockClient) Close() error {
	m.isClosed = true
	return nil
}

// IsClosed returns whether the mock client has been closed
func (m *MockClient) IsClosed() bool {
	return m.isClosed
}

// GetSocketPath returns the socket path for testing
func (m *MockClient) GetSocketPath() string {
	return m.socketPath
}

// GetRefreshInterval returns the refresh interval for testing
func (m *MockClient) GetRefreshInterval() time.Duration {
	return m.refreshInterval
}

// createMockSVID creates a fake SVID for testing
func createMockSVID() (*x509svid.SVID, error) {
	// Generate a test private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	// Create a test certificate
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "test-workload",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour), // Valid for 24 hours
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	// Add SPIFFE ID as SAN
	spiffeID, _ := spiffeid.FromString("spiffe://test.domain/workload")
	template.URIs = []*url.URL{spiffeID.URL()}

	// Create the certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, err
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, err
	}

	// Create and return the SVID
	return &x509svid.SVID{
		ID:           spiffeID,
		Certificates: []*x509.Certificate{cert},
		PrivateKey:   privateKey,
	}, nil
}
