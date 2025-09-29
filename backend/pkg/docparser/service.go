package docparser

import (
	"context"
	"fmt"
	"time"
)

// DocumentParserService 文档解析服务
type DocumentParserService struct {
	client *Client
}

// NewDocumentParserService 创建文档解析服务
func NewDocumentParserService(apiKey string, options ...ClientOption) *DocumentParserService {
	return &DocumentParserService{
		client: NewClient(apiKey, options...),
	}
}

// ParseDocumentResult 文档解析结果
type ParseDocumentResult struct {
	FileID   string    `json:"file_id"`
	Content  string    `json:"content"`
	FileType string    `json:"file_type"`
	Filename string    `json:"filename"`
	Title    string    `json:"title"`
	UploadAt time.Time `json:"upload_at"`
}

// ParseDocument 解析文档（一站式接口）
// 该方法会自动完成：上传文件 -> 获取解析内容 -> 删除文件
func (s *DocumentParserService) ParseDocument(ctx context.Context, filePath string) (*ParseDocumentResult, error) {
	// 1. 上传文件
	uploadResp, err := s.client.UploadFile(ctx, &FileUploadRequest{
		FilePath: filePath,
		Purpose:  "file-extract",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// 确保在函数结束时删除文件
	defer func() {
		if _, delErr := s.client.DeleteFile(context.Background(), uploadResp.ID); delErr != nil {
			// 记录删除失败，但不影响主流程
			// 这里可以根据需要添加日志记录
		}
	}()

	// 2. 等待文件处理完成（如果需要）
	if uploadResp.Status != "ok" {
		return nil, fmt.Errorf("file upload failed with status: %s, details: %s",
			uploadResp.Status, uploadResp.StatusDetails)
	}

	// 3. 获取文件内容
	contentResp, err := s.client.GetFileContent(ctx, uploadResp.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file content: %w", err)
	}

	// 4. 构造结果
	result := &ParseDocumentResult{
		FileID:   uploadResp.ID,
		Content:  contentResp.Content,
		FileType: contentResp.FileType,
		Filename: contentResp.Filename,
		Title:    contentResp.Title,
		UploadAt: uploadResp.GetCreatedTime(),
	}

	return result, nil
}

// ParseDocumentWithoutCleanup 解析文档但不自动删除文件
// 返回文件ID，调用方需要手动删除文件
func (s *DocumentParserService) ParseDocumentWithoutCleanup(ctx context.Context, filePath string) (*ParseDocumentResult, error) {
	// 1. 上传文件
	uploadResp, err := s.client.UploadFile(ctx, &FileUploadRequest{
		FilePath: filePath,
		Purpose:  "file-extract",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// 2. 检查上传状态
	if uploadResp.Status != "ok" {
		return nil, fmt.Errorf("file upload failed with status: %s, details: %s",
			uploadResp.Status, uploadResp.StatusDetails)
	}

	// 3. 获取文件内容
	contentResp, err := s.client.GetFileContent(ctx, uploadResp.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file content: %w", err)
	}

	// 4. 构造结果
	result := &ParseDocumentResult{
		FileID:   uploadResp.ID,
		Content:  contentResp.Content,
		FileType: contentResp.FileType,
		Filename: contentResp.Filename,
		Title:    contentResp.Title,
		UploadAt: uploadResp.GetCreatedTime(),
	}

	return result, nil
}

// CleanupFile 清理文件
func (s *DocumentParserService) CleanupFile(ctx context.Context, fileID string) error {
	_, err := s.client.DeleteFile(ctx, fileID)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// GetClient 获取底层客户端（用于高级用法）
func (s *DocumentParserService) GetClient() *Client {
	return s.client
}
