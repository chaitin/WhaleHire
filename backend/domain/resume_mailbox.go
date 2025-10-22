package domain

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/db"
	"github.com/chaitin/WhaleHire/backend/errcode"
	"github.com/chaitin/WhaleHire/backend/pkg/web"
)

// =========================
// 错误定义
// =========================

var (
	ErrResumeMailboxSettingNotFound   = errcode.ErrResumeMailboxSettingNotFound
	ErrResumeMailboxStatisticNotFound = errcode.ErrResumeMailboxStatisticNotFound
	ErrInvalidCredentials             = errcode.ErrInvalidCredentials
	ErrMailboxConnectionFailed        = errcode.ErrMailboxConnectionFailed
)

// =========================
// 接口定义
// =========================

// ResumeMailboxSettingUsecase 邮箱设置用例接口
type ResumeMailboxSettingUsecase interface {
	// 创建邮箱设置
	Create(ctx context.Context, req *CreateResumeMailboxSettingRequest) (*ResumeMailboxSetting, error)
	// 获取邮箱设置详情
	GetByID(ctx context.Context, id uuid.UUID) (*ResumeMailboxSetting, error)
	// 更新邮箱设置
	Update(ctx context.Context, id uuid.UUID, req *UpdateResumeMailboxSettingRequest) (*ResumeMailboxSetting, error)
	// 删除邮箱设置
	Delete(ctx context.Context, id uuid.UUID) error
	// 获取邮箱设置列表
	List(ctx context.Context, req *ListResumeMailboxSettingsRequest) (*ListResumeMailboxSettingsResponse, error)
	// 测试邮箱连接
	TestConnection(ctx context.Context, req *TestConnectionRequest) error
	// 启用/禁用邮箱设置
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
}

// ResumeMailboxSettingRepo 邮箱设置仓储接口
type ResumeMailboxSettingRepo interface {
	// 创建邮箱设置
	Create(ctx context.Context, req *CreateResumeMailboxSettingRequest) (*db.ResumeMailboxSetting, error)
	// 根据ID获取邮箱设置
	GetByID(ctx context.Context, id uuid.UUID) (*db.ResumeMailboxSetting, error)
	// 更新邮箱设置
	Update(ctx context.Context, id uuid.UUID, req *UpdateResumeMailboxSettingRequest) (*db.ResumeMailboxSetting, error)
	// 删除邮箱设置
	Delete(ctx context.Context, id uuid.UUID) error
	// 获取邮箱设置列表
	List(ctx context.Context, req *ListResumeMailboxSettingsRequest) ([]*db.ResumeMailboxSetting, *db.PageInfo, error)
	// 更新同步状态
	UpdateSyncStatus(ctx context.Context, id uuid.UUID, req *UpdateSyncStatusRequest) error
	// 获取活跃的邮箱设置
	GetActiveSettings(ctx context.Context) ([]*db.ResumeMailboxSetting, error)
}

// ResumeMailboxCursorRepo 邮箱游标仓储接口
type ResumeMailboxCursorRepo interface {
	// 根据邮箱ID获取游标
	GetByMailboxID(ctx context.Context, mailboxID uuid.UUID) (*ResumeMailboxCursor, error)
	// 更新或创建游标
	Upsert(ctx context.Context, mailboxID uuid.UUID, protocolCursor string, lastMessageID string) (*ResumeMailboxCursor, error)
}

// CredentialVault 凭证加密存储接口
type CredentialVault interface {
	// 加密凭证
	Encrypt(ctx context.Context, credential map[string]interface{}) (map[string]interface{}, error)
	// 解密凭证
	Decrypt(ctx context.Context, encryptedCredential map[string]interface{}) (map[string]interface{}, error)
}

// MailboxProtocolAdapter 邮箱协议适配器接口
type MailboxProtocolAdapter interface {
	// TestConnection 测试邮箱连接
	TestConnection(ctx context.Context, config *MailboxConnectionConfig) error
	// Fetch 拉取新邮件
	Fetch(ctx context.Context, config *MailboxConnectionConfig, req *MailboxFetchRequest) (*MailboxFetchResult, error)
	// GetProtocol 返回协议标识
	GetProtocol() string
}

// MailboxAdapterFactory 邮箱适配器工厂接口
type MailboxAdapterFactory interface {
	// GetAdapter 根据协议获取适配器
	GetAdapter(protocol string) (MailboxProtocolAdapter, error)
}

// ResumeMailboxSyncUsecase 邮箱同步用例接口
type ResumeMailboxSyncUsecase interface {
	// SyncNow 手动触发同步
	SyncNow(ctx context.Context, mailboxID uuid.UUID) (*ResumeMailboxSyncResult, error)
}

// ResumeMailboxStatisticUsecase 邮箱统计用例接口
type ResumeMailboxStatisticUsecase interface {
	// 获取邮箱统计详情
	GetByMailboxIDAndDate(ctx context.Context, mailboxID uuid.UUID, date time.Time) (*ResumeMailboxStatistic, error)
	// 获取邮箱统计列表
	List(ctx context.Context, req *ListResumeMailboxStatisticsRequest) (*ListResumeMailboxStatisticsResponse, error)
	// 更新或创建统计记录
	Upsert(ctx context.Context, req *UpsertResumeMailboxStatisticRequest) (*ResumeMailboxStatistic, error)
	// 删除统计记录
	Delete(ctx context.Context, mailboxID uuid.UUID, date time.Time) error
	// 获取统计汇总
	GetSummary(ctx context.Context, req *GetMailboxStatisticsSummaryRequest) (*MailboxStatisticsSummary, error)
}

// ResumeMailboxStatisticRepo 邮箱统计仓储接口
type ResumeMailboxStatisticRepo interface {
	// 根据邮箱ID和日期获取统计
	GetByMailboxIDAndDate(ctx context.Context, mailboxID uuid.UUID, date time.Time) (*db.ResumeMailboxStatistic, error)
	// 获取统计列表
	List(ctx context.Context, req *ListResumeMailboxStatisticsRequest) ([]*db.ResumeMailboxStatistic, *db.PageInfo, error)
	// 更新或创建统计记录
	Upsert(ctx context.Context, req *UpsertResumeMailboxStatisticRequest) (*db.ResumeMailboxStatistic, error)
	// 删除统计记录
	Delete(ctx context.Context, mailboxID uuid.UUID, date time.Time) error
	// 获取统计汇总数据
	GetSummary(ctx context.Context, req *GetMailboxStatisticsSummaryRequest) (*MailboxStatisticsSummary, error)
}

// ResumeMailboxScheduler 邮箱同步调度器接口
type ResumeMailboxScheduler interface {
	// Upsert 注册或更新邮箱调度任务
	Upsert(ctx context.Context, setting *ResumeMailboxSetting) error
	// Remove 移除邮箱调度任务
	Remove(ctx context.Context, mailboxID uuid.UUID) error
}

// =========================
// 请求和响应结构体
// =========================

// CreateResumeMailboxSettingRequest 创建邮箱设置请求
type CreateResumeMailboxSettingRequest struct {
	Name                string                 `json:"name" validate:"required,min=1,max=100"`             // 邮箱设置名称
	EmailAddress        string                 `json:"email_address" validate:"required,email"`            // 邮箱地址
	Protocol            string                 `json:"protocol" validate:"required,oneof=imap pop3"`       // 邮箱协议：imap或pop3
	Host                string                 `json:"host" validate:"required,min=1,max=255"`             // 邮箱服务器地址
	Port                int                    `json:"port" validate:"required,min=1,max=65535"`           // 邮箱服务器端口
	UseSsl              bool                   `json:"use_ssl"`                                            // 是否使用SSL连接
	Folder              *string                `json:"folder,omitempty"`                                   // 邮箱文件夹，可选
	AuthType            string                 `json:"auth_type" validate:"required,oneof=password oauth"` // 认证类型：password或oauth
	EncryptedCredential map[string]interface{} `json:"encrypted_credential" validate:"required"`
	UploaderID          uuid.UUID              `json:"uploader_id" validate:"required"`                                     // 上传者ID
	JobProfileIDs       []uuid.UUID            `json:"job_profile_ids,omitempty"`                                           // 关联的职位档案ID列表，可选
	SyncIntervalMinutes *int                   `json:"sync_interval_minutes,omitempty" validate:"omitempty,min=5,max=1440"` // 同步间隔（分钟），可选，范围5-1440
}

// UpdateResumeMailboxSettingRequest 更新邮箱设置请求
type UpdateResumeMailboxSettingRequest struct {
	Name                *string                 `json:"name,omitempty" validate:"omitempty,min=1,max=100"`                   // 邮箱设置名称，可选
	EmailAddress        *string                 `json:"email_address,omitempty" validate:"omitempty,email"`                  // 邮箱地址，可选
	Protocol            *string                 `json:"protocol,omitempty" validate:"omitempty,oneof=imap pop3"`             // 邮箱协议：imap或pop3，可选
	Host                *string                 `json:"host,omitempty" validate:"omitempty,min=1,max=255"`                   // 邮箱服务器地址，可选
	Port                *int                    `json:"port,omitempty" validate:"omitempty,min=1,max=65535"`                 // 邮箱服务器端口，可选
	UseSsl              *bool                   `json:"use_ssl,omitempty"`                                                   // 是否使用SSL连接，可选
	Folder              *string                 `json:"folder,omitempty"`                                                    // 邮箱文件夹，可选
	AuthType            *string                 `json:"auth_type,omitempty" validate:"omitempty,oneof=password oauth"`       // 认证类型：password或oauth，可选
	EncryptedCredential *map[string]interface{} `json:"encrypted_credential,omitempty"`                                      // 加密后的认证凭据，可选。password类型需要：username(可选,默认使用email_address)、password(必需)；oauth类型需要：username(可选,默认使用email_address)、access_token(必需)
	JobProfileIDs       []uuid.UUID             `json:"job_profile_ids,omitempty"`                                           // 关联的职位档案ID列表，可选
	SyncIntervalMinutes *int                    `json:"sync_interval_minutes,omitempty" validate:"omitempty,min=5,max=1440"` // 同步间隔（分钟），可选，范围5-1440
	Status              *string                 `json:"status,omitempty" validate:"omitempty,oneof=enabled disabled"`        // 状态：enabled或disabled，可选
}

// ListResumeMailboxSettingsRequest 获取邮箱设置列表请求
type ListResumeMailboxSettingsRequest struct {
	web.Pagination

	// 过滤条件
	UploaderID    *uuid.UUID  `json:"uploader_id,omitempty" query:"uploader_id"`                                   // 上传者ID筛选，可选
	JobProfileIDs []uuid.UUID `json:"job_profile_ids,omitempty" query:"job_profile_ids"`                           // 职位档案ID列表筛选，可选
	Status        *string     `json:"status,omitempty" query:"status" validate:"omitempty,oneof=enabled disabled"` // 状态筛选：enabled或disabled，可选
	Protocol      *string     `json:"protocol,omitempty" query:"protocol" validate:"omitempty,oneof=imap pop3"`    // 协议筛选：imap或pop3，可选
}

// ListResumeMailboxSettingsResponse 获取邮箱设置列表响应
type ListResumeMailboxSettingsResponse struct {
	Items []*ResumeMailboxSetting `json:"items"` // 邮箱设置列表
	*db.PageInfo
}

// TestConnectionRequest 测试连接请求
type TestConnectionRequest struct {
	EmailAddress        string                 `json:"email_address" validate:"required,email"`                       // 邮箱地址
	Protocol            string                 `json:"protocol" validate:"required,oneof=imap pop3"`                  // 邮箱协议：imap或pop3
	Host                string                 `json:"host" validate:"required,min=1,max=255"`                        // 邮箱服务器地址
	Port                int                    `json:"port" validate:"required,min=1,max=65535"`                      // 邮箱服务器端口
	UseSsl              bool                   `json:"use_ssl"`                                                       // 是否使用SSL连接
	Folder              *string                `json:"folder,omitempty"`                                              // 邮箱文件夹，可选
	AuthType            string                 `json:"auth_type" validate:"required,oneof=password oauth"`            // 认证类型：password或oauth
	EncryptedCredential map[string]interface{} `json:"encrypted_credential" validate:"required" swaggertype:"object"` // 加密后的认证凭据。password类型需要：username(可选,默认使用email_address)、password(必需)；oauth类型需要：username(可选,默认使用email_address)、access_token(必需)
}

// TestConnectionResponse 测试连接响应
type TestConnectionResponse struct {
	Success bool   `json:"success"` // 连接是否成功
	Message string `json:"message"` // 连接结果消息
}

// GetResumeMailboxSettingRequest 获取邮箱设置详情请求
type GetResumeMailboxSettingRequest struct {
	ID uuid.UUID `json:"id" validate:"required"` // 邮箱设置ID
}

// DeleteResumeMailboxSettingRequest 删除邮箱设置请求
type DeleteResumeMailboxSettingRequest struct {
	ID uuid.UUID `json:"id" validate:"required"` // 邮箱设置ID
}

// UpdateResumeMailboxSettingStatusRequest 更新邮箱设置状态请求
type UpdateResumeMailboxSettingStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=enabled disabled"` // 状态：enabled或disabled
}

// UpdateSyncStatusRequest 更新同步状态请求
type UpdateSyncStatusRequest struct {
	Status       *string    `json:"status,omitempty"`
	LastSyncedAt *time.Time `json:"last_synced_at,omitempty"`
	LastError    *string    `json:"last_error,omitempty"`
	RetryCount   *int       `json:"retry_count,omitempty"`
}

// =========================
// 请求与响应模型（保留原有结构）
// =========================

// CreateResumeMailboxSettingReq 创建邮箱设置请求
type CreateResumeMailboxSettingReq struct {
	Name                string                 `json:"name" validate:"required,max=256"`
	EmailAddress        string                 `json:"email_address" validate:"required,email,max=256"`
	Protocol            string                 `json:"protocol" validate:"required,oneof=imap pop3"`
	Host                string                 `json:"host" validate:"required,max=256"`
	Port                int                    `json:"port" validate:"required,min=1,max=65535"`
	UseSSL              bool                   `json:"use_ssl"`
	Folder              *string                `json:"folder,omitempty" validate:"omitempty,max=256"`
	AuthType            string                 `json:"auth_type" validate:"required,oneof=password oauth"`
	Credential          map[string]interface{} `json:"credential" validate:"required"`
	UploaderID          uuid.UUID              `json:"uploader_id" validate:"required"`
	JobProfileIDs       []uuid.UUID            `json:"job_profile_ids,omitempty"`
	SyncIntervalMinutes *int                   `json:"sync_interval_minutes,omitempty" validate:"omitempty,min=5"`
	Status              *string                `json:"status,omitempty" validate:"omitempty,oneof=enabled disabled"`
}

// UpdateResumeMailboxSettingReq 更新邮箱设置请求
type UpdateResumeMailboxSettingReq struct {
	ID                  uuid.UUID              `json:"-"`
	Name                *string                `json:"name,omitempty" validate:"omitempty,max=256"`
	Host                *string                `json:"host,omitempty" validate:"omitempty,max=256"`
	Port                *int                   `json:"port,omitempty" validate:"omitempty,min=1,max=65535"`
	UseSSL              *bool                  `json:"use_ssl,omitempty"`
	Folder              *string                `json:"folder,omitempty" validate:"omitempty,max=256"`
	AuthType            *string                `json:"auth_type,omitempty" validate:"omitempty,oneof=password oauth"`
	Credential          map[string]interface{} `json:"credential,omitempty"`
	UploaderID          *uuid.UUID             `json:"uploader_id,omitempty"`
	JobProfileIDs       []uuid.UUID            `json:"job_profile_ids,omitempty"`
	SyncIntervalMinutes *int                   `json:"sync_interval_minutes,omitempty" validate:"omitempty,min=5"`
	Status              *string                `json:"status,omitempty" validate:"omitempty,oneof=enabled disabled"`
}

// ListResumeMailboxSettingReq 邮箱设置列表请求
type ListResumeMailboxSettingReq struct {
	web.Pagination
	Status       *string `json:"status,omitempty" query:"status" validate:"omitempty,oneof=enabled disabled"`
	Protocol     *string `json:"protocol,omitempty" query:"protocol" validate:"omitempty,oneof=imap pop3"`
	EmailAddress *string `json:"email_address,omitempty" query:"email_address"`
	UploaderID   *string `json:"uploader_id,omitempty" query:"uploader_id"`
	Keyword      *string `json:"keyword,omitempty" query:"keyword"`
}

// ListResumeMailboxSettingResp 邮箱设置列表响应
type ListResumeMailboxSettingResp struct {
	*db.PageInfo
	Items []*ResumeMailboxSetting `json:"items"`
}

// GetMailboxStatisticsReq 获取邮箱统计请求
type GetMailboxStatisticsReq struct {
	DateFrom *time.Time `json:"date_from,omitempty" query:"date_from"`
	DateTo   *time.Time `json:"date_to,omitempty" query:"date_to"`
	Range    *string    `json:"range,omitempty" query:"range" validate:"omitempty,oneof=7d 30d 90d"`
}

// GetMailboxStatisticsResp 获取邮箱统计响应
type GetMailboxStatisticsResp struct {
	Statistics []*ResumeMailboxStatistic `json:"statistics"`
	Summary    *MailboxStatisticsSummary `json:"summary"`
}

// ListResumeMailboxStatisticsRequest 邮箱统计列表请求
type ListResumeMailboxStatisticsRequest struct {
	web.Pagination
	MailboxID *uuid.UUID `json:"mailbox_id,omitempty" query:"mailbox_id"`                             // 邮箱ID筛选，可选
	DateFrom  *time.Time `json:"date_from,omitempty" query:"date_from"`                               // 开始日期筛选，可选
	DateTo    *time.Time `json:"date_to,omitempty" query:"date_to"`                                   // 结束日期筛选，可选
	Range     *string    `json:"range,omitempty" query:"range" validate:"omitempty,oneof=7d 30d 90d"` // 时间范围筛选：7d、30d、90d，可选
}

// ListResumeMailboxStatisticsResponse 邮箱统计列表响应
type ListResumeMailboxStatisticsResponse struct {
	Items []*ResumeMailboxStatistic `json:"items"` // 邮箱统计列表
	*db.PageInfo
}

// UpsertResumeMailboxStatisticRequest 更新或创建邮箱统计请求
type UpsertResumeMailboxStatisticRequest struct {
	MailboxID          uuid.UUID `json:"mailbox_id" validate:"required"`                             // 邮箱ID
	Date               time.Time `json:"date" validate:"required"`                                   // 统计日期
	SyncedEmails       *int      `json:"synced_emails,omitempty" validate:"omitempty,min=0"`         // 同步邮件数，可选
	ParsedResumes      *int      `json:"parsed_resumes,omitempty" validate:"omitempty,min=0"`        // 解析简历数，可选
	FailedResumes      *int      `json:"failed_resumes,omitempty" validate:"omitempty,min=0"`        // 失败简历数，可选
	SkippedAttachments *int      `json:"skipped_attachments,omitempty" validate:"omitempty,min=0"`   // 跳过附件数，可选
	LastSyncDurationMs *int      `json:"last_sync_duration_ms,omitempty" validate:"omitempty,min=0"` // 最后同步耗时，可选
}

// GetMailboxStatisticsSummaryRequest 获取邮箱统计汇总请求
type GetMailboxStatisticsSummaryRequest struct {
	MailboxID *uuid.UUID `json:"mailbox_id,omitempty" query:"mailbox_id"`                             // 邮箱ID筛选，可选
	DateFrom  *time.Time `json:"date_from,omitempty" query:"date_from"`                               // 开始日期筛选，可选
	DateTo    *time.Time `json:"date_to,omitempty" query:"date_to"`                                   // 结束日期筛选，可选
	Range     *string    `json:"range,omitempty" query:"range" validate:"omitempty,oneof=7d 30d 90d"` // 时间范围筛选：7d、30d、90d，可选
}

// =========================
// 数据模型
// =========================

// ResumeMailboxSetting 邮箱设置数据模型
type ResumeMailboxSetting struct {
	ID                  uuid.UUID              `json:"id"`
	Name                string                 `json:"name"`
	EmailAddress        string                 `json:"email_address"`
	Protocol            string                 `json:"protocol"`
	Host                string                 `json:"host"`
	Port                int                    `json:"port"`
	UseSsl              bool                   `json:"use_ssl"`
	Folder              string                 `json:"folder"`
	AuthType            string                 `json:"auth_type"`
	EncryptedCredential map[string]interface{} `json:"encrypted_credential"`
	UploaderID          uuid.UUID              `json:"uploader_id"`
	UploaderName        string                 `json:"uploader_name,omitempty"`
	JobProfileIDs       []uuid.UUID            `json:"job_profile_ids,omitempty"`
	JobProfileNames     []string               `json:"job_profile_names,omitempty"`
	SyncIntervalMinutes *int                   `json:"sync_interval_minutes,omitempty"`
	Status              string                 `json:"status"`
	LastSyncedAt        *time.Time             `json:"last_synced_at,omitempty"`
	LastError           string                 `json:"last_error"`
	RetryCount          int                    `json:"retry_count"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
}

// From 从数据库实体转换为domain对象
func (s *ResumeMailboxSetting) From(entity *db.ResumeMailboxSetting) *ResumeMailboxSetting {
	if entity == nil {
		return nil
	}

	s.ID = entity.ID
	s.Name = entity.Name
	s.EmailAddress = entity.EmailAddress
	s.Protocol = string(entity.Protocol)
	s.Host = entity.Host
	s.Port = entity.Port
	s.UseSsl = entity.UseSsl
	s.Folder = entity.Folder
	s.AuthType = string(entity.AuthType)
	s.EncryptedCredential = entity.EncryptedCredential
	s.UploaderID = entity.UploaderID
	// 使用多个JobProfileIDs
	s.JobProfileIDs = entity.JobProfileIds
	s.SyncIntervalMinutes = entity.SyncIntervalMinutes
	s.Status = string(entity.Status)
	s.LastSyncedAt = entity.LastSyncedAt
	s.LastError = entity.LastError
	s.RetryCount = entity.RetryCount
	s.CreatedAt = entity.CreatedAt
	s.UpdatedAt = entity.UpdatedAt

	// 填充关联数据
	if entity.Edges.Uploader != nil {
		s.UploaderName = entity.Edges.Uploader.Username
	}
	// JobProfileNames 由usecase层负责填充，这里不设置默认值

	return s
}

// ResumeMailboxCursor 邮箱同步游标数据模型
type ResumeMailboxCursor struct {
	ID             uuid.UUID `json:"id"`
	MailboxID      uuid.UUID `json:"mailbox_id"`
	ProtocolCursor string    `json:"protocol_cursor"`
	LastMessageID  string    `json:"last_message_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// From 从数据库实体转换为domain对象
func (c *ResumeMailboxCursor) From(entity *db.ResumeMailboxCursor) *ResumeMailboxCursor {
	if entity == nil {
		return nil
	}

	c.ID = entity.ID
	c.MailboxID = entity.MailboxID
	c.ProtocolCursor = entity.ProtocolCursor
	c.LastMessageID = entity.LastMessageID
	c.CreatedAt = entity.CreatedAt
	c.UpdatedAt = entity.UpdatedAt

	return c
}

// MailboxConnectionConfig 邮箱连接配置
type MailboxConnectionConfig struct {
	Host         string
	Port         int
	UseSSL       bool
	EmailAddress string
	Folder       string
	AuthType     string
	Credential   map[string]interface{}
}

// MailboxFetchRequest 邮箱拉取请求参数
type MailboxFetchRequest struct {
	Cursor string
	Limit  int
}

// MailboxFetchResult 邮箱拉取结果
type MailboxFetchResult struct {
	Messages      []*MailboxEmail
	NextCursor    string
	LastMessageID string
}

// MailboxEmail 邮件摘要
type MailboxEmail struct {
	MessageID   string
	Subject     string
	ReceivedAt  time.Time
	Attachments []*MailboxAttachment
	RawSize     int64
}

// MailboxAttachment 邮件附件
type MailboxAttachment struct {
	Filename    string
	ContentType string
	Content     []byte
	Size        int64
	Hash        string
	Reader      io.Reader `json:"-"`
}

// ResumeMailboxSyncResult 邮箱同步结果
type ResumeMailboxSyncResult struct {
	MailboxID          uuid.UUID     `json:"mailbox_id"`
	ProcessedEmails    int           `json:"processed_emails"`
	SuccessAttachments int           `json:"success_attachments"`
	FailedAttachments  int           `json:"failed_attachments"`
	SkippedAttachments int           `json:"skipped_attachments"`
	Duration           time.Duration `json:"duration"`
	LastMessageID      string        `json:"last_message_id"`
	Errors             []string      `json:"errors,omitempty"`
}

// ResumeMailboxStatistic 邮箱统计信息
type ResumeMailboxStatistic struct {
	ID                 uuid.UUID `json:"id"`
	MailboxID          uuid.UUID `json:"mailbox_id"`
	Date               time.Time `json:"date"`
	SyncedEmails       int       `json:"synced_emails"`
	ParsedResumes      int       `json:"parsed_resumes"`
	FailedResumes      int       `json:"failed_resumes"`
	SkippedAttachments int       `json:"skipped_attachments"`
	LastSyncDurationMs int       `json:"last_sync_duration_ms"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// From 从数据库实体转换为domain对象
func (s *ResumeMailboxStatistic) From(entity *db.ResumeMailboxStatistic) *ResumeMailboxStatistic {
	if entity == nil {
		return s
	}

	s.ID = entity.ID
	s.MailboxID = entity.MailboxID
	s.Date = entity.Date
	s.SyncedEmails = entity.SyncedEmails
	s.ParsedResumes = entity.ParsedResumes
	s.FailedResumes = entity.FailedResumes
	s.SkippedAttachments = entity.SkippedAttachments
	s.LastSyncDurationMs = entity.LastSyncDurationMs
	s.CreatedAt = entity.CreatedAt
	s.UpdatedAt = entity.UpdatedAt

	return s
}

// MailboxStatisticsSummary 邮箱统计汇总
type MailboxStatisticsSummary struct {
	TotalSyncedEmails       int     `json:"total_synced_emails"`
	TotalParsedResumes      int     `json:"total_parsed_resumes"`
	TotalFailedResumes      int     `json:"total_failed_resumes"`
	TotalSkippedAttachments int     `json:"total_skipped_attachments"`
	AvgSyncDurationMs       float64 `json:"avg_sync_duration_ms"`
	SuccessRate             float64 `json:"success_rate"`
}
