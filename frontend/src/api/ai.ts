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
  
  // 新版字段
  formData.append('user_input', data.user_input)
  if (data.skill_id) {
    formData.append('skill_id', data.skill_id)
  }
  if (data.session_id) {
    formData.append('session_id', data.session_id)
  }

  // 兼容旧版参数
  if (data.prompt) {
    formData.append('prompt', data.prompt)
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
  
  // 原始文件上传
  if (data.rawFiles && data.rawFiles.length > 0) {
    data.rawFiles.forEach(file => {
      formData.append('files', file)
    })
  }

  // 已经上传后的文件信息对象
  if (data.files && data.files.length > 0) {
    formData.append('files_info', JSON.stringify(data.files))
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
