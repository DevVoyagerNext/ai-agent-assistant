import request from '../utils/request'
import type { ApiResponse } from '../types'
import type { Subject } from '../types/subject'
import type { 
  UserInfoRes, 
  ActivityCalendarRes, 
  PublicPrivateNotesRes, 
  SharedNotesRes, 
  LearnedSubjectsRes,
  CollectFolderRes,
  UserSubjectProgressRes,
  RecentSubjectListRes
} from '../types/user'

export const getUserInfo = () => {
  return request.get<ApiResponse<UserInfoRes>>('/user/info')
}

export const getUserActivitiesCalendar = () => {
  return request.get<ApiResponse<ActivityCalendarRes>>('/user/activities/calendar')
}

export const getPublicPrivateNotes = (page = 1, pageSize = 10) => {
  return request.get<ApiResponse<PublicPrivateNotesRes>>('/user/notes/private/public-list', {
    params: { page, pageSize }
  })
}

export const getSharedNotes = (page = 1, pageSize = 10) => {
  return request.get<ApiResponse<SharedNotesRes>>('/user/notes/shares', {
    params: { page, pageSize }
  })
}

export const getLearnedSubjects = () => {
  return request.get<ApiResponse<LearnedSubjectsRes>>('/user/subjects/learned')
}

// 1. 获取该用户的收藏夹
export const getUserCollectFolders = () => {
  return request.get<ApiResponse<CollectFolderRes[]>>('/user/subjects/folders')
}

// 2. 获取该用户点赞的教材
export const getUserLikedSubjects = () => {
  return request.get<ApiResponse<Subject[]>>('/user/subjects/liked')
}

// 3. 获取该用户正在学习的教材
export const getUserLearningSubjects = () => {
  return request.get<ApiResponse<UserSubjectProgressRes[]>>('/user/subjects/learning')
}

// 4. 获取该用户已经学习完成的教材
export const getUserCompletedSubjects = () => {
  return request.get<ApiResponse<UserSubjectProgressRes[]>>('/user/subjects/completed')
}

// 5. 获取该用户最近学习的教材 (分页)
export const getUserRecentSubjects = (page = 1, pageSize = 10) => {
  return request.get<ApiResponse<RecentSubjectListRes>>('/user/subjects/last-learning', {
    params: { page, pageSize }
  })
}
