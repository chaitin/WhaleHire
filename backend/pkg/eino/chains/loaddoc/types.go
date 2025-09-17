package loaddoc

// BatchLoadInput 批量加载输入
type BatchLoadInput struct {
	FilePaths []string `json:"file_paths,omitempty"` // 文件路径列表
	URLs      []string `json:"urls,omitempty"`       // 在线链接列表
}

// LoadError 加载错误
type LoadError struct {
	Source string `json:"source"`
	Error  string `json:"error"`
}
