<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { BadgeCheck, BookCheck, KeyRound, Loader2, Lock, Mail, PenLine, UserPlus } from 'lucide-vue-next'
import { registerAdmin } from '../api/auth'
import { saveAdminAuth } from '../utils/adminAuth'

const router = useRouter()
const loading = ref(false)
const message = ref('')

const form = reactive({
  username: '',
  email: '',
  password: '',
  inviteCode: '',
  signature: ''
})

const handleRegister = async () => {
  if (!form.username || !form.email || !form.password || !form.inviteCode) {
    message.value = '信息未填完整'
    return
  }

  loading.value = true
  message.value = ''
  try {
    const payload = await registerAdmin(form)
    saveAdminAuth(payload)
    router.push('/me')
  } catch (err) {
    message.value = err instanceof Error ? err.message : '注册失败'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <main class="auth-page">
    <section class="auth-card register-card">
      <div class="auth-brand">
        <span class="auth-mark">
          <BookCheck :size="24" />
        </span>
        <div>
          <strong>审核后台</strong>
          <small>创建管理员</small>
        </div>
      </div>

      <div class="auth-header">
        <h1>注册</h1>
        <p>邀请码验证后进入后台</p>
      </div>

      <form class="auth-form" @submit.prevent="handleRegister">
        <label>
          <span><UserPlus :size="18" />用户名</span>
          <input v-model.trim="form.username" type="text" placeholder="用户名" autocomplete="username" />
        </label>

        <label>
          <span><Mail :size="18" />邮箱</span>
          <input v-model.trim="form.email" type="email" placeholder="邮箱" autocomplete="email" />
        </label>

        <label>
          <span><Lock :size="18" />密码</span>
          <input v-model="form.password" type="password" placeholder="至少 8 位" autocomplete="new-password" />
        </label>

        <label>
          <span><KeyRound :size="18" />邀请码</span>
          <input v-model.trim="form.inviteCode" type="text" placeholder="邀请码" />
        </label>

        <label>
          <span><PenLine :size="18" />签名</span>
          <input v-model.trim="form.signature" type="text" placeholder="可选" />
        </label>

        <p v-if="message" class="form-message">{{ message }}</p>

        <button class="primary-button" type="submit" :disabled="loading">
          <Loader2 v-if="loading" class="spin" :size="18" />
          <BadgeCheck v-else :size="18" />
          <span>创建账号</span>
        </button>
      </form>

      <p class="auth-switch">
        已有账号？
        <router-link to="/login">登录</router-link>
      </p>
    </section>
  </main>
</template>
