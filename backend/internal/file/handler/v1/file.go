package v1

import (
	"fmt"
	"io"
	"net/url"
	"strconv"

	"github.com/GoYoko/web"
	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/errcode"
	"github.com/chaitin/WhaleHire/backend/internal/file/usecase"
	"github.com/chaitin/WhaleHire/backend/internal/middleware"
)

type FileHandler struct {
	usecase *usecase.FileUsecase
}

func NewFileHandler(
	w *web.Web,
	usecase *usecase.FileUsecase,
	auth *middleware.AuthMiddleware,
) *FileHandler {
	h := &FileHandler{usecase: usecase}
	g := w.Group("/api/v1/file")
	g.Use(auth.UserAuth())
	g.POST("/upload", web.BaseHandler(h.Upload))
	g.GET("/downloadurl", web.BaseHandler(h.GetDownloadURL))
	g.GET("/stream/:key", web.BaseHandler(h.StreamDownload))
	return h
}

// Upload 上传文件
//
// @Summary		Upload File
// @Description	Upload File
// @Tags			File
// @Accept			multipart/form-data
// @Param			file	formData	file	true	"File"
// @Param			kb_id	formData	string	false	"Knowledge Base ID"
// @Success		200		{object}	domain.ObjectUploadResp
// @Router			/api/v1/file/upload [post]
func (h *FileHandler) Upload(c *web.Context) error {
	ctx := c.Request().Context()
	kbID := c.FormValue("kb_id")
	if kbID == "" {
		kbID = uuid.New().String()
	}
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}
	key, err := h.usecase.UploadFile(ctx, kbID, file)
	if err != nil {
		return err
	}
	return c.Success(domain.ObjectUploadResp{Key: key})
}

// GetDownloadURL 获取文件下载链接
//
//	@Summary		Get Download URL
//	@Description	Get presigned download URL
//	@Tags			File
//	@Accept			json
//	@Produce		json
//	@Param			key	query		string	true	"File Key"
//	@Success		200	{object}	domain.ObjectDownloadResp
//	@Router			/api/v1/file/download [get]
func (h *FileHandler) GetDownloadURL(c *web.Context) error {
	ctx := c.Request().Context()
	key := c.QueryParam("key")
	if key == "" {
		return errcode.ErrMissKey.Wrap(fmt.Errorf("missing key"))
	}
	url, err := h.usecase.GenerateDownloadURL(ctx, key)
	if err != nil {
		return err
	}
	return c.Success(domain.ObjectDownloadResp{URL: url})
}

// StreamDownload 流式下载文件
//
//	@Summary		Stream Download File
//	@Description	Stream download file by key
//	@Tags			File
//	@Accept			json
//	@Produce		application/octet-stream
//	@Param			key	path	string	true	"File Key (URL encoded)"
//	@Success		200	{file}	binary
//	@Router			/api/v1/file/proxy/{key} [get]
func (h *FileHandler) StreamDownload(c *web.Context) error {
	ctx := c.Request().Context()

	// 获取URL编码的key参数
	encodedKey := c.Param("key")
	if encodedKey == "" {
		return errcode.ErrMissKey.Wrap(fmt.Errorf("missing key"))
	}

	// URL解码
	key, err := url.QueryUnescape(encodedKey)
	if err != nil {
		return errcode.ErrMissKey.Wrap(fmt.Errorf("invalid key format: %w", err))
	}

	// 下载文件
	reader, fileInfo, err := h.usecase.DownloadFile(ctx, key)
	if err != nil {
		return err
	}
	defer reader.Close()

	// 设置响应头
	c.Response().Header().Set("Content-Type", fileInfo.ContentType)
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileInfo.OriginalName))
	c.Response().Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size, 10))

	// 流式传输文件内容
	_, err = io.Copy(c.Response(), reader)
	if err != nil {
		return fmt.Errorf("failed to stream file: %w", err)
	}

	return nil
}
