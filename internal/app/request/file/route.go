package file

import (
	"bm_binus/internal/middleware"

	"github.com/labstack/echo/v4"
)

func (h *Handler) Route(v *echo.Group) {
	v.POST("", h.Create, middleware.Authentication)
	v.GET("/:request_id", h.FindByRequestId, middleware.Authentication)
	v.DELETE("/:id", h.Delete, middleware.Authentication)
	v.PUT("/:id", h.Update, middleware.Authentication)
}
