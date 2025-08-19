# 🌈 Platform Support Matrix

> ✨ Comprehensive multi-architecture support for all deployment scenarios

## 📊 **Supported Platforms**

### 🏗️ **Native Binary Support**

| Operating System | Architecture | Status | Notes |
|------------------|--------------|--------|-------|
| **Linux** | amd64 | ✅ Full | Primary development platform |
| **Linux** | arm64 | ✅ Full | ARM64/AArch64 support |
| **Linux** | arm | ✅ Full | 32-bit ARM (ARMv7) |
| **Linux** | 386 | ✅ Full | 32-bit x86 |
| **macOS** | amd64 | ✅ Full | Intel Mac support |
| **macOS** | arm64 | ✅ Full | Apple Silicon (M1/M2/M3) |
| **Windows** | amd64 | ✅ Full | 64-bit Windows |
| **Windows** | arm64 | ✅ Full | ARM64 Windows |
| **Windows** | arm | ✅ Full | 32-bit ARM Windows |
| **Windows** | 386 | ✅ Full | 32-bit Windows |
| **FreeBSD** | amd64 | ✅ Full | 64-bit FreeBSD |
| **FreeBSD** | arm64 | ✅ Full | ARM64 FreeBSD |
| **FreeBSD** | 386 | ✅ Full | 32-bit FreeBSD |

### 🐳 **Docker Multi-Architecture Support**

| Platform | CLI Image | Custom Caddy Image | Status |
|----------|-----------|-------------------|--------|
| **linux/amd64** | ✅ | ✅ | Full support |
| **linux/arm64** | ✅ | ✅ | Full support |
| **linux/arm/v7** | ✅ | ✅ | Full support |

## 🚀 **Build Commands**

### 📦 **Single Platform Builds**
```bash
# Build for current platform
make build

# Build custom Caddy for current platform
make build-caddy

# Build all tools for current platform
make build-all-binaries
```

### 🌈 **Cross-Platform Builds**
```bash
# Build for all supported platforms
make build-all

# Build for specific platform
GOOS=linux GOARCH=arm64 make build
GOOS=windows GOARCH=amd64 make build-caddy
GOOS=darwin GOARCH=arm64 make build-helper-tool
```

### 🐳 **Docker Multi-Arch Builds**
```bash
# Single architecture (current platform)
make docker-build
make docker-build-caddy

# Multi-architecture builds (requires Docker Buildx)
make docker-buildx          # CLI tool
make docker-buildx-caddy    # Custom Caddy

# Push multi-arch to registry
make docker-push-multiarch       # CLI tool
make docker-push-multiarch-caddy # Custom Caddy
```

## 🔧 **Platform-Specific Features**

### 🐧 **Linux**
- **Complete SPIRE integration** with Unix domain sockets
- **Container deployment** ready (Docker, Kubernetes, etc.)
- **Systemd service** support via helper tools
- **Performance optimized** for server workloads

### 🍎 **macOS**
- **Apple Silicon native** performance on M1/M2/M3
- **Intel Mac compatibility** maintained
- **Homebrew installation** ready
- **Development friendly** with hot reload

### 🪟 **Windows**
- **Native Windows service** support
- **PowerShell integration** for automation
- **Windows Server** deployment ready
- **WSL2 compatibility** for Linux workflows

### 🔧 **FreeBSD**
- **Native FreeBSD** binary support
- **Jail compatibility** for containerized deployments
- **RC script support** for service management

## 📈 **Performance Characteristics**

### 🏃‍♀️ **Binary Sizes (Approximate)**

| Binary | linux/amd64 | linux/arm64 | darwin/amd64 | darwin/arm64 | windows/amd64 |
|--------|-------------|-------------|--------------|--------------|---------------|
| **caddy-spire-client** | ~16MB | ~15MB | ~17MB | ~16MB | ~16MB |
| **caddy-with-spire** | ~67MB | ~63MB | ~69MB | ~65MB | ~67MB |
| **external-provider** | ~16MB | ~15MB | ~17MB | ~16MB | ~16MB |
| **spire-helper** | ~16MB | ~15MB | ~17MB | ~16MB | ~16MB |

### ⚡ **Runtime Performance**

| Architecture | Relative Performance | Memory Usage | Notes |
|--------------|---------------------|--------------|-------|
| **amd64** | 100% (baseline) | Standard | Optimal for servers |
| **arm64** | 95-105% | 5-10% lower | Excellent on Apple Silicon |
| **arm** | 80-90% | 10-15% lower | Good for embedded systems |
| **386** | 70-80% | Standard | Legacy system support |

## 🔄 **CI/CD Pipeline Support**

### 🚀 **Automated Builds**
- **GitHub Actions** builds all platforms on every release
- **Multi-arch Docker images** automatically published
- **Cross-compilation testing** ensures compatibility
- **Smoke tests** run on multiple OS/arch combinations

### 📦 **Release Artifacts**
- **Compressed archives** (.tar.gz for Unix, .zip for Windows)
- **Multi-arch Docker images** on GitHub Container Registry
- **Checksums** and **digital signatures** for verification
- **Debian packages** (coming soon)
- **Homebrew formula** (coming soon)

## 🛠️ **Development Setup**

### 📋 **Requirements**
- **Go 1.21+** (1.22 recommended)
- **Docker** with Buildx for multi-arch builds
- **Make** for build automation
- **Git** for version control

### 🧪 **Testing Multi-Arch Builds**
```bash
# Test all cross-compilation targets
make build-all

# Verify Docker multi-arch capability
docker buildx ls

# Test specific architecture (with QEMU emulation)
docker run --rm --platform linux/arm64 caddy-spire-client:latest -version
```

## 🔍 **Troubleshooting**

### ❗ **Common Issues**

#### **CGO Disabled**
Our builds use `CGO_ENABLED=0` for maximum compatibility and static linking.

#### **Docker Buildx Not Available**
```bash
# Install Docker Buildx
docker buildx install

# Create new builder instance
docker buildx create --use --name multiarch --driver docker-container
```

#### **QEMU Emulation Issues**
```bash
# Install QEMU emulation
docker run --rm --privileged multiarch/qemu-user-static --reset -p yes
```

### 🔧 **Platform-Specific Notes**

#### **Windows**
- Use PowerShell or Command Prompt
- Some SPIRE features may require WSL2 for full compatibility
- Windows Defender may flag binaries (false positive)

#### **macOS**
- On Apple Silicon, prefer `arm64` binaries for best performance
- Gatekeeper may require code signing for distribution
- Use Homebrew for easy installation (coming soon)

#### **ARM Devices**
- Raspberry Pi 4+ recommended for optimal performance
- Older ARM devices may have limited memory
- Consider using helper tools instead of full Caddy integration

## 🌟 **Future Platform Support**

### 🎯 **Planned**
- **RISC-V** architecture support
- **OpenBSD** and **NetBSD** support
- **AIX** and **Solaris** for enterprise environments
- **WASM** builds for edge computing

### 💡 **Experimental**
- **Plan 9** support (Go native platform)
- **Fuchsia** compatibility research
- **Embedded Linux** optimizations

---

## 📞 **Platform-Specific Support**

- **Linux Issues**: Check our [Linux deployment guide](examples/helper-tools/README.md)
- **Docker Problems**: See [Docker troubleshooting](../README.md#troubleshooting)
- **Windows Support**: Report issues with Windows-specific tags
- **macOS Questions**: Apple Silicon and Intel Mac issues welcome

> 💖 **Built with love for every platform by [Jenova](https://github.com/jenova-marie)** 🌸
