import { createRouter, createWebHistory } from 'vue-router'
import KnowledgeGraph from '../views/KnowledgeGraph.vue'
import Register from '../views/Register.vue'
import Login from '../views/Login.vue'

const routes = [
  {
    path: '/',
    name: 'KnowledgeGraph',
    component: KnowledgeGraph
  },
  {
    path: '/register',
    name: 'Register',
    component: Register
  },
  {
    path: '/login',
    name: 'Login',
    component: Login
  },
  {
    path: '/study/:node_id',
    name: 'Study',
    component: () => import('../views/StudyRoom.vue')
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
