package usecase

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/domain"
)

var (
	supportedExtensions = map[string]struct{}{
		".pdf":  {},
		".doc":  {},
		".docx": {},
		".txt":  {},
		".rtf":  {},
	}
)

// ResumeMailboxSyncUsecase 邮箱同步用例实现
type ResumeMailboxSyncUsecase struct {
	settingRepo           domain.ResumeMailboxSettingRepo
	cursorRepo            domain.ResumeMailboxCursorRepo
	statisticRepo         domain.ResumeMailboxStatisticRepo
	credentialVault       domain.CredentialVault
	adapterFactory        domain.MailboxAdapterFactory
	resumeUsecase         domain.ResumeUsecase
	jobApplicationUsecase domain.JobApplicationUsecase
	logger                *slog.Logger
}

// NewResumeMailboxSyncUsecase 创建同步用例实例
func NewResumeMailboxSyncUsecase(
	settingRepo domain.ResumeMailboxSettingRepo,
	cursorRepo domain.ResumeMailboxCursorRepo,
	statisticRepo domain.ResumeMailboxStatisticRepo,
	credentialVault domain.CredentialVault,
	adapterFactory domain.MailboxAdapterFactory,
	resumeUsecase domain.ResumeUsecase,
	jobApplicationUsecase domain.JobApplicationUsecase,
	logger *slog.Logger,
) domain.ResumeMailboxSyncUsecase {
	return &ResumeMailboxSyncUsecase{
		settingRepo:           settingRepo,
		cursorRepo:            cursorRepo,
		statisticRepo:         statisticRepo,
		credentialVault:       credentialVault,
		adapterFactory:        adapterFactory,
		resumeUsecase:         resumeUsecase,
		jobApplicationUsecase: jobApplicationUsecase,
		logger:                logger,
	}
}

// SyncNow 手动触发邮箱同步
func (u *ResumeMailboxSyncUsecase) SyncNow(ctx context.Context, mailboxID uuid.UUID) (*domain.ResumeMailboxSyncResult, error) {
	start := time.Now()

	// 加载邮箱配置
	dbSetting, err := u.settingRepo.GetByID(ctx, mailboxID)
	if err != nil {
		return nil, fmt.Errorf("获取邮箱配置失败: %w", err)
	}
	setting := (&domain.ResumeMailboxSetting{}).From(dbSetting)

	if strings.ToLower(setting.Status) != string(consts.MailboxStatusEnabled) {
		return nil, fmt.Errorf("邮箱同步未启用")
	}

	// 解密凭证
	credential, err := u.credentialVault.Decrypt(ctx, setting.EncryptedCredential)
	if err != nil {
		return nil, fmt.Errorf("解密邮箱凭证失败: %w", err)
	}

	adapter, err := u.adapterFactory.GetAdapter(strings.ToLower(setting.Protocol))
	if err != nil {
		return nil, fmt.Errorf("获取协议适配器失败: %w", err)
	}

	// 获取历史游标
	cursor, err := u.cursorRepo.GetByMailboxID(ctx, mailboxID)
	if err != nil {
		return nil, fmt.Errorf("读取邮箱游标失败: %w", err)
	}

	fetchReq := &domain.MailboxFetchRequest{}
	if cursor != nil {
		fetchReq.Cursor = cursor.ProtocolCursor
	}

	config := &domain.MailboxConnectionConfig{
		Host:         setting.Host,
		Port:         setting.Port,
		UseSSL:       setting.UseSsl,
		EmailAddress: setting.EmailAddress,
		Folder:       setting.Folder,
		AuthType:     setting.AuthType,
		Credential:   credential,
	}

	result := &domain.ResumeMailboxSyncResult{
		MailboxID: mailboxID,
	}

	fetchResult, err := adapter.Fetch(ctx, config, fetchReq)
	if err != nil {
		u.logger.Error("邮箱同步失败", slog.String("mailbox", setting.EmailAddress), slog.String("error", err.Error()))
		u.handleSyncFailure(ctx, setting, err)
		return nil, err
	}

	if fetchResult == nil {
		fetchResult = &domain.MailboxFetchResult{
			Messages: []*domain.MailboxEmail{},
		}
	}

	result.ProcessedEmails = len(fetchResult.Messages)

	var jobPositionIDs []string
	if len(setting.JobProfileIDs) > 0 {
		for _, id := range setting.JobProfileIDs {
			jobPositionIDs = append(jobPositionIDs, id.String())
		}
	}

	for _, mailItem := range fetchResult.Messages {
		for _, attachment := range mailItem.Attachments {
			if attachment == nil {
				continue
			}

			if !u.isSupportedAttachment(attachment.Filename) {
				result.SkippedAttachments++
				continue
			}

			reader := bytes.NewReader(attachment.Content)
			req := &domain.UploadResumeReq{
				UploaderID:     setting.UploaderID.String(),
				File:           reader,
				Filename:       attachment.Filename,
				JobPositionIDs: jobPositionIDs,
			}
			sourceStr := string(consts.ResumeSourceTypeEmail)
			req.Source = &sourceStr

			resume, uploadErr := u.resumeUsecase.Upload(ctx, req)
			if uploadErr != nil {
				result.FailedAttachments++
				result.Errors = append(result.Errors, fmt.Sprintf("上传附件失败 %s: %s", attachment.Filename, uploadErr.Error()))
				u.logger.Warn("上传邮件附件失败",
					slog.String("mailbox", setting.EmailAddress),
					slog.String("filename", attachment.Filename),
					slog.String("error", uploadErr.Error()),
				)
				continue
			}

			// 创建岗位关联
			if len(jobPositionIDs) > 0 {
				jobReq := &domain.CreateJobApplicationsReq{
					ResumeID:       resume.ID,
					JobPositionIDs: jobPositionIDs,
					Source:         &sourceStr,
				}
				if _, jobErr := u.jobApplicationUsecase.CreateJobApplications(ctx, jobReq); jobErr != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("创建岗位关联失败 %s: %s", resume.ID, jobErr.Error()))
					u.logger.Warn("邮箱简历创建岗位关联失败",
						slog.String("mailbox", setting.EmailAddress),
						slog.String("resume_id", resume.ID),
						slog.String("error", jobErr.Error()),
					)
				}
			}

			result.SuccessAttachments++
		}
	}

	// 更新游标
	if fetchResult.NextCursor != "" {
		if _, err := u.cursorRepo.Upsert(ctx, mailboxID, fetchResult.NextCursor, fetchResult.LastMessageID); err != nil {
			u.logger.Warn("更新邮箱游标失败",
				slog.String("mailbox", setting.EmailAddress),
				slog.String("error", err.Error()),
			)
		}
	}

	// 更新同步状态
	now := time.Now()
	lastError := ""
	retryCount := 0
	updateReq := &domain.UpdateSyncStatusRequest{
		LastSyncedAt: &now,
		LastError:    &lastError,
		RetryCount:   &retryCount,
	}
	if err := u.settingRepo.UpdateSyncStatus(ctx, mailboxID, updateReq); err != nil {
		u.logger.Warn("更新邮箱同步状态失败",
			slog.String("mailbox", setting.EmailAddress),
			slog.String("error", err.Error()),
		)
	}

	result.LastMessageID = fetchResult.LastMessageID
	result.Duration = time.Since(start)

	// 写入统计数据到数据库
	syncDate, _ := time.Parse("2006-01-02", now.Format("2006-01-02"))
	durationMs := int(result.Duration.Milliseconds())
	statisticReq := &domain.UpsertResumeMailboxStatisticRequest{
		MailboxID:          mailboxID,
		Date:               syncDate,
		SyncedEmails:       &result.ProcessedEmails,
		ParsedResumes:      &result.SuccessAttachments,
		FailedResumes:      &result.FailedAttachments,
		SkippedAttachments: &result.SkippedAttachments,
		LastSyncDurationMs: &durationMs,
	}

	if _, err := u.statisticRepo.Upsert(ctx, statisticReq); err != nil {
		u.logger.Warn("写入邮箱同步统计数据失败",
			slog.String("mailbox", setting.EmailAddress),
			slog.String("error", err.Error()),
		)
	}

	return result, nil
}

func (u *ResumeMailboxSyncUsecase) isSupportedAttachment(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	_, ok := supportedExtensions[ext]
	return ok
}

func (u *ResumeMailboxSyncUsecase) handleSyncFailure(ctx context.Context, setting *domain.ResumeMailboxSetting, syncErr error) {
	now := time.Now()
	errMsg := syncErr.Error()
	retry := setting.RetryCount + 1

	updateReq := &domain.UpdateSyncStatusRequest{
		LastSyncedAt: &now,
		LastError:    &errMsg,
		RetryCount:   &retry,
	}

	if err := u.settingRepo.UpdateSyncStatus(ctx, setting.ID, updateReq); err != nil {
		u.logger.Error("写入邮箱同步失败状态失败",
			slog.String("mailbox", setting.EmailAddress),
			slog.String("error", err.Error()),
		)
	}
}
