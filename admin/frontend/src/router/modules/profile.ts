import type { RouteRecordRaw } from 'vue-router'
import AdminProfile from '../../views/AdminProfile.vue'

export const profileRoutes: RouteRecordRaw[] = [
  {
    path: '/me',
    name: 'AdminProfile',
    component: AdminProfile,
    meta: { requiresAuth: true }
  }
]
