package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
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
	g.DELETE("/conversations", web.BaseHandler(h.DeleteConversation), auth.UserAuth())
	g.POST("/conversations/:id/addmessage", web.BindHandler(h.AddMessageToConversation), auth.UserAuth())

	return h
}

// Generate 生成AI回复
//
//	@Tags			General Agent
//	@Summary		生成AI回复
//	@Description	根据用户输入的提示词和可选的历史对话记录，生成AI智能体的回复内容。支持传入历史消息以保持对话上下文连贯性。
//	@ID				generate
//	@Accept			json
//	@Produce		json
//	@Param			param	body		domain.GenerateReq	true	"生成请求参数，包含prompt(提示词，必填)和history(历史消息数组，可选)"
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

// GenerateStream 流式生成AI回复
//
//	@Tags			General Agent
//	@Summary		流式生成AI回复
//	@Description	以Server-Sent Events(SSE)方式流式生成AI回复，实时返回生成的内容片段。客户端可以实时接收并显示生成过程，提供更好的用户体验。支持超时控制和错误处理。
//	@ID				generate-stream
//	@Accept			json
//	@Produce		text/event-stream
//	@Param			param	body		domain.GenerateReq	true	"生成请求参数，包含prompt(提示词，必填)和history(历史消息数组，可选)"
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

	// 处理conversation_id逻辑
	var conversationID string
	if req.ConversationID != nil && *req.ConversationID != "" {
		// 使用现有对话ID
		conversationID = *req.ConversationID
	} else {
		// 创建新对话
		conv, err := h.usecase.CreateConversation(ctx.Request().Context(), user.ID, &domain.CreateConversationReq{
			Title: "New Conversation", // 可以根据prompt生成更合适的标题
		})
		if err != nil {
			return err
		}
		conversationID = conv.ID
		if err := h.writeSSEEvent(ctx, "data", domain.StreamMetadata{
			Version:        "v1",
			ConversationID: conversationID,
		}); err != nil {
			return err
		}

	}

	// 获取流式读取器
	streamReader, err := h.usecase.GenerateStream(ctx.Request().Context(), &req)
	if err != nil {
		return err
	}
	defer streamReader.Close()

	// 创建带超时的上下文
	ctx_timeout, cancel := context.WithTimeout(ctx.Request().Context(), 5*time.Minute)
	defer cancel()

	// 用于收集完整的AI回复内容
	var fullResponse strings.Builder

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

			// 收集完整回复内容
			fullResponse.WriteString(chunk.Content)

			// 发送数据事件
			response := domain.StreamChunk{
				Content: chunk.Content,
				Done:    chunk.Done,
			}

			if err := h.writeSSEEvent(ctx, "data", response); err != nil {
				return err
			}

			// 如果完成，保存消息到对话并发送完成事件
			if chunk.Done {
				// 保存用户消息
				userMessage := &domain.Message{
					ConversationID: conversationID,
					Role:           "user",
					Content:        &req.Prompt,
					Type:           "text",
				}
				if err := h.usecase.AddMessageToConversation(ctx.Request().Context(), &domain.AddMessageToConversationReq{
					ConversationID: conversationID,
					Message:        userMessage,
				}); err != nil {
					h.logger.Error("Failed to save user message", "error", err)
				}

				// 保存AI回复消息
				aiResponse := fullResponse.String()
				assistantMessage := &domain.Message{
					ConversationID: conversationID,
					Role:           "assistant",
					Content:        &aiResponse,
					Type:           "text",
				}
				if err := h.usecase.AddMessageToConversation(ctx.Request().Context(), &domain.AddMessageToConversationReq{
					ConversationID: conversationID,
					Message:        assistantMessage,
				}); err != nil {
					h.logger.Error("Failed to save assistant message", "error", err)
				}

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

// CreateConversation 创建新对话会话
//
//	@Tags			General Agent
//	@Summary		创建新对话会话
//	@Description	为当前用户创建一个新的对话会话，用于管理和组织AI对话历史。每个对话会话都有唯一的ID和标题，可以包含多条消息记录。
//	@ID				create-conversation
//	@Accept			json
//	@Produce		json
//	@Param			param	body		domain.CreateConversationReq	true	"创建对话请求参数，包含title(对话标题，必填，最大256字符)"
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

// ListConversations 分页获取用户对话列表
//
//	@Tags			General Agent
//	@Summary		分页获取用户对话列表
//	@Description	分页获取当前用户的所有对话会话列表，支持按标题搜索和分页查询。返回对话的基本信息包括ID、标题、状态、创建时间等。
//	@ID				list-conversations
//	@Accept			json
//	@Produce		json
//	@Param			page	query		int		false	"页码，默认1"
//	@Param			size	query		int		false	"每页数量，默认10"
//	@Param			search	query		string	false	"搜索关键词，按对话标题模糊匹配"
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

// GetConversationHistory 获取指定对话的完整历史记录
//
//	@Tags			General Agent
//	@Summary		获取对话历史记录
//	@Description	根据对话ID获取指定对话的完整历史记录，包括所有消息内容、附件信息、消息类型等详细信息。用于恢复和查看完整的对话上下文。
//	@ID				get-conversation-history
//	@Accept			json
//	@Produce		json
//	@Param			conversation_id	query		string	true	"对话ID，用于指定要获取历史记录的对话"
//	@Success		200				{object}	web.Resp{data=domain.Conversation}
//	@Router			/api/v1/general-agent/conversations/history [post]
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

// DeleteConversation 删除指定对话
//
//	@Tags			General Agent
//	@Summary		删除指定对话
//	@Description	根据对话ID删除指定的对话会话及其所有相关的消息记录和附件。此操作不可逆，请谨慎使用。只有对话的创建者才能删除该对话。
//	@ID				delete-conversation
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"对话ID，要删除的对话的唯一标识符"
//	@Success		200	{object}	web.Resp
//	@Router			/api/v1/general-agent/conversations/{id} [delete]
func (h *GeneralAgentHandler) DeleteConversation(c *web.Context) error {
	user := middleware.GetUser(c)
	if user == nil {
		return errcode.ErrPermission
	}

	h.logger.Info(c.QueryParam("id"))
	if err := h.usecase.DeleteConversation(c.Request().Context(), &domain.DeleteConversationReq{
		ConversationID: c.QueryParam("id"),
	}); err != nil {
		return err
	}

	return c.Success(nil)
}

// AddMessageToConversation 向指定对话添加新消息
//
//	@Tags			General Agent
//	@Summary		向对话添加新消息
//	@Description	向指定的对话会话中添加新的消息记录。支持多种消息类型(文本、图片、音频、视频、文件)和角色(用户、助手、系统、智能体)。可以包含附件和元数据信息。
//	@ID				add-message-to-conversation
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string								true	"对话ID，要添加消息的目标对话的唯一标识符"
//	@Param			param	body		domain.AddMessageToConversationReq	true	"添加消息请求参数，包含conversation_id(对话ID)和message(消息对象，包含角色、内容、类型等信息)"
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
