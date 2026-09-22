# 配置说明

[English](configuration.md)

项目只有一个正式的普通配置源：[`../config/config.yaml`](../config/config.yaml)。环境变量文件只是模板或兼容辅助文件，应用不会自动加载它们。

## 配置文件

| 文件 | 用途 | 是否自动加载 |
| --- | --- | --- |
| `config/config.yaml` | 维护公共默认值以及 `development`、`production` 配置段。 | 是 |
| `.env.example` | 环境变量模板和部署参考。 | 否 |
| `.env.dev` | 已废弃的开发环境兼容文件。建议使用 `APP_ENV=development` 和明确的 `GO_API_*` 变量。 | 否 |
| `.env.prod` | 已废弃的生产环境兼容文件。建议使用 `APP_ENV=production` 和部署平台或密钥管理器。 | 否 |
| `CONFIG_FILE` 指向的文件 | 通过 `CONFIG_FILE` 环境变量显式指定的替代 YAML 配置文件。 | 显式配置时加载 |

`.env.dev` 和 `.env.prod` 不会由 Go 应用自动叠加到 `config/config.yaml`。如果在 Shell 中手动执行或加载它们，其中导出的变量会作为普通环境变量覆盖配置。

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

## 配置分组

YAML 配置按照功能划分：

| 分组 | 职责 |
| --- | --- |
| `app` | 应用名称、JWT、文档认证和默认账号设置。 |
| `server` | 监听地址、HTTP 超时、关闭超时和请求限制。 |
| `database` | MySQL/SQLite 连接、连接池和迁移设置。 |
| `log` | 日志级别、格式和输出目标。 |
| `oss` | 阿里云 OSS 凭据、Bucket、Endpoint 和上传设置。 |
| `redis` | 可选缓存、Token 黑名单、限流和降级设置。 |
| `transcoding` | 阿里云 MPS 地域、管道、模板、轮询和 Worker 设置。 |
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
