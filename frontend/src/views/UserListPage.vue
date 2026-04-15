<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, BookOpen, Clock, Activity, Loader2, Star, FolderHeart, FileText, Folder } from 'lucide-vue-next'
import { getUserRecentSubjects, getUserLikedSubjects, getUserCollectFolders, getPrivateNoteDetail } from '../api/user'

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
  }
  // 其他类型暂不跳转或视需求而定
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
                  <Activity :size="12" class="icon-success" />
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
          <div v-for="item in list" :key="item.id" class="folder-card">
            <div class="folder-icon">
              <FolderHeart :size="32" />
            </div>
            <div class="folder-info">
              <h3>{{ item.name }}</h3>
              <p>{{ item.description || '暂无描述' }}</p>
              <span class="badge">{{ item.isPublic ? '公开' : '私密' }}</span>
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
  </div>
</template>

<style scoped>
.list-page {
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

.folder-icon { color: #ff64c8; background: #fff5f5; }
.note-icon.folder { color: #dd5b00; }
.note-icon.file { color: #0075de; }

.icon-blue { color: #0075de; }
.icon-success { color: #1aae39; }
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
