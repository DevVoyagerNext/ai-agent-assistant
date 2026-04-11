import request from '../utils/request'
import type { ApiResponse } from '../types/index'
import type { SubjectNode, SubjectNodeDetail, NodeNote } from '../types/node'

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

// 获取知识点随堂笔记（必须登录）
export const getNodeNote = (nodeId: number) => {
  return request.get<ApiResponse<NodeNote>>(`/nodes/${nodeId}/note`)
}

// 修改知识点的学习状态
export const updateNodeStatus = (nodeId: number, status: 'unstarted' | 'learning' | 'completed') => {
  return request.put<ApiResponse<null>>(`/nodes/${nodeId}/status`, { status })
}

// 评价知识点难易程度
export const updateNodeDifficulty = (nodeId: number, difficulty: 'easy' | 'medium' | 'hard') => {
  return request.put<ApiResponse<null>>(`/nodes/${nodeId}/difficulty`, { difficulty })
}
