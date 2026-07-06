# syntax=docker/dockerfile:1.7
#
# Multi-stage Dockerfile for grux-poc-api.
# Stage 1: golang:1.22-alpine builds a static binary.
# Stage 2: gcr.io/distroless/static:nonroot runs it unprivileged.

# ---- Build stage ----
FROM golang:1.22-alpine AS builder
WORKDIR /app

# Cache go.mod/go.sum first to maximize layer reuse across builds.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO disabled, fully static binary — works under distroless:nonroot.
# -ldflags strip symbol/debug info to shrink the image.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /api ./cmd/api

# ---- Runtime stage ----
FROM gcr.io/distroless/static:nonroot

# Copy the static binary from the builder.
COPY --from=builder /api /api

# Distroless image declares a nonroot user/group (uid/gid 65532).
USER nonroot:nonroot

# Cloud Run injects PORT; default to 8080 inside the binary too.
EXPOSE 8080

ENV PORT=8080

ENTRYPOINT ["/api"]
