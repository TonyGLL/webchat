**🌐 English** · [Español](../es/ci-cd.md)

# CI/CD and container images

How a change becomes a verified, versioned artifact that can run anywhere (today a laptop or a VM, later AWS). Nothing here is tied to one cloud: the deliverable is a pair of container images.

## 1. The pipeline

```mermaid
flowchart LR
    PR[Pull request] --> B
    PR --> F
    PUSH[Push to main<br/>or tag v*] --> B
    PUSH --> F
    subgraph CI["ci.yml"]
      B[Backend<br/>gofmt · vet · build · test -race<br/>Postgres + Redis services]
      F[Frontend<br/>tsc · lint · test · build]
      B --> I
      F --> I
      I[Images<br/>build api + web]
    end
    I -->|pull request| NOPUSH[Build only]
    I -->|main or tag| GHCR[(ghcr.io<br/>webchat-api · webchat-web)]
    CRON[Weekly + dependency changes] --> SEC
    subgraph SECW["security.yml"]
      SEC[govulncheck<br/>bun audit]
    end
    DEP[Dependabot<br/>weekly PRs] -.-> PR
```

| Workflow | Runs on | What it guarantees |
|---|---|---|
| `ci.yml` · backend | every push and pull request | formatting, `go vet`, compilation, all tests with the race detector, and that **every migration applies to an empty PostgreSQL** (the test is skipped locally unless `TEST_DATABASE_URL` points at an empty database) |
| `ci.yml` · frontend | same | type check, lint, tests, production build |
| `ci.yml` · images | after both jobs pass | both Dockerfiles build; on `main` and `v*` tags the images are also published |
| `security.yml` | weekly, and when dependency files change | no reachable Go vulnerability (`govulncheck`) and no high-severity JavaScript advisory (`bun audit`) |
| Dependabot | weekly | pull requests for Go, JavaScript, Docker base images and the workflow actions themselves |

## 2. The images

| Image | Base | Size | Notes |
|---|---|---|---|
| `ghcr.io/<owner>/webchat-api` | `distroless/static`, non-root | ≈ 40 MB | static Go binary; no shell, no package manager |
| `ghcr.io/<owner>/webchat-web` | `node:22-alpine`, non-root | ≈ 315 MB | Next.js `standalone` output; listens on 3100 |

Tags: `sha-<commit>` for every build of `main`, `main` and `latest` for the branch head, and `1.2.3` for a `v1.2.3` tag. Published images are **multi-architecture** (`amd64` and `arm64`), so they run on ordinary x86 hosts and on ARM ones such as AWS Graviton.

The frontend lockfile (`frontend/bun.lock`) is committed: CI and the image install exactly the versions you tested (`--frozen-lockfile`).

Build them yourself:

```bash
docker build -t webchat-api backend
docker build -t webchat-web --build-arg NEXT_PUBLIC_API_URL=https://api.example.com/api/v1 frontend
```

## 3. Decisions behind it

See **D19** in [Design decisions](decisions.md). In short:

- **Images are the deliverable.** Any platform (ECS, EKS, a VM with Docker, Kubernetes) can run an image; scripts tied to one provider cannot move.
- **GHCR first.** It needs no extra credentials (the workflow's own token is enough), so the pipeline works from day one. Switching to Amazon ECR later changes only the login and tag steps.
- **One web image per environment.** Next.js inlines `NEXT_PUBLIC_*` values into the browser bundle at build time, so the API URL is a build argument (set the repository variable `NEXT_PUBLIC_API_URL`). The alternative, reading the URL at runtime, would need a config endpoint or a startup script; it is simple to add if more environments appear.
- **Security checks are separate and non-blocking for unrelated PRs.** A new advisory in a dependency appears without anyone touching the code; it should raise a visible alarm, not stop an unrelated fix.

## 4. Getting ready for AWS

This section records what the application already provides and what is still a decision, so nobody assumes a feature exists that does not. Nothing below is implemented in the repository.

**Already true**

| Property | Why it matters in the cloud |
|---|---|
| Stateless backend, state in PostgreSQL and Redis | run several copies behind a load balancer; scale and replace freely |
| Configuration only through environment variables | maps to task definitions, parameter stores and secret managers without code changes |
| `GET /health` and `GET /ready` | ready-made targets for load-balancer and orchestrator health checks |
| WebSocket events relayed through Redis | any instance can serve any room, so no sticky sessions are needed |
| Heartbeat every ~54 s | stays under the 60 s default idle timeout of an AWS load balancer, so connections are not cut |
| Migrations run at startup under a database lock | several instances can start together without corrupting the schema |
| Non-root, minimal images, multi-architecture | smaller attack surface; cheaper ARM instances are an option |

**Decisions for later**

| Topic | Notes |
|---|---|
| Registry | Amazon ECR instead of GHCR: authenticate the workflow with an **OIDC role** (`aws-actions/configure-aws-credentials`) rather than stored access keys |
| Where it runs | ECS on Fargate, EKS or EC2: all consume the same images |
| Database and Redis | RDS for PostgreSQL and ElastiCache. Refresh tokens live in Redis, so enable persistence or accept that a Redis loss signs everyone out |
| Secrets | `JWT_SECRET`, database and SMTP credentials belong in a secret manager, not in images or compose files (finding H-051) |
| Domains and cookies | The refresh cookie is `SameSite=Strict`: serve the web app and the API from the same registrable domain (for example `app.example.com` and `api.example.com`) |
| TLS and CORS | `CORS_ALLOWED_ORIGINS` must list the real web origin; set secure cookies in production |
| Observability | metrics, traces and alerts are still pending in roadmap phase 5 |

## 5. One-time repository settings

These cannot live in code, so they are easy to forget:

1. **Settings → Actions → General → Workflow permissions:** allow read access; the image job requests `packages: write` itself.
2. **Settings → Secrets and variables → Actions → Variables:** add `NEXT_PUBLIC_API_URL` (without it the web image points at `http://localhost:8081/api/v1`).
3. **Settings → Branches:** require the `Backend (Go)`, `Frontend (Next.js)` and image checks before merging into `main`.
4. After the first publish, set each package's visibility under **Packages**.

## 6. Run the same checks locally

```bash
# backend/ (the migration test needs an EMPTY database; this one is thrown away)
docker run -d --rm --name ci-db -e POSTGRES_PASSWORD=ci -e POSTGRES_DB=ci -p 127.0.0.1:55432:5432 postgres:16-alpine
TEST_DATABASE_URL='postgres://postgres:ci@127.0.0.1:55432/ci?sslmode=disable' go test -race ./...
go run golang.org/x/vuln/cmd/govulncheck@v1.1.4 ./...

# frontend/
bun run lint && bun run test && bun audit --audit-level=high
```
