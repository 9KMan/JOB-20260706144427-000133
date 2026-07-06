#!/usr/bin/env bash
set -euo pipefail

# Bootstrap script for Grux PoC Cloud Run deployment
# Usage: GCP_PROJECT=my-project GCP_REGION=us-central1 ./deploy.sh

: "${GCP_PROJECT:?GCP_PROJECT env var required}"
: "${GCP_REGION:=us-central1}"
: "${SERVICE_NAME:=grux-poc-api}"
: "${REPO_NAME:=grux-poc}"

echo "→ Enabling GCP APIs..."
gcloud services enable run.googleapis.com \
  cloudbuild.googleapis.com \
  artifactregistry.googleapis.com \
  --project="${GCP_PROJECT}"

echo "→ Creating Artifact Registry repo (if not exists)..."
gcloud artifacts repositories create "${REPO_NAME}" \
  --repository-format=docker \
  --location="${GCP_REGION}" \
  --description="Grux PoC container images" \
  --project="${GCP_PROJECT}" 2>/dev/null || echo "  (repo already exists)"

echo "→ Submitting Cloud Build..."
gcloud builds submit \
  --config=deploy/cloudbuild.yaml \
  --project="${GCP_PROJECT}" \
  --substitutions="_REGION=${GCP_REGION},_REPO=${REPO_NAME},_SERVICE=${SERVICE_NAME}"

echo ""
echo "✓ Deployment complete!"
echo "→ Get service URL:"
echo "  gcloud run services describe ${SERVICE_NAME} --region=${GCP_REGION} --project=${GCP_PROJECT} --format='value(status.url)'"
