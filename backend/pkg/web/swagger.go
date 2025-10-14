package web

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

type SwaggerCtx struct {
	BasicAuth *BasicAuth
}

type BasicAuth struct {
	Username string
	Password string
}

type SwaggerOption func(*SwaggerCtx)

func WithBasicAuth(username, password string) SwaggerOption {
	return func(ctx *SwaggerCtx) {
		ctx.BasicAuth = &BasicAuth{
			Username: username,
			Password: password,
		}
	}
}

func (w *Web) Swagger(title, path, doc string, opts ...SwaggerOption) {
	ctx := &SwaggerCtx{}
	for _, opt := range opts {
		opt(ctx)
	}
	w.GET(path+"/swagger.json", &docHandler{doc: doc}, basicAuth(ctx))
	w.GET(path, &swaggerHandler{title: title}, basicAuth(ctx))
}

func basicAuth(ctx *SwaggerCtx) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if ctx.BasicAuth != nil {
				user, pass, ok := c.Request().BasicAuth()
				if !ok || user != ctx.BasicAuth.Username || pass != ctx.BasicAuth.Password {
					c.Response().Header().Set(echo.HeaderWWWAuthenticate, "Basic realm=Restricted")
					return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
				}
			}
			return next(c)
		}
	}
}

type docHandler struct {
	doc string
}

func (d *docHandler) Handle(ctx *Context) error {
	ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	ctx.Response().Header().Set(echo.HeaderContentDisposition, "attachment; filename=swagger.json")
	_, _ = ctx.Response().Write([]byte(d.doc))
	return nil
}

var html = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
    <title>%s</title>
    <!-- Embed elements Elements via Web Component -->
    <script src="https://unpkg.com/@stoplight/elements/web-components.min.js"></script>
    <link rel="stylesheet" href="https://unpkg.com/@stoplight/elements/styles.min.css">
  </head>
  <body>

    <elements-api
      apiDescriptionUrl="%s"
      router="hash"
      layout="sidebar"
    />

  </body>
</html>`

type swaggerHandler struct {
	title string
}

func (s *swaggerHandler) Handle(ctx *Context) error {
	_, _ = ctx.Response().Write(fmt.Appendf(nil, html, s.title, ctx.Request().URL.Path+"/swagger.json"))
	return nil
}
