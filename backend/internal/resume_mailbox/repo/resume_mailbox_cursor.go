package repo

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/db"
	"github.com/chaitin/WhaleHire/backend/db/resumemailboxcursor"
	"github.com/chaitin/WhaleHire/backend/domain"
)

// ResumeMailboxCursorRepo 邮箱游标仓储实现
type ResumeMailboxCursorRepo struct {
	client *db.Client
}

// NewResumeMailboxCursorRepo 创建邮箱游标仓储实例
func NewResumeMailboxCursorRepo(client *db.Client) domain.ResumeMailboxCursorRepo {
	return &ResumeMailboxCursorRepo{
		client: client,
	}
}

// GetByMailboxID 根据邮箱ID获取游标
func (r *ResumeMailboxCursorRepo) GetByMailboxID(ctx context.Context, mailboxID uuid.UUID) (*domain.ResumeMailboxCursor, error) {
	entity, err := r.client.ResumeMailboxCursor.
		Query().
		Where(resumemailboxcursor.MailboxID(mailboxID)).
		Only(ctx)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get resume mailbox cursor: %w", err)
	}

	cursor := &domain.ResumeMailboxCursor{}
	return cursor.From(entity), nil
}

// Upsert 更新或创建游标
func (r *ResumeMailboxCursorRepo) Upsert(ctx context.Context, mailboxID uuid.UUID, protocolCursor string, lastMessageID string) (*domain.ResumeMailboxCursor, error) {
	// 先尝试查询现有游标
	entity, err := r.client.ResumeMailboxCursor.
		Query().
		Where(resumemailboxcursor.MailboxID(mailboxID)).
		Only(ctx)
	if err != nil && !db.IsNotFound(err) {
		return nil, fmt.Errorf("failed to query resume mailbox cursor: %w", err)
	}

	if entity != nil {
		entity, err = r.client.ResumeMailboxCursor.
			UpdateOneID(entity.ID).
			SetProtocolCursor(protocolCursor).
			SetLastMessageID(lastMessageID).
			Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to update resume mailbox cursor: %w", err)
		}
	} else {
		entity, err = r.client.ResumeMailboxCursor.
			Create().
			SetMailboxID(mailboxID).
			SetProtocolCursor(protocolCursor).
			SetLastMessageID(lastMessageID).
			Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create resume mailbox cursor: %w", err)
		}
	}

	cursor := &domain.ResumeMailboxCursor{}
	return cursor.From(entity), nil
}
