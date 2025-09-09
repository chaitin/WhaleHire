package usecase

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/cloudwego/eino/schema"

	"github.com/ptonlix/whalehire/backend/config"
	"github.com/ptonlix/whalehire/backend/db"
	"github.com/ptonlix/whalehire/backend/domain"
	"github.com/ptonlix/whalehire/backend/pkg/eino/models"
)

// GeneralAgentUsecase 通用智能体用例
type GeneralAgentUsecase struct {
	modelFactory *models.ModelFactory
	config       *config.Config
	repo         domain.GeneralAgentRepo
}

// NewGeneralAgentUsecase 创建通用智能体用例
func NewGeneralAgentUsecase(config *config.Config, repo domain.GeneralAgentRepo) *GeneralAgentUsecase {
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
		repo:         repo,
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

// CreateConversation 创建新对话
func (uc *GeneralAgentUsecase) CreateConversation(ctx context.Context, userID string, req *domain.CreateConversationReq) (*domain.Conversation, error) {
	conv := &domain.Conversation{
		UserID:    userID,
		Title:     req.Title,
		Status:    "active",
		Messages:  []*domain.Message{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := uc.repo.SaveConversation(ctx, conv); err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	return conv, nil
}

// GetConversationHistory 获取对话历史
func (uc *GeneralAgentUsecase) GetConversationHistory(ctx context.Context, req *domain.GetConversationHistoryReq) (*domain.Conversation, error) {
	dbConv, err := uc.repo.GetConversationHistory(ctx, req.ConversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation history: %w", err)
	}

	// 转换为 domain 对象
	conv := &domain.Conversation{
		ID:        dbConv.ID.String(),
		UserID:    dbConv.UserID.String(),
		Title:     dbConv.Title,
		Metadata:  dbConv.Metadata,
		Status:    dbConv.Status,
		CreatedAt: dbConv.CreatedAt,
		UpdatedAt: dbConv.UpdatedAt,
	}

	if !dbConv.DeletedAt.IsZero() {
		conv.DeletedAt = &dbConv.DeletedAt
	}

	// 转换消息
	if dbConv.Edges.Messages != nil {
		conv.Messages = make([]*domain.Message, len(dbConv.Edges.Messages))
		for j, dbMsg := range dbConv.Edges.Messages {
			msg := &domain.Message{
				ID:             dbMsg.ID.String(),
				ConversationID: dbMsg.ConversationID.String(),
				Role:           dbMsg.Role,
				Type:           dbMsg.Type,
				CreatedAt:      dbMsg.CreatedAt,
				UpdatedAt:      dbMsg.UpdatedAt,
			}

			if dbMsg.Content != "" {
				msg.Content = &dbMsg.Content
			}
			if dbMsg.AgentName != "" {
				msg.AgentName = &dbMsg.AgentName
			}
			if dbMsg.MediaURL != "" {
				msg.MediaURL = &dbMsg.MediaURL
			}
			if !dbMsg.DeletedAt.IsZero() {
				msg.DeletedAt = &dbMsg.DeletedAt
			}

			conv.Messages[j] = msg
		}
	}

	return conv, nil
}

// ListConversations 分页获取对话列表
func (uc *GeneralAgentUsecase) ListConversations(ctx context.Context, userID string, req *domain.ListConversationsReq) ([]*domain.ListConversationsResp, error) {
	dbConversations, pageInfo, err := uc.repo.ListConversations(ctx, userID, &req.Pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to list conversations: %w", err)
	}

	// 转换为 domain 对象
	conversations := make([]*domain.Conversation, len(dbConversations))
	for i, dbConv := range dbConversations {
		conv := &domain.Conversation{
			ID:        dbConv.ID.String(),
			UserID:    dbConv.UserID.String(),
			Title:     dbConv.Title,
			Metadata:  dbConv.Metadata,
			Status:    dbConv.Status,
			CreatedAt: dbConv.CreatedAt,
			UpdatedAt: dbConv.UpdatedAt,
		}

		if !dbConv.DeletedAt.IsZero() {
			conv.DeletedAt = &dbConv.DeletedAt
		}

		conversations[i] = conv
	}

	// 构造响应
	resp := &domain.ListConversationsResp{
		PageInfo:      pageInfo,
		Conversations: conversations,
	}

	return []*domain.ListConversationsResp{resp}, nil
}

// DeleteConversation 删除对话
func (uc *GeneralAgentUsecase) DeleteConversation(ctx context.Context, req *domain.DeleteConversationReq) error {
	if err := uc.repo.DeleteConversation(ctx, req.ConversationID); err != nil {
		return fmt.Errorf("failed to delete conversation: %w", err)
	}
	return nil
}

// AddMessageToConversation 向对话添加消息
func (uc *GeneralAgentUsecase) AddMessageToConversation(ctx context.Context, req *domain.AddMessageToConversationReq) error {
	return uc.repo.UpdateConversation(ctx, req.ConversationID, func(tx *db.Tx, conv *domain.Conversation) error {
		req.Message.ConversationID = req.ConversationID
		req.Message.CreatedAt = time.Now()
		req.Message.UpdatedAt = time.Now()
		conv.Messages = append(conv.Messages, req.Message)
		conv.UpdatedAt = time.Now()
		return nil
	})
}

// buildMessages 构建消息列表
func (uc *GeneralAgentUsecase) buildMessages(req *domain.GenerateReq) []*schema.Message {
	var messages []*schema.Message

	// 添加历史消息
	for _, msg := range req.History {
		if msg.Content == nil {
			continue
		}
		switch msg.Role {
		case "system":
			messages = append(messages, schema.SystemMessage(*msg.Content))
		case "user":
			messages = append(messages, schema.UserMessage(*msg.Content))
		case "assistant":
			messages = append(messages, schema.AssistantMessage(*msg.Content, nil))
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
