package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"

	"github.com/chaitin/WhaleHire/backend/consts"
)

type ActiveMiddleware struct {
	redis  *redis.Client
	logger *slog.Logger
}

func NewActiveMiddleware(redis *redis.Client, logger *slog.Logger) *ActiveMiddleware {
	return &ActiveMiddleware{
		redis:  redis,
		logger: logger,
	}
}

func (a *ActiveMiddleware) Active(scope string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			switch scope {
			case "admin":
				if user := GetAdmin(c); user != nil {
					if err := a.redis.Set(context.Background(), fmt.Sprintf(consts.AdminActiveKeyFmt, user.ID), time.Now().Unix(), 0).Err(); err != nil {
						a.logger.With("error", err).ErrorContext(c.Request().Context(), "failed to set admin active status in Redis")
					}
				}
			case "user":
				if user := GetUser((c)); user != nil {
					if err := a.redis.Set(context.Background(), fmt.Sprintf(consts.UserActiveKeyFmt, user.ID), time.Now().Unix(), 0).Err(); err != nil {
						a.logger.With("error", err).ErrorContext(c.Request().Context(), "failed to set user active status in Redis")
					}
				}
			}

			return next(c)
		}
	}
}
