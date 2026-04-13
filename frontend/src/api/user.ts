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
  RecentSubjectListRes,
  PrivateNoteResponse,
  CreatePrivateNoteReq
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

// ---------------- 私人笔记分层管理 API ----------------

// 1. 获取私人笔记内容或子文件夹列表 (noteId 为 0 获取根目录)
export const getPrivateNoteDetail = (noteId: number) => {
  return request.get<ApiResponse<PrivateNoteResponse>>(`/user/notes/private/${noteId}`)
}

// 2. 创建私人文件夹或笔记
export const createPrivateNote = (data: CreatePrivateNoteReq) => {
  return request.post<ApiResponse<null>>('/user/notes/private', data)
}

// ----------------------------------------------------

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

// 获取该用户已经收藏的教材
export const getUserCollectedSubjects = () => {
  return request.get<ApiResponse<UserSubjectProgressRes[]>>('/user/subjects/collected')
}

// 6. 点赞或取消点赞教材
export const likeSubject = (id: number) => {
  return request.post<ApiResponse<{ isLiked: boolean }>>(`/user/subjects/${id}/like`)
}

// 7. 创建当前用户的收藏夹
export const createCollectFolder = (data: { name: string; description: string; isPublic: number }) => {
  return request.post<ApiResponse<CollectFolderRes>>('/user/subjects/folders', data)
}

// 8. 把教材添加到用户的收藏夹
export const addSubjectToFolder = (folderId: number, subjectId: number) => {
  return request.post<ApiResponse<null>>(`/user/subjects/folders/${folderId}/subjects`, { subjectId })
}

// 9. 取消教材收藏 (从所有收藏夹移除)
export const uncollectSubject = (id: number) => {
  return request.delete<ApiResponse<null>>(`/user/subjects/${id}/collect`)
}
