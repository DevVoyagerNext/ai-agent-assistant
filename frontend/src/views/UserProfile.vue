<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { getUserInfo } from '../api/user'
import type { UserInfoRes } from '../types/user'
import Toast from '../components/Toast.vue'
import { ArrowLeft, LogOut, Mail, User, Quote, RefreshCcw } from 'lucide-vue-next'

const router = useRouter()
const loading = ref(false)
const userInfo = ref<UserInfoRes | null>(null)

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

const maskedEmail = computed(() => {
  const email = userInfo.value?.email || ''
  const match = email.match(/^(.{2}).*(.{2})@(.+)$/)
  if (!match) return email
  return `${match[1]}****${match[2]}@${match[3]}`
})

const avatarText = computed(() => {
  const name = userInfo.value?.username?.trim() || ''
  return name ? name.slice(0, 1).toUpperCase() : 'U'
})

const fetchUserInfo = async () => {
  loading.value = true
  try {
    const response = await getUserInfo()
    const payload = response.data
    if (payload?.code === 200 && payload.data) {
      userInfo.value = payload.data
    } else {
      showToast(payload?.msg || '获取用户信息失败', 'error')
    }
  } catch (err: any) {
    const code = err?.response?.data?.code
    if (code === 1003) showToast('Token 不存在，请先登录', 'error')
    else if (code === 1004) showToast('Token 无效，请重新登录', 'error')
    else if (code === 1005) showToast('登录已过期，请重新登录', 'error')
    else if (code === 1006) showToast('权限不足', 'error')
    else if (code === 1007) showToast('用户已被禁用', 'error')
    else if (code === 1001) showToast('用户不存在', 'error')
    else showToast(err?.response?.data?.msg || '服务器内部错误', 'error')
  } finally {
    loading.value = false
  }
}

const handleLogout = () => {
  localStorage.removeItem('token')
  localStorage.removeItem('refreshToken')
  localStorage.removeItem('expiresAt')
  localStorage.removeItem('user')
  router.push('/login')
}

onMounted(() => {
  fetchUserInfo()
})
</script>

<template>
  <div class="profile-page">
    <Toast
      v-if="toast.show"
      :message="toast.message"
      :type="toast.type"
      @close="toast.show = false"
    />

    <div class="topbar">
      <button class="ghost-btn" @click="router.push('/')">
        <ArrowLeft :size="18" />
        返回
      </button>

      <div class="topbar-actions">
        <button class="ghost-btn" :disabled="loading" @click="fetchUserInfo">
          <RefreshCcw :size="18" />
          刷新
        </button>
        <button class="danger-btn" @click="handleLogout">
          <LogOut :size="18" />
          退出登录
        </button>
      </div>
    </div>

    <div class="content">
      <div class="card">
        <div class="hero">
          <div class="avatar">
            <img v-if="userInfo?.avatarUrl" :src="userInfo.avatarUrl" alt="avatar" />
            <span v-else>{{ avatarText }}</span>
          </div>

          <div class="title">
            <div class="name-row">
              <h1 class="username">{{ userInfo?.username || '—' }}</h1>
              <span class="badge" :class="{ loading: loading }">
                {{ loading ? '加载中' : '已登录' }}
              </span>
            </div>
            <p class="subtitle">你的学习之旅，从这里继续。</p>
          </div>
        </div>

        <div class="grid">
          <div class="field">
            <div class="label">
              <User :size="16" />
              用户名
            </div>
            <div class="value">{{ userInfo?.username || '—' }}</div>
          </div>

          <div class="field">
            <div class="label">
              <Mail :size="16" />
              邮箱
            </div>
            <div class="value">{{ userInfo?.email ? maskedEmail : '—' }}</div>
          </div>

          <div class="field span-2">
            <div class="label">
              <Quote :size="16" />
              个性签名
            </div>
            <div class="value signature">
              {{ userInfo?.signature || '还没有签名，写一句话介绍自己吧。' }}
            </div>
          </div>
        </div>

        <div class="tips">
          <div class="tip">
            你的登录凭证会通过请求头 x-token 发送到后端（JWT 中间件校验）。
          </div>
          <div class="tip">
            如果提示 Token 过期/无效，点击“退出登录”后重新登录即可。
          </div>
        </div>
      </div>

      <div class="side-card">
        <div class="side-title">快捷入口</div>
        <button class="primary-card" @click="router.push('/')">
          回到知识图谱
        </button>
        <button class="secondary-card" @click="router.push('/study/1')">
          继续学习（示例）
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.profile-page {
  width: 100vw;
  height: 100vh;
  background:
    radial-gradient(1200px 600px at 20% 10%, rgba(59, 130, 246, 0.18), transparent 60%),
    radial-gradient(900px 500px at 80% 30%, rgba(34, 197, 94, 0.16), transparent 55%),
    linear-gradient(180deg, #f8fafc, #eef2ff);
  overflow: hidden;
}

.topbar {
  height: 64px;
  padding: 0 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.topbar-actions {
  display: flex;
  gap: 10px;
}

.ghost-btn,
.danger-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  border-radius: 12px;
  border: 1px solid rgba(148, 163, 184, 0.35);
  background: rgba(255, 255, 255, 0.55);
  backdrop-filter: blur(10px);
  color: #0f172a;
  cursor: pointer;
  transition: transform 0.12s ease, background 0.12s ease, border-color 0.12s ease;
}

.ghost-btn:hover:not(:disabled) {
  transform: translateY(-1px);
  background: rgba(255, 255, 255, 0.75);
  border-color: rgba(99, 102, 241, 0.35);
}

.ghost-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.danger-btn {
  border-color: rgba(248, 113, 113, 0.35);
  color: #7f1d1d;
}

.danger-btn:hover {
  transform: translateY(-1px);
  background: rgba(254, 226, 226, 0.55);
  border-color: rgba(248, 113, 113, 0.55);
}

.content {
  height: calc(100vh - 64px);
  padding: 18px 20px 24px;
  display: grid;
  grid-template-columns: 1fr 320px;
  gap: 16px;
  align-items: start;
}

.card {
  border-radius: 20px;
  border: 1px solid rgba(148, 163, 184, 0.28);
  background: rgba(255, 255, 255, 0.68);
  backdrop-filter: blur(12px);
  box-shadow: 0 18px 55px rgba(15, 23, 42, 0.08);
  padding: 22px;
  overflow: hidden;
}

.hero {
  display: flex;
  gap: 16px;
  align-items: center;
  padding-bottom: 18px;
  border-bottom: 1px dashed rgba(148, 163, 184, 0.35);
}

.avatar {
  width: 76px;
  height: 76px;
  border-radius: 18px;
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.18), rgba(34, 197, 94, 0.14));
  border: 1px solid rgba(148, 163, 184, 0.22);
  display: grid;
  place-items: center;
  overflow: hidden;
  color: #1e293b;
  font-weight: 800;
  font-size: 26px;
}

.avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.title {
  flex: 1;
}

.name-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.username {
  margin: 0;
  font-size: 24px;
  color: #0f172a;
  letter-spacing: 0.2px;
}

.badge {
  font-size: 12px;
  padding: 4px 10px;
  border-radius: 999px;
  background: rgba(34, 197, 94, 0.12);
  color: #166534;
  border: 1px solid rgba(34, 197, 94, 0.22);
}

.badge.loading {
  background: rgba(59, 130, 246, 0.12);
  color: #1d4ed8;
  border-color: rgba(59, 130, 246, 0.22);
}

.subtitle {
  margin: 8px 0 0;
  color: #64748b;
  font-size: 13px;
}

.grid {
  padding-top: 18px;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
}

.field {
  padding: 14px 14px 12px;
  border-radius: 16px;
  border: 1px solid rgba(148, 163, 184, 0.25);
  background: rgba(255, 255, 255, 0.6);
}

.field.span-2 {
  grid-column: span 2;
}

.label {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: #475569;
  font-size: 12px;
  font-weight: 600;
}

.value {
  margin-top: 8px;
  color: #0f172a;
  font-size: 14px;
  font-weight: 600;
}

.signature {
  font-weight: 500;
  color: #334155;
  line-height: 1.7;
}

.tips {
  margin-top: 16px;
  display: grid;
  gap: 10px;
}

.tip {
  padding: 12px 14px;
  border-radius: 14px;
  background: rgba(99, 102, 241, 0.07);
  border: 1px solid rgba(99, 102, 241, 0.14);
  color: #334155;
  font-size: 12px;
  line-height: 1.6;
}

.side-card {
  border-radius: 20px;
  border: 1px solid rgba(148, 163, 184, 0.28);
  background: rgba(255, 255, 255, 0.65);
  backdrop-filter: blur(12px);
  box-shadow: 0 18px 55px rgba(15, 23, 42, 0.07);
  padding: 18px;
  display: grid;
  gap: 12px;
}

.side-title {
  font-weight: 800;
  color: #0f172a;
  letter-spacing: 0.2px;
}

.primary-card,
.secondary-card {
  width: 100%;
  padding: 12px 14px;
  border-radius: 16px;
  border: 1px solid rgba(148, 163, 184, 0.25);
  cursor: pointer;
  transition: transform 0.12s ease, background 0.12s ease, border-color 0.12s ease;
  text-align: left;
  font-weight: 700;
}

.primary-card {
  background: rgba(59, 130, 246, 0.12);
  border-color: rgba(59, 130, 246, 0.22);
  color: #1d4ed8;
}

.secondary-card {
  background: rgba(34, 197, 94, 0.1);
  border-color: rgba(34, 197, 94, 0.2);
  color: #166534;
}

.primary-card:hover,
.secondary-card:hover {
  transform: translateY(-1px);
  background: rgba(255, 255, 255, 0.75);
  border-color: rgba(99, 102, 241, 0.25);
}

@media (max-width: 960px) {
  .content {
    grid-template-columns: 1fr;
  }
}
</style>
