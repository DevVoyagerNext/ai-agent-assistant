# 后台管理

这个目录是独立的后台管理端，包含 Go 接口服务和 Vue 管理页面。

## 目录

- `backend`: 管理端 API，默认端口 `8081`
- `frontend`: 管理端 Vue 页面，默认端口 `5174`

## 后端

后端按前台项目的分层方式组织，使用 Gin + GORM：

- `internal/router`: 路由注册，`enter.go` 汇总，认证、个人信息、教材审批、节点审批、审批记录分文件
- `internal/controller`: HTTP 参数绑定和响应，按认证、个人信息、教材审批、节点审批、审批记录分文件
- `internal/service`: 业务逻辑，`admin_service.go` 汇总依赖，其他文件按业务模块拆分
- `internal/dao`: GORM 数据访问，按用户、统计、教材审批、节点审批、审批记录分文件
- `internal/dto`: 请求和响应结构，按 common、auth、profile、review 拆分
- `internal/model`: 数据表模型，每张核心表独立文件
- `internal/middleware`: 鉴权和跨域中间件
- `internal/config`: 配置加载

后端默认会优先读取项目现有的 `backend/settings.yaml`，复用其中的 MySQL 和 JWT 配置。也可以复制 `backend/config.example.yaml` 为 `backend/config.yaml` 后修改配置。

管理端启动时会自动检查超级管理员账号：如果不存在用户名为 `root` 的账号，会创建 `root / 123456`，角色为 `admin`。如果 `root` 已存在，只会确保它是启用的管理员账号，不会覆盖已有密码。

常用环境变量：

- `ADMIN_PORT`: 管理端 API 端口，默认 `8081`
- `ADMIN_INVITE_CODE`: 管理员注册邀请码，默认 `admin2026`
- `ADMIN_DB_DSN`: 完整 MySQL DSN，设置后会覆盖 YAML 中的 MySQL 配置

启动：

```bash
cd admin/backend
go run .
```

主要接口：

- `POST /v1/admin/register`: 注册管理员
- `POST /v1/admin/login`: 管理员登录
- `GET /v1/admin/me`: 管理员信息和审批统计
- `GET /v1/admin/subjects/pending`: 待审批教材列表
- `POST /v1/admin/subjects/:id/approve`: 审批通过教材
- `POST /v1/admin/subjects/:id/reject`: 驳回教材
- `GET /v1/admin/nodes/pending`: 待审批节点列表
- `POST /v1/admin/nodes/:id/approve`: 审批通过节点草稿
- `POST /v1/admin/nodes/:id/reject`: 驳回节点草稿
- `GET /v1/admin/audit-records`: 审批记录

## 前端

前端按前台项目的分层方式组织：

- `src/api`: 后台接口模块，已拆分为 `auth.ts` 和 `audit.ts`
- `src/utils`: Axios 请求实例和拦截器
- `src/types`: TypeScript 类型，按 common、auth、profile、audit 拆分，并由 `admin.ts` 汇总导出
- `src/router`: 路由和登录守卫，`index.ts` 汇总，模块路由放在 `modules/`
- `src/views`: 登录、注册、个人信息和审批页面

启动：

```bash
cd admin/frontend
npm install
npm run dev
```

如果管理端 API 不是默认地址，可以设置：

```bash
VITE_ADMIN_API_BASE=http://localhost:8081/v1 npm run dev
```
