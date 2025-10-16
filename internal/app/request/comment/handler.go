package comment

import (
	"bm_binus/internal/abstraction"
	"bm_binus/internal/dto"
	"bm_binus/internal/factory"
	"bm_binus/pkg/util/response"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	service Service
}

func NewHandler(f *factory.Factory) *Handler {
	return &Handler{
		service: NewService(f),
	}
}

func (h *Handler) Create(c echo.Context) (err error) {
	payload := new(dto.CommentCreateRequest)
	if err = c.Bind(payload); err != nil {
		return response.ErrorBuilder(http.StatusBadRequest, err, "error bind payload").SendError(c)
	}
	if err = c.Validate(payload); err != nil {
		return response.ErrorBuilder(http.StatusBadRequest, err, "error validate payload").SendError(c)
	}
	data, err := h.service.Create(c.(*abstraction.Context), payload)
	if err != nil {
		return response.ErrorResponse(err).SendError(c)
	}
	return response.SuccessResponse(data).SendSuccess(c)
}

// func (h Handler) FindByRequestId(c echo.Context) (err error) {
// 	payload := new(dto.CommentFindByRequestIDRequest)
// 	if err := c.Bind(payload); err != nil {
// 		return response.ErrorBuilder(http.StatusBadRequest, err, "error bind payload").SendError(c)
// 	}
// 	if err = c.Validate(payload); err != nil {
// 		return response.ErrorBuilder(http.StatusBadRequest, err, "error validate payload").SendError(c)
// 	}
// 	data, err := h.service.FindByRequestId(c.(*abstraction.Context), payload)
// 	if err != nil {
// 		return response.ErrorResponse(err).SendError(c)
// 	}
// 	return response.SuccessResponse(data).SendSuccess(c)
// }

// func (h Handler) Delete(c echo.Context) (err error) {
// 	payload := new(dto.CommentDeleteByIDRequest)
// 	if err := c.Bind(payload); err != nil {
// 		return response.ErrorBuilder(http.StatusBadRequest, err, "error bind payload").SendError(c)
// 	}
// 	if err = c.Validate(payload); err != nil {
// 		return response.ErrorBuilder(http.StatusBadRequest, err, "error validate payload").SendError(c)
// 	}
// 	data, err := h.service.Delete(c.(*abstraction.Context), payload)
// 	if err != nil {
// 		return response.ErrorResponse(err).SendError(c)
// 	}
// 	return response.SuccessResponse(data).SendSuccess(c)
// }

// func (h Handler) Update(c echo.Context) (err error) {
// 	payload := new(dto.CommentUpdateRequest)
// 	if err = c.Bind(payload); err != nil {
// 		return response.ErrorBuilder(http.StatusBadRequest, err, "error bind payload").SendError(c)
// 	}
// 	if err = c.Validate(payload); err != nil {
// 		return response.ErrorBuilder(http.StatusBadRequest, err, "error validate payload").SendError(c)
// 	}
// 	data, err := h.service.Update(c.(*abstraction.Context), payload)
// 	if err != nil {
// 		return response.ErrorResponse(err).SendError(c)
// 	}
// 	return response.SuccessResponse(data).SendSuccess(c)
// }
