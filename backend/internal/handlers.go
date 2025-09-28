package internal

import (
	fileV1 "github.com/chaitin/WhaleHire/backend/internal/file/handler/v1"
	generalagentV1 "github.com/chaitin/WhaleHire/backend/internal/general_agent/handler/v1"
	userV1 "github.com/chaitin/WhaleHire/backend/internal/user/handler/v1"
)

// APIHandlers 包含所有API处理器
type APIHandlers struct {
	UserHandler         *userV1.UserHandler
	GeneralAgentHandler *generalagentV1.GeneralAgentHandler
	FileHandler         *fileV1.FileHandler
}
