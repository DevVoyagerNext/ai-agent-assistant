<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  BadgeCheck,
  BookOpenCheck,
  Check,
  ClipboardList,
  Clock3,
  FileClock,
  GitBranch,
  History,
  Loader2,
  LogOut,
  UserRound,
  X
} from 'lucide-vue-next'
import {
  approveNode,
  approveSubject,
  getAdminMe,
  getAuditRecords,
  getPendingNodes,
  getPendingSubjects,
  rejectNode,
  rejectSubject
} from '../api/audit'
import type { AuditRecord, MeResponse, NodeReviewItem, SubjectReviewItem } from '../types/admin'
import { clearAdminAuth } from '../utils/adminAuth'

type TabKey = 'subjects' | 'nodes' | 'records'
type ReviewTarget = SubjectReviewItem | NodeReviewItem

const router = useRouter()
const loading = ref(false)
const actionLoading = ref<string | null>(null)
const activeTab = ref<TabKey>('subjects')
const me = ref<MeResponse | null>(null)
const pendingSubjects = ref<SubjectReviewItem[]>([])
const pendingNodes = ref<NodeReviewItem[]>([])
const auditRecords = ref<AuditRecord[]>([])
const notice = reactive({ text: '', type: 'success' as 'success' | 'error' })
const rejectDialog = reactive<{
  open: boolean
  targetType: 'subject' | 'node'
  target: ReviewTarget | null
  remark: string
}>({
  open: false,
  targetType: 'subject',
  target: null,
  remark: ''
})

const adminName = computed(() => me.value?.admin.username || '管理员')
const displaySubjectName = (item: SubjectReviewItem) => {
  return item.nameDraft || item.name || '未命名教材'
}

const displayNodeName = (item: NodeReviewItem) => {
  return item.nameDraft || item.name || `节点 #${item.id}`
}

const displayRecordName = (item: AuditRecord) => {
  return item.targetName || item.subjectDraftName || item.subjectName || `目标 #${item.targetId}`
}

const displayDescription = (item: SubjectReviewItem) => {
  return item.descriptionDraft || item.description || '暂无简介'
}

const displayNodeContent = (item: NodeReviewItem) => {
  return item.contentDraft || item.content || '暂无正文草稿'
}

const formatDate = (value?: string | null) => {
  if (!value) return '暂无'
  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  }).format(new Date(value))
}

const showNotice = (text: string, type: 'success' | 'error' = 'success') => {
  notice.text = text
  notice.type = type
  window.setTimeout(() => {
    if (notice.text === text) notice.text = ''
  }, 2400)
}

const loadAll = async () => {
  loading.value = true
  try {
    const [profile, subjects, nodes, records] = await Promise.all([
      getAdminMe(),
      getPendingSubjects(),
      getPendingNodes(),
      getAuditRecords()
    ])
    me.value = profile
    pendingSubjects.value = subjects.list
    pendingNodes.value = nodes.list
    auditRecords.value = records.list
  } catch (err) {
    showNotice(err instanceof Error ? err.message : '加载后台数据失败', 'error')
  } finally {
    loading.value = false
  }
}

const refreshAfterReview = async () => {
  const [profile, subjects, nodes, records] = await Promise.all([
    getAdminMe(),
    getPendingSubjects(),
    getPendingNodes(),
    getAuditRecords()
  ])
  me.value = profile
  pendingSubjects.value = subjects.list
  pendingNodes.value = nodes.list
  auditRecords.value = records.list
}

const handleApproveSubject = async (subject: SubjectReviewItem) => {
  actionLoading.value = `subject-${subject.id}`
  try {
    await approveSubject(subject.id)
    await refreshAfterReview()
    showNotice('教材已通过审核')
  } catch (err) {
    showNotice(err instanceof Error ? err.message : '审批失败', 'error')
  } finally {
    actionLoading.value = null
  }
}

const handleApproveNode = async (node: NodeReviewItem) => {
  actionLoading.value = `node-${node.id}`
  try {
    await approveNode(node.id)
    await refreshAfterReview()
    showNotice('节点已通过审核')
  } catch (err) {
    showNotice(err instanceof Error ? err.message : '审批失败', 'error')
  } finally {
    actionLoading.value = null
  }
}

const openReject = (targetType: 'subject' | 'node', target: ReviewTarget) => {
  rejectDialog.targetType = targetType
  rejectDialog.target = target
  rejectDialog.remark = ''
  rejectDialog.open = true
}

const handleReject = async () => {
  if (!rejectDialog.target) return
  if (!rejectDialog.remark.trim()) {
    showNotice('请填写驳回意见', 'error')
    return
  }

  const id = rejectDialog.target.id
  actionLoading.value = `${rejectDialog.targetType}-${id}`
  try {
    if (rejectDialog.targetType === 'subject') {
      await rejectSubject(id, rejectDialog.remark.trim())
    } else {
      await rejectNode(id, rejectDialog.remark.trim())
    }
    rejectDialog.open = false
    await refreshAfterReview()
    showNotice(rejectDialog.targetType === 'subject' ? '教材已驳回' : '节点已驳回')
  } catch (err) {
    showNotice(err instanceof Error ? err.message : '驳回失败', 'error')
  } finally {
    actionLoading.value = null
  }
}

const rejectTitle = computed(() => {
  if (!rejectDialog.target) return '驳回'
  return rejectDialog.targetType === 'subject'
    ? displaySubjectName(rejectDialog.target as SubjectReviewItem)
    : displayNodeName(rejectDialog.target as NodeReviewItem)
})

const logout = () => {
  clearAdminAuth()
  router.push('/login')
}

onMounted(loadAll)
</script>

<template>
  <main class="admin-shell">
    <aside class="sidebar">
      <div class="brand">
        <span class="brand-icon"><BookOpenCheck :size="24" /></span>
        <div>
          <strong>审核后台</strong>
          <small>内容管理</small>
        </div>
      </div>

      <div class="admin-mini">
        <UserRound :size="28" />
        <div>
          <strong>{{ adminName }}</strong>
          <span>后台管理</span>
        </div>
      </div>

      <nav class="side-nav">
        <button :class="{ active: activeTab === 'subjects' }" @click="activeTab = 'subjects'">
          <span class="nav-main">
            <ClipboardList :size="18" />
            <span>教材审批</span>
          </span>
        </button>
        <button :class="{ active: activeTab === 'nodes' }" @click="activeTab = 'nodes'">
          <span class="nav-main">
            <GitBranch :size="18" />
            <span>节点审批</span>
          </span>
        </button>
        <button :class="{ active: activeTab === 'records' }" @click="activeTab = 'records'">
          <span class="nav-main">
            <History :size="18" />
            <span>审批记录</span>
          </span>
        </button>
      </nav>

      <button class="ghost-button logout-button" @click="logout">
        <LogOut :size="18" />
        <span>退出登录</span>
      </button>
    </aside>

    <section class="workspace">
      <div v-if="notice.text" class="notice" :class="notice.type">
        {{ notice.text }}
      </div>

      <section v-if="activeTab === 'subjects'" class="panel">
        <div class="panel-header">
          <div class="panel-title">
            <h2>待审批教材</h2>
            <span>{{ pendingSubjects.length }} 个待处理</span>
          </div>
        </div>

        <div v-if="pendingSubjects.length === 0" class="empty-state">
          <BookOpenCheck :size="42" />
          <strong>暂无待审批教材</strong>
        </div>

        <div v-else class="review-list">
          <article v-for="subject in pendingSubjects" :key="subject.id" class="review-item">
            <div class="subject-main">
              <div class="subject-title-row">
                <h3>{{ displaySubjectName(subject) }}</h3>
                <span class="status-badge">待审批</span>
              </div>
              <p>{{ displayDescription(subject) }}</p>
              <div class="subject-meta">
                <span>作者：{{ subject.creatorName || `用户 #${subject.creatorId}` }}</span>
                <span>提交时间：{{ formatDate(subject.createdAt) }}</span>
                <span>Slug：{{ subject.slug }}</span>
              </div>
            </div>

            <div class="review-actions">
              <button class="approve-button" :disabled="actionLoading === `subject-${subject.id}`" @click="handleApproveSubject(subject)">
                <Loader2 v-if="actionLoading === `subject-${subject.id}`" class="spin" :size="17" />
                <Check v-else :size="17" />
                <span>通过</span>
              </button>
              <button class="reject-button" :disabled="actionLoading === `subject-${subject.id}`" @click="openReject('subject', subject)">
                <X :size="17" />
                <span>驳回</span>
              </button>
            </div>
          </article>
        </div>
      </section>

      <section v-else-if="activeTab === 'nodes'" class="panel">
        <div class="panel-header">
          <div class="panel-title">
            <h2>待审批节点</h2>
            <span>{{ pendingNodes.length }} 个待处理</span>
          </div>
        </div>

        <div v-if="pendingNodes.length === 0" class="empty-state">
          <GitBranch :size="42" />
          <strong>暂无待审批节点</strong>
        </div>

        <div v-else class="review-list">
          <article v-for="node in pendingNodes" :key="node.id" class="review-item">
            <div class="subject-main">
              <div class="subject-title-row">
                <h3>{{ displayNodeName(node) }}</h3>
                <span class="status-badge">节点续更</span>
              </div>
              <p class="node-preview">{{ displayNodeContent(node) }}</p>
              <div class="subject-meta">
                <span>教材：{{ node.subjectName || `教材 #${node.subjectId}` }}</span>
                <span>作者：{{ node.creatorName || `用户 #${node.creatorId}` }}</span>
                <span>层级：{{ node.level }}</span>
              </div>
            </div>

            <div class="review-actions">
              <button class="approve-button" :disabled="actionLoading === `node-${node.id}`" @click="handleApproveNode(node)">
                <Loader2 v-if="actionLoading === `node-${node.id}`" class="spin" :size="17" />
                <Check v-else :size="17" />
                <span>通过</span>
              </button>
              <button class="reject-button" :disabled="actionLoading === `node-${node.id}`" @click="openReject('node', node)">
                <X :size="17" />
                <span>驳回</span>
              </button>
            </div>
          </article>
        </div>
      </section>

      <section v-else class="panel">
        <div class="panel-header">
          <div class="panel-title">
            <h2>审批记录</h2>
            <span>{{ auditRecords.length }} 条记录</span>
          </div>
        </div>

        <div class="stat-grid record-stats">
          <article class="stat-card pending">
            <FileClock :size="22" />
            <span>待审批</span>
            <strong>{{ me?.stats.pending ?? 0 }}</strong>
          </article>
          <article class="stat-card approved">
            <BadgeCheck :size="22" />
            <span>已通过</span>
            <strong>{{ me?.stats.approved ?? 0 }}</strong>
          </article>
          <article class="stat-card rejected">
            <X :size="22" />
            <span>已驳回</span>
            <strong>{{ me?.stats.rejected ?? 0 }}</strong>
          </article>
          <article class="stat-card today">
            <Clock3 :size="22" />
            <span>今日处理</span>
            <strong>{{ me?.stats.todayReview ?? 0 }}</strong>
          </article>
        </div>

        <div v-if="auditRecords.length === 0" class="empty-state">
          <History :size="42" />
          <strong>暂无审批记录</strong>
        </div>

        <div v-else class="record-table">
          <div class="record-row header">
            <span>对象</span>
            <span>类型</span>
            <span>动作</span>
            <span>管理员</span>
            <span>时间</span>
          </div>
          <div v-for="record in auditRecords" :key="record.id" class="record-row">
            <span>{{ displayRecordName(record) }}</span>
            <span>{{ record.targetType === 'subject' ? '教材' : '节点' }}</span>
            <span>
              <em :class="record.action">{{ record.action === 'approve' ? '通过' : '驳回' }}</em>
            </span>
            <span>{{ record.adminName || `管理员 #${record.adminId}` }}</span>
            <span>{{ formatDate(record.createdAt) }}</span>
          </div>
        </div>
      </section>
    </section>

    <div v-if="rejectDialog.open" class="modal-mask">
      <section class="modal">
        <header>
          <h2>{{ rejectDialog.targetType === 'subject' ? '驳回教材' : '驳回节点' }}</h2>
          <button class="icon-button" @click="rejectDialog.open = false">
            <X :size="18" />
          </button>
        </header>
        <p>{{ rejectTitle }}</p>
        <textarea v-model.trim="rejectDialog.remark" rows="5" placeholder="填写审批意见"></textarea>
        <footer>
          <button class="ghost-button" @click="rejectDialog.open = false">取消</button>
          <button class="reject-button" :disabled="actionLoading === `${rejectDialog.targetType}-${rejectDialog.target?.id}`" @click="handleReject">
            <Loader2 v-if="actionLoading === `${rejectDialog.targetType}-${rejectDialog.target?.id}`" class="spin" :size="17" />
            <X v-else :size="17" />
            <span>确认驳回</span>
          </button>
        </footer>
      </section>
    </div>
  </main>
</template>
