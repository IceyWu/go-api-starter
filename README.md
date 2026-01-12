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
- 🔧 **零外部依赖** - 使用 SQLite，无需安装数据库

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
| CORS | [gin-contrib/cors](https://github.com/gin-contrib/cors) |
| 请求追踪 | [gin-contrib/requestid](https://github.com/gin-contrib/requestid) |
| 响应压缩 | [gin-contrib/gzip](https://github.com/gin-contrib/gzip) |
| 性能分析 | [gin-contrib/pprof](https://github.com/gin-contrib/pprof) |
| 限流 | [golang.org/x/time/rate](https://pkg.go.dev/golang.org/x/time/rate) |
| 验证器 | [Validator](https://github.com/go-playground/validator) |

## 🚀 快速开始

### 环境要求

- Go 1.21+

### 安装运行

```bash
# 克隆项目
git clone https://github.com/yourname/go-api-starter.git
cd go-api-starter

# 安装依赖
go mod tidy

# 复制环境变量配置文件
copy .env.example .env

# 生成 Swagger 文档
swag init -g cmd/server/main.go -o docs

# 运行
go run ./cmd/server

# 或编译后运行
go build -o server ./cmd/server
./server
```

### 启动成功

```
╔════════════════════════════════════════════════════════════╗
║  🚀 go-api-starter started successfully!                   ║
╠════════════════════════════════════════════════════════════╣
║  ➤ Environment:  development                               ║
╠════════════════════════════════════════════════════════════╣
║  ➤ Local:        http://localhost:9527                     ║
║  ➤ Network:      http://192.168.x.x:9527                   ║
╠════════════════════════════════════════════════════════════╣
║  ➤ API Docs:     http://localhost:9527/docs                ║
║  ➤ Swagger:      http://localhost:9527/swagger/index.html  ║
╚════════════════════════════════════════════════════════════╝
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
│   ├── banner/                  # 启动横幅
│   ├── database/                # 数据库连接
│   ├── errors/                  # 错误定义
│   ├── logger/                  # 日志工具
│   ├── response/                # 统一响应
│   └── utils/                   # 工具函数
├── .gitignore
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

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | 健康检查 |
| `GET` | `/health/ready` | 就绪检查 |
| `POST` | `/api/v1/users` | 创建用户 |
| `GET` | `/api/v1/users` | 获取用户列表 |
| `GET` | `/api/v1/users/:id` | 获取单个用户 |
| `PUT` | `/api/v1/users/:id` | 更新用户 |
| `DELETE` | `/api/v1/users/:id` | 删除用户 |
| `GET` | `/debug/pprof/` | 性能分析（开发环境） |

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
```

### 环境变量

支持通过 `.env` 文件或系统环境变量配置：

1. 复制 `.env.example` 为 `.env`
2. 根据需要修改配置
3. 设置 `APP_ENV` 来加载不同环境的配置：
   - `development` 或 `dev` → 加载 `.env.dev`
   - `production` 或 `prod` → 加载 `.env.prod`
   - 未设置 → 加载 `.env`

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `APP_ENV` | 应用环境 | development |
| `SERVER_PORT` | 服务端口 | 9527 |
| `SERVER_MODE` | 运行模式 | debug |
| `DB_DRIVER` | 数据库类型 | sqlite |
| `DB_PATH` | SQLite 路径 | ./data.db |
| `DB_HOST` | MySQL 主机 | localhost |
| `DB_PORT` | MySQL 端口 | 3306 |
| `DB_USER` | MySQL 用户名 | root |
| `DB_PASSWORD` | MySQL 密码 | 123456 |
| `DB_NAME` | MySQL 数据库名 | go_api_starter |
| `LOG_LEVEL` | 日志级别 | debug |

### 限流配置

在 `internal/router/router.go` 中调整限流参数：

```go
// 100 请求/秒，突发 200
rateLimiter := middleware.NewRateLimiter(rate.Limit(100), 200)
```

## 📜 License

MIT License
