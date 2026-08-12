# GitHub Actions Implementation Plan

## Goal

Automate validation, container publishing, and Cloud Run redeployment for both APIs:

- Run Go formatting, vet, and tests, including Firestore emulator tests.
- Build the REST and GraphQL Docker images on pull requests.
- Push both images to GCP Artifact Registry after a successful push to `main`.
- Redeploy both Cloud Run services with their immutable SHA-tagged images.

## 1. Provision GCP identity in the infrastructure repository

- Verify Artifact Registry repository `bryan-quotes-api` exists in project `ls-devx-int-np-e7d3`, region `us-east4`.
- Create a dedicated GitHub Actions service account.
- Grant it `roles/artifactregistry.writer` on that repository only.
- Grant `roles/run.developer` only on `bryan-quotes-rest-api` and `bryan-quotes-graphql-api` in `us-east4`.
- Grant `roles/iam.serviceAccountUser` only on the shared Cloud Run runtime identity `sa-bryan-quotes-api@ls-devx-int-np-e7d3.iam.gserviceaccount.com`.
- Generate one user-managed key for this service account and store its JSON only as a GitHub Environment secret.
- Do not grant project-wide Cloud Run or Artifact Registry permissions.

## 2. Configure GitHub

Create the `quotes-api-deploy` environment, restrict it to `main`, and add these environment variables:

- `GCP_PROJECT_ID=ls-devx-int-np-e7d3`
- `GCP_REGION=us-east4`
- `ARTIFACT_REGISTRY_REPOSITORY=bryan-quotes-api`
- `REST_IMAGE_NAME=quotes-rest-api`
- `GRAPHQL_IMAGE_NAME=quotes-graphql-api`
- `REST_IMAGE_NAME=quotes-rest-api`
- `GRAPHQL_IMAGE_NAME=quotes-graphql-api`
- Secret: `GCP_SERVICE_ACCOUNT_KEY=<entire JSON content of the key for bryan-github-actions@ls-devx-int-np-e7d3.iam.gserviceaccount.com>`

The first five entries are GitHub Environment variables; `GCP_SERVICE_ACCOUNT_KEY` is a GitHub Environment secret. Restrict the `quotes-api-deploy` environment to `main`. Keep the default workflow token read-only; OIDC `id-token: write` is not required.

## 3. Add the workflow

Create one workflow under `.github/workflows/` with actions pinned to full commit SHAs.

### Pull requests targeting `main`

1. Install Go using the version in `go.mod` and enable module caching.
2. Start the Firestore emulator on `127.0.0.1:8085` and wait for readiness.
3. Set `FIRESTORE_EMULATOR_HOST` and `FIRESTORE_PROJECT_ID=quotes-api-emulator`.
4. Run:
   - Non-mutating `gofmt` check
   - `go vet ./...`
   - `go test ./...`
5. Build both `linux/amd64` images without pushing, using BuildKit caching:
   - `build/Dockerfile.rest`
   - `build/Dockerfile.graphql`

Pull-request jobs must not receive GCP credentials or access the publishing environment.

### Pushes to `main`

1. Require the Go checks to pass.
2. Authenticate to GCP with the `GCP_SERVICE_ACCOUNT_KEY` Environment secret.
3. Build both images successfully before pushing.
4. Push each image with the full commit SHA and `latest` tags:
   - `us-east4-docker.pkg.dev/ls-devx-int-np-e7d3/bryan-quotes-api/quotes-rest-api`
   - `us-east4-docker.pkg.dev/ls-devx-int-np-e7d3/bryan-quotes-api/quotes-graphql-api`
5. Update only the image for each Cloud Run service, using the exact SHA-tagged image:
   - `bryan-quotes-rest-api`
   - `bryan-quotes-graphql-api`
6. Wait for each revision to become ready and shift 100% traffic to the latest revision.
7. Report immutable image references, deployed services, and partial-deployment state in the workflow summary.

Deploy REST first and then GraphQL. If either deployment fails, fail the workflow and do not automatically roll back a service that was already updated. Terraform remains the owner of all non-image Cloud Run configuration, including environment variables, secrets, runtime service accounts, ingress, IAM, scaling, and traffic policy.

Add concurrency controls to cancel superseded PR runs and prevent overlapping `main` publications.

## 4. Protect `main`

After the first workflow run establishes check names, configure a GitHub ruleset requiring:

- Go format, vet, and test checks
- REST image build
- GraphQL image build

Prevent merges that bypass required checks.

## Verification

- Validate workflow syntax with `actionlint`.
- Confirm Firestore tests run rather than skip.
- Confirm fork pull requests cannot authenticate or push images.
- Test controlled formatting, test, and Docker build failures.
- Merge to `main` and verify both SHA and `latest` tags in Artifact Registry.
- Verify both Cloud Run services use the intended SHA-tagged images and their latest revisions are ready with 100% traffic.
- Verify IAM is repository/service-scoped, the key is available only through the `quotes-api-deploy` Environment, and no key contents appear in logs.

## Important constraints

- The service account, its least-privilege IAM bindings, and its GitHub Environment key secret must exist before image publishing or deployment can succeed.
- A service-account key is a long-lived credential: create only one dedicated key, restrict the GitHub Environment to `main`, rotate it on a defined schedule and immediately after suspected exposure, then disable and delete the replaced key.
- `latest` is mutable; deployments should use the SHA tag or image digest.
- Two image pushes are not atomic, although building both first reduces partial publication risk.
- Cloud Run updates are image-only. The workflow does not change runtime configuration or roll back partial deployments automatically.
- Multi-architecture builds, image signing, SBOMs, vulnerability scanning, race detection, and coverage are outside this initial scope.
