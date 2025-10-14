package web

import (
	"fmt"
	"sort"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/chaitin/WhaleHire/backend/pkg/web/locale"
	"github.com/chaitin/WhaleHire/backend/pkg/web/validate"
)

type Handler interface {
	Handle(*Context) error
}

type ErrHandle func(error) *Err

type Web struct {
	e      *echo.Echo
	locale *locale.Localizer
	es     []ErrHandle
}

func New() *Web {
	e := echo.New()
	e.Validator = validate.NewCustomValidator()
	return &Web{
		e:      e,
		locale: locale.NewLocalizer(),
		es:     make([]ErrHandle, 0),
	}
}

func (w *Web) Echo() *echo.Echo {
	return w.e
}

func (w *Web) Run(addr string) error {
	return w.e.Start(addr)
}

func (w *Web) SetValidator(v echo.Validator) {
	w.e.Validator = v
}

func (w *Web) SetLocale(locale *locale.Localizer) {
	w.locale = locale
}

func (w *Web) AddErrHandle(h ErrHandle) {
	w.es = append(w.es, h)
}

func (w *Web) Use(m ...echo.MiddlewareFunc) {
	w.e.Use(m...)
}

func (w *Web) POST(path string, h Handler, m ...echo.MiddlewareFunc) {
	w.e.POST(path, func(c echo.Context) error {
		ctx := &Context{
			Context: c,
			locale:  w.locale,
			es:      w.es,
		}
		return h.Handle(ctx)
	}, m...)
}

func (w *Web) GET(path string, h Handler, m ...echo.MiddlewareFunc) {
	w.e.GET(path, func(c echo.Context) error {
		ctx := &Context{
			Context: c,
			locale:  w.locale,
			es:      w.es,
		}
		return h.Handle(ctx)
	}, m...)
}

func (w *Web) PUT(path string, h Handler, m ...echo.MiddlewareFunc) {
	w.e.PUT(path, func(c echo.Context) error {
		ctx := &Context{
			Context: c,
			locale:  w.locale,
			es:      w.es,
		}
		return h.Handle(ctx)
	}, m...)
}

func (w *Web) Group(prefix string, m ...echo.MiddlewareFunc) *Group {
	g := w.e.Group(prefix, m...)
	return &Group{group: g, locale: w.locale, es: w.es}
}

func (w *Web) Routes() []*echo.Route {
	return w.e.Routes()
}

func (w *Web) PrintRoutes() {
	maxMethodLen := 0
	maxPathLen := 0
	clone := make([]*echo.Route, 0)
	for _, v := range w.e.Routes() {
		methodLen := len(v.Method)
		pathLen := len(v.Path)
		if methodLen > maxMethodLen {
			maxMethodLen = methodLen
		}
		if pathLen > maxPathLen {
			maxPathLen = pathLen
		}
		clone = append(clone, v)
	}
	width := maxMethodLen + maxPathLen + 2

	border := "─"
	topBorder := fmt.Sprintf("┌─%s Routes %s┐", border, strings.Repeat(border, width-9))
	bottomBorder := fmt.Sprintf("└%s┘", strings.Repeat(border, width+1))

	sort.Slice(clone, func(i, j int) bool {
		return len(clone[i].Method) > len(clone[j].Method) ||
			(len(clone[i].Method) == len(clone[j].Method) && len(clone[i].Path) > len(clone[j].Path))
	})

	fmt.Println(topBorder)
	for _, v := range clone {
		methodPadding := maxMethodLen - len(v.Method)
		fmt.Printf("│ %s%s %s\n", v.Method, strings.Repeat(" ", methodPadding), v.Path)
	}
	fmt.Println(bottomBorder)
}
