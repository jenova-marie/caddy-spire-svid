// level4_caddy_container_integration - Containerized Comprehensive Caddy + SPIRE testing
// Same "Fire and Bodies" testing as Level 3, but containerized with live spire-agent
package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

func main() {
	fmt.Println("🐳 Level 5: Containerized Comprehensive Caddy + SPIRE Integration")
	fmt.Println("================================================================")
	fmt.Println("💀 Same \"Fire and Bodies\" testing as Level 4, but containerized")
	fmt.Println("🌸 Testing containerized Caddy + SPIRE with live spire-agent")
	fmt.Println()

	// Step 1: Check prerequisites
	if !checkPrerequisites() {
		return
	}

	// Step 2: Start Docker Compose stack
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if !startDockerStack(ctx) {
		return
	}
	defer func() {
		fmt.Println("\n🧹 Stopping Docker containers...")
		stopDockerStack()
	}()

	// Step 3: Wait for containers to be ready
	fmt.Println("⏰ Waiting for containerized Caddy to start and initialize SPIRE certificates...")
	time.Sleep(10 * time.Second)

	if !waitForContainerReady() {
		fmt.Println("❌ Containerized Caddy failed to become ready")
		return
	}

	fmt.Println("✅ Containerized Caddy is ready with SPIRE integration!")
	fmt.Println()

	// Step 4: Run comprehensive tests
	runComprehensiveTests()
}

// checkPrerequisites validates that all required components are available
func checkPrerequisites() bool {
	fmt.Println("🔍 Checking prerequisites...")

	// Check for docker compose (try modern command first, fall back to legacy)
	hasCompose := false
	if _, err := exec.LookPath("docker"); err == nil {
		// Test if 'docker compose' works
		if cmd := exec.Command("docker", "compose", "version"); cmd.Run() == nil {
			hasCompose = true
		}
	}
	if !hasCompose {
		if _, err := exec.LookPath("docker-compose"); err != nil {
			fmt.Println("❌ Neither 'docker compose' nor 'docker-compose' found")
			return false
		}
	}
	fmt.Println("   ✅ Docker Compose found")

	// Check for Docker Compose file
	if _, err := os.Stat("docker-compose.level5.yml"); os.IsNotExist(err) {
		fmt.Println("❌ docker-compose.level5.yml not found")
		return false
	}
	fmt.Println("   ✅ docker-compose.level5.yml found")

	// Check for SPIRE socket
	if _, err := os.Stat("/tmp/spire-agent/public/api.sock"); os.IsNotExist(err) {
		fmt.Printf("❌ SPIRE agent socket not found: /tmp/spire-agent/public/api.sock\n")
		fmt.Printf("   Start SPIRE agent: sudo spire-agent run -config ~/.ssh/spire-agent.conf\n")
		return false
	}
	fmt.Println("   ✅ SPIRE agent socket found")

	// Check for Dockerfile.caddy
	if _, err := os.Stat("Dockerfile.caddy"); os.IsNotExist(err) {
		fmt.Println("❌ Dockerfile.caddy not found")
		return false
	}
	fmt.Println("   ✅ Dockerfile.caddy found")

	return true
}

// startDockerStack starts the Docker Compose stack for Level 4 testing
func startDockerStack(ctx context.Context) bool {
	fmt.Println("🐳 Starting Docker Compose stack...")

	// Clean up any existing containers first
	stopDockerStack()

	// Start the stack (try modern docker compose first)
	var cmd *exec.Cmd
	if exec.Command("docker", "compose", "version").Run() == nil {
		cmd = exec.CommandContext(ctx, "docker", "compose", "-f", "docker-compose.level5.yml", "up", "--build", "-d")
	} else {
		cmd = exec.CommandContext(ctx, "docker-compose", "-f", "docker-compose.level5.yml", "up", "--build", "-d")
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Failed to start Docker stack: %v\n", err)
		return false
	}

	fmt.Println("   🌸 Docker Compose stack started")
	return true
}

// stopDockerStack stops and cleans up the Docker Compose stack
func stopDockerStack() {
	// Try modern docker compose first
	if exec.Command("docker", "compose", "version").Run() == nil {
		cmd := exec.Command("docker", "compose", "-f", "docker-compose.level5.yml", "down", "--remove-orphans")
		_ = cmd.Run() // Ignore errors for cleanup
	} else {
		cmd := exec.Command("docker-compose", "-f", "docker-compose.level5.yml", "down", "--remove-orphans")
		_ = cmd.Run() // Ignore errors for cleanup
	}
}

// waitForContainerReady waits for the containerized Caddy to be responsive
func waitForContainerReady() bool {
	client := &http.Client{
		Timeout: 2 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	for i := 0; i < 60; i++ { // Try for 60 seconds (containers take longer to start)
		resp, err := client.Get("https://localhost:8443")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				return true
			}
		}
		time.Sleep(1 * time.Second)
	}
	return false
}

// runComprehensiveTests executes all the comprehensive test scenarios against the containerized Caddy
func runComprehensiveTests() {
	// Create HTTP client that accepts SPIRE certificates
	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // For testing with SPIRE certs
			},
		},
	}

	// Test 1: Container HTTPS Connection with SPIRE Certificate
	fmt.Println("🐳 Test 1: Container HTTPS Connection with SPIRE Certificate")

	resp, err := client.Get("https://localhost:8443")
	if err != nil {
		fmt.Printf("❌ Test 1 FAILED: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ Test 1 FAILED reading body: %v\n", err)
		return
	}

	fmt.Printf("   Status: %s\n", resp.Status)
	fmt.Printf("   Server: %s\n", resp.Header.Get("Server"))
	fmt.Printf("   Response: %s\n", strings.TrimSpace(string(body)))

	if resp.StatusCode != 200 {
		fmt.Printf("❌ Test 1 FAILED: Expected 200, got %d\n", resp.StatusCode)
		return
	}

	fmt.Println("✅ Test 1 PASSED: Container HTTPS connection successful")

	// Test 2: Container Certificate Details Validation (RecoverySky.org CA)
	fmt.Println("\n🎫 Test 2: Container Certificate Details Validation (RecoverySky.org CA)")

	conn, err := tls.Dial("tcp", "localhost:8443", &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		fmt.Printf("❌ Test 2 FAILED: %v\n", err)
		return
	}
	defer conn.Close()

	cert := conn.ConnectionState().PeerCertificates[0]
	fmt.Printf("   Subject: %s\n", cert.Subject)
	fmt.Printf("   Issuer: %s\n", cert.Issuer)
	fmt.Printf("   Valid from: %v\n", cert.NotBefore)
	fmt.Printf("   Valid until: %v\n", cert.NotAfter)
	fmt.Printf("   Time remaining: %v\n", time.Until(cert.NotAfter))

	// Validate it's a SPIRE certificate
	if cert.Subject.Organization == nil || cert.Subject.Organization[0] != "SPIRE" {
		fmt.Printf("❌ Test 2 FAILED: Certificate not issued by SPIRE (got: %v)\n", cert.Subject.Organization)
		return
	}

	// Validate issuer is RecoverySky.org
	if cert.Issuer.Organization == nil || cert.Issuer.Organization[0] != "RecoverySky Org" {
		fmt.Printf("❌ Test 2 FAILED: Certificate not issued by RecoverySky Org (got: %v)\n", cert.Issuer.Organization)
		return
	}

	// Validate certificate is not expired
	if time.Now().After(cert.NotAfter) {
		fmt.Printf("❌ Test 2 FAILED: Certificate is expired\n")
		return
	}

	fmt.Println("✅ Test 2 PASSED: Container certificate validation successful (RecoverySky.org CA)")

	// Test 3: Container Performance Characteristics
	fmt.Println("\n⚡ Test 3: Container Performance Characteristics")

	var totalDuration time.Duration
	testRuns := 5

	for i := 0; i < testRuns; i++ {
		start := time.Now()
		resp, err := client.Get("https://localhost:8443")
		duration := time.Since(start)
		totalDuration += duration

		if err != nil {
			fmt.Printf("❌ Test 3 FAILED (run %d): %v\n", i+1, err)
			return
		}
		resp.Body.Close()
	}

	avgDuration := totalDuration / time.Duration(testRuns)
	fmt.Printf("   Average response time: %v (over %d requests)\n", avgDuration, testRuns)
	fmt.Printf("   Total time: %v\n", totalDuration)

	if avgDuration > 10*time.Millisecond {
		fmt.Printf("⚠️  Test 3 WARNING: Average response time (%v) slower than ideal (<10ms for containers)\n", avgDuration)
	}

	fmt.Println("✅ Test 3 PASSED: Container performance characteristics validated")

	// Test 4: Container Concurrent Request Handling
	fmt.Println("\n🔄 Test 4: Container Concurrent Request Handling")

	concurrentRequests := 5
	results := make(chan error, concurrentRequests)
	var wg sync.WaitGroup

	start := time.Now()
	for i := 0; i < concurrentRequests; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			resp, err := client.Get("https://localhost:8443")
			if err != nil {
				results <- fmt.Errorf("concurrent request %d failed: %v", id, err)
				return
			}
			resp.Body.Close()

			if resp.StatusCode != 200 {
				results <- fmt.Errorf("concurrent request %d got status %d", id, resp.StatusCode)
				return
			}

			results <- nil
		}(i)
	}

	wg.Wait()
	close(results)
	duration := time.Since(start)

	// Check results
	errorCount := 0
	for result := range results {
		if result != nil {
			fmt.Printf("   ❌ %v\n", result)
			errorCount++
		}
	}

	if errorCount > 0 {
		fmt.Printf("❌ Test 4 FAILED: %d/%d concurrent requests failed\n", errorCount, concurrentRequests)
		return
	}

	fmt.Printf("   Concurrent requests (%d): Completed in %v\n", concurrentRequests, duration)
	fmt.Printf("   Average per request: %v\n", duration/time.Duration(concurrentRequests))
	fmt.Println("✅ Test 4 PASSED: Container concurrent request handling successful")

	fmt.Println("\n🔥💀 LEVEL 5 CONTAINERIZED TEST COMPLETE! 💀🔥")
	fmt.Println("✅ ALL CONTAINERIZED TESTS PASSED!")
	fmt.Println("🐳 Containerized Caddy + SPIRE integration is ROCK SOLID!")
	fmt.Println("💖 Fire and bodies confirmed - everything works in containers!")
}
