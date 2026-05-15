# 🚀 Go API Starter

<p align="center">
  <strong>一个生产就绪的 Go RESTful API 启动模板</strong>
</p>

<p align="center">
  <a href="#特性">特性</a> •
  <a href="#快速开始">快速开始</a> •
  <a href="#项目结构">项目结构</a> •
  <a href="#api-文档">API 文档</a> •
  <a href="#配置说明">配置说明</a>
</p>

---

## ✨ 特性

- 🏗️ **清晰的项目结构** — 清晰分层：model / repository / service / handler / router，配套 DI container
- 🔥 **Gin** — 高性能 HTTP Web 框架
- 📦 **GORM** — ORM，默认 SQLite（零配置），可切换 MySQL
- 📝 **Swagger** — 自动生成的 OpenAPI 文档 + Scalar UI
- ⚙️ **Viper + godotenv** — 多环境配置（yaml + `.env.dev` / `.env.prod`）
- 📊 **Zap** — 结构化日志
- 🔍 **Request ID** — 请求追踪
- 🛡️ **CORS / Gzip / pprof** — 开箱即用
- ⏱️ **多级限流** — 单机 token bucket + Redis 分布式滑动窗口
- 🎯 **Graceful Shutdown** — 优雅停机
- 💊 **Health Checks** — `/health` + `/health/ready`
- 🔐 **JWT + Argon2** — access / refresh token 双令牌
- 🗝️ **RBAC** — 权限空间 + 位图权限 + 角色体系，含路由级权限收集
- 🚫 **Token Blacklist** — 登出 / 批量失效（需 Redis）
- 🔴 **Redis + 内存降级** — Redis 不可用时自动回退到内存缓存
- ☁️ **OSS 文件管理** — 直传 token、分片上传、秒传（MD5）

## 🛠️ 技术栈

| 组件 | 技术 |
|------|------|
| Web 框架 | [Gin](https://github.com/gin-gonic/gin) |
| ORM | [GORM](https://gorm.io/) |
| 数据库 | SQLite / MySQL |
| 配置 | [Viper](https://github.com/spf13/viper) + [godotenv](https://github.com/joho/godotenv) |
| 日志 | [Zap](https://github.com/uber-go/zap) |
| API 文档 | [swag](https://github.com/swaggo/swag) + [gin-swagger](https://github.com/swaggo/gin-swagger) |
| 对象存储 | [Aliyun OSS](https://github.com/aliyun/aliyun-oss-go-sdk) |
| 限流 | [golang.org/x/time](https://pkg.go.dev/golang.org/x/time) + Redis 滑动窗口 |
| 缓存 | [go-redis](https://github.com/redis/go-redis) |
| 验证器 | [validator](https://github.com/go-playground/validator) |

## 🚀 快速开始

### 环境要求

- Go 1.21+
- 可选：MySQL 8.0+（默认使用 SQLite）
- 可选：Redis 6+（默认使用内存缓存）

### 安装运行

```bash
# 克隆项目
git clone https://github.com/yourname/go-api-starter.git
cd go-api-starter

# 安装依赖
go mod tidy

# 复制环境变量配置文件
cp .env.example .env.dev  # Linux / macOS
# copy .env.example .env.dev   # Windows

# 开发模式运行
make dev
```

### Makefile 常用命令

```bash
make build      # 生成 swagger 并编译
make dev        # 开发模式（加载 .env.dev）
make prod       # 生产模式（加载 .env.prod）
make test       # 跑测试
make swagger    # 重新生成 swagger 文档
make clean      # 清理构建产物
make fmt        # 格式化
make lint       # golangci-lint
```

### 启动成功

```
+-----------------------------------------------------------+
|  [*] go-api-starter started successfully!                 |
+-----------------------------------------------------------+
|  > Environment:  development                              |
+-----------------------------------------------------------+
|  > Local:        http://localhost:9527                    |
|  > Network:      http://192.168.x.x:9527                  |
+-----------------------------------------------------------+
|  > API Base:     http://localhost:9527/api/v1             |
|  > API Docs:     http://localhost:9527/docs               |
|  > Swagger:      http://localhost:9527/swagger/index.html |
+-----------------------------------------------------------+
```

## 📁 项目结构

```
go-api-starter/
├── cmd/server/                 # 应用入口
├── config/config.yaml          # 主配置（可被 env 覆盖）
├── docs/                       # Swagger 自动生成
├── internal/
│   ├── config/                 # 配置加载
│   ├── container/              # DI 容器
│   ├── handler/                # HTTP 处理器
│   ├── middleware/             # Gin 中间件
│   ├── model/                  # 数据模型 + DTO
│   ├── repository/             # 数据访问层
│   ├── router/                 # 路由注册（按模块）
│   ├── seed/                   # 权限/管理员种子
│   └── service/                # 业务逻辑
├── pkg/
│   ├── apperrors/              # 应用错误
│   ├── auth/                   # JWT / Argon2
│   ├── banner/                 # 启动横幅
│   ├── cache/                  # Redis / 内存 / Fallback
│   ├── database/               # GORM 初始化
│   ├── i18n/                   # 错误码字典
│   ├── logger/                 # Zap 封装
│   ├── migration/              # AutoMigrate 封装
│   ├── netutil/                # 本机 IP
│   ├── oss/                    # OSS 客户端 + 分片签名
│   ├── response/               # 统一响应 + 分页
│   └── utils/                  # 通用工具
├── migrations/                 # 手写 SQL 迁移（预留）
├── .env.example
├── go.mod
├── Makefile
└── README.md
```

## 📖 API 文档

启动服务后访问：

| 地址 | 说明 |
|------|------|
| http://localhost:9527/docs | Scalar UI |
| http://localhost:9527/swagger/index.html | Swagger UI |
| http://localhost:9527/swagger/doc.json | OpenAPI JSON |
| http://localhost:9527/llms.txt | LLMs.txt（AI 可读接口概览） |
| http://localhost:9527/llms-full.txt | LLMs-full.txt（AI 可读完整文档） |

文档接口由 `DOCS_USER` / `DOCS_PASSWORD` 做 Basic Auth 保护。

### LLMs.txt

项目内置了 [llms.txt](https://llmstxt.org/) 支持，服务启动时根据 Swagger spec 自动生成，无需手动维护。

- `/llms.txt` — 轻量入口，AI 快速了解接口概览
- `/llms-full.txt` — 完整 Markdown 文档，包含参数、响应示例，AI 可据此调用接口

## 🔌 API 端点

### 基础

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | 健康检查 |
| `GET` | `/health/ready` | 就绪检查 |

### 认证

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/auth/register` | 注册 |
| `POST` | `/api/v1/auth/login` | 登录 |
| `POST` | `/api/v1/auth/refresh` | 刷新访问令牌 |
| `POST` | `/api/v1/auth/reset-password/:id` | 管理员重置密码 |
| `POST` | `/api/v1/auth/logout` | 登出（需 Redis） |
| `POST` | `/api/v1/auth/logout-all` | 登出所有设备（需 Redis） |

### 用户

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/users/me` | 当前用户信息 |
| `PUT` | `/api/v1/users/me` | 更新当前用户 |
| `GET` | `/api/v1/users/:sec_uid` | 查看用户 |
| `POST` | `/api/v1/users` | 创建（需权限） |
| `GET` | `/api/v1/users` | 列表（需权限） |
| `PUT` | `/api/v1/users/:sec_uid` | 更新（需权限） |
| `DELETE` | `/api/v1/users/:sec_uid` | 删除（需权限） |

### 权限（RBAC）

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` / `POST` | `/api/v1/permissions/spaces` | 权限空间 |
| `GET` / `POST` | `/api/v1/permissions/permissions` | 权限 |
| `GET` / `POST` | `/api/v1/permissions/roles` | 角色 |
| `POST` | `/api/v1/permissions/roles/:id/permissions` | 为角色分配权限 |
| `POST` | `/api/v1/permissions/users/:sec_uid/roles` | 为用户分配角色 |
| `GET` | `/api/v1/permissions/me/permissions` | 我的权限 |

### 文件 / OSS

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/file/public/upload` | 公开上传（无需鉴权） |
| `POST` | `/api/v1/file/upload/init` | 初始化上传（秒传 / 普通 / 分片自动判断） |
| `POST` | `/api/v1/file/upload/urls` | 分片上传签名 |
| `POST` | `/api/v1/file/upload/complete` | 完成上传并落库 |
| `POST` | `/api/v1/file/upload/abort` | 中止分片上传 |
| `GET` | `/api/v1/file` | 文件列表 |
| `GET` | `/api/v1/file/:sec_uid` | 文件详情 |
| `PUT` | `/api/v1/file/:sec_uid` | 更新（名称 / 可见性） |
| `DELETE` | `/api/v1/file/:sec_uid` | 删除 |

### 上传流程说明

1. 前端计算文件 MD5，调用 `/file/upload/init`
   - 已存在则直接返回文件（秒传）
   - 小于 5MB 返回 `simple` 模式 + OSS 直传 token
   - 大于等于 5MB 返回 `multipart` 模式 + uploadID/key/host 等
2. 按模式上传：
   - `simple`：直接用 token POST 到 OSS
   - `multipart`：调用 `/file/upload/urls` 拿分片签名，PUT 到 OSS；需要续传时同 uploadID 再次调用 `init` 即可拿到已上传分片
3. 调用 `/file/upload/complete` 完成（普通上传传 key+md5，分片上传额外传 upload_id + parts）

## ⚙️ 配置说明

配置文件 `config/config.yaml`，环境变量优先级最高。支持 `.env.dev` / `.env.prod`，通过 `APP_ENV` 决定加载哪个。

### 常用环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `APP_ENV` | 环境 | `development` |
| `SERVER_PORT` | 端口 | `9527` |
| `DB_DRIVER` | `sqlite` / `mysql` / `postgres` | `sqlite` |
| `DB_PATH` | SQLite 路径 | `./data.db` |
| `DB_HOST` / `DB_PORT` / `DB_USER` / `DB_PASSWORD` / `DB_NAME` | MySQL/PG 连接 | — |
| `JWT_SECRET` | JWT 密钥（生产必须改） | — |
| `ADMIN_EMAIL` / `ADMIN_PASSWORD` | 自动创建管理员账号 | — |
| `DOCS_USER` / `DOCS_PASSWORD` | Swagger 页面 Basic Auth | `admin` / `admin123` |
| `REDIS_ENABLED` | 是否启用 Redis | `false` |
| `REDIS_HOST` / `REDIS_PORT` / `REDIS_PASSWORD` / `REDIS_DB` | Redis 连接 | `localhost:6379` |
| `ALICLOUD_OSS_ENDPOINT` | OSS endpoint | — |
| `ALICLOUD_OSS_BUCKET` | OSS bucket | — |
| `ALICLOUD_ACCESS_KEY_ID` | OSS AccessKey ID | — |
| `ALICLOUD_ACCESS_KEY_SECRET` | OSS AccessKey Secret | — |
| `ALICLOUD_OSS_UPLOAD_DIR` | 上传目录前缀 | `go_oss` |
| `OSS_DOMAIN` | 自定义 CDN 域名 | — |

### 生产环境强制校验

`APP_ENV=production` 时，以下配置必须满足：

- `JWT_SECRET` 非空、非默认值、至少 32 字符
- 非 SQLite 时：数据库密码和主机必须显式配置
- OSS endpoint 配置后：AccessKey / Bucket 必须配置

## 📜 License

MIT License
