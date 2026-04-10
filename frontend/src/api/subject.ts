import request from '../utils/request'
import type { ApiResponse } from '../types/index'
import type { Subject, SubjectCategory } from '../types/subject'

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
