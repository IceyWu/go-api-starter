# 配置说明

[English](configuration.md)

项目只有一个正式的应用配置源：[`../config/config.yaml`](../config/config.yaml)。`task dev` 会额外加载 `.env.dev` 和可选的 `.env.dev.local`，作为环境变量覆盖。

## 配置文件

| 文件 | 用途 | 是否自动加载 |
| --- | --- | --- |
| `config/config.yaml` | 维护公共默认值以及 `development`、`production` 配置段。 | 是 |
| `.env.example` | 环境变量模板和部署参考。 | 否 |
| `.env.dev` | `task dev` 加载的开发环境覆盖配置。只放非敏感的本地默认值。 | `task dev` 加载 |
| `.env.dev.local` | 被 Git 忽略的本地开发覆盖配置，用于凭据和机器相关配置。 | 存在时由 `task dev` 加载 |
| `.env.prod` | 已废弃的生产环境兼容文件。建议使用 `APP_ENV=production` 和部署平台或密钥管理器。 | 否 |
| `CONFIG_FILE` 指向的文件 | 通过 `CONFIG_FILE` 环境变量显式指定的替代 YAML 配置文件。 | 显式配置时加载 |

Go 应用本身读取 `config/config.yaml` 和普通环境变量。`task dev` 会先加载 `.env.dev`，再加载 `.env.dev.local`，因此本地密钥应放在后者。`.env.prod` 仍然只是部署模板，不由应用自动加载。

## 环境配置和覆盖优先级

通过 `APP_ENV` 选择配置环境，支持 `development`、`dev`、`production` 和 `prod`，默认值为 `development`。

配置优先级从低到高为：

```text
common
  -> 当前环境配置段（development 或 production）
  -> GO_API_* 环境变量
```

示例：

```text
GO_API_SERVER__PORT=9000
  -> server.port = 9000

GO_API_DATABASE__PASSWORD=change-me
  -> database.password = change-me

GO_API_APP__JWT_SECRET=replace-with-at-least-32-characters
  -> app.jwt_secret = replace-with-at-least-32-characters
```

`GO_API_` 后使用双下划线表示嵌套层级，单下划线仍属于字段名的一部分。

## S3 兼容对象存储

上传层通过 AWS SDK for Go v2 和 S3 兼容抽象工作。同一套配置可以支持 Amazon S3、阿里云 OSS、Cloudflare R2 等兼容对象存储。

Amazon S3 `ap-southeast-2` 示例：

```text
GO_API_STORAGE__ENDPOINT=https://s3.ap-southeast-2.amazonaws.com
GO_API_STORAGE__BUCKET=go-api-starter-s3
GO_API_STORAGE__REGION=ap-southeast-2
GO_API_STORAGE__UPLOAD_DIR=go_api
GO_API_STORAGE__PUBLIC_BASE_URL=
GO_API_STORAGE__FORCE_PATH_STYLE=false
GO_API_STORAGE__ACCESS_KEY_ID=<本地密钥>
GO_API_STORAGE__ACCESS_KEY_SECRET=<本地密钥>
```

访问密钥应放在 `.env.dev.local` 或部署平台的密钥管理器中。Bucket 保持私有，上传使用预签名 URL。IAM 身份需要 Bucket 级别的 `ListBucket`、分片列表权限，以及对象级别的 `GetObject`、`PutObject`、`DeleteObject`、`AbortMultipartUpload`、`ListMultipartUploadParts` 权限。

## 文件元信息与反向地理编码

上传页面开启“处理媒体元信息”后，浏览器会提交 EXIF、拍摄时间、GPS、设备、尺寸和图像特征。服务端会兼容 EXIF 时间格式，并在配置高德 Key 后根据 GPS 补充国家、省、市、区和地址：

```text
GO_API_APP__AMAP_API_KEY=<高德 Web 服务 API Key>
```

不配置 Key 时不会调用外部地理编码服务，GPS 坐标仍会保存，但地址字段保持为空。

## 配置分组

YAML 配置按照功能划分：

| 分组 | 职责 |
| --- | --- |
| `app` | 应用名称、JWT、文档认证和默认账号设置。 |
| `server` | 监听地址、HTTP 超时、关闭超时和请求限制。 |
| `database` | MySQL/SQLite 连接、连接池和迁移设置。 |
| `log` | 日志级别、格式和输出目标。 |
| `storage` | S3 兼容存储的 Endpoint、Bucket、凭据、公开 URL 和上传设置。 |
| `redis` | 可选缓存、Token 黑名单、限流和降级设置。 |
| `transcoding` | 可选的阿里云 MPS 地域、管道、模板、轮询和 Worker 设置。只有配置 MPS 时才设置 `enabled: true`。 |
| `cors` | 允许的来源和跨域行为。 |
| `rate_limit` | 全局、用户、登录、上传和降级限流。 |
| `mail` | SMTP 服务和验证码邮件设置。 |
| `wechat` | 微信小程序凭据和登录设置。 |
| `ws` | WebSocket 来源、认证、心跳和连接限制。 |

视频转码统一使用阿里云 MPS。API 只创建和跟踪任务，独立 Worker 负责轮询 MPS 并持久化结果。

## 推荐用法

开发环境：

```powershell
$env:APP_ENV = "development"
$env:GO_API_APP__JWT_SECRET = "local-only-secret"
task dev
```

`task dev` 会依次加载 `.env.dev` 和 `.env.dev.local`。独立上传页面可使用不带热重载的静态服务器启动：

```powershell
task upload-demo
```

然后打开 `http://127.0.0.1:5501/upload-demo.html`，避免 API 更新本地数据库时被 Live Server 自动刷新页面。

生产环境：

```text
APP_ENV=production
GO_API_APP__JWT_SECRET=<secret-manager-value>
GO_API_APP__DOCS_PASSWORD=<secret-manager-value>
GO_API_DATABASE__PASSWORD=<secret-manager-value>
```

不要提交真实凭据。生产环境必须替换 JWT 密钥、文档账号密码、管理员密码、默认用户密码、CORS 来源以及外部服务凭据。

## 新增配置项

1. 在 `config/config.yaml` 的 `common` 或环境配置段中添加默认值或环境差异。
2. 在 `internal/config/config.go` 中添加对应的强类型字段。
3. 需要环境变量覆盖时，使用推导出的 `GO_API_*` 名称，不需要额外增加绑定代码。
4. 如果配置会影响校验或运行时行为，同时补充或更新配置测试。
