#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT"
mkdir -p dist
VERSION="${VERSION:-dev}"
LDX="-X main.version=${VERSION}"

build() {
  local goos="$1" goarch="$2" suffix="${3:-}"
  local out="dist/yandex-speed-cli-${goos}-${goarch}${suffix}"
  echo "==> $out"
  GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=0 go build -trimpath -ldflags="-s -w $LDX" -o "$out" .
}

# Linux / Windows: x64 (amd64) + x86 (386) + arm64
build linux amd64
build linux 386
build linux arm64
build windows amd64 .exe
build windows 386 .exe
# macOS: x64 Intel + arm64 Apple Silicon (darwin/386 в Go не поддерживается)
build darwin amd64
build darwin arm64

echo "Готово: $ROOT/dist/"
