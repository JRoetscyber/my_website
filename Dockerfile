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

# Copy and compile high-performance C++ WebP converter
COPY webp_converter/ /src/webp_converter/
RUN if [ -f /src/webp_converter/converter.cpp ]; then \
        apk add --no-cache build-base libwebp-dev libgomp && \
        g++ /src/webp_converter/converter.cpp -o /app/converter -lwebp -lm -fopenmp -O3 || true; \
    fi

# Stage 2: Ultra-Lightweight Production Runtime
FROM alpine:3.20

# Install runtime dependencies including Google cwebp and OpenMP runtime
RUN apk add --no-cache ca-certificates tzdata curl libwebp libwebp-tools libgomp

# Security: Non-root user execution
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/ /app/

# Copy views and static assets
COPY go_app/views /app/views
COPY static /app/static

# Persistent data directory for SQLite WAL database & user uploads
RUN mkdir -p /app/data /app/static/uploads /app/static/uploads/webp \
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
