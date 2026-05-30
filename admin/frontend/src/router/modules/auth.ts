import type { RouteRecordRaw } from 'vue-router'
import AdminLogin from '../../views/AdminLogin.vue'
import AdminRegister from '../../views/AdminRegister.vue'

export const authRoutes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'AdminLogin',
    component: AdminLogin,
    meta: { guest: true }
  },
  {
    path: '/register',
    name: 'AdminRegister',
    component: AdminRegister,
    meta: { guest: true }
  }
]
