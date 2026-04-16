<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { accessShareNote } from '../api/user'
import type { ShareAccessRes } from '../types/user'
import Toast from '../components/Toast.vue'
import { Loader2, ArrowLeft, Folder, FileText } from 'lucide-vue-next'

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

const fetchNote = async (nodeId?: number) => {
  loading.value = true
  errorMsg.value = ''
  try {
    const payload = {
      shareToken,
      shareCode,
      privateNoteId: nodeId || 0
    }
    const res = await accessShareNote(payload)
    if (res.data?.code === 200 && res.data.data) {
      noteData.value = res.data.data
      if (nodeId) {
        router.replace({ query: { ...route.query, nodeId: nodeId.toString() } })
      }
    } else {
      errorMsg.value = res.data?.msg || '获取笔记内容失败'
    }
  } catch (err: any) {
    errorMsg.value = err.response?.data?.msg || '获取笔记内容失败'
    if (err.response?.status === 400 || err.response?.data?.code === 400) {
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

const renderedContent = computed(() => {
  if (!noteData.value?.content) return ''
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
          <ArrowLeft :size="16" />
          <span>{{ noteData.parent.title }}</span>
        </button>
        <span v-if="noteData?.parent" class="separator">/</span>
        <div class="current-title-bar" v-if="noteData">
          <Folder v-if="noteData.type === 'folder'" :size="16" class="icon-gray" />
          <FileText v-else :size="16" class="icon-gray" />
          <span class="page-title">{{ noteData.title }}</span>
        </div>
      </div>
      <div class="topbar-right">
         <div class="badge">已分享</div>
      </div>
    </div>

    <div class="content-wrapper">
      <div class="content-container">
        <!-- Loading -->
        <div v-if="loading" class="state-container">
          <Loader2 class="spin" :size="32" />
          <p>加载中...</p>
        </div>

        <!-- Error -->
        <div v-else-if="errorMsg" class="state-container error-state">
          <p class="error-text">{{ errorMsg }}</p>
        </div>

        <!-- Content -->
        <div v-else-if="noteData" class="note-viewer">
          
          <header class="document-header">
            <h1 class="document-title">{{ noteData.title }}</h1>
            <div class="document-meta" v-if="noteData.type === 'folder'">
              <span>{{ noteData.children?.length || 0 }} 个项目</span>
            </div>
          </header>

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
                  <Folder v-if="child.type === 'folder'" :size="20" class="icon-blue" />
                  <FileText v-else :size="20" class="icon-gray" />
                </div>
                <div class="item-info">
                  <h3>{{ child.title }}</h3>
                </div>
                <div class="item-meta">
                  <span>{{ new Date(child.updatedAt).toLocaleDateString() }}</span>
                </div>
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
  </div>
</template>

<style scoped>
:root {
  --notion-black: rgba(0,0,0,0.95);
  --notion-blue: #0075de;
  --notion-blue-hover: #005bab;
  --warm-white: #f6f5f4;
  --warm-dark: #31302e;
  --warm-gray-500: #615d59;
  --warm-gray-300: #a39e98;
  --whisper-border: 1px solid rgba(0,0,0,0.1);
}

.share-access-page {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background-color: #ffffff;
  font-family: "NotionInter", Inter, -apple-system, system-ui, sans-serif;
  color: rgba(0,0,0,0.95);
}

.topbar {
  height: 52px;
  border-bottom: 1px solid rgba(0,0,0,0.1);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  position: sticky;
  top: 0;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(8px);
  z-index: 10;
}

.topbar-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.ghost-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  background: transparent;
  border: none;
  cursor: pointer;
  color: #615d59;
  font-size: 14px;
  padding: 4px 8px;
  border-radius: 4px;
  transition: background 0.2s;
  font-weight: 500;
}

.ghost-btn:hover {
  background: #f6f5f4;
  color: rgba(0,0,0,0.95);
}

.separator {
  color: #a39e98;
  font-size: 14px;
}

.current-title-bar {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 8px;
  border-radius: 4px;
}

.page-title {
  font-size: 14px;
  font-weight: 500;
  color: rgba(0,0,0,0.95);
  margin: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 200px;
}

.badge {
  background: #f2f9ff;
  color: #097fe8;
  padding: 4px 8px;
  border-radius: 9999px;
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.125px;
}

.content-wrapper {
  flex: 1;
  display: flex;
  justify-content: center;
}

.content-container {
  width: 100%;
  max-width: 900px;
  padding: 64px 48px 120px;
}

@media (max-width: 768px) {
  .content-container {
    padding: 32px 24px 80px;
  }
}

.state-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 100px 0;
  color: #615d59;
}

.state-container p {
  margin-top: 16px;
  font-size: 14px;
}

.error-state {
  color: #ef4444;
}

.error-text {
  font-size: 16px;
  font-weight: 500;
}

.icon-blue { color: #0075de; }
.icon-gray { color: #615d59; }

/* Document Header */
.document-header {
  margin-bottom: 32px;
}

.document-title {
  font-size: 40px;
  font-weight: 700;
  line-height: 1.2;
  letter-spacing: -0.625px;
  color: rgba(0,0,0,0.95);
  margin: 0 0 12px 0;
  word-break: break-word;
}

.document-meta {
  font-size: 14px;
  color: #615d59;
}

/* Folder View */
.list-container {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.list-item {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.1s ease;
  border: 1px solid transparent;
}

.list-item:hover {
  background: #f6f5f4;
}

.item-icon {
  margin-right: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.item-info {
  flex: 1;
}

.item-info h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 500;
  color: rgba(0,0,0,0.95);
}

.item-meta {
  font-size: 14px;
  color: #a39e98;
}

.empty-folder {
  padding: 32px 0;
  color: #615d59;
  font-size: 16px;
  border-top: 1px solid rgba(0,0,0,0.1);
}

/* Markdown View */
.markdown-view {
  padding-top: 16px;
}

.markdown-content {
  font-size: 16px;
  line-height: 1.6;
  color: rgba(0,0,0,0.95);
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
