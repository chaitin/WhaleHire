package domain

import (
	"context"
	"time"

	"github.com/GoYoko/web"
	"github.com/ptonlix/whalehire/backend/db"
)

type GeneralAgent interface {
	Generate(ctx context.Context, messages []*Message) (string, error)
	GenerateStream(ctx context.Context, messages []*Message) (StreamReader, error)
}

// GeneralAgentRepo 通用智能体仓储接口
type GeneralAgentRepo interface {
	SaveConversation(ctx context.Context, conversation *Conversation) error
	GetConversationHistory(ctx context.Context, userID string, limit int) ([]*Conversation, error)
	DeleteConversation(ctx context.Context, conversationID string) error
	ListConversations(ctx context.Context, userID string, page *web.Pagination) ([]*Conversation, *db.PageInfo, error)
	UpdateConversation(ctx context.Context, conversationID string, fn func(*db.Tx, *Conversation) error) error
}

// Conversation 对话记录
type Conversation struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Title     string     `json:"title"`
	Messages  []*Message `json:"messages"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type GenerateReq struct {
	Prompt  string     `json:"prompt"`
	History []*Message `json:"history,omitempty"`
}

type GenerateResp struct {
	Answer string `json:"answer"`
}

type Message struct {
	Role    string `json:"role"` // "user", "assistant", "system"
	Content string `json:"content"`
}

type StreamReader interface {
	Recv() (*StreamChunk, error)
	Close() error
}

type StreamChunk struct {
	Content string `json:"content"`
	Done    bool   `json:"done"`
}
