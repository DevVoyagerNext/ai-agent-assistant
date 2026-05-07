export interface FileUploadRes {
  id: number
  fileName: string
  filePath: string
  FileType: string // 首字母大写，与后端返回一致
  fileSize: number
  createdAt: string
}
