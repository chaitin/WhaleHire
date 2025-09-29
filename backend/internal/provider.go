package internal

import (
	"github.com/google/wire"

	fileV1 "github.com/chaitin/WhaleHire/backend/internal/file/handler/v1"
	fileusecase "github.com/chaitin/WhaleHire/backend/internal/file/usecase"
	generalagentV1 "github.com/chaitin/WhaleHire/backend/internal/general_agent/handler/v1"
	generalagentrepo "github.com/chaitin/WhaleHire/backend/internal/general_agent/repo"
	generalagentusecase "github.com/chaitin/WhaleHire/backend/internal/general_agent/usecase"
	middleware "github.com/chaitin/WhaleHire/backend/internal/middleware"
	resumeV1 "github.com/chaitin/WhaleHire/backend/internal/resume/handler/v1"
	resumerepo "github.com/chaitin/WhaleHire/backend/internal/resume/repo"
	resumeservice "github.com/chaitin/WhaleHire/backend/internal/resume/service"
	resumeusecase "github.com/chaitin/WhaleHire/backend/internal/resume/usecase"
	userV1 "github.com/chaitin/WhaleHire/backend/internal/user/handler/v1"
	userrepo "github.com/chaitin/WhaleHire/backend/internal/user/repo"
	userusecase "github.com/chaitin/WhaleHire/backend/internal/user/usecase"
)

// NewAPIHandlers 创建 APIHandlers 实例
func NewAPIHandlers(
	userV1 *userV1.UserHandler,
	generalAgentV1 *generalagentV1.GeneralAgentHandler,
	resumeV1 *resumeV1.ResumeHandler,
	fileV1 *fileV1.FileHandler,
) *APIHandlers {
	return &APIHandlers{
		UserHandler:         userV1,
		GeneralAgentHandler: generalAgentV1,
		ResumeHandler:       resumeV1,
		FileHandler:         fileV1,
	}
}

var Provider = wire.NewSet(
	middleware.NewAuthMiddleware,
	middleware.NewActiveMiddleware,
	middleware.NewReadOnlyMiddleware,
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
	fileusecase.NewFileUsecase,
	fileV1.NewFileHandler,
	NewAPIHandlers,
)
