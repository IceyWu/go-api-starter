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

- 🏗️ **清晰的项目结构** - 遵循 Go 社区最佳实践
- 🔥 **Gin 框架** - 高性能 HTTP Web 框架
- 📦 **GORM** - 强大的 ORM 库
- 📝 **Swagger/Scalar** - 美观的 API 文档界面
- ⚙️ **Viper** - 灵活的配置管理
- 🌍 **Godotenv** - 环境变量管理
- 📊 **Zap** - 高性能结构化日志
- 🔍 **Request ID** - 请求追踪支持
- 🛡️ **CORS** - 跨域资源共享支持
- ⏱️ **Rate Limiting** - API 限流保护
- 🎯 **Graceful Shutdown** - 优雅关闭支持
- 🔄 **Context Propagation** - 完整的上下文传递
- ❌ **Enhanced Error Handling** - 统一的错误处理机制
- 💊 **Health Checks** - 健康检查和就绪检查端点
- 🗜️ **Gzip Compression** - 响应压缩支持
- 📈 **Performance Monitoring** - pprof 性能分析
- 🎯 **完整 CRUD 示例** - 开箱即用的用户管理模块
- 🔐 **权限管理系统** - 基于 RBAC 的权限控制
- ☁️ **阿里云 OSS 集成** - 文件上传与管理
- 🔧 **多数据库支持** - SQLite / MySQL

## 🛠️ 技术栈

| 组件 | 技术 |
|------|------|
| Web 框架 | [Gin](https://github.com/gin-gonic/gin) |
| ORM | [GORM](https://gorm.io/) |
| 数据库 | SQLite / MySQL |
| 配置管理 | [Viper](https://github.com/spf13/viper) |
| 环境变量 | [Godotenv](https://github.com/joho/godotenv) |
| 日志 | [Zap](https://github.com/uber-go/zap) |
| API 文档 | [Swag](https://github.com/swaggo/swag) + [Scalar](https://github.com/scalar/scalar) |
| 对象存储 | [阿里云 OSS](https://github.com/aliyun/aliyun-oss-go-sdk) |
| CORS | [gin-contrib/cors](https://github.com/gin-contrib/cors) |
| 请求追踪 | [gin-contrib/requestid](https://github.com/gin-contrib/requestid) |
| 响应压缩 | [gin-contrib/gzip](https://github.com/gin-contrib/gzip) |
| 性能分析 | [gin-contrib/pprof](https://github.com/gin-contrib/pprof) |
| 限流 | [golang.org/x/time/rate](https://pkg.go.dev/golang.org/x/time/rate) |
| 验证器 | [Validator](https://github.com/go-playground/validator) |

## 🚀 快速开始

### 环境要求

- Go 1.21+
- MySQL 8.0+ (可选，默认使用 SQLite)

### 安装运行

```bash
# 克隆项目
git clone https://github.com/yourname/go-api-starter.git
cd go-api-starter

# 安装依赖
go mod tidy

# 复制环境变量配置文件
copy .env.example .env

# 开发模式运行 (使用 MySQL)
make dev

# 或直接运行 (使用 SQLite)
go run ./cmd/server
```

### Makefile 命令

```bash
make dev      # 开发模式运行
make build    # 编译项目
make swagger  # 生成 Swagger 文档
make clean    # 清理编译产物
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
|  > OpenAPI:      http://localhost:9527/swagger/doc.json   |
+-----------------------------------------------------------+
```

## 📁 项目结构

```
go-api-starter/
├── cmd/
│   └── server/
│       └── main.go              # 应用入口
├── config/
│   └── config.yaml              # 配置文件
├── docs/                        # Swagger 文档 (自动生成)
├── internal/
│   ├── config/                  # 配置加载
│   ├── handler/                 # HTTP 处理器
│   ├── middleware/              # 中间件
│   ├── model/                   # 数据模型
│   ├── repository/              # 数据访问层
│   ├── router/                  # 路由配置
│   └── service/                 # 业务逻辑层
├── pkg/
│   ├── apperrors/               # 应用错误定义
│   ├── banner/                  # 启动横幅
│   ├── database/                # 数据库连接
│   ├── errors/                  # 错误定义
│   ├── logger/                  # 日志工具
│   ├── oss/                     # OSS 客户端
│   ├── response/                # 统一响应
│   └── utils/                   # 工具函数
├── playground/
│   └── shadcn-admin/            # 前端管理后台
├── .env.example
├── go.mod
├── Makefile
└── README.md
```

## 📖 API 文档

启动服务后访问：

| 地址 | 说明 |
|------|------|
| http://localhost:9527/docs | Scalar UI (推荐) |
| http://localhost:9527/swagger/index.html | Swagger UI |
| http://localhost:9527/swagger/doc.json | OpenAPI JSON |

## 🔌 API 端点

### 基础端点

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | 健康检查 |
| `GET` | `/health/ready` | 就绪检查 |

### 用户管理

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/users` | 创建用户 |
| `GET` | `/api/v1/users` | 获取用户列表 |
| `GET` | `/api/v1/users/:id` | 获取单个用户 |
| `PUT` | `/api/v1/users/:id` | 更新用户 |
| `DELETE` | `/api/v1/users/:id` | 删除用户 |

### 权限管理

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/permissions/spaces` | 获取权限空间列表 |
| `POST` | `/api/v1/permissions/spaces` | 创建权限空间 |
| `GET` | `/api/v1/permissions` | 获取权限列表 |
| `POST` | `/api/v1/permissions` | 创建权限 |
| `GET` | `/api/v1/permissions/roles` | 获取角色列表 |
| `POST` | `/api/v1/permissions/roles` | 创建角色 |
| `POST` | `/api/v1/permissions/roles/:id/permissions` | 为角色分配权限 |
| `POST` | `/api/v1/permissions/users/:id/roles` | 为用户分配角色 |
| `GET` | `/api/v1/permissions/me/permissions` | 获取当前用户权限 |

### OSS 文件管理

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/oss/token` | 获取上传令牌 |
| `POST` | `/api/v1/oss/callback` | 上传回调 |
| `GET` | `/api/v1/oss/files` | 获取文件列表 |
| `DELETE` | `/api/v1/oss/files/:id` | 删除文件 |

## ⚙️ 配置说明

配置文件位于 `config/config.yaml`，支持环境变量覆盖：

```yaml
app:
  name: go-api-starter
  env: development

server:
  host: localhost
  port: 9527
  mode: debug

database:
  driver: mysql  # sqlite, mysql
  # SQLite
  path: ./data.db
  # MySQL
  host: localhost
  port: 3306
  username: root
  password: "123456"
  dbname: go_api_starter
  charset: utf8mb4

log:
  level: debug
  format: console

# OSS 配置
oss:
  endpoint: your-bucket.oss-accelerate.aliyuncs.com
  bucket: your-bucket
  access_key_id: ""      # 通过环境变量设置
  access_key_secret: ""  # 通过环境变量设置
  upload_dir: go_oss
  base_path: uploads
  domain: ""             # 自定义域名 (可选)
  max_file_size: 10485760
  token_expire: 1800
```

### OSS 配置说明

| 配置项 | 说明 | 示例 |
|--------|------|------|
| `endpoint` | OSS 访问域名 | `bucket.oss-cn-hangzhou.aliyuncs.com` |
| `bucket` | 存储桶名称 | `my-bucket` |
| `upload_dir` | 上传目录前缀 | `go_oss` |
| `base_path` | 基础路径 | `uploads` |
| `domain` | 自定义 CDN 域名 | `https://cdn.example.com` |
| `max_file_size` | 最大文件大小 (字节) | `10485760` (10MB) |
| `token_expire` | 令牌过期时间 (秒) | `1800` (30分钟) |

**文件存储路径**: `{upload_dir}/{base_path}/{date}/{uuid}.{ext}`  
**示例**: `go_oss/uploads/2026-01-15/abc123.jpg`

**URL 生成规则**:
- 设置 `domain` → `https://cdn.example.com/go_oss/uploads/2026-01-15/abc123.jpg`
- 未设置 `domain` → `https://{endpoint}/go_oss/uploads/2026-01-15/abc123.jpg`

### 环境变量

支持通过 `.env` 文件或系统环境变量配置：

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `APP_ENV` | 应用环境 | development |
| `SERVER_PORT` | 服务端口 | 9527 |
| `SERVER_MODE` | 运行模式 | debug |
| `DB_DRIVER` | 数据库类型 | sqlite |
| `DB_HOST` | MySQL 主机 | localhost |
| `DB_PORT` | MySQL 端口 | 3306 |
| `DB_USER` | MySQL 用户名 | root |
| `DB_PASSWORD` | MySQL 密码 | 123456 |
| `DB_NAME` | MySQL 数据库名 | go_api_starter |
| `OSS_ACCESS_KEY_ID` | OSS AccessKey ID | - |
| `OSS_ACCESS_KEY_SECRET` | OSS AccessKey Secret | - |
| `LOG_LEVEL` | 日志级别 | debug |

## 🖥️ 前端管理后台

项目包含一个基于 React + shadcn/ui 的管理后台：

```bash
cd playground/shadcn-admin
pnpm install
pnpm dev
```

访问 http://localhost:5173

功能包括：
- 用户管理
- 权限管理 (权限空间、权限、角色)
- 文件管理 (OSS 上传、列表、删除)
- 中英文国际化
- 深色/浅色主题

## 📜 License

MIT License
