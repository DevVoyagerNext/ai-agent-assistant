import request from '../utils/request'
import type { ApiResponse } from '../types'
import type { 
  AISessionListRes, 
  AIMessageListRes, 
  AIChatReq, 
  AIChatRes, 
  AIUpdateTitleReq 
} from '../types/ai'

// 1. AI 聊天接口 (发起对话并可附带文件)
export const sendAIChat = (data: AIChatReq) => {
  const formData = new FormData()
  formData.append('prompt', data.prompt)
  
  if (data.sessionId) {
    formData.append('sessionId', data.sessionId.toString())
  }
  if (data.parentId) {
    formData.append('parentId', data.parentId.toString())
  }

  const currentPageUrl = data.currentPageUrl || (typeof window !== 'undefined' ? window.location.href : '')
  if (currentPageUrl) {
    formData.append('currentPageUrl', currentPageUrl)
  }

  const selectedText = data.selectedText?.trim()
    || (typeof window !== 'undefined' ? window.getSelection()?.toString().trim() || '' : '')
  if (selectedText) {
    formData.append('selectedText', selectedText)
  }
  
  if (data.files && data.files.length > 0) {
    data.files.forEach(file => {
      formData.append('files', file)
    })
  }

  return request.post<ApiResponse<AIChatRes>>('/ai/chat', formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

// 2. 获取用户的历史会话列表 (游标分页)
export const getAISessions = (lastId?: number) => {
  return request.get<ApiResponse<AISessionListRes>>('/ai/sessions', {
    params: {
      lastId
    }
  })
}

// 3. 获取具体会话的消息列表 (游标分页，向上拉取)
export const getAISessionMessages = (sessionId: number, lastId?: number) => {
  return request.get<ApiResponse<AIMessageListRes>>(`/ai/sessions/${sessionId}/messages`, {
    params: {
      lastId
    }
  })
}

// 4. 修改会话标题
export const updateAISessionTitle = (sessionId: number, data: AIUpdateTitleReq) => {
  return request.put<ApiResponse<null>>(`/ai/sessions/${sessionId}/title`, data)
}
