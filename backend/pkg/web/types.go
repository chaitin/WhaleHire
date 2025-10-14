package web

import "net/http"

type ErrorID string

const (
	ErrInternal   ErrorID = "err-internal"
	ErrBindParams ErrorID = "err-bind-params"
)

type Pagination struct {
	Page      int    `json:"page" query:"page" validate:"min=1" default:"1"`  // 分页
	Size      int    `json:"size" query:"size" validate:"min=1" default:"10"` // 每页多少条记录
	NextToken string `json:"next_token" query:"next_token"`                   // 下一页标识
}

type PageInfo struct {
	NextToken   string `json:"next_token,omitempty"`
	HasNextPage bool   `json:"has_next_page"`
	TotalCount  int64  `json:"total_count"`
}

type Resp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type Err struct {
	code int
	err  error
	id   ErrorID
	data map[string]any
}

func (e *Err) Error() string {
	if e.err == nil {
		return string(e.id)
	}
	return e.err.Error()
}

func (e *Err) Wrap(err error) *Err {
	e.err = err
	return e
}

func (e *Err) WithData(kv ...any) *Err {
	if len(kv)%2 != 0 {
		panic("[WEB] WithData kv must be even")
	}

	m := make(map[string]any)
	for i := 0; i < len(kv); i += 2 {
		m[kv[i].(string)] = kv[i+1]
	}
	e.data = m
	return e
}

func NewErr(code int, id ErrorID) *Err {
	return &Err{
		code: code,
		id:   id,
	}
}

func NewBadRequestErr(id ErrorID) *Err {
	return &Err{
		code: http.StatusBadRequest,
		id:   id,
	}
}

// BusinessErr 业务错误类型，支持自定义业务错误码和消息
type BusinessErr struct {
	code         int            // HTTP状态码
	businessCode int            // 业务错误码
	err          error          // 原始错误
	id           ErrorID        // 错误ID，用于国际化
	data         map[string]any // 附加数据
}

func (e *BusinessErr) Error() string {
	if e.err == nil {
		return string(e.id)
	}
	return e.err.Error()
}

func (e *BusinessErr) Wrap(err error) *BusinessErr {
	e.err = err
	return e
}

func (e *BusinessErr) WithData(kv ...any) *BusinessErr {
	if len(kv)%2 != 0 {
		panic("[WEB] WithData kv must be even")
	}

	m := make(map[string]any)
	for i := 0; i < len(kv); i += 2 {
		m[kv[i].(string)] = kv[i+1]
	}
	e.data = m
	return e
}

func (e *BusinessErr) HTTPCode() int {
	return e.code
}

func (e *BusinessErr) BusinessCode() int {
	return e.businessCode
}

func (e *BusinessErr) Data() map[string]any {
	return e.data
}

// NewBusinessErr 创建业务错误，HTTP状态码为200，但包含业务错误码
func NewBusinessErr(businessCode int, id ErrorID) *BusinessErr {
	return &BusinessErr{
		code:         http.StatusOK, // HTTP状态码为200
		businessCode: businessCode,
		id:           id,
	}
}

// NewBusinessErrWithHTTP 创建业务错误，可指定HTTP状态码
func NewBusinessErrWithHTTP(httpCode, businessCode int, id ErrorID) *BusinessErr {
	return &BusinessErr{
		code:         httpCode,
		businessCode: businessCode,
		id:           id,
	}
}

// NewBadRequestBusinessErr 创建400错误的业务错误
func NewBadRequestBusinessErr(businessCode int, id ErrorID) *BusinessErr {
	return &BusinessErr{
		code:         http.StatusBadRequest,
		businessCode: businessCode,
		id:           id,
	}
}
