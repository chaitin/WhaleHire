package repo

import (
	"context"
	"fmt"

	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/db"
	"github.com/chaitin/WhaleHire/backend/db/resumemailboxsetting"
	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/google/uuid"
)

// ResumeMailboxSettingRepo 简历邮箱设置仓储实现
type ResumeMailboxSettingRepo struct {
	client *db.Client
}

// NewResumeMailboxSettingRepo 创建简历邮箱设置仓储实例
func NewResumeMailboxSettingRepo(client *db.Client) domain.ResumeMailboxSettingRepo {
	return &ResumeMailboxSettingRepo{
		client: client,
	}
}

// Create 创建简历邮箱设置
func (r *ResumeMailboxSettingRepo) Create(ctx context.Context, req *domain.CreateResumeMailboxSettingRequest) (*db.ResumeMailboxSetting, error) {
	builder := r.client.ResumeMailboxSetting.Create().
		SetName(req.Name).
		SetEmailAddress(req.EmailAddress).
		SetProtocol(req.Protocol).
		SetHost(req.Host).
		SetPort(req.Port).
		SetUseSsl(req.UseSsl).
		SetAuthType(req.AuthType).
		SetEncryptedCredential(req.EncryptedCredential).
		SetUploaderID(req.UploaderID).
		SetStatus(req.Status)

	if req.Folder != nil {
		builder = builder.SetFolder(*req.Folder)
	}

	// 设置JobProfileIDs JSON数组
	if len(req.JobProfileIDs) > 0 {
		builder = builder.SetJobProfileIds(req.JobProfileIDs)
	}

	if req.SyncIntervalMinutes != nil {
		builder = builder.SetSyncIntervalMinutes(*req.SyncIntervalMinutes)
	}

	entity, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create resume mailbox setting: %w", err)
	}

	// 重新查询以获取关联数据
	result, err := r.client.ResumeMailboxSetting.Query().
		Where(resumemailboxsetting.ID(entity.ID)).
		WithUploader().
		First(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query created resume mailbox setting: %w", err)
	}

	return result, nil
}

// GetByID 根据ID获取简历邮箱设置
func (r *ResumeMailboxSettingRepo) GetByID(ctx context.Context, id uuid.UUID) (*db.ResumeMailboxSetting, error) {
	entity, err := r.client.ResumeMailboxSetting.Query().
		Where(resumemailboxsetting.ID(id)).
		WithUploader().
		First(ctx)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, domain.ErrResumeMailboxSettingNotFound
		}
		return nil, fmt.Errorf("failed to get resume mailbox setting: %w", err)
	}

	return entity, nil
}

// Update 更新简历邮箱设置
func (r *ResumeMailboxSettingRepo) Update(ctx context.Context, id uuid.UUID, req *domain.UpdateResumeMailboxSettingRequest) (*db.ResumeMailboxSetting, error) {
	builder := r.client.ResumeMailboxSetting.UpdateOneID(id)

	if req.Name != nil {
		builder = builder.SetName(*req.Name)
	}
	if req.EmailAddress != nil {
		builder = builder.SetEmailAddress(*req.EmailAddress)
	}
	if req.Protocol != nil {
		builder = builder.SetProtocol(*req.Protocol)
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
		builder = builder.SetAuthType(*req.AuthType)
	}
	if req.EncryptedCredential != nil {
		builder = builder.SetEncryptedCredential(*req.EncryptedCredential)
	}

	// 设置JobProfileIDs JSON数组
	if len(req.JobProfileIDs) > 0 {
		builder = builder.SetJobProfileIds(req.JobProfileIDs)
	}

	if req.SyncIntervalMinutes != nil {
		builder = builder.SetSyncIntervalMinutes(*req.SyncIntervalMinutes)
	}
	if req.Status != nil {
		builder = builder.SetStatus(*req.Status)
	}

	entity, err := builder.Save(ctx)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, domain.ErrResumeMailboxSettingNotFound
		}
		return nil, fmt.Errorf("failed to update resume mailbox setting: %w", err)
	}

	// 重新查询以获取关联数据
	result, err := r.client.ResumeMailboxSetting.Query().
		Where(resumemailboxsetting.ID(entity.ID)).
		WithUploader().
		First(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query updated resume mailbox setting: %w", err)
	}

	return result, nil
}

// Delete 删除简历邮箱设置
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

// List 获取简历邮箱设置列表
func (r *ResumeMailboxSettingRepo) List(ctx context.Context, req *domain.ListResumeMailboxSettingsRequest) ([]*db.ResumeMailboxSetting, *db.PageInfo, error) {
	query := r.client.ResumeMailboxSetting.Query().
		WithUploader()

	// 添加筛选条件
	if req.UploaderID != nil {
		query = query.Where(resumemailboxsetting.UploaderIDEQ(*req.UploaderID))
	}

	// 对于JobProfileIDs的筛选，由于现在是JSON数组，需要使用JSON查询
	// 这里暂时跳过JobProfileIDs的筛选，后续可以通过JSON查询实现
	// TODO: 实现JSON数组的查询筛选

	if req.Status != nil {
		query = query.Where(resumemailboxsetting.StatusEQ(*req.Status))
	}
	if req.Protocol != nil {
		query = query.Where(resumemailboxsetting.ProtocolEQ(*req.Protocol))
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

	entities, pageInfo, err := query.Page(ctx, page, size)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list resume mailbox settings: %w", err)
	}

	return entities, pageInfo, nil
}

// UpdateSyncStatus 更新同步状态
func (r *ResumeMailboxSettingRepo) UpdateSyncStatus(ctx context.Context, id uuid.UUID, req *domain.UpdateSyncStatusRequest) error {
	builder := r.client.ResumeMailboxSetting.UpdateOneID(id)

	if req.Status != nil {
		builder = builder.SetStatus(*req.Status)
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

// GetActiveSettings 获取所有启用的邮箱设置
func (r *ResumeMailboxSettingRepo) GetActiveSettings(ctx context.Context) ([]*db.ResumeMailboxSetting, error) {
	entities, err := r.client.ResumeMailboxSetting.Query().
		Where(resumemailboxsetting.StatusEQ(string(consts.MailboxStatusEnabled))).
		WithUploader().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active settings: %w", err)
	}

	return entities, nil
}
