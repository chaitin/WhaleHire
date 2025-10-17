package pkg

import (
	"github.com/google/wire"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/text/language"

	"github.com/chaitin/WhaleHire/backend/pkg/web"
	"github.com/chaitin/WhaleHire/backend/pkg/web/locale"

	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/chaitin/WhaleHire/backend/errcode"
	mid "github.com/chaitin/WhaleHire/backend/internal/middleware"
	"github.com/chaitin/WhaleHire/backend/pkg/docparser"
	"github.com/chaitin/WhaleHire/backend/pkg/ipdb"
	"github.com/chaitin/WhaleHire/backend/pkg/logger"
	"github.com/chaitin/WhaleHire/backend/pkg/session"
	"github.com/chaitin/WhaleHire/backend/pkg/store"
	"github.com/chaitin/WhaleHire/backend/pkg/store/s3"
	"github.com/chaitin/WhaleHire/backend/pkg/version"
)

var Provider = wire.NewSet(
	NewWeb,
	logger.NewLogger,
	store.NewEntDB,
	store.NewRedisCli,
	session.NewSession,
	ipdb.NewIPDB,
	version.NewVersionInfo,
	s3.NewMinioClient,
	docparser.NewDocumentParserServiceFromConfig,
)

func NewWeb(cfg *config.Config, auditMiddleware *mid.AuditMiddleware) *web.Web {
	w := web.New()
	l := locale.NewLocalizerWithFile(language.Chinese, errcode.LocalFS, []string{"locale.zh.toml"})
	w.SetLocale(l)
	w.Use(mid.RequestID())
	w.Use(auditMiddleware.Audit())
	if cfg.Debug {
		w.Use(middleware.Logger())
	}
	return w
}
