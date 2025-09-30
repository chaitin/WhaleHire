package docparser

import "time"

// FileUploadResponse 文件上传响应
type FileUploadResponse struct {
	ID            string `json:"id"`
	Object        string `json:"object"`
	Bytes         int64  `json:"bytes"`
	CreatedAt     int64  `json:"created_at"`
	Filename      string `json:"filename"`
	Purpose       string `json:"purpose"`
	Status        string `json:"status"`
	StatusDetails string `json:"status_details"`
	FileType      string `json:"file_type"`
}

// FileContentResponse 文件内容响应
type FileContentResponse struct {
	Content  string `json:"content"`
	FileType string `json:"file_type"`
	Filename string `json:"filename"`
	Title    string `json:"title"`
	Type     string `json:"type"`
}

// FileDeleteResponse 文件删除响应
type FileDeleteResponse struct {
	Deleted bool   `json:"deleted"`
	ID      string `json:"id"`
	Object  string `json:"object"`
}

// ErrorResponse API错误响应
type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

// FileUploadRequest 文件上传请求参数
type FileUploadRequest struct {
	Purpose  string // 固定为 "file-extract"
	FilePath string // 本地文件路径
}

// GetCreatedTime 获取创建时间
func (f *FileUploadResponse) GetCreatedTime() time.Time {
	return time.Unix(f.CreatedAt, 0)
}
