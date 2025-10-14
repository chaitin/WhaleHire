package web

import (
	"net/http"
)

type bindHandler[T any] struct {
	fn   func(*Context, T) error
	opts []Option
}

// Handle implements Handler.
func (j *bindHandler[T]) Handle(ctx *Context) error {
	for _, opt := range j.opts {
		if err := opt(ctx); err != nil {
			return ctx.Failed(http.StatusBadRequest, ErrBindParams, ctx.err)
		}
	}

	if ctx.err != nil {
		return ctx.Failed(http.StatusBadRequest, ErrBindParams, ctx.err)
	}

	var t T
	if err := ctx.Bind(&t); err != nil {
		return ctx.Failed(http.StatusBadRequest, ErrBindParams, err)
	}
	if err := ctx.Validate(&t); err != nil {
		return ctx.Failed(http.StatusBadRequest, ErrBindParams, err)
	}
	if err := j.fn(ctx, t); err != nil {
		return ctx.Failed(http.StatusInternalServerError, ErrInternal, err)
	}
	return nil
}

func BindHandler[T any](fn func(*Context, T) error, opts ...Option) Handler {
	return &bindHandler[T]{
		fn:   fn,
		opts: opts,
	}
}

type baseHandler struct {
	fn   func(*Context) error
	opts []Option
}

// Handle implements Handler.
func (b *baseHandler) Handle(ctx *Context) error {
	for _, opt := range b.opts {
		if err := opt(ctx); err != nil {
			return err
		}
	}
	if err := b.fn(ctx); err != nil {
		return ctx.Failed(http.StatusInternalServerError, ErrInternal, err)
	}
	return nil
}

func BaseHandler(fn func(*Context) error, opts ...Option) Handler {
	return &baseHandler{
		fn:   fn,
		opts: opts,
	}
}
