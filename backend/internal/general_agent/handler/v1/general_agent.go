package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/GoYoko/web"
	"golang.org/x/time/rate"

	"github.com/ptonlix/whalehire/backend/config"
	"github.com/ptonlix/whalehire/backend/domain"
	"github.com/ptonlix/whalehire/backend/errcode"
	"github.com/ptonlix/whalehire/backend/internal/middleware"
	"github.com/ptonlix/whalehire/backend/pkg/session"
)

// GeneralAgentHandler 通用智能体处理器
type GeneralAgentHandler struct {
	usecase domain.GeneralAgentUsecase
	session *session.Session
	logger  *slog.Logger
	cfg     *config.Config
	limiter *rate.Limiter
}

// NewGeneralAgentHandler 创建通用智能体处理器
func NewGeneralAgentHandler(
	w *web.Web,
	uc domain.GeneralAgentUsecase,
	auth *middleware.AuthMiddleware,
	session *session.Session,
	logger *slog.Logger,
	cfg *config.Config,
) *GeneralAgentHandler {
	h := &GeneralAgentHandler{
		usecase: uc,
		session: session,
		logger:  logger,
		cfg:     cfg,
		limiter: rate.NewLimiter(rate.Every(10*time.Second), 1),
	}

	// 注册路由
	g := w.Group("/api/v1/general-agent")
	g.POST("/generate", web.BindHandler(h.Generate), auth.UserAuth())
	g.POST("/generate-stream", web.BindHandler(h.GenerateStream), auth.UserAuth())

	// 对话记录管理路由
	g.POST("/conversations", web.BindHandler(h.CreateConversation), auth.UserAuth())
	g.GET("/conversations", web.BindHandler(h.ListConversations), auth.UserAuth())
	g.GET("/conversations/history", web.BindHandler(h.GetConversationHistory), auth.UserAuth())
	g.DELETE("/conversations/:id", web.BindHandler(h.DeleteConversation), auth.UserAuth())
	g.POST("/conversations/:id/addmessage", web.BindHandler(h.AddMessageToConversation), auth.UserAuth())

	return h
}

// Generate 生成回复
//
//	@Tags			General Agent
//	@Summary		生成回复
//	@Description	生成回复
//	@ID				generate
//	@Accept			json
//	@Produce		json
//	@Param			param	body		domain.GenerateReq	true	"生成请求参数"
//	@Success		200		{object}	web.Resp{data=domain.GenerateResp}
//	@Router			/api/v1/general-agent/generate [post]
func (h *GeneralAgentHandler) Generate(ctx *web.Context, req domain.GenerateReq) error {
	user := middleware.GetUser(ctx)
	if user == nil {
		return errcode.ErrPermission
	}

	resp, err := h.usecase.Generate(ctx.Request().Context(), &req)
	if err != nil {
		return err
	}

	return ctx.Success(domain.GenerateResp{
		Answer: resp.Answer,
	})
}

// GenerateStream 流式生成回复
//
//	@Tags			General Agent
//	@Summary		流式生成回复
//	@Description	流式生成回复
//	@ID				generate-stream
//	@Accept			json
//	@Produce		text/event-stream
//	@Param			param	body		domain.GenerateReq	true	"生成请求参数"
//	@Success		200		{object}	domain.StreamChunk
//	@Router			/api/v1/general-agent/generate-stream [post]
func (h *GeneralAgentHandler) GenerateStream(ctx *web.Context, req domain.GenerateReq) error {
	user := middleware.GetUser(ctx)
	if user == nil {
		return errcode.ErrPermission
	}

	// 设置SSE响应头
	ctx.Response().Header().Set("Content-Type", "text/event-stream")
	ctx.Response().Header().Set("Cache-Control", "no-cache")
	ctx.Response().Header().Set("Connection", "keep-alive")
	ctx.Response().Header().Set("Access-Control-Allow-Origin", "*")
	ctx.Response().Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	// 获取流式读取器
	streamReader, err := h.usecase.GenerateStream(ctx.Request().Context(), &req)
	if err != nil {
		return err
	}
	defer streamReader.Close()

	// 创建带超时的上下文
	ctx_timeout, cancel := context.WithTimeout(ctx.Request().Context(), 5*time.Minute)
	defer cancel()

	// 流式发送数据
	for {
		select {
		case <-ctx_timeout.Done():
			// 发送超时或取消事件
			if err := h.writeSSEEvent(ctx, "error", map[string]interface{}{
				"message": "Request timeout or cancelled",
			}); err != nil {
				return err
			}
			return nil
		default:
			// 接收流式数据
			chunk, err := streamReader.Recv()
			if err != nil {
				// 发送错误事件
				if writeErr := h.writeSSEEvent(ctx, "error", map[string]interface{}{
					"message": "Stream error: " + err.Error(),
				}); writeErr != nil {
					return writeErr
				}
				return nil
			}

			// 发送数据事件
			response := domain.StreamChunk{
				Content: chunk.Content,
				Done:    chunk.Done,
			}

			if err := h.writeSSEEvent(ctx, "data", response); err != nil {
				return err
			}

			// 如果完成，发送完成事件并退出
			if chunk.Done {
				if err := h.writeSSEEvent(ctx, "done", map[string]interface{}{
					"message": "Stream completed",
				}); err != nil {
					return err
				}
				return nil
			}

			// 刷新响应
			ctx.Response().Flush()
		}
	}
}

// writeSSEEvent 写入SSE事件
func (h *GeneralAgentHandler) writeSSEEvent(ctx *web.Context, event string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// 写入SSE格式的数据
	_, err = fmt.Fprintf(ctx.Response().Writer, "event: %s\ndata: %s\n\n", event, string(jsonData))
	return err
}

// CreateConversation 创建新对话
//
//	@Tags			General Agent
//	@Summary		创建新对话
//	@Description	创建新对话
//	@ID				create-conversation
//	@Accept			json
//	@Produce		json
//	@Param			param	body		domain.CreateConversationReq	true	"创建对话请求参数"
//	@Success		200		{object}	web.Resp{data=domain.Conversation}
//	@Router			/api/v1/general-agent/conversations [post]
func (h *GeneralAgentHandler) CreateConversation(ctx *web.Context, req domain.CreateConversationReq) error {
	user := middleware.GetUser(ctx)
	if user == nil {
		return errcode.ErrPermission
	}

	conv, err := h.usecase.CreateConversation(ctx.Request().Context(), user.ID, &req)
	if err != nil {
		return err
	}

	return ctx.Success(conv)
}

// ListConversations 分页获取对话列表
//
//	@Tags			General Agent
//	@Summary		获取对话列表
//	@Description	分页获取用户的对话列表
//	@ID				list-conversations
//	@Accept			json
//	@Produce		json
//	@Param			page	query		web.Pagination	true	"分页"
//	@Success		200		{object}	web.Resp{data=domain.ListConversationsResp}
//	@Router			/api/v1/general-agent/conversations [get]
func (h *GeneralAgentHandler) ListConversations(c *web.Context, req domain.ListConversationsReq) error {
	user := middleware.GetUser(c)
	if user == nil {
		return errcode.ErrPermission
	}

	resp, err := h.usecase.ListConversations(c.Request().Context(), user.ID, &req)
	if err != nil {
		return err
	}

	return c.Success(resp)
}

// GetConversationHistory 获取对话历史
//
//	@Tags			General Agent
//	@Summary		获取对话历史
//	@Description	获取用户最近的对话历史
//	@ID				get-conversation-history
//	@Accept			json
//	@Produce		json
//	@Param			limit	query	int	false	"限制数量"	default(10)
//	@Success		200	{object}	web.Resp{data=[]domain.Conversation}
//	@Router			/api/v1/general-agent/conversations/history [get]
func (h *GeneralAgentHandler) GetConversationHistory(c *web.Context, req domain.GetConversationHistoryReq) error {
	user := middleware.GetUser(c)
	if user == nil {
		return errcode.ErrPermission
	}
	conversations, err := h.usecase.GetConversationHistory(c.Request().Context(), &req)
	if err != nil {
		return err
	}

	return c.Success(conversations)
}

// DeleteConversation 删除对话
//
//	@Tags			General Agent
//	@Summary		删除对话
//	@Description	删除指定的对话
//	@ID				delete-conversation
//	@Accept			json
//	@Produce		json
//	@Param			id	path	string	true	"对话ID"
//	@Success		200	{object}	web.Resp
//	@Router			/api/v1/general-agent/conversations/{id} [delete]
func (h *GeneralAgentHandler) DeleteConversation(c *web.Context, req domain.DeleteConversationReq) error {
	user := middleware.GetUser(c)
	if user == nil {
		return errcode.ErrPermission
	}

	if err := h.usecase.DeleteConversation(c.Request().Context(), &req); err != nil {
		return err
	}

	return c.Success(nil)
}

// AddMessageToConversation 向对话添加消息
//
//	@Tags			General Agent
//	@Summary		向对话添加消息
//	@Description	向指定对话添加新消息
//	@ID				add-message-to-conversation
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"对话ID"
//	@Param			param	body		AddMessageToConversationReq	true	"添加消息请求参数"
//	@Success		200		{object}	web.Resp
//	@Router			/api/v1/general-agent/conversations/{id}/addmessage [post]
func (h *GeneralAgentHandler) AddMessageToConversation(ctx *web.Context, req domain.AddMessageToConversationReq) error {
	user := middleware.GetUser(ctx)
	if user == nil {
		return errcode.ErrPermission
	}

	if err := h.usecase.AddMessageToConversation(ctx.Request().Context(), &req); err != nil {
		return err
	}

	return ctx.Success(nil)
}

// GenerateWithConversationReq 在对话中生成回复请求
type GenerateWithConversationReq struct {
	ID     string `param:"id" validate:"required"`
	Prompt string `json:"prompt" validate:"required,max=4000"`
}
