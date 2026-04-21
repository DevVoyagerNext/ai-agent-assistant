import { createRouter, createWebHistory } from 'vue-router'
import Register from '../views/Register.vue'
import Login from '../views/Login.vue'

const routes = [
  {
    path: '/',
    component: () => import('../layouts/DefaultLayout.vue'),
    children: [
      {
        path: '',
        name: 'SubjectMarket',
        component: () => import('../views/SubjectMarket.vue')
      },
      {
        path: 'subject/:id',
        name: 'SubjectDetail',
        component: () => import('../views/SubjectDetail.vue')
      },
      {
        path: 'me',
        name: 'UserProfile',
        component: () => import('../views/UserProfile.vue')
      },
      {
        path: 'me/:type',
        name: 'UserListPage',
        component: () => import('../views/UserListPage.vue')
      }
    ]
  },
  {
    path: '/ai-chat',
    name: 'AIChat',
    component: () => import('../views/AIChat.vue')
  },
  {
    path: '/author/subject/:id',
    name: 'AuthorSubject',
    component: () => import('../views/AuthorSubject.vue')
  },
  {
    path: '/auth',
    component: () => import('../layouts/AuthLayout.vue'),
    children: [
      {
        path: '/register',
        name: 'Register',
        component: Register
      },
      {
        path: '/login',
        name: 'Login',
        component: Login
      }
    ]
  },
  {
    path: '/share/verify',
    name: 'ShareVerify',
    component: () => import('../views/ShareVerify.vue')
  },
  {
    path: '/share/access',
    name: 'ShareAccess',
    component: () => import('../views/ShareAccess.vue')
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to) => {
  const protectedRoutes = ['UserProfile']
  const isProtected = protectedRoutes.includes(to.name as string)
  if (!isProtected) return true

  const token = localStorage.getItem('token')
  if (!token) {
    return { name: 'Login' }
  }
  return true
})

export default router
