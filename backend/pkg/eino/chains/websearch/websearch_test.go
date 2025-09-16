package websearch

import (
	"context"
	"testing"

	"github.com/cloudwego/eino-ext/components/tool/duckduckgo/v2"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockToolCallingChatModel 模拟聊天模型
type MockToolCallingChatModel struct {
	mock.Mock
}

func (m *MockToolCallingChatModel) WithTools(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	args := m.Called(tools)
	return args.Get(0).(model.ToolCallingChatModel), args.Error(1)
}

func (m *MockToolCallingChatModel) Generate(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*schema.Message), args.Error(1)
}

func (m *MockToolCallingChatModel) Stream(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*schema.StreamReader[*schema.Message]), args.Error(1)
}

func TestNewWebSearchChain(t *testing.T) {
	ctx := context.Background()

	t.Run("成功创建WebSearchChain", func(t *testing.T) {
		// 创建模拟聊天模型
		mockChat := &MockToolCallingChatModel{}
		mockChat.On("WithTools", mock.AnythingOfType("[]*schema.ToolInfo")).Return(mockChat, nil)

		// 调用函数
		chain, err := NewWebSearchChain(ctx, mockChat)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, chain)
		assert.NotNil(t, chain.GetChain())
		assert.NotEmpty(t, chain.tools)

		// 验证模拟调用
		mockChat.AssertExpectations(t)
	})

	t.Run("聊天模型为nil时应该panic或返回错误", func(t *testing.T) {
		// 这个测试用例检查传入nil聊天模型的情况
		// 注意：根据当前实现，传入nil会导致panic，这里我们捕获panic
		defer func() {
			if r := recover(); r != nil {
				// 预期会panic，这是正常的
				t.Logf("Expected panic caught: %v", r)
			}
		}()

		chain, err := NewWebSearchChain(ctx, nil)
		// 如果没有panic，那么应该返回错误
		if err == nil && chain != nil {
			t.Error("Expected error or panic when chat model is nil")
		}
	})
}

func TestNewWebSearchChainWithConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("使用默认配置创建WebSearchChain", func(t *testing.T) {
		mockChat := &MockToolCallingChatModel{}
		mockChat.On("WithTools", mock.AnythingOfType("[]*schema.ToolInfo")).Return(mockChat, nil)

		chain, err := NewWebSearchChainWithConfig(ctx, mockChat, nil)

		assert.NoError(t, err)
		assert.NotNil(t, chain)
		assert.NotNil(t, chain.GetChain())
		assert.NotEmpty(t, chain.tools)

		mockChat.AssertExpectations(t)
	})

	t.Run("使用自定义配置创建WebSearchChain", func(t *testing.T) {
		mockChat := &MockToolCallingChatModel{}
		mockChat.On("WithTools", mock.AnythingOfType("[]*schema.ToolInfo")).Return(mockChat, nil)

		config := &WebSearchConfig{
			MaxResults: 10,
			Region:     duckduckgo.RegionUS,
		}

		chain, err := NewWebSearchChainWithConfig(ctx, mockChat, config)

		assert.NoError(t, err)
		assert.NotNil(t, chain)
		assert.NotNil(t, chain.GetChain())
		assert.NotEmpty(t, chain.tools)

		mockChat.AssertExpectations(t)
	})

	t.Run("WithTools返回错误时应该失败", func(t *testing.T) {
		mockChat := &MockToolCallingChatModel{}
		// 返回错误时，第一个参数应该是有效的ToolCallingChatModel或者nil，但类型要匹配
		mockChat.On("WithTools", mock.AnythingOfType("[]*schema.ToolInfo")).Return((*MockToolCallingChatModel)(nil), assert.AnError)

		chain, err := NewWebSearchChainWithConfig(ctx, mockChat, nil)

		assert.Error(t, err)
		assert.Nil(t, chain)
		assert.Contains(t, err.Error(), "failed to configure chat model with tools")

		mockChat.AssertExpectations(t)
	})
}

func TestWebSearchConfig(t *testing.T) {
	t.Run("WebSearchConfig字段验证", func(t *testing.T) {
		config := &WebSearchConfig{
			MaxResults: 5,
			Region:     duckduckgo.RegionCN,
		}

		assert.Equal(t, 5, config.MaxResults)
		assert.Equal(t, duckduckgo.RegionCN, config.Region)
	})

	t.Run("空配置应该使用默认值", func(t *testing.T) {
		config := &WebSearchConfig{}

		assert.Equal(t, 0, config.MaxResults)                 // 默认值
		assert.Equal(t, duckduckgo.Region(""), config.Region) // 默认值
	})
}

func TestWebSearchChain_GetChain(t *testing.T) {
	ctx := context.Background()
	mockChat := &MockToolCallingChatModel{}
	mockChat.On("WithTools", mock.AnythingOfType("[]*schema.ToolInfo")).Return(mockChat, nil)

	chain, err := NewWebSearchChain(ctx, mockChat)
	assert.NoError(t, err)
	assert.NotNil(t, chain)

	// 测试GetChain方法
	resultChain := chain.GetChain()
	assert.NotNil(t, resultChain)
	assert.Equal(t, chain.chain, resultChain)

	mockChat.AssertExpectations(t)
}

// 集成测试 - 需要真实的依赖项
func TestNewWebSearchChain_Integration(t *testing.T) {
	t.Skip("跳过集成测试 - 需要真实的聊天模型实现")

	// 这里可以添加使用真实聊天模型的集成测试
	// 例如使用OpenAI或其他实际的模型实现
}

// 基准测试
func BenchmarkNewWebSearchChain(b *testing.B) {
	ctx := context.Background()
	mockChat := &MockToolCallingChatModel{}
	mockChat.On("WithTools", mock.AnythingOfType("[]*schema.ToolInfo")).Return(mockChat, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NewWebSearchChain(ctx, mockChat)
		if err != nil {
			b.Fatal(err)
		}
	}
}
