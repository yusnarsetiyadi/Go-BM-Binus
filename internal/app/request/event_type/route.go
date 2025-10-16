package event_type

import (
	"bm_binus/internal/middleware"

	"github.com/labstack/echo/v4"
)

func (h *Handler) Route(v *echo.Group) {
	v.GET("", h.Find, middleware.Authentication)
	v.POST("", h.Create, middleware.Authentication)
	v.DELETE("/:id", h.Delete, middleware.Authentication)
	v.PUT("/:id", h.Update, middleware.Authentication)
}
