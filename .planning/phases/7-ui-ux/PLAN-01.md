---
phase: 7
plan: 01
type: IMPLEMENTATION
wave: 4
depends_on:
  - phase: 5
  - phase: 6
plans: ["01"]
files_modified:
  - deploy/cloudbuild.yaml
  - deploy/deploy.sh
  - deploy/env.example
  - deploy/README.md
autonomous: true
requirements:
  - Cloud Build config: build → push to Artifact Registry → deploy to Cloud Run
  - Bootstrap script: enable APIs, create Artifact Registry, run build
  - Env var template
  - Deploy instructions
---

# Phase 7 — UI/UX (GCP Cloud Run deploy infrastructure)

## What this phase delivers

The GCP deployment pipeline: Cloud Build config + bootstrap script +
env var template + deploy README. Demonstrates Cloud Run deployment path.

## Files to Create

| File | Purpose |
|---|---|
| `deploy/cloudbuild.yaml` | Cloud Build steps: docker build → push to Artifact Registry → deploy to Cloud Run |
| `deploy/deploy.sh` | Bootstrap: enable GCP APIs + create Artifact Registry + run Cloud Build |
| `deploy/env.example` | Env var template (GCP_PROJECT, GCP_REGION, SERVICE_NAME) |
| `deploy/README.md` | Deploy instructions: prerequisites, setup, deploy, troubleshoot |

## Implementation requirements

### `deploy/cloudbuild.yaml`
```yaml
# Cloud Build pipeline for Grux PoC API
# Steps:
#   1. Build Docker image
#   2. Push to Artifact Registry
#   3. Deploy to Cloud Run
#
# Required substitutions (set via --substitutions or trigger):
#   _REGION: GCP region (default: us-central1)
#   _REPO: Artifact Registry repo name (default: grux-poc)
#   _SERVICE: Cloud Run service name (default: grux-poc-api)

steps:
  - id: build
    name: gcr.io/cloud-builders/docker
    args:
      - build
      - -t
      - ${_REGION}-docker.pkg.dev/${PROJECT_ID}/${_REPO}/${_SERVICE}:${SHORT_SHA}
      - -t
      - ${_REGION}-docker.pkg.dev/${PROJECT_ID}/${_REPO}/${_SERVICE}:latest
      - .

  - id: push
    name: gcr.io/cloud-builders/docker
    args:
      - push
      - --all-tags
      - ${_REGION}-docker.pkg.dev/${PROJECT_ID}/${_REPO}/${_SERVICE}

  - id: deploy
    name: gcr.io/google.com/cloudsdktool/cloud-sdk
    entrypoint: gcloud
    args:
      - run
      - deploy
      - ${_SERVICE}
      - --image=${_REGION}-docker.pkg.dev/${PROJECT_ID}/${_REPO}/${_SERVICE}:${SHORT_SHA}
      - --region=${_REGION}
      - --platform=managed
      - --allow-unauthenticated
      - --port=8080
      - --memory=512Mi
      - --cpu=1
      - --max-instances=10

images:
  - ${_REGION}-docker.pkg.dev/${PROJECT_ID}/${_REPO}/${_SERVICE}:${SHORT_SHA}
  - ${_REGION}-docker.pkg.dev/${PROJECT_ID}/${_REPO}/${_SERVICE}:latest

substitutions:
  _REGION: us-central1
  _REPO: grux-poc
  _SERVICE: grux-poc-api

options:
  logging: CLOUD_LOGGING_ONLY
```

### `deploy/deploy.sh`
```bash
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
```

### `deploy/env.example`
```bash
# GCP project ID (required)
GCP_PROJECT=my-gcp-project

# GCP region (default: us-central1)
GCP_REGION=us-central1

# Cloud Run service name (default: grux-poc-api)
SERVICE_NAME=grux-poc-api

# Artifact Registry repo name (default: grux-poc)
REPO_NAME=grux-poc
```

### `deploy/README.md`
- Prerequisites: gcloud CLI, GCP project, billing enabled
- One-time setup: `./deploy.sh` (with env vars set)
- Verify: `curl $(gcloud run services describe grux-poc-api --region=us-central1 --format='value(status.url)')/health`
- Troubleshoot: Cloud Build logs, Cloud Run logs

## Verification

```bash
bash -n deploy/deploy.sh  # expect: parses without errors
python3 -c "import yaml; yaml.safe_load(open('deploy/cloudbuild.yaml'))"  # expect: valid YAML
test -f deploy/env.example && echo "EXISTS"
test -f deploy/README.md && wc -c deploy/README.md  # expect: > 1000 chars
```

## Acceptance criteria

- [ ] `deploy/cloudbuild.yaml` is valid YAML with 3 build steps
- [ ] `deploy/deploy.sh` is executable (chmod +x)
- [ ] `deploy/env.example` documents all required env vars
- [ ] `deploy/README.md` has prerequisites + deploy + verify + troubleshoot sections