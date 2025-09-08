package usecase

import (
	"context"
	"fmt"
	"io"

	"github.com/cloudwego/eino/schema"

	"github.com/ptonlix/whalehire/backend/config"
	"github.com/ptonlix/whalehire/backend/domain"
	"github.com/ptonlix/whalehire/backend/pkg/eino/models"
)

// GeneralAgentUsecase 通用智能体用例
type GeneralAgentUsecase struct {
	modelFactory *models.ModelFactory
	config       *config.Config
}

// NewGeneralAgentUsecase 创建通用智能体用例
func NewGeneralAgentUsecase(config *config.Config) *GeneralAgentUsecase {
	factory := models.NewModelFactory()

	// 注册OpenAI模型
	openaiConfig := &models.OpenAIConfig{
		APIKey:  config.GeneralAgent.LLM.APIKey,
		BaseURL: config.GeneralAgent.LLM.BaseURL,
		Model:   config.GeneralAgent.LLM.ModelName,
	}
	openaiManager := models.NewOpenAIModelManager(openaiConfig)
	factory.Register(models.ModelTypeOpenAI, config.GeneralAgent.LLM.ModelName, openaiManager)

	return &GeneralAgentUsecase{
		modelFactory: factory,
		config:       config,
	}
}

// Generate 生成回复
func (uc *GeneralAgentUsecase) Generate(ctx context.Context, req *domain.GenerateReq) (*domain.GenerateResp, error) {
	llm, err := uc.modelFactory.GetModel(ctx, models.ModelTypeOpenAI, uc.config.GeneralAgent.LLM.ModelName)
	if err != nil {
		return nil, fmt.Errorf("failed to get model: %w", err)
	}

	messages := uc.buildMessages(req)
	result, err := llm.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("failed to generate: %w", err)
	}

	return &domain.GenerateResp{
		Answer: result.Content,
	}, nil
}

// GenerateStream 流式生成回复
func (uc *GeneralAgentUsecase) GenerateStream(ctx context.Context, req *domain.GenerateReq) (domain.StreamReader, error) {
	llm, err := uc.modelFactory.GetModel(ctx, models.ModelTypeOpenAI, uc.config.GeneralAgent.LLM.ModelName)
	if err != nil {
		return nil, fmt.Errorf("failed to get model: %w", err)
	}

	messages := uc.buildMessages(req)
	streamReader, err := llm.Stream(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("failed to stream: %w", err)
	}

	return &streamReaderAdapter{reader: streamReader}, nil
}

// buildMessages 构建消息列表
func (uc *GeneralAgentUsecase) buildMessages(req *domain.GenerateReq) []*schema.Message {
	var messages []*schema.Message

	// 添加历史消息
	for _, msg := range req.History {
		switch msg.Role {
		case "system":
			messages = append(messages, schema.SystemMessage(msg.Content))
		case "user":
			messages = append(messages, schema.UserMessage(msg.Content))
		case "assistant":
			messages = append(messages, schema.AssistantMessage(msg.Content, nil))
		}
	}

	// 添加当前用户消息
	messages = append(messages, schema.UserMessage(req.Prompt))

	return messages
}

// streamReaderAdapter 流读取器适配器
type streamReaderAdapter struct {
	reader *schema.StreamReader[*schema.Message]
}

func (s *streamReaderAdapter) Recv() (*domain.StreamChunk, error) {
	message, err := s.reader.Recv()
	if err == io.EOF {
		return &domain.StreamChunk{
			Content: "",
			Done:    true,
		}, nil
	}
	if err != nil {
		return nil, err
	}

	return &domain.StreamChunk{
		Content: message.Content,
		Done:    false,
	}, nil
}

func (s *streamReaderAdapter) Close() error {
	s.reader.Close()
	return nil
}
