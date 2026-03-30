# ============================================================================
# Diffsurge Services — Multi-stage Docker Build
#
# Build a specific service:
#   docker build --build-arg SERVICE=api      -t diffsurge-api .
#   docker build --build-arg SERVICE=proxy    -t diffsurge-proxy .
#   docker build --build-arg SERVICE=replayer -t diffsurge-replayer .
#   docker build --build-arg SERVICE=cli      -t diffsurge-cli .
#
# All services share the same builder base and the same distroless runtime
# base, so only one image definition is needed regardless of the target.
# ============================================================================

ARG SERVICE=api

# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

ARG SERVICE

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/       ./cmd/
COPY internal/  ./internal/
COPY pkg/       ./pkg/
COPY configs/   ./configs/

# Produce a fully static binary named "diffsurge"; each service image only
# ever contains the binary for its own component.
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X github.com/diffsurge-org/diffsurge/internal/config.Version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)" \
    -o /build/diffsurge \
    ./cmd/${SERVICE}/main.go

# ── Stage 2: Runtime (distroless — no shell, no package manager) ─────────────
# gcr.io/distroless/static-debian12 bundles only CA certificates and tzdata,
# making it suitable for fully-static Go binaries without alpine overhead.
# The :nonroot tag ensures the process runs as uid 65532 out of the box.
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder --chown=nonroot:nonroot /build/diffsurge /app/diffsurge

EXPOSE 8080

ENTRYPOINT ["/app/diffsurge"]
