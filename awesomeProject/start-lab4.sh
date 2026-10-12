#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"
go mod tidy
go fmt ./...
go test ./...
go run ./cmd/awesomeProject
