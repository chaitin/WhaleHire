package domain

import (
	"context"
	"time"

	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/db"
	"github.com/chaitin/WhaleHire/backend/pkg/web"
	"github.com/google/uuid"
)

// NotificationSettingUsecase 通知设置业务逻辑接口
type NotificationSettingUsecase interface {
	// CreateSetting 创建通知设置
	CreateSetting(ctx context.Context, setting *NotificationSetting) (*NotificationSetting, error)
	// GetSetting 根据ID获取通知设置
	GetSetting(ctx context.Context, id uuid.UUID) (*NotificationSetting, error)
	// GetSettingByNameAndChannel 根据名称和通道类型获取通知设置
	GetSettingByNameAndChannel(ctx context.Context, name string, channel consts.NotificationChannel) (*NotificationSetting, error)
	// GetSettingsByChannel 根据通道类型获取所有通知设置
	GetSettingsByChannel(ctx context.Context, channel consts.NotificationChannel) ([]*NotificationSetting, error)
	// UpdateSetting 更新通知设置
	UpdateSetting(ctx context.Context, setting *NotificationSetting) error
	// DeleteSetting 删除通知设置
	DeleteSetting(ctx context.Context, id uuid.UUID) error
	// ValidateConfig 验证通知设置配置
	ValidateConfig(ctx context.Context, setting *NotificationSetting) error
	// List 获取通知设置列表（支持过滤和分页）
	List(ctx context.Context, req *ListSettingsRequest) (*ListSettingsResponse, error)
	// GetEnabledSettings 获取启用的通知设置
	GetEnabledSettings(ctx context.Context) ([]*NotificationSetting, error)
}

// NotificationSettingRepo 通知设置仓储接口
type NotificationSettingRepo interface {
	// Create 创建通知设置
	Create(ctx context.Context, setting *NotificationSetting) (*db.NotificationSetting, error)
	// GetByID 根据ID获取通知设置
	GetByID(ctx context.Context, id uuid.UUID) (*db.NotificationSetting, error)
	// GetByNameAndChannel 根据名称和通道获取通知设置
	GetByNameAndChannel(ctx context.Context, name string, channel consts.NotificationChannel) (*db.NotificationSetting, error)
	// GetByChannel 根据通道获取所有通知设置
	GetByChannel(ctx context.Context, channel consts.NotificationChannel) ([]*db.NotificationSetting, error)
	// List 获取通知设置列表（支持过滤和分页）
	List(ctx context.Context, req *ListSettingsRequest) ([]*db.NotificationSetting, *db.PageInfo, error)
	// Update 更新通知设置
	Update(ctx context.Context, setting *NotificationSetting) (*db.NotificationSetting, error)
	// Delete 删除通知设置
	Delete(ctx context.Context, id uuid.UUID) error
	// GetEnabledSettings 获取启用的通知设置
	GetEnabledSettings(ctx context.Context) ([]*db.NotificationSetting, error)
}

// NotificationSetting 通知设置领域实体
type NotificationSetting struct {
	ID             uuid.UUID                   `json:"id"`
	Name           string                      `json:"name"`
	Channel        consts.NotificationChannel  `json:"channel"`
	Enabled        bool                        `json:"enabled"`
	DingTalkConfig *NotificationDingTalkConfig `json:"dingtalk_config,omitempty"`
	MaxRetry       int                         `json:"max_retry"`
	Timeout        int                         `json:"timeout"`
	Description    string                      `json:"description"`
	CreatedAt      time.Time                   `json:"created_at"`
	UpdatedAt      time.Time                   `json:"updated_at"`
}

// From 方法用于从 db 实体转换
func (ns *NotificationSetting) From(dbSetting *db.NotificationSetting) *NotificationSetting {
	if dbSetting == nil {
		return nil
	}

	ns.ID = dbSetting.ID
	ns.Name = dbSetting.Name
	ns.Channel = dbSetting.Channel
	ns.Enabled = dbSetting.Enabled
	ns.MaxRetry = dbSetting.MaxRetry
	ns.Timeout = dbSetting.Timeout
	ns.Description = dbSetting.Description
	ns.CreatedAt = dbSetting.CreatedAt
	ns.UpdatedAt = dbSetting.UpdatedAt

	// 转换 DingTalkConfig
	if dbSetting.DingtalkConfig != nil {
		configMap := dbSetting.DingtalkConfig
		ns.DingTalkConfig = &NotificationDingTalkConfig{
			WebhookURL: getStringFromMap(configMap, "webhook_url"),
			Token:      getStringFromMap(configMap, "token"),
			Keywords:   getStringFromMap(configMap, "keywords"),
			Secret:     getStringFromMap(configMap, "secret"),
		}
	}

	return ns
}

// getStringFromMap 从 map 中获取字符串值
func getStringFromMap(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// NotificationDingTalkConfig 钉钉通知配置
type NotificationDingTalkConfig struct {
	WebhookURL string `json:"webhook_url"`        // 钉钉机器人Webhook URL https://oapi.dingtalk.com/robot/send?access_token=XXX
	Token      string `json:"token"`              // 钉钉机器人Token
	Keywords   string `json:"keywords,omitempty"` // 关键词列表，用于 DingDing验证通知
	Secret     string `json:"secret,omitempty"`   // 钉钉机器人Secret
}

// CreateSettingRequest 创建通知设置请求
type CreateSettingRequest struct {
	Name           string                     `json:"name" validate:"required,max=100"`
	Channel        consts.NotificationChannel `json:"channel" validate:"required"`
	Enabled        bool                       `json:"enabled"`
	DingTalkConfig NotificationDingTalkConfig `json:"dingtalk_config,omitempty"`
	MaxRetry       int                        `json:"max_retry" validate:"min=0,max=10" default:"3"`
	Timeout        int                        `json:"timeout" validate:"min=1,max=300" default:"300"`
	Description    string                     `json:"description" validate:"max=500"`
}

// UpdateSettingRequest 更新通知设置请求
type UpdateSettingRequest struct {
	Name           string                     `json:"name" validate:"required,max=100"`
	Channel        consts.NotificationChannel `json:"channel" validate:"required"`
	Enabled        bool                       `json:"enabled"`
	DingTalkConfig NotificationDingTalkConfig `json:"dingtalk_config,omitempty"`
	MaxRetry       int                        `json:"max_retry" validate:"min=0,max=10" default:"3"`
	Timeout        int                        `json:"timeout" validate:"min=1,max=300" default:"300"`
	Description    string                     `json:"description" validate:"max=500"`
}

// GetSettingRequest 获取通知设置请求
type GetSettingRequest struct {
	ID string `param:"id" validate:"required,uuid"`
}

// DeleteSettingRequest 删除通知设置请求
type DeleteSettingRequest struct {
	ID string `param:"id" validate:"required,uuid"`
}

// ListSettingsRequest 获取通知设置列表请求
type ListSettingsRequest struct {
	web.Pagination

	// 过滤条件
	Channel *consts.NotificationChannel `json:"channel,omitempty" query:"channel"` // 通知渠道过滤
	Enabled *bool                       `json:"enabled,omitempty" query:"enabled"` // 启用状态过滤
	Keyword *string                     `json:"keyword,omitempty" query:"keyword"` // 关键词搜索（名称或描述）
}

// ListSettingsResponse 获取通知设置列表响应
type ListSettingsResponse struct {
	Items []*NotificationSetting `json:"items"`
	*db.PageInfo
}

// GetEnabledSettingsRequest 获取启用的通知设置列表请求
type GetEnabledSettingsRequest struct {
	// 暂时不需要参数，预留扩展
}

// GetEnabledSettingsResponse 获取启用的通知设置列表响应
type GetEnabledSettingsResponse struct {
	Settings []*NotificationSetting `json:"settings"`
}

// GetSettingByNameAndChannelRequest 根据名称和通道类型获取通知设置请求
type GetSettingByNameAndChannelRequest struct {
	Name    string                     `param:"name" validate:"required,max=100"`
	Channel consts.NotificationChannel `param:"channel" validate:"required"`
}

// GetSettingsByChannelRequest 根据通道类型获取通知设置列表请求
type GetSettingsByChannelRequest struct {
	Channel consts.NotificationChannel `param:"channel" validate:"required"`
}

// GetSettingsByChannelResponse 根据通道类型获取通知设置列表响应
type GetSettingsByChannelResponse struct {
	Settings []*NotificationSetting `json:"settings"`
}
