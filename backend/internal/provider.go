package internal

import (
	"github.com/google/wire"

	generalagentV1 "github.com/chaitin/WhaleHire/backend/internal/general_agent/handler/v1"
	generalagentrepo "github.com/chaitin/WhaleHire/backend/internal/general_agent/repo"
	generalagentusecase "github.com/chaitin/WhaleHire/backend/internal/general_agent/usecase"
	middleware "github.com/chaitin/WhaleHire/backend/internal/middleware"
	userV1 "github.com/chaitin/WhaleHire/backend/internal/user/handler/v1"
	userrepo "github.com/chaitin/WhaleHire/backend/internal/user/repo"
	userusecase "github.com/chaitin/WhaleHire/backend/internal/user/usecase"
)

// NewAPIHandlers 创建 APIHandlers 实例
func NewAPIHandlers(
	userV1 *userV1.UserHandler,
	generalAgentV1 *generalagentV1.GeneralAgentHandler,
) *APIHandlers {
	return &APIHandlers{
		UserHandler:         userV1,
		GeneralAgentHandler: generalAgentV1,
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
	NewAPIHandlers,
)
