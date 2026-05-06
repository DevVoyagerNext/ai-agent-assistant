package service

import (
	"backend/dto"
	"backend/global"
	"backend/model"
	"backend/pkg/utils"
	"context"
	"errors"
	"path/filepath"
)

type FileService struct{}

// UploadFile 上传文件到七牛云并保存记录
func (s *FileService) UploadFile(ctx context.Context, fileBytes []byte, fileName string, userID uint) (*dto.FileUploadRes, error) {
	if len(fileBytes) == 0 {
		return nil, errors.New("文件内容为空")
	}

	// 1. 上传到七牛云
	fileKey, err := utils.UploadToQiniu(fileBytes, fileName, "Agent/")
	if err != nil {
		return nil, err
	}

	// 2. 获取扩展名作为类型
	ext := filepath.Ext(fileName)
	if len(ext) > 0 {
		ext = ext[1:] // 去掉点号
	} else {
		ext = "unknown"
	}

	// 3. 存入 files 表
	fileRecord := model.File{
		FileName: fileName,
		FilePath: fileKey, // 存储 key
		FileType: ext,
		FileSize: len(fileBytes),
		UserID:   int(userID),
	}

	if err := global.GVA_DB.WithContext(ctx).Create(&fileRecord).Error; err != nil {
		return nil, err
	}

	// 4. 构建返回
	downloadURL := utils.GetQiniuDownloadURL(fileKey)

	return &dto.FileUploadRes{
		ID:        fileRecord.ID,
		FileName:  fileRecord.FileName,
		FilePath:  downloadURL, // 返回完整链接方便前端展示
		FileType:  fileRecord.FileType,
		FileSize:  fileRecord.FileSize,
		CreatedAt: fileRecord.CreatedAt,
	}, nil
}

// GetFileInfo 读取文件信息
func (s *FileService) GetFileInfo(ctx context.Context, fileID uint) (*dto.FileInfoRes, error) {
	var fileRecord model.File
	if err := global.GVA_DB.WithContext(ctx).First(&fileRecord, fileID).Error; err != nil {
		return nil, errors.New("文件不存在")
	}

	downloadURL := utils.GetQiniuDownloadURL(fileRecord.FilePath)

	return &dto.FileInfoRes{
		ID:        fileRecord.ID,
		FileName:  fileRecord.FileName,
		FilePath:  downloadURL,
		FileType:  fileRecord.FileType,
		FileSize:  fileRecord.FileSize,
		CreatedAt: fileRecord.CreatedAt,
	}, nil
}
