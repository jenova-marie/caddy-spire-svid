// Table-driven tests for SPIRE MockClient scenarios
package spire

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jenova-marie/caddy-spire-client/test/mocks"
)

// TestMockClientScenarios runs table-driven tests for MockClient behavior
func TestMockClientScenarios(t *testing.T) {
	scenarios := []struct {
		name            string
		socketPath      string
		refreshInterval time.Duration
		expectError     bool
		expectedSocket  string
	}{
		{
			name:            "default_config_empty_values",
			socketPath:      "",
			refreshInterval: 0,
			expectError:     false,
			expectedSocket:  "", // MockClient will use whatever we pass
		},
		{
			name:            "custom_socket_path",
			socketPath:      "/custom/test/socket",
			refreshInterval: 30 * time.Second,
			expectError:     false,
			expectedSocket:  "/custom/test/socket",
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			t.Logf("🧪 Testing scenario: %s", scenario.name)

			// Create fresh mock for each scenario (better isolation)
			mockClient, err := mocks.NewMockClient(scenario.socketPath, scenario.refreshInterval)

			// Check creation result
			if scenario.expectError {
				assert.Error(t, err, "Expected error for scenario: %s", scenario.name)
				return
			}

			require.NoError(t, err, "Should create mock client successfully for: %s", scenario.name)
			require.NotNil(t, mockClient, "Mock client should not be nil for: %s", scenario.name)

			// Defer cleanup
			defer func() {
				if mockClient != nil {
					mockClient.Close()
				}
			}()

			// Test core properties
			assert.Equal(t, scenario.expectedSocket, mockClient.GetSocketPath(),
				"Socket path should match for: %s", scenario.name)
			assert.Equal(t, scenario.refreshInterval, mockClient.GetRefreshInterval(),
				"Refresh interval should match for: %s", scenario.name)
			assert.False(t, mockClient.IsClosed(),
				"Client should not be closed initially for: %s", scenario.name)

			t.Logf("✅ Basic properties validated for: %s", scenario.name)
		})
	}
}

// TestMockClientCoreMethodsScenarios tests the core methods across different configurations
func TestMockClientCoreMethodsScenarios(t *testing.T) {
	scenarios := []struct {
		name               string
		socketPath         string
		refreshInterval    time.Duration
		testGetTLSConfig   bool
		testGetCurrentSVID bool
		testClose          bool
	}{
		{
			name:               "default_config_all_methods",
			socketPath:         "",
			refreshInterval:    0,
			testGetTLSConfig:   true,
			testGetCurrentSVID: true,
			testClose:          true,
		},
		{
			name:               "custom_socket_all_methods",
			socketPath:         "/test/custom/socket",
			refreshInterval:    45 * time.Second,
			testGetTLSConfig:   true,
			testGetCurrentSVID: true,
			testClose:          true,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			t.Logf("🧪 Testing core methods for: %s", scenario.name)

			// Create fresh mock for this scenario
			mockClient, err := mocks.NewMockClient(scenario.socketPath, scenario.refreshInterval)
			require.NoError(t, err, "Should create mock client for: %s", scenario.name)
			require.NotNil(t, mockClient, "Mock client should not be nil for: %s", scenario.name)

			// Test GetTLSConfig if requested
			if scenario.testGetTLSConfig {
				t.Run("GetTLSConfig", func(t *testing.T) {
					config := mockClient.GetTLSConfig()
					assert.NotNil(t, config, "TLS config should not be nil")
					assert.NotNil(t, config.GetCertificate, "TLS config should have GetCertificate function")
					t.Logf("✅ GetTLSConfig working for: %s", scenario.name)
				})
			}

			// Test GetCurrentSVID if requested
			if scenario.testGetCurrentSVID {
				t.Run("GetCurrentSVID", func(t *testing.T) {
					svid, err := mockClient.GetCurrentSVID()
					assert.NoError(t, err, "Should get SVID without error")
					assert.NotNil(t, svid, "SVID should not be nil")
					assert.NotNil(t, svid.ID, "SVID should have an ID")
					assert.NotEmpty(t, svid.Certificates, "SVID should have certificates")
					assert.Equal(t, "spiffe://test.domain/workload", svid.ID.String(), "SVID ID should match expected")
					t.Logf("✅ GetCurrentSVID working for: %s", scenario.name)
				})
			}

			// Test Close if requested
			if scenario.testClose {
				t.Run("Close", func(t *testing.T) {
					assert.False(t, mockClient.IsClosed(), "Should not be closed initially")

					err := mockClient.Close()
					assert.NoError(t, err, "Should close without error")
					assert.True(t, mockClient.IsClosed(), "Should be closed after Close()")
					t.Logf("✅ Close working for: %s", scenario.name)
				})
			}
		})
	}
}

// TestMockClientConfigurationDefaults specifically tests default value handling
func TestMockClientConfigurationDefaults(t *testing.T) {
	defaultScenarios := []struct {
		name          string
		inputSocket   string
		inputInterval time.Duration
		description   string
	}{
		{
			name:          "completely_empty_config",
			inputSocket:   "",
			inputInterval: 0,
			description:   "Empty string socket and zero duration should be handled gracefully",
		},
		{
			name:          "empty_socket_with_interval",
			inputSocket:   "",
			inputInterval: 60 * time.Second,
			description:   "Empty socket with valid interval should work",
		},
		{
			name:          "socket_with_zero_interval",
			inputSocket:   "/test/socket",
			inputInterval: 0,
			description:   "Valid socket with zero interval should work",
		},
	}

	for _, scenario := range defaultScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			t.Logf("🧪 Testing defaults: %s - %s", scenario.name, scenario.description)

			mockClient, err := mocks.NewMockClient(scenario.inputSocket, scenario.inputInterval)
			require.NoError(t, err, "Should handle default configuration gracefully")
			require.NotNil(t, mockClient, "Mock client should be created")

			// Verify the mock stores exactly what we gave it (no magic defaults)
			assert.Equal(t, scenario.inputSocket, mockClient.GetSocketPath(),
				"Mock should store the exact socket path we provided")
			assert.Equal(t, scenario.inputInterval, mockClient.GetRefreshInterval(),
				"Mock should store the exact refresh interval we provided")

			// Verify basic functionality still works
			svid, err := mockClient.GetCurrentSVID()
			assert.NoError(t, err, "Core functionality should work regardless of config")
			assert.NotNil(t, svid, "Should get valid SVID")

			err = mockClient.Close()
			assert.NoError(t, err, "Should close successfully")

			t.Logf("✅ Default handling validated for: %s", scenario.name)
		})
	}
}

// TestMockClientTLSCertificateScenarios tests that TLS certificates actually work
func TestMockClientTLSCertificateScenarios(t *testing.T) {
	scenarios := []struct {
		name            string
		socketPath      string
		refreshInterval time.Duration
		shouldTestCert  bool
		description     string
	}{
		{
			name:            "default_config_cert_test",
			socketPath:      "",
			refreshInterval: 0,
			shouldTestCert:  true,
			description:     "Default config should provide working TLS certificate",
		},
		{
			name:            "custom_config_cert_test",
			socketPath:      "/test/custom/cert",
			refreshInterval: 60 * time.Second,
			shouldTestCert:  true,
			description:     "Custom config should provide working TLS certificate",
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			t.Logf("🔐 Testing TLS certificate for: %s - %s", scenario.name, scenario.description)

			mockClient, err := mocks.NewMockClient(scenario.socketPath, scenario.refreshInterval)
			require.NoError(t, err, "Should create mock client")
			defer mockClient.Close()

			if scenario.shouldTestCert {
				// Get TLS config
				tlsConfig := mockClient.GetTLSConfig()
				require.NotNil(t, tlsConfig, "TLS config should not be nil")
				require.NotNil(t, tlsConfig.GetCertificate, "TLS config should have GetCertificate function")

				// Actually call GetCertificate to test it works
				cert, err := tlsConfig.GetCertificate(nil)
				assert.NoError(t, err, "GetCertificate should work without error")
				assert.NotNil(t, cert, "Certificate should not be nil")
				assert.NotEmpty(t, cert.Certificate, "Certificate should have certificate data")
				assert.NotNil(t, cert.PrivateKey, "Certificate should have private key")

				// Verify certificate properties
				assert.Len(t, cert.Certificate, 1, "Should have exactly one certificate in chain")

				t.Logf("✅ TLS certificate validated for: %s", scenario.name)
			}
		})
	}
}

// TestMockClientRefreshIntervalScenarios tests various refresh interval configurations
func TestMockClientRefreshIntervalScenarios(t *testing.T) {
	scenarios := []struct {
		name            string
		refreshInterval time.Duration
		expectError     bool
		description     string
	}{
		{
			name:            "zero_refresh_interval",
			refreshInterval: 0,
			expectError:     false,
			description:     "Zero refresh interval should be handled gracefully",
		},
		{
			name:            "short_refresh_interval",
			refreshInterval: 1 * time.Second,
			expectError:     false,
			description:     "Short refresh interval should work",
		},
		{
			name:            "medium_refresh_interval",
			refreshInterval: 30 * time.Second,
			expectError:     false,
			description:     "Medium refresh interval should work",
		},
		{
			name:            "long_refresh_interval",
			refreshInterval: 5 * time.Minute,
			expectError:     false,
			description:     "Long refresh interval should work",
		},
		{
			name:            "very_long_refresh_interval",
			refreshInterval: 24 * time.Hour,
			expectError:     false,
			description:     "Very long refresh interval should work",
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			t.Logf("⏰ Testing refresh interval: %s - %s", scenario.name, scenario.description)

			mockClient, err := mocks.NewMockClient("/test/socket", scenario.refreshInterval)

			if scenario.expectError {
				assert.Error(t, err, "Expected error for scenario: %s", scenario.name)
				return
			}

			require.NoError(t, err, "Should create mock client successfully")
			require.NotNil(t, mockClient, "Mock client should not be nil")
			defer mockClient.Close()

			// Verify the refresh interval was stored correctly
			assert.Equal(t, scenario.refreshInterval, mockClient.GetRefreshInterval(),
				"Refresh interval should match for: %s", scenario.name)

			// Verify functionality still works regardless of interval
			svid, err := mockClient.GetCurrentSVID()
			assert.NoError(t, err, "GetCurrentSVID should work regardless of refresh interval")
			assert.NotNil(t, svid, "SVID should not be nil")

			tlsConfig := mockClient.GetTLSConfig()
			assert.NotNil(t, tlsConfig, "TLS config should work regardless of refresh interval")

			t.Logf("✅ Refresh interval validated for: %s (%v)", scenario.name, scenario.refreshInterval)
		})
	}
}

// TestMockClientErrorScenarios tests error cases and edge conditions
func TestMockClientErrorScenarios(t *testing.T) {
	// Note: Since our MockClient is designed to always succeed for testing,
	// these tests focus on testing our test framework's error handling capabilities
	scenarios := []struct {
		name            string
		socketPath      string
		refreshInterval time.Duration
		testAction      string
		expectError     bool
		description     string
	}{
		{
			name:            "negative_refresh_interval",
			socketPath:      "/test/socket",
			refreshInterval: -1 * time.Second,
			testAction:      "create",
			expectError:     false, // MockClient accepts any value for testing
			description:     "Negative refresh interval handling",
		},
		{
			name: "very_long_socket_path",
			socketPath: "/very/long/path/that/might/cause/issues/in/real/implementations/but/should/work/in/mock/" +
				"even/longer/path/that/goes/beyond/normal/filesystem/limits/for/testing/purposes/only",
			refreshInterval: 30 * time.Second,
			testAction:      "create",
			expectError:     false, // MockClient handles any path for testing
			description:     "Very long socket path handling",
		},
		{
			name:            "special_characters_in_path",
			socketPath:      "/test/socket/with/special/chars/!@#$%^&*()",
			refreshInterval: 30 * time.Second,
			testAction:      "create",
			expectError:     false, // MockClient handles any path for testing
			description:     "Special characters in socket path",
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			t.Logf("💥 Testing error scenario: %s - %s", scenario.name, scenario.description)

			switch scenario.testAction {
			case "create":
				mockClient, err := mocks.NewMockClient(scenario.socketPath, scenario.refreshInterval)

				if scenario.expectError {
					assert.Error(t, err, "Expected error for scenario: %s", scenario.name)
					assert.Nil(t, mockClient, "Client should be nil on error")
					return
				}

				require.NoError(t, err, "Should handle edge case gracefully: %s", scenario.name)
				require.NotNil(t, mockClient, "Mock client should not be nil")
				defer mockClient.Close()

				// Verify the client still works with edge case inputs
				assert.Equal(t, scenario.socketPath, mockClient.GetSocketPath(),
					"Should store exact socket path provided")
				assert.Equal(t, scenario.refreshInterval, mockClient.GetRefreshInterval(),
					"Should store exact refresh interval provided")

				// Test that core functionality still works
				svid, err := mockClient.GetCurrentSVID()
				assert.NoError(t, err, "Core functionality should work with edge case inputs")
				assert.NotNil(t, svid, "SVID should be available")

				t.Logf("✅ Edge case handled gracefully: %s", scenario.name)
			}
		})
	}
}

// TestMockClientConcurrencyScenarios tests concurrent access to mock client
func TestMockClientConcurrencyScenarios(t *testing.T) {
	t.Run("concurrent_svid_access", func(t *testing.T) {
		t.Logf("🔄 Testing concurrent SVID access")

		mockClient, err := mocks.NewMockClient("/test/concurrent", 30*time.Second)
		require.NoError(t, err)
		defer mockClient.Close()

		// Test concurrent access to GetCurrentSVID
		const numGoroutines = 10
		done := make(chan bool, numGoroutines)
		errors := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer func() { done <- true }()

				svid, err := mockClient.GetCurrentSVID()
				if err != nil {
					errors <- err
					return
				}

				if svid == nil {
					errors <- fmt.Errorf("goroutine %d got nil SVID", id)
					return
				}

				if svid.ID.String() != "spiffe://test.domain/workload" {
					errors <- fmt.Errorf("goroutine %d got wrong SPIFFE ID: %s", id, svid.ID.String())
					return
				}
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		// Check for any errors
		close(errors)
		for err := range errors {
			t.Errorf("Concurrent access error: %v", err)
		}

		t.Logf("✅ Concurrent SVID access completed successfully")
	})

	t.Run("concurrent_tls_config_access", func(t *testing.T) {
		t.Logf("🔄 Testing concurrent TLS config access")

		mockClient, err := mocks.NewMockClient("/test/concurrent", 30*time.Second)
		require.NoError(t, err)
		defer mockClient.Close()

		// Test concurrent access to GetTLSConfig
		const numGoroutines = 5
		done := make(chan bool, numGoroutines)
		errors := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer func() { done <- true }()

				config := mockClient.GetTLSConfig()
				if config == nil {
					errors <- fmt.Errorf("goroutine %d got nil TLS config", id)
					return
				}

				if config.GetCertificate == nil {
					errors <- fmt.Errorf("goroutine %d got TLS config without GetCertificate", id)
					return
				}

				// Actually test getting a certificate
				cert, err := config.GetCertificate(nil)
				if err != nil {
					errors <- fmt.Errorf("goroutine %d failed to get certificate: %v", id, err)
					return
				}
				if cert == nil {
					errors <- fmt.Errorf("goroutine %d got nil certificate", id)
					return
				}
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		// Check for any errors
		close(errors)
		for err := range errors {
			t.Errorf("Concurrent TLS access error: %v", err)
		}

		t.Logf("✅ Concurrent TLS config access completed successfully")
	})
}
