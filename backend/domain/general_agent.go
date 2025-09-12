package domain

import (
	"context"
	"time"

	"github.com/GoYoko/web"
	"github.com/ptonlix/whalehire/backend/db"
)

// GenerateAgentUsecase 通用智能体用例接口
type GeneralAgentUsecase interface {
	Generate(ctx context.Context, req *GenerateReq) (*GenerateResp, error)
	GenerateStream(ctx context.Context, req *GenerateReq) (StreamReader, error)
	CreateConversation(ctx context.Context, userID string, req *CreateConversationReq) (*Conversation, error)
	ListConversations(ctx context.Context, userID string, req *ListConversationsReq) (*ListConversationsResp, error)
	DeleteConversation(ctx context.Context, req *DeleteConversationReq) error
	AddMessageToConversation(ctx context.Context, req *AddMessageToConversationReq) error
	GetConversationHistory(ctx context.Context, req *GetConversationHistoryReq) (*Conversation, error)
}

// GeneralAgentRepo 通用智能体仓储接口
type GeneralAgentRepo interface {
	SaveConversation(ctx context.Context, conversation *Conversation) error
	GetConversationHistory(ctx context.Context, conversationID string) (*db.Conversation, error)
	DeleteConversation(ctx context.Context, conversationID string) error
	ListConversations(ctx context.Context, userID string, page *web.Pagination) ([]*db.Conversation, *db.PageInfo, error)
	UpdateConversation(ctx context.Context, conversationID string, fn func(*db.Tx, *db.ConversationUpdateOne) error) error
}

// CreateConversationReq 创建对话请求
type CreateConversationReq struct {
	Title string `json:"title" validate:"required,max=256"`
}

// Conversation 对话记录
type Conversation struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	Title     string                 `json:"title"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Status    string                 `json:"status"`
	Messages  []*Message             `json:"messages"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	DeletedAt *time.Time             `json:"deleted_at,omitempty"`
}

type GenerateReq struct {
	Prompt         string     `json:"prompt"`
	History        []*Message `json:"history,omitempty"`
	ConversationID *string    `json:"conversation_id,omitempty"`
}

type GenerateResp struct {
	Answer string `json:"answer"`
}

type Message struct {
	ID             string                 `json:"id"`
	ConversationID string                 `json:"conversation_id"`
	Role           string                 `json:"role"` // "user", "assistant", "system", "agent"
	Content        *string                `json:"content,omitempty"`
	AgentName      *string                `json:"agent_name,omitempty"`
	Type           string                 `json:"type"` // "text", "image", "audio", "video", "file"
	MediaURL       *string                `json:"media_url,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Attachments    []*Attachment          `json:"attachments,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	DeletedAt      *time.Time             `json:"deleted_at,omitempty"`
}

// Attachment 附件
type Attachment struct {
	ID        string                 `json:"id"`
	MessageID string                 `json:"message_id"`
	FileName  string                 `json:"file_name"`
	FileSize  int64                  `json:"file_size"`
	FileType  string                 `json:"file_type"`
	FileURL   string                 `json:"file_url"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	DeletedAt *time.Time             `json:"deleted_at,omitempty"`
}

type StreamReader interface {
	Recv() (*StreamChunk, error)
	Close() error
}

type StreamChunk struct {
	Content string `json:"content"`
	Done    bool   `json:"done"`
}

type StreamMetadata struct {
	Version        string `json:"version"`
	ConversationID string `json:"conversation_id"`
}

// ListConversationsReq 获取对话列表请求
type ListConversationsReq struct {
	web.Pagination

	Search string `json:"search" query:"search"` // 搜索
}

// ListConversationsResp 对话列表响应
type ListConversationsResp struct {
	*db.PageInfo

	Conversations []*Conversation `json:"conversations"`
}

// GetConversationHistoryReq 获取对话历史请求
type GetConversationHistoryReq struct {
	ConversationID string `json:"conversation_id"  validate:"required"`
}

// DeleteConversationReq 删除对话请求
type DeleteConversationReq struct {
	ConversationID string `json:"conversation_id"  validate:"required"`
}

// AddMessageToConversationReq 向对话添加消息请求
type AddMessageToConversationReq struct {
	ConversationID string   `json:"conversation_id" validate:"required"`
	Message        *Message `json:"message" validate:"required"`
}
