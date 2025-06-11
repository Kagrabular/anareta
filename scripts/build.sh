#!/usr/bin/env bash
set -euo pipefail

echo "Building operator binary..."
go build -o bin/manager main.go

#  Build webhook server if present gonna need this for PR hooks TO-DO
if [ -f webhook/main.go ]; then
  echo "Building webhook server..."
  go build -o bin/webhook-server webhook/main.go
fi

echo "Packaging Helm chart..."
helm package charts/anareta-operator -d bin/

echo "Artifacts in bin/:"
ls -1 bin/

