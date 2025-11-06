# ---------- .dockerignore (should be present in project root) ----------
# .dockerignore -->
# vendor/
# bin/
# *.log
# tmp/
# .git
# node_modules
# Dockerfile
# docker-compose.yml
# .env
#
# ---------- Dockerfile ----------

# ==== BUILD STAGE ====
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install git for go mod downloads if needed
RUN apk add --no-cache git

# Cache deps before copying rest of source code
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build statically-linked binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o api-server ./cmd/api

# ==== RUNTIME STAGE ====
FROM alpine:3.19

# Create app user (non-root)
RUN adduser -D -g '' appuser

WORKDIR /app

# Copy built binary from builder
COPY --from=builder /app/api-server ./api-server

# Copy config files if needed (e.g. migrations, configs)
# COPY --from=builder /app/config.yaml ./
# ... etc

RUN chmod +x ./api-server
USER appuser

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=2s --start-period=5s --retries=3 \
  CMD wget --spider --no-verbose http://localhost:8080/health || exit 1

CMD ["./api-server"]
