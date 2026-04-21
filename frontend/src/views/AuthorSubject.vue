<script setup lang="ts">
import { ref, onMounted, computed, watch, nextTick, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Vditor from 'vditor'
import 'vditor/dist/index.css'
import markdownit from 'markdown-it'
import hljs from 'highlight.js'
import 'highlight.js/styles/github.css'
import { 
  getAuthorInitNodes, 
  getAuthorChildNodes, 
  getAuthorNodeContent, 
  updateAuthorNodeName, 
  updateAuthorNodeContent 
} from '../api/node'
import type { AuthorNode, AuthorNodeContent } from '../types/node'
import { 
  ChevronRight, 
  ChevronDown, 
  FileText, 
  Folder, 
  ArrowLeft,
  CheckCircle2,
  Clock
} from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()
const subjectId = Number(route.params.id)

// 节点树状态
const nodes = ref<AuthorNode[]>([])
const lastNodeId = ref<number | null>(null)
const expandedKeys = ref<Set<number>>(new Set())
const loadingTree = ref(true)

// 当前选中节点状态
const activeNodeId = ref<number | null>(null)
const activeNode = computed(() => nodes.value.find(n => n.id === activeNodeId.value) || null)

// 编辑区状态
const loadingContent = ref(false)
const editName = ref('')
const editContent = ref('')
const originalName = ref('')
const originalContent = ref('')
const saving = ref(false)
const saveStatus = ref<'saved' | 'unsaved' | 'saving' | 'error'>('saved')
const contentInfo = ref<AuthorNodeContent | null>(null)

onBeforeUnmount(() => {
  if (vditorInstance) {
    vditorInstance.destroy()
    vditorInstance = null
  }
})

// 控制是否显示已发布内容面板
const showPublishedPane = ref<boolean>(true)

let vditorInstance: Vditor | null = null

const md = markdownit({
  html: true,
  linkify: true,
  typographer: true,
  highlight: (str: string, lang: string, _attrs: string) => {
    if (lang && hljs.getLanguage(lang)) {
      try {
        return hljs.highlight(str, { language: lang, ignoreIllegals: true }).value
      } catch (__) {}
    }
    return '' // 使用默认转义
  }
})

const renderMarkdown = (text: string) => {
  if (!text) return ''
  return md.render(text)
}

const normalizeAuthorNode = (node: AuthorNode): AuthorNode => ({
  ...node,
  id: Number(node.id),
  subjectId: Number(node.subjectId),
  parentId: Number(node.parentId),
  auditStatus: Number(node.auditStatus),
  hasDraft: Number(node.hasDraft),
  isLeaf: Number(node.isLeaf)
})

let saveTimer: any = null

// 初始化加载
onMounted(async () => {
  await fetchInitNodes()
})

const fetchInitNodes = async () => {
  loadingTree.value = true
  try {
    const res = await getAuthorInitNodes(subjectId)
    if (res.data?.code === 200 && res.data.data) {
      nodes.value = (res.data.data.nodeList || []).map(normalizeAuthorNode)
      lastNodeId.value = res.data.data.lastNodeId
      
      // 展开所有路径上的节点
      nodes.value.forEach(node => {
        if (node.parentId === 0) expandedKeys.value.add(node.id)
      })
      
      if (lastNodeId.value) {
        // 自动选中断点节点
        await handleNodeSelect(lastNodeId.value)
        // 确保断点节点的所有父节点都被展开
        const targetNode = nodes.value.find(n => n.id === lastNodeId.value)
        if (targetNode && targetNode.path) {
          const pathIds = targetNode.path.split('/').filter(p => p).map(Number)
          pathIds.forEach(id => expandedKeys.value.add(id))
        }
      }
    } else {
      console.error('获取初始节点失败', res.data?.msg)
    }
  } catch (err) {
    console.error(err)
  } finally {
    loadingTree.value = false
  }
}

// 获取子节点
const fetchChildren = async (parentId: number) => {
  try {
    const res = await getAuthorChildNodes(parentId)
    if (res.data?.code === 200 && res.data.data) {
      const children = res.data.data.map(normalizeAuthorNode)
      
      // 我们需要将最新的 children 更新到 nodes.value 中，保证草稿状态是最新的
      const childrenMap = new Map(children.map(c => [c.id, c]))
      
      const nextNodes = nodes.value.map(n => {
        if (childrenMap.has(n.id)) {
          const updatedNode = childrenMap.get(n.id)!
          childrenMap.delete(n.id)
          return updatedNode
        }
        return n
      })
      
      // 追加全新的节点
      nodes.value = [...nextNodes, ...Array.from(childrenMap.values())]
    }
  } catch (err) {
    console.error('获取子节点失败', err)
  }
}

// 切换展开/折叠
const toggleExpand = async (node: AuthorNode, event: Event) => {
  event.stopPropagation()
  if (node.isLeaf === 1) return // 如果是叶子节点，不处理展开/折叠

  if (expandedKeys.value.has(node.id)) {
    expandedKeys.value.delete(node.id)
  } else {
    expandedKeys.value.add(node.id)
    // 每次展开都去请求最新数据，保证数据是最新的
    await fetchChildren(node.id)
  }
}

// 选中节点
const handleNodeSelect = async (nodeId: number) => {
  const node = nodes.value.find(n => n.id === nodeId)
  if (!node) return

  // 如果是非叶子节点，点击时自动处理展开逻辑
  if (node.isLeaf === 0) {
    if (!expandedKeys.value.has(node.id)) {
      expandedKeys.value.add(node.id)
      // 每次展开都去请求最新数据
      await fetchChildren(node.id)
    } else {
      // 如果已经展开了，再次点击非叶子节点可以选择折叠
      expandedKeys.value.delete(node.id)
    }
  }

  if (activeNodeId.value === nodeId) return
  
  activeNodeId.value = nodeId
  
  editName.value = node.nameDraft || node.name || '未命名节点'
  originalName.value = editName.value
  
  loadingContent.value = true
  saveStatus.value = 'saved'
  try {
    const res = await getAuthorNodeContent(nodeId)
    if (res.data?.code === 200 && res.data.data) {
      contentInfo.value = res.data.data
      // 优先显示草稿，如果没有草稿则显示正式内容
      editContent.value = res.data.data.hasDraft === 1 ? res.data.data.contentDraft : res.data.data.content
      originalContent.value = editContent.value

      // Initialize or update Vditor
      nextTick(() => {
        if (vditorInstance) {
          vditorInstance.setValue(editContent.value)
        } else {
          vditorInstance = new Vditor('vditor-container', {
            mode: 'ir',
            cdn: '/vditor',
            minHeight: 0,
            height: '100%',
            toolbarConfig: { pin: true },
            toolbar: [
              'headings', 'bold', 'italic', 'strike', 'link', '|', 
              'list', 'ordered-list', 'check', 'outdent', 'indent', '|', 
              'quote', 'line', 'code', 'inline-code', 'insert-before', 'insert-after', '|', 
              'undo', 'redo'
            ],
            cache: { enable: false },
            input: (value) => {
              editContent.value = value
              // 触发 watcher 的保存逻辑
            }
          })
        }
      })
    } else {
      editContent.value = ''
      originalContent.value = ''
      contentInfo.value = null
      nextTick(() => {
        if (vditorInstance) {
          vditorInstance.setValue('')
        }
      })
    }
  } catch (err) {
    console.error('获取节点内容失败', err)
    editContent.value = ''
    originalContent.value = ''
    contentInfo.value = null
  } finally {
    loadingContent.value = false
  }
}

// 自动保存逻辑
watch([editName, editContent], ([newName, newContent]) => {
  if (newName !== originalName.value || newContent !== originalContent.value) {
    saveStatus.value = 'unsaved'
    if (saveTimer) clearTimeout(saveTimer)
    saveTimer = setTimeout(() => {
      saveDraft()
    }, 2000) // 2秒防抖自动保存
  }
})

const saveDraft = async () => {
  if (!activeNodeId.value) return
  saveStatus.value = 'saving'
  saving.value = true
  
  try {
    let success = true
    // 保存名称
    if (editName.value !== originalName.value) {
      const resName = await updateAuthorNodeName(activeNodeId.value, editName.value)
      if (resName.data?.code === 200) {
        originalName.value = editName.value
        // 更新列表中的节点
        const node = nodes.value.find(n => n.id === activeNodeId.value)
        if (node) {
          node.nameDraft = editName.value
          node.hasDraft = 1
        }
      } else {
        success = false
      }
    }
    
    // 保存内容
    if (editContent.value !== originalContent.value) {
      const resContent = await updateAuthorNodeContent(activeNodeId.value, editContent.value)
      if (resContent.data?.code === 200) {
        originalContent.value = editContent.value
      } else {
        success = false
      }
    }
    
    if (success) {
      saveStatus.value = 'saved'
    } else {
      saveStatus.value = 'error'
    }
  } catch (err) {
    console.error('保存失败', err)
    saveStatus.value = 'error'
  } finally {
    saving.value = false
  }
}

interface AuthorTreeNode extends AuthorNode {
  children: AuthorTreeNode[]
}

interface VisibleTreeNode extends AuthorNode {
  depth: number
}

// 构建树形结构供渲染
const treeNodes = computed<AuthorTreeNode[]>(() => {
  const buildTree = (parentId: number): AuthorTreeNode[] => {
    return nodes.value
      .filter(n => Number(n.parentId) === parentId)
      .map(n => ({
        ...n,
        children: buildTree(n.id)
      }))
  }
  return buildTree(0)
})

// 将当前展开状态下可见的节点拍平成一维列表，支持无限层级渲染
const visibleTreeNodes = computed<VisibleTreeNode[]>(() => {
  const result: VisibleTreeNode[] = []

  const walk = (items: AuthorTreeNode[], depth: number) => {
    items.forEach(item => {
      result.push({
        ...item,
        depth
      })

      if (expandedKeys.value.has(item.id) && item.children.length > 0) {
        walk(item.children, depth + 1)
      }
    })
  }

  walk(treeNodes.value, 0)
  return result
})

</script>

<template>
  <div class="author-layout">
    <!-- 侧边栏 -->
    <aside class="sidebar">
      <div class="sidebar-header">
        <button class="back-btn" @click="router.back()">
          <ArrowLeft :size="16" />
          返回
        </button>
        <h2 class="sidebar-title">教材目录</h2>
      </div>
      
      <div class="sidebar-content">
        <div v-if="loadingTree" class="loading-state">加载中...</div>
        <div v-else-if="treeNodes.length === 0" class="empty-state">暂无目录节点</div>
        <div v-else class="tree-container">
          <div
            v-for="node in visibleTreeNodes"
            :key="node.id"
            class="tree-node"
            :class="{ 'active': activeNodeId === node.id }"
            @click="handleNodeSelect(node.id)"
          >
            <div class="node-content" :style="{ paddingLeft: `${8 + node.depth * 16}px` }">
              <span v-if="node.isLeaf !== 1" class="expand-icon" @click="toggleExpand(node, $event)">
                <ChevronDown v-if="expandedKeys.has(node.id)" :size="14" />
                <ChevronRight v-else :size="14" />
              </span>
              <FileText v-if="node.isLeaf === 1" :size="14" class="node-icon file" />
              <Folder v-else :size="14" class="node-icon folder" />
              <span class="node-title">{{ node.nameDraft || node.name || '未命名' }}</span>
              <span v-if="node.status === 'draft' || node.hasDraft === 1" class="draft-dot"></span>
            </div>
          </div>
        </div>
      </div>
    </aside>

    <!-- 主编辑区 -->
    <main class="editor-main">
      <div v-if="!activeNodeId" class="empty-editor">
        <FileText :size="48" class="empty-icon" />
        <p>请在左侧选择一个节点进行编辑</p>
      </div>
      
      <div v-else-if="loadingContent" class="loading-editor">
        <div class="skeleton-title"></div>
        <div class="skeleton-content"></div>
        <div class="skeleton-content"></div>
        <div class="skeleton-content" style="width: 60%"></div>
      </div>
      
      <div v-else class="editor-container">
        <header class="editor-header">
          <div class="header-status">
            <span v-if="contentInfo?.auditStatus === 1" class="badge warning">审核中</span>
            <span v-else-if="contentInfo?.auditStatus === 2" class="badge success">已发布</span>
            <span v-else-if="contentInfo?.auditStatus === 3" class="badge error">被驳回</span>
            <span v-else class="badge default">草稿</span>
            
            <div class="save-status">
              <span v-if="saveStatus === 'saving'" class="status-text warning"><loader-circle class="spin" :size="14" /> 保存中...</span>
              <span v-else-if="saveStatus === 'saved'" class="status-text success"><CheckCircle2 :size="14" /> 已保存</span>
              <span v-else-if="saveStatus === 'unsaved'" class="status-text warning"><Clock :size="14" /> 未保存</span>
              <span v-else-if="saveStatus === 'error'" class="status-text error">保存失败</span>
            </div>
          </div>
          <div class="header-actions">
          </div>
        </header>
        
        <div class="editor-body">
          <!-- 左侧：草稿编辑 -->
          <div class="pane draft-pane">
            <div class="pane-header" style="justify-content: flex-start; gap: 16px;">
              <span class="pane-title">草稿编辑</span>
              <button class="ghost-button" @click="showPublishedPane = !showPublishedPane" title="显示/隐藏已发布内容">
                {{ showPublishedPane ? '隐藏已发布' : '对比已发布' }}
              </button>
            </div>
            
            <div class="pane-content pane-content-vditor">
              <div class="draft-edit-area">
                <input 
                  v-model="editName" 
                  class="title-input" 
                  placeholder="输入草稿标题..."
                  autocomplete="off"
                />
                <div id="vditor-container" class="vditor-wrapper"></div>
              </div>
            </div>
          </div>

          <!-- 右侧：已上线字段 -->
          <div v-if="showPublishedPane" class="pane published-pane">
            <div class="pane-header">
              <span class="pane-title">已发布内容</span>
            </div>
            <div class="pane-content">
              <div class="published-name">{{ activeNode?.name || '无标题' }}</div>
              <div class="markdown-body" v-html="renderMarkdown(contentInfo?.content || '')"></div>
              <div v-if="!contentInfo?.content" class="empty-content">暂无已发布正文</div>
            </div>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<style scoped>
.author-layout {
  display: flex;
  height: 100vh;
  width: 100vw;
  background-color: #ffffff;
  color: rgba(0, 0, 0, 0.95);
  font-family: Inter, -apple-system, system-ui, sans-serif;
  overflow: hidden;
}

/* 侧边栏 */
.sidebar {
  width: 280px;
  background-color: #f6f5f4;
  border-right: 1px solid rgba(0, 0, 0, 0.1);
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}

.sidebar-header {
  padding: 16px;
  border-bottom: 1px solid rgba(0, 0, 0, 0.05);
}

.back-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  background: transparent;
  border: none;
  color: #615d59;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
  margin-left: -8px;
  margin-bottom: 12px;
}

.back-btn:hover {
  background-color: rgba(0, 0, 0, 0.05);
  color: rgba(0, 0, 0, 0.95);
}

.sidebar-title {
  font-size: 16px;
  font-weight: 600;
  margin: 0;
}

.sidebar-content {
  flex: 1;
  overflow-y: auto;
  padding: 12px 0;
}

.tree-container {
  display: flex;
  flex-direction: column;
}

.tree-node {
  padding: 4px 12px;
  cursor: pointer;
  user-select: none;
}

.tree-node:hover {
  background-color: rgba(0, 0, 0, 0.05);
}

.tree-node.active {
  background-color: rgba(0, 0, 0, 0.08);
  font-weight: 500;
}

.node-content {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 0;
  border-radius: 4px;
}

.expand-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  border-radius: 4px;
  color: #a39e98;
}

.expand-icon:hover {
  background-color: rgba(0, 0, 0, 0.1);
  color: #615d59;
}

.node-icon {
  color: #615d59;
}

.node-icon.folder {
  color: #a39e98;
}

.node-title {
  font-size: 14px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex: 1;
}

.draft-dot {
  width: 6px;
  height: 6px;
  background-color: #dd5b00;
  border-radius: 50%;
  margin-right: 8px;
}

.tree-children {
  display: flex;
  flex-direction: column;
}

/* 主编辑区 */
.editor-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  background-color: #ffffff;
  overflow: hidden;
}

.empty-editor {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #a39e98;
}

.empty-icon {
  margin-bottom: 16px;
  opacity: 0.5;
}

.editor-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  max-width: 1200px;
  width: 100%;
  margin: 0 auto;
  padding: 0 32px;
}

.editor-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 32px 0 16px;
}

.header-status {
  display: flex;
  gap: 8px;
  align-items: center;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.badge {
  padding: 4px 8px;
  border-radius: 9999px;
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.125px;
}

.badge.draft { background-color: #f6f5f4; color: #615d59; }
.badge.warning { background-color: #fff3e0; color: #dd5b00; }
.badge.success { background-color: #e8f5e9; color: #1aae39; }
.badge.error { background-color: #ffebee; color: #d32f2f; }

.header-actions {
  display: flex;
  align-items: center;
  gap: 16px;
}

.save-status {
  font-size: 14px;
}

.status-text {
  display: flex;
  align-items: center;
  gap: 4px;
  color: #a39e98;
}

.status-text.success { color: #1aae39; }
.status-text.warning { color: #dd5b00; }
.status-text.error { color: #d32f2f; }

.btn-primary {
  display: flex;
  align-items: center;
  gap: 6px;
  background-color: #0075de;
  color: #ffffff;
  border: 1px solid transparent;
  border-radius: 4px;
  padding: 6px 12px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary:hover:not(:disabled) {
  background-color: #005bab;
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.editor-body {
  flex: 1;
  display: flex;
  gap: 24px;
  padding-bottom: 32px;
  min-height: 0;
}

.pane {
  flex: 1;
  display: flex;
  flex-direction: column;
  background-color: #f6f5f4;
  border-radius: 12px;
  border: 1px solid rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.pane-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid rgba(0, 0, 0, 0.05);
  background-color: #ffffff;
}

.pane-title {
  font-size: 14px;
  font-weight: 600;
  color: #615d59;
}

.draft-header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.ghost-button {
  background: transparent;
  border: 1px solid rgba(0, 0, 0, 0.1);
  padding: 4px 12px;
  font-size: 12px;
  font-weight: 500;
  color: #615d59;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
}

.ghost-button:hover {
  background: rgba(0, 0, 0, 0.05);
  color: rgba(0, 0, 0, 0.95);
}



.pane-content {
  flex: 1;
  padding: 24px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  background-color: #ffffff;
}

.pane-content-vditor {
  padding: 0; /* Let vditor manage its own padding */
  display: flex;
  flex-direction: column;
  overflow: hidden; /* Important for Vditor to scroll itself */
}

.vditor-wrapper {
  flex: 1;
  min-height: 0; /* Crucial for flex scrolling inside Vditor */
  border: none !important;
}

/* Override Vditor internal styles to match Notion */
.vditor-wrapper :deep(.vditor) {
  border: none !important;
}
.vditor-wrapper :deep(.vditor-toolbar) {
  border-bottom: 1px solid rgba(0, 0, 0, 0.05);
  background-color: #f6f5f4;
  padding: 8px 16px;
}
.vditor-wrapper :deep(.vditor-reset) {
  padding: 24px;
  font-family: Inter, -apple-system, system-ui, sans-serif;
  color: rgba(0,0,0,0.95);
}

/* 隐藏 vditor 默认的提示框 */
.vditor-wrapper :deep(.vditor-tip) {
  display: none !important;
}

.draft-edit-area, .draft-preview-area {
  display: flex;
  flex-direction: column;
  flex: 1;
}

.published-name {
  font-size: 32px;
  font-weight: 700;
  color: rgba(0, 0, 0, 0.95);
  margin-bottom: 24px;
  letter-spacing: -0.625px;
}

.empty-content {
  color: #a39e98;
  font-size: 14px;
  font-style: italic;
  margin-top: 16px;
}

.title-input {
  font-size: 32px;
  font-weight: 700;
  border: none;
  outline: none;
  color: rgba(0, 0, 0, 0.95);
  background: transparent;
  padding: 24px 24px 0 24px;
  margin-bottom: 0;
  letter-spacing: -0.625px;
}

.title-input::placeholder {
  color: #a39e98;
}



/* Markdown Styles inside panes */
.markdown-body {
  font-size: 15px;
  line-height: 1.6;
  color: rgba(0,0,0,0.95);
}

.markdown-body :deep(h1),
.markdown-body :deep(h2),
.markdown-body :deep(h3),
.markdown-body :deep(h4) {
  font-weight: 600;
  margin-top: 1.5em;
  margin-bottom: 0.5em;
  color: rgba(0,0,0,0.95);
}

.markdown-body :deep(p) {
  margin-bottom: 1em;
}

.markdown-body :deep(code) {
  background-color: #f6f5f4;
  padding: 0.2em 0.4em;
  border-radius: 4px;
  font-family: ui-monospace, SFMono-Regular, Consolas, monospace;
  font-size: 0.9em;
}

.markdown-body :deep(pre code) {
  background-color: transparent;
  padding: 0;
  color: inherit;
}

.markdown-body :deep(pre) {
  background-color: #f6f5f4;
  padding: 16px;
  border-radius: 8px;
  overflow-x: auto;
  margin-bottom: 1em;
  border: 1px solid rgba(0,0,0,0.05);
}

.markdown-body :deep(blockquote) {
  border-left: 4px solid #dddddd;
  padding-left: 16px;
  margin-left: 0;
  color: #615d59;
}

/* 骨架屏 */
.loading-state, .empty-state {
  padding: 24px;
  text-align: center;
  color: #a39e98;
  font-size: 14px;
}

.loading-editor {
  padding: 64px;
  max-width: 900px;
  width: 100%;
  margin: 0 auto;
}

.skeleton-title {
  height: 48px;
  width: 60%;
  background-color: #f6f5f4;
  border-radius: 8px;
  margin-bottom: 32px;
  animation: pulse 1.5s infinite;
}

.skeleton-content {
  height: 16px;
  width: 100%;
  background-color: #f6f5f4;
  border-radius: 4px;
  margin-bottom: 12px;
  animation: pulse 1.5s infinite;
}

@keyframes pulse {
  0% { opacity: 0.6; }
  50% { opacity: 1; }
  100% { opacity: 0.6; }
}
</style>
