package internal

import (
	"github.com/google/wire"

	"github.com/chaitin/WhaleHire/backend/domain"
	auditV1 "github.com/chaitin/WhaleHire/backend/internal/audit/handler/v1"
	auditrepo "github.com/chaitin/WhaleHire/backend/internal/audit/repo"
	auditusecase "github.com/chaitin/WhaleHire/backend/internal/audit/usecase"
	departmentV1 "github.com/chaitin/WhaleHire/backend/internal/department/handler/v1"
	departmentrepo "github.com/chaitin/WhaleHire/backend/internal/department/repo"
	departmentusecase "github.com/chaitin/WhaleHire/backend/internal/department/usecase"
	fileV1 "github.com/chaitin/WhaleHire/backend/internal/file/handler/v1"
	fileusecase "github.com/chaitin/WhaleHire/backend/internal/file/usecase"
	generalagentV1 "github.com/chaitin/WhaleHire/backend/internal/general_agent/handler/v1"
	generalagentrepo "github.com/chaitin/WhaleHire/backend/internal/general_agent/repo"
	generalagentusecase "github.com/chaitin/WhaleHire/backend/internal/general_agent/usecase"
	jobapplicationV1 "github.com/chaitin/WhaleHire/backend/internal/job_application/handler/v1"
	jobapplicationrepo "github.com/chaitin/WhaleHire/backend/internal/job_application/repo"
	jobapplicationusecase "github.com/chaitin/WhaleHire/backend/internal/job_application/usecase"
	jobprofileV1 "github.com/chaitin/WhaleHire/backend/internal/jobprofile/handler/v1"
	jobprofilerepo "github.com/chaitin/WhaleHire/backend/internal/jobprofile/repo"
	jobprofileservice "github.com/chaitin/WhaleHire/backend/internal/jobprofile/service"
	jobprofileusecase "github.com/chaitin/WhaleHire/backend/internal/jobprofile/usecase"
	middleware "github.com/chaitin/WhaleHire/backend/internal/middleware"
	notificationadapter "github.com/chaitin/WhaleHire/backend/internal/notification/adapter"
	notificationV1 "github.com/chaitin/WhaleHire/backend/internal/notification/handler/v1"
	notificationrepo "github.com/chaitin/WhaleHire/backend/internal/notification/repo"
	notificationusecase "github.com/chaitin/WhaleHire/backend/internal/notification/usecase"
	notificationworker "github.com/chaitin/WhaleHire/backend/internal/notification/worker"
	resumeV1 "github.com/chaitin/WhaleHire/backend/internal/resume/handler/v1"
	resumerepo "github.com/chaitin/WhaleHire/backend/internal/resume/repo"
	resumeservice "github.com/chaitin/WhaleHire/backend/internal/resume/service"
	resumeusecase "github.com/chaitin/WhaleHire/backend/internal/resume/usecase"
	resumemailboxadapter "github.com/chaitin/WhaleHire/backend/internal/resume_mailbox/adapter"
	resumemailV1 "github.com/chaitin/WhaleHire/backend/internal/resume_mailbox/handler/v1"
	resumemailboxsettingrepo "github.com/chaitin/WhaleHire/backend/internal/resume_mailbox/repo"
	resumemailboxscheduler "github.com/chaitin/WhaleHire/backend/internal/resume_mailbox/scheduler"
	resumemailboxsettingusecase "github.com/chaitin/WhaleHire/backend/internal/resume_mailbox/usecase"
	screeningV1 "github.com/chaitin/WhaleHire/backend/internal/screening/handler/v1"
	screeningrepo "github.com/chaitin/WhaleHire/backend/internal/screening/repo"
	screeningservice "github.com/chaitin/WhaleHire/backend/internal/screening/service"
	screeningusecase "github.com/chaitin/WhaleHire/backend/internal/screening/usecase"
	userV1 "github.com/chaitin/WhaleHire/backend/internal/user/handler/v1"
	userrepo "github.com/chaitin/WhaleHire/backend/internal/user/repo"
	userusecase "github.com/chaitin/WhaleHire/backend/internal/user/usecase"
)

// NewAPIHandlers 创建 APIHandlers 实例
func NewAPIHandlers(
	userV1 *userV1.UserHandler,
	generalAgentV1 *generalagentV1.GeneralAgentHandler,
	resumeV1 *resumeV1.ResumeHandler,
	jobProfileV1 *jobprofileV1.JobProfileHandler,
	jobApplicationV1 *jobapplicationV1.JobApplicationHandler,
	fileV1 *fileV1.FileHandler,
	departmentV1 *departmentV1.DepartmentHandler,
	screeningV1 *screeningV1.ScreeningHandler,
	auditV1 *auditV1.AuditHandler,
) *APIHandlers {
	return &APIHandlers{
		UserHandler:           userV1,
		GeneralAgentHandler:   generalAgentV1,
		ResumeHandler:         resumeV1,
		JobProfileHandler:     jobProfileV1,
		JobApplicationHandler: jobApplicationV1,
		FileHandler:           fileV1,
		DepartmentHandler:     departmentV1,
		ScreeningHandler:      screeningV1,
		AuditHandler:          auditV1,
	}
}

var Provider = wire.NewSet(
	middleware.NewAuthMiddleware,
	middleware.NewActiveMiddleware,
	middleware.NewReadOnlyMiddleware,
	middleware.NewAuditMiddleware,
	auditrepo.NewAuditRepo,
	auditusecase.NewAuditUsecase,
	auditV1.NewAuditHandler,
	userV1.NewUserHandler,
	userrepo.NewUserRepo,
	userusecase.NewUserUsecase,
	generalagentV1.NewGeneralAgentHandler,
	generalagentrepo.NewGeneralAgentRepo,
	generalagentusecase.NewGeneralAgentUsecase,
	resumeV1.NewResumeHandler,
	resumerepo.NewResumeRepo,
	resumeusecase.NewResumeUsecase,
	resumeservice.NewParserService,
	resumeservice.NewStorageService,
	jobprofileV1.NewJobProfileHandler,
	jobprofilerepo.NewJobProfileRepo,
	jobprofilerepo.NewJobSkillMetaRepo,
	jobprofileservice.NewJobProfileParserService,
	jobprofileservice.NewJobProfilePromptService,
	jobprofileusecase.NewJobProfileUsecase,
	jobapplicationV1.NewJobApplicationHandler,
	jobapplicationrepo.NewJobApplicationRepo,
	jobapplicationusecase.NewJobApplicationUsecase,
	departmentrepo.NewDepartmentRepo,
	departmentusecase.NewDepartmentUsecase,
	departmentV1.NewDepartmentHandler,
	fileusecase.NewFileUsecase,
	fileV1.NewFileHandler,
	screeningrepo.NewScreeningRepo,
	screeningrepo.NewScreeningNodeRunRepo,
	screeningservice.NewMatchingService,
	screeningusecase.NewScreeningUsecase,
	screeningV1.NewScreeningHandler,
	notificationV1.NewNotificationSettingHandler,
	notificationrepo.NewNotificationEventRepo,
	notificationrepo.NewNotificationSettingRepo,
	notificationusecase.NewNotificationUsecase,
	notificationusecase.NewNotificationSettingUsecase,
	notificationadapter.NewDingTalkAdapter,
	notificationworker.NewNotificationWorker,
	resumemailboxadapter.NewAdapterFactory,
	resumemailboxsettingrepo.NewResumeMailboxSettingRepo,
	resumemailboxsettingrepo.NewResumeMailboxCursorRepo,
	resumemailboxsettingrepo.NewResumeMailboxStatisticRepo,
	resumemailboxscheduler.NewScheduler,
	NewResumeMailboxScheduler,
	resumemailboxsettingusecase.NewResumeMailboxSettingUsecase,
	resumemailboxsettingusecase.NewResumeMailboxSyncUsecase,
	resumemailboxsettingusecase.NewResumeMailboxStatisticUsecase,
	resumemailV1.NewResumeMailboxSettingHandler,
	resumemailV1.NewResumeMailboxStatisticHandler,
	NewAPIHandlers,
)

// NewResumeMailboxScheduler 创建ResumeMailboxScheduler接口实现
func NewResumeMailboxScheduler(scheduler *resumemailboxscheduler.Scheduler) domain.ResumeMailboxScheduler {
	return scheduler
}
