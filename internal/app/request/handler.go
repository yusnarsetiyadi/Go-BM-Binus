package request

import (
	"bm_binus/internal/abstraction"
	"bm_binus/internal/app/request/comment"
	"bm_binus/internal/app/request/event_type"
	"bm_binus/internal/app/request/file"
	"bm_binus/internal/dto"
	"bm_binus/internal/factory"
	"bm_binus/pkg/util/response"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type handler struct {
	service Service

	EventTypeHandler event_type.Handler
	CommentHandler   comment.Handler
	FileHandler      file.Handler
}

func NewHandler(f *factory.Factory) *handler {
	return &handler{
		service: NewService(f),

		EventTypeHandler: *event_type.NewHandler(f),
		CommentHandler:   *comment.NewHandler(f),
		FileHandler:      *file.NewHandler(f),
	}
}

func (h *handler) Create(c echo.Context) (err error) {
	payload := new(dto.RequestCreateRequest)

	if err = c.Bind(payload); err != nil {
		return response.ErrorBuilder(http.StatusBadRequest, err, "error bind payload").SendError(c)
	}
	if err = c.Validate(payload); err != nil {
		return response.ErrorBuilder(http.StatusBadRequest, err, "error validate payload").SendError(c)
	}

	contentType := c.Request().Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		if err := c.Request().ParseMultipartForm(64 << 20); err != nil {
			return response.ErrorBuilder(http.StatusBadRequest, err, "error bind multipart/form-data").SendError(c)
		}
		payload.Files = c.Request().MultipartForm.File["files"]
	}

	data, err := h.service.Create(c.(*abstraction.Context), payload)
	if err != nil {
		return response.ErrorResponse(err).SendError(c)
	}
	return response.SuccessResponse(data).SendSuccess(c)
}
