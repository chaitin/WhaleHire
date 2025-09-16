package loaddoc

// Document 文档结构，基于 Eino 框架的 Document 概念
type Document struct {
	ID       string         `json:"id"`
	Content  string         `json:"content"`
	Metadata map[string]any `json:"metadata,omitempty"`
	Source   string         `json:"source"`
	Type     DocumentType   `json:"type"`
}

// DocumentType 文档类型
type DocumentType string

const (
	DocumentTypeFile DocumentType = "file"
	DocumentTypeURL  DocumentType = "url"
)

// BatchLoadInput 批量加载输入
type BatchLoadInput struct {
	FilePaths []string `json:"file_paths,omitempty"` // 文件路径列表
	URLs      []string `json:"urls,omitempty"`       // 在线链接列表
}

// BatchLoadResult 批量加载结果
type BatchLoadResult struct {
	Documents []*Document `json:"documents"`
	Errors    []LoadError `json:"errors,omitempty"`
}

// LoadError 加载错误
type LoadError struct {
	Source string `json:"source"`
	Error  string `json:"error"`
}
