package web

import (
	"github.com/labstack/echo/v4"

	"github.com/chaitin/WhaleHire/backend/pkg/web/locale"
)

type Group struct {
	group  *echo.Group
	locale *locale.Localizer
	es     []ErrHandle
}

func (g *Group) Use(m ...echo.MiddlewareFunc) {
	g.group.Use(m...)
}

func (g *Group) POST(path string, h Handler, m ...echo.MiddlewareFunc) {
	g.group.POST(path, func(c echo.Context) error {
		ctx := &Context{
			Context: c,
			locale:  g.locale,
			es:      g.es,
		}
		return h.Handle(ctx)
	}, m...)
}

func (g *Group) GET(path string, h Handler, m ...echo.MiddlewareFunc) {
	g.group.GET(path, func(c echo.Context) error {
		ctx := &Context{
			Context: c,
			locale:  g.locale,
			es:      g.es,
		}
		return h.Handle(ctx)
	}, m...)
}

func (g *Group) PUT(path string, h Handler, m ...echo.MiddlewareFunc) {
	g.group.PUT(path, func(c echo.Context) error {
		ctx := &Context{
			Context: c,
			locale:  g.locale,
			es:      g.es,
		}
		return h.Handle(ctx)
	}, m...)
}

func (g *Group) DELETE(path string, h Handler, m ...echo.MiddlewareFunc) {
	g.group.DELETE(path, func(c echo.Context) error {
		ctx := &Context{
			Context: c,
			locale:  g.locale,
			es:      g.es,
		}
		return h.Handle(ctx)
	}, m...)
}

func (g *Group) Any(path string, h Handler, m ...echo.MiddlewareFunc) {
	g.group.Any(path, func(c echo.Context) error {
		ctx := &Context{
			Context: c,
			locale:  g.locale,
			es:      g.es,
		}
		return h.Handle(ctx)
	}, m...)
}

func (g *Group) Group(prefix string, middleware ...echo.MiddlewareFunc) *Group {
	gg := g.group.Group(prefix, middleware...)
	g.group = gg
	return g
}
