package repo

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/chaitin/WhaleHire/backend/pkg/web"
	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/db"
	"github.com/chaitin/WhaleHire/backend/db/conversation"
	"github.com/chaitin/WhaleHire/backend/db/message"
	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/pkg/entx"
)

type GeneralAgentRepo struct {
	db *db.Client
}

func NewGeneralAgentRepo(db *db.Client) domain.GeneralAgentRepo {
	return &GeneralAgentRepo{db: db}
}

func (r *GeneralAgentRepo) SaveConversation(ctx context.Context, conv *domain.Conversation) error {
	return entx.WithTx(ctx, r.db, func(tx *db.Tx) error {
		// 创建或更新对话
		var conversationID uuid.UUID
		var err error

		if conv.ID == "" {
			// 创建新对话
			userID, err := uuid.Parse(conv.UserID)
			if err != nil {
				return err
			}

			create := tx.Conversation.Create().
				SetUserID(userID).
				SetTitle(conv.Title).
				SetStatus(conv.Status)

			if conv.AgentName != "" {
				create = create.SetAgentName(conv.AgentName)
			}

			if conv.Metadata != nil {
				create = create.SetMetadata(conv.Metadata)
			}

			dbConv, err := create.Save(ctx)
			if err != nil {
				return err
			}
			conversationID = dbConv.ID
			conv.ID = conversationID.String()
		} else {
			// 更新现有对话
			conversationID, err = uuid.Parse(conv.ID)
			if err != nil {
				return err
			}

			update := tx.Conversation.UpdateOneID(conversationID).
				SetTitle(conv.Title).
				SetStatus(conv.Status)

			if conv.AgentName != "" {
				update = update.SetAgentName(conv.AgentName)
			}

			if conv.Metadata != nil {
				update = update.SetMetadata(conv.Metadata)
			}

			if err := update.Exec(ctx); err != nil {
				return err
			}
		}

		// 保存消息
		for _, msg := range conv.Messages {
			if msg.ID == "" {
				// 创建新消息
				create := tx.Message.Create().
					SetConversationID(conversationID).
					SetRole(msg.Role).
					SetType(msg.Type)

				if msg.Content != nil {
					create = create.SetContent(*msg.Content)
				}
				if msg.AgentName != nil {
					create = create.SetAgentName(*msg.AgentName)
				}
				if msg.MediaURL != nil {
					create = create.SetMediaURL(*msg.MediaURL)
				}

				dbMsg, err := create.Save(ctx)
				if err != nil {
					return err
				}
				msg.ID = dbMsg.ID.String()
				msg.ConversationID = conversationID.String()
			}
		}

		return nil
	})
}

func (r *GeneralAgentRepo) GetConversationHistory(ctx context.Context, conversationID string) (*db.Conversation, error) {
	cid, err := uuid.Parse(conversationID)
	if err != nil {
		return nil, err
	}

	dbConv, err := r.db.Conversation.Query().
		Where(conversation.ID(cid)).
		WithMessages(func(q *db.MessageQuery) {
			q.Order(message.ByCreatedAt(sql.OrderAsc()))
		}).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	return dbConv, nil
}

func (r *GeneralAgentRepo) DeleteConversation(ctx context.Context, conversationID string) error {
	cid, err := uuid.Parse(conversationID)
	if err != nil {
		return err
	}

	return entx.WithTx(ctx, r.db, func(tx *db.Tx) error {
		// 软删除相关消息
		if err := tx.Message.Update().
			Where(message.ConversationID(cid)).
			SetDeletedAt(time.Now()).
			Exec(ctx); err != nil {
			return err
		}

		// 软删除对话
		return tx.Conversation.UpdateOneID(cid).
			SetDeletedAt(time.Now()).
			Exec(ctx)
	})
}

func (r *GeneralAgentRepo) ListConversations(ctx context.Context, userID string, page *web.Pagination) ([]*db.Conversation, *db.PageInfo, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, nil, err
	}

	q := r.db.Conversation.Query().
		Where(conversation.UserID(uid)).
		Order(conversation.ByUpdatedAt(sql.OrderDesc()))

	dbConversations, pageInfo, err := q.Page(ctx, page.Page, page.Size)
	if err != nil {
		return nil, nil, err
	}

	return dbConversations, pageInfo, nil
}

func (r *GeneralAgentRepo) UpdateConversation(ctx context.Context, conversationID string, fn func(*db.Tx, *db.ConversationUpdateOne) error) error {
	cid, err := uuid.Parse(conversationID)
	if err != nil {
		return err
	}

	return entx.WithTx(ctx, r.db, func(tx *db.Tx) error {
		// 获取更新器
		update := tx.Conversation.UpdateOneID(cid)

		// 执行业务逻辑
		if err := fn(tx, update); err != nil {
			return err
		}

		// 执行更新
		return update.Exec(ctx)
	})
}
