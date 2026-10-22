# ----------------------------------------------------
# ARMSS Photo Studio -- Go Backend Dockerfile
# Multi-stage build: static binary for 100% reliability
# ----------------------------------------------------
FROM golang:1.24-bookworm AS builder

WORKDIR /app

# Download modules first for better Docker layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Build a statically-linked Linux binary (CGO_ENABLED=0 avoids libc deps)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o armss-photo-studio ./cmd/server

# ----------------------------------------------------
# Minimal Production Runtime Image
# ----------------------------------------------------
FROM debian:bookworm-slim

WORKDIR /app

# Install wget and ca-certificates for healthcheck + outbound HTTPS calls
# (minio-go and AI provider clients need TLS)
RUN apt-get update -qq && \
    apt-get install -y --no-install-recommends wget ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Copy timezone data from builder (no apt-get tzdata needed)
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy compiled binary
COPY --from=builder /app/armss-photo-studio /app/armss-photo-studio

# Copy static assets required at runtime (fonts, brand assets)
COPY --from=builder /app/assets /app/assets

# Persistent local storage directory (used when STORAGE_DRIVER=local)
RUN mkdir -p /app/data/storage

ENV PORT=8080 \
    ENV=production \
    STORAGE_DRIVER=local \
    LOCAL_STORAGE_DIR=/app/data/storage

EXPOSE 8080

# Health check -- /healthz is a fast liveness probe (no DB/S3 deps)
# /readyz is the deeper dependency check -- used by the readiness gate
HEALTHCHECK --interval=10s --timeout=5s --start-period=25s --retries=12 \
    CMD wget -qO- http://127.0.0.1:${PORT:-8080}/healthz > /dev/null || exit 1

CMD ["/app/armss-photo-studio"]
