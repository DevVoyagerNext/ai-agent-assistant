<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { BookCheck, Lock, Loader2, LogIn, Mail } from 'lucide-vue-next'
import { loginAdmin } from '../api/auth'
import { saveAdminAuth } from '../utils/adminAuth'

const router = useRouter()
const loading = ref(false)
const message = ref('')

const form = reactive({
  account: '',
  password: ''
})

const handleLogin = async () => {
  if (!form.account || !form.password) {
    message.value = '账号和密码不能为空'
    return
  }

  loading.value = true
  message.value = ''
  try {
    const payload = await loginAdmin(form)
    saveAdminAuth(payload)
    router.push('/me')
  } catch (err) {
    message.value = err instanceof Error ? err.message : '登录失败'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <main class="auth-page">
    <section class="auth-card">
      <div class="auth-brand">
        <span class="auth-mark">
          <BookCheck :size="24" />
        </span>
        <div>
          <strong>审核后台</strong>
          <small>管理员入口</small>
        </div>
      </div>

      <div class="auth-header">
        <h1>登录</h1>
        <p>使用管理员账号进入</p>
      </div>

      <form class="auth-form" @submit.prevent="handleLogin">
        <label>
          <span><Mail :size="18" />账号</span>
          <input v-model.trim="form.account" type="text" placeholder="管理员账号" autocomplete="username" />
        </label>

        <label>
          <span><Lock :size="18" />密码</span>
          <input v-model="form.password" type="password" placeholder="密码" autocomplete="current-password" />
        </label>

        <p v-if="message" class="form-message">{{ message }}</p>

        <button class="primary-button" type="submit" :disabled="loading">
          <Loader2 v-if="loading" class="spin" :size="18" />
          <LogIn v-else :size="18" />
          <span>进入后台</span>
        </button>
      </form>

      <p class="auth-switch">
        没有账号？
        <router-link to="/register">注册</router-link>
      </p>
    </section>
  </main>
</template>
