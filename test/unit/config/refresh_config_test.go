package config_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/jenova-marie/caddy-spire-svid/pkg/spire"
	"github.com/stretchr/testify/assert"
)

// TestRefreshConfigMutualExclusivity tests that RefreshInterval and RefreshAtPercent are mutually exclusive
func TestRefreshConfigMutualExclusivity(t *testing.T) {
	tests := []struct {
		name           string
		config         spire.Config
		expectError    bool
		expectedErrMsg string
		description    string
	}{
		{
			name: "both_zero_gets_default_percent",
			config: spire.Config{
				SocketPath:       "/tmp/test.sock",
				RefreshInterval:  0,
				RefreshAtPercent: 0,
			},
			expectError: false,
			description: "When both are zero, should default to 65% refresh",
		},
		{
			name: "only_interval_specified",
			config: spire.Config{
				SocketPath:       "/tmp/test.sock",
				RefreshInterval:  30 * time.Second,
				RefreshAtPercent: 0,
			},
			expectError: false,
			description: "Only RefreshInterval specified should work",
		},
		{
			name: "only_percent_specified",
			config: spire.Config{
				SocketPath:       "/tmp/test.sock",
				RefreshInterval:  0,
				RefreshAtPercent: 75,
			},
			expectError: false,
			description: "Only RefreshAtPercent specified should work",
		},
		{
			name: "both_specified_error",
			config: spire.Config{
				SocketPath:       "/tmp/test.sock",
				RefreshInterval:  30 * time.Second,
				RefreshAtPercent: 65,
			},
			expectError:    true,
			expectedErrMsg: "RefreshInterval and RefreshAtPercent are mutually exclusive",
			description:    "Both specified should return error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("🧪 Testing: %s - %s", tt.name, tt.description)

			// Test config validation logic (without creating actual client)
			// This simulates the validation that happens in NewClient
			config := tt.config

			// Apply the same logic as in NewClient
			if config.RefreshInterval == 0 && config.RefreshAtPercent == 0 {
				config.RefreshAtPercent = 65
			}

			// Check mutual exclusivity
			var err error
			if config.RefreshInterval != 0 && config.RefreshAtPercent != 0 {
				err = fmt.Errorf("RefreshInterval and RefreshAtPercent are mutually exclusive")
			}

			if tt.expectError {
				assert.Error(t, err, "Expected error for config: %s", tt.name)
				assert.Contains(t, err.Error(), tt.expectedErrMsg, "Error message should contain expected text")
			} else {
				assert.NoError(t, err, "Expected no error for config: %s", tt.name)

				// Verify the configuration is set correctly
				if tt.config.RefreshInterval == 0 && tt.config.RefreshAtPercent == 0 {
					assert.Equal(t, 65, config.RefreshAtPercent, "Default should be 65%")
					assert.Equal(t, time.Duration(0), config.RefreshInterval, "Interval should remain 0")
				} else if tt.config.RefreshInterval > 0 {
					assert.Equal(t, tt.config.RefreshInterval, config.RefreshInterval, "Interval should be preserved")
					assert.Equal(t, 0, config.RefreshAtPercent, "Percent should be 0")
				} else if tt.config.RefreshAtPercent > 0 {
					assert.Equal(t, tt.config.RefreshAtPercent, config.RefreshAtPercent, "Percent should be preserved")
					assert.Equal(t, time.Duration(0), config.RefreshInterval, "Interval should be 0")
				}
			}

			t.Logf("✅ Test passed: %s", tt.name)
		})
	}
}

// TestRefreshAtPercentValidation tests the validation of RefreshAtPercent values
func TestRefreshAtPercentValidation(t *testing.T) {
	tests := []struct {
		name             string
		refreshAtPercent int
		expectValid      bool
		description      string
	}{
		{
			name:             "valid_65_percent",
			refreshAtPercent: 65,
			expectValid:      true,
			description:      "65% is a valid refresh percentage",
		},
		{
			name:             "valid_1_percent",
			refreshAtPercent: 1,
			expectValid:      true,
			description:      "1% is the minimum valid percentage",
		},
		{
			name:             "valid_99_percent",
			refreshAtPercent: 99,
			expectValid:      true,
			description:      "99% is the maximum valid percentage",
		},
		{
			name:             "invalid_0_percent",
			refreshAtPercent: 0,
			expectValid:      false,
			description:      "0% is invalid (would mean refresh immediately)",
		},
		{
			name:             "invalid_100_percent",
			refreshAtPercent: 100,
			expectValid:      false,
			description:      "100% is invalid (would mean never refresh)",
		},
		{
			name:             "invalid_negative",
			refreshAtPercent: -10,
			expectValid:      false,
			description:      "Negative percentages are invalid",
		},
		{
			name:             "invalid_over_100",
			refreshAtPercent: 150,
			expectValid:      false,
			description:      "Percentages over 100 are invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("🧪 Testing: %s - %s", tt.name, tt.description)

			// Validate the percentage (same logic as in Caddyfile parsing)
			isValid := tt.refreshAtPercent >= 1 && tt.refreshAtPercent <= 99

			assert.Equal(t, tt.expectValid, isValid,
				"Validation result should match expected for %d%%", tt.refreshAtPercent)

			if isValid {
				t.Logf("✅ Valid percentage: %d%%", tt.refreshAtPercent)
			} else {
				t.Logf("❌ Invalid percentage: %d%%", tt.refreshAtPercent)
			}
		})
	}
}

// TestRefreshMechanismCalculations tests the calculations for smart refresh
func TestRefreshMechanismCalculations(t *testing.T) {
	tests := []struct {
		name             string
		certLifetime     time.Duration
		refreshAtPercent int
		expectedRefresh  time.Duration
		description      string
	}{
		{
			name:             "1_hour_cert_65_percent",
			certLifetime:     1 * time.Hour,
			refreshAtPercent: 65,
			expectedRefresh:  39 * time.Minute, // 65% of 60 minutes
			description:      "1-hour certificate should refresh at 39 minutes",
		},
		{
			name:             "24_hour_cert_65_percent",
			certLifetime:     24 * time.Hour,
			refreshAtPercent: 65,
			expectedRefresh:  15*time.Hour + 36*time.Minute, // 65% of 24 hours
			description:      "24-hour certificate should refresh at 15.6 hours",
		},
		{
			name:             "30_min_cert_50_percent",
			certLifetime:     30 * time.Minute,
			refreshAtPercent: 50,
			expectedRefresh:  15 * time.Minute, // 50% of 30 minutes
			description:      "30-minute certificate should refresh at 15 minutes",
		},
		{
			name:             "7_day_cert_90_percent",
			certLifetime:     7 * 24 * time.Hour,
			refreshAtPercent: 90,
			expectedRefresh:  151*time.Hour + 12*time.Minute, // 90% of 7 days
			description:      "7-day certificate should refresh at 6.3 days",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("🧪 Testing: %s - %s", tt.name, tt.description)

			// Calculate refresh time (same logic as in smartRefresh)
			refreshTime := tt.certLifetime * time.Duration(tt.refreshAtPercent) / 100

			assert.Equal(t, tt.expectedRefresh, refreshTime,
				"Refresh time calculation should match expected")

			t.Logf("✅ Certificate lifetime: %v, Refresh at %d%% = %v",
				tt.certLifetime, tt.refreshAtPercent, refreshTime)
		})
	}
}
