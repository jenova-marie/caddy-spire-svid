# 🌸 Caddy Module Integration Guide

This guide shows how to integrate the `caddy-spire-client` with Caddy as a custom module.

## Creating a Caddy Module

Here's an example of how to create a Caddy module that uses SPIFFE/SPIRE certificates:

```go
package caddyspire

import (
    "context"
    "time"

    "github.com/caddyserver/caddy/v2"
    "github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
    "github.com/caddyserver/caddy/v2/modules/caddytls"
    "github.com/jenova-marie/caddy-spire-client/pkg/spire"
)

func init() {
    caddy.RegisterModule(SpireModule{})
}

// SpireModule implements a Caddy module for SPIFFE/SPIRE integration
type SpireModule struct {
    SocketPath      string        `json:"socket_path,omitempty"`
    RefreshInterval time.Duration `json:"refresh_interval,omitempty"`
    
    client *spire.Client
}

// CaddyModule returns the Caddy module information
func (SpireModule) CaddyModule() caddy.ModuleInfo {
    return caddy.ModuleInfo{
        ID:  "tls.certificates.spire",
        New: func() caddy.Module { return new(SpireModule) },
    }
}

// Provision sets up the module
func (s *SpireModule) Provision(ctx caddy.Context) error {
    if s.SocketPath == "" {
        s.SocketPath = "/tmp/spire-agent/public/api.sock"
    }
    if s.RefreshInterval == 0 {
        s.RefreshInterval = 30 * time.Second
    }

    config := spire.Config{
        SocketPath:      s.SocketPath,
        RefreshInterval: s.RefreshInterval,
    }

    client, err := spire.NewClient(config)
    if err != nil {
        return err
    }

    s.client = client
    return nil
}

// Cleanup shuts down the module
func (s *SpireModule) Cleanup() error {
    if s.client != nil {
        return s.client.Close()
    }
    return nil
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler
func (s *SpireModule) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
    for d.Next() {
        for d.NextBlock(0) {
            switch d.Val() {
            case "socket_path":
                if !d.NextArg() {
                    return d.ArgErr()
                }
                s.SocketPath = d.Val()
            case "refresh_interval":
                if !d.NextArg() {
                    return d.ArgErr()
                }
                dur, err := time.ParseDuration(d.Val())
                if err != nil {
                    return d.Errf("invalid duration: %v", err)
                }
                s.RefreshInterval = dur
            default:
                return d.Errf("unknown subdirective: %s", d.Val())
            }
        }
    }
    return nil
}

// GetCertificate implements the certificate provider interface
func (s *SpireModule) GetCertificate(ctx context.Context, hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
    tlsConfig := s.client.GetTLSConfig()
    return tlsConfig.GetCertificate(hello)
}

// Interface guards
var (
    _ caddy.Provisioner     = (*SpireModule)(nil)
    _ caddy.CleanerUpper    = (*SpireModule)(nil)
    _ caddyfile.Unmarshaler = (*SpireModule)(nil)
    _ caddytls.CertificateLoader = (*SpireModule)(nil)
)
```

## Caddyfile Configuration

```caddyfile
https://example.com {
    tls {
        certificates spire {
            socket_path /tmp/spire-agent/public/api.sock
            refresh_interval 30s
        }
    }
    
    respond "Hello from SPIFFE-secured Caddy! 🌸"
}
```

## JSON Configuration

```json
{
    "apps": {
        "http": {
            "servers": {
                "example": {
                    "listen": [":443"],
                    "routes": [
                        {
                            "match": [{"host": ["example.com"]}],
                            "handle": [
                                {
                                    "handler": "static_response",
                                    "body": "Hello from SPIFFE-secured Caddy! 🌸"
                                }
                            ]
                        }
                    ]
                }
            }
        },
        "tls": {
            "certificates": {
                "spire": {
                    "socket_path": "/tmp/spire-agent/public/api.sock",
                    "refresh_interval": "30s"
                }
            }
        }
    }
}
```

## Building Caddy with the Module

Create a `main.go` file for your custom Caddy build:

```go
package main

import (
    caddycmd "github.com/caddyserver/caddy/v2/cmd"
    _ "github.com/caddyserver/caddy/v2/modules/standard"
    _ "your-module-path/caddyspire"
)

func main() {
    caddycmd.Main()
}
```

Build your custom Caddy:

```bash
go mod init caddy-with-spire
go get github.com/caddyserver/caddy/v2
go get github.com/jenova-marie/caddy-spire-client
go build -o caddy-spire
```

💖 Now you have a Caddy server with SPIFFE/SPIRE certificate management!
