# 🧪 Unit Tests

This directory contains unit tests that validate individual components and logic without external dependencies.

## 📁 Directory Structure

```
unit/
├── config/           # Configuration-related tests
│   └── refresh_config_test.go  # Tests for refresh_interval and refresh_at_percent
└── README.md         # This file
```

## 🎯 Test Categories

### `/config` - Configuration Tests
Tests that validate configuration logic, including:
- Configuration validation
- Default value application
- Mutual exclusivity rules
- Value range validation
- Configuration calculations

**Example tests:**
- `refresh_config_test.go` - Tests refresh mechanism configuration

## 🚀 Running Unit Tests

### Run all unit tests:
```bash
go test ./test/unit/...
```

### Run specific category:
```bash
# Configuration tests only
go test ./test/unit/config/...
```

### Run with verbose output:
```bash
go test -v ./test/unit/...
```

### Run with coverage:
```bash
go test -cover ./test/unit/...
```

## 📝 Adding New Unit Tests

When adding new unit tests:

1. **Determine the category**: Is it config, validation, calculation, etc.?
2. **Create appropriate subdirectory** if needed
3. **Use descriptive test names** that explain what's being tested
4. **Follow the naming convention**: `*_test.go`
5. **Use table-driven tests** where appropriate

### Example Structure:
```go
package config_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestFeatureName(t *testing.T) {
    tests := []struct {
        name        string
        input       interface{}
        expected    interface{}
        expectError bool
    }{
        // Test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test logic
        })
    }
}
```

## 🌟 Best Practices

1. **Keep tests focused**: Each test should validate one specific behavior
2. **Use meaningful names**: Test names should describe what they're testing
3. **Avoid external dependencies**: Unit tests should not require SPIRE agent, network, etc.
4. **Use mocks when needed**: For testing interfaces and dependencies
5. **Document complex tests**: Add comments explaining non-obvious test logic

## 🔗 Related

- [Integration Tests](../integration/README.md) - Tests requiring external dependencies
- [Mock Tests](../mocks/) - Mock implementations for testing
- [Main Test README](../README.md) - Overview of entire test framework
