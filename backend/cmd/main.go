package main

import (
	"context"

	"github.com/GoYoko/web"
	"github.com/google/wire"

	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/chaitin/WhaleHire/backend/docs"
	"github.com/chaitin/WhaleHire/backend/internal"
	"github.com/chaitin/WhaleHire/backend/pkg"
	"github.com/chaitin/WhaleHire/backend/pkg/service"
	"github.com/chaitin/WhaleHire/backend/pkg/store"
)

// @title WhaleHire API
// @version 1.0
// @description WhaleHire API
func main() {
	s, err := newServer()
	if err != nil {
		panic(err)
	}

	s.version.Print()
	s.logger.With("config", s.config).Debug("config")

	if s.config.Debug {
		s.web.Swagger("WhaleHire API", "/reference", string(docs.SwaggerJSON), web.WithBasicAuth("mc", "mc88"))
	}

	s.web.PrintRoutes()

	if err := store.MigrateSQL(s.config, s.logger); err != nil {
		panic(err)
	}

	if err := s.userV1.InitAdmin(); err != nil {
		panic(err)
	}

	svc := service.NewService(service.WithPprof())
	svc.Add(s)
	if err := svc.Run(); err != nil {
		panic(err)
	}
}

// Name implements service.Servicer.
func (s *Server) Name() string {
	return "Server"
}

// Start implements service.Servicer.
func (s *Server) Start() error {
	return s.web.Run(s.config.Server.Addr)
}

// Stop implements service.Servicer.
func (s *Server) Stop() error {
	return s.web.Echo().Shutdown(context.Background())
}

//lint:ignore U1000 unused for wire
var appSet = wire.NewSet(
	wire.FieldsOf(new(*config.Config), "Logger"),
	config.Init,
	pkg.Provider,
	internal.Provider,
)
