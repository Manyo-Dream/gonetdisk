# GoNetDisk - AI 代理开发指南

> 本文档供 AI 编码代理阅读，用于快速理解项目结构和开发规范。

---

## 1. 项目概述

**GoNetDisk** 是一个基于 Go 语言开发的网络云盘/文件存储系统后端服务。项目采用分层架构设计，实现了用户管理、文件上传、文件哈希去重（秒传）等核心功能。

### 核心功能
- 用户注册/登录/信息管理（JWT 认证）
- 文件上传（支持单文件，当前仅支持 jpg 格式）
- 文件哈希去重（MD5 秒传基础）
- 物理文件与用户文件分离的双表存储设计

### 项目状态
当前处于早期开发阶段，用户模块基础功能已完成，文件上传功能部分实现。

---

## 2. 技术栈

| 类别 | 技术 | 版本 | 用途 |
|------|------|------|------|
| 语言 | Go | 1.25.1 | 开发语言 |
| Web 框架 | Gin | v1.11.0 | HTTP 服务、路由、中间件 |
| ORM | GORM | v1.31.1 | 数据库操作 |
| 数据库驱动 | MySQL Driver | v1.6.0 | MySQL 连接 |
| 配置管理 | Viper | v1.21.0 | YAML 配置读取 |
| 认证 | JWT (golang-jwt/jwt) | v5.3.0 | Token 生成与验证 |
| 密码加密 | bcrypt (golang.org/x/crypto) | v0.46.0 | 密码哈希存储 |

---

## 3. 项目结构

```
GoNetDisk/
├── cmd/
│   └── server/
│       └── main.go              # 应用程序入口
├── configs/
│   ├── config.yaml              # 业务配置（YAML 格式）
│   └── config.go                # 配置结构体定义和加载逻辑
├── internal/                    # 内部代码（不可被外部导入）
│   ├── controller/              # 控制器层 - HTTP 请求处理
│   │   ├── file_controller.go
│   │   └── user_controller.go
│   ├── service/                 # 业务逻辑层 - 核心业务处理
│   │   ├── file_service.go
│   │   └── user_service.go
│   ├── repository/              # 数据访问层 - 数据库操作
│   │   ├── file_repoe.go        # 注意：文件名拼写有误
│   │   └── user_repo.go
│   ├── model/                   # 数据模型层 - GORM 实体定义
│   │   ├── file.go
│   │   └── user.go
│   ├── dto/                     # 数据传输对象 - 请求/响应结构
│   │   ├── flie_request.go      # 注意：文件名拼写有误
│   │   ├── flie_response.go     # 注意：文件名拼写有误
│   │   ├── user_request.go
│   │   └── user_response.go
│   ├── middleware/              # 中间件
│   │   └── auth.go              # JWT 认证中间件
│   ├── router/                  # 路由定义
│   │   └── router.go
│   └── util/                    # 工具类
│       └── jwt.go               # JWT 管理器
├── pkg/                         # 公共包（可被外部导入）
│   └── database/
│       └── mysql.go             # 数据库连接初始化
├── docker/
│   ├── docker-compose.yaml      # MySQL 服务定义
│   ├── init/
│   │   └── init.sql             # 数据库初始化脚本
│   └── mysql_data/              # MySQL 数据持久化目录（gitignore）
├── storage/
│   └── uploads/                 # 文件上传存储目录（gitignore）
├── ai-docs/                     # AI 辅助开发文档
│   ├── current/                 # 当前开发文档
│   └── template/                # 文档模板
├── docs/                        # 项目文档
├── scripts/                     # 脚本文件（预留）
├── go.mod                       # Go 模块定义
└── go.sum                       # Go 依赖校验
```

---

## 4. 架构设计

### 4.1 分层架构

项目采用经典的分层架构，依赖关系自上而下：

```
HTTP 层 (Gin)
  ├── Router: 路由分发 (/api/v1/...)
  ├── Middleware: 认证、日志
  └── Controller: 请求处理、参数校验

业务逻辑层 (Service)
  └── 处理核心业务逻辑、事务管理、数据编排

数据访问层 (Repository)
  └── 数据库操作、查询封装

数据模型层 (Model)
  └── GORM 实体定义

数据库 (MySQL)
```

### 4.2 依赖注入流程

```
main.go
  ├── configs.LoadConfig()      # 加载配置文件
  ├── database.InitDB()         # 初始化数据库连接
  ├── util.NewJWTManager()      # 创建 JWT 管理器
  └── router.SetupRouter()      # 设置路由，完成依赖注入
```

在 `router.SetupRouter()` 中完成各层的组装：
```go
userRepo := repository.NewUserRepo(db)
userService := service.NewUserService(userRepo, jwtManager)
userController := controller.NewUserController(userService)
```

---

## 5. 数据模型

### 5.1 用户表 (user)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT UNSIGNED | 主键，自增 |
| username | VARCHAR(64) | 用户名，唯一索引 |
| email | VARCHAR(255) | 邮箱，唯一索引 |
| password_hash | VARCHAR(255) | 密码哈希（bcrypt） |
| avatar_url | VARCHAR(500) | 头像地址 |
| used_space | BIGINT UNSIGNED | 已用空间（字节） |
| total_space | BIGINT UNSIGNED | 总空间配额（默认 1GB） |
| status | INT | 状态：0正常，1禁用 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |
| deleted_at | DATETIME | 软删除时间（gorm.DeletedAt） |

### 5.2 物理文件表 (physical_file)

用于存储实际文件元数据，通过文件哈希实现去重：

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT UNSIGNED | 主键，自增 |
| file_hash | CHAR(64) | 文件唯一哈希（MD5） |
| file_name | VARCHAR(255) | 原始文件名 |
| file_ext | VARCHAR(20) | 文件扩展名 |
| file_size | BIGINT | 文件实际大小 |
| file_path | VARCHAR(500) | 物理存储路径 |
| storage_type | VARCHAR(20) | 存储方式：local, oss, s3 |
| ref_count | INT UNSIGNED | 引用计数 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

### 5.3 用户文件表 (user_file)

用户视角的文件目录结构：

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT UNSIGNED | 主键，自增 |
| user_id | BIGINT UNSIGNED | 所属用户ID |
| physical_id | BIGINT UNSIGNED | 关联物理文件ID（目录为NULL） |
| parent_id | BIGINT UNSIGNED | 父文件夹ID，0为根目录 |
| file_name | VARCHAR(255) | 用户显示的文件名 |
| file_ext | VARCHAR(20) | 扩展名 |
| path_stack | TEXT | 族谱路径，如 /0/1/5/ |
| is_dir | BOOLEAN | 是否为文件夹 |

---

## 6. API 接口

### 基础路径

所有 API 均以 `/api/v1` 为前缀。

### 用户接口

| 方法 | 路径 | 功能 | 认证 | 请求体 |
|------|------|------|------|--------|
| POST | /user/register | 用户注册 | 否 | `{"username": "...", "email": "...", "password": "..."}` |
| POST | /user/login | 用户登录 | 否 | `{"email": "...", "password": "..."}` |
| GET | /user/info | 获取用户信息 | 是 | `{"email": "..."}` |
| PUT | /user/info | 更新用户信息 | 是 | `{"username": "...", "email": "...", "avatar_url": "..."}` |

### 文件接口

| 方法 | 路径 | 功能 | 认证 | 请求体 |
|------|------|------|------|--------|
| POST | /file/upload | 文件上传 | 是 | multipart/form-data (`parent_id`, `file`) |

### 认证方式

使用 JWT Bearer Token：
```
Authorization: Bearer <token>
```

---

## 7. 配置说明

### 配置文件

配置文件位于 `configs/config.yaml`：

```yaml
server:
  port: 9090              # 服务端口
  host: "0.0.0.0"         # 监听地址
  mode: debug             # 运行模式：debug/release

database:
  type: mysql
  host: "localhost"
  port: 3306
  user: "root"
  password: "gonetdisk"
  name: "gonetdisk"
  charset: "utf8mb4"
  parseTime: true
  loc: "Local"
  max_idle_conns: 10      # 最大空闲连接数
  max_open_conns: 100     # 最大打开连接数
  log_mode: "info"        # 日志级别

jwt:
  secret: "123321"        # JWT 密钥
  expire_hours: 24        # Token 过期时间（小时）
```

### 环境变量

项目未使用环境变量覆盖配置，但 `.gitignore` 排除了敏感配置文件：
- `configs/config.dev.yaml`
- `.env` 文件

---

## 8. 构建与运行

### 环境要求

- Go 1.25.1+
- MySQL 8.0+
- Docker & Docker Compose（可选，用于运行 MySQL）

### 启动步骤

```bash
# 1. 启动 MySQL（使用 Docker）
cd docker
docker-compose up -d
cd ..

# 2. 下载依赖
go mod download

# 3. 运行服务
go run cmd/server/main.go

# 服务将监听 http://localhost:9090
```

### 构建可执行文件

```bash
# Windows
go build -o bin/server.exe cmd/server/main.go

# Linux/Mac
go build -o bin/server cmd/server/main.go
```

---

## 9. 开发规范

### 9.1 代码风格

- 使用标准 Go 代码格式（`gofmt`）
- 命名规范：
  - 结构体：PascalCase（如 `UserService`）
  - 接口：PascalCase + `er` 后缀（如 `Reader`）
  - 私有函数/变量：camelCase
  - 常量：UPPER_SNAKE_CASE 或 PascalCase
- 注释使用中文，函数注释说明功能和参数
- 错误信息使用中文

### 9.2 包结构规范

- `internal/` 存放不可被外部导入的代码
- `pkg/` 存放可被外部导入的公共包
- 每个包应有单一职责

### 9.3 DTO 命名规范

- 请求 DTO：`XxxRequest`
- 响应 DTO：`XxxResponse`
- 使用 struct tag 进行参数校验：`binding:"required,email"`

### 9.4 数据库规范

- 使用 GORM 进行数据库操作
- 模型定义使用蛇形命名字段（如 `password_hash`）
- 使用软删除（`gorm.DeletedAt`）
- 自动迁移在 `database.InitDB()` 中执行

---

## 10. 已知问题与注意事项

### 10.1 需要修复的问题

1. **配置文件硬编码路径**
   - 文件：`cmd/server/main.go` 第 14 行
   - 问题：使用了 Windows 绝对路径 `E:\Go_Project\GoNetDisk\configs\config.yaml`
   - 建议：改为相对路径 `./configs/config.yaml`

2. **文件名拼写错误**
   - `internal/repository/file_repoe.go` → 应为 `file_repo.go`
   - `internal/dto/flie_request.go` → 应为 `file_request.go`
   - `internal/dto/flie_response.go` → 应为 `file_response.go`

3. **JWT 密钥过于简单**
   - 当前使用 `"123321"`，生产环境必须更换为强密钥

4. **文件上传限制**
   - 当前仅支持 `.jpg` 格式（硬编码限制）
   - 文件大小限制为 10MB

### 10.2 开发注意事项

- 数据库表名使用单数形式（GORM `SingularTable: true`）
- 文件上传后存储逻辑未完成（仅保存元数据）
- Auth 中间件已创建但未应用到路由

---

## 11. 后续开发规划

| 序号 | 功能模块 | 状态 |
|------|----------|------|
| 1 | 文件上传 | 进行中（基础功能完成） |
| 2 | 文件下载 | 待开发 |
| 3 | 文件管理（目录、移动、重命名） | 待开发 |
| 4 | 文件分享 | 待开发 |
| 5 | 存储管理 | 待开发 |
| 6 | 用户管理 | 进行中 |
| 7 | 文件预览 | 待开发 |
| 8 | 文件同步 | 待开发 |
| 9 | 回收站 | 待开发 |
| 10 | 管理后台 | 待开发 |

---

## 12. 参考资料

- [Gin 框架文档](https://gin-gonic.com/docs/)
- [GORM 文档](https://gorm.io/docs/)
- [JWT-Go 文档](https://github.com/golang-jwt/jwt)
- [Viper 文档](https://github.com/spf13/viper)

---

*文档更新时间：2026-02-06*
