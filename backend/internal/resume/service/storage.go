package service

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"

	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/pkg/store/s3"
)

// StorageService 文件存储服务实现
type StorageService struct {
	minioClient  *s3.MinioClient
	bucket       string
	maxFileSize  int64
	allowedTypes []string
	logger       *slog.Logger
}

// NewStorageService 创建新的存储服务实例
func NewStorageService(minioClient *s3.MinioClient, cfg *config.Config, logger *slog.Logger) domain.StorageService {
	return &StorageService{
		minioClient:  minioClient,
		bucket:       cfg.S3.BucketName,
		maxFileSize:  cfg.FileStorage.MaxFileSize,
		allowedTypes: cfg.FileStorage.AllowedTypes,
		logger:       logger,
	}
}

// Upload 上传文件
func (s *StorageService) Upload(ctx context.Context, file io.Reader, filename string) (*domain.FileInfo, error) {
	// 验证文件类型
	if err := s.validateFileType(filename); err != nil {
		return nil, err
	}

	// 生成唯一文件名
	storedName := s.generateFilename(filename)

	// 创建对象路径 (按年月组织)
	now := time.Now()
	objectName := filepath.Join("resumes", now.Format("2006"), now.Format("01"), storedName)

	// 上传文件到MinIO
	info, err := s.minioClient.Client.PutObject(ctx, s.bucket, objectName, file, -1, minio.PutObjectOptions{
		ContentType: s.getContentType(filename),
		UserMetadata: map[string]string{
			"originalname": filename,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("上传文件失败: %w", err)
	}

	// 检查文件大小
	if info.Size > s.maxFileSize {
		// 删除已上传的文件
		if err := s.minioClient.Client.RemoveObject(ctx, s.bucket, objectName, minio.RemoveObjectOptions{}); err != nil {
			// 记录删除失败的错误，但不影响主要错误的返回
			s.logger.Warn("删除文件失败", "error", err, "objectName", objectName)
		}
		return nil, fmt.Errorf("文件大小超出限制: %d bytes", info.Size)
	}

	// 生成访问URL
	fileURL := s.generateFileURL(objectName)

	return &domain.FileInfo{
		OriginalName: filename,
		StoredName:   storedName,
		FilePath:     objectName, // 使用对象路径作为文件路径
		FileURL:      fileURL,
		Size:         info.Size,
		ContentType:  s.getContentType(filename),
	}, nil
}

// Download 下载文件
func (s *StorageService) Download(ctx context.Context, url string) (io.Reader, error) {
	objectName, err := s.getObjectName(url)
	if err != nil {
		return nil, err
	}

	object, err := s.minioClient.Client.GetObject(ctx, s.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("下载文件失败: %w", err)
	}

	return object, nil
}

// Delete 删除文件
func (s *StorageService) Delete(ctx context.Context, url string) error {
	objectName, err := s.getObjectName(url)
	if err != nil {
		return err
	}

	if err := s.minioClient.Client.RemoveObject(ctx, s.bucket, objectName, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("删除文件失败: %w", err)
	}

	return nil
}

// GetLocalPath 从URL获取对象名称（保持接口兼容性）
func (s *StorageService) GetLocalPath(url string) (string, error) {
	return s.getObjectName(url)
}

// getObjectName 从URL提取对象名称
func (s *StorageService) getObjectName(url string) (string, error) {
	// 从URL中提取对象名称
	// 假设URL格式为: http://domain/bucket/object/path
	parts := strings.Split(url, "/")
	if len(parts) < 4 {
		return "", fmt.Errorf("无效的文件URL: %s", url)
	}

	// 获取bucket后面的路径作为对象名称
	objectName := strings.Join(parts[4:], "/")
	if objectName == "" {
		return "", fmt.Errorf("无法从URL提取对象名称: %s", url)
	}

	return objectName, nil
}

// validateFileType 验证文件类型
func (s *StorageService) validateFileType(filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))

	for _, allowedType := range s.allowedTypes {
		if ext == allowedType {
			return nil
		}
	}

	return fmt.Errorf("不支持的文件类型: %s", ext)
}

// generateFilename 生成唯一文件名
func (s *StorageService) generateFilename(originalName string) string {
	ext := filepath.Ext(originalName)
	uuid := uuid.New().String()
	timestamp := time.Now().Format("20060102150405")
	return fmt.Sprintf("%s_%s%s", uuid, timestamp, ext)
}

// generateFileURL 生成文件访问URL
func (s *StorageService) generateFileURL(objectName string) string {
	// 使用MinIO客户端生成预签名URL，有效期为24小时
	ctx := context.Background()
	expires := 24 * time.Hour

	url, err := s.minioClient.SignURL(ctx, s.bucket, objectName, expires)
	if err != nil {
		// 如果生成预签名URL失败，回退到公共访问URL
		return fmt.Sprintf("http://%s/%s/%s", s.minioClient.Client.EndpointURL().Host, s.bucket, objectName)
	}

	return url
}

// getContentType 根据文件扩展名获取内容类型
func (s *StorageService) getContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".doc":
		return "application/msword"
	case ".txt":
		return "text/plain"
	default:
		return "application/octet-stream"
	}
}
