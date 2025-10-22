package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"entgo.io/ent/dialect/sql"
	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/db"
	"github.com/chaitin/WhaleHire/backend/db/notificationsetting"
	"github.com/chaitin/WhaleHire/backend/domain"
)

// notificationSettingRepo 通知设置仓储实现
type notificationSettingRepo struct {
	client *db.Client
}

// NewNotificationSettingRepo 创建通知设置仓储实例
func NewNotificationSettingRepo(client *db.Client) domain.NotificationSettingRepo {
	return &notificationSettingRepo{
		client: client,
	}
}

// Create 创建通知设置
func (r *notificationSettingRepo) Create(ctx context.Context, setting *domain.NotificationSetting) (*db.NotificationSetting, error) {
	create := r.client.NotificationSetting.Create().
		SetName(setting.Name).
		SetChannel(setting.Channel).
		SetEnabled(setting.Enabled).
		SetMaxRetry(setting.MaxRetry).
		SetTimeout(setting.Timeout).
		SetDescription(setting.Description)

	if setting.DingTalkConfig != nil {
		// 将 types.NotificationDingTalkConfig 转换为 map[string]interface{}
		configMap := map[string]interface{}{
			"webhook_url": setting.DingTalkConfig.WebhookURL,
			"token":       setting.DingTalkConfig.Token,
			"keywords":    setting.DingTalkConfig.Keywords,
			"secret":      setting.DingTalkConfig.Secret,
		}
		create.SetDingtalkConfig(configMap)
	}

	entity, err := create.Save(ctx)
	if err != nil {
		return nil, err
	}

	return entity, nil
}

// GetByID 根据ID获取通知设置
func (r *notificationSettingRepo) GetByID(ctx context.Context, id uuid.UUID) (*db.NotificationSetting, error) {
	entity, err := r.client.NotificationSetting.
		Query().
		Where(notificationsetting.ID(id)).
		First(ctx)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return entity, nil
}

// GetByNameAndChannel 根据名称和通道获取通知设置
func (r *notificationSettingRepo) GetByNameAndChannel(ctx context.Context, name string, channel consts.NotificationChannel) (*db.NotificationSetting, error) {
	entity, err := r.client.NotificationSetting.
		Query().
		Where(
			notificationsetting.Name(name),
			notificationsetting.Channel(channel),
		).
		First(ctx)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return entity, nil
}

// GetByChannel 根据通道获取所有通知设置
func (r *notificationSettingRepo) GetByChannel(ctx context.Context, channel consts.NotificationChannel) ([]*db.NotificationSetting, error) {
	entities, err := r.client.NotificationSetting.
		Query().
		Where(notificationsetting.Channel(channel)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return entities, nil
}

// List 获取通知设置列表（支持过滤和分页）
func (r *notificationSettingRepo) List(ctx context.Context, req *domain.ListSettingsRequest) ([]*db.NotificationSetting, *db.PageInfo, error) {
	query := r.client.NotificationSetting.Query()

	// 应用过滤条件
	if req != nil {
		if req.Channel != nil {
			query = query.Where(notificationsetting.Channel(*req.Channel))
		}
		if req.Enabled != nil {
			query = query.Where(notificationsetting.Enabled(*req.Enabled))
		}
		if req.Keyword != nil && *req.Keyword != "" {
			keyword := strings.TrimSpace(*req.Keyword)
			if keyword != "" {
				query = query.Where(notificationsetting.Or(
					notificationsetting.NameContainsFold(keyword),
					notificationsetting.DescriptionContainsFold(keyword),
				))
			}
		}
	}

	// 排序：按更新时间倒序
	query = query.Order(notificationsetting.ByUpdatedAt(sql.OrderDesc()))

	// 分页
	page := 1
	pageSize := 20
	if req != nil {
		if req.Page > 0 {
			page = req.Page
		}
		if req.Size > 0 {
			pageSize = req.Size
		}
	}

	items, pageInfo, err := query.Page(ctx, page, pageSize)
	if err != nil {
		return nil, nil, fmt.Errorf("list notification settings failed: %w", err)
	}

	return items, pageInfo, nil
}

// Update 更新通知设置
func (r *notificationSettingRepo) Update(ctx context.Context, setting *domain.NotificationSetting) (*db.NotificationSetting, error) {
	update := r.client.NotificationSetting.
		UpdateOneID(setting.ID).
		SetName(setting.Name).
		SetChannel(setting.Channel).
		SetEnabled(setting.Enabled).
		SetMaxRetry(setting.MaxRetry).
		SetTimeout(setting.Timeout).
		SetDescription(setting.Description)

	if setting.DingTalkConfig != nil {
		// 将 domain.NotificationDingTalkConfig 转换为 map[string]interface{}
		configMap := map[string]interface{}{
			"webhook_url": setting.DingTalkConfig.WebhookURL,
			"token":       setting.DingTalkConfig.Token,
			"keywords":    setting.DingTalkConfig.Keywords,
			"secret":      setting.DingTalkConfig.Secret,
		}
		update.SetDingtalkConfig(configMap)
	} else {
		update.ClearDingtalkConfig()
	}

	entity, err := update.Save(ctx)
	if err != nil {
		return nil, err
	}

	return entity, nil
}

// Delete 删除通知设置
func (r *notificationSettingRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.client.NotificationSetting.DeleteOneID(id).Exec(ctx)
}

// GetEnabledSettings 获取启用的通知设置
func (r *notificationSettingRepo) GetEnabledSettings(ctx context.Context) ([]*db.NotificationSetting, error) {
	entities, err := r.client.NotificationSetting.
		Query().
		Where(notificationsetting.Enabled(true)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return entities, nil
}
