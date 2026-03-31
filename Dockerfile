# ============================================================================
# DiffSurge CLI — Multi-stage Docker Image
#
# Build:
#   docker build -t diffsurge/cli .
#
# Usage:
#   docker run -v $(pwd):/work diffsurge/cli surge --help
#   docker run -v $(pwd):/work diffsurge/cli surge capture --port 8080
#   docker run -v $(pwd):/work diffsurge/cli surge replay --against v2
#
# The final image is FROM scratch (~0 MB overhead). Only the statically-linked
# Go binary, CA certs, and timezone data are included.
# ============================================================================

# ── Stage 1: Build ──────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

COPY diffsurge-go/go.mod diffsurge-go/go.sum ./
RUN go mod download

COPY diffsurge-go/cmd/       ./cmd/
COPY diffsurge-go/internal/  ./internal/
COPY diffsurge-go/pkg/       ./pkg/
COPY diffsurge-go/configs/   ./configs/

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /build/surge \
    ./cmd/cli/main.go

# ── Stage 2: Scratch runtime ───────────────────────────────────────────────
FROM scratch

# CA certificates for HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Binary
COPY --from=builder /build/surge /usr/local/bin/surge

WORKDIR /work

ENTRYPOINT ["surge"]
