package usecase

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"

	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/pkg/store/s3"
)

type FileUsecase struct {
	logger   *slog.Logger
	s3Client *s3.MinioClient
	config   *config.Config
}

func NewFileUsecase(logger *slog.Logger, s3Client *s3.MinioClient, config *config.Config) *FileUsecase {
	return &FileUsecase{
		s3Client: s3Client,
		logger:   logger,
		config:   config,
	}
}

func (u *FileUsecase) UploadFile(ctx context.Context, kbID string, file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	ext := strings.ToLower(filepath.Ext(file.Filename))
	filename := fmt.Sprintf("%s/%s%s", kbID, uuid.New().String(), ext)

	size := file.Size

	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = mime.TypeByExtension(ext)
	}

	resp, err := u.s3Client.PutObject(
		ctx,
		domain.Bucket,
		filename,
		src,
		size,
		minio.PutObjectOptions{
			ContentType: contentType,
			UserMetadata: map[string]string{
				"originalname": file.Filename,
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("upload failed: %w", err)
	}

	return resp.Key, nil
}

func (u *FileUsecase) UploadFileFromBytes(ctx context.Context, kbID string, filename string, fileBytes []byte) (string, error) {
	// Create a reader from the byte slice
	reader := bytes.NewReader(fileBytes)

	ext := strings.ToLower(filepath.Ext(filename))
	s3Filename := fmt.Sprintf("%s/%s%s", kbID, uuid.New().String(), ext)

	size := int64(len(fileBytes))

	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		// Fallback content type if extension not recognized
		contentType = "application/octet-stream"
	}

	resp, err := u.s3Client.PutObject(
		ctx,
		domain.Bucket,
		s3Filename,
		reader,
		size,
		minio.PutObjectOptions{
			ContentType: contentType,
			UserMetadata: map[string]string{
				"originalname": filename,
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("upload failed: %w", err)
	}

	return resp.Key, nil
}

func (u *FileUsecase) UploadFileFromReader(
	ctx context.Context,
	kbID string,
	filename string,
	reader io.Reader,
	size int64, // 必须提供对象大小
) (string, error) {
	// 生成唯一文件名
	ext := strings.ToLower(filepath.Ext(filename))
	s3Filename := fmt.Sprintf("%s/%s%s", kbID, uuid.New().String(), ext)

	// 获取内容类型
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream" // 默认类型
	}

	// 上传到 S3
	_, err := u.s3Client.PutObject(
		ctx,
		domain.Bucket,
		s3Filename,
		reader,
		size, // 必须提供对象大小
		minio.PutObjectOptions{
			ContentType: contentType,
			UserMetadata: map[string]string{
				"originalname": filename,
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("S3 upload failed: %w", err)
	}

	return s3Filename, nil
}

func (u *FileUsecase) GenerateDownloadURL(ctx context.Context, key string) (string, error) {
	return u.s3Client.SignURL(ctx, domain.Bucket, key, time.Hour)
}

// DownloadFile 直接下载文件流
func (u *FileUsecase) DownloadFile(ctx context.Context, key string) (io.ReadCloser, *domain.FileInfo, error) {
	// 获取对象信息
	objInfo, err := u.s3Client.Client.StatObject(ctx, domain.Bucket, key, minio.StatObjectOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get object info: %w", err)
	}

	// 获取文件流
	object, err := u.s3Client.Client.GetObject(ctx, domain.Bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get object: %w", err)
	}

	// 构造文件信息
	fileInfo := &domain.FileInfo{
		OriginalName: objInfo.UserMetadata["originalname"],
		StoredName:   filepath.Base(key),
		FilePath:     key,
		Size:         objInfo.Size,
		ContentType:  objInfo.ContentType,
	}

	// 如果没有原始文件名，使用存储的文件名
	if fileInfo.OriginalName == "" {
		fileInfo.OriginalName = fileInfo.StoredName
	}

	return object, fileInfo, nil
}
