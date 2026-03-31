#!/usr/bin/env bash
#
# build-binaries.sh — Cross-compile DiffSurge CLI binaries using GoReleaser.
#
# Usage:
#   ./scripts/build-binaries.sh            # snapshot build (no publish)
#   ./scripts/build-binaries.sh --release  # tagged release (publishes to GitHub)
#
# Prerequisites:
#   - Go 1.24+
#   - GoReleaser v2+  (brew install goreleaser)
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
GO_MODULE_DIR="$PROJECT_ROOT/diffsurge-go"

# ── Ensure goreleaser is available ──────────────────────────────────────────
if ! command -v goreleaser &>/dev/null; then
  echo "❌  goreleaser not found. Install it first:"
  echo "    brew install goreleaser    # macOS"
  echo "    go install github.com/goreleaser/goreleaser/v2@latest"
  exit 1
fi

echo "📦  GoReleaser version: $(goreleaser --version | head -1)"

# ── Create .goreleaser.yml if it doesn't exist yet ──────────────────────────
GORELEASER_CFG="$GO_MODULE_DIR/.goreleaser.yml"
if [[ ! -f "$GORELEASER_CFG" ]]; then
  echo "⚙️   Generating default .goreleaser.yml …"
  cat > "$GORELEASER_CFG" <<'YAML'
# yaml-language-server: $schema=https://goreleaser.com/static/schema.json
version: 2

project_name: diffsurge

before:
  hooks:
    - go mod tidy

builds:
  - id: surge
    main: ./cmd/surge
    binary: surge
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.ShortCommit}}
      - -X main.date={{.Date}}

archives:
  - id: default
    format: tar.gz
    format_overrides:
      - goos: windows
        format: zip
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"

checksum:
  name_template: "checksums.txt"

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
      - "^ci:"
      - "^chore:"

release:
  github:
    owner: ankitbuildstuff
    name: diffsurge
  draft: false
  prerelease: auto
YAML
  echo "✅  Created $GORELEASER_CFG"
fi

# ── Run GoReleaser ──────────────────────────────────────────────────────────
cd "$GO_MODULE_DIR"

if [[ "${1:-}" == "--release" ]]; then
  echo "🚀  Running tagged release …"
  goreleaser release --clean
else
  echo "🔨  Running snapshot build (no publish) …"
  goreleaser build --snapshot --clean
  echo ""
  echo "✅  Binaries are in: $GO_MODULE_DIR/dist/"
  ls -lh dist/*/surge* 2>/dev/null || echo "(check dist/ for output)"
fi
