<script setup lang="ts">
import { ref, onMounted, computed, watch, nextTick, onBeforeUnmount, reactive } from 'vue'
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
  createKnowledgeNode,
  updateKnowledgeNodeDraft,
  upsertKnowledgeContent
} from '../api/node'
import type { AuthorNode, AuthorNodeContent } from '../types/node'
import Toast from '../components/Toast.vue'
import {
  ChevronRight,
  ChevronDown,
  FileText,
  Folder,
  ArrowLeft,
  CheckCircle2,
  Clock,
  MessageSquare,
  Send,
  Bot,
  User,
  Pencil,
  LoaderCircle,
  Plus,
  ChevronUp,
  ArrowUpCircle,
  ArrowDownCircle,
  Copy,
  List,
  AlignVerticalSpaceAround,
  Wrench,
  Loader2
} from 'lucide-vue-next'
import {
  getAISessions,
  getAISessionMessages
} from '../api/ai'
import type { AIChatMessage, AIChatSession } from '../types/ai'

const route = useRoute()
const router = useRouter()
const subjectId = Number(route.params.id)
const TITLE_MAX_LENGTH = 150
const SAVE_DEBOUNCE_MS = 800
const SAVE_THROTTLE_MS = 1800

// 节点树状态
const isAccordionMode = ref(true)
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
const creatingChild = ref(false)
const createChildVisible = ref(false)
const newChildName = ref('')
const isHydratingEditor = ref(false)
const pendingSaveAfterCurrent = ref(false)
const toast = reactive({
  show: false,
  message: '',
  type: 'success' as 'success' | 'error'
})
let saveDebounceTimer: ReturnType<typeof setTimeout> | null = null
let saveThrottleTimer: ReturnType<typeof setTimeout> | null = null
let lastSaveTimestamp = 0

const showToast = (message: string, type: 'success' | 'error' = 'success') => {
  toast.message = message
  toast.type = type
  toast.show = true
}

const activeNodeNameReadonly = computed(() => activeNode.value?.parentId === 0)

onBeforeUnmount(() => {
  if (vditorInstance) {
    vditorInstance.destroy()
    vditorInstance = null
  }
  if (saveDebounceTimer) clearTimeout(saveDebounceTimer)
  if (saveThrottleTimer) clearTimeout(saveThrottleTimer)
})

// 控制是否显示已发布内容面板
const showPublishedPane = ref<boolean>(true)

// 右侧面板状态
const rightPanelTab = ref<'published' | 'ai'>('published')
const aiViewMode = ref<'entry' | 'chat'>('entry')
const aiHistoryDrawerOpen = ref(false)

// AI 助手相关状态
const aiMessages = ref<AIChatMessage[]>([])
const aiInput = ref('')
const aiSending = ref(false)
const aiMessagesContainer = ref<HTMLElement | null>(null)
const sessions = ref<AIChatSession[]>([])
const currentSessionId = ref<number | null>(null)
const loadingSessions = ref(false)
const hasMoreSessions = ref(true)
const loadingMessages = ref(false)
const hasMoreMessages = ref(true)
const reasoningCollapsedMap = reactive<Record<number, boolean>>({})
const latestSelectedText = ref('')
let activeChatAbortController: AbortController | null = null
const copyToast = ref('')
let copyToastTimer: ReturnType<typeof setTimeout> | null = null

const getCurrentSelectionText = () => window.getSelection()?.toString().trim() || ''

const syncSelectedText = () => {
  latestSelectedText.value = getCurrentSelectionText()
}

const appendAIContext = (formData: FormData) => {
  formData.set('currentPageUrl', window.location.href)

  const selectedText = latestSelectedText.value || getCurrentSelectionText()
  if (selectedText) {
    formData.set('selectedText', selectedText)
  } else {
    formData.delete('selectedText')
  }
}

const currentAISessionTitle = computed(() => {
  if (!currentSessionId.value) return '新对话'
  return sessions.value.find(item => item.id === currentSessionId.value)?.title || '历史对话'
})

const isStreamingAssistantMessage = (msg: AIChatMessage) => {
  if (!aiSending.value || msg.role !== 'assistant') return false
  const lastMessage = aiMessages.value[aiMessages.value.length - 1]
  return lastMessage?.id === msg.id
}

const toggleReasoning = (messageId: number) => {
  reasoningCollapsedMap[messageId] = !reasoningCollapsedMap[messageId]
}

const isReasoningCollapsed = (msg: AIChatMessage) => {
  if (reasoningCollapsedMap[msg.id] !== undefined) {
    return reasoningCollapsedMap[msg.id]
  }

  const isStreaming = aiSending.value && aiMessages.value[aiMessages.value.length - 1]?.id === msg.id
  if (isStreaming && !msg.content && msg.reasoning && msg.reasoning.length > 200) {
    return true
  }

  return false
}

// 加载会话列表
const loadSessions = async (reset = false) => {
  if (loadingSessions.value || (!hasMoreSessions.value && !reset)) return
  loadingSessions.value = true
  try {
    const lastId = reset ? undefined : sessions.value[sessions.value.length - 1]?.id
    const res = await getAISessions(lastId)
    if (res.data?.code === 200 && res.data.data) {
      if (reset) {
        sessions.value = res.data.data.list || []
      } else {
        sessions.value.push(...(res.data.data.list || []))
      }
      hasMoreSessions.value = res.data.data.hasMore
    }
  } catch (error) {
    console.error('加载会话失败', error)
  } finally {
    loadingSessions.value = false
  }
}

// 选择会话
const selectSession = async (sessionId: number) => {
  if (currentSessionId.value === sessionId) return
  currentSessionId.value = sessionId
  await loadMessages(sessionId, true)
  aiViewMode.value = 'chat'
  aiHistoryDrawerOpen.value = false
  scrollToBottom()
}

// 加载消息
const loadMessages = async (sessionId: number, reset = false) => {
  if (loadingMessages.value || (!hasMoreMessages.value && !reset)) return
  loadingMessages.value = true
  try {
    const lastId = reset ? undefined : aiMessages.value[0]?.id
    const res = await getAISessionMessages(sessionId, lastId)
    if (res.data?.code === 200 && res.data.data) {
      const list = res.data.data.list || []
      if (reset) {
        aiMessages.value = list
      } else {
        aiMessages.value.unshift(...list)
      }
      hasMoreMessages.value = res.data.data.hasMore
    }
  } catch (error) {
    console.error('加载消息失败', error)
  } finally {
    loadingMessages.value = false
  }
}

// 新建对话
const startNewAIChat = () => {
  currentSessionId.value = null
  aiViewMode.value = 'chat'
  aiHistoryDrawerOpen.value = false
  aiMessages.value = [
    {
      id: 0,
      sessionId: 0,
      parentId: null,
      role: 'assistant',
      content: '你好！我是你的 AI 助教。我可以帮你润色内容、生成代码示例或回答相关知识点。你想聊聊什么？',
      status: 'active',
      createdAt: new Date().toISOString()
    }
  ]
  hasMoreMessages.value = false
}

const openAIHistory = async () => {
  aiHistoryDrawerOpen.value = true
  if (!sessions.value.length) {
    await loadSessions(true)
  }
}

const closeAIHistory = () => {
  aiHistoryDrawerOpen.value = false
}

// SSE 处理相关
const parseSSEJson = <T>(data: string): T | null => {
  try {
    return JSON.parse(data) as T
  } catch (error) {
    console.error('SSE JSON 解析失败', error, data)
    return null
  }
}

const normalizeMessageChunk = (chunk: string) => {
  let trimmed = chunk.trim()
  if (trimmed.startsWith('"') && trimmed.endsWith('"')) {
    const parsed = parseSSEJson<string>(trimmed)
    if (typeof parsed === 'string') trimmed = parsed
  }
  try {
    const binaryString = atob(trimmed)
    const bytes = new Uint8Array(binaryString.length)
    for (let i = 0; i < binaryString.length; i++) {
      bytes[i] = binaryString.charCodeAt(i)
    }
    return new TextDecoder('utf-8').decode(bytes)
  } catch (error) {
    return trimmed
  }
}

const extractPdfDownloadUrl = (content: string) => {
  const markdownLinkMatch = content.match(/\[[^\]]*\.pdf[^\]]*\]\((\/v1\/ai\/exports\/[^)\s]+\.pdf)\)/i)
  if (markdownLinkMatch?.[1]) {
    return markdownLinkMatch[1]
  }

  const bareUrlMatch = content.match(/(^|[\s(])(\/v1\/ai\/exports\/[^\s)"']+\.pdf)/i)
  if (bareUrlMatch?.[2]) {
    return bareUrlMatch[2]
  }

  return ''
}

const triggerPdfDownload = async (downloadUrl: string) => {
  if (!downloadUrl) return

  try {
    const token = localStorage.getItem('token') || ''
    const res = await fetch(`http://localhost:8080${downloadUrl}`, {
      headers: {
        'x-token': token
      }
    })
    
    if (!res.ok) {
      throw new Error(`下载失败: ${res.status}`)
    }

    const blob = await res.blob()
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    // 尝试从 URL 提取文件名
    const fileNameMatch = downloadUrl.match(/\/([^/]+\.pdf)$/i)
    a.download = fileNameMatch ? fileNameMatch[1] : 'export.pdf'
    document.body.appendChild(a)
    a.click()
    a.remove()
    URL.revokeObjectURL(url)
  } catch (error) {
    console.error('PDF 下载失败:', error)
  }
}

const extractSSEEvents = (buffer: string) => {
  const events: { event: string; data: string }[] = []
  let rest = buffer
  while (true) {
    const delimiterIndex = rest.indexOf('\n\n')
    if (delimiterIndex === -1) break
    const rawEvent = rest.slice(0, delimiterIndex).trim()
    rest = rest.slice(delimiterIndex + 2)
    if (!rawEvent) continue
    let eventName = 'message'
    const dataLines: string[] = []
    rawEvent.split('\n').forEach((line) => {
      if (line.startsWith('event:')) {
        eventName = line.slice(6).trim()
      } else if (line.startsWith('data:')) {
        dataLines.push(line.slice(5).trimStart())
      }
    })
    events.push({ event: eventName, data: dataLines.join('\n') })
  }
  return { events, rest }
}

const showCopyToast = (message: string) => {
  copyToast.value = message
  if (copyToastTimer) clearTimeout(copyToastTimer)
  copyToastTimer = setTimeout(() => {
    copyToast.value = ''
  }, 1800)
}

const scrollToBottom = () => {
  nextTick(() => {
    if (aiMessagesContainer.value) {
      aiMessagesContainer.value.scrollTop = aiMessagesContainer.value.scrollHeight
    }
  })
}

const scrollToNextUserMsg = (direction: 'up' | 'down') => {
  const container = aiMessagesContainer.value
  if (!container) return

  const userMessages = aiMessages.value.filter(message => message.role === 'user')
  if (!userMessages.length) return

  let closestIdx = 0
  let minDiff = Number.POSITIVE_INFINITY

  userMessages.forEach((message, index) => {
    const element = document.getElementById(`ai-msg-${message.id}`)
    if (!element) return

    const diff = Math.abs(element.offsetTop - container.scrollTop - container.clientHeight / 3)
    if (diff < minDiff) {
      minDiff = diff
      closestIdx = index
    }
  })

  if (direction === 'down' && closestIdx === userMessages.length - 1) {
    scrollToBottom()
    return
  }

  const targetIdx = direction === 'up'
    ? Math.max(0, closestIdx - 1)
    : Math.min(userMessages.length - 1, closestIdx + 1)

  const targetElement = document.getElementById(`ai-msg-${userMessages[targetIdx].id}`)
  targetElement?.scrollIntoView({ behavior: 'smooth', block: 'center' })
}

const copyText = async (text: string) => {
  if (!text) return
  try {
    await navigator.clipboard.writeText(text)
    showCopyToast('内容已复制到剪贴板')
  } catch (error) {
    console.error('复制内容失败', error)
  }
}

const handleCodeCopy = async (event: MouseEvent) => {
  const button = (event.target as HTMLElement).closest('.code-copy-btn') as HTMLElement | null
  if (!button) return

  const rawCode = button.getAttribute('data-code') || ''
  const textarea = document.createElement('textarea')
  textarea.innerHTML = rawCode
  const decodedCode = textarea.value

  try {
    await navigator.clipboard.writeText(decodedCode)
    button.classList.add('copied')
    showCopyToast('代码已复制到剪贴板')
    setTimeout(() => {
      button.classList.remove('copied')
    }, 1800)
  } catch (error) {
    console.error('复制代码失败', error)
  }
}

const sendAIMessage = async () => {
  const prompt = aiInput.value.trim()
  if (!prompt || aiSending.value) return

  const tempId = Date.now()
  const parentId = aiMessages.value.length > 0 ? aiMessages.value[aiMessages.value.length - 1].id : 0
  
  const userMsg = reactive<AIChatMessage>({
    id: tempId,
    sessionId: currentSessionId.value || 0,
    parentId: parentId || null,
    role: 'user',
    content: prompt,
    status: 'active',
    createdAt: new Date().toISOString()
  })
  
  aiMessages.value.push(userMsg)
  aiInput.value = ''
  aiSending.value = true
  
  const assistantMsgId = Date.now() + 1
  const assistantMsg = reactive<AIChatMessage>({
    id: assistantMsgId,
    sessionId: currentSessionId.value || 0,
    parentId: tempId,
    role: 'assistant',
    content: '',
    reasoning: '',
    status: 'active',
    createdAt: new Date().toISOString()
  })
  aiMessages.value.push(assistantMsg)
  scrollToBottom()

  activeChatAbortController?.abort()
  const abortController = new AbortController()
  activeChatAbortController = abortController
  let streamFinished = false

  try {
    const token = localStorage.getItem('token') || ''
    const reqData = new FormData()
    reqData.append('prompt', prompt)
    if (currentSessionId.value) reqData.append('sessionId', currentSessionId.value.toString())
    if (parentId) reqData.append('parentId', parentId.toString())
    appendAIContext(reqData)

    const response = await fetch('http://localhost:8080/v1/ai/chat', {
      method: 'POST',
      headers: { 'x-token': token },
      body: reqData,
      signal: abortController.signal
    })

    if (!response.ok) throw new Error(`请求失败: ${response.status}`)
    if (!response.body) throw new Error('流式响应体为空')

    const reader = response.body.getReader()
    const decoder = new TextDecoder('utf-8')
    let sseBuffer = ''

    while (true) {
      const { value, done } = await reader.read()
      if (done) break

      sseBuffer += decoder.decode(value, { stream: true }).replace(/\r\n/g, '\n').replace(/\r/g, '\n')
      const { events, rest } = extractSSEEvents(sseBuffer)
      sseBuffer = rest

      for (const event of events) {
        if (event.event === 'meta') {
          const data = parseSSEJson<{ sessionId: number; messageId: number }>(event.data)
          if (data) {
            assistantMsg.sessionId = data.sessionId
            assistantMsg.id = data.messageId
            userMsg.sessionId = data.sessionId
            currentSessionId.value = data.sessionId
            loadSessions(true)
          }
        } else if (event.event === 'message') {
          assistantMsg.content += normalizeMessageChunk(event.data)
          scrollToBottom()
        } else if (event.event === 'reasoning') {
          assistantMsg.reasoning = (assistantMsg.reasoning || '') + normalizeMessageChunk(event.data)
          scrollToBottom()
        } else if (event.event === 'tool') {
          const toolText = normalizeMessageChunk(event.data)
          if (!assistantMsg.toolLogs) {
            assistantMsg.toolLogs = []
          }
          assistantMsg.toolLogs.push(toolText)
          scrollToBottom()
        } else if (event.event === 'done') {
          streamFinished = true
          aiSending.value = false
          
          // Check for download URL in the final message content
          const downloadUrl = extractPdfDownloadUrl(assistantMsg.content)
          triggerPdfDownload(downloadUrl)
          
          abortController.abort()
          return
        }
      }
    }
  } catch (err) {
    if (!abortController.signal.aborted || !streamFinished) {
      console.error('AI 聊天失败', err)
      assistantMsg.content = '抱歉，我现在遇到了一点问题，请稍后再试。'
    }
  } finally {
    if (activeChatAbortController === abortController) activeChatAbortController = null
    aiSending.value = false
  }
}

// 初始化加载会话
onMounted(() => {
  loadSessions(true)
  document.addEventListener('selectionchange', syncSelectedText)
  syncSelectedText()
})

onBeforeUnmount(() => {
  activeChatAbortController?.abort()
  if (copyToastTimer) clearTimeout(copyToastTimer)
  document.removeEventListener('selectionchange', syncSelectedText)
  window.removeEventListener('mousemove', onDragLeft)
  window.removeEventListener('mousemove', onDragRight)
  window.removeEventListener('mouseup', stopDrag)
})

let vditorInstance: Vditor | null = null

const md = markdownit({
  breaks: true,
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

md.renderer.rules.fence = (tokens, idx, options, _env, _slf) => {
  const token = tokens[idx]
  const info = token.info ? token.info.trim() : ''
  const langName = info.split(/\s+/g)[0]

  let highlighted = ''
  if (options.highlight) {
    highlighted = options.highlight(token.content, langName, '') || ''
  }

  if (!highlighted) {
    highlighted = md.utils.escapeHtml(token.content)
  }

  const rawCode = md.utils.escapeHtml(token.content)

  return `
    <div class="code-block-wrapper">
      <div class="code-block-header">
        <span class="code-lang">${langName || 'text'}</span>
        <button class="code-copy-btn" data-code="${rawCode}" title="复制代码">
          <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="14" height="14" x="8" y="8" rx="2" ry="2"/><path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"/></svg>
        </button>
      </div>
      <pre class="hljs"><code>${highlighted}</code></pre>
    </div>\n`
}

const stabilizeMarkdownForStreaming = (text: string) => {
  let processed = text

  const fenceMatches = processed.match(/```/g)
  if (fenceMatches && fenceMatches.length % 2 !== 0) {
    processed += '\n```\n'
  }

  const inlineCodeSegments = processed.replace(/```[\s\S]*?```/g, '')
  const backtickMatches = inlineCodeSegments.match(/`/g)
  if (backtickMatches && backtickMatches.length % 2 !== 0) {
    processed += '`'
  }

  const boldMatches = processed.match(/\*\*/g)
  if (boldMatches && boldMatches.length % 2 !== 0) {
    processed += '**'
  }

  return processed
}

const normalizePlainTextFenceContent = (text: string) => {
  return text
    .replace(/\\r\\n/g, '\n')
    .replace(/\\n/g, '\n')
    .replace(/\\r/g, '\n')
    .replace(/\\t/g, '\t')
}

const renderMarkdown = (text: string, isStreaming = false) => {
  if (!text) return ''

  let processed = text
  processed = processed.replace(/```([^\n`]*)\n([\s\S]*?)```/g, (match, info, code) => {
    const lang = info.trim().toLowerCase()
    if (!lang || ['text', 'plain', 'plaintext', 'txt'].includes(lang)) {
      return `\`\`\`${info}\n${normalizePlainTextFenceContent(code)}\`\`\``
    }
    return match
  })

  if (isStreaming) {
    processed = stabilizeMarkdownForStreaming(processed)
  }

  return md.render(processed)
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

const hasPendingDraftChanges = () => {
  return editName.value !== originalName.value || editContent.value !== originalContent.value
}

const updateEditorNameSilently = (value: string) => {
  isHydratingEditor.value = true
  editName.value = value
  originalName.value = value
  nextTick(() => {
    isHydratingEditor.value = false
  })
}

const clearAutoSaveTimers = () => {
  if (saveDebounceTimer) {
    clearTimeout(saveDebounceTimer)
    saveDebounceTimer = null
  }
  if (saveThrottleTimer) {
    clearTimeout(saveThrottleTimer)
    saveThrottleTimer = null
  }
}

const scheduleAutoSave = () => {
  if (!activeNodeId.value || !hasPendingDraftChanges()) return
  if (saveDebounceTimer) clearTimeout(saveDebounceTimer)
  saveDebounceTimer = setTimeout(() => {
    void queueSaveDraft()
  }, SAVE_DEBOUNCE_MS)
}

const queueSaveDraft = async () => {
  if (!activeNodeId.value || !hasPendingDraftChanges()) {
    saveStatus.value = 'saved'
    return
  }

  if (saving.value) {
    pendingSaveAfterCurrent.value = true
    return
  }

  const now = Date.now()
  const remaining = SAVE_THROTTLE_MS - (now - lastSaveTimestamp)
  if (remaining > 0) {
    if (saveThrottleTimer) clearTimeout(saveThrottleTimer)
    saveThrottleTimer = setTimeout(() => {
      void persistDraft()
    }, remaining)
    return
  }

  await persistDraft()
}

const flushPendingSave = async () => {
  clearAutoSaveTimers()
  if (hasPendingDraftChanges()) {
    await persistDraft()
  }
}

const persistDraft = async () => {
  if (!activeNodeId.value) return

  const currentNodeId = activeNodeId.value
  const currentNode = nodes.value.find(node => node.id === currentNodeId) || null
  const trimmedName = editName.value.trim()

  if (!trimmedName) {
    saveStatus.value = 'error'
    showToast('节点标题不能为空', 'error')
    return
  }

  if (trimmedName.length > TITLE_MAX_LENGTH) {
    saveStatus.value = 'error'
    showToast(`节点标题不能超过 ${TITLE_MAX_LENGTH} 个字符`, 'error')
    return
  }

  if (!hasPendingDraftChanges()) {
    saveStatus.value = 'saved'
    return
  }

  saveStatus.value = 'saving'
  saving.value = true

  try {
    let success = true

    if (!activeNodeNameReadonly.value && trimmedName !== originalName.value) {
      const resName = await updateKnowledgeNodeDraft(currentNodeId, {
        subjectId,
        nameDraft: trimmedName
      })

      if (resName.data?.code === 200) {
        updateEditorNameSilently(trimmedName)
        if (currentNode) {
          currentNode.nameDraft = trimmedName
          currentNode.hasDraft = 1
        }
      } else {
        success = false
      }
    }

    if (editContent.value !== originalContent.value) {
      const resContent = await upsertKnowledgeContent(currentNodeId, {
        contentDraft: editContent.value
      })

      if (resContent.data?.code === 200) {
        originalContent.value = editContent.value
        if (contentInfo.value) {
          contentInfo.value = {
            ...contentInfo.value,
            contentDraft: editContent.value,
            hasDraft: 1
          }
        }
        if (currentNode) {
          currentNode.hasDraft = 1
        }
      } else {
        success = false
      }
    }

    if (success) {
      lastSaveTimestamp = Date.now()
      saveStatus.value = 'saved'
    } else {
      saveStatus.value = 'error'
      showToast('自动保存失败，请稍后重试', 'error')
    }
  } catch (err: any) {
    console.error('保存失败', err)
    saveStatus.value = 'error'
    showToast(err?.response?.data?.msg || '自动保存失败，请稍后重试', 'error')
  } finally {
    saving.value = false
    if (saveThrottleTimer) {
      clearTimeout(saveThrottleTimer)
      saveThrottleTimer = null
    }
    if (pendingSaveAfterCurrent.value) {
      pendingSaveAfterCurrent.value = false
      void queueSaveDraft()
    }
  }
}

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
      
      const nextNodes = nodes.value.filter(n => !Number.isNaN(n.id)).map(n => {
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
    if (isAccordionMode.value) {
      // 手风琴模式：收起同级其他节点
      // 找出所有与当前节点 parentId 相同的节点
      const siblings = nodes.value.filter(n => n.parentId === node.parentId)
      siblings.forEach(sibling => {
        if (sibling.id !== node.id && expandedKeys.value.has(sibling.id)) {
          expandedKeys.value.delete(sibling.id)
        }
      })
    }

    expandedKeys.value.add(node.id)
    // 每次展开都去请求最新数据，保证数据是最新的
    await fetchChildren(node.id)
  }
}

// 选中节点
const handleNodeSelect = async (nodeId: number) => {
  const node = nodes.value.find(n => n.id === nodeId)
  if (!node) return

  if (activeNodeId.value && activeNodeId.value !== nodeId) {
    await flushPendingSave()
  }

  // 如果是非叶子节点，点击时自动处理展开逻辑
  if (node.isLeaf === 0) {
    if (!expandedKeys.value.has(node.id)) {
      if (isAccordionMode.value) {
        // 手风琴模式：收起同级其他节点
        const siblings = nodes.value.filter(n => n.parentId === node.parentId)
        siblings.forEach(sibling => {
          if (sibling.id !== node.id && expandedKeys.value.has(sibling.id)) {
            expandedKeys.value.delete(sibling.id)
          }
        })
      }
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
  createChildVisible.value = false
  newChildName.value = ''
  
  isHydratingEditor.value = true
  editName.value = node.nameDraft || node.name || '未命名节点'
  originalName.value = editName.value
  
  loadingContent.value = true
  saveStatus.value = 'saved'

  if (vditorInstance) {
    try {
      vditorInstance.destroy()
    } catch (e) {
      // ignore
    }
    vditorInstance = null
  }

  let newContent = ''

  try {
    const res = await getAuthorNodeContent(nodeId)
    if (res.data?.code === 200 && res.data.data) {
      contentInfo.value = res.data.data
      // 优先显示草稿，如果没有草稿则显示正式内容
      newContent = res.data.data.hasDraft === 1 ? res.data.data.contentDraft : res.data.data.content
    } else {
      contentInfo.value = null
    }
  } catch (err) {
    console.error('获取节点内容失败', err)
    contentInfo.value = null
  }

  editContent.value = newContent
  originalContent.value = newContent
  loadingContent.value = false

  await nextTick()

  // Initialize Vditor on the newly created DOM element
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
    }
  })

  isHydratingEditor.value = false
}

// 自动保存逻辑
watch([editName, editContent], ([newName, newContent]) => {
  if (!activeNodeId.value || isHydratingEditor.value) return

  if (newName !== originalName.value || newContent !== originalContent.value) {
    saveStatus.value = 'unsaved'
    scheduleAutoSave()
  } else if (!saving.value) {
    saveStatus.value = 'saved'
  }
})

const openCreateChild = () => {
  if (!activeNodeId.value) return
  newChildName.value = ''
  createChildVisible.value = true
}

const cancelCreateChild = () => {
  createChildVisible.value = false
  newChildName.value = ''
}

const createChildNode = async () => {
  if (!activeNodeId.value) return

  const nameDraft = newChildName.value.trim()
  if (!nameDraft) {
    showToast('新节点标题不能为空', 'error')
    return
  }
  if (nameDraft.length > TITLE_MAX_LENGTH) {
    showToast(`新节点标题不能超过 ${TITLE_MAX_LENGTH} 个字符`, 'error')
    return
  }

  if (creatingChild.value) return
  creatingChild.value = true

  await flushPendingSave()

  try {
    const parentId = activeNodeId.value
    const res = await createKnowledgeNode({
      subjectId,
      parentId,
      nameDraft
    })

    if (res.data?.code === 200 && res.data.data) {
      const data = res.data.data as any
      const newId = typeof data === 'object' ? Number(data.id || data.nodeId) : Number(data)
      const parentNode = nodes.value.find(node => node.id === parentId)
      if (parentNode) {
        parentNode.isLeaf = 0
      }
      
      // 乐观更新：将新节点预先推入 nodes.value 中，防止因为数据库延迟导致 fetchChildren 获取不到
      const isExist = nodes.value.find(n => n.id === newId)
      if (!isExist) {
        nodes.value.push({
          id: newId,
          subjectId: subjectId,
          parentId: parentId,
          name: '',
          nameDraft: nameDraft,
          status: 'draft',
          auditStatus: 0,
          hasDraft: 1,
          path: parentNode ? `${parentNode.path}/${newId}` : `${newId}`,
          isLeaf: 1,
          sortOrder: 99999 // 默认排在最后
        } as any)
      }

      expandedKeys.value.add(parentId)
      await fetchChildren(parentId)
      cancelCreateChild()
      showToast('子节点创建成功')
      
      // 立刻跳转选中新节点
      await handleNodeSelect(newId)
      
      // 滚动树列表到新节点位置，确保用户能看到
      await nextTick()
      setTimeout(() => {
        const activeEl = document.querySelector('.tree-node.active')
        if (activeEl) {
          activeEl.scrollIntoView({ behavior: 'smooth', block: 'center' })
        }
      }, 100)
    } else {
      showToast(res.data?.msg || '创建子节点失败', 'error')
    }
  } catch (err: any) {
    console.error('创建子节点失败', err)
    showToast(err?.response?.data?.msg || '创建子节点失败', 'error')
  } finally {
    creatingChild.value = false
  }
}

// --- Layout Dragging Logic ---
const sidebarWidth = ref(280)
const rightPanelWidth = ref(380)
const isDraggingLeft = ref(false)
const isDraggingRight = ref(false)

const startDragLeft = (e: MouseEvent) => {
  e.preventDefault()
  isDraggingLeft.value = true
  document.body.style.cursor = 'col-resize'
  window.addEventListener('mousemove', onDragLeft)
  window.addEventListener('mouseup', stopDrag)
}

const onDragLeft = (e: MouseEvent) => {
  if (!isDraggingLeft.value) return
  let newWidth = e.clientX
  if (newWidth < 200) newWidth = 200
  if (newWidth > 600) newWidth = 600
  sidebarWidth.value = newWidth
}

const startDragRight = (e: MouseEvent) => {
  e.preventDefault()
  isDraggingRight.value = true
  document.body.style.cursor = 'col-resize'
  window.addEventListener('mousemove', onDragRight)
  window.addEventListener('mouseup', stopDrag)
}

const onDragRight = (e: MouseEvent) => {
  if (!isDraggingRight.value) return
  let newWidth = window.innerWidth - e.clientX
  if (newWidth < 250) newWidth = 250
  if (newWidth > 800) newWidth = 800
  rightPanelWidth.value = newWidth
}

const stopDrag = () => {
  if (isDraggingLeft.value || isDraggingRight.value) {
    isDraggingLeft.value = false
    isDraggingRight.value = false
    document.body.style.cursor = ''
    window.removeEventListener('mousemove', onDragLeft)
    window.removeEventListener('mousemove', onDragRight)
    window.removeEventListener('mouseup', stopDrag)
    // 拖拽结束后，如果使用了 Vditor 等可能需要重新计算宽度的组件，可以触发 resize
    window.dispatchEvent(new Event('resize'))
  }
}

interface AuthorTreeNode extends AuthorNode {
  children: AuthorTreeNode[]
  hasDescendantDraft?: boolean
}

interface VisibleTreeNode extends AuthorNode {
  depth: number
  hasDescendantDraft?: boolean
}

// 构建树形结构供渲染
const treeNodes = computed<AuthorTreeNode[]>(() => {
  const buildTree = (parentId: number): AuthorTreeNode[] => {
    return nodes.value
      .filter(n => Number(n.parentId) === parentId)
      .map(n => {
        const children = buildTree(n.id)
        const hasDescendantDraft = children.some(
          child => child.status === 'draft' || child.hasDraft === 1 || child.hasDescendantDraft
        )
        return {
          ...n,
          children,
          hasDescendantDraft
        }
      })
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
        depth,
        hasDescendantDraft: item.hasDescendantDraft
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
    <Toast
      v-if="toast.show"
      :message="toast.message"
      :type="toast.type"
      @close="toast.show = false"
    />
    <!-- 侧边栏 -->
    <aside class="sidebar" :style="{ width: sidebarWidth + 'px' }">
      <div class="sidebar-header">
        <button class="back-btn" @click="router.back()">
          <ArrowLeft :size="16" />
          返回
        </button>
        <div class="sidebar-title-wrapper">
          <h2 class="sidebar-title">教材目录</h2>
          <button 
            class="accordion-toggle-btn" 
            @click="isAccordionMode = !isAccordionMode"
            :title="isAccordionMode ? '当前为手风琴模式，点击切换为普通模式' : '当前为普通模式，点击切换为手风琴模式'"
          >
            <AlignVerticalSpaceAround v-if="isAccordionMode" :size="14" />
            <List v-else :size="14" />
          </button>
        </div>
      </div>
      
      <div class="sidebar-content">
        <div v-if="loadingTree" class="loading-state">加载中...</div>
        <div v-else-if="treeNodes.length === 0" class="empty-state">暂无目录节点</div>
        <TransitionGroup name="tree-list" tag="div" v-else class="tree-container">
          <div
            v-for="node in visibleTreeNodes"
            :key="node.id"
            class="tree-node"
            :class="{ 'active': activeNodeId === node.id }"
            @click="handleNodeSelect(node.id)"
          >
            <div class="node-content" :style="{ paddingLeft: `${8 + node.depth * 16}px` }">
              <span v-if="node.isLeaf !== 1" class="expand-icon" @click="toggleExpand(node, $event)">
                <ChevronRight :size="14" :class="['expand-svg', { 'is-expanded': expandedKeys.has(node.id) }]" />
              </span>
              <span v-else class="expand-placeholder"></span>
              <div class="node-icon-wrapper">
                <FileText v-if="node.isLeaf === 1" :size="14" class="node-icon file" />
                <Folder v-else :size="14" class="node-icon folder" />
                <Pencil v-if="node.status === 'draft' || node.hasDraft === 1" :size="12" class="draft-dot-icon" fill="currentColor" />
                <span v-else-if="node.hasDescendantDraft" class="descendant-draft-dot"></span>
              </div>
              <span class="node-title">{{ node.nameDraft || node.name || '未命名' }}</span>
            </div>
          </div>
        </TransitionGroup>
      </div>
    </aside>

    <!-- 左侧边框 (拖拽条) -->
    <div class="resizer left-resizer" @mousedown="startDragLeft"></div>

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
            <button class="btn-primary" type="button" @click="openCreateChild" :disabled="creatingChild">
              <Plus :size="14" />
              新建子节点
            </button>
          </div>
        </header>
        
        <div class="editor-body">
          <!-- 左侧：草稿编辑 -->
          <div class="pane draft-pane">
            <div class="pane-content pane-content-vditor">
              <div class="draft-edit-area">
                <div v-if="createChildVisible" class="create-child-panel">
                  <div class="create-child-header">
                    <span class="create-child-title">在当前节点下创建子节点</span>
                    <span class="create-child-hint">输入标题后立即创建并自动选中新节点</span>
                  </div>
                  <div class="create-child-actions">
                    <input
                      v-model="newChildName"
                      class="child-input"
                      placeholder="输入新节点标题..."
                      maxlength="150"
                      autocomplete="off"
                      @keydown.enter.prevent="createChildNode"
                    />
                    <button class="btn-primary" type="button" @click="createChildNode" :disabled="creatingChild">
                      <LoaderCircle v-if="creatingChild" :size="14" class="spin" />
                      <Plus v-else :size="14" />
                      创建
                    </button>
                    <button class="btn-secondary" type="button" @click="cancelCreateChild" :disabled="creatingChild">
                      取消
                    </button>
                  </div>
                </div>
                <input 
                  v-model="editName" 
                  class="title-input" 
                  placeholder="输入草稿标题..."
                  autocomplete="off"
                  maxlength="150"
                  :readonly="activeNodeNameReadonly"
                  :class="{ readonly: activeNodeNameReadonly }"
                />
                <div class="field-tip" v-if="activeNodeNameReadonly">
                  <span>顶级节点标题不可在此处修改，请通过教材修改接口同步。</span>
                </div>
                <div id="vditor-container" class="vditor-wrapper"></div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </main>

    <!-- 右侧边框 (拖拽条) -->
    <div class="resizer right-resizer" @mousedown="startDragRight"></div>

    <!-- 右侧多功能面板 -->
    <aside class="right-panel" :class="{ 'collapsed': !showPublishedPane, 'is-dragging': isDraggingRight }" :style="showPublishedPane ? { width: rightPanelWidth + 'px' } : {}">
      <div class="panel-toggle" @click="showPublishedPane = !showPublishedPane">
        <ChevronRight v-if="showPublishedPane" :size="16" />
        <MessageSquare v-else :size="16" />
      </div>

      <template v-if="showPublishedPane">
        <div class="panel-tabs">
          <div 
            class="panel-tab" 
            :class="{ active: rightPanelTab === 'published' }"
            @click="rightPanelTab = 'published'"
          >
            <FileText :size="14" />
            已发布
          </div>
          <div 
            class="panel-tab" 
            :class="{ active: rightPanelTab === 'ai' }"
            @click="rightPanelTab = 'ai'"
          >
            <Sparkles :size="14" />
            AI 助手
          </div>
        </div>

        <div class="panel-content">
          <!-- 已发布内容展示 -->
          <div v-if="rightPanelTab === 'published'" class="published-view">
            <div class="published-header">
              <span class="published-title">{{ activeNode?.name || '无标题' }}</span>
            </div>
            <div class="published-body">
              <div class="markdown-body" v-html="renderMarkdown(contentInfo?.content || '')"></div>
              <div v-if="!contentInfo?.content" class="empty-content">暂无已发布正文</div>
            </div>
          </div>

          <!-- AI 聊天界面 -->
          <div v-else-if="rightPanelTab === 'ai'" class="ai-view">
            <Transition name="toast">
              <div v-if="copyToast" class="toast-notification">
                {{ copyToast }}
              </div>
            </Transition>

            <div class="ai-topbar" :class="{ compact: aiViewMode !== 'chat' }">
              <div class="ai-topbar-title">
                <Sparkles :size="16" />
                <span>{{ aiViewMode === 'chat' ? currentAISessionTitle : 'AI 助手' }}</span>
              </div>
              <div class="ai-topbar-actions">
                <button class="ai-topbar-btn" type="button" @click="openAIHistory" title="查看历史对话">
                  <MessageSquare :size="16" />
                </button>
                <button class="ai-topbar-btn" type="button" @click="startNewAIChat" title="新建对话">
                  <Plus :size="16" />
                </button>
              </div>
            </div>

            <div v-if="aiViewMode === 'entry'" class="ai-entry">
              <div class="ai-entry-hero">
                <div class="ai-entry-logo">
                  <Sparkles :size="20" />
                </div>
                <h3>今天想让 AI 帮你做什么？</h3>
                <p>在写作和编辑过程中，随时切换历史会话，或者开启一个新的上下文。</p>
              </div>
              <div class="ai-entry-actions">
                <button class="ai-entry-card" type="button" @click="openAIHistory">
                  <div class="ai-entry-card-icon">
                    <MessageSquare :size="18" />
                  </div>
                  <div class="ai-entry-card-body">
                    <span class="ai-entry-card-title">历史对话</span>
                    <span class="ai-entry-card-desc">查看并快速切换之前的会话</span>
                  </div>
                </button>
                <button class="ai-entry-card" type="button" @click="startNewAIChat">
                  <div class="ai-entry-card-icon">
                    <Plus :size="18" />
                  </div>
                  <div class="ai-entry-card-body">
                    <span class="ai-entry-card-title">新建对话</span>
                    <span class="ai-entry-card-desc">创建一个全新的上下文，重新开始提问</span>
                  </div>
                </button>
              </div>
            </div>

            <template v-else>
              <div class="ai-messages" ref="aiMessagesContainer" @click="handleCodeCopy">
                <div
                  v-for="msg in aiMessages"
                  :id="'ai-msg-' + msg.id"
                  :key="msg.id"
                  class="ai-message"
                  :class="msg.role"
                >
                  <div class="message-avatar">
                    <Bot v-if="msg.role === 'assistant'" :size="14" />
                    <User v-else :size="14" />
                  </div>
                  <div class="message-content">
                    <div class="message-bubble-container">
                      <div class="message-bubble">
                      
                      <!-- Tool Block -->
                      <div v-if="msg.toolLogs && msg.toolLogs.length > 0" class="tool-block">
                        <div class="tool-summary">
                          <Loader2 v-if="isStreamingAssistantMessage(msg) && !msg.content" class="spin" :size="12" />
                          <Wrench v-else :size="12" />
                          <span>工具调用过程 ({{ msg.toolLogs.length }} 步)</span>
                        </div>
                        <div class="tool-logs">
                          <div v-for="(log, idx) in msg.toolLogs" :key="idx" class="tool-item">
                            <span class="tool-dot"></span>
                            <span>{{ log }}</span>
                          </div>
                        </div>
                      </div>

                      <div v-if="msg.role === 'assistant' && msg.reasoning" class="ai-reasoning">
                          <div class="reasoning-header" @click="toggleReasoning(msg.id)">
                            <LoaderCircle
                              v-if="isStreamingAssistantMessage(msg) && !msg.content"
                              :size="12"
                              class="spin"
                            />
                            <Sparkles v-else :size="12" />
                            <span>思考过程</span>
                            <ChevronDown
                              v-if="isReasoningCollapsed(msg)"
                              :size="12"
                              class="collapse-icon collapsed"
                            />
                            <ChevronUp v-else :size="12" class="collapse-icon" />
                          </div>
                          <div v-show="!isReasoningCollapsed(msg)" class="reasoning-body">
                            <div
                              class="markdown-body"
                              v-html="renderMarkdown(msg.reasoning, isStreamingAssistantMessage(msg) && !msg.content)"
                            ></div>
                          </div>
                        </div>

                        <div
                          v-if="msg.role === 'assistant' && !msg.content && !msg.reasoning && isStreamingAssistantMessage(msg)"
                          class="message-typing"
                        >
                          <span class="dot"></span>
                          <span class="dot"></span>
                          <span class="dot"></span>
                        </div>

                        <div
                          v-else-if="msg.role === 'assistant' && !msg.content && msg.reasoning && isStreamingAssistantMessage(msg)"
                          class="message-status"
                        >
                          <LoaderCircle class="spin" :size="14" />
                          <span>正在整理回答...</span>
                        </div>

                        <div
                          v-if="msg.role === 'assistant' && msg.content"
                          class="markdown-body"
                          :class="{ 'is-streaming': isStreamingAssistantMessage(msg) }"
                          v-html="renderMarkdown(msg.content, isStreamingAssistantMessage(msg))"
                        ></div>

                        <div
                          v-else-if="msg.role === 'user'"
                          class="markdown-body"
                          v-html="renderMarkdown(msg.content)"
                        ></div>
                      </div>

                      <button
                        v-if="msg.content"
                        class="copy-btn"
                        type="button"
                        title="复制"
                        @click.stop="copyText(msg.content)"
                      >
                        <Copy :size="16" />
                      </button>
                    </div>
                  </div>
                </div>
              </div>

              <div v-if="aiMessages.some(msg => msg.role === 'user')" class="global-nav-actions">
                <button class="nav-btn" type="button" title="上一条我的消息" @click="scrollToNextUserMsg('up')">
                  <ArrowUpCircle :size="22" />
                </button>
                <button class="nav-btn" type="button" title="下一条我的消息" @click="scrollToNextUserMsg('down')">
                  <ArrowDownCircle :size="22" />
                </button>
              </div>

              <div class="ai-input-area">
                <textarea 
                  v-model="aiInput" 
                  placeholder="询问 AI 助教..." 
                  @keydown.enter.prevent="sendAIMessage"
                  :disabled="aiSending"
                  rows="1"
                ></textarea>
                <button class="ai-send-btn" @click="sendAIMessage" :disabled="!aiInput.trim() || aiSending">
                  <Send v-if="!aiSending" :size="16" />
                  <LoaderCircle v-else :size="16" class="spin" />
                </button>
              </div>
            </template>

            <Transition name="ai-drawer">
              <div v-if="aiHistoryDrawerOpen" class="ai-history-overlay" @click.self="closeAIHistory">
                <div class="ai-history-drawer">
                  <div class="ai-history-drawer-header">
                    <div class="ai-history-drawer-title">
                      <MessageSquare :size="16" />
                      <span>历史对话</span>
                    </div>
                    <button class="ai-history-close" type="button" @click="closeAIHistory">
                      <ChevronRight :size="16" />
                    </button>
                  </div>
                  <div class="ai-history-drawer-body">
                    <button
                      v-for="session in sessions"
                      :key="session.id"
                      class="ai-history-item"
                      type="button"
                      @click="selectSession(session.id)"
                    >
                      <div class="ai-history-item-icon">
                        <MessageSquare :size="16" />
                      </div>
                      <div class="ai-history-item-body">
                        <span class="ai-history-item-title">{{ session.title }}</span>
                        <span class="ai-history-item-time">{{ new Date(session.updatedAt).toLocaleString() }}</span>
                      </div>
                    </button>
                    <div v-if="loadingSessions" class="ai-history-empty">
                      正在加载历史对话...
                    </div>
                    <div v-else-if="!sessions.length" class="ai-history-empty">
                      暂无历史对话，点击“新建对话”开始使用。
                    </div>
                  </div>
                </div>
              </div>
            </Transition>
          </div>
        </div>
      </template>
    </aside>
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
  width: 280px; /* 默认值，将被内联样式覆盖 */
  background-color: #f6f5f4;
  border-right: 1px solid rgba(0, 0, 0, 0.1);
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}

/* 拖拽边框 */
.resizer {
  width: 4px;
  background-color: transparent;
  cursor: col-resize;
  position: relative;
  z-index: 10;
  transition: background-color 0.2s;
  flex-shrink: 0;
}

.resizer:hover,
.resizer:active {
  background-color: rgba(63, 58, 53, 0.2);
}

.left-resizer {
  margin-left: -2px;
  margin-right: -2px;
}

.right-resizer {
  margin-left: -2px;
  margin-right: -2px;
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

.sidebar-title-wrapper {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: 8px;
}

.sidebar-title {
  font-size: 16px;
  font-weight: 600;
  margin: 0;
}

.accordion-toggle-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  color: rgba(0, 0, 0, 0.45);
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  transition: all 0.2s;
}

.accordion-toggle-btn:hover {
  background-color: rgba(0, 0, 0, 0.05);
  color: rgba(0, 0, 0, 0.85);
}

.sidebar-content {
  flex: 1;
  overflow-y: auto;
  padding: 12px 0;
}

.tree-container {
  display: flex;
  flex-direction: column;
  position: relative;
}

/* Tree Node Animations */
.tree-list-move,
.tree-list-enter-active,
.tree-list-leave-active {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
}

.tree-list-enter-from,
.tree-list-leave-to {
  opacity: 0;
  max-height: 0 !important;
  padding-top: 0 !important;
  padding-bottom: 0 !important;
  margin-top: 0 !important;
  margin-bottom: 0 !important;
  transform: translateX(-8px);
}

.tree-list-enter-to,
.tree-list-leave-from {
  opacity: 1;
  max-height: 40px; /* Assuming 36px is the typical height */
  transform: translateX(0);
}

.tree-node {
  padding: 4px 12px;
  cursor: pointer;
  user-select: none;
  max-height: 40px; /* Setting explicit max-height for animation */
  box-sizing: border-box;
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
  cursor: pointer;
}

.expand-svg {
  transition: transform 0.25s cubic-bezier(0.4, 0, 0.2, 1);
}

.expand-svg.is-expanded {
  transform: rotate(90deg);
}

.expand-placeholder {
  width: 20px;
  flex-shrink: 0;
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

.node-icon-wrapper {
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  margin-right: 6px;
}

.node-title {
  font-size: 14px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex: 1;
}

.draft-dot-icon {
  position: absolute;
  top: -6px;
  right: -8px;
  color: #3b82f6; /* 静态蓝色 */
  background-color: transparent; /* 透明背景 */
  flex-shrink: 0;
  z-index: 1;
}

.descendant-draft-dot {
  position: absolute;
  top: -3px;
  right: -5px;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: #eab308; /* 黄色 */
  z-index: 1;
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

.btn-secondary {
  display: flex;
  align-items: center;
  gap: 6px;
  background-color: #ffffff;
  color: #615d59;
  border: 1px solid rgba(0, 0, 0, 0.12);
  border-radius: 4px;
  padding: 6px 12px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-secondary:hover:not(:disabled) {
  background-color: #f6f5f4;
  color: rgba(0, 0, 0, 0.95);
}

.btn-secondary:disabled {
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

/* 右侧面板 */
.right-panel {
  width: 380px; /* 默认值，将被内联样式覆盖 */
  background-color: #f6f5f4;
  border-left: 1px solid rgba(0, 0, 0, 0.1);
  display: flex;
  flex-direction: column;
  position: relative;
  transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  flex-shrink: 0;
}

.right-panel.collapsed {
  width: 40px !important; /* 强制覆盖内联宽度的拖拽状态 */
}

.right-panel.is-dragging {
  transition: none !important; /* 拖拽时禁用过渡动画以保证丝滑 */
}

.panel-toggle {
  position: absolute;
  left: -12px;
  top: 50%;
  transform: translateY(-50%);
  width: 24px;
  height: 24px;
  background: #ffffff;
  border: 1px solid rgba(0, 0, 0, 0.1);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  z-index: 10;
  color: #615d59;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.panel-toggle:hover {
  background: #f6f5f4;
  color: #000000;
}

.panel-tabs {
  display: flex;
  padding: 16px 16px 0;
  gap: 4px;
  border-bottom: 1px solid rgba(0, 0, 0, 0.05);
}

.panel-tab {
  padding: 8px 12px;
  font-size: 13px;
  font-weight: 500;
  color: #615d59;
  cursor: pointer;
  border-radius: 6px 6px 0 0;
  display: flex;
  align-items: center;
  gap: 6px;
  transition: all 0.2s;
}

.panel-tab:hover {
  background-color: rgba(0, 0, 0, 0.05);
}

.panel-tab.active {
  background-color: #ffffff;
  color: #000000;
  box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.02);
}

.panel-content {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  background-color: #ffffff;
}

/* 已发布视图 */
.published-view {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.published-header {
  padding: 20px 24px;
  border-bottom: 1px solid rgba(0, 0, 0, 0.05);
}

.published-title {
  font-size: 18px;
  font-weight: 600;
  color: #000000;
}

.published-body {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
}

/* AI 助手样式 */
.ai-view {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  position: relative;
  background:
    radial-gradient(circle at top, rgba(93, 84, 74, 0.08), transparent 36%),
    linear-gradient(180deg, #fcfbf9 0%, #f8f6f3 100%);
}

.ai-topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 14px 16px 10px;
  background-color: rgba(252, 251, 249, 0.72);
  backdrop-filter: blur(10px);
  z-index: 2;
}

.ai-topbar.compact {
  background-color: transparent;
}

.ai-topbar-title {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
  font-size: 14px;
  font-weight: 600;
  color: #3b3834;
}

.ai-topbar-title span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ai-topbar-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.ai-topbar-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  padding: 0;
  border: 1px solid rgba(63, 58, 53, 0.09);
  border-radius: 50%;
  background-color: rgba(255, 255, 255, 0.84);
  color: #5c5752;
  cursor: pointer;
  box-shadow: 0 4px 10px rgba(63, 58, 53, 0.04);
  transition: all 0.2s ease;
}

.ai-topbar-btn:hover {
  border-color: rgba(63, 58, 53, 0.18);
  background-color: #ffffff;
  color: #3f3a35;
}

.ai-entry {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 28px 24px 32px;
}

.ai-entry-hero {
  text-align: center;
  margin-bottom: 24px;
}

.ai-entry-logo {
  width: 52px;
  height: 52px;
  margin: 0 auto 16px;
  border-radius: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #3f3a35 0%, #6b635b 100%);
  color: #ffffff;
  box-shadow: 0 10px 24px rgba(63, 58, 53, 0.18);
}

.ai-entry-hero h3 {
  margin: 0;
  font-size: 24px;
  line-height: 1.3;
  color: #24211f;
}

.ai-entry-hero p {
  margin: 10px 0 0;
  font-size: 13px;
  line-height: 1.7;
  color: #8a847d;
}

.ai-entry-actions {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.ai-entry-card {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 16px 18px;
  border: 1px solid rgba(63, 58, 53, 0.08);
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.92);
  text-align: left;
  cursor: pointer;
  box-shadow: 0 10px 24px rgba(63, 58, 53, 0.05);
  transition: transform 0.2s ease, box-shadow 0.2s ease, border-color 0.2s ease, background 0.2s ease;
}

.ai-entry-card:hover {
  transform: translateY(-1px);
  border-color: rgba(63, 58, 53, 0.16);
  background: #ffffff;
  box-shadow: 0 14px 28px rgba(63, 58, 53, 0.08);
}

.ai-entry-card-icon {
  width: 40px;
  height: 40px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #f2efeb;
  color: #3f3a35;
  flex-shrink: 0;
}

.ai-entry-card-body {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.ai-entry-card-title {
  font-size: 14px;
  font-weight: 600;
  color: #24211f;
}

.ai-entry-card-desc {
  font-size: 12px;
  color: #8a847d;
  line-height: 1.5;
}

.ai-history-overlay {
  position: absolute;
  inset: 0;
  z-index: 4;
  background: rgba(32, 28, 24, 0.18);
  backdrop-filter: blur(4px);
}

.ai-history-drawer {
  width: 88%;
  max-width: 320px;
  height: 100%;
  background: rgba(252, 251, 249, 0.98);
  border-right: 1px solid rgba(63, 58, 53, 0.08);
  box-shadow: 12px 0 32px rgba(32, 28, 24, 0.12);
  display: flex;
  flex-direction: column;
}

.ai-history-drawer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 16px;
  border-bottom: 1px solid rgba(63, 58, 53, 0.08);
}

.ai-history-drawer-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 600;
  color: #2f2b27;
}

.ai-history-close {
  width: 28px;
  height: 28px;
  border-radius: 999px;
  border: 1px solid rgba(63, 58, 53, 0.08);
  background: #ffffff;
  color: #5c5752;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.ai-history-drawer-body {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 14px 12px 16px;
}

.ai-history-item {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px;
  border: 1px solid rgba(0, 0, 0, 0.08);
  border-radius: 16px;
  background-color: #ffffff;
  text-align: left;
  cursor: pointer;
  transition: all 0.2s ease;
}

.ai-history-item:hover {
  border-color: rgba(63, 58, 53, 0.18);
  box-shadow: 0 10px 20px rgba(63, 58, 53, 0.06);
}

.ai-history-item-icon {
  width: 36px;
  height: 36px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  background-color: #f2efeb;
  color: #4b4843;
}

.ai-history-item-body {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.ai-history-item-title {
  font-size: 13px;
  font-weight: 600;
  color: #24211f;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.ai-history-item-time {
  font-size: 11px;
  color: #9a948d;
}

.ai-history-empty {
  padding: 24px 12px;
  text-align: center;
  font-size: 12px;
  color: #8a847d;
}

.ai-drawer-enter-active,
.ai-drawer-leave-active {
  transition: opacity 0.22s ease;
}

.ai-drawer-enter-active .ai-history-drawer,
.ai-drawer-leave-active .ai-history-drawer {
  transition: transform 0.22s ease;
}

.ai-drawer-enter-from,
.ai-drawer-leave-to {
  opacity: 0;
}

.ai-drawer-enter-from .ai-history-drawer,
.ai-drawer-leave-to .ai-history-drawer {
  transform: translateX(-18px);
}

.ai-messages {
  flex: 1;
  overflow-y: auto;
  padding: 8px 20px 20px;
  display: flex;
  flex-direction: column;
  gap: 22px;
  scroll-behavior: smooth;
}

.ai-message {
  display: flex;
  gap: 12px;
  max-width: 94%;
}

.ai-message.assistant {
  align-self: flex-start;
}

.ai-message.user {
  align-self: flex-end;
  flex-direction: row-reverse;
}

.message-avatar {
  width: 30px;
  height: 30px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 8px 18px rgba(63, 58, 53, 0.08);
}

.ai-message.assistant .message-avatar {
  background-color: #4b4843;
  color: white;
}

.ai-message.user .message-avatar {
  background-color: #a39e98;
  color: white;
}

.message-content {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-width: 100%;
}

.message-bubble-container {
  display: flex;
  align-items: flex-end;
  gap: 8px;
}

.ai-message.user .message-bubble-container {
  flex-direction: row-reverse;
}

.message-bubble {
  padding: 12px 14px;
  border-radius: 16px;
  font-size: 14px;
  line-height: 1.6;
  position: relative;
  max-width: 100%;
  overflow-x: auto;
}

.ai-message.assistant .message-bubble {
  background-color: rgba(255, 255, 255, 0.9);
  border: 1px solid rgba(63, 58, 53, 0.06);
  box-shadow: 0 8px 24px rgba(63, 58, 53, 0.05);
  border-top-left-radius: 6px;
}

/* Tool Block Styles */
.tool-block {
  margin-bottom: 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  background: rgba(59, 130, 246, 0.05);
  border: 1px solid rgba(59, 130, 246, 0.1);
  padding: 10px 12px;
  border-radius: 8px;
}

.tool-summary {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: #3b82f6;
  font-weight: 500;
}

.tool-logs {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding-left: 6px;
  border-left: 2px solid rgba(59, 130, 246, 0.2);
  margin-left: 6px;
}

.tool-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: #64748b;
}

.tool-dot {
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background: #3b82f6;
  flex-shrink: 0;
}

.ai-message.user .message-bubble {
  background: linear-gradient(135deg, #4b4843 0%, #605850 100%);
  color: white;
  border-top-right-radius: 6px;
  box-shadow: 0 10px 24px rgba(63, 58, 53, 0.12);
}

/* Typing Cursor Effect */
.is-streaming::after {
  content: '▋';
  display: inline-block;
  animation: blink 1s step-end infinite;
  color: #3b82f6;
  margin-left: 2px;
  vertical-align: baseline;
}

@keyframes blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0; }
}

.ai-message.user :deep(.markdown-body) {
  color: white;
}

.message-bubble.loading {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #8a847d;
}

.message-status {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #8a847d;
  font-size: 13px;
}

.message-typing {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  min-height: 22px;
}

.message-typing .dot {
  width: 6px;
  height: 6px;
  border-radius: 999px;
  background-color: #8a847d;
  animation: typing-bounce 1.2s infinite ease-in-out;
}

.message-typing .dot:nth-child(2) {
  animation-delay: 0.15s;
}

.message-typing .dot:nth-child(3) {
  animation-delay: 0.3s;
}

.copy-btn {
  opacity: 0;
  background: transparent;
  border: none;
  color: #9a948d;
  cursor: pointer;
  padding: 4px;
  border-radius: 8px;
  transition: all 0.2s ease;
  flex-shrink: 0;
}

.ai-message:hover .copy-btn {
  opacity: 1;
}

.copy-btn:hover {
  color: #4b4843;
  background: rgba(0, 0, 0, 0.04);
}

.global-nav-actions {
  position: absolute;
  right: 18px;
  bottom: 96px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  opacity: 0.28;
  transition: opacity 0.2s ease;
  z-index: 3;
}

.global-nav-actions:hover {
  opacity: 1;
}

.nav-btn {
  width: 38px;
  height: 38px;
  border-radius: 999px;
  border: 1px solid rgba(63, 58, 53, 0.08);
  background: rgba(255, 255, 255, 0.92);
  color: #6c655f;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  box-shadow: 0 8px 24px rgba(63, 58, 53, 0.08);
  transition: all 0.2s ease;
}

.nav-btn:hover {
  color: #4b4843;
  background: #ffffff;
  border-color: rgba(63, 58, 53, 0.16);
}

.ai-input-area {
  margin: 0 16px 16px;
  padding: 10px;
  border: 1px solid rgba(63, 58, 53, 0.08);
  border-radius: 18px;
  background-color: rgba(255, 255, 255, 0.92);
  box-shadow: 0 10px 30px rgba(63, 58, 53, 0.08);
  display: flex;
  gap: 8px;
  align-items: flex-end;
}

.ai-input-area textarea {
  flex: 1;
  border: none;
  border-radius: 12px;
  padding: 10px 12px;
  font-size: 14px;
  resize: none;
  max-height: 120px;
  transition: background-color 0.2s;
  background-color: transparent;
  line-height: 1.6;
}

.ai-input-area textarea:focus {
  outline: none;
  background-color: rgba(0, 0, 0, 0.015);
}

.ai-send-btn {
  width: 36px;
  height: 36px;
  border-radius: 12px;
  background: linear-gradient(135deg, #4b4843 0%, #605850 100%);
  color: white;
  border: none;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: opacity 0.2s, transform 0.2s ease;
  flex-shrink: 0;
  margin-bottom: 2px;
  box-shadow: 0 10px 24px rgba(63, 58, 53, 0.16);
}

.ai-send-btn:hover:not(:disabled) {
  transform: translateY(-1px);
}

.ai-send-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

@keyframes typing-bounce {
  0%, 80%, 100% {
    transform: scale(0.8);
    opacity: 0.5;
  }
  40% {
    transform: scale(1);
    opacity: 1;
  }
}

/* 思考过程样式 */
.ai-reasoning {
  border: 1px solid rgba(63, 58, 53, 0.05);
  border-radius: 14px;
  background-color: rgba(255, 255, 255, 0.66);
  overflow: hidden;
}

.reasoning-header {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 11px 12px;
  color: #8a847d;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
}

.collapse-icon {
  margin-left: auto;
  transition: transform 0.2s ease;
}

.collapse-icon.collapsed {
  transform: rotate(-90deg);
}

.reasoning-body {
  padding: 0 12px 12px;
  border-top: 1px solid rgba(63, 58, 53, 0.05);
}

.toast-notification {
  position: absolute;
  top: 14px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 6;
  padding: 8px 14px;
  border-radius: 999px;
  background: rgba(36, 33, 31, 0.92);
  color: #ffffff;
  font-size: 12px;
  box-shadow: 0 10px 24px rgba(0, 0, 0, 0.16);
}

.toast-enter-active,
.toast-leave-active {
  transition: opacity 0.18s ease, transform 0.18s ease;
}

.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateX(-50%) translateY(-4px);
}

.ai-messages::-webkit-scrollbar {
  width: 4px;
}

.ai-messages::-webkit-scrollbar-thumb {
  background: rgba(0,0,0,0.1);
  border-radius: 2px;
}

.ai-messages::-webkit-scrollbar-track {
  background: transparent;
}

.ai-history-drawer-body::-webkit-scrollbar {
  width: 4px;
}

.ai-history-drawer-body::-webkit-scrollbar-thumb {
  background: rgba(0, 0, 0, 0.1);
  border-radius: 2px;
}

:deep(.code-block-wrapper) {
  position: relative;
  margin: 16px 0;
  border-radius: 12px;
  overflow: hidden;
  background-color: #f4f5f9;
}

:deep(.code-block-header) {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 14px;
  background-color: #f4f5f9;
  color: #333;
  font-size: 12px;
  border-bottom: 2px solid #ffffff;
}

:deep(.code-lang) {
  text-transform: lowercase;
}

:deep(.code-copy-btn) {
  background: transparent;
  border: none;
  color: #475569;
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

:deep(.code-copy-btn:hover) {
  background-color: rgba(0, 0, 0, 0.05);
  color: #1e293b;
}

:deep(.code-copy-btn.copied) {
  color: #10b981;
}

:deep(.markdown-body pre) {
  margin: 0;
  padding: 16px;
  overflow: auto;
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
  background-color: #ffffff;
}

.vditor-wrapper {
  flex: 1;
  min-height: 0; /* Crucial for flex scrolling inside Vditor */
  border: none !important;
}

/* Override Vditor internal styles to match modern clean design */
.vditor-wrapper :deep(.vditor) {
  border: none !important;
  background-color: transparent !important;
}

.vditor-wrapper :deep(.vditor-toolbar) {
  border-bottom: 1px solid rgba(226, 232, 240, 0.8) !important;
  background: #ffffff !important;
  padding: 4px 48px 8px !important; /* 上间距缩小到4px，进一步拉近标题和工具栏 */
  width: 100% !important;
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  align-items: center;
  justify-content: flex-start;
  position: sticky;
  top: 0;
  z-index: 10;
  box-shadow: none !important;
  max-width: 1000px !important;
  margin: 0 auto !important;
  box-sizing: border-box !important;
  border-radius: 0 !important;
  border-left: none !important;
  border-right: none !important;
  border-top: none !important;
}

.vditor-wrapper :deep(.vditor-toolbar__item) {
  margin: 0 !important;
}

.vditor-wrapper :deep(.vditor-toolbar__item > button) {
  color: #64748b !important;
  border-radius: 8px !important;
  padding: 6px !important;
  width: 36px !important;
  height: 36px !important;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  background: transparent !important;
  border: 1px solid transparent !important;
}

.vditor-wrapper :deep(.vditor-toolbar__item > button:hover) {
  background: #ffffff !important;
  color: #2563eb !important;
  box-shadow: 0 2px 6px rgba(0,0,0,0.05) !important;
  transform: translateY(-1px);
}

.vditor-wrapper :deep(.vditor-toolbar__item > button.vditor-menu--current) {
  background: #eff6ff !important;
  color: #2563eb !important;
  border-color: #bfdbfe !important;
  box-shadow: inset 0 2px 4px rgba(0,0,0,0.02) !important;
}

.vditor-wrapper :deep(.vditor-toolbar__divider) {
  height: 20px !important;
  margin: 0 8px !important;
  border-left: 1px solid #cbd5e1 !important;
}

.vditor-wrapper :deep(.vditor-reset) {
  padding: 24px 60px 40px !important; /* 减小正文上方的内边距，让内容更紧凑 */
  width: 100% !important;
  max-width: 1000px !important;
  margin: 0 auto !important;
  box-sizing: border-box !important;
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif !important;
  color: #334155 !important;
  font-size: 16px !important;
  line-height: 1.8 !important;
  background-color: #ffffff !important;
  min-height: calc(100% - 150px) !important;
  border-left: none !important;
  border-right: none !important;
  border-bottom: none !important;
  border-radius: 0 0 8px 8px !important;
  box-shadow: none !important;
  margin-bottom: 40px !important;
}

.vditor-wrapper :deep(.vditor-reset h1),
.vditor-wrapper :deep(.vditor-reset h2),
.vditor-wrapper :deep(.vditor-reset h3),
.vditor-wrapper :deep(.vditor-reset h4) {
  color: #0f172a !important;
  font-weight: 600 !important;
  margin-top: 2em !important;
  margin-bottom: 1em !important;
}

/* 隐藏 vditor 默认的提示框 */
.vditor-wrapper :deep(.vditor-tip) {
  display: none !important;
}

.draft-edit-area, .draft-preview-area {
  display: flex;
  flex-direction: column;
  flex: 1;
  background-color: #ffffff;
}

.create-child-panel {
  max-width: 1000px;
  width: calc(100% - 48px);
  margin: 20px auto 8px;
  padding: 16px 20px;
  box-sizing: border-box;
  border: 1px solid rgba(59, 130, 246, 0.15);
  border-radius: 12px;
  background: #eff6ff;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.05);
}

.create-child-header {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-bottom: 12px;
}

.create-child-title {
  font-size: 14px;
  font-weight: 600;
  color: #1e3a8a;
}

.create-child-hint {
  font-size: 12px;
  color: #64748b;
}

.create-child-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.child-input {
  flex: 1;
  min-width: 0;
  height: 38px;
  padding: 0 14px;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  background: #ffffff;
  font-size: 14px;
  color: #0f172a;
  outline: none;
  transition: all 0.2s ease;
}

.child-input:focus {
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.15);
}

.published-name {
  font-size: 36px;
  font-weight: 800;
  color: #0f172a;
  margin-bottom: 24px;
  letter-spacing: -0.02em;
}

.empty-content {
  color: #94a3b8;
  font-size: 14px;
  font-style: italic;
  margin-top: 16px;
}

.title-input {
  font-size: 36px;
  font-weight: 800;
  border: none;
  outline: none;
  color: #0f172a;
  background: transparent;
  padding: 16px 60px 8px; /* 顶部距离减少使标题整体上移，底部留出8px防止文字下半部分被截断 */
  margin: 0 auto; /* 去除负边距，避免被下方工具栏物理遮挡 */
  width: 100%;
  max-width: 1000px;
  box-sizing: border-box;
  letter-spacing: -0.02em;
  background-color: #ffffff;
  box-shadow: none;
  border-left: none;
  border-right: none;
  border-top: none;
  border-bottom: none;
  border-radius: 8px 8px 0 0;
  line-height: 1.5; /* 恢复正常的行高，防止输入框内部文字被上下裁剪 */
}

.title-input::placeholder {
  color: #cbd5e1;
}

.title-input.readonly {
  color: #64748b;
  cursor: not-allowed;
}

.field-tip {
  padding: 4px 60px 8px;
  font-size: 13px;
  line-height: 1.6;
  color: #64748b;
  max-width: 1000px;
  margin: 0 auto;
  width: 100%;
  box-sizing: border-box;
  background-color: #ffffff;
  box-shadow: none;
  border-left: none;
  border-right: none;
  border-top: none;
  border-bottom: none;
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
