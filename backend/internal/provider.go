package internal

import (
	"github.com/google/wire"

	middleware "github.com/ptonlix/whalehire/backend/internal/middleware"
	userV1 "github.com/ptonlix/whalehire/backend/internal/user/handler/v1"
	userrepo "github.com/ptonlix/whalehire/backend/internal/user/repo"
	userusecase "github.com/ptonlix/whalehire/backend/internal/user/usecase"
)

// NewAPIHandlers 创建 APIHandlers 实例
func NewAPIHandlers(
	userV1 *userV1.UserHandler,
) *APIHandlers {
	return &APIHandlers{
		UserHandler: userV1,
	}
}

var Provider = wire.NewSet(
	middleware.NewAuthMiddleware,
	middleware.NewActiveMiddleware,
	middleware.NewReadOnlyMiddleware,
	userV1.NewUserHandler,
	userrepo.NewUserRepo,
	userusecase.NewUserUsecase,
)
