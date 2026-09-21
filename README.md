# 🚀 Go API Starter

一个生产就绪的 Go RESTful API 脚手架。

## ✨ 特性

- 🏗️ 清晰分层（model / repository / service / handler / router）+ DI container
- 🔐 JWT + Argon2 双令牌认证，Token Blacklist
- 🗝️ 位图 RBAC 权限体系（权限空间 + 角色 + 路由级自动收集）
- 📧 邮箱验证码（注册/登录/重置密码）
- ☁️ OSS 文件管理（直传、分片、秒传）
- 🌐 WebSocket Hub（心跳、指令/ack、认证）
- 💬 微信小程序一键登录
- ⏱️ 多级限流（单机 + Redis 分布式）
- 🔴 Redis + 内存自动降级
- 📝 Huma OpenAPI + Scalar UI + LLMs.txt

## 🚀 快速开始

```bash
git clone https://github.com/IceyWu/go-api-starter
cd go-api-starter

# 安装并使用 Task（项目命令统一由 Taskfile 管理）
go install github.com/go-task/task/v3/cmd/task@latest
task deps

task dev
```

常用校验命令：`task test`、`task check`、`task sqlc`、`task migrate`。

生产环境需要先安装 Atlas，执行迁移并单独运行 API 和 MPS Worker：

```bash
task migrate
task prod
task worker
```

CI 会自动执行 `task check`、`task security`、Atlas migration 校验和 Linux 构建。生产环境必须通过环境变量覆盖 JWT、文档账号、管理员密码、默认用户密码、CORS 来源及外部服务凭据。

## 📖 文档

| 地址 | 说明 |
|------|------|
| `/docs` | Scalar API 文档 |
| `/openapi.json` | Complete OpenAPI JSON (Basic Auth) |
| `/huma-openapi.json` | Huma generated operation document |
| `/llms.txt` | AI 可读接口概览 |
| `/llms-full.txt` | AI 可读完整文档 |
| `/ws` | WebSocket 入口 |

## 🔌 核心 API

| 模块 | 端点 | 说明 |
|------|------|------|
| 认证 | `POST /api/v1/auth/register` | 注册（需验证码） |
| | `POST /api/v1/auth/login` | 登录 |
| | `POST /api/v1/auth/wx-login` | 微信登录 |
| | `POST /api/v1/auth/self-reset-password` | 自助重置密码 |
| 验证码 | `POST /api/v1/verification/send` | 发送验证码 |
| | `POST /api/v1/verification/verify` | 校验验证码 |
| 用户 | `GET/PUT /api/v1/users/me` | 当前用户 |
| | `CRUD /api/v1/users/:uid` | 用户管理 |
| 权限 | `/api/v1/permissions/*` | 空间/权限/角色 CRUD |
| 文件 | `POST /api/v1/file/upload/init` | 上传初始化 |
| | `POST /api/v1/file/upload/complete` | 完成上传 |
| WebSocket | `GET /ws` | 长连接 |

## ⚙️ 配置

配置统一维护在 [`config/config.yaml`](config/config.yaml) 中，`APP_ENV` 选择 `development` 或 `production` 配置段。

环境变量使用 `GO_API_` 前缀覆盖配置，双下划线表示层级，例如 `GO_API_SERVER__PORT=9000` 覆盖 `server.port`。完整规则见 [AGENTS.md](./AGENTS.md#environment-variables)，示例见 [.env.example](.env.example)。

视频转码统一提交到阿里云 MPS。API server 只负责创建任务，独立的 MPS Worker 负责轮询和结果落库；项目不再包含本地转码 worker。

## 🤖 AI Agents

项目包含 [AGENTS.md](./AGENTS.md)，为 AI 编码代理提供构建命令、代码规范和架构上下文。

## 📜 License

MIT
