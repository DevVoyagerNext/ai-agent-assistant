import { createRouter, createWebHistory } from 'vue-router'
import { authRoutes } from './modules/auth'
import { profileRoutes } from './modules/profile'
import { setupRouterGuard } from './guard'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      redirect: '/me'
    },
    ...authRoutes,
    ...profileRoutes
  ]
})

setupRouterGuard(router)

export default router
