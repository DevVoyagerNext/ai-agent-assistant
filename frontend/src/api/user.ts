import request from '../utils/request'
import type { ApiResponse } from '../types'
import type { UserInfoRes } from '../types/user'

export const getUserInfo = () => {
  return request.get<ApiResponse<UserInfoRes>>('/user/info')
}
