import request from '../utils/request'
import type { ApiResponse } from '../types/index'
import type { SubjectNode, SubjectNodeDetail, NodeNote, AuthorInitRes, AuthorNode, AuthorNodeContent } from '../types/node'
import { validateNoteContent } from '../utils/noteValidation'

// 获取教材顶级知识点
export const getTopNodes = (subjectId: number) => {
  return request.get<ApiResponse<SubjectNode[]>>('/nodes/top', {
    params: { subjectId }
  })
}

// 获取某个知识点的直属子知识点
export const getChildNodes = (nodeId: number) => {
  return request.get<ApiResponse<SubjectNode[]>>(`/nodes/${nodeId}/children`)
}

// 获取知识点详情（含正文/层级等）
export const getNodeDetail = (nodeId: number) => {
  return request.get<ApiResponse<SubjectNodeDetail>>(`/nodes/${nodeId}/detail`)
}

// 获取知识点层级路径 (用于定位节点)
export const getNodePath = (nodeId: number) => {
  return request.get<ApiResponse<SubjectNode[]>>('/nodes/path', {
    params: { nodeId }
  })
}

// 获取知识点随堂笔记（必须登录）
export const getNodeNote = (nodeId: number) => {
  return request.get<ApiResponse<NodeNote>>(`/nodes/${nodeId}/note`)
}

// 保存/创建随堂笔记
export const saveNodeNote = (nodeId: number, data: { noteContent: string; isImportant: number }) => {
  const noteContent = validateNoteContent(data.noteContent)
  return request.post<ApiResponse<null>>(`/nodes/${nodeId}/note`, { ...data, noteContent })
}

// 修改知识点的学习状态
export const updateNodeStatus = (nodeId: number, status: 'unstarted' | 'learning' | 'completed') => {
  return request.put<ApiResponse<null>>(`/nodes/${nodeId}/status`, { status })
}

// 评价知识点难易程度
export const updateNodeDifficulty = (nodeId: number, difficulty: 'easy' | 'medium' | 'hard') => {
  return request.put<ApiResponse<null>>(`/nodes/${nodeId}/difficulty`, { difficulty })
}

// ========== 创作者视角接口 ==========

// 1. 查询创作知识点 (含断点)
export const getAuthorInitNodes = (subjectId: number) => {
  return request.get<ApiResponse<AuthorInitRes>>('/nodes/author-init', {
    params: { subjectId }
  })
}

// 2. 创作者获取子节点列表接口
export const getAuthorChildNodes = (nodeId: number) => {
  return request.get<ApiResponse<AuthorNode[]>>(`/nodes/${nodeId}/author-children`)
}

// 3. 创作者获取节点内容接口
export const getAuthorNodeContent = (nodeId: number) => {
  return request.get<ApiResponse<AuthorNodeContent>>(`/nodes/${nodeId}/author-content`)
}

// 4. 创作者修改知识点名称草稿 (猜测的PUT接口，基于上下文)
export const updateAuthorNodeName = (nodeId: number, nameDraft: string) => {
  return request.put<ApiResponse<null>>(`/nodes/${nodeId}/author-name`, { nameDraft })
}

// 5. 创作者修改正文草稿 (猜测的PUT接口，基于上下文)
export const updateAuthorNodeContent = (nodeId: number, contentDraft: string) => {
  return request.put<ApiResponse<null>>(`/nodes/${nodeId}/author-content`, { contentDraft })
}
