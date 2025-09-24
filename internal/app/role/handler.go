package role

import (
	"bm_binus/internal/abstraction"
	"bm_binus/internal/factory"
	"bm_binus/pkg/util/response"

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

func (h handler) Find(c echo.Context) (err error) {
	data, err := h.service.Find(c.(*abstraction.Context))
	if err != nil {
		return response.ErrorResponse(err).SendError(c)
	}
	return response.SuccessResponse(data).SendSuccess(c)
}

func (h handler) Export(c echo.Context) (err error) {
	filename, data, err := h.service.Export(c.(*abstraction.Context))
	if err != nil {
		return response.ErrorResponse(err).SendError(c)
	}
	return response.SendExcelData(c, filename, *data)
}
