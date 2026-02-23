#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUTPUT="${REPO_ROOT}/bin/g2o"

echo "Building g2o..."
go build -o "$OUTPUT" "$REPO_ROOT"
echo "Built: $OUTPUT"
