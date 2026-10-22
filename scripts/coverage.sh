#!/bin/sh
# coverage.sh runs the whole test suite with whole-program coverage and hands the profile to
# scripts/coverage.ant, which prints the figures and fails below the thresholds.
# Usage: scripts/coverage.sh [overall-min] [package-min]
set -eu
profile=${COVERPROFILE:-$(mktemp)}
bin=$(mktemp -d)/ant
go test -coverpkg=./... -coverprofile="$profile" ./... > /dev/null
go build -o "$bin" ./cmd/ant
exec "$bin" run scripts/coverage.ant "$profile" "${1:-90}" "${2:-80}"
