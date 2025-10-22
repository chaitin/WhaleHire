package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"
	"sync"
	"text/template"
	"time"

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

// getDingTalkClients 获取所有启用的钉钉客户端
func (d *DingTalkAdapter) getDingTalkClients(ctx context.Context) ([]*dingtalk.DingTalk, error) {
	settings, err := d.settingUsecase.GetSettingsByChannel(ctx, consts.NotificationChannelDingTalk)
	if err != nil {
		return nil, fmt.Errorf("failed to get dingtalk settings: %w", err)
	}

	if len(settings) == 0 {
		return nil, fmt.Errorf("no dingtalk configurations found")
	}

	var clients []*dingtalk.DingTalk
	for _, setting := range settings {
		// 只处理启用的配置
		if !setting.Enabled || setting.DingTalkConfig == nil {
			continue
		}

		client := dingtalk.InitDingTalkWithSecret(setting.DingTalkConfig.Token, setting.DingTalkConfig.Secret,
			dingtalk.WithInitSendTimeout(10*time.Second)) // 增加超时时间到10秒
		clients = append(clients, client)
	}

	if len(clients) == 0 {
		return nil, fmt.Errorf("no enabled dingtalk configurations found")
	}

	return clients, nil
}

// broadcastMessage 群发消息到所有钉钉群
func (d *DingTalkAdapter) broadcastMessage(ctx context.Context, title, content string) error {
	clients, err := d.getDingTalkClients(ctx)
	if err != nil {
		return fmt.Errorf("failed to get dingtalk clients: %w", err)
	}

	// 使用 channel 收集错误
	errChan := make(chan error, len(clients))

	// 使用信号量限制并发数，避免同时发送过多请求
	semaphore := make(chan struct{}, 3) // 最多同时发送3个请求
	var wg sync.WaitGroup

	// 并发发送到所有钉钉群，但限制并发数
	for _, client := range clients {
		wg.Add(1)
		go func(c *dingtalk.DingTalk) {
			defer wg.Done()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 添加随机延迟，进一步避免请求冲突
			time.Sleep(time.Duration(len(clients)) * 50 * time.Millisecond)

			// 重试机制：最多重试3次
			var lastErr error
			for retry := 0; retry < 3; retry++ {
				if err := c.SendMarkDownMessage(title, content); err != nil {
					lastErr = err
					// 指数退避：第一次重试等待1秒，第二次2秒，第三次4秒
					if retry < 2 {
						time.Sleep(time.Duration(1<<retry) * time.Second)
					}
					continue
				}
				// 发送成功
				errChan <- nil
				return
			}
			// 重试3次后仍然失败
			errChan <- fmt.Errorf("failed to send message to dingtalk group after 3 retries: %w", lastErr)
		}(client)
	}

	// 等待所有goroutine完成
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// 收集所有结果
	var errors []error
	for err := range errChan {
		if err != nil {
			errors = append(errors, err)
		}
	}

	// 如果有错误，记录并返回
	if len(errors) > 0 {
		d.logger.ErrorContext(ctx, "Some dingtalk messages failed to send",
			slog.Int("total_groups", len(clients)),
			slog.Int("failed_groups", len(errors)),
			slog.Int("success_groups", len(clients)-len(errors)),
		)

		// 如果所有群都发送失败，返回错误
		if len(errors) == len(clients) {
			return fmt.Errorf("all dingtalk groups failed to receive message")
		}

		// 部分失败，记录警告但不返回错误
		d.logger.WarnContext(ctx, "Partial success in broadcasting dingtalk message",
			slog.Int("success_groups", len(clients)-len(errors)),
			slog.Int("failed_groups", len(errors)),
		)
	} else {
		d.logger.InfoContext(ctx, "Successfully broadcasted message to all dingtalk groups",
			slog.Int("total_groups", len(clients)),
		)
	}

	return nil
}

// sendResumeParseCompletedNotification 发送简历解析完成通知
func (d *DingTalkAdapter) sendResumeParseCompletedNotification(ctx context.Context, event *domain.NotificationEvent) error {
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

	// 群发消息到所有钉钉群
	if err := d.broadcastMessage(ctx, title, content); err != nil {
		d.logger.ErrorContext(ctx, "Failed to broadcast resume parse notification",
			slog.String("error", err.Error()),
			slog.String("event_id", event.ID.String()),
		)
		return fmt.Errorf("failed to broadcast dingtalk message: %w", err)
	}

	return nil
}

// sendBatchResumeParseCompletedNotification 发送批量简历解析完成通知
func (d *DingTalkAdapter) sendBatchResumeParseCompletedNotification(ctx context.Context, event *domain.NotificationEvent) error {
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
		CompletedAt:  payload.CompletedAt.In(time.Local).Format("2006-01-02 15:04:05"),
	}

	content, err := d.renderTemplate("batch_resume_parse_completed", templateData)
	if err != nil {
		return fmt.Errorf("failed to render batch template: %w", err)
	}

	// 群发消息到所有钉钉群
	if err := d.broadcastMessage(ctx, title, content); err != nil {
		d.logger.ErrorContext(ctx, "Failed to broadcast batch resume parse notification",
			slog.String("error", err.Error()),
			slog.String("event_id", event.ID.String()),
		)
		return fmt.Errorf("failed to broadcast dingtalk message: %w", err)
	}

	return nil
}

// sendMatchingCompletedNotification 发送匹配完成通知
func (d *DingTalkAdapter) sendMatchingCompletedNotification(ctx context.Context, event *domain.NotificationEvent) error {
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

	// 群发消息到所有钉钉群
	if err := d.broadcastMessage(ctx, title, content); err != nil {
		d.logger.ErrorContext(ctx, "Failed to broadcast matching notification",
			slog.String("error", err.Error()),
			slog.String("event_id", event.ID.String()),
		)
		return fmt.Errorf("failed to broadcast dingtalk message: %w", err)
	}

	return nil
}

// sendScreeningTaskCompletedNotification 发送筛选任务完成通知
func (d *DingTalkAdapter) sendScreeningTaskCompletedNotification(ctx context.Context, event *domain.NotificationEvent) error {
	// 将 event.Payload 反序列化为 ScreeningTaskCompletedPayload 结构体
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	var payload domain.ScreeningTaskCompletedPayload
	if errJson := json.Unmarshal(payloadBytes, &payload); errJson != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", errJson)
	}

	title := "📋 智能匹配任务完成"

	// 使用模板渲染通知内容
	templateData := struct {
		TaskID      string
		JobID       string
		JobName     string
		UserName    string
		TotalCount  int
		PassedCount int
		CompletedAt string
	}{
		TaskID:      payload.TaskID.String(),
		JobName:     payload.JobName,
		UserName:    payload.UserName,
		TotalCount:  payload.TotalCount,
		PassedCount: payload.PassedCount,
		CompletedAt: payload.CompletedAt.Format("2006-01-02 15:04:05"),
	}

	content, err := d.renderTemplate("screening_task_completed", templateData)
	if err != nil {
		return fmt.Errorf("failed to render screening template: %w", err)
	}

	// 群发消息到所有钉钉群
	if err := d.broadcastMessage(ctx, title, content); err != nil {
		d.logger.ErrorContext(ctx, "Failed to broadcast screening notification",
			slog.String("error", err.Error()),
			slog.String("event_id", event.ID.String()),
		)
		return fmt.Errorf("failed to broadcast dingtalk message: %w", err)
	}

	return nil
}

// GetChannelType 获取通道类型
func (d *DingTalkAdapter) GetChannelType() consts.NotificationChannel {
	return consts.NotificationChannelDingTalk
}
