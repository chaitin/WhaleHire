//go:build wireinject
// +build wireinject

package main

import (
	"log/slog"

	"github.com/google/wire"

	"github.com/chaitin/WhaleHire/backend/pkg/web"

	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/chaitin/WhaleHire/backend/db"
	departmentV1 "github.com/chaitin/WhaleHire/backend/internal/department/handler/v1"
	fileV1 "github.com/chaitin/WhaleHire/backend/internal/file/handler/v1"
	generalagentV1 "github.com/chaitin/WhaleHire/backend/internal/general_agent/handler/v1"
	jobapplicationV1 "github.com/chaitin/WhaleHire/backend/internal/job_application/handler/v1"
	jobprofileV1 "github.com/chaitin/WhaleHire/backend/internal/jobprofile/handler/v1"
	resumeV1 "github.com/chaitin/WhaleHire/backend/internal/resume/handler/v1"
	screeningV1 "github.com/chaitin/WhaleHire/backend/internal/screening/handler/v1"
	userV1 "github.com/chaitin/WhaleHire/backend/internal/user/handler/v1"
	"github.com/chaitin/WhaleHire/backend/pkg/version"
)

type Server struct {
	config           *config.Config
	web              *web.Web
	ent              *db.Client
	logger           *slog.Logger
	userV1           *userV1.UserHandler
	resumeV1         *resumeV1.ResumeHandler
	generalagentV1   *generalagentV1.GeneralAgentHandler
	jobprofileV1     *jobprofileV1.JobProfileHandler
	departmentV1     *departmentV1.DepartmentHandler
	jobapplicationV1 *jobapplicationV1.JobApplicationHandler
	screeningV1      *screeningV1.ScreeningHandler
	fileV1           *fileV1.FileHandler
	version          *version.VersionInfo
}

func newServer() (*Server, error) {
	wire.Build(
		wire.Struct(new(Server), "*"),
		appSet,
	)
	return &Server{}, nil
}
