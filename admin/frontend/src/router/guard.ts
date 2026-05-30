import type { Router } from 'vue-router'
import { isAdminSessionValid } from '../utils/adminAuth'

export const setupRouterGuard = (router: Router) => {
  router.beforeEach(to => {
    const hasSession = isAdminSessionValid()
    if (to.meta.requiresAuth && !hasSession) {
      return { name: 'AdminLogin' }
    }
    if (to.meta.guest && hasSession) {
      return { name: 'AdminProfile' }
    }
    return true
  })
}
