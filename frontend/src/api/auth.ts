import request from '../utils/request'
import type { ApiResponse } from '../types/index'
import type { AuthData, LoginRequest, RegisterRequest, SendCodeRequest } from '../types/auth'

export const login = (data: LoginRequest) => {
  return request.post<ApiResponse<AuthData>>('/user/login', data)
}

export const register = (data: RegisterRequest) => {
  return request.post<ApiResponse<AuthData>>('/user/register', data)
}

export const sendCode = (data: SendCodeRequest) => {
  return request.post<ApiResponse<unknown>>('/user/send-email', data)
}
