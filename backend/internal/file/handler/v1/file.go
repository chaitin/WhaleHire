package v1

import (
	"fmt"

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
	g.GET("/download", web.BaseHandler(h.Download))
	return h
}

// Upload
//
//	@Summary		Upload File
//	@Description	Upload File
//	@Tags			File
//	@Accept			multipart/form-data
//	@Param			file	formData	file	true	"File"
//	@Param			kb_id	formData	string	false	"Knowledge Base ID"
//	@Success		200		{object}	domain.ObjectUploadResp
//	@Router			/api/v1/file/upload [post]
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

// Download
//
//	@Summary		Download File
//	@Description	Get presigned download URL
//	@Tags			File
//	@Accept			json
//	@Produce		json
//	@Param			key	query		string	true	"File Key"
//	@Success		200	{object}	domain.ObjectDownloadResp
//	@Router			/api/v1/file/download [get]
func (h *FileHandler) Download(c *web.Context) error {
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
