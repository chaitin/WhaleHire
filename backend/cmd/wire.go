//go:build wireinject
// +build wireinject

package main

import (
	"log/slog"

	"github.com/google/wire"

	"github.com/GoYoko/web"

	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/chaitin/WhaleHire/backend/db"
	fileV1 "github.com/chaitin/WhaleHire/backend/internal/file/handler/v1"
	generalagentV1 "github.com/chaitin/WhaleHire/backend/internal/general_agent/handler/v1"
	userV1 "github.com/chaitin/WhaleHire/backend/internal/user/handler/v1"
	"github.com/chaitin/WhaleHire/backend/pkg/version"
)

type Server struct {
	config         *config.Config
	web            *web.Web
	ent            *db.Client
	logger         *slog.Logger
	userV1         *userV1.UserHandler
	generalagentV1 *generalagentV1.GeneralAgentHandler
	fileV1         *fileV1.FileHandler
	version        *version.VersionInfo
}

func newServer() (*Server, error) {
	wire.Build(
		wire.Struct(new(Server), "*"),
		appSet,
	)
	return &Server{}, nil
}
