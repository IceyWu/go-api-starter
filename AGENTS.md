# AGENTS.md

## Project overview

Go API Starter — 基于 Go + Gin + GORM 的 RESTful API 脚手架。采用清晰分层架构：model → repository → service → handler → router，通过 DI container 管理依赖。

技术栈：Gin、GORM（SQLite/MySQL）、Zap 日志、Viper 配置、Redis 缓存（可选）、阿里云 OSS、WebSocket、JWT + Argon2 认证、位图 RBAC 权限。

## Setup commands

```bash
# 安装依赖
go mod tidy

# 生成 Swagger 文档
swag init -g cmd/server/main.go -o docs

# 开发模式运行（加载 .env.dev）
make dev

# 生产模式运行（加载 .env.prod）
make prod

# 编译
make build
```

## Testing instructions

```bash
# 运行所有测试
go test -v ./...

# 运行单个包测试
go test -v ./internal/model/...

# 运行 lint
golangci-lint run
```

- 修改代码后务必确保 `go build ./...` 编译通过
- 修改 handler/swagger 注释后需重新运行 `swag init -g cmd/server/main.go -o docs`

## Code style

- 使用标准 Go 代码规范（gofmt）
- 错误处理：使用 `pkg/apperrors` 包装，传递 i18n 错误码
- 所有对外 API 使用 `uid`（22 字符 URL-safe base64）作为资源标识，不暴露数字自增 ID
- Model 层同时承载数据模型和 DTO（Request/Response）
- Repository 层只做数据访问，不含业务逻辑
- Service 层承载业务逻辑
- Handler 层做参数绑定、校验，调用 service，返回统一响应
- 统一响应格式通过 `pkg/response` 包
- 删除操作统一使用硬删除（`Unscoped().Delete`），模型中不保留 `DeletedAt` 字段

## Architecture

```
cmd/server/main.go          → 入口，初始化所有组件
internal/config/            → 配置加载（Viper + godotenv）
internal/container/         → DI 容器（sync.Once 懒加载）
internal/model/             → GORM 模型 + Request/Response DTO
internal/repository/        → 数据访问层（接口 + 实现）
internal/service/           → 业务逻辑层（接口 + 实现）
internal/handler/           → HTTP 处理器（Gin handler）
internal/router/            → 路由注册（按模块分文件）
internal/middleware/        → Gin 中间件
internal/ws/                → WebSocket Hub
internal/seed/              → 权限/管理员种子数据
pkg/                        → 可复用的基础包
```

## Key conventions

- 配置优先级：环境变量 > .env 文件 > config.yaml
- `APP_ENV` 决定加载 `.env.dev` 或 `.env.prod`
- Redis 不可用时自动降级为内存缓存
- 权限系统使用位图（bitwise）实现，每个权限空间最多 64 个权限
- 路由注册时通过中间件自动收集权限码，启动时自动 seed 到数据库
- 验证码通过 cache backend（Redis/内存）存储，60 秒有效，60 秒内不可重发
- 邮件验证码通过 SMTP 发送，支持 `MAIL_MOCK_SEND=true` 跳过实际发送（开发模式）

### 分页接口规范

所有列表接口统一使用 `pkg/response.Pagination` 结构：

**请求参数**（Query）：

| 参数 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `page` | int | 页码，从 1 开始 | `1` |
| `page_size` | int | 每页数量，最大 100；传 `-1` 获取全部 | `10` |
| `sort` | string | 排序，格式 `field,asc` 或 `field,desc` | `id desc` |

**响应结构**：

```json
{
  "code": 0,
  "data": {
    "list": [],
    "page": 1,
    "page_size": 10,
    "total": 100,
    "total_pages": 10
  }
}
```

**Handler 使用方式**：

```go
func (h *XxxHandler) List(c *gin.Context) {
    p, ok := BindPagination(c)
    if !ok {
        return
    }

    items, total, err := h.service.List(ctx, p.GetOffset(), p.GetPageSize(), p.GetSort())
    if err != nil {
        c.Error(err)
        return
    }

    response.SuccessWithPage(c, items, total, p)
}
```

- `BindPagination(c)` 统一绑定并校验分页参数
- `response.SuccessWithPage` 自动计算 `total_pages` 并返回标准分页结构
- Repository 层接收 `offset, limit int, sort string`，不依赖 Pagination 结构

## Security considerations

- JWT secret 生产环境必须 ≥ 32 字符
- 密码使用 Argon2 哈希
- 注册/登录/重置密码均需验证码校验
- Token blacklist 需 Redis 支持
- 文档页有 Basic Auth 保护（`DOCS_USER` / `DOCS_PASSWORD`）
- 所有删除为硬删除，不可恢复

## Environment variables

配置优先级：环境变量 > `.env.dev` / `.env.prod` > `config/config.yaml`

`APP_ENV` 决定加载哪个 env 文件（`development` → `.env.dev`，`production` → `.env.prod`）。

### 应用 & 服务器

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `APP_ENV` | 运行环境 | `development` |
| `SERVER_PORT` | 监听端口 | `9527` |
| `JWT_SECRET` | JWT 密钥（生产 ≥ 32 字符） | — |
| `ACCESS_TOKEN_DAYS` | Access Token 有效天数 | `7` |
| `REFRESH_TOKEN_DAYS` | Refresh Token 有效天数 | `30` |
| `ADMIN_EMAIL` | 自动创建管理员邮箱（留空跳过） | — |
| `ADMIN_PASSWORD` | 管理员密码 | — |

### 数据库

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `DB_DRIVER` | `sqlite` / `mysql` / `postgres` | `sqlite` |
| `DB_PATH` | SQLite 文件路径 | `./data.db` |
| `DB_HOST` / `DB_PORT` | MySQL/PG 主机端口 | — |
| `DB_USER` / `DB_PASSWORD` / `DB_NAME` | MySQL/PG 凭证 | — |

### Redis

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `REDIS_ENABLED` | 是否启用 Redis | `false` |
| `REDIS_HOST` / `REDIS_PORT` | Redis 地址 | `localhost:6379` |
| `REDIS_PASSWORD` | Redis 密码 | — |
| `REDIS_DB` | 数据库编号 | `0` |
| `REDIS_ENABLE_FALLBACK` | Redis 不可用时降级内存 | `true` |

### 邮件验证码（SMTP）

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `MAIL_ENABLED` | 是否启用邮件 | `false` |
| `MAIL_HOST` | SMTP 服务器 | — |
| `MAIL_PORT` | 端口（465=SSL, 587=STARTTLS） | `587` |
| `MAIL_USER` | SMTP 账号 | — |
| `MAIL_PASS` | SMTP 密码 | — |
| `MAIL_FROM` | 发件人显示名 | — |
| `MAIL_MOCK_SEND` | `true` 跳过实际发送（开发用） | `false` |

### OSS（阿里云对象存储）

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `ALICLOUD_OSS_ENDPOINT` | OSS Endpoint | — |
| `ALICLOUD_OSS_BUCKET` | Bucket 名称 | — |
| `ALICLOUD_ACCESS_KEY_ID` | AccessKey ID | — |
| `ALICLOUD_ACCESS_KEY_SECRET` | AccessKey Secret | — |
| `ALICLOUD_OSS_UPLOAD_DIR` | 上传目录前缀 | `go_oss` |
| `OSS_DOMAIN` | 自定义 CDN 域名 | — |

### 微信小程序

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `WX_APPID` | 小程序 AppID | — |
| `WX_SECRET` | 小程序 AppSecret | — |
| `DEFAULT_USER_PASSWORD` | 微信注册用户默认密码 | `123456` |

### WebSocket

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `WS_KEY` | WebSocket 连接认证 Key | — |

### 文档保护

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `DOCS_USER` | Swagger 页面 Basic Auth 用户名 | `admin` |
| `DOCS_PASSWORD` | Swagger 页面 Basic Auth 密码 | `admin123` |

## Adding new modules (新增模块注意事项)

新增一个业务模块时，按以下步骤保持风格统一：

### 1. Model (`internal/model/`)

- 创建 `xxx.go`，包含 GORM 模型 + Request/Response DTO
- 不使用 `DeletedAt` 软删除字段
- 对外资源必须有 `UID` 字段（22 字符 base64），不暴露自增 ID
- 在 `model/registry.go` 的 `AllModels()` 中注册模型以启用 AutoMigrate

### 2. Repository (`internal/repository/`)

- 创建 `xxx_repository.go`，实现数据访问
- 在 `interfaces.go` 中定义对应接口
- 删除方法使用 `db.Unscoped().Delete()` 硬删除
- 错误变量命名：`ErrXxxNotFound`

### 3. Service (`internal/service/`)

- 创建 `xxx_service.go`，实现业务逻辑
- 在 `interfaces.go` 中定义对应接口
- 使用 `pkg/apperrors` 包装错误，传递 `pkg/i18n` 中的错误码
- 新增错误码需同时更新 `pkg/i18n/codes.go`、`zh_cn.go`、`en_us.go`

### 4. Handler (`internal/handler/`)

- 创建 `xxx_handler.go`
- 参数绑定用 `c.ShouldBindJSON` / `c.ShouldBindQuery`
- 错误统一通过 `c.Error()` 交给全局 error handler
- 成功响应用 `response.Success` / `response.Created`
- 添加 Swagger 注释（`// @Summary` 等）

### 5. Router (`internal/router/`)

- 创建 `xxx_routes.go`，定义路由注册函数 `registerXxxRoutes`
- 在 `router.go` 中调用该函数
- 需要权限控制的路由使用 `permMw.RequirePermission("CODE")` 中间件

### 6. Container (`internal/container/`)

- 在 `container.go` 中添加 `sync.Once` 字段
- 在 `container_repositories.go` 添加 Repository getter
- 在 `container_handlers.go` 添加 Service / Handler getter

### 7. 配置（如需要）

- 在 `internal/config/config.go` 添加配置结构体和 `viper.BindEnv`
- 在 `.env.example` 中添加对应变量
- 生产强制校验在 `internal/config/validation.go` 中添加

### 8. 验证

- `go build ./...` 编译通过
- `swag init -g cmd/server/main.go -o docs` 重新生成文档
- 更新此 AGENTS.md 的 Environment variables 部分（如有新增环境变量）
