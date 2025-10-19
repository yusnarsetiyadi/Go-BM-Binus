package ahphistory

import (
	"bm_binus/internal/abstraction"
	"bm_binus/internal/dto"
	"bm_binus/internal/factory"
	"bm_binus/pkg/util/response"
	"net/http"

	"github.com/labstack/echo/v4"
)

type handler struct {
	service Service
}

func NewHandler(f *factory.Factory) *handler {
	return &handler{
		service: NewService(f),
	}
}

func (h *handler) Create(c echo.Context) (err error) {
	payload := new(dto.AhpHistoryCreateRequest)
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

func (h handler) Find(c echo.Context) (err error) {
	data, err := h.service.Find(c.(*abstraction.Context))
	if err != nil {
		return response.ErrorResponse(err).SendError(c)
	}
	return response.SuccessResponse(data).SendSuccess(c)
}

func (h handler) FindById(c echo.Context) (err error) {
	payload := new(dto.AhpHistoryFindByIDRequest)
	if err := c.Bind(payload); err != nil {
		return response.ErrorBuilder(http.StatusBadRequest, err, "error bind payload").SendError(c)
	}
	if err = c.Validate(payload); err != nil {
		return response.ErrorBuilder(http.StatusBadRequest, err, "error validate payload").SendError(c)
	}
	data, err := h.service.FindById(c.(*abstraction.Context), payload)
	if err != nil {
		return response.ErrorResponse(err).SendError(c)
	}
	return response.SuccessResponse(data).SendSuccess(c)
}

func (h handler) Delete(c echo.Context) (err error) {
	payload := new(dto.AhpHistoryDeleteByIDRequest)
	if err := c.Bind(payload); err != nil {
		return response.ErrorBuilder(http.StatusBadRequest, err, "error bind payload").SendError(c)
	}
	if err = c.Validate(payload); err != nil {
		return response.ErrorBuilder(http.StatusBadRequest, err, "error validate payload").SendError(c)
	}
	data, err := h.service.Delete(c.(*abstraction.Context), payload)
	if err != nil {
		return response.ErrorResponse(err).SendError(c)
	}
	return response.SuccessResponse(data).SendSuccess(c)
}
