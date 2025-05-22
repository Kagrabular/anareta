#!/usr/bin/env bash
set -euo pipefail

echo "Applying CRDs..."
kubectl apply -f config/crd/bases

NAMESPACE="anareta-system"
echo "Ensuring namespace $NAMESPACE..."
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

echo "Applying RBAC resources..."
kubectl apply -f config/rbac

echo "Deploying controller manager..."
kubectl apply -f config/manager/manager.yaml

echo "Installing Helm chart..."
helm upgrade --install anareta-operator charts/anareta-operator -n $NAMESPACE --set installCRDs=false

echo "Resources in $NAMESPACE:"
kubectl get all -n $NAMESPACE
