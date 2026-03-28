# GoNetDisk

<p align="center">
  <b>基于 Go 的轻量级网盘后端服务</b>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25.1-00ADD8?style=flat-square&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/Gin-1.11.0-00ADD8?style=flat-square" alt="Gin">
  <img src="https://img.shields.io/badge/GORM-1.31.1-blue?style=flat-square" alt="GORM">
  <img src="https://img.shields.io/badge/MySQL-8.0-4479A1?style=flat-square&logo=mysql&logoColor=white" alt="MySQL">
  <img src="https://img.shields.io/badge/License-MIT-green?style=flat-square" alt="License">
</p>

---

GoNetDisk 当前是一套最小可用的网盘后端，已经实现用户体系、JWT 鉴权、单文件上传、文件哈希去重、目录创建和基础配额累加。下载、列表、删除、分享、预览、同步、回收站、管理后台等能力还未在 Go 代码中落地。

## 当前能力

- 用户注册 / 登录
- JWT Token 鉴权
- 获取 / 更新当前用户信息
- 单文件上传
- 物理文件 MD5 去重
- 基础目录创建
- 用户已用空间累计
- Docker 启动 MySQL 开发环境

## 架构

```text
┌──────────────────────────────────┐
│            HTTP 层 (Gin)         │
│ Router -> Middleware -> Controller│
├──────────────────────────────────┤
│          Service 业务层          │
│ 鉴权身份使用 / 上传编排 / 配额校验 │
├──────────────────────────────────┤
│        Repository 数据访问层      │
│           GORM 查询更新           │
├──────────────────────────────────┤
│        MySQL + 本地文件系统       │
└──────────────────────────────────┘
```

## 项目结构

```text
GoNetDisk/
├── cmd/server/              # 服务入口
├── configs/                 # 配置结构和 YAML
├── internal/
│   ├── controller/          # HTTP 控制器
│   ├── dto/                 # 请求/响应 DTO
│   ├── middleware/          # JWT 中间件
│   ├── model/               # GORM 模型
│   ├── repository/          # 数据访问层
│   ├── router/              # 路由装配
│   ├── service/             # 业务逻辑
│   └── util/                # JWT 等工具
├── pkg/database/            # 数据库初始化
├── docker/                  # MySQL Docker 配置
├── storage/temp/            # 临时上传目录
├── storage/uploads/         # 正式文件目录
└── ai-docs/                 # AI 协作文档
```

## 技术栈

| 类别 | 技术 |
|------|------|
| 语言 | Go 1.25.1 |
| Web 框架 | Gin 1.11.0 |
| ORM | GORM 1.31.1 |
| 数据库 | MySQL 8.0 |
| 配置 | Viper 1.21.0 |
| 认证 | golang-jwt/jwt/v5 5.3.0 |
| 密码哈希 | bcrypt |

## 快速开始

### 环境要求

- Go 1.25.1+
- MySQL 8.0+，或 Docker

### 1. 启动数据库

```bash
cd docker
docker-compose up -d
```

默认会创建 `gonetdisk` 数据库并执行 `docker/init/init.sql`。

### 2. 检查配置

编辑 `configs/config.yaml`：

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

注意：`server.mode`、数据库连接池参数和 `log_mode` 目前在代码里还没有完全接线，配置文件和实际运行行为并不完全一致。

### 3. 启动服务

```bash
go run cmd/server/main.go
```

入口通过相对路径加载 `./configs/config.yaml`，因此应在仓库根目录执行。

默认监听地址：`http://localhost:9090`

### 4. 局域网访问与上传

服务端当前按 `0.0.0.0:9090` 监听，因此同一局域网内的其他机器可以直接访问这个后端。项目仍然没有内置前端页面，但浏览器前端、桌面客户端、Postman、`curl` 都可以直接调用上传接口。

接入步骤：

1. 在服务端机器执行 `ipconfig`，确认局域网 IPv4 地址，例如 `192.168.1.50`。
2. 确认操作系统防火墙已经放行 TCP `9090` 入站。
3. 客户端把请求地址从 `http://localhost:9090` 改成 `http://192.168.1.50:9090`。
4. 浏览器客户端可以直接发起跨域上传请求；服务端已经放行 `Authorization`、`Content-Type` 和 `multipart/form-data` 预检请求。

局域网上传示例：

```bash
curl -X POST http://192.168.1.50:9090/api/v1/file/upload \
  -H "Authorization: Bearer <your-token>" \
  -F "parent_id=0" \
  -F "file=@/path/to/photo.jpg"
```

### 5. 编译基线校验

当前仓库没有测试文件，但可以先做一次编译链路校验：

```powershell
$env:GOCACHE = (Join-Path $PWD '.gocache')
go test ./...
```

## API

基础前缀：`/api/v1`

### 用户模块

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|:----:|
| POST | `/user/register` | 注册 | 否 |
| POST | `/user/login` | 登录 | 否 |
| GET | `/user/info` | 获取当前用户信息 | 是 |
| PUT | `/user/info` | 更新当前用户信息 | 是 |

### 文件模块

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|:----:|
| POST | `/file/upload` | 上传文件 | 是 |

上传接口使用 `multipart/form-data`：

- `file`: 必填，上传文件
- `parent_id`: 选填，父目录 ID，根目录传 `0`

### 文件夹模块

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|:----:|
| POST | `/folder/create` | 创建文件夹 | 是 |

创建文件夹请求支持 JSON 或表单：

- `folder_name`: 必填，文件夹名
- `parent_id`: 选填，父目录 ID，根目录传 `0`

## 请求示例

### 注册

```bash
curl -X POST http://localhost:9090/api/v1/user/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"test@example.com","password":"123456"}'
```

### 登录

```bash
curl -X POST http://localhost:9090/api/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"123456"}'
```

### 上传文件

```bash
curl -X POST http://localhost:9090/api/v1/file/upload \
  -H "Authorization: Bearer <your-token>" \
  -F "parent_id=0" \
  -F "file=@/path/to/photo.jpg"
```

如果从局域网其他机器访问，只需要把 `localhost` 替换成服务端机器的局域网 IP。

### 创建文件夹

```bash
curl -X POST http://localhost:9090/api/v1/folder/create \
  -H "Authorization: Bearer <your-token>" \
  -H "Content-Type: application/json" \
  -d '{"folder_name":"docs","parent_id":0}'
```

## 数据设计

当前 Go 代码实际迁移和使用的核心表有三张：

- `user`
- `physical_file`
- `user_file`

设计要点：

- `physical_file` 负责物理文件元数据和去重
- `user_file` 负责用户目录视图
- 相同内容的文件只存一份物理文件，通过 `file_hash` 复用
- 当前摘要算法为 `md5`

`docker/init/init.sql` 里还存在 `role`、`admin`、`permission`、`role_permission` 表草案，但当前没有对应的 Go 实现。

## 当前可优化问题

- 配置项存在但未完全生效：`server.mode`、连接池参数、`log_mode`
- 上传返回的 `download_url` 仍是本地路径，没有下载接口
- 用户模块错误码过于粗糙，`409` 使用过多
- 上传输入校验不足，缺少文件名净化和 MIME 白名单
- 上传目录和最大文件大小硬编码
- 空间管理只有上传增量累加，没有查询和校正能力
- 仓库没有自动化测试

## 开发路线

- [x] 用户注册 / 登录
- [x] JWT 鉴权
- [x] 文件上传
- [x] 文件哈希去重
- [x] 文件夹创建
- [ ] 文件下载
- [ ] 文件列表 / 重命名 / 移动 / 删除
- [ ] 文件分享
- [ ] 文件预览
- [ ] 文件同步
- [ ] 回收站
- [ ] 完整存储管理
- [ ] 管理后台

## License

MIT
