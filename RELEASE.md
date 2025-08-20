# 🎉 Caddy SPIRE Client Release Guide

This document outlines the complete release process for the Caddy SPIRE Client project, including binary releases and Docker image publishing.

## 📋 Prerequisites

- [ ] Go 1.24.2+ installed
- [ ] Docker with buildx support
- [ ] GitHub CLI or access to GitHub releases
- [ ] GitHub Container Registry access
- [ ] All tests passing (`make test-all`)

## 🌸 Release Checklist

### 1. Pre-Release Preparation

```bash
# Ensure you're on the main branch and up-to-date
git checkout main
git pull origin main

# Run full test suite
make test-all

# Clean build environment
make clean-all
```

### 2. Version Management

```bash
# Update version in relevant files:
# - Makefile (VERSION_TAG variable)
# - Dockerfile.caddy (CustomVersion ldflags)
# - Any other version references

# Example: Update VERSION_TAG in Makefile
# VERSION_TAG ?= v1.0.5
```

### 3. Binary Release Build

```bash
# Build cross-platform release artifacts
make release

# Verify artifacts were created
ls -la dist/archives/

# Expected files:
# - caddy-spire-svid-linux-amd64.tar.gz
# - caddy-spire-svid-linux-arm64.tar.gz
# - caddy-spire-svid-darwin-amd64.tar.gz
# - caddy-spire-svid-darwin-arm64.tar.gz
# - caddy-spire-svid-windows-amd64.zip
# - caddy-spire-svid-windows-arm64.zip
```

### 4. Docker Image Release

#### Setup Docker Environment (One-time)

```bash
# Setup Docker buildx for multi-architecture builds
make docker-setup

# Login to GitHub Container Registry
make docker-login
# Use your GitHub username and personal access token
```

#### Build and Push Docker Images

```bash
# Option A: Build and push multi-architecture image
make docker-push

# Option B: Build architectures separately (for debugging)
make docker-build-amd64
make docker-build-arm64

# Option C: Build locally without pushing (for testing)
make docker-build
```

#### Verify Docker Images

```bash
# Check that images were pushed successfully
docker buildx imagetools inspect ghcr.io/jenova-marie/caddy-spire-svid-caddy:latest
docker buildx imagetools inspect ghcr.io/jenova-marie/caddy-spire-svid-caddy:v1.0.4

# Test pulling images
docker pull ghcr.io/jenova-marie/caddy-spire-svid-caddy:latest
```

### 5. Create GitHub Release

#### Tag the Release

```bash
# Create and push git tag
git tag v1.0.4
git push origin v1.0.4
```

#### Create GitHub Release (Manual)

1. Go to GitHub repository → Releases → "Create a new release"
2. Choose the tag: `v1.0.4`
3. Release title: `🌸 Caddy SPIRE Client v1.0.4`
4. Upload binary artifacts from `dist/archives/`
5. Write release notes with changelog

#### Example Release Notes Template

```markdown
## 🌸 Caddy SPIRE Client v1.0.4

### ✨ New Features
- [List new features]

### 🐛 Bug Fixes
- [List bug fixes]

### 🔧 Improvements
- [List improvements]

### 📦 Installation

#### Binaries
Download the appropriate binary for your platform:

**Linux AMD64:**
```bash
curl -L https://github.com/jenova-marie/caddy-spire-svid/releases/download/v1.0.4/caddy-spire-svid-linux-amd64.tar.gz | tar xz
sudo mv caddy-spire-svid /usr/local/bin/
```

**Linux ARM64:**
```bash
curl -L https://github.com/jenova-marie/caddy-spire-svid/releases/download/v1.0.4/caddy-spire-svid-linux-arm64.tar.gz | tar xz
sudo mv caddy-spire-svid /usr/local/bin/
```

#### Docker Image
```bash
# Pull the Caddy + SPIRE image
docker pull ghcr.io/jenova-marie/caddy-spire-svid-caddy:v1.0.4

# Or use latest
docker pull ghcr.io/jenova-marie/caddy-spire-svid-caddy:latest
```

### 🔧 Usage
```bash
caddy-spire-svid -help
```

💖 Made with love by Jenova
```

## 🚀 Quick Release Commands

For experienced maintainers, here's the condensed release flow:

```bash
# Pre-flight checks
make test-all && make clean-all

# Build binaries
make release

# Setup Docker (if not done)
make docker-setup && make docker-login

# Build and push Docker images
make docker-push

# Tag and create GitHub release
git tag v1.0.X
git push origin v1.0.X
# Then create GitHub release with dist/archives/* files
```

## 🛠️ Troubleshooting

### Docker Build Issues

```bash
# Check buildx status
docker buildx ls

# Reset buildx if needed
docker buildx rm multiarch-builder
make docker-setup

# Build single architecture for debugging
make docker-build-arm64  # or docker-build-amd64 or docker-build-armv7
```

### GitHub Actions CI/CD Issues

- **GHCR push fails**: `GITHUB_TOKEN` cannot push to GHCR - use `GHCR_TOKEN` secret instead
- **Add GHCR_TOKEN secret**: Repo Settings → Secrets → Actions → New secret `GHCR_TOKEN` with your PAT
- **Permission denied**: Check GitHub token has `write:packages` and `read:packages` scopes

### Binary Build Issues

```bash
# Clean and rebuild
make clean-all
make release

# Check Go environment
go version
go env GOOS GOARCH
```

### Registry Authentication Issues

```bash
# Re-login to GHCR
docker logout ghcr.io
make docker-login

# Verify access
docker buildx imagetools inspect ghcr.io/jenova-marie/caddy-spire-svid-caddy:latest
```

## 📊 Post-Release Verification

```bash
# Test binary download and execution
curl -L https://github.com/jenova-marie/caddy-spire-svid/releases/download/v1.0.X/caddy-spire-svid-linux-amd64.tar.gz | tar xz
./caddy-spire-svid -version

# Test Docker image
docker run --rm ghcr.io/jenova-marie/caddy-spire-svid-caddy:v1.0.X version

# Verify multi-arch support
docker buildx imagetools inspect ghcr.io/jenova-marie/caddy-spire-svid-caddy:v1.0.X
```

## 🌸 Release Notes

Keep track of what changed in each release:

- **v1.0.4** - [Date] - Description
- **v1.0.3** - [Date] - Description
- **v1.0.2** - [Date] - Description

---

💖 **Happy releasing!** This process ensures consistent, high-quality releases across all platforms and architectures.
