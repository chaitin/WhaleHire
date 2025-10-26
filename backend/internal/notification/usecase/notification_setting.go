package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/domain"
)

type notificationSettingUsecase struct {
	repo   domain.NotificationSettingRepo
	logger *slog.Logger
}

// NewNotificationSettingUsecase 创建通知设置业务逻辑实例
func NewNotificationSettingUsecase(
	repo domain.NotificationSettingRepo,
	logger *slog.Logger,
) domain.NotificationSettingUsecase {
	return &notificationSettingUsecase{
		repo:   repo,
		logger: logger,
	}
}

func (u *notificationSettingUsecase) CreateSetting(ctx context.Context, setting *domain.NotificationSetting) (*domain.NotificationSetting, error) {
	// 验证配置
	if err := u.ValidateConfig(ctx, setting); err != nil {
		u.logger.Error("validate notification setting config failed", "error", err)
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	// 检查同一名称和通道的配置是否已存在
	existingEntity, err := u.repo.GetByNameAndChannel(ctx, setting.Name, setting.Channel)
	if err == nil && existingEntity != nil {
		return nil, fmt.Errorf("名称为 %s 的 %s 通道配置已存在", setting.Name, setting.Channel)
	}

	// 创建设置，使用数据库生成的ID
	createdEntity, err := u.repo.Create(ctx, setting)
	if err != nil {
		u.logger.Error("create notification setting failed", "error", err)
		return nil, fmt.Errorf("创建通知设置失败: %w", err)
	}

	// 使用From方法将数据库实体转换为领域对象
	result := &domain.NotificationSetting{}
	result = result.From(createdEntity)

	u.logger.Info("notification setting created", "id", result.ID.String(), "name", result.Name, "channel", string(result.Channel))
	return result, nil
}

func (u *notificationSettingUsecase) GetSetting(ctx context.Context, id uuid.UUID) (*domain.NotificationSetting, error) {
	entity, err := u.repo.GetByID(ctx, id)
	if err != nil {
		u.logger.Error("get notification setting failed", "id", id.String(), "error", err)
		return nil, fmt.Errorf("获取通知设置失败: %w", err)
	}

	setting := &domain.NotificationSetting{}
	setting.From(entity)
	return setting, nil
}

func (u *notificationSettingUsecase) GetSettingByNameAndChannel(ctx context.Context, name string, channel consts.NotificationChannel) (*domain.NotificationSetting, error) {
	entity, err := u.repo.GetByNameAndChannel(ctx, name, channel)
	if err != nil {
		u.logger.Error("get notification setting by name and channel failed", "name", name, "channel", string(channel), "error", err)
		return nil, fmt.Errorf("获取通知设置失败: %w", err)
	}

	setting := &domain.NotificationSetting{}
	setting.From(entity)
	return setting, nil
}

func (u *notificationSettingUsecase) GetSettingsByChannel(ctx context.Context, channel consts.NotificationChannel) ([]*domain.NotificationSetting, error) {
	entities, err := u.repo.GetByChannel(ctx, channel)
	if err != nil {
		u.logger.Error("get notification settings by channel failed", "channel", string(channel), "error", err)
		return nil, fmt.Errorf("获取通知设置失败: %w", err)
	}

	settings := make([]*domain.NotificationSetting, len(entities))
	for i, entity := range entities {
		setting := &domain.NotificationSetting{}
		setting.From(entity)
		settings[i] = setting
	}
	return settings, nil
}

func (u *notificationSettingUsecase) UpdateSetting(ctx context.Context, setting *domain.NotificationSetting) error {
	// 验证配置
	if err := u.ValidateConfig(ctx, setting); err != nil {
		u.logger.Error("validate notification setting config failed", "error", err)
		return fmt.Errorf("配置验证失败: %w", err)
	}

	// 检查设置是否存在
	_, err := u.repo.GetByID(ctx, setting.ID)
	if err != nil {
		u.logger.Error("get notification setting for update failed", "id", setting.ID.String(), "error", err)
		return fmt.Errorf("通知设置不存在: %w", err)
	}

	// 检查同一名称和通道的配置是否已存在（排除当前设置）
	existingByNameAndChannelEntity, errNameChannel := u.repo.GetByNameAndChannel(ctx, setting.Name, setting.Channel)
	if errNameChannel == nil && existingByNameAndChannelEntity != nil {
		existingByNameAndChannel := &domain.NotificationSetting{}
		existingByNameAndChannel.From(existingByNameAndChannelEntity)
		if existingByNameAndChannel.ID != setting.ID {
			return fmt.Errorf("名称为 %s 的 %s 通道配置已存在", setting.Name, setting.Channel)
		}
	}

	// 更新设置
	_, err = u.repo.Update(ctx, setting)
	if err != nil {
		u.logger.Error("update notification setting failed", "error", err)
		return fmt.Errorf("更新通知设置失败: %w", err)
	}

	u.logger.Info("notification setting updated", "id", setting.ID.String(), "name", setting.Name, "channel", string(setting.Channel))
	return nil
}

func (u *notificationSettingUsecase) DeleteSetting(ctx context.Context, id uuid.UUID) error {
	// 检查设置是否存在
	_, err := u.repo.GetByID(ctx, id)
	if err != nil {
		u.logger.Error("get notification setting for delete failed", "id", id.String(), "error", err)
		return fmt.Errorf("通知设置不存在: %w", err)
	}

	// 删除设置
	if err := u.repo.Delete(ctx, id); err != nil {
		u.logger.Error("delete notification setting failed", "id", id.String(), "error", err)
		return fmt.Errorf("删除通知设置失败: %w", err)
	}

	u.logger.Info("notification setting deleted", "id", id.String())
	return nil
}

// List 获取通知设置列表（支持过滤和分页）
func (u *notificationSettingUsecase) List(ctx context.Context, req *domain.ListSettingsRequest) (*domain.ListSettingsResponse, error) {
	entities, pageInfo, err := u.repo.List(ctx, req)
	if err != nil {
		u.logger.Error("list notification settings failed", "error", err)
		return nil, fmt.Errorf("获取通知设置列表失败: %w", err)
	}

	// 转换为领域对象
	settings := make([]*domain.NotificationSetting, 0, len(entities))
	for _, entity := range entities {
		setting := &domain.NotificationSetting{}
		setting.From(entity)
		settings = append(settings, setting)
	}

	return &domain.ListSettingsResponse{
		Items:    settings,
		PageInfo: pageInfo,
	}, nil
}

func (u *notificationSettingUsecase) GetEnabledSettings(ctx context.Context) ([]*domain.NotificationSetting, error) {
	entities, err := u.repo.GetEnabledSettings(ctx)
	if err != nil {
		u.logger.Error("get enabled notification settings failed", "error", err)
		return nil, fmt.Errorf("获取启用的通知设置失败: %w", err)
	}

	settings := make([]*domain.NotificationSetting, len(entities))
	for i, entity := range entities {
		setting := &domain.NotificationSetting{}
		setting.From(entity)
		settings[i] = setting
	}
	return settings, nil
}

func (u *notificationSettingUsecase) ValidateConfig(ctx context.Context, setting *domain.NotificationSetting) error {
	if setting == nil {
		return fmt.Errorf("通知设置不能为空")
	}

	// 验证通道
	if setting.Channel == "" {
		return fmt.Errorf("通知通道不能为空")
	}

	// 验证重试次数
	if setting.MaxRetry < 0 {
		return fmt.Errorf("最大重试次数不能为负数")
	}

	// 验证超时时间
	if setting.Timeout <= 0 {
		return fmt.Errorf("超时时间必须大于0")
	}

	// 根据通道类型验证配置
	switch setting.Channel {
	case consts.NotificationChannelDingTalk:
		if setting.DingTalkConfig == nil {
			return fmt.Errorf("钉钉通知配置不能为空")
		}
		if setting.DingTalkConfig.WebhookURL == "" {
			return fmt.Errorf("钉钉Webhook URL不能为空")
		}
	default:
		return fmt.Errorf("不支持的通知通道: %s", setting.Channel)
	}

	return nil
}
