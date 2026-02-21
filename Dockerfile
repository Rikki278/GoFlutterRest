# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install dependencies faster by caching the module layer separately
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build a statically-linked binary (no CGO needed for pgx driver)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server ./cmd/api/...

# ── Stage 2: Run ──────────────────────────────────────────────────────────────
FROM alpine:3.20

# ca-certificates is required for TLS connections (e.g. Render PostgreSQL + sslmode=require)
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/server .

# Render injects PORT at runtime — the app reads it via SERVER_PORT or PORT
EXPOSE 8080

CMD ["./server"]
