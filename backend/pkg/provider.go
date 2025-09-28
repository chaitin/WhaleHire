package pkg

import (
	"github.com/google/wire"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/text/language"

	"github.com/GoYoko/web"
	"github.com/GoYoko/web/locale"

	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/chaitin/WhaleHire/backend/errcode"
	mid "github.com/chaitin/WhaleHire/backend/internal/middleware"
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
)

func NewWeb(cfg *config.Config) *web.Web {
	w := web.New()
	l := locale.NewLocalizerWithFile(language.Chinese, errcode.LocalFS, []string{"locale.zh.toml"})
	w.SetLocale(l)
	w.Use(mid.RequestID())
	if cfg.Debug {
		w.Use(middleware.Logger())
	}
	return w
}
