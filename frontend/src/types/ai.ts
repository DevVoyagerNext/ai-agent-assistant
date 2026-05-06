export interface AIChatSession {
  id: number
  title: string
  modelId: string
  updatedAt: string
  createdAt: string
}

export interface AISessionListRes {
  list: AIChatSession[]
  hasMore: boolean
}

export interface AIChatMessage {
  id: number
  sessionId: number
  parentId: number | null
  role: 'user' | 'assistant'
  content: string
  reasoning?: string
  toolLogs?: string[]
  status: 'active' | 'deleted' | 'hidden'
  createdAt: string
}

export interface AIMessageListRes {
  list: AIChatMessage[]
  hasMore: boolean
}

export interface AIChatFile {
  file_url: string
  file_name: string
  file_type: string
  file_size: number
}

export interface AIChatReq {
  skill_id?: string
  user_input: string
  session_id?: string
  file_urls?: string[]
  files?: AIChatFile[] // 新增：支持文件对象数组
  
  // 兼容旧版参数
  prompt?: string
  parentId?: number
  currentPageUrl?: string
  selectedText?: string
  rawFiles?: File[] // 重命名：前端上传文件时原始文件对象
}

export interface AIChatRes {
  reply: string
  sessionId: number
  messageId: number
}

export interface AIUpdateTitleReq {
  title: string
}