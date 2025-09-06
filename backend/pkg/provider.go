package pkg

import (
	"github.com/google/wire"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/text/language"

	"github.com/GoYoko/web"
	"github.com/GoYoko/web/locale"

	"github.com/ptonlix/whalehire/backend/config"
	"github.com/ptonlix/whalehire/backend/errcode"
	mid "github.com/ptonlix/whalehire/backend/internal/middleware"
	"github.com/ptonlix/whalehire/backend/pkg/ipdb"
	"github.com/ptonlix/whalehire/backend/pkg/logger"
	"github.com/ptonlix/whalehire/backend/pkg/session"
	"github.com/ptonlix/whalehire/backend/pkg/store"
	"github.com/ptonlix/whalehire/backend/pkg/version"
)

var Provider = wire.NewSet(
	NewWeb,
	logger.NewLogger,
	store.NewEntDB,
	store.NewRedisCli,
	session.NewSession,
	ipdb.NewIPDB,
	version.NewVersionInfo,
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
