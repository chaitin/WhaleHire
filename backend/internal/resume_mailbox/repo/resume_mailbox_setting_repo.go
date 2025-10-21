package repo

import (
	"context"
	"fmt"

	"github.com/chaitin/WhaleHire/backend/db"
	"github.com/chaitin/WhaleHire/backend/db/resumemailboxsetting"
	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/google/uuid"
)

// ResumeMailboxSettingRepo 邮箱设置仓储实现
type ResumeMailboxSettingRepo struct {
	client *db.Client
}

// NewResumeMailboxSettingRepo 创建邮箱设置仓储实例
func NewResumeMailboxSettingRepo(client *db.Client) domain.ResumeMailboxSettingRepo {
	return &ResumeMailboxSettingRepo{
		client: client,
	}
}

// Create 创建邮箱设置
func (r *ResumeMailboxSettingRepo) Create(ctx context.Context, req *domain.CreateResumeMailboxSettingRequest) (*db.ResumeMailboxSetting, error) {
	builder := r.client.ResumeMailboxSetting.Create().
		SetName(req.Name).
		SetEmailAddress(req.EmailAddress).
		SetProtocol(resumemailboxsetting.Protocol(req.Protocol)).
		SetHost(req.Host).
		SetPort(req.Port).
		SetUseSsl(req.UseSsl).
		SetAuthType(resumemailboxsetting.AuthType(req.AuthType)).
		SetEncryptedCredential(req.EncryptedCredential).
		SetUploaderID(req.UploaderID).
		SetStatus(resumemailboxsetting.StatusEnabled)

	if req.Folder != nil {
		builder = builder.SetFolder(*req.Folder)
	}

	if req.JobProfileID != nil {
		builder = builder.SetJobProfileID(*req.JobProfileID)
	}

	if req.SyncIntervalMinutes != nil {
		builder = builder.SetSyncIntervalMinutes(*req.SyncIntervalMinutes)
	}

	entity, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create resume mailbox setting: %w", err)
	}

	// 加载关联数据
	entity, err = r.client.ResumeMailboxSetting.Query().
		Where(resumemailboxsetting.ID(entity.ID)).
		WithUploader().
		WithJobProfile().
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load created resume mailbox setting: %w", err)
	}

	return entity, nil
}

// GetByID 根据ID获取邮箱设置
func (r *ResumeMailboxSettingRepo) GetByID(ctx context.Context, id uuid.UUID) (*db.ResumeMailboxSetting, error) {
	entity, err := r.client.ResumeMailboxSetting.
		Query().
		Where(resumemailboxsetting.ID(id)).
		WithUploader().
		WithJobProfile().
		Only(ctx)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, domain.ErrResumeMailboxSettingNotFound
		}
		return nil, fmt.Errorf("failed to get resume mailbox setting: %w", err)
	}

	return entity, nil
}

// Update 更新邮箱设置
func (r *ResumeMailboxSettingRepo) Update(ctx context.Context, id uuid.UUID, req *domain.UpdateResumeMailboxSettingRequest) (*db.ResumeMailboxSetting, error) {
	builder := r.client.ResumeMailboxSetting.UpdateOneID(id)

	if req.Name != nil {
		builder = builder.SetName(*req.Name)
	}
	if req.EmailAddress != nil {
		builder = builder.SetEmailAddress(*req.EmailAddress)
	}
	if req.Protocol != nil {
		builder = builder.SetProtocol(resumemailboxsetting.Protocol(*req.Protocol))
	}
	if req.Host != nil {
		builder = builder.SetHost(*req.Host)
	}
	if req.Port != nil {
		builder = builder.SetPort(*req.Port)
	}
	if req.UseSsl != nil {
		builder = builder.SetUseSsl(*req.UseSsl)
	}
	if req.Folder != nil {
		builder = builder.SetFolder(*req.Folder)
	}
	if req.AuthType != nil {
		builder = builder.SetAuthType(resumemailboxsetting.AuthType(*req.AuthType))
	}
	if req.EncryptedCredential != nil {
		builder = builder.SetEncryptedCredential(*req.EncryptedCredential)
	}
	if req.JobProfileID != nil {
		builder = builder.SetJobProfileID(*req.JobProfileID)
	}
	if req.SyncIntervalMinutes != nil {
		builder = builder.SetSyncIntervalMinutes(*req.SyncIntervalMinutes)
	}
	if req.Status != nil {
		builder = builder.SetStatus(resumemailboxsetting.Status(*req.Status))
	}

	entity, err := builder.Save(ctx)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, domain.ErrResumeMailboxSettingNotFound
		}
		return nil, fmt.Errorf("failed to update resume mailbox setting: %w", err)
	}

	// 加载关联数据
	entity, err = r.client.ResumeMailboxSetting.Query().
		Where(resumemailboxsetting.ID(entity.ID)).
		WithUploader().
		WithJobProfile().
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load updated resume mailbox setting: %w", err)
	}

	return entity, nil
}

// Delete 删除邮箱设置
func (r *ResumeMailboxSettingRepo) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.client.ResumeMailboxSetting.DeleteOneID(id).Exec(ctx)
	if err != nil {
		if db.IsNotFound(err) {
			return domain.ErrResumeMailboxSettingNotFound
		}
		return fmt.Errorf("failed to delete resume mailbox setting: %w", err)
	}
	return nil
}

// List 获取邮箱设置列表
func (r *ResumeMailboxSettingRepo) List(ctx context.Context, req *domain.ListResumeMailboxSettingsRequest) ([]*db.ResumeMailboxSetting, *db.PageInfo, error) {
	query := r.client.ResumeMailboxSetting.Query().
		WithUploader().
		WithJobProfile()

	// 添加过滤条件
	if req.UploaderID != nil {
		query = query.Where(resumemailboxsetting.UploaderID(*req.UploaderID))
	}
	if req.JobProfileID != nil {
		query = query.Where(resumemailboxsetting.JobProfileID(*req.JobProfileID))
	}
	if req.Status != nil {
		query = query.Where(resumemailboxsetting.StatusEQ(resumemailboxsetting.Status(*req.Status)))
	}
	if req.Protocol != nil {
		query = query.Where(resumemailboxsetting.ProtocolEQ(resumemailboxsetting.Protocol(*req.Protocol)))
	}

	// 分页查询
	page := req.Page
	if page <= 0 {
		page = 1
	}
	size := req.Size
	if size <= 0 {
		size = 10
	}

	entities, pageInfo, err := query.
		Order(db.Desc(resumemailboxsetting.FieldCreatedAt)).
		Page(ctx, page, size)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list resume mailbox settings: %w", err)
	}

	return entities, pageInfo, nil
}

// UpdateSyncStatus 更新同步状态
func (r *ResumeMailboxSettingRepo) UpdateSyncStatus(ctx context.Context, id uuid.UUID, req *domain.UpdateSyncStatusRequest) error {
	builder := r.client.ResumeMailboxSetting.UpdateOneID(id)

	if req.Status != nil {
		builder = builder.SetStatus(resumemailboxsetting.Status(*req.Status))
	}
	if req.LastSyncedAt != nil {
		builder = builder.SetLastSyncedAt(*req.LastSyncedAt)
	}
	if req.LastError != nil {
		builder = builder.SetLastError(*req.LastError)
	}
	if req.RetryCount != nil {
		builder = builder.SetRetryCount(*req.RetryCount)
	}

	err := builder.Exec(ctx)
	if err != nil {
		if db.IsNotFound(err) {
			return domain.ErrResumeMailboxSettingNotFound
		}
		return fmt.Errorf("failed to update sync status: %w", err)
	}

	return nil
}

// GetActiveSettings 获取活跃的邮箱设置
func (r *ResumeMailboxSettingRepo) GetActiveSettings(ctx context.Context) ([]*db.ResumeMailboxSetting, error) {
	entities, err := r.client.ResumeMailboxSetting.Query().
		Where(resumemailboxsetting.StatusEQ(resumemailboxsetting.StatusEnabled)).
		WithUploader().
		WithJobProfile().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active resume mailbox settings: %w", err)
	}

	return entities, nil
}
