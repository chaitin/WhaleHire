package usecase

import (
	"context"
	"fmt"

	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/google/uuid"
)

// ResumeMailboxSettingUsecase 邮箱设置用例实现
type ResumeMailboxSettingUsecase struct {
	repo            domain.ResumeMailboxSettingRepo
	credentialVault domain.CredentialVault
}

// NewResumeMailboxSettingUsecase 创建邮箱设置用例实例
func NewResumeMailboxSettingUsecase(
	repo domain.ResumeMailboxSettingRepo,
	credentialVault domain.CredentialVault,
) domain.ResumeMailboxSettingUsecase {
	return &ResumeMailboxSettingUsecase{
		repo:            repo,
		credentialVault: credentialVault,
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

	return setting, nil
}

// Delete 删除邮箱设置
func (u *ResumeMailboxSettingUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	err := u.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete resume mailbox setting: %w", err)
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
	// 解密凭证信息
	decryptedCredential, err := u.credentialVault.Decrypt(ctx, req.EncryptedCredential)
	if err != nil {
		return fmt.Errorf("failed to decrypt credential: %w", err)
	}

	// 这里应该实现实际的邮箱连接测试逻辑
	// 由于涉及到具体的邮箱协议实现，这里先返回成功
	// 在实际项目中，需要根据 Protocol 类型调用相应的连接测试逻辑

	// 验证必要的凭证字段
	if req.AuthType == "password" {
		if _, ok := decryptedCredential["password"]; !ok {
			return domain.ErrInvalidCredentials
		}
	} else if req.AuthType == "oauth" {
		if _, ok := decryptedCredential["access_token"]; !ok {
			return domain.ErrInvalidCredentials
		}
	}

	// TODO: 实现实际的邮箱连接测试
	// 1. 根据 Protocol (imap/pop3) 创建相应的客户端
	// 2. 使用解密后的凭证进行连接测试
	// 3. 如果是 IMAP，还需要测试指定的 Folder 是否存在

	return nil
}

// UpdateStatus 启用/禁用邮箱设置
func (u *ResumeMailboxSettingUsecase) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	// 验证状态值
	if status != "enabled" && status != "disabled" {
		return fmt.Errorf("invalid status: %s", status)
	}

	updateReq := &domain.UpdateResumeMailboxSettingRequest{
		Status: &status,
	}

	_, err := u.repo.Update(ctx, id, updateReq)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	return nil
}
