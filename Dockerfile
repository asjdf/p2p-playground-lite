# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binaries
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/daemon ./cmd/daemon
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/controller ./cmd/controller

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create app user
RUN addgroup -S p2pgroup && adduser -S p2puser -G p2pgroup

# Copy binaries from builder
COPY --from=builder /bin/daemon /usr/local/bin/daemon
COPY --from=builder /bin/controller /usr/local/bin/controller

# Copy example configs
COPY configs/daemon.example.yaml /etc/p2p-playground/daemon.yaml
COPY configs/controller.example.yaml /etc/p2p-playground/controller.yaml

# Create data directories
RUN mkdir -p /data/packages /data/apps /data/keys /data/keys/trusted

# Copy trusted public keys for signature verification
COPY --chown=p2puser:p2pgroup configs/keys/trusted/*.pub /data/keys/trusted/

# Ensure directories exist and set ownership
RUN mkdir -p /data/keys/trusted && chown -R p2puser:p2pgroup /data

# Switch to non-root user
USER p2puser

# Set working directory
WORKDIR /data

# Expose default ports
EXPOSE 9000

# Default command
CMD ["daemon", "start", "-c", "/etc/p2p-playground/daemon.yaml"]
