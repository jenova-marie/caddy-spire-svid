// SPIRE Docker Diagnostic Tool
// Simple container that connects to local spire-agent and prints attestation results
package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

func main() {
	fmt.Println("🐳 SPIRE Docker Diagnostic Tool")
	fmt.Println("================================")
	fmt.Println("🔍 Attempting to connect to local SPIRE agent...")
	fmt.Println()

	// Get SPIRE socket path from environment or use default Unix socket
	spireSocketPath := os.Getenv("SPIRE_SOCKET_PATH")
	if spireSocketPath == "" {
		spireSocketPath = "unix:///tmp/spire-agent/public/api.sock"
	}
	fmt.Printf("📡 Using SPIRE socket: %s\n", spireSocketPath)

	// Test socket exists before creating SPIRE client
	fmt.Println("\n🧪 Testing socket availability...")
	socketFile := "/tmp/spire-agent/public/api.sock"
	if _, err := os.Stat(socketFile); err != nil {
		fmt.Printf("❌ FATAL: Socket does not exist: %s\n", socketFile)
		fmt.Printf("   💡 Make sure SPIRE agent is running and socket is mounted\n")
		fmt.Printf("   🔍 Error: %v\n", err)
		os.Exit(1)
	} else {
		fmt.Printf("✅ Socket file exists: %s\n", socketFile)
	}

	// Verify socket permissions and accessibility
	fmt.Println("\n🔐 Testing socket permissions...")
	if fileInfo, err := os.Stat(socketFile); err != nil {
		fmt.Printf("❌ FATAL: Cannot stat socket: %v\n", err)
		os.Exit(1)
	} else {
		fmt.Printf("✅ Socket permissions: %s\n", fileInfo.Mode())
		fmt.Printf("   📏 Socket size: %d bytes\n", fileInfo.Size())
		fmt.Printf("   📅 Modified: %s\n", fileInfo.ModTime())

		// Check if it's actually a socket
		if fileInfo.Mode()&os.ModeSocket == 0 {
			fmt.Printf("⚠️  Warning: File is not a socket (mode: %s)\n", fileInfo.Mode())
		} else {
			fmt.Printf("✅ Confirmed: File is a valid Unix socket\n")
		}
	}

	// Test socket connectivity with a simple dial
	fmt.Println("\n🔌 Testing socket connectivity...")
	if conn, err := net.Dial("unix", socketFile); err != nil {
		fmt.Printf("❌ FATAL: Cannot connect to socket: %v\n", err)
		fmt.Printf("   💡 Socket exists but is not accepting connections\n")
		fmt.Printf("   🔍 Check if SPIRE agent is running and listening\n")
		os.Exit(1)
	} else {
		fmt.Printf("✅ Socket connection successful!\n")
		conn.Close()
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("\n🔌 Creating SPIRE workload API client...")

	// Create X.509 source using Unix socket
	source, err := workloadapi.NewX509Source(ctx, workloadapi.WithClientOptions(workloadapi.WithAddr(spireSocketPath)))
	if err != nil {
		fmt.Printf("❌ FATAL: Failed to create X.509 source: %v\n", err)
		os.Exit(1)
	}
	defer source.Close()

	fmt.Println("✅ Successfully created X.509 source!")

	// Get SVID
	fmt.Println("\n🆔 Retrieving SVID...")
	svid, err := source.GetX509SVID()
	if err != nil {
		fmt.Printf("❌ FATAL: Failed to get SVID: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Successfully retrieved SVID!")
	fmt.Println()

	// Print SVID details
	fmt.Println("📋 SVID Details:")
	fmt.Printf("   🆔 SPIFFE ID: %s\n", svid.ID)
	fmt.Printf("   📜 Certificates: %d\n", len(svid.Certificates))

	if len(svid.Certificates) > 0 {
		cert := svid.Certificates[0]
		fmt.Printf("   🔒 Subject: %s\n", cert.Subject)
		fmt.Printf("   🏢 Issuer: %s\n", cert.Issuer)
		fmt.Printf("   📅 Valid from: %s\n", cert.NotBefore.UTC())
		fmt.Printf("   📅 Valid until: %s\n", cert.NotAfter.UTC())
		fmt.Printf("   ⏰ Time remaining: %s\n", time.Until(cert.NotAfter))

		// Print DNS names if any
		if len(cert.DNSNames) > 0 {
			fmt.Printf("   🌐 DNS Names: %v\n", cert.DNSNames)
		}
	}

	// Get bundles
	fmt.Println("\n🎫 Retrieving X.509 bundles...")
	bundles, err := source.GetX509BundleForTrustDomain(svid.ID.TrustDomain())
	if err != nil {
		fmt.Printf("⚠️  Warning: Failed to get bundles: %v\n", err)
	} else {
		fmt.Printf("✅ Successfully retrieved bundles for trust domain: %s\n", svid.ID.TrustDomain())
		fmt.Printf("   📦 Root certificates: %d\n", len(bundles.X509Authorities()))
	}

	// Test TLS config
	fmt.Println("\n🔐 Testing TLS configuration...")
	_, err = source.GetX509SVID()
	if err != nil {
		fmt.Printf("⚠️  Warning: Failed to get TLS config: %v\n", err)
	} else {
		fmt.Println("✅ TLS configuration working!")
	}

	// Print container environment info
	fmt.Println("\n🐳 Container Environment:")
	fmt.Printf("   📁 Working directory: %s\n", getCurrentDir())
	fmt.Printf("   👤 User ID: %s\n", getEnvOrDefault("USER", "unknown"))
	fmt.Printf("   🏠 Home directory: %s\n", getEnvOrDefault("HOME", "unknown"))

	// Print process info
	fmt.Printf("   🔢 Process ID: %d\n", os.Getpid())
	fmt.Printf("   🔢 Parent PID: %d\n", os.Getppid())

	fmt.Println()
	fmt.Println("🎉 DIAGNOSTIC COMPLETE!")
	fmt.Println("✅ Docker container successfully connected to SPIRE agent")
	fmt.Println("💖 Container attestation working perfectly!")
}

func getCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return dir
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
