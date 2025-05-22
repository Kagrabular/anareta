#!/usr/bin/env bash
set -euo pipefail

# Generate deepcopy methods and CRD manifests
echo "Generating deepcopy and CRD manifests..."
make generate
make manifests

echo "CRD YAMLs written to config/crd/bases/"

