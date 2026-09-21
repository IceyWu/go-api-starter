# AGENTS.md

## Project overview

Go API Starter — 基于 Go + Chi + Huma + sqlc + Atlas 的 RESTful API 脚手架。采用清晰分层架构：model → repository → service → handler → router，通过 DI container 管理依赖。

技术栈：Chi、Huma、sqlc、Atlas、slog 日志、Koanf 配置、Redis 缓存（可选）、阿里云 OSS、WebSocket、JWT + Argon2 认证、位图 RBAC 权限。

## Setup commands

```bash
# 安装依赖
# Taskfile 是项目命令的唯一入口
task deps

# 生成 SQL 类型代码
task sqlc

# 开发模式运行（APP_ENV=development）
task dev

# 生产模式运行（APP_ENV=production）
task prod

# 编译
task build
```

## Command policy

项目构建、测试、格式化、静态检查、代码生成、依赖整理、迁移和运行命令必须通过 `Taskfile.yml` 执行。代理和开发者不得在项目流程中直接执行 `go test`、`go build`、`go vet`、`go fmt`、`go mod tidy` 或 `go run`；这些命令只能作为 Taskfile 任务的内部实现。

常用入口：

- `task build`：编译 API server
- `task test`：运行全部测试
- `task check`：执行格式化、sqlc 生成、测试、编译、vet 和 diff 检查
- `task fmt` / `task vet`：单独执行格式化或静态检查
- `task sqlc`：生成类型安全 SQL 代码
- `task migrate`：执行 Atlas 数据库迁移
- `task atlas-validate`：校验 SQLite/MySQL Atlas migration
- `task security`：执行 Go 依赖漏洞扫描
- `task dev` / `task prod` / `task worker`：运行对应进程

新增项目命令时，先在 `Taskfile.yml` 中增加任务，再通过 `task <name>` 调用，不直接把 Go 命令作为文档或操作入口。

MPS 转码轮询使用独立进程：

```bash
task worker
```

## Testing instructions

```bash
# 运行所有测试
task test

# 运行单个包测试
task test-package PACKAGE=internal/model/...

# 运行 lint
task lint
```

- 单包测试使用 `task test-package PACKAGE=<package>`；项目完整验证必须使用 `task test` 或 `task check`
- 修改代码后务必确保 `task build` 或 `task check` 通过
- `/openapi.json` 是完整业务 OpenAPI 契约；`/huma-openapi.json` 仅用于查看当前已注册的 Huma typed operations
- 当前 HTTP 处理器通过 `internal/transport` 的 Chi transport helpers 运行，不再使用 Gin/Gin-compatible 命名

## Code style

- 使用标准 Go 代码规范（gofmt）
- 错误处理：使用 `internal/platform/apperrors` 包装，传递 i18n 错误码
- 所有对外 API 使用 `uid`（22 字符 URL-safe base64）作为资源标识，不暴露数字自增 ID
- Model 层同时承载数据模型和 DTO（Request/Response）
- Repository 层只做数据访问，不含业务逻辑
- Service 层承载业务逻辑
- Handler 层做参数绑定、校验，调用 service，返回统一响应
- 统一响应格式通过 `internal/platform/response` 包
- 删除策略由模型定义决定；需要审计恢复的模型使用 `DeletedAt`，其他模型使用硬删除

## Architecture

```
cmd/server/main.go          → 极薄启动入口
cmd/migrate/main.go         → 执行版本化数据库迁移
internal/app/               → 应用启动、生命周期和优雅关闭
internal/config/            → 配置加载（Koanf + YAML + GO_API_* 环境变量）
internal/container/         → DI 容器（sync.Once 懒加载）
internal/model/             → SQL 模型 + Request/Response DTO
internal/repository/        → 数据访问层（接口 + 实现）
internal/service/           → 业务逻辑层（接口 + 实现）
internal/handler/           → HTTP 处理器
internal/router/            → 路由注册（按模块分文件）
internal/middleware/        → HTTP 中间件
internal/ws/                → WebSocket Hub
internal/seed/              → 权限/管理员种子数据
internal/platform/          → 应用私有基础设施（数据库、缓存、OSS、日志、指标等）
internal/transport/         → Chi HTTP transport helpers and response writers
```

## Key conventions

- 配置优先级：`GO_API_*` 环境变量 > 当前环境配置段 > `common` 配置段
- `APP_ENV` 选择 `development` 或 `production` 配置段
- 双下划线表示配置层级：`GO_API_SERVER__PORT` → `server.port`
- Redis 不可用时自动降级为内存缓存
- 权限系统使用位图（bitwise）实现，每个权限空间最多 64 个权限
- 路由注册时通过中间件自动收集权限码，启动时自动 seed 到数据库
- 验证码通过 cache backend（Redis/内存）存储，60 秒有效，60 秒内不可重发
- 邮件验证码通过 SMTP 发送，支持 `MAIL_MOCK_SEND=true` 跳过实际发送（开发模式）

### 分页接口规范

所有列表接口统一使用 `internal/platform/response.Pagination` 结构：

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
func (h *XxxHandler) List(c *httpx.Context) {
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

配置文件为唯一的普通配置来源：`config/config.yaml`。

- `APP_ENV`：选择配置环境，支持 `development`、`dev`、`production`、`prod`，默认是 `development`。
- `CONFIG_FILE`：可选，指定配置文件路径，默认是 `./config/config.yaml`。
- `GO_API_*`：覆盖配置文件中的任意字段。
- 环境变量名使用双下划线表示嵌套层级，单下划线保留在字段名中。

示例：

```bash
APP_ENV=production
GO_API_SERVER__PORT=9000
GO_API_DATABASE__PASSWORD=change-me
GO_API_APP__JWT_SECRET=replace-with-at-least-32-characters
GO_API_OSS__ACCESS_KEY_SECRET=...
```

映射规则：

```text
GO_API_SERVER__PORT
  -> server.port

GO_API_DATABASE__PASSWORD
  -> database.password

```

新增配置时只需要：

1. 在 `config/config.yaml` 的 `common` 或环境段中添加配置。
2. 在 `internal/config/config.go` 的强类型结构体中添加字段。
3. 如需环境变量覆盖，直接使用 `GO_API_` + 路径转换后的变量名，不需要新增绑定代码。

## Adding new modules (新增模块注意事项)

新增一个业务模块时，按以下步骤保持风格统一：

### 1. Model (`internal/model/`)

- 创建 `xxx.go`，包含 SQL 模型 + Request/Response DTO
- 不使用 `DeletedAt` 软删除字段
- 对外资源必须有 `UID` 字段（22 字符 base64），不暴露自增 ID
- 数据库结构由 Atlas SQL migrations 和 `db/schema.sql` 管理，`sqlc generate` 同步类型安全查询代码

### 2. Repository (`internal/repository/`)

- 创建 `xxx_repository.go`，实现数据访问
- 在 `interfaces.go` 中定义对应接口
- 删除方法使用 `db.Unscoped().Delete()` 硬删除
- 错误变量命名：`ErrXxxNotFound`

### 3. Service (`internal/service/`)

- 创建 `xxx_service.go`，实现业务逻辑
- 在 `interfaces.go` 中定义对应接口
- 使用 `internal/platform/apperrors` 包装错误，传递 `internal/platform/i18n` 中的错误码
- 新增错误码需同时更新 `internal/platform/i18n/codes.go`、`zh_cn.go`、`en_us.go`

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

- 在 `config/config.yaml` 添加普通配置和环境差异
- 在 `internal/config/config.go` 添加对应强类型字段
- 如需环境变量覆盖，使用 `GO_API_` 前缀和双下划线层级规则，无需维护绑定清单
- 生产强制校验在 `internal/config/validation.go` 中添加

### 8. 验证

- `task check` 通过
- 更新此 AGENTS.md 的 Environment variables 部分（如有新增环境变量）
