# GoNetDisk - Go语言网盘项目规划文档

## 一、项目需求分析

### 1.1 核心功能需求
- **用户认证系统**
  - 用户注册（用户名、邮箱、密码）
  - 用户登录（支持JWT令牌认证）
  - 密码加密存储（使用bcrypt）
  - 用户信息管理

- **文件管理功能**
  - 文件上传（支持大文件分片上传）
  - 文件下载（支持断点续传）
  - 文件列表查看
  - 文件删除
  - 文件重命名
  - 文件夹管理

- **存储管理**
  - 用户存储空间配额管理
  - 文件元数据存储
  - 文件实际存储（本地文件系统）

### 1.2 非功能性需求
- **安全性**
  - 密码加密存储
  - JWT令牌认证
  - 文件访问权限控制
  - 防止路径遍历攻击

- **性能**
  - 支持大文件上传（>100MB）
  - 并发文件处理
  - 数据库连接池

- **可扩展性**
  - 模块化设计
  - 支持未来扩展云存储
  - API版本管理

## 二、技术选型

### 2.1 后端技术栈
- **编程语言**: Go 1.21+
- **Web框架**: Gin (轻量级、高性能)
- **ORM**: GORM (功能完善、易用)
- **数据库**: MySQL 8.0 (关系型数据库)
- **缓存**: Redis (会话管理、缓存)
- **认证**: JWT (JSON Web Token)
- **配置管理**: Viper
- **日志**: Zap (高性能日志库)
- **验证**: go-playground/validator

### 2.2 前端技术栈（可选）
- **框架**: Vue 3 + Vite
- **UI组件库**: Element Plus
- **HTTP客户端**: Axios
- **状态管理**: Pinia

### 2.3 开发工具
- **容器化**: Docker + Docker Compose
- **API文档**: Swagger
- **代码规范**: golangci-lint
- **版本控制**: Git

## 三、项目架构设计

### 3.1 整体架构
```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   前端应用   │────▶│   API网关   │────▶│  业务服务层  │
└─────────────┘     └─────────────┘     └─────────────┘
                                              │
                    ┌─────────────────────────┼─────────────────────────┐
                    ▼                         ▼                         ▼
            ┌─────────────┐           ┌─────────────┐           ┌─────────────┐
            │  用户服务   │           │  文件服务   │           │  存储服务   │
            └─────────────┘           └─────────────┘           └─────────────┘
                    │                         │                         │
                    └─────────────────────────┼─────────────────────────┘
                                              ▼
                                    ┌─────────────────────┐
                                    │   数据访问层(DAO)    │
                                    └─────────────────────┘
                                              │
                    ┌─────────────────────────┼─────────────────────────┐
                    ▼                         ▼                         ▼
            ┌─────────────┐           ┌─────────────┐           ┌─────────────┐
            │   MySQL     │           │   Redis     │           │  文件系统   │
            └─────────────┘           └─────────────┘           └─────────────┘
```

### 3.2 分层架构
- **Controller层**: 处理HTTP请求，参数验证
- **Service层**: 业务逻辑处理
- **Repository层**: 数据访问
- **Model层**: 数据模型定义
- **Middleware层**: 中间件（认证、日志、CORS等）

## 四、数据库设计

### 4.1 用户表 (users)
```sql
CREATE TABLE users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE COMMENT '用户名',
    email VARCHAR(100) NOT NULL UNIQUE COMMENT '邮箱',
    password_hash VARCHAR(255) NOT NULL COMMENT '密码哈希',
    storage_used BIGINT UNSIGNED DEFAULT 0 COMMENT '已使用存储空间(字节)',
    storage_limit BIGINT UNSIGNED DEFAULT 1073741824 COMMENT '存储空间限制(字节)，默认1GB',
    status TINYINT DEFAULT 1 COMMENT '状态：1-正常，0-禁用',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_username (username),
    INDEX idx_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';
```

### 4.2 文件表 (files)
```sql
CREATE TABLE files (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    filename VARCHAR(255) NOT NULL COMMENT '文件名',
    original_name VARCHAR(255) NOT NULL COMMENT '原始文件名',
    file_path VARCHAR(500) NOT NULL COMMENT '文件存储路径',
    file_size BIGINT UNSIGNED NOT NULL COMMENT '文件大小(字节)',
    file_type VARCHAR(100) COMMENT '文件类型(MIME)',
    file_hash VARCHAR(64) COMMENT '文件哈希(SHA256)',
    parent_id BIGINT UNSIGNED DEFAULT 0 COMMENT '父文件夹ID，0表示根目录',
    is_directory TINYINT DEFAULT 0 COMMENT '是否为文件夹：1-是，0-否',
    status TINYINT DEFAULT 1 COMMENT '状态：1-正常，0-已删除',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_parent_id (parent_id),
    INDEX idx_file_hash (file_hash)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文件表';
```

### 4.3 分享表 (shares) - 可选扩展
```sql
CREATE TABLE shares (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    file_id BIGINT UNSIGNED NOT NULL COMMENT '文件ID',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '创建者用户ID',
    share_code VARCHAR(32) NOT NULL UNIQUE COMMENT '分享码',
    password VARCHAR(255) COMMENT '访问密码',
    expire_time DATETIME COMMENT '过期时间',
    download_count INT UNSIGNED DEFAULT 0 COMMENT '下载次数',
    max_download_count INT UNSIGNED COMMENT '最大下载次数',
    status TINYINT DEFAULT 1 COMMENT '状态：1-有效，0-已失效',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_share_code (share_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='分享表';
```

## 五、项目目录结构

```
GoNetDisk/
├── cmd/
│   └── server/
│       └── main.go                 # 程序入口
├── internal/
│   ├── controller/                 # 控制器层
│   │   ├── user_controller.go
│   │   ├── file_controller.go
│   │   └── auth_controller.go
│   ├── service/                    # 业务逻辑层
│   │   ├── user_service.go
│   │   ├── file_service.go
│   │   └── auth_service.go
│   ├── repository/                 # 数据访问层
│   │   ├── user_repository.go
│   │   └── file_repository.go
│   ├── model/                      # 数据模型
│   │   ├── user.go
│   │   ├── file.go
│   │   └── response.go
│   ├── middleware/                 # 中间件
│   │   ├── auth.go
│   │   ├── cors.go
│   │   └── logger.go
│   ├── config/                     # 配置
│   │   └── config.go
│   └── util/                       # 工具函数
│       ├── jwt.go
│       ├── password.go
│       └── file.go
├── pkg/                            # 公共包
│   ├── logger/
│   └── database/
├── storage/                        # 文件存储目录
│   └── uploads/
├── configs/                        # 配置文件
│   ├── config.yaml
│   └── config.dev.yaml
├── scripts/                        # 脚本文件
│   ├── init.sql
│   └── migrate.sql
├── docs/                           # 文档
│   └── api/
├── docker/                         # Docker配置
│   ├── Dockerfile
│   └── docker-compose.yml
├── .gitignore
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## 六、核心功能实现思路

### 6.1 用户注册登录流程

**注册流程:**
1. 接收用户注册信息（用户名、邮箱、密码）
2. 验证参数（用户名格式、邮箱格式、密码强度）
3. 检查用户名和邮箱是否已存在
4. 使用bcrypt加密密码
5. 创建用户记录
6. 返回成功信息

**登录流程:**
1. 接收登录信息（用户名/邮箱、密码）
2. 验证用户是否存在
3. 验证密码是否正确
4. 生成JWT令牌
5. 返回令牌和用户信息

### 6.2 文件上传流程

**单文件上传:**
1. 验证用户身份（JWT）
2. 检查用户存储空间是否足够
3. 接收文件流
4. 计算文件哈希（SHA256）
5. 检查文件是否已存在（去重）
6. 保存文件到存储系统
7. 更新用户存储空间
8. 保存文件元数据到数据库
9. 返回文件信息

**大文件分片上传:**
1. 前端将文件分片
2. 每个分片独立上传
3. 服务端验证分片完整性
4. 合并分片
5. 验证完整文件哈希
6. 清理临时分片文件

### 6.3 文件下载流程

**普通下载:**
1. 验证用户身份
2. 验证文件访问权限
3. 检查文件是否存在
4. 设置响应头（Content-Type, Content-Disposition）
5. 流式传输文件内容

**断点续传:**
1. 解析Range请求头
2. 定位文件起始位置
3. 设置Content-Range响应头
4. 返回指定范围的数据

## 七、Docker配置方案

### 7.1 Dockerfile
```dockerfile
# 多阶段构建
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server cmd/server/main.go

# 运行阶段
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /root/
COPY --from=builder /app/server .
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/storage ./storage

EXPOSE 8080
CMD ["./server"]
```

### 7.2 docker-compose.yml
```yaml
version: '3.8'

services:
  mysql:
    image: mysql:8.0
    container_name: gonetdisk-mysql
    environment:
      MYSQL_ROOT_PASSWORD: root123
      MYSQL_DATABASE: gonetdisk
      MYSQL_USER: gonetdisk
      MYSQL_PASSWORD: gonetdisk123
    ports:
      - "3306:3306"
    volumes:
      - mysql-data:/var/lib/mysql
      - ./scripts/init.sql:/docker-entrypoint-initdb.d/init.sql
    networks:
      - gonetdisk-network

  redis:
    image: redis:7-alpine
    container_name: gonetdisk-redis
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    networks:
      - gonetdisk-network

  app:
    build:
      context: .
      dockerfile: docker/Dockerfile
    container_name: gonetdisk-app
    ports:
      - "8080:8080"
    volumes:
      - ./storage:/root/storage
    environment:
      - GIN_MODE=release
      - DB_HOST=mysql
      - DB_PORT=3306
      - DB_USER=gonetdisk
      - DB_PASSWORD=gonetdisk123
      - DB_NAME=gonetdisk
      - REDIS_HOST=redis
      - REDIS_PORT=6379
    depends_on:
      - mysql
      - redis
    networks:
      - gonetdisk-network

volumes:
  mysql-data:
  redis-data:

networks:
  gonetdisk-network:
    driver: bridge
```

## 八、API接口设计

### 8.1 用户认证接口
```
POST   /api/v1/auth/register    # 用户注册
POST   /api/v1/auth/login       # 用户登录
POST   /api/v1/auth/logout      # 用户登出
GET    /api/v1/auth/info        # 获取当前用户信息
```

### 8.2 文件管理接口
```
POST   /api/v1/files/upload     # 上传文件
GET    /api/v1/files            # 获取文件列表
GET    /api/v1/files/:id        # 获取文件详情
GET    /api/v1/files/:id/download # 下载文件
DELETE /api/v1/files/:id        # 删除文件
PUT    /api/v1/files/:id        # 更新文件信息
POST   /api/v1/files/folder     # 创建文件夹
```

### 8.3 用户管理接口
```
GET    /api/v1/users/profile    # 获取用户资料
PUT    /api/v1/users/profile    # 更新用户资料
GET    /api/v1/users/storage    # 获取存储空间信息
```

## 九、安全考虑

### 9.1 认证安全
- 使用JWT进行无状态认证
- Token设置合理的过期时间
- 密码使用bcrypt加密（cost=10）
- 实现Token刷新机制

### 9.2 文件安全
- 验证文件类型（MIME类型）
- 限制文件大小
- 防止路径遍历攻击
- 文件名消毒处理
- 实现文件访问权限控制

### 9.3 API安全
- 实现请求频率限制
- 输入参数验证
- SQL注入防护（使用ORM）
- XSS防护
- CORS配置

## 十、开发规范

### 10.1 代码规范
- 遵循Go官方代码规范
- 使用gofmt格式化代码
- 使用golangci-lint进行代码检查
- 函数和变量使用驼峰命名
- 导出函数添加注释

### 10.2 Git规范
- 使用语义化版本
- 提交信息格式：`type(scope): description`
  - feat: 新功能
  - fix: 修复bug
  - docs: 文档更新
  - style: 代码格式调整
  - refactor: 重构
  - test: 测试相关
  - chore: 构建/工具相关

### 10.3 测试规范
- 单元测试覆盖率>80%
- 集成测试覆盖核心流程
- 使用testify测试框架

## 十一、性能优化

### 11.1 数据库优化
- 合理使用索引
- 使用连接池
- 慢查询优化
- 读写分离（可选）

### 11.2 缓存策略
- Redis缓存用户信息
- Redis缓存文件元数据
- 实现缓存失效机制

### 11.3 文件传输优化
- 支持断点续传
- 使用流式传输
- 实现文件压缩（可选）

## 十二、监控和日志

### 12.1 日志管理
- 使用Zap高性能日志库
- 日志分级：DEBUG、INFO、WARN、ERROR
- 日志文件轮转
- 结构化日志输出

### 12.2 监控指标
- API响应时间
- 文件上传/下载速度
- 存储空间使用率
- 用户活跃度

## 十三、部署方案

### 13.1 开发环境
- 本地运行MySQL和Redis
- 使用热重载工具（air）
- 开发模式配置

### 13.2 生产环境
- Docker容器化部署
- Nginx反向代理
- HTTPS配置
- 定期数据备份

## 十四、后续扩展方向

1. **功能扩展**
   - 文件分享功能
   - 文件预览（图片、PDF、视频）
   - 文件版本管理
   - 回收站功能
   - 文件搜索

2. **存储扩展**
   - 支持对象存储（OSS、S3）
   - 分布式文件存储
   - CDN加速

3. **协作功能**
   - 文件协作编辑
   - 团队空间
   - 权限管理

4. **移动端支持**
   - 移动端API
   - 移动端App

## 十五、开发时间估算

- 项目搭建和配置：1-2天
- 数据库设计和实现：1天
- 用户认证功能：2-3天
- 文件上传功能：3-4天
- 文件下载功能：2-3天
- 文件管理功能：2-3天
- Docker配置和部署：1-2天
- 测试和优化：2-3天
- 文档编写：1天

**总计：约15-20个工作日**

## 十六、参考资料

- [Gin框架文档](https://gin-gonic.com/docs/)
- [GORM文档](https://gorm.io/docs/)
- [JWT最佳实践](https://jwt.io/introduction)
- [Go语言最佳实践](https://go.dev/doc/effective_go)
- [Docker最佳实践](https://docs.docker.com/develop/dev-best-practices/)