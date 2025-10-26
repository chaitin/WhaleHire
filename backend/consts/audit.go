package consts

// OperatorType 操作者类型
type OperatorType string

const (
	OperatorTypeUser  OperatorType = "user"  // 普通用户
	OperatorTypeAdmin OperatorType = "admin" // 管理员
)

// OperationType 操作类型
type OperationType string

const (
	OperationTypeCreate OperationType = "create" // 创建
	OperationTypeUpdate OperationType = "update" // 更新
	OperationTypeDelete OperationType = "delete" // 删除
	OperationTypeView   OperationType = "view"   // 查看（敏感数据）
	OperationTypeLogin  OperationType = "login"  // 登录
	OperationTypeLogout OperationType = "logout" // 登出
)

// ResourceType 资源类型
type ResourceType string

const (
	ResourceTypeUser                   ResourceType = "user"                     // 用户
	ResourceTypeAdmin                  ResourceType = "admin"                    // 管理员
	ResourceTypeRole                   ResourceType = "role"                     // 角色
	ResourceTypeDepartment             ResourceType = "department"               // 部门
	ResourceTypeJobPosition            ResourceType = "job_position"             // 职位
	ResourceTypeResume                 ResourceType = "resume"                   // 简历
	ResourceTypeScreening              ResourceType = "screening"                // 筛选任务
	ResourceTypeSetting                ResourceType = "setting"                  // 系统设置
	ResourceTypeAttachment             ResourceType = "attachment"               // 附件
	ResourceTypeConversation           ResourceType = "conversation"             // 对话
	ResourceTypeMessage                ResourceType = "message"                  // 消息
	ResourceTypeNotificationSetting    ResourceType = "notification_setting"     // 通知设置
	ResourceTypeResumeMailboxSetting   ResourceType = "resume_mailbox_setting"   // 简历邮箱设置
	ResourceTypeResumeMailboxStatistic ResourceType = "resume_mailbox_statistic" // 简历邮箱统计
)

// AuditLogStatus 审计日志状态
type AuditLogStatus string

const (
	AuditLogStatusSuccess AuditLogStatus = "success" // 成功
	AuditLogStatusFailed  AuditLogStatus = "failed"  // 失败
)
