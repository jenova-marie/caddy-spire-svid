// caddy-spire-client is a demonstration tool that shows how to integrate
// SPIFFE/SPIRE workload API with Caddy server for automated certificate management.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spiffe/go-spiffe/v2/svid/x509svid"

	"github.com/jenova-marie/caddy-spire-client/pkg/spire"
)

const (
	defaultSocketPath      = "/tmp/spire-agent/public/api.sock"
	defaultRefreshInterval = 30 * time.Second
)

var (
	socketPath      = flag.String("socket", defaultSocketPath, "Path to SPIRE agent socket")
	refreshInterval = flag.Duration("refresh", defaultRefreshInterval, "SVID refresh interval")
	showVersion     = flag.Bool("version", false, "Show version information")
	help            = flag.Bool("help", false, "Show help information")
)

// Version information (set by build process)
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

func main() {
	flag.Parse()

	if *help {
		printHelp()
		return
	}

	if *showVersion {
		printVersion()
		return
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal...")
		cancel()
	}()

	// Create SPIRE client
	config := spire.Config{
		SocketPath:      *socketPath,
		RefreshInterval: *refreshInterval,
	}

	client, err := spire.NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create SPIRE client: %v", err)
	}
	defer client.Close()

	log.Printf("🌸 Caddy SPIRE Client started successfully!")
	log.Printf("Socket: %s", *socketPath)
	log.Printf("Refresh interval: %v", *refreshInterval)

	// Get TLS config that can be used by Caddy
	tlsConfig := client.GetTLSConfig()
	if tlsConfig == nil {
		log.Fatal("Failed to get TLS configuration")
	}

	log.Println("TLS configuration ready for Caddy integration")
	log.Println("Current SVID:")
	var svid *x509svid.SVID
	svid, err = client.GetCurrentSVID()
	if err != nil {
		log.Printf("Warning: Could not get current SVID: %v", err)
	} else if len(svid.Certificates) > 0 {
		cert := svid.Certificates[0]
		log.Printf("  Subject=%s, Expires=%v", cert.Subject, cert.NotAfter)
		log.Printf("  SPIFFE ID=%s", svid.ID)
	}

	// In a real Caddy integration, you would pass tlsConfig to Caddy's
	// certificate management system here. For this demo, we just keep running.
	log.Println("✨ Client running... Press Ctrl+C to stop")

	// Wait for shutdown signal
	<-ctx.Done()
	log.Println("💖 Caddy SPIRE Client shutting down gracefully...")
}

func printVersion() {
	fmt.Printf("caddy-spire-client %s\n", Version)
	fmt.Printf("Git commit: %s\n", GitCommit)
	fmt.Printf("Build time: %s\n", BuildTime)
}

func printHelp() {
	fmt.Println("🌸 Caddy SPIRE Client - SPIFFE/SPIRE integration for Caddy")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Printf("  %s [OPTIONS]\n", os.Args[0])
	fmt.Println()
	fmt.Println("OPTIONS:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Printf("  %s\n", os.Args[0])
	fmt.Printf("  %s -socket /custom/path/api.sock -refresh 1m\n", os.Args[0])
	fmt.Println()
	fmt.Println("For more information, visit: https://github.com/jenova-marie/caddy-spire-client")
}
