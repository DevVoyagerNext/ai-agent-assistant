<script setup lang="ts">
import { ref, onMounted, computed, reactive } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import markdownit from 'markdown-it'
import hljs from 'highlight.js'
import 'highlight.js/styles/github.css' // 切换回浅色主题，适合灰色背景
import { 
  getTopNodes, getChildNodes, getNodeDetail, getNodeNote,
  updateNodeStatus, updateNodeDifficulty
} from '../api/node'
import { getSubjectDetail } from '../api/subject'
import { 
  likeSubject, 
  getUserCollectFolders, 
  createCollectFolder, 
  addSubjectToFolder 
} from '../api/user'
import type { SubjectNode, SubjectNodeDetail, NodeNote } from '../types/node'
import type { Subject } from '../types/subject'
import type { CollectFolderRes } from '../types/user'
import TreeItem from '../components/TreeItem.vue'
import Toast from '../components/Toast.vue'
import {
  getNodeProgressStatusLabel,
  normalizeNodeProgressStatus,
  type NodeProgressStatus
} from '../utils/nodeProgress'
import { 
  FileText, ArrowLeft, 
  Edit3, BookOpen, 
  Clock, Award, BookOpenCheck, Loader2,
  LayoutList, LayoutPanelLeft, CheckCircle2,
  Heart, Bookmark, Plus, X, Globe, Lock
} from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()
const subjectId = Number(route.params.id)
const isLoggedIn = computed(() => !!localStorage.getItem('token'))
const isLiked = ref(false)
const isCollected = ref(false)
const updatingLike = ref(false)

// ----------------- 收藏相关 -----------------
const showCollectModal = ref(false)
const folders = ref<CollectFolderRes[]>([])
const loadingFolders = ref(false)
const showCreateFolder = ref(false)
const newFolderName = ref('')
const newFolderDesc = ref('')
const newFolderIsPublic = ref(0)
const creatingFolder = ref(false)
const addingToFolder = ref<number | null>(null)

// 获取教材详情（获取点赞、收藏状态）
const fetchSubjectDetail = async () => {
  if (!isLoggedIn.value) return
  try {
    const res = await getSubjectDetail(subjectId)
    if (res.data?.code === 200 && res.data.data) {
      isLiked.value = res.data.data.isLiked
      isCollected.value = res.data.data.isCollected
    }
  } catch (error) {
    console.error('获取教材详情失败', error)
  }
}

const handleCollectClick = async () => {
  if (!isLoggedIn.value) {
    showToast('请登录后再收藏', 'error')
    router.push('/login')
    return
  }
  
  showCollectModal.value = true
  fetchFolders()
}

const fetchFolders = async () => {
  loadingFolders.value = true
  try {
    const res = await getUserCollectFolders()
    if (res.data?.code === 200 && res.data.data) {
      folders.value = res.data.data
    }
  } catch (error) {
    console.error('获取收藏夹失败', error)
  } finally {
    loadingFolders.value = false
  }
}

const handleCreateFolder = async () => {
  if (!newFolderName.value.trim()) {
    showToast('请输入文件夹名称', 'error')
    return
  }
  
  creatingFolder.value = true
  try {
    const res = await createCollectFolder({
      name: newFolderName.value,
      description: newFolderDesc.value,
      isPublic: newFolderIsPublic.value
    })
    if (res.data?.code === 200) {
      showToast('创建成功')
      showCreateFolder.value = false
      newFolderName.value = ''
      newFolderDesc.value = ''
      fetchFolders()
    }
  } catch (error: any) {
    showToast(error.response?.data?.msg || '创建失败', 'error')
  } finally {
    creatingFolder.value = false
  }
}

const handleAddToFolder = async (folderId: number) => {
  addingToFolder.value = folderId
  try {
    const res = await addSubjectToFolder(folderId, subjectId)
    if (res.data?.code === 200) {
      showToast('已添加到收藏夹')
      isCollected.value = true
      showCollectModal.value = false
    }
  } catch (error: any) {
    showToast(error.response?.data?.msg || '添加失败', 'error')
  } finally {
    addingToFolder.value = null
  }
}

const handleLikeClick = async () => {
  if (!isLoggedIn.value) {
    showToast('请登录后再点赞', 'error')
    router.push('/login')
    return
  }

  if (updatingLike.value) return
  
  updatingLike.value = true
  try {
    const res = await likeSubject(subjectId)
    if (res.data?.code === 200 && res.data.data) {
      isLiked.value = res.data.data.isLiked
      showToast(isLiked.value ? '已点赞' : '已取消点赞')
    }
  } catch (error: any) {
    console.error('点赞失败', error)
    showToast(error.response?.data?.msg || '操作失败', 'error')
  } finally {
    updatingLike.value = false
  }
}

// ----------------- 目录树相关 -----------------
interface TreeNode extends SubjectNode {
  expanded?: boolean;
  children?: TreeNode[];
  loadingChildren?: boolean;
}

const topNodes = ref<TreeNode[]>([])
const loadingTree = ref(false)
const expandMode = ref<'normal' | 'accordion'>('normal') // 模式一：normal, 模式二：accordion

const toggleExpandMode = () => {
  expandMode.value = expandMode.value === 'normal' ? 'accordion' : 'normal'
  showToast(`已切换为${expandMode.value === 'normal' ? '自由' : '手风琴'}模式`)
}

const fetchTopNodes = async () => {
  loadingTree.value = true
  try {
    const res = await getTopNodes(subjectId)
    if (res.data?.code === 200 && res.data.data) {
      topNodes.value = res.data.data.map(node => ({ 
        ...node, 
        expanded: false, 
        children: [],
        userProgressStatus: normalizeNodeProgressStatus(node.userProgressStatus)
      }))
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
      node.children = res.data.data.map(child => ({ 
        ...child, 
        expanded: false, 
        children: [],
        userProgressStatus: normalizeNodeProgressStatus(child.userProgressStatus)
      }))
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
    // 如果节点当前是收起状态，则需要执行展开逻辑（包含手风琴模式的处理）
    if (!node.expanded) {
      await toggleExpand(node)
    }
  }
}

// ----------------- 正文和笔记相关 -----------------
// 初始化 Markdown 解析器
const md = markdownit({
  html: true,
  linkify: true,
  typographer: true,
  highlight: (str: string, lang: string, _attrs: string) => {
    if (lang && hljs.getLanguage(lang)) {
      try {
        return hljs.highlight(str, { language: lang, ignoreIllegals: true }).value;
      } catch (__) {}
    }
    return ''; // 使用默认转义
  }
})

// 统一包装代码块，应用自定义样式
md.renderer.rules.fence = (tokens, idx, options, _env, _slf) => {
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
const updatingDifficulty = ref(false)

// ----------------- 反馈弹窗相关 -----------------
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

const renderedMarkdown = computed(() => {
  if (!nodeDetail.value?.content) return ''
  // 渲染前先解码可能存在的乱码实体
  const decodedContent = decodeEntities(nodeDetail.value.content)
  return md.render(decodedContent)
})

const currentNodeStatus = computed<NodeProgressStatus>(() => (
  normalizeNodeProgressStatus(nodeDetail.value?.userProgressStatus)
))

const currentNodeStatusLabel = computed(() => getNodeProgressStatusLabel(currentNodeStatus.value))

// 缓存对象：以 nodeId 为 key
const detailCache = new Map<number, SubjectNodeDetail>()
const noteCache = new Map<number, NodeNote | null>()

// 内部函数：同步更新所有地方的节点状态
const syncNodeStatus = (id: number, status: NodeProgressStatus) => {
  console.log(`[SyncNodeStatus] NodeID: ${id}, NewStatus: ${status}`)
  // 1. 更新当前详情状态
  if (currentNodeId.value === id && nodeDetail.value) {
    nodeDetail.value.userProgressStatus = status
  }
  
  // 2. 更新树形目录中的节点状态
  const updateNodeInTree = (nodes: TreeNode[]) => {
    for (const node of nodes) {
      if (node.id === id) {
        console.log(`[SyncNodeStatus] Found in tree, updating status.`, node.name)
        node.userProgressStatus = status
        return true
      }
      if (node.children && updateNodeInTree(node.children)) {
        return true
      }
    }
    return false
  }
  const found = updateNodeInTree(topNodes.value)
  if (!found) {
    console.warn(`[SyncNodeStatus] NodeID: ${id} not found in current tree structure.`)
  }
  
  // 3. 更新缓存
  if (detailCache.has(id)) {
    const cached = detailCache.get(id)
    if (cached) {
      cached.userProgressStatus = status
    }
  }
}

const selectNode = async (id: number) => {
  currentNodeId.value = id
  
  // 1. 检查缓存：如果命中，直接使用并返回
  if (detailCache.has(id)) {
    console.log(`[Cache Hit] Node ID: ${id}`)
    nodeDetail.value = detailCache.get(id) || null
    nodeNote.value = noteCache.get(id) || null
    
    // 缓存命中的情况下，如果状态是未开始，也需要自动触发“学习中”
    if (nodeDetail.value?.userProgressStatus === 'unstarted' && isLoggedIn.value) {
      autoStartLearning(id)
    }
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

      // 优先继承目录树中已存在的状态，避免接口返回异常值导致状态颜色丢失。
      let existingStatus: NodeProgressStatus = 'unstarted'
      const findStatusInTree = (nodes: TreeNode[]) => {
        for (const node of nodes) {
          if (node.id === id) {
            existingStatus = normalizeNodeProgressStatus(node.userProgressStatus)
            return true
          }
          if (node.children && findStatusInTree(node.children)) return true
        }
        return false
      }
      findStatusInTree(topNodes.value)

      detail.userProgressStatus = normalizeNodeProgressStatus(detail.userProgressStatus, existingStatus)
      
      nodeDetail.value = detail
      detailCache.set(id, detail) // 存入缓存
      
      // 如果获取到的详情显示状态为“未开始”，且已登录，则自动标记为“学习中”
      if (detail.userProgressStatus === 'unstarted' && isLoggedIn.value) {
        autoStartLearning(id)
      }
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

// 自动将“未开始”状态更新为“学习中”
const autoStartLearning = async (id: number) => {
  try {
    const res = await updateNodeStatus(id, 'learning')
    if (res.data?.code === 200) {
      syncNodeStatus(id, 'learning')
    }
  } catch (error) {
    console.error(`[AutoStartLearning] 自动更新状态为学习中失败 (NodeID: ${id})`, error)
  }
}

const updateStatus = async () => {
  if (!currentNodeId.value || updatingStatus.value) return
  
  updatingStatus.value = true
  try {
    const res = await updateNodeStatus(currentNodeId.value, 'completed')
    if (res.data?.code === 200) {
      showToast('恭喜！你已完成该知识点的学习')
      syncNodeStatus(currentNodeId.value, 'completed')
    } else {
      showToast(res.data?.msg || '更新状态失败', 'error')
    }
  } catch (error: any) {
    console.error('更新学习状态失败', error)
    showToast(error.response?.data?.msg || '服务器连接失败', 'error')
  } finally {
    updatingStatus.value = false
  }
}

// 刷新当前节点的难度统计数据
const refreshNodeCounts = async (id: number) => {
  try {
    const res = await getNodeDetail(id)
    if (res.data?.code === 200 && res.data.data) {
      const detail = res.data.data
      
      // 更新当前显示的详情中的计数值
      if (nodeDetail.value && nodeDetail.value.id === id) {
        nodeDetail.value.easyCount = detail.easyCount
        nodeDetail.value.mediumCount = detail.mediumCount
        nodeDetail.value.hardCount = detail.hardCount
      }
      
      // 同时更新缓存中的数据
      const cached = detailCache.get(id)
      if (cached) {
        cached.easyCount = detail.easyCount
        cached.mediumCount = detail.mediumCount
        cached.hardCount = detail.hardCount
      }
    }
  } catch (error) {
    console.error('刷新难度计数失败', error)
  }
}

const handleDifficultyClick = async (difficulty: 'easy' | 'medium' | 'hard') => {
  if (!isLoggedIn.value) {
    showToast('请登录后再评价难度', 'error')
    router.push('/login')
    return
  }
  
  if (!currentNodeId.value || updatingDifficulty.value) return
  
  updatingDifficulty.value = true
  try {
    const res = await updateNodeDifficulty(currentNodeId.value, difficulty)
    if (res.data?.code === 200) {
      showToast('评价成功！感谢你的反馈')
    } else {
      showToast(res.data?.msg || '评价失败', 'error')
    }
  } catch (error: any) {
    // 处理异常，特别是“已经评价过”的业务提示
    const msg = error.response?.data?.msg || '评价失败'
    showToast(msg, 'error')
    console.error('评价难度失败', error)
  } finally {
    // 无论评价成功还是报错（可能是已评价过），都从后端拉取最新计数来刷新数据
    if (currentNodeId.value) {
      await refreshNodeCounts(currentNodeId.value)
    }
    updatingDifficulty.value = false
  }
}

onMounted(() => {
  if (!subjectId) {
    router.replace('/')
    return
  }
  fetchSubjectDetail()
  fetchTopNodes()
})

// ----------------- 布局拖拽相关 -----------------
const sidebarWidth = ref(320)
const noteSidebarWidth = ref(360)
const isResizingLeft = ref(false)
const isResizingRight = ref(false)

const startResizingLeft = (e: MouseEvent) => {
  isResizingLeft.value = true
  document.addEventListener('mousemove', handleMouseMove)
  document.addEventListener('mouseup', stopResizing)
  document.body.style.cursor = 'col-resize'
  document.body.style.userSelect = 'none'
}

const startResizingRight = (e: MouseEvent) => {
  isResizingRight.value = true
  document.addEventListener('mousemove', handleMouseMove)
  document.addEventListener('mouseup', stopResizing)
  document.body.style.cursor = 'col-resize'
  document.body.style.userSelect = 'none'
}

const handleMouseMove = (e: MouseEvent) => {
  if (isResizingLeft.value) {
    const newWidth = e.clientX
    if (newWidth > 200 && newWidth < 600) {
      sidebarWidth.value = newWidth
    }
  } else if (isResizingRight.value) {
    const newWidth = window.innerWidth - e.clientX
    if (newWidth > 250 && newWidth < 600) {
      noteSidebarWidth.value = newWidth
    }
  }
}

const stopResizing = () => {
  isResizingLeft.value = false
  isResizingRight.value = false
  document.removeEventListener('mousemove', handleMouseMove)
  document.removeEventListener('mouseup', stopResizing)
  document.body.style.cursor = ''
  document.body.style.userSelect = ''
}

const goBack = () => {
  router.push('/')
}
</script>

<template>
  <div class="study-layout">
    <Toast 
      v-if="toast.show" 
      :message="toast.message" 
      :type="toast.type" 
      @close="toast.show = false" 
    />
    <!-- 左侧目录树 -->
    <aside class="sidebar" :style="{ width: sidebarWidth + 'px' }">
      <div class="sidebar-header">
        <button class="back-btn" @click="goBack">
          <ArrowLeft :size="18" />
          <span>返回大厅</span>
        </button>
        <div class="course-title-wrap">
          <BookOpenCheck class="title-icon" :size="20" />
          <h3>教材目录</h3>
          <button 
             class="mode-toggle-btn" 
             :title="expandMode === 'normal' ? '当前模式：自由模式 (点击切换为手风琴)' : '当前模式：手风琴模式 (点击切换为自由)'"
             @click="toggleExpandMode"
           >
             <LayoutList v-if="expandMode === 'normal'" :size="16" />
             <LayoutPanelLeft v-else :size="16" />
           </button>
          <button 
            class="mode-toggle-btn like-btn" 
            :class="{ liked: isLiked }"
            :title="isLiked ? '取消点赞' : '点赞教材'"
            @click="handleLikeClick"
          >
            <Heart :size="16" :fill="isLiked ? 'currentColor' : 'none'" />
          </button>
          <button 
            class="mode-toggle-btn collect-btn" 
            :class="{ collected: isCollected }"
            :title="isCollected ? '已收藏' : '收藏教材'"
            @click="handleCollectClick"
          >
            <Bookmark :size="16" :fill="isCollected ? 'currentColor' : 'none'" />
          </button>
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

    <!-- 左侧拉伸条 -->
    <div 
      class="resizer left-resizer" 
      :class="{ active: isResizingLeft }"
      @mousedown="startResizingLeft"
    ></div>

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
              <span 
                class="diff-tag easy" 
                title="点击评价：太简单了" 
                @click="handleDifficultyClick('easy')"
              >易 {{ nodeDetail.easyCount }}</span>
              <span 
                class="diff-tag medium" 
                title="点击评价：正合适" 
                @click="handleDifficultyClick('medium')"
              >中 {{ nodeDetail.mediumCount }}</span>
              <span 
                class="diff-tag hard" 
                title="点击评价：太难了" 
                @click="handleDifficultyClick('hard')"
              >难 {{ nodeDetail.hardCount }}</span>
            </div>
            <div class="status-badge" :class="currentNodeStatus">
              <Award v-if="currentNodeStatus === 'completed'" :size="14" />
              <Clock v-else :size="14" />
              <span>{{ currentNodeStatusLabel }}</span>
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
            <div v-if="currentNodeStatus === 'completed'" class="completed-status">
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

    <!-- 右侧拉伸条 -->
    <div 
      class="resizer right-resizer" 
      :class="{ active: isResizingRight }"
      @mousedown="startResizingRight"
    ></div>

    <!-- 右侧随堂笔记区 -->
    <aside class="note-sidebar" :style="{ width: noteSidebarWidth + 'px' }">
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

    <!-- 收藏弹窗 -->
    <Teleport to="body">
      <div v-if="showCollectModal" class="modal-overlay" @click.self="showCollectModal = false">
        <div class="modal-content collect-modal">
          <header class="modal-header">
            <h3>添加到收藏夹</h3>
            <button class="close-btn" @click="showCollectModal = false">
              <X :size="20" />
            </button>
          </header>

          <div class="modal-body">
            <div v-if="loadingFolders" class="loading-folders">
              <Loader2 class="spin" :size="24" />
              <span>正在加载收藏夹...</span>
            </div>
            <div v-else class="folder-list">
              <div 
                v-for="folder in folders" 
                :key="folder.id" 
                class="folder-item"
                @click="handleAddToFolder(folder.id)"
              >
                <div class="folder-info">
                  <Bookmark :size="18" class="folder-icon" />
                  <div class="folder-detail">
                    <span class="folder-name">{{ folder.name }}</span>
                    <span class="folder-desc">{{ folder.description || '暂无描述' }}</span>
                  </div>
                </div>
                <div class="folder-meta">
                  <Globe v-if="folder.isPublic" :size="14" title="公开" />
                  <Lock v-else :size="14" title="私有" />
                  <Loader2 v-if="addingToFolder === folder.id" class="spin" :size="14" />
                </div>
              </div>

              <div v-if="folders.length === 0 && !showCreateFolder" class="empty-folders">
                <Bookmark :size="48" />
                <p>你还没有收藏夹，快去创建一个吧</p>
              </div>
            </div>

            <!-- 创建文件夹表单 -->
            <div v-if="showCreateFolder" class="create-folder-form">
              <div class="form-group">
                <label>文件夹名称</label>
                <input v-model="newFolderName" type="text" placeholder="例如：我的考研教材" />
              </div>
              <div class="form-group">
                <label>描述</label>
                <textarea v-model="newFolderDesc" placeholder="可选：简单介绍一下这个收藏夹"></textarea>
              </div>
              <div class="form-group row">
                <label>是否公开</label>
                <div class="radio-group">
                  <label>
                    <input v-model="newFolderIsPublic" type="radio" :value="0" />
                    <span>私有</span>
                  </label>
                  <label>
                    <input v-model="newFolderIsPublic" type="radio" :value="1" />
                    <span>公开</span>
                  </label>
                </div>
              </div>
              <div class="form-actions">
                <button class="cancel-btn" @click="showCreateFolder = false">取消</button>
                <button class="confirm-btn" :disabled="creatingFolder" @click="handleCreateFolder">
                  <Loader2 v-if="creatingFolder" class="spin" :size="16" />
                  <span>创建并保存</span>
                </button>
              </div>
            </div>
          </div>

          <footer v-if="!showCreateFolder" class="modal-footer">
            <button class="create-btn-trigger" @click="showCreateFolder = true">
              <Plus :size="16" />
              <span>新建文件夹</span>
            </button>
          </footer>
        </div>
      </div>
    </Teleport>
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
  background: #fcfdfe;
  border-right: 1px solid #edf2f7;
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  min-width: 200px;
  max-width: 600px;
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
  gap: 12px;
  color: #1e293b;
}

.mode-toggle-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  background: #eff6ff;
  border: 1px solid #dbeafe;
  border-radius: 6px;
  color: #3b82f6;
  cursor: pointer;
  transition: all 0.2s;
  flex-shrink: 0;
}

.mode-toggle-btn:hover {
  background: #dbeafe;
  color: #2563eb;
  border-color: #bfdbfe;
}

.mode-toggle-btn.like-btn.liked {
  color: #ef4444;
  background: #fef2f2;
  border-color: #fecaca;
}

.mode-toggle-btn.like-btn.liked:hover {
  background: #fee2e2;
}

.mode-toggle-btn.collect-btn.collected {
  color: #f59e0b;
  background: #fffbeb;
  border-color: #fef3c7;
}

.mode-toggle-btn.collect-btn.collected:hover {
  background: #fef3c7;
}

/* 弹窗样式 */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.4);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: #fff;
  border-radius: 16px;
  width: 100%;
  max-width: 440px;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  max-height: 80vh;
}

.modal-header {
  padding: 20px 24px;
  border-bottom: 1px solid #f1f5f9;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.modal-header h3 {
  font-size: 18px;
  font-weight: 700;
  color: #1e293b;
  margin: 0;
}

.close-btn {
  background: transparent;
  border: none;
  color: #94a3b8;
  cursor: pointer;
  padding: 4px;
  border-radius: 6px;
  transition: all 0.2s;
}

.close-btn:hover {
  background: #f1f5f9;
  color: #64748b;
}

.modal-body {
  padding: 24px;
  overflow-y: auto;
  flex: 1;
}

.loading-folders, .empty-folders {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 0;
  color: #94a3b8;
  gap: 12px;
}

.folder-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.folder-item {
  padding: 16px;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  cursor: pointer;
  transition: all 0.2s;
}

.folder-item:hover {
  border-color: #3b82f6;
  background: #eff6ff;
}

.folder-info {
  display: flex;
  align-items: center;
  gap: 16px;
  flex: 1;
  min-width: 0;
}

.folder-icon {
  color: #f59e0b;
  flex-shrink: 0;
}

.folder-detail {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.folder-name {
  font-weight: 600;
  color: #1e293b;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.folder-desc {
  font-size: 12px;
  color: #64748b;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.folder-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #94a3b8;
  margin-left: 12px;
}

.modal-footer {
  padding: 16px 24px;
  border-top: 1px solid #f1f5f9;
  background: #fcfdfe;
}

.create-btn-trigger {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 12px;
  background: #fff;
  border: 1px dashed #cbd5e1;
  border-radius: 10px;
  color: #64748b;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.create-btn-trigger:hover {
  border-color: #3b82f6;
  color: #3b82f6;
  background: #eff6ff;
}

.create-folder-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-group.row {
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
}

.form-group label {
  font-size: 14px;
  font-weight: 600;
  color: #475569;
}

.form-group input, .form-group textarea {
  padding: 12px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 14px;
  outline: none;
  transition: border-color 0.2s;
}

.form-group input:focus, .form-group textarea:focus {
  border-color: #3b82f6;
}

.form-group textarea {
  height: 80px;
  resize: none;
}

.radio-group {
  display: flex;
  gap: 16px;
}

.radio-group label {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  font-weight: normal;
}

.form-actions {
  display: flex;
  gap: 12px;
  margin-top: 8px;
}

.form-actions button {
  flex: 1;
  padding: 12px;
  border-radius: 10px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.cancel-btn {
  background: #f1f5f9;
  border: none;
  color: #64748b;
}

.cancel-btn:hover {
  background: #e2e8f0;
}

.confirm-btn {
  background: #3b82f6;
  border: none;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.confirm-btn:hover:not(:disabled) {
  background: #2563eb;
}

.confirm-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.course-title-wrap h3 {
  font-size: 17px;
  font-weight: 700;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  min-width: 0;
}

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
  cursor: pointer;
  transition: all 0.2s;
  user-select: none;
}

.diff-tag:hover {
  transform: translateY(-1px);
  filter: brightness(0.95);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.diff-tag:active {
  transform: translateY(0);
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
  background: #fff;
  border-left: 1px solid #edf2f7;
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  min-width: 250px;
  max-width: 600px;
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

/* 拖拽条样式 */
.resizer {
  width: 4px;
  background: transparent;
  cursor: col-resize;
  transition: background 0.2s, width 0.2s;
  z-index: 10;
  flex-shrink: 0;
  position: relative;
}

.resizer:hover, .resizer.active {
  background: #3b82f6;
  width: 4px;
}

/* 增加热区，方便抓取 */
.resizer::after {
  content: '';
  position: absolute;
  top: 0;
  bottom: 0;
  left: -4px;
  right: -4px;
}

.left-resizer {
  margin-right: -4px;
}

.right-resizer {
  margin-left: -4px;
}
</style>
