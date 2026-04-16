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
  CreatePrivateNoteReq,
  SharePrivateNoteReq,
  SharePrivateNoteRes
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

// 1. 获取私人笔记列表 (分页)
export const getPrivateNoteDetail = (noteId: number, scope = 2, page = 1, pageSize = 20) => {
  return request.get<ApiResponse<PrivateNoteResponse>>(`/user/notes/private/${noteId}`, {
    params: { scope, page, pageSize }
  })
}

// 创建新的私人笔记或文件夹
export const createPrivateNote = (data: { parentId: number, type: 'folder' | 'markdown', title: string, isPublic: 0 | 1 }) => {
  return request.post<ApiResponse<null>>('/user/notes/private', data)
}

// 2. 删除私人笔记 (文件或文件夹)
export const deletePrivateNote = (noteId: number) => {
  return request.delete<ApiResponse<null>>(`/user/notes/private/${noteId}`)
}

// 3. 修改私人笔记内容 (仅限文件)
export const updatePrivateNoteContent = (noteId: number, content: string) => {
  return request.put<ApiResponse<null>>(`/user/notes/private/${noteId}/content`, { content })
}

// 3.5 修改私人笔记标题 (仅限文件)
export const updatePrivateNoteName = (noteId: number, title: string) => {
  return request.put<ApiResponse<null>>(`/user/notes/private/${noteId}/title`, { title })
}

// 4. 修改文件或文件夹标题
export const updatePrivateNoteTitle = (noteId: number, title: string) => {
  return request.put<ApiResponse<null>>(`/user/notes/private/${noteId}/title`, { title })
}

// 5. 修改文件或文件夹公开状态
export const updatePrivateNotePublic = (noteId: number, isPublic: 0 | 1) => {
  return request.put<ApiResponse<null>>(`/user/notes/private/${noteId}/public`, { isPublic })
}

// 6. 分享私人笔记
export const sharePrivateNote = (noteId: number, expiresAt: string) => {
  return request.post<ApiResponse<SharePrivateNoteRes>>(`/user/notes/private/${noteId}/share`, { expiresAt })
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

// 2. 获取该用户点赞的教材 (分页)
export const getUserLikedSubjects = (page = 1, pageSize = 20) => {
  return request.get<ApiResponse<RecentSubjectListRes>>('/user/subjects/liked', {
    params: { page, pageSize }
  })
}

// 3. 获取该用户正在学习的教材 (分页)
export const getUserLearningSubjects = (page = 1, pageSize = 20) => {
  return request.get<ApiResponse<RecentSubjectListRes>>('/user/subjects/learning', {
    params: { page, pageSize }
  })
}

// 4. 获取该用户已经学习完成的教材 (分页)
export const getUserCompletedSubjects = (page = 1, pageSize = 20) => {
  return request.get<ApiResponse<RecentSubjectListRes>>('/user/subjects/completed', {
    params: { page, pageSize }
  })
}

// 5. 获取该用户最近学习的教材 (分页)
export const getUserRecentSubjects = (page = 1, pageSize = 10) => {
  return request.get<ApiResponse<RecentSubjectListRes>>('/user/subjects/last-learning', {
    params: { page, pageSize }
  })
}

// 6. 获取该用户已经收藏的教材 (分页)
export const getUserCollectedSubjects = (page = 1, pageSize = 20) => {
  return request.get<ApiResponse<RecentSubjectListRes>>('/user/subjects/collected', {
    params: { page, pageSize }
  })
}

// 7. 获取某个收藏夹下的教材 (分页)
export const getSubjectsInFolder = (folderId: number, page = 1, pageSize = 20) => {
  return request.get<ApiResponse<RecentSubjectListRes>>(`/user/subjects/folders/${folderId}`, {
    params: { page, pageSize }
  })
}

// 8. 修改收藏夹公开状态
export const updateCollectFolderPublic = (folderId: number, isPublic: 0 | 1) => {
  return request.put<ApiResponse<null>>(`/user/subjects/folders/${folderId}/public`, { isPublic })
}

// 9. 修改收藏夹名称
export const updateCollectFolderName = (folderId: number, name: string) => {
  return request.put<ApiResponse<null>>(`/user/subjects/folders/${folderId}/name`, { name })
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
