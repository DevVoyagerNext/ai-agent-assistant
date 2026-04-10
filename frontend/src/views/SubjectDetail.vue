<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getTopNodes, getChildNodes, getNodeDetail, getNodeNote } from '../api/node'
import type { SubjectNode, SubjectNodeDetail, NodeNote } from '../types/node'
import { 
  ChevronRight, ChevronDown, FileText, ArrowLeft, 
  Edit3, CheckCircle2, Circle, BookOpen, 
  Clock, Award, BookOpenCheck, Loader2
} from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()
const subjectId = Number(route.params.id)
const isLoggedIn = computed(() => !!localStorage.getItem('token'))

// ----------------- 目录树相关 -----------------
interface TreeNode extends SubjectNode {
  expanded?: boolean;
  children?: TreeNode[];
  loadingChildren?: boolean;
}

const topNodes = ref<TreeNode[]>([])
const loadingTree = ref(false)

const fetchTopNodes = async () => {
  loadingTree.value = true
  try {
    const res = await getTopNodes(subjectId)
    if (res.data?.code === 200 && res.data.data) {
      topNodes.value = res.data.data.map(node => ({ ...node, expanded: false, children: [] }))
      // 默认加载第一个顶级节点
      if (topNodes.value.length > 0) {
        await toggleNode(topNodes.value[0])
      }
    }
  } catch (error) {
    console.error('获取顶级节点失败', error)
  } finally {
    loadingTree.value = false
  }
}

const toggleNode = async (node: TreeNode) => {
  // 如果是叶子节点，直接加载详情
  if (node.isLeaf === 1) {
    await selectNode(node.id)
    return
  }

  // 展开/收起
  node.expanded = !node.expanded
  
  // 如果展开且没有子节点，则去拉取
  if (node.expanded && (!node.children || node.children.length === 0)) {
    node.loadingChildren = true
    try {
      const res = await getChildNodes(node.id)
      if (res.data?.code === 200 && res.data.data) {
        node.children = res.data.data.map(child => ({ ...child, expanded: false, children: [] }))
      }
    } catch (error) {
      console.error('获取子节点失败', error)
    } finally {
      node.loadingChildren = false
    }
  }

  // 加载当前节点的详情和笔记
  await selectNode(node.id)
}

// ----------------- 正文和笔记相关 -----------------
const currentNodeId = ref<number | null>(null)
const nodeDetail = ref<SubjectNodeDetail | null>(null)
const nodeNote = ref<NodeNote | null>(null)
const loadingDetail = ref(false)

const selectNode = async (id: number) => {
  if (currentNodeId.value === id) return
  currentNodeId.value = id
  loadingDetail.value = true
  
  try {
    const resDetail = await getNodeDetail(id)
    if (resDetail.data?.code === 200 && resDetail.data.data) {
      nodeDetail.value = resDetail.data.data
    } else {
      nodeDetail.value = null
    }

    if (isLoggedIn.value) {
      const resNote = await getNodeNote(id)
      if (resNote.data?.code === 200 && resNote.data.data) {
        nodeNote.value = resNote.data.data
      } else {
        nodeNote.value = null
      }
    }
  } catch (error) {
    console.error('获取详情或笔记失败', error)
  } finally {
    loadingDetail.value = false
  }
}

onMounted(() => {
  if (!subjectId) {
    router.replace('/')
    return
  }
  fetchTopNodes()
})

const goBack = () => {
  router.push('/')
}
</script>

<template>
  <div class="study-layout">
    <!-- 左侧目录树 -->
    <aside class="sidebar">
      <div class="sidebar-header">
        <button class="back-btn" @click="goBack">
          <ArrowLeft :size="18" />
          <span>返回大厅</span>
        </button>
        <div class="course-title-wrap">
          <BookOpenCheck class="title-icon" :size="20" />
          <h3>教材目录</h3>
        </div>
      </div>
      
      <div class="tree-container">
        <div v-if="loadingTree" class="loading-state">
          <Loader2 class="spin" :size="24" />
          <span>正在构建知识树...</span>
        </div>
        <div v-else-if="topNodes.length === 0" class="empty-state">
          <FileText :size="32" />
          <span>暂无目录数据</span>
        </div>
        <ul v-else class="tree-list">
          <template v-for="node in topNodes" :key="node.id">
            <li class="tree-item">
              <div 
                class="node-content" 
                :class="{ 
                  active: currentNodeId === node.id,
                  'is-parent': node.isLeaf === 0
                }"
                @click="toggleNode(node)"
              >
                <span class="expand-icon" v-if="node.isLeaf === 0">
                  <ChevronDown v-if="node.expanded" :size="14" />
                  <ChevronRight v-else :size="14" />
                </span>
                <span class="expand-icon placeholder" v-else></span>
                
                <CheckCircle2 v-if="node.userProgressStatus === 'completed'" class="status-icon completed" :size="16" />
                <Circle v-else class="status-icon unstarted" :size="16" />
                
                <span class="node-name">{{ node.name }}</span>
              </div>
              
              <!-- 子节点递归渲染 (支持多层) -->
              <transition name="fade">
                <ul v-if="node.expanded && node.children && node.children.length > 0" class="tree-list sub-list">
                  <li v-for="child in node.children" :key="child.id" class="tree-item">
                    <div 
                      class="node-content sub-content" 
                      :class="{ active: currentNodeId === child.id }"
                      @click="toggleNode(child)"
                    >
                      <span class="expand-icon" v-if="child.isLeaf === 0">
                        <ChevronDown v-if="child.expanded" :size="14" />
                        <ChevronRight v-else :size="14" />
                      </span>
                      <span class="expand-icon placeholder" v-else></span>

                      <CheckCircle2 v-if="child.userProgressStatus === 'completed'" class="status-icon completed" :size="16" />
                      <Circle v-else class="status-icon unstarted" :size="16" />

                      <span class="node-name">{{ child.name }}</span>
                    </div>
                    
                    <ul v-if="child.expanded && child.children && child.children.length > 0" class="tree-list sub-list">
                      <li v-for="grandchild in child.children" :key="grandchild.id" class="tree-item">
                        <div 
                          class="node-content grand-content" 
                          :class="{ active: currentNodeId === grandchild.id }"
                          @click="selectNode(grandchild.id)"
                        >
                          <span class="expand-icon placeholder"></span>
                          <CheckCircle2 v-if="grandchild.userProgressStatus === 'completed'" class="status-icon completed" :size="16" />
                          <Circle v-else class="status-icon unstarted" :size="16" />
                          <span class="node-name">{{ grandchild.name }}</span>
                        </div>
                      </li>
                    </ul>
                    <div v-if="child.loadingChildren" class="loading-sub">
                      <Loader2 class="spin" :size="14" />
                    </div>
                  </li>
                </ul>
              </transition>
              <div v-if="node.loadingChildren" class="loading-sub">
                <Loader2 class="spin" :size="14" />
              </div>
            </li>
          </template>
        </ul>
      </div>
    </aside>

    <!-- 中间正文区 -->
    <main class="main-content">
      <div v-if="loadingDetail" class="content-skeleton">
        <div class="skeleton-header"></div>
        <div class="skeleton-line"></div>
        <div class="skeleton-line short"></div>
        <div class="skeleton-body"></div>
      </div>
      <div v-else-if="nodeDetail" class="content-area">
        <header class="content-header">
          <div class="breadcrumb">
            <BookOpen :size="14" />
            <span>教材正文</span>
          </div>
          <h1>{{ nodeDetail.name }}</h1>
          <div class="meta-info">
            <div class="difficulty-tags">
              <span class="diff-tag easy">易 {{ nodeDetail.easyCount }}</span>
              <span class="diff-tag medium">中 {{ nodeDetail.mediumCount }}</span>
              <span class="diff-tag hard">难 {{ nodeDetail.hardCount }}</span>
            </div>
            <div class="status-badge" :class="nodeDetail.userProgressStatus">
              <Award v-if="nodeDetail.userProgressStatus === 'completed'" :size="14" />
              <Clock v-else :size="14" />
              <span>{{ nodeDetail.userProgressStatus === 'completed' ? '已掌握' : nodeDetail.userProgressStatus === 'learning' ? '学习中' : '未开始' }}</span>
            </div>
          </div>
        </header>

        <article class="markdown-article">
          <div class="content-body">
            <!-- 预留 Markdown 渲染，目前显示原始内容 -->
            <div class="markdown-placeholder">
              {{ nodeDetail.content }}
            </div>
          </div>
        </article>
      </div>
      <div v-else class="empty-view">
        <div class="empty-illustration">
          <FileText :size="64" />
        </div>
        <h2>开启你的沉浸式学习</h2>
        <p>在左侧选择感兴趣的知识点，即刻开启深度学习之旅</p>
      </div>
    </main>

    <!-- 右侧随堂笔记区 -->
    <aside class="note-sidebar">
      <div class="note-header">
        <div class="note-title">
          <Edit3 :size="18" />
          <h3>随堂笔记</h3>
        </div>
      </div>
      
      <div class="note-main">
        <div v-if="!isLoggedIn" class="note-login-guide">
          <div class="guide-icon">🔒</div>
          <p>随堂笔记已锁定</p>
          <p class="sub-p">登录后即可随时记录学习感悟</p>
          <button class="primary-btn" @click="router.push('/login')">去登录</button>
        </div>
        <div v-else-if="!currentNodeId" class="note-empty">
          <p>选中一个知识点来记录笔记吧</p>
        </div>
        <div v-else class="note-active">
          <div class="editor-wrap">
            <textarea 
              v-model="nodeNote!.noteContent"
              class="note-textarea" 
              placeholder="在这里输入你的随堂笔记..."
              readonly
            ></textarea>
          </div>
          <div class="editor-footer">
            <div class="save-status">
              <Clock :size="12" />
              <span v-if="nodeNote?.updatedAt">上次同步: {{ nodeNote.updatedAt }}</span>
              <span v-else>尚未记录笔记</span>
            </div>
            <button class="action-btn disabled">保存 (只读)</button>
          </div>
        </div>
      </div>
    </aside>
  </div>
</template>

<style scoped>
.study-layout {
  display: flex;
  height: 100vh;
  width: 100vw;
  background: #fff;
  overflow: hidden;
  color: #1e293b;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
}

/* 通用动画 */
.spin { animation: spin 1s linear infinite; }
@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }

.fade-enter-active, .fade-leave-active { transition: opacity 0.3s; }
.fade-enter-from, .fade-leave-to { opacity: 0; }

/* 左侧目录树 */
.sidebar {
  width: 320px;
  background: #fcfdfe;
  border-right: 1px solid #edf2f7;
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}

.sidebar-header {
  padding: 24px 20px 16px;
  background: #fff;
}

.back-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: #f1f5f9;
  border: none;
  color: #475569;
  font-size: 13px;
  font-weight: 500;
  padding: 6px 12px;
  border-radius: 6px;
  cursor: pointer;
  margin-bottom: 20px;
  transition: all 0.2s;
}

.back-btn:hover { background: #e2e8f0; color: #0f172a; }

.course-title-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #1e293b;
}

.course-title-wrap h3 { font-size: 17px; font-weight: 700; }
.title-icon { color: #3b82f6; }

.tree-container {
  flex: 1;
  overflow-y: auto;
  padding: 8px 0;
}

.tree-list { list-style: none; padding: 0; margin: 0; }
.sub-list { background: rgba(248, 250, 252, 0.5); }

.node-content {
  display: flex;
  align-items: center;
  padding: 10px 16px;
  cursor: pointer;
  transition: all 0.2s;
  color: #475569;
  font-size: 14px;
  border-left: 3px solid transparent;
}

.node-content:hover { background: #f1f5f9; color: #0f172a; }
.node-content.active {
  background: #eff6ff;
  color: #2563eb;
  font-weight: 600;
  border-left-color: #3b82f6;
}

.sub-content { padding-left: 34px; }
.grand-content { padding-left: 52px; }

.expand-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  color: #94a3b8;
  margin-right: 6px;
}

.expand-icon.placeholder { visibility: hidden; }

.status-icon { margin-right: 8px; flex-shrink: 0; }
.status-icon.completed { color: #10b981; }
.status-icon.unstarted { color: #cbd5e1; }

.node-name { flex: 1; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

.loading-sub { padding: 8px 16px 8px 52px; display: flex; color: #3b82f6; }

/* 中间正文区 */
.main-content {
  flex: 1;
  background: #fff;
  overflow-y: auto;
  position: relative;
}

.content-area {
  max-width: 880px;
  margin: 0 auto;
  padding: 60px 40px;
}

.content-header { margin-bottom: 40px; }

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #94a3b8;
  margin-bottom: 12px;
}

.content-header h1 {
  font-size: 32px;
  font-weight: 800;
  color: #0f172a;
  margin-bottom: 20px;
  line-height: 1.2;
}

.meta-info { display: flex; align-items: center; justify-content: space-between; }

.difficulty-tags { display: flex; gap: 8px; }

.diff-tag {
  font-size: 12px;
  padding: 2px 10px;
  border-radius: 4px;
  font-weight: 600;
}

.diff-tag.easy { background: #d1fae5; color: #065f46; }
.diff-tag.medium { background: #fef3c7; color: #92400e; }
.diff-tag.hard { background: #fee2e2; color: #991b1b; }

.status-badge {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 600;
  padding: 4px 12px;
  border-radius: 20px;
}

.status-badge.completed { background: #10b981; color: #fff; }
.status-badge.learning { background: #3b82f6; color: #fff; }
.status-badge.unstarted { background: #f1f5f9; color: #64748b; }

.content-body {
  font-size: 17px;
  line-height: 1.8;
  color: #334155;
}

.markdown-placeholder {
  white-space: pre-wrap;
  font-family: inherit;
  background: #f8fafc;
  padding: 24px;
  border-radius: 12px;
  border: 1px solid #f1f5f9;
}

/* 右侧笔记区 */
.note-sidebar {
  width: 360px;
  background: #fff;
  border-left: 1px solid #edf2f7;
  display: flex;
  flex-direction: column;
}

.note-header {
  padding: 24px 20px;
  border-bottom: 1px solid #f1f5f9;
}

.note-title {
  display: flex;
  align-items: center;
  gap: 10px;
  color: #7c3aed;
}

.note-title h3 { font-size: 17px; font-weight: 700; color: #1e293b; }

.note-main { flex: 1; display: flex; flex-direction: column; padding: 20px; background: #fcfdfe; }

.note-login-guide, .note-empty, .empty-view {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
}

.guide-icon { font-size: 40px; margin-bottom: 16px; }
.note-login-guide p { font-weight: 600; color: #1e293b; margin-bottom: 4px; }
.sub-p { font-size: 13px; color: #94a3b8; margin-bottom: 20px; }

.primary-btn {
  background: #7c3aed;
  color: #fff;
  border: none;
  padding: 10px 24px;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
}

.note-active { flex: 1; display: flex; flex-direction: column; gap: 16px; }

.editor-wrap {
  flex: 1;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.02);
  display: flex;
}

.note-textarea {
  flex: 1;
  border: none;
  resize: none;
  padding: 16px;
  font-size: 15px;
  line-height: 1.6;
  color: #334155;
  outline: none;
}

.editor-footer { display: flex; flex-direction: column; gap: 12px; }

.save-status {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #94a3b8;
}

.action-btn {
  width: 100%;
  padding: 10px;
  border-radius: 8px;
  font-weight: 600;
  border: none;
}

.action-btn.disabled { background: #f1f5f9; color: #cbd5e1; cursor: not-allowed; }

/* 骨架屏 */
.content-skeleton { padding: 60px 40px; max-width: 800px; margin: 0 auto; width: 100%; }
.skeleton-header { height: 40px; width: 60%; background: #f1f5f9; border-radius: 8px; margin-bottom: 24px; }
.skeleton-line { height: 16px; width: 100%; background: #f1f5f9; border-radius: 4px; margin-bottom: 12px; }
.skeleton-line.short { width: 40%; }
.skeleton-body { height: 300px; width: 100%; background: #f1f5f9; border-radius: 12px; margin-top: 40px; }

.empty-view h2 { margin-top: 20px; font-size: 22px; font-weight: 700; color: #1e293b; }
.empty-view p { color: #94a3b8; margin-top: 8px; max-width: 280px; }
.empty-illustration { color: #e2e8f0; }
</style>