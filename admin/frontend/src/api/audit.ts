import request from '../utils/request'
import { unwrap } from './helper'
import type {
  ApiResponse,
  AuditRecord,
  ListResponse,
  MeResponse,
  NodeReviewItem,
  SubjectReviewItem
} from '../types/admin'

export const getAdminMe = () => {
  return unwrap(request.get<ApiResponse<MeResponse>>('/admin/me'))
}

export const getPendingSubjects = () => {
  return unwrap(request.get<ApiResponse<ListResponse<SubjectReviewItem>>>('/admin/subjects/pending'))
}

export const getPendingNodes = () => {
  return unwrap(request.get<ApiResponse<ListResponse<NodeReviewItem>>>('/admin/nodes/pending'))
}

export const getAuditRecords = () => {
  return unwrap(request.get<ApiResponse<ListResponse<AuditRecord>>>('/admin/audit-records'))
}

export const approveSubject = (id: number, remark = '') => {
  return unwrap(request.post<ApiResponse<null>>(`/admin/subjects/${id}/approve`, { remark }))
}

export const rejectSubject = (id: number, remark: string) => {
  return unwrap(request.post<ApiResponse<null>>(`/admin/subjects/${id}/reject`, { remark }))
}

export const approveNode = (id: number, remark = '') => {
  return unwrap(request.post<ApiResponse<null>>(`/admin/nodes/${id}/approve`, { remark }))
}

export const rejectNode = (id: number, remark: string) => {
  return unwrap(request.post<ApiResponse<null>>(`/admin/nodes/${id}/reject`, { remark }))
}
