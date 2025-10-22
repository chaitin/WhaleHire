package main

import (
	"context"

	"github.com/chaitin/WhaleHire/backend/pkg"
	"github.com/chaitin/WhaleHire/backend/pkg/web"
	"github.com/google/wire"

	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/chaitin/WhaleHire/backend/docs"
	"github.com/chaitin/WhaleHire/backend/internal"
	"github.com/chaitin/WhaleHire/backend/pkg/service"
	"github.com/chaitin/WhaleHire/backend/pkg/store"

	// 新增：将通知 Worker 作为一个独立的服务加入生命周期管理
	"github.com/chaitin/WhaleHire/backend/internal/notification/worker"
	resumemailboxscheduler "github.com/chaitin/WhaleHire/backend/internal/resume_mailbox/scheduler"
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

	// 记录外部服务配置信息
	s.logger.Info("外部服务配置检查开始")
	s.logger.Info("文档解析服务配置",
		"base_url", s.config.DocumentParser.BaseURL,
		"api_key_length", len(s.config.DocumentParser.APIKey),
		"timeout", s.config.DocumentParser.Timeout)
	s.logger.Info("通用代理服务配置",
		"base_url", s.config.GeneralAgent.LLM.BaseURL,
		"api_key_length", len(s.config.GeneralAgent.LLM.APIKey),
		"model", s.config.GeneralAgent.LLM.ModelName)
	s.logger.Info("嵌入服务配置",
		"api_endpoint", s.config.Embedding.APIEndpoint,
		"api_key_length", len(s.config.Embedding.APIKey),
		"model_name", s.config.Embedding.ModelName)
	s.logger.Info("外部服务配置检查完成")

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

	s.logger.Info("服务启动完成，开始监听请求", "addr", s.config.Server.Addr)
	svc := service.NewService(service.WithPprof())
	svc.Add(s)
	svc.Add(resumemailboxscheduler.NewServicer(s.resumeMailboxScheduler))
	// 新增：将通知 Worker 封装为 Servicer，交由 Service 管理
	svc.Add(worker.NewServicer(s.notificationWorker))
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
	internal.QueueProvider,
)
