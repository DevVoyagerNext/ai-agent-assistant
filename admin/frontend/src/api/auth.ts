import request from '../utils/request'
import { unwrap } from './helper'
import type { ApiResponse, AuthResponse, LoginPayload, RegisterPayload } from '../types/admin'

export const loginAdmin = (payload: LoginPayload) => {
  return unwrap(request.post<ApiResponse<AuthResponse>>('/admin/login', payload))
}

export const registerAdmin = (payload: RegisterPayload) => {
  return unwrap(request.post<ApiResponse<AuthResponse>>('/admin/register', payload))
}
