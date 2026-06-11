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
- 📝 Swagger + Scalar UI + LLMs.txt

## 🚀 快速开始

```bash
git clone https://github.com/IceyWu/go-api-starter
cd go-api-starter

go mod tidy
cp .env.example .env.dev

make dev
```

## 📖 文档

| 地址 | 说明 |
|------|------|
| `/docs` | Scalar API 文档 |
| `/swagger/doc.json` | OpenAPI JSON |
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

配置优先级：环境变量 > `.env.dev` / `.env.prod` > `config/config.yaml`

通过 `APP_ENV` 决定加载哪个 env 文件。完整变量及说明见 [AGENTS.md](./AGENTS.md#environment-variables)，快速示例见 [.env.example](.env.example)。

## 🤖 AI Agents

项目包含 [AGENTS.md](./AGENTS.md)，为 AI 编码代理提供构建命令、代码规范和架构上下文。

## 📜 License

MIT
