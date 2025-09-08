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
	"github.com/ptonlix/whalehire/backend/internal/general_agent/usecase"
	"github.com/ptonlix/whalehire/backend/internal/middleware"
	"github.com/ptonlix/whalehire/backend/pkg/session"
)

// GeneralAgentHandler 通用智能体处理器
type GeneralAgentHandler struct {
	usecase *usecase.GeneralAgentUsecase
	session *session.Session
	logger  *slog.Logger
	cfg     *config.Config
	limiter *rate.Limiter
}

// NewGeneralAgentHandler 创建通用智能体处理器
func NewGeneralAgentHandler(
	w *web.Web,
	uc *usecase.GeneralAgentUsecase,
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
func (h *GeneralAgentHandler) Generate(c *web.Context, req domain.GenerateReq) error {
	generateReq := &domain.GenerateReq{
		Prompt:  req.Prompt,
		History: req.History,
	}

	resp, err := h.usecase.Generate(c.Request().Context(), generateReq)
	if err != nil {
		return err
	}

	return c.Success(domain.GenerateResp{
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
func (h *GeneralAgentHandler) GenerateStream(c *web.Context, req domain.GenerateReq) error {
	generateReq := &domain.GenerateReq{
		Prompt:  req.Prompt,
		History: req.History,
	}

	// 设置SSE响应头
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().Header().Set("Access-Control-Allow-Origin", "*")
	c.Response().Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	// 获取流式读取器
	streamReader, err := h.usecase.GenerateStream(c.Request().Context(), generateReq)
	if err != nil {
		return err
	}
	defer streamReader.Close()

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Minute)
	defer cancel()

	// 流式发送数据
	for {
		select {
		case <-ctx.Done():
			// 发送超时或取消事件
			if err := h.writeSSEEvent(c, "error", map[string]interface{}{
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
				if writeErr := h.writeSSEEvent(c, "error", map[string]interface{}{
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

			if err := h.writeSSEEvent(c, "data", response); err != nil {
				return err
			}

			// 如果完成，发送完成事件并退出
			if chunk.Done {
				if err := h.writeSSEEvent(c, "done", map[string]interface{}{
					"message": "Stream completed",
				}); err != nil {
					return err
				}
				return nil
			}

			// 刷新响应
			c.Response().Flush()
		}
	}
}

// writeSSEEvent 写入SSE事件
func (h *GeneralAgentHandler) writeSSEEvent(c *web.Context, event string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// 写入SSE格式的数据
	_, err = fmt.Fprintf(c.Response().Writer, "event: %s\ndata: %s\n\n", event, string(jsonData))
	return err
}
