# Configuration guide

[简体中文](configuration.zh-CN.md)

This project has one normal application configuration source: [`../config/config.yaml`](../config/config.yaml). The `task dev` command additionally loads `.env.dev` and the optional `.env.dev.local` file as environment-variable overrides.

## Configuration files

| File | Purpose | Loaded automatically? |
| --- | --- | --- |
| `config/config.yaml` | Shared defaults plus `development` and `production` profiles. | Yes |
| `.env.example` | Documented environment-variable template for local setup and deployment reference. | No |
| `.env.dev` | Development overrides loaded by `task dev`. Keep non-secret local defaults here. | By `task dev` |
| `.env.dev.local` | Ignored local development overrides for credentials and machine-specific values. | By `task dev` when present |
| `.env.prod` | Deprecated production compatibility file. Use `APP_ENV=production` and a secret manager or deployment environment instead. | No |
| `CONFIG_FILE` target | Optional replacement YAML file selected by the `CONFIG_FILE` environment variable. | Only when explicitly configured |

The Go application itself reads `config/config.yaml` and normal environment variables. `task dev` loads `.env.dev` first and `.env.dev.local` second, so the local file is the right place for secrets. `.env.prod` remains a deployment template and is not loaded by the application.

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

## S3-compatible object storage

The upload layer uses the AWS SDK for Go v2 through an S3-compatible provider interface. The same settings work with Amazon S3, Alibaba Cloud OSS, Cloudflare R2, and other compatible providers.

Amazon S3 in `ap-southeast-2`:

```text
GO_API_STORAGE__ENDPOINT=https://s3.ap-southeast-2.amazonaws.com
GO_API_STORAGE__BUCKET=go-api-starter-s3
GO_API_STORAGE__REGION=ap-southeast-2
GO_API_STORAGE__UPLOAD_DIR=go_api
GO_API_STORAGE__PUBLIC_BASE_URL=
GO_API_STORAGE__FORCE_PATH_STYLE=false
GO_API_STORAGE__ACCESS_KEY_ID=<local-secret>
GO_API_STORAGE__ACCESS_KEY_SECRET=<local-secret>
```

Keep the access key variables in `.env.dev.local` or a deployment secret manager. The bucket should remain private; uploads use presigned URLs. The IAM principal needs bucket-level `ListBucket` and multipart-list permissions, plus object-level `GetObject`, `PutObject`, `DeleteObject`, `AbortMultipartUpload`, and `ListMultipartUploadParts` permissions.

## File metadata and reverse geocoding

When the playground's “process media metadata” option is enabled, the browser submits EXIF, capture time, GPS, device, dimensions, and image features. The server accepts common EXIF time formats and can enrich GPS coordinates through AMap when configured:

```text
GO_API_APP__AMAP_API_KEY=<AMap Web service API key>
```

Without this key, the server does not call an external geocoding service; GPS coordinates are still stored and address fields remain empty.

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

`task dev` loads `.env.dev` and then `.env.dev.local`. The standalone upload page can be served without live reload:

```powershell
task upload-demo
```

Open `http://127.0.0.1:5501/upload-demo.html`. This avoids Live Server refreshing the page when the API updates the local database.

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
