<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, BookOpen, Clock, Activity, Loader2, Star, FolderHeart, FileText, Folder, X, ToggleRight, ToggleLeft, ChevronRight } from 'lucide-vue-next'
import { getUserRecentSubjects, getUserLikedSubjects, getUserCollectFolders, getPrivateNoteDetail, getSubjectsInFolder, updateCollectFolderPublic } from '../api/user'

const route = useRoute()
const router = useRouter()
const type = computed(() => route.params.type as string)

const list = ref<any[]>([])
const loading = ref(false)
const error = ref('')

const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const hasMore = computed(() => list.value.length < total.value)

// --- 收藏夹弹窗相关 ---
const showFolderModal = ref(false)
const selectedFolder = ref<any>(null)
const folderSubjects = ref<any[]>([])
const loadingFolderSubjects = ref(false)
const folderPage = ref(1)
const folderTotal = ref(0)
const folderHasMore = computed(() => folderSubjects.value.length < folderTotal.value)

const openFolderModal = async (folder: any) => {
  selectedFolder.value = folder
  showFolderModal.value = true
  folderPage.value = 1
  folderSubjects.value = []
  await fetchFolderSubjects()
}

const fetchFolderSubjects = async (isLoadMore = false) => {
  if (!selectedFolder.value) return
  loadingFolderSubjects.value = true
  try {
    const res = await getSubjectsInFolder(selectedFolder.value.id, folderPage.value, pageSize.value)
    if (res.data?.code === 200 && res.data.data) {
      const data = res.data.data
      const newList = data.list || []
      folderSubjects.value = isLoadMore ? [...folderSubjects.value, ...newList] : newList
      folderTotal.value = data.total || 0
    } else {
      // 如果接口返回非200或无数据，确保列表为空
      if (!isLoadMore) folderSubjects.value = []
      folderTotal.value = 0
    }
  } catch (err) {
    console.error('Fetch folder subjects error:', err)
    // 发生错误时也确保列表为空，停止加载
    if (!isLoadMore) folderSubjects.value = []
    folderTotal.value = 0
  } finally {
    loadingFolderSubjects.value = false
  }
}

const loadMoreFolderSubjects = () => {
  if (loadingFolderSubjects.value || !folderHasMore.value) return
  folderPage.value++
  fetchFolderSubjects(true)
}

const handleFolderItemClick = (item: any) => {
  const url = `/subject/${item.id}` + (item.lastNodeId ? `?nodeId=${item.lastNodeId}` : '')
  router.push(url)
}

const updatingFolderIds = ref<Set<number>>(new Set())

const handleToggleFolderPublic = async (folder: any) => {
  if (!folder?.id) return
  if (updatingFolderIds.value.has(folder.id)) return

  updatingFolderIds.value.add(folder.id)
  updatingFolderIds.value = new Set(updatingFolderIds.value)

  const currentPublic = !!folder.isPublic
  const nextPublic: 0 | 1 = currentPublic ? 0 : 1
  const prevValue = folder.isPublic
  folder.isPublic = nextPublic

  try {
    const res = await updateCollectFolderPublic(folder.id, nextPublic)
    if (res.data?.code !== 200) {
      folder.isPublic = prevValue
    }
  } catch (err) {
    folder.isPublic = prevValue
  } finally {
    updatingFolderIds.value.delete(folder.id)
    updatingFolderIds.value = new Set(updatingFolderIds.value)
  }
}
// ----------------------

const titleMap: Record<string, string> = {
  'recent-learning': '最近学习',
  'liked-subjects': '点赞的教材',
  'collections': '我的收藏夹',
  'private-notes': '私人笔记'
}

const pageTitle = computed(() => titleMap[type.value] || '列表')

const formatDate = (dateStr: string) => {
  if (!dateStr) return '未知时间'
  const d = new Date(dateStr)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

const getCoverStyle = (id: number) => {
  const palettes: Array<[string, string]> = [
    ['#3b82f6', '#8b5cf6'],
    ['#06b6d4', '#3b82f6'],
    ['#22c55e', '#14b8a6'],
    ['#f97316', '#ef4444'],
    ['#a855f7', '#ec4899']
  ]
  const [from, to] = palettes[id % palettes.length]
  return { background: `linear-gradient(135deg, ${from}, ${to})` }
}

const fetchList = async (isLoadMore = false) => {
  if (!isLoadMore) {
    loading.value = true
    list.value = []
  }
  error.value = ''
  
  try {
    if (type.value === 'recent-learning') {
      const res = await getUserRecentSubjects(page.value, pageSize.value)
      if (res.data?.code === 200) {
        const normalized = (res.data.data.list || []).map(item => {
          const subject = (item as any)?.subject ?? item ?? {}
          return {
            ...subject,
            status: (item as any)?.status ?? subject?.status ?? '',
            lastStudyTime: (item as any)?.lastStudyTime ?? subject?.lastStudyTime ?? ''
          }
        })
        list.value = isLoadMore ? [...list.value, ...normalized] : normalized
        total.value = res.data.data.total || 0
      }
    } else if (type.value === 'liked-subjects') {
      const res = await getUserLikedSubjects(page.value, pageSize.value)
      if (res.data?.code === 200) {
        const normalized = (res.data.data.list || []).map(item => {
          const subject = (item as any)?.subject ?? item ?? {}
          return {
            ...subject,
            status: (item as any)?.status ?? subject?.status ?? '',
            lastStudyTime: (item as any)?.lastStudyTime ?? subject?.lastStudyTime ?? ''
          }
        })
        list.value = isLoadMore ? [...list.value, ...normalized] : normalized
        total.value = res.data.data.total || 0
      }
    } else if (type.value === 'collections') {
      // 不分页
      const res = await getUserCollectFolders()
      if (res.data?.code === 200) {
        list.value = res.data.data || []
        total.value = list.value.length
      }
    } else if (type.value === 'private-notes') {
      const res = await getPrivateNoteDetail(0, 2, page.value, pageSize.value)
      if (res.data?.code === 200 && res.data.data) {
        const data = res.data.data
        if (data.type === 'folder') {
          const children = Array.isArray(data.children) ? data.children : []
          list.value = isLoadMore ? [...list.value, ...children] : children
          total.value = data.total || 0
        }
      }
    }
  } catch (err: any) {
    error.value = err.response?.data?.msg || '加载失败'
  } finally {
    loading.value = false
  }
}

const loadMore = () => {
  if (loading.value || !hasMore.value) return
  page.value++
  fetchList(true)
}

const handleItemClick = (item: any) => {
  if (type.value === 'recent-learning' || type.value === 'liked-subjects') {
    const url = `/subject/${item.id}` + (item.lastNodeId ? `?nodeId=${item.lastNodeId}` : '')
    router.push(url)
  } else if (type.value === 'collections') {
    openFolderModal(item)
  }
}

onMounted(() => {
  fetchList()
})
</script>

<template>
  <div class="list-page">
    <div class="topbar">
      <button class="ghost-btn" @click="router.push('/me')">
        <ArrowLeft :size="18" />
        返回个人主页
      </button>
      <h2 class="page-title">{{ pageTitle }}</h2>
      <div style="width: 100px;"></div>
    </div>

    <div class="content-container">
      <div v-if="loading && page === 1" class="loading-state">
        <Loader2 class="spin" :size="32" />
        <p>加载中...</p>
      </div>

      <div v-else-if="error" class="error-state">
        <p>{{ error }}</p>
        <button class="primary-btn" @click="fetchList(false)">重试</button>
      </div>

      <div v-else-if="!list.length" class="empty-state">
        <p>暂无数据</p>
      </div>

      <div v-else class="list-grid">
        <!-- 教材样式 (最近学习/点赞) -->
        <template v-if="type === 'recent-learning' || type === 'liked-subjects'">
          <div v-for="item in list" :key="item.id" class="subject-card" @click="handleItemClick(item)">
            <div class="subject-cover" :style="getCoverStyle(item.id)">
              <BookOpen :size="24" />
            </div>
            <div class="subject-info">
              <h3>{{ item.name }}</h3>
              <p class="desc">{{ item.description || '暂无简介' }}</p>
              <div class="meta">
                <span v-if="item.progressPercent !== undefined" class="meta-item">
                  <Activity :size="12" class="icon-green" />
                  进度: {{ item.progressPercent }}%
                </span>
                <span v-if="item.lastStudyTime" class="meta-item">
                  <Clock :size="12" class="icon-brown" />
                  {{ formatDate(item.lastStudyTime) }}
                </span>
              </div>
            </div>
          </div>
        </template>

        <!-- 收藏夹样式 -->
        <template v-else-if="type === 'collections'">
          <div v-for="item in list" :key="item.id" class="folder-card" @click="handleItemClick(item)">
            <div class="folder-icon">
              <FolderHeart :size="32" />
            </div>
            <div class="folder-info">
              <h3>{{ item.name }}</h3>
              <p>{{ item.description || '暂无描述' }}</p>
              <button
                class="badge-btn"
                :class="{ public: !!item.isPublic }"
                :disabled="updatingFolderIds.has(item.id)"
                @click.stop="handleToggleFolderPublic(item)"
              >
                <Loader2 v-if="updatingFolderIds.has(item.id)" class="spin" :size="16" />
                <ToggleRight v-else-if="!!item.isPublic" :size="16" />
                <ToggleLeft v-else :size="16" />
                <span>{{ !!item.isPublic ? '公开' : '私密' }}</span>
              </button>
            </div>
          </div>
        </template>

        <!-- 私人笔记样式 -->
        <template v-else-if="type === 'private-notes'">
          <div v-for="item in list" :key="item.id" class="note-card">
            <div class="note-icon" :class="item.type">
              <Folder v-if="item.type === 'folder'" :size="24" />
              <FileText v-else :size="24" />
            </div>
            <div class="note-info">
              <h3>{{ item.title }}</h3>
              <div class="meta">
                <span>{{ formatDate(item.updatedAt) }}</span>
                <span class="badge" :class="{ public: item.isPublic }">{{ item.isPublic ? '公开' : '私密' }}</span>
              </div>
            </div>
          </div>
        </template>
      </div>

      <div v-if="hasMore" class="load-more-container">
        <button class="load-more-btn" :disabled="loading" @click="loadMore">
          <Loader2 v-if="loading" class="spin" :size="16" />
          <span v-else>加载更多</span>
        </button>
      </div>
      <div v-else-if="list.length > 0 && type !== 'collections'" class="no-more">
        没有更多数据了
      </div>
    </div>

    <!-- 收藏夹教材列表弹窗 -->
    <Teleport to="body">
      <div v-if="showFolderModal" class="modal-overlay" @click.self="showFolderModal = false">
        <div class="modal-content large-modal">
          <header class="modal-header">
            <div class="header-left">
              <FolderHeart :size="24" class="icon-teal" />
              <div class="header-text">
                <h3>{{ selectedFolder?.name }}</h3>
                <p>{{ selectedFolder?.description || '收藏夹教材列表' }}</p>
              </div>
            </div>
            <button class="close-btn" @click="showFolderModal = false">
              <X :size="24" />
            </button>
          </header>

          <div class="modal-body">
            <div v-if="loadingFolderSubjects && folderPage === 1" class="modal-loading">
              <Loader2 class="spin" :size="32" />
              <p>加载中...</p>
            </div>
            <div v-else-if="!folderSubjects.length" class="modal-empty">
              <p>该收藏夹下暂无教材</p>
            </div>
            <div v-else class="modal-list">
              <div v-for="item in folderSubjects" :key="item.id" class="subject-row" @click="handleFolderItemClick(item)">
                <div class="subject-cover-mini" :style="getCoverStyle(item.id)">
                  <BookOpen :size="16" />
                </div>
                <div class="subject-row-info">
                  <h4>{{ item.name }}</h4>
                  <p>{{ item.description || '暂无简介' }}</p>
                </div>
                <div class="subject-row-meta">
                  <span v-if="item.progressPercent !== undefined" class="meta-item">
                    <Activity :size="12" class="icon-green" />
                    {{ item.progressPercent }}%
                  </span>
                  <ChevronRight :size="18" class="row-arrow" />
                </div>
              </div>

              <div v-if="folderHasMore" class="modal-load-more">
                <button class="ghost-btn" :disabled="loadingFolderSubjects" @click="loadMoreFolderSubjects">
                  <Loader2 v-if="loadingFolderSubjects" class="spin" :size="16" />
                  <span v-else>查看更多</span>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.list-page {
  --notion-black: rgba(0,0,0,0.95);
  --notion-white: #ffffff;
  --notion-blue: #0075de;
  --notion-blue-hover: #005bab;
  --warm-white: #f6f5f4;
  --warm-dark: #31302e;
  --warm-gray-500: #615d59;
  --warm-gray-300: #a39e98;
  --whisper-border: 1px solid rgba(0,0,0,0.1);
  --card-shadow: rgba(0,0,0,0.04) 0px 4px 18px, rgba(0,0,0,0.027) 0px 2.025px 7.84688px, rgba(0,0,0,0.02) 0px 0.8px 2.925px, rgba(0,0,0,0.01) 0px 0.175px 1.04062px;
  --deep-shadow: rgba(0,0,0,0.01) 0px 1px 3px, rgba(0,0,0,0.02) 0px 3px 7px, rgba(0,0,0,0.02) 0px 7px 15px, rgba(0,0,0,0.04) 0px 14px 28px, rgba(0,0,0,0.05) 0px 23px 52px;
  
  width: 100vw;
  height: 100vh;
  background: var(--notion-white);
  color: var(--notion-black);
  font-family: "NotionInter", Inter, -apple-system, system-ui, sans-serif;
  display: flex;
  flex-direction: column;
}

.topbar {
  height: 64px;
  background: var(--notion-white);
  border-bottom: 1px solid rgba(0,0,0,0.1);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  flex-shrink: 0;
}

.page-title {
  margin: 0;
  font-size: 18px;
  font-weight: 700;
  color: var(--notion-black);
  letter-spacing: -0.2px;
}

.ghost-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  border-radius: 4px;
  border: 1px solid rgba(0,0,0,0.1);
  background: transparent;
  color: var(--notion-black);
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.1s ease;
}

.ghost-btn:hover {
  background: rgba(0,0,0,0.05);
}

.content-container {
  flex: 1;
  overflow-y: auto;
  padding: 40px 24px;
  max-width: 1200px;
  margin: 0 auto;
  width: 100%;
}

.list-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 24px;
}

.subject-card, .folder-card, .note-card {
  background: var(--notion-white);
  border-radius: 12px;
  padding: 24px;
  border: 1px solid rgba(0,0,0,0.1);
  box-shadow: rgba(0,0,0,0.04) 0px 4px 18px, rgba(0,0,0,0.027) 0px 2.025px 7.84688px;
  display: flex;
  gap: 20px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.subject-card:hover, .folder-card:hover, .note-card:hover {
  transform: translateY(-2px);
  box-shadow: rgba(0,0,0,0.06) 0px 6px 22px, rgba(0,0,0,0.04) 0px 3px 10px;
  background: var(--warm-white);
}

.subject-cover {
  width: 60px;
  height: 80px;
  border-radius: 8px;
  border: 1px solid rgba(0,0,0,0.1);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  flex-shrink: 0;
}

.folder-icon, .note-icon {
  width: 56px;
  height: 56px;
  border-radius: 12px;
  background: var(--warm-white);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--warm-gray-500);
  flex-shrink: 0;
  border: 1px solid rgba(0,0,0,0.05);
}

.folder-icon { color: #2a9d99; background: #f0f9f9; }
.note-icon.folder { color: #ff64c8; }
.note-icon.file { color: #0075de; }

.icon-blue { color: #0075de; }
.icon-green { color: #1aae39; }
.icon-teal { color: #2a9d99; }
.icon-warning { color: #dd5b00; }
.icon-danger { color: #eb5757; }
.icon-pink { color: #ff64c8; }
.icon-purple { color: #391c57; }
.icon-brown { color: #523410; }

.subject-info, .folder-info, .note-info {
  flex: 1;
  min-width: 0;
}

.subject-info h3, .folder-info h3, .note-info h3 {
  margin: 0 0 8px;
  font-size: 16px;
  font-weight: 700;
  color: var(--notion-black);
  letter-spacing: -0.2px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.folder-info p {
  font-size: 13px;
  color: var(--warm-gray-500);
  margin: 0 0 12px;
}

.badge-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 12px;
  border-radius: 999px;
  background: rgba(0,0,0,0.05);
  color: var(--warm-gray-500);
  border: none;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.badge-btn:hover {
  background: rgba(0,0,0,0.1);
}

.badge-btn.public {
  background: #f2f9ff;
  color: #0075de;
}

.badge-btn.public:hover {
  background: #e1f0ff;
}

.desc {
  font-size: 13px;
  color: var(--warm-gray-500);
  margin: 0 0 12px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  line-height: 1.5;
}

.meta {
  display: flex;
  align-items: center;
  gap: 16px;
  font-size: 12px;
  color: var(--warm-gray-300);
  font-weight: 500;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

.badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 999px;
  background: var(--warm-white);
  color: var(--warm-gray-500);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.badge.public {
  background: #f2f9ff;
  color: #0075de;
}

/* Modal Styles */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.4);
  backdrop-filter: blur(2px);
  display: grid;
  place-items: center;
  z-index: 1000;
}

.modal-content.large-modal {
  width: 90%;
  max-width: 800px;
  max-height: 85vh;
  background: var(--notion-white);
  border-radius: 12px;
  box-shadow: var(--deep-shadow);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid rgba(0,0,0,0.1);
}

.modal-header {
  padding: 24px;
  border-bottom: 1px solid rgba(0,0,0,0.1);
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.header-left {
  display: flex;
  gap: 16px;
  align-items: center;
}

.header-text h3 {
  margin: 0;
  font-size: 20px;
  font-weight: 700;
  color: var(--notion-black);
}

.header-text p {
  margin: 4px 0 0;
  font-size: 14px;
  color: var(--warm-gray-500);
}

.close-btn {
  background: transparent;
  border: none;
  color: var(--warm-gray-300);
  cursor: pointer;
  padding: 4px;
  border-radius: 6px;
  transition: all 0.2s ease;
}

.close-btn:hover {
  background: var(--warm-white);
  color: var(--notion-black);
}

.modal-body {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
}

.modal-loading, .modal-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 0;
  color: var(--warm-gray-300);
  gap: 16px;
}

.modal-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.subject-row {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  border-radius: 8px;
  background: transparent;
  cursor: pointer;
  transition: all 0.1s ease;
  gap: 16px;
}

.subject-row:hover {
  background: var(--warm-white);
}

.subject-cover-mini {
  width: 32px;
  height: 44px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  flex-shrink: 0;
}

.subject-row-info {
  flex: 1;
  min-width: 0;
}

.subject-row-info h4 {
  margin: 0 0 4px;
  font-size: 15px;
  font-weight: 600;
  color: var(--notion-black);
}

.subject-row-info p {
  margin: 0;
  font-size: 13px;
  color: var(--warm-gray-500);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.subject-row-meta {
  display: flex;
  align-items: center;
  gap: 16px;
  color: var(--warm-gray-300);
  font-size: 13px;
  font-weight: 500;
}

.row-arrow {
  color: var(--warm-gray-300);
  transition: transform 0.2s;
}

.subject-row:hover .row-arrow {
  transform: translateX(4px);
  color: var(--notion-blue);
}

.modal-load-more {
  display: flex;
  justify-content: center;
  margin-top: 16px;
  padding-bottom: 8px;
}

.loading-state, .error-state, .empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 20px;
  color: var(--warm-gray-300);
  gap: 16px;
  text-align: center;
}

.spin { animation: spin 1s linear infinite; }
@keyframes spin { 100% { transform: rotate(360deg); } }

.primary-btn {
  padding: 8px 24px;
  background: #0075de;
  color: white;
  border: none;
  border-radius: 6px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s;
}

.primary-btn:hover { background: #005bab; }

.load-more-container {
  display: flex;
  justify-content: center;
  margin-top: 48px;
}

.load-more-btn {
  padding: 10px 32px;
  border-radius: 999px;
  border: 1px solid rgba(0,0,0,0.1);
  background: var(--notion-white);
  color: var(--notion-black);
  font-weight: 600;
  font-size: 14px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: all 0.2s ease;
  box-shadow: 0 2px 8px rgba(0,0,0,0.05);
}

.load-more-btn:hover:not(:disabled) {
  background: var(--warm-white);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
}

.load-more-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.no-more {
  text-align: center;
  margin-top: 48px;
  color: var(--warm-gray-300);
  font-size: 14px;
  font-weight: 500;
}
</style>
