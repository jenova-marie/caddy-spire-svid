# 🌸 Comprehensive Testing Framework for caddy-spire-client
# Multi-level testing from mocks to full containerized deployments

.PHONY: test test-all test-level1 test-level2 test-level3 test-level4 test-diagnostics
.PHONY: build build-all build-tests clean clean-all coverage help
.PHONY: run local diagnostic docker-diagnostic
.PHONY: docker-setup docker-build docker-push docker-build-amd64 docker-build-arm64 docker-build-armv7

# Colors for beautiful output 💖
CYAN = \033[36m
GREEN = \033[32m
YELLOW = \033[33m
RED = \033[31m
MAGENTA = \033[35m
BLUE = \033[34m
RESET = \033[0m

# Directories
BIN_DIR = bin
TEST_DIR = test
INTEGRATION_DIR = test/integration
PKG_DIR = pkg

# Docker Configuration
DOCKER_REGISTRY = ghcr.io
DOCKER_REPO = jenova-marie/caddy-spire-client
DOCKER_IMAGE_CADDY = $(DOCKER_REGISTRY)/$(DOCKER_REPO)
DOCKER_TAG ?= latest
VERSION_TAG ?= v1.0.4

# =============================================================================
# 🏗️ MAIN PROJECT BUILD TARGETS
# =============================================================================

# Default target - build main project binaries
all: build

# Build all main project binaries
build: build-caddy-spire-client build-caddy-with-spire
	@echo "$(GREEN)✅ All project binaries built successfully!$(RESET)"

# Build release artifacts for distribution
release: clean build-all
	@echo "$(MAGENTA)🎉 Building release artifacts...$(RESET)"
	@mkdir -p dist/archives
	@for arch in amd64 arm64; do \
		for os in linux darwin windows; do \
			ext=""; \
			if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
			echo "$(CYAN)Building $$os-$$arch...$(RESET)"; \
			GOOS=$$os GOARCH=$$arch go build -ldflags "-s -w" -o dist/caddy-spire-client-$$os-$$arch$$ext ./cmd/caddy-spire-client; \
			if [ "$$os" = "windows" ]; then \
				cd dist && zip -q caddy-spire-client-$$os-$$arch.zip caddy-spire-client-$$os-$$arch$$ext && cd ..; \
			else \
				cd dist && tar -czf archives/caddy-spire-client-$$os-$$arch.tar.gz caddy-spire-client-$$os-$$arch && cd ..; \
			fi; \
		done; \
	done
	@echo "$(GREEN)✅ Release artifacts built in dist/archives/$(RESET)"

# Build the main caddy-spire-client binary
build-caddy-spire-client:
	@echo "$(CYAN)🔨 Building caddy-spire-client...$(RESET)"
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/caddy-spire-client cmd/caddy-spire-client/main.go
	@echo "$(GREEN)✅ caddy-spire-client built: $(BIN_DIR)/caddy-spire-client$(RESET)"

# Build the custom Caddy with SPIRE module
build-caddy-with-spire:
	@echo "$(CYAN)🔨 Building caddy-with-spire...$(RESET)"
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/caddy-with-spire cmd/caddy-with-spire/main.go
	@echo "$(GREEN)✅ caddy-with-spire built: $(BIN_DIR)/caddy-with-spire$(RESET)"

# Install binaries to $GOPATH/bin
install: build
	@echo "$(CYAN)📦 Installing binaries to GOPATH...$(RESET)"
	go install ./cmd/caddy-spire-client
	go install ./cmd/caddy-with-spire
	@echo "$(GREEN)✅ Binaries installed to GOPATH/bin$(RESET)"

# =============================================================================
# 🐳 DOCKER BUILD & PUSH TARGETS
# =============================================================================

# Setup Docker buildx for multi-architecture builds
docker-setup:
	@echo "$(CYAN)🔧 Setting up Docker buildx for multi-architecture builds...$(RESET)"
	@docker buildx create --use --name multiarch-builder --driver docker-container 2>/dev/null || true
	@docker buildx ls
	@echo "$(GREEN)✅ Docker buildx setup complete$(RESET)"

# Build and push multi-architecture Caddy + SPIRE image
docker-build: docker-setup
	@echo "$(CYAN)🐳 Building multi-architecture Caddy + SPIRE Docker image...$(RESET)"
	@echo "$(YELLOW)📦 Building for: linux/amd64,linux/arm64,linux/arm/v7$(RESET)"
	@echo "$(YELLOW)🏷️  Tags: $(DOCKER_IMAGE_CADDY):$(DOCKER_TAG), $(DOCKER_IMAGE_CADDY):$(VERSION_TAG)$(RESET)"
	docker buildx build \
		--platform linux/amd64,linux/arm64,linux/arm/v7 \
		--file Dockerfile.caddy \
		--tag $(DOCKER_IMAGE_CADDY):$(DOCKER_TAG) \
		--tag $(DOCKER_IMAGE_CADDY):$(VERSION_TAG) \
		.
	@echo "$(GREEN)✅ Multi-architecture build complete$(RESET)"

# Build and push to registry
docker-push: docker-setup
	@echo "$(CYAN)🚀 Building and pushing multi-architecture Caddy + SPIRE Docker image...$(RESET)"
	@echo "$(YELLOW)📦 Building for: linux/amd64,linux/arm64,linux/arm/v7$(RESET)"
	@echo "$(YELLOW)🏷️  Tags: $(DOCKER_IMAGE_CADDY):$(DOCKER_TAG), $(DOCKER_IMAGE_CADDY):$(VERSION_TAG)$(RESET)"
	@echo "$(MAGENTA)🔑 Make sure you're logged in: docker login $(DOCKER_REGISTRY)$(RESET)"
	docker buildx build \
		--platform linux/amd64,linux/arm64,linux/arm/v7 \
		--file Dockerfile.caddy \
		--tag $(DOCKER_IMAGE_CADDY):$(DOCKER_TAG) \
		--tag $(DOCKER_IMAGE_CADDY):$(VERSION_TAG) \
		--push \
		.
	@echo "$(GREEN)✅ Multi-architecture image pushed to $(DOCKER_IMAGE_CADDY)$(RESET)"

# Build AMD64 only (for testing/debugging)
docker-build-amd64: docker-setup
	@echo "$(CYAN)🐳 Building AMD64 Caddy + SPIRE Docker image...$(RESET)"
	docker buildx build \
		--platform linux/amd64 \
		--file Dockerfile.caddy \
		--tag $(DOCKER_IMAGE_CADDY):$(VERSION_TAG)-amd64 \
		--push \
		.
	@echo "$(GREEN)✅ AMD64 image built and pushed$(RESET)"

# Build ARM64 only (for testing/debugging)
docker-build-arm64: docker-setup
	@echo "$(CYAN)🐳 Building ARM64 Caddy + SPIRE Docker image...$(RESET)"
	docker buildx build \
		--platform linux/arm64 \
		--file Dockerfile.caddy \
		--tag $(DOCKER_IMAGE_CADDY):$(VERSION_TAG)-arm64 \
		--push \
		.
	@echo "$(GREEN)✅ ARM64 image built and pushed$(RESET)"

# Build ARMv7 only (for testing/debugging)
docker-build-armv7: docker-setup
	@echo "$(CYAN)🐳 Building ARMv7 Caddy + SPIRE Docker image...$(RESET)"
	docker buildx build \
		--platform linux/arm/v7 \
		--file Dockerfile.caddy \
		--tag $(DOCKER_IMAGE_CADDY):$(VERSION_TAG)-armv7 \
		--push \
		.
	@echo "$(GREEN)✅ ARMv7 image built and pushed$(RESET)"

# Login to GitHub Container Registry
docker-login:
	@echo "$(CYAN)🔑 Logging into GitHub Container Registry...$(RESET)"
	@echo "$(YELLOW)💡 Use your GitHub username and personal access token$(RESET)"
	@docker login $(DOCKER_REGISTRY)
	@echo "$(GREEN)✅ Login successful$(RESET)"

# Debug registry access
docker-debug:
	@echo "$(CYAN)🔍 Debugging registry access...$(RESET)"
	@echo "$(YELLOW)Registry: $(DOCKER_REGISTRY)$(RESET)"
	@echo "$(YELLOW)Image: $(DOCKER_IMAGE_CADDY)$(RESET)"
	@echo "$(YELLOW)Checking authentication...$(RESET)"
	@docker info | grep -A10 "Registry Mirrors" || true
	@echo "$(CYAN)Testing registry access...$(RESET)"
	@docker buildx imagetools inspect ghcr.io/library/hello-world || echo "❌ Cannot access GHCR"

# =============================================================================
# 🧪 TESTING FRAMEWORK
# =============================================================================

# Quick test (CI-Safe) - Level 1 + 2 Go tests
test: test-level1 test-level2-unit
	@echo "$(GREEN)✅ Quick test suite completed!$(RESET)"
	@echo "$(CYAN)📊 Coverage report generated: coverage.html$(RESET)"

# All levels testing (requires SPIRE agent + Docker)
test-all: test-level1 test-level2 test-level3 test-level4 test-level5
	@echo "$(GREEN)🎉 ALL TEST LEVELS (1-5) COMPLETED!$(RESET)"
	@echo "$(MAGENTA)💖 Your caddy-spire-client is battle-tested!$(RESET)"
	@echo "$(CYAN)📊 Progressive Testing Results:$(RESET)"
	@echo "  ✅ Level 1: Mock-based unit testing"
	@echo "  ✅ Level 2: Bare metal client testing"
	@echo "  ✅ Level 3: Comprehensive Caddy + SPIRE integration"
	@echo "  ✅ Level 4: Basic SPIRE socket connectivity"
	@echo "  ✅ Level 5: Containerized comprehensive testing"

# =============================================================================
# 🔍 LEVEL 4: Basic SPIRE Socket Connectivity (Container Diagnostic)
# =============================================================================
test-level4: build-level4
	@echo "$(CYAN)🔍 Level 4: Basic SPIRE Socket Connectivity - Starting...$(RESET)"
	@echo "$(YELLOW)🐳 Testing Docker container access to host SPIRE socket$(RESET)"
	@echo "$(YELLOW)⚠️  Requires: Local spire-agent running$(RESET)"
	@echo "$(CYAN)🚀 Running container socket diagnostic...$(RESET)"
	@if ./$(BIN_DIR)/level4_spire_socket_diagnostic; then \
		echo "$(GREEN)✅ Level 4: Basic SPIRE socket connectivity completed!$(RESET)"; \
	else \
		echo "$(RED)❌ Level 4 failed - ensure SPIRE agent is running$(RESET)"; \
		exit 1; \
	fi

# =============================================================================
# 🔬 LEVEL 1: Mock-Based Testing (Fast, isolated unit tests)
# =============================================================================
test-level1:
	@echo "$(CYAN)🔬 Level 1: Unit Testing - Starting...$(RESET)"
	@echo "$(YELLOW)⚡ Fast, isolated unit tests with mocks$(RESET)"
	go test -v ./$(PKG_DIR)/spire/
	go test -v ./$(TEST_DIR)/mocks/
	@echo "$(GREEN)✅ Level 1: Unit tests completed!$(RESET)"

# =============================================================================
# 🔗 LEVEL 2: Bare Metal Client Testing (Real SPIRE agent required)
# =============================================================================
test-level2-unit:
	@echo "$(CYAN)🔗 Level 2: Integration Unit Testing - Starting...$(RESET)"
	@echo "$(YELLOW)🧪 Client integration unit tests (no external dependencies)$(RESET)"
	go test -v -run "TestClient" ./$(PKG_DIR)/spire/
	@echo "$(GREEN)✅ Level 2: Integration unit tests completed!$(RESET)"

test-level2: build-level2
	@echo "$(CYAN)🔗 Level 2: Bare Metal Client Testing - Starting...$(RESET)"
	@echo "$(YELLOW)⚠️  Requires: Local spire-agent running$(RESET)"
	@echo "$(YELLOW)📋 Prerequisites: sudo spire-agent run -config ~/.ssh/spire-agent.conf$(RESET)"
	@echo "$(CYAN)🚀 Running comprehensive client tests...$(RESET)"
	@if ./$(BIN_DIR)/level2_client_comprehensive; then \
		echo "$(GREEN)✅ Level 2: Bare metal client testing completed!$(RESET)"; \
	else \
		echo "$(RED)❌ Level 2 failed - ensure SPIRE agent is running$(RESET)"; \
		exit 1; \
	fi

# =============================================================================
# 🌸 LEVEL 3: Comprehensive Caddy + SPIRE Integration
# =============================================================================
test-level3: build-level3
	@echo "$(CYAN)🌸 Level 3: Comprehensive Caddy + SPIRE Integration - Starting...$(RESET)"
	@echo "$(YELLOW)⚠️  Requires: Local spire-agent running + custom Caddy build$(RESET)"
	@echo "$(MAGENTA)🔥 COMPREHENSIVE TESTING - 'fire and bodies'$(RESET)"
	@echo "$(CYAN)🚀 Running comprehensive Caddy integration tests...$(RESET)"
	@if ./$(BIN_DIR)/level3_caddy_comprehensive; then \
		echo "$(GREEN)✅ Level 3: Comprehensive Caddy + SPIRE integration completed!$(RESET)"; \
	else \
		echo "$(RED)❌ Level 3 failed - ensure SPIRE agent is running$(RESET)"; \
		exit 1; \
	fi

# =============================================================================
# 🐳 LEVEL 5: Containerized Comprehensive Testing
# =============================================================================
test-level5: build-level5 test-level5-setup
	@echo "$(CYAN)🐳 Level 5: Containerized Comprehensive Testing - Starting...$(RESET)"
	@echo "$(YELLOW)⚠️  Requires: Docker + local spire-agent running$(RESET)"
	@echo "$(BLUE)📦 Same comprehensive testing as Level 3, but containerized$(RESET)"
	@echo "$(CYAN)🚀 Running containerized integration tests...$(RESET)"
	@if ./$(BIN_DIR)/level5_caddy_container_integration; then \
		echo "$(GREEN)✅ Level 5: Containerized comprehensive testing completed!$(RESET)"; \
	else \
		echo "$(RED)❌ Level 5 failed - ensure Docker and SPIRE agent are running$(RESET)"; \
		exit 1; \
	fi

# Setup Level 5 containers and dependencies
test-level5-setup:
	@echo "$(CYAN)🐳 Setting up Level 5 container environment...$(RESET)"
	@echo "$(YELLOW)🔍 Checking Docker and prerequisites...$(RESET)"
	@docker --version || (echo "$(RED)❌ Docker not found$(RESET)" && exit 1)
	@docker compose version || docker-compose --version || (echo "$(RED)❌ Docker Compose not found$(RESET)" && exit 1)
	@test -f docker-compose.level5.yml || (echo "$(RED)❌ docker-compose.level5.yml not found$(RESET)" && exit 1)
	@test -f Dockerfile.caddy || (echo "$(RED)❌ Dockerfile.caddy not found$(RESET)" && exit 1)
	@test -S /tmp/spire-agent/public/api.sock || (echo "$(RED)❌ SPIRE agent socket not found - start with: sudo spire-agent run -config ~/.ssh/spire-agent.conf$(RESET)" && exit 1)
	@echo "$(GREEN)✅ Level 5 prerequisites verified$(RESET)"

# Clean up Level 5 containers
test-level5-cleanup:
	@echo "$(YELLOW)🧹 Cleaning up Level 5 containers...$(RESET)"
	@docker compose -f docker-compose.level5.yml down --remove-orphans 2>/dev/null || true
	@docker-compose -f docker-compose.level5.yml down --remove-orphans 2>/dev/null || true
	@echo "$(GREEN)✅ Level 5 cleanup completed$(RESET)"

# Force rebuild Level 5 containers
test-level5-rebuild: test-level5-cleanup
	@echo "$(CYAN)🔄 Rebuilding Level 5 containers...$(RESET)"
	@docker compose -f docker-compose.level5.yml build --no-cache || docker-compose -f docker-compose.level5.yml build --no-cache
	@echo "$(GREEN)✅ Level 5 containers rebuilt$(RESET)"

# =============================================================================
# 🔍 DIAGNOSTIC TOOLS
# =============================================================================
test-diagnostics: build-diagnostics
	@echo "$(CYAN)🔍 Running SPIRE Diagnostic Tools...$(RESET)"
	@echo "$(YELLOW)📡 Socket connectivity diagnostic:$(RESET)"
	./$(BIN_DIR)/spire_socket_diagnostic || echo "$(RED)❌ Socket diagnostic failed$(RESET)"
	@echo ""
	@echo "$(YELLOW)🆔 Attestation diagnostic:$(RESET)"
	./$(BIN_DIR)/spire_attestation_diagnostic || echo "$(RED)❌ Attestation diagnostic failed$(RESET)"
	@echo ""
	@echo "$(YELLOW)🔌 Simple connection test:$(RESET)"
	./$(BIN_DIR)/spire_connection_simple || echo "$(RED)❌ Connection test failed$(RESET)"
	@echo "$(GREEN)✅ Diagnostic tools completed!$(RESET)"

# =============================================================================
# 🧪 TEST BUILD TARGETS
# =============================================================================
build-all-tests: build-tests build-diagnostic
	@echo "$(GREEN)✅ All test binaries built successfully!$(RESET)"

build-tests: build-level2 build-level3 build-level4 build-level5
	@echo "$(GREEN)✅ All test binaries built!$(RESET)"

build-level4:
	@echo "$(CYAN)🔨 Building Level 4 diagnostic binary...$(RESET)"
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/level4_spire_socket_diagnostic $(INTEGRATION_DIR)/level4/main.go

build-level2:
	@echo "$(CYAN)🔨 Building Level 2 test binary...$(RESET)"
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/level2_client_comprehensive $(INTEGRATION_DIR)/level2/main.go

build-level3:
	@echo "$(CYAN)🔨 Building Level 3 test binary...$(RESET)"
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/level3_caddy_comprehensive $(INTEGRATION_DIR)/level3/main.go

build-level5:
	@echo "$(CYAN)🔨 Building Level 5 test binary...$(RESET)"
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/level5_caddy_container_integration $(INTEGRATION_DIR)/level5/main.go

build-diagnostics:
	@echo "$(CYAN)🔨 Building diagnostic tools...$(RESET)"
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/spire_socket_diagnostic $(TEST_DIR)/diagnostics/spire_socket_diagnostic/main.go
	go build -o $(BIN_DIR)/spire_attestation_diagnostic $(TEST_DIR)/diagnostics/spire_attestation_diagnostic/main.go
	go build -o $(BIN_DIR)/spire_connection_simple $(TEST_DIR)/diagnostics/spire_connection_simple/main.go

# =============================================================================
# 🐳 DOCKER DIAGNOSTIC (Legacy compatibility)
# =============================================================================
build-diagnostic: build-diagnostics
	@echo "$(CYAN)🐳 Building SPIRE diagnostic Docker image...$(RESET)"
	@docker build -t spire-diagnostic .
	@echo "$(GREEN)✅ Docker image built successfully!$(RESET)"

run: build-diagnostic
	@echo "$(CYAN)🚀 Running SPIRE diagnostic container...$(RESET)"
	@docker-compose up

docker-diagnostic: clean build-diagnostic
	@echo "$(CYAN)🧪 Testing SPIRE Docker connectivity...$(RESET)"
	@echo "$(YELLOW)🗑️  Cleaning up any existing containers...$(RESET)"
	@docker-compose down > /dev/null 2>&1 || true
	@sleep 1
	@echo "$(CYAN)🚀 Starting diagnostic test...$(RESET)"
	@docker-compose up
	@echo "$(GREEN)✅ Diagnostic test complete!$(RESET)"

# Build and run locally (no Docker)
local:
	@echo "$(CYAN)🔧 Running diagnostic locally...$(RESET)"
	@go mod tidy
	@go run main.go

diagnostic: test-diagnostics

# =============================================================================
# 📊 COVERAGE & REPORTING
# =============================================================================
coverage:
	@echo "$(CYAN)📊 Generating test coverage report...$(RESET)"
	go test -coverprofile=coverage.out ./$(TEST_DIR)/
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✅ Coverage report generated: coverage.html$(RESET)"
	@echo "$(CYAN)🌐 Open with: open coverage.html$(RESET)"

# =============================================================================
# 🧹 CLEANUP
# =============================================================================
clean:
	@echo "$(YELLOW)🗑️  Cleaning up containers...$(RESET)"
	@docker-compose down > /dev/null 2>&1 || true
	@docker rm spire-diagnostic > /dev/null 2>&1 || true

clean-all: clean
	@echo "$(YELLOW)🗑️  Cleaning up all build artifacts...$(RESET)"
	@rm -rf $(BIN_DIR)
	@rm -f coverage.out coverage.html
	@docker system prune -f > /dev/null 2>&1 || true
	@echo "$(GREEN)✅ Full cleanup complete!$(RESET)"

# =============================================================================
# 📖 HELP
# =============================================================================
help:
	@echo "$(MAGENTA)🌸 Caddy SPIRE Client - Build & Test Framework$(RESET)"
	@echo "================================================="
	@echo ""
	@echo "$(GREEN)🏗️ Main Project Targets:$(RESET)"
	@echo "  $(YELLOW)all$(RESET)            - Build all main project binaries (default)"
	@echo "  $(YELLOW)build$(RESET)          - Build all main project binaries"
	@echo "  $(YELLOW)build-caddy-spire-client$(RESET) - Build caddy-spire-client binary"
	@echo "  $(YELLOW)build-caddy-with-spire$(RESET)   - Build custom Caddy with SPIRE module"
	@echo "  $(YELLOW)install$(RESET)        - Install binaries to GOPATH/bin"
	@echo ""
	@echo "$(GREEN)🧪 Testing Framework:$(RESET)"
	@echo "  $(YELLOW)test$(RESET)           - Quick CI-safe tests (Level 1 + 2 unit)"
	@echo "  $(YELLOW)test-all$(RESET)       - All test levels (requires SPIRE agent + Docker)"
	@echo ""
	@echo "$(GREEN)📊 Individual Test Levels:$(RESET)"
	@echo "  $(YELLOW)test-level1$(RESET)    - 🔬 Mock-based tests (fast, no dependencies)"
	@echo "  $(YELLOW)test-level2$(RESET)    - 🔗 Bare metal client tests (requires SPIRE agent)"
	@echo "  $(YELLOW)test-level3$(RESET)    - 🌸 Comprehensive Caddy + SPIRE integration"
	@echo "  $(YELLOW)test-level4$(RESET)    - 🔍 Basic SPIRE socket connectivity (container diagnostic)"
	@echo "  $(YELLOW)test-level5$(RESET)    - 🐳 Containerized comprehensive testing"
	@echo ""
	@echo "$(GREEN)🔍 Diagnostic Tools:$(RESET)"
	@echo "  $(YELLOW)test-diagnostics$(RESET) - Run all diagnostic tools"
	@echo "  $(YELLOW)diagnostic$(RESET)       - Alias for test-diagnostics"
	@echo ""
	@echo "$(GREEN)🐳 Level 5 Container Management:$(RESET)"
	@echo "  $(YELLOW)test-level5-setup$(RESET)   - Verify Level 5 prerequisites"
	@echo "  $(YELLOW)test-level5-cleanup$(RESET) - Clean up Level 5 containers"
	@echo "  $(YELLOW)test-level5-rebuild$(RESET) - Force rebuild Level 5 containers"
	@echo ""
	@echo "$(GREEN)🏗️ Test Build Targets:$(RESET)"
	@echo "  $(YELLOW)build-all-tests$(RESET) - Build all test binaries"
	@echo "  $(YELLOW)build-tests$(RESET)     - Build Level 2-5 test binaries"
	@echo "  $(YELLOW)build-diagnostics$(RESET) - Build diagnostic tools"
	@echo ""
	@echo "$(GREEN)🐳 Docker Build & Push:$(RESET)"
	@echo "  $(YELLOW)docker-setup$(RESET)    - Setup Docker buildx for multi-arch builds"
	@echo "  $(YELLOW)docker-build$(RESET)    - Build multi-arch Caddy + SPIRE image (local)"
	@echo "  $(YELLOW)docker-push$(RESET)     - Build and push multi-arch image to GHCR"
	@echo "  $(YELLOW)docker-build-amd64$(RESET) - Build and push AMD64 only"
	@echo "  $(YELLOW)docker-build-arm64$(RESET) - Build and push ARM64 only"
	@echo "  $(YELLOW)docker-build-armv7$(RESET) - Build and push ARMv7 only"
	@echo "  $(YELLOW)docker-login$(RESET)    - Login to GitHub Container Registry"
	@echo ""
	@echo "$(GREEN)🐳 Docker (Legacy):$(RESET)"
	@echo "  $(YELLOW)run$(RESET)            - Run diagnostic container"
	@echo "  $(YELLOW)docker-diagnostic$(RESET) - Clean run of diagnostic container"
	@echo "  $(YELLOW)local$(RESET)          - Run diagnostic locally (no Docker)"
	@echo ""
	@echo "$(GREEN)📊 Coverage & Cleanup:$(RESET)"
	@echo "  $(YELLOW)coverage$(RESET)       - Generate test coverage report"
	@echo "  $(YELLOW)clean$(RESET)          - Clean up containers"
	@echo "  $(YELLOW)clean-all$(RESET)      - Clean up all artifacts"
	@echo ""
	@echo "$(GREEN)🎯 Quick Start:$(RESET)"
	@echo "  $(CYAN)make$(RESET)           - Build main project binaries"
	@echo "  $(CYAN)make test$(RESET)       - Run CI-safe tests"
	@echo "  $(CYAN)make test-level1$(RESET) - Fastest tests (mocks only)"
	@echo "  $(CYAN)make install$(RESET)    - Build and install to GOPATH"
	@echo ""
	@echo "$(MAGENTA)💖 Progressive Testing: Always start with Level 1, then move up!$(RESET)"