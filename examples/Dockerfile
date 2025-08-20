# 🌸 Simplified Caddy SPIRE Docker Image
# Use our custom Caddy build with SPIRE module included

# Use official lightweight Caddy image with SPIRE support
FROM ghcr.io/jenova-marie/caddy-spire-svid:latest

# Install bind-tools for DNS debugging (nslookup, dig, etc.)
RUN apk add --no-cache bind-tools curl

# Create necessary directories
RUN mkdir -p /srv /etc/caddy/sites /etc/caddy/headers /etc/caddy/shared /var/lib/caddy /var/log/caddy /data

# Set working directory
WORKDIR /srv

# Copy static HTML files
COPY ./html/ /srv/

# Copy Caddyfile and sites configuration
COPY ./Caddyfile /etc/caddy/Caddyfile
COPY ./sites/ /etc/caddy/sites/
COPY ./headers/ /etc/caddy/headers/
COPY ./shared/ /etc/caddy/shared/

# Expose HTTP and HTTPS
EXPOSE 80 443

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:80/health || exit 1

# Run Caddy with explicit config and adapter
CMD ["caddy", "run", "--config", "/etc/caddy/Caddyfile", "--adapter", "caddyfile"]
