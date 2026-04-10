<script setup lang="ts">
import { computed, reactive } from 'vue'
import { useRouter } from 'vue-router'
import Toast from '../components/Toast.vue'
import Skeleton from '../components/Skeleton.vue'
import { 
  ArrowLeft, LogOut, RefreshCcw, 
  Activity, BookOpen, Share2, Book, Users, Star, Layers 
} from 'lucide-vue-next'
import { useUserProfile } from '../composables/useUserProfile'
import ActivityCalendar from '../components/ActivityCalendar.vue'

const router = useRouter()
const {
  userInfo,
  activities,
  publicPrivateNotes,
  sharedNotes,
  learnedSubjects,
  loadingUserInfo,
  loadingActivities,
  loadingPublicPrivateNotes,
  loadingSharedNotes,
  loadingLearnedSubjects,
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
        <button class="ghost-btn" :disabled="isGlobalLoading" @click="handleRefresh">
          <RefreshCcw :size="18" :class="{ 'spin': isGlobalLoading }" />
          刷新
        </button>
        <button class="danger-btn" @click="handleLogout">
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
            <h2>在学/已学教材</h2>
          </div>
          
          <div v-if="loadingLearnedSubjects" class="list-skeleton">
            <div v-for="i in 2" :key="i" class="list-item-skeleton">
              <Skeleton width="48px" height="64px" border-radius="6px" />
              <div class="skeleton-col">
                <Skeleton width="60%" height="16px" />
                <Skeleton width="100%" height="8px" style="margin-top: 10px;" />
              </div>
            </div>
          </div>
          <div v-else class="subject-list">
            <div v-for="subject in learnedSubjects" :key="subject.subjectId" class="subject-item">
              <div class="subject-cover">
                <img v-if="subject.coverImage" :src="subject.coverImage" :alt="subject.subjectName" />
                <Book v-else :size="24" color="#94a3b8" />
              </div>
              <div class="subject-info">
                <h4>{{ subject.subjectName }}</h4>
                <div class="progress-bar-wrap">
                  <div class="progress-bar">
                    <div 
                      class="progress-inner" 
                      :style="{ width: `${Math.min(100, (subject.learned / subject.total) * 100)}%` }"
                    ></div>
                  </div>
                  <span class="progress-text">{{ subject.learned }}/{{ subject.total }}</span>
                </div>
              </div>
            </div>
            <div v-if="!learnedSubjects.length" class="empty-state">
              暂无学习数据
            </div>
          </div>
        </div>
      </div>

      <div class="side-column">
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
            <div v-for="note in publicPrivateNotes" :key="note.id" class="note-item">
              <h4>{{ note.title }}</h4>
              <span class="date">{{ formatDate(note.updatedAt) }}</span>
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
            <h2>已分享笔记</h2>
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

.subject-info h4 {
  margin: 0 0 10px 0;
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
  padding: 12px;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.5);
  border: 1px solid rgba(148, 163, 184, 0.15);
  transition: background 0.2s;
}

.note-item:hover {
  background: rgba(255, 255, 255, 0.8);
}

.note-item h4 {
  margin: 0 0 6px 0;
  font-size: 14px;
  color: #1e293b;
  line-height: 1.4;
}

.note-item .date {
  font-size: 12px;
  color: #94a3b8;
}

.note-item.shared {
  display: flex;
  flex-direction: column;
  gap: 8px;
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
