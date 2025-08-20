// Test file for SPIRE client - Unit tests
package spire

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jenova-marie/caddy-spire-svid/test/mocks"
)

// TestNewClientBasic tests that we can create a real client (will fail without SPIRE agent, which is expected)
func TestNewClientBasic(t *testing.T) {
	// Skip this test in Phase 1 - it hangs without real SPIRE agent
	t.Skip("Skipping real client test in Phase 1 - would hang without SPIRE agent")
}

// TestConfigDefaults verifies that default configuration values work
func TestConfigDefaults(t *testing.T) {
	assert.Equal(t, "/tmp/spire-agent/public/api.sock", DefaultSpireSocketPath)
}

// TestClientInterface tests that our real client implements the expected interface
func TestClientInterface(t *testing.T) {
	// This is Phase 1 - we're just testing that our interface works conceptually
	// We'll use our mock to validate the interface without needing real SPIRE

	mockClient, err := mocks.NewMockClient("/test/socket", 30*time.Second)
	require.NoError(t, err, "Mock client should be created successfully")

	// Test that mock implements the interface we expect real client to have
	t.Run("GetTLSConfig", func(t *testing.T) {
		config := mockClient.GetTLSConfig()
		assert.NotNil(t, config, "Should return TLS config")
		assert.NotNil(t, config.GetCertificate, "TLS config should have GetCertificate")
	})

	t.Run("GetCurrentSVID", func(t *testing.T) {
		svid, err := mockClient.GetCurrentSVID()
		assert.NoError(t, err, "Should get SVID")
		assert.NotNil(t, svid, "SVID should not be nil")
		assert.NotEmpty(t, svid.Certificates, "SVID should have certificates")
	})

	t.Run("Close", func(t *testing.T) {
		err := mockClient.Close()
		assert.NoError(t, err, "Should close successfully")
		assert.True(t, mockClient.IsClosed(), "Should be marked as closed")
	})
}

// TestPhase1_FrameworkValidation is our main Phase 1 test
func TestPhase1_FrameworkValidation(t *testing.T) {
	t.Log("🌸 Phase 1: Testing Framework Validation")

	// Test 1: Mock client creation works
	t.Run("MockClient_Creation", func(t *testing.T) {
		mockClient, err := mocks.NewMockClient("/test/socket", 45*time.Second)
		require.NoError(t, err, "Should create mock client")
		defer mockClient.Close()

		// Verify properties
		assert.Equal(t, "/test/socket", mockClient.GetSocketPath())
		assert.Equal(t, 45*time.Second, mockClient.GetRefreshInterval())
	})

	// Test 2: Mock provides working TLS config
	t.Run("MockClient_TLSConfig", func(t *testing.T) {
		mockClient, err := mocks.NewMockClient("/test", 30*time.Second)
		require.NoError(t, err)
		defer mockClient.Close()

		config := mockClient.GetTLSConfig()
		require.NotNil(t, config)

		// Test that we can actually get a certificate
		cert, err := config.GetCertificate(nil)
		assert.NoError(t, err)
		assert.NotNil(t, cert)
		assert.NotEmpty(t, cert.Certificate)
	})

	// Test 3: Mock provides valid SVID
	t.Run("MockClient_SVID", func(t *testing.T) {
		mockClient, err := mocks.NewMockClient("/test", 30*time.Second)
		require.NoError(t, err)
		defer mockClient.Close()

		svid, err := mockClient.GetCurrentSVID()
		assert.NoError(t, err)
		assert.NotNil(t, svid)
		assert.Equal(t, "spiffe://test.domain/workload", svid.ID.String())
		assert.NotEmpty(t, svid.Certificates)

		// Verify certificate is valid
		cert := svid.Certificates[0]
		now := time.Now()
		assert.True(t, now.After(cert.NotBefore), "Certificate should be valid now")
		assert.True(t, now.Before(cert.NotAfter), "Certificate should not be expired")
	})

	t.Log("✅ Phase 1: Framework validation complete - ready for Phase 2!")
}
