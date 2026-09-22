<p align="center">
  <img src="https://raw.githubusercontent.com/IceyWu/go-api-starter/main/public/logo.svg" alt="go-api-starter logo" width="96" height="96" />
</p>

<h1 align="center">go-api-starter</h1>

<p align="center">
  面向生产的 Go API 脚手架，提供 OpenAPI、RBAC、文件上传、MPS 视频转码和工程化工具链。
</p>

<p align="center">
  简体中文 · <a href="README.md">English</a>
</p>
<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/github/license/IceyWu/go-api-starter" alt="License" /></a>
  <a href="https://github.com/IceyWu/go-api-starter/actions/workflows/ci.yml"><img src="https://img.shields.io/github/actions/workflow/status/IceyWu/go-api-starter/ci.yml?label=CI" alt="CI status" /></a>
  <a href="go.mod"><img src="https://img.shields.io/github/go-mod/go-version/IceyWu/go-api-starter" alt="Go version" /></a>
</p>

`go-api-starter` 基于 Chi、Huma、sqlc、Atlas 和 Task 构建，覆盖认证、用户、权限、文件上传、WebSocket 消息以及阿里云 MPS 异步视频转码。

## 功能特性

- 清晰分层：model、repository、service、handler、router 和 DI container。
- JWT access/refresh 双 Token 认证和 Token 黑名单。
- 权限空间、CRUD 权限、角色以及路由级权限控制。
- S3 兼容对象存储直传、分片上传、断点续传、秒传和文件记录持久化。
- 可选的独立 MPS Worker，负责异步视频转码、轮询和结果落库。
- 支持 API Key、心跳、指令和 ack 的 WebSocket Hub。
- 微信小程序登录、Redis 分布式限流和内存降级。
- 健康检查、Prometheus Metrics、Scalar API 文档、OpenAPI JSON 和 `llms.txt`。

## 快速开始

```bash
git clone https://github.com/IceyWu/go-api-starter.git
cd go-api-starter

# 安装固定版本的 Task。
go install github.com/go-task/task/v3/cmd/task@v3.53.1

task deps
task dev
```

项目命令统一由 [Task](https://taskfile.dev/) 管理：

```bash
task --list
task test
task check
task security
task migrate
task worker
```

## Docker Compose

Docker Compose 会启动 MySQL、Redis 和 API，并在 API 提供服务前自动执行 Atlas 迁移：

```bash
docker compose up --build api
```

API 就绪检查地址为 `http://localhost:8080/health/ready`。需要启动 MPS Worker 时：

```bash
docker compose --profile worker up --build
```

Compose 中的账号和密钥仅用于本地开发，部署到其他环境前必须全部替换。

## 项目架构

```text
HTTP 客户端
    -> Chi transport 和中间件
    -> Huma / 传统路由适配层
    -> handler
    -> service
    -> repository 和 platform 适配层
    -> MySQL 或 SQLite / Redis / S3 兼容对象存储 / MPS

MPS Worker
    -> 任务管理器
    -> 阿里云 MPS
    -> 轮询和结果持久化
    -> 可选 Webhook 通知
```

API 进程不执行本地视频转码。启用 MPS 时，API 负责创建和跟踪任务，独立 Worker 轮询 MPS 并写入最终视频变体。

## 文档和接口

| 地址 | 说明 |
| --- | --- |
| `/docs` | Scalar 接口文档（Basic Auth） |
| `/openapi.json` | 完整 OpenAPI 文档（Basic Auth） |
| `/swagger/doc.json` | 兼容旧客户端的 OpenAPI 地址 |
| `/llms.txt` | AI 可读接口概览 |
| `/llms-full.txt` | AI 可读完整接口文档 |
| `/health` | 存活检查 |
| `/health/ready` | 数据库和缓存就绪检查 |
| `/metrics` | Prometheus 指标 |
| `/ws` | WebSocket 入口 |

文件上传使用与存储厂商无关的上传会话流程：

```text
POST   /api/v1/uploads
POST   /api/v1/uploads/{id}/complete
DELETE /api/v1/uploads/{id}
```

客户端只接收预签名 URL 并提交分片 ETag；对象 Key、大小和类型由服务端生成并校验。

开发环境默认端口为 `9527`，生产环境默认为 `8080`。

## 配置

唯一配置源是 [`config/config.yaml`](config/config.yaml)，通过 `APP_ENV=development` 或 `APP_ENV=production` 选择环境配置。

环境变量使用 `GO_API_` 前缀覆盖最终配置，双下划线表示嵌套层级：

```text
GO_API_SERVER__PORT=9000
GO_API_DATABASE__PASSWORD=replace-me
GO_API_APP__JWT_SECRET=replace-with-at-least-32-characters
```

完整的配置文件职责、加载优先级和环境变量规则见[中文配置说明](docs/configuration.zh-CN.md)。

本地开发时，`task dev` 会加载 `.env.dev` 和可选的、被 Git 忽略的 `.env.dev.local`；对象存储凭据应放在后者。独立上传 Demo 可通过 `task upload-demo` 启动，地址为 `http://127.0.0.1:5501/upload-demo.html`。

## 开发与验证

```bash
task fmt
task sqlc
task test
task test-integration
task check
task atlas-validate
task security
task build-linux-all
```

如果配置了 `MYSQL_TEST_DSN`，集成测试会连接真实 MySQL。GitHub Actions 会自动启动 MySQL、执行迁移并运行集成测试。

## 项目结构

```text
cmd/                 server、worker 和迁移入口
config/              通用配置和环境配置
db/                  SQL schema 和 sqlc 查询
docs/                纳入版本控制的 OpenAPI 文档
internal/            应用代码和平台适配层
migrations/          SQLite 和 MySQL 的 Atlas 迁移
public/              Logo 和静态资源
Taskfile.yml         项目命令约束
Dockerfile           生产容器构建
docker-compose.yml   API、MySQL、Redis 和 Worker 编排
```

## 许可证

MIT
