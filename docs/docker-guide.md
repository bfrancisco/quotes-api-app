# Docker Build and Run Guide

This repository provides separate Docker images for the REST and GraphQL APIs.

## Prerequisites

- Docker Desktop or Docker Engine is installed and running.
- Run all commands from the repository root.

Confirm Docker is available:

```bash
docker version
```

## Build the REST API image

```bash
docker build --no-cache -f build/Dockerfile.rest \
  -t quotes-rest-api \
  -t us-east4-docker.pkg.dev/ls-devx-int-np-e7d3/bryan-quotes-api/quotes-rest-api:latest \
  .
```

Run the REST API, mapping the container's port `8080` to the same local port:

```bash
docker run --rm --name quotes-rest-api -p 8080:8080 quotes-rest-api
```

In another terminal, check the service:

```bash
curl http://localhost:8080/v1/health
```

Stop the foreground container with `Ctrl+C`. The `--rm` option removes the stopped container automatically.

## Build the GraphQL API image

```bash
docker build --no-cache -f build/Dockerfile.graphql \
  -t quotes-graphql-api \
  -t us-east4-docker.pkg.dev/ls-devx-int-np-e7d3/bryan-quotes-api/quotes-graphql-api:latest \
  .
```

Run the GraphQL API:

```bash
docker run --rm --name quotes-graphql-api -p 8081:8081 quotes-graphql-api
```

Open the GraphQL playground at:

```text
http://localhost:8081/
```

Send a GraphQL request to `/query`. For example:

```bash
curl -X POST http://localhost:8081/query \
  -H 'Content-Type: application/json' \
  -d '{"query":"{ quotes { id text author } }"}'
```

Stop the foreground container with `Ctrl+C`.

## Run containers in the background

Use detached mode (`-d`) to keep a container running after the terminal is available again:

```bash
docker run -d --name quotes-rest-api -p 8080:8080 quotes-rest-api
docker run -d --name quotes-graphql-api -p 8081:8081 quotes-graphql-api
```

View active containers:

```bash
docker ps
```

View logs:

```bash
docker logs quotes-rest-api
docker logs quotes-graphql-api
```

Follow logs continuously:

```bash
docker logs --follow quotes-rest-api
```

Stop and remove background containers:

```bash
docker rm --force quotes-rest-api quotes-graphql-api
```

## List and remove images

List local images:

```bash
docker image ls
```

Remove the application images when they are no longer needed:

```bash
docker image rm quotes-rest-api quotes-graphql-api
```

## Docker push to Artifact Registry repository
```bash
docker push <artifact_registry_address>/<image-tag>
```

## GitHub Actions publishing and deployment

The GitHub Actions workflow builds both images with BuildKit caching after validation succeeds on `main`. It publishes each image with the full Git commit SHA and the mutable `latest` tag, then updates the matching Cloud Run service with the SHA-tagged image.

| Image | Cloud Run service |
| --- | --- |
| `quotes-rest-api` | `bryan-quotes-rest-api` |
| `quotes-graphql-api` | `bryan-quotes-graphql-api` |

`latest` is convenient for local testing, but deployments must use the immutable SHA tag (or a digest). The workflow updates images only; Terraform remains responsible for Cloud Run configuration and IAM.

## WSL Docker Desktop credential-helper workaround

On WSL, Docker Desktop may fail before building with an error similar to:

```text
error getting credentials: fork/exec /usr/bin/docker-credential-desktop.exe: exec format error
```

That is a Docker Desktop / WSL interoperability issue, rather than a Dockerfile or application error. To build public images without using the broken credential helper, use a temporary empty Docker configuration:

```bash
temp_docker_config=$(mktemp -d)
DOCKER_CONFIG="$temp_docker_config" docker build --no-cache -f build/Dockerfile.rest -t quotes-rest-api .
rm -rf "$temp_docker_config"
```

For GraphQL, replace the Dockerfile path and image name:

```bash
temp_docker_config=$(mktemp -d)
DOCKER_CONFIG="$temp_docker_config" docker build --no-cache -f build/Dockerfile.graphql -t quotes-graphql-api .
rm -rf "$temp_docker_config"
```

This workaround is suitable for public Docker Hub images. It does not provide credentials for private registries.

For a permanent fix, quit Docker Desktop, run `wsl --shutdown` in Windows PowerShell, then restart Docker Desktop and open a new WSL terminal.

## Rebuild after source changes

Run the appropriate build command again after changing Go source code or Dockerfile instructions. `--no-cache` forces every build layer to run again; omit it for faster builds that reuse unchanged cached layers.
