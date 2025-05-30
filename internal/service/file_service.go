package service

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"reisen-be/internal/model"
	"strings"
)

type FileService struct {
	dataDir      string
}

func NewFileService(dataDir string) *FileService {
	return &FileService{dataDir: dataDir}
}

// SaveAvatar 保存用户头像
func (s *FileService) SaveAvatar(userID model.UserId, file *multipart.FileHeader) (string, error) {
	// 验证文件大小
	if file.Size > 2097152 {
		return "", errors.New("文件大小超过限制")
	}

	// 验证文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := false
	for _, t := range []string{ ".jpg", ".jpeg", ".png", ".gif" } {
		if strings.ToLower(t) == ext {
			allowed = true
			break
		}
	}
	if !allowed {
		return "", errors.New("不支持的文件类型")
	}

	// 打开上传文件
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// 生成唯一文件名
	filename := fmt.Sprintf("%d%s", userID, ext)
	filePath := filepath.Join(s.dataDir, "avatar", filename)

	// 创建目标文件
	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// 复制文件内容
	if _, err = io.Copy(dst, src); err != nil {
		return "", err
	}

	return filename, nil
}
