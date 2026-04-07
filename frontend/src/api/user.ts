import request from '../utils/request'
import type { ApiResponse } from '../types'
import type { 
  UserInfoRes, 
  ActivityCalendarRes, 
  PublicPrivateNotesRes, 
  SharedNotesRes, 
  LearnedSubjectsRes 
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
