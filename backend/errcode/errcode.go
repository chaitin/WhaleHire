package errcode

import (
	"embed"

	"github.com/chaitin/WhaleHire/backend/pkg/web"
)

//go:embed locale.*.toml
var LocalFS embed.FS

var (
	// ========== 通用/系统模块 (10000-19999) ==========
	ErrPermission   = web.NewBadRequestBusinessErr(10000, "err-permission")
	ErrInvalidParam = web.NewBadRequestBusinessErr(10001, "err-invalid-param")
	ErrMissKey      = web.NewBadRequestBusinessErr(10002, "err-miss-key")
	ErrOnlyAdmin    = web.NewBadRequestBusinessErr(10003, "err-only-admin")

	// ========== 用户管理模块 (20000-29999) ==========
	ErrUserNotFound        = web.NewBadRequestBusinessErr(20000, "err-user-not-found")
	ErrUserLock            = web.NewBadRequestBusinessErr(20001, "err-user-lock")
	ErrPassword            = web.NewBadRequestBusinessErr(20002, "err-password")
	ErrAccountAlreadyExist = web.NewBadRequestBusinessErr(20003, "err-account-already-exist")
	ErrInviteCodeInvalid   = web.NewBadRequestBusinessErr(20004, "err-invite-code-invalid")
	ErrEmailInvalid        = web.NewBadRequestBusinessErr(20005, "err-email-invalid")
	ErrOAuthStateInvalid   = web.NewBadRequestBusinessErr(20006, "err-oauth-state-invalid")
	ErrUnsupportedPlatform = web.NewBadRequestBusinessErr(20007, "err-unsupported-platform")
	ErrNotInvited          = web.NewBadRequestBusinessErr(20008, "err-not-invited")
	ErrDingtalkNotEnabled  = web.NewBadRequestBusinessErr(20009, "err-dingtalk-not-enabled")
	ErrCustomNotEnabled    = web.NewBadRequestBusinessErr(20010, "err-custom-not-enabled")
	ErrUserLimit           = web.NewBadRequestBusinessErr(20011, "err-user-limit")

	// ========== 简历管理模块 (30000-39999) ==========
	// 预留简历相关错误码

	// ========== 职位管理模块 (40000-49999) ==========
	ErrJobProfileRequired   = web.NewBadRequestBusinessErr(40000, "err-jobprofile-required")
	ErrJobSkillMetaRequired = web.NewBadRequestBusinessErr(40001, "err-jobskillmeta-required")
	ErrJobProfileHasResumes = web.NewBadRequestBusinessErr(40002, "err-jobprofile-has-resumes")

	// ========== 求职申请模块 (50000-59999) ==========
	// 预留求职申请相关错误码

	// ========== 部门管理模块 (60000-69999) ==========
	ErrDepartmentRequired = web.NewBadRequestBusinessErr(60000, "err-department-required")

	// ========== 筛选任务模块 (70000-79999) ==========
	ErrScreeningTaskNotFound     = web.NewBadRequestBusinessErr(70000, "err-screening-task-not-found")
	ErrScreeningTaskRunning      = web.NewBadRequestBusinessErr(70001, "err-screening-task-running")
	ErrScreeningTaskDeleteFailed = web.NewBadRequestBusinessErr(70002, "err-screening-task-delete-failed")

	// ========== 通知设置模块 (80000-89999) ==========
	ErrNotificationSettingCreateFailed = web.NewBadRequestBusinessErr(80000, "err-notification-setting-create-failed")
	ErrNotificationSettingGetFailed    = web.NewBadRequestBusinessErr(80001, "err-notification-setting-get-failed")
	ErrNotificationSettingListFailed   = web.NewBadRequestBusinessErr(80002, "err-notification-setting-list-failed")
	ErrNotificationSettingUpdateFailed = web.NewBadRequestBusinessErr(80003, "err-notification-setting-update-failed")
	ErrNotificationSettingDeleteFailed = web.NewBadRequestBusinessErr(80004, "err-notification-setting-delete-failed")
	ErrNotificationSettingNotFound     = web.NewBadRequestBusinessErr(80005, "err-notification-setting-not-found")

	// ========== 简历邮箱模块 (90000-99999) ==========
	ErrResumeMailboxSettingNotFound   = web.NewBadRequestBusinessErr(90000, "err-resume-mailbox-setting-not-found")
	ErrResumeMailboxStatisticNotFound = web.NewBadRequestBusinessErr(90003, "err-resume-mailbox-statistic-not-found")
	ErrInvalidCredentials             = web.NewBadRequestBusinessErr(90001, "err-invalid-credentials")
	ErrMailboxConnectionFailed        = web.NewBadRequestBusinessErr(90002, "err-mailbox-connection-failed")
)
