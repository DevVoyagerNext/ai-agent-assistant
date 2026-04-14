<script setup lang="ts">
import { computed, reactive, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import Toast from '../components/Toast.vue'
import Skeleton from '../components/Skeleton.vue'
import { 
  ArrowLeft, LogOut, RefreshCcw, 
  Activity, BookOpen, Share2, Book, Users, Star, Layers, FolderHeart, Bookmark,
  Folder, FileText, ToggleRight, ToggleLeft, Edit3, X, Loader2
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
      <div class="main-column">
        <!-- User Info Card -->
        <div class="card user-card">
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
              <span class="stat-label"><Users :size="14" /> 粉丝</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ userInfo?.followingCount || 0 }}</span>
              <span class="stat-label"><Star :size="14" /> 关注</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ userInfo?.learnedSubjectsCount || 0 }}</span>
              <span class="stat-label"><Book :size="14" /> 已学课程</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ userInfo?.sharedNotesCount || 0 }}</span>
              <span class="stat-label"><Share2 :size="14" /> 分享笔记</span>
            </div>
          </div>
        </div>

        <!-- Activity Calendar Card -->
        <div class="card">
          <div class="card-header">
            <Activity :size="20" class="icon-primary" />
            <h2>活跃度日历</h2>
          </div>
          
          <div v-if="loadingActivities" class="activity-skeleton">
            <Skeleton v-for="i in 14" :key="i" width="100%" height="30px" border-radius="4px" />
          </div>
          <ActivityCalendar v-else :items="activities" />
        </div>

        <!-- Learned Subjects Card -->
        <div class="card">
          <div class="card-header">
            <Layers :size="20" class="icon-success" />
            <h2>最近学习 (在学/已学)</h2>
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
                      <Clock :size="12" />
                      {{ item.lastStudyTime ? formatDate(item.lastStudyTime) : '未知时间' }}
                    </span>
                    <span class="meta-item">
                      <Activity :size="12" />
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
      </div>

      <div class="side-column">
        <!-- Liked Subjects -->
        <div class="card">
          <div class="card-header">
            <Star :size="20" class="icon-primary" />
            <h2>点赞的教材</h2>
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
              <h4>{{ subject.name }}</h4>
              <span v-if="subject.progressPercent > 0" class="date">进度 {{ subject.progressPercent }}%</span>
            </div>
            <div v-if="!likedSubjects.length" class="empty-state">
              暂无点赞教材
            </div>
          </div>
        </div>

        <!-- Collect Folders -->
        <div class="card">
          <div class="card-header">
            <FolderHeart :size="20" class="icon-danger" />
            <h2>我的收藏夹</h2>
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
              <h4>{{ folder.name }}</h4>
              <span class="date">{{ folder.isPublic ? '公开' : '私密' }}</span>
            </div>
            <div v-if="!collectFolders.length" class="empty-state">
              暂无收藏夹
            </div>
          </div>
        </div>
        <!-- Private Notes -->
        <div class="card">
          <div class="card-header">
            <BookOpen :size="20" class="icon-warning" />
            <h2>私人笔记</h2>
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
                  <Folder v-if="note.type === 'folder'" :size="16" class="note-type-icon folder" />
                  <FileText v-else :size="16" class="note-type-icon file" />
                  <h4 @click="openRename(note)">{{ note.title }}</h4>
                  <Edit3 :size="12" class="edit-icon-mini" @click="openRename(note)" />
                </div>
                <span class="date">{{ formatDate(note.updatedAt) }}</span>
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
        <div class="card">
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
                <h4>{{ note.nodeName }}</h4>
                <span class="views"><Activity :size="12" /> {{ note.viewCount }}</span>
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
.profile-page {
  width: 100vw;
  height: 100vh;
  background:
    radial-gradient(1200px 600px at 20% 10%, rgba(59, 130, 246, 0.18), transparent 60%),
    radial-gradient(900px 500px at 80% 30%, rgba(34, 197, 94, 0.16), transparent 55%),
    linear-gradient(180deg, #f8fafc, #eef2ff);
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

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  100% { transform: rotate(360deg); }
}

.content-wrapper {
  flex: 1;
  padding: 18px 20px 24px;
  display: grid;
  grid-template-columns: 1fr 360px;
  gap: 20px;
  align-items: start;
  overflow-y: auto;
}

.main-column,
.side-column {
  display: flex;
  flex-direction: column;
  gap: 20px;
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

.card-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 18px;
}

.card-header h2 {
  font-size: 18px;
  margin: 0;
  color: #0f172a;
}

.icon-primary { color: #3b82f6; }
.icon-success { color: #22c55e; }
.icon-warning { color: #f59e0b; }
.icon-info { color: #0ea5e9; }

/* Hero Section */
.hero, .hero-skeleton {
  display: flex;
  gap: 16px;
  align-items: center;
  padding-bottom: 18px;
  border-bottom: 1px dashed rgba(148, 163, 184, 0.35);
}

.title-skeleton {
  display: flex;
  flex-direction: column;
  justify-content: center;
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

.subtitle {
  margin: 8px 0 0;
  color: #64748b;
  font-size: 13px;
}

/* Stats */
.stats-grid, .stats-skeleton {
  padding-top: 18px;
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 14px;
}

.stat-item {
  padding: 16px 12px;
  border-radius: 16px;
  border: 1px solid rgba(148, 163, 184, 0.25);
  background: rgba(255, 255, 255, 0.6);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
}

.stat-value {
  font-size: 20px;
  font-weight: 700;
  color: #0f172a;
}

.stat-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #64748b;
  font-weight: 500;
}

/* Activity Grid */
.activity-grid, .activity-skeleton {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(24px, 1fr));
  gap: 6px;
}

.activity-box {
  width: 100%;
  aspect-ratio: 1;
  border-radius: 4px;
  background: rgba(148, 163, 184, 0.15);
  transition: transform 0.2s;
}

.activity-box:hover {
  transform: scale(1.1);
}

.low-activity { background: rgba(34, 197, 94, 0.3); }
.mid-activity { background: rgba(34, 197, 94, 0.6); }
.high-activity { background: rgba(34, 197, 94, 1); }

/* List Common */
.list-skeleton {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.list-item-skeleton {
  display: flex;
  gap: 12px;
  align-items: center;
}

.skeleton-col {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.empty-state {
  text-align: center;
  padding: 24px;
  color: #94a3b8;
  font-size: 13px;
}

/* Subject List */
.subject-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.subject-item {
  display: flex;
  gap: 16px;
  align-items: center;
  padding: 12px;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.4);
  border: 1px solid rgba(148, 163, 184, 0.15);
}

.subject-cover {
  width: 48px;
  height: 64px;
  border-radius: 6px;
  background: #f1f5f9;
  display: grid;
  place-items: center;
  overflow: hidden;
}

.subject-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.subject-info {
  flex: 1;
}

.subject-title-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.subject-status-icons {
  display: flex;
  gap: 8px;
  color: #f59e0b;
}

.subject-info h4 {
  margin: 0;
  font-size: 15px;
  color: #1e293b;
}

.progress-bar-wrap {
  display: flex;
  align-items: center;
  gap: 10px;
}

.progress-bar {
  flex: 1;
  height: 6px;
  background: rgba(148, 163, 184, 0.2);
  border-radius: 999px;
  overflow: hidden;
}

.progress-inner {
  height: 100%;
  background: #3b82f6;
  border-radius: 999px;
  transition: width 0.3s ease;
}

.progress-text {
  font-size: 12px;
  color: #64748b;
  min-width: 40px;
}

/* Note List */
.note-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.note-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.5);
  border: 1px solid rgba(148, 163, 184, 0.15);
  transition: all 0.2s;
}

.note-item:hover {
  background: rgba(255, 255, 255, 0.8);
  border-color: rgba(59, 130, 246, 0.3);
  transform: translateX(2px);
}

.note-main-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.note-title-line {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
  min-width: 0;
}

.note-title-line h4 {
  margin: 0;
  font-size: 14px;
  color: #1e293b;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  cursor: pointer;
}

.note-title-line h4:hover {
  color: #3b82f6;
  text-decoration: underline;
}

.edit-icon-mini {
  color: #94a3b8;
  opacity: 0;
  cursor: pointer;
  transition: all 0.2s;
}

.note-item:hover .edit-icon-mini {
  opacity: 1;
}

.edit-icon-mini:hover {
  color: #3b82f6;
}

.note-type-icon {
  flex-shrink: 0;
}

.note-type-icon.folder {
  color: #f59e0b;
}

.note-type-icon.file {
  color: #3b82f6;
}

.note-item .date {
  font-size: 11px;
  color: #94a3b8;
}

.note-item-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.toggle-public-mini {
  background: transparent;
  border: none;
  color: #94a3b8;
  cursor: pointer;
  display: flex;
  align-items: center;
  padding: 4px;
  border-radius: 4px;
  transition: all 0.2s;
}

.toggle-public-mini.isPublic {
  color: #3b82f6;
}

.toggle-public-mini:hover {
  background: rgba(148, 163, 184, 0.1);
}

.note-item.shared {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.note-item.note-ellipsis {
  justify-content: center;
  opacity: 0.75;
}

.note-item.note-ellipsis h4 {
  margin: 0;
}

.note-item.note-ellipsis:hover {
  background: rgba(255, 255, 255, 0.5);
  transform: none;
}

.note-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.views {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  color: #64748b;
  background: rgba(148, 163, 184, 0.1);
  padding: 2px 6px;
  border-radius: 4px;
}

.note-meta {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: #64748b;
}

/* Modal Styles */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(4px);
}

.small-modal {
  width: 380px;
  background: #fff;
  border-radius: 16px;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
}

.modal-header {
  padding: 16px 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid #f1f5f9;
}

.modal-header h3 {
  margin: 0;
  font-size: 16px;
  color: #1e293b;
}

.close-btn {
  background: transparent;
  border: none;
  color: #94a3b8;
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
}

.close-btn:hover {
  background: #f1f5f9;
}

.modal-body {
  padding: 20px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-group label {
  font-size: 13px;
  font-weight: 600;
  color: #64748b;
}

.form-group input {
  padding: 10px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  outline: none;
  transition: all 0.2s;
}

.form-group input:focus {
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.modal-footer {
  padding: 16px 20px;
  border-top: 1px solid #f1f5f9;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.cancel-btn {
  padding: 8px 16px;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
  background: #fff;
  color: #64748b;
  cursor: pointer;
  font-weight: 500;
}

.confirm-btn {
  padding: 8px 16px;
  border-radius: 8px;
  border: none;
  background: #3b82f6;
  color: #fff;
  cursor: pointer;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 6px;
}

.confirm-btn:hover {
  background: #2563eb;
}

.confirm-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

@media (max-width: 1024px) {
  .content-wrapper {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
