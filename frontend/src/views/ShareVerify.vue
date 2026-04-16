<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getShareBasicInfo } from '../api/user'
import type { ShareBasicInfoRes } from '../types/user'
import Toast from '../components/Toast.vue'
import { Loader2, User, FileText, Folder, Lock, AlertCircle, ArrowRight } from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()
const shareToken = route.query.token as string

const loading = ref(true)
const shareInfo = ref<ShareBasicInfoRes | null>(null)
const errorMsg = ref('')
const shareCode = ref('')
const verifying = ref(false)

const toast = ref({
  show: false,
  message: '',
  type: 'error' as 'success' | 'error'
})

const showToast = (message: string, type: 'success' | 'error' = 'error') => {
  toast.value = { show: true, message, type }
}

onMounted(async () => {
  if (!shareToken) {
    errorMsg.value = '无效的分享链接：缺少 token 参数'
    loading.value = false
    return
  }

  try {
    const res = await getShareBasicInfo(shareToken)
    if (res.data?.code === 200 && res.data.data) {
      shareInfo.value = res.data.data
      if (!shareInfo.value.isActive) {
        errorMsg.value = '该分享链接已被取消'
      } else if (shareInfo.value.isExpired) {
        errorMsg.value = '该分享链接已过期'
      }
    } else {
      errorMsg.value = res.data?.msg || '获取分享信息失败'
    }
  } catch (err: any) {
    errorMsg.value = err.response?.data?.msg || '获取分享信息失败'
  } finally {
    loading.value = false
  }
})

const handleVerify = async () => {
  if (!shareCode.value || shareCode.value.length !== 4) {
    showToast('请输入4位提取码')
    return
  }

  verifying.value = true
  
  // 在此处我们直接跳转到访问页，并将 token 和 code 传递过去
  // 因为分享的根节点访问通常可以依赖 token 和 code（无需特定的 nodeId 或者传 0）
  // 具体的权限校验在 accessShareNote 接口中进行
  setTimeout(() => {
    verifying.value = false
    showToast('提取码验证成功，即将加载内容...', 'success')
    router.push({
      path: '/share/access',
      query: {
        token: shareToken,
        code: shareCode.value,
        nodeId: shareInfo.value?.noteId?.toString() || '0'
      }
    })
  }, 800)
}
</script>

<template>
  <div class="share-verify-page">
    <Toast 
      v-if="toast.show" 
      :message="toast.message" 
      :type="toast.type" 
      @close="toast.show = false" 
    />

    <div class="verify-container">
      <!-- Loading State -->
      <div v-if="loading" class="loading-state">
        <Loader2 class="spin" :size="48" />
        <p>正在获取分享信息...</p>
      </div>

      <!-- Error State (Invalid/Expired/Cancelled) -->
      <div v-else-if="errorMsg" class="error-state">
        <AlertCircle :size="64" class="icon-error" />
        <h2>分享不可用</h2>
        <p>{{ errorMsg }}</p>
        <button class="back-btn" @click="router.push('/')">返回首页</button>
      </div>

      <!-- Normal State (Need Code) -->
      <div v-else-if="shareInfo" class="verify-card">
        <div class="author-info">
          <div class="avatar">
            <img v-if="shareInfo.authorAvatar" :src="shareInfo.authorAvatar" alt="avatar" />
            <div v-else class="avatar-placeholder">
              <User :size="24" />
            </div>
          </div>
          <div class="author-details">
            <h3>{{ shareInfo.authorName }}</h3>
            <p>给你分享了私人笔记</p>
          </div>
        </div>

        <div class="note-info">
          <div class="note-icon">
            <Folder v-if="shareInfo.noteType === 'folder'" :size="32" class="icon-pink" />
            <FileText v-else :size="32" class="icon-blue" />
          </div>
          <h2 class="note-title">{{ shareInfo.noteTitle }}</h2>
        </div>

        <div class="verify-form">
          <div class="input-wrapper">
            <Lock :size="20" class="input-icon" />
            <input 
              v-model="shareCode" 
              type="text" 
              placeholder="请输入提取码" 
              maxlength="4"
              @keyup.enter="handleVerify"
            />
          </div>
          <button class="submit-btn" :disabled="verifying || shareCode.length !== 4" @click="handleVerify">
            <Loader2 v-if="verifying" class="spin" :size="20" />
            <span v-else>提取文件</span>
            <ArrowRight v-if="!verifying" :size="20" />
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
:root {
  --notion-black: rgba(0,0,0,0.95);
  --notion-blue: #0075de;
  --warm-gray-500: #615d59;
  --warm-gray-300: #a39e98;
}

.share-verify-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #f6f5f4;
  font-family: "NotionInter", Inter, -apple-system, system-ui, sans-serif;
  padding: 20px;
}

.verify-container {
  width: 100%;
  max-width: 420px;
}

.loading-state, .error-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 48px 24px;
  background: white;
  border-radius: 16px;
  box-shadow: 0 4px 24px rgba(0,0,0,0.06);
}

.loading-state p {
  margin-top: 16px;
  color: var(--warm-gray-500);
  font-size: 16px;
}

.icon-error {
  color: #ef4444;
  margin-bottom: 24px;
}

.error-state h2 {
  font-size: 24px;
  color: var(--notion-black);
  margin-bottom: 12px;
}

.error-state p {
  color: var(--warm-gray-500);
  margin-bottom: 32px;
}

.back-btn {
  padding: 10px 24px;
  border-radius: 8px;
  border: 1px solid rgba(0,0,0,0.1);
  background: white;
  color: var(--notion-black);
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.back-btn:hover {
  background: #f9f9f9;
}

.verify-card {
  background: white;
  border-radius: 16px;
  padding: 32px;
  box-shadow: 0 4px 24px rgba(0,0,0,0.06);
}

.author-info {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 32px;
  padding-bottom: 24px;
  border-bottom: 1px solid rgba(0,0,0,0.08);
}

.avatar {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  overflow: hidden;
  background: #f0f0f0;
}

.avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.avatar-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--warm-gray-300);
}

.author-details h3 {
  font-size: 16px;
  font-weight: 600;
  color: var(--notion-black);
  margin-bottom: 4px;
}

.author-details p {
  font-size: 14px;
  color: var(--warm-gray-500);
}

.note-info {
  text-align: center;
  margin-bottom: 40px;
}

.note-icon {
  margin-bottom: 16px;
  display: flex;
  justify-content: center;
}

.icon-pink { color: #ec4899; }
.icon-blue { color: #3b82f6; }

.note-title {
  font-size: 20px;
  font-weight: 600;
  color: var(--notion-black);
  line-height: 1.4;
  word-break: break-all;
}

.verify-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.input-icon {
  position: absolute;
  left: 16px;
  color: var(--warm-gray-300);
}

.input-wrapper input {
  width: 100%;
  padding: 16px 16px 16px 48px;
  border: 1px solid rgba(0,0,0,0.1);
  border-radius: 12px;
  font-size: 16px;
  outline: none;
  transition: all 0.2s;
  background: #fafafa;
}

.input-wrapper input:focus {
  background: white;
  border-color: #8b5cf6;
  box-shadow: 0 0 0 3px rgba(139, 92, 246, 0.1);
}

.submit-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  width: 100%;
  padding: 16px;
  border-radius: 12px;
  border: none;
  background: #8b5cf6;
  color: white;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.submit-btn:hover:not(:disabled) {
  background: #7c3aed;
  transform: translateY(-1px);
}

.submit-btn:disabled {
  background: #c4b5fd;
  cursor: not-allowed;
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
