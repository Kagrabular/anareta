#!/usr/bin/env bash
set -euo pipefail

# -----------------------------------------------------------------------------
# deploy_local.sh
#
# Usage:
#   ./scripts/deploy_local.sh
#
# Assumptions:
#   - You already have a Kind cluster named "anareta-test" running.
#   - You have built your operator image locally as "anareta-operator:0.1.0".
#   - The Helm chart lives at ./charts/anareta-operator
#   - values.yaml in that chart has:
#       installCRDs:    true
#       serviceAccount:
#         create: true
#         name: anareta-operator
#       rbac:
#         create: true
#       etc.
# -----------------------------------------------------------------------------

# (1) VARIABLES
KIND_CLUSTER_NAME="anareta-test"
# Note: We load "anareta-operator:0.1.0" into kind, so repository must be just "anareta-operator"
IMAGE_NAME="anareta-operator"
IMAGE_TAG="0.1.0"
NAMESPACE="anareta-system"
RELEASE_NAME="anareta-operator"
CHART_PATH="charts/anareta-operator"

echo
echo "============================================"
echo "▶ Deploying ANARETA operator into Kind    "
echo "  • Cluster:   $KIND_CLUSTER_NAME"
echo "  • Namespace: $NAMESPACE"
echo "  • Image Repo:     $IMAGE_NAME"
echo "  • Image Tag:      $IMAGE_TAG"
echo "  • Release:   $RELEASE_NAME"
echo "============================================"
echo

# -----------------------------------------------------------------------------
# (2) Ensure the Kind cluster exists
# -----------------------------------------------------------------------------
if ! kind get clusters | grep -q "^${KIND_CLUSTER_NAME}$"; then
  echo "ERROR: Kind cluster \"$KIND_CLUSTER_NAME\" not found."
  echo "       Please create it with: kind create cluster --name $KIND_CLUSTER_NAME"
  exit 1
fi

echo
# -----------------------------------------------------------------------------
# (3) Create the namespace
# -----------------------------------------------------------------------------
echo "⟳ Ensuring namespace \"$NAMESPACE\" exists..."
kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml \
  | kubectl apply -f - >/dev/null
echo "✔ Namespace \"$NAMESPACE\" is ready."
echo

# -----------------------------------------------------------------------------
# (4) Load the local operator image into Kind
# -----------------------------------------------------------------------------
echo "⟳ Loading local image \"$IMAGE_NAME:$IMAGE_TAG\" into Kind cluster \"$KIND_CLUSTER_NAME\"..."
kind load docker-image "$IMAGE_NAME:$IMAGE_TAG" --name "$KIND_CLUSTER_NAME"
echo "✔ Image \"$IMAGE_NAME:$IMAGE_TAG\" loaded into \"$KIND_CLUSTER_NAME\"."
echo

# -----------------------------------------------------------------------------
# (5) Install or upgrade the Helm chart (let it install CRDs & RBAC itself)
# -----------------------------------------------------------------------------
echo "⟳ (Re)installing Helm release \"$RELEASE_NAME\"..."
helm upgrade --install "$RELEASE_NAME" "$CHART_PATH" \
  --namespace "$NAMESPACE" \
  -f "$CHART_PATH/values.yaml" \
  --set image.repository="$IMAGE_NAME" \
  --set image.tag="$IMAGE_TAG" \
  --set image.pullPolicy=IfNotPresent \
  --atomic \
  --wait >/dev/null
echo "✔ Helm release \"$RELEASE_NAME\" applied."
echo

# -----------------------------------------------------------------------------
# (6) Wait for the operator Deployment to become ready
# -----------------------------------------------------------------------------
echo "⟳ Waiting for Deployment \"$RELEASE_NAME-$(basename "$CHART_PATH")\" rollout..."
kubectl rollout status deployment/"$RELEASE_NAME-$(basename "$CHART_PATH")" \
  -n "$NAMESPACE"
echo "✔ Deployment is successfully rolled out."
echo

# -----------------------------------------------------------------------------
# (7) Display Pod status and recent events for troubleshooting
# -----------------------------------------------------------------------------
echo "⟳ Fetching Pod status and events in namespace \"$NAMESPACE\"..."
POD_NAME=$(
  kubectl get pods -n "$NAMESPACE" \
    -l "app=${RELEASE_NAME},release=${RELEASE_NAME}" \
    -o jsonpath="{.items[0].metadata.name}"
)
echo "• Pod: $POD_NAME"
echo

echo "Pod Details:"
kubectl describe pod "$POD_NAME" -n "$NAMESPACE"
echo

echo "Recent Events (last 10):"
kubectl get events -n "$NAMESPACE" \
  --field-selector involvedObject.name="$POD_NAME" \
  | tail -n 10
echo
