import request from '../utils/request'
import type { ApiResponse } from '../types'
import type { FileUploadRes } from '../types/file'

/**
 * 上传文件到服务器 (后端会转存至七牛云)
 * @param file 原始文件对象
 */
export const uploadFile = (file: File) => {
  const formData = new FormData()
  formData.append('file', file)
  
  return request.post<ApiResponse<FileUploadRes>>('/files/upload', formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}
