# RBAC权限管理系统 + 大模型API资源审核平台 技术文档 v4.0.0

---

## 目录

- [第1章 项目概述](#第1章-项目概述)
  - [1.1 项目简介](#11-项目简介)
  - [1.2 核心功能](#12-核心功能)
  - [1.3 技术选型](#13-技术选型)
  - [1.4 版本说明](#14-版本说明)
- [第2章 系统架构](#第2章-系统架构)
  - [2.1 整体架构设计](#21-整体架构设计)
  - [2.2 依赖注入架构（uber-go/fx）](#22-依赖注入架构uber-gofx)
  - [2.3 后端分层架构](#23-后端分层架构)
  - [2.4 前端架构设计](#24-前端架构设计)
  - [2.5 请求处理完整流程](#25-请求处理完整流程)
  - [2.6 消息通知数据流](#26-消息通知数据流)
- [第3章 目录结构详解](#第3章-目录结构详解)
  - [3.1 项目根目录](#31-项目根目录)
  - [3.2 后端目录结构](#32-后端目录结构)
  - [3.3 前端目录结构](#33-前端目录结构)
- [第4章 uber-go/fx依赖注入框架详解](#第4章-uber-gofx依赖注入框架详解)
  - [4.1 fx框架核心概念](#41-fx框架核心概念)
  - [4.2 Module()函数详解](#42-module函数详解)
  - [4.3 Provide组件构造顺序](#43-provide组件构造顺序)
  - [4.4 Invoke启动钩子详解](#44-invoke启动钩子详解)
  - [4.5 Lifecycle生命周期管理](#45-lifecycle生命周期管理)
  - [4.6 优雅关停顺序](#46-优雅关停顺序)
- [第5章 缓存架构设计](#第5章-缓存架构设计)
  - [5.1 两级缓存架构](#51-两级缓存架构)
  - [5.2 缓存分类与TTL策略](#52-缓存分类与ttl策略)
  - [5.3 缓存三大问题防护](#53-缓存三大问题防护)
  - [5.4 cache.Client核心结构体详解](#54-cacheclient核心结构体详解)
  - [5.5 Fetch方法完整流程代码级解析](#55-fetch方法完整流程代码级解析)
  - [5.6 L1本地缓存实现（local.go）](#56-l1本地缓存实现localgo)
  - [5.7 布隆过滤器实现（bloom.go）](#57-布隆过滤器实现bloomgo)
  - [5.8 Redis熔断器实现（circuit.go）](#58-redis熔断器实现circuitgo)
  - [5.9 缓存预热机制](#59-缓存预热机制)
  - [5.10 缓存失效策略](#510-缓存失效策略)
- [第6章 网关限流架构](#第6章-网关限流架构)
  - [6.1 三层限流模型](#61-三层限流模型)
  - [6.2 令牌桶算法原理与Lua脚本](#62-令牌桶算法原理与lua脚本)
  - [6.3 RateLimiter结构体详解](#63-ratelimiter结构体详解)
  - [6.4 限流熔断器实现](#64-限流熔断器实现)
  - [6.5 本地内存令牌桶降级实现](#65-本地内存令牌桶降级实现)
  - [6.6 限流Key设计与路径归一化](#66-限流key设计与路径归一化)
  - [6.7 白名单机制](#67-白名单机制)
  - [6.8 限流统计与后台清理](#68-限流统计与后台清理)
- [第7章 后端架构逐模块详解](#第7章-后端架构逐模块详解)
  - [7.1 入口层 main.go](#71-入口层-maingo)
  - [7.2 应用初始化层 app/init.go](#72-应用初始化层-appinitgo)
  - [7.3 模块注册层 app/module.go（完整代码解析）](#73-模块注册层-appmodulego完整代码解析)
  - [7.4 配置层 config/config.go](#74-配置层-configconfiggo)
  - [7.5 模型层 model/model.go（完整代码解析）](#75-模型层-modelmodelgo完整代码解析)
  - [7.6 路由层 router/router.go（完整代码解析）](#76-路由层-routerroutergo完整代码解析)
  - [7.7 中间件层](#77-中间件层)
  - [7.8 数据访问层 Repository](#78-数据访问层-repository)
  - [7.9 业务逻辑层 Service](#79-业务逻辑层-service)
  - [7.10 处理器层 Handler](#710-处理器层-handler)
- [第8章 公共包 pkg 详解](#第8章-公共包-pkg-详解)
  - [8.1 response 统一响应封装](#81-response-统一响应封装)
  - [8.2 jwt JWT工具包](#82-jwt-jwt工具包)
  - [8.3 casbin 权限引擎](#83-casbin-权限引擎)
  - [8.4 cache 两级缓存包](#84-cache-两级缓存包)
  - [8.5 mq RabbitMQ消息队列（完整代码解析）](#85-mq-rabbitmq消息队列完整代码解析)
  - [8.6 ws WebSocket实时通信（完整代码解析）](#86-ws-websocket实时通信完整代码解析)
- [第9章 Casbin权限模型详解](#第9章-casbin权限模型详解)
  - [9.1 RBAC权限模型设计](#91-rbac权限模型设计)
  - [9.2 模型定义文件](#92-模型定义文件)
  - [9.3 权限校验完整流程](#93-权限校验完整流程)
  - [9.4 策略管理与初始化](#94-策略管理与初始化)
- [第10章 审核系统业务架构](#第10章-审核系统业务架构)
  - [10.1 业务流程概述](#101-业务流程概述)
  - [10.2 状态机设计](#102-状态机设计)
  - [10.3 2分钟撤回机制（代码级详解）](#103-2分钟撤回机制代码级详解)
  - [10.4 乐观锁并发控制](#104-乐观锁并发控制)
  - [10.5 事务边界设计](#105-事务边界设计)
  - [10.6 AuditService核心方法逐行解析](#106-auditservice核心方法逐行解析)
- [第11章 RabbitMQ消息队列](#第11章-rabbitmq消息队列)
  - [11.1 架构设计：Fanout广播模式](#111-架构设计fanout广播模式)
  - [11.2 Exchange与Queue规划](#112-exchange与queue规划)
  - [11.3 Client结构体详解](#113-client结构体详解)
  - [11.4 连接建立与拓扑声明](#114-连接建立与拓扑声明)
  - [11.5 可靠投递七重保障](#115-可靠投递七重保障)
  - [11.6 死信队列设计](#116-死信队列设计)
  - [11.7 自动重连机制](#117-自动重连机制)
  - [11.8 消息消费与手动ACK](#118-消息消费与手动ack)
  - [11.9 定时对账补发机制](#119-定时对账补发机制)
- [第12章 WebSocket实时通信](#第12章-websocket实时通信)
  - [12.1 Hub架构设计](#121-hub架构设计)
  - [12.2 Client与Hub结构体详解](#122-client与hub结构体详解)
  - [12.3 连接注册与管理](#123-连接注册与管理)
  - [12.4 心跳与超时机制（ReadPump/WritePump）](#124-心跳与超时机制readpumpwritepump)
  - [12.5 Redis PubSub跨实例广播](#125-redis-pubsub跨实例广播)
  - [12.6 幂等去重机制（LRU缓存）](#126-幂等去重机制lru缓存)
  - [12.7 消息推送策略](#127-消息推送策略)
  - [12.8 单用户连接数限制](#128-单用户连接数限制)
- [第13章 前端架构详解](#第13章-前端架构详解)
  - [13.1 技术选型与版本](#131-技术选型与版本)
  - [13.2 应用入口 main.ts](#132-应用入口-maints)
  - [13.3 根组件 App.vue](#133-根组件-appvue)
  - [13.4 状态管理 Pinia（user.ts详解）](#134-状态管理-piniausersts详解)
  - [13.5 路由系统 Vue Router（index.ts详解）](#135-路由系统-vue-routerindexts详解)
  - [13.6 Axios请求封装 request.ts](#136-axios请求封装-requestts)
  - [13.7 API接口层](#137-api接口层)
  - [13.8 布局组件 Layout.vue](#138-布局组件-layoutvue)
  - [13.9 双角色UI设计体系](#139-双角色ui设计体系)
  - [13.10 WebSocket客户端（websocket.ts详解）](#1310-websocket客户端websocketts详解)
  - [13.11 页面组件详解](#1311-页面组件详解)
  - [13.12 错误页面系统](#1312-错误页面系统)
- [第14章 数据库设计](#第14章-数据库设计)
  - [14.1 ER关系图](#141-er关系图)
  - [14.2 users 用户表（字段逐列详解）](#142-users-用户表字段逐列详解)
  - [14.3 roles 角色表](#143-roles-角色表)
  - [14.4 permissions 权限表](#144-permissions-权限表)
  - [14.5 user_roles 用户角色关联表](#145-user_roles-用户角色关联表)
  - [14.6 role_permissions 角色权限关联表](#146-role_permissions-角色权限关联表)
  - [14.7 casbin_rule Casbin策略表](#147-casbin_rule-casbin策略表)
  - [14.8 api_resources 资源表](#148-api_resources-资源表)
  - [14.9 api_audit_applications 审核申请表（含乐观锁）](#149-api_audit_applications-审核申请表含乐观锁)
  - [14.10 sys_messages 系统消息表（含MQ投递标记）](#1410-sys_messages-系统消息表含mq投递标记)
  - [14.11 索引设计说明（GIN/B-tree复合索引）](#1411-索引设计说明ginb-tree复合索引)
- [第15章 API接口文档](#第15章-api接口文档)
  - [15.1 通用说明](#151-通用说明)
  - [15.2 公开接口（无需认证）](#152-公开接口无需认证)
  - [15.3 基础认证接口（JWT即可）](#153-基础认证接口jwt即可)
  - [15.4 管理员接口（需Casbin权限）](#154-管理员接口需casbin权限)
  - [15.5 错误码说明](#155-错误码说明)
- [第16章 部署指南](#第16章-部署指南)
  - [16.1 Docker Compose一键部署](#161-docker-compose一键部署)
  - [16.2 本地开发环境搭建](#162-本地开发环境搭建)
  - [16.3 Nginx配置详解](#163-nginx配置详解)
  - [16.4 服务健康检查](#164-服务健康检查)
- [第17章 开发规范与指南](#第17章-开发规范与指南)
  - [17.1 新增接口开发流程](#171-新增接口开发流程)
  - [17.2 缓存使用规范](#172-缓存使用规范)
  - [17.3 代码注释规范](#173-代码注释规范)
  - [17.4 错误处理规范](#174-错误处理规范)
  - [17.5 测试指南](#175-测试指南)
- [第18章 生产环境风险规避清单](#第18章-生产环境风险规避清单)
- [第19章 配置文件详解](#第19章-配置文件详解)
  - [19.1 config.yaml完整配置说明](#191-configyaml-完整配置说明)
  - [19.2 环境变量覆盖](#192-环境变量覆盖)
- [第20章 Docker部署详解](#第20章-docker部署详解)
  - [20.1 后端Dockerfile](#201-后端dockerfile)
  - [20.2 前端Dockerfile（多阶段构建）](#202-前端dockerfile多阶段构建)
  - [20.3 docker-compose.yml完整配置](#203-docker-composeyml-完整配置)
- [第21章 监控与可观测性](#第21章-监控与可观测性)
  - [21.1 日志规范](#211-日志规范)
  - [21.2 关键监控指标](#212-关键监控指标)
  - [21.3 /health健康检查端点](#213-health-健康检查端点)
- [第22章 常见问题排查指南](#第22章-常见问题排查指南)
  - [22.1 403 Forbidden问题排查](#221-403-forbidden问题排查)
  - [22.2 WebSocket连接失败排查](#222-websocket连接失败排查)
  - [22.3 消息收不到实时通知排查](#223-消息收不到实时通知排查)
  - [22.4 缓存问题排查](#224-缓存问题排查)
  - [22.5 服务启动失败排查](#225-服务启动失败排查)

---

## 第1章 项目概述

### 1.1 项目简介

本项目是一个**生产级RBAC权限管理系统 + 大模型API资源审核平台**，基于Go (Gin) + Vue 3 + PostgreSQL + Redis + RabbitMQ构建。系统采用前后端分离架构，使用**uber-go/fx**作为依赖注入框架管理组件生命周期，集成Casbin作为权限控制引擎，提供完整的用户/角色/权限管理、资源申请审核、实时消息通知、多级缓存防护、网关限流、优雅关停等生产级特性。

**设计理念**：
- **严格分层**：Handler → Service → Repository 三层架构，职责清晰
- **依赖注入**：uber-go/fx自动管理组件构造顺序和生命周期，支持优雅关停
- **防御式编程**：缓存三大问题防护、三层限流、熔断器、乐观锁
- **消息驱动**：PostgreSQL事务先提交，再发RabbitMQ，确保数据一致性（无幽灵消息）
- **实时推送**：RabbitMQ Fanout → Redis PubSub → WebSocket 三级消息分发
- **Go 1.26特性**：全面使用log/slog结构化日志、signal.NotifyContext优雅关停（由fx内置实现）
- **代码注释**：所有后端代码包含详细中文注释，解释设计意图和关键逻辑

### 1.2 核心功能

| 功能模块 | 详细功能描述 |
|---------|------------|
| **用户认证** | 用户注册、登录、JWT Token认证、bcrypt密码加密、Token过期自动刷新 |
| **用户管理** | 用户CRUD、用户启用/禁用、用户角色分配、用户信息脱敏返回（不含密码） |
| **角色管理** | 角色CRUD、角色权限分配、角色状态管理、内置admin/user角色 |
| **权限管理** | 权限CRUD、权限分类（API级）、权限与路由自动同步、启动时种子数据初始化 |
| **Casbin权限控制** | RBAC模型、接口级细粒度权限校验、策略持久化到PostgreSQL、自动重载 |
| **资源清单管理** | 大模型API资源CRUD、资源分类（聊天/代码/图像/语音/嵌入）、资源状态管理 |
| **资源申请审核** | 用户提交API使用申请、管理员审核（通过/驳回）、2分钟可撤回窗口 |
| **2分钟撤回机制** | 服务端Redis TTL控制撤回窗口（前端不可篡改）、PG时间戳兜底、乐观锁防并发 |
| **实时消息通知** | WebSocket实时弹窗、RabbitMQ异步解耦、Redis PubSub跨网关广播、站内信持久化 |
| **离线消息兜底** | 管理员/用户登录时拉取未读消息、PG作为唯一可信持久层、定时对账补发 |
| **两级缓存架构** | L1本地内存缓存（sync.Map）+ L2 Redis分布式缓存，完整防穿透/击穿/雪崩 |
| **布隆过滤器** | 启动预热合法ID，拦截不存在数据的查询请求（防穿透第一层） |
| **网关三层限流** | IP级（30req/s）+ 接口级（100req/s）+ 用户级（20req/s）令牌桶限流 |
| **熔断器保护** | Redis故障时自动降级为本地内存限流/缓存，自动探测恢复（三态熔断） |
| **仪表盘统计** | 管理员：用户数/角色数/权限数/待审核数；普通用户：个人申请统计 |
| **双角色UI** | 管理员端：琥珀金色系，完整管理菜单；普通用户端：靛蓝系，功能菜单 |
| **错误处理体系** | 统一JSON响应格式、HTTP状态码与code严格匹配、Panic恢复中间件、自定义错误页面 |
| **优雅关停** | fx.Lifecycle管理HTTP Server.Shutdown()等待活跃请求、限流器/Hub/MQ资源按序清理 |

### 1.3 技术选型

#### 后端技术栈

| 技术 | 版本 | 详细用途 |
|-----|-----|---------|
| Go | 1.26.4 | 后端开发语言，使用log/slog等新特性 |
| Gin | v1.11.0 | Web框架，HTTP路由、中间件链、Context管理 |
| GORM | v1.31.0 | ORM框架，PostgreSQL操作、关联预加载、事务、自动迁移 |
| Casbin | v2.11.0 | 权限控制引擎，RBAC模型，gorm-adapter持久化策略 |
| uber-go/fx | - | **依赖注入容器**，自动管理组件构造顺序、生命周期、优雅关停 |
| PostgreSQL | 15 | 关系型数据库，唯一可信持久层，存储业务数据和消息 |
| Redis | 7 | 分布式缓存、PubSub跨网关广播、撤回窗口TTL控制、限流计数器 |
| RabbitMQ | 3 | 消息队列，异步解耦、通知分发、削峰填谷，Fanout广播模式 |
| gorilla/websocket | - | WebSocket实现，实时弹窗推送、心跳保活 |
| go-redis | v8 | Redis客户端，支持PubSub、Lua脚本、分布式锁 |
| golang-jwt | v5 | JWT Token生成、解析、验证，HS256签名 |
| google/uuid | - | UUID生成，对外暴露ID，避免主键泄露 |
| golang.org/x/sync | - | singleflight合并并发请求，防缓存击穿 |
| log/slog | 标准库 | Go 1.26结构化JSON日志，分级日志（Info/Warn/Error） |

#### 前端技术栈

| 技术 | 版本 | 详细用途 |
|-----|-----|---------|
| Vue | 3.5.x | 渐进式前端框架，Composition API + `<script setup>`语法 |
| Vite | 7.x | 下一代构建工具，开发服务器HMR极速热更新 |
| TypeScript | 5.x | 类型安全的JavaScript超集，编译时类型检查 |
| Pinia | 3.x | Vue官方推荐状态管理，替代Vuex，Composition API风格 |
| Vue Router | 4.x | 路由管理，路由守卫、动态路由、权限控制 |
| Element Plus | 2.x | 企业级Vue 3 UI组件库，中文语言包 |
| Axios | 1.x | HTTP客户端，请求/响应拦截器、统一错误处理 |
| Sass | 1.x | CSS预处理器，变量、嵌套、混合宏 |
| @element-plus/icons-vue | - | Element Plus官方图标库 |

#### 基础设施

| 技术 | 版本 | 用途 |
|-----|-----|-----|
| Docker | - | 容器化部署，环境一致性 |
| Docker Compose | - | 多服务编排，一键启动PG/Redis/RabbitMQ/前后端 |
| Nginx | latest | 前端Web服务器、API反向代理、SPA路由支持、错误拦截 |
| PostgreSQL | 15-alpine | 数据库服务 |
| Redis | 7-alpine | 缓存服务、PubSub广播 |
| RabbitMQ | 3-management-alpine | 消息队列服务（含管理界面） |

### 1.4 版本说明

- **文档版本**：v4.0.0
- **最后更新**：2026年6月27日
- **适用版本**：RBAC权限管理系统 + 大模型API资源审核平台 v3.2.0
- **Go版本要求**：>= 1.26.4
- **Node.js版本要求**：>= 18.x
- **依赖注入框架**：uber-go/fx（从v4.0文档开始正式说明）
- **演示账号**：admin/admin123（管理员）、user/user123（普通用户）
- **服务端口**：前端3000，后端8080

---

## 第2章 系统架构

### 2.1 整体架构设计

系统采用**前后端分离 + 事件驱动 + 依赖注入**架构，Nginx作为统一入口，整体分为五层：客户端层、接入层、应用层、消息层、数据层。

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              客户端层 (Client)                                   │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐        │
│  │  Web浏览器    │  │  Postman     │  │  移动端      │  │  第三方系统   │        │
│  │  (Vue3 SPA)  │  │  (API测试)   │  │  (预留)      │  │  (预留)      │        │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘        │
└─────────┼─────────────────┼─────────────────┼─────────────────┼────────────────┘
          │                 │                 │                 │
          └─────────────────┴─────────────────┴─────────────────┘
                                    │ HTTP/HTTPS, WSS
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              接入层 (Nginx)                                      │
│  ┌───────────────────────────────────────────────────────────────────────────┐  │
│  │  /         → try_files $uri /index.html (Vue SPA history模式)             │  │
│  │  /api/*    → proxy_pass http://backend:8080 (API反向代理)                  │  │
│  │  /ws       → proxy_pass http://backend:8080 (WebSocket反向代理)            │  │
│  │  错误页面  → 拦截4xx/5xx返回SPA错误路由                                    │  │
│  └───────────────────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────┬──────────────────────────────────────────┘
                                       │
                    ┌──────────────────┴──────────────────┐
                    │                                     │
                    ▼                                     ▼
┌─────────────────────────────┐           ┌─────────────────────────────────────┐
│       前端应用 (Frontend)    │           │         后端应用 (Backend)           │
│  ┌───────────────────────┐  │           │  ┌───────────────────────────────┐  │
│  │  Vue 3 + TypeScript   │  │           │  │  uber-go/fx 依赖注入容器       │  │
│  │  Pinia State Mgmt     │  │           │  │  组件自动构造 + 生命周期管理   │  │
│  │  Vue Router (守卫)    │  │           │  └───────────────┬───────────────┘  │
│  │  Element Plus UI      │  │           │                  │                   │
│  │  Axios (拦截器)       │  │           │  ┌───────────────▼───────────────┐  │
│  │  WebSocket Client     │  │           │  │  Gin Web Framework            │  │
│  └───────────────────────┘  │           │  │  全局中间件链：                │  │
│                             │           │  │  1. CustomRecovery (Panic恢复)│  │
│                             │           │  │  2. CORS (跨域)               │  │
│                             │           │  │  3. RateLimit (IP+接口限流)   │  │
│                             │           │  │  4. RequestLogger (请求日志)  │  │
│                             │           │  └───────────────┬───────────────┘  │
│                             │           │                  │                   │
│                             │           │  ┌───────────────▼───────────────┐  │
│                             │           │  │  路由层 (三层权限架构)          │  │
│                             │           │  │  1. 公开接口 (/login, /health)│  │
│                             │           │  │  2. 基础认证 (JWTAuth)        │  │
│                             │           │  │  3. 管理员 (CasbinAuth)       │  │
│                             │           │  └───────────────┬───────────────┘  │
│                             │           │                  │                   │
│                             │           │  ┌───────────────▼───────────────┐  │
│                             │           │  │  Handler Layer                │  │
│                             │           │  │  - AuthHandler    认证        │  │
│                             │           │  │  - UserHandler    用户管理    │  │
│                             │           │  │  - RoleHandler    角色管理    │  │
│                             │           │  │  - PermissionHandler 权限    │  │
│                             │           │  │  - DashboardHandler 仪表盘   │  │
│                             │           │  │  - AuditHandler   审核        │  │
│                             │           │  │  - ResourceHandler 资源管理   │  │
│                             │           │  │  - WsHandler      WebSocket   │  │
│                             │           │  └───────────────┬───────────────┘  │
│                             │           │                  │                   │
│                             │           │  ┌───────────────▼───────────────┐  │
│                             │           │  │  Service Layer (业务逻辑+缓存)  │  │
│                             │           │  │  - AuthService                │  │
│                             │           │  │  - UserService                │  │
│                             │           │  │  - RoleService                │  │
│                             │           │  │  - PermissionService          │  │
│                             │           │  │  - DashboardService           │  │
│                             │           │  │  - AuditService (审核+消息)   │  │
│                             │           │  │  - ResourceService            │  │
│                             │           │  └───────────────┬───────────────┘  │
│                             │           │                  │                   │
│                             │           │  ┌───────────────▼───────────────┐  │
│                             │           │  │  Repository Layer (数据访问)    │  │
│                             │           │  │  - UserRepo / RoleRepo        │  │
│                             │           │  │  - PermissionRepo             │  │
│                             │           │  │  - AuditRepository (审核+消息)│  │
│                             │           │  │  - ResourceRepo               │  │
│                             │           │  │  - Casbin策略操作             │  │
│                             │           └──────────────────┼──────────────────┘
│                             │                              │
└─────────────────────────────┘                              │
                                                             │
        ┌────────────────────────────────────────────────────┼───────────────────────────────┐
        │                       消息层                        │                               │
        │  ┌──────────────────────┐  ┌──────────────────────┐▼──────┐  ┌─────────────────┐   │
        │  │  RabbitMQ (Fanout)   │  │  Redis PubSub        │缓存   │  │  PostgreSQL     │   │
        │  │  - 异步解耦分发       │  │  - 跨网关实例广播    │&限流  │  │  唯一可信持久层  │   │
        │  │  - 手动ACK/死信队列   │  │  - 撤回窗口TTL       │      │  │  (业务+消息)    │   │
        │  │  - 定时对账补发       │  │  - 分布式锁限流      │      │  └─────────────────┘   │
        │  └──────────┬───────────┘  └──────────┬───────────┘      │                         │
        │             │                         │                  │                         │
        │             └─────────────────────────┘                  │                         │
        │                       │                                  │                         │
        │                       ▼                                  ▼                         │
        │            ┌──────────────────────┐           ┌─────────────────────┐             │
        │            │  WebSocket Hub       │◄──────────┤ L1本地/L2Redis缓存  │             │
        │            │  - 连接管理/心跳     │           │  防穿透/击穿/雪崩    │             │
        │            │  - 幂等去重(LRU)     │           └─────────────────────┘             │
        │            │  - 管理员房间/用户直推│                                                 │
        │            └──────────────────────┘                                                 │
        └────────────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
                          ┌──────────────────┐
                          │  浏览器实时弹窗   │
                          └──────────────────┘
```

### 2.2 依赖注入架构（uber-go/fx）

项目使用 **uber-go/fx** 作为依赖注入（DI）框架，这是与v3文档最大的区别。fx框架负责：

1. **自动构造组件**：根据构造函数的参数类型自动解析依赖，按顺序构造
2. **生命周期管理**：通过fx.Lifecycle管理组件的启动（OnStart）和关闭（OnStop）
3. **优雅关停**：收到中断信号时，按OnStop注册逆序关闭所有资源
4. **可选依赖**：Redis/RabbitMQ失败时返回nil，服务降级为PG-only模式仍可运行

**fx依赖图（自底向上）**：
```
config.Config (从YAML/环境变量加载)
    ↓
*gorm.DB (PostgreSQL连接)
    ↓
*redis.Client (可选，失败为nil)
*mq.Client (RabbitMQ，可选，失败为nil)
    ↓
*cache.Client (两级缓存，依赖rdb)
*ws.Hub (WebSocket Hub，依赖rdb做PubSub)
    ↓
Repositories (UserRepo/RoleRepo/... 依赖db)
    ↓
Services (AuthService/UserService/... 依赖Repos+Cache+MQ+WS)
    ↓
Handlers (AuthHandler/UserHandler/... 依赖Services)
    ↓
router.Handlers (聚合所有Handler)
*router.EngineWrapper (Gin引擎+限流器)
    ↓
http.Server (通过fx.Lifecycle启动和优雅关闭)
```

### 2.3 后端分层架构

后端严格遵循 **Handler → Service → Repository → Database/Cache** 四层分层架构，由fx自动管理依赖注入。各层职责边界清晰：

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     Handler Layer (处理器层)                            │
│  职责: 参数绑定校验、HTTP协议处理、调用Service、封装统一响应              │
│  约束: 不包含任何业务逻辑，不直接操作DB/Cache，只做HTTP相关处理          │
│  组件: AuthHandler, UserHandler, RoleHandler, PermissionHandler,       │
│        DashboardHandler, AuditHandler, ResourceHandler, WsHandler      │
└──────────────────────────────┬──────────────────────────────────────────┘
                               │  调用业务方法 (传Context + DTO)
                               ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                     Service Layer (业务逻辑层)                          │
│  职责: 核心业务逻辑、事务管理、缓存策略、DTO转换、消息发送、乐观锁        │
│  约束: 可独立单元测试，不直接处理HTTP请求，不写原始SQL                    │
│  组件: AuthService, UserService, RoleService, PermissionService,       │
│        DashboardService, AuditService(审核+MQ+WS), ResourceService      │
│  关键: 事务内不做Redis/MQ网络IO；先提交PG事务，再发MQ消息                │
└──────────────────────────────┬──────────────────────────────────────────┘
                               │  数据访问 (传Entity/Query，支持事务)
                               ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                  Repository Layer (数据访问层)                           │
│  职责: 数据库CRUD封装、Casbin策略操作、GORM查询、事务Begin/Commit/Rollback│
│  约束: 只做数据访问，不包含业务逻辑，返回Entity/error                     │
│  组件: UserRepository, RoleRepository, PermissionRepository,           │
│        AuditRepository(审核+消息), ResourceRepository                   │
└──────────────────────────────┬──────────────────────────────────────────┘
                               │  数据库/缓存操作
              ┌────────────────┴────────────────┐
              ▼                                 ▼
┌─────────────────────────────┐   ┌─────────────────────────────┐
│  PostgreSQL (主存储)         │   │  Cache (L1本地 + L2 Redis)  │
│  - 业务数据 (users/roles/..)│   │  - L1: sync.Map 进程内缓存  │
│  - Casbin策略表             │   │  - L2: Redis 分布式缓存     │
│  - 审核申请 + 系统消息       │   │  - 防穿透/击穿/雪崩完整防护  │
│  - 唯一可信数据源            │   │  - 热点数据预热             │
└─────────────────────────────┘   └─────────────────────────────┘
```

### 2.4 前端架构设计

前端采用Vue 3 Composition API + TypeScript，Pinia状态管理，Vue Router路由守卫控制权限，Axios拦截器统一处理请求响应。

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         View Layer (页面视图)                           │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────┐  │
│  │  Login   │ │ Layout   │ │ Dashboard│ │ 用户管理 │ │ 角色/权限    │  │
│  │  登录页  │ │ 布局容器 │ │ 仪表盘   │ │ Users    │ │ Roles/Perms  │  │
│  └──────────┘ └────┬─────┘ └──────────┘ └──────────┘ └──────────────┘  │
│  ┌──────────┐ ┌────┴─────┐ ┌──────────┐ ┌──────────┐ ┌──────────────┐  │
│  │ 资源清单 │ │ 申请资源 │ │ 我的申请 │ │ 审核管理 │ │ 消息中心     │  │
│  │ResourceLi│ │ApplyRes  │ │MyApp     │ │AuditList │ │ Messages     │  │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────────┘  │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │  Error Pages: 401/403/404/500 错误页面                          │  │
│  └──────────────────────────────────────────────────────────────────┘  │
└──────────────────────────────┬──────────────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                       Component Layer (组件层)                          │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │  业务组件: NotificationBell(通知铃铛)                             │  │
│  │  布局组件: Sidebar(侧边栏), Header(顶栏)                         │  │
│  │  Element Plus组件: ElTable, ElForm, ElDialog, ElMenu...          │  │
│  └──────────────────────────────────────────────────────────────────┘  │
└──────────────────────────────┬──────────────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                        State Layer (Pinia状态)                          │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │  user store: token, userInfo, roles, isAdmin, 登录登出逻辑       │  │
│  │  持久化: localStorage存储token，刷新页面自动恢复登录态            │  │
│  └──────────────────────────────────────────────────────────────────┘  │
└──────────────────────────────┬──────────────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                        API Layer (接口层)                               │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐   │
│  │ request.ts   │ │   auth.ts    │ │   user.ts    │ │  rbac.ts     │   │
│  │ Axios实例封装 │ │ 登录/注册/   │ │ 用户管理API  │ │ 角色/权限API │   │
│  │ 请求/响应拦截│ │ 用户信息     │ │              │ │              │   │
│  └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘   │
│  ┌──────────────┐ ┌──────────────────────────────────────────────────┐  │
│  │ resource.ts  │ │ audit.ts        审核/消息/资源相关API            │  │
│  │ 资源相关API  │ │                                                    │  │
│  └──────────────┘ └──────────────────────────────────────────────────┘  │
└──────────────────────────────┬──────────────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                       Router Layer (路由层)                             │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │  Vue Router: 路由守卫 beforeEach，动态菜单，权限路由              │  │
│  │  白名单: /login, /401, /403, /404, /500                        │  │
│  │  登录守卫: 无token跳转/login，有token预加载用户信息              │  │
│  │  管理员守卫: 非admin访问管理员路由跳转/403                       │  │
│  └──────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────┘
```

### 2.5 请求处理完整流程

一个API请求从浏览器到返回响应经过的完整链路：

```
1. 用户在浏览器点击操作
   ↓
2. Vue组件调用API层方法（如 auditApi.submit(data)）
   ↓
3. Axios请求拦截器自动添加 Authorization: Bearer <token> 头
   ↓
4. Nginx接收请求
   ├─ /api/* 反向代理到 backend:8080
   └─ WebSocket /ws 升级连接
   ↓
5. Gin Engine接收请求（fx已完成所有组件初始化）
   ↓
6. 【全局中间件链按顺序执行】
   ├─ ① CustomRecovery(): defer recover()捕获panic，返回500 JSON
   ├─ ② CORS(): 设置Access-Control-Allow-*跨域头
   ├─ ③ RateLimit(limiter):
   │    ├─ 白名单检查（/health, /ws开头跳过）
   │    ├─ IP级限流: rate:ip:{client_ip} 令牌桶 30req/s, burst 60
   │    │    └─ Redis可用走Lua脚本，不可用降级本地令牌桶（熔断器控制）
   │    └─ 接口级限流: rate:api:{METHOD}:{normalized_path} 100req/s, burst 200
   └─ ④ requestLogger(): 记录请求日志，按状态码分级（Info/Warn/Error）
   ↓
7. 路由匹配（router.go中fx.Invoke调用registerRoutes注册）
   ├─ NoRoute → 404 JSON
   ├─ NoMethod → 405 JSON
   └─ 匹配到对应Handler
   ↓
8. 【路由分组中间件】
   ├─ 公开接口组 → 直接到Handler
   ├─ auth组 → JWTAuth() + UserRateLimit()
   │    ├─ JWTAuth(): 解析Bearer Token，验证JWT签名有效期
   │    │    ├─ 失败 → 401
   │    │    └─ 成功 → c.Set("user_id"/"username"/"roles")
   │    └─ UserRateLimit(): 用户级限流 rate:user:{username} 20req/s, burst 40
   └─ admin组 → CasbinAuth()
        ├─ 获取sub=username, obj=path, act=method
        ├─ casbin.Enforcer.Enforce(sub, obj, act)
        ├─ 失败 → 403
        └─ 成功 → 继续
   ↓
9. Handler层处理（由fx构造并注入）
   ├─ c.ShouldBindJSON() / ShouldBindQuery() 绑定参数（binding标签校验）
   ├─ 参数校验失败 → 400
   ├─ 调用Service层方法，传c.Request.Context()
   └─ 接收结果
   ↓
10. Service层业务逻辑（以提交审核申请为例，fx注入所有依赖）
    ├─ 查询申请人信息（通过UserRepo）
    ├─ 构造AuditApplication实体（UUID、状态pending）
    ├─ 查询所有管理员ID（通过UserRepo）
    ├─ 【开启PG事务】tx := auditRepo.BeginTx()
    │    ├─ 保存申请到api_audit_applications
    │    └─ 批量保存管理员消息到sys_messages
    ├─ 【提交事务】tx.Commit()
    │    （注意：事务内不做Redis/MQ操作，这是关键原则）
    ├─ Redis设置2分钟撤回TTL: cache:audit:withdraw:{app_id} = 1, EX 120
    ├─ 失效相关缓存（cache.Client.Invalidate*方法）: 待审数缓存、消息未读数缓存、申请列表缓存
    ├─ 【发送MQ消息】（事务提交后才发，避免幽灵消息）
    │    └─ mqClient.Publish() → Fanout Exchange audit.fanout
    └─ 【本地WS推送】wsHub.SendToAdminsLocal()（快速给本地在线管理员，不等待MQ）
   ↓
11. Repository层数据访问
    ├─ GORM链式查询: Where/Preload/Order/Limit/Offset
    ├─ 乐观锁: UPDATE ... WHERE version = ? AND id = ?
    └─ 事务操作: tx.Create/tx.Where(tx).Update
   ↓
12. 缓存读取/写入流程（跨层调用cache.Client.Fetch，fx注入）
    ├─ L1本地缓存（sync.Map）命中 → 直接返回
    ├─ L2 Redis命中 → 回填L1后返回
    ├─ 空值标记（__NULL__）→ 返回not found
    ├─ 逻辑过期 → 返回旧值+异步重建（rebuildCh通道）
    ├─ singleflight合并同key并发请求（golang.org/x/sync/singleflight）
    ├─ Redis SETNX分布式锁（多实例防击穿）
    ├─ double-check（加锁后再查一次缓存）
    └─ 查DB → 写缓存（TTL±10%抖动防雪崩）→ 回填布隆过滤器
   ↓
13. 返回响应（response包统一格式）
    ├─ response.Success(c, data) → HTTP 200, {code:200, data:...}
    └─ response.Fail(c, code, msg) → HTTP状态码=code, {code, message}
   ↓
14. MQ消费（异步流程，不阻塞HTTP响应，每个fx实例都有独立队列）
    ├─ 所有实例的独占队列 audit.instance.{hostname}-xxxx 收到消息（Fanout广播）
    ├─ 手动ACK消费（消费成功才Ack，失败Nack重试→死信）
    ├─ 根据message_id查PG获取完整消息（MQ只传轻量ID）
    ├─ wsHub.SendToUserLocal/SendToAdminsLocal → 推送给本地连接
    └─ 更新消息mq_delivered = true
   ↓
15. WebSocket推送
    ├─ Hub事件循环收到消息（通过register/unregister/broadcast/direct/room channel）
    ├─ 幂等去重检查（LRU缓存5分钟窗口）
    ├─ 写入Client.Send channel
    └─ Client.WritePump() 发送到浏览器
   ↓
16. 前端WebSocket收到消息（websocket.ts）
    ├─ 去重判断（根据notification.id）
    ├─ ElNotification弹出通知
    ├─ 更新Pinia store中的未读数
    └─ 如果在消息列表页，自动刷新列表
   ↓
17. 优雅关停（fx收到os.Interrupt信号触发）
    ├─ HTTP Server.Shutdown() 等待活跃请求完成（10s超时）
    ├─ AuditService.Stop() 停止定时补发协程
    ├─ wsHub.Stop() 关闭所有WebSocket连接
    ├─ mqClient.Close() 等待消费者完成在途消息
    ├─ engine.Stop() 停止限流器后台清理协程
    └─ cacheClient.Close() 停止异步重建协程，关闭Redis连接
```

### 2.6 消息通知数据流

实时通知是本系统的核心复杂链路，采用 **PG事务先提交 → MQ异步分发 → Redis PubSub跨实例 → WS本地推送** 架构：

```
【用户提交申请】
     │
     ├─ ① [PG事务] 写入api_audit_applications (pending)
     ├─ ② [PG事务] 批量写入sys_messages给每个管理员（is_read=false, mq_delivered=false）
     ├─ ③ [事务COMMIT] ← 关键！消息先落盘，确保不丢（无幽灵消息）
     │
     ├─ ④ [Redis] SETEX cache:audit:withdraw:{id} 120s (撤回窗口)
     ├─ ⑤ [Redis] DEL 待审数缓存 + SCAN DEL 相关列表缓存
     │
     ├─ ⑥ [MQ] Publish 轻量消息 {message_id, target_type: "admins"}
     │         到 audit.fanout (Fanout Exchange，每个实例都收到)
     │
     ├─ ⑦ [本地WS] 直接wsHub.SendToAdminsLocal() 快速通知本实例管理员（快速路径）
     │
     └─ HTTP响应返回给用户 (不等待MQ/WS完成，低延迟)

【MQ消费（每个fx启动的实例都执行）】
     │
     ├─ 实例1的独占队列 audit.instance.hostname-xxxx 收到消息
     ├─ 实例2的独占队列 audit.instance.hostname-yyyy 收到消息
     ├─ ...
     │
     ├─ 每个实例:
     │    ├─ 根据message_id查PG sys_messages表获取详情
     │    ├─ wsHub.SendToAdminsLocal(notification) 推给本地管理员连接
     │    ├─ Ack消息（消费成功确认，手动ACK模式）
     │    └─ UPDATE sys_messages SET mq_delivered = true
     │
     └─ 【定时对账 每30s，fx启动时StartRetryScheduler】
          └─ SELECT * FROM sys_messages WHERE mq_delivered = false LIMIT 50
              └─ 重新Publish到MQ（兜底极端情况：MQ当时不可用、消息丢失等）

【在线管理员收到实时弹窗】
     └─ WebSocket notification事件 → ElNotification弹窗 + 铃铛红点+1

【离线管理员兜底】
     └─ 登录时（fx已完成初始化）:
          ├─ GET /api/messages/unread-count 返回未读数+待审数
          └─ 进入消息页: GET /api/messages?unread=true 拉取未读消息
```

---

## 第3章 目录结构详解

### 3.1 项目根目录

```
casbin_demo/
├── backend/                    # 后端Go项目（fx DI框架）
├── frontend/                   # 前端Vue3项目
├── scripts/                    # 测试脚本
│   ├── test-api.ps1           # Windows PowerShell API测试脚本
│   └── test-api.sh            # Linux/Mac shell API测试脚本
├── docker-compose.yml          # Docker Compose编排（PG+Redis+RabbitMQ+前后端）
├── Makefile                    # make命令集合（build/run/test等）
├── init.sql                    # 数据库初始化SQL（备用）
└── TECHNICAL_DOCUMENTATION.md  # 本技术文档
```

### 3.2 后端目录结构

```
backend/
├── cmd/
│   └── server/
│       └── main.go                 # 应用入口：仅一行 fx.New(app.Module()).Run()
├── internal/                       # 内部包（不可被外部项目import）
│   ├── app/
│   │   ├── init.go                # 初始化工具函数：setupSlog/mustLoadConfig/mustInitDB/mustSeedData
│   │   └── module.go              # fx模块注册：Provide+Invoke，所有组件构造和生命周期
│   ├── config/
│   │   └── config.go              # 配置结构体定义、YAML加载、viper环境变量绑定、默认值
│   ├── handler/                   # HTTP处理器层（fx构造，依赖Service）
│   │   ├── auth_handler.go        # 登录/注册/用户信息/消息相关接口
│   │   ├── user_handler.go        # 用户CRUD、分配角色
│   │   ├── role_handler.go        # 角色CRUD、分配权限
│   │   ├── permission_handler.go  # 权限CRUD
│   │   ├── dashboard_handler.go   # 仪表盘统计
│   │   ├── audit_handler.go       # 提交/审核/撤回申请、消息列表/已读
│   │   ├── resource_handler.go    # 资源清单CRUD、可用资源列表
│   │   └── ws_handler.go          # WebSocket连接升级（JWT从query取）
│   ├── middleware/                 # Gin中间件
│   │   ├── jwt.go                 # JWT认证中间件
│   │   ├── casbin.go              # Casbin RBAC权限中间件 + CustomRecovery
│   │   └── ratelimit.go           # 三层限流中间件+熔断器+本地令牌桶+Lua脚本
│   ├── model/
│   │   └── model.go               # 所有数据模型：Entity/DTO/Request/Response/常量
│   ├── repository/                 # 数据访问层（fx构造，依赖*gorm.DB）
│   │   ├── database.go            # PG连接初始化、GORM配置、连接池、AutoMigrate、种子数据
│   │   ├── redis.go               # Redis客户端全局变量
│   │   ├── user_repo.go           # 用户数据访问
│   │   ├── role_repo.go           # 角色+Casbin策略数据访问
│   │   ├── permission_repo.go     # 权限数据访问
│   │   ├── audit_repo.go          # 审核申请+系统消息数据访问（含乐观锁、事务）
│   │   └── resource_repo.go       # API资源数据访问
│   ├── router/
│   │   └── router.go              # 路由注册：三层路由架构、EngineWrapper、优雅关停
│   └── service/                   # 业务逻辑层（fx构造，依赖Repos+Cache+MQ+WS）
│       ├── auth_service.go        # 登录/注册/密码验证/JWT生成
│       ├── user_service.go        # 用户CRUD+缓存+DTO转换
│       ├── role_service.go        # 角色CRUD+Casbin策略同步+缓存
│       ├── permission_service.go  # 权限CRUD+缓存
│       ├── dashboard_service.go   # 仪表盘统计（带缓存）
│       ├── audit_service.go       # 审核申请提交/撤回/审核、MQ/WS推送、定时补发
│       └── resource_service.go    # 资源清单管理+缓存
├── pkg/                            # 公共可复用包（无业务逻辑）
│   ├── response/
│   │   └── response.go            # 统一响应封装：Success/Fail + 快捷错误方法
│   ├── jwt/
│   │   └── jwt.go                 # JWT生成/解析/验证（HS256）
│   ├── casbin/
│   │   ├── casbin.go              # Casbin Enforcer初始化、gorm-adapter持久化
│   │   └── model.conf             # RBAC模型配置文件（备用）
│   ├── cache/
│   │   ├── cache.go               # 缓存客户端：两级缓存+三大问题防护+Fetch核心方法
│   │   ├── local.go               # L1本地内存缓存（sync.Map+TTL+逻辑过期）
│   │   ├── bloom.go               # 布隆过滤器实现（位数组+7哈希）
│   │   └── circuit.go             # Redis熔断器（三态：Closed/Open/Half-Open）
│   ├── mq/
│   │   └── rabbitmq.go            # RabbitMQ客户端：Fanout广播、手动ACK、死信、自动重连
│   └── ws/
│       └── hub.go                 # WebSocket Hub：连接管理、心跳、Redis PubSub、幂等LRU去重
├── tests/                          # 测试文件
│   ├── model_test.go
│   ├── jwt_test.go
│   ├── handler_test.go
│   └── audit_service_test.go
├── config.yaml                     # 后端配置文件（PG/Redis/JWT/MQ等）
├── config-docker.yaml              # Docker环境专用配置
├── Dockerfile                      # 后端Docker镜像构建
├── .dockerignore
├── go.mod                          # Go模块定义
└── go.sum                          # Go依赖校验
```

### 3.3 前端目录结构

```
frontend/
├── src/
│   ├── main.ts                     # 应用入口：创建Vue实例、注册插件（Element Plus/Pinia/Router）
│   ├── App.vue                     # 根组件：<router-view/> + 全局样式
│   ├── env.d.ts                    # TypeScript类型声明（Vite环境）
│   ├── api/                        # API接口层
│   │   ├── request.ts              # Axios实例创建、请求/响应拦截器、401跳转登录
│   │   ├── auth.ts                 # 认证API：login/register/getUserInfo
│   │   ├── user.ts                 # 用户管理API
│   │   ├── rbac.ts                 # 角色/权限管理API
│   │   ├── resource.ts             # 资源清单API
│   │   └── audit.ts                # 审核申请/消息/待审数API
│   ├── assets/
│   │   └── style.scss              # 全局样式、主题变量（管理员琥珀金/用户靛蓝）
│   ├── components/                  # 公共组件
│   │   ├── ErrorPage.vue           # 通用错误页面组件（401/403/404/500共用）
│   │   └── NotificationBell.vue    # 顶部通知铃铛组件（红点+未读数+下拉列表）
│   ├── router/
│   │   └── index.ts                # 路由配置、全局守卫（beforeEach权限控制）
│   ├── store/
│   │   └── user.ts                 # Pinia用户状态Store（token/userInfo/isAdmin）
│   ├── utils/
│   │   └── websocket.ts            # WebSocket客户端封装（心跳/重连/去重/消息分发）
│   └── views/                      # 页面视图
│       ├── Login.vue               # 登录/注册页（双Tab切换）
│       ├── Layout.vue              # 主布局容器（侧边栏+顶栏+内容区，双角色菜单）
│       ├── Dashboard.vue           # 仪表盘（管理员/用户显示不同内容）
│       ├── Users.vue               # 用户管理（管理员）
│       ├── Roles.vue               # 角色管理（管理员）
│       ├── Permissions.vue         # 权限管理（管理员）
│       ├── ResourceList.vue        # 资源清单浏览（所有登录用户）
│       ├── ApplyResource.vue       # 申请资源（普通用户）
│       ├── MyApplications.vue      # 我的申请列表（普通用户）
│       ├── AuditList.vue           # 审核管理（管理员）
│       ├── Messages.vue            # 消息中心（所有用户，时间线样式）
│       └── error/                  # 错误页面目录
│           ├── 401.vue
│           ├── 403.vue
│           ├── 404.vue
│           └── 500.vue
├── public/                         # 静态资源（不经过构建）
├── index.html                      # HTML入口模板
├── vite.config.ts                  # Vite构建配置（代理、路径别名@）
├── tsconfig.json                   # TypeScript配置
├── tsconfig.node.json              # Node环境TS配置
├── package.json                    # 前端依赖
├── package-lock.json
├── nginx.conf                      # 生产环境Nginx配置（反向代理+SPA路由）
├── Dockerfile                      # 前端Docker镜像构建（多阶段构建）
└── .dockerignore
```

---

## 第4章 uber-go/fx依赖注入框架详解

### 4.1 fx框架核心概念

uber-go/fx是一个Go语言的依赖注入框架，核心概念：

| 概念 | 说明 | 在本项目中的应用 |
|-----|------|----------------|
| `fx.Provide` | 注册组件构造函数，fx根据参数类型自动解析依赖 | `provideConfig`, `provideDB`, `NewUserRepository`, `NewAuditService`等 |
| `fx.Invoke` | 注册启动函数，在所有Provide完成后执行 | `setupLogger`, `autoMigrate`, `seedData`, `registerRoutes`, `startHTTPServer` |
| `fx.Lifecycle` | 生命周期管理，通过`lc.Append(fx.Hook{OnStart, OnStop})`注册启动/关闭钩子 | HTTP Server启动/关闭、缓存客户端关闭、WS Hub停止等 |
| `fx.New` | 创建fx应用，传入Module等Option | main.go中`fx.New(app.Module()).Run()` |
| `fx.Run` | 启动应用，阻塞直到收到中断信号，然后按序执行OnStop | main.go唯一一行核心代码 |

**fx的核心优势**：
1. 无需手动管理组件构造顺序，fx自动解析依赖DAG（有向无环图）
2. 可选依赖优雅降级：Redis/MQ连接失败时返回nil，Service层判断nil即可降级
3. 优雅关停自动处理：收到SIGINT/SIGTERM时按OnStop注册逆序关闭
4. 代码解耦：组件只需声明依赖，无需关心如何构造依赖

### 4.2 Module()函数详解

Module()函数定义在 [module.go](file:///d:/Programming/Agent_demo/casbin_demo/backend/internal/app/module.go)，是整个应用的组装中心：

```go
func Module() fx.Option {
    return fx.Options(
        fx.Provide(
            // 基础资源
            provideConfig,      // 加载config.yaml
            provideDB,          // 初始化PostgreSQL
            provideRedis,       // 初始化Redis（失败返回nil）
            provideMQ,          // 初始化RabbitMQ（失败返回nil，降级PG-only）
            provideWSHub,       // 创建WebSocket Hub（依赖rdb做PubSub）
            provideCacheClient, // 创建两级缓存客户端（依赖rdb，注册Lifecycle OnStop）

            // Repository 层（fx自动注入*gorm.DB）
            repository.NewUserRepository,
            repository.NewRoleRepository,
            repository.NewPermissionRepository,
            repository.NewAuditRepository,
            repository.NewResourceRepository,

            // Service 层（fx自动注入Repos+Cache+MQ+WS）
            service.NewAuthService,
            service.NewUserService,
            service.NewRoleService,
            service.NewPermissionService,
            service.NewDashboardService,
            service.NewAuditService,
            service.NewResourceService,

            // Handler 层（fx自动注入Services）
            handler.NewAuthHandler,
            handler.NewUserHandler,
            handler.NewRoleHandler,
            handler.NewPermissionHandler,
            handler.NewDashboardHandler,
            handler.NewAuditHandler,
            handler.NewResourceHandler,
            handler.NewWsHandler,

            // Router 层（fx自动注入Handlers+Config）
            provideHandlers,    // 聚合所有Handler到router.Handlers结构体
            provideEngine,      // 创建Gin引擎+限流器，注册Lifecycle OnStop
        ),
        fx.Invoke(
            setupLogger,        // 初始化slog结构化日志
            autoMigrate,        // GORM AutoMigrate新增表
            seedData,           // 幂等初始化种子数据（admin/user账号、角色、权限、Casbin策略）
            initCasbin,         // 初始化Casbin Enforcer并加载策略
            warmupCache,        // 缓存预热：角色/权限列表+布隆过滤器用户ID
            startWSHub,         // 启动WS Hub事件循环
            startMQConsumers,   // 注入MQ消息处理器+启动定时补发协程
            registerRoutes,     // 注册所有HTTP路由
            startHTTPServer,    // 启动HTTP Server（注册Lifecycle OnStart/OnStop）
        ),
    )
}
```

### 4.3 Provide组件构造顺序

fx根据构造函数的参数类型自动推导构造顺序，无需手动指定：

**第1层：无依赖基础组件**
- `provideConfig() *config.Config` - 无参数，最先构造
- `provideWSHub(rdb *redis.Client) *ws.Hub` - 等待Redis

**第2层：依赖Config**
- `provideDB(cfg *config.Config) *gorm.DB`
- `provideRedis(cfg *config.Config) *redis.Client` - 失败返回nil不panic
- `provideMQ(cfg *config.Config) *mq.Client` - 失败返回nil不panic

**第3层：依赖基础资源**
- `provideCacheClient(lc fx.Lifecycle, rdb *redis.Client) *cache.Client` - 注册OnStop关闭
- `provideEngine(cfg *config.Config, lc fx.Lifecycle, mqClient *mq.Client, wsHub *ws.Hub, auditSvc *service.AuditService) *router.EngineWrapper` - 注册OnStop关停组件

**第4层：Repository层（依赖*gorm.DB）**
- `repository.NewUserRepository(db *gorm.DB) *UserRepository`
- `repository.NewRoleRepository(db *gorm.DB) *RoleRepository`
- `repository.NewPermissionRepository(db *gorm.DB) *PermissionRepository`
- `repository.NewAuditRepository(db *gorm.DB) *AuditRepository`
- `repository.NewResourceRepository(db *gorm.DB) *ResourceRepository`

**第5层：Service层（依赖Repos+Cache+MQ+WS）**
- `service.NewAuditService(auditRepo, userRepo, cacheClient, rdb, mqClient, wsHub) *AuditService` - 依赖最多
- 其他Service类似

**第6层：Handler层（依赖Services）**
- 每个Handler构造函数接收对应的Service

**第7层：聚合层**
- `provideHandlers(auth, user, role, perm, dash, audit, resource, wsHandler) *router.Handlers`
- 等待所有Handler构造完成后聚合

### 4.4 Invoke启动钩子详解

Invoke函数在所有Provide构造完成后按注册顺序执行：

| Invoke函数 | 执行时机 | 职责 |
|-----------|---------|-----|
| `setupLogger()` | 最先 | 配置slog为JSON格式，打印启动横幅 |
| `autoMigrate(auditRepo, resourceRepo)` | 数据库就绪后 | GORM自动迁移审核、消息、资源表 |
| `seedData(db *gorm.DB)` | AutoMigrate后 | 幂等创建admin/user角色、账号、权限、Casbin策略 |
| `initCasbin(db *gorm.DB)` | 种子数据后 | 初始化Casbin Enforcer，从DB加载策略 |
| `warmupCache(cache, userRepo, roleRepo, permRepo)` | Casbin就绪后 | 预热角色/权限列表到缓存，加载用户ID到布隆过滤器 |
| `startWSHub(hub *ws.Hub)` | 缓存预热后 | 启动WS Hub事件循环goroutine |
| `startMQConsumers(auditSvc)` | WS启动后 | 设置MQ消息处理器，启动定时补发协程 |
| `registerRoutes(engine, h)` | 所有组件就绪后 | 注册三层路由架构到Gin引擎 |
| `startHTTPServer(lc, cfg, engine)` | 最后 | 通过fx.Lifecycle OnStart在goroutine中启动HTTP Server |

### 4.5 Lifecycle生命周期管理

通过`fx.Lifecycle`的`Append`方法注册Hook，fx管理启动和关闭：

```go
// 示例：startHTTPServer中的Lifecycle使用
lc.Append(fx.Hook{
    OnStart: func(ctx context.Context) error {
        go func() {
            slog.Info("server listening", "addr", addr)
            if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
                slog.Error("server failed to start", "error", err)
            }
        }()
        return nil
    },
    OnStop: func(ctx context.Context) error {
        slog.Info("shutdown signal received, gracefully stopping...")
        shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
        defer cancel()
        return srv.Shutdown(shutdownCtx) // 等待活跃请求完成
    },
})
```

### 4.6 优雅关停顺序

fx收到os.Interrupt信号（Ctrl+C）或SIGTERM（docker stop）时，按OnStop注册**逆序**执行：

```
OnStop执行顺序（与OnStart注册顺序相反）：
1. startHTTPServer.OnStop
   └─ http.Server.Shutdown() → 停止接受新请求，等待活跃请求完成（10s超时）
2. provideEngine.OnStop（EngineWrapper）
   ├─ auditSvc.Stop() → 停止定时补发协程，wg.Wait()等待goroutine退出
   ├─ wsHub.Stop() → 关闭所有WebSocket连接，停止PubSub订阅，wg.Wait()
   ├─ mqClient.Close() → 停止消费者，等待在途消息ACK，关闭连接
   └─ engine.Stop() → 停止限流器后台清理goroutine
3. provideCacheClient.OnStop
   ├─ c.Close() → 停止异步重建协程，关闭local cache清理goroutine
   └─ rdb.Close() → 关闭Redis连接
```

这个顺序确保：
- 先停止接受新请求，处理完在途请求
- 然后停止消息和WebSocket相关组件，避免新消息进来
- 最后关闭缓存和数据库连接

---

## 第5章 缓存架构设计

### 5.1 两级缓存架构

系统采用 **L1本地内存缓存 + L2 Redis分布式缓存** 的两级缓存架构，请求链路如下：

```
请求 → L1本地缓存（sync.Map，微秒级） → 命中则直接返回
     ↓ 未命中
     L2 Redis（毫秒级网络往返） → 命中则回填L1后返回
     ↓ 未命中/熔断
     DB（PostgreSQL，唯一可信数据源） → 写入两级缓存后返回
```

各层级职责：

| 层级 | 实现 | TTL | 用途 |
|-----|------|-----|------|
| L1 | sync.Map + RWMutex | 30s~2min | 微秒级访问，抗Redis抖动，短TTL保证最终一致性 |
| L2 | Redis | 5min~30min±10%抖动 | 分布式缓存，多实例共享，防穿透/击穿/雪崩 |
| DB | PostgreSQL | 持久化 | 唯一可信数据源，缓存miss/失效时兜底 |

### 5.2 缓存分类与TTL策略

缓存按数据特性分类，使用不同TTL和缓存策略：

| 缓存类型 | TTL | 抖动 | 策略 | 示例Key |
|---------|-----|------|------|---------|
| 热点数据（HotData） | 30min | ±10% | 逻辑过期+异步重建 | 角色列表、权限列表 |
| 查询缓存（Query） | 5min | ±10% | 普通TTL | 用户列表分页、申请列表 |
| 仪表盘统计 | 2min | ±10% | 普通TTL | `cache:dashboard:stats`、待审数 |
| 配置数据 | 1h | ±10% | 普通TTL | 系统配置（预留） |
| 空值标记 | 60s | 无抖动 | 防穿透 | `__NULL__`标记 |
| 本地缓存 | 30s | 无 | 极短TTL | L1层独立TTL |
| 撤回窗口TTL | 120s | 无抖动 | 精确控制窗口 | `cache:audit:withdraw:{id}` |
| 分布式锁 | 10s | 无 | SETNX自动过期 | `lock:{cache_key}` |

TTL抖动实现见 [cache.go#L233-L240](file:///d:/Programming/Agent_demo/casbin_demo/backend/pkg/cache/cache.go#L233-L240)：

```go
func (c *Client) withJitter(ttl time.Duration) time.Duration {
    if c.jitter <= 0 || ttl <= 0 {
        return ttl
    }
    delta := time.Duration(float64(ttl) * c.jitter)  // jitter=0.1
    offset := time.Duration(rand.Int63n(int64(delta*2))) - delta  // [-10%, +10%]
    return ttl + offset
}
```

**为什么要加抖动**：如果大批Key同时写入缓存且TTL相同，它们会在同一时刻集体过期，导致瞬间大量请求打到DB（缓存雪崩）。随机抖动±10%将过期时间分散开，避免集体失效。

### 5.3 缓存三大问题防护

#### 5.3.1 缓存穿透防护（三层）

缓存穿透：查询**不存在**的数据，绕过缓存直打DB。攻击者可构造大量不存在ID的请求压垮DB。

| 防护层 | 实现位置 | 原理 |
|-------|---------|------|
| 布隆过滤器 | [bloom.go](file:///d:/Programming/Agent_demo/casbin_demo/backend/pkg/cache/bloom.go) | 启动时预热合法ID，请求到达时先过过滤器，不存在直接拦截（100%准确） |
| 空值缓存 | [cache.go#L300-L305](file:///d:/Programming/Agent_demo/casbin_demo/backend/pkg/cache/cache.go#L300-L305) | DB查询为空时写入`__NULL__`标记，短TTL(60s)，后续请求直接返回not found |
| IP限流 | [ratelimit.go](file:///d:/Programming/Agent_demo/casbin_demo/backend/internal/middleware/ratelimit.go) | 网关层IP级令牌桶限流，拦截高频恶意请求 |

**注意**：认证接口（登录/注册）不缓存空值，避免攻击者用不存在的用户名刷缓存占用内存。

#### 5.3.2 缓存击穿防护（三层）

缓存击穿：**热点Key过期瞬间**，大量并发请求同时打到DB。

| 防护层 | 实现位置 | 原理 |
|-------|---------|------|
| singleflight | [cache.go#L496](file:///d:/Programming/Agent_demo/casbin_demo/backend/pkg/cache/cache.go#L496) | 同一进程内合并同Key并发请求，只有一个协程查DB，其他等待结果 |
| 分布式锁 | [cache.go#L513-L533](file:///d:/Programming/Agent_demo/casbin_demo/backend/pkg/cache/cache.go#L513-L533) | 多实例部署时通过Redis SETNX互斥锁，全局只有一个实例查DB |
| 逻辑过期 | [cache.go#L322-L340](file:///d:/Programming/Agent_demo/casbin_demo/backend/pkg/cache/cache.go#L322-L340) | 热点数据Redis TTL设为0（永不过期），值内携带逻辑过期时间，过期时返回旧数据+异步重建 |

#### 5.3.3 缓存雪崩防护（四层）

缓存雪崩：大量Key同时过期或Redis宕机，DB压力骤增引发级联故障。

| 防护层 | 实现位置 | 原理 |
|-------|---------|------|
| TTL随机抖动 | `withJitter()` | 避免大批Key同时过期 |
| 多级缓存 | L1 sync.Map | Redis不可用时L1本地缓存继续提供服务（短TTL兜底） |
| 熔断器 | [circuit.go](file:///d:/Programming/Agent_demo/casbin_demo/backend/pkg/cache/circuit.go) | Redis连续失败时自动熔断（三态：Closed/Open/Half-Open），快速降级到DB，防止雪崩扩散 |
| 网关限流 | RateLimit中间件 | IP级+接口级限流，即使缓存全挂也限制DB入口流量 |

### 5.4 cache.Client核心结构体详解

[Client结构体](file:///d:/Programming/Agent_demo/casbin_demo/backend/pkg/cache/cache.go#L123-L136) 是缓存系统的核心：

```go
type Client struct {
    rdb         *redis.Client          // L2 Redis客户端（可为nil，表示PG-only模式）
    sf          singleflight.Group     // singleflight合并同Key并发请求（防击穿-进程内）
    local       *LocalCache            // L1本地内存缓存（sync.Map实现）
    bloom       *BloomFilter           // 布隆过滤器（防穿透第一层）
    cb          *CircuitBreaker        // Redis熔断器（防雪崩）
    ctx         context.Context        // 全局context
    jitter      float64                // TTL抖动比例（默认0.1即±10%）
    rebuildCh   chan rebuildTask       // 逻辑过期异步重建通道（缓冲100）
    rebuildOnce sync.Once              // 确保重建worker只启动一次
    closeOnce   sync.Once              // 确保Close只执行一次
    closed      chan struct{}          // 关闭信号
}
```

构造函数 [NewClient](file:///d:/Programming/Agent_demo/casbin_demo/backend/pkg/cache/cache.go#L152-L164)：

```go
func NewClient(rdb *redis.Client) *Client {
    c := &Client{
        rdb:    rdb,
        local:  NewLocalCache(TTLLocal),           // L1缓存默认TTL 30s
        bloom:  NewBloomFilter(1<<20, 7),          // 1M位≈128KB，7个哈希函数
        cb:     NewCircuitBreaker("redis", 5, 30*time.Second), // 连续5次失败熔断，30s后半开
        ctx:    context.Background(),
        jitter: 0.1,                                // ±10% TTL抖动
        closed: make(chan struct{}),
    }
    c.startRebuildWorker()  // 启动异步重建协程
    return c
}
```

### 5.5 Fetch方法完整流程代码级解析

[Fetch方法](file:///d:/Programming/Agent_demo/casbin_demo/backend/pkg/cache/cache.go#L441-L565) 是缓存读取的核心入口，实现了完整的多级缓存+三大问题防护。逐段解析：

**Step 1-2: 布隆过滤器拦截（防穿透）**

```go
if opt.UseBloom && c.bloom != nil {
    bloomKey := key
    if !c.bloom.Contains(bloomKey) {
        return nil, false, nil  // 布隆过滤器判定不存在，直接返回（100%准确）
    }
}
```

布隆过滤器`Contains`返回false表示元素**一定不存在**，此时不查Redis也不查DB，直接返回nil。返回true表示**可能存在**（有小概率误判），继续后续流程。

**Step 3: 多级缓存查找（L1→L2）**

```go
data, found, isNull, fromLocal, lexpired := c.GetBytes(key)
if found {
    if isNull {
        return nil, false, nil  // 命中空值标记，返回not found（防穿透第二层）
    }
    // 处理逻辑过期条目
    var entry logicalEntry
    if json.Unmarshal(data, &entry) == nil && entry.Logical {
        if time.Now().After(entry.ExpireAt) || lexpired {
            // 逻辑过期：触发异步重建，立即返回旧数据（防击穿）
            select {
            case c.rebuildCh <- rebuildTask{key: key, ttl: opt.TTL, loader: loader}:
            default:  // 通道满则跳过，不阻塞
            }
        }
        json.Unmarshal(entry.Data, result)
        return result, true, nil  // 返回旧数据（即使已过期）
    }
    // 普通缓存命中
    json.Unmarshal(data, result)
    return result, true, nil
}
```

[GetBytes方法](file:///d:/Programming/Agent_demo/casbin_demo/backend/pkg/cache/cache.go#L250-L284) 的查找顺序：
1. 先查L1本地缓存（sync.Map，无网络开销）
2. L1未命中时检查熔断器，若Open则跳过Redis
3. 查Redis，成功则RecordSuccess()，失败则RecordFailure()
4. Redis命中后回填L1（空值标记TTL 60s，普通数据TTL 30s）

**Step 4-5: singleflight + 分布式锁（防击穿）**

```go
v, err, _ := c.sf.Do(key, func() (interface{}, error) {
    // double-check：等待锁期间可能已有其他协程写入缓存
    data2, found2, isNull2, _, _ := c.GetBytes(key)
    if found2 { /* 缓存已有数据，直接返回 */ }

    // 分布式锁（多实例防击穿）
    var unlock func()
    acquired := false
    if opt.UseDistLock && c.Enabled() {
        lockKey := lockPrefix + key
        acquired, unlock = c.acquireLock(lockKey)  // SETNX，TTL 10s
        if acquired {
            defer unlock()
        } else {
            time.Sleep(50 * time.Millisecond)  // 等待其他实例完成DB查询
            // triple-check：等待后再查一次缓存
            data3, found3, isNull3, _, _ := c.GetBytes(key)
            if found3 { /* 返回缓存数据 */ }
        }
    }

    // 执行loader从DB加载数据
    loaded, err := loader()
    if err != nil { return nil, err }
    if loaded == nil {
        c.SetNull(key)  // DB返回nil，写入空值标记（防穿透）
        return nil, nil
    }

    // 写入缓存（逻辑过期 或 普通TTL+抖动）
    if opt.UseLogicalExp {
        c.SetLogical(key, loaded, opt.TTL)
    } else {
        c.SetJSON(key, loaded, opt.TTL)  // 内部调用withJitter
    }

    // 回填布隆过滤器（新数据ID加入）
    if opt.UseBloom && c.bloom != nil {
        c.bloom.Add(key)
    }
    return loaded, nil
})
```

Fetch流程完整状态图：

```
请求进入Fetch
    │
    ├─[UseBloom]──► BloomFilter.Contains?
    │                  ├─ false → return nil（一定不存在，防穿透）
    │                  └─ true → 继续
    │
    ├─ GetBytes (L1→L2)
    │   ├─ 命中NULL → return nil（空值缓存，防穿透）
    │   ├─ 命中逻辑过期 → 返回旧数据 + 异步rebuild（防击穿）
    │   └─ 命中普通数据 → return data
    │
    └─ 缓存未命中
        │
        └─ singleflight.Do (合并并发)
            │
            ├─ double-check (GetBytes)
            │   └─ 命中 → return
            │
            ├─ [UseDistLock] SETNX分布式锁?
            │   ├─ 获取成功 → 查DB
            │   ├─ 获取失败 → sleep 50ms → triple-check → 命中则return
            │   └─ 未获取到有效数据 → 查DB
            │
            ├─ loader()查DB
            │   ├─ 返回nil → SetNull (防穿透)
            │   └─ 返回data → SetJSON/SetLogical (TTL±jitter防雪崩)
            │                  → BloomFilter.Add
            │
            └─ return loaded
```

### 5.6 L1本地缓存实现（local.go）

[LocalCache](file:///d:/Programming/Agent_demo/casbin_demo/backend/pkg/cache/local.go#L35-L41) 使用`sync.RWMutex`保护`map[string]*localItem`，支持两种缓存模式：

```go
type localItem struct {
    data        []byte    // JSON序列化后的字节
    expireAt    time.Time // 普通模式：绝对过期时间
    isNull      bool      // 是否是空值标记__NULL__
    logicalExp  bool      // 是否为逻辑过期条目
    logicalTime time.Time // 逻辑过期时间（热点数据）
}
```

**普通TTL模式**（Set方法）：过期后在Get时惰性删除 + 后台GC每分钟定期清理。

**逻辑过期模式**（SetLogical方法）：永不过期，不被GC清理。过期时返回旧数据并由上层触发异步重建。

后台GC协程 [gc()](file:///d:/Programming/Agent_demo/casbin_demo/backend/pkg/cache/local.go#L70-L88)：
- 每分钟扫描一次
- 只清理普通TTL条目中已过期的
- 逻辑过期条目不清理（由异步重建更新）
- Stop()通过close(stopCh)停止，防止goroutine泄漏

### 5.7 布隆过滤器实现（bloom.go）

[BloomFilter](file:///d:/Programming/Agent_demo/casbin_demo/backend/pkg/cache/bloom.go#L25-L31) 使用位数组+双重哈希技术：

```go
type BloomFilter struct {
    m      uint         // 位数组大小（默认1<<20 = 1,048,576 bits ≈ 128KB）
    k      uint         // 哈希函数个数（默认7）
    bitset []bool       // 位数组（bool切片，简单直观）
    mu     sync.RWMutex // 读写锁支持并发读
    count  uint         // 已添加元素计数
}
```

**双重哈希（Double Hashing）**：通过FNV-1a算法生成两个独立哈希值`hash1`和`hash2`，然后线性组合生成k个位置：

```go
func (bf *BloomFilter) locations(key string) []uint {
    h := fnv.New64a()
    h.Write([]byte(key))
    hash1 := h.Sum64()
    h.Reset()
    h.Write([]byte(key + "|salt"))
    hash2 := h.Sum64()
    locs := make([]uint, bf.k)
    for i := uint(0); i < bf.k; i++ {
        locs[i] = uint((hash1 + uint64(i)*hash2) % uint64(bf.m))
    }
    return locs
}
```

比使用k个不同哈希函数更高效，且分布性良好。误判率约为0.8%（n=100000, m=1<<20, k=7）。

**Contains方法的核心保证**：如果任何一个哈希位置为false，元素**一定不存在**（100%准确）。这正是防穿透利用的特性。

### 5.8 Redis熔断器实现（circuit.go）

[CircuitBreaker](file:///d:/Programming/Agent_demo/casbin_demo/backend/pkg/cache/circuit.go#L30-L39) 实现三态熔断模式：

```go
type CircuitBreaker struct {
    mu           sync.Mutex
    failures     int           // 当前连续失败次数
    maxFailures  int           // 熔断阈值（默认5）
    open         bool          // 熔断器是否打开
    openUntil    time.Time     // 打开截止时间（半开探测时机）
    resetTimeout time.Duration // 熔断持续时间（默认30s）
    semiOpen     bool          // 是否处于半开状态
    name         string        // 熔断器名称
}
```

状态转换：

```
Closed（正常放行）
    │
    ├─ RecordSuccess() → failures=0（重置失败计数）
    │
    └─ RecordFailure() → failures++
        └─ failures >= maxFailures → Open（熔断）

Open（快速失败，不访问Redis）
    │
    └─ Allow()检查：time.Now() > openUntil → Half-Open（放一个探测请求）

Half-Open（半开探测）
    ├─ RecordSuccess() → Closed（探测成功，恢复正常）
    └─ RecordFailure() → Open（探测失败，重新熔断）
```

在缓存GetBytes中的使用：
- Allow()返回false → 跳过Redis，直接返回未命中（降级到DB/singleflight）
- Redis操作成功 → RecordSuccess()
- Redis操作失败 → RecordFailure()

### 5.9 缓存预热机制

启动时通过fx.Invoke调用`warmupCache`，执行以下预热：
1. 加载角色全量列表到缓存（热点数据，逻辑过期）
2. 加载权限全量列表到缓存（热点数据，逻辑过期）
3. 将所有用户ID加入布隆过滤器（防穿透）

通过 [Warmup方法](file:///d:/Programming/Agent_demo/casbin_demo/backend/pkg/cache/cache.go#L576-L596) 批量执行，记录成功/失败数量。

### 5.10 缓存失效策略

**写操作必须主动失效相关缓存**，遵循以下原则：

| 操作 | 失效策略 | 实现方法 |
|-----|---------|---------|
| 用户CRUD | SCAN+DEL `cache:user:*` + 仪表盘缓存 | `InvalidateUserCaches()` |
| 角色CRUD | SCAN+DEL `cache:role:*` + 权限缓存 | `InvalidateRoleCaches()` |
| 权限CRUD | SCAN+DEL `cache:permission:*` + 仪表盘 | `InvalidatePermissionCaches()` |
| 提交/审核/撤回申请 | DEL待审数 + SCAN+DEL申请列表 + 消息未读数 | `invalidateAuditCaches()` |
| 标记消息已读 | DEL对应用户未读数缓存 | `cache.Delete(msgUnreadKey+uid)` |

SCAN+DEL使用Redis SCAN游标迭代，每批100个DEL，避免KEYS命令阻塞Redis。撤回成功后主动DEL撤回TTL缓存，防止二次撤回。

---

## 第6章 网关限流架构

### 6.1 三层限流模型

系统在网关层实施**三层限流**防护，从外到内依次检查，任意一层拒绝即返回HTTP 429：

```
请求进入
    │
    ├─[0] 白名单检查 (/health, /ws开头) → 直接放行
    │
    ├─[1] IP级限流（全局中间件RateLimit）
    │    key: rate:ip:{client_ip}
    │    配置: 30 req/s, burst 60
    │    目的: 防DDoS/CC攻击，单IP恶意刷请求
    │
    ├─[2] 接口级限流（全局中间件RateLimit）
    │    key: rate:api:{METHOD}:{normalized_path}
    │    配置: 100 req/s, burst 200
    │    目的: 保护全局接口不被压垮，防服务雪崩
    │    注意: 路径归一化（数字ID替换为:id），防止绕过
    │
    └─[3] 用户级限流（JWTAuth之后UserRateLimit）
         key: rate:user:{username}
         配置: 20 req/s, burst 40
         目的: 防止单个登录用户滥用API资源
```

### 6.2 令牌桶算法原理与Lua脚本

令牌桶算法相比固定窗口/滑动窗口的优势：
- **支持突发流量**：桶内积累的令牌可一次性消耗，应对合理突发
- **平滑限流**：匀速补充令牌，输出速率平稳
- **无固定窗口边界问题**：不会出现窗口切换时的双倍流量突刺

**Redis分布式令牌桶**使用Lua脚本保证原子性（[ratelimit.go#L424-L457](file:///d:/Programming/Agent_demo/casbin_demo/backend/internal/middleware/ratelimit.go#L424-L457)）：

```lua
-- KEYS[1] = 限流key
-- ARGV[1] = rate（每秒补充令牌数）
-- ARGV[2] = capacity（桶容量/突发上限）
-- ARGV[3] = now（当前时间戳，秒浮点）
-- ARGV[4] = ttl（key过期秒数）
local key = KEYS[1]
local rate = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local ttl = tonumber(ARGV[4])

-- 读取当前桶状态
local data = redis.call('HMGET', key, 'tokens', 'last_refill')
local tokens = tonumber(data[1])
local lastRefill = tonumber(data[2])

-- 首次访问：桶满
if tokens == nil then
    tokens = capacity
    lastRefill = now
else
    -- 计算经过时间补充令牌
    local elapsed = now - lastRefill
    tokens = tokens + elapsed * rate
    if tokens > capacity then tokens = capacity end
    lastRefill = now
end

-- 无令牌：拒绝
if tokens < 1 then
    redis.call('HMSET', key, 'tokens', tokens, 'last_refill', lastRefill)
    redis.call('EXPIRE', key, ttl)
    return 0
end

-- 消耗1个令牌：放行
tokens = tokens - 1
redis.call('HMSET', key, 'tokens', tokens, 'last_refill', lastRefill)
redis.call('EXPIRE', key, ttl)
return 1
```

Lua脚本在Redis端原子执行，避免竞态条件。Key设置5分钟TTL自动过期，不永久占内存。

### 6.3 RateLimiter结构体详解

[RateLimiter](file:///d:/Programming/Agent_demo/casbin_demo/backend/internal/middleware/ratelimit.go#L172-L202)：

```go
type RateLimiter struct {
    mu       sync.Mutex
    buckets  map[string]*rateBucket  // 本地内存令牌桶（降级用）
    stopCh   chan struct{}
    once     sync.Once
    cb       *circuitBreaker         // Redis熔断器
    stats    RateLimitStats          // 限流统计（atomic操作）

    ipRate, ipCapacity     float64   // IP限流配置
    apiRate, apiCapacity   float64   // 接口限流配置
    userRate, userCapacity float64   // 用户限流配置

    whitelistPrefixes []string       // 白名单前缀（/ws）
    whitelistPaths    map[string]bool // 白名单精确路径（/health）
    maxLocalBuckets   int            // 本地桶上限（防内存溢出）
    redisClient       *redis.Client  // Redis客户端（nil则纯本地限流）
}
```

[rateBucket](file:///d:/Programming/Agent_demo/casbin_demo/backend/internal/middleware/ratelimit.go#L137-L142) 本地令牌桶：

```go
type rateBucket struct {
    tokens     float64   // 当前令牌数
    maxTokens  float64   // 桶容量
    refillRate float64   // 每秒补充速率
    lastRefill time.Time // 上次补充时间
}
```

### 6.4 限流熔断器实现

限流熔断器 [circuitBreaker](file:///d:/Programming/Agent_demo/casbin_demo/backend/internal/middleware/ratelimit.go#L53-L61) 与缓存熔断器类似但使用atomic操作（更高并发性能）：

- **failureThreshold=5**：连续5次Redis操作失败则熔断
- **successThreshold=3**：半开状态下连续3次成功则关闭
- **resetTimeout=30s**：熔断30秒后进入半开探测
- 熔断时Redis Fallbacks计数+1，降级为本地限流

### 6.5 本地内存令牌桶降级实现

[allowLocal](file:///d:/Programming/Agent_demo/casbin_demo/backend/internal/middleware/ratelimit.go#L364-L398)：

```go
func (rl *RateLimiter) allowLocal(key string, rate, capacity float64) bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    b, ok := rl.buckets[key]
    if !ok {
        // 本地桶数量上限检查
        if len(rl.buckets) >= rl.maxLocalBuckets {
            return true  // 桶满了，放行（保护服务可用性优先）
        }
        b = &rateBucket{tokens: capacity, maxTokens: capacity, refillRate: rate, lastRefill: time.Now()}
        rl.buckets[key] = b
    }

    // 补充令牌
    now := time.Now()
    elapsed := now.Sub(b.lastRefill).Seconds()
    b.tokens += elapsed * b.refillRate
    if b.tokens > b.maxTokens { b.tokens = b.maxTokens }
    b.lastRefill = now

    if b.tokens < 1 { return false }
    b.tokens--
    return true
}
```

### 6.6 限流Key设计与路径归一化

**Key命名规范**：

| 维度 | Key格式 | 示例 |
|-----|--------|------|
| IP | `rate:ip:{ip}` | `rate:ip:192.168.1.1` |
| 接口 | `rate:api:{METHOD}:{normalized_path}` | `rate:api:GET:/api/users/:id` |
| 用户 | `rate:user:{username}` | `rate:user:admin` |

**路径归一化** [normalizePath](file:///d:/Programming/Agent_demo/casbin_demo/backend/internal/middleware/ratelimit.go#L631-L659)：将URL中的纯数字段替换为`:id`，防止不同ID生成不同key导致限流失效：

- `/api/users/123` → `/api/users/:id`
- `/api/audit/applications/456` → `/api/audit/applications/:id`
- `/api/resources` → `/api/resources`（不变）
- UUID（含字母和`-`）不替换，只匹配纯数字段

### 6.7 白名单机制

白名单路径不限流：
- **精确匹配**：`/health`（Docker/K8s健康检查探针）
- **前缀匹配**：`/ws`开头（WebSocket连接有自己的心跳/频控机制）

### 6.8 限流统计与后台清理

[RateLimitStats](file:///d:/Programming/Agent_demo/casbin_demo/backend/internal/middleware/ratelimit.go#L145-L152) 记录：
- TotalRequests：总请求数
- Allowed：通过数
- RejectedIP/RejectedAPI/RejectedUser：各层拒绝数
- RedisFallbacks：Redis降级次数（监控熔断器状态）

后台清理 [cleanup()](file:///d:/Programming/Agent_demo/casbin_demo/backend/internal/middleware/ratelimit.go#L325-L361)：
- 每5分钟扫描一次
- 删除10分钟内无请求且令牌已满的桶（非活跃）
- 桶总数超过上限(10000)时强制清空（保护内存，短暂放通）
- Stop()通过close(stopCh)停止

限流触发时返回标准HTTP 429响应，包含`Retry-After`和`X-RateLimit-*`头：

```go
c.Header("Retry-After", "1")
c.Header("X-RateLimit-Limit", "60")
c.Header("X-RateLimit-Remaining", "0")
c.AbortWithStatusJSON(429, gin.H{
    "code": 429, "message": "请求过于频繁，请稍后再试", "type": "ip_rate_limit",
})
```

---

## 第7章 后端架构逐模块详解

### 7.1 入口层 main.go

[main.go](file:///d:/Programming/Agent_demo/casbin_demo/backend/cmd/server/main.go) 极其简洁，唯一职责是启动fx应用：

```go
func main() {
    fx.New(app.Module()).Run()
}
```

- `fx.New()` 创建fx应用，传入Module中定义的所有组件
- `.Run()` 阻塞运行，启动所有OnStart钩子，等待os.Interrupt信号
- 收到信号后按OnStop逆序执行优雅关停

**设计原则**：main.go只做启动，不包含任何业务逻辑或资源初始化代码。

### 7.2 应用初始化层 app/init.go

[init.go](file:///d:/Programming/Agent_demo/casbin_demo/backend/internal/app/init.go) 提供基础初始化工具函数：

| 函数 | 职责 |
|-----|------|
| `setupSlog()` | 配置slog为JSON格式输出到stdout，Level=Info |
| `mustLoadConfig()` | 加载config.yaml，失败降级默认配置 |
| `mustInitDB(cfg)` | 初始化PostgreSQL连接（GORM），失败panic |
| `mustSeedData(db)` | 幂等执行种子数据初始化，失败panic |

这些函数被module.go中的fx.Invoke/Provide调用。

### 7.3 模块注册层 app/module.go（完整代码解析）

[module.go](file:///d:/Programming/Agent_demo/casbin_demo/backend/internal/app/module.go) 是整个应用的组装中心，定义了`Module() fx.Option`。它通过fx.Provide注册所有组件构造函数，通过fx.Invoke注册启动钩子。

**fx.Provide 组件注册（按依赖层次）**：

1. **基础资源层（无业务依赖）**：
   - `provideConfig` → `*config.Config`：加载YAML配置+环境变量
   - `provideDB` → `*gorm.DB`：PostgreSQL连接，失败panic（必须成功）
   - `provideRedis` → `*redis.Client`：Redis连接，**失败返回nil不panic**（可选依赖）
   - `provideMQ` → `*mq.Client`：RabbitMQ连接，**失败返回nil不panic**（可选依赖，PG-only降级）
   - `provideWSHub` → `*ws.Hub`：WebSocket Hub，依赖rdb做PubSub
   - `provideCacheClient` → `*cache.Client`：两级缓存客户端，注册Lifecycle OnStop

2. **Repository层（依赖*gorm.DB）**：
   - `repository.NewUserRepository` → `*UserRepository`
   - `repository.NewRoleRepository` → `*RoleRepository`
   - `repository.NewPermissionRepository` → `*PermissionRepository`
   - `repository.NewAuditRepository` → `*AuditRepository`
   - `repository.NewResourceRepository` → `*ResourceRepository`

3. **Service层（依赖Repos + Cache + MQ + WS）**：
   - `service.NewAuthService`、`NewUserService`、`NewRoleService`...
   - `NewAuditService`依赖最多：auditRepo + userRepo + cacheClient + rdb + mqClient + wsHub

4. **Handler层（依赖Services）**：
   - 每个Handler构造函数接收对应的Service

5. **聚合层**：
   - `provideHandlers` → `*router.Handlers`：聚合所有Handler到一个结构体
   - `provideEngine` → `*router.EngineWrapper`：创建Gin引擎+限流器，注册OnStop

**fx.Invoke 启动钩子（按顺序执行）**：

| 顺序 | 函数 | 作用 |
|-----|------|------|
| 1 | `setupLogger` | 配置slog JSON日志 |
| 2 | `autoMigrate` | GORM AutoMigrate新增表（审核、消息、资源表） |
| 3 | `seedData` | 幂等种子数据：admin/user角色、账号、权限、Casbin策略 |
| 4 | `initCasbin` | 初始化Casbin Enforcer，从DB加载策略 |
| 5 | `warmupCache` | 预热角色/权限列表缓存，加载用户ID到布隆过滤器 |
| 6 | `startWSHub` | 启动WS Hub事件循环goroutine |
| 7 | `startMQConsumers` | 设置MQ消息handler，启动定时补发协程 |
| 8 | `registerRoutes` | 注册三层路由到Gin引擎 |
| 9 | `startHTTPServer` | 通过fx.Lifecycle OnStart在goroutine中启动HTTP Server |

**可选依赖降级策略**：
- provideRedis失败时返回`nil, nil`（不返回error），fx正常继续
- provideMQ失败时返回`nil, nil`，Service层判断`mqClient == nil`则降级为PG-only模式
- 降级后消息仍落PG，由定时对账补发（每30s扫描mq_delivered=false的消息）兜底
- WebSocket不依赖MQ仍可工作，本地直接推送

### 7.4 配置层 config/config.go

[Config结构体](file:///d:/Programming/Agent_demo/casbin_demo/backend/internal/config/config.go#L22-L28) 聚合所有子配置：

```go
type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
    Redis    RedisConfig    `mapstructure:"redis"`
    JWT      JWTConfig      `mapstructure:"jwt"`
    RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq"`
}
```

**配置来源优先级**（从高到低）：
1. 环境变量（如`DATABASE_HOST`、`REDIS_PORT`）
2. YAML配置文件（config.yaml）
3. 代码内置默认值（Default()函数）

[Load函数](file:///d:/Programming/Agent_demo/casbin_demo/backend/internal/config/config.go#L99-L143) 使用viper：
- `SetEnvKeyReplacer(strings.NewReplacer(".", "_"))`：将`database.host`映射到`DATABASE_HOST`
- `AutomaticEnv()`：自动绑定环境变量
- `applyDefaults()`：零值字段填充默认值
- DSN()方法构建PostgreSQL连接字符串，Addr()方法构建Redis host:port

### 7.5 模型层 model/model.go（完整代码解析）

[model.go](file:///d:/Programming/Agent_demo/casbin_demo/backend/internal/model/model.go) 定义所有数据结构，分为三类：**数据库Entity**、**请求DTO**、**响应DTO**。

#### 7.5.1 核心Entity

**User**（users表）：
```go
type User struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    UUID      string         `gorm:"uniqueIndex;size:36;not null" json:"uuid"`      // 对外暴露ID
    Username  string         `gorm:"uniqueIndex;size:50;not null" json:"username"` // 登录名
    Password  string         `gorm:"size:255;not null" json:"-"`                   // bcrypt哈希，永不序列化
    Nickname  string         `gorm:"size:50" json:"nickname"`
    Email     string         `gorm:"size:100" json:"email"`
    Avatar    string         `gorm:"size:255" json:"avatar"`
    Status    int            `gorm:"default:1" json:"status"`
    Roles     []Role         `gorm:"many2many:user_roles;" json:"roles,omitempty"` // 多对多
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // 软删除
}
```

关键点：`Password`字段使用`json:"-"`确保永不出现在API响应中；UUID作为对外暴露标识避免主键自增ID泄露。

**Role**（roles表）、**Permission**（permissions表）：类似的多对多关系设计。

**AuditApplication**（api_audit_applications表）- 审核申请核心实体：
- 复合索引：`idx_applicant_status (applicant_id, status)`、`idx_status_created (status, created_at)`
- **乐观锁**：`Version int`字段，更新时`WHERE version = ?`防并发修改
- 状态常量：Pending(0)/Approved(1)/Rejected(2)/Withdrawn(3)

**SysMessage**（sys_messages表）- 系统消息：
- 复合索引：`idx_receiver_read_created (receiver_id, is_read, created_at)`覆盖未读数+列表查询
- `MQDelivered bool`标记MQ是否投递成功，用于定时对账补发
- 外键关联：BusinessType+BusinessID关联业务数据

**Resource**（api_resources表）- API资源清单：
- Type字段：llm_chat/llm_code/image_gen/asr/tts/embedding/other
- Tags字段：JSON数组字符串存储分类标签

#### 7.5.2 请求/响应DTO

所有请求DTO使用`binding`标签定义校验规则：
```go
type LoginRequest struct {
    Username string `json:"username" binding:"required"`        // 必填
    Password string `json:"password" binding:"required"`
}
type CreateAuditRequest struct {
    ResourceName string `json:"resource_name" binding:"required,max=200"` // 必填+最大200字符
    ResourceType string `json:"resource_type" binding:"required,max=50"`
    Purpose      string `json:"purpose" binding:"required"`
    ExpectedQPS  int    `json:"expected_qps" binding:"min=0"`             // 最小值0
}
```

响应DTO做脱敏处理，不返回Password等敏感字段：
```go
type UserResponse struct {
    ID        uint      `json:"id"`
    UUID      string    `json:"uuid"`
    Username  string    `json:"username"`
    Nickname  string    `json:"nickname"`
    Roles     []string  `json:"roles"`  // 只返回角色名称列表
    // Password字段不存在于此结构体
}
```

### 7.6 路由层 router/router.go（完整代码解析）

[router.go](file:///d:/Programming/Agent_demo/casbin_demo/backend/internal/router/router.go) 负责Gin引擎创建、全局中间件装配、三层路由注册。

#### 7.6.1 NewEngine - 引擎创建与全局中间件

[NewEngine](file:///d:/Programming/Agent_demo/casbin_demo/backend/internal/router/router.go#L64-L95) 创建Gin引擎并注册全局中间件（按顺序执行）：

```go
func NewEngine(mode string) *EngineWrapper {
    r := gin.New()  // 不使用Default()，手动选择中间件
    limiter := middleware.DefaultAPIRateLimiter()

    r.Use(middleware.CustomRecovery())  // 1. Panic恢复（最外层，兜底）
    r.Use(cors.New(cors.Config{...}))  // 2. CORS跨域
    r.Use(middleware.RateLimit(limiter)) // 3. IP+接口限流
    r.Use(requestLogger())             // 4. 请求日志（最内层，能拿到最终响应状态码）

    r.NoRoute(func(c *gin.Context) { response.NotFound(c, "请求的资源不存在") })
    r.NoMethod(func(c *gin.Context) { response.MethodNotAllowed(c, "请求方法不被允许") })

    return &EngineWrapper{Engine: r, rateLimiter: limiter}
}
```

**中间件执行顺序（栈式）**：
```
请求进入
  → CustomRecovery (defer recover)
    → CORS (设置跨域头)
      → RateLimit (IP+接口限流)
        → RequestLogger (记录开始时间)
          → 路由匹配的Handler
            → JWTAuth/CasbinAuth (路由组中间件)
              → 具体Handler处理
        ← RequestLogger (c.Next()后记录日志)
      ← RateLimit
    ← CORS
  ← CustomRecovery
响应返回
```

#### 7.6.2 RegisterRoutes - 三层路由架构

[RegisterRoutes](file:///d:/Programming/Agent_demo/casbin_demo/backend/internal/router/router.go#L136-L219) 实现三层权限路由：

**第一层：公开接口（无中间件）**
```go
r.GET("/health", healthHandler)  // 健康检查
r.GET("/ws", h.WS.Connect)       // WebSocket（JWT从query参数取）
api.POST("/login", h.Auth.Login)
api.POST("/register", h.Auth.Register)
```

**第二层：基础认证接口（JWTAuth + UserRateLimit）**
```go
auth := api.Group("")
auth.Use(middleware.JWTAuth())       // JWT认证
auth.Use(middleware.UserRateLimit(ew.RateLimiter())) // 用户级限流
{
    auth.GET("/userinfo", ...)
    auth.GET("/dashboard", ...)
    auth.POST("/audit/applications", ...)       // 提交申请
    auth.POST("/audit/applications/:id/withdraw", ...) // 撤回
    auth.GET("/audit/my-applications", ...)
    auth.GET("/messages/*", ...)
    auth.GET("/resources/*", ...)
}
```

**第三层：管理员接口（CasbinAuth RBAC校验）**
```go
admin := auth.Group("")
admin.Use(middleware.CasbinAuth())  // Casbin RBAC权限校验
{
    admin.GET("/users", ...)       // 用户管理
    admin.POST("/users", ...)
    admin.GET("/roles", ...)       // 角色管理
    admin.GET("/permissions", ...) // 权限管理
    admin.GET("/audit/applications", ...)    // 查看所有申请
    admin.POST("/audit/applications/:id/review", ...) // 审核
    admin.POST("/resources", ...)  // 资源CRUD
}
```

嵌套关系：`admin`组在`auth`组内部，所以管理员接口自动继承JWT认证+用户级限流，再额外加CasbinAuth校验。

### 7.7 中间件层

#### 7.7.1 JWTAuth中间件

[jwt.go](file:///d:/Programming/Agent_demo/casbin_demo/backend/internal/middleware/jwt.go#L22-L50)：
1. 从`Authorization`头提取`Bearer <token>`
2. 验证格式（必须"Bearer "开头）
3. 调用`jwtpkg.ParseToken()`验证签名和有效期
4. 将`user_id/uuid/username/roles`写入Gin Context
5. 失败返回401

#### 7.7.2 CasbinAuth中间件

[casbin.go](file:///d:/Programming/Agent_demo/casbin_demo/backend/internal/middleware/casbin.go#L26-L60)：
1. 从Context获取username（sub）
2. 获取请求路径path（obj）和HTTP方法method（act）
3. 调用`casbinpkg.Enforcer.Enforce(sub, obj, act)`校验
4. Enforcer未初始化（降级场景）直接放行
5. 失败返回403，错误返回500

#### 7.7.3 CustomRecovery中间件

[casbin.go#L63-L76](file:///d:/Programming/Agent_demo/casbin_demo/backend/internal/middleware/casbin.go#L63-L76)：
- `defer recover()`捕获所有panic
- slog.Error记录错误日志（path、method、error）
- 返回JSON格式500响应（不是Gin默认的HTML错误页）

#### 7.7.4 requestLogger中间件

[router.go#L222-L248](file:///d:/Programming/Agent_demo/casbin_demo/backend/internal/router/router.go#L222-L248)：
- c.Next()前记录start时间
- c.Next()后按状态码分级日志：≥500 Error，≥400 Warn，其他Info
- 记录method、path、status、latency_ms、client_ip

### 7.8 数据访问层 Repository

Repository层封装所有数据库操作，返回Entity/error，不包含业务逻辑。

| Repository | 文件 | 核心职责 |
|-----------|------|---------|
| UserRepository | user_repo.go | 用户CRUD、FindByID、FindAdminUserIDs、FindByUsername |
| RoleRepository | role_repo.go | 角色CRUD、角色权限关联、Casbin策略同步 |
| PermissionRepository | permission_repo.go | 权限CRUD |
| AuditRepository | audit_repo.go | 审核申请CRUD、乐观锁更新、消息批量创建、分页查询、未读数统计 |
| ResourceRepository | resource_repo.go | 资源CRUD、分页查询 |

[AuditRepository](file:///d:/Programming/Agent_demo/casbin_demo/backend/internal/repository/audit_repo.go) 是最复杂的Repository，核心方法：
- `BeginTx()`：开启GORM事务
- `Create(tx, app)`：在事务中创建申请
- `UpdateStatusWithVersion(tx, appID, version, status, ...)`：乐观锁更新（`WHERE id=? AND version=?`）
- `WithdrawApplication(tx, appID, userID, version, reason)`：撤回（校验所有权+状态+乐观锁）
- `CreateMessagesBatch(tx, msgs)`：批量创建消息
- `FindUnDeliveredMessages(batchSize)`：查找mq_delivered=false的消息（定时补发用）
- `CountUnreadMessages(userID)`：统计未读数
- `ListMessages/ListByApplicant/ListAll`：分页查询（Preload关联、Count总条数）

### 7.9 业务逻辑层 Service

Service层是核心业务逻辑所在，处理事务、缓存、消息、DTO转换。

| Service | 核心职责 |
|---------|---------|
| AuthService | 登录验证、密码校验(bcrypt)、JWT生成、用户注册 |
| UserService | 用户CRUD+缓存失效、DTO转换(User→UserResponse)、角色分配 |
| RoleService | 角色CRUD+Casbin策略同步、权限分配、缓存失效 |
| PermissionService | 权限CRUD+缓存失效 |
| DashboardService | 仪表盘统计（用户数/角色数/权限数/待审数，带缓存） |
| AuditService | 审核申请提交/撤回/审核、MQ/WS推送、定时补发、消息管理 |
| ResourceService | 资源CRUD+缓存、可用资源列表 |

**Service层关键原则**：
1. **事务内不做Redis/MQ网络IO**：先提交PG事务，再操作Redis/MQ
2. **先提交PG事务，再发MQ消息**：防止"幽灵消息"（消息发出去了但事务回滚）
3. **写操作主动失效缓存**：增删改后Delete/DeleteByPattern相关缓存Key
4. **缓存读取使用cache.Fetch**：统一走两级缓存+三大问题防护

### 7.10 处理器层 Handler

Handler层负责HTTP协议处理：参数绑定、校验、调用Service、封装响应。不包含业务逻辑。

Handler方法统一模式：
```go
func (h *XxxHandler) SomeAction(c *gin.Context) {
    // 1. 绑定参数（ShouldBindJSON/ShouldBindQuery）
    var req SomeRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, "参数错误")
        return
    }
    // 2. 获取用户ID（从JWT Context）
    userID, _ := c.Get("user_id")
    // 3. 调用Service
    result, err := h.svc.SomeMethod(c.Request.Context(), userID.(uint), &req)
    if err != nil {
        response.Fail(c, http.StatusBadRequest, err.Error())
        return
    }
    // 4. 返回成功响应
    response.Success(c, result)
}
```

| Handler | 关键方法 |
|---------|---------|
| AuthHandler | Login、Register、GetUserInfo |
| UserHandler | List、Create、Update、Delete、AssignRole |
| RoleHandler | List、Create、Update、Delete、AssignPermission |
| PermissionHandler | List、Create、Update、Delete |
| DashboardHandler | GetStats |
| AuditHandler | Submit、Withdraw、Review、ListMyApplications、ListAllApplications、GetDetail、GetUnreadCount、GetPendingCount、ListMessages、MarkMessageRead、MarkAllRead |
| ResourceHandler | ListResources、ListActiveResources、GetResource、CreateResource、UpdateResource、DeleteResource |
| WsHandler | Connect（WebSocket升级） |

---

## 第8章 公共包 pkg 详解

### 8.1 response 统一响应封装

[response.go](file:///d:/Programming/Agent_demo/casbin_demo/backend/pkg/response/response.go) 提供统一JSON响应格式。

**核心原则**：HTTP状态码与响应体`code`字段严格一致。

```go
type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}
```

| 方法 | HTTP状态码 | code | 用途 |
|-----|-----------|------|------|
| Success(c, data) | 200 | 200 | 成功返回数据 |
| SuccessWithMessage(c, msg, data) | 200 | 200 | 成功+自定义消息 |
| Fail(c, code, msg) | code | code | 通用失败（code=HTTP状态码） |
| Unauthorized(c, msg) | 401 | 401 | 未认证 |
| Forbidden(c, msg) | 403 | 403 | 无权限 |
| BadRequest(c, msg) | 400 | 400 | 参数错误 |
| NotFound(c, msg) | 404 | 404 | 资源不存在 |
| MethodNotAllowed(c, msg) | 405 | 405 | 方法不允许 |
| ServerError(c, msg) | 500 | 500 | 服务器内部错误 |

### 8.2 jwt JWT工具包

[jwt.go](file:///d:/Programming/Agent_demo/casbin_demo/backend/pkg/jwt/jwt.go) 使用HS256算法签名。

```go
type Claims struct {
    UserID   uint     `json:"user_id"`
    UUID     string   `json:"uuid"`
    Username string   `json:"username"`  // Casbin策略sub使用
    Roles    []string `json:"roles"`
    jwt.RegisteredClaims  // 标准声明（ExpiresAt/IssuedAt/NotBefore/Issuer）
}
```

- `GenerateToken(userID, uuid, username, roles)`：使用config.JWT.Secret签名，有效期由ExpireHours配置（默认24h）
- `ParseToken(tokenString)`：验证签名和有效期，返回Claims

### 8.3 casbin 权限引擎

[casbin.go](file:///d:/Programming/Agent_demo/casbin_demo/backend/pkg/casbin/casbin.go) 初始化Casbin Enforcer，使用gorm-adapter将策略持久化到PostgreSQL的casbin_rule表。

全局变量`Enforcer`在fx.Invoke的`initCasbin`中初始化，种子数据时自动添加admin用户的所有权限策略和user用户的基础权限策略。

RBAC模型定义（model.conf）：
- sub = username（用户名）
- obj = URL路径（如/api/users）
- act = HTTP方法（GET/POST/PUT/DELETE）
- 角色继承：g(r, p) 定义用户-角色关系

### 8.4 cache 两级缓存包

已在第5章详细解析，包含cache.go、local.go、bloom.go、circuit.go四个文件。

### 8.5 mq RabbitMQ消息队列（完整代码解析）

[rabbitmq.go](file:///d:/Programming/Agent_demo/casbin_demo/backend/pkg/mq/rabbitmq.go) 实现可靠消息投递。

#### 8.5.1 架构：Fanout广播模式

```
                          ┌──────────┐
Publish ───────────────►│ Fanout   ├─────► Instance-1 Queue (exclusive, auto-delete)
                         │ Exchange ├─────► Instance-2 Queue (exclusive, auto-delete)
                         │audit.fanout├───► Instance-3 Queue ...
                         └──────────┘
```

**为什么用fanout而不是shared queue**：shared queue模式下多实例竞争消费，只有一个实例拿到消息，如果该实例没有对应在线用户则消息无法实时推送。fanout模式每个实例都收到消息，各自推送给本地连接的用户，确保不丢消息。

#### 8.5.2 拓扑设计

[setupTopology](file:///d:/Programming/Agent_demo/casbin_demo/backend/pkg/mq/rabbitmq.go#L208-L263) 声明：

1. **死信交换机(DLX)**：`audit.dlx`（direct类型，持久化）
2. **死信队列(DLQ)**：`audit.dlq`（quorum队列，持久化），绑定到DLX
3. **业务交换机**：`audit.fanout`（fanout类型，持久化），配置x-dead-letter-exchange指向DLX
4. **实例独占队列**：`audit.instance.{hostname}-{random}`（exclusive + auto-delete），绑定到fanout交换机，配置死信参数

队列使用**quorum队列**（Raft仲裁）比普通镜像队列更可靠，适合关键业务消息。

#### 8.5.3 可靠投递七重保障

| 保障 | 实现 |
|-----|------|
| ① 消息持久化 | `DeliveryMode: amqp.Persistent`，交换机/队列durable=true |
| ② 手动ACK | `Consume(autoAck=false)`，业务处理成功才Ack |
| ③ 失败重试 | `Nack(requeue=true)` 重新入队，失败进入DLQ |
| ④ 死信队列 | 超过重试次数的消息进入DLQ，不会丢失 |
| ⑤ 自动重连 | `watchConnection`监听连接断开，指数退避重连 |
| ⑥ 轻量消息 | MQ只传message_id，详情从PG加载，减小MQ压力 |
| ⑦ 定时对账 | 每30s扫描mq_delivered=false的消息重新Publish |

#### 8.5.4 消息消费流程

[processMessage](file:///d:/Programming/Agent_demo/casbin_demo/backend/pkg/mq/rabbitmq.go#L310-L336)：
1. 解析JSON消息体
2. 通过`handler`原子指针获取业务处理函数（SetHandler注入）
3. 调用handler处理（handler中查PG→WS推送→标记mq_delivered=true）
4. 成功→Ack；失败→sleep 500ms后Nack(requeue=true)重试
5. JSON解析失败→Nack(requeue=false)直接进DLQ（避免无限循环）

#### 8.5.5 自动重连机制

[reconnect](file:///d:/Programming/Agent_demo/casbin_demo/backend/pkg/mq/rabbitmq.go#L181-L205)：
- 最多重试30次
- 延迟递增：`3s * min(attempt, 10)`（3s, 6s, ..., 30s）
- 每次重连依次执行：connect() → setupTopology() → startConsumer()
- 超过最大重试次数后放弃，等待定时对账补发兜底

#### 8.5.6 Close优雅关闭

```go
func (c *Client) Close() error {
    c.closed.Store(true)
    close(c.stopCh)
    c.wg.Wait()  // 等待消费者goroutine退出（处理完在途消息）
    // 关闭channel和connection
}
```

### 8.6 ws WebSocket实时通信（完整代码解析）

[hub.go](file:///d:/Programming/Agent_demo/casbin_demo/backend/pkg/ws/hub.go) 实现WebSocket Hub，支持多实例通过Redis PubSub广播。

#### 8.6.1 Hub架构

```
单实例：  Browser ←→ Client ←→ Hub (channel通信)
多实例：  Browser ←→ Client ←→ Hub ←→ Redis PubSub ←→ 其他实例Hub ←→ Client
```

Hub通过Go channel实现线程安全的事件驱动，避免显式锁竞争（注册/注销/广播都通过channel发送到eventLoop单协程处理）。

#### 8.6.2 核心数据结构

```go
type Hub struct {
    clients    map[*Client]bool                // 所有连接
    userConns  map[uint]map[*Client]bool       // 用户ID→连接集合（支持多端同时在线）
    roomConns  map[string]map[*Client]bool     // 房间→连接集合（admin房间）
    register   chan *Client                    // 注册channel
    unregister chan *Client                    // 注销channel
    broadcast  chan []byte                     // 广播channel
    direct     chan *directMessage             // 用户直推channel
    room       chan *roomMessage               // 房间广播channel
    rdb        *redis.Client                   // Redis客户端（PubSub跨实例）
    dedupLL    *list.List                      // LRU去重链表
    dedupMap   map[string]*list.Element        // LRU去重索引
    // ...wg, stopCh, mu等
}
```

#### 8.6.3 事件循环

[eventLoop](file:///d:/Programming/Agent_demo/casbin_demo/backend/pkg/ws/hub.go#L265-L282) 单协程处理所有事件，无需加锁：

```go
func (h *Hub) eventLoop() {
    for {
        select {
        case client := <-h.register:   h.handleRegister(client)
        case client := <-h.unregister: h.removeClient(client)
        case message := <-h.broadcast: h.localBroadcastAll(message)
        case dm := <-h.direct:         h.sendToUserLocal(dm.UserID, dm.Data)
        case rm := <-h.room:           h.sendToRoomLocal(rm.Room, rm.Data)
        case <-h.stopCh:               return
        }
    }
}
```

#### 8.6.4 连接管理与单用户连接数限制

[handleRegister](file:///d:/Programming/Agent_demo/casbin_demo/backend/pkg/ws/hub.go#L285-L321)：
- 检查用户现有连接数，超过MaxConnsPerUser(5)时关闭最旧连接
- 注册到clients、userConns
- admin角色用户自动加入admin房间

#### 8.6.5 心跳机制（ReadPump/WritePump）

[ReadPump](file:///d:/Programming/Agent_demo/casbin_demo/backend/pkg/ws/hub.go#L557-L588)：
- 设置ReadDeadline为60s（PongWait）
- PongHandler重置ReadDeadline（收到客户端Pong说明连接活跃）
- 接收客户端文本"ping"，回复"pong"
- MaxMessageSize=1024字节（防止大消息攻击）

[WritePump](file:///d:/Programming/Agent_demo/casbin_demo/backend/pkg/ws/hub.go#L591-L632)：
- 每25s（PingPeriod）发送Ping控制帧
- 从Send channel读取消息批量写入（NextWriter+合并写）
- SetWriteDeadline=10s防止慢连接阻塞

#### 8.6.6 Redis PubSub跨实例广播

[subscribeRedis](file:///d:/Programming/Agent_demo/casbin_demo/backend/pkg/ws/hub.go#L171-L194)：
- 订阅固定频道`ws:broadcast`
- 收到消息后反序列化为WsMessage
- 幂等去重检查后推送给本地连接
- PubSub连接断开时自动重连（resubscribeRedis，最多10次）

[publishToRedis](file:///d:/Programming/Agent_demo/casbin_demo/backend/pkg/ws/hub.go#L519-L528)：3s超时，失败仅Warn不阻断业务。

**关键**：`SendToAdminsLocal`和`SendToUserLocal`只推本地，不经过Redis PubSub（MQ消费端使用，避免消息循环）；`SendToUser`和`BroadcastToAdmins`先推本地再PubSub跨实例（业务直接调用时使用）。

#### 8.6.7 幂等去重机制（LRU缓存）

[isDuplicate](file:///d:/Programming/Agent_demo/casbin_demo/backend/pkg/ws/hub.go#L357-L386) 使用LRU（Least Recently Used）缓存：
- DedupWindow=5分钟：同一ID消息5分钟内不重复弹窗
- DedupMaxEntries=10000：最多缓存1万条消息ID
- 使用list.List+map实现O(1)查找和移动
- 新消息插入链表头部，超容量从尾部淘汰最旧的
- 每分钟定时清理过期条目

---

## 第9章 Casbin权限模型详解

### 9.1 RBAC权限模型设计

系统使用Casbin的RBAC模型，核心元素：

| 元素 | 含义 | 对应值 |
|-----|------|-------|
| sub (subject) | 主体（谁） | 用户名（如admin、user） |
| obj (object) | 客体（什么资源） | API路径（如/api/users） |
| act (action) | 动作（什么操作） | HTTP方法（GET、POST、PUT、DELETE） |
| g (group) | 角色继承 | 用户→角色映射（如admin属于admin角色） |

策略规则格式：`p, sub, obj, act` 或 `g, user, role`

示例策略：
```
p, admin, /api/users, GET       # admin可以GET /api/users
p, admin, /api/users, POST      # admin可以POST /api/users
p, user, /api/dashboard, GET    # user可以GET /api/dashboard
g, admin, admin                 # admin用户拥有admin角色
g, user, user                   # user用户拥有user角色
```

### 9.2 模型定义文件

[model.conf](file:///d:/Programming/Agent_demo/casbin_demo/backend/pkg/casbin/model.conf)：

```ini
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
```

Matcher解释：
- `g(r.sub, p.sub)`：请求的用户是否继承（属于）策略中定义的角色
- `r.obj == p.obj`：请求路径精确匹配策略路径
- `r.act == p.act`：请求方法精确匹配策略方法
- 三个条件同时满足才放行

### 9.3 权限校验完整流程

```
HTTP请求进入CasbinAuth中间件
    │
    ├─ c.Get("username") → sub（如"admin"）
    ├─ c.Request.URL.Path → obj（如"/api/users"）
    ├─ c.Request.Method → act（如"GET"）
    │
    └─ Enforcer.Enforce(sub, obj, act)
        │
        ├─ Casbin加载策略（从casbin_rule表，gorm-adapter）
        ├─ 匹配g规则：用户是否拥有某角色
        ├─ 匹配p规则：角色是否有(obj, act)的权限
        │
        ├─ true → c.Next() 放行
        └─ false → 403 Forbidden
```

### 9.4 策略管理与初始化

种子数据初始化时（[database.go](file:///d:/Programming/Agent_demo/casbin_demo/backend/internal/repository/database.go)的SeedData）：

1. 创建admin/user角色
2. 创建admin/user账号（bcrypt密码：admin123/user123）
3. 创建所有权限条目（对应每个管理员API）
4. 为admin角色分配所有权限p策略
5. 为user角色分配基础权限（dashboard、资源浏览、个人消息等）
6. 添加g策略（用户-角色关系）

角色权限变更时（RoleService.AssignPermission），同步更新Casbin策略并ReloadEnforcer。

---

## 第10章 审核系统业务架构

### 10.1 业务流程概述

审核系统是平台核心业务，完整流程：

```
[用户提交申请]
    │
    ├─ 用户填写资源申请表单（资源名、类型、API、用途、QPS等）
    ├─ POST /api/audit/applications
    │
    ├─ 后端AuditService.SubmitApplication:
    │   ├─ 查询申请人信息
    │   ├─ PG事务：写入申请(pending) + 批量写入管理员消息
    │   ├─ 事务COMMIT（关键！先落盘再发消息）
    │   ├─ Redis设置2分钟撤回TTL
    │   ├─ 失效相关缓存（待审数、申请列表、未读数）
    │   ├─ MQ Fanout广播通知所有实例
    │   └─ 本地WS快速推送在线管理员
    │
    └─ 管理员实时收到WebSocket弹窗通知

[2分钟撤回窗口]
    │
    ├─ 用户在2分钟内可撤回（Redis TTL优先校验）
    ├─ Redis不可用时PG时间戳兜底
    ├─ 撤回使用乐观锁防止并发修改
    │
    └─ 撤回后：管理员收到撤回通知，申请状态变为"已撤回"

[管理员审核]
    │
    ├─ 管理员查看待审列表 → 点击审核
    ├─ POST /api/audit/applications/:id/review
    │
    ├─ 后端AuditService.ReviewApplication:
    │   ├─ 校验申请存在且状态为pending
    │   ├─ PG事务：乐观锁更新状态(approved/rejected) + 写入申请人消息
    │   ├─ 事务COMMIT
    │   ├─ 删除撤回TTL
    │   ├─ MQ通知申请人
    │   └─ WS推送审核结果给申请人
    │
    └─ 申请人收到WebSocket弹窗通知审核结果
```

### 10.2 状态机设计

```
            用户提交
  ┌──────────────────────────┐
  │                          ▼
  │                    ┌───────────┐
  │    ┌──────────────►│  Pending  │◄──────────────┐
  │    │ 撤回窗口内     │ (待审核)   │               │
  │    │               └─────┬─────┘               │
  │    │                     │                     │
  │    │            ┌────────┴────────┐            │
  │    │            │                 │            │
  │    │     管理员通过          管理员驳回          │
  │    │            │                 │            │
  │    │            ▼                 ▼            │
  │    │      ┌───────────┐    ┌───────────┐      │
  │    │      │ Approved  │    │ Rejected  │      │
  │    │      │ (已通过)   │    │ (已驳回)   │      │
  │    │      └───────────┘    └───────────┘      │
  │    │                                           │
  │    │ 撤回                               (终态，不可再操作)
  │    │                                           │
  │    └──────────┐                      ┌─────────┘
  │               ▼                      │
  │         ┌───────────┐                │
  └─────────│ Withdrawn │────────────────┘
            │ (已撤回)   │
            └───────────┘
                 │
                 └──► 终态
```

状态转换规则：
- Pending → Approved：管理员审核通过
- Pending → Rejected：管理员审核驳回
- Pending → Withdrawn：用户在2分钟窗口内撤回
- Approved/Rejected/Withdrawn：终态，不可再变更

### 10.3 2分钟撤回机制（代码级详解）

[canWithdraw](file:///d:/Programming/Agent_demo/casbin_demo/backend/internal/service/audit_service.go#L78-L90) 双重校验：

```go
func (s *AuditService) canWithdraw(app *model.AuditApplication) bool {
    if app.Status != model.AuditStatusPending {
        return false  // 非待审核状态不可撤回
    }
    // 第一层：Redis TTL校验（优先，精确控制）
    if s.rdb != nil {
        key := fmt.Sprintf("%s%d", withdrawCachePrefix, app.ID)
        exists, err := s.rdb.Exists(context.Background(), key).Result()
        if err == nil && exists > 0 {
            return true  // Redis key存在，窗口有效
        }
    }
    // 第二层：PG时间戳兜底（Redis不可用时降级）
    return time.Since(app.CreatedAt) < withdrawWindowDuration
}
```

**为什么用Redis TTL而不是前端计时**：
- 前端时间可篡改（修改系统时间绕过限制）
- Redis TTL由服务端精确控制，SET时EX=120s，到期自动删除
- Redis不可用时PG创建时间兜底，保证功能可用

撤回窗口流程：
1. 提交申请时：`setWithdrawTTL(appID)` → Redis SETEX key 120s
2. 查询详情时：`canWithdraw(app)` → 返回CanWithdraw和WithdrawRemain毫秒数
3. 撤回时：`canWithdraw(app)`校验 → 乐观锁更新状态
4. 撤回成功/审核完成后：`delWithdrawTTL(appID)` → 主动删除key

### 10.4 乐观锁并发控制

审核和撤回操作都使用乐观锁防止并发修改：

[UpdateStatusWithVersion](file:///d:/Programming/Agent_demo/casbin_demo/backend/internal/repository/audit_repo.go) 执行SQL等效于：
```sql
UPDATE api_audit_applications 
SET status = ?, reviewer_id = ?, version = version + 1, updated_at = ?
WHERE id = ? AND version = ? AND status = 0;
```

如果UPDATE影响行数为0，说明：
- 版本已变（其他协程已修改）→ 返回`ErrOptimisticLockConflict`
- 状态不是pending（已被处理）→ 返回`ErrNotPending`

Service层捕获错误返回友好提示："申请状态已变更，请刷新后重试"。

### 10.5 事务边界设计

**核心原则：事务内不做Redis/MQ网络IO**。

SubmitApplication事务边界：
```go
tx := s.auditRepo.BeginTx()
// ─── 事务内操作（只有DB操作）───
s.auditRepo.Create(tx, app)                    // 写入申请
s.auditRepo.CreateMessagesBatch(tx, adminMsgs) // 写入管理员消息
tx.Commit()  // ← 提交事务！

// ─── 事务外操作（Redis/MQ/WS，事务已提交）───
s.setWithdrawTTL(app.ID)           // Redis设置撤回TTL
s.mqNotifyAdmins(ctx, m)           // 发送MQ消息
s.wsHub.SendToAdminsLocal(...)     // 本地WS推送
s.invalidateAuditCaches()          // 失效缓存
```

**为什么先提交事务再发MQ**：如果先发MQ再提交事务，MQ消费者可能在事务提交前就收到消息并查PG，查不到数据导致错误；更严重的是如果事务回滚，MQ消息已经发出去了（幽灵消息），用户看到通知但数据不存在。

### 10.6 AuditService核心方法逐行解析

#### SubmitApplication（提交申请）

[audit_service.go#L214-L320](file:///d:/Programming/Agent_demo/casbin_demo/backend/internal/service/audit_service.go#L214-L320)：

1. **L215-222**：查询申请人信息，Nickname为空则用Username
2. **L224-239**：构造AuditApplication实体，生成UUID，状态pending，设置CreatedAt/UpdatedAt
3. **L241-245**：查询所有管理员ID（用于创建通知消息）
4. **L247-283**：**PG事务块**
   - BeginTx开启事务
   - Create写入申请记录
   - 循环构造管理员SysMessage列表（每个管理员一条）
   - CreateMessagesBatch批量写入消息
   - Commit提交事务，失败则Rollback
5. **L285**：Redis设置2分钟撤回TTL
6. **L286-288**：遍历管理员消息，MQ异步广播
7. **L290-312**：本地WS快速推送（不等待MQ，本实例管理员立即收到通知）
8. **L314-315**：失效相关缓存（待审数、申请列表、消息未读数）

#### ReviewApplication（审核申请）

[audit_service.go#L322-L422](file:///d:/Programming/Agent_demo/casbin_demo/backend/internal/service/audit_service.go#L322-L422)：

1. **L323-329**：查询申请，校验状态为pending
2. **L330-337**：查询审核人信息
3. **L339-349**：根据Approved确定状态和通知内容
4. **L351-389**：**PG事务块**
   - UpdateStatusWithVersion乐观锁更新状态
   - 创建申请人的SysMessage（审核结果通知）
   - Commit提交事务
5. **L391**：删除撤回TTL（申请已被处理，不能再撤回）
6. **L392**：MQ通知申请人
7. **L394-416**：WS推送审核结果给申请人
8. **L418-419**：失效缓存

#### WithdrawApplication（撤回申请）

[audit_service.go#L424-L542](file:///d:/Programming/Agent_demo/casbin_demo/backend/internal/service/audit_service.go#L424-L542)：

1. **L425-437**：校验申请存在、所有权（ApplicantID==userID）、状态pending、撤回窗口有效
2. **L439-488**：**PG事务块**
   - WithdrawApplication（乐观锁更新状态为withdrawn+撤回原因+撤回时间）
   - 批量创建管理员撤回通知消息
   - Commit提交
3. **L490**：删除撤回TTL
4. **L491-517**：MQ通知管理员 + WS推送管理员
5. **L519-536**：WS给申请人推送撤回成功确认通知
6. **L538-539**：失效缓存

#### HandleMQMessage（MQ消息消费处理）

[audit_service.go#L666-L713](file:///d:/Programming/Agent_demo/casbin_demo/backend/internal/service/audit_service.go#L666-L713)：

1. 根据TargetType（admins/user）查询PG获取完整消息
2. 构造WsNotification
3. WS推送给对应目标（管理员房间/指定用户）
4. 标记消息mq_delivered=true
5. 失效未读数缓存

#### StartRetryScheduler + retryUndelivered（定时对账补发）

[audit_service.go#L724-L791](file:///d:/Programming/Agent_demo/casbin_demo/backend/internal/service/audit_service.go#L724-L791)：

- 每30s执行一次
- 查询mq_delivered=false的消息（最多50条）
- 重新Publish到MQ
- 成功后标记delivered
- 兜底极端情况：MQ当时不可用、消息丢失、消费者崩溃等

---

## 第11章 RabbitMQ消息队列

（注：核心实现在8.5节已详解，本节补充部署和配置细节）

### 11.1 Exchange与Queue规划

| 名称 | 类型 | 持久化 | 用途 |
|-----|------|-------|------|
| audit.fanout | fanout | 是 | 业务通知广播交换机 |
| audit.dlx | direct | 是 | 死信交换机 |
| audit.instance.{tag} | exclusive+auto-delete | quorum | 每个实例独占消费队列 |
| audit.dlq | durable | quorum | 死信队列，存放处理失败的消息 |

### 11.2 Prefetch与QoS

```go
ch.Qos(PrefetchCount, 0, false)  // PrefetchCount=20
```

每个消费者预取20条消息，避免消息堆积在单消费者但其他消费者空闲。

### 11.3 消息体设计

```go
type NotificationMessage struct {
    MessageID      uint      `json:"message_id"`     // 唯一标识，PG查详情
    TargetType     string    `json:"target_type"`    // "admins"或"user"
    TargetID       uint      `json:"target_id"`      // 指定用户ID（TargetType=user时）
    Type           string    `json:"type"`           // 消息类型
    BusinessID     uint      `json:"business_id"`    // 业务ID
    BusinessType   string    `json:"business_type"`  // 业务类型
    CreatedAt      time.Time `json:"created_at"`
    IdempotencyKey string    `json:"idempotency_key"` // 幂等键（用message_id）
}
```

**MQ消息只传ID不传详情**，减小MQ带宽和内存压力，详情从PG查询获取最新数据。

---

## 第12章 WebSocket实时通信

（注：核心实现在8.6节已详解，本节补充消息推送策略和连接协议）

### 12.1 WebSocket连接协议

**连接端点**：`GET /ws?token=<jwt_token>`

- WsHandler.Connect方法从query参数获取token
- 验证JWT有效性
- 升级HTTP连接为WebSocket
- 创建Client并注册到Hub
- 启动ReadPump和WritePump协程

**消息格式**（JSON）：

客户端→服务端：
```json
{"type": "ping"}          // 心跳
```

服务端→客户端：
```json
{
  "type": "notification",
  "target_type": "room",
  "room": "admin",
  "data": {
    "type": "new_application",
    "title": "新的审核申请",
    "content": "张三提交了...",
    "business_type": "audit_application",
    "business_id": 123,
    "timestamp": "2026-06-26T10:00:00Z",
    "id": "new-app-123"
  },
  "timestamp": "2026-06-26T10:00:00Z"
}
```

### 12.2 消息推送策略

| 方法 | 本地推送 | Redis PubSub | 使用场景 |
|-----|---------|-------------|---------|
| SendToAdminsLocal | ✅ | ❌ | MQ消费端使用（避免循环广播） |
| SendToUserLocal | ✅ | ❌ | MQ消费端使用 |
| SendToUser | ✅ | ✅ | 业务代码直接调用 |
| BroadcastToAdmins | ✅ | ✅ | 业务代码直接调用 |

**防循环**：MQ消费时使用Local版本不经过PubSub，因为MQ已经通过Fanout确保所有实例都收到了消息。

### 12.3 单用户连接数限制

MaxConnsPerUser=5，同一用户最多5个WebSocket连接（多端同时在线）。超出时关闭最旧连接（FIFO策略），保护服务器内存。

---

## 第13章 前端架构详解

### 13.1 技术选型与版本

前端基于Vue 3.5 + TypeScript + Vite 7 + Pinia 3 + Vue Router 4 + Element Plus 2 + Axios。

### 13.2 应用入口 main.ts

创建Vue应用实例，注册插件（Element Plus、Pinia、Router），挂载到#app。

### 13.3 根组件 App.vue

包含`<router-view/>`作为路由出口，引入全局样式。

### 13.4 状态管理 Pinia（user.ts详解）

[user.ts](file:///d:/Programming/Agent_demo/casbin_demo/frontend/src/store/user.ts) 使用Composition API风格defineStore：

```ts
export const useUserStore = defineStore('user', () => {
  const token = ref<string>(localStorage.getItem('token') || '')
  const userInfo = ref<UserInfo | null>(null)

  const setToken = (newToken: string) => {
    token.value = newToken
    localStorage.setItem('token', newToken)  // 持久化到localStorage
  }

  const login = async (params: LoginParams) => {
    const res = await loginApi(params)
    setToken(res.data.token)
    userInfo.value = res.data.user
    return res.data
  }

  const fetchUserInfo = async () => {
    const res = await getUserInfo()
    userInfo.value = res.data
    return res.data
  }

  const logout = () => {
    token.value = ''
    userInfo.value = null
    localStorage.removeItem('token')
  }

  const hasRole = (role: string) => userInfo.value?.roles.includes(role) ?? false

  return { token, userInfo, login, fetchUserInfo, logout, hasRole }
})
```

**Token持久化**：存储在localStorage，刷新页面自动恢复登录态。

### 13.5 路由系统 Vue Router（index.ts详解）

[index.ts](file:///d:/Programming/Agent_demo/casbin_demo/frontend/src/router/index.ts) 实现路由守卫权限控制：

```ts
router.beforeEach(async (to, from, next) => {
  NProgress.start()
  const token = localStorage.getItem('token')
  const userStore = useUserStore()

  // 错误页面直接放行
  if (['/401','/403','/404','/500'].includes(to.path)) { next(); return }

  // 需要认证的路由：无token → 跳登录
  if (to.meta.requiresAuth !== false && !token) {
    next(to.path === '/login' ? '/' : '/login')
    return
  }

  // 已登录访问登录页 → 跳首页
  if (to.path === '/login' && token) { next('/dashboard'); return }

  // 需要认证且有token：预加载用户信息
  if (to.meta.requiresAuth !== false && token) {
    if (!userStore.userInfo) {
      try {
        await userStore.fetchUserInfo()
      } catch {
        userStore.logout()
        next('/login')
        return
      }
    }
    // 管理员路由：非admin → 403
    if (to.meta.adminOnly && !userStore.hasRole('admin')) {
      next('/403')
      return
    }
  }
  next()
})
```

**路由meta字段**：
- `requiresAuth: false`：公开页面（登录页、错误页）
- `adminOnly: true`：仅管理员可访问
- 默认：所有Layout下的路由都需要认证

### 13.6 Axios请求封装 request.ts

创建Axios实例，配置baseURL、超时、请求/响应拦截器：
- 请求拦截器：自动添加`Authorization: Bearer <token>`头
- 响应拦截器：401自动跳登录页，统一错误提示

### 13.7 API接口层

| 文件 | 对应后端 | 接口 |
|-----|---------|------|
| auth.ts | AuthHandler | login、register、getUserInfo |
| user.ts | UserHandler | 用户CRUD、分配角色 |
| rbac.ts | RoleHandler+PermissionHandler | 角色/权限CRUD、分配 |
| resource.ts | ResourceHandler | 资源CRUD、列表、详情 |
| audit.ts | AuditHandler | 申请提交/撤回/审核、消息列表/已读、未读数/待审数 |

### 13.8 布局组件 Layout.vue

主布局容器包含：
- **侧边栏（Sidebar）**：根据角色显示不同菜单（admin琥珀金色系/user靛蓝色系）
- **顶栏（Header）**：通知铃铛、用户信息、角色标签、退出按钮
- **内容区**：`<router-view/>`渲染页面
- **NProgress**：页面加载进度条

管理员菜单：仪表盘、资源管理、审核管理、消息通知、用户管理、角色管理、权限管理
普通用户菜单：仪表盘、资源清单、申请资源、我的申请、消息通知

### 13.9 双角色UI设计体系

| 角色 | 色系 | CSS变量 | 菜单风格 |
|-----|------|--------|---------|
| 管理员 | 琥珀金色 | #f59e0b ~ #f97316 | 深色侧边栏渐变+金色边框 |
| 普通用户 | 靛蓝色 | #6366f1 ~ #06b6d4 | 科技蓝紫渐变 |

### 13.10 WebSocket客户端（websocket.ts详解）

[websocket.ts](file:///d:/Programming/Agent_demo/casbin_demo/frontend/src/utils/websocket.ts) 实现完整的WebSocket客户端：

```ts
class WebSocketService {
  private ws: WebSocket | null = null
  private reconnectAttempts = 0
  private maxReconnectAttempts = 10
  private heartbeatTimer: number | null = null
  private handlers: Map<string, MessageHandler[]> = new Map()
  private shownMessages: Map<string, number> = new Map()  // 前端去重

  connect(token: string) { ... }     // 建立连接，wss/ws自动判断
  private doConnect() { ... }        // 创建WS实例，绑定事件
  private scheduleReconnect() { ... } // 指数退避重连（3s,6s,...15s）
  private startHeartbeat() { ... }   // 每25s发送"ping"心跳
  private handleMessage(msg) { ... } // 消息分发+ElNotification弹窗+去重
  private isDuplicate(msgId) { ... } // 5分钟窗口去重
  on(event, handler) { ... }         // 注册消息处理器
  disconnect() { ... }               // 主动断开（登出时）
}
```

**通知弹窗**：
- 新申请：ElNotification info类型，点击跳转到审核页
- 审核结果：success（通过）/warning（驳回），点击跳转到我的申请
- 申请撤回：warning类型
- 弹窗后自动fetchCounts刷新未读数/待审数

### 13.11 页面组件详解

| 页面 | 角色 | 功能 |
|-----|------|------|
| Login.vue | 所有人 | 登录/注册双Tab切换 |
| Dashboard.vue | 所有登录用户 | 管理员：用户/角色/权限/待审统计；用户：个人申请统计+使用指南 |
| Users.vue | 管理员 | 用户列表、创建、编辑、删除、分配角色 |
| Roles.vue | 管理员 | 角色列表、创建、编辑、删除、分配权限 |
| Permissions.vue | 管理员 | 权限列表、创建、编辑、删除 |
| ResourceList.vue | 所有登录用户 | 资源清单浏览，分类筛选、搜索、分页 |
| ApplyResource.vue | 普通用户 | 填写申请表单提交 |
| MyApplications.vue | 普通用户 | 个人申请列表，撤回操作，查看详情 |
| AuditList.vue | 管理员 | 所有申请列表，审核操作（通过/驳回+意见） |
| Messages.vue | 所有登录用户 | 时间线样式消息列表，已读/未读，全部已读，跳转到业务页 |

### 13.12 错误页面系统

通用错误组件 [ErrorPage.vue](file:///d:/Programming/Agent_demo/casbin_demo/frontend/src/components/ErrorPage.vue) 支持401/403/404/500四种错误码，根据token是否存在决定"返回首页"的跳转目标（有token→dashboard，无token→login），修复了之前403页面的重定向循环问题。

---

## 第14章 数据库设计

### 14.1 ER关系图

```
┌──────┐       ┌───────────┐       ┌────────────┐
│ users│──────<│ user_roles │>──────│   roles    │
└──┬───┘       └───────────┘       └─────┬──────┘
   │                                     │
   │                              ┌──────┴───────┐
   │                              │role_permissions│
   │                              └──────┬───────┘
   │                                     │
   │                                     ▼
   │                              ┌────────────┐
   │                              │permissions │
   │                              └────────────┘
   │
   ├─────────────────< api_audit_applications (申请人)
   │
   └─────────────────< sys_messages (接收人)
                              │
                              └── api_audit_applications (business关联)

┌──────────────────┐
│  casbin_rule     │ ← Casbin策略持久化（gorm-adapter自动管理）
└──────────────────┘
┌──────────────────┐
│  api_resources   │ ← API资源清单
└──────────────────┘
```

### 14.2 users 用户表（字段逐列详解）

| 字段 | 类型 | 约束 | 说明 |
|-----|------|------|------|
| id | uint | PK, autoinc | 自增主键（内部使用，不对外暴露） |
| uuid | varchar(36) | unique, not null | 对外暴露的唯一标识 |
| username | varchar(50) | unique, not null | 登录用户名 |
| password | varchar(255) | not null | bcrypt哈希（cost=10） |
| nickname | varchar(50) | - | 昵称 |
| email | varchar(100) | - | 邮箱 |
| avatar | varchar(255) | - | 头像URL |
| status | int | default 1 | 1=启用, 0=禁用 |
| created_at | timestamp | - | 创建时间 |
| updated_at | timestamp | - | 更新时间 |
| deleted_at | timestamp | index, null | 软删除时间 |

### 14.3 roles 角色表

| 字段 | 类型 | 约束 | 说明 |
|-----|------|------|------|
| id | uint | PK | 主键 |
| name | varchar(50) | unique, not null | 角色标识（admin/user），Casbin g策略使用 |
| label | varchar(100) | - | 显示名称（管理员/普通用户） |
| description | varchar(255) | - | 描述 |
| status | int | default 1 | 状态 |

### 14.4 permissions 权限表

| 字段 | 类型 | 约束 | 说明 |
|-----|------|------|------|
| id | uint | PK | 主键 |
| name | varchar(100) | unique, not null | 权限标识（如user:list） |
| label | varchar(100) | - | 显示名称 |
| path | varchar(255) | - | API路径，Casbin obj匹配 |
| method | varchar(20) | - | HTTP方法，Casbin act匹配 |

### 14.5 user_roles 用户角色关联表

| 字段 | 类型 | 说明 |
|-----|------|------|
| user_id | uint | 用户ID（FK→users.id） |
| role_id | uint | 角色ID（FK→roles.id） |

联合主键：(user_id, role_id)

### 14.6 role_permissions 角色权限关联表

| 字段 | 类型 | 说明 |
|-----|------|------|
| role_id | uint | 角色ID（FK→roles.id） |
| permission_id | uint | 权限ID（FK→permissions.id） |

联合主键：(role_id, permission_id)

### 14.7 casbin_rule Casbin策略表

gorm-adapter自动管理，字段：
| 字段 | 说明 |
|-----|------|
| id | 主键 |
| ptype | 策略类型（p/g） |
| v0 | sub（角色名 或 用户名） |
| v1 | obj（API路径） |
| v2 | act（HTTP方法） |

### 14.8 api_resources 资源表

| 字段 | 类型 | 说明 |
|-----|------|------|
| id | uint PK | 主键 |
| uuid | varchar(36) unique | UUID |
| name | varchar(200) not null | 资源名称 |
| type | varchar(50) index | 资源类型 |
| api_name | varchar(200) not null | API名 |
| description | text | 描述 |
| provider | varchar(100) | 厂商 |
| version | varchar(50) | 版本 |
| default_qps | int default 10 | 默认QPS |
| max_qps | int default 100 | 最大QPS |
| status | int index default 1 | 1=可用, 0=不可用 |
| docs_url | varchar(500) | 文档链接 |
| tags | text | JSON标签数组 |

### 14.9 api_audit_applications 审核申请表（含乐观锁）

| 字段 | 类型 | 约束 | 说明 |
|-----|------|------|------|
| id | uint | PK | 主键 |
| uuid | varchar(36) | unique | UUID |
| applicant_id | uint | idx_applicant_status | 申请人ID |
| applicant_name | varchar(50) | - | 申请人姓名（冗余） |
| resource_name | varchar(200) | not null | 资源名 |
| resource_type | varchar(50) | not null | 资源类型 |
| api_name | varchar(200) | not null | API名 |
| api_description | text | - | API描述 |
| purpose | text | not null | 申请用途 |
| expected_qps | int | default 0 | 预期QPS |
| contact_info | varchar(200) | - | 联系方式 |
| status | int | idx_applicant_status, idx_status_created | 0=待审,1=通过,2=驳回,3=撤回 |
| reviewer_id | uint | index, null | 审核人ID |
| reviewer_name | varchar(50) | - | 审核人姓名 |
| review_comment | text | - | 审核意见 |
| withdraw_reason | text | - | 撤回原因 |
| withdrawn_at | timestamp | null | 撤回时间 |
| reviewed_at | timestamp | null | 审核时间 |
| version | int | not null default 0 | **乐观锁版本号** |
| created_at | timestamp | idx_status_created | 创建时间 |

**乐观锁实现**：更新时`SET version=version+1 WHERE id=? AND version=?`，version不匹配则影响行数为0。

### 14.10 sys_messages 系统消息表（含MQ投递标记）

| 字段 | 类型 | 约束 | 说明 |
|-----|------|------|------|
| id | uint | PK | 主键 |
| uuid | varchar(36) | unique | UUID |
| receiver_id | uint | idx_receiver_read_created | 接收人ID |
| type | varchar(50) | index | 消息类型 |
| title | varchar(200) | not null | 消息标题 |
| content | text | - | 消息内容 |
| business_type | varchar(50) | idx_business | 业务类型 |
| business_id | uint | idx_business | 业务ID |
| is_read | bool | idx_receiver_read_created default false | 已读标记 |
| mq_delivered | bool | idx_mq_delivered default false | MQ投递标记（定时补发用） |
| created_at | timestamp | idx_receiver_read_created, idx_mq_delivered | 创建时间 |

### 14.11 索引设计说明（B-tree复合索引）

| 索引 | 覆盖查询 |
|-----|---------|
| idx_applicant_status (applicant_id, status) | 用户查自己的申请列表+状态筛选 |
| idx_status_created (status, created_at) | 管理员按状态筛选+时间排序 |
| idx_receiver_read_created (receiver_id, is_read, created_at) | 未读数统计+消息列表分页（覆盖索引） |
| idx_business (business_type, business_id) | 按业务反查消息 |
| idx_mq_delivered (mq_delivered, created_at) | 定时补发扫描未投递消息 |

GIN索引在当前PostgreSQL版本中用于全文搜索，本系统使用B-tree复合索引覆盖高频查询场景。

---

## 第15章 API接口文档

### 15.1 通用说明

- **Base URL**：`/api`
- **认证方式**：Bearer Token（Header: `Authorization: Bearer <jwt>`）
- **响应格式**：`{code: int, message: string, data?: any}`
- **分页参数**：`page`（从1开始）、`page_size`（1-100）
- **演示账号**：admin/admin123（管理员）、user/user123（普通用户）

### 15.2 公开接口（无需认证）

| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | /health | 健康检查（返回服务状态、版本、特性列表） |
| GET | /ws?token=xxx | WebSocket连接升级 |
| POST | /api/login | 登录（Body: {username, password}） |
| POST | /api/register | 注册（Body: {username, password, nickname?, email?}） |

POST /api/login 响应示例：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "token": "eyJhbGc...",
    "user": {
      "id": 1, "username": "admin", "nickname": "管理员",
      "roles": ["admin"], "status": 1
    }
  }
}
```

### 15.3 基础认证接口（JWT即可）

| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | /api/userinfo | 获取当前用户信息 |
| GET | /api/dashboard | 获取仪表盘数据（按角色返回不同内容） |
| POST | /api/audit/applications | 提交审核申请 |
| POST | /api/audit/applications/:id/withdraw | 撤回申请（2分钟窗口内） |
| GET | /api/audit/my-applications?page=&page_size=&status= | 我的申请列表 |
| GET | /api/audit/applications/:id | 申请详情 |
| GET | /api/messages/unread-count | 未读消息数+待审数 |
| GET | /api/messages?page=&page_size=&unread= | 消息列表 |
| PUT | /api/messages/:id/read | 标记消息已读 |
| PUT | /api/messages/read-all | 全部标记已读 |
| GET | /api/resources?page=&page_size=&type=&keyword= | 资源列表 |
| GET | /api/resources/active | 可用资源列表 |
| GET | /api/resources/:id | 资源详情 |

### 15.4 管理员接口（需Casbin权限）

| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | /api/users | 用户列表 |
| POST | /api/users | 创建用户 |
| PUT | /api/users/:id | 更新用户 |
| DELETE | /api/users/:id | 删除用户 |
| POST | /api/users/assign-role | 分配角色 |
| GET | /api/roles | 角色列表 |
| POST | /api/roles | 创建角色 |
| PUT | /api/roles/:id | 更新角色 |
| DELETE | /api/roles/:id | 删除角色 |
| POST | /api/roles/assign-permission | 分配权限 |
| GET | /api/permissions | 权限列表 |
| POST | /api/permissions | 创建权限 |
| PUT | /api/permissions/:id | 更新权限 |
| DELETE | /api/permissions/:id | 删除权限 |
| GET | /api/audit/applications?page=&page_size=&status=&applicant= | 所有申请列表 |
| POST | /api/audit/applications/:id/review | 审核申请（Body: {approved, comment}） |
| GET | /api/audit/pending-count | 待审核数量 |
| POST | /api/resources | 创建资源 |
| PUT | /api/resources/:id | 更新资源 |
| DELETE | /api/resources/:id | 删除资源 |

### 15.5 错误码说明

| HTTP状态码 | code | 含义 | 典型场景 |
|-----------|------|------|---------|
| 200 | 200 | 成功 | 请求正常处理 |
| 400 | 400 | 请求参数错误 | 参数缺失、格式错误、校验失败 |
| 401 | 401 | 未认证 | Token缺失/无效/过期 |
| 403 | 403 | 无权限 | Casbin校验失败、非所有者操作 |
| 404 | 404 | 资源不存在 | 申请/用户/资源不存在 |
| 405 | 405 | 方法不允许 | 使用错误的HTTP方法 |
| 429 | 429 | 请求过于频繁 | 触发IP/接口/用户级限流 |
| 500 | 500 | 服务器内部错误 | 数据库异常、未预期panic |

---

## 第16章 部署指南

### 16.1 Docker Compose一键部署

项目根目录[docker-compose.yml](file:///d:/Programming/Agent_demo/casbin_demo/docker-compose.yml)定义了完整服务编排：

```yaml
services:
  postgres:   # PostgreSQL 15，端口5432
  redis:      # Redis 7，端口6379
  rabbitmq:   # RabbitMQ 3-management（含管理界面），端口5672/15672
  backend:    # Go后端，端口8080，依赖PG/Redis/RabbitMQ
  frontend:   # Nginx前端，端口80/443，依赖backend
```

一键启动：
```bash
docker-compose up -d
```

启动顺序：PG/Redis/RabbitMQ先启动 → backend等待PG就绪后启动 → frontend最后启动。

### 16.2 本地开发环境搭建

**前置要求**：Go 1.26.4+、Node.js 18+、PostgreSQL 15、Redis 7、RabbitMQ 3

**后端启动**：
```bash
cd backend
go mod download
go run cmd/server/main.go
# 默认监听 :8080，自动连接PG/Redis/RabbitMQ（失败降级）
```

**前端启动**：
```bash
cd frontend
npm install
npm run dev
# 默认监听 :3000，Vite代理/api到localhost:8080
```

### 16.3 Nginx配置详解

[nginx.conf](file:///d:/Programming/Agent_demo/casbin_demo/frontend/nginx.conf) 配置要点：
- `/` → try_files $uri /index.html（SPA history模式支持）
- `/api/` → proxy_pass http://backend:8080（API反向代理）
- `/ws` → proxy_pass http://backend:8080（WebSocket反向代理，配置Upgrade头）
- gzip压缩、静态资源缓存

### 16.4 服务健康检查

- **后端**：`GET /health` 返回200 + JSON状态信息
- **Docker**：healthcheck配置定期curl /health
- **前端**：Nginx直接提供静态文件，只要Nginx运行即可

---

## 第17章 开发规范与指南

### 17.1 新增接口开发流程

1. 在model.go中定义Request/Response DTO（加binding校验标签）
2. 在Repository层添加数据访问方法（接收`*gorm.DB`支持事务）
3. 在Service层添加业务逻辑方法（事务边界清晰，缓存+MQ在事务外）
4. 在Handler层添加HTTP处理方法（参数绑定→调用Service→封装响应）
5. 在module.go中RegisterDependencies（fx自动注入，无需手动构造）
6. 在router.go对应路由组中注册路由（公开/auth/admin三层选择）
7. 如需权限控制：在SeedData中添加对应p策略，或通过权限管理API分配

### 17.2 缓存使用规范

1. **查询操作**：使用`cache.Fetch(key, opt, result, loader)`统一走两级缓存
2. **写操作**：完成DB事务后主动调用`cache.Delete/DeleteByPattern`失效相关缓存
3. **热点数据**：使用`HotDataOptions()`启用逻辑过期+异步重建
4. **空值处理**：loader返回nil时Fetch自动SetNull，无需手动处理
5. **敏感数据**：绝不在缓存中存储Password等字段，缓存DTO而非Entity
6. **认证接口**：登录不缓存结果，每次验证密码
7. **Key命名**：使用`cache.CacheKey(module, typ, id)`辅助函数生成

### 17.3 代码注释规范

所有后端代码必须包含详细中文注释：
- 包注释（package上方）：说明包的职责、设计思路
- 结构体注释：字段说明、用途
- 函数注释：职责、参数、返回值、关键逻辑说明
- 关键行注释：解释"为什么这么做"而不是"做了什么"

### 17.4 错误处理规范

1. Handler层错误返回对应的HTTP状态码（400/401/403/404/500）
2. Service层返回`error`，不直接写HTTP响应
3. Repository层返回`error`，不做业务判断
4. 错误信息面向用户（中文友好提示），详细错误通过slog记录
5. 使用errors.Is判断特定错误类型（乐观锁冲突等）
6. panic只在启动初始化阶段使用（mustXxx函数），运行时通过CustomRecovery兜底

### 17.5 测试指南

- 单元测试：Service层核心逻辑可独立测试（mock Repository）
- Handler测试：使用httptest构造请求
- JWT测试：测试token生成和解析
- 模型测试：确保DTO/Entity字段标签正确
- 运行测试：`go test ./... -v`

---

## 第18章 生产环境风险规避清单

| 风险 | 规避措施 | 实现位置 |
|-----|---------|---------|
| **缓存穿透** | 布隆过滤器+空值缓存+IP限流 | cache/bloom.go, SetNull(), ratelimit.go |
| **缓存击穿** | singleflight+分布式锁+逻辑过期 | cache.go sf.Do/acquireLock/SetLogical |
| **缓存雪崩** | TTL±10%抖动+多级缓存+熔断器+限流 | withJitter(), LocalCache, CircuitBreaker, RateLimit |
| **幽灵消息** | 先提交PG事务，再发MQ消息 | audit_service.go 事务边界设计 |
| **消息重复推送** | WebSocket端LRU幂等去重+前端5分钟去重 | hub.go isDuplicate, websocket.ts shownMessages |
| **消息丢失** | 持久化+手动ACK+死信队列+自动重连+定时对账 | rabbitmq.go 七重保障 |
| **撤回窗口被篡改** | 服务端Redis TTL控制+PG时间戳兜底 | canWithdraw()/setWithdrawTTL() |
| **并发审核冲突** | 乐观锁（version字段） | UpdateStatusWithVersion |
| **Redis宕机级联故障** | 三态熔断器+本地限流/缓存降级 | circuit.go, allowRedis→allowLocal |
| **RabbitMQ宕机** | 降级PG-only模式+定时对账补发 | provideMQ返回nil, retryUndelivered |
| **DDoS/CC攻击** | 三层限流（IP+接口+用户）+白名单 | ratelimit.go |
| **限流绕过** | 路径归一化（数字ID替换:id） | normalizePath() |
| **本地内存溢出** | 本地桶上限(10000)+定期清理 | cleanup() maxLocalBuckets |
| **Goroutine泄漏** | 所有后台goroutine通过stopCh+wg.Wait()停止 | LocalCache.gc, RateLimiter.cleanup, Hub.Run, AuditService.wgGo |
| **单用户连接过多** | 单用户最大5个WebSocket连接 | handleRegister MaxConnsPerUser |
| **大消息攻击** | WebSocket MaxMessageSize=1024 | ReadPump SetReadLimit |
| **密码泄露** | 响应DTO不含Password字段(json:"-")，bcrypt加密 | model.User Password字段 |
| **Panic崩溃** | CustomRecovery中间件defer recover | casbin.go CustomRecovery |
| **优雅关停** | fx.Lifecycle管理OnStop逆序关闭 | module.go startHTTPServer OnStop |
| **主键泄露** | 使用UUID对外暴露，自增ID仅内部使用 | 所有Entity的UUID字段 |
| **跨域问题** | CORS中间件配置AllowCredentials | router.go cors.New |
| **WebSocket断连** | 前端指数退避自动重连+心跳保活 | websocket.ts scheduleReconnect/startHeartbeat |
| **Redis PubSub断连** | 自动重连（最多10次） | hub.go resubscribeRedis |
| **MQ消息体过大** | 只传message_id，详情从PG加载 | NotificationMessage结构体设计 |
| **多实例竞争消费** | Fanout广播模式+每实例独占队列 | rabbitmq.go setupTopology |
| **Redis PubSub循环广播** | MQ消费端使用Local版本不PubSub | SendToAdminsLocal vs BroadcastToAdmins |

---

## 第19章 配置文件详解

### 19.1 config.yaml 完整配置说明

[config.yaml](file:///d:/Programming/Agent_demo/casbin_demo/backend/config.yaml)：

```yaml
server:
  port: 8080              # HTTP服务监听端口
  mode: debug             # Gin模式：debug/release/test

database:
  host: localhost         # PostgreSQL主机
  port: 5432              # PostgreSQL端口
  user: postgres          # 用户名
  password: postgres      # 密码（生产环境建议用环境变量覆盖）
  dbname: casbin_demo     # 数据库名
  sslmode: disable        # SSL模式（生产环境建议require）
  max_idle_conns: 10      # 连接池最大空闲连接
  max_open_conns: 100     # 连接池最大打开连接
  conn_max_lifetime: 3600 # 连接最大生命周期（秒）

redis:
  host: localhost         # Redis主机
  port: 6379              # Redis端口
  password: ""            # Redis密码
  db: 0                   # DB编号
  pool_size: 10           # 连接池大小

jwt:
  secret: "your-secret-key-change-in-production"  # JWT签名密钥（生产必须修改）
  expire_hours: 24        # Token过期时间（小时）

rabbitmq:
  host: localhost         # RabbitMQ主机
  port: 5672              # AMQP端口
  user: guest             # 用户名
  password: guest         # 密码
  vhost: "/"              # 虚拟主机
```

### 19.2 环境变量覆盖

所有配置项都可以通过环境变量覆盖，规则：将`.`替换为`_`，全大写。例如：
- `DATABASE_HOST` 覆盖 database.host
- `DATABASE_PASSWORD` 覆盖 database.password
- `JWT_SECRET` 覆盖 jwt.secret
- `REDIS_PASSWORD` 覆盖 redis.password

Docker/K8s部署时通过环境变量注入敏感配置，避免明文写入配置文件。

---

## 第20章 Docker部署详解

### 20.1 后端Dockerfile

[backend/Dockerfile](file:///d:/Programming/Agent_demo/casbin_demo/backend/Dockerfile) 采用多阶段构建：

```dockerfile
# 构建阶段：Go 1.26镜像
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server cmd/server/main.go

# 运行阶段：alpine最小镜像
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/server .
COPY config-docker.yaml config.yaml
EXPOSE 8080
CMD ["./server"]
```

### 20.2 前端Dockerfile（多阶段构建）

[frontend/Dockerfile](file:///d:/Programming/Agent_demo/casbin_demo/frontend/Dockerfile)：

```dockerfile
# 构建阶段：Node.js 18
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# 运行阶段：Nginx alpine
FROM nginx:alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

### 20.3 docker-compose.yml 完整配置

[docker-compose.yml](file:///d:/Programming/Agent_demo/casbin_demo/docker-compose.yml)：

| 服务 | 镜像 | 端口 | 依赖 | 健康检查 |
|-----|------|------|------|---------|
| postgres | postgres:15-alpine | 5432 | - | pg_isready |
| redis | redis:7-alpine | 6379 | - | redis-cli ping |
| rabbitmq | rabbitmq:3-management-alpine | 5672, 15672 | - | rabbitmq-diagnostics ping |
| backend | 自建镜像 | 8080 | postgres, redis, rabbitmq | curl /health |
| frontend | 自建镜像(nginx) | 80 | backend | curl -f / |

启动顺序：基础设施服务 → backend（等待PG就绪）→ frontend。

---

## 第21章 监控与可观测性

### 21.1 日志规范

所有日志使用Go 1.26标准库`log/slog`，JSON结构化输出，字段包括：
- `ts`：时间戳（RFC3339）
- `level`：日志级别（INFO/WARN/ERROR）
- `msg`：日志消息
- `path`：请求路径
- `method`：HTTP方法
- `status`：响应状态码
- `latency_ms`：耗时毫秒
- `client_ip`：客户端IP
- `error`：错误信息（ERROR级别）

**日志级别说明**：
- INFO：正常请求、启动信息、关键业务事件
- WARN：可恢复错误、限流触发、熔断器状态变化
- ERROR：数据库错误、Panic恢复、未预期异常

### 21.2 关键监控指标

| 指标类别 | 指标名 | 监控方式 | 告警阈值 |
|---------|-------|---------|---------|
| HTTP | 请求QPS | slog日志统计或Prometheus | - |
| HTTP | 错误率(4xx/5xx) | 日志统计 | 5xx > 1%告警 |
| HTTP | P95/P99延迟 | 日志统计 | P99 > 500ms告警 |
| 限流 | IP限流触发次数 | RateLimitStats | 突增告警 |
| 限流 | Redis降级次数 | RedisFallbacks计数 | >0告警（Redis异常） |
| 缓存 | 缓存命中率 | Fetch命中/总请求 | <70%告警 |
| 缓存 | 熔断器状态 | CircuitBreaker open状态 | open告警 |
| MQ | 消息堆积数 | RabbitMQ管理API | >1000告警 |
| MQ | 死信队列消息数 | audit.dlq队列长度 | >0告警 |
| WS | 在线连接数 | Hub.Clients() | - |
| DB | 连接池使用率 | GORM Stats | >80%告警 |

### 21.3 /health 健康检查端点

`GET /health` 返回服务健康状态：

```json
{
  "code": 200,
  "data": {
    "status": "ok",
    "version": "v3.2.0",
    "time": "2026-06-27T00:00:00Z",
    "features": {
      "redis": true,
      "rabbitmq": true,
      "websocket": true
    },
    "stats": {
      "ws_connections": 5,
      "rate_limit_rejected_ip": 12,
      "rate_limit_fallbacks": 0
    }
  }
}
```

features字段显示可选依赖是否可用，降级时对应字段为false。

---

## 第22章 常见问题排查指南

### 22.1 403 Forbidden问题排查

| 可能原因 | 排查步骤 | 解决方案 |
|---------|---------|---------|
| 用户未分配对应角色 | 检查user_roles表是否有记录 | 通过用户管理API分配角色 |
| 角色未分配对应权限 | 检查role_permissions + casbin_rule表 | 通过角色管理分配权限 |
| Casbin策略未重载 | 重启服务或调用ReloadEnforcer | 重启后端服务 |
| 访问管理员接口但非admin | 检查userInfo.roles是否包含admin | 确认账号角色 |
| 路由路径不匹配 | 检查请求路径与策略path是否精确匹配 | 路径需完全一致（含/api前缀） |

### 22.2 WebSocket连接失败排查

1. 检查Nginx配置是否包含WebSocket Upgrade头
2. 确认token有效（JWT未过期）
3. 检查/ws路径是否被限流拦截（白名单已配置，正常不会）
4. 浏览器控制台查看WS连接状态码（101=成功，401=token无效）

### 22.3 消息收不到实时通知排查

1. 检查RabbitMQ是否连接成功（/health features.rabbitmq）
2. 如RabbitMQ未连接，降级为定时对账（30s补发），检查sys_messages.mq_delivered字段
3. 检查WebSocket连接状态（浏览器控制台Network→WS）
4. 检查幂等去重：同一消息ID 5分钟内不重复推送
5. 查看死信队列audit.dlq是否有堆积消息

### 22.4 缓存问题排查

- **缓存不生效**：检查cache.Enabled()，Redis连接失败会降级但本地缓存仍工作
- **数据不一致**：写操作是否调用了Invalidate*方法失效缓存
- **Redis CPU高**：检查是否使用了KEYS命令（代码中使用SCAN替代）

### 22.5 服务启动失败排查

1. 检查PostgreSQL是否可连接（数据库必须启动）
2. Redis/RabbitMQ连接失败不影响启动（降级模式），查看Warn日志
3. 检查config.yaml配置是否正确（主机、端口、密码）
4. 查看slog ERROR日志的具体错误信息

---

**文档版本**：v4.0.0
**最后更新**：2026年6月27日
**适用版本**：RBAC权限管理系统 + 大模型API资源审核平台 v3.2.0
**Go版本要求**：>= 1.26.4
**Node.js版本要求**：>= 18.x
**演示账号**：admin/admin123（管理员）、user/user123（普通用户）
**服务端口**：前端3000（开发）/80（生产），后端8080
**管理界面**：RabbitMQ管理界面 http://localhost:15672 (guest/guest)
