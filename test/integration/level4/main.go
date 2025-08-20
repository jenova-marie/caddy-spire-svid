// test0_spire_socket_diagnostic - Basic SPIRE socket connectivity test
// Simple container test to verify Docker + SPIRE socket mounting works
package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

func main() {
	fmt.Println("🔍 Level 4: Basic SPIRE Socket Connectivity Test")
	fmt.Println("==============================================")
	fmt.Println("🐳 Testing Docker container access to host SPIRE socket")
	fmt.Println()

	// Step 1: Check prerequisites
	if !checkPrerequisites() {
		return
	}

	// Step 2: Run the diagnostic container
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if !runDiagnosticContainer(ctx) {
		fmt.Println("❌ Level 4 FAILED!")
		fmt.Println("   Check the container logs above for details")
		os.Exit(1)
	}

	fmt.Println("🎉 Level 4 completed successfully!")
	fmt.Println("✅ Docker container can access host SPIRE socket")
	fmt.Println("✅ SPIRE agent CLI works in container")
	fmt.Println("✅ Container workload attestation successful")
}

// checkPrerequisites validates that all required components are available
func checkPrerequisites() bool {
	fmt.Println("🔍 Checking prerequisites...")

	// Check for Docker
	if _, err := exec.LookPath("docker"); err != nil {
		fmt.Println("❌ Docker not found in PATH")
		return false
	}
	fmt.Println("   ✅ Docker found")

	// Check for Docker Compose
	hasCompose := false
	if exec.Command("docker", "compose", "version").Run() == nil {
		hasCompose = true
	} else if _, err := exec.LookPath("docker-compose"); err == nil {
		hasCompose = true
	}
	if !hasCompose {
		fmt.Println("❌ Neither 'docker compose' nor 'docker-compose' found")
		return false
	}
	fmt.Println("   ✅ Docker Compose found")

	// Check for Docker Compose file
	if _, err := os.Stat("docker-compose.level4.yml"); os.IsNotExist(err) {
		fmt.Println("❌ docker-compose.level4.yml not found")
		return false
	}
	fmt.Println("   ✅ docker-compose.level4.yml found")

	// Check for SPIRE socket
	if _, err := os.Stat("/tmp/spire-agent/public/api.sock"); os.IsNotExist(err) {
		fmt.Printf("❌ SPIRE agent socket not found: /tmp/spire-agent/public/api.sock\n")
		fmt.Printf("   Start SPIRE agent: sudo spire-agent run -config ~/.ssh/spire-agent.conf\n")
		return false
	}
	fmt.Println("   ✅ SPIRE agent socket found")

	// Check for Dockerfile
	if _, err := os.Stat("Dockerfile.spire-test"); os.IsNotExist(err) {
		fmt.Println("❌ Dockerfile.spire-test not found")
		return false
	}
	fmt.Println("   ✅ Dockerfile.spire-test found")

	return true
}

// runDiagnosticContainer runs the SPIRE socket diagnostic container
func runDiagnosticContainer(ctx context.Context) bool {
	fmt.Println("🐳 Starting SPIRE socket diagnostic container...")

	// Clean up any existing containers first
	cleanupContainer()

	// Run the diagnostic container with proper exit code propagation
	var cmd *exec.Cmd
	if exec.Command("docker", "compose", "version").Run() == nil {
		cmd = exec.CommandContext(ctx, "docker", "compose", "-f", "docker-compose.level4.yml", "up", "--build", "--exit-code-from", "caddy-spire-svid-test")
	} else {
		cmd = exec.CommandContext(ctx, "docker-compose", "-f", "docker-compose.level4.yml", "up", "--build", "--exit-code-from", "caddy-spire-svid-test")
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	fmt.Println("   🌸 Diagnostic container completed")

	// Clean up
	cleanupContainer()

	if err != nil {
		fmt.Printf("❌ Diagnostic container failed: %v\n", err)
		fmt.Println("   Check container logs above for details")
		return false
	}

	return true
}

// cleanupContainer stops and removes any existing test containers
func cleanupContainer() {
	// Try modern docker compose first
	if exec.Command("docker", "compose", "version").Run() == nil {
		cmd := exec.Command("docker", "compose", "-f", "docker-compose.level4.yml", "down", "--remove-orphans")
		_ = cmd.Run() // Ignore errors for cleanup
	} else {
		cmd := exec.Command("docker-compose", "-f", "docker-compose.level4.yml", "down", "--remove-orphans")
		_ = cmd.Run() // Ignore errors for cleanup
	}
}
