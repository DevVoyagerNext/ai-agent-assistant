package dto

import "time"

// FileUploadRes 文件上传返回信息
type FileUploadRes struct {
	ID        uint      `json:"id"`
	FileName  string    `json:"fileName"`
	FilePath  string    `json:"filePath"` // 可以是完整下载链接
	FileType  string    `json:"fileType"`
	FileSize  int       `json:"fileSize"`
	CreatedAt time.Time `json:"createdAt"`
}

// FileInfoRes 文件信息返回
type FileInfoRes struct {
	ID        uint      `json:"id"`
	FileName  string    `json:"fileName"`
	FilePath  string    `json:"filePath"` // 完整下载链接
	FileType  string    `json:"fileType"`
	FileSize  int       `json:"fileSize"`
	CreatedAt time.Time `json:"createdAt"`
}
