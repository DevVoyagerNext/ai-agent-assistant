<script setup lang="ts">
import { computed, reactive, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import Toast from '../components/Toast.vue'
import Skeleton from '../components/Skeleton.vue'
import { 
  ArrowLeft, LogOut, RefreshCcw, 
  Activity, BookOpen, Share2, Book, Users, Star, Layers, FolderHeart,
  Folder, FileText, ToggleRight, ToggleLeft, Edit3, X, Loader2, ChevronRight, ChevronLeft, Clock, Save, FolderPlus, FilePlus, Trash2, Bot
} from 'lucide-vue-next'
import { useUserProfile } from '../composables/useUserProfile'
import ActivityCalendar from '../components/ActivityCalendar.vue'
import { 
  updatePrivateNoteTitle, updatePrivateNotePublic, 
  updateCollectFolderPublic, updateCollectFolderName,
  getSubjectsInFolder, getPrivateNoteDetail, updatePrivateNoteContent,
  createPrivateNote, sharePrivateNote, updateShareNoteStatus, updateShareNoteExpire, deleteSharedNote
} from '../api/user'

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
  createdSubjects,
  
  loadingUserInfo,
  loadingActivities,
  loadingPublicPrivateNotes,
  loadingSharedNotes,
  loadingLearnedSubjects,
  loadingCollectFolders,
  loadingLikedSubjects,
  loadingRecentSubjects,
  loadingCreatedSubjects,
  
  errorUserInfo,
  refreshAll
} = useUserProfile()

// --- 用户创建的教材分类 ---
const publishedNoDraftSubjects = computed(() => 
  createdSubjects.value.filter(s => s.status === 'published' && s.hasDraft === 0)
)

const publishedWithDraftSubjects = computed(() => 
  createdSubjects.value.filter(s => s.status === 'published' && s.hasDraft === 1)
)

const draftSubjects = computed(() => 
  createdSubjects.value.filter(s => s.status === 'draft')
)

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
const pendingRenameFolder = ref<any>(null)

const openRename = (note: any) => {
  pendingRenameNote.value = note
  pendingRenameFolder.value = null
  renameTitle.value = note.title
  showRenameModal.value = true
}

const openRenameFolder = (folder: any) => {
  pendingRenameFolder.value = folder
  pendingRenameNote.value = null
  renameTitle.value = folder.name
  showRenameModal.value = true
}

const handleRename = async () => {
  if (!renameTitle.value.trim()) return
  renaming.value = true
  try {
    if (pendingRenameNote.value) {
      const res = await updatePrivateNoteTitle(pendingRenameNote.value.id, renameTitle.value)
      if (res.data?.code === 200) {
        pendingRenameNote.value.title = renameTitle.value
        showToast('重命名成功', 'success')
        showRenameModal.value = false
      } else {
        showToast(res.data?.msg || '重命名失败', 'error')
      }
    } else if (pendingRenameFolder.value) {
      const res = await updateCollectFolderName(pendingRenameFolder.value.id, renameTitle.value)
      if (res.data?.code === 200) {
        pendingRenameFolder.value.name = renameTitle.value
        showToast('重命名成功', 'success')
        showRenameModal.value = false
      } else {
        showToast(res.data?.msg || '重命名失败', 'error')
      }
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
    } else {
      showToast(res.data?.msg || '修改失败', 'error')
    }
  } catch (err: any) {
    showToast(err.response?.data?.msg || '修改失败', 'error')
  }
}

const updatingFolderIds = ref<Set<number>>(new Set())

// --- 收藏夹教材列表弹窗相关 ---
const showFolderModal = ref(false)
const selectedFolder = ref<any>(null)
const folderSubjects = ref<any[]>([])
const loadingFolderSubjects = ref(false)
const folderPage = ref(1)
const folderTotal = ref(0)
const pageSize = ref(20)
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
      if (!isLoadMore) folderSubjects.value = []
      folderTotal.value = 0
    }
  } catch (err: any) {
    console.error('Fetch folder subjects error:', err)
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

const handleFolderSubjectClick = (item: any) => {
  const url = `/subject/${item.id}` + (item.lastNodeId ? `?nodeId=${item.lastNodeId}` : '')
  router.push(url)
}

// --- 私人笔记内容弹窗相关 ---
const showPrivateNoteModal = ref(false)
const loadingNoteDetail = ref(false)
const noteDetail = ref<any>(null) // PrivateNoteResponse
const notePathStack = ref<any[]>([]) // 用于返回上一级，存储 {id, title}

const openPrivateNoteModal = async (note: any) => {
  notePathStack.value = []
  showPrivateNoteModal.value = true
  await fetchPrivateNoteDetail(note.id, note.title)
}

const fetchPrivateNoteDetail = async (id: number, title: string) => {
  loadingNoteDetail.value = true
  try {
    const res = await getPrivateNoteDetail(id, 2) // scope 2 = 全部
    if (res.data?.code === 200) {
      noteDetail.value = res.data.data
      // 初始化编辑内容
      if (noteDetail.value.type === 'markdown' && noteDetail.value.content) {
        editNoteContent.value = noteDetail.value.content.content
        syncStatus.value = 'saved'
      }
      
      // 只有在成功获取到数据后才推入栈（如果是从列表点击或者点击子项）
      // 避免重复推入
      const isAlreadyInStack = notePathStack.value.some(p => p.id === id)
      if (!isAlreadyInStack) {
        notePathStack.value.push({ id, title })
      }
    }
  } catch (err: any) {
    console.error('Fetch private note error:', err)
    showToast('获取笔记内容失败', 'error')
  } finally {
    loadingNoteDetail.value = false
  }
}

const handleNoteBack = async () => {
  if (notePathStack.value.length <= 1) {
    showPrivateNoteModal.value = false
    return
  }
  notePathStack.value.pop() // 弹出当前
  const prev = notePathStack.value[notePathStack.value.length - 1]
  await fetchPrivateNoteDetail(prev.id, prev.title)
}

const handleNoteChildClick = async (child: any) => {
  await fetchPrivateNoteDetail(child.id, child.title)
}

// --- 私人笔记创建相关 ---
const showCreateNoteModal = ref(false)
const creatingNote = ref(false)
const createNoteForm = reactive({
  title: '',
  type: 'folder' as 'folder' | 'markdown',
  isPublic: 0 as 0 | 1
})

const openCreateNote = (type: 'folder' | 'markdown') => {
  createNoteForm.type = type
  createNoteForm.title = ''
  createNoteForm.isPublic = 0
  showCreateNoteModal.value = true
}

const handleCreateNote = async () => {
  if (!createNoteForm.title.trim()) return
  
  const currentFolderId = notePathStack.value.length > 0 
    ? notePathStack.value[notePathStack.value.length - 1].id 
    : 0

  creatingNote.value = true
  try {
    const res = await createPrivateNote({
      parentId: currentFolderId,
      type: createNoteForm.type,
      title: createNoteForm.title,
      isPublic: createNoteForm.isPublic
    })
    
    if (res.data?.code === 200) {
      showToast('创建成功', 'success')
      showCreateNoteModal.value = false
      // 刷新当前文件夹内容或根列表
      if (currentFolderId === 0) {
        await handleRefresh() // 刷新主页面
      } else {
        const currentFolder = notePathStack.value[notePathStack.value.length - 1]
        await fetchPrivateNoteDetail(currentFolder.id, currentFolder.title)
      }
    } else {
      showToast(res.data?.msg || '创建失败', 'error')
    }
  } catch (err: any) {
    showToast(err.response?.data?.msg || '创建失败', 'error')
  } finally {
    creatingNote.value = false
  }
}

// --- 私人笔记内容编辑与同步相关 ---
const editNoteContent = ref('')
const savingNoteContent = ref(false)
const syncStatus = ref<'saved' | 'saving' | 'error'>('saved')
let debounceTimer: any = null

const handleNoteContentInput = () => {
  syncStatus.value = 'saving'
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    handleAutoSaveNoteContent()
  }, 1000) // 1秒防抖
}

const handleAutoSaveNoteContent = async () => {
  if (!noteDetail.value?.content?.id) return
  savingNoteContent.value = true
  try {
    const res = await updatePrivateNoteContent(noteDetail.value.content.id, editNoteContent.value)
    if (res.data?.code === 200) {
      noteDetail.value.content.content = editNoteContent.value
      syncStatus.value = 'saved'
    } else {
      syncStatus.value = 'error'
      showToast(res.data?.msg || '自动保存失败', 'error')
    }
  } catch (err: any) {
    syncStatus.value = 'error'
    console.error('Auto save error:', err)
  } finally {
    savingNoteContent.value = false
  }
}

const handleToggleNotePublicInModal = async (customNote?: any) => {
  const targetNote = customNote || noteDetail.value?.content || noteDetail.value
  if (!targetNote?.id) return
  
  const currentPublic = !!targetNote.isPublic
  const newPublic: 0 | 1 = currentPublic ? 0 : 1
  
  try {
    const res = await updatePrivateNotePublic(targetNote.id, newPublic)
    if (res.data?.code === 200) {
      targetNote.isPublic = newPublic
      // 同时更新外层列表中的状态（如果匹配）
      if (pendingRenameNote.value && pendingRenameNote.value.id === targetNote.id) {
        pendingRenameNote.value.isPublic = newPublic
      }
      showToast(newPublic === 1 ? '已设为公开' : '已设为私密', 'success')
    } else {
      showToast(res.data?.msg || '修改失败', 'error')
    }
  } catch (err: any) {
    showToast(err.response?.data?.msg || '修改失败', 'error')
  }
}

// --- 私人笔记分享相关 ---
const showShareModal = ref(false)
const sharing = ref(false)
const pendingShareNote = ref<any>(null)
const shareExpiresAt = ref('')
const shareResult = ref<{ shareToken: string, shareCode: string, expiresAt: string } | null>(null)
const windowOrigin = computed(() => window.location.origin)

const openShareModal = (note: any) => {
  pendingShareNote.value = note
  shareExpiresAt.value = new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString().slice(0, 16) // 默认7天后
  shareResult.value = null
  showShareModal.value = true
}

const handleShare = async () => {
  if (!pendingShareNote.value?.id || !shareExpiresAt.value) return
  sharing.value = true
  try {
    // 转换日期格式为 API 要求的 "YYYY-MM-DD HH:mm:ss"
    const formattedDate = shareExpiresAt.value.replace('T', ' ') + ':00'
    const res = await sharePrivateNote(pendingShareNote.value.id, formattedDate)
    if (res.data?.code === 200 && res.data.data) {
      shareResult.value = res.data.data
      showToast('分享链接已生成', 'success')
      // 刷新分享列表
      refreshAll()
    } else {
      showToast(res.data?.msg || '分享失败', 'error')
    }
  } catch (err: any) {
    showToast(err.response?.data?.msg || '分享失败', 'error')
  } finally {
    sharing.value = false
  }
}

const copyToClipboard = (text: string | undefined) => {
  if (!text) return
  navigator.clipboard.writeText(text).then(() => {
    showToast('已复制到剪贴板', 'success')
  }).catch(() => {
    showToast('复制失败', 'error')
  })
}

const handleToggleShareStatus = async (shareId: number, isActive: 0 | 1) => {
  const actionText = isActive === 1 ? '重新分享' : '取消分享'
  if (isActive === 0 && !confirm('确定要取消该分享吗？取消后已有链接将失效。')) return
  
  try {
    const res = await updateShareNoteStatus(shareId, isActive)
    if (res.data?.code === 200) {
      showToast(`${actionText}成功`, 'success')
      // 刷新分享列表
      refreshAll()
    } else {
      showToast(res.data?.msg || `${actionText}失败`, 'error')
    }
  } catch (err: any) {
    showToast(err.response?.data?.msg || `${actionText}失败`, 'error')
  }
}

const handleDeleteSharedNote = async (shareId: number) => {
  if (!confirm('确定要删除该分享记录吗？删除后不可恢复。')) return
  
  try {
    const res = await deleteSharedNote(shareId)
    if (res.data?.code === 200) {
      showToast('删除分享成功', 'success')
      refreshAll()
    } else {
      showToast(res.data?.msg || '删除分享失败', 'error')
    }
  } catch (err: any) {
    showToast(err.response?.data?.msg || '删除分享失败', 'error')
  }
}

// --- 修改分享时间相关 ---
const showUpdateExpireModal = ref(false)
const updatingExpire = ref(false)
const pendingUpdateShareId = ref<number | null>(null)
const updateExpireMode = ref<'minutes' | 'date'>('minutes')
const selectedMinutes = ref<number | 'custom'>(43200) // 默认 1个月，或者 'custom'
const customDurationValue = ref<number>(1)
const customDurationUnit = ref<number>(1440) // 默认单位为 天 (1440分钟)
const selectedExpireDate = ref('')

const quickTimeOptions: { label: string; value: number | 'custom' }[] = [
  { label: '1 小时', value: 60 },
  { label: '1 天', value: 1440 },
  { label: '7 天', value: 10080 },
  { label: '30 天', value: 43200 },
  { label: '自定义', value: 'custom' }
]

const customUnitOptions = [
  { label: '分钟', value: 1 },
  { label: '小时', value: 60 },
  { label: '天', value: 1440 },
  { label: '个月(30天)', value: 43200 }
]

const openUpdateExpireModal = (shareId: number) => {
  pendingUpdateShareId.value = shareId
  updateExpireMode.value = 'minutes'
  selectedMinutes.value = 43200
  selectedExpireDate.value = new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString().slice(0, 16)
  showUpdateExpireModal.value = true
}

const handleUpdateExpire = async () => {
  if (!pendingUpdateShareId.value) return
  updatingExpire.value = true
  
  try {
    let expireMinutes = 0
    let expireAt = undefined
    
    if (updateExpireMode.value === 'minutes') {
      if (selectedMinutes.value === 'custom') {
        if (!customDurationValue.value || customDurationValue.value <= 0) {
          showToast('请输入有效的自定义时长', 'error')
          updatingExpire.value = false
          return
        }
        expireMinutes = customDurationValue.value * customDurationUnit.value
      } else {
        expireMinutes = selectedMinutes.value as number
      }
    } else {
      // 日期模式
      if (!selectedExpireDate.value) {
        showToast('请选择过期时间', 'error')
        updatingExpire.value = false
        return
      }
      expireAt = selectedExpireDate.value.replace('T', ' ') + ':00'
    }
    
    const res = await updateShareNoteExpire(pendingUpdateShareId.value, expireMinutes, expireAt)
    if (res.data?.code === 200) {
      showToast('修改时间成功', 'success')
      showUpdateExpireModal.value = false
      refreshAll()
    } else {
      showToast(res.data?.msg || '修改时间失败', 'error')
    }
  } catch (err: any) {
    showToast(err.response?.data?.msg || '修改时间失败', 'error')
  } finally {
    updatingExpire.value = false
  }
}

// ------------------------------

const handleToggleCollectionPublic = async (folder: any) => {
  if (!folder?.id) return
  if (updatingFolderIds.value.has(folder.id)) return

  updatingFolderIds.value.add(folder.id)
  updatingFolderIds.value = new Set(updatingFolderIds.value)

  const currentPublic = !!folder.isPublic
  const newPublic: 0 | 1 = currentPublic ? 0 : 1

  try {
    const res = await updateCollectFolderPublic(folder.id, newPublic)
    if (res.data?.code === 200) {
      folder.isPublic = newPublic
      showToast(newPublic === 1 ? '已设为公开' : '已设为私密', 'success')
    } else {
      showToast(res.data?.msg || '修改失败', 'error')
    }
  } catch (err: any) {
    showToast(err.response?.data?.msg || '修改失败', 'error')
  } finally {
    updatingFolderIds.value.delete(folder.id)
    updatingFolderIds.value = new Set(updatingFolderIds.value)
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
        <button class="primary-btn" @click="router.push('/ai-chat')" style="background-color: #8b5cf6;">
          <Bot :size="18" />
          AI 助手
        </button>
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
            <Users :size="18" class="icon-blue" />
            <span>关于我</span>
          </div>
          <div class="nav-item" @click="scrollTo('activity')">
            <Activity :size="18" class="icon-green" />
            <span>活跃度</span>
          </div>
          <div class="nav-item" @click="scrollTo('collections')">
            <FolderHeart :size="18" class="icon-pink" />
            <span>收藏夹</span>
          </div>
          <div class="nav-item" @click="scrollTo('recent')">
            <Layers :size="18" class="icon-teal" />
            <span>最近学习</span>
          </div>
          <div class="nav-item" @click="scrollTo('created')">
            <Book :size="18" class="icon-blue" />
            <span>创建教材</span>
          </div>
          <div class="nav-item" @click="scrollTo('liked')">
            <Star :size="18" class="icon-orange" />
            <span>点赞教材</span>
          </div>
          <div class="nav-item" @click="scrollTo('private')">
            <BookOpen :size="18" class="icon-teal" />
            <span>私人笔记</span>
          </div>
          <div class="nav-item" @click="scrollTo('shared')">
            <Share2 :size="18" class="icon-purple" />
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
              <span class="stat-label"><Users :size="14" class="icon-blue" /> 粉丝</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ userInfo?.followingCount || 0 }}</span>
              <span class="stat-label"><Star :size="14" class="icon-orange" /> 关注</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ userInfo?.learnedSubjectsCount || 0 }}</span>
              <span class="stat-label"><Book :size="14" class="icon-teal" /> 已学课程</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ userInfo?.sharedNotesCount || 0 }}</span>
              <span class="stat-label"><Share2 :size="14" class="icon-purple" /> 分享笔记</span>
            </div>
          </div>
        </div>

        <!-- Activity Card -->
        <div id="activity" class="card">
          <div class="card-header">
            <Activity :size="20" class="icon-green" />
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
            <FolderHeart :size="20" class="icon-pink" />
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
            <div v-for="folder in collectFolders" :key="folder.id" class="note-item" @click="openFolderModal(folder)">
              <div class="note-main-info">
                <div class="note-title-line">
                  <Folder :size="16" class="note-type-icon folder icon-pink" />
                  <h4 @click.stop="openRenameFolder(folder)">{{ folder.name }}</h4>
                  <Edit3 :size="12" class="edit-icon-mini" @click.stop="openRenameFolder(folder)" />
                </div>
              </div>
              <div class="note-item-actions">
                <button 
                  class="toggle-public-mini" 
                  :class="{ isPublic: !!folder.isPublic }"
                  @click.stop="handleToggleCollectionPublic(folder)"
                  :title="!!folder.isPublic ? '已公开' : '已私密'"
                  :disabled="updatingFolderIds.has(folder.id)"
                >
                  <Loader2 v-if="updatingFolderIds.has(folder.id)" class="spin" :size="18" />
                  <ToggleRight v-else-if="!!folder.isPublic" :size="18" />
                  <ToggleLeft v-else :size="18" />
                  <span>{{ !!folder.isPublic ? '公开' : '私密' }}</span>
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
                      <Activity :size="12" class="icon-green" />
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

        <!-- Created Subjects -->
        <div id="created" class="card">
          <div class="card-header">
            <Layers :size="20" class="icon-blue" />
            <h2>创建的教材</h2>
          </div>
          
          <div v-if="loadingCreatedSubjects" class="list-skeleton">
            <div v-for="i in 3" :key="i" class="list-item-skeleton" style="padding: 12px 0;">
              <div class="skeleton-col">
                <Skeleton width="80%" height="16px" />
                <Skeleton width="40%" height="12px" style="margin-top: 8px;" />
              </div>
            </div>
          </div>
          <div v-else class="note-list">
            <!-- 已发布（无草稿） -->
            <div v-if="publishedNoDraftSubjects.length" class="subject-group">
              <h3 class="group-title">已发布</h3>
              <div 
                v-for="subject in publishedNoDraftSubjects" 
                :key="subject.id" 
                class="note-item"
                @click="router.push(`/author/subject/${subject.id}`)"
                style="cursor: pointer;"
              >
                <div class="note-title-line">
                  <Book :size="16" class="note-type-icon icon-teal" />
                  <h4>{{ subject.name }}</h4>
                  <span class="status-badge published">已发布</span>
                  <span class="date">{{ formatDate(subject.createdAt) }}</span>
                </div>
              </div>
            </div>

            <!-- 已发布（有草稿） -->
            <div v-if="publishedWithDraftSubjects.length" class="subject-group">
              <h3 class="group-title">已发布 (有草稿待处理)</h3>
              <div 
                v-for="subject in publishedWithDraftSubjects" 
                :key="subject.id" 
                class="note-item"
                @click="router.push(`/author/subject/${subject.id}`)"
                style="cursor: pointer;"
              >
                <div class="note-title-line">
                  <Book :size="16" class="note-type-icon icon-orange" />
                  <h4>{{ subject.name }}</h4>
                  <span class="status-badge has-draft">有草稿</span>
                  <span class="date">{{ formatDate(subject.createdAt) }}</span>
                </div>
              </div>
            </div>

            <!-- 未发布 -->
            <div v-if="draftSubjects.length" class="subject-group">
              <h3 class="group-title">未发布 (草稿)</h3>
              <div 
                v-for="subject in draftSubjects" 
                :key="subject.id" 
                class="note-item"
                @click="router.push(`/author/subject/${subject.id}`)"
                style="cursor: pointer;"
              >
                <div class="note-title-line">
                  <Book :size="16" class="note-type-icon icon-gray" />
                  <h4>{{ subject.nameDraft || subject.name || '未命名教材' }}</h4>
                  <span class="status-badge draft">草稿</span>
                  <span class="date">{{ formatDate(subject.createdAt) }}</span>
                </div>
              </div>
            </div>

            <div v-if="!createdSubjects.length" class="empty-state">
              暂无创建的教材
            </div>
          </div>
        </div>

        <!-- Liked Subjects -->
        <div id="liked" class="card">
          <div class="card-header clickable-header" @click="router.push('/me/liked-subjects')">
            <Star :size="20" class="icon-orange" />
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
              <div class="note-title-line">
                <Book :size="16" class="note-type-icon icon-teal" />
                <h4>{{ subject.name }}</h4>
                <span v-if="subject.progressPercent > 0" class="date">进度 {{ subject.progressPercent }}%</span>
              </div>
            </div>
            <div v-if="!likedSubjects.length" class="empty-state">
              暂无点赞教材
            </div>
          </div>
        </div>

        <!-- Private Notes -->
        <div id="private" class="card">
          <div class="card-header clickable-header" @click="router.push('/me/private-notes')">
            <BookOpen :size="20" class="icon-teal" />
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
            <div v-for="note in privateNotesPreview" :key="note.id" class="note-item" @click="openPrivateNoteModal(note)">
              <div class="note-main-info">
                <div class="note-title-line">
                  <Folder v-if="note.type === 'folder'" :size="16" class="note-type-icon folder icon-pink" />
                  <FileText v-else :size="16" class="note-type-icon file icon-blue" />
                  <h4 @click.stop="openRename(note)">{{ note.title }}</h4>
                  <Edit3 :size="12" class="edit-icon-mini" @click.stop="openRename(note)" />
                  <span v-if="note.isShared" class="status-badge" style="margin-left: auto;">已分享</span>
                  <span class="date" :style="note.isShared ? 'margin-left: 12px;' : 'margin-left: auto;'">{{ formatDate(note.updatedAt) }}</span>
                </div>
              </div>
              <div class="note-item-actions">
                <button 
                  class="share-btn-mini"
                  :class="{ 'is-shared': note.isShared }"
                  :disabled="note.isShared"
                  @click.stop="note.isShared ? null : openShareModal(note)"
                  :title="note.isShared ? '已分享' : '分享笔记'"
                >
                  <Share2 :size="18" />
                </button>
                <button 
                  class="toggle-public-mini" 
                  :class="{ isPublic: note.isPublic === 1 }"
                  @click.stop="handleTogglePublic(note)"
                  :title="note.isPublic === 1 ? '已公开' : '已私密'"
                >
                  <ToggleRight v-if="note.isPublic === 1" :size="18" />
                  <ToggleLeft v-else :size="18" />
                  <span>{{ note.isPublic === 1 ? '公开' : '私密' }}</span>
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
            <Share2 :size="20" class="icon-purple" />
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
            <div 
              v-for="note in sharedNotes" 
              :key="note.id" 
              class="note-item shared"
              :class="{ 'inactive': !note.isActive }"
              @click="openPrivateNoteModal({ id: note.nodeId, title: note.noteTitle || note.nodeName, type: note.noteType })"
              style="cursor: pointer;"
            >
              <div class="note-header">
                <div class="note-title-line">
                  <Folder v-if="note.noteType === 'folder'" :size="16" class="icon-pink" />
                  <FileText v-else :size="16" class="icon-blue" />
                  <h4>{{ note.noteTitle || note.nodeName }}</h4>
                  <span v-if="!note.isActive" class="status-badge inactive-badge">已取消</span>
                </div>
                <span class="views"><Activity :size="12" class="icon-green" /> {{ note.viewCount }}</span>
              </div>
              <div class="note-meta">
                <div class="meta-left">
                  <span>提取码: {{ note.shareCode }}</span>
                  <span>至: {{ formatDate(note.expiresAt) }}</span>
                </div>
                <div class="meta-right">
                  <template v-if="note.isActive">
                    <button 
                      class="ghost-btn-mini" 
                      @click.stop="copyToClipboard(note.shareCode)"
                      title="复制提取码"
                    >
                      <Layers :size="14" />
                      <span>提取码</span>
                    </button>
                    <button 
                      class="ghost-btn-mini" 
                      @click.stop="copyToClipboard(`${windowOrigin}/share/verify?token=${note.shareToken}`)"
                      title="复制分享链接"
                    >
                      <Share2 :size="14" />
                      <span>链接</span>
                    </button>
                    <button 
                      class="ghost-btn-mini" 
                      @click.stop="openUpdateExpireModal(note.id)"
                      title="修改过期时间"
                    >
                      <Clock :size="14" />
                      <span>改日期</span>
                    </button>
                    <button 
                      class="danger-btn-mini" 
                      @click.stop="handleToggleShareStatus(note.id, 0)"
                      title="取消分享"
                    >
                      <X :size="14" />
                      <span>取消分享</span>
                    </button>
                    <button 
                      class="danger-btn-mini" 
                      @click.stop="handleDeleteSharedNote(note.id)"
                      title="删除分享"
                    >
                      <Trash2 :size="14" />
                      <span>删除分享</span>
                    </button>
                  </template>
                  <template v-else>
                    <button 
                      class="ghost-btn-mini" 
                      @click.stop="handleToggleShareStatus(note.id, 1)"
                      title="重新分享"
                      style="color: var(--notion-blue); border-color: rgba(0, 117, 222, 0.2);"
                    >
                      <RefreshCcw :size="14" />
                      <span>重新分享</span>
                    </button>
                    <button 
                      class="danger-btn-mini" 
                      @click.stop="handleDeleteSharedNote(note.id)"
                      title="删除分享"
                    >
                      <Trash2 :size="14" />
                      <span>删除分享</span>
                    </button>
                  </template>
                </div>
              </div>
            </div>
            <div v-if="!sharedNotes.length" class="empty-state">
              暂无分享笔记
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 新建笔记/文件夹弹窗 -->
    <Teleport to="body">
      <div v-if="showCreateNoteModal" class="modal-overlay create-modal-overlay" @click.self="showCreateNoteModal = false">
        <div class="modal-content small-modal">
          <header class="modal-header">
            <div class="header-left">
              <FolderPlus v-if="createNoteForm.type === 'folder'" :size="20" class="icon-pink" />
              <FilePlus v-else :size="20" class="icon-blue" />
              <h3>新建{{ createNoteForm.type === 'folder' ? '文件夹' : '笔记' }}</h3>
            </div>
            <button class="close-btn" @click="showCreateNoteModal = false">
              <X :size="20" />
            </button>
          </header>
          <div class="modal-body">
            <div class="form-group">
              <label>名称</label>
              <input 
                v-model="createNoteForm.title" 
                type="text" 
                :placeholder="'请输入' + (createNoteForm.type === 'folder' ? '文件夹' : '笔记') + '名称'" 
                maxlength="255" 
                @keyup.enter="handleCreateNote" 
              />
            </div>
            <div class="form-group">
              <label>隐私设置</label>
              <div class="privacy-selector">
                <button 
                  class="privacy-option" 
                  :class="{ active: createNoteForm.isPublic === 0 }"
                  @click="createNoteForm.isPublic = 0"
                >
                  <ToggleLeft :size="18" />
                  <span>私密</span>
                </button>
                <button 
                  class="privacy-option" 
                  :class="{ active: createNoteForm.isPublic === 1 }"
                  @click="createNoteForm.isPublic = 1"
                >
                  <ToggleRight :size="18" />
                  <span>公开</span>
                </button>
              </div>
              <p class="form-tip">{{ createNoteForm.isPublic === 1 ? '公开内容所有人可见' : '仅自己可见' }}</p>
            </div>
          </div>
          <footer class="modal-footer">
            <div class="form-actions">
              <button class="cancel-btn" @click="showCreateNoteModal = false">取消</button>
              <button class="confirm-btn" :disabled="creatingNote || !createNoteForm.title.trim()" @click="handleCreateNote">
                <Loader2 v-if="creatingNote" class="spin" :size="16" />
                <span>确认创建</span>
              </button>
            </div>
          </footer>
        </div>
      </div>
    </Teleport>

    <!-- 重命名弹窗 -->
    <Teleport to="body">
      <div v-if="showRenameModal" class="modal-overlay rename-modal-overlay" @click.self="showRenameModal = false">
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
              <input v-model="renameTitle" type="text" placeholder="请输入新名称" maxlength="255" @keyup.enter="handleRename" />
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

    <!-- 分享笔记弹窗 -->
    <Teleport to="body">
      <div v-if="showShareModal" class="modal-overlay share-modal-overlay" @click.self="showShareModal = false">
        <div class="modal-content small-modal">
          <header class="modal-header">
            <div class="header-left">
              <Share2 :size="20" class="icon-purple" />
              <h3>分享笔记</h3>
            </div>
            <button class="close-btn" @click="showShareModal = false">
              <X :size="20" />
            </button>
          </header>
          <div class="modal-body">
            <template v-if="!shareResult">
              <div class="share-note-info">
                <p>分享内容: <strong>{{ pendingShareNote?.title || pendingShareNote?.noteTitle || pendingShareNote?.nodeName }}</strong></p>
              </div>
              <div class="form-group">
                <label>过期时间</label>
                <input 
                  v-model="shareExpiresAt" 
                  type="datetime-local" 
                  required
                />
                <p class="form-tip">过期后分享链接将失效</p>
              </div>
            </template>
            <template v-else>
              <div class="share-success">
                <div class="success-icon">
                  <Share2 :size="48" class="icon-purple" />
                </div>
                <h4>分享链接已生成</h4>
                <div class="share-details">
                  <div class="detail-row">
                    <span class="label">分享 Token:</span>
                    <div class="copy-box" @click="copyToClipboard(shareResult.shareToken)">
                      <code>{{ shareResult.shareToken }}</code>
                      <Layers :size="14" />
                    </div>
                  </div>
                  <div class="detail-row">
                    <span class="label">分享链接:</span>
                    <div class="copy-box" @click="copyToClipboard(`${windowOrigin}/share/verify?token=${shareResult.shareToken}`)">
                      <code>{{ `${windowOrigin}/share/verify?token=${shareResult.shareToken}` }}</code>
                      <Layers :size="14" />
                    </div>
                  </div>
                  <div class="detail-row">
                    <span class="label">提取码:</span>
                    <div class="copy-box" @click="copyToClipboard(shareResult.shareCode)">
                      <code class="share-code">{{ shareResult.shareCode }}</code>
                      <Layers :size="14" />
                    </div>
                  </div>
                  <div class="detail-row">
                    <span class="label">有效期至:</span>
                    <span class="value">{{ formatDate(shareResult.expiresAt) }}</span>
                  </div>
                </div>
              </div>
            </template>
          </div>
          <footer class="modal-footer">
            <div class="form-actions">
              <button class="cancel-btn" @click="showShareModal = false">{{ shareResult ? '关闭' : '取消' }}</button>
              <button v-if="!shareResult" class="confirm-btn purple-btn" :disabled="sharing || !shareExpiresAt" @click="handleShare">
                <Loader2 v-if="sharing" class="spin" :size="16" />
                <span>生成分享链接</span>
              </button>
            </div>
          </footer>
        </div>
      </div>
    </Teleport>

    <!-- 修改分享时间弹窗 -->
    <Teleport to="body">
      <div v-if="showUpdateExpireModal" class="modal-overlay share-modal-overlay" @click.self="showUpdateExpireModal = false">
        <div class="modal-content small-modal">
          <header class="modal-header">
            <div class="header-left">
              <Clock :size="20" class="icon-purple" />
              <h3>修改分享过期时间</h3>
            </div>
            <button class="close-btn" @click="showUpdateExpireModal = false">
              <X :size="20" />
            </button>
          </header>
          <div class="modal-body">
            <div class="form-group">
              <label>过期模式</label>
              <div class="toggle-group">
                <button 
                  class="toggle-btn" 
                  :class="{ active: updateExpireMode === 'minutes' }" 
                  @click="updateExpireMode = 'minutes'"
                >按时长</button>
                <button 
                  class="toggle-btn" 
                  :class="{ active: updateExpireMode === 'date' }" 
                  @click="updateExpireMode = 'date'"
                >按日期</button>
              </div>
            </div>

            <!-- 按时长模式 -->
            <div v-if="updateExpireMode === 'minutes'" class="form-group">
              <label>选择分享时长</label>
              <div class="quick-options-grid">
                <button 
                  v-for="opt in quickTimeOptions" 
                  :key="opt.value"
                  class="quick-option-btn"
                  :class="{ active: selectedMinutes === opt.value }"
                  @click="selectedMinutes = opt.value"
                >
                  {{ opt.label }}
                </button>
              </div>

              <!-- 自定义时长输入 -->
              <div v-if="selectedMinutes === 'custom'" class="custom-duration-input" style="margin-top: 16px;">
                <div class="input-with-select">
                  <input 
                    type="number" 
                    v-model="customDurationValue" 
                    min="1" 
                    class="number-input"
                    placeholder="请输入时长"
                  />
                  <select v-model="customDurationUnit" class="unit-select">
                    <option v-for="unit in customUnitOptions" :key="unit.value" :value="unit.value">
                      {{ unit.label }}
                    </option>
                  </select>
                </div>
              </div>
            </div>

            <!-- 按日期模式 -->
            <div v-else class="form-group">
              <label>选择过期日期</label>
              <input 
                v-model="selectedExpireDate" 
                type="datetime-local" 
                required
              />
              <p class="form-tip">精确到具体的过期时间点</p>
            </div>
          </div>
          <footer class="modal-footer">
            <div class="form-actions">
              <button class="cancel-btn" @click="showUpdateExpireModal = false">取消</button>
              <button class="confirm-btn purple-btn" :disabled="updatingExpire" @click="handleUpdateExpire">
                <Loader2 v-if="updatingExpire" class="spin" :size="16" />
                <span>保存修改</span>
              </button>
            </div>
          </footer>
        </div>
      </div>
    </Teleport>

    <!-- 收藏夹教材列表弹窗 -->
    <Teleport to="body">
      <div v-if="showFolderModal" class="modal-overlay folder-modal-overlay" @click.self="showFolderModal = false">
        <div class="modal-content large-modal">
          <header class="modal-header">
            <div class="header-left">
              <FolderHeart :size="24" class="icon-pink" />
              <div class="header-text">
                <div class="modal-title-row" @click="openRenameFolder(selectedFolder)">
                  <h3>{{ selectedFolder?.name }}</h3>
                  <Edit3 :size="16" class="edit-icon-mini" />
                </div>
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
              <div v-for="item in folderSubjects" :key="item.id" class="subject-row" @click="handleFolderSubjectClick(item)">
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

    <!-- 私人笔记内容弹窗 -->
    <Teleport to="body">
      <div v-if="showPrivateNoteModal" class="modal-overlay note-modal-overlay" @click.self="showPrivateNoteModal = false">
        <div class="modal-content large-modal">
          <header class="modal-header">
            <div class="header-left">
              <button class="back-btn" @click="handleNoteBack">
                <ChevronLeft :size="24" />
              </button>
              <div class="header-text">
                <div class="modal-title-row" @click="openRename(noteDetail?.content || {id: notePathStack[notePathStack.length-1]?.id, title: notePathStack[notePathStack.length-1]?.title})">
                  <h3>{{ notePathStack[notePathStack.length - 1]?.title }}</h3>
                  <Edit3 :size="16" class="edit-icon-mini" />
                </div>
                <p v-if="noteDetail?.type === 'folder'">私人文件夹 · {{ noteDetail.total }} 个项目</p>
                <p v-else>私人笔记</p>
              </div>
              <!-- 顶部标题旁的公开按钮 -->
              <button 
                v-if="noteDetail"
                class="toggle-public-mini header-privacy-toggle" 
                :class="{ isPublic: !!(noteDetail.content?.isPublic ?? noteDetail.isPublic) }"
                @click.stop="handleToggleNotePublicInModal()"
                :title="!!(noteDetail.content?.isPublic ?? noteDetail.isPublic) ? '已公开' : '已私密'"
              >
                <ToggleRight v-if="!!(noteDetail.content?.isPublic ?? noteDetail.isPublic)" :size="18" />
                <ToggleLeft v-else :size="18" />
                <span>{{ !!(noteDetail.content?.isPublic ?? noteDetail.isPublic) ? '公开' : '私密' }}</span>
              </button>
            </div>
            <div class="header-right-actions">
              <button 
                v-if="noteDetail"
                class="icon-btn share-btn-header" 
                :class="{ 'is-shared': noteDetail.content?.isShared || noteDetail.isShared }"
                :disabled="noteDetail.content?.isShared || noteDetail.isShared"
                @click="(noteDetail.content?.isShared || noteDetail.isShared) ? null : openShareModal(noteDetail.content || noteDetail)"
                :title="(noteDetail.content?.isShared || noteDetail.isShared) ? '已分享' : '分享笔记'"
              >
                <Share2 :size="20" />
              </button>
              <button 
                v-if="noteDetail?.type === 'folder'"
                class="icon-btn create-btn-header" 
                @click="openCreateNote('folder')"
                title="新建文件夹"
              >
                <FolderPlus :size="20" />
              </button>
              <button 
                v-if="noteDetail?.type === 'folder'"
                class="icon-btn create-btn-header" 
                @click="openCreateNote('markdown')"
                title="新建笔记"
              >
                <FilePlus :size="20" />
              </button>
              <button class="close-btn" @click="showPrivateNoteModal = false">
                <X :size="24" />
              </button>
            </div>
          </header>

          <div class="modal-body">
            <div v-if="loadingNoteDetail" class="modal-loading">
              <Loader2 class="spin" :size="32" />
              <p>加载中...</p>
            </div>
            <div v-else-if="!noteDetail" class="modal-empty">
              <p>加载失败</p>
            </div>
            
            <!-- 文件夹视图 -->
            <div v-else-if="noteDetail.type === 'folder'" class="modal-list">
              <div v-for="child in noteDetail.children" :key="child.id" class="subject-row" @click="handleNoteChildClick(child)">
                <Folder v-if="child.type === 'folder'" :size="18" class="icon-pink" />
                <FileText v-else :size="18" class="icon-blue" />
                <div class="subject-row-info">
                  <div class="note-title-line">
                    <h4 @click.stop="openRename(child)">{{ child.title }}</h4>
                    <Edit3 :size="12" class="edit-icon-mini" @click.stop="openRename(child)" />
                    <span v-if="child.isShared" class="status-badge" style="margin-left: auto; margin-right: 16px;">已分享</span>
                  </div>
                </div>
                <div class="subject-row-meta">
                  <!-- 子项列表中的分享按钮 -->
                  <button 
                    class="share-btn-mini"
                    :class="{ 'is-shared': child.isShared }"
                    :disabled="child.isShared"
                    @click.stop="child.isShared ? null : openShareModal(child)"
                    :title="child.isShared ? '已分享' : '分享笔记'"
                  >
                    <Share2 :size="16" />
                  </button>
                  <!-- 子项列表中的公开按钮 -->
                  <button 
                    class="toggle-public-mini" 
                    :class="{ isPublic: !!child.isPublic }"
                    @click.stop="handleToggleNotePublicInModal(child)"
                    :title="!!child.isPublic ? '已公开' : '已私密'"
                  >
                    <ToggleRight v-if="!!child.isPublic" :size="16" />
                    <ToggleLeft v-else :size="16" />
                    <span>{{ !!child.isPublic ? '公开' : '私密' }}</span>
                  </button>
                  <span>{{ formatDate(child.updatedAt) }}</span>
                  <ChevronRight :size="18" class="row-arrow" />
                </div>
              </div>
              <div v-if="!noteDetail.children?.length" class="modal-empty">
                <p>文件夹为空</p>
              </div>
            </div>

            <!-- 文件内容视图 -->
            <div v-else-if="noteDetail.type === 'markdown'" class="markdown-content">
              <div class="content-header">
                <!-- 顶部标题旁的公开按钮已经在 header 中，此处显示同步状态 -->
                <div class="header-placeholder"></div>
              </div>
              
              <div class="content-editor">
                <textarea 
                  v-model="editNoteContent" 
                  @input="handleNoteContentInput"
                  placeholder="请输入笔记内容..."
                  maxlength="1000"
                ></textarea>
                <div class="editor-footer">
                  <div class="sync-indicator" :class="syncStatus">
                    <template v-if="syncStatus === 'saving'">
                      <Loader2 class="spin" :size="14" />
                      <span>同步中...</span>
                    </template>
                    <template v-else-if="syncStatus === 'saved'">
                      <Save :size="14" />
                      <span>已同步</span>
                    </template>
                    <template v-else-if="syncStatus === 'error'">
                      <X :size="14" />
                      <span>同步失败</span>
                    </template>
                  </div>
                  <span>{{ editNoteContent.length }} / 1000</span>
                </div>
              </div>
            </div>
          </div>
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

/* 隐私选择器 */
.privacy-selector {
  display: flex;
  gap: 8px;
  margin-top: 8px;
}

.privacy-option {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 8px;
  border-radius: 6px;
  border: 1px solid rgba(0,0,0,0.1);
  background: white;
  cursor: pointer;
  font-size: 13px;
  color: var(--warm-gray-500);
  transition: all 0.2s;
}

.privacy-option.active {
  border-color: var(--notion-blue);
  color: var(--notion-blue);
  background: rgba(0, 117, 222, 0.05);
}

/* 分享相关样式 */
.share-btn-mini {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 4px;
  border: none;
  background: transparent;
  color: var(--warm-gray-500);
  cursor: pointer;
  transition: all 0.2s;
}

.share-btn-mini:hover:not(:disabled) {
  background: rgba(139, 92, 246, 0.1);
  color: #8b5cf6;
}

.share-btn-header {
  color: #8b5cf6 !important;
}

.share-btn-header:hover:not(:disabled) {
  background: rgba(139, 92, 246, 0.1) !important;
}

.share-btn-mini.is-shared,
.share-btn-header.is-shared {
  color: var(--warm-gray-300) !important;
  cursor: not-allowed;
  opacity: 0.6;
}

.share-btn-mini.is-shared:hover,
.share-btn-header.is-shared:hover {
  background: transparent !important;
}

.purple-btn {
  background: #8b5cf6 !important;
  color: white !important;
}

.purple-btn:hover {
  background: #7c3aed !important;
}

.share-note-info {
  margin-bottom: 16px;
  padding: 12px;
  background: var(--warm-white);
  border-radius: 8px;
  font-size: 14px;
}

.share-success {
  text-align: center;
  padding: 16px 0;
}

.success-icon {
  margin-bottom: 16px;
}

.share-success h4 {
  font-size: 18px;
  margin-bottom: 24px;
  color: var(--warm-dark);
}

.share-details {
  background: var(--warm-white);
  padding: 16px;
  border-radius: 12px;
  text-align: left;
}

.detail-row {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 16px;
}

.detail-row:last-child {
  margin-bottom: 0;
}

.detail-row .label {
  font-size: 12px;
  color: var(--warm-gray-500);
  font-weight: 500;
}

.copy-box {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  background: white;
  border: 1px solid rgba(0,0,0,0.1);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}

.copy-box:hover {
  border-color: #8b5cf6;
  background: rgba(139, 92, 246, 0.02);
}

.copy-box code {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 13px;
  color: var(--warm-dark);
  word-break: break-all;
}

.copy-box .share-code {
  font-size: 18px;
  font-weight: bold;
  letter-spacing: 2px;
  color: #8b5cf6;
}

.toggle-group {
  display: flex;
  background: var(--warm-white);
  padding: 4px;
  border-radius: 8px;
  margin-bottom: 16px;
}

.toggle-btn {
  flex: 1;
  padding: 8px;
  border: none;
  background: transparent;
  color: var(--warm-gray-500);
  font-size: 14px;
  font-weight: 500;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.toggle-btn.active {
  background: white;
  color: var(--notion-black);
  box-shadow: 0 2px 4px rgba(0,0,0,0.05);
}

.quick-options-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
}

.quick-option-btn {
  padding: 12px;
  border: 1px solid rgba(0,0,0,0.1);
  background: white;
  border-radius: 8px;
  color: var(--warm-gray-500);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.quick-option-btn:hover {
  background: rgba(139, 92, 246, 0.05);
  border-color: #c4b5fd;
}

.quick-option-btn.active {
  background: rgba(139, 92, 246, 0.1);
  border-color: #8b5cf6;
  color: #8b5cf6;
}

.input-with-select {
  display: flex;
  align-items: center;
  border: 1px solid rgba(0,0,0,0.1);
  border-radius: 8px;
  overflow: hidden;
  background: white;
}

.number-input {
  flex: 1;
  padding: 12px 16px;
  border: none;
  font-size: 14px;
  outline: none;
  background: transparent;
  color: var(--notion-black);
}

.unit-select {
  padding: 12px 16px;
  border: none;
  border-left: 1px solid rgba(0,0,0,0.1);
  background: var(--warm-white);
  color: var(--warm-gray-500);
  font-size: 14px;
  font-weight: 500;
  outline: none;
  cursor: pointer;
}

.unit-select:focus {
  background: white;
}

.detail-row .value {
  font-size: 14px;
  color: var(--warm-dark);
}

.form-tip {
  margin-top: 6px;
  font-size: 12px;
  color: var(--warm-gray-300);
}

.header-right-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.icon-btn {
  background: transparent;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  outline: none;
}

.create-btn-header {
  color: var(--warm-gray-500);
  padding: 6px;
  border-radius: 4px;
  transition: all 0.2s;
}

.create-btn-header:hover {
  background: rgba(0,0,0,0.05);
  color: var(--notion-black);
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

.icon-blue { color: var(--notion-blue); }
.icon-green { color: #1aae39; }
.icon-orange { color: #dd5b00; }
.icon-danger { color: #eb5757; }
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
  color: var(--notion-blue);
}

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
  display: flex;
  align-items: center;
  gap: 6px;
  background: rgba(0,0,0,0.05);
  border: none;
  color: var(--warm-gray-500);
  cursor: pointer;
  padding: 4px 10px;
  border-radius: 9999px;
  font-size: 12px;
  font-weight: 600;
  transition: all 0.2s ease;
}

.toggle-public-mini:hover {
  background: rgba(0,0,0,0.1);
}

.toggle-public-mini.isPublic {
  background: #f2f9ff;
  color: var(--notion-blue);
}

.toggle-public-mini.isPublic:hover {
  background: #e1f0ff;
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

.note-item.shared.inactive {
  background: var(--warm-white);
  border-color: rgba(0,0,0,0.05);
}

.status-badge {
  font-size: 12px;
  padding: 2px 6px;
  border-radius: 4px;
  font-weight: 500;
  margin-left: 8px;
}

.inactive-badge {
  background: rgba(0,0,0,0.05);
  color: var(--warm-gray-500);
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
  align-items: center;
  font-size: 12px;
  color: var(--warm-gray-300);
}

.meta-left {
  display: flex;
  gap: 16px;
  color: var(--warm-gray-500);
  font-size: 13px;
}

.meta-right {
  display: flex;
  gap: 8px;
}

.ghost-btn-mini {
  display: flex;
  align-items: center;
  gap: 4px;
  background: transparent;
  border: 1px solid rgba(0,0,0,0.1);
  color: var(--warm-gray-500);
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  transition: all 0.2s ease;
}

.ghost-btn-mini:hover {
  background: rgba(0,0,0,0.05);
  color: var(--notion-black);
}

.danger-btn-mini {
  display: flex;
  align-items: center;
  gap: 4px;
  background: transparent;
  border: 1px solid rgba(239, 68, 68, 0.2);
  color: #ef4444;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  transition: all 0.2s ease;
}

.danger-btn-mini:hover {
  background: rgba(239, 68, 68, 0.05);
  border-color: #ef4444;
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

.rename-modal-overlay {
  z-index: 1300; /* 确保在展示弹窗之上 */
}

.create-modal-overlay {
  z-index: 1310; /* 确保在展示弹窗之上 */
}

.folder-modal-overlay {
  z-index: 1050;
}

.note-modal-overlay {
  z-index: 1060;
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

.modal-content.large-modal {
  max-width: 800px;
  max-height: 85vh;
  display: flex;
  flex-direction: column;
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

.modal-title-row {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
}

.modal-title-row:hover .edit-icon-mini {
  opacity: 1;
  color: var(--notion-blue);
}

.header-left {
  display: flex;
  gap: 16px;
  align-items: center;
}

.back-btn {
  background: transparent;
  border: none;
  color: var(--warm-gray-500);
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.back-btn:hover {
  background: var(--warm-white);
  color: var(--notion-black);
}

.header-text h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 700;
}

.header-text p {
  margin: 4px 0 0;
  font-size: 13px;
  color: var(--warm-gray-500);
}

.modal-body {
  padding: 24px 20px;
  overflow-y: auto;
}

.markdown-content {
  color: var(--notion-black);
  line-height: 1.6;
}

.content-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 12px;
  border-bottom: var(--whisper-border);
}

.status-badge {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 4px;
  background: #f2f9ff;
  color: var(--notion-blue);
  font-weight: 600;
}

.status-badge.private {
  background: rgba(0,0,0,0.05);
  color: var(--warm-gray-500);
}

.sync-indicator {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 500;
  transition: all 0.3s ease;
}

.sync-indicator.saving {
  color: var(--notion-blue);
}

.sync-indicator.saved {
  color: #1aae39;
}

.sync-indicator.error {
  color: #eb5757;
}

.content-header .actions {
  display: flex;
  gap: 12px;
}

.edit-btn, .save-btn, .cancel-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border-radius: 4px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  border: 1px solid transparent;
}

.edit-btn {
  background: var(--warm-white);
  color: var(--notion-black);
}

.edit-btn:hover {
  background: rgba(0,0,0,0.1);
}

.save-btn {
  background: var(--notion-blue);
  color: white;
}

.save-btn:hover {
  background: var(--notion-blue-hover);
}

.save-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.cancel-btn {
  background: transparent;
  color: var(--warm-gray-500);
}

.cancel-btn:hover {
  color: var(--notion-black);
}

.content-editor textarea {
  width: 100%;
  min-height: 300px;
  padding: 16px;
  border-radius: 8px;
  border: 1px solid #ddd;
  font-size: 15px;
  line-height: 1.6;
  font-family: inherit;
  resize: vertical;
  transition: all 0.2s ease;
}

.content-editor textarea:focus {
  outline: none;
  border-color: var(--notion-blue);
  box-shadow: 0 0 0 3px rgba(0,117,222,0.1);
}

.editor-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 12px;
  font-size: 12px;
  color: var(--warm-gray-300);
}

.markdown-content .content-body {
  font-size: 15px;
  white-space: pre-wrap;
  word-break: break-word;
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

.header-privacy-toggle {
  margin-left: 12px;
}

.subject-row-meta {
  display: flex;
  align-items: center;
  gap: 16px;
  color: var(--warm-gray-300);
  font-size: 13px;
  font-weight: 500;
}

.subject-row-meta .toggle-public-mini {
  padding: 2px 8px;
  font-size: 11px;
}

.subject-row-meta .toggle-public-mini svg {
  width: 14px;
  height: 14px;
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

/* 教材分组样式 */
.subject-group {
  margin-bottom: 24px;
}

.subject-group:last-child {
  margin-bottom: 0;
}

.group-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--warm-gray-500);
  margin-bottom: 12px;
  padding-left: 10px;
  border-left: 3px solid var(--notion-blue);
}

.status-badge.published {
  background: rgba(16, 185, 129, 0.1);
  color: #10b981;
}

.status-badge.has-draft {
  background: rgba(245, 158, 11, 0.1);
  color: #f59e0b;
}

.status-badge.draft {
  background: rgba(107, 114, 128, 0.1);
  color: #6b7280;
}

.icon-gray {
  color: #6b7280;
}
</style>
