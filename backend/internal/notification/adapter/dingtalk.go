package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"
	"text/template"

	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/pkg/dingtalk"
)

// DingTalkAdapter 钉钉发送适配器
type DingTalkAdapter struct {
	settingUsecase domain.NotificationSettingUsecase
	logger         *slog.Logger
	templates      map[string]*template.Template
}

// NewDingTalkAdapter 创建钉钉适配器
func NewDingTalkAdapter(settingUsecase domain.NotificationSettingUsecase, logger *slog.Logger) *DingTalkAdapter {
	adapter := &DingTalkAdapter{
		settingUsecase: settingUsecase,
		logger:         logger,
		templates:      make(map[string]*template.Template),
	}

	// 初始化模板
	adapter.initTemplates()

	return adapter
}

// initTemplates 初始化通知模板
func (d *DingTalkAdapter) initTemplates() {
	templateDir := "templates/notification"

	templateFiles := map[string]string{
		"resume_parse_success":         "resume_parse_success.tmpl",
		"resume_parse_failure":         "resume_parse_failure.tmpl",
		"batch_resume_parse_completed": "batch_resume_parse_completed.tmpl",
		"job_matching_completed":       "job_matching_completed.tmpl",
		"screening_task_completed":     "screening_task_completed.tmpl",
	}

	for name, filename := range templateFiles {
		tmplPath := filepath.Join(templateDir, filename)
		tmpl, err := template.ParseFiles(tmplPath)
		if err != nil {
			d.logger.Error("Failed to parse template", slog.String("template", name), slog.String("error", err.Error()))
			continue
		}
		d.templates[name] = tmpl
	}
}

// renderTemplate 渲染模板
func (d *DingTalkAdapter) renderTemplate(templateName string, data interface{}) (string, error) {
	tmpl, exists := d.templates[templateName]
	if !exists {
		return "", fmt.Errorf("template %s not found", templateName)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", templateName, err)
	}

	return buf.String(), nil
}

// SendNotification 发送通知
func (d *DingTalkAdapter) SendNotification(ctx context.Context, event *domain.NotificationEvent) error {
	switch event.EventType {
	case consts.NotificationEventTypeResumeParseCompleted:
		return d.sendResumeParseCompletedNotification(ctx, event)
	case consts.NotificationEventTypeBatchResumeParseCompleted:
		return d.sendBatchResumeParseCompletedNotification(ctx, event)
	case consts.NotificationEventTypeJobMatchingCompleted:
		return d.sendMatchingCompletedNotification(ctx, event)
	case consts.NotificationEventTypeScreeningTaskCompleted:
		return d.sendScreeningTaskCompletedNotification(ctx, event)
	default:
		return fmt.Errorf("unsupported event type: %s", event.EventType)
	}
}

// getDingTalkClient 获取钉钉客户端
func (d *DingTalkAdapter) getDingTalkClient(ctx context.Context) (*dingtalk.DingTalk, error) {
	setting, err := d.settingUsecase.GetSettingByChannel(ctx, consts.NotificationChannelDingTalk)
	if err != nil {
		return nil, fmt.Errorf("failed to get dingtalk setting: %w", err)
	}

	if setting == nil || setting.DingTalkConfig == nil {
		return nil, fmt.Errorf("dingtalk configuration not found")
	}

	client := dingtalk.InitDingTalkWithSecret(setting.DingTalkConfig.Token, setting.DingTalkConfig.Secret)
	return client, nil
}

// sendResumeParseCompletedNotification 发送简历解析完成通知
func (d *DingTalkAdapter) sendResumeParseCompletedNotification(ctx context.Context, event *domain.NotificationEvent) error {
	client, err := d.getDingTalkClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to get dingtalk client: %w", err)
	}

	// 将 event.Payload 反序列化为 ResumeParseCompletedPayload 结构体
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	var payload domain.ResumeParseCompletedPayload
	if errJson := json.Unmarshal(payloadBytes, &payload); errJson != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", errJson)
	}

	var title string
	var templateName string
	var templateData interface{}

	if payload.Success {
		title = "✅ 简历解析成功"
		templateName = "resume_parse_success"
		templateData = struct {
			ResumeID    string
			UserID      string
			FileName    string
			ProcessedAt string
		}{
			ResumeID:    payload.ResumeID.String(),
			UserID:      payload.UserID.String(),
			FileName:    payload.FileName,
			ProcessedAt: payload.ParsedAt.Format("2006-01-02 15:04:05"),
		}
	} else {
		title = "❌ 简历解析失败"
		templateName = "resume_parse_failure"
		templateData = struct {
			ResumeID    string
			UserID      string
			FileName    string
			ErrorMsg    string
			ProcessedAt string
		}{
			ResumeID:    payload.ResumeID.String(),
			UserID:      payload.UserID.String(),
			FileName:    payload.FileName,
			ErrorMsg:    payload.ErrorMsg,
			ProcessedAt: payload.ParsedAt.Format("2006-01-02 15:04:05"),
		}
	}

	// 使用模板渲染通知内容
	content, err := d.renderTemplate(templateName, templateData)
	if err != nil {
		return fmt.Errorf("failed to render resume template: %w", err)
	}

	if err := client.SendMarkDownMessage(title, content); err != nil {
		d.logger.ErrorContext(ctx, "Failed to send resume notification",
			slog.String("error", err.Error()),
			slog.String("event_id", event.ID.String()),
		)
		return fmt.Errorf("failed to send dingtalk message: %w", err)
	}

	return nil
}

// sendBatchResumeParseCompletedNotification 发送批量简历解析完成通知
func (d *DingTalkAdapter) sendBatchResumeParseCompletedNotification(ctx context.Context, event *domain.NotificationEvent) error {
	client, err := d.getDingTalkClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to get dingtalk client: %w", err)
	}

	// 直接反序列化为 BatchResumeParseCompletedPayload 结构体
	var payload domain.BatchResumeParseCompletedPayload

	// 将 event.Payload 重新序列化为 JSON，然后反序列化为结构体
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if errJson := json.Unmarshal(payloadBytes, &payload); errJson != nil {
		return fmt.Errorf("failed to unmarshal payload to BatchResumeParseCompletedPayload: %w", err)
	}

	// 计算成功率
	successRate := float64(0)
	if payload.TotalCount > 0 {
		successRate = float64(payload.SuccessCount) / float64(payload.TotalCount) * 100
	}

	title := "📊 批量简历解析任务完成"

	// 使用模板渲染通知内容
	templateData := struct {
		TaskID       string
		UploaderName string
		TotalCount   int
		SuccessCount int
		FailedCount  int
		SuccessRate  float64
		Source       string
		CompletedAt  string
	}{
		TaskID:       payload.TaskID,
		UploaderName: payload.UploaderName,
		TotalCount:   payload.TotalCount,
		SuccessCount: payload.SuccessCount,
		FailedCount:  payload.FailedCount,
		SuccessRate:  successRate,
		Source:       payload.Source,
		CompletedAt:  payload.CompletedAt.Format("2006-01-02 15:04:05"),
	}

	content, err := d.renderTemplate("batch_resume_parse_completed", templateData)
	if err != nil {
		return fmt.Errorf("failed to render batch template: %w", err)
	}

	if err := client.SendMarkDownMessage(title, content); err != nil {
		d.logger.ErrorContext(ctx, "Failed to send batch resume parse notification",
			slog.String("error", err.Error()),
			slog.String("event_id", event.ID.String()),
		)
		return fmt.Errorf("failed to send dingtalk message: %w", err)
	}

	return nil
}

// sendMatchingCompletedNotification 发送匹配完成通知
func (d *DingTalkAdapter) sendMatchingCompletedNotification(ctx context.Context, event *domain.NotificationEvent) error {
	client, err := d.getDingTalkClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to get dingtalk client: %w", err)
	}

	// 将 event.Payload 反序列化为 JobMatchingCompletedPayload 结构体
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	var payload domain.JobMatchingCompletedPayload
	if errJson := json.Unmarshal(payloadBytes, &payload); errJson != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", errJson)
	}

	title := "🎯 职位匹配完成"

	// 使用模板渲染通知内容
	templateData := struct {
		JobID       string
		ResumeID    string
		UserID      string
		MatchScore  float64
		ProcessedAt string
	}{
		JobID:       payload.JobID.String(),
		ResumeID:    payload.ResumeID.String(),
		UserID:      payload.UserID.String(),
		MatchScore:  payload.MatchScore,
		ProcessedAt: payload.MatchedAt.Format("2006-01-02 15:04:05"),
	}

	content, err := d.renderTemplate("job_matching_completed", templateData)
	if err != nil {
		return fmt.Errorf("failed to render matching template: %w", err)
	}

	if err := client.SendMarkDownMessage(title, content); err != nil {
		d.logger.ErrorContext(ctx, "Failed to send matching notification",
			slog.String("error", err.Error()),
			slog.String("event_id", event.ID.String()),
		)
		return fmt.Errorf("failed to send dingtalk message: %w", err)
	}

	return nil
}

// sendScreeningTaskCompletedNotification 发送筛选任务完成通知
func (d *DingTalkAdapter) sendScreeningTaskCompletedNotification(ctx context.Context, event *domain.NotificationEvent) error {
	client, err := d.getDingTalkClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to get dingtalk client: %w", err)
	}

	// 将 event.Payload 反序列化为 ScreeningTaskCompletedPayload 结构体
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	var payload domain.ScreeningTaskCompletedPayload
	if errJson := json.Unmarshal(payloadBytes, &payload); errJson != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", errJson)
	}

	title := "📋 简历筛选任务完成"

	// 使用模板渲染通知内容
	templateData := struct {
		TaskID      string
		JobID       string
		JobName     string
		UserName    string
		TotalCount  int
		PassedCount int
		AvgScore    float64
		CompletedAt string
	}{
		TaskID:      payload.TaskID.String(),
		JobName:     payload.JobName,
		UserName:    payload.UserName,
		TotalCount:  payload.TotalCount,
		PassedCount: payload.PassedCount,
		AvgScore:    payload.AvgScore,
		CompletedAt: payload.CompletedAt.Format("2006-01-02 15:04:05"),
	}

	content, err := d.renderTemplate("screening_task_completed", templateData)
	if err != nil {
		return fmt.Errorf("failed to render screening template: %w", err)
	}

	if err := client.SendMarkDownMessage(title, content); err != nil {
		d.logger.ErrorContext(ctx, "Failed to send screening notification",
			slog.String("error", err.Error()),
			slog.String("event_id", event.ID.String()),
		)
		return fmt.Errorf("failed to send dingtalk message: %w", err)
	}

	return nil
}

// GetChannelType 获取通道类型
func (d *DingTalkAdapter) GetChannelType() consts.NotificationChannel {
	return consts.NotificationChannelDingTalk
}
