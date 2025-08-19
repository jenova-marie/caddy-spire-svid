# SPIRE Docker Diagnostic Tool
# Simple testing of Docker + SPIRE connectivity
# Plus full orchestration with containerized SPIRE agent

.PHONY: build run test clean local help build-agent full-test full-clean

# Colors for output
CYAN = \033[36m
GREEN = \033[32m
YELLOW = \033[33m
RED = \033[31m
RESET = \033[0m

# Build the Docker image
build:
	@echo "$(CYAN)🐳 Building SPIRE diagnostic Docker image...$(RESET)"
	@docker build -t spire-diagnostic .
	@echo "$(GREEN)✅ Docker image built successfully!$(RESET)"

# Run the diagnostic container
run: build
	@echo "$(CYAN)🚀 Running SPIRE diagnostic container...$(RESET)"
	@docker-compose up

# Clean run - remove any existing container first
test: clean build
	@echo "$(CYAN)🧪 Testing SPIRE Docker connectivity...$(RESET)"
	@echo "$(YELLOW)🗑️  Cleaning up any existing containers...$(RESET)"
	@docker-compose down > /dev/null 2>&1 || true
	@sleep 1
	@echo "$(CYAN)🚀 Starting diagnostic test...$(RESET)"
	@docker-compose up
	@echo "$(GREEN)✅ Diagnostic test complete!$(RESET)"

# Clean up containers and images
clean:
	@echo "$(YELLOW)🗑️  Cleaning up containers...$(RESET)"
	@docker-compose down > /dev/null 2>&1 || true
	@docker rm spire-diagnostic > /dev/null 2>&1 || true
	@echo "$(GREEN)✅ Cleanup complete!$(RESET)"

# Build and run locally (no Docker)
local:
	@echo "$(CYAN)🔧 Running diagnostic locally...$(RESET)"
	@go mod tidy
	@go run main.go

# Build SPIRE agent container
build-agent:
	@echo "$(CYAN)🏗️  Building SPIRE agent Docker image...$(RESET)"
	@docker build -f Dockerfile.agent -t spire-agent-docker .
	@echo "$(GREEN)✅ SPIRE agent image built successfully!$(RESET)"

# Simple SPIRE agent + diagnostic test ONLY  
full-test: full-clean
	@echo "$(CYAN)🚀 Starting SPIRE agent + diagnostic test...$(RESET)"
	@echo "$(YELLOW)💫 This solves macOS Docker Desktop Unix socket issues!$(RESET)"
	@docker-compose -f docker-compose.full-orchestration.yml up --build
	@echo "$(CYAN)⏳ Waiting for services to start...$(RESET)"
	@sleep 30
	@echo "$(CYAN)🧪 Running full orchestration diagnostic...$(RESET)"
	@docker-compose -f docker-compose.full-orchestration.yml logs spire-diagnostic-full
	@echo "$(GREEN)✅ Full orchestration test complete!$(RESET)"
	@echo "$(YELLOW)🌐 Test Caddy at: https://localhost:8443$(RESET)"

# Clean up full orchestration
full-clean:
	@echo "$(YELLOW)🗑️  Cleaning up full orchestration...$(RESET)"
	@docker-compose -f docker-compose.full-orchestration.yml down -v > /dev/null 2>&1 || true
	@docker system prune -f > /dev/null 2>&1 || true
	@echo "$(GREEN)✅ Full cleanup complete!$(RESET)"

# Show help
help:
	@echo "$(CYAN)SPIRE Docker Diagnostic Tool$(RESET)"
	@echo "============================"
	@echo ""
	@echo "$(GREEN)Basic Diagnostic Tests:$(RESET)"
	@echo "  $(YELLOW)build$(RESET)      - Build diagnostic Docker image"
	@echo "  $(YELLOW)run$(RESET)        - Run diagnostic container"
	@echo "  $(YELLOW)test$(RESET)       - Clean run of diagnostic"
	@echo "  $(YELLOW)local$(RESET)      - Run diagnostic locally (no Docker)"
	@echo "  $(YELLOW)clean$(RESET)      - Clean up containers"
	@echo ""
	@echo "$(GREEN)Full Orchestration (Solves macOS Issues):$(RESET)"
	@echo "  $(YELLOW)build-agent$(RESET) - Build SPIRE agent container"
	@echo "  $(YELLOW)full-test$(RESET)   - Full SPIRE + Caddy orchestration"
	@echo "  $(YELLOW)full-clean$(RESET)  - Clean up full orchestration"
	@echo ""
	@echo "  $(YELLOW)help$(RESET)       - Show this help"
