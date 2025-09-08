package models

import (
	"context"
	"fmt"
	"sync"

	"github.com/cloudwego/eino/components/model"
)

// ==================== 基础定义 ====================

// ModelType 模型类型
type ModelType string

const (
	ModelTypeOpenAI ModelType = "openai"
)

// ModelKey 模型唯一标识 (类型 + 名称)
type ModelKey struct {
	Type ModelType
	Name string
}

// ModelManager 模型管理器接口
type ModelManager interface {
	Initialize(ctx context.Context) error
	GetModel() model.ToolCallingChatModel
	IsInitialized() bool
	Close() error
}

// ==================== 工厂实现 ====================

// ModelFactory 模型工厂
type ModelFactory struct {
	mu       sync.RWMutex
	managers map[ModelKey]*lazyManager
}

// lazyManager 支持懒加载初始化
type lazyManager struct {
	manager ModelManager
	once    sync.Once
	initErr error
}

// NewModelFactory 创建工厂
func NewModelFactory() *ModelFactory {
	return &ModelFactory{
		managers: make(map[ModelKey]*lazyManager),
	}
}

// Register 注册模型管理器
func (f *ModelFactory) Register(modelType ModelType, name string, manager ModelManager) {
	f.mu.Lock()
	defer f.mu.Unlock()
	key := ModelKey{Type: modelType, Name: name}
	f.managers[key] = &lazyManager{manager: manager}
}

// GetModel 获取模型 (懒加载初始化)
func (f *ModelFactory) GetModel(ctx context.Context, modelType ModelType, name string) (model.ToolCallingChatModel, error) {
	f.mu.RLock()
	lazy, exists := f.managers[ModelKey{Type: modelType, Name: name}]
	f.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("model %s:%s not registered", modelType, name)
	}

	// 懒加载初始化
	lazy.once.Do(func() {
		lazy.initErr = lazy.manager.Initialize(ctx)
	})

	if lazy.initErr != nil {
		return nil, fmt.Errorf("failed to initialize model %s:%s: %w", modelType, name, lazy.initErr)
	}

	return lazy.manager.GetModel(), nil
}

// IsInitialized 判断是否初始化
func (f *ModelFactory) IsInitialized(modelType ModelType, name string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	lazy, exists := f.managers[ModelKey{Type: modelType, Name: name}]
	if !exists {
		return false
	}
	return lazy.manager.IsInitialized()
}

// Close 关闭所有模型
func (f *ModelFactory) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	var errs []error
	for key, lazy := range f.managers {
		if err := lazy.manager.Close(); err != nil {
			errs = append(errs, fmt.Errorf("%s:%s close error: %w", key.Type, key.Name, err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors during close: %v", errs)
	}
	return nil
}
