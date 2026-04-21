<script setup lang="ts">
import { ref, reactive, onMounted, onBeforeUnmount, nextTick, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import markdownit from 'markdown-it'
import mathjax3 from 'markdown-it-mathjax3'
import hljs from 'highlight.js'
import 'highlight.js/styles/github.css'
import { 
  getAISessions, 
  getAISessionMessages, 
  updateAISessionTitle 
} from '../api/ai'
import type { AIChatSession, AIChatMessage } from '../types/ai'
import { 
  Plus, MessageSquare, Send, Edit3, 
  Loader2, ArrowLeft, Bot, User, Sparkles, ChevronDown, ChevronUp,
  ArrowUpCircle, ArrowDownCircle, Copy
} from 'lucide-vue-next'

const router = useRouter()
const route = useRoute()

const md = markdownit({
  breaks: true,
  linkify: true,
  highlight: (str: string, lang: string, _attrs: string) => {
    if (lang && hljs.getLanguage(lang)) {
      try {
        return hljs.highlight(str, { language: lang, ignoreIllegals: true }).value
      } catch (__) {}
    }
    return '' // Use default escaping
  }
})

md.use(mathjax3)

// Custom fence rule to wrap in hljs classes, consistent with SubjectDetail.vue
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

const stabilizeMarkdownForStreaming = (content: string) => {
  let processed = content

  // 未闭合的代码块会让 markdown-it 在流式阶段无法正确高亮，先临时补齐。
  const fenceMatches = processed.match(/```/g)
  if (fenceMatches && fenceMatches.length % 2 !== 0) {
    processed += '\n```\n'
  }

  // 补齐未闭合的行内代码，避免整段后续文本都被当成 code span。
  const inlineCodeSegments = processed.replace(/```[\s\S]*?```/g, '')
  const backtickMatches = inlineCodeSegments.match(/`/g)
  if (backtickMatches && backtickMatches.length % 2 !== 0) {
    processed += '`'
  }

  // 补齐未闭合的加粗标记，提升流式阶段的可读性。
  const boldMatches = processed.match(/\*\*/g)
  if (boldMatches && boldMatches.length % 2 !== 0) {
    processed += '**'
  }

  return processed
}

const normalizePlainTextFenceContent = (content: string) => {
  return content
    .replace(/\\r\\n/g, '\n')
    .replace(/\\n/g, '\n')
    .replace(/\\r/g, '\n')
    .replace(/\\t/g, '\t')
}

const renderMarkdown = (content: string, isStreaming = false) => {
  if (!content) return ''
  
  let processed = content

  // 纯文本代码块中的转义换行需要还原为真实换行，避免 \n 被直接显示出来。
  processed = processed.replace(/```([^\n`]*)\n([\s\S]*?)```/g, (match, info, code) => {
    const lang = info.trim().toLowerCase()
    if (!lang || ['text', 'plain', 'plaintext', 'txt'].includes(lang)) {
      return `\`\`\`${info}\n${normalizePlainTextFenceContent(code)}\`\`\``
    }
    return match
  })
  
  // 1. 支持 ```math 和 ```latex 代码块自动转换为 LaTeX 块级公式
  processed = processed.replace(/```(?:math|latex)\n([\s\S]*?)```/g, '$$\n$1\n$$')
  
  // 2. 剥离行内代码块中的数学公式 (例如 `$u_1$` -> $u_1$)
  processed = processed.replace(/`(\$[^`]+?\$)`/g, '$1')
  
  // 3. 针对常见的算法题输入格式（如果没有被正确包裹）进行智能转换
  // 检测紧跟在【输入格式】或输入格式等字眼后面的纯文本代码块
  processed = processed.replace(/(输入格式.*?)\n+```(?:text)?\n([\s\S]*?)```/g, (match, prefix, code) => {
    if (/[a-zA-Z]_[a-zA-Z0-9{}]/.test(code) && !/[;=")(]/.test(code)) {
      const matrix = code.trim().split('\n').map((line: string) => line.trim().replace(/\s+/g, ' & ')).join(' \\\\\n')
      return `${prefix}\n$$\n\\begin{matrix}\n${matrix}\n\\end{matrix}\n$$`
    }
    return match
  })

  // 4. 替换非代码块中可能遗漏的下标情况 (仅在确认为公式结构时，但这较危险，暂不全局替换)
  if (isStreaming) {
    processed = stabilizeMarkdownForStreaming(processed)
  }

  return md.render(processed)
}

const isStreamingAssistantMessage = (msg: AIChatMessage) => {
  if (!isSending.value || msg.role !== 'assistant') return false
  const lastMessage = messages.value[messages.value.length - 1]
  return lastMessage?.id === msg.id
}

// ===== Sidebar / Sessions State =====
const sessions = ref<AIChatSession[]>([])
const hasMoreSessions = ref(true)
const loadingSessions = ref(false)
const currentSessionId = ref<number | null>(null)

// ===== Chat Area State =====
const messages = ref<AIChatMessage[]>([])
const hasMoreMessages = ref(true)
const loadingMessages = ref(false)
const inputContent = ref('')
const isSending = ref(false)
const messagesContainer = ref<HTMLElement | null>(null)
let activeChatAbortController: AbortController | null = null

// ===== Reasoning Collapse State =====
// 记录哪些消息的深度思考区域是被折叠的
// key: messageId, value: boolean (true = collapsed)
const reasoningCollapsedMap = ref<Record<number, boolean>>({})

// 切换折叠状态
const toggleReasoning = (msgId: number) => {
  reasoningCollapsedMap.value[msgId] = !reasoningCollapsedMap.value[msgId]
}

// 检查某个思考过程是否应折叠
const isReasoningCollapsed = (msg: AIChatMessage) => {
  const isStreaming = isStreamingAssistantMessage(msg)
  
  // 用户已经手动改变过状态的话，以用户的状态为准
  if (reasoningCollapsedMap.value[msg.id] !== undefined) {
    return reasoningCollapsedMap.value[msg.id]
  }
  
  // 默认策略：
  // 1. 如果正在流式思考中且字数超过 200，自动收起（折叠）
  // 2. 如果思考已经结束（开始输出 content 或已经完毕），自动展开（不折叠）
  if (isStreaming && !msg.content && msg.reasoning && msg.reasoning.length > 200) {
    return true
  }
  
  return false
}

// ===== Rename State =====
const renamingSessionId = ref<number | null>(null)
const renameTitle = ref('')
const isRenaming = ref(false)

// ===== Initialization =====
onMounted(async () => {
  await loadSessions(true)
  if (route.query.sessionId) {
    const sId = parseInt(route.query.sessionId as string)
    if (!isNaN(sId)) {
      await selectSession(sId)
    }
  }

  // Add global event listener for copying code blocks
  messagesContainer.value?.addEventListener('click', handleCodeCopy)
})

onBeforeUnmount(() => {
  if (activeChatAbortController) {
    activeChatAbortController.abort()
  }
  messagesContainer.value?.removeEventListener('click', handleCodeCopy)
})

// ===== Toast Logic =====
const toastMsg = ref('')
const showToastTimer = ref<any>(null)
const showToast = (msg: string) => {
  toastMsg.value = msg
  if (showToastTimer.value) clearTimeout(showToastTimer.value)
  showToastTimer.value = setTimeout(() => {
    toastMsg.value = ''
  }, 2000)
}

// ===== Copy Code Block =====
const handleCodeCopy = async (e: MouseEvent) => {
  const btn = (e.target as HTMLElement).closest('.code-copy-btn') as HTMLElement
  if (btn) {
    const rawCode = btn.getAttribute('data-code') || ''
    // Decode HTML entities
    const textarea = document.createElement('textarea')
    textarea.innerHTML = rawCode
    const decodedCode = textarea.value
    
    try {
      await navigator.clipboard.writeText(decodedCode)
      btn.classList.add('copied')
      showToast('代码已复制到剪贴板')
      setTimeout(() => {
        btn.classList.remove('copied')
      }, 2000)
    } catch (err) {
      console.error('Failed to copy code:', err)
    }
  }
}

// ===== Session Logic =====
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

const handleScrollSessions = (e: Event) => {
  const target = e.target as HTMLElement
  if (target.scrollTop + target.clientHeight >= target.scrollHeight - 20) {
    loadSessions()
  }
}

const startNewChat = () => {
  currentSessionId.value = null
  messages.value = []
  hasMoreMessages.value = false
  router.replace('/ai-chat')
}

const selectSession = async (sessionId: number) => {
  if (currentSessionId.value === sessionId) return
  currentSessionId.value = sessionId
  router.replace({ query: { sessionId } })
  await loadMessages(sessionId, true)
  scrollToBottom()
}

// ===== Rename Logic =====
const openRename = (session: AIChatSession) => {
  renamingSessionId.value = session.id
  renameTitle.value = session.title
}

const cancelRename = () => {
  renamingSessionId.value = null
  renameTitle.value = ''
}

const confirmRename = async (session: AIChatSession) => {
  if (!renameTitle.value.trim() || renameTitle.value === session.title) {
    cancelRename()
    return
  }
  isRenaming.value = true
  try {
    const res = await updateAISessionTitle(session.id, { title: renameTitle.value })
    if (res.data?.code === 200) {
      session.title = renameTitle.value
    }
  } catch (error) {
    console.error('重命名失败', error)
  } finally {
    isRenaming.value = false
    cancelRename()
  }
}

// ===== Messages Logic =====
const loadMessages = async (sessionId: number, reset = false) => {
  if (loadingMessages.value || (!hasMoreMessages.value && !reset)) return
  loadingMessages.value = true
  
  const oldScrollHeight = messagesContainer.value?.scrollHeight || 0
  
  try {
    const lastId = reset ? undefined : messages.value[0]?.id
    const res = await getAISessionMessages(sessionId, lastId)
    if (res.data?.code === 200 && res.data.data) {
      const list = res.data.data.list || []
      if (reset) {
        messages.value = list
      } else {
        messages.value.unshift(...list)
      }
      hasMoreMessages.value = res.data.data.hasMore
      
      // Restore scroll position after loading more
      if (!reset) {
        nextTick(() => {
          if (messagesContainer.value) {
            const newScrollHeight = messagesContainer.value.scrollHeight
            messagesContainer.value.scrollTop = newScrollHeight - oldScrollHeight
          }
        })
      }
    }
  } catch (error) {
    console.error('加载消息失败', error)
  } finally {
    loadingMessages.value = false
  }
}

const handleScrollMessages = (e: Event) => {
  const target = e.target as HTMLElement
  if (target.scrollTop <= 20 && currentSessionId.value) {
    loadMessages(currentSessionId.value)
  }
}

const scrollToBottom = () => {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
}

// ===== Chat Input Logic =====
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

  // gin 在 SSE 中可能把字符串序列化为 JSON string，这里统一还原。
  if (trimmed.startsWith('"') && trimmed.endsWith('"')) {
    const parsed = parseSSEJson<string>(trimmed)
    if (typeof parsed === 'string') {
      trimmed = parsed
    }
  }

  // 后端现已将消息内容改为 Base64 编码传输，在此处进行解码
  try {
    // 浏览器 atob 解码 Base64 后得到的是 binary string，需要转换回正确的 UTF-8 字符
    const binaryString = atob(trimmed)
    const bytes = new Uint8Array(binaryString.length)
    for (let i = 0; i < binaryString.length; i++) {
      bytes[i] = binaryString.charCodeAt(i)
    }
    return new TextDecoder('utf-8').decode(bytes)
  } catch (error) {
    // 如果不是合法的 base64，或者解码失败，降级返回原字符串
    return trimmed
  }
}

type SSEEvent = {
  event: string
  data: string
}

const extractSSEEvents = (buffer: string) => {
  const events: SSEEvent[] = []
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

    events.push({
      event: eventName,
      data: dataLines.join('\n')
    })
  }

  return { events, rest }
}

const sendMessage = async () => {
  const prompt = inputContent.value.trim()
  if (!prompt) return
  if (isSending.value) return

  // Optimistic UI
  const tempId = Date.now()
  const parentId = messages.value.length > 0 ? messages.value[messages.value.length - 1].id : 0
  
  const userMsg = reactive<AIChatMessage>({
    id: tempId,
    sessionId: currentSessionId.value || 0,
    parentId: parentId || null,
    role: 'user',
    content: prompt,
    status: 'active',
    createdAt: new Date().toISOString()
  })
  
  messages.value.push(userMsg)
  
  const reqData = new FormData()
  reqData.append('prompt', prompt)
  if (currentSessionId.value) {
    reqData.append('sessionId', currentSessionId.value.toString())
  }
  if (parentId) {
    reqData.append('parentId', parentId.toString())
  }
  
  inputContent.value = ''
  isSending.value = true
  nextTick(() => adjustTextareaHeight())
  
  // Create an empty assistant message for streaming
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
  messages.value.push(assistantMsg)
  
  scrollToBottom()

  activeChatAbortController?.abort()
  const abortController = new AbortController()
  activeChatAbortController = abortController
  let streamFinished = false

  try {
    const token = localStorage.getItem('token') || ''

    const response = await fetch('http://localhost:8080/v1/ai/chat', {
      method: 'POST',
      headers: {
        'x-token': token
      },
      body: reqData,
      signal: abortController.signal
    })

    const contentType = response.headers.get('content-type') || ''
    if (!response.ok) {
      throw new Error(`请求失败: ${response.status}`)
    }
    if (!contentType.includes('text/event-stream')) {
      throw new Error(`接口未返回流式内容: ${contentType || 'unknown'}`)
    }
    if (!response.body) {
      throw new Error('流式响应体为空')
    }

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
          if (!data) {
            continue
          }

          assistantMsg.sessionId = data.sessionId
          assistantMsg.id = data.messageId
          userMsg.sessionId = data.sessionId
          currentSessionId.value = data.sessionId

          if (route.query.sessionId !== String(data.sessionId)) {
            router.replace({ query: { sessionId: data.sessionId } })
          }
          void loadSessions(true)
        } else if (event.event === 'message') {
          assistantMsg.content += normalizeMessageChunk(event.data)
          scrollToBottom()
        } else if (event.event === 'reasoning') {
          assistantMsg.reasoning = (assistantMsg.reasoning || '') + normalizeMessageChunk(event.data)
          scrollToBottom()
        } else if (event.event === 'done') {
          streamFinished = true
          isSending.value = false
          abortController.abort()
          return
        }
      }
    }

    if (!streamFinished && !abortController.signal.aborted) {
      throw new Error('流式连接意外关闭')
    }
  } catch (error: any) {
    if (!abortController.signal.aborted || !streamFinished) {
      console.error('发送失败', error)
      alert('发送失败或连接断开')
    }
  } finally {
    if (activeChatAbortController === abortController) {
      activeChatAbortController = null
    }
    isSending.value = false
  }
}

const handleTextareaKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    sendMessage()
  }
}

const currentSessionTitle = computed(() => {
  if (!currentSessionId.value) return '新对话'
  const s = sessions.value.find(s => s.id === currentSessionId.value)
  return s ? s.title : '对话'
})

// ===== Context Navigation =====
const scrollToNextUserMsg = (direction: 'up' | 'down') => {
  const container = messagesContainer.value
  if (!container) return
  
  const userMsgs = messages.value.filter(m => m.role === 'user')
  if (userMsgs.length === 0) return

  let closestIdx = 0
  let minDiff = Infinity

  userMsgs.forEach((msg, idx) => {
    const el = document.getElementById(`msg-${msg.id}`)
    if (el) {
      // 找到离当前视图中心最近的用户消息
      const diff = Math.abs(el.offsetTop - container.scrollTop - container.clientHeight / 3)
      if (diff < minDiff) {
        minDiff = diff
        closestIdx = idx
      }
    }
  })

  // 如果点击的是“向下”且当前已经是最后一条用户消息，则跳转到最底部
  if (direction === 'down' && closestIdx === userMsgs.length - 1) {
    scrollToBottom()
    return
  }

  let targetIdx = closestIdx
  if (direction === 'up') {
    targetIdx = Math.max(0, closestIdx - 1)
  } else {
    targetIdx = Math.min(userMsgs.length - 1, closestIdx + 1)
  }

  const targetMsgId = userMsgs[targetIdx].id
  const targetEl = document.getElementById(`msg-${targetMsgId}`)
  if (targetEl) {
    targetEl.scrollIntoView({ behavior: 'smooth', block: 'center' })
  }
}

// ===== Copy Text =====
const copyText = async (text: string) => {
  if (!text) return
  try {
    await navigator.clipboard.writeText(text)
    showToast('内容已复制到剪贴板')
  } catch (err) {
    console.error('Failed to copy', err)
  }
}

// ===== Textarea Auto-resize =====
const chatInputRef = ref<HTMLTextAreaElement | null>(null)
const adjustTextareaHeight = () => {
  const el = chatInputRef.value
  if (!el) return
  if (!inputContent.value) {
    el.style.height = 'auto'
    return
  }
  el.style.height = 'auto'
  el.style.height = Math.min(el.scrollHeight, 150) + 'px'
}

</script>

<template>
  <div class="ai-chat-layout">
    <!-- Global Toast Notification -->
    <Transition name="toast">
      <div v-if="toastMsg" class="toast-notification">
        {{ toastMsg }}
      </div>
    </Transition>

    <!-- Sidebar -->
    <aside class="chat-sidebar">
      <div class="sidebar-header">
        <button class="back-home-btn" @click="router.push('/')">
          <ArrowLeft :size="20" />
        </button>
        <button class="new-chat-btn" @click="startNewChat">
          <Plus :size="16" />
          <span>新对话</span>
        </button>
      </div>

      <div class="session-list" @scroll="handleScrollSessions">
        <div 
          v-for="session in sessions" 
          :key="session.id"
          class="session-item"
          :class="{ active: currentSessionId === session.id }"
          @click="selectSession(session.id)"
        >
          <MessageSquare :size="16" class="session-icon" />
          
          <div v-if="renamingSessionId === session.id" class="rename-input-wrap" @click.stop>
            <input 
              v-model="renameTitle" 
              class="rename-input"
              autoFocus
              @keyup.enter="confirmRename(session)"
              @keyup.esc="cancelRename"
              @blur="confirmRename(session)"
            />
          </div>
          <span v-else class="session-title">{{ session.title }}</span>

          <div class="session-actions" v-if="renamingSessionId !== session.id">
            <button class="icon-btn-small" @click.stop="openRename(session)">
              <Edit3 :size="14" />
            </button>
          </div>
        </div>
        
        <div v-if="loadingSessions" class="loading-more">
          <Loader2 class="spin" :size="16" />
        </div>
      </div>
    </aside>

    <!-- Main Chat Area -->
    <main class="chat-main">
      <header class="chat-header">
        <div class="mobile-menu-btn" @click="router.push('/')">
          <ArrowLeft :size="20" />
        </div>
        <h2 class="current-title">{{ currentSessionTitle }}</h2>
        <div class="header-right-actions">
          <button class="user-profile-btn" @click="router.push('/profile')" title="个人信息">
            <User :size="18" />
          </button>
        </div>
      </header>

      <div class="messages-container" ref="messagesContainer" @scroll="handleScrollMessages">
        <div v-if="loadingMessages" class="loading-more">
          <Loader2 class="spin" :size="20" /> 加载历史消息...
        </div>

        <div v-if="messages.length === 0 && !loadingMessages && !currentSessionId" class="empty-chat">
          <Sparkles class="sparkle-icon" :size="48" />
          <h3>今天想聊点什么？</h3>
          <p>我可以帮你解答编程问题、分析代码、或者讨论前沿技术。</p>
        </div>

        <div 
          v-for="(msg, index) in messages" 
          :key="msg.id"
          :id="'msg-' + msg.id"
          class="message-row"
          :class="[msg.role, { 'mt-extra': msg.role === 'user' && index > 0 && messages[index-1].role === 'assistant' }]"
        >
          <div class="avatar-wrap" v-if="msg.role === 'assistant'">
            <Bot :size="24" class="bot-avatar" />
          </div>
          <div class="message-content-wrap">
            <!-- Render files if any (mainly for user) -->
            
            <div class="message-bubble-container">
              <div class="message-bubble" :class="{ 'markdown-body': msg.role === 'assistant' }">
                
                <!-- Reasoning Block (only show if it exists) -->
                <div 
                  v-if="msg.role === 'assistant' && msg.reasoning" 
                  class="reasoning-block"
                >
                  <div class="reasoning-summary" @click="toggleReasoning(msg.id)">
                    <Loader2 v-if="isStreamingAssistantMessage(msg) && !msg.content" class="spin icon-sm" :size="14" />
                    <Sparkles v-else class="icon-sm" :size="14" />
                    <span>深度思考过程</span>
                    <ChevronDown v-if="isReasoningCollapsed(msg)" class="icon-sm chevron" :size="14" />
                    <ChevronUp v-else class="icon-sm chevron" :size="14" />
                  </div>
                  
                  <div v-show="!isReasoningCollapsed(msg)" class="reasoning-content md-content" v-html="renderMarkdown(msg.reasoning, isStreamingAssistantMessage(msg) && !msg.content)"></div>
                  
                  <!-- 底部收起按钮，只有当内容展开且内容足够长时才显示 -->
                  <div v-show="!isReasoningCollapsed(msg) && msg.reasoning.length > 200" class="reasoning-footer" @click="toggleReasoning(msg.id)">
                    <span>收起思考过程</span>
                    <ChevronUp class="icon-sm chevron" :size="14" />
                  </div>
                </div>

                <!-- Typing Indicator for Assistant when content is empty and no reasoning yet -->
                <div v-if="msg.role === 'assistant' && msg.content === '' && !msg.reasoning && isSending" class="typing">
                  <span class="dot"></span><span class="dot"></span><span class="dot"></span>
                </div>
                
                <!-- Normal Content -->
                <div
                  v-if="msg.role === 'assistant' && msg.content !== ''"
                  v-html="renderMarkdown(msg.content, isStreamingAssistantMessage(msg))"
                  class="md-content"
                ></div>
                
                <!-- User Content -->
                <span v-if="msg.role === 'user'" class="msg-text">{{ msg.content }}</span>
              </div>

              <!-- Copy Action -->
              <button class="copy-btn" v-if="msg.content" @click="copyText(msg.content)" title="复制">
                <Copy :size="16" />
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Global Context Navigation Buttons -->
      <div class="global-nav-actions" v-if="messages.some(m => m.role === 'user')">
        <button class="nav-btn" @click="scrollToNextUserMsg('up')" title="上一条我的消息">
          <ArrowUpCircle :size="24" />
        </button>
        <button class="nav-btn" @click="scrollToNextUserMsg('down')" title="下一条我的消息">
          <ArrowDownCircle :size="24" />
        </button>
      </div>

      <div class="chat-input-area">
        <div class="input-box">
          <textarea 
            ref="chatInputRef"
            v-model="inputContent"
            class="chat-textarea"
            placeholder="给 AI 助手发送消息..."
            rows="2"
            @keydown="handleTextareaKeydown"
            @input="adjustTextareaHeight"
          ></textarea>
          
          <div class="input-actions">
            <!-- Occupy left space to keep right-aligned flex layout looking balanced or simply justify-content: flex-end -->
            <div class="spacer"></div>
            
            <button 
              class="send-btn" 
              :class="{ active: inputContent.trim() && !isSending }"
              :disabled="isSending || !inputContent.trim()"
              @click="sendMessage"
            >
              <Send :size="18" />
            </button>
          </div>
        </div>
        <p class="input-tip">AI 助手可能会生成不准确的信息，请注意核实。Shift + Enter 换行</p>
      </div>
    </main>
  </div>
</template>

<style scoped>
.ai-chat-layout {
  display: flex;
  height: 100vh;
  width: 100vw;
  background: #ffffff;
  font-family: system-ui, -apple-system, sans-serif;
  overflow: hidden;
}

/* ===== Sidebar ===== */
.chat-sidebar {
  width: 260px;
  background: #f9f9f9;
  border-right: 1px solid rgba(0,0,0,0.08);
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}

.sidebar-header {
  padding: 16px;
  display: flex;
  align-items: center;
  gap: 12px;
}

.back-home-btn {
  background: transparent;
  border: none;
  color: #666;
  cursor: pointer;
  padding: 8px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.back-home-btn:hover {
  background: rgba(0,0,0,0.05);
}

.new-chat-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  background: #ffffff;
  border: 1px solid rgba(0,0,0,0.1);
  padding: 8px 12px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  color: #333;
  cursor: pointer;
  transition: all 0.2s;
}

.new-chat-btn:hover {
  background: #f0f0f0;
}

.session-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.session-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 8px;
  cursor: pointer;
  color: #444;
  transition: all 0.2s ease;
  position: relative;
  overflow: hidden; /* 防止 hover 导致内容溢出 */
}

.session-item:hover {
  background: rgba(0,0,0,0.05);
}

.session-item.active {
  background: #e3e5e8;
  font-weight: 500;
}

.session-icon {
  flex-shrink: 0;
  color: #666;
}

.session-title {
  flex: 1;
  font-size: 14px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  transition: padding-right 0.2s ease; /* 平滑收缩右侧空间 */
}

.rename-input-wrap {
  flex: 1;
}

.rename-input {
  width: 100%;
  border: 1px solid #ccc;
  border-radius: 4px;
  padding: 2px 4px;
  font-size: 14px;
  outline: none;
}

.session-actions {
  position: absolute;
  right: 8px;
  display: flex;
  align-items: center;
  background: linear-gradient(90deg, transparent, #f9f9f9 20%);
  padding-left: 12px;
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.2s ease;
}

.session-item:hover .session-actions {
  opacity: 1;
  pointer-events: auto;
}

.session-item:hover .session-title {
  padding-right: 28px; /* 为按钮腾出空间 */
}

.session-item.active:hover .session-actions {
  background: linear-gradient(90deg, transparent, #e3e5e8 20%);
}

.icon-btn-small {
  background: transparent;
  border: none;
  color: #666;
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  transition: background 0.2s, color 0.2s;
}

.icon-btn-small:hover {
  background: rgba(0,0,0,0.1);
  color: #333;
}

.loading-more {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 12px;
  color: #999;
  font-size: 12px;
  gap: 8px;
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  100% { transform: rotate(360deg); }
}

/* ===== Main Chat Area ===== */
.chat-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: #ffffff;
  position: relative;
}

.chat-header {
  height: 60px;
  border-bottom: 1px solid rgba(0,0,0,0.05);
  display: flex;
  align-items: center;
  padding: 0 24px;
}

.header-right-actions {
  margin-left: auto;
  display: flex;
  align-items: center;
}

.user-profile-btn {
  background: #f0f0f0;
  border: 1px solid #e2e8f0;
  color: #64748b;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s;
}

.user-profile-btn:hover {
  background: #e2e8f0;
  color: #3b82f6;
}

.mobile-menu-btn {
  display: none;
  margin-right: 16px;
  cursor: pointer;
  color: #666;
}

.current-title {
  font-size: 16px;
  font-weight: 600;
  color: #333;
  margin: 0;
}

.messages-container {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 24px;
  scroll-behavior: smooth;
}

.empty-chat {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #666;
  text-align: center;
  max-width: 400px;
  margin: 0 auto;
  transform: translateX(-55px);
}

.sparkle-icon {
  color: #3b82f6;
  margin-bottom: 16px;
}

.empty-chat h3 {
  font-size: 24px;
  color: #333;
  margin-bottom: 8px;
}

.message-row {
  display: flex;
  gap: 16px;
  max-width: 850px;
  margin: 0 auto;
  width: 100%;
}

.message-row.user {
  flex-direction: row-reverse;
  padding-right: 120px;
  box-sizing: border-box;
}

.message-row.user.mt-extra {
  margin-top: 32px;
}

.avatar-wrap {
  flex-shrink: 0;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f0f0f0;
}

.bot-avatar {
  color: #3b82f6;
}

.user-avatar {
  color: #666;
}

.message-content-wrap {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-width: 85%;
  position: relative;
}

.message-row.user .message-content-wrap {
  align-items: flex-end;
}

.message-bubble-container {
  display: flex;
  align-items: flex-end;
  gap: 8px;
}

.message-row.user .message-bubble-container {
  flex-direction: row-reverse;
}

.copy-btn {
  opacity: 0;
  background: transparent;
  border: none;
  color: #94a3b8;
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  transition: all 0.2s;
  flex-shrink: 0;
}

.message-row:hover .copy-btn {
  opacity: 1;
}

.copy-btn:hover {
  color: #3b82f6;
  background: #f1f5f9;
}

.global-nav-actions {
  position: absolute;
  right: 24px;
  bottom: 120px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  opacity: 0.2;
  transition: opacity 0.3s;
  z-index: 10;
}

.global-nav-actions:hover {
  opacity: 1;
}

.nav-btn {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 50%;
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #64748b;
  cursor: pointer;
  box-shadow: 0 2px 6px rgba(0,0,0,0.05);
  transition: all 0.2s;
}

.nav-btn:hover {
  color: #3b82f6;
  border-color: #bfdbfe;
  background: #eff6ff;
}

.msg-files {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.msg-file-item {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  background: #f1f5f9;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  font-size: 12px;
  color: #475569;
}

.message-bubble {
  padding: 12px 16px;
  border-radius: 12px;
  font-size: 15px;
  line-height: 1.6;
  color: #333;
  white-space: pre-wrap;
  word-break: break-word;
  max-width: 100%;
  overflow-x: auto;
}

.message-bubble.markdown-body {
  white-space: normal;
}

/* Markdown 样式 */
:deep(.md-content p) {
  margin: 0 0 12px 0;
}
:deep(.md-content p:last-child) {
  margin-bottom: 0;
}
:deep(.code-block-wrapper) {
  position: relative;
  margin: 16px 0;
  border-radius: 12px;
  overflow: hidden;
  background-color: #f4f5f9;
  border: none;
}

:deep(.code-block-header) {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 16px;
  background-color: #f4f5f9;
  color: #333;
  font-size: 13px;
  font-family: system-ui, -apple-system, sans-serif;
  border-bottom: 3px solid #ffffff;
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
  transition: all 0.2s;
}

:deep(.code-copy-btn:hover) {
  background-color: rgba(0, 0, 0, 0.05);
  color: #1e293b;
}

:deep(.code-copy-btn.copied) {
  color: #10b981;
}

:deep(.md-content pre) {
  margin: 0;
  padding: 16px;
  overflow: auto;
  font-family: ui-monospace, SFMono-Regular, SF Mono, Menlo, Consolas, Liberation Mono, monospace;
  font-size: 13px;
  line-height: 1.5;
  background: transparent;
}
:deep(.md-content code) {
  background-color: rgba(175, 184, 193, 0.2);
  padding: 0.2em 0.4em;
  border-radius: 6px;
  font-family: ui-monospace, SFMono-Regular, SF Mono, Menlo, Consolas, Liberation Mono, monospace;
  font-size: 13px;
}
:deep(.md-content pre code) {
  background-color: transparent;
  padding: 0;
  border-radius: 0;
}
:deep(.md-content ul), :deep(.md-content ol) {
  padding-left: 24px;
  margin-top: 0;
  margin-bottom: 12px;
}
:deep(.md-content li) {
  margin-bottom: 4px;
}
:deep(.md-content h1), :deep(.md-content h2), :deep(.md-content h3), :deep(.md-content h4), :deep(.md-content h5), :deep(.md-content h6) {
  margin-top: 24px;
  margin-bottom: 12px;
  font-weight: 600;
  line-height: 1.25;
}
:deep(.md-content table) {
  border-spacing: 0;
  border-collapse: collapse;
  margin-bottom: 12px;
  width: 100%;
}
:deep(.md-content table th), :deep(.md-content table td) {
  padding: 6px 13px;
  border: 1px solid #d0d7de;
}
:deep(.md-content table tr:nth-child(2n)) {
  background-color: #f6f8fa;
}
:deep(.md-content a) {
  color: #0969da;
  text-decoration: none;
}
:deep(.md-content a:hover) {
  text-decoration: underline;
}

.message-row.user .message-bubble {
  background: #3b82f6;
  color: white;
  border-bottom-right-radius: 4px;
}

.message-row.assistant .message-bubble {
  border-bottom-left-radius: 4px;
}

/* Reasoning Block Styles */
.reasoning-block {
  margin-bottom: 12px;
  background-color: rgba(0, 0, 0, 0.03);
  border-radius: 8px;
  border: 1px solid rgba(0, 0, 0, 0.05);
}

.reasoning-summary {
  cursor: pointer;
  padding: 8px 12px;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: #666;
  user-select: none;
  font-weight: 500;
  outline: none;
}

.reasoning-summary::-webkit-details-marker {
  display: none;
}

.reasoning-summary:hover {
  color: #333;
}

.chevron {
  margin-left: auto;
  opacity: 0.6;
}

.reasoning-content {
  padding: 0 12px 12px 12px;
  font-size: 14px;
  color: #555;
  border-top: 1px dashed rgba(0, 0, 0, 0.05);
  margin-top: 4px;
  padding-top: 8px;
}

.reasoning-footer {
  cursor: pointer;
  padding: 8px 12px;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 4px;
  font-size: 12px;
  color: #666;
  border-top: 1px dashed rgba(0, 0, 0, 0.05);
  user-select: none;
}

.reasoning-footer:hover {
  color: #333;
  background-color: rgba(0, 0, 0, 0.02);
  border-bottom-left-radius: 8px;
  border-bottom-right-radius: 8px;
}

.icon-sm {
  flex-shrink: 0;
}

.typing {
  display: flex;
  align-items: center;
  gap: 4px;
  height: 24px;
  overflow: hidden; /* 防止点溢出导致滚动条 */
  padding: 0 4px;
}

.dot {
  width: 6px;
  height: 6px;
  background: #999;
  border-radius: 50%;
  animation: typing 1.4s infinite ease-in-out;
}

.dot:nth-child(1) { animation-delay: -0.32s; }
.dot:nth-child(2) { animation-delay: -0.16s; }

@keyframes typing {
  0%, 80%, 100% { transform: scale(0); }
  40% { transform: scale(1); }
}

/* ===== Input Area ===== */
.chat-input-area {
  padding: 16px 24px 24px;
  background: #ffffff;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.input-box {
  display: flex;
  flex-direction: column;
  gap: 8px;
  background: #ffffff;
  border: 1px solid rgba(0,0,0,0.05);
  border-radius: 30px;
  padding: 15px 16px 12px 20px;
  width: calc(100% - 98px);
  max-width: 752px;
  margin-right: 98px;
  box-sizing: border-box;
  box-shadow: 0 4px 10px rgba(0, 0, 0, 0.08);
  transition: border-color 0.2s, box-shadow 0.2s;
}

.input-box:focus-within {
  border-color: #3b82f6;
  box-shadow: 0 4px 20px rgba(59, 130, 246, 0.15);
}

.input-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.chat-textarea {
  width: 100%;
  border: none;
  background: transparent;
  outline: none;
  font-size: 15px;
  color: #333;
  resize: none;
  min-height: 48px;
  max-height: 150px;
  padding: 4px 0;
  line-height: 1.5;
  font-family: inherit;
}

.send-btn {
  background: #e4e4e7;
  color: #3a3a3d;
  border: none;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: not-allowed;
  transition: all 0.2s;
}

.send-btn.active {
  background: #3b82f6;
  color: #ffffff;
  cursor: pointer;
}

.send-btn.active:hover {
  background: #2563eb;
}

.input-tip {
  font-size: 12px;
  color: #999;
  margin-top: 12px;
  text-align: center;
  width: calc(100% - 98px);
  max-width: 752px;
  margin-right: 98px;
}

@media (max-width: 768px) {
  .chat-sidebar {
    display: none; /* Mobile logic: can be toggled via menu */
  }
  .mobile-menu-btn {
    display: block;
  }
}

/* ===== Toast Notification ===== */
.toast-notification {
  position: fixed;
  top: 20px;
  left: 50%;
  transform: translateX(-50%);
  background: #333;
  color: white;
  padding: 8px 16px;
  border-radius: 8px;
  font-size: 14px;
  z-index: 9999;
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
}

.toast-enter-active, .toast-leave-active {
  transition: all 0.3s ease;
}
.toast-enter-from, .toast-leave-to {
  opacity: 0;
  transform: translate(-50%, -20px);
}
</style>
