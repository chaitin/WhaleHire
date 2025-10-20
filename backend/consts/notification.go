package consts

// NotificationEventType 通知事件类型
type NotificationEventType string

const (
	// NotificationEventTypeResumeParseCompleted 简历解析完成
	NotificationEventTypeResumeParseCompleted NotificationEventType = "resume_parse_completed"
	// NotificationEventTypeBatchResumeParseCompleted 批量简历解析完成
	NotificationEventTypeBatchResumeParseCompleted NotificationEventType = "batch_resume_parse_completed"
	// NotificationEventTypeJobMatchingCompleted 岗位匹配完成
	NotificationEventTypeJobMatchingCompleted NotificationEventType = "job_matching_completed"
	// NotificationEventTypeScreeningTaskCompleted 筛选任务完成
	NotificationEventTypeScreeningTaskCompleted NotificationEventType = "screening_task_completed"
)

// NotificationChannel 通知渠道
type NotificationChannel string

const (
	// NotificationChannelDingTalk 钉钉
	NotificationChannelDingTalk NotificationChannel = "dingtalk"
	// NotificationChannelWebhook Webhook
	NotificationChannelWebhook NotificationChannel = "webhook"
	// NotificationChannelEmail 邮件
	NotificationChannelEmail NotificationChannel = "email"
)

// NotificationStatus 通知状态
type NotificationStatus string

const (
	// NotificationStatusPending 待发送
	NotificationStatusPending NotificationStatus = "pending"
	// NotificationStatusDelivering 发送中
	NotificationStatusDelivering NotificationStatus = "delivering"
	// NotificationStatusDelivered 已发送
	NotificationStatusDelivered NotificationStatus = "delivered"
	// NotificationStatusFailed 发送失败
	NotificationStatusFailed NotificationStatus = "failed"
	// NotificationStatusCancelled 已取消
	NotificationStatusCancelled NotificationStatus = "cancelled"
)
