package docparser

import (
	"errors"
	"fmt"
)

// 预定义错误类型
var (
	ErrInvalidAPIKey     = errors.New("invalid API key")
	ErrFileNotFound      = errors.New("file not found")
	ErrFileTooLarge      = errors.New("file too large")
	ErrUnsupportedFormat = errors.New("unsupported file format")
	ErrUploadFailed      = errors.New("file upload failed")
	ErrProcessingFailed  = errors.New("file processing failed")
	ErrNetworkError      = errors.New("network error")
	ErrTimeout           = errors.New("request timeout")
)

// APIError API错误结构
type APIError struct {
	StatusCode int    `json:"status_code"`
	Type       string `json:"type"`
	Message    string `json:"message"`
	RequestID  string `json:"request_id,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error (status %d): %s - %s", e.StatusCode, e.Type, e.Message)
}

// NewAPIError 创建API错误
func NewAPIError(statusCode int, errorType, message string) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Type:       errorType,
		Message:    message,
	}
}

// IsRetryableError 判断错误是否可重试
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// 检查是否为API错误
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		// 5xx错误通常可以重试
		return apiErr.StatusCode >= 500 && apiErr.StatusCode < 600
	}

	// 网络错误和超时错误可以重试
	return errors.Is(err, ErrNetworkError) || errors.Is(err, ErrTimeout)
}

// IsAuthError 判断是否为认证错误
func IsAuthError(err error) bool {
	if err == nil {
		return false
	}

	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 401 || apiErr.Type == "incorrect_api_key_error"
	}

	return errors.Is(err, ErrInvalidAPIKey)
}

// IsFileError 判断是否为文件相关错误
func IsFileError(err error) bool {
	if err == nil {
		return false
	}

	return errors.Is(err, ErrFileNotFound) ||
		errors.Is(err, ErrFileTooLarge) ||
		errors.Is(err, ErrUnsupportedFormat) ||
		errors.Is(err, ErrUploadFailed) ||
		errors.Is(err, ErrProcessingFailed)
}

// WrapError 包装错误，添加上下文信息
func WrapError(err error, operation string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", operation, err)
}