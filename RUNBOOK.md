# RUNBOOK — Grux PoC Operations

## Local Development

### Run the Go backend
```bash
make run                          # start on :8080
```

### Run tests
```bash
go test ./... -v                  # all tests with verbose output
```

### Build the Docker image
```bash
make docker-build                 # builds grux-poc-api:latest
make docker-run                   # runs on localhost:8080
```

### Run the Flutter app
```bash
cd mobile
flutter pub get                  # install dependencies
flutter run -d chrome            # run in Chrome (web target)
flutter run -d <device-id>       # run on physical device
flutter test                     # run widget tests
```

## GCP Deployment

### Prerequisites
- `gcloud` CLI installed and authenticated (`gcloud auth login`)
- GCP project with billing enabled
- Required IAM roles: Cloud Run Admin, Cloud Build Editor, Artifact Registry Admin

### One-time setup
```bash
export GCP_PROJECT=my-gcp-project
export GCP_REGION=us-central1
./deploy/deploy.sh                # enables APIs, creates repo, runs Cloud Build
```

### Verify deployment
```bash
SERVICE_URL=$(gcloud run services describe grux-poc-api \
  --region=$GCP_REGION \
  --project=$GCP_PROJECT \
  --format='value(status.url)')

curl $SERVICE_URL/health
# {"status":"ok","service":"grux-poc","version":"0.1.0"}
```

### Redeploy (after code changes)
```bash
gcloud builds submit \
  --config=deploy/cloudbuild.yaml \
  --project=$GCP_PROJECT \
  --substitutions="_REGION=$GCP_REGION,_REPO=grux-poc,_SERVICE=grux-poc-api"
```

### View logs
```bash
gcloud run services logs read grux-poc-api \
  --region=$GCP_REGION \
  --project=$GCP_PROJECT \
  --limit=50
```

## Troubleshooting

### Build fails: `go: command not found`
- Install Go 1.22+: https://go.dev/doc/install
- Verify: `go version`

### Tests fail: `cannot find package`
- Run `go mod tidy` to resolve dependencies
- Run `go mod verify` to validate

### Cloud Build fails: `API not enabled`
- Run `./deploy/deploy.sh` which enables all required APIs
- Or manually: `gcloud services enable run.googleapis.com cloudbuild.googleapis.com artifactregistry.googleapis.com`

### Cloud Run returns 500
- Check logs: `gcloud run services logs read grux-poc-api --region=$GCP_REGION`
- Verify env vars: `gcloud run services describe grux-poc-api --region=$GCP_REGION`
- Common cause: PORT not set to 8080 (Cloud Run sets $PORT automatically)

### Flutter app can't reach backend
- Check `API_BASE_URL` in `mobile/lib/services/api_client.dart`
- For Android emulator, use `10.0.2.2` instead of `localhost`
- For physical device on same WiFi, use your machine's LAN IP
- For production, use the Cloud Run URL

### CORS errors in browser
- The PoC sets CORS to `*` for development
- For production, restrict to specific origins in `internal/api/middleware.go`

## Performance Baselines (PoC)

| Metric | Target | How to measure |
|---|---|---|
| Cold start (Cloud Run) | < 2s | `time curl $SERVICE_URL/health` |
| Request latency p95 | < 50ms | Cloud Run metrics |
| Memory usage | < 256 MB | Cloud Run metrics |
| Test coverage | ≥ 60% | `go test -cover ./...` |

## Production Migration Checklist

When promoting from PoC to production:

- [ ] Replace in-memory store with Postgres (Cloud SQL)
- [ ] Add Google Workspace SSO authentication
- [ ] Add OpenTelemetry tracing + Cloud Trace integration
- [ ] Add Cloud Armor for DDoS protection
- [ ] Add Cloud Monitoring alerts (error rate, latency p95)
- [ ] Add CI/CD (GitHub Actions → Cloud Build → Cloud Run)
- [ ] Add staging environment (separate Cloud Run service)
- [ ] Add custom domain + SSL cert (Cloud Run domain mappings)
- [ ] Add backup strategy (Cloud SQL automated backups)
- [ ] Add runbook for common incidents (DB connection loss, API rate limits)

## References

- [Go documentation](https://go.dev/doc/)
- [Gin framework](https://gin-gonic.com/docs/)
- [Flutter documentation](https://docs.flutter.dev/)
- [GCP Cloud Run docs](https://cloud.google.com/run/docs)
- [Cloud Build configuration reference](https://cloud.google.com/build/docs/configuring-builds/create-basic-configuration)