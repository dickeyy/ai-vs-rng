## syntax=docker/dockerfile:1

# ---------- Build stage ----------
FROM golang:1.24-alpine AS builder

# Enable Go modules and configure environment
ENV CGO_ENABLED=0 \
    GO111MODULE=on

WORKDIR /app

# Install build dependencies (git for fetching modules)
RUN apk add --no-cache git ca-certificates && update-ca-certificates

# Cache go.mod and go.sum first for better layer caching
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy the rest of the source
COPY . .

# Build the binary
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -o /bin/cis-320 ./

# ---------- Runtime stage ----------
FROM gcr.io/distroless/static:nonroot AS runtime

WORKDIR /

# Copy binary
COPY --from=builder /bin/cis-320 /bin/cis-320
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy runtime assets needed by the app
COPY symbols.txt /symbols.txt

# Expose nothing by default; app is CLI
USER nonroot:nonroot

ENTRYPOINT ["/bin/cis-320"]


