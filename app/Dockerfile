# ── Build stage ──────────────────────────────────────────────────────────────

FROM golang:1.23.3-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download


COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /monitoring-service ./cmd/main.go