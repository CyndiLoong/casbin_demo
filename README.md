# casbin_demo · 生产级 RBAC 权限管理 + 大模型 API 资源审核平台

[![Go](https://img.shields.io/badge/Go-1.22-00ADD8?logo=go)](https://go.dev)
[![Gin](https://img.shields.io/badge/Gin-1.10-008ECB)](https://gin-gonic.com)
[![Vue](https://img.shields.io/badge/Vue-3.4-4FC08D?logo=vue.js)](https://vuejs.org)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.4-3178C6?logo=typescript)](https://www.typescriptlang.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?logo=postgresql)](https://www.postgresql.org)
[![Redis](https://img.shields.io/badge/Redis-7-DC382D?logo=redis)](https://redis.io)
[![RabbitMQ](https://img.shields.io/badge/RabbitMQ-3-FF6600?logo=rabbitmq)](https://www.rabbitmq.com)
[![Fx](https://img.shields.io/badge/uber--go/fx-1.24-5E50A1)](https://github.com/uber-go/fx)
[![Casbin](https://img.shields.io/badge/Casbin-v2-6646EA)](https://casbin.org)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker)](https://www.docker.com)

> 一个前后端分离的 **RBAC 权限管理系统 + 大模型 API 资源审核平台**。后端使用 Go (Gin) + uber-go/fx + Casbin + GORM，前端使用 Vue 3 + Element Plus + Pinia，通过 PostgreSQL + Redis + RabbitMQ 支撑缓存、限流、实时通知等生产级特性。

---

## 目录

- [功能概览](#功能概览)
- [技术栈](#技术栈)
- [系统架构](#系统架构)
- [目录结构](#目录结构)
- [快速开始（Docker Compose）](#快速开始docker-compose)
- [本地开发](#本地开发)
- [环境变量与配置](#环境变量与配置)
- [默认账号](#默认账号)
- [API 概览](#api-概览)
- [数据库设计](#数据库设计)
- [部署](#部署)
- [技术文档](#技术文档)
- [许可证](#许可证)

---

## 功能概览

### 权限与用户
- 用户注册 / 登录 / JWT 认证 / bcrypt 密码加密 / Token 自动刷新
- 用户、角色、权限 CRUD，内置 `admin` / `user` 角色
- Casbin RBAC 模型，接口级细粒度权限校验，策略持久化到 PostgreSQL 并自动重载

### 资源审核平台
- 大模型 API 资源清单（聊天 / 代码 / 图像 / 语音 / 嵌入 五大分类）
- 用户提交使用申请 → 管理员审核（通过 / 驳回）
- **2 分钟撤回机制**：Redis TTL 控制撤回窗口（前端不可篡改），PG 时间戳兜底，乐观锁防并发
- 站内信持久化，离线消息登录兜底

### 实时通知
- RabbitMQ Fanout 广播 → Redis PubSub 跨网关 → WebSocket 实时弹窗
- 单用户连接数限制、心跳保活、幂等去重（LRU）

### 生产级基础设施
- **两级缓存**：L1 本地内存（`sync.Map`）+ L2 Redis
- **布隆过滤器**：缓存穿透第一道防线，启动预热合法 ID
- **三层限流**：IP 级 30 req/s + 接口级 100 req/s + 用户级 20 req/s（令牌桶 + Lua）
- **熔断器**：Redis 故障自动降级为本地实现，三态自动探测恢复
- **优雅关停**：`fx.Lifecycle` 管理 HTTP Server、Hub、MQ、限流器按序释放
- 结构化日志（`log/slog`）、Panic 恢复中间件、统一错误码、自定义 401/403/404/500 页面

---

## 技术栈

### 后端 `backend/`
| 类别 | 依赖 |
|------|------|
| 语言 | Go 1.22 |
| Web 框架 | Gin v1.10 |
| 依赖注入 | uber-go/fx v1.24 |
| ORM | GORM v1.25 + Postgres Driver |
| 权限引擎 | Casbin v2 + gorm-adapter v3 |
| 缓存 | go-redis v8 |
| 限流 | gin-contrib/cors + 自研 Lua 令牌桶 |
| 消息队列 | RabbitMQ AMQP (`amqp091-go`) |
| WebSocket | gorilla/websocket |
| 鉴权 | golang-jwt v5 |
| 配置 | spf13/viper |
| 加密 | golang.org/x/crypto (bcrypt) |

### 前端 `frontend/`
| 类别 | 依赖 |
|------|------|
| 框架 | Vue 3.4 + TypeScript 5.4 |
| 构建 | Vite 5 |
| UI | Element Plus 2.6 + Element Icons |
| 状态 | Pinia 2 |
| 路由 | Vue Router 4 |
| HTTP | Axios |
| 样式 | SCSS + Animate.css |
| 进度条 | NProgress |

### 中间件
- PostgreSQL 16（端口映射 `5433` → 容器内 `5432`）
- Redis 7（`6380` → `6379`）
- RabbitMQ 3（`5672`/`15672` 管理台）

---

## 系统架构

```
                          ┌──────────────────┐
  Browser ── Element Plus │   Vue 3 SPA      │
   (3000)   Axios/WS      │  Pinia / Router   │
                          └────────┬─────────┘
                                   │  HTTP :8080 / WS
                          ┌────────▼─────────┐
                          │  Gin + uber-go/fx │
                          │  ┌─── Middleware ─┴───┐
                          │  │ JWT · Casbin ·     │
                          │  │ RateLimit · CORS · │
                          │  │ Recovery           │
                          │  └───┬────────────────┘
                          │      ├── Service ── Repository ── PostgreSQL 16
                          │      ├── Cache L1+L2            ── Redis 7
                          │      └── MQ (Fanout) ── Redis PubSub ── WS Hub
                          └─────────────────────┘        └── 跨实例广播
```

**数据流**：请求 → JWT 鉴权 → Casbin RBAC 校验 → Handler → Service → Repository → PG；写操作事务提交后发 RabbitMQ Fanout，订阅方通过 Redis PubSub 跨实例广播到所有 WebSocket Hub，Hub 再推送给前端；离线消息登录时从 PG 拉取。

详细架构设计、模块解析、缓存/限流/MQ/WS 深度说明见 [TECHNICAL_DOCUMENTATION.md](./TECHNICAL_DOCUMENTATION.md)（22 章）。

---

## 目录结构

```
casbin_demo/
├── backend/
│   ├── cmd/server/main.go              # fx App 入口
│   ├── internal/
│   │   ├── app/                        # fx 模块注册 & 应用初始化
│   │   ├── config/                     # Viper 配置
│   │   ├── handler/                   # 8 个 Handler（auth、user、role、permission、resource、audit、dashboard、ws）
│   │   ├── middleware/                # JWT / Casbin / RateLimit
│   │   ├── model/                     # GORM 模型 & Casbin 策略
│   │   ├── repository/                 # DB / Redis / 各资源 Repo
│   │   ├── router/                    # Gin 路由
│   │   └── service/                   # 业务逻辑
│   ├── pkg/
│   │   ├── cache/                     # L1+L2 缓存 / 布隆过滤器 / 熔断器
│   │   ├── casbin/                    # 权限引擎封装 + model.conf
│   │   ├── jwt/                       # JWT 工具
│   │   ├── mq/                        # RabbitMQ Fanout 客户端
│   │   ├── response/                  # 统一响应封装
│   │   └── ws/                        # WebSocket Hub
│   ├── tests/                         # go test
│   ├── Dockerfile                     # 多阶段 Go 构建
│   ├── config.yaml                    # 默认配置（本地）
│   └── config-docker.yaml             # 容器内配置
├── frontend/
│   ├── src/
│   │   ├── api/                       # Axios 接口层（auth/user/rbac/resource/audit）
│   │   ├── components/                # 通用组件（通知铃铛、错误页）
│   │   ├── router/                    # Vue Router
│   │   ├── store/                     # Pinia user store
│   │   ├── utils/websocket.ts         # WS 客户端（带心跳 & 自动重连）
│   │   └── views/                     # 10+ 页面（登录、仪表盘、用户、角色、权限、
│   │                                  # 资源、申请、审核、我的申请、消息、错误页）
│   ├── Dockerfile                     # 多阶段 Vite 构建 + Nginx 托管
│   ├── nginx.conf
│   ├── vite.config.ts / tsconfig.json
│   └── package.json
├── scripts/
│   ├── test-api.ps1                   # PowerShell 接口测试脚本
│   └── test-api.sh
├── docker-compose.yml                 # 一键拉起 postgres/redis/rabbitmq/backend/frontend
├── Makefile                           # build/run/test/docker-* 常用命令
├── init.sql                           # PG 初始化辅助脚本
└── TECHNICAL_DOCUMENTATION.md         # 22 章深度技术文档
```

---

## 快速开始（Docker Compose）

```powershell
# 在项目根目录执行（Windows PowerShell）
docker compose up -d --build

# 查看日志
docker compose logs -f backend
docker compose logs -f frontend

# 停止并清理数据卷
docker compose down -v
```

启动完成后：

| 服务 | 地址 |
|------|------|
| 前端（Nginx 托管） | http://localhost:3000 |
| 后端 API | http://localhost:8080 |
| 健康检查 | http://localhost:8080/health |
| RabbitMQ 管理台 | http://localhost:15672 （guest / guest） |
| PostgreSQL | `localhost:5433`（postgres / postgres / `casbin_demo`） |
| Redis | `localhost:6380` |

> 首次启动时 GORM AutoMigrate 会自动创建表结构，无需手动执行 `init.sql`（它是备用参考）。

---

## 本地开发

### 前置依赖
- Go 1.22+
- Node.js 18+（建议 20 LTS）
- PostgreSQL 16（本地 5432 或映射 5433）
- Redis 7（本地 6379 或映射 6380）
- RabbitMQ 3（本地 5672，管理台 15672）

### 1. 启动中间件（示例用 Docker 临时拉起）
```powershell
docker run -d --name casbin-pg  -p 5433:5432 -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=casbin_demo postgres:16-alpine
docker run -d --name casbin-redis -p 6380:6379 redis:7-alpine
docker run -d --name casbin-mq   -p 5672:5672 -p 15672:15672 -e RABBITMQ_DEFAULT_USER=guest -e RABBITMQ_DEFAULT_PASS=guest rabbitmq:3-management-alpine
```

### 2. 后端
```powershell
cd backend
go mod tidy
go run ./cmd/server
# 监听 :8080，读取 ./config.yaml
```

如需容器环境变量覆盖，复制 `config-docker.yaml` 或设置环境变量（见 [环境变量与配置](#环境变量与配置)）。

### 3. 前端
```powershell
cd frontend
npm install
npm run dev
# 默认 Vite :5173，已代理 /api 到 :8080；Nginx 构建后托管在 :80
```

### 使用 Makefile
```powershell
make install-deps      # 后端 go mod tidy + 前端 npm install
make build             # 两端打包
make run               # 后端运行
make frontend-run      # 前端 Vite 开发模式
make test-backend      # go test -v ./...
make test-api          # 运行 scripts/test-api.ps1
make docker-up         # docker compose up -d --build
make docker-down       # docker compose down -v
make clean             # 清理二进制与 dist
```

---

## 环境变量与配置

后端读取 `backend/config.yaml`，并允许以下环境变量覆盖（也可用 `backend/config-docker.yaml`）：

| 配置节 | 环境变量 | 默认值（docker） | 说明 |
|--------|----------|-----------------|------|
| database.host | `DATABASE_HOST` | `postgres` | PG 主机 |
| database.port | `DATABASE_PORT` | `5432` | PG 端口 |
| database.user | `DATABASE_USER` | `postgres` | PG 用户 |
| database.password | `DATABASE_PASSWORD` | `postgres` | PG 密码 |
| database.dbname | `DATABASE_DBNAME` | `casbin_demo` | PG 数据库 |
| redis.host | `REDIS_HOST` | `redis` | Redis 主机 |
| redis.port | `REDIS_PORT` | `6379` | Redis 端口 |
| rabbitmq.host | `RABBITMQ_HOST` | `rabbitmq` | MQ 主机 |
| rabbitmq.port | `RABBITMQ_PORT` | `5672` | MQ 端口 |
| rabbitmq.user | `RABBITMQ_USER` | `guest` | MQ 用户 |
| rabbitmq.password | `RABBITMQ_PASSWORD` | `guest` | MQ 密码 |
| jwt.secret | — | 见 yaml | **请生产环境替换** |
| jwt.expire_hours | — | `24` | Token 有效期 |
| server.port | — | `8080` | API 端口 |
| server.mode | — | `debug` | `debug` / `release` |

前端环境变量见 `frontend/vite.config.ts`（默认 `/api` → `:8080`）。

---

## 默认账号

系统首次启动时 GORM AutoMigrate + 种子初始化内置：

| 角色 | 用户名 | 密码 | 说明 |
|------|--------|------|------|
| admin | `admin` | `admin123` | 完整管理权限（琥珀金色系 UI） |
| user  | `alice` | `alice123` | 普通用户（靛蓝系 UI） |

> 生产环境请立即修改密码并替换 JWT secret。

---

## API 概览

所有接口以 `/api` 前缀访问，统一返回 `{"code": 0, "msg": "...", "data": ...}` 格式。

### 公开
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/auth/register` | 注册 |
| POST | `/api/auth/login` | 登录（返回 JWT） |
| GET  | `/health` | 健康检查 |

### 基础认证（Bearer JWT）
| 方法 | 路径 | 说明 |
|------|------|------|
| GET    | `/api/users/profile` | 当前用户 |
| GET    | `/api/dashboard`     | 仪表盘统计（按角色返回） |
| GET    | `/api/resources`       | 大模型资源清单 |
| POST   | `/api/audit/apply`    | 提交申请 |
| GET    | `/api/audit/my`       | 我的申请列表 |
| DELETE | `/api/audit/:id`      | 2 分钟内撤回 |
| GET    | `/api/messages`       | 站内信 |
| WS     | `/api/ws`             | 实时通知（带心跳） |

### 管理员（Casbin）
| 方法 | 路径 | 说明 |
|------|------|------|
| CRUD | `/api/users`, `/api/roles`, `/api/permissions` | RBAC 管理 |
| CRUD | `/api/resources`                                 | 资源管理 |
| GET    | `/api/audit`         | 审核队列 |
| POST   | `/api/audit/:id/approve` / `reject` | 审核操作 |

完整错误码、权限策略、请求/响应示例见 [TECHNICAL_DOCUMENTATION.md §15](./TECHNICAL_DOCUMENTATION.md)。

---

## 数据库设计

GORM AutoMigrate 自动创建，核心表：

| 表 | 说明 |
|----|------|
| `users` / `roles` / `permissions` | RBAC 三要素 |
| `user_roles` / `role_permissions` | 多对多关联 |
| `casbin_rule` | Casbin 策略（g、p、p2 规则） |
| `api_resources` | 大模型 API 资源清单（聊天/代码/图像/语音/嵌入） |
| `api_audit_applications` | 审核申请（含乐观锁 `version`） |
| `sys_messages` | 系统消息（含 MQ 投递标记 `mq_delivered`） |

索引策略：用户邮箱唯一；审核表按 `status`、`applicant_id` 复合索引；Casbin 策略按 `p_type` 索引。ER 图与字段详解见 [TECHNICAL_DOCUMENTATION.md §14](./TECHNICAL_DOCUMENTATION.md)。

---

## 部署

- **Docker Compose**：见 [快速开始](#快速开始docker-compose)（后端 Alpine 多阶段构建约 20 MB 运行时；前端 Vite 构建 + Nginx 托管）。
- **生产注意**：
  1. 替换 `jwt.secret`（配置文件或环境变量）
  2. `server.mode=release`
  3. 修改默认账号密码
  4. Postgres / Redis / RabbitMQ 启用持久卷与密码
  5. Nginx 添加 HTTPS、Gzip、代理超时
- **健康检查**：`GET /health`。

详细 Dockerfile、Nginx 配置、风险清单见 [TECHNICAL_DOCUMENTATION.md §16–22](./TECHNICAL_DOCUMENTATION.md)。

---

## 技术文档

本仓库包含一份 **22 章、深度代码级解析** 的技术文档：

📘 [TECHNICAL_DOCUMENTATION.md](./TECHNICAL_DOCUMENTATION.md)

覆盖：架构、fx 依赖注入、两级缓存、布隆过滤器、三层限流、Casbin RBAC、审核状态机、RabbitMQ Fanout + 死信、WebSocket Hub + Redis PubSub、前端 Pinia/Vue Router、数据库 ER、API、Docker、监控、常见问题排查。

---

## 许可证

MIT License.

本项目用于学习与演示生产级 Go 后端架构设计思想。
