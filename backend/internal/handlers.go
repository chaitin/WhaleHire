package internal

import (
	departmentV1 "github.com/chaitin/WhaleHire/backend/internal/department/handler/v1"
	fileV1 "github.com/chaitin/WhaleHire/backend/internal/file/handler/v1"
	generalagentV1 "github.com/chaitin/WhaleHire/backend/internal/general_agent/handler/v1"
	jobapplicationV1 "github.com/chaitin/WhaleHire/backend/internal/job_application/handler/v1"
	jobprofileV1 "github.com/chaitin/WhaleHire/backend/internal/jobprofile/handler/v1"
	resumeV1 "github.com/chaitin/WhaleHire/backend/internal/resume/handler/v1"
	screeningV1 "github.com/chaitin/WhaleHire/backend/internal/screening/handler/v1"
	userV1 "github.com/chaitin/WhaleHire/backend/internal/user/handler/v1"
)

// APIHandlers 包含所有API处理器
type APIHandlers struct {
	UserHandler           *userV1.UserHandler
	GeneralAgentHandler   *generalagentV1.GeneralAgentHandler
	ResumeHandler         *resumeV1.ResumeHandler
	JobProfileHandler     *jobprofileV1.JobProfileHandler
	JobApplicationHandler *jobapplicationV1.JobApplicationHandler
	FileHandler           *fileV1.FileHandler
	DepartmentHandler     *departmentV1.DepartmentHandler
	ScreeningHandler      *screeningV1.ScreeningHandler
}
