# AI Agent Assistant

一个面向学习场景的 AI 助手项目，提供教材浏览、知识点学习、AI 对话、私人笔记、收藏夹与用户成长数据等能力。

项目采用前后端分离架构：

- `backend/`：Go + Gin + Gorm + Redis 的后端服务
- `frontend/`：Vue 3 + TypeScript + Vite 的前端应用
- `gin-vue-admin/`：仓库中额外保留的管理端/参考工程代码

## 项目简介

这个项目主要围绕“AI + 学习内容组织”展开，目标是把教材、知识点、学习笔记和 AI 助教能力整合在一起，形成一个可持续使用的个人学习助手。

当前代码中已经覆盖了以下核心能力：

- 用户注册、登录、Token 刷新
- 教材分类、教材列表、教材搜索、教材详情
- 教材创建、草稿编辑、发布、点赞、收藏、收藏夹管理
- 知识点树浏览、知识点详情、作者编辑入口
- 学习进度记录、节点状态更新、节点难度标记、随堂笔记
- 私人笔记创建、重命名、公开/私密切换、分享与访问
- AI 对话、历史会话、会话标题修改、消息记录查看
- 用户资料页、学习活跃度日历、最近学习内容汇总

## 技术栈

### 后端

- Go
- Gin
- Gorm
- MySQL
- Redis
- Viper
- Zap
- JWT
- CloudWeGo Eino / OpenAI 兼容模型接入

### 前端

- Vue 3
- TypeScript
- Vite
- Vue Router
- Pinia
- Axios
- Markdown-It
- Highlight.js
- `@antv/x6`
- `vditor`

## 主要页面

根据当前前端路由，项目包含以下页面：

- `/`：教材市场
- `/subject/:id`：教材详情页
- `/ai-chat`：AI 对话页
- `/me`：个人中心
- `/me/:type`：用户列表/内容列表页
- `/author/subject/:id`：作者教材编辑页
- `/login`、`/register`：登录注册页
- `/share/verify`、`/share/access`：笔记分享访问页

## 后端接口模块

当前后端以 `/v1` 为统一前缀，主要分为以下模块：

- `user`：注册、登录、刷新令牌、用户信息、活动统计
- `subjects`：教材列表、搜索、分类、详情
- `user/subjects`：教材创建、发布、点赞、收藏、收藏夹、学习状态
- `nodes`：知识点树、节点详情、节点编辑、学习状态、难度标记、随堂笔记
- `user/notes/private`：私人笔记、文件夹、公开设置、分享访问
- `ai`：AI 聊天、会话列表、历史消息、导出下载

## 项目结构

```text
ai-agent-assistant/
├── backend/                 # Go 后端服务
│   ├── config/              # 配置结构定义
│   ├── controller/          # 控制器层
│   ├── dao/                 # 数据访问层
│   ├── dto/                 # 请求/响应 DTO
│   ├── initialize/          # 基础组件初始化
│   ├── middleware/          # 中间件
│   ├── model/               # 数据模型
│   ├── pkg/                 # 通用工具、错误码、响应封装
│   ├── router/              # 路由注册
│   ├── service/             # 业务逻辑层
│   ├── main.go              # 后端启动入口
│   └── settings.example.yaml
├── frontend/                # Vue 3 前端应用
│   ├── src/api/             # 接口请求封装
│   ├── src/components/      # 通用组件
│   ├── src/composables/     # 组合式逻辑
│   ├── src/layouts/         # 页面布局
│   ├── src/router/          # 路由配置
│   ├── src/store/           # Pinia 状态管理
│   ├── src/types/           # 类型定义
│   ├── src/utils/           # 工具函数
│   └── src/views/           # 页面视图
├── gin-vue-admin/           # 仓库中保留的管理端/参考代码
├── docker-compose.yaml      # 预留的一键启动文件
└── README.md
```

## 环境要求

建议本地环境如下：

- Go `1.25+`
- Node.js `18+`
- npm `9+`
- MySQL `8.x`
- Redis `6.x` 或更高版本

## 后端配置

后端示例配置文件位于：

```text
backend/settings.example.yaml
```

建议先复制为实际运行配置文件：

```bash
cd backend
copy settings.example.yaml settings.yaml
```

你需要根据本地环境修改以下配置：

- `system.base-url`：后端对外访问地址
- `mysql.*`：数据库连接信息
- `redis.*`：Redis 连接信息
- `jwt.*`：JWT 密钥与过期时间
- `email.*`：邮箱服务配置
- `ai.*`：AI 模型平台地址、API Key、模型名

## 本地启动

### 1. 启动后端

```bash
cd backend
go mod tidy
go run main.go
```

默认情况下，后端运行在：

```text
http://localhost:8080
```

### 2. 启动前端

```bash
cd frontend
npm install
npm run dev
```

前端启动后，按终端输出访问本地开发地址，通常是：

```text
http://localhost:5173
```

## 开发说明

### 前端

- 采用 `Vue 3 + TypeScript + Vite`
- 接口统一放在 `src/api/`
- 页面路由位于 `src/router/index.ts`
- 主要业务页面位于 `src/views/`

### 后端

- 采用分层结构：`controller -> service -> dao`
- 配置通过 `Viper` 加载
- 日志使用 `Zap`
- 路由统一在 `router/` 下注册
- 认证通过 `JWT` 中间件处理

## 当前状态

从仓库内容来看，项目已经具备较完整的核心业务骨架，适合继续补充以下内容：

- 数据库初始化脚本与表结构说明
- 完整的接口文档
- 部署文档
- `docker-compose.yaml` 实际编排内容
- 测试用账号与演示数据说明

## 后续可补充

如果你准备继续完善这个仓库，推荐下一步补这些文档：

- `docs/api.md`：接口清单
- `docs/deploy.md`：部署流程
- `docs/database.md`：数据库表设计
- `docs/architecture.md`：系统架构图

## License

如果你后续准备开源，建议补充 `LICENSE` 文件，并在这里声明协议类型。
