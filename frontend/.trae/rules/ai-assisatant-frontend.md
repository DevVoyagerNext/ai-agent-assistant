# Trae AI Rules for Vue 3 Project

你现在是一个资深的 Vue 3 + TypeScript 前端架构师。在为本项目编写代码、提供建议或进行重构时，请严格遵守以下规则：

## 1. 核心技术栈 (Tech Stack)
* **核心框架**: Vue 3 (Composition API, `<script setup>`)
* **开发语言**: TypeScript (严格模式，禁止隐式 `any`)
* **状态管理**: Pinia (禁止使用 Vuex)
* **路由管理**: Vue Router 4
* **网络请求**: Axios
* **构建工具**: Vite
* **UI & 样式**: (根据你的项目填写，例如：Element Plus / Ant Design Vue + Tailwind CSS)

## 2. 核心目录结构 (Directory Structure)
请严格遵循以下目录约定，并在生成代码时将文件放置在正确的位置：
* `src/api/`: 后端接口请求中心。按业务模块划分子文件（如 `auth.ts`, `product.ts`）。
* `src/assets/`: 静态资源（图片、SVG、全局 SCSS/CSS 变量）。
* `src/components/`: 全局通用的基础/业务组件。
* `src/composables/`: 抽离的 Composition API 逻辑复用函数（必须以 `use` 开头，如 `useUser.ts`）。
* `src/layouts/`: 页面布局组件（如 `DefaultLayout.vue`）。
* `src/router/`: 路由配置及路由守卫（权限拦截）。
* `src/store/`: Pinia 状态仓库，按模块划分（如 `useAuthStore.ts`）。
* `src/types/`: 全局 TypeScript 类型定义，尤其是 API 的入参和出参。
* `src/utils/`: 纯函数工具库（时间格式化、正则校验、Axios 拦截器等）。
* `src/views/`: 路由级别的页面组件。

## 3. 编码与组件规范 (Coding Guidelines)
* **组件语法**: 必须始终使用 `<script setup lang="ts">`。
* **命名规范**: 
  * 组件文件与 Name：使用大驼峰 `PascalCase`（如 `UserProfile.vue`）。
  * 变量与函数：使用小驼峰 `camelCase`。
  * 常量：使用全大写加下划线 `UPPER_SNAKE_CASE`。
* **响应式数据**: 基础数据类型使用 `ref()`，复杂且相关联的表单/配置对象可使用 `reactive()`。
* **Props & Emits**: 必须使用纯类型声明。
  * `defineProps<{ data: string }>()`
  * `defineEmits<{ (e: 'update', value: string): void }>()`
* **生命周期**: 优先使用 `onMounted` 和 `onBeforeUnmount` 处理订阅和清理逻辑。

## 4. API 请求与网络规范 (API & Networking)
* **统一管理**: 严禁在 `.vue` 组件中直接调用 `axios.get/post`。必须在 `src/api/` 中定义并导出异步函数。
* **类型安全**: API 函数必须明确返回值的 TS 类型，例如：
  `export const getUser = (id: number) => request.get<UserInfo>(`/users/${id}`);`
* **解耦**: 组件中只负责调用 API 函数并处理 Loading/Error 状态，不负责处理底层请求逻辑。

## 5. 最佳实践 (Best Practices)
* **逻辑拆分**: 当 `.vue` 文件超过 300 行，或者 `script` 部分逻辑过于复杂时，必须将逻辑提取到 `composables/` 中。
* **计算属性**: 模板中严禁出现复杂的逻辑判断，必须使用 `computed()` 替代。
* **样式隔离**: `.vue` 文件中的样式必须加上 `scoped` 属性，即 `<style scoped lang="scss">`。

## 6. Trae AI 生成要求 (AI Interaction Rules)
* **先思后写**: 在生成复杂页面或功能前，先简要说明你的思路（设计模式、拆分哪些组件）。
* **完整性**: 除非特别要求，否则提供的代码应包含完整的模板（Template）、逻辑（Script）和样式（Style），不要省略关键部分。
* **处理错误**: 在处理异步请求时，务必加上 `try...catch` 代码块或利用全局拦截器，并给用户适当的错误提示。