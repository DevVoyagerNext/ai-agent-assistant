import request from '../utils/request'
import type { ApiResponse } from '../types/index'
import type { Subject, SubjectCategory, SubjectSearchRes } from '../types/subject'
import type { CreatedSubjectItem } from '../types/user'

// 获取教材分类
export const getSubjectCategories = () => {
  return request.get<ApiResponse<SubjectCategory[]>>('/subjects/categories')
}

// 获取所有教材
export const getAllSubjects = () => {
  return request.get<ApiResponse<Subject[]>>('/subjects')
}

// 通过分类获取该分类下的教材
export const getSubjectsByCategory = (categoryId: number) => {
  return request.get<ApiResponse<Subject[]>>(`/subjects/category/${categoryId}`)
}

export const searchSubjects = (keyword: string, page = 1, pageSize = 20) => {
  return request.get<ApiResponse<SubjectSearchRes>>('/subjects/search', {
    params: { keyword, page, pageSize }
  })
}

// 获取教材详情
export const getSubjectDetail = (id: number) => {
  return request.get<ApiResponse<Subject>>(`/subjects/${id}`)
}

// 作者第一次发布教材：进入后台待审核队列，审核通过后才会正式 published
export const publishSubject = (subjectId: number) => {
  return request.put<ApiResponse<CreatedSubjectItem>>(`/user/subjects/${subjectId}/publish`)
}
