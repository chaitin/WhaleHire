package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/google/uuid"
)

// ResumeMailboxSettingUsecase 邮箱设置用例实现
type ResumeMailboxSettingUsecase struct {
	repo              domain.ResumeMailboxSettingRepo
	credentialVault   domain.CredentialVault
	adapterFactory    domain.MailboxAdapterFactory
	scheduler         domain.ResumeMailboxScheduler
	jobProfileUsecase domain.JobProfileUsecase
}

// NewResumeMailboxSettingUsecase 创建邮箱设置用例实例
func NewResumeMailboxSettingUsecase(
	repo domain.ResumeMailboxSettingRepo,
	credentialVault domain.CredentialVault,
	adapterFactory domain.MailboxAdapterFactory,
	scheduler domain.ResumeMailboxScheduler,
	jobProfileUsecase domain.JobProfileUsecase,
) domain.ResumeMailboxSettingUsecase {
	return &ResumeMailboxSettingUsecase{
		repo:              repo,
		credentialVault:   credentialVault,
		adapterFactory:    adapterFactory,
		scheduler:         scheduler,
		jobProfileUsecase: jobProfileUsecase,
	}
}

// Create 创建邮箱设置
func (u *ResumeMailboxSettingUsecase) Create(ctx context.Context, req *domain.CreateResumeMailboxSettingRequest) (*domain.ResumeMailboxSetting, error) {
	// 加密凭证信息
	encryptedCredential, err := u.credentialVault.Encrypt(ctx, req.EncryptedCredential)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt credential: %w", err)
	}

	// 创建新的请求对象，使用加密后的凭证
	createReq := *req
	createReq.EncryptedCredential = encryptedCredential

	// 调用仓储层创建
	dbSetting, err := u.repo.Create(ctx, &createReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create resume mailbox setting: %w", err)
	}

	// 使用 From 方法转换为 domain 类型
	setting := &domain.ResumeMailboxSetting{}
	setting.From(dbSetting)

	if u.scheduler != nil {
		if err := u.scheduler.Upsert(ctx, setting); err != nil {
			return nil, fmt.Errorf("注册邮箱同步任务失败: %w", err)
		}
	}

	return setting, nil
}

// GetByID 获取邮箱设置详情
func (u *ResumeMailboxSettingUsecase) GetByID(ctx context.Context, id uuid.UUID) (*domain.ResumeMailboxSetting, error) {
	dbSetting, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 使用 From 方法转换为 domain 类型
	setting := &domain.ResumeMailboxSetting{}
	setting.From(dbSetting)

	// 填充JobProfileNames
	if err := u.fillJobProfileNames(ctx, setting); err != nil {
		return nil, err
	}

	// 解密凭证信息（仅在需要时）
	if len(setting.EncryptedCredential) > 0 {
		decryptedCredential, err := u.credentialVault.Decrypt(ctx, setting.EncryptedCredential)
		if err != nil {
			// 解密失败时记录错误但不阻断返回
			// 在实际使用中可能需要根据业务需求决定是否返回错误
			setting.EncryptedCredential = map[string]interface{}{
				"error": "failed to decrypt credential",
			}
		} else {
			setting.EncryptedCredential = decryptedCredential
		}
	}

	return setting, nil
}

// fillJobProfileNames 填充JobProfileNames字段
func (u *ResumeMailboxSettingUsecase) fillJobProfileNames(ctx context.Context, setting *domain.ResumeMailboxSetting) error {
	if len(setting.JobProfileIDs) == 0 {
		setting.JobProfileNames = []string{}
		return nil
	}

	// 将UUID转换为字符串
	ids := make([]string, len(setting.JobProfileIDs))
	for i, id := range setting.JobProfileIDs {
		ids[i] = id.String()
	}

	// 批量获取JobProfile
	profiles, err := u.jobProfileUsecase.GetByIDs(ctx, ids)
	if err != nil {
		return fmt.Errorf("failed to get job profiles: %w", err)
	}

	// 提取名称
	names := make([]string, len(profiles))
	for i, profile := range profiles {
		names[i] = profile.Name
	}

	setting.JobProfileNames = names
	return nil
}

// Update 更新邮箱设置
func (u *ResumeMailboxSettingUsecase) Update(ctx context.Context, id uuid.UUID, req *domain.UpdateResumeMailboxSettingRequest) (*domain.ResumeMailboxSetting, error) {
	// 如果需要更新凭证，先加密
	updateReq := *req
	if req.EncryptedCredential != nil {
		encryptedCredential, err := u.credentialVault.Encrypt(ctx, *req.EncryptedCredential)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt credential: %w", err)
		}
		updateReq.EncryptedCredential = &encryptedCredential
	}

	// 调用仓储层更新
	dbSetting, err := u.repo.Update(ctx, id, &updateReq)
	if err != nil {
		return nil, fmt.Errorf("failed to update resume mailbox setting: %w", err)
	}

	// 使用 From 方法转换为 domain 类型
	setting := &domain.ResumeMailboxSetting{}
	setting.From(dbSetting)

	if u.scheduler != nil {
		if err := u.scheduler.Upsert(ctx, setting); err != nil {
			return nil, fmt.Errorf("更新邮箱调度任务失败: %w", err)
		}
	}

	return setting, nil
}

// Delete 删除邮箱设置
func (u *ResumeMailboxSettingUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	err := u.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete resume mailbox setting: %w", err)
	}

	if u.scheduler != nil {
		if err := u.scheduler.Remove(ctx, id); err != nil {
			return fmt.Errorf("移除邮箱调度任务失败: %w", err)
		}
	}

	return nil
}

// List 获取邮箱设置列表
func (u *ResumeMailboxSettingUsecase) List(ctx context.Context, req *domain.ListResumeMailboxSettingsRequest) (*domain.ListResumeMailboxSettingsResponse, error) {
	// 设置默认分页参数
	if req.Size <= 0 {
		req.Size = 20
	}
	if req.Size > 100 {
		req.Size = 100
	}

	dbSettings, pageInfo, err := u.repo.List(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list resume mailbox settings: %w", err)
	}

	// 转换为 domain 类型
	items := make([]*domain.ResumeMailboxSetting, len(dbSettings))
	for i, dbSetting := range dbSettings {
		setting := &domain.ResumeMailboxSetting{}
		setting.From(dbSetting)

		// 填充JobProfileNames
		if err := u.fillJobProfileNames(ctx, setting); err != nil {
			// 记录错误但不阻断返回，JobProfileNames将保持为空
			_ = err // 明确忽略错误
		}

		items[i] = setting
	}

	// 对于列表查询，不解密凭证信息以提高性能
	// 如果需要凭证信息，应该调用 GetByID 方法

	return &domain.ListResumeMailboxSettingsResponse{
		Items:    items,
		PageInfo: pageInfo,
	}, nil
}

// TestConnection 测试邮箱连接
func (u *ResumeMailboxSettingUsecase) TestConnection(ctx context.Context, req *domain.TestConnectionRequest) error {
	// 检查凭证是否已经是解密后的格式
	credential := req.EncryptedCredential

	// 如果凭证包含 encrypted_data 和 algorithm 字段，说明还需要解密
	if _, hasEncryptedData := credential["encrypted_data"].(string); hasEncryptedData {
		if algorithm, hasAlgorithm := credential["algorithm"].(string); hasAlgorithm && algorithm == "AES-256-GCM" {
			// 需要解密
			decryptedCredential, err := u.credentialVault.Decrypt(ctx, credential)
			if err != nil {
				return fmt.Errorf("failed to decrypt credential: %w", err)
			}
			credential = decryptedCredential
		}
	}

	// 验证必要的凭证字段
	if req.AuthType == string(consts.MailboxAuthTypePassword) {
		if _, ok := credential["password"]; !ok {
			return domain.ErrInvalidCredentials
		}
	} else if req.AuthType == string(consts.MailboxAuthTypeOAuth) {
		if _, ok := credential["access_token"]; !ok {
			return domain.ErrInvalidCredentials
		}
	}

	adapter, err := u.adapterFactory.GetAdapter(strings.ToLower(req.Protocol))
	if err != nil {
		return err
	}

	folder := ""
	if req.Folder != nil {
		folder = *req.Folder
	}

	config := &domain.MailboxConnectionConfig{
		Host:         req.Host,
		Port:         req.Port,
		UseSSL:       req.UseSsl,
		EmailAddress: req.EmailAddress,
		Folder:       folder,
		AuthType:     req.AuthType,
		Credential:   credential,
	}

	if err := adapter.TestConnection(ctx, config); err != nil {
		return fmt.Errorf("邮箱连接失败: %w", err)
	}

	return nil
}

// UpdateStatus 启用/禁用邮箱设置
func (u *ResumeMailboxSettingUsecase) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	// 验证状态值
	if status != string(consts.MailboxStatusEnabled) && status != string(consts.MailboxStatusDisabled) {
		return fmt.Errorf("invalid status: %s", status)
	}

	updateReq := &domain.UpdateResumeMailboxSettingRequest{
		Status: &status,
	}

	entity, err := u.repo.Update(ctx, id, updateReq)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	if u.scheduler != nil {
		setting := (&domain.ResumeMailboxSetting{}).From(entity)
		if err := u.scheduler.Upsert(ctx, setting); err != nil {
			return fmt.Errorf("更新邮箱调度任务失败: %w", err)
		}
	}

	return nil
}
