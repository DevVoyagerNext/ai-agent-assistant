<script setup lang="ts">
import { computed, reactive, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import Toast from '../components/Toast.vue'
import Skeleton from '../components/Skeleton.vue'
import { 
  ArrowLeft, LogOut, RefreshCcw, 
  Activity, BookOpen, Share2, Book, Users, Star, Layers, FolderHeart, Bookmark,
  Folder, FileText, ToggleRight, ToggleLeft, Edit3, X, Loader2, ChevronRight
} from 'lucide-vue-next'
import { useUserProfile } from '../composables/useUserProfile'
import ActivityCalendar from '../components/ActivityCalendar.vue'
import { updatePrivateNoteTitle, updatePrivateNotePublic } from '../api/user'

const router = useRouter()
const isAuthenticated = computed(() => !!localStorage.getItem('token'))

onMounted(() => {
  if (!isAuthenticated.value) {
    router.replace('/login')
  }
})
const {
  userInfo,
  activities,
  publicPrivateNotes,
  sharedNotes,
  collectFolders,
  likedSubjects,
  recentSubjects,
  
  loadingUserInfo,
  loadingActivities,
  loadingPublicPrivateNotes,
  loadingSharedNotes,
  loadingLearnedSubjects,
  loadingCollectFolders,
  loadingLikedSubjects,
  loadingRecentSubjects,
  
  errorUserInfo,
  refreshAll
} = useUserProfile()

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

const avatarText = computed(() => {
  const name = userInfo.value?.username?.trim() || ''
  return name ? name.slice(0, 1).toUpperCase() : 'U'
})

const isGlobalLoading = computed(() => 
  loadingUserInfo.value || 
  loadingActivities.value || 
  loadingPublicPrivateNotes.value || 
  loadingSharedNotes.value || 
  loadingLearnedSubjects.value
)

const handleRefresh = () => {
  if (!isAuthenticated.value) {
    router.replace('/login')
    return
  }
  refreshAll()
  if (!errorUserInfo.value) {
    showToast('刷新成功', 'success')
  }
}

const handleLogout = () => {
  localStorage.removeItem('token')
  localStorage.removeItem('refreshToken')
  localStorage.removeItem('expiresAt')
  localStorage.removeItem('user')
  router.push('/login')
}

// 简单的日期格式化
const formatDate = (dateStr: string) => {
  const d = new Date(dateStr)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

const privateNotesPreview = computed(() => {
  if (!publicPrivateNotes.value?.length) return []
  return publicPrivateNotes.value.slice(0, 5)
})

const privateNotesHasMore = computed(() => (publicPrivateNotes.value?.length || 0) > 5)

// ----------------- 私人笔记修改相关 -----------------
const showRenameModal = ref(false)
const renameTitle = ref('')
const renaming = ref(false)
const pendingRenameNote = ref<any>(null)

const openRename = (note: any) => {
  pendingRenameNote.value = note
  renameTitle.value = note.title
  showRenameModal.value = true
}

const handleRename = async () => {
  if (!renameTitle.value.trim() || !pendingRenameNote.value) return
  renaming.value = true
  try {
    const res = await updatePrivateNoteTitle(pendingRenameNote.value.id, renameTitle.value)
    if (res.data?.code === 200) {
      pendingRenameNote.value.title = renameTitle.value
      showToast('重命名成功', 'success')
      showRenameModal.value = false
    }
  } catch (err: any) {
    showToast(err.response?.data?.msg || '重命名失败', 'error')
  } finally {
    renaming.value = false
  }
}

const handleTogglePublic = async (note: any) => {
  const newPublic = note.isPublic === 1 ? 0 : 1
  try {
    const res = await updatePrivateNotePublic(note.id, newPublic as 0 | 1)
    if (res.data?.code === 200) {
      note.isPublic = newPublic as 0 | 1
      showToast(newPublic === 1 ? '已设为公开' : '已设为私密', 'success')
    }
  } catch (err: any) {
    showToast(err.response?.data?.msg || '修改失败', 'error')
  }
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
  return {
    background: `linear-gradient(135deg, ${from}, ${to})`
  }
}

const scrollTo = (id: string) => {
  const el = document.getElementById(id)
  if (el) {
    el.scrollIntoView({ behavior: 'smooth' })
  }
}
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
        返回大厅
      </button>

      <div class="topbar-actions">
        <button class="ghost-btn" :disabled="isGlobalLoading" @click="handleRefresh">
          <RefreshCcw :size="18" :class="{ 'spin': isGlobalLoading }" />
          刷新
        </button>
        <button v-if="isAuthenticated" class="danger-btn" @click="handleLogout">
          <LogOut :size="18" />
          退出登录
        </button>
      </div>
    </div>

    <div class="content-wrapper">
      <aside class="side-nav">
        <div class="nav-group">
          <div class="nav-item" @click="scrollTo('user-info')">
            <Users :size="18" class="icon-primary" />
            <span>关于我</span>
          </div>
          <div class="nav-item" @click="scrollTo('activity')">
            <Activity :size="18" class="icon-success" />
            <span>活跃度</span>
          </div>
          <div class="nav-item" @click="scrollTo('collections')">
            <FolderHeart :size="18" class="icon-danger" />
            <span>收藏夹</span>
          </div>
          <div class="nav-item" @click="scrollTo('recent')">
            <Layers :size="18" class="icon-teal" />
            <span>最近学习</span>
          </div>
          <div class="nav-item" @click="scrollTo('liked')">
            <Star :size="18" class="icon-warning" />
            <span>点赞教材</span>
          </div>
          <div class="nav-item" @click="scrollTo('private')">
            <BookOpen :size="18" class="icon-success" />
            <span>私人笔记</span>
          </div>
          <div class="nav-item" @click="scrollTo('shared')">
            <Share2 :size="18" class="icon-info" />
            <span>分享笔记</span>
          </div>
        </div>
      </aside>

      <div class="main-column">
        <!-- User Info Card -->
        <div id="user-info" class="card user-card">
          <div v-if="loadingUserInfo" class="hero-skeleton">
            <Skeleton width="76px" height="76px" border-radius="18px" />
            <div class="title-skeleton">
              <Skeleton width="150px" height="28px" />
              <Skeleton width="200px" height="16px" style="margin-top: 8px;" />
            </div>
          </div>
          <div v-else class="hero">
            <div class="avatar">
              <img v-if="userInfo?.avatarUrl" :src="userInfo.avatarUrl" alt="avatar" />
              <span v-else>{{ avatarText }}</span>
            </div>

            <div class="title">
              <div class="name-row">
                <h1 class="username">{{ userInfo?.username || '—' }}</h1>
              </div>
              <p class="subtitle">{{ userInfo?.signature || '保持学习，保持好奇' }}</p>
            </div>
          </div>

          <div v-if="loadingUserInfo" class="stats-skeleton">
            <Skeleton v-for="i in 4" :key="i" width="100%" height="60px" border-radius="12px" />
          </div>
          <div v-else class="stats-grid">
            <div class="stat-item">
              <span class="stat-value">{{ userInfo?.followersCount || 0 }}</span>
              <span class="stat-label"><Users :size="14" class="icon-primary" /> 粉丝</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ userInfo?.followingCount || 0 }}</span>
              <span class="stat-label"><Star :size="14" class="icon-primary" /> 关注</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ userInfo?.learnedSubjectsCount || 0 }}</span>
              <span class="stat-label"><Book :size="14" class="icon-primary" /> 已学课程</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ userInfo?.sharedNotesCount || 0 }}</span>
              <span class="stat-label"><Share2 :size="14" class="icon-primary" /> 分享笔记</span>
            </div>
          </div>
        </div>

        <!-- Activity Calendar Card -->
        <div id="activity" class="card">
          <div class="card-header">
            <Activity :size="20" class="icon-success" />
            <h2>活跃度</h2>
          </div>
          
          <div v-if="loadingActivities" class="activity-skeleton">
            <Skeleton v-for="i in 14" :key="i" width="100%" height="30px" border-radius="4px" />
          </div>
          <ActivityCalendar v-else :items="activities" />
        </div>

        <!-- Collect Folders -->
        <div id="collections" class="card">
          <div class="card-header clickable-header" @click="router.push('/me/collections')">
            <FolderHeart :size="20" class="icon-danger" />
            <h2>我的收藏夹</h2>
            <ChevronRight :size="18" class="header-arrow" />
          </div>
          
          <div v-if="loadingCollectFolders" class="list-skeleton">
            <div v-for="i in 2" :key="i" class="list-item-skeleton" style="padding: 12px 0;">
              <div class="skeleton-col">
                <Skeleton width="80%" height="16px" />
                <Skeleton width="40%" height="12px" style="margin-top: 8px;" />
              </div>
            </div>
          </div>
          <div v-else class="note-list">
            <div v-for="folder in collectFolders" :key="folder.id" class="note-item">
              <div class="note-main-info">
                <div class="note-title-line">
                  <FolderHeart :size="16" class="note-type-icon icon-danger" />
                  <h4>{{ folder.name }}</h4>
                </div>
              </div>
              <div class="note-item-actions">
                <button 
                  class="toggle-public-mini" 
                  :class="{ isPublic: folder.isPublic }"
                  :title="folder.isPublic ? '已公开' : '已私密'"
                >
                  <ToggleRight v-if="folder.isPublic" :size="18" />
                  <ToggleLeft v-else :size="18" />
                </button>
              </div>
            </div>
            <div v-if="!collectFolders.length" class="empty-state">
              暂无收藏夹
            </div>
          </div>
        </div>

        <!-- Learned Subjects Card -->
        <div id="recent" class="card">
          <div class="card-header clickable-header" @click="router.push('/me/recent-learning')">
            <Layers :size="20" class="icon-teal" />
            <h2>最近学习 (在学/已学)</h2>
            <ChevronRight :size="18" class="header-arrow" />
          </div>
          
          <div v-if="loadingRecentSubjects" class="list-skeleton">
            <div v-for="i in 2" :key="i" class="list-item-skeleton">
              <Skeleton width="48px" height="64px" border-radius="6px" />
              <div class="skeleton-col">
                <Skeleton width="60%" height="16px" />
                <Skeleton width="100%" height="8px" style="margin-top: 10px;" />
              </div>
            </div>
          </div>
          <div v-else class="subject-list">
            <div v-for="item in recentSubjects" :key="item.id" class="subject-item">
              <div class="subject-info" @click="router.push(`/subject/${item.id}?nodeId=${item.lastNodeId}`)">
                <div class="subject-cover" :style="getCoverStyle(item.id)">
                  <BookOpen :size="20" />
                </div>
                <div class="subject-text">
                  <h3>{{ item.name }}</h3>
                  <div class="subject-meta">
                    <span class="meta-item">
                      <Clock :size="12" class="icon-brown" />
                      {{ item.lastStudyTime ? formatDate(item.lastStudyTime) : '未知时间' }}
                    </span>
                    <span class="meta-item">
                      <Activity :size="12" class="icon-success" />
                      进度 {{ item.progressPercent }}%
                    </span>
                  </div>
                </div>
              </div>
            </div>
            <div v-if="!recentSubjects.length" class="empty-state">
              暂无学习数据
            </div>
          </div>
        </div>

        <!-- Liked Subjects -->
        <div id="liked" class="card">
          <div class="card-header clickable-header" @click="router.push('/me/liked-subjects')">
            <Star :size="20" class="icon-warning" />
            <h2>点赞的教材</h2>
            <ChevronRight :size="18" class="header-arrow" />
          </div>
          
          <div v-if="loadingLikedSubjects" class="list-skeleton">
            <div v-for="i in 2" :key="i" class="list-item-skeleton" style="padding: 12px 0;">
              <div class="skeleton-col">
                <Skeleton width="80%" height="16px" />
              </div>
            </div>
          </div>
          <div v-else class="note-list">
            <div 
              v-for="subject in likedSubjects" 
              :key="subject.id" 
              class="note-item"
              @click="router.push(`/subject/${subject.id}`)"
              style="cursor: pointer;"
            >
              <div class="note-main-info">
                <div class="note-title-line">
                  <Star :size="16" class="note-type-icon icon-warning" />
                  <h4>{{ subject.name }}</h4>
                </div>
              </div>
              <span v-if="subject.progressPercent > 0" class="date">进度 {{ subject.progressPercent }}%</span>
            </div>
            <div v-if="!likedSubjects.length" class="empty-state">
              暂无点赞教材
            </div>
          </div>
        </div>

        <!-- Private Notes -->
        <div id="private" class="card">
          <div class="card-header clickable-header" @click="router.push('/me/private-notes')">
            <BookOpen :size="20" class="icon-success" />
            <h2>私人笔记</h2>
            <ChevronRight :size="18" class="header-arrow" />
          </div>
          
          <div v-if="loadingPublicPrivateNotes" class="list-skeleton">
            <div v-for="i in 3" :key="i" class="list-item-skeleton" style="padding: 12px 0;">
              <div class="skeleton-col">
                <Skeleton width="80%" height="16px" />
                <Skeleton width="40%" height="12px" style="margin-top: 8px;" />
              </div>
            </div>
          </div>
          <div v-else class="note-list">
            <div v-for="note in privateNotesPreview" :key="note.id" class="note-item">
              <div class="note-main-info">
                <div class="note-title-line">
                  <Folder v-if="note.type === 'folder'" :size="16" class="note-type-icon icon-warning" />
                  <FileText v-else :size="16" class="note-type-icon icon-info" />
                  <h4 @click="openRename(note)">{{ note.title }}</h4>
                  <Edit3 :size="12" class="edit-icon-mini" @click="openRename(note)" />
                  <span class="date">{{ formatDate(note.updatedAt) }}</span>
                </div>
              </div>
              <div class="note-item-actions">
                <button 
                  class="toggle-public-mini" 
                  :class="{ isPublic: note.isPublic === 1 }"
                  @click="handleTogglePublic(note)"
                  :title="note.isPublic === 1 ? '已公开' : '已私密'"
                >
                  <ToggleRight v-if="note.isPublic === 1" :size="18" />
                  <ToggleLeft v-else :size="18" />
                </button>
              </div>
            </div>
            <div v-if="privateNotesHasMore" class="note-item note-ellipsis">
              <h4>...</h4>
              <span class="date">...</span>
            </div>
            <div v-if="!publicPrivateNotes.length" class="empty-state">
              暂无私人笔记
            </div>
          </div>
        </div>

        <!-- Shared Notes -->
        <div id="shared" class="card">
          <div class="card-header">
            <Share2 :size="20" class="icon-info" />
            <h2>分享笔记</h2>
          </div>
          
          <div v-if="loadingSharedNotes" class="list-skeleton">
            <div v-for="i in 3" :key="i" class="list-item-skeleton" style="padding: 12px 0;">
              <div class="skeleton-col">
                <Skeleton width="70%" height="16px" />
                <Skeleton width="50%" height="12px" style="margin-top: 8px;" />
              </div>
            </div>
          </div>
          <div v-else class="note-list">
            <div v-for="note in sharedNotes" :key="note.id" class="note-item shared">
              <div class="note-header">
                <div class="note-title-line">
                  <Share2 :size="16" class="note-type-icon icon-info" />
                  <h4>{{ note.nodeName }}</h4>
                </div>
                <span class="views"><Activity :size="12" class="icon-success" /> {{ note.viewCount }}</span>
              </div>
              <div class="note-meta">
                <span>Token: {{ note.shareToken }}</span>
                <span>至: {{ formatDate(note.expiresAt) }}</span>
              </div>
            </div>
            <div v-if="!sharedNotes.length" class="empty-state">
              暂无分享笔记
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 重命名弹窗 -->
    <Teleport to="body">
      <div v-if="showRenameModal" class="modal-overlay" @click.self="showRenameModal = false">
        <div class="modal-content small-modal">
          <header class="modal-header">
            <h3>重命名</h3>
            <button class="close-btn" @click="showRenameModal = false">
              <X :size="20" />
            </button>
          </header>
          <div class="modal-body">
            <div class="form-group">
              <label>新名称</label>
              <input v-model="renameTitle" type="text" placeholder="请输入新名称" maxlength="255" />
            </div>
          </div>
          <footer class="modal-footer">
            <div class="form-actions">
              <button class="cancel-btn" @click="showRenameModal = false">取消</button>
              <button class="confirm-btn" :disabled="renaming" @click="handleRename">
                <Loader2 v-if="renaming" class="spin" :size="16" />
                <span>确认</span>
              </button>
            </div>
          </footer>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
:root {
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
}

.profile-page {
  width: 100vw;
  height: 100vh;
  background: var(--notion-white);
  color: var(--notion-black);
  font-family: "NotionInter", Inter, -apple-system, system-ui, sans-serif;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.topbar {
  height: 64px;
  flex-shrink: 0;
  padding: 0 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: var(--whisper-border);
  background: var(--notion-white);
}

.topbar-actions {
  display: flex;
  gap: 12px;
}

.ghost-btn,
.danger-btn,
.primary-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border-radius: 4px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.1s ease;
  border: 1px solid transparent;
}

.ghost-btn {
  background: transparent;
  color: var(--notion-black);
}

.ghost-btn:hover:not(:disabled) {
  background: rgba(0,0,0,0.05);
}

.ghost-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.primary-btn {
  background: var(--notion-blue);
  color: white;
}

.primary-btn:hover:not(:disabled) {
  background: var(--notion-blue-hover);
}

.danger-btn {
  color: #eb5757;
  background: transparent;
}

.danger-btn:hover {
  background: rgba(235, 87, 87, 0.1);
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  100% { transform: rotate(360deg); }
}

.content-wrapper {
  flex: 1;
  padding: 40px 24px;
  display: flex;
  flex-direction: row;
  justify-content: center;
  gap: 48px;
  overflow-y: auto;
  width: 100%;
}

.side-nav {
  position: sticky;
  top: 0;
  height: fit-content;
  width: 180px;
  flex-shrink: 0;
  padding-top: 8px;
}

.nav-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  border-radius: 6px;
  color: var(--warm-gray-500);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  user-select: none;
}

.nav-item:hover {
  background: var(--warm-white);
  color: var(--notion-black);
}

.nav-item:active {
  background: rgba(0,0,0,0.1);
}

.nav-item span {
  white-space: nowrap;
}

@media (max-width: 1024px) {
  .side-nav {
    display: none;
  }
  .content-wrapper {
    padding: 24px 16px;
  }
}

@media (max-width: 640px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

.main-column {
  display: flex;
  flex-direction: column;
  gap: 32px;
  width: 100%;
  max-width: 900px;
}

.card {
  border-radius: 12px;
  border: var(--whisper-border);
  background: var(--notion-white);
  padding: 32px;
  box-shadow: var(--card-shadow);
  transition: box-shadow 0.2s ease;
}

.card:hover {
  box-shadow: rgba(0,0,0,0.06) 0px 6px 22px, rgba(0,0,0,0.04) 0px 3px 10px;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 24px;
}

.clickable-header {
  cursor: pointer;
  transition: background 0.1s ease;
  padding: 6px 12px;
  margin-left: -12px;
  border-radius: 6px;
}

.clickable-header:hover {
  background: var(--warm-white);
}

.header-arrow {
  margin-left: auto;
  color: var(--warm-gray-300);
  transition: transform 0.2s ease;
}

.clickable-header:hover .header-arrow {
  transform: translateX(4px);
  color: var(--notion-blue);
}

.card-header h2 {
  font-size: 20px;
  font-weight: 700;
  margin: 0;
  color: var(--notion-black);
  letter-spacing: -0.25px;
}

.icon-primary { color: var(--notion-blue); }
.icon-success { color: #1aae39; }
.icon-warning { color: #dd5b00; }
.icon-danger { color: #eb5757; }
.icon-info { color: #0075de; }
.icon-teal { color: #2a9d99; }
.icon-pink { color: #ff64c8; }
.icon-purple { color: #391c57; }
.icon-brown { color: #523410; }

/* Hero Section */
.hero, .hero-skeleton {
  display: flex;
  gap: 28px;
  align-items: center;
  padding-bottom: 32px;
  border-bottom: 1px solid rgba(0,0,0,0.05);
}

.avatar {
  width: 96px;
  height: 96px;
  border-radius: 20px;
  background: var(--warm-white);
  border: var(--whisper-border);
  display: grid;
  place-items: center;
  overflow: hidden;
  color: var(--notion-black);
  font-weight: 700;
  font-size: 40px;
}

.avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.name-row {
  display: flex;
  align-items: baseline;
  gap: 12px;
}

.username {
  margin: 0;
  font-size: 40px;
  font-weight: 700;
  color: var(--notion-black);
  letter-spacing: -1.5px;
  line-height: 1.1;
}

.subtitle {
  margin: 8px 0 0;
  color: var(--warm-gray-500);
  font-size: 18px;
  font-weight: 400;
  line-height: 1.5;
}

/* Stats */
.stats-grid, .stats-skeleton {
  padding-top: 32px;
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
}

.stat-item {
  padding: 16px;
  border-radius: 12px;
  background: var(--warm-white);
  border: 1px solid transparent;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  transition: all 0.2s ease;
}

.stat-item:hover {
  background: var(--notion-white);
  border-color: rgba(0,0,0,0.1);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0,0,0,0.05);
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
  color: var(--notion-black);
  letter-spacing: -0.5px;
}

.stat-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--warm-gray-500);
  font-weight: 600;
}

/* List Common */
.empty-state {
  text-align: center;
  padding: 48px;
  color: var(--warm-gray-300);
  font-size: 15px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

/* Subject List */
.subject-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.subject-item {
  padding: 16px;
  border-radius: 12px;
  background: var(--notion-white);
  border: var(--whisper-border);
  cursor: pointer;
  transition: all 0.2s ease;
}

.subject-item:hover {
  background: var(--warm-white);
  transform: translateX(4px);
}

.subject-info {
  display: flex;
  gap: 16px;
  align-items: center;
}

.subject-cover {
  width: 48px;
  height: 64px;
  border-radius: 6px;
  border: var(--whisper-border);
  display: grid;
  place-items: center;
  color: white;
  flex-shrink: 0;
}

.subject-text {
  flex: 1;
}

.subject-text h3 {
  margin: 0 0 6px;
  font-size: 16px;
  font-weight: 700;
  color: var(--notion-black);
  letter-spacing: -0.2px;
}

.subject-meta {
  display: flex;
  gap: 16px;
  font-size: 13px;
  color: var(--warm-gray-500);
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

/* Note List */
.note-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.note-item {
  padding: 12px;
  border-radius: 8px;
  background: transparent;
  display: flex;
  align-items: center;
  justify-content: space-between;
  transition: all 0.1s ease;
  cursor: pointer;
}

.note-item:hover {
  background: var(--warm-white);
}

.note-main-info {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.note-title-line {
  display: flex;
  align-items: center;
  gap: 12px;
}

.note-type-icon {
  color: var(--warm-gray-500);
}

.note-type-icon.folder { color: #f2994a; }
.note-type-icon.file { color: #2d9cdb; }

.note-title-line h4 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--notion-black);
}

.note-title-line .date {
  font-size: 12px;
  color: var(--warm-gray-300);
  margin-left: auto;
  margin-right: 12px;
}

.edit-icon-mini {
  color: var(--warm-gray-300);
  opacity: 0;
  transition: all 0.2s ease;
}

.edit-icon-mini:hover {
  color: var(--notion-blue);
  transform: scale(1.2);
}

.note-item:hover .edit-icon-mini {
  opacity: 1;
}

.note-item-actions {
  display: flex;
  gap: 8px;
}

.toggle-public-mini {
  background: transparent;
  border: none;
  color: var(--warm-gray-300);
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  transition: all 0.2s ease;
}

.toggle-public-mini:hover {
  background: rgba(0,0,0,0.05);
  color: var(--notion-black);
}

.toggle-public-mini.isPublic {
  color: var(--notion-blue);
}

.note-ellipsis h4 {
  color: var(--warm-gray-300);
}

/* Shared Notes */
.note-item.shared {
  flex-direction: column;
  align-items: stretch;
  gap: 8px;
  cursor: default;
}

.note-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.note-header h4 {
  margin: 0;
  font-size: 14px;
  font-weight: 700;
}

.views {
  font-size: 12px;
  color: var(--warm-gray-500);
  display: flex;
  align-items: center;
  gap: 4px;
}

.note-meta {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: var(--warm-gray-300);
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

.modal-content {
  background: var(--notion-white);
  border-radius: 12px;
  box-shadow: var(--deep-shadow);
  border: var(--whisper-border);
  width: 90%;
  max-width: 400px;
  overflow: hidden;
}

.modal-header {
  padding: 16px 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: var(--whisper-border);
}

.modal-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 700;
}

.close-btn {
  background: transparent;
  border: none;
  color: var(--warm-gray-300);
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
}

.close-btn:hover {
  background: var(--warm-white);
  color: var(--notion-black);
}

.modal-body {
  padding: 24px 20px;
}

.form-group label {
  display: block;
  font-size: 13px;
  font-weight: 600;
  color: var(--warm-gray-500);
  margin-bottom: 8px;
}

.form-group input {
  width: 100%;
  padding: 10px 12px;
  border-radius: 6px;
  border: 1px solid #ddd;
  font-size: 14px;
  color: var(--notion-black);
  transition: all 0.2s ease;
}

.form-group input:focus {
  outline: none;
  border-color: var(--notion-blue);
  box-shadow: 0 0 0 3px rgba(0,117,222,0.1);
}

.modal-footer {
  padding: 16px 20px;
  background: var(--warm-white);
  display: flex;
  justify-content: flex-end;
}

.form-actions {
  display: flex;
  gap: 12px;
}

.confirm-btn {
  background: var(--notion-blue);
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 6px;
  font-weight: 600;
  cursor: pointer;
}

.confirm-btn:hover:not(:disabled) {
  background: var(--notion-blue-hover);
}

.cancel-btn {
  background: white;
  color: var(--notion-black);
  border: var(--whisper-border);
  padding: 8px 16px;
  border-radius: 6px;
  font-weight: 600;
  cursor: pointer;
}

@media (max-width: 640px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
