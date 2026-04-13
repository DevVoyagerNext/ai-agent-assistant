<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { register, sendCode as sendCodeApi } from '../api/auth'
import { validateEmail, validatePassword, validateUsername, validateCode } from '../utils/validate'
import Toast from '../components/Toast.vue'
import { User, Mail, Lock, ShieldCheck, PenTool, Loader2 } from 'lucide-vue-next'

const router = useRouter()
const loading = ref(false)
const errorMsg = ref('')
const sendLoading = ref(false)
const countdown = ref(0)
const timer = ref<number | null>(null)
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
  username: '',
  email: '',
  password: '',
  code: '',
  signature: ''
})

const validate = () => {
  const userErr = validateUsername(form.username)
  if (userErr) return userErr

  const emailErr = validateEmail(form.email)
  if (emailErr) return emailErr

  const passErr = validatePassword(form.password)
  if (passErr) return passErr

  const codeErr = validateCode(form.code)
  if (codeErr) return codeErr

  // Signature: 不超过30个字
  if (form.signature.length > 30) {
    return '个性签名不能超过30个字'
  }

  return null
}

const startCountdown = () => {
  countdown.value = 60
  timer.value = window.setInterval(() => {
    if (countdown.value > 0) {
      countdown.value--
    } else {
      if (timer.value) {
        clearInterval(timer.value)
      }
    }
  }, 1000)
}

const handleSendCode = async () => {
  const emailErr = validateEmail(form.email)
  if (emailErr) {
    showToast(emailErr, 'error')
    return
  }

  sendLoading.value = true
  try {
    const response = await sendCodeApi({ email: form.email })
    const payload = response.data
    if (payload?.code === 200) {
      showToast('验证码发送成功！')
      startCountdown()
    } else {
      showToast(payload?.msg || '发送验证码失败', 'error')
    }
  } catch (err: any) {
    showToast(err.response?.data?.msg || '服务器连接失败', 'error')
  } finally {
    sendLoading.value = false
  }
}

const handleRegister = async () => {
  const err = validate()
  if (err) {
    showToast(err, 'error')
    return
  }

  loading.value = true
  errorMsg.value = ''

  try {
    const response = await register(form)
    const payload = response.data
    if (payload?.code === 200 && payload.data) {
      const { token, refreshToken, expiresAt, userId, user } = payload.data
      localStorage.setItem('token', token)
      localStorage.setItem('refreshToken', refreshToken)
      localStorage.setItem('expiresAt', String(expiresAt))
      localStorage.setItem('userId', String(userId))
      localStorage.setItem('user', JSON.stringify(user))
      showToast('注册成功！正在跳转...')
      setTimeout(() => router.push('/me'), 1500)
    } else {
      showToast(payload?.msg || '注册失败', 'error')
    }
  } catch (err: any) {
    showToast(err.response?.data?.msg || '服务器连接失败', 'error')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="register-container">
    <Toast 
      v-if="toast.show" 
      :message="toast.message" 
      :type="toast.type" 
      @close="toast.show = false" 
    />
    
    <div class="register-box">
      <div class="register-header">
        <h1>创建账号</h1>
        <p>开启您的 AI 学习之旅</p>
      </div>

      <form @submit.prevent="handleRegister" class="register-form">
        <div class="form-group">
          <label><User :size="18" /> 用户名</label>
          <input v-model="form.username" type="text" placeholder="1-10位中英文或下划线" required />
        </div>

        <div class="form-group">
          <label><Mail :size="18" /> QQ 邮箱</label>
          <input v-model="form.email" type="email" placeholder="例如：12345@qq.com" required />
        </div>

        <div class="form-group">
          <label><Lock :size="18" /> 密码</label>
          <input v-model="form.password" type="password" placeholder="8-20位，含多种字符" required />
        </div>

        <div class="form-group">
          <label><ShieldCheck :size="18" /> 验证码</label>
          <div class="code-input-group">
            <input v-model="form.code" type="text" placeholder="4位验证码" required />
            <button 
              type="button" 
              class="send-code-btn" 
              :disabled="sendLoading || countdown > 0"
              @click="handleSendCode"
            >
              <Loader2 v-if="sendLoading" class="spin" :size="16" />
              <span v-else>{{ countdown > 0 ? `${countdown}s` : '发送验证码' }}</span>
            </button>
          </div>
        </div>

        <div class="form-group">
          <label><PenTool :size="18" /> 个性签名 (选填)</label>
          <input v-model="form.signature" type="text" placeholder="不超过30个字" />
        </div>

        <div v-if="errorMsg" class="error-tip">{{ errorMsg }}</div>

        <button type="submit" :disabled="loading" class="submit-btn">
          <Loader2 v-if="loading" class="spin" :size="20" />
          <span v-else>立即注册</span>
        </button>
      </form>

      <div class="register-footer">
        <p>已有账号？ <router-link to="/login">去登录</router-link></p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.register-container {
  width: 100vw;
  height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
}

.register-box {
  width: 400px;
  background: rgba(255, 255, 255, 0.7);
  padding: 40px;
  border-radius: 16px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
  z-index: 10;
}

.register-header {
  text-align: center;
  margin-bottom: 30px;
}

.register-header h1 {
  font-size: 24px;
  color: #1e293b;
  margin-bottom: 8px;
}

.register-header p {
  font-size: 14px;
  color: #64748b;
}

.register-form {
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
  padding: 10px 12px;
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

.code-input-group {
  display: flex;
  gap: 10px;
}

.code-input-group input {
  flex: 1;
}

.send-code-btn {
  padding: 0 15px;
  background: rgba(255, 255, 255, 0.4);
  border: none;
  border-radius: 8px;
  color: #3b82f6;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.2s;
  min-width: 90px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.send-code-btn:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.6);
}

.send-code-btn:disabled {
  color: #94a3b8;
  cursor: not-allowed;
  opacity: 0.8;
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

.register-footer {
  margin-top: 24px;
  text-align: center;
  font-size: 14px;
  color: #64748b;
}

.register-footer a {
  color: #3b82f6;
  text-decoration: none;
  font-weight: 500;
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
