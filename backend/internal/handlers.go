package internal

import (
	userV1 "github.com/ptonlix/whalehire/backend/internal/user/handler/v1"
)

// APIHandlers 包含所有API处理器
type APIHandlers struct {
	UserHandler *userV1.UserHandler
}
