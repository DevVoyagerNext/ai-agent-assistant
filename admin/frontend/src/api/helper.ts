import type { ApiResponse } from '../types/admin'

export const unwrap = async <T>(promise: Promise<{ data: ApiResponse<T> }>) => {
  const response = await promise
  if (response.data.code !== 200) {
    throw new Error(response.data.msg || '请求失败')
  }
  return response.data.data
}
