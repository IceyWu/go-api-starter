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
- 📊 **Zap** - 高性能结构化日志
- 🎯 **完整 CRUD 示例** - 开箱即用的用户管理模块
- 🔧 **零外部依赖** - 使用 SQLite，无需安装数据库

## 🛠️ 技术栈

| 组件 | 技术 |
|------|------|
| Web 框架 | [Gin](https://github.com/gin-gonic/gin) |
| ORM | [GORM](https://gorm.io/) |
| 数据库 | SQLite / MySQL |
| 配置管理 | [Viper](https://github.com/spf13/viper) |
| 日志 | [Zap](https://github.com/uber-go/zap) |
| API 文档 | [Swag](https://github.com/swaggo/swag) + [Scalar](https://github.com/scalar/scalar) |
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
| `POST` | `/api/v1/users` | 创建用户 |
| `GET` | `/api/v1/users` | 获取用户列表 |
| `GET` | `/api/v1/users/:id` | 获取单个用户 |
| `PUT` | `/api/v1/users/:id` | 更新用户 |
| `DELETE` | `/api/v1/users/:id` | 删除用户 |

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

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `SERVER_PORT` | 服务端口 | 9527 |
| `SERVER_MODE` | 运行模式 | debug |
| `DATABASE_DRIVER` | 数据库类型 | sqlite |
| `DATABASE_PATH` | SQLite 路径 | ./data.db |
| `DATABASE_HOST` | MySQL 主机 | localhost |
| `DATABASE_PORT` | MySQL 端口 | 3306 |
| `DATABASE_USERNAME` | MySQL 用户名 | root |
| `DATABASE_PASSWORD` | MySQL 密码 | 123456 |
| `DATABASE_DBNAME` | MySQL 数据库名 | go_api_starter |
| `LOG_LEVEL` | 日志级别 | debug |

## 📜 License

MIT License
