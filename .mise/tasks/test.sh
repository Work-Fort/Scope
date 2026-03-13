#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
#MISE description="Run tests with race detection and coverage"
set -euo pipefail

mkdir -p build
go test -v -race -coverprofile=build/coverage.out ./...
