# ── Build stage ──────────────────────────────────────────────────────────────

FROM golang:1.23.3-alpine AS builder

WORKDIR /app

# Copy dependency files first (better layer caching)
COPY go.mod go.sum ./
RUN go mod download


COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /monitoring-service ./cmd/main.go

# ── Run stage ──────────────────────────────────────────────────────────────
FROM alpine:latest

WORKDIR /app

# Add CA certs for any HTTPS calls
RUN apk --no-cache add ca-certificates

# Copy only the binary from the build stage
COPY --from=builder /monitoring-service .

EXPOSE 8080

CMD ["./monitoring-service"]