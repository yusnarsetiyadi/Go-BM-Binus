package event_type

import (
	"bm_binus/internal/abstraction"
	"bm_binus/internal/factory"
	"bm_binus/pkg/util/response"

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

func (h Handler) Find(c echo.Context) (err error) {
	data, err := h.service.Find(c.(*abstraction.Context))
	if err != nil {
		return response.ErrorResponse(err).SendError(c)
	}
	return response.SuccessResponse(data).SendSuccess(c)
}

// func (h *Handler) Create(c echo.Context) (err error) {
// 	payload := new(dto.EventTypeCreateRequest)
// 	if err = c.Bind(payload); err != nil {
// 		return response.ErrorBuilder(http.StatusBadRequest, err, "error bind payload").SendError(c)
// 	}
// 	if err = c.Validate(payload); err != nil {
// 		return response.ErrorBuilder(http.StatusBadRequest, err, "error validate payload").SendError(c)
// 	}
// 	data, err := h.service.Create(c.(*abstraction.Context), payload)
// 	if err != nil {
// 		return response.ErrorResponse(err).SendError(c)
// 	}
// 	return response.SuccessResponse(data).SendSuccess(c)
// }

// func (h Handler) Delete(c echo.Context) (err error) {
// 	payload := new(dto.TaskTypeDeleteByIDRequest)
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
// 	payload := new(dto.TaskTypeUpdateRequest)
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
