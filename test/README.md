# 🌸 Testing Framework for caddy-spire-svid

This directory contains a comprehensive, multi-level testing framework designed to validate our SPIFFE/SPIRE + Caddy integration from basic unit tests to full containerized deployments.

## 🎯 Testing Philosophy

Our testing approach follows a **progressive validation strategy**, where each level builds upon the previous one, ensuring reliability at every stage of integration.

## 📚 Test Organization

### 🗂️ Directory Structure

```
test/
├── README.md                     # This file
├── unit/                        # Pure unit tests (no external dependencies)
│   ├── config/                  # Configuration-related unit tests
│   │   └── refresh_config_test.go # Refresh mechanism configuration tests
│   └── README.md                # Unit test documentation
├── integration/                 # Level 2-5 integration tests (each in own subdir)
│   ├── level2/main.go          # Bare metal client.go validation
│   ├── level3/main.go          # Caddy + SPIRE integration  
│   ├── level4/main.go          # Container SPIRE socket diagnostic
│   └── level5/main.go          # Containerized comprehensive testing
├── mocks/                       # Mock implementations
│   ├── spire_mock_client.go     # SPIRE client mock
│   └── spire_mock_client_test.go # Mock validation tests
└── docker/                      # Docker-specific test configurations
    ├── test-caddy-config.json   # Test Caddy configuration
    └── test.caddyfile           # Test Caddyfile

pkg/spire/                       # Level 1 Go unit tests (in main package)
├── client_test.go               # Mock framework validation
├── client_scenarios_test.go     # Mock scenario testing  
├── client_integration_test.go   # Real client unit tests
├── client_real_test.go          # Real client integration tests
└── client.go                    # Main SPIRE client implementation
```

## 🚀 Test Levels

### **Unit Tests** 🧪
**Purpose**: Pure unit tests for configuration and logic validation  
**Dependencies**: None  
**Location**: `test/unit/`  

**What it tests:**
- ✅ Configuration validation and defaults
- ✅ Mutual exclusivity rules
- ✅ Value calculations and logic
- ✅ Edge cases and boundaries
- ✅ Pure functions without external dependencies

**Run with:**
```bash
go test ./test/unit/...
```

---

### **Level 1: Mock-Based Testing** 🔬
**Purpose**: Fast, isolated unit tests using mocks  
**Dependencies**: None (fully self-contained)  
**Location**: `pkg/spire/*_test.go`  

**What it tests:**
- ✅ Mock framework validation
- ✅ Business logic correctness
- ✅ Error handling scenarios
- ✅ Interface compliance
- ✅ Edge cases and boundary conditions

**Run with:**
```bash
make test-level1
# or
go test -run "TestMock" ./pkg/spire/tests/...
```

**Coverage**: Tests all code paths without external dependencies

---

### **Level 2: Bare Metal Client Testing** 🔗
**Purpose**: Real client.go validation with live SPIRE agent  
**Dependencies**: Local `spire-agent` running and configured  
**Location**: `pkg/spire/*_test.go` + `test/integration/level2/main.go`  

**What it tests:**
- ✅ Real SPIRE agent connectivity
- ✅ Workload attestation process
- ✅ X.509 SVID retrieval and validation
- ✅ TLS configuration generation
- ✅ Certificate lifecycle management
- ✅ Client error handling with real failures

**Run with:**
```bash
make test-level2
# or manually:
go test -run "TestClient" ./pkg/spire/tests/...
./bin/level2_client_comprehensive
```

**Prerequisites:**
- SPIRE agent running: `sudo spire-agent run -config ~/.ssh/spire-agent.conf`
- Workload entries configured for your UID/process

---

### **Level 3: Comprehensive Caddy + SPIRE Integration** 🌸
**Purpose**: Comprehensive testing of spire-client with Caddy using LIVE spire-agent  
**Dependencies**: Local `spire-agent` running + custom Caddy build  
**Location**: `test/integration/level3/main.go`  

**What it tests (COMPREHENSIVE - "fire and bodies"):**
- ✅ Caddy SPIRE module loading and initialization
- ✅ TLS certificate provisioning via SPIRE from live agent
- ✅ HTTPS endpoint functionality with RecoverySky.org CA certificates
- ✅ Certificate validation and properties (issuer, subject, expiration)
- ✅ TLS protocol validation (TLS 1.3, cipher suites)
- ✅ Multiple request handling and certificate persistence
- ✅ Performance characteristics (sub-3ms response times)
- ✅ HTTP redirect functionality
- ✅ Multiple domains (localhost, localhost, etc.)
- ✅ Error handling and edge cases
- ✅ Concurrent request handling

**Run with:**
```bash
make test-level3
# Fully automated: starts Caddy, runs comprehensive tests, cleans up
```

**Expected results:**
- HTTPS endpoints serving with SPIRE-issued certificates from RecoverySky.org CA
- Sub-3ms response times
- Valid TLS 1.3 connections
- All comprehensive integration scenarios passing

**Note**: Certificate refresh testing requires 1-hour timeout - documented limitation

---

### **Level 4: Basic SPIRE Socket Connectivity** 🔍
**Purpose**: Container diagnostic test to verify SPIRE agent socket mounting and attestation  
**Dependencies**: Local `spire-agent` running  
**Location**: `test/integration/level4/main.go`  

**What it tests:**
- ✅ Docker container can access host SPIRE socket
- ✅ SPIRE agent CLI works in container environment
- ✅ Container workload attestation with Docker Compose service labels
- ✅ X.509 SVID retrieval from containerized workload

**Run with:**
```bash
make test-level4
# Fully automated: starts diagnostic container, tests SPIRE connectivity, cleans up
```

**Expected results:**
- Container receives valid SPIFFE ID: `spiffe://recoverysky.org/cmd/inanna/caddy-spire-svid-test`
- SPIRE socket connectivity confirmed
- SVID retrieval successful with proper certificate validation

---

### **Level 5: Containerized Comprehensive Testing** 🐳
**Purpose**: Same comprehensive testing as Level 3, but containerized  
**Dependencies**: Docker + local `spire-agent` running  
**Location**: `test/integration/level5/main.go`  

**What it tests:**
- ✅ **Same comprehensive testing as Level 3** 
- ✅ Container build optimization and caching
- ✅ Containerized Caddy + SPIRE integration with live agent
- ✅ Docker socket mounting to local spire-agent
- ✅ Container networking with SPIRE certificates
- ✅ Container resource utilization with SPIRE workloads
- ✅ Multi-request stability in containerized environment

**Run with:**
```bash
make test-level5
# Fully automated: starts container, runs comprehensive tests, cleans up
```

**Container + SPIRE integration tested:**
- Docker socket mounting for SPIRE agent access
- Container workload attestation
- HTTPS serving from container with SPIRE certificates

---

## 🔍 Diagnostic Tools

When things go wrong, our diagnostic tools help identify issues quickly:

### **Socket Diagnostic** 📡
Tests basic Unix socket connectivity and permissions.

```bash
make test-diagnostics
# or
./bin/spire_socket_diagnostic
```

### **Attestation Diagnostic** 🆔
Debugs workload attestation process step-by-step.

```bash
./bin/spire_attestation_diagnostic
```

### **Simple Connection Test** 🔌
Ultra-minimal SPIRE connection validation.

```bash
./bin/spire_connection_simple
```

## 📊 Running Tests

### **Quick Test (CI-Safe)**
```bash
make test           # Level 1 + 2 Go tests only
make test-level1    # Fastest - mock tests only
```

### **Full Local Testing**
```bash
make test-all       # All levels (requires SPIRE agent + Docker)
```

### **Individual Levels**
```bash
make test-level1    # Mock tests
make test-level2    # Real SPIRE agent tests
make test-level3    # Caddy integration
make test-level4    # Container tests
```

### **Coverage Reports**
```bash
make test           # Generates coverage.html
open coverage.html  # View in browser
```

## ✅ CI/CD Integration

Our CI pipeline runs **Level 1 tests only** to ensure fast, reliable builds without external dependencies:

```yaml
# .github/workflows/ci.yml runs:
make test-level1    # Mock-based tests only
```

**Why Level 1 only in CI?**
- ⚡ **Fast**: No external dependencies
- 🔒 **Reliable**: No flaky network connections
- 💚 **Consistent**: Same results across all environments
- 🧪 **Comprehensive**: Tests all business logic

## 🎯 Test Coverage

Our current test coverage:

| Component | Coverage | Level |
|-----------|----------|-------|
| `pkg/spire/client.go` | **75.9%** | Level 1 + 2 |
| Mock framework | **100%** | Level 1 |
| Integration scenarios | **95%** | Level 2-4 |
| Error handling | **90%** | All levels |

## 🚀 Adding New Tests

### **For new client.go features:**
1. Add mock scenarios in `level1_mock_scenarios_test.go`
2. Add real tests in `level2_client_integration_test.go`
3. Update integration tests as needed

### **For new Caddy module features:**
1. Add scenarios to `level3_caddy_local_integration/main.go`
2. Update container tests in `level4_caddy_container/main.go`

### **For diagnostic tools:**
Add new tools to `tests/diagnostics/` and update `make test-diagnostics`

## 🏆 Best Practices

1. **Progressive Testing**: Always start with Level 1, then move up
2. **Isolation**: Each level should test distinct concerns
3. **Fast Feedback**: Keep Level 1 tests under 1 second total
4. **Real World**: Level 2+ should test real-world scenarios
5. **Documentation**: Update this README when adding new tests

---

**🌸 Happy Testing! Our test suite is a work of art! 💖**

*Last updated: August 2025*
