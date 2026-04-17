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
  status: 'active' | string
  createdAt: string
  // For frontend UI only
  files?: File[] 
}

export interface AIMessageListRes {
  list: AIChatMessage[]
  hasMore: boolean
}

export interface AIChatReq {
  prompt: string
  sessionId?: number
  parentId?: number
  files?: File[]
}

export interface AIChatRes {
  reply: string
  sessionId: number
  messageId: number
}

export interface AIUpdateTitleReq {
  title: string
}