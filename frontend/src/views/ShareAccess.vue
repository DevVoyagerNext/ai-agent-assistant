<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { accessShareNote } from '../api/user'
import type { ShareAccessRes } from '../types/user'
import Toast from '../components/Toast.vue'
import { Loader2, ArrowLeft, Folder, FileText, ChevronRight } from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()

const shareToken = route.query.token as string
const shareCode = route.query.code as string
const queryNodeId = route.query.nodeId ? parseInt(route.query.nodeId as string) : undefined

const loading = ref(true)
const noteData = ref<ShareAccessRes | null>(null)
const errorMsg = ref('')

const toast = ref({
  show: false,
  message: '',
  type: 'error' as 'success' | 'error'
})

const showToast = (message: string, type: 'success' | 'error' = 'error') => {
  toast.value = { show: true, message, type }
}

const fetchNote = async (nodeId?: number) => {
  loading.value = true
  errorMsg.value = ''
  try {
    const payload = {
      share_token: shareToken,
      share_code: shareCode,
      // 如果没有传 nodeId，后端接口通常可以通过 token 和 code 访问到分享的根节点，
      // 具体逻辑依后端实现而定。如果必需 nodeId，则需要在验证页面获取并传递过来。
      private_node_id: nodeId || 0 
    }
    const res = await accessShareNote(payload)
    if (res.data?.code === 200 && res.data.data) {
      noteData.value = res.data.data
      // 更新 URL 以支持书签和刷新
      if (nodeId) {
        router.replace({ query: { ...route.query, nodeId: nodeId.toString() } })
      }
    } else {
      errorMsg.value = res.data?.msg || '获取笔记内容失败'
    }
  } catch (err: any) {
    errorMsg.value = err.response?.data?.msg || '获取笔记内容失败'
    if (err.response?.status === 400 || err.response?.data?.code === 400) {
      // 提取码错误或过期，跳回验证页
      setTimeout(() => {
        router.replace(`/share/verify?token=${shareToken}`)
      }, 1500)
    }
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  if (!shareToken || !shareCode) {
    errorMsg.value = '缺少必要的分享参数'
    loading.value = false
    setTimeout(() => {
      router.replace('/')
    }, 2000)
    return
  }
  fetchNote(queryNodeId)
})

const handleChildClick = (childId: number) => {
  fetchNote(childId)
}

const handleBackToParent = () => {
  if (noteData.value?.parent?.id) {
    fetchNote(noteData.value.parent.id)
  }
}

// 简单的 Markdown 换行渲染
const renderedContent = computed(() => {
  if (!noteData.value?.content) return ''
  // 将 \n 替换为 <br> 以简单展示，实际项目中可能需要 markdown 渲染库如 marked
  return noteData.value.content.replace(/\n/g, '<br>')
})

</script>

<template>
  <div class="share-access-page">
    <Toast 
      v-if="toast.show" 
      :message="toast.message" 
      :type="toast.type" 
      @close="toast.show = false" 
    />

    <div class="topbar">
      <div class="topbar-left">
        <button v-if="noteData?.parent" class="ghost-btn" @click="handleBackToParent">
          <ArrowLeft :size="18" />
          返回上一级
        </button>
        <span v-if="noteData?.parent" class="separator">/</span>
        <h1 class="page-title">{{ noteData?.title || '加载中...' }}</h1>
      </div>
    </div>

    <div class="content-container">
      <!-- Loading -->
      <div v-if="loading" class="state-container">
        <Loader2 class="spin" :size="48" />
        <p>正在加载内容...</p>
      </div>

      <!-- Error -->
      <div v-else-if="errorMsg" class="state-container">
        <p class="error-text">{{ errorMsg }}</p>
      </div>

      <!-- Content -->
      <div v-else-if="noteData" class="note-viewer">
        
        <!-- 文件夹视图 -->
        <div v-if="noteData.type === 'folder'" class="folder-view">
          <div v-if="!noteData.children || noteData.children.length === 0" class="empty-folder">
            该文件夹为空
          </div>
          <div v-else class="list-container">
            <div 
              v-for="child in noteData.children" 
              :key="child.id" 
              class="list-item"
              @click="handleChildClick(child.id)"
            >
              <div class="item-icon">
                <Folder v-if="child.type === 'folder'" :size="24" class="icon-pink" />
                <FileText v-else :size="24" class="icon-blue" />
              </div>
              <div class="item-info">
                <h3>{{ child.title }}</h3>
                <span>{{ new Date(child.updatedAt).toLocaleDateString() }}</span>
              </div>
              <ChevronRight :size="20" class="icon-gray" />
            </div>
          </div>
        </div>

        <!-- Markdown 视图 -->
        <div v-else-if="noteData.type === 'markdown'" class="markdown-view">
          <div class="markdown-content" v-html="renderedContent"></div>
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
  --warm-white: #f6f5f4;
}

.share-access-page {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background-color: white;
  font-family: "NotionInter", Inter, -apple-system, system-ui, sans-serif;
}

.topbar {
  height: 60px;
  border-bottom: 1px solid rgba(0,0,0,0.1);
  display: flex;
  align-items: center;
  padding: 0 24px;
  position: sticky;
  top: 0;
  background: white;
  z-index: 10;
}

.topbar-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.ghost-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  background: transparent;
  border: none;
  cursor: pointer;
  color: var(--warm-gray-500);
  font-size: 14px;
  padding: 6px 8px;
  border-radius: 4px;
  transition: background 0.2s;
}

.ghost-btn:hover {
  background: var(--warm-white);
  color: var(--notion-black);
}

.separator {
  color: var(--warm-gray-300);
}

.page-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--notion-black);
  margin: 0;
}

.content-container {
  flex: 1;
  max-width: 900px;
  margin: 0 auto;
  width: 100%;
  padding: 40px 24px;
}

.state-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 100px 0;
  color: var(--warm-gray-500);
}

.state-container p {
  margin-top: 16px;
}

.error-text {
  color: #ef4444;
  font-size: 16px;
}

.icon-pink { color: #ec4899; }
.icon-blue { color: #3b82f6; }
.icon-gray { color: var(--warm-gray-300); }

/* Folder View */
.list-container {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.list-item {
  display: flex;
  align-items: center;
  padding: 16px;
  border-radius: 8px;
  border: 1px solid rgba(0,0,0,0.08);
  cursor: pointer;
  transition: all 0.2s;
}

.list-item:hover {
  background: var(--warm-white);
  border-color: rgba(0,0,0,0.15);
}

.item-icon {
  margin-right: 16px;
}

.item-info {
  flex: 1;
}

.item-info h3 {
  margin: 0 0 4px 0;
  font-size: 16px;
  color: var(--notion-black);
  font-weight: 500;
}

.item-info span {
  font-size: 12px;
  color: var(--warm-gray-500);
}

.empty-folder {
  text-align: center;
  padding: 64px 0;
  color: var(--warm-gray-500);
  background: var(--warm-white);
  border-radius: 8px;
}

/* Markdown View */
.markdown-content {
  font-size: 16px;
  line-height: 1.6;
  color: var(--notion-black);
  white-space: pre-wrap;
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
