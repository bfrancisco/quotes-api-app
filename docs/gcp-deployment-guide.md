# Deploy to GCP: a beginner guide

This guide deploys the REST and GraphQL APIs as **two Cloud Run services**. Their Docker images live in **Artifact Registry** and both services use **Firestore** for durable, shared quotes.

> [!IMPORTANT]
> Before deployment, run the Firestore emulator integration tests described in **Step 3**. Cloud Run must use `STORAGE_MODE=firestore`; the default `memory` mode is for local development only and loses data on restart.

## 1. What you need

- A Google Cloud project with billing enabled.
- Permission to view and change the team's infrastructure repository.
- `gcloud`, Docker, Git, Go, and Terraform installed locally.
- A Google account to use with IAP.
- A domain name and DNS access if browser access through IAP is required.

Choose one region, for example `us-central1`, and use it consistently for Artifact Registry and Cloud Run.

```bash
export PROJECT_ID="YOUR_GCP_PROJECT_ID"
export REGION="us-central1"
export REPOSITORY="quotes-api"
export REST_SERVICE="quotes-rest-api"
export GRAPHQL_SERVICE="quotes-graphql-api"

gcloud auth login
gcloud config set project "$PROJECT_ID"
gcloud auth application-default login
```

Check that the selected project is correct before continuing:

```bash
gcloud config get-value project
```

## 2. Learn the Terraform repository first

The infrastructure repository is the source of truth for GCP resources. Do **not** create duplicate load balancers, DNS zones, or IAM resources from this application repository.

Before adding resources, find out:

1. Where Terraform state is stored and how it is locked.
2. How environments are organized, such as development and production.
3. Which Terraform module manages the GCP project, DNS, and HTTPS load balancer.
4. How pull requests run `terraform plan` and how protected-main applies changes.
5. The team's naming, labels, IAM, and secret-management conventions.

Create a Terraform pull request that follows those conventions. It should provision, or connect this app to, the following resources:

| Resource | Why it is needed |
| --- | --- |
| Artifact Registry Docker repository | Stores REST and GraphQL container images. |
| Firestore database and indexes | Stores quotes durably for both APIs. |
| Cloud Run runtime service account | Lets the APIs access Firestore without key files. |
| Two Cloud Run services | Runs REST and GraphQL independently. |
| Workload Identity Federation | Lets GitHub Actions authenticate without a service-account key. |
| Serverless NEGs, routes, and IAP | Routes browser traffic and protects it with Google sign-in. |

### Terraform ownership boundary

- **Terraform manages:** Cloud Run settings, service accounts, IAM, Firestore, Artifact Registry, networking, load balancer, and IAP.
- **Application CI manages:** building images and updating Cloud Run to a specific immutable image tag.

Keep image tags out of normal Terraform drift correction. Otherwise a later infrastructure apply can unintentionally roll an application release back.

> [!WARNING]
> Protect Firestore from accidental deletion with the team's approved lifecycle/deletion-protection approach. A Terraform destroy must not silently remove production quotes.

## 3. Make the application ready

Before deploying, implement and test these changes in this repository:

The application selects its repository from environment variables. Cloud Run must provide:

| Variable | Cloud Run value | Local default / purpose |
| --- | --- | --- |
| `PORT` | Set automatically by Cloud Run | `8080` for REST, `8081` for GraphQL |
| `STORAGE_MODE` | `firestore` | `memory` |
| `FIRESTORE_PROJECT_ID` | GCP project ID | Required in Firestore mode |
| `FIRESTORE_DATABASE_ID` | Optional non-default database ID | Empty uses Firestore's `(default)` database |
| `FIRESTORE_COLLECTION` | `quotes` or the Terraform-approved collection | `quotes` |
| `SEED_QUOTES` | Do not set | Set to `true` only for intentional local memory demos |

Both entrypoints handle Cloud Run `SIGTERM` and drain HTTP requests for up to 30 seconds. They never seed Firestore; production sample data is intentionally out of scope.

The Firestore adapter implements [the existing repository contract](../internal/repository/quote_repository.go). It stores quote UUIDs as document IDs and uses `text`, `author`, and `createdAt` fields. Terraform must define the composite index for queries ordered by `author`, `createdAt`, and document ID when it is required by the selected Firestore database configuration.

Run Firestore emulator integration tests before deployment. Start the emulator using your team's approved Firebase/Google Cloud tooling, set its endpoint, then run:

```bash
export FIRESTORE_EMULATOR_HOST="127.0.0.1:8085"
export FIRESTORE_PROJECT_ID="quotes-api-emulator"
gofmt -w .
go vet ./...
go test ./...
```

The REST API should keep its health check at `/v1/health`. The GraphQL API serves the playground at `/` and accepts operations at `/query`.

## 4. Build images and publish them

After Terraform creates the Artifact Registry repository, configure Docker authentication:

```bash
gcloud auth configure-docker "${REGION}-docker.pkg.dev"
```

Tag each image with the exact source revision rather than `latest`:

```bash
export IMAGE_TAG="$(git rev-parse --short HEAD)"
export IMAGE_BASE="${REGION}-docker.pkg.dev/${PROJECT_ID}/${REPOSITORY}"

docker build -f build/Dockerfile.rest \
  -t "${IMAGE_BASE}/rest:${IMAGE_TAG}" .
docker build -f build/Dockerfile.graphql \
  -t "${IMAGE_BASE}/graphql:${IMAGE_TAG}" .

docker push "${IMAGE_BASE}/rest:${IMAGE_TAG}"
docker push "${IMAGE_BASE}/graphql:${IMAGE_TAG}"
```

Confirm that both images exist:

```bash
gcloud artifacts docker images list "$IMAGE_BASE"
```

## 5. Deploy the two Cloud Run services

Terraform should create the services and their configuration first. For an initial learning deployment, update each service to its immutable image:

```bash
gcloud run services update "$REST_SERVICE" \
  --image="${IMAGE_BASE}/rest:${IMAGE_TAG}" \
  --region="$REGION"

gcloud run services update "$GRAPHQL_SERVICE" \
  --image="${IMAGE_BASE}/graphql:${IMAGE_TAG}" \
  --region="$REGION"
```

The Cloud Run service configuration must:

- attach the Terraform-created runtime service account;
- set the app's storage mode and Firestore collection configuration;
- allow the Firestore API through IAM;
- use restricted ingress, such as `internal-and-cloud-load-balancing`, when IAP/load-balancer access is used;
- avoid unauthenticated direct access.

Check the deployed revisions and logs:

```bash
gcloud run services describe "$REST_SERVICE" --region="$REGION"
gcloud run services describe "$GRAPHQL_SERVICE" --region="$REGION"
gcloud logging read \
  "resource.type=cloud_run_revision AND resource.labels.service_name=${REST_SERVICE}" \
  --limit=20
```

## 6. Set up browser access with IAP

Cloud Run IAM is well suited to service-to-service callers, but it is inconvenient for a browser GraphQL playground. For team members signing in with Google accounts, use an external HTTPS Application Load Balancer with IAP.

In the infrastructure repository:

1. Create serverless NEGs for the REST and GraphQL Cloud Run services.
2. Add two backend services and route distinct paths, for example `/api/rest/*` and `/api/graphql/*`.
3. Attach the HTTPS load balancer to your DNS name and wait for the managed certificate to become active.
4. Enable IAP for both backends.
5. Grant the IAP HTTPS Resource Accessor role to the intended Google Group or users.

Verify with two accounts: an allowed account must reach the APIs after Google sign-in, and an unapproved account must be denied. Use the load-balancer URL, not the direct Cloud Run URL.

## 7. Automate future releases

Use two GitHub Actions workflows, aligned with the team's existing Terraform workflow:

1. **Infrastructure repository:** pull requests run formatting, validation, and a Terraform plan; protected-main applies approved infrastructure.
2. **Application repository:** pull requests run formatting, `go vet`, and `go test ./...`; protected-main builds both images, pushes SHA tags, and updates both Cloud Run services.

Configure GitHub Actions to use Workload Identity Federation. Do not save a downloaded Google service-account JSON key as a GitHub secret.

## 8. Verify and roll back

After a release:

1. Open the IAP-protected GraphQL path in a browser and run a health query.
2. Call REST `GET /v1/health` through the IAP-protected path.
3. Create a quote through REST and read it through GraphQL.
4. Deploy a new revision or allow Cloud Run to scale out; confirm the quote remains available.
5. Check Cloud Run logs for errors.

To roll back, redeploy the previous known-good image tag to both services. Keep every successful SHA tag in Artifact Registry so a rollback never depends on rebuilding old source.

## Next improvements

- Add Cloud Monitoring uptime checks and alert policies.
- Set Cloud Run minimum/maximum instances after measuring traffic and cost.
- Add rate limiting or Cloud Armor if the service is exposed beyond the team.
- Use Firebase Authentication instead of IAP if the API will serve consumer users rather than an internal Google-account group.
