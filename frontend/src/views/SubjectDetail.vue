<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import markdownit from 'markdown-it'
import hljs from 'highlight.js'
import 'highlight.js/styles/github.css' // 切换回浅色主题，适合灰色背景
import { 
  getTopNodes, getChildNodes, getNodeDetail, getNodeNote,
  updateNodeStatus
} from '../api/node'
import type { SubjectNode, SubjectNodeDetail, NodeNote } from '../types/node'
import TreeItem from '../components/TreeItem.vue'
import { 
  FileText, ArrowLeft, 
  Edit3, BookOpen, 
  Clock, Award, BookOpenCheck, Loader2,
  LayoutList, LayoutPanelLeft, CheckCircle2
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
const expandMode = ref<'normal' | 'accordion'>('normal') // 模式一：normal, 模式二：accordion

const fetchTopNodes = async () => {
  loadingTree.value = true
  try {
    const res = await getTopNodes(subjectId)
    if (res.data?.code === 200 && res.data.data) {
      topNodes.value = res.data.data.map(node => ({ ...node, expanded: false, children: [] }))
      if (topNodes.value.length > 0) {
        await selectNode(topNodes.value[0].id)
      }
    }
  } catch (error) {
    console.error('获取顶级节点失败', error)
  } finally {
    loadingTree.value = false
  }
}

const ensureChildrenLoaded = async (node: TreeNode) => {
  if (node.isLeaf === 1) return
  if (node.children && node.children.length > 0) return

  node.loadingChildren = true
  try {
    const res = await getChildNodes(node.id)
    if (res.data?.code === 200 && res.data.data) {
      node.children = res.data.data.map(child => ({ ...child, expanded: false, children: [] }))
    } else {
      node.children = []
    }
  } catch (error) {
    console.error('获取子节点失败', error)
    node.children = []
  } finally {
    node.loadingChildren = false
  }
}

const findParentAndSiblings = (nodes: TreeNode[], targetId: number): { parent: TreeNode | null, siblings: TreeNode[] } | null => {
  for (const node of nodes) {
    if (node.children) {
      const found = node.children.find(c => c.id === targetId)
      if (found) {
        return { parent: node, siblings: node.children }
      }
      const result = findParentAndSiblings(node.children, targetId)
      if (result) return result
    }
  }
  // 检查是否在顶级节点中
  if (nodes.find(n => n.id === targetId)) {
    return { parent: null, siblings: nodes }
  }
  return null
}

const toggleExpand = async (node: TreeNode) => {
  if (node.isLeaf === 1) return
  
  const targetExpanded = !node.expanded
  
  // 模式二：手风琴模式，且当前是准备展开时
  if (expandMode.value === 'accordion' && targetExpanded) {
    const result = findParentAndSiblings(topNodes.value, node.id)
    if (result) {
      // 收起同级其他节点
      result.siblings.forEach(s => {
        if (s.id !== node.id) {
          s.expanded = false
        }
      })
    }
  }

  node.expanded = targetExpanded
  
  if (node.expanded) {
    await ensureChildrenLoaded(node)
  }
}

const handleNodeClick = async (node: TreeNode) => {
  await selectNode(node.id)
  if (node.isLeaf === 0) {
    node.expanded = true
    await ensureChildrenLoaded(node)
  }
}

// ----------------- 正文和笔记相关 -----------------
// 初始化 Markdown 解析器
const md = markdownit({
  html: true,
  linkify: true,
  typographer: true,
  highlight: (str: string, lang: string, attrs: string) => {
    if (lang && hljs.getLanguage(lang)) {
      try {
        return hljs.highlight(str, { language: lang, ignoreIllegals: true }).value;
      } catch (__) {}
    }
    return ''; // 使用默认转义
  }
})

// 统一包装代码块，应用自定义样式
const defaultFence = md.renderer.rules.fence || function(tokens, idx, options, env, slf) {
  return slf.renderToken(tokens, idx, options);
};

md.renderer.rules.fence = (tokens, idx, options, env, slf) => {
  const token = tokens[idx];
  const info = token.info ? token.info.trim() : '';
  const langName = info.split(/\s+/g)[0];
  
  let highlighted = '';
  if (options.highlight) {
    // markdown-it 14.x options.highlight expects 3 arguments: (str, lang, attrs)
    highlighted = options.highlight(token.content, langName, '') || '';
  }

  if (!highlighted) {
    highlighted = md.utils.escapeHtml(token.content);
  }

  return `<pre class="hljs"><code>${highlighted}</code></pre>\n`;
}

// 解码 HTML 实体（如 &#x20;）
const decodeEntities = (text: string) => {
  if (!text) return ''
  return text
    .replace(/&#x([0-9a-fA-F]+);/g, (_, hex) => String.fromCharCode(parseInt(hex, 16)))
    .replace(/&#([0-9]+);/g, (_, dec) => String.fromCharCode(parseInt(dec, 10)))
    .replace(/&nbsp;/g, ' ')
    .replace(/&lt;/g, '<')
    .replace(/&gt;/g, '>')
    .replace(/&amp;/g, '&')
    .replace(/&quot;/g, '"')
    .replace(/&apos;/g, "'")
}

const currentNodeId = ref<number | null>(null)
const nodeDetail = ref<SubjectNodeDetail | null>(null)
const nodeNote = ref<NodeNote | null>(null)
const loadingDetail = ref(false)
const updatingStatus = ref(false)

const renderedMarkdown = computed(() => {
  if (!nodeDetail.value?.content) return ''
  // 渲染前先解码可能存在的乱码实体
  const decodedContent = decodeEntities(nodeDetail.value.content)
  return md.render(decodedContent)
})

// 缓存对象：以 nodeId 为 key
const detailCache = new Map<number, SubjectNodeDetail>()
const noteCache = new Map<number, NodeNote | null>()

const selectNode = async (id: number) => {
  currentNodeId.value = id
  
  // 1. 检查缓存：如果命中，直接使用并返回
  if (detailCache.has(id)) {
    console.log(`[Cache Hit] Node ID: ${id}`)
    nodeDetail.value = detailCache.get(id) || null
    nodeNote.value = noteCache.get(id) || null
    return
  }

  loadingDetail.value = true
  
  try {
    // 并发请求详情和笔记（如果已登录）
    const detailPromise = getNodeDetail(id)
    const notePromise = isLoggedIn.value ? getNodeNote(id) : Promise.resolve({ data: { code: 200, data: null } })

    const [resDetail, resNote] = await Promise.all([detailPromise, notePromise])

    // 处理详情
    console.log('[GetNodeDetail]', { nodeId: id, response: resDetail.data })
    if (resDetail.data?.code === 200 && resDetail.data.data) {
      const detail = resDetail.data.data
      nodeDetail.value = detail
      detailCache.set(id, detail) // 存入缓存
    } else {
      nodeDetail.value = null
    }

    // 处理笔记
    if (isLoggedIn.value) {
      console.log('[GetNodeNote]', { nodeId: id, response: resNote.data })
      if (resNote.data?.code === 200 && resNote.data.data) {
        const note = resNote.data.data
        nodeNote.value = note
        noteCache.set(id, note) // 存入缓存
      } else {
        nodeNote.value = null
        noteCache.set(id, null) // 即使没有笔记也缓存 null，避免重复请求
      }
    } else {
      nodeNote.value = null
    }
  } catch (error) {
    console.error('获取详情或笔记失败', error)
  } finally {
    loadingDetail.value = false
  }
}

const updateStatus = async () => {
  if (!currentNodeId.value || updatingStatus.value) return
  
  updatingStatus.value = true
  try {
    const res = await updateNodeStatus(currentNodeId.value, 'completed')
    if (res.data?.code === 200) {
      // 1. 更新当前详情状态
      if (nodeDetail.value) {
        nodeDetail.value.userProgressStatus = 'completed'
      }
      
      // 2. 更新树形目录中的节点状态
      const updateNodeInTree = (nodes: TreeNode[]) => {
        for (const node of nodes) {
          if (node.id === currentNodeId.value) {
            node.userProgressStatus = 'completed'
            return true
          }
          if (node.children && updateNodeInTree(node.children)) {
            return true
          }
        }
        return false
      }
      updateNodeInTree(topNodes.value)
      
      // 3. 更新缓存
      if (detailCache.has(currentNodeId.value)) {
        const cached = detailCache.get(currentNodeId.value)
        if (cached) cached.userProgressStatus = 'completed'
      }
    }
  } catch (error) {
    console.error('更新学习状态失败', error)
  } finally {
    updatingStatus.value = false
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
          <div class="mode-switch-btns">
            <button 
              class="mode-btn" 
              :class="{ active: expandMode === 'normal' }"
              title="自由模式：可展开多个节点"
              @click="expandMode = 'normal'"
            >
              <LayoutList :size="16" />
            </button>
            <button 
              class="mode-btn" 
              :class="{ active: expandMode === 'accordion' }"
              title="手风琴模式：展开时自动收起同级"
              @click="expandMode = 'accordion'"
            >
              <LayoutPanelLeft :size="16" />
            </button>
          </div>
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
          <TreeItem
            v-for="node in topNodes"
            :key="node.id"
            :node="node"
            :level="0"
            :active-id="currentNodeId"
            @node-click="handleNodeClick"
            @toggle-expand="toggleExpand"
          />
        </ul>
      </div>
    </aside>

    <!-- 中间正文区 -->
    <main class="main-content">
      <div v-if="!nodeDetail && loadingDetail" class="content-skeleton">
        <div class="skeleton-header"></div>
        <div class="skeleton-line"></div>
        <div class="skeleton-line short"></div>
        <div class="skeleton-body"></div>
      </div>
      <div v-else-if="nodeDetail" class="content-area">
        <div v-if="loadingDetail" class="content-loading-overlay">
          <Loader2 class="spin" :size="28" />
        </div>
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
            <div 
              class="markdown-content" 
              v-html="renderedMarkdown"
            ></div>
          </div>
          
          <!-- 叶子节点底部操作区 -->
          <div v-if="nodeDetail.isLeaf === 1" class="article-footer">
            <div v-if="nodeDetail.userProgressStatus === 'completed'" class="completed-status">
              <CheckCircle2 :size="20" />
              <span>学习完毕</span>
            </div>
            <button 
              v-else 
              class="complete-btn" 
              :disabled="updatingStatus"
              @click="updateStatus"
            >
              <Loader2 v-if="updatingStatus" class="spin" :size="18" />
              <BookOpenCheck v-else :size="18" />
              <span>我已学完</span>
            </button>
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
              class="note-textarea" 
              :value="nodeNote?.noteContent || ''"
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
  justify-content: space-between;
}

.mode-switch-btns {
  display: flex;
  background: #f1f5f9;
  padding: 3px;
  border-radius: 6px;
  gap: 2px;
}

.mode-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: none;
  background: transparent;
  color: #64748b;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
}

.mode-btn:hover {
  color: #0f172a;
}

.mode-btn.active {
  background: #fff;
  color: #3b82f6;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.course-title-wrap h3 { font-size: 17px; font-weight: 700; flex: 1; }
.title-icon { color: #3b82f6; }

.tree-container {
  flex: 1;
  overflow-y: auto;
  padding: 8px 0;
}

.tree-list { list-style: none; padding: 0; margin: 0; }

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
  position: relative;
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
  font-size: 16px;
  line-height: 1.8;
  color: #334155;
}

/* Markdown 样式增强 */
.markdown-content {
  line-height: 1.8;
  word-wrap: break-word;
}

.markdown-content :deep(h1) { font-size: 2em; margin-bottom: 16px; padding-bottom: 8px; border-bottom: 1px solid #edf2f7; }
.markdown-content :deep(h2) { font-size: 1.5em; margin: 24px 0 16px; padding-bottom: 4px; border-bottom: 1px solid #edf2f7; }
.markdown-content :deep(h3) { font-size: 1.25em; margin: 20px 0 12px; }
.markdown-content :deep(p) { margin-bottom: 16px; }
.markdown-content :deep(ul), .markdown-content :deep(ol) { padding-left: 2em; margin-bottom: 16px; }
.markdown-content :deep(li) { margin-bottom: 4px; }
.markdown-content :deep(code) { 
  background: #f1f5f9; /* 浅灰色背景 */
  padding: 2px 6px; 
  border-radius: 4px; 
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
  font-size: 0.9em;
  color: #334155; /* 深灰色文字，清晰可读 */
}
.markdown-content :deep(pre) {
  padding: 16px;
  border-radius: 12px;
  overflow-x: auto;
  margin: 20px 0;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
}
.markdown-content :deep(pre.hljs) {
  background: #f6f8fa; /* 经典的灰色背景 */
  color: #1e293b;    /* 深色文字 */
  border: 1px solid #e2e8f0; /* 添加一个淡边框，增加质感 */
}
.markdown-content :deep(pre code) {
  padding: 0;
  background: transparent;
  font-size: 14px;
  line-height: 1.5;
  color: inherit;
}
.markdown-content :deep(blockquote) {
  border-left: 4px solid #e2e8f0;
  padding-left: 16px;
  color: #64748b;
  margin: 16px 0;
  font-style: italic;
}
.markdown-content :deep(img) {
  max-width: 100%;
  border-radius: 8px;
  margin: 16px 0;
}
.markdown-content :deep(table) {
  width: 100%;
  border-collapse: collapse;
  margin: 20px 0;
}
.markdown-content :deep(th), .markdown-content :deep(td) {
  border: 1px solid #e2e8f0;
  padding: 12px;
  text-align: left;
}
.markdown-content :deep(th) {
  background: #f8fafc;
}

.content-loading-overlay {
  position: absolute;
  top: 24px;
  right: 24px;
  width: 44px;
  height: 44px;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(226, 232, 240, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #3b82f6;
  box-shadow: 0 10px 18px rgba(15, 23, 42, 0.08);
  pointer-events: none;
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

.markdown-content :deep(a:hover) {
  text-decoration: underline;
}

/* 底部操作区 */
.article-footer {
  margin-top: 60px;
  padding-top: 40px;
  border-top: 1px solid #f1f5f9;
  display: flex;
  justify-content: center;
}

.complete-btn {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  background: #3b82f6;
  color: #fff;
  border: none;
  padding: 14px 40px;
  border-radius: 12px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
}

.complete-btn:hover:not(:disabled) {
  background: #2563eb;
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(59, 130, 246, 0.4);
}

.complete-btn:active:not(:disabled) {
  transform: translateY(0);
}

.complete-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.completed-status {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: #10b981;
  background: #f0fdf4;
  padding: 12px 32px;
  border-radius: 12px;
  font-weight: 600;
  font-size: 16px;
  border: 1px solid #dcfce7;
}

/* 中间正文区骨架屏 */
.content-skeleton { padding: 60px 40px; max-width: 800px; margin: 0 auto; width: 100%; }
.skeleton-header { height: 40px; width: 60%; background: #f1f5f9; border-radius: 8px; margin-bottom: 24px; }
.skeleton-line { height: 16px; width: 100%; background: #f1f5f9; border-radius: 4px; margin-bottom: 12px; }
.skeleton-line.short { width: 40%; }
.skeleton-body { height: 300px; width: 100%; background: #f1f5f9; border-radius: 12px; margin-top: 40px; }

.empty-view h2 { margin-top: 20px; font-size: 22px; font-weight: 700; color: #1e293b; }
.empty-view p { color: #94a3b8; margin-top: 8px; max-width: 280px; }
.empty-illustration { color: #e2e8f0; }
</style>
