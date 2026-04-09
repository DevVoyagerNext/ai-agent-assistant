import { createRouter, createWebHistory } from 'vue-router'
import KnowledgeGraph from '../views/KnowledgeGraph.vue'
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
        path: 'graph',
        name: 'KnowledgeGraph',
        component: KnowledgeGraph
      },
      {
        path: 'study/:node_id',
        name: 'Study',
        component: () => import('../views/StudyRoom.vue')
      },
      {
        path: 'me',
        name: 'UserProfile',
        component: () => import('../views/UserProfile.vue')
      }
    ]
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
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to) => {
  const protectedRoutes = ['UserProfile', 'KnowledgeGraph', 'Study']
  const isProtected = protectedRoutes.includes(to.name as string)
  if (!isProtected) return true

  const token = localStorage.getItem('token')
  if (!token) {
    return { name: 'Login' }
  }
  return true
})

export default router
