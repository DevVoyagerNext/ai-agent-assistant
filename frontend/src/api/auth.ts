import request from '../utils/request'
import type { ApiResponse } from '../types/index'
import type { AuthData } from '../types/auth'

export const login = (data: any) => {
  return request.post<ApiResponse<AuthData>>('/user/login', data)
}

export const register = (data: any) => {
  return request.post<ApiResponse<AuthData>>('/user/register', data)
}

export const sendCode = (data: any) => {
  return request.post<ApiResponse<any>>('/user/send-code', data)
}
