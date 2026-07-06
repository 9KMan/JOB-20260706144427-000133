# Grux PoC — GCP Cloud Run Deployment

This directory contains the Cloud Run deployment pipeline for the Grux PoC
Go API. It uses Cloud Build to build the container, push it to Artifact
Registry, and deploy to Cloud Run in one shot.

## Prerequisites

- **gcloud CLI** installed and authenticated (`gcloud auth login`).
- A **GCP project** with billing enabled.
- The `cloudbuild.googleapis.com` IAM permission for the project (granted by
  default to project owners; service accounts can be granted the
  `Cloud Build Service Account` role).
- **Docker** is NOT required locally — Cloud Build runs the build.
- The Go API Dockerfile lives at the repo root as `Dockerfile`.

## Quick start

1. Copy the env template and fill in your project ID:
   ```bash
   cp deploy/env.example deploy/.env
   # edit deploy/.env: set GCP_PROJECT=your-actual-project-id
   set -a; source deploy/.env; set +a
   ```

2. Run the bootstrap script (one-time per project):
   ```bash
   ./deploy/deploy.sh
   ```
   This will:
   - Enable Cloud Run, Cloud Build, and Artifact Registry APIs.
   - Create the Artifact Registry repo (idempotent).
   - Submit the Cloud Build, which builds, pushes, and deploys.

3. (Optional) override defaults:
   ```bash
   GCP_PROJECT=my-prod-project \
   GCP_REGION=us-east1 \
   SERVICE_NAME=grux-poc-api-v2 \
   REPO_NAME=grux-images \
   ./deploy/deploy.sh
   ```

## Verify the deployment

After `deploy.sh` finishes, fetch the service URL and curl the health
endpoint:

```bash
SERVICE_URL=$(gcloud run services describe grux-poc-api \
  --region=us-central1 --project="${GCP_PROJECT}" \
  --format='value(status.url)')

curl "${SERVICE_URL}/health"
# → {"status":"ok"}

curl "${SERVICE_URL}/items"
# → []
```

If the API requires auth, drop `--allow-unauthenticated` from the Cloud Build
deploy step and use `gcloud run services proxy` or an identity token.

## How it works

The pipeline (`deploy/cloudbuild.yaml`) runs three steps in sequence:

1. **build** — `docker build` tags the image with both `${SHORT_SHA}` and
   `latest` so Cloud Run can deploy by immutable SHA while humans can still
   reference `:latest` for rollbacks.
2. **push** — pushes both tags to Artifact Registry at
   `${_REGION}-docker.pkg.dev/${PROJECT_ID}/${_REPO}/${_SERVICE}`.
3. **deploy** — `gcloud run deploy` with `--allow-unauthenticated`, port
   8080, 512Mi memory, 1 CPU, max 10 instances. Tune these via the
   `cloudbuild.yaml` deploy step.

### Substitutions

| Var | Default | Purpose |
|---|---|---|
| `_REGION` | `us-central1` | GCP region for Artifact Registry + Cloud Run |
| `_REPO` | `grux-poc` | Artifact Registry repo name |
| `_SERVICE` | `grux-poc-api` | Cloud Run service name |

Override on the gcloud command with
`--substitutions=_REGION=...,_REPO=...,_SERVICE=...`.

## Manual deploy (without the bootstrap)

```bash
gcloud builds submit \
  --config=deploy/cloudbuild.yaml \
  --project="${GCP_PROJECT}" \
  --substitutions="_REGION=us-central1,_REPO=grux-poc,_SERVICE=grux-poc-api"
```

## Continuous deployment

To deploy on every push to `main`:

```bash
gcloud builds triggers create github \
  --repo-name=grux-poc \
  --repo-owner=YOUR_ORG \
  --branch-pattern="^main$" \
  --build-config=deploy/cloudbuild.yaml \
  --project="${GCP_PROJECT}" \
  --region="${GCP_REGION}"
```

## Troubleshooting

### Build fails: "API [cloudbuild.googleapis.com] not enabled"
Run `./deploy/deploy.sh` once — it enables the required APIs.

### Build fails: "Permission denied" pushing to Artifact Registry
The Cloud Build default service account
(`${PROJECT_NUMBER}@cloudbuild.gserviceaccount.com`) needs the
**Artifact Registry Writer** role:
```bash
gcloud projects add-iam-policy-binding "${GCP_PROJECT}" \
  --member="serviceAccount:${PROJECT_NUMBER}@cloudbuild.gserviceaccount.com" \
  --role="roles/artifactregistry.writer"
```

### Cloud Run deploy fails: "Container failed to start"
Check Cloud Run logs:
```bash
gcloud run services logs read grux-poc-api \
  --region=us-central1 --project="${GCP_PROJECT}" --limit=50
```
Common cause: the API listens on a port other than 8080. Update
`--port=8080` in `cloudbuild.yaml` to match your API's `PORT` env var.

### Cloud Run deploy fails: "Revision ... is not ready"
Inspect revisions:
```bash
gcloud run revisions list \
  --service=grux-poc-api --region=us-central1 --project="${GCP_PROJECT}"
```
Then read logs of the failing revision:
```bash
gcloud run revisions logs read REVISION_NAME --project="${GCP_PROJECT}"
```

### Build succeeds but image is stale
The `:latest` tag is overwritten every build, but Cloud Run deploys by
SHA. To force a redeploy of `:latest`:
```bash
gcloud run services update grux-poc-api \
  --region=us-central1 --project="${GCP_PROJECT}" \
  --image="${_REGION}-docker.pkg.dev/${PROJECT_ID}/grux-poc/grux-poc-api:latest"
```

### Get the deployment SHA
```bash
gcloud builds list --project="${GCP_PROJECT}" --limit=1 --format='value(images)'
```

## Cleanup

```bash
gcloud run services delete grux-poc-api \
  --region=us-central1 --project="${GCP_PROJECT}"

gcloud artifacts repositories delete grux-poc \
  --location=us-central1 --project="${GCP_PROJECT}"
```
