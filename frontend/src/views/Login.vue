<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { login } from '../api/auth'
import { validateEmail, validatePassword } from '../utils/validate'
import Toast from '../components/Toast.vue'
import { Mail, Lock, Loader2 } from 'lucide-vue-next'

const router = useRouter()
const loading = ref(false)
const errorMsg = ref('')
const toast = reactive({
  show: false,
  message: '',
  type: 'success' as 'success' | 'error'
})

const showToast = (message: string, type: 'success' | 'error' = 'success') => {
  toast.message = message
  toast.type = type
  toast.show = true
}

const form = reactive({
  email: '',
  password: ''
})

const validate = () => {
  const emailErr = validateEmail(form.email)
  if (emailErr) return emailErr

  const passErr = validatePassword(form.password)
  if (passErr) return passErr

  return null
}

const handleLogin = async () => {
  const err = validate()
  if (err) {
    showToast(err, 'error')
    return
  }

  loading.value = true
  errorMsg.value = ''

  try {
    const response = await login(form)
    const payload = response.data
    if (payload?.code === 200 && payload.data) {
      const { token, refreshToken, expiresAt, user } = payload.data
      localStorage.setItem('token', token)
      localStorage.setItem('refreshToken', refreshToken)
      localStorage.setItem('expiresAt', String(expiresAt))
      localStorage.setItem('user', JSON.stringify(user))
      showToast('登录成功！')
      setTimeout(() => router.push('/me'), 1500)
    } else {
      showToast(payload?.msg || '登录失败', 'error')
    }
  } catch (err: any) {
    showToast(err.response?.data?.msg || '服务器连接失败', 'error')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-container">
    <Toast 
      v-if="toast.show" 
      :message="toast.message" 
      :type="toast.type" 
      @close="toast.show = false" 
    />
    
    <div class="login-box">
      <div class="login-header">
        <h1>欢迎回来</h1>
        <p>登录以继续您的 AI 学习之旅</p>
      </div>

      <form @submit.prevent="handleLogin" class="login-form">
        <div class="form-group">
          <label><Mail :size="18" /> QQ 邮箱</label>
          <input v-model="form.email" type="email" placeholder="例如：12345@qq.com" required />
        </div>

        <div class="form-group">
          <label><Lock :size="18" /> 密码</label>
          <input v-model="form.password" type="password" placeholder="请输入您的密码" required />
        </div>

        <div v-if="errorMsg" class="error-tip">{{ errorMsg }}</div>

        <button type="submit" :disabled="loading" class="submit-btn">
          <Loader2 v-if="loading" class="spin" :size="20" />
          <span v-else>立即登录</span>
        </button>
      </form>

      <div class="login-footer">
        <p>还没有账号？ <router-link to="/register">立即注册</router-link></p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-container {
  width: 100vw;
  height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
}

.login-box {
  width: 400px;
  background: rgba(255, 255, 255, 0.7);
  padding: 40px;
  border-radius: 16px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
  z-index: 10;
}

.login-header {
  text-align: center;
  margin-bottom: 30px;
}

.login-header h1 {
  font-size: 24px;
  color: #1e293b;
  margin-bottom: 8px;
}

.login-header p {
  font-size: 14px;
  color: #64748b;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-group label {
  font-size: 14px;
  color: #475569;
  display: flex;
  align-items: center;
  gap: 6px;
  font-weight: 500;
}

.form-group input {
  padding: 12px;
  border: none;
  border-radius: 8px;
  outline: none;
  font-size: 14px;
  transition: all 0.2s;
  background: rgba(255, 255, 255, 0.5);
}

.form-group input:focus {
  background: rgba(255, 255, 255, 0.8);
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.2);
}

.submit-btn {
  margin-top: 10px;
  padding: 12px;
  background: #3b82f6;
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 8px;
  transition: background 0.2s;
}

.submit-btn:hover:not(:disabled) {
  background: #2563eb;
}

.submit-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.error-tip {
  color: #ef4444;
  font-size: 12px;
  text-align: center;
}

.login-footer {
  margin-top: 24px;
  text-align: center;
  font-size: 14px;
  color: #64748b;
}

.login-footer a {
  color: #3b82f6;
  text-decoration: none;
  font-weight: 500;
}

.spin {
  animation: spin 1s linear infinite;
}

.inline-icon {
  margin-left: 4px;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
