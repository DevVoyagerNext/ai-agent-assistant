<script setup lang="ts">
import { ref, onMounted, nextTick, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import MarkdownIt from 'markdown-it'
import hljs from 'highlight.js'
import 'highlight.js/styles/github.css'
import { fetchEventSource } from '@microsoft/fetch-event-source'
import { 
  getAISessions, 
  getAISessionMessages, 
  updateAISessionTitle 
} from '../api/ai'
import type { AIChatSession, AIChatMessage } from '../types/ai'
import { 
  Plus, MessageSquare, Send, Paperclip, X, MoreVertical, Edit3, 
  Trash2, Loader2, ArrowLeft, Bot, User, Sparkles
} from 'lucide-vue-next'

const router = useRouter()
const route = useRoute()

const md = new MarkdownIt({
  breaks: true,
  linkify: true,
  highlight: function (str: string, lang: string, _attrs: string) {
    if (lang && hljs.getLanguage(lang)) {
      try {
        return '<pre class="hljs"><code>' + hljs.highlight(str, { language: lang, ignoreIllegals: true }).value + '</code></pre>'
      } catch (__) {}
    }
    return '<pre class="hljs"><code>' + md.utils.escapeHtml(str) + '</code></pre>'
  }
})

const renderMarkdown = (content: string) => {
  if (!content) return ''
  return md.render(content)
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
const selectedFiles = ref<File[]>([])
const isSending = ref(false)
const messagesContainer = ref<HTMLElement | null>(null)

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
})

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
const handleFileSelect = (e: Event) => {
  const target = e.target as HTMLInputElement
  if (!target.files) return
  
  const newFiles = Array.from(target.files)
  if (selectedFiles.value.length + newFiles.length > 3) {
    alert('最多只能上传 3 个文件')
    return
  }
  
  for (const file of newFiles) {
    if (file.size > 5 * 1024 * 1024) {
      alert(`文件 ${file.name} 超过 5MB 限制`)
      continue
    }
    selectedFiles.value.push(file)
  }
  
  // reset input
  target.value = ''
}

const removeFile = (index: number) => {
  selectedFiles.value.splice(index, 1)
}

const sendMessage = async () => {
  const prompt = inputContent.value.trim()
  if (!prompt && selectedFiles.value.length === 0) return
  if (isSending.value) return

  // Optimistic UI
  const tempId = Date.now()
  const parentId = messages.value.length > 0 ? messages.value[messages.value.length - 1].id : 0
  
  const userMsg: AIChatMessage = {
    id: tempId,
    sessionId: currentSessionId.value || 0,
    parentId: parentId || null,
    role: 'user',
    content: prompt,
    status: 'active',
    createdAt: new Date().toISOString(),
    files: [...selectedFiles.value]
  }
  
  messages.value.push(userMsg)
  
  const reqData = new FormData()
  reqData.append('prompt', prompt)
  if (currentSessionId.value) {
    reqData.append('sessionId', currentSessionId.value.toString())
  }
  if (parentId) {
    reqData.append('parentId', parentId.toString())
  }
  if (selectedFiles.value.length > 0) {
    selectedFiles.value.forEach(file => {
      reqData.append('files', file)
    })
  }
  
  inputContent.value = ''
  selectedFiles.value = []
  isSending.value = true
  
  // Create an empty assistant message for streaming
  const assistantMsgId = Date.now() + 1
  const assistantMsg: AIChatMessage = {
    id: assistantMsgId,
    sessionId: currentSessionId.value || 0,
    parentId: tempId,
    role: 'assistant',
    content: '',
    status: 'active',
    createdAt: new Date().toISOString()
  }
  messages.value.push(assistantMsg)
  
  scrollToBottom()

  try {
    const token = localStorage.getItem('token') || ''
    
    await fetchEventSource('http://localhost:8080/v1/ai/chat', {
      method: 'POST',
      headers: {
        'x-token': token
      },
      body: reqData,
      onmessage(ev) {
        if (ev.event === 'meta') {
          try {
            const data = JSON.parse(ev.data)
            assistantMsg.sessionId = data.sessionId
            assistantMsg.id = data.messageId
            
            if (!currentSessionId.value) {
              currentSessionId.value = data.sessionId
              router.replace({ query: { sessionId: data.sessionId } })
              loadSessions(true) // refresh sidebar
            }
          } catch (e) {
            console.error('Failed to parse meta:', e)
          }
        } else if (ev.event === 'message') {
          assistantMsg.content += ev.data
          scrollToBottom()
        } else if (ev.event === 'done') {
          isSending.value = false
        }
      },
      onerror(err) {
        console.error('EventSource error:', err)
        isSending.value = false
        throw err // trigger catch
      },
      onclose() {
        isSending.value = false
      }
    })
  } catch (error: any) {
    console.error('发送失败', error)
    alert('发送失败或连接断开')
  } finally {
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

</script>

<template>
  <div class="ai-chat-layout">
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
          v-for="msg in messages" 
          :key="msg.id"
          class="message-row"
          :class="msg.role"
        >
          <div class="avatar-wrap">
            <Bot v-if="msg.role === 'assistant'" :size="24" class="bot-avatar" />
            <User v-else :size="24" class="user-avatar" />
          </div>
          <div class="message-content-wrap">
            <!-- Render files if any (mainly for user) -->
            <div v-if="msg.files && msg.files.length > 0" class="msg-files">
              <div v-for="(f, i) in msg.files" :key="i" class="msg-file-item">
                <Paperclip :size="12" />
                <span class="file-name">{{ f.name }}</span>
              </div>
            </div>
            
            <div class="message-bubble" :class="{ 'markdown-body': msg.role === 'assistant' }" v-show="msg.content !== ''">
              <div v-if="msg.role === 'assistant'" v-html="renderMarkdown(msg.content)" class="md-content"></div>
              <span v-else class="msg-text">{{ msg.content }}</span>
            </div>
          </div>
        </div>
        
        <div v-if="isSending && messages[messages.length - 1]?.content === ''" class="message-row assistant">
          <div class="avatar-wrap">
            <Bot :size="24" class="bot-avatar" />
          </div>
          <div class="message-content-wrap">
            <div class="message-bubble typing">
              <span class="dot"></span><span class="dot"></span><span class="dot"></span>
            </div>
          </div>
        </div>
      </div>

      <div class="chat-input-area">
        <!-- Selected Files Preview -->
        <div v-if="selectedFiles.length > 0" class="selected-files-preview">
          <div v-for="(file, idx) in selectedFiles" :key="idx" class="preview-file-tag">
            <Paperclip :size="14" />
            <span class="file-name">{{ file.name }}</span>
            <button class="remove-file-btn" @click="removeFile(idx)">
              <X :size="14" />
            </button>
          </div>
        </div>

        <div class="input-box">
          <label class="attach-btn" title="上传文件 (最多3个)">
            <input type="file" multiple @change="handleFileSelect" style="display: none;" />
            <Paperclip :size="20" />
          </label>
          
          <textarea 
            v-model="inputContent"
            class="chat-textarea"
            placeholder="给 AI 助手发送消息..."
            rows="1"
            @keydown="handleTextareaKeydown"
          ></textarea>
          
          <button 
            class="send-btn" 
            :class="{ active: inputContent.trim() || selectedFiles.length > 0 }"
            :disabled="isSending || (!inputContent.trim() && selectedFiles.length === 0)"
            @click="sendMessage"
          >
            <Send :size="18" />
          </button>
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
  transition: background 0.2s;
  position: relative;
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
  display: none;
}

.session-item:hover .session-actions {
  display: flex;
}

.icon-btn-small {
  background: transparent;
  border: none;
  color: #666;
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
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
  max-width: 800px;
  margin: 0 auto;
  width: 100%;
}

.message-row.user {
  flex-direction: row-reverse;
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
  max-width: 80%;
}

.message-row.user .message-content-wrap {
  align-items: flex-end;
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
  background: #f4f4f5;
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
:deep(.md-content pre) {
  background-color: #f6f8fa;
  border-radius: 6px;
  padding: 12px;
  overflow: auto;
  margin: 12px 0;
  font-family: ui-monospace, SFMono-Regular, SF Mono, Menlo, Consolas, Liberation Mono, monospace;
  font-size: 13px;
  line-height: 1.45;
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

.typing {
  display: flex;
  align-items: center;
  gap: 4px;
  height: 24px;
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

.selected-files-preview {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 12px;
  width: 100%;
  max-width: 800px;
}

.preview-file-tag {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 10px;
  background: #f1f5f9;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 13px;
  color: #333;
}

.remove-file-btn {
  background: transparent;
  border: none;
  color: #94a3b8;
  cursor: pointer;
  display: flex;
  align-items: center;
  padding: 2px;
  border-radius: 4px;
}

.remove-file-btn:hover {
  background: #e2e8f0;
  color: #ef4444;
}

.input-box {
  display: flex;
  align-items: flex-end;
  gap: 12px;
  background: #f4f4f5;
  border: 1px solid rgba(0,0,0,0.1);
  border-radius: 16px;
  padding: 8px 16px;
  width: 100%;
  max-width: 800px;
  transition: border-color 0.2s;
}

.input-box:focus-within {
  border-color: #3b82f6;
  background: #ffffff;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}

.attach-btn {
  color: #666;
  cursor: pointer;
  padding: 8px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s;
}

.attach-btn:hover {
  background: rgba(0,0,0,0.05);
}

.chat-textarea {
  flex: 1;
  border: none;
  background: transparent;
  outline: none;
  font-size: 15px;
  color: #333;
  resize: none;
  max-height: 150px;
  padding: 8px 0;
  line-height: 1.5;
  font-family: inherit;
}

.send-btn {
  background: #e4e4e7;
  color: #a1a1aa;
  border: none;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: not-allowed;
  transition: all 0.2s;
  margin-bottom: 4px;
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
}

@media (max-width: 768px) {
  .chat-sidebar {
    display: none; /* Mobile logic: can be toggled via menu */
  }
  .mobile-menu-btn {
    display: block;
  }
}
</style>