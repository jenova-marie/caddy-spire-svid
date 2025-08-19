package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

func main() {
	socketPath := "/tmp/spire-agent/public/api.sock"

	fmt.Printf("🧪 Basic Unix Socket Connectivity Test\n")
	fmt.Printf("Socket: %s\n", socketPath)

	// Check if socket exists
	if stat, err := os.Stat(socketPath); err != nil {
		fmt.Printf("❌ Socket doesn't exist: %v\n", err)
		os.Exit(1)
	} else {
		fmt.Printf("✅ Socket exists\n")
		fmt.Printf("   Mode: %s\n", stat.Mode())
		fmt.Printf("   Size: %d\n", stat.Size())
	}

	// Try to connect
	fmt.Printf("\n🔌 Attempting connection...\n")
	conn, err := net.DialTimeout("unix", socketPath, 5*time.Second)
	if err != nil {
		fmt.Printf("❌ Connection failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Connected successfully!\n")
	conn.Close()
	fmt.Printf("✅ Connection closed cleanly\n")
	fmt.Printf("🎉 Basic socket connectivity works!\n")
}
