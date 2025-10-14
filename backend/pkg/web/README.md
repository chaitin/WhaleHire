# web

web 是基于 Echo 的 web 框架，为 WhaleHire 项目提供了一套完整的 web 开发解决方案。

## 主要功能

- 参数自动绑定和校验（支持默认值）
- 统一的错误处理和国际化
- 分页功能支持
- 路由分组管理
- 集成 Swagger UI 文档
- 业务错误码支持
- 自定义错误处理器

## 快速开始

### 1. 基本使用

```go
package main

import (
    "github.com/chaitin/WhaleHire/backend/pkg/web"
)

func main() {
    w := web.New()
    
    // 注册路由
    w.GET("/", web.BaseHandler(Hello))
    
    // 启动服务器
    if err := w.Run(":8080"); err != nil {
        panic(err)
    }
}

func Hello(ctx *web.Context) error {
    return ctx.Success("Hello, World!")
}
```

### 2. 带参数的处理器

```go
type UserRequest struct {
    ID   string `json:"id" query:"id" validate:"required"`
    Name string `json:"name" form:"name" default:"Anonymous"`
    Age  int    `json:"age" form:"age" validate:"min=0,max=120" default:"18"`
}

w.POST("/user", web.BindHandler(CreateUser))

func CreateUser(ctx *web.Context, req UserRequest) error {
    // 参数已自动绑定、验证并应用默认值
    user := map[string]interface{}{
        "id":   req.ID,
        "name": req.Name,
        "age":  req.Age,
    }
    return ctx.Success(user)
}
```

### 3. 分页支持

```go
w.GET("/users", web.BindHandler(ListUsers, web.WithPage()))

func ListUsers(ctx *web.Context, req ListUsersRequest) error {
    page := ctx.Page()
    // page.Page: 当前页码（默认1）
    // page.Size: 每页大小（默认10，最大100）
    // page.NextToken: 下一页标识
    
    users := getUsersWithPagination(page.Page, page.Size)
    return ctx.Success(users)
}
```

### 4. 路由分组

```go
// API v1 分组
v1 := w.Group("/api/v1")
v1.GET("/users", web.BindHandler(ListUsers))
v1.POST("/users", web.BindHandler(CreateUser))

// 管理员分组（带中间件）
admin := w.Group("/admin", authMiddleware)
admin.GET("/stats", web.BaseHandler(GetStats))
admin.DELETE("/users/:id", web.BindHandler(DeleteUser))
```

### 5. Swagger 文档

```go
//go:embed swagger.json
var swaggerJSON []byte

// 基本配置
w.Swagger("WhaleHire API", "/docs", string(swaggerJSON))

// 带基础认证的配置
w.Swagger("WhaleHire API", "/docs", string(swaggerJSON), 
    web.WithBasicAuth("admin", "password"))
```

## 错误处理

### 1. 定义国际化错误消息

创建 `locale.zh.toml` 文件：
```toml
[err-user-not-found]
other = "用户不存在"

[err-permission-denied]
other = "权限不足"

[err-invalid-email]
other = "邮箱格式无效: {{.email}}"
```

创建 `locale.en.toml` 文件：
```toml
[err-user-not-found]
other = "User not found"

[err-permission-denied]
other = "Permission denied"

[err-invalid-email]
other = "Invalid email format: {{.email}}"
```

### 2. 注册国际化配置

```go
import (
    "embed"
    "golang.org/x/text/language"
    "github.com/chaitin/WhaleHire/backend/pkg/web/locale"
)

//go:embed locale.*.toml
var LocaleFS embed.FS

func main() {
    w := web.New()
    
    // 设置国际化
    l := locale.NewLocalizerWithFile(
        language.Chinese, 
        LocaleFS, 
        []string{"locale.zh.toml", "locale.en.toml"},
    )
    w.SetLocale(l)
}
```

### 3. 定义错误类型

```go
import (
    "net/http"
    "github.com/chaitin/WhaleHire/backend/pkg/web"
)

var (
    ErrUserNotFound     = web.NewErr(http.StatusNotFound, "err-user-not-found")
    ErrPermissionDenied = web.NewErr(http.StatusForbidden, "err-permission-denied")
    ErrInvalidEmail     = web.NewBadRequestErr("err-invalid-email")
)

// 业务错误（带业务错误码）
var (
    ErrBusinessLogic = web.NewBusinessErr(1001, "err-business-logic")
    ErrDataConflict  = web.NewBusinessErrWithHTTP(http.StatusConflict, 1002, "err-data-conflict")
)
```

### 4. 使用错误

```go
func GetUser(ctx *web.Context, req GetUserRequest) error {
    user, err := userService.GetByID(req.ID)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return ErrUserNotFound.Wrap(err)
        }
        return err
    }
    
    return ctx.Success(user)
}

func ValidateEmail(ctx *web.Context, req EmailRequest) error {
    if !isValidEmail(req.Email) {
        return ErrInvalidEmail.WithData("email", req.Email)
    }
    
    return ctx.Success("Email is valid")
}
```

### 5. 自定义错误处理器

```go
w.AddErrHandle(func(err error) *web.Err {
    // 处理特定类型的错误
    if dbErr, ok := err.(*DatabaseError); ok {
        return web.NewErr(http.StatusServiceUnavailable, "err-database-unavailable")
    }
    return nil // 返回 nil 表示不处理此错误
})
```

## 高级功能

### 1. 自定义验证器

```go
import (
    "github.com/go-playground/validator/v10"
    "github.com/chaitin/WhaleHire/backend/pkg/web/validate"
)

validator := validate.NewCustomValidator()
w.SetValidator(validator)
```

### 2. 中间件使用

```go
import "github.com/labstack/echo/v4/middleware"

// 全局中间件
w.Use(middleware.Logger())
w.Use(middleware.Recover())
w.Use(middleware.CORS())

// 路由组中间件
api := w.Group("/api", authMiddleware, rateLimitMiddleware)
```

### 3. 响应格式

成功响应：
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "id": "123",
        "name": "John Doe"
    }
}
```

错误响应：
```json
{
    "code": 404,
    "message": "用户不存在 [trace_id: abc123]"
}
```

业务错误响应：
```json
{
    "code": 400,
    "business_code": 1001,
    "message": "业务逻辑错误 [trace_id: abc123]"
}
```

### 4. 分页响应

```go
type PagedResponse struct {
    Data     []User    `json:"data"`
    PageInfo *PageInfo `json:"page_info"`
}

type PageInfo struct {
    NextToken   string `json:"next_token,omitempty"`
    HasNextPage bool   `json:"has_next_page"`
    TotalCount  int64  `json:"total_count"`
}
```

## API 参考

### Web 实例方法

- `New() *Web` - 创建新的 Web 实例
- `Run(addr string) error` - 启动服务器
- `Echo() *echo.Echo` - 获取底层 Echo 实例
- `SetValidator(v echo.Validator)` - 设置自定义验证器
- `SetLocale(locale *locale.Localizer)` - 设置国际化配置
- `AddErrHandle(h ErrHandle)` - 添加错误处理器
- `Use(m ...echo.MiddlewareFunc)` - 添加全局中间件

### HTTP 方法

- `GET(path string, h Handler, m ...echo.MiddlewareFunc)`
- `POST(path string, h Handler, m ...echo.MiddlewareFunc)`
- `PUT(path string, h Handler, m ...echo.MiddlewareFunc)`
- `Group(prefix string, m ...echo.MiddlewareFunc) *Group`

### 处理器类型

- `BaseHandler(fn func(*Context) error, opts ...Option) Handler` - 无参数处理器
- `BindHandler[T any](fn func(*Context, T) error, opts ...Option) Handler` - 带参数处理器

### Context 方法

- `Success(data any) error` - 返回成功响应
- `Failed(code int, id ErrorID, err error) error` - 返回错误响应
- `Page() *Pagination` - 获取分页信息
- `ErrMsg(id ErrorID, data map[string]any) (string, error)` - 获取错误消息

### 选项

- `WithPage() Option` - 启用分页功能

## 最佳实践

1. **错误处理**: 使用统一的错误类型和国际化消息
2. **参数验证**: 利用结构体标签进行自动验证和默认值设置
3. **分页**: 对列表接口使用 `WithPage()` 选项
4. **路由组织**: 使用路由分组管理相关的 API
5. **中间件**: 合理使用中间件处理横切关注点
6. **文档**: 集成 Swagger 文档提供 API 参考

## 依赖

- [Echo](https://github.com/labstack/echo) - Web 框架
- [go-playground/validator](https://github.com/go-playground/validator) - 参数验证
- [go-i18n](https://github.com/nicksnyder/go-i18n) - 国际化支持
- [xid](https://github.com/rs/xid) - 唯一 ID 生成
