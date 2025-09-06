package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/ptonlix/whalehire/backend/config"
)

type ReadOnlyMiddleware struct {
	cfg *config.Config
}

func NewReadOnlyMiddleware(cfg *config.Config) *ReadOnlyMiddleware {
	return &ReadOnlyMiddleware{cfg: cfg}
}

func (m *ReadOnlyMiddleware) Guard() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if m.cfg.ReadOnly && c.Request().Method != http.MethodGet {
				return c.JSON(http.StatusOK, echo.Map{
					"code":    -1,
					"message": "只读模式下不支持该操作",
				})
			}
			return next(c)
		}
	}
}
