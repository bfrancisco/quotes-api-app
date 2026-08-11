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
- Configure Workload Identity Federation for:
  - Repository: `bfrancisco/quotes-api-app`
  - Branch: `main`
  - GitHub environment: `quotes-api-deploy`
- Export the WIF provider resource name and service-account email.
- Do not create a service-account JSON key or grant project-wide Cloud Run or Artifact Registry permissions.

## 2. Configure GitHub

Create the `quotes-api-deploy` environment, restrict it to `main`, and add these environment variables:

- `GCP_PROJECT_ID=ls-devx-int-np-e7d3`
- `GCP_REGION=us-east4`
- `ARTIFACT_REGISTRY_REPOSITORY=bryan-quotes-api`
- `REST_IMAGE_NAME=quotes-rest-api`
- `GRAPHQL_IMAGE_NAME=quotes-graphql-api`
- `GCP_WORKLOAD_IDENTITY_PROVIDER=projects/408378328351/locations/global/workloadIdentityPools/github/providers/bryan-quotes-api-provider`
- `GCP_SERVICE_ACCOUNT=bryan-github-actions@ls-devx-int-np-e7d3.iam.gserviceaccount.com`

These are resource identifiers, not credentials. Keep the default workflow token read-only and grant `id-token: write` only to the publishing/deployment job.

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
2. Authenticate to GCP through WIF using a short-lived OIDC token.
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
- Verify IAM is repository/service-scoped and no JSON key exists.

## Important constraints

- WIF and IAM must exist before image publishing or deployment can succeed.
- `latest` is mutable; deployments should use the SHA tag or image digest.
- Two image pushes are not atomic, although building both first reduces partial publication risk.
- Cloud Run updates are image-only. The workflow does not change runtime configuration or roll back partial deployments automatically.
- Multi-architecture builds, image signing, SBOMs, vulnerability scanning, race detection, and coverage are outside this initial scope.
