package models

import (
	"context"
	"fmt"
	"sync"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
)

// OpenAIConfig OpenAI 模型配置
type OpenAIConfig struct {
	APIKey         string `json:"api_key" yaml:"api_key"`
	BaseURL        string `json:"base_url" yaml:"base_url"`
	Model          string `json:"model" yaml:"model"`
	ResponseFormat string `json:"response_format" yaml:"response_format"` // 输出格式: "json_object", "text", ""(默认)
}

// OpenAIModelManager OpenAI模型管理器
type OpenAIModelManager struct {
	config *OpenAIConfig
	model  model.ToolCallingChatModel
	mu     sync.RWMutex
}

// NewOpenAIModelManager 创建 OpenAI 管理器
func NewOpenAIModelManager(config *OpenAIConfig) *OpenAIModelManager {
	return &OpenAIModelManager{config: config}
}

// Initialize 初始化模型
func (m *OpenAIModelManager) Initialize(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.model != nil {
		return nil // 已初始化
	}

	if m.config.APIKey == "" {
		return fmt.Errorf("OpenAI API key is required")
	}
	if m.config.Model == "" {
		m.config.Model = "gpt-3.5-turbo"
	}

	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: m.config.BaseURL,
		Model:   m.config.Model,
		APIKey:  m.config.APIKey,
		ResponseFormat: m.getResponseFormat(),
	})
	if err != nil {
		return err
	}

	m.model = chatModel
	return nil
}

// GetModel 获取模型
func (m *OpenAIModelManager) GetModel() model.ToolCallingChatModel {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.model
}

// IsInitialized 是否已初始化
func (m *OpenAIModelManager) IsInitialized() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.model != nil
}

// Close 关闭模型 (如果需要)
func (m *OpenAIModelManager) Close() error {
	// OpenAI SDK 一般不需要关闭，这里预留
	return nil
}

// getResponseFormat 根据配置获取响应格式
func (m *OpenAIModelManager) getResponseFormat() *openai.ChatCompletionResponseFormat {
	switch m.config.ResponseFormat {
	case "json_object":
		return &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		}
	case "text":
		return &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeText,
		}
	default:
		// 默认不设置响应格式，让 OpenAI 自动选择
		return nil
	}
}
