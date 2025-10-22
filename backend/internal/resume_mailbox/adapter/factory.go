package adapter

import (
	"fmt"
	"strings"
	"sync"

	"log/slog"

	"github.com/chaitin/WhaleHire/backend/domain"
)

// Factory 协议适配器工厂实现
type Factory struct {
	adapters map[string]domain.MailboxProtocolAdapter
	mu       sync.RWMutex
}

// NewAdapterFactory 创建新的协议适配器工厂
func NewAdapterFactory(logger *slog.Logger) domain.MailboxAdapterFactory {
	f := &Factory{
		adapters: make(map[string]domain.MailboxProtocolAdapter),
	}

	// 注册内置适配器
	f.RegisterAdapter(NewIMAPAdapter(logger))

	return f
}

// RegisterAdapter 注册适配器
func (f *Factory) RegisterAdapter(adapter domain.MailboxProtocolAdapter) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.adapters[strings.ToLower(adapter.GetProtocol())] = adapter
}

// GetAdapter 获取指定协议的适配器
func (f *Factory) GetAdapter(protocol string) (domain.MailboxProtocolAdapter, error) {
	f.mu.RLock()
	adapter, ok := f.adapters[strings.ToLower(protocol)]
	f.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("暂不支持的邮箱协议: %s", protocol)
	}
	return adapter, nil
}
