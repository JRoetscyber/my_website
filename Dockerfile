# Stage 1: High-Performance Go Compiler
FROM golang:alpine AS builder

WORKDIR /src

ENV GOTOOLCHAIN=auto \
    CGO_ENABLED=0

RUN apk add --no-cache git ca-certificates tzdata

# Cache Go modules layer
COPY go_app/go.mod go_app/go.sum ./
RUN go mod download

# Copy Go source code
COPY go_app/ ./

# Build minimal, stripped, static Go binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/server ./cmd/server

# Stage 2: Ultra-Lightweight Production Runtime
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata curl

# Security: Non-root user execution
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server /app/server

# Copy views and static assets
COPY go_app/views /app/views
COPY static /app/static

# Persistent data directory for SQLite WAL database & user uploads
RUN mkdir -p /app/data /app/static/uploads \
    && chown -R appuser:appgroup /app

USER appuser

EXPOSE 5000

ENV PORT=5000 \
    ENV=production \
    DATABASE_URL=/app/data/jo4dev.db \
    TEMPLATES_DIR=/app/views \
    STATIC_DIR=/app/static

HEALTHCHECK --interval=15s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:5000/health || exit 1

ENTRYPOINT ["/app/server"]
