# RBAC权限管理系统技术文档 v1.0.0

---

## 目录

- [第1章 项目概述](#第1章-项目概述)
  - [1.1 项目简介](#11-项目简介)
  - [1.2 核心功能](#12-核心功能)
  - [1.3 技术选型](#13-技术选型)
  - [1.4 版本说明](#14-版本说明)
- [第2章 系统架构](#第2章-系统架构)
  - [2.1 整体架构设计](#21-整体架构设计)
  - [2.2 后端分层架构](#22-后端分层架构)
  - [2.3 前端架构设计](#23-前端架构设计)
  - [2.4 请求处理流程](#24-请求处理流程)
- [第3章 目录结构](#第3章-目录结构)
  - [3.1 项目根目录结构](#31-项目根目录结构)
  - [3.2 后端目录结构](#32-后端目录结构)
  - [3.3 前端目录结构](#33-前端目录结构)
- [第4章 缓存架构设计](#第4章-缓存架构设计)
  - [4.1 两级缓存架构](#41-两级缓存架构)
  - [4.2 缓存分类与TTL策略](#42-缓存分类与ttl策略)
  - [4.3 缓存问题防护方案](#43-缓存问题防护方案)
  - [4.4 缓存预热机制](#44-缓存预热机制)
  - [4.5 缓存失效策略](#45-缓存失效策略)
- [第5章 后端架构详解](#第5章-后端架构详解)
  - [5.1 入口层设计](#51-入口层设计)
  - [5.2 应用装配层（fx依赖注入）](#52-应用装配层fx依赖注入)
  - [5.3 路由层](#53-路由层)
  - [5.4 配置层](#54-配置层)
  - [5.5 模型层](#55-模型层)
  - [5.6 数据访问层（Repository）](#56-数据访问层repository)
  - [5.7 业务逻辑层（Service）](#57-业务逻辑层service)
  - [5.8 中间件层](#58-中间件层)
  - [5.9 处理器层（Handler）](#59-处理器层handler)
  - [5.10 公共包（pkg）](#510-公共包pkg)
- [第6章 Casbin权限模型详解](#第6章-casbin权限模型详解)
  - [6.1 RBAC权限模型](#61-rbac权限模型)
  - [6.2 模型定义文件](#62-模型定义文件)
  - [6.3 权限校验流程](#63-权限校验流程)
  - [6.4 策略管理](#64-策略管理)
- [第7章 前端架构详解](#第7章-前端架构详解)
  - [7.1 技术选型与版本](#71-技术选型与版本)
  - [7.2 应用入口设计](#72-应用入口设计)
  - [7.3 状态管理（Pinia）](#73-状态管理pinia)
  - [7.4 路由系统设计](#74-路由系统设计)
  - [7.5 Axios请求封装](#75-axios请求封装)
  - [7.6 API接口层设计](#76-api接口层设计)
  - [7.7 布局组件设计](#77-布局组件设计)
  - [7.8 错误页面系统](#78-错误页面系统)
  - [7.9 样式与主题设计](#79-样式与主题设计)
- [第8章 错误处理机制](#第8章-错误处理机制)
  - [8.1 后端错误处理](#81-后端错误处理)
  - [8.2 前端错误处理](#82-前端错误处理)
  - [8.3 Nginx错误拦截](#83-nginx错误拦截)
  - [8.4 错误状态码映射表](#84-错误状态码映射表)
- [第9章 审核系统架构设计](#第9章-审核系统架构设计)
  - [9.1 系统概述与业务流程](#91-系统概述与业务流程)
  - [9.2 核心架构组件分工](#92-核心架构组件分工)
  - [9.3 状态机与状态流转](#93-状态机与状态流转)
  - [9.4 2分钟撤回机制设计](#94-2分钟撤回机制设计)
  - [9.5 并发安全保障](#95-并发安全保障)
  - [9.6 生产环境风险规避清单](#96-生产环境风险规避清单)
- [第10章 WebSocket实时通信](#第10章-websocket实时通信)
  - [10.1 WebSocket Hub架构设计](#101-websocket-hub架构设计)
  - [10.2 消息类型与事件定义](#102-消息类型与事件定义)
  - [10.3 连接管理与心跳机制](#103-连接管理与心跳机制)
  - [10.4 前端WebSocket客户端设计](#104-websocket前端客户端设计)
  - [10.5 离线消息与PG兜底方案](#105-离线消息与pg兜底方案)
  - [10.6 多网关跨实例广播](#106-多网关跨实例广播)
- [第11章 RabbitMQ消息队列](#第11章-rabbitmq消息队列)
  - [11.1 消息队列架构设计](#111-消息队列架构设计)
  - [11.2 Exchange与Queue规划](#112-exchange与queue规划)
  - [11.3 可靠投递保障机制](#113-可靠投递保障机制)
  - [11.4 死信队列与异常处理](#114-死信队列与异常处理)
  - [11.5 定时对账与补发机制](#115-定时对账与补发机制)
  - [11.6 高可用降级策略](#116-高可用降级策略)
- [第12章 API接口文档](#第12章-api接口文档)
  - [12.1 通用说明](#121-通用说明)
  - [12.2 认证接口](#122-认证接口)
  - [12.3 用户管理接口](#123-用户管理接口)
  - [12.4 角色管理接口](#124-角色管理接口)
  - [12.5 权限管理接口](#125-权限管理接口)
  - [12.6 仪表盘接口](#126-仪表盘接口)
  - [12.7 审核系统接口](#127-审核系统接口)
  - [12.8 错误码说明](#128-错误码说明)
- [第13章 部署指南](#第13章-部署指南)
  - [13.1 Docker Compose部署](#131-docker-compose部署)
  - [13.2 本地开发环境部署](#132-本地开发环境部署)
  - [13.3 Nginx配置详解](#133-nginx配置详解)
  - [13.4 环境变量配置说明](#134-环境变量配置说明)
  - [13.5 服务健康检查](#135-服务健康检查)
- [第14章 开发指南](#第14章-开发指南)
  - [14.1 新增接口开发流程](#141-新增接口开发流程)
  - [14.2 fx依赖注入规范](#142-fx依赖注入规范)
  - [14.3 缓存使用规范](#143-缓存使用规范)
  - [14.4 代码规范](#144-代码规范)
  - [14.5 测试规范](#145-测试规范)
  - [14.6 常见问题与解决方案](#146-常见问题与解决方案)
- [第15章 核心依赖版本](#第15章-核心依赖版本)
  - [15.1 后端依赖](#151-后端依赖)
  - [15.2 前端依赖](#152-前端依赖)
  - [15.3 基础设施依赖](#153-基础设施依赖)

---

## 第1章 项目概述

### 1.1 项目简介

本项目是一个基于Go (Gin) + Vue 3 + PostgreSQL + Redis构建的生产级RBAC（Role-Based Access Control，基于角色的访问控制）权限管理系统。系统采用前后端分离架构，集成Casbin作为权限控制引擎，提供完整的用户、角色、权限管理功能，同时实现了多级缓存方案、优雅重启、规范错误处理等生产级特性。

系统设计遵循Go语言工程最佳实践，使用uber-go/fx实现依赖注入，保证代码的可测试性和可维护性。后端严格遵循Handler→Service→Repository分层架构，前端采用Vue 3 Composition API + TypeScript实现类型安全的开发体验。

### 1.2 核心功能

| 功能模块 | 功能描述 |
|---------|---------|
| 用户认证 | 用户注册、登录、JWT Token认证、密码加密存储 |
| 用户管理 | 用户CRUD、用户状态管理、用户角色分配 |
| 角色管理 | 角色CRUD、角色权限分配、角色状态管理 |
| 权限管理 | 权限CRUD、权限分类（菜单/按钮/API）、权限树展示 |
| 权限控制 | 基于Casbin的RBAC权限模型、接口级权限校验、细粒度权限控制 |
| 资源审核 | 大模型API资源申请、管理员审核、申请撤回（2分钟窗口）、审核状态流转 |
| 实时通知 | WebSocket实时推送、RabbitMQ消息队列、Redis PubSub跨网关广播、站内消息 |
| 数据缓存 | L1本地缓存+L2 Redis分布式缓存、缓存预热、缓存失效策略 |
| 缓存防护 | 防穿透（布隆过滤器+空值缓存）、防击穿（singleflight+分布式锁）、防雪崩（TTL抖动+熔断器） |
| 仪表盘 | 用户/角色/权限/待审核统计数据、数据可视化 |
| 错误处理 | 统一响应格式、自定义错误页面、Panic恢复、404/405处理 |
| 服务治理 | 优雅关停、IP限流、熔断器、服务降级、结构化日志 |

### 1.3 技术选型

#### 后端技术栈

| 技术 | 版本 | 用途 |
|-----|-----|-----|
| Go | 1.26.4 | 后端开发语言，使用log/slog、sync.OnceValue等新特性 |
| Gin | v1.11.0 | Web框架，处理HTTP请求路由和中间件 |
| GORM | v1.31.0 | ORM框架，数据库操作 |
| Casbin | v2.11.0 | 权限控制引擎，实现RBAC模型 |
| PostgreSQL | - | 关系型数据库，存储业务数据和消息持久化 |
| Redis | - | 分布式缓存、PubSub跨网关广播、撤回TTL控制 |
| RabbitMQ | - | 消息队列，异步解耦、通知推送、削峰填谷 |
| WebSocket (gorilla) | - | 实时通信，前端弹窗通知 |
| uber-go/fx | v1.24.1 | 依赖注入框架，管理组件生命周期 |
| go-redis | v9.7.3 | Redis客户端 |
| golang-jwt | v5.3.0 | JWT Token生成与验证 |
| crypto/bcrypt | - | 密码加密 |
| log/slog | - | Go 1.26标准库结构化日志 |

#### 前端技术栈

| 技术 | 版本 | 用途 |
|-----|-----|-----|
| Vue | v3.5.13 | 前端框架，Composition API |
| Vite | v7.2.4 | 构建工具，开发服务器 |
| TypeScript | v5.9.3 | 类型安全的JavaScript超集 |
| Pinia | v3.0.3 | 状态管理库 |
| Vue Router | v4.6.4 | 前端路由管理 |
| Element Plus | v2.13.7 | UI组件库 |
| Axios | v1.15.0 | HTTP客户端 |
| Sass | v1.99.2 | CSS预处理器 |

#### 基础设施

| 技术 | 版本 | 用途 |
|-----|-----|-----|
| Docker | - | 容器化部署 |
| Docker Compose | - | 多服务编排 |
| Nginx | latest | 前端Web服务器、反向代理、错误拦截 |
| PostgreSQL | 15-alpine | 数据库服务 |
| Redis | 7-alpine | 缓存服务、PubSub广播 |
| RabbitMQ | 3-management-alpine | 消息队列服务 |

### 1.4 版本说明

- **文档版本**：v1.1.0
- **最后更新**：2026年6月25日
- **适用版本**：RBAC权限管理系统 + 资源审核系统 v1.1.0
- **Go版本要求**：>= 1.26.4
- **Node.js版本要求**：>= 18.x

---

## 第2章 系统架构

### 2.1 整体架构设计

系统采用前后端分离架构，通过Nginx作为统一入口，实现静态资源服务和API反向代理。整体架构分为四层：客户端层、接入层、应用层、数据层。

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              客户端层 (Client)                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │  Web浏览器    │  │  Postman     │  │  移动端      │  │  第三方系统   │     │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘     │
└─────────┼─────────────────┼─────────────────┼─────────────────┼─────────────┘
          │                 │                 │                 │
          └─────────────────┴─────────────────┴─────────────────┘
                                    │ HTTP/HTTPS
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              接入层 (Nginx)                                  │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │  静态资源服务  │  API反向代理  │  错误拦截跳转  │  AJAX/浏览器请求区分  │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
│                                    │                                        │
│  ┌─────────────────────────────────┴─────────────────────────────────────┐  │
│  │  /api/* → proxy_pass http://backend:8080                              │  │
│  │  /*     → try_files $uri /index.html                                  │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────┬──────────────────────────────────────┘
                                       │
                    ┌──────────────────┴──────────────────┐
                    │                                     │
                    ▼                                     ▼
┌─────────────────────────────┐           ┌─────────────────────────────────────┐
│       前端应用 (Frontend)    │           │         后端应用 (Backend)           │
│  ┌───────────────────────┐  │           │  ┌───────────────────────────────┐  │
│  │  Vue 3 + TypeScript   │  │           │  │  Gin Web Framework            │  │
│  │  Pinia State Mgmt     │  │           │  │  Middleware (JWT/Casbin/Rate) │  │
│  │  Vue Router           │  │           │  └───────────────┬───────────────┘  │
│  │  Element Plus UI      │  │           │                  │                   │
│  │  Axios HTTP Client    │  │           │  ┌───────────────▼───────────────┐  │
│  └───────────────────────┘  │           │  │  Handler Layer                │  │
│                             │           │  │  - UserHandler                │  │
│                             │           │  │  - RoleHandler                │  │
│                             │           │  │  - PermissionHandler          │  │
│                             │           │  │  - DashboardHandler           │  │
│                             │           │  │  - AuthHandler                │  │
│                             │           │  └───────────────┬───────────────┘  │
│                             │           │                  │                   │
│                             │           │  ┌───────────────▼───────────────┐  │
│                             │           │  │  Service Layer (业务逻辑+缓存)  │  │
│                             │           │  │  - UserService                │  │
│                             │           │  │  - RoleService                │  │
│                             │           │  │  - PermissionService          │  │
│                             │           │  │  - AuthService                │  │
│                             │           │  │  - DashboardService           │  │
│                             │           │  └───────────────┬───────────────┘  │
│                             │           │                  │                   │
│                             │           │  ┌───────────────▼───────────────┐  │
│                             │           │  │  Repository Layer (数据访问)    │  │
│                             │           │  └───────────────┬───────────────┘  │
│                             │           └──────────────────┼──────────────────┘
│                             │                              │
└─────────────────────────────┘                              │
                                                             │
                    ┌────────────────────────────────────────┼──────────────────┐
                    │                                        │                  │
                    ▼                                        ▼                  ▼
┌─────────────────────────────────────┐  ┌─────────────────────────────┐  ┌──────────────┐
│      数据层 - PostgreSQL            │  │     数据层 - Redis          │  │  Casbin策略  │
│  ┌───────────────────────────────┐  │  │  ┌───────────────────────┐  │  │   存储       │
│  │  users  │  roles  │ user_roles│  │  │  │  L1: 本地内存缓存     │  │  │  (GORM)      │
│  │  permissions │ role_perms      │  │  │  │  L2: Redis分布式缓存  │  │  │              │
│  │  casbin_rule (策略表)          │  │  │  │  缓存防护: 穿透/击穿/雪崩│ │  │              │
│  └───────────────────────────────┘  │  │  └───────────────────────┘  │  └──────────────┘
└─────────────────────────────────────┘  └─────────────────────────────┘
```

### 2.2 后端分层架构

后端严格遵循**Handler → Service → Repository → Database/Cache**四层分层架构，遵循依赖倒置原则，各层职责清晰。

```
┌─────────────────────────────────────────────────────────────┐
│                     Handler Layer (处理器层)                  │
│  职责: 参数校验、请求解析、调用Service、响应封装                │
│  特点: 不包含业务逻辑，只做HTTP协议相关处理                    │
│  组件: UserHandler, RoleHandler, PermissionHandler等         │
└──────────────────────────────┬──────────────────────────────┘
                               │  调用业务方法
                               ▼
┌─────────────────────────────────────────────────────────────┐
│                     Service Layer (业务逻辑层)                │
│  职责: 核心业务逻辑实现、缓存策略、事务管理、权限逻辑          │
│  特点: 可独立测试，不直接处理HTTP请求                          │
│  组件: UserService, RoleService, PermissionService等         │
│  缓存: 集成Cache包，实现两级缓存和缓存防护                    │
└──────────────────────────────┬──────────────────────────────┘
                               │  数据访问
                               ▼
┌─────────────────────────────────────────────────────────────┐
│                  Repository Layer (数据访问层)                │
│  职责: 数据库CRUD封装、Casbin策略操作、数据持久化             │
│  特点: 只做数据访问，不包含业务逻辑                            │
│  组件: UserRepo, RoleRepo, PermissionRepo, CasbinRepo        │
└──────────────────────────────┬──────────────────────────────┘
                               │  数据库/缓存操作
              ┌────────────────┴────────────────┐
              ▼                                 ▼
┌─────────────────────────────┐   ┌─────────────────────────────┐
│       Database (PostgreSQL)  │   │        Cache (Redis+本地)   │
│  存储: 用户、角色、权限、策略  │   │  缓存: 热点数据、查询结果    │
└─────────────────────────────┘   └─────────────────────────────┘
```

### 2.3 前端架构设计

前端采用Vue 3 Composition API + TypeScript开发，使用Pinia进行状态管理，Vue Router管理路由，Axios封装HTTP请求，Element Plus提供UI组件。

```
┌─────────────────────────────────────────────────────────────┐
│                         View Layer                           │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────────┐ │
│  │  Login   │ │ Layout   │ │  User    │ │ Error Pages      │ │
│  │  Page    │ │  Container│ │  Mgmt    │ │ 401/403/404/500  │ │
│  └────┬─────┘ └────┬─────┘ └────┬─────┘ └────────┬─────────┘ │
└───────┼──────────────┼────────────┼────────────────┼──────────┘
        │              │            │                │
        ▼              ▼            ▼                ▼
┌─────────────────────────────────────────────────────────────┐
│                       Component Layer                        │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  业务组件 (UserForm, RoleForm, PermissionTree...)      │  │
│  │  布局组件 (Sidebar, Header, Breadcrumb...)             │  │
│  └───────────────────────────────────────────────────────┘  │
└──────────────────────────────┬──────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
│                        State Layer                          │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  Pinia Stores: user.ts, app.ts                         │  │
│  │  管理: 用户信息、Token、菜单状态、全局配置                │  │
│  └───────────────────────────────────────────────────────┘  │
└──────────────────────────────┬──────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
│                        API Layer                            │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────────────┐ │
│  │  request.ts  │ │   auth.ts    │ │ user.ts/role.ts...   │ │
│  │ (Axios封装)  │ │  (认证API)   │ │  (业务API)           │ │
│  └──────────────┘ └──────────────┘ └──────────────────────┘ │
└──────────────────────────────┬──────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
│                       Router Layer                          │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  Vue Router: 路由守卫、动态路由、权限控制、错误路由     │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### 2.4 请求处理流程

一个完整的API请求处理流程包含以下步骤：

```
客户端发起请求
    │
    ▼
Nginx接收请求
    ├─ 静态资源请求 → 返回前端文件
    └─ /api/* 请求 → 反向代理到后端:8080
            │
            ▼
Gin Engine接收请求
    │
    ▼
全局中间件处理
    ├─ Logger() → 请求日志
    ├─ Cors() → 跨域处理
    ├─ CustomRecovery() → Panic恢复
    ├─ RateLimiter() → IP限流
    └─ 错误处理中间件
            │
            ▼
路由匹配
    ├─ NoRoute → 返回404 JSON
    ├─ NoMethod → 返回405 JSON
    └─ 匹配成功继续
            │
            ▼
认证中间件 (JWTAuth)
    ├─ 跳过白名单路径 (登录、注册)
    ├─ 提取Token (Header/Query)
    ├─ 解析验证Token
    ├─ 将用户信息存入Context
    └─ 失败 → 返回401
            │
            ▼
权限中间件 (Casbin)
    ├─ 提取用户ID、请求路径、方法
    ├─ 调用casbin.Enforcer.Enforce()校验
    ├─ 通过 → 继续
    └─ 失败 → 返回403
            │
            ▼
Handler层处理
    ├─ ShouldBindJSON/Query() 绑定参数
    ├─ 参数校验 (binding标签)
    ├─ 调用Service层方法
    └─ 返回统一响应格式
            │
            ▼
Service层处理
    ├─ 参数验证、业务逻辑处理
    ├─ 查询缓存 → 命中则直接返回
    ├─ 缓存未命中 → singleflight合并请求
    ├─ 调用Repository层查询DB
    ├─ 布隆过滤器校验 (防穿透)
    ├─ 查询结果写入缓存
    └─ 熔断器保护 (防雪崩)
            │
            ▼
Repository层处理
    ├─ GORM操作数据库
    ├─ Casbin策略操作
    └─ 返回数据/错误
            │
            ▼
返回响应
    ├─ 成功 → {code:200, message:"success", data:...}
    └─ 失败 → {code:错误码, message:"错误信息"} HTTP状态码匹配code
```

---

## 第3章 目录结构

### 3.1 项目根目录结构

```
casbin_demo/
├── backend/                    # 后端Go项目目录
│   ├── cmd/
│   │   └── server/
│   │       └── main.go         # 应用入口，仅一个main函数
│   ├── internal/               # 内部应用代码（不可外部导入）
│   │   ├── app/                # 应用装配层 (fx)
│   │   ├── config/             # 配置层
│   │   ├── handler/            # 处理器层
│   │   ├── middleware/          # 中间件层
│   │   ├── model/              # 模型层 (DB实体+DTO)
│   │   ├── repository/         # 数据访问层
│   │   ├── router/             # 路由层
│   │   └── service/            # 业务逻辑层
│   ├── pkg/                    # 公共包（可复用）
│   │   ├── cache/              # 两级缓存实现
│   │   ├── casbin/             # Casbin初始化
│   │   ├── database/           # 数据库初始化
│   │   ├── jwt/                # JWT工具
│   │   ├── limiter/            # 限流器
│   │   ├── logger/             # 日志工具
│   │   ├── redis/              # Redis客户端
│   │   └── response/           # 统一响应封装
│   ├── config.yaml             # 后端配置文件
│   ├── go.mod                  # Go模块定义
│   └── go.sum                  # Go依赖校验
├── frontend/                   # 前端Vue3项目目录
│   ├── src/
│   │   ├── api/                # API接口层
│   │   ├── assets/             # 静态资源
│   │   ├── components/         # 公共组件
│   │   ├── layouts/            # 布局组件
│   │   ├── router/             # 路由配置
│   │   ├── stores/             # Pinia状态管理
│   │   ├── styles/             # 全局样式
│   │   ├── utils/              # 工具函数
│   │   ├── views/              # 页面视图
│   │   ├── App.vue             # 根组件
│   │   └── main.ts             # 前端入口
│   ├── Dockerfile              # 前端Docker构建文件
│   ├── nginx.conf              # Nginx配置
│   ├── package.json            # 前端依赖定义
│   ├── tsconfig.json           # TypeScript配置
│   └── vite.config.ts          # Vite配置
├── docker-compose.yml          # Docker Compose编排
├── Makefile                    # 常用命令集合
└── TECHNICAL_DOCUMENTATION.md  # 本文档
```

### 3.2 后端目录结构

```
backend/
├── cmd/
│   └── server/
│       └── main.go                 # 应用入口：fx.New(app.Module()).Run()
├── internal/
│   ├── app/
│   │   └── module.go              # fx.Module()定义，所有组件注册中心
│   ├── config/
│   │   └── config.go              # 配置结构体定义、Viper加载逻辑
│   ├── handler/
│   │   ├── user_handler.go        # 用户管理处理器
│   │   ├── role_handler.go        # 角色管理处理器
│   │   ├── permission_handler.go  # 权限管理处理器
│   │   ├── auth_handler.go        # 认证处理器（登录注册）
│   │   └── dashboard_handler.go   # 仪表盘处理器
│   ├── middleware/
│   │   ├── cors.go                # 跨域中间件
│   │   ├── jwt.go                 # JWT认证中间件
│   │   ├── casbin.go              # Casbin权限中间件
│   │   ├── ratelimit.go           # IP限流中间件
│   │   └── recovery.go            # Panic恢复中间件
│   ├── model/
│   │   ├── user.go                # User实体、DTO、表单
│   │   ├── role.go                # Role实体、DTO、表单
│   │   └── permission.go          # Permission实体、DTO、表单
│   ├── repository/
│   │   ├── user_repo.go           # 用户数据访问
│   │   ├── role_repo.go           # 角色数据访问
│   │   └── permission_repo.go     # 权限数据访问
│   ├── router/
│   │   └── router.go              # 路由注册、EngineWrapper定义
│   └── service/
│       ├── user_service.go        # 用户业务逻辑
│       ├── role_service.go        # 角色业务逻辑
│       ├── permission_service.go  # 权限业务逻辑
│       ├── auth_service.go        # 认证业务逻辑
│       └── dashboard_service.go   # 仪表盘业务逻辑
└── pkg/
    ├── cache/
    │   └── cache.go               # 两级缓存实现、缓存防护逻辑
    ├── casbin/
    │   └── casbin.go              # Casbin Enforcer初始化、GORM适配器
    ├── database/
    │   └── database.go            # PostgreSQL连接池初始化、GORM配置
    ├── jwt/
    │   └── jwt.go                 # JWT Token生成、解析、验证
    ├── limiter/
    │   └── limiter.go             # 令牌桶限流器、熔断器
    ├── logger/
    │   └── logger.go              # slog初始化、日志格式配置
    ├── redis/
    │   └── redis.go               # Redis客户端初始化
    └── response/
        └── response.go            # Success/Fail响应封装、状态码匹配
```

### 3.3 前端目录结构

```
frontend/
└── src/
    ├── api/
    │   ├── request.ts             # Axios实例封装、拦截器
    │   ├── auth.ts                # 认证相关API
    │   ├── user.ts                # 用户管理API
    │   ├── role.ts                # 角色管理API
    │   ├── permission.ts          # 权限管理API
    │   └── dashboard.ts           # 仪表盘API
    ├── assets/
    │   └── logo.png               # 静态图片资源
    ├── components/
    │   └── HelloWorld.vue         # 示例组件（可删除）
    ├── layouts/
    │   └── MainLayout.vue         # 主布局组件
    ├── router/
    │   ├── index.ts               # 路由配置、守卫定义
    │   └── routes.ts              # 路由表定义
    ├── stores/
    │   ├── user.ts                # 用户状态Store
    │   └── app.ts                 # 应用全局状态Store
    ├── styles/
    │   └── index.scss             # 全局样式、Element Plus主题
    ├── utils/
    │   └── storage.ts             # localStorage/sessionStorage封装
    ├── views/
    │   ├── Login.vue              # 登录/注册页面
    │   ├── Layout.vue             # 布局容器页面
    │   ├── Dashboard.vue          # 仪表盘页面
    │   ├── UserManagement.vue     # 用户管理页面
    │   ├── RoleManagement.vue     # 角色管理页面
    │   ├── PermissionManagement.vue  # 权限管理页面
    │   └── error/                 # 错误页面目录
    │       ├── ErrorPage.vue      # 错误页面通用组件
    │       ├── 401.vue            # 401未授权页面
    │       ├── 403.vue            # 403禁止访问页面
    │       ├── 404.vue            # 404未找到页面
    │       └── 500.vue            # 500服务器错误页面
    ├── App.vue                    # 根组件
    └── main.ts                    # 应用入口：创建Vue实例、注册插件
```

---

## 第4章 缓存架构设计

### 4.1 两级缓存架构

系统采用**L1本地内存缓存 + L2 Redis分布式缓存**的两级缓存架构，在保证性能的同时确保数据一致性。

```
┌─────────────────────────────────────────────────────────────────┐
│                     两级缓存数据读取流程                          │
│                                                                 │
│  业务请求                                                        │
│      │                                                          │
│      ▼                                                          │
│  ┌─────────────┐    命中     ┌──────────────────┐                │
│  │ L1本地缓存   │──────────→ │  返回数据（最快）  │                │
│  │ (进程内)     │            └──────────────────┘                │
│  └──────┬──────┘                                                 │
│         │ 未命中                                                 │
│         ▼                                                       │
│  ┌─────────────┐    命中     ┌──────────────────┐                │
│  │  singleflight│──────────→ │  合并请求，防击穿  │                │
│  │  (请求合并)  │           └──────────────────┘                │
│  └──────┬──────┘                                                 │
│         │ 执行查询                                               │
│         ▼                                                       │
│  ┌─────────────┐    命中     ┌──────────────────┐                │
│  │ L2 Redis缓存 │──────────→ │ 回写L1，返回数据  │                │
│  │ (分布式)     │           └──────────────────┘                │
│  └──────┬──────┘                                                 │
│         │ 未命中                                                 │
│         ▼                                                       │
│  ┌─────────────┐    不存在   ┌──────────────────┐                │
│  │ 布隆过滤器    │──────────→ │  返回空值（防穿透）│                │
│  └──────┬──────┘            └──────────────────┘                │
│         │ 存在                                                   │
│         ▼                                                       │
│  ┌─────────────┐    熔断     ┌──────────────────┐                │
│  │ 熔断器检查   │──────────→ │  服务降级返回      │                │
│  └──────┬──────┘            └──────────────────┘                │
│         │ 正常                                                   │
│         ▼                                                       │
│  ┌─────────────┐    分布式锁  ┌──────────────────┐               │
│  │ 查询数据库   │◄──────────│  双重检查防击穿    │               │
│  └──────┬──────┘            └──────────────────┘                │
│         │ 查询结果                                                │
│         ├──────────────────────────────────────────┐             │
│         ▼                                          ▼             │
│  ┌─────────────┐                           ┌─────────────┐       │
│  │ 空值缓存     │                           │ 写入Redis   │       │
│  │ (60s TTL)   │                           │ (TTL±10%抖动)│       │
│  └─────────────┘                           └──────┬──────┘       │
│                                                  │               │
│                                                  ▼               │
│                                         ┌─────────────┐          │
│                                         │ 回写L1本地   │          │
│                                         └──────┬──────┘          │
│                                                │                 │
│                                                ▼                 │
│                                       ┌──────────────────┐       │
│                                       │    返回数据       │       │
│                                       └──────────────────┘       │
└─────────────────────────────────────────────────────────────────┘
```

### 4.2 缓存分类与TTL策略

缓存按照数据特性分为五类，每类采用不同的TTL策略：

| 缓存分类 | TTL基础值 | TTL抖动 | 适用场景 | Key前缀示例 |
|---------|----------|--------|---------|------------|
| **热点数据缓存** | 30分钟 | ±10% | 角色列表、权限列表、用户信息等高频访问、变更少的数据 | `cache:hot:roles` |
| **查询缓存** | 5分钟 | ±10% | 用户列表、分页查询等动态查询结果 | `cache:query:users:page:1` |
| **配置缓存** | 1小时 | ±10% | 系统配置、常量数据等 | `cache:config:*` |
| **统计缓存** | 2分钟 | ±10% | 仪表盘统计数据、计数数据 | `cache:stats:*` |
| **空值缓存** | 60秒 | 无抖动 | 数据库不存在的数据（防穿透） | `cache:null:*` |

**TTL随机抖动实现**：为防止缓存雪崩，所有非空值缓存的TTL会在基础值上随机±10%，避免大量缓存同时过期。

```go
// TTL抖动计算示例（cache包实现）
func (c *Cache) getTTLWithJitter(category string) time.Duration {
    baseTTL := c.getBaseTTL(category)
    // 空值缓存不抖动
    if category == "null" {
        return baseTTL
    }
    // ±10%随机抖动
    jitter := time.Duration(rand.Float64()*0.2-0.1) * float64(baseTTL)
    return baseTTL + jitter
}
```

### 4.3 缓存问题防护方案

系统针对缓存三大问题（穿透、击穿、雪崩）实现了多层防护：

#### 4.3.1 缓存穿透防护

**问题描述**：查询不存在的数据，请求直接穿透到数据库，导致数据库压力过大。

**防护方案**：
1. **布隆过滤器**：启动时预热ID集合到布隆过滤器，查询前先检查，不存在直接返回
2. **空值缓存**：数据库查询为空时，缓存空值标记（NullValue = "NULL_CACHE"），TTL 60秒，无抖动
3. **IP限流**：对高频请求IP进行限流（20req/s），防止恶意攻击

```go
// 空值缓存逻辑
var data User
err := c.Get(ctx, key, &data)
if err == nil {
    if data.ID == 0 && reflect.ValueOf(data).IsZero() {
        return nil, errors.New("data not found (null cache)")
    }
    return &data, nil
}
```

#### 4.3.2 缓存击穿防护

**问题描述**：热点key过期瞬间，大量并发请求同时打到数据库，导致数据库压力骤增。

**防护方案**：
1. **singleflight.Group**：Go标准库扩展，合并相同key的并发请求，只有一个请求实际查询DB
2. **分布式锁**（Redis SETNX）：获取锁后双重检查缓存，防止多实例下的击穿问题
3. **逻辑过期**：热点数据不设置物理过期，设置逻辑过期时间，后台异步更新

```go
// singleflight使用示例
v, err, _ := c.sf.Do(key, func() (interface{}, error) {
    // 双重检查：获取锁前再查一次缓存
    var data T
    if err := c.Get(ctx, key, &data); err == nil {
        return &data, nil
    }
    // 分布式锁保护
    lockKey := "lock:" + key
    if c.redisClient.SetNX(ctx, lockKey, 1, 10*time.Second).Val() {
        defer c.redisClient.Del(ctx, lockKey)
        // 双重检查：获取锁后再查一次
        // ... 查询数据库
    }
    // 查询数据库并写缓存
})
```

#### 4.3.3 缓存雪崩防护

**问题描述**：大量缓存同时过期，或Redis服务宕机，导致所有请求打到数据库，引发数据库雪崩。

**防护方案**：
1. **TTL随机抖动±10%**：避免大量key同时过期
2. **L1本地缓存兜底**：Redis不可用时使用本地缓存数据
3. **熔断器（Circuit Breaker）**：数据库连续失败5次触发熔断，30秒内直接返回降级响应
4. **服务降级**：熔断期间返回默认数据或友好提示，不打数据库
5. **Redis集群部署**：主从+哨兵提高可用性

```go
// 熔断器实现
type CircuitBreaker struct {
    failures    int64
    lastFailure time.Time
    state       int32 // 0:关闭, 1:打开(熔断)
}

func (cb *CircuitBreaker) Allow() error {
    if atomic.LoadInt32(&cb.state) == 1 {
        if time.Since(cb.lastFailure) < 30*time.Second {
            return errors.New("circuit breaker open")
        }
        atomic.StoreInt32(&cb.state, 0)
    }
    return nil
}

func (cb *CircuitBreaker) RecordFailure() {
    atomic.AddInt64(&cb.failures, 1)
    cb.lastFailure = time.Now()
    if atomic.LoadInt64(&cb.failures) >= 5 {
        atomic.StoreInt32(&cb.state, 1)
    }
}
```

### 4.4 缓存预热机制

服务启动时自动预热热点数据到Redis和本地缓存，避免启动初期大量请求穿透到数据库：

```go
// 缓存预热流程（在app/module.go的Invoke中执行）
func PreloadCache(
    roleService *service.RoleService,
    permService *service.PermissionService,
) {
    slog.Info("开始缓存预热...")

    // 1. 预热角色列表（热点数据，30min TTL）
    roles, _ := roleService.GetAllRoles()
    slog.Info("角色列表预热完成", "count", len(roles))

    // 2. 预热权限列表（热点数据，30min TTL）
    perms, _ := permService.GetAllPermissions()
    slog.Info("权限列表预热完成", "count", len(perms))

    // 3. 预热布隆过滤器：加载所有用户ID、角色ID、权限ID
    // ...

    slog.Info("缓存预热完成")
}
```

### 4.5 缓存失效策略

写操作（增/删/改）必须主动失效相关缓存，使用**SCAN + DEL**非阻塞方式批量删除缓存key，避免使用KEYS命令阻塞Redis：

```go
// 缓存失效示例
func (s *UserService) invalidateUserCache(ctx context.Context, userID uint) {
    // 1. 删除具体用户缓存
    c.redisClient.Del(ctx, "cache:hot:user:"+string(rune(userID)))

    // 2. SCAN删除相关查询缓存（非阻塞）
    var cursor uint64
    for {
        keys, nextCursor, err := c.redisClient.Scan(ctx, cursor, "cache:query:users:*", 100).Result()
        if err != nil {
            break
        }
        if len(keys) > 0 {
            c.redisClient.Del(ctx, keys...)
        }
        cursor = nextCursor
        if cursor == 0 {
            break
        }
    }

    // 3. 清除L1本地缓存
    c.localCache.Delete("cache:hot:user:" + string(rune(userID)))
}
```

**缓存失效策略表**：

| 操作 | 失效缓存范围 | 执行方式 |
|-----|------------|---------|
| 新增用户 | 用户列表查询缓存 | SCAN+DEL |
| 修改用户 | 该用户缓存 + 用户列表缓存 | DEL + SCAN+DEL |
| 删除用户 | 该用户缓存 + 用户列表缓存 | DEL + SCAN+DEL |
| 角色变更 | 角色缓存 + 用户角色相关缓存 | DEL + SCAN+DEL |
| 权限变更 | 权限缓存 + 角色权限缓存 + 用户权限缓存 | DEL + SCAN+DEL |

---

## 第5章 后端架构详解

### 5.1 入口层设计

`main.go`设计极致简洁，仅包含一个`main`函数，所有组件创建、依赖注入、生命周期管理委托给fx框架。

**文件位置**：[main.go](file:///d:/Programming/Agent_demo/casbin_demo/backend/cmd/server/main.go)

```go
package main

import (
    "go.uber.org/fx"
    "casbin-demo/internal/app"
)

// main 应用入口
// 设计原则：main函数仅做应用启动，所有初始化逻辑交给fx框架
func main() {
    // fx.New()创建应用容器，app.Module()注册所有组件
    // Run()启动应用，阻塞等待信号，实现优雅关停
    fx.New(app.Module()).Run()
}
```

**设计优势**：
- `main.go`零业务逻辑，极简设计
- 依赖注入由fx管理，组件解耦
- fx自动处理生命周期（OnStart/OnStop）
- 优雅关停由signal.NotifyContext支持（Go 1.26特性）

### 5.2 应用装配层（fx依赖注入）

应用装配层是fx的Module定义，是所有组件的注册中心，统一管理组件的构造函数和生命周期。

**文件位置**：[module.go](file:///d:/Programming/Agent_demo/casbin_demo/backend/internal/app/module.go)

#### 5.2.1 Module()核心结构

```go
package app

import (
    "go.uber.org/fx"
    // 导入所有需要的包...
)

// Module 应用模块，注册所有组件到fx容器
// fx.Module返回fx.Option，包含Provide和Invoke两部分
func Module() fx.Option {
    return fx.Module("casbin-demo",
        // Provide: 注册构造函数，fx自动解析依赖关系
        fx.Provide(
            // 1. 基础组件
            config.NewConfig,         // 配置加载
            logger.NewLogger,         // 日志初始化
            database.NewDatabase,     // 数据库初始化
            redis.NewRedisClient,     // Redis客户端初始化
            cache.NewCache,           // 缓存客户端初始化

            // 2. Casbin权限引擎
            casbin.NewEnforcer,       // Casbin Enforcer

            // 3. Repository层
            repository.NewUserRepo,
            repository.NewRoleRepo,
            repository.NewPermissionRepo,

            // 4. Service层
            service.NewAuthService,
            service.NewUserService,
            service.NewRoleService,
            service.NewPermissionService,
            service.NewDashboardService,

            // 5. Handler层
            handler.NewAuthHandler,
            handler.NewUserHandler,
            handler.NewRoleHandler,
            handler.NewPermissionHandler,
            handler.NewDashboardHandler,

            // 6. 路由层
            router.NewHandlers,       // Handlers聚合
            router.NewEngineWrapper,  // EngineWrapper

            // 7. 限流器
            limiter.NewRateLimiter,
        ),

        // Invoke: 注册启动后执行的函数，用于副作用操作
        fx.Invoke(
            router.RegisterRoutes,    // 注册路由
            cache.PreloadCache,       // 缓存预热
        ),
    )
}
```

#### 5.2.2 依赖注入链路图

```
fx容器解析依赖顺序（从底层到上层）：

config.NewConfig()
       ↓
logger.NewLogger(config)
       ↓
database.NewDatabase(config, logger)
redis.NewRedisClient(config, logger)
       ↓
cache.NewCache(redisClient, logger)
casbin.NewEnforcer(db, logger)
       ↓
repository.NewUserRepo(db, logger)
repository.NewRoleRepo(db, casbin, logger)
repository.NewPermissionRepo(db, casbin, logger)
       ↓
service.NewUserService(userRepo, cache, logger)
service.NewRoleService(roleRepo, cache, logger)
service.NewPermissionService(permRepo, cache, logger)
...
       ↓
handler.NewUserHandler(userService, logger)
handler.NewRoleHandler(roleService, logger)
...
       ↓
router.NewHandlers(authHandler, userHandler, roleHandler, permHandler, dashboardHandler)
router.NewEngineWrapper(config, handlers, rateLimiter, logger)
       ↓
fx.Invoke:
  router.RegisterRoutes(engineWrapper)
  cache.PreloadCache(roleService, permService)
```

#### 5.2.3 sync.OnceValue单例模式

所有基础组件（DB、Redis、Casbin等）使用Go 1.26新增的`sync.OnceValue`实现并发安全的单例模式：

**示例 - database.go**：
```go
var getDB = sync.OnceValue(func() *gorm.DB {
    // 初始化逻辑只执行一次
    dsn := ...
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{...})
    if err != nil {
        slog.Error("数据库连接失败", "error", err)
        panic(err)
    }
    // 自动迁移
    db.AutoMigrate(&model.User{}, &model.Role{}, &model.Permission{})
    return db
})

// NewDatabase 构造函数，供fx调用
func NewDatabase() *gorm.DB {
    return getDB() // 多次调用返回同一个实例
}
```

**sync.OnceValue优势**：
- 比`sync.Once`更简洁，直接返回值
- 并发安全，初始化逻辑只执行一次
- 初始化失败会panic，快速失败
- Go 1.26标准库，无需额外依赖

### 5.3 路由层

路由层负责Gin引擎创建、全局中间件注册、路由分组配置，定义`EngineWrapper`包装Gin引擎和需要优雅关停的资源（如限流器）。

**文件位置**：[router.go](file:///d:/Programming/Agent_demo/casbin_demo/backend/internal/router/router.go)

#### 5.3.1 核心结构体

```go
// EngineWrapper 包装Gin引擎和需要优雅关闭的资源
// 实现了fx的Start/Stop接口，支持优雅关停
type EngineWrapper struct {
    Engine      *gin.Engine
    RateLimiter *limiter.RateLimiter
}

// Handlers 聚合所有Handler实例，方便路由注册时注入
type Handlers struct {
    Auth       *handler.AuthHandler
    User       *handler.UserHandler
    Role       *handler.RoleHandler
    Permission *handler.PermissionHandler
    Dashboard  *handler.DashboardHandler
}
```

#### 5.3.2 NewEngine - 创建Gin引擎

```go
func NewEngine(
    cfg *config.Config,
    handlers *Handlers,
    rl *limiter.RateLimiter,
    log *slog.Logger,
) *EngineWrapper {
    // 设置Gin模式
    if cfg.Server.Mode == "release" {
        gin.SetMode(gin.ReleaseMode)
    }

    r := gin.New()

    // 注册全局中间件（按顺序执行）
    r.Use(gin.Logger())                                    // 1. 请求日志
    r.Use(middleware.Cors())                               // 2. 跨域处理
    r.Use(middleware.CustomRecovery(log))                  // 3. Panic恢复
    r.Use(middleware.RateLimitMiddleware(rl))              // 4. IP限流

    // 注册NoRoute/NoMethod处理器（返回JSON而非HTML）
    r.NoRoute(func(c *gin.Context) {
        response.NotFound(c, "资源不存在")
    })
    r.NoMethod(func(c *gin.Context) {
        response.MethodNotAllowed(c, "请求方法不允许")
    })

    return &EngineWrapper{
        Engine:      r,
        RateLimiter: rl,
    }
}
```

#### 5.3.3 RegisterRoutes - 路由注册

```go
func RegisterRoutes(ew *EngineWrapper, handlers *Handlers) {
    r := ew.Engine

    // 健康检查接口（无需认证）
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    // API v1 路由组
    v1 := r.Group("/api/v1")
    {
        // 公开接口（无需认证）
        auth := v1.Group("/auth")
        {
            auth.POST("/register", handlers.Auth.Register)
            auth.POST("/login", handlers.Auth.Login)
        }

        // 需要认证的接口组
        authenticated := v1.Group("")
        authenticated.Use(middleware.JWTAuth())
        {
            // 认证相关（获取用户信息）
            authenticated.GET("/userinfo", handlers.Auth.GetUserInfo)

            // 用户管理（需要权限）
            users := authenticated.Group("/users")
            users.Use(middleware.CasbinMiddleware())
            {
                users.GET("", handlers.User.List)
                users.POST("", handlers.User.Create)
                users.GET("/:id", handlers.User.Get)
                users.PUT("/:id", handlers.User.Update)
                users.DELETE("/:id", handlers.User.Delete)
            }

            // 角色管理
            roles := authenticated.Group("/roles")
            roles.Use(middleware.CasbinMiddleware())
            {
                roles.GET("", handlers.Role.List)
                roles.POST("", handlers.Role.Create)
                roles.GET("/:id", handlers.Role.Get)
                roles.PUT("/:id", handlers.Role.Update)
                roles.DELETE("/:id", handlers.Role.Delete)
                roles.POST("/:id/permissions", handlers.Role.AssignPermissions)
            }

            // 权限管理
            perms := authenticated.Group("/permissions")
            perms.Use(middleware.CasbinMiddleware())
            {
                perms.GET("", handlers.Permission.List)
                perms.POST("", handlers.Permission.Create)
                perms.GET("/:id", handlers.Permission.Get)
                perms.PUT("/:id", handlers.Permission.Update)
                perms.DELETE("/:id", handlers.Permission.Delete)
            }

            // 仪表盘
            dashboard := authenticated.Group("/dashboard")
            dashboard.Use(middleware.CasbinMiddleware())
            {
                dashboard.GET("/stats", handlers.Dashboard.GetStats)
            }
        }
    }

    // fx生命周期：启动HTTP服务器
    ew.startServer()
}
```

#### 5.3.4 优雅关停实现

```go
func (ew *EngineWrapper) startServer() {
    srv := &http.Server{
        Addr:    ":8080",
        Handler: ew.Engine,
    }

    // fx OnStart: 异步启动服务器
    fxStart := func(ctx context.Context) error {
        go func() {
            slog.Info("HTTP服务器启动", "addr", ":8080")
            if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
                slog.Error("服务器启动失败", "error", err)
            }
        }()
        return nil
    }

    // fx OnStop: 优雅关停
    fxStop := func(ctx context.Context) error {
        slog.Info("正在关闭HTTP服务器...")
        shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
        defer cancel()

        // 停止限流器后台goroutine
        ew.RateLimiter.Stop()

        // 关停HTTP服务器，等待活跃请求完成
        return srv.Shutdown(shutdownCtx)
    }
}
```

### 5.4 配置层

配置层使用Viper加载配置，支持**YAML文件、环境变量、默认值**三级配置，环境变量优先级最高。

**文件位置**：[config.go](file:///d:/Programming/Agent_demo/casbin_demo/backend/internal/config/config.go)

```go
// Config 全局配置结构体
type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
    Redis    RedisConfig    `mapstructure:"redis"`
    JWT      JWTConfig      `mapstructure:"jwt"`
}

type ServerConfig struct {
    Mode string `mapstructure:"mode"` // debug/release/test
    Port int    `mapstructure:"port"`
}

type DatabaseConfig struct {
    Host     string `mapstructure:"host"`
    Port     int    `mapstructure:"port"`
    User     string `mapstructure:"user"`
    Password string `mapstructure:"password"`
    DBName   string `mapstructure:"dbname"`
    SSLMode  string `mapstructure:"sslmode"`
}

// NewConfig 加载配置
func NewConfig() *Config {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath(".")
    viper.AddConfigPath("./backend")

    // 环境变量前缀：CASBIN_DEMO_，例如CASBIN_DEMO_DATABASE_HOST
    viper.SetEnvPrefix("CASBIN_DEMO")
    viper.AutomaticEnv()
    viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

    // 设置默认值
    viper.SetDefault("server.mode", "debug")
    viper.SetDefault("server.port", 8080)
    // ...更多默认值

    // 读取配置文件（忽略错误，可通过环境变量配置）
    if err := viper.ReadInConfig(); err != nil {
        slog.Warn("配置文件读取失败，使用环境变量和默认值", "error", err)
    }

    var cfg Config
    if err := viper.Unmarshal(&cfg); err != nil {
        slog.Error("配置解析失败", "error", err)
        panic(err)
    }
    return &cfg
}
```

### 5.5 模型层

模型层包含GORM数据库实体（Entity）和数据传输对象（DTO/Form），使用结构体标签定义数据库映射和验证规则。

#### 5.5.1 User模型

**文件位置**：[user.go](file:///d:/Programming/Agent_demo/casbin_demo/backend/internal/model/user.go)

```go
// User 用户实体（数据库表: users）
type User struct {
    ID        uint           `gorm:"primarykey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
    Username  string         `gorm:"uniqueIndex;size:50;not null" json:"username"`
    Password  string         `gorm:"size:100;not null" json:"-"` // json:"-" 不返回密码
    Nickname  string         `gorm:"size:50" json:"nickname"`
    Email     string         `gorm:"size:100" json:"email"`
    Status    int            `gorm:"default:1" json:"status"` // 1:启用 0:禁用
    Roles     []Role         `gorm:"many2many:user_roles;" json:"roles,omitempty"`
}

// RegisterForm 注册表单
type RegisterForm struct {
    Username string `json:"username" binding:"required,min=3,max=50"`
    Password string `json:"password" binding:"required,min=6,max=50"`
    Nickname string `json:"nickname" binding:"max=50"`
    Email    string `json:"email" binding:"omitempty,email"`
}

// LoginForm 登录表单
type LoginForm struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
}

// UserDTO 用户返回DTO（脱敏，不含密码）
type UserDTO struct {
    ID        uint      `json:"id"`
    Username  string    `json:"username"`
    Nickname  string    `json:"nickname"`
    Email     string    `json:"email"`
    Status    int       `json:"status"`
    Roles     []RoleDTO `json:"roles,omitempty"`
    CreatedAt time.Time `json:"created_at"`
}
```

### 5.6 数据访问层（Repository）

Repository层封装所有数据库操作，向上提供统一的数据访问接口，不包含业务逻辑。

**UserRepo示例**：
```go
type UserRepo struct {
    db     *gorm.DB
    logger *slog.Logger
}

func NewUserRepo(db *gorm.DB, logger *slog.Logger) *UserRepo {
    return &UserRepo{db: db, logger: logger}
}

// Create 创建用户
func (r *UserRepo) Create(user *model.User) error {
    return r.db.Create(user).Error
}

// GetByID 根据ID查询用户（预加载Roles）
func (r *UserRepo) GetByID(id uint) (*model.User, error) {
    var user model.User
    err := r.db.Preload("Roles").First(&user, id).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

// GetByUsername 根据用户名查询（用于登录）
func (r *UserRepo) GetByUsername(username string) (*model.User, error) {
    var user model.User
    err := r.db.Where("username = ?", username).First(&user).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

// List 分页查询用户列表
func (r *UserRepo) List(page, pageSize int) ([]model.User, int64, error) {
    var users []model.User
    var total int64

    query := r.db.Model(&model.User{})
    query.Count(&total)
    err := r.db.Preload("Roles").
        Offset((page - 1) * pageSize).
        Limit(pageSize).
        Order("id DESC").
        Find(&users).Error
    return users, total, err
}
```

### 5.7 业务逻辑层（Service）

Service层实现核心业务逻辑，包括缓存策略、事务管理、DTO转换、权限逻辑等。

**UserService示例**：
```go
type UserService struct {
    repo  *repository.UserRepo
    cache *cache.Cache
    logger *slog.Logger
}

func NewUserService(repo *repository.UserRepo, cache *cache.Cache, logger *slog.Logger) *UserService {
    return &UserService{repo: repo, cache: cache, logger: logger}
}

// GetByID 带缓存的用户查询
func (s *UserService) GetByID(ctx context.Context, id uint) (*model.UserDTO, error) {
    cacheKey := fmt.Sprintf("cache:hot:user:%d", id)

    // 1. 查缓存（两级缓存+防护逻辑在cache包内部实现）
    var dto model.UserDTO
    err := s.cache.Get(ctx, cacheKey, &dto)
    if err == nil {
        return &dto, nil
    }

    // 2. 缓存未命中，查询数据库
    user, err := s.repo.GetByID(id)
    if err != nil {
        // 查询为空，设置空值缓存（防穿透）
        s.cache.SetNull(ctx, cacheKey)
        return nil, errors.New("用户不存在")
    }

    // 3. 转换为DTO
    dto = model.UserDTO{
        ID:       user.ID,
        Username: user.Username,
        Nickname: user.Nickname,
        Email:    user.Email,
        Status:   user.Status,
    }

    // 4. 转换角色为DTO
    for _, role := range user.Roles {
        dto.Roles = append(dto.Roles, model.RoleDTO{
            ID:   role.ID,
            Name: role.Name,
        })
    }

    // 5. 写入缓存（热点数据，30min±10% TTL）
    s.cache.Set(ctx, cacheKey, dto, "hot")

    return &dto, nil
}

// Create 创建用户（密码加密+缓存失效）
func (s *UserService) Create(ctx context.Context, form *model.CreateUserForm) (*model.UserDTO, error) {
    // 1. 检查用户名是否已存在
    exist, _ := s.repo.GetByUsername(form.Username)
    if exist != nil {
        return nil, errors.New("用户名已存在")
    }

    // 2. 密码bcrypt加密
    hashedPwd, err := bcrypt.GenerateFromPassword([]byte(form.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, err
    }

    // 3. 创建用户实体
    user := &model.User{
        Username: form.Username,
        Password: string(hashedPwd),
        Nickname: form.Nickname,
        Email:    form.Email,
        Status:   1,
    }

    // 4. 保存到数据库
    if err := s.repo.Create(user); err != nil {
        return nil, err
    }

    // 5. 失效用户列表缓存
    s.invalidateUserListCache(ctx)

    return s.toDTO(user), nil
}
```

### 5.8 中间件层

中间件层包含JWT认证、Casbin权限校验、CORS跨域、限流、Panic恢复等中间件。

#### 5.8.1 JWT认证中间件

```go
func JWTAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. 从Header获取Token
        token := c.GetHeader("Authorization")
        if token == "" {
            token = c.Query("token")
        }

        if token == "" {
            response.Unauthorized(c, "未提供认证令牌")
            c.Abort()
            return
        }

        // 2. 解析Bearer Token
        token = strings.TrimPrefix(token, "Bearer ")

        // 3. 解析验证Token
        claims, err := jwt.ParseToken(token)
        if err != nil {
            response.Unauthorized(c, "无效的认证令牌")
            c.Abort()
            return
        }

        // 4. 将用户信息存入Context
        c.Set("userID", claims.UserID)
        c.Set("username", claims.Username)

        c.Next()
    }
}
```

#### 5.8.2 Casbin权限中间件

```go
func CasbinMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. 从Context获取用户ID
        userIDAny, exists := c.Get("userID")
        if !exists {
            response.Unauthorized(c, "未登录")
            c.Abort()
            return
        }
        userID := userIDAny.(uint)

        // 2. 获取请求路径和方法
        obj := c.Request.URL.Path
        act := c.Request.Method

        // 3. Casbin权限校验
        e := casbin.GetEnforcer()
        ok, err := e.Enforce(fmt.Sprintf("%d", userID), obj, act)
        if err != nil {
            response.InternalError(c, "权限校验失败")
            c.Abort()
            return
        }

        if !ok {
            response.Forbidden(c, "权限不足")
            c.Abort()
            return
        }

        c.Next()
    }
}
```

#### 5.8.3 CustomRecovery Panic恢复中间件

```go
func CustomRecovery(logger *slog.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                // 1. 记录Panic日志（结构化）
                stack := make([]byte, 4096)
                n := runtime.Stack(stack, false)
                logger.Error("服务器内部Panic",
                    "error", err,
                    "path", c.Request.URL.Path,
                    "method", c.Request.Method,
                    "stack", string(stack[:n]),
                )

                // 2. 返回500 JSON响应（不泄露堆栈信息）
                response.InternalError(c, "服务器内部错误")
                c.Abort()
            }
        }()
        c.Next()
    }
}
```

### 5.9 处理器层（Handler）

Handler层负责HTTP协议相关处理：参数绑定、参数校验、调用Service、封装响应。

**UserHandler示例**：
```go
type UserHandler struct {
    service *service.UserService
    logger  *slog.Logger
}

func NewUserHandler(service *service.UserService, logger *slog.Logger) *UserHandler {
    return &UserHandler{service: service, logger: logger}
}

// List 用户列表
// @Summary 获取用户列表
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页条数" default(10)
// @Success 200 {object} response.Response
// @Router /api/v1/users [get]
func (h *UserHandler) List(c *gin.Context) {
    // 1. 绑定查询参数
    var query struct {
        Page     int `form:"page,default=1"`
        PageSize int `form:"page_size,default=10"`
    }
    if err := c.ShouldBindQuery(&query); err != nil {
        response.BadRequest(c, "参数错误: "+err.Error())
        return
    }

    // 2. 调用Service
    list, total, err := h.service.List(c.Request.Context(), query.Page, query.PageSize)
    if err != nil {
        response.InternalError(c, err.Error())
        return
    }

    // 3. 返回成功响应
    response.Success(c, gin.H{
        "list":  list,
        "total": total,
        "page":  query.Page,
        "page_size": query.PageSize,
    })
}

// Create 创建用户
func (h *UserHandler) Create(c *gin.Context) {
    var form model.CreateUserForm
    // 1. 绑定JSON参数（使用binding标签自动校验）
    if err := c.ShouldBindJSON(&form); err != nil {
        response.BadRequest(c, "参数错误: "+err.Error())
        return
    }

    // 2. 调用Service
    user, err := h.service.Create(c.Request.Context(), &form)
    if err != nil {
        response.BadRequest(c, err.Error())
        return
    }

    // 3. 返回成功
    response.Success(c, user)
}
```

### 5.10 公共包（pkg）

#### 5.10.1 response包 - 统一响应封装

**文件位置**：[response.go](file:///d:/Programming/Agent_demo/casbin_demo/backend/pkg/response/response.go)

统一响应格式，HTTP状态码与body.code严格匹配：

```go
// Response 统一响应结构
type Response struct {
    Code    int         `json:"code"`    // 业务码，与HTTP状态码一致
    Message string      `json:"message"` // 消息
    Data    interface{} `json:"data,omitempty"` // 数据
}

// Success 成功响应 (HTTP 200, code=200)
func Success(c *gin.Context, data interface{}) {
    c.JSON(http.StatusOK, Response{
        Code:    http.StatusOK,
        Message: "success",
        Data:    data,
    })
}

// Fail 失败响应（HTTP状态码与code参数一致）
func Fail(c *gin.Context, code int, message string) {
    c.JSON(code, Response{
        Code:    code,
        Message: message,
    })
}

// 常用错误响应快捷方法
func BadRequest(c *gin.Context, message string)  { Fail(c, 400, message) }
func Unauthorized(c *gin.Context, message string){ Fail(c, 401, message) }
func Forbidden(c *gin.Context, message string)   { Fail(c, 403, message) }
func NotFound(c *gin.Context, message string)    { Fail(c, 404, message) }
func InternalError(c *gin.Context, message string){ Fail(c, 500, message) }
```

#### 5.10.2 jwt包 - JWT工具

```go
type Claims struct {
    UserID   uint   `json:"user_id"`
    Username string `json:"username"`
    jwt.RegisteredClaims
}

// GenerateToken 生成JWT Token
func GenerateToken(userID uint, username string) (string, error) {
    claims := Claims{
        UserID:   userID,
        Username: username,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24小时过期
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "casbin-demo",
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtSecret)
}
```

---

## 第6章 Casbin权限模型详解

### 6.1 RBAC权限模型

系统采用Casbin标准的RBAC（基于角色的访问控制）模型，核心是**用户-角色-权限**三层关系：

```
┌─────────┐      多对多      ┌─────────┐      多对多      ┌─────────────┐
│  用户    │◄──────────────►│  角色    │◄──────────────►│   权限       │
│ (User)  │   user_roles    │ (Role)  │  role_perms    │ (Permission) │
└─────────┘                 └─────────┘                └─────────────┘
                                                        │
                                                        ▼
                                               ┌─────────────────┐
                                               │  API资源/方法    │
                                               │  /api/v1/users   │
                                               │  GET/POST/...    │
                                               └─────────────────┘
```

**Casbin中的角色继承关系**：
- 用户通过`g`策略分配给角色（`g, user_id, role_id`）
- 角色通过`p`策略绑定权限（`p, role_id, /api/v1/users, GET`）
- Casbin自动处理角色继承，用户拥有分配角色的所有权限

### 6.2 模型定义文件

**文件位置**：`backend/configs/rbac_model.conf`（或代码中定义）

```ini
[request_definition]
r = sub, obj, act
# sub: 访问主体（用户ID，字符串形式）
# obj: 访问对象（API路径，如 /api/v1/users）
# act: 访问动作（HTTP方法，如 GET/POST/PUT/DELETE）

[policy_definition]
p = sub, obj, act
# 策略定义格式与请求对应

[role_definition]
g = _, _
# 用户-角色映射：g, 用户ID, 角色ID

[policy_effect]
e = some(where (p.eft == allow))
# 只要有一条策略允许，最终结果为允许

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
# 匹配规则：
# 1. g(r.sub, p.sub): 请求用户属于某个角色（角色继承匹配）
# 2. r.obj == p.obj: 请求路径匹配
# 3. r.act == p.act: 请求方法匹配
```

### 6.3 权限校验流程

```
HTTP请求到达Casbin中间件
    │
    ▼
提取参数：sub=用户ID, obj=请求路径, act=请求方法
    │
    ▼
调用e.Enforce(sub, obj, act)
    │
    ├─ 遍历所有p策略（角色-资源-方法）
    │       │
    │       ├─ 匹配资源路径(obj)和方法(act)
    │       │       │
    │       │       └─ 匹配成功的角色集合：allowed_roles
    │       │
    │       └─ 无匹配 → 继续遍历
    │
    ├─ 遍历所有g策略（用户-角色映射）
    │       │
    │       └─ 检查用户拥有的角色是否在allowed_roles中
    │
    ├─ 存在匹配 → 返回true（允许访问）
    │
    └─ 无匹配 → 返回false（禁止访问）
```

### 6.4 策略管理

策略存储在数据库`casbin_rule`表中（通过gorm-adapter），通过Casbin API进行管理：

```go
// 分配角色给用户
e.AddGroupingPolicy(fmt.Sprintf("%d", userID), fmt.Sprintf("%d", roleID))

// 为角色添加权限
e.AddPolicy(fmt.Sprintf("%d", roleID), "/api/v1/users", "GET")

// 删除角色权限
e.RemovePolicy(fmt.Sprintf("%d", roleID), "/api/v1/users", "POST")

// 重新加载策略（修改后调用）
e.LoadPolicy()

// 保存策略到存储（内存变更后持久化）
e.SavePolicy()
```

**casbin_rule表结构**：

| 字段 | 类型 | 说明 | 示例 |
|-----|-----|-----|-----|
| id | bigint | 主键 | 1 |
| ptype | varchar(100) | 策略类型 | p（权限策略）/ g（角色映射） |
| v0 | varchar(100) | 主体 | 角色ID / 用户ID |
| v1 | varchar(100) | 对象 | API路径 / 角色ID |
| v2 | varchar(100) | 动作 | HTTP方法 / 空 |
| v3-v5 | varchar(100) | 预留 | - |

**策略示例数据**：

| ptype | v0 | v1 | v2 | 说明 |
|-------|----|----|----|-----|
| g | 1 | 1 | - | 用户1拥有角色1（admin） |
| p | 1 | /api/v1/users | GET | 角色1可以查看用户列表 |
| p | 1 | /api/v1/users | POST | 角色1可以创建用户 |
| p | 2 | /api/v1/users | GET | 角色2（普通用户）可以查看用户列表 |

---

## 第7章 前端架构详解

### 7.1 技术选型与版本

| 技术 | 版本 | 用途说明 |
|-----|-----|---------|
| Vue | 3.5.13 | 渐进式前端框架，使用Composition API和`<script setup>`语法 |
| TypeScript | 5.9.3 | JavaScript类型超集，提供编译时类型检查 |
| Vite | 7.2.4 | 下一代前端构建工具，开发环境极速热更新 |
| Pinia | 3.0.3 | Vue官方推荐状态管理库，替代Vuex |
| Vue Router | 4.6.4 | Vue.js官方路由管理器 |
| Element Plus | 2.13.7 | 基于Vue 3的企业级UI组件库 |
| Axios | 1.15.0 | HTTP客户端，用于发起API请求 |
| Sass | 1.99.2 | CSS预处理器，支持变量、嵌套、混合等特性 |
| @element-plus/icons-vue | 2.3.1 | Element Plus官方图标库 |

### 7.2 应用入口设计

**文件位置**：[main.ts](file:///d:/Programming/Agent_demo/casbin_demo/frontend/src/main.ts)

应用入口负责创建Vue实例、注册全局插件（Pinia、Router、Element Plus）：

```typescript
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'

import App from './App.vue'
import router from './router'
import './styles/index.scss'

const app = createApp(App)

// 注册所有Element Plus图标
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

app.use(createPinia())
app.use(router)
app.use(ElementPlus, {
  locale: zhCn,  // 中文语言包
  size: 'default'
})

app.mount('#app')
```

**根组件App.vue**：仅包含路由出口，使用transition实现页面切换动画。

### 7.3 状态管理（Pinia）

使用Pinia进行状态管理，定义两个核心Store：user store管理用户认证状态，app store管理全局UI状态。

#### 7.3.1 用户状态Store

**文件位置**：[user.ts](file:///d:/Programming/Agent_demo/casbin_demo/frontend/src/stores/user.ts)

```typescript
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { getToken, setToken, removeToken } from '@/utils/storage'
import { login, register, getUserInfo } from '@/api/auth'
import type { LoginForm, RegisterForm, UserInfo } from '@/api/auth'

export const useUserStore = defineStore('user', () => {
  // 状态
  const token = ref<string>(getToken() || '')
  const userInfo = ref<UserInfo | null>(null)

  // 计算属性
  const isLoggedIn = computed(() => !!token.value)
  const username = computed(() => userInfo.value?.username || '')
  const roles = computed(() => userInfo.value?.roles || [])

  // 登录
  async function loginAction(form: LoginForm) {
    const res = await login(form)
    token.value = res.data.token
    setToken(res.data.token)
    await fetchUserInfo()
    return res
  }

  // 注册
  async function registerAction(form: RegisterForm) {
    const res = await register(form)
    token.value = res.data.token
    setToken(res.data.token)
    await fetchUserInfo()
    return res
  }

  // 获取用户信息
  async function fetchUserInfo() {
    const res = await getUserInfo()
    userInfo.value = res.data
    return res.data
  }

  // 退出登录
  function logout() {
    token.value = ''
    userInfo.value = null
    removeToken()
  }

  return {
    token,
    userInfo,
    isLoggedIn,
    username,
    roles,
    loginAction,
    registerAction,
    fetchUserInfo,
    logout
  }
})
```

#### 7.3.2 本地存储封装

**文件位置**：[storage.ts](file:///d:/Programming/Agent_demo/casbin_demo/frontend/src/utils/storage.ts)

```typescript
const TOKEN_KEY = 'rbac_token'

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token)
}

export function removeToken(): void {
  localStorage.removeItem(TOKEN_KEY)
}
```

### 7.4 路由系统设计

**文件位置**：[index.ts](file:///d:/Programming/Agent_demo/casbin_demo/frontend/src/router/index.ts)

路由系统使用Vue Router 4，实现路由守卫（登录验证）、动态标题、错误路由处理。

```typescript
import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { ElMessage } from 'element-plus'

// 路由表定义
const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { title: '登录', public: true }
  },
  {
    path: '/',
    component: () => import('@/views/Layout.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '仪表盘' }
      },
      {
        path: 'users',
        name: 'UserManagement',
        component: () => import('@/views/UserManagement.vue'),
        meta: { title: '用户管理' }
      },
      {
        path: 'roles',
        name: 'RoleManagement',
        component: () => import('@/views/RoleManagement.vue'),
        meta: { title: '角色管理' }
      },
      {
        path: 'permissions',
        name: 'PermissionManagement',
        component: () => import('@/views/PermissionManagement.vue'),
        meta: { title: '权限管理' }
      }
    ]
  },
  // 错误页面路由
  { path: '/401', name: 'Error401', component: () => import('@/views/error/401.vue'), meta: { public: true, title: '401' } },
  { path: '/403', name: 'Error403', component: () => import('@/views/error/403.vue'), meta: { public: true, title: '403' } },
  { path: '/404', name: 'Error404', component: () => import('@/views/error/404.vue'), meta: { public: true, title: '404' } },
  { path: '/500', name: 'Error500', component: () => import('@/views/error/500.vue'), meta: { public: true, title: '500' } },
  // 404兜底
  { path: '/:pathMatch(.*)*', redirect: '/404' }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫：登录验证
router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  document.title = `${to.meta.title || 'RBAC系统'} - RBAC权限管理系统`

  // 公开页面直接放行
  if (to.meta.public) {
    next()
    return
  }

  // 未登录跳转到登录页
  if (!userStore.isLoggedIn) {
    ElMessage.warning('请先登录')
    next('/login')
    return
  }

  next()
})

export default router
```

### 7.5 Axios请求封装

**文件位置**：[request.ts](file:///d:/Programming/Agent_demo/casbin_demo/frontend/src/api/request.ts)

封装Axios实例，配置请求拦截器（添加Token和X-Requested-With头）和响应拦截器（统一错误处理）：

```typescript
import axios, { type AxiosInstance, type AxiosResponse, type InternalAxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'
import router from '@/router'
import { getToken } from '@/utils/storage'

// 响应类型定义
export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

// 创建Axios实例
const service: AxiosInstance = axios.create({
  baseURL: '/api/v1',
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json',
    'X-Requested-With': 'XMLHttpRequest' // 标识为AJAX请求，Nginx据此区分
  }
})

// 请求拦截器
service.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = getToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// 响应拦截器
service.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>) => {
    const res = response.data
    // code=200表示成功
    if (res.code !== 200) {
      ElMessage.error(res.message || '请求失败')
      return Promise.reject(new Error(res.message || '请求失败'))
    }
    return res
  },
  (error) => {
    const { response } = error

    if (response) {
      const { status } = response

      switch (status) {
        case 401:
          ElMessage.error('登录已过期，请重新登录')
          const userStore = useUserStore()
          userStore.logout()
          router.push('/login')
          break
        case 403:
          ElMessage.error('没有权限访问该资源')
          router.push('/403')
          break
        case 404:
          ElMessage.error('请求的资源不存在')
          break
        case 500:
          ElMessage.error('服务器内部错误')
          break
        default:
          ElMessage.error(response.data?.message || '网络错误')
      }
    } else {
      ElMessage.error('网络连接失败，请检查网络')
    }

    return Promise.reject(error)
  }
)

export default service
```

### 7.6 API接口层设计

API层按模块拆分，每个模块对应一个文件，定义该模块的所有API请求：

**auth.ts - 认证API**：
```typescript
import request from './request'

export interface LoginForm {
  username: string
  password: string
}

export interface RegisterForm {
  username: string
  password: string
  nickname?: string
  email?: string
}

export interface UserInfo {
  id: number
  username: string
  nickname: string
  email: string
  roles: Array<{ id: number; name: string }>
}

// 登录
export function login(data: LoginForm) {
  return request.post<any, ApiResponse<{ token: string }>>('/auth/login', data)
}

// 注册
export function register(data: RegisterForm) {
  return request.post<any, ApiResponse<{ token: string }>>('/auth/register', data)
}

// 获取用户信息
export function getUserInfo() {
  return request.get<any, ApiResponse<UserInfo>>('/userinfo')
}
```

其他API模块（user.ts、role.ts、permission.ts、dashboard.ts）结构类似，分别封装对应模块的CRUD接口。

### 7.7 布局组件设计

主布局组件[Layout.vue](file:///d:/Programming/Agent_demo/casbin_demo/frontend/src/views/Layout.vue)采用**侧边栏+顶部栏+内容区**经典管理后台布局：

```
┌─────────────────────────────────────────────────────────┐
│  Logo  │  折叠按钮          用户名▼  退出登录            │ ← Header
├────────┼────────────────────────────────────────────────┤
│        │                                                │
│  菜单  │              主内容区                           │
│  -仪表盘│         <router-view />                        │
│  -用户  │                                                │
│  -角色  │                                                │
│  -权限  │                                                │
│        │                                                │
└────────┴────────────────────────────────────────────────┘
        ← Sidebar(可折叠) →      ← Main Content →
```

**核心功能特性**：
- 侧边栏可折叠，折叠状态持久化到Pinia
- 响应式设计，小屏自动折叠
- Element Plus Menu组件实现导航菜单
- 顶部显示用户名和退出登录按钮
- 使用`<router-view>`配合`<transition>`实现页面切换动画
- 使用`v-loading`实现加载状态

### 7.8 错误页面系统

错误页面系统包含两个核心部分：通用错误组件ErrorPage.vue和四个具体错误页面（401/403/404/500）。

**ErrorPage.vue设计**：
- 接收`code`、`title`、`description`三个props
- 大号数字错误码（使用响应式字体）
- 错误标题和描述文字
- 操作按钮（返回首页/去登录）
- 统一的视觉风格

**401页面**：
- 错误码：401
- 标题：未授权访问
- 描述：您需要登录后才能访问此页面
- 按钮：去登录（跳转/login）、返回首页

**403页面**：
- 错误码：403
- 标题：访问被拒绝
- 描述：抱歉，您没有权限访问此页面
- 按钮：返回首页

**404页面**：
- 错误码：404
- 标题：页面未找到
- 描述：抱歉，您访问的页面不存在
- 按钮：返回首页

**500页面**：
- 错误码：500
- 标题：服务器错误
- 描述：抱歉，服务器遇到了一些问题
- 按钮：返回首页

**样式特点**：
- 使用玻璃拟态效果（backdrop-filter: blur）
- 渐变背景
- 按钮悬停动效
- 响应式布局适配移动端

### 7.9 样式与主题设计

**全局样式文件**：[index.scss](file:///d:/Programming/Agent_demo/casbin_demo/frontend/src/styles/index.scss)

核心样式设计：
1. **Element Plus主题定制**：通过CSS变量覆盖Element Plus默认主题色（使用蓝色系#409eff）
2. **全局重置样式**：统一margin/padding，设置box-sizing: border-box
3. **登录页玻璃拟态效果**：
   - backdrop-filter: blur(10px)
   - 半透明白色背景
   - 圆角阴影
4. **过渡动画**：页面切换使用fade-transform动画
5. **响应式断点**：适配移动端、平板、桌面端
6. **滚动条美化**：自定义滚动条样式

---

## 第8章 错误处理机制

系统构建了**后端→前端→Nginx**三层错误处理机制，确保不同场景下用户都能获得友好的错误提示。

### 8.1 后端错误处理

后端错误处理遵循以下原则：
- HTTP状态码与body.code严格一致
- 不返回栈信息给前端，内部使用slog记录
- 所有错误返回统一JSON格式：`{code, message, data}`

#### 8.1.1 Response包状态码规范

| 函数 | HTTP状态码 | body.code | 使用场景 |
|-----|-----------|----------|---------|
| Success | 200 | 200 | 请求成功 |
| BadRequest | 400 | 400 | 参数错误、验证失败 |
| Unauthorized | 401 | 401 | 未登录、Token无效/过期 |
| Forbidden | 403 | 403 | 权限不足 |
| NotFound | 404 | 404 | 资源不存在 |
| MethodNotAllowed | 405 | 405 | 请求方法不允许 |
| InternalError | 500 | 500 | 服务器内部错误、Panic |

#### 8.1.2 Panic恢复机制

通过CustomRecovery中间件捕获所有Handler中的Panic：
1. 捕获recover()
2. 使用slog记录ERROR级别日志，包含path、method、stack trace
3. 返回500 JSON响应给前端
4. 不中断服务，单个请求Panic不影响其他请求

#### 8.1.3 404/405处理

```go
// NoRoute：未匹配到路由时返回404 JSON（而非Gin默认404 HTML）
r.NoRoute(func(c *gin.Context) {
    response.NotFound(c, "请求的资源不存在")
})

// NoMethod：请求方法不匹配时返回405 JSON
r.NoMethod(func(c *gin.Context) {
    response.MethodNotAllowed(c, "请求方法不允许")
})
```

### 8.2 前端错误处理

前端错误处理包含Axios响应拦截器和路由守卫两部分：

#### 8.2.1 Axios拦截器错误处理

响应拦截器按HTTP状态码分类处理：

| 状态码 | 处理逻辑 |
|-------|---------|
| 401 | 清除Token和用户信息，ElMessage提示"登录已过期"，跳转/login |
| 403 | ElMessage提示"没有权限"，跳转/403错误页 |
| 404 | ElMessage提示"资源不存在" |
| 500 | ElMessage提示"服务器错误" |
| 网络错误 | ElMessage提示"网络连接失败" |
| 业务错误(code≠200) | ElMessage显示后端返回的message |

#### 8.2.2 路由守卫错误处理

- 未登录访问受保护页面：ElMessage.warning提示，跳转/login
- 登录后访问/login页：自动跳转/dashboard（已登录用户不需要再登录）

### 8.3 Nginx错误拦截

**关键设计**：Nginx区分**AJAX请求**和**浏览器直接访问**，采用不同的错误处理策略：

- AJAX请求（带`X-Requested-With: XMLHttpRequest`头）：让后端JSON响应直接返回，前端JS处理
- 浏览器直接访问（无X-Requested-With头）：Nginx拦截错误状态码，返回对应HTML错误页

**Nginx配置**：
```nginx
server {
    listen 80;
    server_name localhost;

    # 开启错误拦截
    proxy_intercept_errors on;

    # API请求代理到后端
    location /api/ {
        proxy_pass http://backend:8080;
        # ... 代理配置
    }

    # 错误页面处理：通过map指令区分AJAX和浏览器请求
    error_page 401 /401;
    error_page 403 /403;
    error_page 404 /404;
    error_page 500 502 503 504 /500;

    # 错误页面location：判断X-Requested-With头
    location ~ ^/(401|403|404|500)$ {
        internal;
        if ($http_x_requested_with = XMLHttpRequest) {
            # AJAX请求：返回JSON，让前端处理
            return 200 '{"code":$status,"message":"error"}';
            add_header Content-Type application/json;
        }
        # 浏览器直接访问：返回SPA的错误路由页面（try_files到index.html由前端路由处理）
        root /usr/share/nginx/html;
        try_files /index.html =404;
    }

    # 前端静态资源
    location / {
        root /usr/share/nginx/html;
        try_files $uri $uri/ /index.html; # Vue Router history模式
    }
}
```

**错误处理流程图**：
```
浏览器输入URL直接访问/api/userinfo（无Token）
    │
    ▼
Nginx接收请求 → 转发到后端
    │
    ▼
后端返回401 JSON: {code:401, message:"未授权"}
    │
    ▼
Nginx拦截401（proxy_intercept_errors on）
    │
    ├─ 检查X-Requested-With头
    │       │
    │       ├─ 有XMLHttpRequest → 返回JSON给前端处理
    │       │
    │       └─ 无 → 返回前端index.html，前端路由跳转到/401页面
    │
    ▼
用户看到友好的401错误页面（带"去登录"按钮）
```

### 8.4 错误状态码映射表

| HTTP状态码 | 场景 | 后端处理 | 前端AJAX处理 | 浏览器直接访问 |
|-----------|-----|---------|-------------|---------------|
| 200 | 请求成功 | 返回{code:200, data:...} | 返回resolve(data) | 正常显示数据 |
| 400 | 参数错误 | response.BadRequest | ElMessage提示错误消息 | Nginx返回400（本项目未做特殊拦截） |
| 401 | 未登录/Token无效 | response.Unauthorized | 清除Token跳转登录 | 显示401页面（去登录按钮） |
| 403 | 权限不足 | response.Forbidden | 跳转/403页面 | 显示403页面 |
| 404 | 资源不存在/路由不存在 | response.NotFound | ElMessage提示 | 显示404页面 |
| 405 | 方法不允许 | response.MethodNotAllowed | ElMessage提示 | Nginx默认（JSON或错误页） |
| 500 | 服务器错误/Panic | response.InternalError / Recovery | ElMessage提示 | 显示500页面 |

---

## 第9章 审核系统架构设计

### 9.1 系统概述与业务流程

审核系统是大模型API资源申请的核心审批流程，实现了从用户提交申请到管理员审核再到结果通知的全链路闭环。系统采用**PostgreSQL + RabbitMQ + Redis + WebSocket**的企业级架构组合，确保消息不丢、数据一致、高并发抗峰值。

#### 9.1.1 业务角色

| 角色 | 职责 |
|-----|-----|
| 普通用户 | 提交大模型API资源申请、查看申请状态、2分钟内撤回申请、接收审核结果通知 |
| 管理员 | 查看待审核列表、审核申请（通过/驳回）、接收新申请实时通知、查看历史审核记录 |

#### 9.1.2 核心业务流程

```
用户端                                 服务端                               管理员端
  │                                      │                                      │
  ├─ 填写资源申请表单                     │                                      │
  │  （资源名称/类型/API/Purpose/QPS等）  │                                      │
  └─────────────提交申请────────────────▶│                                      │
                                        │                                      │
                                        ├─ 1. PG事务写入申请表（状态=待审核）    │
                                        ├─ 2. 生成站内消息（未读）              │
                                        ├─ 3. Redis写2分钟撤回TTL              │
                                        ├─ 4. 发送RabbitMQ消息（新申请通知）    │
                                        └──────────────────────────────────────│
                                                                               │
                                                                               ├─ 管理员收到WebSocket弹窗通知
                                                                               │  （红点+通知列表实时更新）
                                                                               ├─ 点击查看详情→跳转审核页
                                                                               ├─ 审核操作：通过 / 驳回
                                                                               └────────────审核提交──────────▶│
                                                                                            │
                                        │◀──────────────────────────────────────┘
                                        │
                                        ├─ 1. 乐观锁校验状态（必须是待审核）    │
                                        ├─ 2. PG事务更新申请状态+审核信息      │
                                        ├─ 3. 生成审核结果站内消息             │
                                        ├─ 4. 发送RabbitMQ消息（审核结果）     │
                                        └──────────────────────────────────────│
  │                                      │                                      │
  ├─ 用户收到WebSocket弹窗通知           │                                      │
  │  （审核通过/驳回，可查看详情）        │                                      │
  └──────────────────────────────────────┘                                      │
```

#### 9.1.3 撤回业务流程

```
用户端                                 服务端                               管理员端
  │                                      │                                      │
  ├─ 提交申请后2分钟内                    │                                      │
  │  （倒计时显示，服务端Redis TTL控制）   │                                      │
  └─────────────撤回申请────────────────▶│                                      │
                                        │                                      │
                                        ├─ 1. 校验：申请人归属 + 状态=待审核    │
                                        ├─ 2. 校验：Redis TTL是否过期（兜底）   │
                                        ├─ 3. 乐观锁更新状态为已撤回           │
                                        ├─ 4. 删除Redis撤回TTL Key             │
                                        ├─ 5. 生成撤回站内消息                 │
                                        ├─ 6. 发送RabbitMQ撤回通知             │
                                        └──────────────────────────────────────│
                                                                               │
                                                                               ├─ 管理员收到撤回通知
                                                                               │  （申请从待审核列表消失）
                                                                               └───────────────────────────────┘
```

### 9.2 核心架构组件分工

| 组件 | 核心职责 | 唯一性定位 |
|-----|---------|-----------|
| **PostgreSQL** | 唯一可信持久层 | 存储申请表、站内消息全量数据；事务保证申请/撤回数据一致性；提供历史消息、已读/未读、审计记录；时间戳兜底判断2分钟撤回窗口 |
| **RabbitMQ** | 异步解耦 + 推送分发 + 削峰 | 用户提交申请、撤回申请异步推送通知；解耦主流程，HTTP响应不阻塞；多路由支持（新申请/撤回通知/审核结果）；可靠投递、死信重试、异常补发 |
| **Redis** | 高速热点 + 时效控制 + 跨网关广播 | 缓存管理员角色列表、未读消息计数；2分钟撤回窗口TTL控制（核心时效）；Redis PubSub实现多网关WebSocket全局广播；分布式限流、在线用户状态 |
| **WebSocket** | 前端实时弹窗 | 在线管理员/用户实时弹窗；区分「新申请通知 / 撤回通知 / 审核结果」弹窗类型；断线重连、离线消息PG兜底 |
| **Casbin** | 权限控制 | 动态获取全部管理员接收人；权限校验、防越权撤回、防越权审核 |
| **网关限流** | 流量防护 | 外层防CC、防刷；内层业务用户级限流 |

### 9.3 状态机与状态流转

#### 9.3.1 申请状态定义

| 状态值 | 状态名 | 说明 |
|-------|-------|-----|
| 0 | 待审核 (pending) | 申请已提交，等待管理员审核 |
| 1 | 已通过 (approved) | 管理员审核通过，资源已授权 |
| 2 | 已驳回 (rejected) | 管理员审核驳回，需重新申请 |
| 3 | 已撤回 (withdrawn) | 用户主动撤回申请 |

#### 9.3.2 状态流转图

```
  ┌─────────────┐
  │   待审核     │ 0
  └──────┬──────┘
         │
   ┌─────┴─────┐
   │           │
   ▼           ▼
┌──────┐  ┌──────┐     ┌─────────┐
│ 已通过│  │ 已驳回│     │  已撤回  │
│  1   │  │  2   │     │   3     │
└──────┘  └──────┘     └─────────┘
```

#### 9.3.3 状态流转规则

| 当前状态 | 操作 | 目标状态 | 操作人 | 校验条件 |
|---------|-----|---------|-------|---------|
| - | 提交申请 | 待审核 (0) | 申请人 | 参数校验通过 |
| 待审核 (0) | 审核通过 | 已通过 (1) | 管理员 | 状态=待审核 + 乐观锁Version匹配 |
| 待审核 (0) | 审核驳回 | 已驳回 (2) | 管理员 | 状态=待审核 + 乐观锁Version匹配 |
| 待审核 (0) | 撤回申请 | 已撤回 (3) | 申请人本人 | 状态=待审核 + 2分钟窗口内 + 申请人归属校验 + 乐观锁 |
| 已通过 (1) | - | - | - | 终态，不可变更 |
| 已驳回 (2) | - | - | - | 终态，不可变更 |
| 已撤回 (3) | - | - | - | 终态，不可变更 |

### 9.4 2分钟撤回机制设计

#### 9.4.1 设计原则

**服务端绝对控制**：撤回窗口的判断完全由服务端控制，前端倒计时仅作展示，不可作为撤回依据。防止用户篡改客户端时间无限撤回。

#### 9.4.2 双重时效校验机制

| 校验层级 | 实现方式 | 精度 | 作用 |
|---------|---------|-----|-----|
| **L1 Redis TTL** | Redis Key设置120秒过期（`withdraw:{app_id}`） | 毫秒级 | 快速判断、高性能、原子操作 |
| **L2 PG时间戳兜底** | `created_at` + 120秒与当前时间比较 | 秒级 | Redis故障时兜底、数据一致性最终保证 |

#### 9.4.3 撤回校验流程

```
用户发起撤回请求
       │
       ▼
  ┌────────────┐
  │ 鉴权中间件  │ → 校验JWT Token
  └──────┬─────┘
         │
         ▼
  ┌────────────┐
  │ Casbin权限  │ → 校验用户是否有撤回权限
  └──────┬─────┘
         │
         ▼
  ┌────────────┐
  │ 申请人归属  │ → 校验申请人ID == 当前用户ID（防越权）
  └──────┬─────┘
         │
         ▼
  ┌────────────┐
  │ 状态校验    │ → 状态必须是"待审核"(0)
  └──────┬─────┘
         │
         ▼
  ┌────────────┐
  │ Redis TTL  │ → EXISTS withdraw:{app_id} 判断是否在窗口内
  └──────┬─────┘
         │
         ▼
  ┌────────────┐
  │ PG时间戳兜底│ → now - created_at <= 120s（双重保险）
  └──────┬─────┘
         │
         ▼
  ┌────────────┐
  │ 乐观锁更新  │ → UPDATE ... WHERE id=? AND version=? AND status=0
  └──────┬─────┘
         │
         ▼
  ┌────────────┐
  │ 后置操作    │ → 删除Redis TTL Key / 发MQ通知 / 生成站内消息
  └────────────┘
```

#### 9.4.4 Redis Key设计

```
withdraw:ttl:{application_id}   →  无实际值，仅用TTL判断存在性，TTL=120s
audit:pending:count             →  待审核数量缓存，TTL=30s
```

### 9.5 并发安全保障

#### 9.5.1 乐观锁机制

使用数据库 `version` 字段实现乐观锁，防止"审核中同时被撤回"导致的数据状态错乱：

```go
// 撤回时的SQL
UPDATE audit_applications 
SET status = 3, version = version + 1, withdraw_reason = ?, withdrawn_at = NOW()
WHERE id = ? AND version = ? AND status = 0

// 审核时的SQL
UPDATE audit_applications 
SET status = ?, version = version + 1, reviewer_id = ?, review_comment = ?, reviewed_at = NOW()
WHERE id = ? AND version = ? AND status = 0
```

**冲突场景处理**：
- 管理员点击审核的同时用户撤回 → 后提交者会收到"申请状态已变更"的错误提示
- 两个管理员同时审核同一条申请 → 后提交者会收到"申请状态已变更"的错误提示

#### 9.5.2 幂等性保障

| 操作 | 幂等策略 |
|-----|---------|
| 提交申请 | 幂等Key + 同一用户同一API短时间内去重 |
| 审核操作 | 状态前置校验（必须是待审核）+ 乐观锁 |
| 撤回操作 | 状态前置校验（必须是待审核）+ 乐观锁 |
| MQ消息消费 | 消息去重表 + 消费幂等Key |

### 9.6 生产环境风险规避清单

#### 9.6.1 致命风险（必须100%修复）

| 风险描述 | 本系统解决方案 | 对应代码位置 |
|---------|--------------|------------|
| 先发MQ、后提交DB事务 → 弹窗成功但数据库回滚，数据幽灵消息 | **先DB事务提交，再发MQ**。事务失败不发消息，杜绝幽灵消息 | [audit_service.go](file:///workspace/backend/internal/service/audit_service.go) |
| MQ开启自动ACK → 推送失败但消息删除，消息永久丢失 | **手动ACK**，消费成功后才确认，失败则进入重试/死信队列 | [rabbitmq.go](file:///workspace/backend/pkg/mq/rabbitmq.go) |
| 多网关共用同一个MQ消费队列 → 只有部分管理员收到弹窗 | **fanout交换机 + 每网关实例独立队列**，所有实例都能收到消息，再由各实例的WebSocket Hub分发给在线用户 | [rabbitmq.go](file:///workspace/backend/pkg/mq/rabbitmq.go) |
| 依赖前端时间判断2分钟撤回 → 用户篡改时间无限撤回 | **服务端Redis TTL + PG时间戳双重校验**，前端倒计时仅作展示 | [audit_service.go](file:///workspace/backend/internal/service/audit_service.go) |
| 撤回未做状态乐观锁 → 审核中同时被撤回，数据状态错乱 | **version字段乐观锁**，更新时校验版本号和当前状态 | [audit_repo.go](file:///workspace/backend/internal/repository/audit_repo.go) |
| 事务内写Redis/MQ网络IO → 长事务、DB锁堆积、高并发死锁 | **事务内仅写DB**，Redis/MQ等网络IO操作放在事务提交之后 | [audit_service.go](file:///workspace/backend/internal/service/audit_service.go) |
| 撤回未校验申请人归属 → 可越权撤回他人申请 | **申请人ID归属校验** + Casbin权限双重校验 | [audit_service.go](file:///workspace/backend/internal/service/audit_service.go) |
| Redis PubSub无离线兜底 → 断线期间消息丢失 | **PG站内消息表兜底**，用户上线后拉取未读消息 | [audit_repo.go](file:///workspace/backend/internal/repository/audit_repo.go) |

#### 9.6.2 高风险（上线前必须整改）

| 风险描述 | 本系统解决方案 |
|---------|--------------|
| MQ消息体过大携带业务详情 → 网络拥堵、消费卡顿 | **消息体只传message_id**，消费端查PG获取详情 |
| 没有MQ失败补发机制 → MQ宕机消息永久丢失 | **定时对账调度器**（30s间隔）扫描未发送/未达状态，自动补发 |
| 未做缓存删除更新 → 管理员消息红点永久脏数据 | **操作后主动失效缓存**，确保数据一致性 |
| 协程闭包变量捕获问题 → 推送管理员列表错乱 | **使用索引遍历**，避免range闭包变量共享问题 |
| WebSocket无心跳、无重复连接剔除 → 在线用户状态错乱、重复弹窗 | **心跳机制（ping/pong 30s间隔） + 单用户连接数限制（MaxConnsPerUser=5）** |
| 未做幂等控制 → 同一消息多次弹窗刷屏 | **消息去重 + 幂等Key**，前端根据消息ID去重 |

#### 9.6.3 中风险（影响体验、稳定性）

| 风险描述 | 本系统解决方案 |
|---------|--------------|
| 未建立PG消息表GIN索引 → 未读消息查询全表扫描慢 | **B-tree复合索引**（user_id + is_read + created_at）覆盖高频查询 |
| 未限制单用户WS连接数 → 连接泄露、内存暴涨 | **MaxConnsPerUser=5**，超出时踢掉最早连接 |
| 撤回成功不主动删Redis缓存 → 缓存残留可二次撤回 | **撤回成功后DEL撤回TTL Key**，同时删除待审核计数缓存 |
| MQ无死信队列 → 异常消息无法排查 | **DLX死信交换机 + 死信队列**，异常消息入DLQ便于排查 |
| 未做定时对账补发 → 极端故障消息缺失无法发现 | **StartRetryScheduler定时调度**，周期性扫描补偿 |
| Redis、MQ无连接重连 → 中间件重启后服务不自动恢复 | **连接池自动重连** + 指数退避重试策略 |

---

## 第10章 WebSocket实时通信

### 10.1 WebSocket Hub架构设计

#### 10.1.1 Hub核心组件

| 组件 | 职责 | 实现位置 |
|-----|-----|---------|
| Hub | 连接管理中心，维护所有在线连接，负责消息广播、连接注册/注销 | [hub.go](file:///workspace/backend/pkg/ws/hub.go) |
| Client | 单个WebSocket连接抽象，负责消息收发、心跳保活、连接清理 | [hub.go](file:///workspace/backend/pkg/ws/hub.go) |
| Message | 消息结构，包含类型、接收人、业务数据 | [hub.go](file:///workspace/backend/pkg/ws/hub.go) |

#### 10.1.2 Hub架构图

```
                    ┌──────────────────────────────────────┐
                    │            Hub (全局单例)             │
                    │                                      │
                    │  ┌──────────────────────────────┐   │
                    │  │  clients: map[*Client]bool   │   │
                    │  │  所有在线连接集合             │   │
                    │  └──────────────────────────────┘   │
                    │                                      │
                    │  ┌──────────────────────────────┐   │
                    │  │  userClients: map[int][]*Client │ │
                    │  │  用户ID → 连接列表映射         │   │
                    │  └──────────────────────────────┘   │
                    │                                      │
                    │  broadcast  chan *Message            │
                    │  register   chan *Client             │
                    │  unregister chan *Client             │
                    │  userMessage chan *UserMessage       │
                    └──────────────┬───────────────────────┘
                                   │
                ┌──────────────────┼──────────────────┐
                │                  │                  │
                ▼                  ▼                  ▼
         ┌──────────┐        ┌──────────┐       ┌──────────┐
         │ Client 1 │        │ Client 2 │       │ Client N │
         │ user_id:1│        │ user_id:2│       │ user_id:N│
         └──────────┘        └──────────┘       └──────────┘
```

### 10.2 消息类型与事件定义

#### 10.2.1 通用消息结构

```json
{
  "type": "new_application",
  "data": { ... },
  "timestamp": 1719000000
}
```

#### 10.2.2 消息类型清单

| 消息类型 (type) | 接收人 | 触发场景 | data字段说明 |
|----------------|-------|---------|-------------|
| `new_application` | 所有管理员 | 用户提交新申请 | 申请摘要信息（id、申请人、资源名、类型） |
| `application_withdrawn` | 所有管理员 | 用户撤回申请 | 申请ID、撤回原因 |
| `review_result` | 申请人 | 管理员审核通过/驳回 | 申请ID、审核结果、审核意见 |
| `withdraw_confirmed` | 申请人 | 撤回操作确认（自己撤回自己也通知） | 申请ID、撤回状态 |
| `message_unread_count` | 特定用户 | 未读消息数变化 | unread_count |
| `ping` | 所有在线用户 | 心跳检测 | - |
| `pong` | 服务端 | 心跳响应 | - |

#### 10.2.3 消息发送优先级

| 优先级 | 消息类型 | 说明 |
|-------|---------|-----|
| P0 | review_result | 审核结果通知，用户最关心的实时反馈 |
| P0 | new_application | 新申请通知，管理员需要及时处理 |
| P1 | application_withdrawn | 撤回通知，更新管理员视图 |
| P1 | withdraw_confirmed | 撤回确认，用户操作回执 |
| P2 | message_unread_count | 未读计数更新，非紧急 |

### 10.3 连接管理与心跳机制

#### 10.3.1 连接生命周期

```
客户端发起WebSocket连接（/ws?token=xxx）
       │
       ▼
  ┌──────────────┐
  │  Token校验    │ → JWT鉴权，获取用户ID
  └──────┬───────┘
         │
         ▼
  ┌──────────────┐
  │  连接数限制   │ → 单用户最多5个连接，超出踢最早的
  └──────┬───────┘
         │
         ▼
  ┌──────────────┐
  │  注册到Hub    │ → clients + userClients 双索引
  └──────┬───────┘
         │
         ▼
  ┌──────────────┐
  │  发送未读计数 │ → 初始推送未读消息数
  └──────┬───────┘
         │
         ▼
  ┌──────────────┐
  │  消息收发循环 │ → readPump + writePump 双协程
  │  心跳保活     │ → ping 30s间隔，超时60s断开
  └──────┬───────┘
         │
         ▼
  ┌──────────────┐
  │  连接断开     │ → 从Hub注销，清理资源
  └──────────────┘
```

#### 10.3.2 心跳机制

| 参数 | 值 | 说明 |
|-----|-----|-----|
| 心跳间隔 | 30秒 | 服务端每30秒发一次ping |
| 超时时间 | 60秒 | 客户端60秒未响应pong则断开 |
| 写超时 | 10秒 | WebSocket写入超时时间 |
| 读缓冲区 | 1024字节 | 单条消息最大长度 |

#### 10.3.3 连接数限制策略

```go
const MaxConnsPerUser = 5  // 单用户最大连接数

// 超出限制时的处理：踢出最早建立的连接
// 保证同一用户在多端登录/多标签页场景下正常使用，
// 同时防止连接泄露导致内存暴涨
```

### 10.4 前端WebSocket客户端设计

#### 10.4.1 客户端特性

| 特性 | 实现方式 |
|-----|---------|
| **自动重连** | 断线后指数退避重连（1s → 2s → 4s → 8s → 最大30s） |
| **断线检测** | 心跳超时检测 + onclose事件触发重连 |
| **事件订阅** | 事件总线模式，支持on/off订阅/取消订阅 |
| **消息去重** | 基于消息ID的去重机制，防止重复弹窗 |
| **自动重发** | 连接恢复后自动拉取离线消息 |

#### 10.4.2 前端事件监听示例

```typescript
// 监听新申请通知（管理员）
wsService.on('new_application', (data) => {
  Notification({ title: '新申请待审核', message: data.resource_name })
  notificationCount.value++
})

// 监听审核结果（申请人）
wsService.on('review_result', (data) => {
  if (data.approved) {
    ElMessage.success(`申请「${data.resource_name}」已通过`)
  } else {
    ElMessage.warning(`申请「${data.resource_name}」被驳回`)
  }
  refreshApplicationList()
})

// 监听撤回确认（申请人）
wsService.on('withdraw_confirmed', (data) => {
  ElMessage.info('申请已撤回')
  refreshApplicationList()
})
```

### 10.5 离线消息与PG兜底方案

#### 10.5.1 设计思路

Redis PubSub + WebSocket 仅保障**在线时的实时推送**，用户离线期间的消息通过 **PostgreSQL站内消息表** 持久化存储，用户上线后主动拉取。

#### 10.5.2 离线消息补偿流程

```
用户建立WebSocket连接
       │
       ▼
  ┌──────────────┐
  │  连接成功     │
  └──────┬───────┘
         │
         ▼
  ┌──────────────┐
  │  拉取未读消息 │ → GET /messages/unread
  └──────┬───────┘
         │
         ▼
  ┌──────────────┐
  │  更新未读计数 │ → 前端红点展示
  └──────┬───────┘
         │
         ▼
  ┌──────────────┐
  │  用户点击通知 │ → 标记已读
  └──────────────┘
```

#### 10.5.3 消息表设计要点

- 索引：`(user_id, is_read, created_at DESC)` 复合索引，覆盖"查询某用户未读消息列表"高频查询
- 分页：历史消息分页查询，避免一次性加载过多
- 归档：定期归档历史已读消息，控制表体积

### 10.6 多网关跨实例广播

#### 10.6.1 问题背景

单实例WebSocket仅能推送本实例连接的用户。多网关（多实例部署）场景下，新申请通知需要推送给**所有网关节点上的在线管理员**。

#### 10.6.2 解决方案：Redis PubSub + MQ双保险

```
                   ┌─────────────────────────────────┐
                   │       用户提交申请               │
                   └──────────────┬──────────────────┘
                                  │
                                  ▼
                   ┌─────────────────────────────────┐
                   │  Gateway A (接收请求的实例)      │
                   │  - 写PG                          │
                   │  - 写Redis TTL                   │
                   └──────────────┬──────────────────┘
                                  │
                     ┌────────────┴────────────┐
                     ▼                         ▼
            ┌──────────────────┐      ┌──────────────────┐
            │   RabbitMQ       │      │  Redis PubSub    │
            │   (可靠投递)      │      │  (实时广播)      │
            └────────┬─────────┘      └────────┬─────────┘
                     │                          │
         ┌───────────┴───────────┐    ┌─────────┴───────────┐
         ▼                       ▼    ▼                     ▼
  ┌──────────────┐       ┌──────────────┐            ┌──────────────┐
  │  Gateway A   │       │  Gateway B   │            │  Gateway C   │
  │  WebSocket   │       │  WebSocket   │            │  WebSocket   │
  │  Hub         │       │  Hub         │            │  Hub         │
  └──────┬───────┘       └──────┬───────┘            └──────┬───────┘
         │                      │                           │
         ▼                      ▼                           ▼
  ┌──────────────┐       ┌──────────────┐            ┌──────────────┐
  │  在线管理员A1 │       │  在线管理员B1 │            │  在线管理员C1 │
  │  在线管理员A2 │       │  在线管理员B2 │            │  在线管理员C2 │
  └──────────────┘       └──────────────┘            └──────────────┘
```

**两种机制对比**：

| 机制 | 实时性 | 可靠性 | 适用场景 |
|-----|-------|-------|---------|
| Redis PubSub | 毫秒级 | 仅在线有效，不持久化 | 多网关实时广播，在线用户秒级送达 |
| RabbitMQ | 百毫秒级 | 持久化、可重试、死信队列 | 可靠投递保障，离线消息PG兜底 |

**为什么双保险**：
- Redis PubSub负责**实时性**，让在线用户立即收到通知
- RabbitMQ负责**可靠性**，确保消息不丢，消费失败可重试
- 即使其中一个组件故障，另一个仍能保证消息送达（降级）

---

## 第11章 RabbitMQ消息队列

### 11.1 消息队列架构设计

#### 11.1.1 核心设计原则

| 原则 | 说明 |
|-----|-----|
| **异步解耦** | 主流程（提交申请/审核）不阻塞于通知推送，HTTP快速响应 |
| **削峰填谷** | 高并发申请场景下，MQ缓冲流量，消费端按能力消费 |
| **可靠投递** | 手动ACK + 持久化 + 死信队列 + 定时补发，确保消息不丢 |
| **轻量消息** | 消息体只传ID，详情查DB，避免消息过大 |

#### 11.1.2 消息生产者与消费者

| 消息类型 | 生产者 | 消费者 | 触发时机 |
|---------|-------|-------|---------|
| 新申请通知 | 提交申请接口 | 所有网关实例 | 用户提交大模型API申请 |
| 撤回通知 | 撤回申请接口 | 所有网关实例 | 用户在2分钟内撤回申请 |
| 审核结果通知 | 审核接口 | 对应网关实例 | 管理员审核通过/驳回 |

### 11.2 Exchange与Queue规划

#### 11.2.1 Exchange设计

| Exchange名称 | 类型 | 用途 |
|-------------|-----|-----|
| `audit.notifications` | fanout | 审核通知广播交换机（新申请、撤回） |
| `audit.review` | direct | 审核结果定向交换机（按用户路由） |
| `audit.dlx` | fanout | 死信交换机 |

#### 11.2.2 Queue设计

**广播类队列（每实例一个，fanout）**：

| 队列名 | 绑定交换机 | 说明 |
|-------|-----------|-----|
| `audit.notifications.{instance_id}` | audit.notifications | 每实例独立队列，确保所有实例都能收到广播消息 |

**定向队列**：

| 队列名 | 绑定交换机 | Routing Key | 说明 |
|-------|-----------|-------------|-----|
| `audit.review.user` | audit.review | user.{user_id} | 按用户ID路由审核结果 |

**死信队列**：

| 队列名 | 绑定交换机 | 说明 |
|-------|-----------|-----|
| `audit.dlq` | audit.dlx | 死信队列，存放消费失败的消息，便于排查和人工处理 |

#### 11.2.3 为什么广播用fanout而不是direct？

**问题**：如果所有网关共用同一个队列消费新申请通知，消息会被轮询分发给各消费者，导致只有部分管理员收到弹窗。

**解决方案**：使用fanout交换机，每个网关实例绑定一个**独立队列**，广播消息会复制到所有队列，所有实例都能收到完整通知。

```
                      ┌──────────────────┐
                      │  Exchange:       │
                      │  audit.notifica- │
                      │  tions (fanout)  │
                      └────────┬─────────┘
                               │
              ┌────────────────┼────────────────┐
              ▼                ▼                ▼
    ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
    │ Queue: notify.  │ │ Queue: notify.  │ │ Queue: notify.  │
    │ gw-A            │ │ gw-B            │ │ gw-C            │
    └────────┬────────┘ └────────┬────────┘ └────────┬────────┘
             │                   │                   │
             ▼                   ▼                   ▼
    ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
    │  Gateway A      │ │  Gateway B      │ │  Gateway C      │
    │  Consumer       │ │  Consumer       │ │  Consumer       │
    └─────────────────┘ └─────────────────┘ └─────────────────┘
```

### 11.3 可靠投递保障机制

#### 11.3.1 四重可靠保障

| 层级 | 机制 | 作用 |
|-----|-----|-----|
| 1 | **消息持久化** | Queue和Message都设置durable=true，MQ重启消息不丢 |
| 2 | **生产者确认** | Publisher Confirm，确保消息成功到达Broker |
| 3 | **手动ACK** | 消费成功后才ack，失败则nack/requeue |
| 4 | **定时对账补发** | 周期性扫描数据库状态，补偿丢失的消息 |

#### 11.3.2 消费确认机制

```go
// 消费流程
1. 接收消息
2. 处理业务逻辑（推送WebSocket等）
3. 处理成功 → Ack(false)  确认消息
4. 处理失败 → Nack(true)  重新入队重试
   - 重试超过3次 → 进入死信队列
```

**重试策略**：
- 第1次失败：立即重试
- 第2次失败：延迟5秒重试
- 第3次失败：进入死信队列，人工介入

### 11.4 死信队列与异常处理

#### 11.4.1 死信触发条件

| 触发条件 | 说明 |
|---------|-----|
| 消息被Nack且requeue=false | 消费端主动拒绝，不重试 |
| 消息TTL过期 | 消息在队列中存活时间超过设定值 |
| 队列达到最大长度 | 队列满了，新消息进入死信 |

#### 11.4.2 死信队列用途

- **问题排查**：死信消息保留完整上下文，便于定位消费失败原因
- **人工处理**：对于无法自动恢复的消息，人工介入处理
- **监控告警**：死信队列堆积告警，提示系统异常

### 11.5 定时对账与补发机制

#### 11.5.1 为什么需要定时对账？

即使有了持久化、ACK、死信队列，仍可能出现消息丢失的极端场景：
- MQ宕机期间生产者消息未发送成功
- 网络分区导致消息丢失
- 消费者处理成功但ACK时网络断开

**定时对账**是最后一道防线，确保消息最终可达。

#### 11.5.2 对账机制设计

| 参数 | 值 | 说明 |
|-----|-----|-----|
| 调度间隔 | 30秒 | 每30秒执行一次对账扫描 |
| 扫描范围 | 最近1小时 | 扫描1小时内创建但状态不匹配的记录 |
| 补发次数上限 | 3次 | 单条消息最多补发3次，避免无限重试 |

#### 11.5.3 对账流程

```
定时调度触发（每30s）
       │
       ▼
  ┌──────────────┐
  │ 扫描最近1h内 │
  │ 未发送成功的  │
  │ 消息记录      │
  └──────┬───────┘
         │
         ▼
  ┌──────────────┐
  │ 补发次数<3？  │
  └──────┬───────┘
         │是
         ▼
  ┌──────────────┐
  │ 重新发送MQ    │
  │ 补发次数+1    │
  └──────────────┘

         否
         │
         ▼
  ┌──────────────┐
  │ 标记为失败    │
  │ 告警通知      │
  └──────────────┘
```

### 11.6 高可用降级策略

#### 11.6.1 MQ故障降级

当RabbitMQ不可用时，系统自动降级：

| 降级层级 | 触发条件 | 降级策略 |
|---------|---------|---------|
| L1 | MQ连接失败 | 跳过MQ发送，直接通过Redis PubSub广播 + PG持久化兜底 |
| L2 | Redis也故障 | 仅靠PG持久化 + 用户主动刷新页面获取最新状态 |
| L3 | PG也故障 | 直接拒绝服务，返回503，保证数据不脏写 |

#### 11.6.2 降级设计原则

- **优雅降级**：功能从"实时推送"降级为"刷新可见"，但数据一致性始终保证
- **自动恢复**：中间件恢复后，连接自动重连，功能自动恢复
- **可观测**：降级期间有日志和告警，便于运维感知

---

## 第12章 API接口文档

### 9.1 通用说明

#### 9.1.1 基础信息

- **Base URL**：`http://localhost/api/v1`（生产环境）或 `http://localhost:3000/api/v1`（开发环境）
- **认证方式**：Bearer Token（JWT），放在请求头`Authorization: Bearer {token}`
- **Content-Type**：`application/json`
- **字符编码**：UTF-8

#### 9.1.2 统一响应格式

**成功响应 (HTTP 200)**：
```json
{
  "code": 200,
  "message": "success",
  "data": { ... }
}
```

**失败响应 (HTTP状态码 = code)**：
```json
{
  "code": 400,
  "message": "错误描述信息"
}
```

#### 9.1.3 分页参数说明

列表接口通用分页参数：

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|-----|-----|-----|-------|-----|
| page | int | 否 | 1 | 页码，从1开始 |
| page_size | int | 否 | 10 | 每页条数 |

分页响应格式：
```json
{
  "code": 200,
  "data": {
    "list": [ ... ],
    "total": 100,
    "page": 1,
    "page_size": 10
  }
}
```

### 9.2 认证接口

#### 9.2.1 用户注册

- **URL**：`POST /api/v1/auth/register`
- **认证**：不需要
- **请求体**：

| 字段 | 类型 | 必填 | 说明 |
|-----|-----|-----|-----|
| username | string | 是 | 用户名，3-50字符 |
| password | string | 是 | 密码，6-50字符 |
| nickname | string | 否 | 昵称，最长50字符 |
| email | string | 否 | 邮箱格式 |

```json
{
  "username": "admin",
  "password": "123456",
  "nickname": "管理员",
  "email": "T1T2c@PjXkDek.47v"
}
```

- **成功响应 (200)**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

- **失败响应 (400)**：用户名已存在
```json
{
  "code": 400,
  "message": "用户名已存在"
}
```

#### 9.2.2 用户登录

- **URL**：`POST /api/v1/auth/login`
- **认证**：不需要
- **请求体**：

| 字段 | 类型 | 必填 | 说明 |
|-----|-----|-----|-----|
| username | string | 是 | 用户名 |
| password | string | 是 | 密码 |

```json
{
  "username": "admin",
  "password": "123456"
}
```

- **成功响应 (200)**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

- **失败响应 (401)**：用户名或密码错误
```json
{
  "code": 401,
  "message": "用户名或密码错误"
}
```

#### 9.2.3 获取当前用户信息

- **URL**：`GET /api/v1/userinfo`
- **认证**：需要（Bearer Token）
- **请求参数**：无
- **成功响应 (200)**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "username": "admin",
    "nickname": "管理员",
    "email": "T1T2c@PjXkDek.47v",
    "status": 1,
    "roles": [
      {
        "id": 1,
        "name": "admin",
        "description": "超级管理员"
      }
    ],
    "created_at": "2026-06-25T10:00:00Z"
  }
}
```

- **失败响应 (401)**：Token无效或过期

### 9.3 用户管理接口

#### 9.3.1 获取用户列表

- **URL**：`GET /api/v1/users`
- **认证**：需要（需要users:list权限）
- **查询参数**：`page`, `page_size`
- **成功响应 (200)**：
```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "id": 1,
        "username": "admin",
        "nickname": "管理员",
        "email": "T1T2c@PjXkDek.47v",
        "status": 1,
        "roles": [{"id": 1, "name": "admin"}],
        "created_at": "2026-06-25T10:00:00Z"
      }
    ],
    "total": 10,
    "page": 1,
    "page_size": 10
  }
}
```

#### 9.3.2 创建用户

- **URL**：`POST /api/v1/users`
- **认证**：需要（需要users:create权限）
- **请求体**：

| 字段 | 类型 | 必填 | 说明 |
|-----|-----|-----|-----|
| username | string | 是 | 用户名 |
| password | string | 是 | 密码 |
| nickname | string | 否 | 昵称 |
| email | string | 否 | 邮箱 |
| role_ids | number[] | 否 | 角色ID数组 |

#### 9.3.3 获取用户详情

- **URL**：`GET /api/v1/users/:id`
- **认证**：需要（需要users:read权限）
- **路径参数**：id - 用户ID

#### 9.3.4 更新用户

- **URL**：`PUT /api/v1/users/:id`
- **认证**：需要（需要users:update权限）

#### 9.3.5 删除用户

- **URL**：`DELETE /api/v1/users/:id`
- **认证**：需要（需要users:delete权限）

### 9.4 角色管理接口

#### 9.4.1 获取角色列表

- **URL**：`GET /api/v1/roles`
- **认证**：需要（需要roles:list权限）

#### 9.4.2 创建角色

- **URL**：`POST /api/v1/roles`
- **认证**：需要（需要roles:create权限）
- **请求体**：

| 字段 | 类型 | 必填 | 说明 |
|-----|-----|-----|-----|
| name | string | 是 | 角色名称（英文标识） |
| description | string | 否 | 角色描述 |
| permission_ids | number[] | 否 | 权限ID数组 |

#### 9.4.3 获取角色详情

- **URL**：`GET /api/v1/roles/:id`
- **认证**：需要（需要roles:read权限）

#### 9.4.4 更新角色

- **URL**：`PUT /api/v1/roles/:id`
- **认证**：需要（需要roles:update权限）

#### 9.4.5 删除角色

- **URL**：`DELETE /api/v1/roles/:id`
- **认证**：需要（需要roles:delete权限）

#### 9.4.6 分配角色权限

- **URL**：`POST /api/v1/roles/:id/permissions`
- **认证**：需要
- **请求体**：
```json
{
  "permission_ids": [1, 2, 3]
}
```

### 9.5 权限管理接口

#### 9.5.1 获取权限列表

- **URL**：`GET /api/v1/permissions`
- **认证**：需要（需要permissions:list权限）
- **成功响应**：返回权限树结构

#### 9.5.2 创建权限

- **URL**：`POST /api/v1/permissions`
- **认证**：需要
- **请求体**：

| 字段 | 类型 | 必填 | 说明 |
|-----|-----|-----|-----|
| name | string | 是 | 权限名称 |
| code | string | 是 | 权限标识（如users:list） |
| type | string | 是 | 类型：menu/button/api |
| parent_id | number | 否 | 父权限ID |
| path | string | 否 | API路径 |
| method | string | 否 | HTTP方法 |

#### 9.5.3-9.5.5 详情/更新/删除

与用户/角色接口类似，路径为`/api/v1/permissions/:id`。

### 9.6 仪表盘接口

#### 9.6.1 获取统计数据

- **URL**：`GET /api/v1/dashboard/stats`
- **认证**：需要
- **成功响应 (200)**：
```json
{
  "code": 200,
  "data": {
    "user_count": 100,
    "role_count": 5,
    "permission_count": 30,
    "recent_users": [...]
  }
}
```

### 12.7 审核系统接口

审核系统接口用于大模型API资源的申请、审核、撤回等操作，所有接口均需JWT认证。

#### 12.7.1 提交资源申请

- **URL**：`POST /api/v1/audit/applications`
- **认证**：需要（普通用户权限）
- **权限**：`audit:application:create`
- **描述**：普通用户提交大模型API资源审核申请

**请求体**：

| 字段 | 类型 | 必填 | 说明 |
|-----|-----|-----|-----|
| resource_name | string | 是 | 资源名称，如"GPT-4 API"，最大100字符 |
| resource_type | string | 是 | 资源类型：llm(大语言模型)/embedding(向量模型)/image(图像模型)/audio(音频模型) |
| api_name | string | 是 | API标识名称，如"gpt-4"、"text-embedding-3"，最大50字符 |
| api_description | string | 否 | API详细描述，最大500字符 |
| purpose | string | 是 | 使用用途说明，最大1000字符 |
| expected_qps | int | 是 | 预估QPS，范围1-10000 |
| contact_info | string | 是 | 联系方式（邮箱/电话），最大100字符 |

```json
{
  "resource_name": "GPT-4 Turbo API",
  "resource_type": "llm",
  "api_name": "gpt-4-turbo",
  "api_description": "GPT-4 Turbo大语言模型，支持128K上下文",
  "purpose": "用于内部智能客服系统的开发和测试，预计日均调用量约1万次",
  "expected_qps": 50,
  "contact_info": "T1T2c@PjXkDek.47v"
}
```

**成功响应 (200)**：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "uuid": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "applicant_id": 2,
    "applicant_name": "张三",
    "resource_name": "GPT-4 Turbo API",
    "resource_type": "llm",
    "api_name": "gpt-4-turbo",
    "api_description": "GPT-4 Turbo大语言模型，支持128K上下文",
    "purpose": "用于内部智能客服系统的开发和测试",
    "expected_qps": 50,
    "contact_info": "T1T2c@PjXkDek.47v",
    "status": 0,
    "status_text": "待审核",
    "created_at": "2026-06-25T10:30:00Z",
    "can_withdraw": true,
    "withdraw_remain_ms": 119500
  }
}
```

**失败响应 (400)**：参数校验失败
```json
{
  "code": 400,
  "message": "resource_name不能为空"
}
```

**失败响应 (403)**：无提交权限
```json
{
  "code": 403,
  "message": "没有提交申请的权限"
}
```

#### 12.7.2 获取我的申请列表

- **URL**：`GET /api/v1/audit/applications/my`
- **认证**：需要
- **描述**：获取当前登录用户的申请列表，支持分页和状态筛选

**查询参数**：

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|-----|-----|-----|-------|-----|
| page | int | 否 | 1 | 页码，从1开始 |
| page_size | int | 否 | 10 | 每页条数，最大100 |
| status | int | 否 | - | 按状态筛选：0-待审核 1-已通过 2-已驳回 3-已撤回 |

**成功响应 (200)**：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "uuid": "a1b2c3d4-...",
        "resource_name": "GPT-4 Turbo API",
        "resource_type": "llm",
        "api_name": "gpt-4-turbo",
        "status": 0,
        "status_text": "待审核",
        "created_at": "2026-06-25T10:30:00Z",
        "can_withdraw": true,
        "withdraw_remain_ms": 119500
      }
    ],
    "total": 25,
    "page": 1,
    "page_size": 10
  }
}
```

#### 12.7.3 获取全部申请列表（管理员）

- **URL**：`GET /api/v1/audit/applications`
- **认证**：需要（管理员权限）
- **权限**：`audit:application:list`
- **描述**：管理员获取所有申请列表，支持分页、状态筛选、关键词搜索

**查询参数**：

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|-----|-----|-----|-------|-----|
| page | int | 否 | 1 | 页码，从1开始 |
| page_size | int | 否 | 10 | 每页条数，最大100 |
| status | int | 否 | - | 按状态筛选：0-待审核 1-已通过 2-已驳回 3-已撤回 |
| keyword | string | 否 | - | 关键词搜索（资源名称/申请人/API名称） |

**成功响应 (200)**：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "uuid": "a1b2c3d4-...",
        "applicant_id": 2,
        "applicant_name": "张三",
        "resource_name": "GPT-4 Turbo API",
        "resource_type": "llm",
        "api_name": "gpt-4-turbo",
        "purpose": "用于内部智能客服系统",
        "expected_qps": 50,
        "status": 0,
        "status_text": "待审核",
        "created_at": "2026-06-25T10:30:00Z",
        "can_withdraw": false
      }
    ],
    "total": 100,
    "page": 1,
    "page_size": 10
  }
}
```

#### 12.7.4 获取申请详情

- **URL**：`GET /api/v1/audit/applications/:id`
- **认证**：需要
- **描述**：根据ID获取申请详情
- **权限说明**：
  - 管理员可查看所有申请
  - 普通用户只能查看自己的申请

**路径参数**：

| 参数 | 类型 | 说明 |
|-----|-----|-----|
| id | int | 申请ID |

**成功响应 (200)**：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "uuid": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "applicant_id": 2,
    "applicant_name": "张三",
    "resource_name": "GPT-4 Turbo API",
    "resource_type": "llm",
    "api_name": "gpt-4-turbo",
    "api_description": "GPT-4 Turbo大语言模型，支持128K上下文",
    "purpose": "用于内部智能客服系统的开发和测试",
    "expected_qps": 50,
    "contact_info": "T1T2c@PjXkDek.47v",
    "status": 0,
    "status_text": "待审核",
    "reviewer_id": null,
    "reviewer_name": null,
    "review_comment": null,
    "reviewed_at": null,
    "created_at": "2026-06-25T10:30:00Z",
    "can_withdraw": true,
    "withdraw_remain_ms": 118000,
    "withdraw_reason": null,
    "withdrawn_at": null
  }
}
```

**失败响应 (403)**：越权查看他人申请
```json
{
  "code": 403,
  "message": "无权查看该申请"
}
```

**失败响应 (404)**：申请不存在
```json
{
  "code": 404,
  "message": "申请不存在"
}
```

#### 12.7.5 审核申请（管理员）

- **URL**：`POST /api/v1/audit/applications/:id/review`
- **认证**：需要（管理员权限）
- **权限**：`audit:application:review`
- **描述**：管理员对申请进行审核，通过或驳回

**路径参数**：

| 参数 | 类型 | 说明 |
|-----|-----|-----|
| id | int | 申请ID |

**请求体**：

| 字段 | 类型 | 必填 | 说明 |
|-----|-----|-----|-----|
| approved | boolean | 是 | 审核结果：true-通过 false-驳回 |
| comment | string | 否 | 审核意见，驳回时建议填写，最大500字符 |

```json
{
  "approved": true,
  "comment": "申请合理，已开通50QPS额度，使用期间请注意监控"
}
```

**成功响应 (200)**：

```json
{
  "code": 200,
  "message": "审核成功"
}
```

**失败响应 (400)**：申请状态不是待审核
```json
{
  "code": 400,
  "message": "申请状态已变更，无法审核"
}
```

**失败响应 (404)**：申请不存在
```json
{
  "code": 404,
  "message": "申请不存在"
}
```

#### 12.7.6 撤回申请

- **URL**：`POST /api/v1/audit/applications/:id/withdraw`
- **认证**：需要
- **权限**：`audit:application:withdraw`
- **描述**：用户在提交后2分钟内撤回申请
- **限制**：
  - 只能撤回自己提交的申请
  - 只能撤回状态为"待审核"的申请
  - 必须在提交后2分钟内撤回（服务端时间校验）

**路径参数**：

| 参数 | 类型 | 说明 |
|-----|-----|-----|
| id | int | 申请ID |

**请求体**：

| 字段 | 类型 | 必填 | 说明 |
|-----|-----|-----|-----|
| reason | string | 否 | 撤回原因，最大200字符 |

```json
{
  "reason": "信息填写有误，需要重新提交"
}
```

**成功响应 (200)**：

```json
{
  "code": 200,
  "message": "撤回成功"
}
```

**失败响应 (400)**：超过撤回时间窗口
```json
{
  "code": 400,
  "message": "已超过撤回时间（2分钟），无法撤回"
}
```

**失败响应 (400)**：申请状态不是待审核
```json
{
  "code": 400,
  "message": "当前状态无法撤回"
}
```

**失败响应 (403)**：越权撤回他人申请
```json
{
  "code": 403,
  "message": "无权撤回该申请"
}
```

#### 12.7.7 获取待审核数量（管理员）

- **URL**：`GET /api/v1/audit/applications/pending/count`
- **认证**：需要（管理员权限）
- **权限**：`audit:application:list`
- **描述**：获取当前待审核的申请数量，用于导航栏红点提示

**成功响应 (200)**：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "count": 15
  }
}
```

#### 12.7.8 WebSocket连接

- **URL**：`GET /ws` (WebSocket协议)
- **认证**：需要（Token通过query参数传递）
- **描述**：建立WebSocket长连接，接收实时通知消息

**连接参数**：

| 参数 | 位置 | 说明 |
|-----|-----|-----|
| token | query | JWT Token，如 `/ws?token=eyJhbGciOiJIUzI1NiIs...` |

**消息格式**：

```json
{
  "type": "new_application",
  "data": {
    "id": 1,
    "resource_name": "GPT-4 Turbo API",
    "applicant_name": "张三"
  },
  "timestamp": 1719292200
}
```

**消息类型列表**：

| type | 接收人 | 触发场景 | data说明 |
|------|-------|---------|---------|
| `new_application` | 管理员 | 用户提交新申请 | {id, resource_name, applicant_name, api_name} |
| `application_withdrawn` | 管理员 | 用户撤回申请 | {id, reason} |
| `review_result` | 申请人 | 管理员审核完成 | {id, approved, comment, resource_name} |
| `withdraw_confirmed` | 申请人 | 撤回操作确认 | {id, status} |
| `message_unread_count` | 所有用户 | 未读消息数变化 | {unread_count} |
| `ping` | 所有用户 | 心跳检测 | - |

#### 12.7.9 审核系统状态码

| 状态值 | 状态名 | 说明 |
|-------|-------|-----|
| 0 | 待审核 | 申请已提交，等待管理员审核 |
| 1 | 已通过 | 管理员审核通过 |
| 2 | 已驳回 | 管理员审核驳回 |
| 3 | 已撤回 | 用户主动撤回 |

### 12.8 错误码说明

| 错误码 | HTTP状态 | 说明 | 处理建议 |
|-------|---------|-----|---------|
| 200 | 200 OK | 请求成功 | - |
| 400 | 400 Bad Request | 参数错误、业务逻辑错误 | 检查请求参数 |
| 401 | 401 Unauthorized | 未登录、Token无效/过期 | 重新登录 |
| 403 | 403 Forbidden | 权限不足 | 联系管理员分配权限 |
| 404 | 404 Not Found | 资源不存在 | 检查请求URL和ID |
| 405 | 405 Method Not Allowed | 请求方法错误 | 检查HTTP方法（GET/POST等） |
| 500 | 500 Internal Server Error | 服务器内部错误 | 查看后端日志，联系开发人员 |

---

## 第13章 部署指南

### 10.1 Docker Compose部署

使用Docker Compose一键部署完整服务（PostgreSQL + Redis + Backend + Frontend/Nginx）。

#### 10.1.1 环境要求

| 软件 | 最低版本 |
|-----|---------|
| Docker | 20.10+ |
| Docker Compose | v2+ |
| 内存 | ≥ 2GB |
| 磁盘 | ≥ 10GB |

#### 10.1.2 部署步骤

```bash
# 1. 克隆项目
git clone <project-url>
cd casbin_demo

# 2. 构建并启动所有服务（后台运行）
docker-compose up -d --build

# 3. 查看服务状态
docker-compose ps

# 4. 查看服务日志
docker-compose logs -f backend
docker-compose logs -f frontend

# 5. 停止服务
docker-compose down

# 6. 停止服务并删除数据卷（清空数据库）
docker-compose down -v
```

#### 10.1.3 服务访问地址

部署成功后，访问以下地址：

| 服务 | 地址 | 说明 |
|-----|-----|-----|
| 前端应用 | http://localhost | Nginx服务，包含前端和API代理 |
| 后端API | http://localhost/api/v1 | 通过Nginx代理到后端 |
| 后端直连 | http://localhost:8080 | 后端服务直接端口（不推荐） |

#### 10.1.4 docker-compose.yml服务配置

```yaml
version: '3.8'

services:
  # PostgreSQL数据库
  postgres:
    image: postgres:15-alpine
    container_name: rbac-postgres
    environment:
      POSTGRES_DB: casbin_demo
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres123
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - rbac-network
    restart: unless-stopped

  # Redis缓存
  redis:
    image: redis:7-alpine
    container_name: rbac-redis
    command: redis-server --requirepass redis123
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "-a", "redis123", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - rbac-network
    restart: unless-stopped

  # 后端Go应用
  backend:
    build:
      context: ./backend
    container_name: rbac-backend
    environment:
      CASBIN_DEMO_DATABASE_HOST: postgres
      CASBIN_DEMO_DATABASE_PASSWORD: postgres123
      CASBIN_DEMO_REDIS_HOST: redis
      CASBIN_DEMO_REDIS_PASSWORD: redis123
      CASBIN_DEMO_SERVER_MODE: release
    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - rbac-network
    restart: unless-stopped

  # 前端Nginx服务
  frontend:
    build:
      context: ./frontend
    container_name: rbac-frontend
    ports:
      - "80:80"
    depends_on:
      - backend
    networks:
      - rbac-network
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:

networks:
  rbac-network:
    driver: bridge
```

#### 10.1.5 数据持久化

数据通过Docker Volume持久化：
- `postgres_data`：PostgreSQL数据库文件
- `redis_data`：Redis持久化数据（RDB/AOF）

备份数据：
```bash
# 备份PostgreSQL
docker exec rbac-postgres pg_dump -U postgres casbin_demo > backup.sql

# 恢复PostgreSQL
cat backup.sql | docker exec -i rbac-postgres psql -U postgres casbin_demo
```

### 10.2 本地开发环境部署

本地开发适合二次开发和调试，需要分别启动后端和前端开发服务器。

#### 10.2.1 启动基础设施（PostgreSQL + Redis）

```bash
# 使用Docker启动数据库和缓存（不启动应用）
docker-compose up -d postgres redis

# 验证服务是否就绪
docker-compose ps
```

#### 10.2.2 后端开发环境

```bash
# 进入后端目录
cd backend

# 安装依赖（Go 1.26.4+）
go mod download

# 复制配置文件（如需要修改）
cp config.yaml.example config.yaml
# 编辑config.yaml配置数据库和Redis连接信息

# 运行后端（热重载可使用air或fresh）
go run cmd/server/main.go

# 或使用Makefile
make run-backend
```

后端启动成功标志：
```
time=... level=INFO msg="HTTP服务器启动" addr=:8080
time=... level=INFO msg="开始缓存预热..."
time=... level=INFO msg="角色列表预热完成" count=...
time=... level=INFO msg="权限列表预热完成" count=...
time=... level=INFO msg="缓存预热完成"
```

后端API地址：http://localhost:8080/api/v1

#### 10.2.3 前端开发环境

```bash
# 进入前端目录
cd frontend

# 安装依赖（Node.js 18+）
npm install

# 启动开发服务器（端口3000）
npm run dev

# 或使用Makefile
make run-frontend
```

前端开发服务器特性：
- 热更新（HMR）
- API代理：`/api/*`自动转发到`http://localhost:8080`
- Vite配置见[vite.config.ts](file:///d:/Programming/Agent_demo/casbin_demo/frontend/vite.config.ts)

前端访问地址：http://localhost:3000

### 10.3 Nginx配置详解

前端Nginx配置文件：[nginx.conf](file:///d:/Programming/Agent_demo/casbin_demo/frontend/nginx.conf)

配置说明：

```nginx
server {
    listen 80;
    server_name localhost;

    # 开启gzip压缩
    gzip on;
    gzip_types text/plain application/json application/javascript text/css;

    # 开启代理错误拦截（关键：后端返回的错误状态码会被Nginx处理）
    proxy_intercept_errors on;

    # API代理配置
    location /api/ {
        proxy_pass http://backend:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_connect_timeout 60s;
        proxy_read_timeout 60s;
    }

    # 错误页面配置
    error_page 401 /401;
    error_page 403 /403;
    error_page 404 /404;
    error_page 500 502 503 504 /500;

    # 错误页面处理：区分AJAX请求和浏览器直接访问
    location ~ ^/(401|403|404|500)$ {
        internal;  # 仅内部重定向可访问

        # AJAX请求（带X-Requested-With头）返回JSON
        if ($http_x_requested_with = XMLHttpRequest) {
            return 200 '{"code":$status,"message":"请求错误，请刷新重试"}';
            add_header Content-Type application/json;
        }

        # 浏览器直接访问返回前端页面，由前端路由处理显示错误页
        root /usr/share/nginx/html;
        try_files /index.html =404;
    }

    # 前端静态资源 + Vue Router history模式
    location / {
        root /usr/share/nginx/html;
        index index.html;
        try_files $uri $uri/ /index.html;  # 所有未匹配路由返回index.html
        # 静态资源缓存
        location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff2?)$ {
            expires 7d;
            add_header Cache-Control "public, immutable";
        }
    }
}
```

### 10.4 环境变量配置说明

后端配置支持通过环境变量覆盖，环境变量前缀为`CASBIN_DEMO_`，以下划线代替点：

| 环境变量 | 对应配置 | 默认值 | 说明 |
|---------|---------|-------|-----|
| CASBIN_DEMO_SERVER_MODE | server.mode | debug | 运行模式：debug/release/test |
| CASBIN_DEMO_SERVER_PORT | server.port | 8080 | HTTP监听端口 |
| CASBIN_DEMO_DATABASE_HOST | database.host | localhost | PostgreSQL主机 |
| CASBIN_DEMO_DATABASE_PORT | database.port | 5432 | PostgreSQL端口 |
| CASBIN_DEMO_DATABASE_USER | database.user | postgres | 数据库用户名 |
| CASBIN_DEMO_DATABASE_PASSWORD | database.password | - | 数据库密码 |
| CASBIN_DEMO_DATABASE_DBNAME | database.dbname | casbin_demo | 数据库名 |
| CASBIN_DEMO_DATABASE_SSLMODE | database.sslmode | disable | SSL模式 |
| CASBIN_DEMO_REDIS_HOST | redis.host | localhost | Redis主机 |
| CASBIN_DEMO_REDIS_PORT | redis.port | 6379 | Redis端口 |
| CASBIN_DEMO_REDIS_PASSWORD | redis.password | - | Redis密码 |
| CASBIN_DEMO_REDIS_DB | redis.db | 0 | Redis数据库编号 |
| CASBIN_DEMO_JWT_SECRET | jwt.secret | casbin-demo-secret | JWT签名密钥（生产环境务必修改） |
| CASBIN_DEMO_JWT_EXPIRE | jwt.expire | 24h | Token过期时间 |

### 10.5 服务健康检查

#### 10.5.1 健康检查接口

后端提供统一健康检查接口：`GET /health`（无需认证）

```bash
curl http://localhost:8080/health
# 响应：{"status":"ok"}
```

#### 10.5.2 Docker健康检查

docker-compose.yml中已配置PostgreSQL和Redis的健康检查，backend依赖它们健康后才启动。

检查容器健康状态：
```bash
docker inspect --format='{{.State.Health.Status}}' rbac-postgres
# healthy / unhealthy / starting
```

#### 10.5.3 服务自检清单

部署完成后按以下清单验证：

| 检查项 | 预期结果 | 验证命令/操作 |
|-------|---------|-------------|
| PostgreSQL可连接 | 连接成功 | `docker exec -it rbac-postgres psql -U postgres -d casbin_demo -c "\dt"` |
| Redis可连接 | PONG响应 | `docker exec -it rbac-redis redis-cli -a redis123 ping` |
| 后端服务启动 | HTTP 200 | `curl http://localhost:8080/health` |
| 前端页面可访问 | 显示登录页 | 浏览器访问http://localhost |
| API代理正常 | 返回JSON | `curl http://localhost/api/v1/health` |
| 注册/登录功能 | 获得Token | 前端注册/登录测试 |
| 数据库表自动迁移 | 表已创建 | PostgreSQL中查看users/roles/permissions/casbin_rule表 |
| 缓存预热完成 | 日志显示预热完成 | `docker-compose logs backend | grep "缓存预热"` |

---

## 第14章 开发指南

### 11.1 新增接口开发流程

新增一个API接口需要修改以下四层代码，按顺序开发：

#### 步骤1：定义Model（Entity/DTO/Form）
在`internal/model/`下新增或修改实体、DTO、表单结构体。

#### 步骤2：实现Repository
在`internal/repository/`下新增数据访问方法，封装数据库CRUD操作。

#### 步骤3：实现Service
在`internal/service/`下新增业务逻辑方法，加入缓存策略（查询缓存、空值缓存、失效策略）。

#### 步骤4：实现Handler
在`internal/handler/`下新增Handler方法，处理参数绑定和响应封装。

#### 步骤5：注册路由
在`internal/router/router.go`的`RegisterRoutes`函数中添加路由，配置正确的中间件。

#### 步骤6：注册到fx
在`internal/app/module.go`中添加新组件的Provide（如果新增了Repo/Service/Handler）。

#### 步骤7：前端API和页面
- 在`frontend/src/api/`新增接口请求
- 在`frontend/src/views/`新增页面组件
- 在`frontend/src/router/`添加路由配置

#### 步骤8：添加Casbin权限策略
启动后在角色管理中为对应角色分配新接口的权限（或在数据库中添加casbin_rule记录）。

### 11.2 fx依赖注入规范

1. **构造函数命名**：`NewXxx(依赖1, 依赖2) *Xxx`
2. **构造函数职责**：仅做初始化和依赖注入，不做业务逻辑
3. **单例组件**：所有组件在fx容器中默认单例
4. **依赖顺序**：构造函数的参数顺序就是依赖顺序，fx自动解析
5. **公共组件**：pkg下的组件优先使用`sync.OnceValue`保证单例
6. **OnStart/OnStop**：需要启停逻辑的组件使用fx.Hook注册生命周期钩子

**正确示例**：
```go
func NewUserService(repo *repository.UserRepo, cache *cache.Cache, logger *slog.Logger) *UserService {
    return &UserService{
        repo:   repo,
        cache:  cache,
        logger: logger,
    }
}
```

### 11.3 缓存使用规范

1. **查询接口必须加缓存**：根据数据类型选择缓存分类（hot/query/config/stats）
2. **写操作必须失效缓存**：使用SCAN+DEL批量删除相关缓存，禁止使用KEYS
3. **空查询必须缓存空值**：防穿透，TTL固定60秒
4. **空值标记判断**：通过ID=0或reflect.IsZero()判断是空值缓存还是真实数据
5. **不得缓存敏感数据**：密码、密钥等敏感信息禁止写入缓存
6. **认证接口不缓存**：登录、注册、获取当前用户信息等接口不使用缓存
7. **缓存key命名规范**：`cache:{分类}:{业务}:{标识}`，例如：
   - `cache:hot:user:1`（用户ID=1的热点缓存）
   - `cache:query:users:page:1:size:10`（用户分页查询缓存）
   - `cache:null:user:999`（用户ID=999的空值缓存）

### 11.4 代码规范

#### 11.4.1 Go代码规范

1. **Go 1.26+新特性使用**：
   - 日志统一使用`log/slog`结构化日志，禁止fmt.Println
   - 单例模式使用`sync.OnceValue`
   - 信号处理使用`signal.NotifyContext`
   - 切片操作使用`slices`包函数
   - 错误合并使用`errors.Join`

2. **注释规范**：
   - 所有 exported 类型、函数、方法必须有注释
   - 注释用中文，详细说明功能、参数、返回值
   - 包注释在package语句前

3. **错误处理**：
   - 所有error必须处理，不能用_忽略（除了fmt.Println等明确忽略的）
   - 错误信息首字母小写，不带标点
   - 对外暴露的错误使用errors.New或fmt.Errorf包装

4. **命名规范**：
   - 结构体、接口：大驼峰（UserService）
   - 私有字段：小驼峰（userRepo）
   - 常量：大写下划线（MaxPageSize）
   - JSON字段：小写下划线（user_name）

5. **HTTP响应规范**：
   - 必须使用pkg/response包返回响应
   - HTTP状态码必须与body.code一致
   - 禁止直接c.JSON返回未包装的响应

#### 11.4.2 前端代码规范

1. **TypeScript规范**：
   - 所有组件、函数必须定义类型
   - 禁止使用any（特殊情况需注释说明）
   - API请求和响应必须定义interface

2. **Vue 3规范**：
   - 使用Composition API和`<script setup lang="ts">`
   - ref/reactive按需使用，ref用于基本类型，reactive用于对象
   - composables函数命名以use开头（useUserStore, useRouter）

3. **样式规范**：
   - 使用SCSS，嵌套深度不超过3层
   - 全局样式在styles/index.scss中定义
   - 组件样式使用scoped

### 11.5 测试规范

1. **单元测试文件命名**：`xxx_test.go`，与被测文件同目录
2. **测试函数命名**：`TestXxx(t *testing.T)`
3. **表驱动测试**：多个测试case使用表驱动
4. **Mock依赖**：Service层测试Mock Repository，Handler层测试Mock Service
5. **运行测试**：
```bash
# 运行所有测试
go test ./...

# 运行带覆盖率的测试
go test -cover ./...

# 运行指定包测试
go test ./internal/service/... -v
```

### 11.6 常见问题与解决方案

#### Q1：启动报错"database connection failed"
**原因**：PostgreSQL连接配置错误或服务未启动。
**解决方案**：
1. 检查PostgreSQL容器是否运行：`docker-compose ps`
2. 确认config.yaml中数据库配置正确
3. 如果使用Docker部署backend，数据库host应为`postgres`而非localhost
4. 查看PostgreSQL日志：`docker-compose logs postgres`

#### Q2：启动报错"redis connection refused"
**原因**：Redis未启动或密码配置错误。
**解决方案**：
1. 检查Redis容器状态
2. 确认redis密码配置一致
3. Docker部署时Redis host应为`redis`

#### Q3：登录后访问接口返回403
**原因**：用户角色没有对应接口的权限。
**解决方案**：
1. 确认用户已分配角色
2. 确认角色已分配对应API的权限
3. 检查casbin_rule表中是否有对应的p策略记录
4. 如修改了策略，重启后端或调用e.LoadPolicy()重新加载

#### Q4：前端页面刷新404
**原因**：Vue Router使用history模式，Nginx需要配置try_files。
**解决方案**：
- 确保Nginx配置中有`try_files $uri $uri/ /index.html;`
- 开发环境Vite已配置historyApiFallback，不会出现此问题

#### Q5：缓存数据不一致
**原因**：写操作未正确失效缓存。
**解决方案**：
1. 检查写操作是否调用了invalidateXxxCache
2. 确认缓存key模式匹配正确
3. 开发环境可临时禁用缓存调试
4. 手动清除Redis：`docker exec rbac-redis redis-cli -a redis123 FLUSHDB`

#### Q6：Casbin权限修改不生效
**原因**：Casbin Enforcer缓存了策略。
**解决方案**：
1. 修改策略后调用`e.LoadPolicy()`重新加载
2. 或重启后端服务
3. 开发环境可暂时关闭Casbin缓存测试

#### Q7：跨域问题（CORS error）
**原因**：前端域名与API域名不同且未配置CORS。
**解决方案**：
- 开发环境Vite已配置代理，不会跨域
- 生产环境通过Nginx代理，API和前端同源
- 如需跨域，后端已配置CORS中间件，允许所有来源（生产环境建议限制）

---

## 第15章 核心依赖版本

### 12.1 后端依赖

[go.mod](file:///d:/Programming/Agent_demo/casbin_demo/backend/go.mod)核心依赖版本：

| 依赖包 | 版本 | 用途 |
|-------|-----|-----|
| Go | 1.26.4 | 编程语言版本 |
| github.com/gin-gonic/gin | v1.11.0 | Web框架 |
| gorm.io/gorm | v1.31.0 | ORM框架 |
| gorm.io/driver/postgres | v1.5.11 | PostgreSQL驱动 |
| github.com/casbin/casbin/v2 | v2.111.0 | 权限控制引擎 |
| github.com/casbin/gorm-adapter/v3 | v3.28.0 | Casbin GORM存储适配器 |
| go.uber.org/fx | v1.24.1 | 依赖注入框架 |
| github.com/redis/go-redis/v9 | v9.7.3 | Redis客户端 |
| github.com/golang-jwt/jwt/v5 | v5.3.0 | JWT Token处理 |
| golang.org/x/crypto | v0.41.0 | bcrypt密码加密 |

### 12.2 前端依赖

[package.json](file:///d:/Programming/Agent_demo/casbin_demo/frontend/package.json)核心依赖版本：

| 依赖包 | 版本 | 用途 |
|-------|-----|-----|
| vue | ^3.5.13 | 前端框架 |
| vue-router | ^4.6.4 | 路由管理 |
| pinia | ^3.0.3 | 状态管理 |
| element-plus | ^2.13.7 | UI组件库 |
| @element-plus/icons-vue | ^2.3.1 | 图标库 |
| axios | ^1.15.0 | HTTP客户端 |
| typescript | ~5.9.3 | TypeScript语言 |
| vite | ^7.2.4 | 构建工具 |
| sass | ^1.99.2 | CSS预处理器 |

### 12.3 基础设施依赖

| 软件 | 镜像版本 | 用途 |
|-----|---------|-----|
| PostgreSQL | postgres:15-alpine | 关系型数据库 |
| Redis | redis:7-alpine | 分布式缓存 |
| Nginx | nginx:alpine | Web服务器、反向代理（前端容器内） |
| Go (构建阶段) | golang:1.26-alpine | 后端编译阶段基础镜像 |
| Node.js (构建阶段) | node:20-alpine | 前端编译阶段基础镜像 |

---

**文档结束**