package docparser

import (
	"fmt"
	"time"

	"github.com/chaitin/WhaleHire/backend/config"
)

// NewDocumentParserServiceFromConfig 从配置创建文档解析服务
func NewDocumentParserServiceFromConfig(cfg *config.Config) *DocumentParserService {
	client := NewClientFromConfig(cfg)
	return &DocumentParserService{
		client: client,
	}
}

// NewClientFromConfig 从配置创建客户端
func NewClientFromConfig(cfg *config.Config) *Client {
	options := []ClientOption{
		WithBaseURL(cfg.DocumentParser.BaseURL),
		WithTimeout(time.Duration(cfg.DocumentParser.Timeout) * time.Second),
	}

	return NewClient(cfg.DocumentParser.APIKey, options...)
}

// ValidateConfig 验证配置
func ValidateConfig(cfg *config.Config) error {
	if cfg.DocumentParser.APIKey == "" {
		return fmt.Errorf("document parser API key is required")
	}
	if cfg.DocumentParser.BaseURL == "" {
		return fmt.Errorf("document parser base URL is required")
	}
	return nil
}
