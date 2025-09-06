//go:build wireinject
// +build wireinject

package main

import (
	"log/slog"

	"github.com/google/wire"

	"github.com/GoYoko/web"

	"github.com/ptonlix/whalehire/backend/config"
	"github.com/ptonlix/whalehire/backend/db"
	userV1 "github.com/ptonlix/whalehire/backend/internal/user/handler/v1"
	"github.com/ptonlix/whalehire/backend/pkg/version"
)

type Server struct {
	config  *config.Config
	web     *web.Web
	ent     *db.Client
	logger  *slog.Logger
	userV1  *userV1.UserHandler
	version *version.VersionInfo
}

func newServer() (*Server, error) {
	wire.Build(
		wire.Struct(new(Server), "*"),
		appSet,
	)
	return &Server{}, nil
}
