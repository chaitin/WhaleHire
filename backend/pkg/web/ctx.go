package web

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/rs/xid"

	"github.com/chaitin/WhaleHire/backend/pkg/web/locale"
)

var logger = slog.Default()

type Context struct {
	echo.Context
	err    error
	locale *locale.Localizer
	page   *Pagination
	es     []ErrHandle
}

type Option func(*Context) error

func WithPage() Option {
	return func(ctx *Context) error {
		p := ctx.QueryParam("page")
		size := ctx.QueryParam("size")
		nt := ctx.QueryParam("next_token")
		i, _ := strconv.Atoi(p)
		pgi, _ := strconv.Atoi(size)
		page := &Pagination{
			Page:      i,
			Size:      pgi,
			NextToken: nt,
		}
		if page.NextToken == "" && page.Page == 0 {
			page.Page = 1
		}
		if page.Size == 0 {
			page.Size = 10
		}
		if page.Size > 100 {
			page.Size = 100
		}
		ctx.page = page
		return nil
	}
}

func (c *Context) Page() *Pagination {
	return c.page
}

func (c *Context) ErrMsg(id ErrorID, data map[string]any) (string, error) {
	return c.locale.Message(c.Request().Header.Get("Accept-Language"), string(id), data)
}

func (c *Context) Failed(code int, id ErrorID, err error) error {
	for _, handle := range c.es {
		if e := handle(err); e != nil {
			return c.failed(e.code, e.id, err, e.data)
		}
	}

	// 处理BusinessErr类型
	if e, ok := err.(*BusinessErr); ok {
		fmt.Printf("BusinessErr: %v\n", e)
		return c.businessFailed(e)
	}

	if e, ok := err.(*Err); ok {
		if e.err == nil {
			if msg, ee := c.ErrMsg(e.id, e.data); ee != nil {
				e.err = err
			} else {
				e.err = errors.New(msg)
			}
		}
		return c.failed(e.code, e.id, e.err, e.data)
	}
	return c.failed(code, id, err, nil)
}

func (c *Context) failed(code int, id ErrorID, err error, data map[string]any) error {
	traceID := xid.New().String()
	msg, ee := c.ErrMsg(id, data)
	if ee != nil {
		msg = "Internal Server Error"
	}
	logger.With("trace_id", traceID).With("err", err).Warn("request failed")
	c.Set("err-msg", err.Error())
	return c.JSON(code, Resp{
		Code:    code,
		Message: fmt.Sprintf("%s [trace_id: %s]", msg, traceID),
	})
}

func (c *Context) Success(r any) error {
	return c.JSON(http.StatusOK, Resp{
		Code:    0,
		Message: "success",
		Data:    r,
	})
}

// businessFailed 处理业务错误
func (c *Context) businessFailed(err *BusinessErr) error {
	traceID := xid.New().String()
	msg, ee := c.ErrMsg(err.id, err.data)
	if ee != nil {
		msg = "Business Error"
	}

	logger.With("trace_id", traceID).With("business_code", err.BusinessCode()).With("err", err).Warn("business error")
	c.Set("err-msg", err.Error())

	return c.JSON(err.HTTPCode(), Resp{
		Code:    err.BusinessCode(),
		Message: fmt.Sprintf("%s [trace_id: %s]", msg, traceID),
	})
}
