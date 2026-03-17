# GoNetDisk

<p align="center">
  <b>基于 Go 语言的网络云盘系统</b>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25.1-00ADD8?style=flat-square&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/Gin-1.11.0-00ADD8?style=flat-square" alt="Gin">
  <img src="https://img.shields.io/badge/GORM-1.31.1-blue?style=flat-square" alt="GORM">
  <img src="https://img.shields.io/badge/MySQL-8.0-4479A1?style=flat-square&logo=mysql&logoColor=white" alt="MySQL">
  <img src="https://img.shields.io/badge/License-MIT-green?style=flat-square" alt="License">
</p>

---

GoNetDisk 是一个轻量级的网络云盘/文件存储系统后端服务，采用经典分层架构设计，支持用户管理、文件上传、文件哈希去重（秒传）等核心功能。

## ✨ 特性

- 🔐 用户注册 / 登录，JWT Token 认证
- 📤 文件上传，支持按日期自动归档
- ⚡ 文件哈希去重，相同文件秒传
- 📁 物理文件与用户文件双表分离，节省存储空间
- 🐳 Docker 一键部署 MySQL 环境
- 🏗️ 清晰的分层架构，易于扩展

## 🏛️ 架构

```
┌─────────────────────────────────┐
│         HTTP 层 (Gin)           │
│  Router → Middleware → Controller│
├─────────────────────────────────┤
│       业务逻辑层 (Service)       │
│   事务管理 / 业务编排 / 去重逻辑  │
├─────────────────────────────────┤
│       数据访问层 (Repository)     │
│        GORM 数据库操作           │
├─────────────────────────────────┤
│           MySQL 8.0             │
└─────────────────────────────────┘
```

## 📂 项目结构

```
GoNetDisk/
├── cmd/server/              # 应用入口
│   └── main.go
├── configs/                 # 配置文件
│   ├── config.yaml
│   └── config.go
├── internal/                # 内部业务代码
│   ├── controller/          # 控制器层 - 请求处理
│   ├── service/             # 业务逻辑层
│   ├── repository/          # 数据访问层
│   ├── model/               # 数据模型 (GORM)
│   ├── dto/                 # 数据传输对象
│   ├── middleware/          # 中间件 (JWT 认证)
│   ├── router/              # 路由定义
│   └── util/                # 工具类 (JWT Manager)
├── pkg/database/            # 数据库连接
├── docker/                  # Docker 部署配置
├── storage/uploads/         # 文件存储目录
└── ai-docs/                 # 项目文档
```

## 🛠️ 技术栈

| 类别 | 技术 | 用途 |
|------|------|------|
| 语言 | Go 1.25.1 | 开发语言 |
| Web 框架 | Gin v1.11.0 | HTTP 路由与中间件 |
| ORM | GORM v1.31.1 | 数据库操作 |
| 数据库 | MySQL 8.0 | 数据持久化 |
| 配置管理 | Viper v1.21.0 | YAML 配置读取 |
| 认证 | golang-jwt v5.3.0 | JWT Token |
| 密码加密 | bcrypt | 密码哈希存储 |
| 容器化 | Docker Compose | 环境部署 |

## 🚀 快速开始

### 环境要求

- Go 1.25.1+
- MySQL 8.0+（或使用 Docker）

### 1. 克隆项目

```bash
git clone https://github.com/Manyo-Dream/gonetdisk.git
cd gonetdisk
```

### 2. 启动数据库

```bash
cd docker
docker-compose up -d
```

这会自动创建 `gonetdisk` 数据库并执行初始化 SQL 脚本。

### 3. 修改配置

编辑 `configs/config.yaml`，根据实际环境调整数据库连接信息：

```yaml
server:
  port: 9090
  host: "0.0.0.0"
  mode: debug

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
  max_idle_conns: 10
  max_open_conns: 100
  log_mode: "info"

jwt:
  secret: "your-secret-key"
  expiresHours: 24
```

### 4. 运行服务

```bash
go run cmd/server/main.go
```

服务入口当前通过相对路径加载 `./configs/config.yaml`，请在仓库根目录执行上面的命令。

服务启动后默认监听 `http://localhost:9090`。

`server.mode` 字段已经存在于配置中，但当前入口还没有调用 `gin.SetMode`，因此这个配置项暂未生效。

### 5. 基线校验

当前仓库还没有单元测试文件，但可以先跑一遍基础测试命令确认编译链路正常：

```powershell
$env:GOCACHE = (Join-Path $PWD '.gocache')
go test ./...
```

## 📡 API 接口

基础路径：`/api/v1`

### 用户模块

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|:----:|
| POST | `/user/register` | 用户注册 | ✗ |
| POST | `/user/login` | 用户登录 | ✗ |
| GET | `/user/info` | 获取用户信息 | ✓ |
| PUT | `/user/info` | 更新用户信息 | ✓ |

`/user/info` 相关接口当前使用 JWT 中的用户身份，不再信任客户端传入的邮箱或用户 ID。

### 文件模块

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|:----:|
| POST | `/file/upload` | 文件上传 | ✓ |

### 示例

**注册**
```bash
curl -X POST http://localhost:9090/api/v1/user/register \
  -H "Content-Type: application/json" \
  -d '{"username": "test", "email": "test@example.com", "password": "123456"}'
```

**登录**
```bash
curl -X POST http://localhost:9090/api/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "123456"}'
```

**上传文件**
```bash
curl -X POST http://localhost:9090/api/v1/file/upload \
  -H "Authorization: Bearer <your-token>" \
  -F "parentID=0" \
  -F "file=@/path/to/photo.jpg"
```

上传接口使用 `multipart/form-data`，当前支持的表单字段如下：

- `file`: 必填，上传的文件
- `parentID`: 选填，父目录 ID，根目录可传 `0`

当前实现还有两个需要注意的点：

- 单文件大小限制为 `100MB`
- 返回字段 `download_url` 目前是服务端保存的本地文件路径，还不是可直接访问的 HTTP 下载地址

## 💾 数据库设计

项目采用物理文件与用户文件双表分离的设计：

```
┌──────────┐       ┌────────────────┐       ┌──────────┐
│   User   │ 1───N │   UserFile     │ N───1 │PhysicalFile│
│          │       │                │       │            │
│ id       │       │ user_id (FK)   │       │ id         │
│ username │       │ physical_id(FK)│       │ file_hash  │
│ email    │       │ file_name      │       │ file_path  │
│       │ parent_id      │       │ ref_count  │
└──────────┘       └────────────────┘       └────────────┘
```

- 相同内容的文件只存储一份物理文件，通过哈希去重实现秒传
- 用户文件表维护每个用户独立的文件目录视图
- 引用计数 `ref_count` 追踪物理文件被引用次数
- 当前代码中的文件哈希算法实际使用 `md5`

## 🗺️ 开发路线

- [x] 用户注册 / 登录
- [x] JWT 认证
- [x] 文件上传
- [x] 文件哈希去重
- [ ] 文件下载
- [ ] 文件管理（重命名 / 移动 / 删除）
- [ ] 文件分享
- [ ] 文件预览
- [ ] 回收站
- [ ] 存储空间管理
- [ ] 管理后台

## 📄 License

MIT
