#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
#MISE description="Build the workfort binary"
#MISE sources=["**/*.go", "go.mod", "go.sum"]
#MISE outputs=["build/workfort"]
set -euo pipefail

TAGS=""

# Build shell SPA if web/shell exists and has node_modules
if [ -d "web/shell/node_modules" ]; then
  mise run build:web
  TAGS="-tags spa"
fi

mkdir -p build
VERSION=$(git rev-parse --short HEAD 2>/dev/null || echo "dev")
go build $TAGS -ldflags "-X github.com/Work-Fort/Scope/cmd.Version=${VERSION}" -o build/workfort .
echo "✓ Build complete"
