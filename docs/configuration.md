# Configuration guide

[简体中文](configuration.zh-CN.md)

This project has one normal configuration source: [`../config/config.yaml`](../config/config.yaml). Environment files are templates or compatibility helpers; they are not additional configuration sources loaded automatically by the application.

## Configuration files

| File | Purpose | Loaded automatically? |
| --- | --- | --- |
| `config/config.yaml` | Shared defaults plus `development` and `production` profiles. | Yes |
| `.env.example` | Documented environment-variable template for local setup and deployment reference. | No |
| `.env.dev` | Deprecated development compatibility file. Use `APP_ENV=development` and explicit `GO_API_*` variables instead. | No |
| `.env.prod` | Deprecated production compatibility file. Use `APP_ENV=production` and a secret manager or deployment environment instead. | No |
| `CONFIG_FILE` target | Optional replacement YAML file selected by the `CONFIG_FILE` environment variable. | Only when explicitly configured |

The `.env.dev` and `.env.prod` files are not layered on top of `config/config.yaml` by the Go application. If you source one from a shell, its exported variables become normal environment-variable overrides.

## Profile and override order

Set `APP_ENV` to choose a profile. Supported values are `development`, `dev`, `production`, and `prod`; the default is `development`.

Values are resolved in this order, from lowest to highest priority:

```text
common
  -> selected profile (development or production)
  -> GO_API_* environment variables
```

For example:

```text
GO_API_SERVER__PORT=9000
  -> server.port = 9000

GO_API_DATABASE__PASSWORD=change-me
  -> database.password = change-me

GO_API_APP__JWT_SECRET=replace-with-at-least-32-characters
  -> app.jwt_secret = replace-with-at-least-32-characters
```

`GO_API_` variables use double underscores for nested keys. A single underscore remains part of a field name.

## Configuration groups

The YAML file is organized by capability:

| Group | Responsibility |
| --- | --- |
| `app` | Application identity, JWT, documentation credentials, and default account settings. |
| `server` | Bind address, HTTP timeouts, shutdown timeout, and request limits. |
| `database` | MySQL/SQLite connection, pool, and migration settings. |
| `log` | Log level, format, and output destination. |
| `storage` | S3-compatible endpoint, bucket, credentials, public URL, and upload settings. |
| `redis` | Optional cache, blacklist, rate-limit, and fallback settings. |
| `transcoding` | Optional Alibaba Cloud MPS region, pipeline, templates, polling, and worker settings. Set `enabled: true` only when MPS is configured. |
| `cors` | Allowed origins and HTTP cross-origin behavior. |
| `rate_limit` | Global, user, login, upload, and fallback rate limits. |
| `mail` | SMTP server and verification-mail settings. |
| `wechat` | WeChat Mini Program credentials and login settings. |
| `ws` | WebSocket origin, authentication, heartbeat, and connection limits. |

Video transcoding uses Alibaba Cloud MPS. The API creates and tracks tasks; the independent Worker polls MPS and persists results.

## Recommended usage

Local development:

```powershell
$env:APP_ENV = "development"
$env:GO_API_APP__JWT_SECRET = "local-only-secret"
task dev
```

Production:

```text
APP_ENV=production
GO_API_APP__JWT_SECRET=<secret-manager-value>
GO_API_APP__DOCS_PASSWORD=<secret-manager-value>
GO_API_DATABASE__PASSWORD=<secret-manager-value>
```

Do not commit real credentials. Production must replace JWT secrets, documentation credentials, administrator credentials, default user credentials, CORS origins, and external-service credentials.

## Adding a setting

1. Add the shared default or profile-specific value to `config/config.yaml`.
2. Add the matching strongly typed field in `internal/config/config.go`.
3. Use the derived `GO_API_*` name when an environment override is needed; no separate binding code is required.
4. Add or update configuration tests when the setting changes validation or runtime behavior.
