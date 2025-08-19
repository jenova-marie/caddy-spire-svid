// Test for the mock client to ensure our test framework works
package mocks

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMockClient validates that our mock client works properly
func TestMockClient(t *testing.T) {
	t.Run("Create_MockClient_Success", func(t *testing.T) {
		mockClient, err := NewMockClient("/test/socket", 30*time.Second)

		require.NoError(t, err, "Should create mock client without error")
		assert.NotNil(t, mockClient, "Mock client should not be nil")
		assert.Equal(t, "/test/socket", mockClient.GetSocketPath())
		assert.Equal(t, 30*time.Second, mockClient.GetRefreshInterval())
		assert.False(t, mockClient.IsClosed(), "Mock client should not be closed initially")
	})

	t.Run("GetTLSConfig_Success", func(t *testing.T) {
		mockClient, err := NewMockClient("/test/socket", 30*time.Second)
		require.NoError(t, err)

		tlsConfig := mockClient.GetTLSConfig()
		assert.NotNil(t, tlsConfig, "TLS config should not be nil")
		assert.NotNil(t, tlsConfig.GetCertificate, "TLS config should have GetCertificate function")

		// Test that GetCertificate actually works
		cert, err := tlsConfig.GetCertificate(nil)
		assert.NoError(t, err, "Should get certificate without error")
		assert.NotNil(t, cert, "Certificate should not be nil")
		assert.NotEmpty(t, cert.Certificate, "Certificate should have certificate data")
	})

	t.Run("GetCurrentSVID_Success", func(t *testing.T) {
		mockClient, err := NewMockClient("/test/socket", 30*time.Second)
		require.NoError(t, err)

		svid, err := mockClient.GetCurrentSVID()
		assert.NoError(t, err, "Should get SVID without error")
		assert.NotNil(t, svid, "SVID should not be nil")
		assert.NotNil(t, svid.ID, "SVID should have an ID")
		assert.NotEmpty(t, svid.Certificates, "SVID should have certificates")
		assert.NotNil(t, svid.PrivateKey, "SVID should have a private key")

		// Verify SPIFFE ID
		assert.Equal(t, "spiffe://test.domain/workload", svid.ID.String())

		// Verify certificate is valid
		cert := svid.Certificates[0]
		assert.True(t, time.Now().After(cert.NotBefore), "Certificate should be valid now")
		assert.True(t, time.Now().Before(cert.NotAfter), "Certificate should not be expired")
	})

	t.Run("Close_Success", func(t *testing.T) {
		mockClient, err := NewMockClient("/test/socket", 30*time.Second)
		require.NoError(t, err)

		assert.False(t, mockClient.IsClosed(), "Should not be closed initially")

		err = mockClient.Close()
		assert.NoError(t, err, "Should close without error")
		assert.True(t, mockClient.IsClosed(), "Should be closed after Close()")
	})
}

// TestMockClientBehavior tests that our mock behaves like a real client would
func TestMockClientBehavior(t *testing.T) {
	t.Run("Multiple_GetCurrentSVID_Calls", func(t *testing.T) {
		mockClient, err := NewMockClient("/test/socket", 30*time.Second)
		require.NoError(t, err)

		// Multiple calls should return the same SVID
		svid1, err1 := mockClient.GetCurrentSVID()
		svid2, err2 := mockClient.GetCurrentSVID()

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.Equal(t, svid1.ID.String(), svid2.ID.String(), "Should return consistent SVID ID")
	})

	t.Run("Multiple_GetTLSConfig_Calls", func(t *testing.T) {
		mockClient, err := NewMockClient("/test/socket", 30*time.Second)
		require.NoError(t, err)

		// Multiple calls should return working TLS configs
		config1 := mockClient.GetTLSConfig()
		config2 := mockClient.GetTLSConfig()

		assert.NotNil(t, config1)
		assert.NotNil(t, config2)

		// Both should be able to get certificates
		cert1, err1 := config1.GetCertificate(nil)
		cert2, err2 := config2.GetCertificate(nil)

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NotNil(t, cert1)
		assert.NotNil(t, cert2)
	})
}
