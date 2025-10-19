package ahphistory

import (
	"bm_binus/internal/middleware"

	"github.com/labstack/echo/v4"
)

func (h *handler) Route(v *echo.Group) {
	v.POST("", h.Create, middleware.Authentication)
	v.GET("", h.Find, middleware.Authentication)
	v.GET("/:id", h.FindById, middleware.Authentication)
	v.DELETE("/:id", h.Delete, middleware.Authentication)
}
