<p align="center">
  <img src="https://raw.githubusercontent.com/IceyWu/go-api-starter/main/public/logo.svg" alt="go-api-starter logo" width="112" height="112" />
</p>

<h1 align="center">go-api-starter</h1>

<p align="center">
  A production-oriented Go API starter for authentication, RBAC, file uploads, and asynchronous media processing.
</p>

<p align="center">
  <a href="README.zh-CN.md">简体中文</a> · English
</p>

<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/github/license/IceyWu/go-api-starter" alt="License" /></a>
  <a href="https://github.com/IceyWu/go-api-starter/actions/workflows/ci.yml"><img src="https://img.shields.io/github/actions/workflow/status/IceyWu/go-api-starter/ci.yml?label=CI" alt="CI status" /></a>
  <a href="go.mod"><img src="https://img.shields.io/github/go-mod/go-version/IceyWu/go-api-starter" alt="Go version" /></a>
</p>

`go-api-starter` is built with Chi, Huma, sqlc, Atlas, and Task. It provides a structured foundation for production APIs while keeping infrastructure concerns explicit and replaceable.

## Highlights

- Layered architecture with `handler`, `service`, `repository`, `router`, and dependency injection.
- JWT access/refresh authentication, Argon2 password hashing, and token blacklist support.
- Permission spaces, roles, CRUD permissions, and route-level authorization.
- S3-compatible direct upload, multipart upload, resumable upload, instant upload, and persisted file records.
- Optional independent Alibaba Cloud MPS Worker for asynchronous video transcoding, polling, and result persistence.
- WebSocket Hub with API key authentication, heartbeat, commands, and acknowledgements.
- WeChat Mini Program login, Redis distributed rate limiting, and in-memory fallback.
- Health checks, Prometheus metrics, Scalar API documentation, OpenAPI JSON, and `llms.txt`.

## Architecture

```text
HTTP client
    -> Chi transport and middleware
    -> Huma / compatibility route adapters
    -> handlers
    -> services
    -> repositories and platform adapters
    -> MySQL / Redis / S3-compatible storage / MPS

MPS Worker
    -> task manager
    -> Alibaba Cloud MPS
    -> polling and result persistence
    -> optional webhook notification
```

The API process does not perform local video transcoding. When enabled, it creates and tracks MPS tasks; the independent Worker polls MPS and persists the resulting video variants.

## Quick start

### Requirements

- Go 1.27.2+
- [Task](https://taskfile.dev/)
- MySQL 8+ and Redis for a production-like environment
- Docker Compose (optional)

### Local development

```bash
git clone https://github.com/IceyWu/go-api-starter.git
cd go-api-starter
cp .env.example .env.dev

# Install the repository-pinned Task version.
go install github.com/go-task/task/v3/cmd/task@v3.54.0

task deps
task dev
```

All project commands are managed by `Taskfile.yml`:

```bash
task --list
task test
task check
task test CLI_ARGS="-race"
task security
task migrate
task worker
```

### Docker Compose

Start MySQL, Redis, and the API with automatic Atlas migrations:

```bash
docker compose up --build api
```

Docker 镜像构建默认将 Go 包编译并发限制为 2，避免在资源较小的部署服务器上占满内存。资源受限时可显式设为 1：

```bash
docker compose build --build-arg GO_BUILD_PARALLELISM=1 api
docker compose up api
```

有更多构建资源时，可通过 `GO_BUILD_PARALLELISM` build arg 提高并发；该值只限制 Docker 构建阶段，不限制运行中的 API 容器。

Start the MPS Worker as well:

```bash
docker compose --profile worker up --build
```

Compose credentials are intended for local development only. Replace every credential and secret before using another environment.

`task docker-smoke` verifies a disposable API/MySQL/Redis stack without host ports or persistent volumes. `task test-integration` accepts `MYSQL_TEST_DSN` and `REDIS_TEST_ADDR`; `REQUIRE_INTEGRATION=true` rejects missing service configuration. Race tests use `task test CLI_ARGS="-race"` and require a C compiler with `CGO_ENABLED=1`.

Compose pins MySQL 26.7.0 and Redis 8.10.2. MySQL uses a separate `mysql-data-26-7` volume; existing `mysql-data` volumes are preserved. To reuse data from MySQL 8.4, back it up and follow the [supported upgrade path](https://dev.mysql.com/doc/refman/26.7/en/upgrade-paths.html) through MySQL 9.7 LTS, rather than attaching the old volume directly.

`task update` upgrades the stable Go toolchain and application/test modules. Run `task check`, `task lint`, `task security`, and `task atlas-validate` afterward. Tool and container versions are pinned separately in Taskfile, Dockerfile, Compose, and CI.

## API documentation

| Path | Description |
| --- | --- |
| `/docs` | Scalar API documentation (Basic Auth) |
| `/openapi.json` | Complete OpenAPI document (Basic Auth) |
| `/swagger/doc.json` | OpenAPI endpoint for legacy clients |
| `/llms.txt` | AI-readable API overview |
| `/llms-full.txt` | AI-readable full API documentation |
| `/health` | Liveness check |
| `/health/ready` | Database and cache readiness check |
| `/metrics` | Prometheus metrics |
| `/ws` | WebSocket entry point |

Upload sessions use a provider-neutral flow:

```text
POST   /api/v1/uploads
POST   /api/v1/uploads/{id}/complete
DELETE /api/v1/uploads/{id}
```

The client only receives presigned URLs and submits part ETags. Object keys,
size, and content type are controlled and verified by the server.

The default development port is `9527`; the default production port is `8080`.

## Configuration

The single configuration source is [`config/config.yaml`](config/config.yaml). Select the environment with `APP_ENV=development` or `APP_ENV=production`.

Environment variables use the `GO_API_` prefix and override the final configuration. Double underscores represent nested keys:

```text
GO_API_SERVER__PORT=9000
GO_API_DATABASE__PASSWORD=replace-me
GO_API_APP__JWT_SECRET=replace-with-at-least-32-characters
```

See the complete [configuration guide](docs/configuration.md), [`AGENTS.md`](AGENTS.md#environment-variables), and [`.env.example`](.env.example).

For local development, `task dev` loads `.env.dev` and the optional ignored `.env.dev.local`; keep storage credentials in the latter.

## Development and verification

```bash
task fmt
task sqlc
task test
task test-integration
task check
task atlas-validate
task security
task build-all TARGET_OS=linux TARGET_ARCH=amd64
```

CI starts MySQL and Redis, applies migrations, and runs unit/integration tests together with the race detector once. It also verifies the Docker stack. Tests stay alongside their Go packages; external-service tests use the `integration` build tag. `cmd/` contains only server, worker, and migration entry points.

## Project layout

```text
cmd/                 server, worker, and migration entry points
config/              shared and environment-specific configuration
db/                  SQL schema and sqlc queries
docs/                version-controlled OpenAPI documentation
internal/            application code and platform adapters
migrations/          SQLite and MySQL Atlas migrations
public/              logo and static assets
Taskfile.yml         project command constraints
Dockerfile           production container build
docker-compose.yml   API, MySQL, Redis, and Worker orchestration
```

## License

MIT
