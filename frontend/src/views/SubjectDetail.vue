<script setup lang="ts">
import { ref, onMounted, computed, reactive, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import markdownit from 'markdown-it'
import hljs from 'highlight.js'
import 'highlight.js/styles/github.css' // 切换回浅色主题，适合灰色背景
import { 
  getTopNodes, getChildNodes, getNodeDetail, getNodeNote,
  updateNodeStatus, updateNodeDifficulty, saveNodeNote, getNodePath
} from '../api/node'
import { getSubjectDetail } from '../api/subject'
import { 
  likeSubject, 
  getUserCollectFolders, 
  createCollectFolder, 
  addSubjectToFolder,
  uncollectSubject,
  getPrivateNoteDetail,
  createPrivateNote,
  deletePrivateNote,
  updatePrivateNoteContent,
  updatePrivateNoteTitle,
  updatePrivateNotePublic
} from '../api/user'
import type { SubjectNode, SubjectNodeDetail, NodeNote } from '../types/node'
import type { 
  CollectFolderRes, 
  PrivateNoteBase,
  PrivateMarkdownDetail 
} from '../types/user'
import TreeItem from '../components/TreeItem.vue'
import Toast from '../components/Toast.vue'
import {
  getNodeProgressStatusLabel,
  normalizeNodeProgressStatus,
  type NodeProgressStatus
} from '../utils/nodeProgress'
import { NOTE_MAX_LENGTH, validateNoteContent } from '../utils/noteValidation'
import { 
  FileText, ArrowLeft, 
  Edit3, BookOpen, 
  Clock, Award, BookOpenCheck, Loader2,
  LayoutList, LayoutPanelLeft, CheckCircle2,
  Heart, Bookmark, Plus, X, Globe, Lock,
  StickyNote, ChevronRight, Copy,
  Folder, FolderPlus, FilePlus, ChevronLeft, Trash2,
  ToggleRight, ToggleLeft
} from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()
const subjectId = Number(route.params.id)
const isLoggedIn = computed(() => !!localStorage.getItem('token'))
const isLiked = ref(false)
const isCollected = ref(false)
const collectFolderId = ref<number | null>(null) // 教材所属的收藏夹ID
const lastNodeId = ref<number | null>(null) // 最近学习的知识点ID
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
      collectFolderId.value = res.data.data.collectFolderId || null
      lastNodeId.value = res.data.data.lastNodeId || null
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
  
  // 如果已经收藏，点击则是取消收藏（从所有收藏夹移除）
  if (isCollected.value) {
    try {
      const res = await uncollectSubject(subjectId)
      if (res.data?.code === 200) {
        showToast('已取消收藏')
        isCollected.value = false
        collectFolderId.value = null
      }
    } catch (error: any) {
      showToast(error.response?.data?.msg || '取消收藏失败', 'error')
    }
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
      collectFolderId.value = folderId
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
    // 优先从 URL 参数获取，如果没有，则使用 fetchSubjectDetail 拿到的 lastNodeId
    const targetNodeId = Number(route.query.nodeId) || lastNodeId.value
    
    // 如果有有效的 nodeId，使用 getNodePath 接口定位
    if (targetNodeId && !isNaN(targetNodeId)) {
      const resPath = await getNodePath(targetNodeId)
      if (resPath.data?.code === 200 && resPath.data.data) {
        const pathNodes = resPath.data.data
        if (pathNodes.length > 0) {
          // 1. 获取顶级节点（路径的第一项）并初始化 topNodes
          const resTop = await getTopNodes(subjectId)
          if (resTop.data?.code === 200 && resTop.data.data) {
            topNodes.value = resTop.data.data.map(node => ({
              ...node,
              expanded: false,
              children: [],
              userProgressStatus: normalizeNodeProgressStatus(node.userProgressStatus)
            }))

            // 2. 按照路径深度逐层展开并加载子节点
            let currentNodes = topNodes.value
            for (let i = 0; i < pathNodes.length; i++) {
              const pNode = pathNodes[i]
              const foundNode = currentNodes.find(n => n.id === pNode.id)
              
              if (foundNode) {
                // 如果不是最后一个节点（即路径中的中间节点），则展开并加载子节点
                if (i < pathNodes.length - 1) {
                  foundNode.expanded = true
                  await ensureChildrenLoaded(foundNode)
                  currentNodes = foundNode.children || []
                } else {
                  // 最后一个节点，直接选中
                  await selectNode(foundNode.id)
                  // 如果是文件夹，也展开一下
                  if (foundNode.isLeaf === 0) {
                    foundNode.expanded = true
                    await ensureChildrenLoaded(foundNode)
                  }
                }
              }
            }
            return // 成功处理路径逻辑后退出
          }
        }
      }
    }

    // 默认逻辑：获取顶级节点并选中第一个
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
    console.error('获取知识点树失败', error)
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
const noteEditContent = ref('') // 编辑中的笔记内容
const isAppendingToPrivate = ref(false)
const isOverwritingPrivate = ref(false)
const lastSavedContent = ref('') // 上次保存的内容，用于减少无谓请求
const isNoteImportant = ref(1) // 默认标记为重要
const loadingDetail = ref(false)
const updatingStatus = ref(false)
const updatingDifficulty = ref(false)
const savingNote = ref(false)
const saveStatusText = ref<'saving' | 'saved' | 'error-empty' | 'error-too-long' | 'error-xss' | 'error-net' | ''>('')
let debounceTimer: ReturnType<typeof setTimeout> | null = null

const appendClassNoteToCurrentPrivate = async () => {
  if (!isLoggedIn.value) {
    showToast('请先登录再使用私人笔记', 'error')
    router.push('/login')
    return
  }
  if (!currentPrivateNote.value) {
    showToast('请先在私人笔记中打开一个笔记文件', 'error')
    return
  }
  if (!noteEditContent.value.trim()) {
    showToast('随堂笔记为空，无法追加', 'error')
    return
  }
  if (isAppendingToPrivate.value) return
  isAppendingToPrivate.value = true
  try {
    const separator = currentPrivateNote.value.content ? '\n\n---\n\n' : ''
    const newContent = `${currentPrivateNote.value.content || ''}${separator}${noteEditContent.value}`
    const res = await updatePrivateNoteContent(currentPrivateNote.value.id, newContent)
    if (res.data?.code === 200) {
      currentPrivateNote.value.content = newContent
      lastPrivateSyncAt.value = formatDateTime(new Date())
      setPrivateSyncStatus('saved')
      showToast('已将随堂笔记追加到当前私人笔记', 'success')
    } else {
      showToast(res.data?.msg || '追加失败', 'error')
    }
  } catch (error: any) {
    showToast(error?.response?.data?.msg || '追加失败', 'error')
  } finally {
    isAppendingToPrivate.value = false
  }
}

const overwriteClassNoteToCurrentPrivate = async () => {
  if (!isLoggedIn.value) {
    showToast('请先登录再使用私人笔记', 'error')
    router.push('/login')
    return
  }
  if (!currentPrivateNote.value) {
    showToast('请先在私人笔记中打开一个笔记文件', 'error')
    return
  }
  if (!noteEditContent.value.trim()) {
    showToast('随堂笔记为空，无法覆盖', 'error')
    return
  }
  if (isOverwritingPrivate.value) return
  isOverwritingPrivate.value = true
  try {
    const newContent = noteEditContent.value
    const res = await updatePrivateNoteContent(currentPrivateNote.value.id, newContent)
    if (res.data?.code === 200) {
      currentPrivateNote.value.content = newContent
      lastPrivateSyncAt.value = formatDateTime(new Date())
      setPrivateSyncStatus('saved')
      showToast('已用随堂笔记覆盖当前私人笔记', 'success')
    } else {
      showToast(res.data?.msg || '覆盖失败', 'error')
    }
  } catch (error: any) {
    showToast(error?.response?.data?.msg || '覆盖失败', 'error')
  } finally {
    isOverwritingPrivate.value = false
  }
}

// ----------------- 私人笔记相关 -----------------
const noteType = ref<'class' | 'private'>('class')
const privateNotes = ref<PrivateNoteBase[]>([])
const currentPrivateNote = ref<PrivateMarkdownDetail | null>(null)
const privateNavStack = ref<{ id: number; title: string }[]>([{ id: 0, title: '根目录' }])
const loadingPrivateNotes = ref(false)
const currentFolderId = computed(() => privateNavStack.value[privateNavStack.value.length - 1].id)
const lastPrivateSyncAt = ref('')
const privateSyncStatus = ref<'saving' | 'saved' | ''>('')
const privateSyncStatusTimer = ref<number | null>(null)

const formatDateTime = (date: Date) => {
  const yyyy = date.getFullYear()
  const mm = date.getMonth() + 1
  const dd = date.getDate()
  const HH = String(date.getHours()).padStart(2, '0')
  const MM = String(date.getMinutes()).padStart(2, '0')
  const SS = String(date.getSeconds()).padStart(2, '0')
  return `${yyyy}-${mm}-${dd} ${HH}:${MM}:${SS}`
}

const formatDateTimeText = (value: string) => {
  const d = new Date(value)
  if (!Number.isNaN(d.getTime())) return formatDateTime(d)
  return value
}

const setPrivateSyncStatus = (status: 'saving' | 'saved' | '') => {
  if (privateSyncStatusTimer.value) {
    clearTimeout(privateSyncStatusTimer.value)
    privateSyncStatusTimer.value = null
  }
  privateSyncStatus.value = status
  if (status === 'saved') {
    privateSyncStatusTimer.value = window.setTimeout(() => {
      privateSyncStatus.value = ''
      privateSyncStatusTimer.value = null
    }, 1500)
  }
}

const privateNotesTotal = ref(0)
const privateNotesPage = ref(1)
const privateNotesPageSize = ref(20)

const fetchPrivateContent = async (noteId: number, page = 1) => {
  if (!isLoggedIn.value) return
  loadingPrivateNotes.value = true
  privateNotesPage.value = page
  try {
    const res = await getPrivateNoteDetail(noteId, 2, page, privateNotesPageSize.value)
    if (res.data?.code === 200) {
      if (!res.data.data) {
        privateNotes.value = []
        privateNotesTotal.value = 0
        currentPrivateNote.value = null
        lastPrivateSyncAt.value = ''
        setPrivateSyncStatus('')
        return
      }
      const data = res.data.data
      if (data.type === 'folder') {
        privateNotes.value = Array.isArray(data.children) ? data.children : []
        privateNotesTotal.value = data.total || 0
        currentPrivateNote.value = null
        lastPrivateSyncAt.value = ''
        setPrivateSyncStatus('')
      } else {
        privateNotes.value = []
        privateNotesTotal.value = 0
        currentPrivateNote.value = data.content ?? null
        lastPrivateSyncAt.value = currentPrivateNote.value?.updatedAt ? formatDateTimeText(currentPrivateNote.value.updatedAt) : ''
        setPrivateSyncStatus('')
      }
    }
  } catch (error: any) {
    const status = error?.response?.status
    const code = error?.response?.data?.code
    if (noteId === 0 && (status === 404 || code === 404)) {
      privateNotes.value = []
      currentPrivateNote.value = null
      lastPrivateSyncAt.value = ''
      setPrivateSyncStatus('')
      return
    }
    console.error('获取私人笔记失败', error)
    showToast(error?.response?.data?.msg || '获取私人笔记失败', 'error')
  } finally {
    loadingPrivateNotes.value = false
  }
}

const handlePrivateItemClick = (item: PrivateNoteBase) => {
  if (item.type === 'folder') {
    privateNavStack.value.push({ id: item.id, title: item.title })
    fetchPrivateContent(item.id, 1)
  } else {
    fetchPrivateContent(item.id, 1)
  }
}

const goBackPrivate = () => {
  if (privateNavStack.value.length > 1) {
    // 如果当前正在查看笔记正文，点击返回应停留在当前文件夹
    if (currentPrivateNote.value) {
      currentPrivateNote.value = null
      fetchPrivateContent(currentFolderId.value, 1)
    } else {
      privateNavStack.value.pop()
      fetchPrivateContent(currentFolderId.value, 1)
    }
  } else if (currentPrivateNote.value) {
    currentPrivateNote.value = null
    fetchPrivateContent(0, 1)
  }
}

const toggleNoteType = (type: 'class' | 'private') => {
  noteType.value = type
  if (type === 'private') {
    if (!privateNavStack.value.length) {
      privateNavStack.value = [{ id: 0, title: '根目录' }]
    }
    if (!currentPrivateNote.value && privateNotes.value.length === 0) {
      fetchPrivateContent(currentFolderId.value, 1)
    }
  }
}

// ----------------- 新建私人笔记/文件夹 -----------------
const showCreatePrivateModal = ref(false)
const createPrivateType = ref<'folder' | 'markdown'>('markdown')
const createPrivateTitle = ref('')
const createPrivateContent = ref('')
const createPrivateIsPublic = ref(false)
const creatingPrivate = ref(false)

const openCreatePrivate = (type: 'folder' | 'markdown') => {
  createPrivateType.value = type
  createPrivateTitle.value = ''
  createPrivateContent.value = ''
  createPrivateIsPublic.value = false
  showCreatePrivateModal.value = true
}

const handleCreatePrivate = async () => {
  if (!createPrivateTitle.value.trim()) {
    showToast('标题不能为空', 'error')
    return
  }
  
  creatingPrivate.value = true
  try {
    const res = await createPrivateNote({
      parentId: currentFolderId.value,
      type: createPrivateType.value,
      title: createPrivateTitle.value,
      content: createPrivateType.value === 'markdown' ? createPrivateContent.value : undefined,
      isPublic: createPrivateIsPublic.value ? 1 : 0
    })
    
    if (res.data?.code === 200) {
      showToast('创建成功')
      showCreatePrivateModal.value = false
      fetchPrivateContent(currentFolderId.value, 1)
    } else {
      showToast(res.data?.msg || '创建失败', 'error')
    }
  } catch (error: any) {
    showToast(error.response?.data?.msg || '创建失败', 'error')
  } finally {
    creatingPrivate.value = false
  }
}

const showDeletePrivateModal = ref(false)
const deletingPrivate = ref(false)
const pendingDeleteItem = ref<PrivateNoteBase | null>(null)

// ----------------- 修改私人笔记相关 -----------------
const showRenamePrivateModal = ref(false)
const renamePrivateTitle = ref('')
const renamingPrivate = ref(false)
const pendingRenameItem = ref<{ id: number; title: string } | null>(null)

const openRenamePrivate = (item: PrivateNoteBase) => {
  pendingRenameItem.value = item
  renamePrivateTitle.value = item.title
  showRenamePrivateModal.value = true
}

const openRenameCurrentPrivate = () => {
  if (!currentPrivateNote.value) return
  pendingRenameItem.value = { id: currentPrivateNote.value.id, title: currentPrivateNote.value.title }
  renamePrivateTitle.value = currentPrivateNote.value.title
  showRenamePrivateModal.value = true
}

const handleRenamePrivate = async () => {
  if (!renamePrivateTitle.value.trim() || !pendingRenameItem.value) return
  renamingPrivate.value = true
  setPrivateSyncStatus('saving')
  try {
    const res = await updatePrivateNoteTitle(pendingRenameItem.value.id, renamePrivateTitle.value)
    if (res.data?.code === 200) {
      showToast('重命名成功')
      showRenamePrivateModal.value = false
      lastPrivateSyncAt.value = formatDateTime(new Date())
      setPrivateSyncStatus('saved')
      // 如果当前正在查看这个笔记的正文，同步更新标题
      if (currentPrivateNote.value && currentPrivateNote.value.id === pendingRenameItem.value.id) {
        currentPrivateNote.value.title = renamePrivateTitle.value
        fetchPrivateContent(currentPrivateNote.value.id, 1)
      } else {
        fetchPrivateContent(currentFolderId.value, privateNotesPage.value)
      }
    }
  } catch (error: any) {
    setPrivateSyncStatus('')
    showToast(error.response?.data?.msg || '重命名失败', 'error')
  } finally {
    renamingPrivate.value = false
  }
}

const handleTogglePublic = async (item: PrivateNoteBase | PrivateMarkdownDetail) => {
  const newPublic = item.isPublic === 1 ? 0 : 1
  try {
    setPrivateSyncStatus('saving')
    const res = await updatePrivateNotePublic(item.id, newPublic as 0 | 1)
    if (res.data?.code === 200) {
      item.isPublic = newPublic as 0 | 1
      showToast(newPublic === 1 ? '已设为公开' : '已设为私密')
      lastPrivateSyncAt.value = formatDateTime(new Date())
      setPrivateSyncStatus('saved')
    } else {
      setPrivateSyncStatus('')
      showToast(res.data?.msg || '修改公开状态失败', 'error')
    }
  } catch (error: any) {
    setPrivateSyncStatus('')
    showToast(error.response?.data?.msg || '修改公开状态失败', 'error')
  }
}

const savingPrivateContent = ref(false)
const privateContentDebounceTimer = ref<number | null>(null)

const handlePrivateContentChange = (content: string) => {
  if (privateContentDebounceTimer.value) {
    clearTimeout(privateContentDebounceTimer.value)
  }
  
  privateContentDebounceTimer.value = window.setTimeout(async () => {
    if (!currentPrivateNote.value) return
    savingPrivateContent.value = true
    setPrivateSyncStatus('saving')
    try {
      const res = await updatePrivateNoteContent(currentPrivateNote.value.id, content)
      if (res.data?.code === 200) {
        lastPrivateSyncAt.value = formatDateTime(new Date())
        currentPrivateNote.value.updatedAt = lastPrivateSyncAt.value
        setPrivateSyncStatus('saved')
      } else {
        setPrivateSyncStatus('')
        showToast(res.data?.msg || '同步失败', 'error')
      }
    } catch (error: any) {
      setPrivateSyncStatus('')
      showToast(error?.response?.data?.msg || '同步失败', 'error')
    } finally {
      savingPrivateContent.value = false
    }
  }, 1000)
}

const openDeletePrivate = (item: PrivateNoteBase) => {
  pendingDeleteItem.value = item
  showDeletePrivateModal.value = true
}

const handleDeletePrivate = async () => {
  if (!pendingDeleteItem.value || deletingPrivate.value) return
  deletingPrivate.value = true
  try {
    const res = await deletePrivateNote(pendingDeleteItem.value.id)
    if (res.data?.code === 200) {
      showToast('已移入回收站')
      showDeletePrivateModal.value = false
      pendingDeleteItem.value = null
      fetchPrivateContent(currentFolderId.value, privateNotesPage.value)
      return
    }
    showToast(res.data?.msg || '删除失败', 'error')
  } catch (error: any) {
    showToast(error?.response?.data?.msg || '删除失败', 'error')
  } finally {
    deletingPrivate.value = false
  }
}

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

const copyToClipboard = async (text: string, successMessage: string) => {
  if (!text) {
    showToast('内容为空，无法复制', 'error')
    return
  }
  try {
    await navigator.clipboard.writeText(text)
    showToast(successMessage, 'success')
  } catch {
    showToast('复制失败，请检查浏览器权限', 'error')
  }
}

const copyNodeTitle = () => {
  if (!nodeDetail.value) return
  copyToClipboard(nodeDetail.value.name || '', '标题已复制')
}

const copyNodeContent = () => {
  if (!nodeDetail.value) return
  copyToClipboard(nodeDetail.value.content || '', '教材正文已复制')
}

const copyClassNote = () => {
  copyToClipboard(noteEditContent.value || '', '随堂笔记已复制')
}

const copyPrivateNote = () => {
  if (!currentPrivateNote.value) return
  copyToClipboard(currentPrivateNote.value.content || '', '私人笔记已复制')
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
  if (debounceTimer) {
    clearTimeout(debounceTimer)
    debounceTimer = null
  }
  currentNodeId.value = id
  
  // 1. 检查缓存：如果命中，直接使用并返回
  if (detailCache.has(id)) {
    console.log(`[Cache Hit] Node ID: ${id}`)
    nodeDetail.value = detailCache.get(id) || null
    nodeNote.value = noteCache.get(id) || null
    noteEditContent.value = nodeNote.value?.noteContent || ''
    lastSavedContent.value = noteEditContent.value // 同步上次保存内容
    saveStatusText.value = '' // 重置状态
    isNoteImportant.value = nodeNote.value?.isImportant ?? 1
    
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
        noteEditContent.value = note.noteContent || ''
        lastSavedContent.value = noteEditContent.value // 同步上次保存内容
        saveStatusText.value = '' // 重置状态
        isNoteImportant.value = note.isImportant ?? 1
        noteCache.set(id, note) // 存入缓存
      } else {
        nodeNote.value = null
        noteEditContent.value = ''
        lastSavedContent.value = ''
        saveStatusText.value = ''
        isNoteImportant.value = 1
        noteCache.set(id, null) // 即使没有笔记也缓存 null，避免重复请求
      }
    } else {
      nodeNote.value = null
      noteEditContent.value = ''
      lastSavedContent.value = ''
      saveStatusText.value = ''
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

onMounted(async () => {
  if (!subjectId) {
    router.replace('/')
    return
  }
  await fetchSubjectDetail()
  fetchTopNodes()
})

// ----------------- 布局拖拽相关 -----------------
const sidebarWidth = ref(320)
const noteSidebarWidth = ref(360)
const isResizingLeft = ref(false)
const isResizingRight = ref(false)

const startResizingLeft = () => {
  isResizingLeft.value = true
  document.addEventListener('mousemove', handleMouseMove)
  document.addEventListener('mouseup', stopResizing)
  document.body.style.cursor = 'col-resize'
  document.body.style.userSelect = 'none'
}

const startResizingRight = () => {
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

const handleSaveNote = async () => {
  if (!currentNodeId.value || savingNote.value) return
  // 如果内容没变，不发送请求
  if (noteEditContent.value === lastSavedContent.value) return
  
  savingNote.value = true
  saveStatusText.value = 'saving'
  try {
    const noteContent = validateNoteContent(noteEditContent.value)
    const res = await saveNodeNote(currentNodeId.value, {
      noteContent,
      isImportant: isNoteImportant.value
    })
    if (res.data?.code === 200) {
      noteEditContent.value = noteContent
      lastSavedContent.value = noteContent
      saveStatusText.value = 'saved'
      
      // 更新当前显示的时间
      const now = new Date().toLocaleString('zh-CN', { hour12: false }).replace(/\//g, '-')
      if (nodeNote.value) {
        nodeNote.value.noteContent = noteContent
        nodeNote.value.updatedAt = now
      } else {
        nodeNote.value = {
          id: 0,
          nodeId: currentNodeId.value,
          noteContent,
          isImportant: isNoteImportant.value,
          updatedAt: now
        }
      }
      noteCache.set(currentNodeId.value, nodeNote.value)
      
      // 3秒后清除“已同步”字样
      setTimeout(() => {
        if (saveStatusText.value === 'saved') saveStatusText.value = ''
      }, 3000)
    }
  } catch (error: any) {
    console.error('保存笔记失败', error)
    const msg = String(error?.message || '')
    if (msg.includes('不能为空')) {
      saveStatusText.value = 'error-empty'
    } else if (msg.includes('不能超过')) {
      saveStatusText.value = 'error-too-long'
    } else if (msg.includes('非法字符')) {
      saveStatusText.value = 'error-xss'
    } else {
      saveStatusText.value = 'error-net'
    }
  } finally {
    savingNote.value = false
  }
}

// 自动保存监听逻辑
watch([noteEditContent, isNoteImportant], ([newContent, newImportant]) => {
  if (!currentNodeId.value) return
  
  if (!newContent.trim()) {
    if (lastSavedContent.value) {
      saveStatusText.value = 'error-empty'
    } else {
      saveStatusText.value = ''
    }
    if (debounceTimer) {
      clearTimeout(debounceTimer)
      debounceTimer = null
    }
    return
  }

  if (newContent.length > NOTE_MAX_LENGTH) {
    saveStatusText.value = 'error-too-long'
    if (debounceTimer) {
      clearTimeout(debounceTimer)
      debounceTimer = null
    }
    return
  }

  if (newContent.includes('<') || newContent.includes('>')) {
    saveStatusText.value = 'error-xss'
    if (debounceTimer) {
      clearTimeout(debounceTimer)
      debounceTimer = null
    }
    return
  }

  // 仅当内容或重要程度发生实际变化时才触发防抖保存
  const isContentChanged = newContent !== lastSavedContent.value
  const isImportantChanged = nodeNote.value ? newImportant !== nodeNote.value.isImportant : false
  
  if (!isContentChanged && !isImportantChanged) return
  
  if (debounceTimer) clearTimeout(debounceTimer)
  
  debounceTimer = setTimeout(() => {
    handleSaveNote()
  }, 1000) // 1秒防抖
})

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
          <div class="content-title-line">
            <h1>{{ nodeDetail.name }}</h1>
            <button 
              class="icon-action-btn copy-icon-btn" 
              type="button"
              title="复制标题"
              @click="copyNodeTitle"
            >
              <Copy :size="16" />
            </button>
          </div>
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
            <div class="copy-actions-overlay content-body-copy-actions">
              <button 
                class="icon-action-btn copy-icon-btn" 
                type="button"
                title="复制教材正文"
                @click="copyNodeContent"
              >
                <Copy :size="16" />
              </button>
            </div>
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
        <div class="note-tabs">
          <button 
            class="tab-btn" 
            :class="{ active: noteType === 'class' }"
            @click="toggleNoteType('class')"
          >
            <Edit3 :size="16" />
            <span>随堂笔记</span>
          </button>
          <button 
            class="tab-btn" 
            :class="{ active: noteType === 'private' }"
            @click="toggleNoteType('private')"
          >
            <StickyNote :size="16" />
            <span>私人笔记</span>
          </button>
        </div>
        <div v-if="noteType === 'class'" class="class-note-header-actions">
          <button 
            class="append-to-private-btn" 
            :disabled="!currentPrivateNote || !noteEditContent.trim() || isAppendingToPrivate"
            @click="appendClassNoteToCurrentPrivate"
            title="将当前随堂笔记内容追加到右侧当前打开的私人笔记末尾"
          >
            <Loader2 v-if="isAppendingToPrivate" class="spin" :size="12" />
            <StickyNote v-else :size="12" />
            <span>追加到当前私人笔记</span>
          </button>
          <button 
            class="append-to-private-btn danger" 
            :disabled="!currentPrivateNote || !noteEditContent.trim() || isOverwritingPrivate"
            @click="overwriteClassNoteToCurrentPrivate"
            title="用当前随堂笔记内容覆盖右侧当前打开的私人笔记"
          >
            <Loader2 v-if="isOverwritingPrivate" class="spin" :size="12" />
            <FileText v-else :size="12" />
            <span>覆盖当前私人笔记</span>
          </button>
        </div>
      </div>
      
      <div class="note-main">
        <div v-if="!isLoggedIn" class="note-login-guide">
          <div class="guide-icon">🔒</div>
          <p>{{ noteType === 'class' ? '随堂笔记' : '私人笔记' }}已锁定</p>
          <p class="sub-p">登录后即可随时记录学习感悟</p>
          <button class="primary-btn" @click="router.push('/login')">去登录</button>
        </div>

        <!-- 随堂笔记内容 -->
        <template v-else-if="noteType === 'class'">
          <div v-if="!currentNodeId" class="note-empty">
            <p>选中一个知识点来记录笔记吧</p>
          </div>
          <div v-else class="note-active">
            <div class="editor-wrap">
              <div class="copy-actions-overlay note-editor-copy-actions">
                <button 
                  class="icon-action-btn copy-icon-btn"
                  type="button"
                  title="复制随堂笔记"
                  @click="copyClassNote"
                >
                  <Copy :size="14" />
                </button>
              </div>
              <textarea 
                class="note-textarea" 
                v-model="noteEditContent"
                placeholder="在这里输入你的随堂笔记..."
                :maxlength="NOTE_MAX_LENGTH"
              ></textarea>
            </div>
            <div class="editor-footer">
              <div class="save-status">
                <Clock :size="12" />
                <span v-if="nodeNote?.updatedAt">上次同步: {{ nodeNote.updatedAt }}</span>
                <span v-else>尚未记录笔记</span>
              </div>
              <div class="autosave-indicator">
                <span v-if="saveStatusText === 'saving'" class="status-saving">
                  <Loader2 class="spin" :size="12" /> 正在同步...
                </span>
                <span v-else-if="saveStatusText === 'saved'" class="status-saved">
                  已同步到云端
                </span>
                <span v-else-if="saveStatusText === 'error-empty'" class="status-error">
                  内容不能为空
                </span>
                <span v-else-if="saveStatusText === 'error-too-long'" class="status-error">
                  最多 {{ NOTE_MAX_LENGTH }} 字符
                </span>
                <span v-else-if="saveStatusText === 'error-xss'" class="status-error">
                  包含非法字符
                </span>
                <span v-else-if="saveStatusText === 'error-net'" class="status-error">
                  同步失败，请检查网络
                </span>
              </div>
            </div>
          </div>
        </template>

        <!-- 私人笔记内容 -->
        <template v-else>
          <div class="private-note-container">
            <div class="private-nav-bar">
              <div class="nav-left-group">
                <button 
                  v-if="privateNavStack.length > 1 || currentPrivateNote" 
                  class="nav-back-btn"
                  @click="goBackPrivate"
                >
                  <ChevronLeft :size="16" />
                  <span>返回</span>
                </button>
                <div v-else class="nav-placeholder"></div>
              </div>
              
              <div class="nav-center-group">
                <div 
                  class="nav-current-title" 
                  @click="currentPrivateNote ? openRenameCurrentPrivate() : null"
                  :style="currentPrivateNote ? { cursor: 'pointer' } : {}"
                >
                  {{ currentPrivateNote ? (currentPrivateNote?.title || '笔记') : (privateNavStack[privateNavStack.length - 1]?.title || '根目录') }}
                </div>

                <div v-if="currentPrivateNote" class="nav-file-actions">
                  <button class="icon-action-btn" title="重命名" @click="openRenameCurrentPrivate">
                    <Edit3 :size="16" />
                  </button>
                  <button 
                    class="icon-action-btn nav-toggle-btn" 
                    :class="{ isPublic: currentPrivateNote.isPublic === 1 }"
                    @click="handleTogglePublic(currentPrivateNote)"
                    :title="currentPrivateNote.isPublic === 1 ? '点击设为私密' : '点击设为公开'"
                  >
                    <ToggleRight v-if="currentPrivateNote.isPublic === 1" :size="18" />
                    <ToggleLeft v-else :size="18" />
                  </button>
                </div>
              </div>

              <div class="nav-actions">
                <button 
                  v-if="!currentPrivateNote"
                  class="icon-action-btn" 
                  title="新建文件夹"
                  @click="openCreatePrivate('folder')"
                >
                  <FolderPlus :size="16" />
                </button>
                <button 
                  v-if="!currentPrivateNote"
                  class="icon-action-btn" 
                  title="新建笔记"
                  @click="openCreatePrivate('markdown')"
                >
                  <FilePlus :size="16" />
                </button>
              </div>
            </div>

            <div class="private-body" :class="{ 'is-loading': loadingPrivateNotes }">
              <div v-if="loadingPrivateNotes" class="private-loading-overlay">
                <Loader2 class="spin" :size="24" />
                <span>正在获取内容...</span>
              </div>

              <div v-if="currentPrivateNote" class="private-content-view">
                <div class="view-body">
                  <div class="copy-actions-overlay private-editor-copy-actions">
                    <button 
                      class="icon-action-btn copy-icon-btn"
                      type="button"
                      title="复制当前私人笔记"
                      @click="copyPrivateNote"
                    >
                      <Copy :size="16" />
                    </button>
                  </div>
                  <textarea 
                    v-model="currentPrivateNote.content" 
                    class="private-editor" 
                    placeholder="输入笔记内容..."
                    @input="handlePrivateContentChange(($event.target as HTMLTextAreaElement).value)"
                  ></textarea>
                </div>
              </div>

              <div v-else class="private-list-view">
                <div v-if="privateNotes.length === 0" class="note-empty">
                  <StickyNote :size="32" class="empty-icon" />
                  <p>当前文件夹暂无笔记</p>
                  <button class="create-btn" @click="openCreatePrivate('markdown')">立即创建</button>
                </div>
                <div v-else class="private-note-list">
                  <div 
                    v-for="item in privateNotes" 
                    :key="item.id" 
                    class="private-note-item"
                    @click="handlePrivateItemClick(item)"
                  >
                    <div class="item-icon">
                      <Folder v-if="item.type === 'folder'" :size="18" class="folder-icon-color" />
                      <FileText v-else :size="18" class="file-icon-color" />
                    </div>
                    <div class="note-info">
                      <div class="title-row">
                        <span class="note-title" @click.stop="openRenamePrivate(item)">{{ item.title }}</span>
                        <Edit3 :size="12" class="edit-icon" @click.stop="openRenamePrivate(item)" />
                      </div>
                      <span class="note-date">{{ new Date(item.updatedAt).toLocaleDateString() }}</span>
                    </div>
                    <div class="private-item-actions">
                      <button 
                        class="toggle-public-mini" 
                        :class="{ isPublic: item.isPublic === 1 }"
                        @click.stop="handleTogglePublic(item)"
                        :title="item.isPublic === 1 ? '已公开' : '私密'"
                      >
                        <ToggleRight v-if="item.isPublic === 1" :size="16" />
                        <ToggleLeft v-else :size="16" />
                      </button>
                      <button class="item-delete-btn" title="删除" @click.stop="openDeletePrivate(item)">
                        <Trash2 :size="16" />
                      </button>
                      <ChevronRight :size="14" class="arrow-icon" />
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <div v-if="currentPrivateNote" class="private-footer-tip">
              <div class="save-status">
                <Clock :size="12" />
                <span v-if="lastPrivateSyncAt">上次同步: {{ lastPrivateSyncAt }}</span>
                <span v-else>尚未同步</span>
              </div>
              <div class="autosave-indicator">
                <span v-if="privateSyncStatus === 'saving'" class="status-saving">
                  <Loader2 class="spin" :size="12" /> 正在同步...
                </span>
                <span v-else-if="privateSyncStatus === 'saved'" class="status-saved">
                  已同步到云端
                </span>
              </div>
            </div>
          </div>
        </template>
      </div>
    </aside>

    <!-- 收藏弹窗 -->
    <Teleport to="body">
      <!-- 新建私人笔记/文件夹弹窗 -->
      <div v-if="showCreatePrivateModal" class="modal-overlay" @click.self="showCreatePrivateModal = false">
        <div class="modal-content small-modal">
          <header class="modal-header">
            <h3>新建{{ createPrivateType === 'folder' ? '文件夹' : '笔记' }}</h3>
            <button class="close-btn" @click="showCreatePrivateModal = false">
              <X :size="20" />
            </button>
          </header>
          <div class="modal-body">
            <!-- 提示创建位置 -->
            <div class="location-tip">
              <Folder :size="14" />
              <span>创建在：{{ privateNavStack[privateNavStack.length - 1]?.title || '根目录' }}</span>
            </div>
            
            <div class="form-group" style="margin-top: 16px;">
              <label>标题</label>
              <input 
                v-model="createPrivateTitle" 
                type="text" 
                :placeholder="createPrivateType === 'folder' ? '请输入文件夹名称' : '请输入笔记标题'"
                maxlength="255"
              />
            </div>
            <div class="form-group row" style="margin-top: 16px;">
              <label>公开</label>
              <label class="public-checkbox">
                <input v-model="createPrivateIsPublic" type="checkbox" />
                <span>{{ createPrivateIsPublic ? '公开' : '不公开' }}</span>
              </label>
            </div>
            <div v-if="createPrivateType === 'markdown'" class="form-group" style="margin-top: 16px;">
              <label>内容</label>
              <textarea 
                v-model="createPrivateContent" 
                placeholder="请输入笔记内容..."
                maxlength="1000"
              ></textarea>
            </div>
          </div>
          <footer class="modal-footer">
            <div class="form-actions">
              <button class="cancel-btn" @click="showCreatePrivateModal = false">取消</button>
              <button 
                class="confirm-btn" 
                :disabled="creatingPrivate"
                @click="handleCreatePrivate"
              >
                <Loader2 v-if="creatingPrivate" class="spin" :size="16" />
                <span v-else>确认创建</span>
              </button>
            </div>
          </footer>
        </div>
      </div>

      <div v-if="showDeletePrivateModal" class="modal-overlay" @click.self="showDeletePrivateModal = false">
        <div class="modal-content small-modal">
          <header class="modal-header">
            <h3>删除确认</h3>
            <button class="close-btn" @click="showDeletePrivateModal = false">
              <X :size="20" />
            </button>
          </header>
          <div class="modal-body">
            <div class="delete-tip">
              确认删除「{{ pendingDeleteItem?.title }}」吗？删除后将移入回收站。
            </div>
          </div>
          <footer class="modal-footer">
            <div class="form-actions">
              <button class="cancel-btn" @click="showDeletePrivateModal = false">取消</button>
              <button class="danger-btn" :disabled="deletingPrivate" @click="handleDeletePrivate">
                <Loader2 v-if="deletingPrivate" class="spin" :size="16" />
                <span v-else>确认删除</span>
              </button>
            </div>
          </footer>
        </div>
      </div>

      <!-- 重命名弹窗 -->
      <div v-if="showRenamePrivateModal" class="modal-overlay" @click.self="showRenamePrivateModal = false">
        <div class="modal-content small-modal">
          <header class="modal-header">
            <h3>重命名</h3>
            <button class="close-btn" @click="showRenamePrivateModal = false">
              <X :size="20" />
            </button>
          </header>
          <div class="modal-body">
            <div class="form-group">
              <label>新名称</label>
              <input 
                v-model="renamePrivateTitle" 
                type="text" 
                placeholder="请输入新名称"
                maxlength="255"
              />
            </div>
          </div>
          <footer class="modal-footer">
            <div class="form-actions">
              <button class="cancel-btn" @click="showRenamePrivateModal = false">取消</button>
              <button 
                class="confirm-btn" 
                :disabled="renamingPrivate"
                @click="handleRenamePrivate"
              >
                <Loader2 v-if="renamingPrivate" class="spin" :size="16" />
                <span v-else>确认</span>
              </button>
            </div>
          </footer>
        </div>
      </div>

      <!-- 收藏弹窗 -->
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

.public-checkbox {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  user-select: none;
  font-weight: normal;
  color: #475569;
}

.form-group input[type="checkbox"] {
  width: 16px;
  height: 16px;
  padding: 0;
  border-radius: 4px;
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
  position: relative;
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
  left: 24px;
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

.copy-actions-overlay {
  position: absolute;
  z-index: 3;
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 0;
  border-radius: 0;
  background: transparent;
  backdrop-filter: none;
  border: none;
  box-shadow: none;
}

.copy-icon-btn {
  padding: 2px !important;
  border-radius: 6px !important;
  background: transparent !important;
  color: rgba(148, 163, 184, 0.55) !important;
}

.copy-icon-btn:hover {
  background: transparent !important;
  color: #3b82f6 !important;
}

.copy-icon-btn:disabled {
  opacity: 0.35;
}

.content-body-copy-actions {
  top: -18px;
  right: -15px;
}

.note-editor-copy-actions {
  top: 10px;
  right: 10px;
}

.private-editor-copy-actions {
  top: 10px;
  right: 10px;
}

.content-title-line {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin-bottom: 20px;
}

.content-title-line h1 {
  margin-bottom: 0;
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
  padding: 16px 20px;
  border-bottom: 1px solid #f1f5f9;
}

.note-tabs {
  display: flex;
  background: #f1f5f9;
  padding: 4px;
  border-radius: 10px;
  width: 100%;
}

.tab-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 8px 0;
  border: none;
  background: transparent;
  color: #64748b;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  border-radius: 8px;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.tab-btn:hover {
  color: #1e293b;
}

.tab-btn.active {
  background: white;
  color: #3b82f6;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.note-main { flex: 1; display: flex; flex-direction: column; background: #fcfdfe; }

.class-note-header-actions {
  margin-top: 8px;
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: 8px;
}

.append-to-private-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 6px 10px;
  border-radius: 999px;
  border: 1px solid #e2e8f0;
  background: #f8fafc;
  color: #475569;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
}

.append-to-private-btn:hover:not(:disabled) {
  background: #eff6ff;
  border-color: #bfdbfe;
  color: #1d4ed8;
}

.append-to-private-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.append-to-private-btn.danger {
  background: #fef2f2;
  border-color: #fecaca;
  color: #b91c1c;
}

.append-to-private-btn.danger:hover:not(:disabled) {
  background: #fee2e2;
  border-color: #fca5a5;
  color: #991b1b;
}

/* 私人笔记容器 */
.private-note-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #fff;
}

.private-body {
  position: relative;
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.private-body.is-loading .private-content-view,
.private-body.is-loading .private-list-view {
  opacity: 0.6;
}

.private-loading-overlay {
  position: absolute;
  inset: 0;
  z-index: 2;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  background: rgba(255, 255, 255, 0.65);
  backdrop-filter: blur(2px);
  color: #64748b;
  font-size: 13px;
}

/* 导航栏 */
.private-nav-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid #f1f5f9;
  background: #fcfdfe;
}

.nav-left-group {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 120px;
}

.nav-center-group {
  display: flex;
  align-items: center;
  gap: 4px;
}

.nav-file-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

.nav-toggle-btn.isPublic {
  color: #3b82f6;
}

.nav-back-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  background: transparent;
  border: none;
  color: #64748b;
  font-size: 13px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
  transition: all 0.2s;
}

.nav-back-btn:hover {
  background: #f1f5f9;
  color: #3b82f6;
}

.nav-placeholder { width: 60px; }

.nav-current-title {
  font-size: 14px;
  font-weight: 700;
  color: #1e293b;
  max-width: 150px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.nav-actions {
  display: flex;
  gap: 8px;
  min-width: 120px;
  justify-content: flex-end;
}

.icon-action-btn {
  background: transparent;
  border: none;
  color: #64748b;
  cursor: pointer;
  padding: 6px;
  border-radius: 6px;
  transition: all 0.2s;
  display: flex;
  align-items: center;
}

.icon-action-btn:hover {
  background: #eff6ff;
  color: #3b82f6;
}

/* 列表视图 */
.private-list-view {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.private-note-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: #fff;
  border: 1px solid #f1f5f9;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  margin-bottom: 8px;
}

.private-note-item:hover {
  border-color: #dbeafe;
  background: #fcfdfe;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.03);
}

.item-icon {
  display: flex;
  align-items: center;
}

.note-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.title-row {
  display: flex;
  align-items: center;
  gap: 6px;
}

.note-title {
  font-size: 14px;
  font-weight: 600;
  color: #1e293b;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.note-title:hover {
  color: #3b82f6;
  text-decoration: underline;
}

.edit-icon {
  color: #94a3b8;
  opacity: 0;
  transition: opacity 0.2s;
  cursor: pointer;
}

.private-note-item:hover .edit-icon {
  opacity: 1;
}

.edit-icon:hover {
  color: #3b82f6;
}

.note-date {
  font-size: 11px;
  color: #94a3b8;
}

.private-item-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  opacity: 0.6;
  transition: opacity 0.2s;
}

.private-note-item:hover .private-item-actions {
  opacity: 1;
}

.toggle-public-mini {
  background: transparent;
  border: none;
  padding: 4px;
  border-radius: 4px;
  color: #94a3b8;
  cursor: pointer;
  display: flex;
  align-items: center;
  transition: all 0.2s;
}

.toggle-public-mini.isPublic {
  color: #3b82f6;
}

.toggle-public-mini:hover {
  background: #f1f5f9;
}

.item-delete-btn {
  background: transparent;
  border: none;
  padding: 6px;
  border-radius: 6px;
  color: #94a3b8;
  cursor: pointer;
  display: flex;
  align-items: center;
  transition: all 0.2s;
}

.item-delete-btn:hover {
  background: #fee2e2;
  color: #ef4444;
}

.arrow-icon {
  color: #cbd5e1;
}

.folder-icon-color { color: #f59e0b; }
.file-icon-color { color: #3b82f6; }

/* 正文视图 */
.private-content-view {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: #fff;
  position: relative;
}

.private-content-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 20px;
  border-bottom: 1px solid #f1f5f9;
}

.title-edit-area {
  flex: 1;
  margin-right: 16px;
}

.content-title-input {
  width: 100%;
  border: 1px solid transparent;
  background: transparent;
  font-size: 16px;
  font-weight: 700;
  color: #1e293b;
  padding: 4px 8px;
  border-radius: 4px;
  transition: all 0.2s;
}

.content-title-input:hover,
.content-title-input:focus {
  background: #f8fafc;
  border-color: #e2e8f0;
  outline: none;
}

.content-header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.toggle-public-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: #f1f5f9;
  border: 1px solid #e2e8f0;
  border-radius: 20px;
  color: #64748b;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.toggle-public-btn.isPublic {
  background: #eff6ff;
  border-color: #dbeafe;
  color: #3b82f6;
}

.toggle-public-btn:hover {
  opacity: 0.8;
}

.save-status-mini {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #94a3b8;
}

.view-body {
  flex: 1;
  padding: 0;
  overflow: hidden;
  position: relative;
}

.private-editor {
  width: 100%;
  height: 100%;
  padding: 20px;
  border: none;
  resize: none;
  font-size: 14px;
  line-height: 1.6;
  color: #334155;
  background: #fff;
  outline: none;
  font-family: inherit;
}

.private-footer-tip {
  padding: 12px 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  background: #fcfdfe;
  border-top: 1px solid #f1f5f9;
}

/* 弹窗样式增强 */
.small-modal {
  width: 400px;
  max-width: 90vw;
}

.location-tip {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: #f8fafc;
  border-radius: 6px;
  color: #64748b;
  font-size: 13px;
  border: 1px solid #f1f5f9;
}

.location-tip span {
  font-weight: 600;
  color: #3b82f6;
}

.delete-tip {
  padding: 12px 14px;
  background: #f8fafc;
  border: 1px solid #f1f5f9;
  border-radius: 10px;
  color: #475569;
  font-size: 14px;
  line-height: 1.5;
}

.cancel-btn {
  padding: 10px 20px;
  border: 1px solid #e2e8f0;
  background: #fff;
  border-radius: 8px;
  font-weight: 600;
  color: #64748b;
  cursor: pointer;
}

.danger-btn {
  padding: 10px 20px;
  border: none;
  background: #ef4444;
  color: #fff;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
}

.danger-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.confirm-btn {
  padding: 10px 20px;
  border: none;
  background: #3b82f6;
  color: #fff;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
}

.confirm-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.note-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  padding: 60px 0;
  color: #94a3b8;
}

.create-btn {
  margin-top: 16px;
  background: #fff;
  border: 1px solid #3b82f6;
  color: #3b82f6;
  padding: 8px 24px;
  border-radius: 20px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.create-btn:hover {
  background: #eff6ff;
  transform: scale(1.05);
}

.empty-icon {
  color: #f1f5f9;
}

.note-login-guide, .note-empty, .empty-view {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 20px;
}

.note-active { flex: 1; display: flex; flex-direction: column; gap: 16px; padding: 20px; position: relative; }

.primary-btn {
  background: #3b82f6;
  color: #fff;
  border: none;
  padding: 10px 24px;
  border-radius: 10px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.2);
}

.primary-btn:hover {
  background: #2563eb;
  transform: translateY(-1px);
  box-shadow: 0 6px 16px rgba(59, 130, 246, 0.3);
}

.editor-wrap {
  flex: 1;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.02);
  display: flex;
  position: relative;
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

.editor-footer {
  padding: 12px 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #fcfdfe;
  border-top: 1px solid #f1f5f9;
}

.save-status {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #94a3b8;
  font-size: 12px;
}

.autosave-indicator {
  font-size: 12px;
  font-weight: 500;
}

.status-saving {
  color: #3b82f6;
  display: flex;
  align-items: center;
  gap: 4px;
}

.status-saved {
  color: #10b981;
}

.status-error {
  color: #ef4444;
}



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
