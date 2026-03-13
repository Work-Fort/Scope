#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
#MISE description="Build the workfort binary"
#MISE sources=["**/*.go", "go.mod", "go.sum"]
#MISE outputs=["build/workfort"]
set -euo pipefail

# Build shell SPA if web/shell exists and has node_modules
if [ -d "web/shell/node_modules" ]; then
  echo "Building shell SPA..."
  (cd web/shell && pnpm build)
  rm -rf cmd/web/placeholder
  cp -r web/shell/dist cmd/web/placeholder
fi

mkdir -p build
VERSION=$(git rev-parse --short HEAD 2>/dev/null || echo "dev")
go build -ldflags "-X github.com/Work-Fort/Scope/cmd.Version=${VERSION}" -o build/workfort .
echo "✓ Build complete"
